package database

import "time"

type TimeboxRow struct {
	ID             string    `db:"id"`
	WorkspaceID    string    `db:"workspace_id"`
	TaskID         *string   `db:"task_id"`
	OwnerID        string    `db:"owner_id"`
	CategoryID     *string   `db:"category_id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	ScheduledStart time.Time `db:"scheduled_start"`
	ScheduledEnd   time.Time `db:"scheduled_end"`
	PlannedMinutes int       `db:"planned_minutes"`
	ActualMinutes  int       `db:"actual_minutes"`
	Status         string    `db:"status"`
	IsBuffer       bool      `db:"is_buffer"`
	OwnerName      *string   `db:"owner_name"`
	OwnerAvatar    *string   `db:"owner_avatar"`
	CategoryName   *string   `db:"category_name"`
	CategoryColor  *string   `db:"category_color"`
	TaskTitle      *string   `db:"task_title"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type TimeLogRow struct {
	ID              string     `db:"id"`
	TimeboxID       string     `db:"timebox_id"`
	WorkspaceID     string     `db:"workspace_id"`
	StartedAt       time.Time  `db:"started_at"`
	EndedAt         *time.Time `db:"ended_at"`
	DurationSeconds int        `db:"duration_seconds"`
	Source          string     `db:"source"`
	Note            *string    `db:"note"`
	CreatedBy       string     `db:"created_by"`
}
