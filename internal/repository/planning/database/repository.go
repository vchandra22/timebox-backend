package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"timebox-backend/internal/config"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/repository/dbexecutor"
	planningrepo "timebox-backend/internal/repository/planning"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

const postgresUniqueViolation = "23505"

type Repository struct {
	db         config.PostgreSQL
	dbExecutor *dbexecutor.Executor
}

func NewRepository(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{db: db, dbExecutor: dbExecutor}
}

func (r *Repository) ListCategories(ctx context.Context, workspaceID string) ([]entity.Category, error) {
	var rows []CategoryRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListCategories, workspaceID); err != nil {
		return nil, planningError(err)
	}
	categories := make([]entity.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, row.toEntity())
	}
	return categories, nil
}

func (r *Repository) CreateCategory(ctx context.Context, category entity.Category) (entity.Category, error) {
	var row CategoryRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateCategory, category.WorkspaceID, category.Name, category.Color))
	return row.toEntity(), err
}

func (r *Repository) FindCategory(ctx context.Context, id string) (entity.Category, error) {
	var row CategoryRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindCategory, id))
	return row.toEntity(), err
}

func (r *Repository) UpdateCategory(ctx context.Context, category entity.Category) (entity.Category, error) {
	var row CategoryRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateCategory, category.ID, category.Name, category.Color))
	return row.toEntity(), err
}

func (r *Repository) DeleteCategory(ctx context.Context, id string) error {
	var deletedID string
	return planningError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteCategory, id))
}

func (r *Repository) ListTags(ctx context.Context, workspaceID, q string) ([]entity.Tag, error) {
	var rows []TagRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTags, workspaceID, q); err != nil {
		return nil, planningError(err)
	}
	tags := make([]entity.Tag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, row.toEntity())
	}
	return tags, nil
}

func (r *Repository) CreateTag(ctx context.Context, tag entity.Tag) (entity.Tag, error) {
	var row TagRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateTag, tag.WorkspaceID, tag.Name))
	return row.toEntity(), err
}

func (r *Repository) FindTag(ctx context.Context, id string) (entity.Tag, error) {
	var row TagRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindTag, id))
	return row.toEntity(), err
}

func (r *Repository) UpdateTag(ctx context.Context, tag entity.Tag) (entity.Tag, error) {
	var row TagRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateTag, tag.ID, tag.Name))
	return row.toEntity(), err
}

func (r *Repository) DeleteTag(ctx context.Context, id string) error {
	var deletedID string
	return planningError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteTag, id))
}

func (r *Repository) ListGoals(ctx context.Context, filter planningrepo.GoalFilter) ([]entity.Goal, int, error) {
	var total int
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &total, QueryCountGoals, filter.WorkspaceID, filter.Q, filter.Status, filter.PinnedSet, filter.Pinned); err != nil {
		return nil, 0, planningError(err)
	}
	var rows []GoalRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListGoals, filter.WorkspaceID, filter.Q, filter.Status, filter.PinnedSet, filter.Pinned, filter.Limit, filter.Offset); err != nil {
		return nil, 0, planningError(err)
	}
	goals := make([]entity.Goal, 0, len(rows))
	for _, row := range rows {
		goals = append(goals, row.toEntity())
	}
	return goals, total, nil
}

func (r *Repository) CreateGoal(ctx context.Context, goal entity.Goal) (entity.Goal, error) {
	var row GoalRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateGoal, goal.WorkspaceID, goal.CreatedBy, goal.Title, goal.Description, goal.TargetDate, goal.IsPinned))
	return row.toEntity(), err
}

func (r *Repository) FindGoal(ctx context.Context, id string) (entity.Goal, error) {
	var row GoalRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindGoal, id))
	return row.toEntity(), err
}

func (r *Repository) UpdateGoal(ctx context.Context, goal entity.Goal) (entity.Goal, error) {
	var row GoalRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateGoal, goal.ID, goal.Title, goal.Description, goal.TargetDate, goal.Status, goal.IsPinned))
	return row.toEntity(), err
}

