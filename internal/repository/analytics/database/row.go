package database

import "time"

type StreakRow struct {
	CompletedDays int        `db:"completed_days"`
	LastDate      *time.Time `db:"last_date"`
}

type LeaderboardRow struct {
	UserID    string  `db:"user_id"`
	FullName  string  `db:"full_name"`
	AvatarURL *string `db:"avatar_url"`
	Value     int     `db:"value"`
}

type TimelineRow struct {
	ID             string    `db:"id"`
	Title          string    `db:"title"`
	ScheduledStart time.Time `db:"scheduled_start"`
	ScheduledEnd   time.Time `db:"scheduled_end"`
	Status         string    `db:"status"`
}

type SummaryRow struct {
	PlannedMinutes     int     `db:"planned_minutes"`
	ActualMinutes      int     `db:"actual_minutes"`
	CompletedTimeboxes int     `db:"completed_timeboxes"`
	TotalTimeboxes     int     `db:"total_timeboxes"`
	CompletionRate     float64 `db:"completion_rate"`
}

type CategoryDistributionRow struct {
	CategoryID     *string `db:"category_id"`
	CategoryName   string  `db:"category_name"`
	PlannedMinutes int     `db:"planned_minutes"`
	ActualMinutes  int     `db:"actual_minutes"`
}

type TeamTodayRow struct {
	UserID              string  `db:"user_id"`
	FullName            string  `db:"full_name"`
	AvatarURL           *string `db:"avatar_url"`
	CurrentTimeboxID    *string `db:"current_timebox_id"`
	CurrentTimeboxTitle *string `db:"current_timebox_title"`
	CurrentStatus       *string `db:"current_status"`
	PlannedMinutes      int     `db:"planned_minutes"`
	ActualMinutes       int     `db:"actual_minutes"`
	CompletionRate      float64 `db:"completion_rate"`
}

type ReportItemRow struct {
	Key            string  `db:"key"`
	Label          string  `db:"label"`
	PlannedMinutes int     `db:"planned_minutes"`
	ActualMinutes  int     `db:"actual_minutes"`
	Percentage     float64 `db:"percentage"`
}

type TrendRow struct {
	Date               string  `db:"date"`
	PlannedMinutes     int     `db:"planned_minutes"`
	ActualMinutes      int     `db:"actual_minutes"`
	CompletedTimeboxes int     `db:"completed_timeboxes"`
	TotalTimeboxes     int     `db:"total_timeboxes"`
	CompletionRate     float64 `db:"completion_rate"`
}

type WorkloadRow struct {
	UserID             string `db:"user_id"`
	FullName           string `db:"full_name"`
	PlannedMinutes     int    `db:"planned_minutes"`
	ActualMinutes      int    `db:"actual_minutes"`
	CompletedTimeboxes int    `db:"completed_timeboxes"`
	OverrunCount       int    `db:"overrun_count"`
}

type SearchRow struct {
	ID      string `db:"id"`
	Title   string `db:"title"`
	Snippet string `db:"snippet"`
}

type SavedViewRow struct {
	ID           string    `db:"id"`
	WorkspaceID  string    `db:"workspace_id"`
	UserID       string    `db:"user_id"`
	Name         string    `db:"name"`
	ResourceType string    `db:"resource_type"`
	FilterJSON   []byte    `db:"filter_json"`
	Shared       bool      `db:"shared"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type ActivityLogRow struct {
	ID           string    `db:"id"`
	WorkspaceID  *string   `db:"workspace_id"`
	ActorID      *string   `db:"actor_id"`
	ActorName    *string   `db:"actor_name"`
	Action       string    `db:"action"`
	ResourceType string    `db:"resource_type"`
	ResourceID   *string   `db:"resource_id"`
	OldValue     []byte    `db:"old_value"`
	NewValue     []byte    `db:"new_value"`
	IPAddress    *string   `db:"ip_address"`
	UserAgent    *string   `db:"user_agent"`
	CreatedAt    time.Time `db:"created_at"`
}
