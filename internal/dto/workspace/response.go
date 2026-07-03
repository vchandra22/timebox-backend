package workspacedto

import "time"

type WorkspaceSettingsResponse struct {
	LeaderboardEnabled            bool `json:"leaderboard_enabled"`
	DefaultPlannerIntervalMinutes int  `json:"default_planner_interval_minutes"`
}

type WorkspaceOwnerResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type WorkspaceResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Slug        string                     `json:"slug"`
	LogoURL     *string                    `json:"logo_url"`
	Timezone    string                     `json:"timezone"`
	Role        string                     `json:"role,omitempty"`
	OwnerID     string                     `json:"owner_id,omitempty"`
	Owner       *WorkspaceOwnerResponse    `json:"owner,omitempty"`
	Settings    *WorkspaceSettingsResponse `json:"settings,omitempty"`
	MemberCount int                        `json:"member_count,omitempty"`
	Status      string                     `json:"status,omitempty"`
	CreatedAt   *time.Time                 `json:"created_at,omitempty"`
	UpdatedAt   *time.Time                 `json:"updated_at,omitempty"`
}

type WorkspaceInvitationResponse struct {
	InviteID  string    `json:"invite_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TeamSummaryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WorkspaceMemberResponse struct {
	UserID    string                `json:"user_id"`
	FullName  string                `json:"full_name"`
	Email     string                `json:"email"`
	AvatarURL *string               `json:"avatar_url"`
	Role      string                `json:"role"`
	Status    string                `json:"status"`
	Teams     []TeamSummaryResponse `json:"teams"`
	JoinedAt  time.Time             `json:"joined_at"`
}

type WorkspaceMemberUpdateResponse struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	TeamIDs   []string  `json:"team_ids"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TeamResponse struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	MemberCount int        `json:"member_count,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