func (r *Repository) ArchiveGoal(ctx context.Context, id string) error {
	var archivedID string
	return planningError(r.dbExecutor.Get(ctx, r.db.Conn, &archivedID, QueryArchiveGoal, id))
}

func (r *Repository) ListTasks(ctx context.Context, filter planningrepo.TaskFilter) ([]entity.Task, int, error) {
	tagIDs := strings.Join(filter.TagIDs, ",")
	var total int
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &total, QueryCountTasks, filter.WorkspaceID, filter.Q, filter.Status, filter.Priority, filter.GoalID, filter.AssigneeID, filter.CategoryID, filter.IncludeDone, tagIDs); err != nil {
		return nil, 0, planningError(err)
	}
	var rows []TaskRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTasks, filter.WorkspaceID, filter.Q, filter.Status, filter.Priority, filter.GoalID, filter.AssigneeID, filter.CategoryID, filter.IncludeDone, tagIDs, filter.Limit, filter.Offset); err != nil {
		return nil, 0, planningError(err)
	}
	tasks := make([]entity.Task, 0, len(rows))
	for _, row := range rows {
		task := row.toEntity()
		task.Tags = r.taskTags(ctx, task.ID)
		tasks = append(tasks, task)
	}
	return tasks, total, nil
}

func (r *Repository) CreateTask(ctx context.Context, task entity.Task) (entity.Task, error) {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return entity.Task{}, err
	}
	defer tx.Rollback()

	var row TaskRow
	err = planningError(r.dbExecutor.Get(ctx, tx, &row, QueryCreateTask, task.WorkspaceID, task.GoalID, task.AssigneeID, task.CategoryID, task.CreatedBy, task.Title, task.Description, task.Priority, task.EstimatedMinutes))
	if err != nil {
		return entity.Task{}, err
	}
	if err := r.replaceTaskTags(ctx, tx, row.ID, row.WorkspaceID, task.TagIDs); err != nil {
		return entity.Task{}, err
	}
	for index, item := range task.Checklist {
		if err := r.dbExecutor.Exec(ctx, tx, QueryInsertChecklist, row.ID, item.Title, (index+1)*1000); err != nil {
			return entity.Task{}, planningError(err)
		}
	}
	created := row.toEntity()
	created.TagIDs = task.TagIDs
	return created, tx.Commit()
}

func (r *Repository) FindTask(ctx context.Context, id string) (entity.Task, error) {
	var row TaskRow
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindTask, id); err != nil {
		return entity.Task{}, planningError(err)
	}
	task := row.toEntity()
	task.Tags = r.taskTags(ctx, task.ID)
	task.Checklist = r.taskChecklist(ctx, task.ID)
	return task, nil
}

func (r *Repository) UpdateTask(ctx context.Context, task entity.Task) (entity.Task, error) {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return entity.Task{}, err
	}
	defer tx.Rollback()

	var row TaskRow
	err = planningError(r.dbExecutor.Get(ctx, tx, &row, QueryUpdateTask, task.ID, task.GoalID, task.AssigneeID, task.CategoryID, task.Title, task.Description, task.Status, task.Priority, task.EstimatedMinutes))
	if err != nil {
		return entity.Task{}, err
	}
	if task.TagIDsSet {
		if err := r.replaceTaskTags(ctx, tx, row.ID, row.WorkspaceID, task.TagIDs); err != nil {
			return entity.Task{}, err
		}
	}
	updated := row.toEntity()
	updated.TagIDs = task.TagIDs
	return updated, tx.Commit()
}

func (r *Repository) DeleteTask(ctx context.Context, id string) error {
	var deletedID string
	return planningError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteTask, id))
}

func (r *Repository) MoveTask(ctx context.Context, id, toStatus string, position int) (entity.TaskMove, error) {
	var row TaskMoveRow
	err := planningError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryMoveTask, id, toStatus, position))
	return row.toEntity(), err
}

