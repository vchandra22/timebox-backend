package execution

import (
	"context"
	"errors"
	"time"

	"timebox-backend/internal/entity"
)

var (
	ErrNotFound            = errors.New("execution resource not found")
	ErrTimerNotFound       = errors.New("timer not found")
	ErrTimerAlreadyRunning = errors.New("timer already running")
)

type TimeboxFilter struct {
	WorkspaceID string
	Date        *time.Time
	StartDate   *time.Time
	EndDate     *time.Time
	OwnerID     string
	Status      string
	CategoryID  string
	GoalID      string
}

type Repository interface {
	ListTimeboxes(ctx context.Context, filter TimeboxFilter) ([]entity.Timebox, error)
	CreateTimebox(ctx context.Context, timebox entity.Timebox) (entity.Timebox, error)
	FindTimebox(ctx context.Context, id string) (entity.Timebox, error)
	UpdateTimebox(ctx context.Context, timebox entity.Timebox) (entity.Timebox, error)
	DeleteTimebox(ctx context.Context, id string) error
	UpdateTimeboxStatus(ctx context.Context, id, status string, actualMinutes int) (entity.Timebox, error)
	CreateTimeLog(ctx context.Context, log entity.TimeLog) (entity.TimeLog, error)
	CloseRunningLog(ctx context.Context, timeboxID string, endedAt time.Time, note *string) error
	ListTimeLogs(ctx context.Context, timeboxID string) ([]entity.TimeLog, error)
	SumTimeLogSeconds(ctx context.Context, timeboxID string) (int, error)
}

type TimerRepository interface {
	Get(ctx context.Context, userID string) (entity.TimerState, error)
	Save(ctx context.Context, state entity.TimerState) error
	Delete(ctx context.Context, userID string) error
}
