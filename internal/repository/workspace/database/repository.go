package database

import (
	"context"
	"database/sql"
	"errors"

	"timebox-backend/internal/config"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/repository/dbexecutor"
	workspacerepo "timebox-backend/internal/repository/workspace"

	"github.com/jackc/pgx/v5/pgconn"
)

const postgresUniqueViolation = "23505"

type Repository struct {
	db         config.PostgreSQL
	dbExecutor *dbexecutor.Executor
}

func NewRepository(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{db: db, dbExecutor: dbExecutor}
}

func (r *Repository) Create(ctx context.Context, workspace entity.Workspace) (entity.Workspace, error) {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return entity.Workspace{}, err
	}
	defer tx.Rollback()

	var row WorkspaceRow
	err = workspaceError(r.dbExecutor.Get(ctx, tx, &row, QueryCreateWorkspace, workspace.Name, workspace.Slug, workspace.Timezone, workspace.LogoURL, workspace.OwnerID))
	if err != nil {
		return entity.Workspace{}, err
	}
	if err := r.dbExecutor.Exec(ctx, tx, QueryCreateOwnerMember, row.ID, workspace.OwnerID); err != nil {
		return entity.Workspace{}, workspaceError(err)
	}
	return row.toEntity(), tx.Commit()
}

func (r *Repository) List(ctx context.Context, filter workspacerepo.ListFilter) ([]entity.Workspace, int, error) {
	var total int
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &total, QueryCountWorkspaces, filter.UserID, filter.Q, filter.Status); err != nil {
		return nil, 0, workspaceError(err)
	}

	var rows []WorkspaceRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListWorkspaces, filter.UserID, filter.Q, filter.Status, filter.Limit, filter.Offset); err != nil {
		return nil, 0, workspaceError(err)
	}

	workspaces := make([]entity.Workspace, 0, len(rows))
	for _, row := range rows {
		workspaces = append(workspaces, row.toEntity())
	}
	return workspaces, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (entity.Workspace, error) {
	var row WorkspaceRow
	err := workspaceError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindWorkspaceByID, id))
	return row.toEntity(), err
}

func (r *Repository) Update(ctx context.Context, workspace entity.Workspace) (entity.Workspace, error) {
	var row WorkspaceRow
	err := workspaceError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateWorkspace, workspace.ID, workspace.Name, workspace.Slug, workspace.Timezone, workspace.LogoURL, workspace.Settings.LeaderboardEnabled, workspace.Settings.DefaultPlannerIntervalMinutes))
	return row.toEntity(), err
}

func (r *Repository) FindMember(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	var row WorkspaceMemberRow
	err := workspaceError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindMember, workspaceID, userID))
	return row.toEntity(), err
}

func (r *Repository) ListMembers(ctx context.Context, filter workspacerepo.ListFilter) ([]entity.WorkspaceMember, int, error) {
	var total int
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &total, QueryCountMembers, filter.UserID, filter.Role, filter.Status, filter.Q); err != nil {
		return nil, 0, workspaceError(err)
	}

	var rows []WorkspaceMemberRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListMembers, filter.UserID, filter.Role, filter.Status, filter.Q, filter.Limit, filter.Offset); err != nil {
		return nil, 0, workspaceError(err)
	}

	members := make([]entity.WorkspaceMember, 0, len(rows))
	for _, row := range rows {
		member := row.toEntity()
		member.Teams = r.memberTeams(ctx, filter.UserID, member.UserID)
		members = append(members, member)
	}
	return members, total, nil
}

func (r *Repository) InviteMember(ctx context.Context, invitation entity.WorkspaceInvitation, teamIDs []string) (entity.WorkspaceInvitation, error) {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return entity.WorkspaceInvitation{}, err
	}
	defer tx.Rollback()

	var row WorkspaceInvitationRow
	err = workspaceError(r.dbExecutor.Get(ctx, tx, &row, QueryCreateInvitation, invitation.WorkspaceID, invitation.Email, invitation.Role, invitation.ExpiresAt))
	if err != nil {
		return entity.WorkspaceInvitation{}, err
	}

	var userID string
	if err := tx.GetContext(ctx, &userID, QueryFindUserIDByEmail, invitation.Email); err == nil {
		if err := r.dbExecutor.Exec(ctx, tx, QueryUpsertInvitedMember, invitation.WorkspaceID, userID, invitation.Role); err != nil {
			return entity.WorkspaceInvitation{}, workspaceError(err)
		}
		for _, teamID := range teamIDs {
			if err := r.dbExecutor.Exec(ctx, tx, QueryUpsertTeamMember, teamID, userID, invitation.WorkspaceID); err != nil {
				return entity.WorkspaceInvitation{}, workspaceError(err)
			}
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return entity.WorkspaceInvitation{}, workspaceError(err)
	}

	return row.toEntity(), tx.Commit()
}