func (r *Repository) replaceTaskTags(ctx context.Context, tx dbTx, taskID, workspaceID string, tagIDs []string) error {
	if err := r.dbExecutor.Exec(ctx, tx, QueryDeleteTaskTags, taskID); err != nil {
		return planningError(err)
	}
	for _, tagID := range tagIDs {
		if tagID == "" {
			continue
		}
		if err := r.dbExecutor.Exec(ctx, tx, QueryInsertTaskTag, taskID, tagID, workspaceID); err != nil {
			return planningError(err)
		}
	}
	return nil
}

type dbTx interface{ sqlx.ExtContext }

func (r *Repository) taskTags(ctx context.Context, taskID string) []entity.Tag {
	var rows []TagRow
	// ponytail: N+1 is acceptable for small paginated task pages; batch if list pages get large.
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTaskTags, taskID); err != nil {
		return nil
	}
	tags := make([]entity.Tag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, row.toEntity())
	}
	return tags
}

func (r *Repository) taskChecklist(ctx context.Context, taskID string) []entity.TaskChecklist {
	var rows []ChecklistRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListTaskChecklist, taskID); err != nil {
		return nil
	}
	items := make([]entity.TaskChecklist, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toEntity())
	}
	return items
}

func planningError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return planningrepo.ErrNotFound
	}
	if isUniqueViolation(err) {
		return planningrepo.ErrConflict
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation
}

func (r CategoryRow) toEntity() entity.Category {
	return entity.Category{ID: r.ID, WorkspaceID: r.WorkspaceID, Name: r.Name, Color: r.Color, IsDefault: r.IsDefault, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (r TagRow) toEntity() entity.Tag {
	return entity.Tag{ID: r.ID, WorkspaceID: r.WorkspaceID, Name: r.Name, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (r GoalRow) toEntity() entity.Goal {
	return entity.Goal{
		ID:              r.ID,
		WorkspaceID:     r.WorkspaceID,
		CreatedBy:       r.CreatedBy,
		CreatedByName:   r.CreatedByName,
		Title:           r.Title,
		Description:     r.Description,
		TargetDate:      r.TargetDate,
		Status:          r.Status,
		IsPinned:        r.IsPinned,
		PlannedMinutes:  r.PlannedMinutes,
		ActualMinutes:   r.ActualMinutes,
		CompletedBlocks: r.CompletedBlocks,
		ProgressPercent: r.ProgressPercent,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func (r TaskRow) toEntity() entity.Task {
	task := entity.Task{
		ID:               r.ID,
		WorkspaceID:      r.WorkspaceID,
		GoalID:           r.GoalID,
		AssigneeID:       r.AssigneeID,
		CategoryID:       r.CategoryID,
		CreatedBy:        r.CreatedBy,
		Title:            r.Title,
		Description:      r.Description,
		Status:           r.Status,
		Priority:         r.Priority,
		EstimatedMinutes: r.EstimatedMinutes,
		Position:         r.Position,
		TimeboxesCount:   r.TimeboxesCount,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
	if r.AssigneeID != nil && r.AssigneeName != nil {
		task.Assignee = &entity.UserSummary{ID: *r.AssigneeID, FullName: *r.AssigneeName, AvatarURL: r.AssigneeAvatar}
	}
	if r.GoalID != nil && r.GoalTitle != nil {
		task.Goal = &entity.GoalSummary{ID: *r.GoalID, Title: *r.GoalTitle}
	}
	return task
}

func (r ChecklistRow) toEntity() entity.TaskChecklist {
	return entity.TaskChecklist{ID: r.ID, Title: r.Title, IsDone: r.IsDone, Position: r.Position}
}

func (r TaskMoveRow) toEntity() entity.TaskMove {
	return entity.TaskMove{ID: r.ID, WorkspaceID: r.WorkspaceID, FromStatus: r.FromStatus, ToStatus: r.ToStatus, Position: r.Position, UpdatedAt: r.UpdatedAt}
}
