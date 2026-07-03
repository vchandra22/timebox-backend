package planning

import (
	"context"
	"errors"

	"timebox-backend/internal/entity"
)

var (
	ErrNotFound = errors.New("planning resource not found")
	ErrConflict = errors.New("planning resource conflict")
)

type GoalFilter struct {
	WorkspaceID string
	Q           string
	Status      string
	PinnedSet   bool
	Pinned      bool
	Limit       int
	Offset      int
}

type TaskFilter struct {
	WorkspaceID string
	Q           string
	Status      string
	Priority    string
	GoalID      string
	AssigneeID  string
	CategoryID  string
	TagIDs      []string
	IncludeDone bool
	Sort        string
	Order       string
	Limit       int
	Offset      int
}

type Repository interface {
	ListCategories(ctx context.Context, workspaceID string) ([]entity.Category, error)
	CreateCategory(ctx context.Context, category entity.Category) (entity.Category, error)
	FindCategory(ctx context.Context, id string) (entity.Category, error)
	UpdateCategory(ctx context.Context, category entity.Category) (entity.Category, error)
	DeleteCategory(ctx context.Context, id string) error
	ListTags(ctx context.Context, workspaceID, q string) ([]entity.Tag, error)
	CreateTag(ctx context.Context, tag entity.Tag) (entity.Tag, error)
	FindTag(ctx context.Context, id string) (entity.Tag, error)
	UpdateTag(ctx context.Context, tag entity.Tag) (entity.Tag, error)
	DeleteTag(ctx context.Context, id string) error
	ListGoals(ctx context.Context, filter GoalFilter) ([]entity.Goal, int, error)
	CreateGoal(ctx context.Context, goal entity.Goal) (entity.Goal, error)
	FindGoal(ctx context.Context, id string) (entity.Goal, error)
	UpdateGoal(ctx context.Context, goal entity.Goal) (entity.Goal, error)
	ArchiveGoal(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filter TaskFilter) ([]entity.Task, int, error)
	CreateTask(ctx context.Context, task entity.Task) (entity.Task, error)
	FindTask(ctx context.Context, id string) (entity.Task, error)
	UpdateTask(ctx context.Context, task entity.Task) (entity.Task, error)
	DeleteTask(ctx context.Context, id string) error
	MoveTask(ctx context.Context, id, toStatus string, position int) (entity.TaskMove, error)
}
