package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"timebox-backend/internal/config"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/repository/dbexecutor"
	executionrepo "timebox-backend/internal/repository/execution"
)

type Repository struct {
	db         config.PostgreSQL
	dbExecutor *dbexecutor.Executor
}

func NewRepository(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{db: db, dbExecutor: dbExecutor}
}

func (r *Repository) ListTimeboxes(ctx context.Context, filter executionrepo.TimeboxFilter) ([]entity.Timebox, error) {
	var rows []TimeboxRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTimeboxes, filter.WorkspaceID, filter.Date, filter.StartDate, filter.EndDate, filter.OwnerID, filter.Status, filter.CategoryID, filter.GoalID); err != nil {
		return nil, executionError(err)
	}
	timeboxes := make([]entity.Timebox, 0, len(rows))
	for _, row := range rows {
		timeboxes = append(timeboxes, row.toEntity())
	}
	return timeboxes, nil
}

func (r *Repository) CreateTimebox(ctx context.Context, timebox entity.Timebox) (entity.Timebox, error) {
	var id string
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &id, QueryCreateTimebox, timebox.WorkspaceID, timebox.TaskID, timebox.OwnerID, timebox.CategoryID, timebox.Title, timebox.Description, timebox.ScheduledStart, timebox.ScheduledEnd, timebox.IsBuffer); err != nil {
		return entity.Timebox{}, executionError(err)
	}
	return r.FindTimebox(ctx, id)
}

func (r *Repository) FindTimebox(ctx context.Context, id string) (entity.Timebox, error) {
	var row TimeboxRow
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindTimebox, id); err != nil {
		return entity.Timebox{}, executionError(err)
	}
	return row.toEntity(), nil
}

func (r *Repository) UpdateTimebox(ctx context.Context, timebox entity.Timebox) (entity.Timebox, error) {
	var id string
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &id, QueryUpdateTimebox, timebox.ID, timebox.TaskID, timebox.OwnerID, timebox.CategoryID, timebox.Title, timebox.Description, timebox.ScheduledStart, timebox.ScheduledEnd, timebox.Status, timebox.IsBuffer); err != nil {
		return entity.Timebox{}, executionError(err)
	}
	return r.FindTimebox(ctx, id)
}

func (r *Repository) DeleteTimebox(ctx context.Context, id string) error {
	var deletedID string
	return executionError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteTimebox, id))
}

func (r *Repository) UpdateTimeboxStatus(ctx context.Context, id, status string, actualMinutes int) (entity.Timebox, error) {
	var updatedID string
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &updatedID, QueryUpdateTimeboxStatus, id, status, actualMinutes); err != nil {
		return entity.Timebox{}, executionError(err)
	}
	return r.FindTimebox(ctx, updatedID)
}

func (r *Repository) CreateTimeLog(ctx context.Context, log entity.TimeLog) (entity.TimeLog, error) {
	var row TimeLogRow
	err := executionError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateTimeLog, log.TimeboxID, log.StartedAt, log.EndedAt, log.Source, log.Note, log.CreatedBy))
	return row.toEntity(), err
}

func (r *Repository) CloseRunningLog(ctx context.Context, timeboxID string, endedAt time.Time, note *string) error {
	return executionError(r.dbExecutor.Exec(ctx, r.db.Conn, QueryCloseRunningLog, timeboxID, endedAt, note))
}

func (r *Repository) ListTimeLogs(ctx context.Context, timeboxID string) ([]entity.TimeLog, error) {
	var rows []TimeLogRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTimeLogs, timeboxID); err != nil {
		return nil, executionError(err)
	}
	logs := make([]entity.TimeLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, row.toEntity())
	}
	return logs, nil
}

func (r *Repository) SumTimeLogSeconds(ctx context.Context, timeboxID string) (int, error) {
	var total int
	err := executionError(r.dbExecutor.Get(ctx, r.db.Conn, &total, QuerySumTimeLogSeconds, timeboxID))
	return total, err
}

func executionError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return executionrepo.ErrNotFound
	}
	return err
}

func (r TimeboxRow) toEntity() entity.Timebox {
	timebox := entity.Timebox{
		ID:             r.ID,
		WorkspaceID:    r.WorkspaceID,
		TaskID:         r.TaskID,
		OwnerID:        r.OwnerID,
		CategoryID:     r.CategoryID,
		Title:          r.Title,
		Description:    r.Description,
		ScheduledStart: r.ScheduledStart,
		ScheduledEnd:   r.ScheduledEnd,
		PlannedMinutes: r.PlannedMinutes,
		ActualMinutes:  r.ActualMinutes,
		Status:         r.Status,
		IsBuffer:       r.IsBuffer,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.OwnerName != nil {
		timebox.Owner = &entity.UserSummary{ID: r.OwnerID, FullName: *r.OwnerName, AvatarURL: r.OwnerAvatar}
	}
	if r.CategoryID != nil && r.CategoryName != nil && r.CategoryColor != nil {
		timebox.Category = &entity.CategorySummary{ID: *r.CategoryID, Name: *r.CategoryName, Color: *r.CategoryColor}
	}
	if r.TaskID != nil && r.TaskTitle != nil {
		timebox.Task = &entity.TaskSummary{ID: *r.TaskID, Title: *r.TaskTitle}
	}
	return timebox
}

func (r TimeLogRow) toEntity() entity.TimeLog {
	return entity.TimeLog{ID: r.ID, TimeboxID: r.TimeboxID, WorkspaceID: r.WorkspaceID, StartedAt: r.StartedAt, EndedAt: r.EndedAt, DurationSeconds: r.DurationSeconds, Source: r.Source, Note: r.Note, CreatedBy: r.CreatedBy}
}
