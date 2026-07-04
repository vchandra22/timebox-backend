package analyticsdto

import "time"

type UserSummaryResponse struct {
	ID        string  `json:"id"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type StreakResponse struct {
	WorkspaceID       string          `json:"workspace_id"`
	CurrentStreak     int             `json:"current_streak"`
	LongestStreak     int             `json:"longest_streak"`
	LastCompletedDate *string         `json:"last_completed_date"`
	Badges            []BadgeResponse `json:"badges"`
}

type BadgeResponse struct {
	Code     string    `json:"code"`
	Title    string    `json:"title"`
	EarnedAt time.Time `json:"earned_at"`
}

type LeaderboardResponse struct {
	Metric string                    `json:"metric"`
	Period string                    `json:"period"`
	Items  []LeaderboardItemResponse `json:"items"`
}

type LeaderboardItemResponse struct {
	Rank  int                 `json:"rank"`
	User  UserSummaryResponse `json:"user"`
	Value int                 `json:"value"`
}

type PersonalDashboardResponse struct {
	Date                 string                         `json:"date"`
	Timezone             string                         `json:"timezone"`
	ActiveTimer          any                            `json:"active_timer"`
	TodayTimeline        []TimelineItemResponse         `json:"today_timeline"`
	Summary              DashboardSummaryResponse       `json:"summary"`
	CategoryDistribution []CategoryDistributionResponse `json:"category_distribution"`
	OverrunOrMissed      []TimelineItemResponse         `json:"overrun_or_missed"`
}

type WorkspaceDashboardResponse struct {
	Date                 string                         `json:"date"`
	Summary              DashboardSummaryResponse       `json:"summary"`
	TeamToday            []TeamTodayResponse            `json:"team_today"`
	CategoryDistribution []CategoryDistributionResponse `json:"category_distribution"`
	Trend                []TrendItemResponse            `json:"trend"`
}

type DashboardSummaryResponse struct {
	PlannedMinutes        int     `json:"planned_minutes,omitempty"`
	ActualMinutes         int     `json:"actual_minutes,omitempty"`
	CompletedTimeboxes    int     `json:"completed_timeboxes,omitempty"`
	TotalTimeboxes        int     `json:"total_timeboxes,omitempty"`
	CompletionRate        float64 `json:"completion_rate,omitempty"`
	CurrentStreak         int     `json:"current_streak,omitempty"`
	TeamFocusMinutesToday int     `json:"team_focus_minutes_today,omitempty"`
	TeamFocusMinutesWeek  int     `json:"team_focus_minutes_week,omitempty"`
	ActiveMembersNow      int     `json:"active_members_now,omitempty"`
	OverrunRiskMembers    int     `json:"overrun_risk_members,omitempty"`
}

type TimelineItemResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	ScheduledStart time.Time `json:"scheduled_start"`
	ScheduledEnd   time.Time `json:"scheduled_end"`
	Status         string    `json:"status"`
}

type CategoryDistributionResponse struct {
	CategoryID     *string `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	PlannedMinutes int     `json:"planned_minutes"`
	ActualMinutes  int     `json:"actual_minutes"`
}

type TeamTodayResponse struct {
	User           UserSummaryResponse   `json:"user"`
	CurrentTimebox *TimelineItemResponse `json:"current_timebox"`
	PlannedMinutes int                   `json:"planned_minutes"`
	ActualMinutes  int                   `json:"actual_minutes"`
	CompletionRate float64               `json:"completion_rate"`
}

type ReportRangeResponse struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type ReportSummaryResponse struct {
	PlannedMinutes  int `json:"planned_minutes"`
	ActualMinutes   int `json:"actual_minutes"`
	VarianceMinutes int `json:"variance_minutes"`
}

type TimeAllocationReportResponse struct {
	Range   ReportRangeResponse   `json:"range"`
	GroupBy string                `json:"group_by"`
	Summary ReportSummaryResponse `json:"summary"`
	Items   []ReportItemResponse  `json:"items"`
}

type ReportItemResponse struct {
	Key             string  `json:"key"`
	Label           string  `json:"label"`
	PlannedMinutes  int     `json:"planned_minutes"`
	ActualMinutes   int     `json:"actual_minutes"`
	VarianceMinutes int     `json:"variance_minutes"`
	Percentage      float64 `json:"percentage"`
}

type ProductivityTrendResponse struct {
	Interval string              `json:"interval"`
	Items    []TrendItemResponse `json:"items"`
}

type TrendItemResponse struct {
	Date               string  `json:"date"`
	PlannedMinutes     int     `json:"planned_minutes"`
	ActualMinutes      int     `json:"actual_minutes"`
	CompletedTimeboxes int     `json:"completed_timeboxes"`
	TotalTimeboxes     int     `json:"total_timeboxes"`
	CompletionRate     float64 `json:"completion_rate"`
}

type TeamWorkloadResponse struct {
	Range ReportRangeResponse        `json:"range"`
	Items []TeamWorkloadItemResponse `json:"items"`
}

type TeamWorkloadItemResponse struct {
	User               UserSummaryResponse `json:"user"`
	PlannedMinutes     int                 `json:"planned_minutes"`
	ActualMinutes      int                 `json:"actual_minutes"`
	CompletedTimeboxes int                 `json:"completed_timeboxes"`
	OverrunCount       int                 `json:"overrun_count"`
	RiskLevel          string              `json:"risk_level"`
}

type SearchResponse struct {
	Query   string                          `json:"query"`
	Results map[string][]SearchItemResponse `json:"results"`
}

type SearchItemResponse struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type SavedViewResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	ResourceType string         `json:"resource_type"`
	FilterJSON   map[string]any `json:"filter_json,omitempty"`
	Shared       bool           `json:"shared"`
}

type ActivityLogResponse struct {
	ID           string               `json:"id"`
	WorkspaceID  *string              `json:"workspace_id"`
	Actor        *UserSummaryResponse `json:"actor"`
	Action       string               `json:"action"`
	ResourceType string               `json:"resource_type"`
	ResourceID   *string              `json:"resource_id"`
	OldValue     map[string]any       `json:"old_value"`
	NewValue     map[string]any       `json:"new_value"`
	IPAddress    *string              `json:"ip_address"`
	UserAgent    *string              `json:"user_agent"`
	CreatedAt    time.Time            `json:"created_at"`
}
