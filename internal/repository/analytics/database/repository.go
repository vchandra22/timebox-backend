package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"timebox-backend/internal/config"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/repository/analytics"
	"timebox-backend/internal/repository/dbexecutor"
)

type Repository struct {
	db         config.PostgreSQL
	dbExecutor *dbexecutor.Executor
}

func NewRepository(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{db: db, dbExecutor: dbExecutor}
}

func (r *Repository) Streak(ctx context.Context, workspaceID, userID string) (entity.Streak, error) {
	var row StreakRow
	err := r.db.Conn.GetContext(ctx, &row, `
		SELECT COUNT(DISTINCT scheduled_start::date)::int AS completed_days, MAX(scheduled_start::date) AS last_date
		FROM timeboxes
		WHERE workspace_id = $1 AND owner_id = $2 AND status = 'completed' AND deleted_at IS NULL
	`, workspaceID, userID)
	if err != nil {
		return entity.Streak{}, analyticsError(err)
	}
	return entity.Streak{WorkspaceID: workspaceID, CurrentStreak: row.CompletedDays, LongestStreak: row.CompletedDays, LastCompletedDate: row.LastDate}, nil
}

func (r *Repository) Leaderboard(ctx context.Context, workspaceID, metric, period string, limit int) (entity.Leaderboard, error) {
	start := periodStart(period)
	var rows []LeaderboardRow
	err := r.db.Conn.SelectContext(ctx, &rows, `
		SELECT u.id::text AS user_id, u.full_name, u.avatar_url,
			CASE WHEN $3 = 'completed_timeboxes' THEN COUNT(tb.id)::int ELSE COALESCE(SUM(tb.actual_minutes), 0)::int END AS value
		FROM workspace_members wm
		JOIN users u ON u.id = wm.user_id
		LEFT JOIN timeboxes tb ON tb.owner_id = wm.user_id AND tb.workspace_id = wm.workspace_id AND tb.deleted_at IS NULL AND tb.scheduled_start >= $2
		WHERE wm.workspace_id = $1 AND wm.status = 'active'
		GROUP BY u.id, u.full_name, u.avatar_url
		ORDER BY value DESC
		LIMIT $4
	`, workspaceID, start, metric, limit)
	if err != nil {
		return entity.Leaderboard{}, analyticsError(err)
	}
	items := make([]entity.LeaderboardItem, 0, len(rows))
	for i, row := range rows {
		items = append(items, entity.LeaderboardItem{Rank: i + 1, User: entity.UserSummary{ID: row.UserID, FullName: row.FullName, AvatarURL: row.AvatarURL}, Value: row.Value})
	}
	return entity.Leaderboard{Metric: metric, Period: period, Items: items}, nil
}

func (r *Repository) PersonalDashboard(ctx context.Context, workspaceID, userID string, date time.Time) (entity.PersonalDashboard, error) {
	summary, _ := r.summary(ctx, workspaceID, userID, date, date.AddDate(0, 0, 1))
	timeline := r.timeline(ctx, workspaceID, userID, date, date.AddDate(0, 0, 1))
	categories := r.categoryDistribution(ctx, workspaceID, userID, date, date.AddDate(0, 0, 1))
	streak, _ := r.Streak(ctx, workspaceID, userID)
	summary.CurrentStreak = streak.CurrentStreak
	return entity.PersonalDashboard{Date: date.Format("2006-01-02"), Timezone: "UTC", TodayTimeline: timeline, Summary: summary, CategoryDistribution: categories, OverrunOrMissed: nil}, nil
}

func (r *Repository) WorkspaceDashboard(ctx context.Context, workspaceID string, date time.Time) (entity.WorkspaceDashboard, error) {
	summary, _ := r.summary(ctx, workspaceID, "", date, date.AddDate(0, 0, 1))
	weekSummary, _ := r.summary(ctx, workspaceID, "", date.AddDate(0, 0, -7), date.AddDate(0, 0, 1))
	summary.TeamFocusMinutesToday = summary.ActualMinutes
	summary.TeamFocusMinutesWeek = weekSummary.ActualMinutes
	return entity.WorkspaceDashboard{Date: date.Format("2006-01-02"), Summary: summary, TeamToday: r.teamToday(ctx, workspaceID, date), CategoryDistribution: r.categoryDistribution(ctx, workspaceID, "", date, date.AddDate(0, 0, 1)), Trend: r.trend(ctx, workspaceID, date.AddDate(0, 0, -6), date.AddDate(0, 0, 1), "day")}, nil
}

