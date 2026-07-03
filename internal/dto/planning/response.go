package planningdto

import "time"

type CategoryResponse struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id,omitempty"`
	Name        string     `json:"name"`
	Color       string     `json:"color"`
	IsDefault   bool       `json:"is_default,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type CategoryDeleteResponse struct {
	MovedTimeboxesToCategoryID *string `json:"moved_timeboxes_to_category_id"`
}

type TagResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	Name        string `json:"name"`
}

type GoalResponse struct {
	ID              string            `json:"id"`
	WorkspaceID     string            `json:"workspace_id,omitempty"`
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	TargetDate      *string           `json:"target_date"`
	Status          string            `json:"status,omitempty"`
	IsPinned        bool              `json:"is_pinned"`
	ProgressPercent float64           `json:"progress_percent,omitempty"`
	Progress        *GoalProgressResp `json:"progress,omitempty"`
	CreatedBy       *CreatedByResp    `json:"created_by,omitempty"`
	CreatedAt       *time.Time        `json:"created_at,omitempty"`
	UpdatedAt       *time.Time        `json:"updated_at,omitempty"`
}

type GoalProgressResp struct {
	PlannedMinutes     int     `json:"planned_minutes"`
	ActualMinutes      int     `json:"actual_minutes"`
	CompletedTimeboxes int     `json:"completed_timeboxes"`
	ProgressPercent    float64 `json:"progress_percent"`
}

type CreatedByResp struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type TaskResponse struct {
	ID               string                  `json:"id"`
	WorkspaceID      string                  `json:"workspace_id,omitempty"`
	GoalID           *string                 `json:"goal_id"`
	AssigneeID       *string                 `json:"assignee_id,omitempty"`
	CategoryID       *string                 `json:"category_id"`
	Title            string                  `json:"title"`
	Description      string                  `json:"description,omitempty"`
	Status           string                  `json:"status"`
	Priority         string                  `json:"priority"`
	EstimatedMinutes *int                    `json:"estimated_minutes"`
	Position         int                     `json:"position"`
	Assignee         *UserSummaryResponse    `json:"assignee,omitempty"`
	Goal             *GoalSummaryResponse    `json:"goal,omitempty"`
	Tags             []TagResponse           `json:"tags,omitempty"`
	Checklist        []TaskChecklistResponse `json:"checklist,omitempty"`
	TimeboxesCount   int                     `json:"timeboxes_count,omitempty"`
	CreatedAt        *time.Time              `json:"created_at,omitempty"`
	UpdatedAt        *time.Time              `json:"updated_at,omitempty"`
}

type UserSummaryResponse struct {
	ID        string  `json:"id"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
}

type GoalSummaryResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type TaskChecklistResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	IsDone   bool   `json:"is_done"`
	Position int    `json:"position"`
}

type TaskMoveResponse struct {
	ID         string    `json:"id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Position   int       `json:"position"`
	UpdatedAt  time.Time `json:"updated_at"`
}
