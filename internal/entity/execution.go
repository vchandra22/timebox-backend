package entity

import "time"

type Timebox struct {
	ID             string
	WorkspaceID    string
	TaskID         *string
	OwnerID        string
	CategoryID     *string
	Title          string
	Description    string
	ScheduledStart time.Time
	ScheduledEnd   time.Time
	PlannedMinutes int
	ActualMinutes  int
	Status         string
	IsBuffer       bool
	Owner          *UserSummary
	Category       *CategorySummary
	Task           *TaskSummary
	Warnings       []Warning
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CategorySummary struct {
	ID    string
	Name  string
	Color string
}

type TaskSummary struct {
	ID    string
	Title string
}

type Warning struct {
	Code              string
	Message           string
	ConflictTimeboxID *string
}

type TimerState struct {
	TimeboxID        string
	UserID           string
	Status           string
	StartedAt        time.Time
	PausedAt         *time.Time
	PlannedMinutes   int
	ElapsedSeconds   int
	RemainingSeconds int
	ServerTime       time.Time
	Timebox          *Timebox
}

type TimeLog struct {
	ID              string
	TimeboxID       string
	WorkspaceID     string
	StartedAt       time.Time
	EndedAt         *time.Time
	DurationSeconds int
	Source          string
	Note            *string
	CreatedBy       string
}

type TimerCompletion struct {
	TimeboxID       string
	Status          string
	CompletedAt     time.Time
	SkippedAt       time.Time
	PlannedMinutes  int
	ActualMinutes   int
	VarianceMinutes int
	StreakUpdated   bool
}
