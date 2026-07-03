package service

import (
	"context"
	"errors"
	"time"

	"timebox-backend/internal/entity"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

const (
	WorkspaceRoleOwner  = "owner"
	WorkspaceRoleAdmin  = "admin"
	WorkspaceRoleMember = "member"
	WorkspaceRoleViewer = "viewer"

	WorkspaceMemberActive   = "active"
	WorkspaceMemberInvited  = "invited"
	WorkspaceMemberInactive = "inactive"
	WorkspaceMemberRemoved  = "removed"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrWorkspaceSlugUsed = errors.New("workspace slug already exists")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidRole       = errors.New("invalid role")
	ErrInvalidStatus     = errors.New("invalid status")
)

type WorkspaceService struct {
	repo workspacerepo.Repository
}

type WorkspaceSettingsPatch struct {
	LeaderboardEnabled            *bool
	DefaultPlannerIntervalMinutes *int
}

func newWorkspaceService(repo workspacerepo.Repository) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

func (s *WorkspaceService) Create(ctx context.Context, userID string, workspace entity.Workspace) (entity.Workspace, error) {
	if _, err := time.LoadLocation(workspace.Timezone); err != nil {
		return entity.Workspace{}, ErrInvalidTimezone
	}
	workspace.OwnerID = userID
	created, err := s.repo.Create(ctx, workspace)
	return created, workspaceError(err)
}

func (s *WorkspaceService) List(ctx context.Context, userID, q, status string, page, limit int) ([]entity.Workspace, int, error) {
	workspaces, total, err := s.repo.List(ctx, workspacerepo.ListFilter{UserID: userID, Q: q, Status: status, Limit: limit, Offset: (page - 1) * limit})
	return workspaces, total, workspaceError(err)
}

func (s *WorkspaceService) FindByID(ctx context.Context, userID, id string) (entity.Workspace, error) {
	if err := s.requireMember(ctx, id, userID); err != nil {
		return entity.Workspace{}, err
	}
	workspace, err := s.repo.FindByID(ctx, id)
	return workspace, workspaceError(err)
}

func (s *WorkspaceService) Update(ctx context.Context, userID string, patch entity.Workspace, settingsPatch *WorkspaceSettingsPatch) (entity.Workspace, error) {
	if err := s.requireManager(ctx, patch.ID, userID); err != nil {
		return entity.Workspace{}, err
	}
	current, err := s.repo.FindByID(ctx, patch.ID)
	if err != nil {
		return entity.Workspace{}, workspaceError(err)
	}
	if patch.Name != "" {
		current.Name = patch.Name
	}
	if patch.Slug != "" {
		current.Slug = patch.Slug
	}
	if patch.Timezone != "" {
		if _, err := time.LoadLocation(patch.Timezone); err != nil {
			return entity.Workspace{}, ErrInvalidTimezone
		}
		current.Timezone = patch.Timezone
	}
	if patch.LogoURL != nil {
		current.LogoURL = patch.LogoURL
	}
	if settingsPatch != nil {
		if settingsPatch.LeaderboardEnabled != nil {
			current.Settings.LeaderboardEnabled = *settingsPatch.LeaderboardEnabled
		}
		if settingsPatch.DefaultPlannerIntervalMinutes != nil {
			current.Settings.DefaultPlannerIntervalMinutes = *settingsPatch.DefaultPlannerIntervalMinutes
		}
	}
	updated, err := s.repo.Update(ctx, current)
	return updated, workspaceError(err)
}

func (s *WorkspaceService) InviteMember(ctx context.Context, actorID, workspaceID string, invitation entity.WorkspaceInvitation, teamIDs []string) (entity.WorkspaceInvitation, error) {
	if err := s.requireManager(ctx, workspaceID, actorID); err != nil {
		return entity.WorkspaceInvitation{}, err
	}
	if !validRole(invitation.Role) || invitation.Role == WorkspaceRoleOwner {
		return entity.WorkspaceInvitation{}, ErrInvalidRole
	}
	invitation.WorkspaceID = workspaceID
	invitation.Status = WorkspaceMemberInvited
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	created, err := s.repo.InviteMember(ctx, invitation, teamIDs)
	return created, workspaceError(err)
}

func (s *WorkspaceService) ListMembers(ctx context.Context, actorID, workspaceID, role, status, q string, page, limit int) ([]entity.WorkspaceMember, int, error) {
	if err := s.requireMember(ctx, workspaceID, actorID); err != nil {
		return nil, 0, err
	}
	members, total, err := s.repo.ListMembers(ctx, workspacerepo.ListFilter{UserID: workspaceID, Role: role, Status: status, Q: q, Limit: limit, Offset: (page - 1) * limit})
	return members, total, workspaceError(err)
}

func (s *WorkspaceService) UpdateMember(ctx context.Context, actorID, workspaceID, userID string, patch entity.WorkspaceMember) (entity.WorkspaceMember, error) {
	actor, err := s.manager(ctx, workspaceID, actorID)
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	current, err := s.repo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return entity.WorkspaceMember{}, workspaceError(err)
	}
	if current.Role == WorkspaceRoleOwner && actor.Role != WorkspaceRoleOwner {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	if patch.Role != "" {
		if !validRole(patch.Role) {
			return entity.WorkspaceMember{}, ErrInvalidRole
		}
		current.Role = patch.Role
	}
	if patch.Status != "" {
		if !validStatus(patch.Status) {
			return entity.WorkspaceMember{}, ErrInvalidStatus
		}
		current.Status = patch.Status
	}
	current.TeamIDs = patch.TeamIDs
	current.TeamIDsSet = patch.TeamIDsSet
	updated, err := s.repo.UpdateMember(ctx, current)
	return updated, workspaceError(err)
}

func (s *WorkspaceService) ListTeams(ctx context.Context, actorID, workspaceID string) ([]entity.Team, error) {
	if err := s.requireMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	teams, err := s.repo.ListTeams(ctx, workspaceID)
	return teams, workspaceError(err)
}

func (s *WorkspaceService) CreateTeam(ctx context.Context, actorID, workspaceID string, team entity.Team) (entity.Team, error) {
	if err := s.requireManager(ctx, workspaceID, actorID); err != nil {
		return entity.Team{}, err
	}
	team.WorkspaceID = workspaceID
	created, err := s.repo.CreateTeam(ctx, team)
	return created, workspaceError(err)
}

func (s *WorkspaceService) UpdateTeam(ctx context.Context, actorID string, patch entity.Team) (entity.Team, error) {
	current, err := s.repo.FindTeam(ctx, patch.ID)
	if err != nil {
		return entity.Team{}, workspaceError(err)
	}
	if err := s.requireManager(ctx, current.WorkspaceID, actorID); err != nil {
		return entity.Team{}, err
	}
	if patch.Name != "" {
		current.Name = patch.Name
	}
	if patch.Description != "" {
		current.Description = patch.Description
	}
	updated, err := s.repo.UpdateTeam(ctx, current)
	return updated, workspaceError(err)
}

func (s *WorkspaceService) DeleteTeam(ctx context.Context, actorID, teamID string) error {
	team, err := s.repo.FindTeam(ctx, teamID)
	if err != nil {
		return workspaceError(err)
	}
	if err := s.requireManager(ctx, team.WorkspaceID, actorID); err != nil {
		return err
	}
	return workspaceError(s.repo.DeleteTeam(ctx, teamID))
}

func (s *WorkspaceService) requireMember(ctx context.Context, workspaceID, userID string) error {
	member, err := s.repo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return workspaceError(err)
	}
	if member.Status != WorkspaceMemberActive {
		return ErrForbidden
	}
	return nil
}

func (s *WorkspaceService) requireManager(ctx context.Context, workspaceID, userID string) error {
	_, err := s.manager(ctx, workspaceID, userID)
	return err
}

func (s *WorkspaceService) manager(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, err := s.repo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return entity.WorkspaceMember{}, workspaceError(err)
	}
	if member.Status != WorkspaceMemberActive || (member.Role != WorkspaceRoleOwner && member.Role != WorkspaceRoleAdmin) {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	return member, nil
}

func validRole(role string) bool {
	return role == WorkspaceRoleOwner || role == WorkspaceRoleAdmin || role == WorkspaceRoleMember || role == WorkspaceRoleViewer
}

func validStatus(status string) bool {
	return status == WorkspaceMemberActive || status == WorkspaceMemberInvited || status == WorkspaceMemberInactive || status == WorkspaceMemberRemoved
}

func workspaceError(err error) error {
	if errors.Is(err, workspacerepo.ErrNotFound) {
		return ErrWorkspaceNotFound
	}
	if errors.Is(err, workspacerepo.ErrSlugExists) {
		return ErrWorkspaceSlugUsed
	}
	return err
}
