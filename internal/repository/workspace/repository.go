package workspace

import (
	"context"
	"errors"

	"timebox-backend/internal/entity"
)

var (
	ErrNotFound     = errors.New("workspace not found")
	ErrSlugExists   = errors.New("workspace slug already exists")
	ErrMemberExists = errors.New("workspace member already exists")
)

type ListFilter struct {
	UserID string
	Q      string
	Status string
	Role   string
	Limit  int
	Offset int
}

type Repository interface {
	Create(ctx context.Context, workspace entity.Workspace) (entity.Workspace, error)
	List(ctx context.Context, filter ListFilter) ([]entity.Workspace, int, error)
	FindByID(ctx context.Context, id string) (entity.Workspace, error)
	Update(ctx context.Context, workspace entity.Workspace) (entity.Workspace, error)
	FindMember(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error)
	ListMembers(ctx context.Context, filter ListFilter) ([]entity.WorkspaceMember, int, error)
	InviteMember(ctx context.Context, invitation entity.WorkspaceInvitation, teamIDs []string) (entity.WorkspaceInvitation, error)
	UpdateMember(ctx context.Context, member entity.WorkspaceMember) (entity.WorkspaceMember, error)
	ListTeams(ctx context.Context, workspaceID string) ([]entity.Team, error)
	CreateTeam(ctx context.Context, team entity.Team) (entity.Team, error)
	FindTeam(ctx context.Context, id string) (entity.Team, error)
	UpdateTeam(ctx context.Context, team entity.Team) (entity.Team, error)
	DeleteTeam(ctx context.Context, id string) error
}
