package planningdto

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required,max=120"`
	Color string `json:"color" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

type CreateTagRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=80"`
}

type UpdateTagRequest struct {
	Name *string `json:"name"`
}

type CreateGoalRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description"`
	TargetDate  string `json:"target_date"`
	IsPinned    bool   `json:"is_pinned"`
}

type UpdateGoalRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	TargetDate  *string `json:"target_date"`
	Status      *string `json:"status"`
	IsPinned    *bool   `json:"is_pinned"`
}

type CreateTaskRequest struct {
	GoalID           *string                  `json:"goal_id"`
	AssigneeID       *string                  `json:"assignee_id"`
	CategoryID       *string                  `json:"category_id"`
	Title            string                   `json:"title" binding:"required,max=200"`
	Description      string                   `json:"description"`
	Priority         string                   `json:"priority"`
	EstimatedMinutes *int                     `json:"estimated_minutes"`
	TagIDs           []string                 `json:"tag_ids"`
	Checklist        []TaskChecklistCreateReq `json:"checklist"`
}

type TaskChecklistCreateReq struct {
	Title string `json:"title" binding:"required"`
}

type UpdateTaskRequest struct {
	GoalID           *string   `json:"goal_id"`
	AssigneeID       *string   `json:"assignee_id"`
	CategoryID       *string   `json:"category_id"`
	Title            *string   `json:"title"`
	Description      *string   `json:"description"`
	Status           *string   `json:"status"`
	Priority         *string   `json:"priority"`
	EstimatedMinutes *int      `json:"estimated_minutes"`
	TagIDs           *[]string `json:"tag_ids"`
}

type MoveTaskRequest struct {
	WorkspaceID  string  `json:"workspace_id" binding:"required"`
	ToStatus     string  `json:"to_status" binding:"required"`
	Position     int     `json:"position" binding:"required"`
	BeforeTaskID *string `json:"before_task_id"`
	AfterTaskID  *string `json:"after_task_id"`
}
