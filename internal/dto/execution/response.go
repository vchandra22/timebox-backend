package executiondto

import "time"

type TimeboxResponse struct {
	ID             string               `json:"id"`
	WorkspaceID    string               `json:"workspace_id,omitempty"`
	TaskID         *string              `json:"task_id"`
	OwnerID        string               `json:"owner_id,omitempty"`
	CategoryID     *string              `json:"category_id"`
	Title          string               `json:"title"`
	Description    string               `json:"description,omitempty"`
	ScheduledStart time.Time            `json:"scheduled_start"`
	ScheduledEnd   time.Time            `json:"scheduled_end"`
	PlannedMinutes int                  `json:"planned_minutes"`
	ActualMinutes  int                  `json:"actual_minutes,omitempty"`
	Status         string               `json:"status"`
	IsBuffer       bool                 `json:"is_buffer"`
	Owner          *UserSummaryResponse `json:"owner,omitempty"`
	Category       *CategorySummaryResp `json:"category,omitempty"`
	Task           *TaskSummaryResponse `json:"task,omitempty"`
	Warnings       []WarningResponse    `json:"warnings,omitempty"`
	CreatedAt      *time.Time           `json:"created_at,omitempty"`
	UpdatedAt      *time.Time           `json:"updated_at,omitempty"`
}

type UserSummaryResponse struct {
	ID        string  `json:"id"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
}

type CategorySummaryResp struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type TaskSummaryResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type WarningResponse struct {
	Code              string  `json:"code"`
	Message           string  `json:"message"`
	ConflictTimeboxID *string `json:"conflict_timebox_id,omitempty"`
}

type TimerStateResponse struct {
	TimeboxID        string           `json:"timebox_id"`
	Status           string           `json:"status"`
	StartedAt        time.Time        `json:"started_at,omitempty"`
	PausedAt         *time.Time       `json:"paused_at,omitempty"`
	PlannedMinutes   int              `json:"planned_minutes,omitempty"`
	ElapsedSeconds   int              `json:"elapsed_seconds"`
	RemainingSeconds int              `json:"remaining_seconds"`
	ServerTime       time.Time        `json:"server_time,omitempty"`
	Timebox          *TimeboxResponse `json:"timebox,omitempty"`
	ResumedAt        *time.Time       `json:"resumed_at,omitempty"`
}

type TimerCompletionResponse struct {
	TimeboxID       string    `json:"timebox_id"`
	Status          string    `json:"status"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	SkippedAt       time.Time `json:"skipped_at,omitempty"`
	PlannedMinutes  int       `json:"planned_minutes,omitempty"`
	ActualMinutes   int       `json:"actual_minutes,omitempty"`
	VarianceMinutes int       `json:"variance_minutes,omitempty"`
	StreakUpdated   bool      `json:"streak_updated,omitempty"`
}

type TimeLogResponse struct {
	ID              string     `json:"id"`
	TimeboxID       string     `json:"timebox_id"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationSeconds int        `json:"duration_seconds"`
	Source          string     `json:"source"`
	Note            *string    `json:"note"`
	CreatedBy       string     `json:"created_by,omitempty"`
}