func (r *Repository) TimeAllocationReport(ctx context.Context, filter analytics.ReportFilter) (entity.TimeAllocationReport, error) {
	items := r.reportItems(ctx, filter)
	summary := reportSummary(items)
	if summary.ActualMinutes > 0 {
		for i := range items {
			items[i].Percentage = (float64(items[i].ActualMinutes) / float64(summary.ActualMinutes)) * 100
		}
	}
	return entity.TimeAllocationReport{Range: reportRange(filter), GroupBy: filter.GroupBy, Summary: summary, Items: items}, nil
}

func (r *Repository) ProductivityTrendReport(ctx context.Context, filter analytics.ReportFilter) (entity.ProductivityTrendReport, error) {
	return entity.ProductivityTrendReport{Interval: filter.Interval, Items: r.trend(ctx, filter.WorkspaceID, filter.StartDate, filter.EndDate.AddDate(0, 0, 1), filter.Interval)}, nil
}

func (r *Repository) TeamWorkloadReport(ctx context.Context, filter analytics.ReportFilter) (entity.TeamWorkloadReport, error) {
	var rows []WorkloadRow
	_ = r.db.Conn.SelectContext(ctx, &rows, `
		SELECT u.id::text AS user_id, u.full_name, COALESCE(SUM(tb.planned_minutes), 0)::int AS planned_minutes,
			COALESCE(SUM(tb.actual_minutes), 0)::int AS actual_minutes,
			COUNT(tb.id) FILTER (WHERE tb.status = 'completed')::int AS completed_timeboxes,
			COUNT(tb.id) FILTER (WHERE tb.actual_minutes > tb.planned_minutes)::int AS overrun_count
		FROM workspace_members wm
		JOIN users u ON u.id = wm.user_id
		LEFT JOIN timeboxes tb ON tb.owner_id = wm.user_id AND tb.workspace_id = wm.workspace_id AND tb.deleted_at IS NULL AND tb.scheduled_start >= $2 AND tb.scheduled_start < $3
		WHERE wm.workspace_id = $1 AND wm.status = 'active'
		GROUP BY u.id, u.full_name
		ORDER BY actual_minutes DESC
	`, filter.WorkspaceID, filter.StartDate, filter.EndDate.AddDate(0, 0, 1))
	items := make([]entity.TeamWorkloadItem, 0, len(rows))
	for _, row := range rows {
		risk := "normal"
		if row.OverrunCount > 2 {
			risk = "high"
		}
		items = append(items, entity.TeamWorkloadItem{User: entity.UserSummary{ID: row.UserID, FullName: row.FullName}, PlannedMinutes: row.PlannedMinutes, ActualMinutes: row.ActualMinutes, CompletedTimeboxes: row.CompletedTimeboxes, OverrunCount: row.OverrunCount, RiskLevel: risk})
	}
	return entity.TeamWorkloadReport{Range: reportRange(filter), Items: items}, nil
}

func (r *Repository) Search(ctx context.Context, workspaceID, query string, types []string, limit int) (entity.SearchResults, error) {
	result := entity.SearchResults{Query: query}
	typeSet := map[string]bool{}
	for _, item := range types {
		typeSet[item] = true
	}
	all := len(typeSet) == 0
	if all || typeSet["task"] {
		result.Tasks = r.searchRows(ctx, `SELECT id::text, title, description AS snippet FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL AND (title ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%') LIMIT $3`, workspaceID, query, limit)
	}
	if all || typeSet["timebox"] {
		result.Timeboxes = r.searchRows(ctx, `SELECT id::text, title, description AS snippet FROM timeboxes WHERE workspace_id = $1 AND deleted_at IS NULL AND (title ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%') LIMIT $3`, workspaceID, query, limit)
	}
	if all || typeSet["goal"] {
		result.Goals = r.searchRows(ctx, `SELECT id::text, title, description AS snippet FROM goals WHERE workspace_id = $1 AND (title ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%') LIMIT $3`, workspaceID, query, limit)
	}
	if all || typeSet["comment"] {
		result.Comments = r.searchRows(ctx, `SELECT id::text, body AS title, body AS snippet FROM comments WHERE workspace_id = $1 AND deleted_at IS NULL AND body ILIKE '%' || $2 || '%' LIMIT $3`, workspaceID, query, limit)
	}
	if all || typeSet["attachment"] {
		result.Attachments = r.searchRows(ctx, `SELECT id::text, file_name AS title, url AS snippet FROM attachments WHERE workspace_id = $1 AND deleted_at IS NULL AND file_name ILIKE '%' || $2 || '%' LIMIT $3`, workspaceID, query, limit)
	}
	return result, nil
}

