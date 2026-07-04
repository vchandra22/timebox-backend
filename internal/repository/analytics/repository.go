package analytics

import (
	"context"
	"errors"
	"time"

	"timebox-backend/internal/entity"
)

var ErrNotFound = errors.New("analytics resource not found")

type ReportFilter struct {
	WorkspaceID string
	StartDate   time.Time
	EndDate     time.Time
	GroupBy     string
	UserID      string
	CategoryID  string
	GoalID      string
	Interval    string
}

type ActivityLogFilter struct {
	WorkspaceID  string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

type Repository interface {
	Streak(ctx context.Context, workspaceID, userID string) (entity.Streak, error)
	Leaderboard(ctx context.Context, workspaceID, metric, period string, limit int) (entity.Leaderboard, error)
	PersonalDashboard(ctx context.Context, workspaceID, userID string, date time.Time) (entity.PersonalDashboard, error)
	WorkspaceDashboard(ctx context.Context, workspaceID string, date time.Time) (entity.WorkspaceDashboard, error)
	TimeAllocationReport(ctx context.Context, filter ReportFilter) (entity.TimeAllocationReport, error)
	ProductivityTrendReport(ctx context.Context, filter ReportFilter) (entity.ProductivityTrendReport, error)
	TeamWorkloadReport(ctx context.Context, filter ReportFilter) (entity.TeamWorkloadReport, error)
	Search(ctx context.Context, workspaceID, query string, types []string, limit int) (entity.SearchResults, error)
	ListSavedViews(ctx context.Context, workspaceID, userID, resourceType string) ([]entity.SavedView, error)
	CreateSavedView(ctx context.Context, view entity.SavedView) (entity.SavedView, error)
	FindSavedView(ctx context.Context, id string) (entity.SavedView, error)
	UpdateSavedView(ctx context.Context, view entity.SavedView) (entity.SavedView, error)
	DeleteSavedView(ctx context.Context, id string) error
	ListActivityLogs(ctx context.Context, filter ActivityLogFilter) ([]entity.ActivityLog, int, error)
}
