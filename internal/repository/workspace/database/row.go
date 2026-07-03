package database

import "time"

type WorkspaceRow struct {
	ID                            string    `db:"id"`
	Name                          string    `db:"name"`
	Slug                          string    `db:"slug"`
	Timezone                      string    `db:"timezone"`
	LogoURL                       *string   `db:"logo_url"`
	OwnerID                       string    `db:"owner_id"`
	OwnerName                     string    `db:"owner_name"`
	Role                          string    `db:"role"`
	MemberCount                   int       `db:"member_count"`
	Status                        string    `db:"status"`
	LeaderboardEnabled            bool      `db:"leaderboard_enabled"`
	DefaultPlannerIntervalMinutes int       `db:"default_planner_interval_minutes"`
	CreatedAt                     time.Time `db:"created_at"`
	UpdatedAt                     time.Time `db:"updated_at"`
}

type WorkspaceMemberRow struct {
	WorkspaceID string    `db:"workspace_id"`
	UserID      string    `db:"user_id"`
	FullName    string    `db:"full_name"`
	Email       string    `db:"email"`
	AvatarURL   *string   `db:"avatar_url"`
	Role        string    `db:"role"`
	Status      string    `db:"status"`
	JoinedAt    time.Time `db:"joined_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type WorkspaceInvitationRow struct {
	ID          string    `db:"id"`
	WorkspaceID string    `db:"workspace_id"`
	Email       string    `db:"email"`
	Role        string    `db:"role"`
	Status      string    `db:"status"`
	ExpiresAt   time.Time `db:"expires_at"`
}

type TeamRow struct {
	ID          string    `db:"id"`
	WorkspaceID string    `db:"workspace_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	MemberCount int       `db:"member_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type TeamSummaryRow struct {
	UserID string `db:"user_id"`
	ID     string `db:"id"`
	Name   string `db:"name"`
}
