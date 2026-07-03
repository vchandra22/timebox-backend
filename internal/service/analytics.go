package service

import (
	"context"
	"errors"
	"time"

	"timebox-backend/internal/entity"
	analyticsrepo "timebox-backend/internal/repository/analytics"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

var (
	ErrAnalyticsNotFound      = errors.New("analytics resource not found")
	ErrInvalidAnalyticsRange  = errors.New("invalid analytics date range")
	ErrInvalidAnalyticsFilter = errors.New("invalid analytics filter")
)

type AnalyticsService struct {
	repo          analyticsrepo.Repository
	workspaceRepo workspacerepo.Repository
}

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

func newAnalyticsService(repo analyticsrepo.Repository, workspaceRepo workspacerepo.Repository) *AnalyticsService {
	return &AnalyticsService{repo: repo, workspaceRepo: workspaceRepo}
}

func (s *AnalyticsService) Streak(ctx context.Context, actorID, workspaceID string) (entity.Streak, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.Streak{}, err
	}
	streak, err := s.repo.Streak(ctx, workspaceID, actorID)
	return streak, analyticsError(err)
}

func (s *AnalyticsService) Leaderboard(ctx context.Context, actorID, workspaceID, metric, period string, limit int) (entity.Leaderboard, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.Leaderboard{}, err
	}
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return entity.Leaderboard{}, err
	}
	if !workspace.Settings.LeaderboardEnabled {
		return entity.Leaderboard{}, ErrForbidden
	}
	if metric == "" {
		metric = "focus_minutes"
	}
	if period == "" {
		period = "week"
	}
	if limit < 1 {
		limit = 10
	}
	leaderboard, err := s.repo.Leaderboard(ctx, workspaceID, metric, period, limit)
	return leaderboard, analyticsError(err)
}

func (s *AnalyticsService) PersonalDashboard(ctx context.Context, actorID, workspaceID string, date time.Time) (entity.PersonalDashboard, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.PersonalDashboard{}, err
	}
	dashboard, err := s.repo.PersonalDashboard(ctx, workspaceID, actorID, date)
	return dashboard, analyticsError(err)
}

func (s *AnalyticsService) WorkspaceDashboard(ctx context.Context, actorID, workspaceID string, date time.Time) (entity.WorkspaceDashboard, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.WorkspaceDashboard{}, err
	}
	dashboard, err := s.repo.WorkspaceDashboard(ctx, workspaceID, date)
	return dashboard, analyticsError(err)
}

func (s *AnalyticsService) TimeAllocationReport(ctx context.Context, actorID string, filter ReportFilter) (entity.TimeAllocationReport, error) {
	if err := s.validateReport(ctx, actorID, filter.WorkspaceID, filter.StartDate, filter.EndDate); err != nil {
		return entity.TimeAllocationReport{}, err
	}
	if filter.GroupBy == "" {
		filter.GroupBy = "category"
	}
	report, err := s.repo.TimeAllocationReport(ctx, analyticsrepo.ReportFilter(filter))
	return report, analyticsError(err)
}

func (s *AnalyticsService) ProductivityTrendReport(ctx context.Context, actorID string, filter ReportFilter) (entity.ProductivityTrendReport, error) {
	if err := s.validateReport(ctx, actorID, filter.WorkspaceID, filter.StartDate, filter.EndDate); err != nil {
		return entity.ProductivityTrendReport{}, err
	}
	if filter.Interval == "" {
		filter.Interval = "day"
	}
	report, err := s.repo.ProductivityTrendReport(ctx, analyticsrepo.ReportFilter(filter))
	return report, analyticsError(err)
}

func (s *AnalyticsService) TeamWorkloadReport(ctx context.Context, actorID string, filter ReportFilter) (entity.TeamWorkloadReport, error) {
	if err := s.validateReport(ctx, actorID, filter.WorkspaceID, filter.StartDate, filter.EndDate); err != nil {
		return entity.TeamWorkloadReport{}, err
	}
	report, err := s.repo.TeamWorkloadReport(ctx, analyticsrepo.ReportFilter(filter))
	return report, analyticsError(err)
}