func (r *Repository) ListSavedViews(ctx context.Context, workspaceID, userID, resourceType string) ([]entity.SavedView, error) {
	var rows []SavedViewRow
	err := r.db.Conn.SelectContext(ctx, &rows, `
		SELECT id, workspace_id::text, user_id::text, name, resource_type, filter_json, shared, created_at, updated_at
		FROM saved_views
		WHERE workspace_id = $1 AND resource_type = $2 AND deleted_at IS NULL AND (user_id = $3 OR shared = TRUE)
		ORDER BY created_at DESC
	`, workspaceID, resourceType, userID)
	return savedViewEntities(rows), analyticsError(err)
}

func (r *Repository) CreateSavedView(ctx context.Context, view entity.SavedView) (entity.SavedView, error) {
	var row SavedViewRow
	err := r.db.Conn.GetContext(ctx, &row, `
		INSERT INTO saved_views (workspace_id, user_id, name, resource_type, filter_json, shared)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id::text, user_id::text, name, resource_type, filter_json, shared, created_at, updated_at
	`, view.WorkspaceID, view.UserID, view.Name, view.ResourceType, savedViewFilterJSON(view.FilterJSON), view.Shared)
	return row.toEntity(), analyticsError(err)
}

func (r *Repository) FindSavedView(ctx context.Context, id string) (entity.SavedView, error) {
	var row SavedViewRow
	err := r.db.Conn.GetContext(ctx, &row, `SELECT id, workspace_id::text, user_id::text, name, resource_type, filter_json, shared, created_at, updated_at FROM saved_views WHERE id = $1 AND deleted_at IS NULL`, id)
	return row.toEntity(), analyticsError(err)
}

func (r *Repository) UpdateSavedView(ctx context.Context, view entity.SavedView) (entity.SavedView, error) {
	var row SavedViewRow
	err := r.db.Conn.GetContext(ctx, &row, `
		UPDATE saved_views SET name = $2, filter_json = $3, shared = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, workspace_id::text, user_id::text, name, resource_type, filter_json, shared, created_at, updated_at
	`, view.ID, view.Name, savedViewFilterJSON(view.FilterJSON), view.Shared)
	return row.toEntity(), analyticsError(err)
}

func (r *Repository) DeleteSavedView(ctx context.Context, id string) error {
	var deletedID string
	return analyticsError(r.db.Conn.GetContext(ctx, &deletedID, `UPDATE saved_views SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING id`, id))
}