func (r *Repository) UpdateMember(ctx context.Context, member entity.WorkspaceMember) (entity.WorkspaceMember, error) {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	defer tx.Rollback()

	var row WorkspaceMemberRow
	err = workspaceError(r.dbExecutor.Get(ctx, tx, &row, QueryUpdateMember, member.WorkspaceID, member.UserID, member.Role, member.Status))
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	if member.TeamIDsSet {
		if err := r.dbExecutor.Exec(ctx, tx, QuerySoftDeleteTeamMembersForUser, member.WorkspaceID, member.UserID); err != nil {
			return entity.WorkspaceMember{}, workspaceError(err)
		}
		for _, teamID := range member.TeamIDs {
			if err := r.dbExecutor.Exec(ctx, tx, QueryUpsertTeamMember, teamID, member.UserID, member.WorkspaceID); err != nil {
				return entity.WorkspaceMember{}, workspaceError(err)
			}
		}
	}
	updated := row.toEntity()
	updated.TeamIDs = member.TeamIDs
	return updated, tx.Commit()
}

func (r *Repository) ListTeams(ctx context.Context, workspaceID string) ([]entity.Team, error) {
	var rows []TeamRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTeams, workspaceID); err != nil {
		return nil, workspaceError(err)
	}

	teams := make([]entity.Team, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, row.toEntity())
	}
	return teams, nil
}

func (r *Repository) CreateTeam(ctx context.Context, team entity.Team) (entity.Team, error) {
	var row TeamRow
	err := workspaceError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateTeam, team.WorkspaceID, team.Name, team.Description))
	return row.toEntity(), err
}

func (r *Repository) FindTeam(ctx context.Context, id string) (entity.Team, error) {
	var row TeamRow
	err := workspaceError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindTeam, id))
	return row.toEntity(), err
}

func (r *Repository) UpdateTeam(ctx context.Context, team entity.Team) (entity.Team, error) {
	var row TeamRow
	err := workspaceError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateTeam, team.ID, team.Name, team.Description))
	return row.toEntity(), err
}

func (r *Repository) DeleteTeam(ctx context.Context, id string) error {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var deletedID string
	if err := r.dbExecutor.Get(ctx, tx, &deletedID, QueryDeleteTeam, id); err != nil {
		return workspaceError(err)
	}
	if err := r.dbExecutor.Exec(ctx, tx, QuerySoftDeleteTeamMembers, id); err != nil {
		return workspaceError(err)
	}
	return tx.Commit()
}

func (r *Repository) memberTeams(ctx context.Context, workspaceID, userID string) []entity.TeamSummary {
	var rows []TeamSummaryRow
	// ponytail: N+1 is enough for paginated member lists; batch this if member pages get large.
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListMemberTeams, workspaceID, userID); err != nil {
		return nil
	}
	teams := make([]entity.TeamSummary, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, entity.TeamSummary{ID: row.ID, Name: row.Name})
	}
	return teams
}

func workspaceError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return workspacerepo.ErrNotFound
	}
	if isUniqueViolation(err) {
		return workspacerepo.ErrSlugExists
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation
}

func (r WorkspaceRow) toEntity() entity.Workspace {
	return entity.Workspace{
		ID:          r.ID,
		Name:        r.Name,
		Slug:        r.Slug,
		Timezone:    r.Timezone,
		LogoURL:     r.LogoURL,
		OwnerID:     r.OwnerID,
		OwnerName:   r.OwnerName,
		Role:        r.Role,
		MemberCount: r.MemberCount,
		Status:      r.Status,
		Settings: entity.WorkspaceSettings{
			LeaderboardEnabled:            r.LeaderboardEnabled,
			DefaultPlannerIntervalMinutes: r.DefaultPlannerIntervalMinutes,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func (r WorkspaceMemberRow) toEntity() entity.WorkspaceMember {
	return entity.WorkspaceMember{
		WorkspaceID: r.WorkspaceID,
		UserID:      r.UserID,
		FullName:    r.FullName,
		Email:       r.Email,
		AvatarURL:   r.AvatarURL,
		Role:        r.Role,
		Status:      r.Status,
		JoinedAt:    r.JoinedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (r WorkspaceInvitationRow) toEntity() entity.WorkspaceInvitation {
	return entity.WorkspaceInvitation{
		ID:          r.ID,
		WorkspaceID: r.WorkspaceID,
		Email:       r.Email,
		Role:        r.Role,
		Status:      r.Status,
		ExpiresAt:   r.ExpiresAt,
	}
}

func (r TeamRow) toEntity() entity.Team {
	return entity.Team{
		ID:          r.ID,
		WorkspaceID: r.WorkspaceID,
		Name:        r.Name,
		Description: r.Description,
		MemberCount: r.MemberCount,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