func (s *AnalyticsService) Search(ctx context.Context, actorID, workspaceID, query string, types []string, limit int) (entity.SearchResults, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.SearchResults{}, err
	}
	if query == "" {
		return entity.SearchResults{}, ErrInvalidAnalyticsFilter
	}
	if limit < 1 {
		limit = 10
	}
	results, err := s.repo.Search(ctx, workspaceID, query, types, limit)
	return results, analyticsError(err)
}

func (s *AnalyticsService) ListSavedViews(ctx context.Context, actorID, workspaceID, resourceType string) ([]entity.SavedView, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	if resourceType == "" {
		return nil, ErrInvalidAnalyticsFilter
	}
	views, err := s.repo.ListSavedViews(ctx, workspaceID, actorID, resourceType)
	return views, analyticsError(err)
}

func (s *AnalyticsService) CreateSavedView(ctx context.Context, actorID string, view entity.SavedView) (entity.SavedView, error) {
	if _, err := s.requireWorkspaceMember(ctx, view.WorkspaceID, actorID); err != nil {
		return entity.SavedView{}, err
	}
	view.UserID = actorID
	created, err := s.repo.CreateSavedView(ctx, view)
	return created, analyticsError(err)
}

func (s *AnalyticsService) UpdateSavedView(ctx context.Context, actorID string, patch entity.SavedView) (entity.SavedView, error) {
	current, err := s.repo.FindSavedView(ctx, patch.ID)
	if err != nil {
		return entity.SavedView{}, analyticsError(err)
	}
	if current.UserID != actorID {
		return entity.SavedView{}, ErrForbidden
	}
	if patch.Name != "" {
		current.Name = patch.Name
	}
	if patch.FilterJSON != nil {
		current.FilterJSON = patch.FilterJSON
	}
	if patch.SharedSet {
		current.Shared = patch.Shared
	}
	updated, err := s.repo.UpdateSavedView(ctx, current)
	return updated, analyticsError(err)
}

func (s *AnalyticsService) DeleteSavedView(ctx context.Context, actorID, id string) error {
	view, err := s.repo.FindSavedView(ctx, id)
	if err != nil {
		return analyticsError(err)
	}
	if view.UserID != actorID {
		return ErrForbidden
	}
	return analyticsError(s.repo.DeleteSavedView(ctx, id))
}

func (s *AnalyticsService) ActivityLogs(ctx context.Context, actorID string, filter ActivityLogFilter) ([]entity.ActivityLog, int, error) {
	if filter.WorkspaceID == "" {
		return nil, 0, ErrInvalidAnalyticsFilter
	}
	if _, err := s.requireWorkspaceManager(ctx, filter.WorkspaceID, actorID); err != nil {
		return nil, 0, err
	}
	logs, total, err := s.repo.ListActivityLogs(ctx, analyticsrepo.ActivityLogFilter(filter))
	return logs, total, analyticsError(err)
}

func (s *AnalyticsService) validateReport(ctx context.Context, actorID, workspaceID string, startDate, endDate time.Time) error {
	if !endDate.Before(startDate) {
		_, err := s.requireWorkspaceMember(ctx, workspaceID, actorID)
		return err
	}
	return ErrInvalidAnalyticsRange
}

func (s *AnalyticsService) requireWorkspaceManager(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, err := s.requireWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	if member.Role != WorkspaceRoleOwner && member.Role != WorkspaceRoleAdmin {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	return member, nil
}

func (s *AnalyticsService) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	if workspaceID == "" {
		return entity.WorkspaceMember{}, ErrInvalidAnalyticsFilter
	}
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return entity.WorkspaceMember{}, ErrForbidden
		}
		return entity.WorkspaceMember{}, err
	}
	if member.Status != WorkspaceMemberActive {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	return member, nil
}

func analyticsError(err error) error {
	if errors.Is(err, analyticsrepo.ErrNotFound) {
		return ErrAnalyticsNotFound
	}
	return err
}