func (r *Repository) ListActivityLogs(ctx context.Context, filter analytics.ActivityLogFilter) ([]entity.ActivityLog, int, error) {
	var total int
	_ = r.db.Conn.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM activity_logs al
		WHERE ($1 = '' OR al.workspace_id::text = $1)
			AND ($2 = '' OR al.actor_id::text = $2)
			AND ($3 = '' OR al.action = $3)
			AND ($4 = '' OR al.resource_type = $4)
			AND ($5 = '' OR al.resource_id::text = $5)
			AND ($6::timestamptz IS NULL OR al.created_at >= $6)
			AND ($7::timestamptz IS NULL OR al.created_at <= $7)
	`, filter.WorkspaceID, filter.ActorID, filter.Action, filter.ResourceType, filter.ResourceID, filter.StartDate, filter.EndDate)
	var rows []ActivityLogRow
	err := r.db.Conn.SelectContext(ctx, &rows, `
		SELECT al.id, al.workspace_id::text, al.actor_id::text, u.full_name AS actor_name, al.action, al.resource_type, al.resource_id::text,
			al.old_value, al.new_value, al.ip_address, al.user_agent, al.created_at
		FROM activity_logs al
		LEFT JOIN users u ON u.id = al.actor_id
		WHERE ($1 = '' OR al.workspace_id::text = $1)
			AND ($2 = '' OR al.actor_id::text = $2)
			AND ($3 = '' OR al.action = $3)
			AND ($4 = '' OR al.resource_type = $4)
			AND ($5 = '' OR al.resource_id::text = $5)
			AND ($6::timestamptz IS NULL OR al.created_at >= $6)
			AND ($7::timestamptz IS NULL OR al.created_at <= $7)
		ORDER BY al.created_at DESC
		LIMIT $8 OFFSET $9
	`, filter.WorkspaceID, filter.ActorID, filter.Action, filter.ResourceType, filter.ResourceID, filter.StartDate, filter.EndDate, filter.Limit, filter.Offset)
	return activityLogEntities(rows), total, analyticsError(err)
}

func (r *Repository) summary(ctx context.Context, workspaceID, userID string, start, end time.Time) (entity.DashboardSummary, error) {
	var row SummaryRow
	err := r.db.Conn.GetContext(ctx, &row, `
		SELECT COALESCE(SUM(planned_minutes), 0)::int AS planned_minutes, COALESCE(SUM(actual_minutes), 0)::int AS actual_minutes,
			COUNT(id) FILTER (WHERE status = 'completed')::int AS completed_timeboxes, COUNT(id)::int AS total_timeboxes,
			CASE WHEN COUNT(id) = 0 THEN 0 ELSE ROUND((COUNT(id) FILTER (WHERE status = 'completed')::numeric / COUNT(id)::numeric) * 100, 2)::float END AS completion_rate
		FROM timeboxes
		WHERE workspace_id = $1 AND deleted_at IS NULL AND scheduled_start >= $2 AND scheduled_start < $3 AND ($4 = '' OR owner_id::text = $4)
	`, workspaceID, start, end, userID)
	return entity.DashboardSummary{PlannedMinutes: row.PlannedMinutes, ActualMinutes: row.ActualMinutes, CompletedTimeboxes: row.CompletedTimeboxes, TotalTimeboxes: row.TotalTimeboxes, CompletionRate: row.CompletionRate}, analyticsError(err)
}

func (r *Repository) timeline(ctx context.Context, workspaceID, userID string, start, end time.Time) []entity.TimelineItem {
	var rows []TimelineRow
	_ = r.db.Conn.SelectContext(ctx, &rows, `SELECT id::text, title, scheduled_start, scheduled_end, status FROM timeboxes WHERE workspace_id = $1 AND deleted_at IS NULL AND scheduled_start >= $2 AND scheduled_start < $3 AND ($4 = '' OR owner_id::text = $4) ORDER BY scheduled_start`, workspaceID, start, end, userID)
	items := make([]entity.TimelineItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.TimelineItem{ID: row.ID, Title: row.Title, ScheduledStart: row.ScheduledStart, ScheduledEnd: row.ScheduledEnd, Status: row.Status})
	}
	return items
}

func (r *Repository) categoryDistribution(ctx context.Context, workspaceID, userID string, start, end time.Time) []entity.CategoryDistribution {
	var rows []CategoryDistributionRow
	_ = r.db.Conn.SelectContext(ctx, &rows, `SELECT tb.category_id::text, COALESCE(c.name, 'Uncategorized') AS category_name, COALESCE(SUM(tb.planned_minutes),0)::int AS planned_minutes, COALESCE(SUM(tb.actual_minutes),0)::int AS actual_minutes FROM timeboxes tb LEFT JOIN categories c ON c.id = tb.category_id WHERE tb.workspace_id = $1 AND tb.deleted_at IS NULL AND tb.scheduled_start >= $2 AND tb.scheduled_start < $3 AND ($4 = '' OR tb.owner_id::text = $4) GROUP BY tb.category_id, c.name`, workspaceID, start, end, userID)
	items := make([]entity.CategoryDistribution, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.CategoryDistribution{CategoryID: row.CategoryID, CategoryName: row.CategoryName, PlannedMinutes: row.PlannedMinutes, ActualMinutes: row.ActualMinutes})
	}
	return items
}

func (r *Repository) teamToday(ctx context.Context, workspaceID string, date time.Time) []entity.TeamTodayItem {
	var rows []TeamTodayRow
	_ = r.db.Conn.SelectContext(ctx, &rows, `SELECT u.id::text AS user_id, u.full_name, u.avatar_url, NULL::text AS current_timebox_id, NULL::text AS current_timebox_title, NULL::text AS current_status, COALESCE(SUM(tb.planned_minutes),0)::int AS planned_minutes, COALESCE(SUM(tb.actual_minutes),0)::int AS actual_minutes, 0::float AS completion_rate FROM workspace_members wm JOIN users u ON u.id = wm.user_id LEFT JOIN timeboxes tb ON tb.owner_id = wm.user_id AND tb.workspace_id = wm.workspace_id AND tb.deleted_at IS NULL AND tb.scheduled_start >= $2 AND tb.scheduled_start < $3 WHERE wm.workspace_id = $1 AND wm.status = 'active' GROUP BY u.id, u.full_name, u.avatar_url`, workspaceID, date, date.AddDate(0, 0, 1))
	items := make([]entity.TeamTodayItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.TeamTodayItem{User: entity.UserSummary{ID: row.UserID, FullName: row.FullName, AvatarURL: row.AvatarURL}, PlannedMinutes: row.PlannedMinutes, ActualMinutes: row.ActualMinutes, CompletionRate: row.CompletionRate})
	}
	return items
}

func (r *Repository) trend(ctx context.Context, workspaceID string, start, end time.Time, interval string) []entity.TrendItem {
	trunc := "day"
	if interval == "week" || interval == "month" {
		trunc = interval
	}
	var rows []TrendRow
	query := `SELECT to_char(date_trunc('` + trunc + `', scheduled_start), 'YYYY-MM-DD') AS date, COALESCE(SUM(planned_minutes),0)::int AS planned_minutes, COALESCE(SUM(actual_minutes),0)::int AS actual_minutes, COUNT(id) FILTER (WHERE status = 'completed')::int AS completed_timeboxes, COUNT(id)::int AS total_timeboxes, CASE WHEN COUNT(id) = 0 THEN 0 ELSE ROUND((COUNT(id) FILTER (WHERE status = 'completed')::numeric / COUNT(id)::numeric) * 100, 2)::float END AS completion_rate FROM timeboxes WHERE workspace_id = $1 AND deleted_at IS NULL AND scheduled_start >= $2 AND scheduled_start < $3 GROUP BY 1 ORDER BY 1`
	_ = r.db.Conn.SelectContext(ctx, &rows, query, workspaceID, start, end)
	items := make([]entity.TrendItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.TrendItem{Date: row.Date, PlannedMinutes: row.PlannedMinutes, ActualMinutes: row.ActualMinutes, CompletedTimeboxes: row.CompletedTimeboxes, TotalTimeboxes: row.TotalTimeboxes, CompletionRate: row.CompletionRate})
	}
	return items
}

func (r *Repository) reportItems(ctx context.Context, filter analytics.ReportFilter) []entity.ReportItem {
	keyExpr, labelExpr := reportGroupExpr(filter.GroupBy)
	query := `SELECT ` + keyExpr + ` AS key, ` + labelExpr + ` AS label, COALESCE(SUM(tb.planned_minutes),0)::int AS planned_minutes, COALESCE(SUM(tb.actual_minutes),0)::int AS actual_minutes, 0::float AS percentage FROM timeboxes tb LEFT JOIN categories c ON c.id = tb.category_id LEFT JOIN tasks t ON t.id = tb.task_id LEFT JOIN goals g ON g.id = t.goal_id LEFT JOIN users u ON u.id = tb.owner_id WHERE tb.workspace_id = $1 AND tb.deleted_at IS NULL AND tb.scheduled_start >= $2 AND tb.scheduled_start < $3 AND ($4 = '' OR tb.owner_id::text = $4) AND ($5 = '' OR tb.category_id::text = $5) AND ($6 = '' OR g.id::text = $6) GROUP BY 1, 2 ORDER BY actual_minutes DESC`
	var rows []ReportItemRow
	_ = r.db.Conn.SelectContext(ctx, &rows, query, filter.WorkspaceID, filter.StartDate, filter.EndDate.AddDate(0, 0, 1), filter.UserID, filter.CategoryID, filter.GoalID)
	items := make([]entity.ReportItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.ReportItem{Key: row.Key, Label: row.Label, PlannedMinutes: row.PlannedMinutes, ActualMinutes: row.ActualMinutes, VarianceMinutes: row.ActualMinutes - row.PlannedMinutes, Percentage: row.Percentage})
	}
	return items
}

func reportGroupExpr(groupBy string) (string, string) {
	switch groupBy {
	case "goal":
		return "COALESCE(g.id::text, 'none')", "COALESCE(g.title, 'No Goal')"
	case "user":
		return "u.id::text", "u.full_name"
	case "day":
		return "to_char(tb.scheduled_start, 'YYYY-MM-DD')", "to_char(tb.scheduled_start, 'YYYY-MM-DD')"
	default:
		return "COALESCE(c.id::text, 'none')", "COALESCE(c.name, 'Uncategorized')"
	}
}

func reportSummary(items []entity.ReportItem) entity.ReportSummary {
	var summary entity.ReportSummary
	for _, item := range items {
		summary.PlannedMinutes += item.PlannedMinutes
		summary.ActualMinutes += item.ActualMinutes
	}
	summary.VarianceMinutes = summary.ActualMinutes - summary.PlannedMinutes
	return summary
}

func reportRange(filter analytics.ReportFilter) entity.ReportRange {
	return entity.ReportRange{StartDate: filter.StartDate.Format("2006-01-02"), EndDate: filter.EndDate.Format("2006-01-02")}
}

func (r *Repository) searchRows(ctx context.Context, query, workspaceID, search string, limit int) []entity.SearchResult {
	var rows []SearchRow
	_ = r.db.Conn.SelectContext(ctx, &rows, query, workspaceID, search, limit)
	results := make([]entity.SearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, entity.SearchResult{ID: row.ID, Title: row.Title, Snippet: row.Snippet})
	}
	return results
}

func savedViewEntities(rows []SavedViewRow) []entity.SavedView {
	items := make([]entity.SavedView, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toEntity())
	}
	return items
}

func activityLogEntities(rows []ActivityLogRow) []entity.ActivityLog {
	items := make([]entity.ActivityLog, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toEntity())
	}
	return items
}

func (r SavedViewRow) toEntity() entity.SavedView {
	filter := map[string]any{}
	_ = json.Unmarshal(r.FilterJSON, &filter)
	return entity.SavedView{ID: r.ID, WorkspaceID: r.WorkspaceID, UserID: r.UserID, Name: r.Name, ResourceType: r.ResourceType, FilterJSON: filter, Shared: r.Shared, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func savedViewFilterJSON(filter map[string]any) []byte {
	if filter == nil {
		filter = map[string]any{}
	}
	value, _ := json.Marshal(filter)
	return value
}

func (r ActivityLogRow) toEntity() entity.ActivityLog {
	oldValue, newValue := map[string]any{}, map[string]any{}
	_ = json.Unmarshal(r.OldValue, &oldValue)
	_ = json.Unmarshal(r.NewValue, &newValue)
	var actor *entity.UserSummary
	if r.ActorID != nil && r.ActorName != nil {
		actor = &entity.UserSummary{ID: *r.ActorID, FullName: *r.ActorName}
	}
	return entity.ActivityLog{ID: r.ID, WorkspaceID: r.WorkspaceID, Actor: actor, Action: r.Action, ResourceType: r.ResourceType, ResourceID: r.ResourceID, OldValue: oldValue, NewValue: newValue, IPAddress: r.IPAddress, UserAgent: r.UserAgent, CreatedAt: r.CreatedAt}
}

func periodStart(period string) time.Time {
	now := time.Now()
	switch period {
	case "month":
		return now.AddDate(0, -1, 0)
	case "day":
		return now.AddDate(0, 0, -1)
	default:
		return now.AddDate(0, 0, -7)
	}
}

func analyticsError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return analytics.ErrNotFound
	}
	return err
}
