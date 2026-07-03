package entity

import "time"

type Streak struct {
	WorkspaceID       string
	CurrentStreak     int
	LongestStreak     int
	LastCompletedDate *time.Time
	Badges            []Badge
}

type Badge struct {
	Code     string
	Title    string
	EarnedAt time.Time
}

type Leaderboard struct {
	Metric string
	Period string
	Items  []LeaderboardItem
}

type LeaderboardItem struct {
	Rank  int
	User  UserSummary
	Value int
}

type DashboardSummary struct {
	PlannedMinutes        int
	ActualMinutes         int
	CompletedTimeboxes    int
	TotalTimeboxes        int
	CompletionRate        float64
	CurrentStreak         int
	TeamFocusMinutesToday int
	TeamFocusMinutesWeek  int
	ActiveMembersNow      int
	OverrunRiskMembers    int
}

type TimelineItem struct {
	ID             string
	Title          string
	ScheduledStart time.Time
	ScheduledEnd   time.Time
	Status         string
}

type CategoryDistribution struct {
	CategoryID     *string
	CategoryName   string
	PlannedMinutes int
	ActualMinutes  int
}

type PersonalDashboard struct {
	Date                 string
	Timezone             string
	TodayTimeline        []TimelineItem
	Summary              DashboardSummary
	CategoryDistribution []CategoryDistribution
	OverrunOrMissed      []TimelineItem
}

type WorkspaceDashboard struct {
	Date                 string
	Summary              DashboardSummary
	TeamToday            []TeamTodayItem
	CategoryDistribution []CategoryDistribution
	Trend                []TrendItem
}

type TeamTodayItem struct {
	User           UserSummary
	CurrentTimebox *TimelineItem
	PlannedMinutes int
	ActualMinutes  int
	CompletionRate float64
}

type ReportRange struct {
	StartDate string
	EndDate   string
}

type ReportSummary struct {
	PlannedMinutes  int
	ActualMinutes   int
	VarianceMinutes int
}

type TimeAllocationReport struct {
	Range   ReportRange
	GroupBy string
	Summary ReportSummary
	Items   []ReportItem
}

type ReportItem struct {
	Key             string
	Label           string
	PlannedMinutes  int
	ActualMinutes   int
	VarianceMinutes int
	Percentage      float64
}

type ProductivityTrendReport struct {
	Interval string
	Items    []TrendItem
}

type TrendItem struct {
	Date               string
	PlannedMinutes     int
	ActualMinutes      int
	CompletedTimeboxes int
	TotalTimeboxes     int
	CompletionRate     float64
}

type TeamWorkloadReport struct {
	Range ReportRange
	Items []TeamWorkloadItem
}

type TeamWorkloadItem struct {
	User               UserSummary
	PlannedMinutes     int
	ActualMinutes      int
	CompletedTimeboxes int
	OverrunCount       int
	RiskLevel          string
}

type SearchResult struct {
	ID      string
	Title   string
	Snippet string
}

type SearchResults struct {
	Query       string
	Tasks       []SearchResult
	Timeboxes   []SearchResult
	Goals       []SearchResult
	Comments    []SearchResult
	Attachments []SearchResult
}

type SavedView struct {
	ID           string
	WorkspaceID  string
	UserID       string
	Name         string
	ResourceType string
	FilterJSON   map[string]any
	Shared       bool
	SharedSet    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ActivityLog struct {
	ID           string
	WorkspaceID  *string
	Actor        *UserSummary
	Action       string
	ResourceType string
	ResourceID   *string
	OldValue     map[string]any
	NewValue     map[string]any
	IPAddress    *string
	UserAgent    *string
	CreatedAt    time.Time
}
