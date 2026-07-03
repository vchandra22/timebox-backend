package entity

import "time"

type WorkspaceSettings struct {
	LeaderboardEnabled            bool
	DefaultPlannerIntervalMinutes int
}

type Workspace struct {
	ID          string
	Name        string
	Slug        string
	Timezone    string
	LogoURL     *string
	OwnerID     string
	OwnerName   string
	Role        string
	MemberCount int
	Status      string
	Settings    WorkspaceSettings
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMember struct {
	WorkspaceID string
	UserID      string
	FullName    string
	Email       string
	AvatarURL   *string
	Role        string
	Status      string
	TeamIDs     []string
	TeamIDsSet  bool
	Teams       []TeamSummary
	JoinedAt    time.Time
	UpdatedAt   time.Time
}

type WorkspaceInvitation struct {
	ID          string
	WorkspaceID string
	Email       string
	Role        string
	Status      string
	ExpiresAt   time.Time
}

type Team struct {
	ID          string
	WorkspaceID string
	Name        string
	Description string
	MemberCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TeamSummary struct {
	ID   string
	Name string
}
