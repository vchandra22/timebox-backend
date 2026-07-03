package database

import "time"

type CategoryRow struct {
	ID          string    `db:"id"`
	WorkspaceID string    `db:"workspace_id"`
	Name        string    `db:"name"`
	Color       string    `db:"color"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type TagRow struct {
	ID          string    `db:"id"`
	WorkspaceID string    `db:"workspace_id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type GoalRow struct {
	ID              string     `db:"id"`
	WorkspaceID     string     `db:"workspace_id"`
	CreatedBy       string     `db:"created_by"`
	CreatedByName   string     `db:"created_by_name"`
	Title           string     `db:"title"`
	Description     string     `db:"description"`
	TargetDate      *time.Time `db:"target_date"`
	Status          string     `db:"status"`
	IsPinned        bool       `db:"is_pinned"`
	PlannedMinutes  int        `db:"planned_minutes"`
	ActualMinutes   int        `db:"actual_minutes"`
	CompletedBlocks int        `db:"completed_blocks"`
	ProgressPercent float64    `db:"progress_percent"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type TaskRow struct {
	ID               string    `db:"id"`
	WorkspaceID      string    `db:"workspace_id"`
	GoalID           *string   `db:"goal_id"`
	AssigneeID       *string   `db:"assignee_id"`
	CategoryID       *string   `db:"category_id"`
	CreatedBy        string    `db:"created_by"`
	Title            string    `db:"title"`
	Description      string    `db:"description"`
	Status           string    `db:"status"`
	Priority         string    `db:"priority"`
	EstimatedMinutes *int      `db:"estimated_minutes"`
	Position         int       `db:"position"`
	AssigneeName     *string   `db:"assignee_name"`
	AssigneeAvatar   *string   `db:"assignee_avatar"`
	GoalTitle        *string   `db:"goal_title"`
	TimeboxesCount   int       `db:"timeboxes_count"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type ChecklistRow struct {
	ID       string `db:"id"`
	Title    string `db:"title"`
	IsDone   bool   `db:"is_done"`
	Position int    `db:"position"`
}

type TaskMoveRow struct {
	ID          string    `db:"id"`
	WorkspaceID string    `db:"workspace_id"`
	FromStatus  string    `db:"from_status"`
	ToStatus    string    `db:"to_status"`
	Position    int       `db:"position"`
	UpdatedAt   time.Time `db:"updated_at"`
}
