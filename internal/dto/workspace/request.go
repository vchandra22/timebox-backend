package workspacedto

type CreateWorkspaceRequest struct {
	Name     string  `json:"name" binding:"required,max=120"`
	Slug     string  `json:"slug" binding:"required,max=140"`
	Timezone string  `json:"timezone" binding:"required"`
	LogoURL  *string `json:"logo_url"`
}

type WorkspaceSettingsRequest struct {
	LeaderboardEnabled            *bool `json:"leaderboard_enabled"`
	DefaultPlannerIntervalMinutes *int  `json:"default_planner_interval_minutes"`
}

type UpdateWorkspaceRequest struct {
	Name     *string                   `json:"name"`
	Slug     *string                   `json:"slug"`
	Timezone *string                   `json:"timezone"`
	LogoURL  *string                   `json:"logo_url"`
	Settings *WorkspaceSettingsRequest `json:"settings"`
}

type InviteWorkspaceMemberRequest struct {
	Email   string   `json:"email" binding:"required,email"`
	Role    string   `json:"role" binding:"required"`
	TeamIDs []string `json:"team_ids"`
}

type UpdateWorkspaceMemberRequest struct {
	Role    *string   `json:"role"`
	Status  *string   `json:"status"`
	TeamIDs *[]string `json:"team_ids"`
}

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,max=120"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
