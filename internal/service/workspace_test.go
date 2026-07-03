package service

import (
	"context"
	"errors"
	"testing"

	"timebox-backend/internal/entity"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

func TestWorkspacePermissions(t *testing.T) {
	repo := &workspaceRepoStub{members: map[string]entity.WorkspaceMember{
		"ws-1:owner-1":  {WorkspaceID: "ws-1", UserID: "owner-1", Role: WorkspaceRoleOwner, Status: WorkspaceMemberActive},
		"ws-1:admin-1":  {WorkspaceID: "ws-1", UserID: "admin-1", Role: WorkspaceRoleAdmin, Status: WorkspaceMemberActive},
		"ws-1:member-1": {WorkspaceID: "ws-1", UserID: "member-1", Role: WorkspaceRoleMember, Status: WorkspaceMemberActive},
	}}
	svc := newWorkspaceService(repo)

	if _, err := svc.CreateTeam(context.Background(), "member-1", "ws-1", entity.Team{Name: "Backend"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("CreateTeam err = %v, want ErrForbidden", err)
	}
	if _, err := svc.UpdateMember(context.Background(), "admin-1", "ws-1", "owner-1", entity.WorkspaceMember{Role: WorkspaceRoleMember}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("UpdateMember err = %v, want ErrForbidden", err)
	}
	if _, err := svc.InviteMember(context.Background(), "owner-1", "ws-1", entity.WorkspaceInvitation{Role: WorkspaceRoleOwner}, nil); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("InviteMember err = %v, want ErrInvalidRole", err)
	}
}

type workspaceRepoStub struct {
	members map[string]entity.WorkspaceMember
}

func (r *workspaceRepoStub) Create(context.Context, entity.Workspace) (entity.Workspace, error) {
	return entity.Workspace{}, nil
}

func (r *workspaceRepoStub) List(context.Context, workspacerepo.ListFilter) ([]entity.Workspace, int, error) {
	return nil, 0, nil
}

func (r *workspaceRepoStub) FindByID(context.Context, string) (entity.Workspace, error) {
	return entity.Workspace{}, nil
}

func (r *workspaceRepoStub) Update(context.Context, entity.Workspace) (entity.Workspace, error) {
	return entity.Workspace{}, nil
}

func (r *workspaceRepoStub) FindMember(_ context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, ok := r.members[workspaceID+":"+userID]
	if !ok {
		return entity.WorkspaceMember{}, workspacerepo.ErrNotFound
	}
	return member, nil
}

func (r *workspaceRepoStub) ListMembers(context.Context, workspacerepo.ListFilter) ([]entity.WorkspaceMember, int, error) {
	return nil, 0, nil
}

func (r *workspaceRepoStub) InviteMember(_ context.Context, invitation entity.WorkspaceInvitation, _ []string) (entity.WorkspaceInvitation, error) {
	return invitation, nil
}

func (r *workspaceRepoStub) UpdateMember(_ context.Context, member entity.WorkspaceMember) (entity.WorkspaceMember, error) {
	return member, nil
}

func (r *workspaceRepoStub) ListTeams(context.Context, string) ([]entity.Team, error) {
	return nil, nil
}

func (r *workspaceRepoStub) CreateTeam(_ context.Context, team entity.Team) (entity.Team, error) {
	return team, nil
}

func (r *workspaceRepoStub) FindTeam(context.Context, string) (entity.Team, error) {
	return entity.Team{}, nil
}

func (r *workspaceRepoStub) UpdateTeam(_ context.Context, team entity.Team) (entity.Team, error) {
	return team, nil
}

func (r *workspaceRepoStub) DeleteTeam(context.Context, string) error {
	return nil
}
