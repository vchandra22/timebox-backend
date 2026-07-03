package service

import (
	"context"
	"errors"

	"timebox-backend/internal/entity"
	planningrepo "timebox-backend/internal/repository/planning"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

const (
	GoalStatusActive    = "active"
	GoalStatusArchived  = "archived"
	GoalStatusCompleted = "completed"

	TaskStatusBacklog    = "backlog"
	TaskStatusScheduled  = "scheduled"
	TaskStatusInProgress = "in_progress"
	TaskStatusDone       = "done"
	TaskStatusCancelled  = "cancelled"

	TaskPriorityLow    = "low"
	TaskPriorityMedium = "medium"
	TaskPriorityHigh   = "high"
	TaskPriorityUrgent = "urgent"
)

var (
	ErrPlanningNotFound  = errors.New("planning resource not found")
	ErrPlanningConflict  = errors.New("planning resource conflict")
	ErrInvalidGoalStatus = errors.New("invalid goal status")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrInvalidPriority   = errors.New("invalid priority")
)

type PlanningService struct {
	repo          planningrepo.Repository
	workspaceRepo workspacerepo.Repository
}

type GoalListFilter struct {
	WorkspaceID string
	Q           string
	Status      string
	PinnedSet   bool
	Pinned      bool
	Limit       int
	Offset      int
}

type TaskListFilter struct {
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

func newPlanningService(repo planningrepo.Repository, workspaceRepo workspacerepo.Repository) *PlanningService {
	return &PlanningService{repo: repo, workspaceRepo: workspaceRepo}
}

func (s *PlanningService) ListCategories(ctx context.Context, actorID, workspaceID string) ([]entity.Category, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	categories, err := s.repo.ListCategories(ctx, workspaceID)
	return categories, planningError(err)
}

func (s *PlanningService) CreateCategory(ctx context.Context, actorID, workspaceID string, category entity.Category) (entity.Category, error) {
	if _, err := s.requireWorkspaceWriter(ctx, workspaceID, actorID); err != nil {
		return entity.Category{}, err
	}
	category.WorkspaceID = workspaceID
	created, err := s.repo.CreateCategory(ctx, category)
	return created, planningError(err)
}

func (s *PlanningService) UpdateCategory(ctx context.Context, actorID string, patch entity.Category) (entity.Category, error) {
	current, err := s.repo.FindCategory(ctx, patch.ID)
	if err != nil {
		return entity.Category{}, planningError(err)
	}
	if _, err := s.requireWorkspaceWriter(ctx, current.WorkspaceID, actorID); err != nil {
		return entity.Category{}, err
	}
	if patch.Name != "" {
		current.Name = patch.Name
	}
	if patch.Color != "" {
		current.Color = patch.Color
	}
	updated, err := s.repo.UpdateCategory(ctx, current)
	return updated, planningError(err)
}

func (s *PlanningService) DeleteCategory(ctx context.Context, actorID, id string) error {
	category, err := s.repo.FindCategory(ctx, id)
	if err != nil {
		return planningError(err)
	}
	if _, err := s.requireWorkspaceManager(ctx, category.WorkspaceID, actorID); err != nil {
		return err
	}
	return planningError(s.repo.DeleteCategory(ctx, id))
}

func (s *PlanningService) ListTags(ctx context.Context, actorID, workspaceID, q string) ([]entity.Tag, error) {
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	tags, err := s.repo.ListTags(ctx, workspaceID, q)
	return tags, planningError(err)
}

func (s *PlanningService) CreateTag(ctx context.Context, actorID string, tag entity.Tag) (entity.Tag, error) {
	if _, err := s.requireWorkspaceWriter(ctx, tag.WorkspaceID, actorID); err != nil {
		return entity.Tag{}, err
	}
	created, err := s.repo.CreateTag(ctx, tag)
	return created, planningError(err)
}

func (s *PlanningService) UpdateTag(ctx context.Context, actorID string, patch entity.Tag) (entity.Tag, error) {
	current, err := s.repo.FindTag(ctx, patch.ID)
	if err != nil {
		return entity.Tag{}, planningError(err)
	}
	if _, err := s.requireWorkspaceWriter(ctx, current.WorkspaceID, actorID); err != nil {
		return entity.Tag{}, err
	}
	if patch.Name != "" {
		current.Name = patch.Name
	}
	updated, err := s.repo.UpdateTag(ctx, current)
	return updated, planningError(err)
}

func (s *PlanningService) DeleteTag(ctx context.Context, actorID, id string) error {
	tag, err := s.repo.FindTag(ctx, id)
	if err != nil {
		return planningError(err)
	}
	if _, err := s.requireWorkspaceWriter(ctx, tag.WorkspaceID, actorID); err != nil {
		return err
	}
	return planningError(s.repo.DeleteTag(ctx, id))
}

func (s *PlanningService) ListGoals(ctx context.Context, actorID string, filter GoalListFilter) ([]entity.Goal, int, error) {
	if _, err := s.requireWorkspaceMember(ctx, filter.WorkspaceID, actorID); err != nil {
		return nil, 0, err
	}
	goals, total, err := s.repo.ListGoals(ctx, planningrepo.GoalFilter(filter))
	return goals, total, planningError(err)
}

func (s *PlanningService) CreateGoal(ctx context.Context, actorID, workspaceID string, goal entity.Goal) (entity.Goal, error) {
	if _, err := s.requireWorkspaceWriter(ctx, workspaceID, actorID); err != nil {
		return entity.Goal{}, err
	}
	goal.WorkspaceID = workspaceID
	goal.CreatedBy = actorID
	created, err := s.repo.CreateGoal(ctx, goal)
	return created, planningError(err)
}

func (s *PlanningService) FindGoal(ctx context.Context, actorID, id string) (entity.Goal, error) {
	goal, err := s.repo.FindGoal(ctx, id)
	if err != nil {
		return entity.Goal{}, planningError(err)
	}
	if _, err := s.requireWorkspaceMember(ctx, goal.WorkspaceID, actorID); err != nil {
		return entity.Goal{}, err
	}
	return goal, nil
}

func (s *PlanningService) UpdateGoal(ctx context.Context, actorID string, patch entity.Goal) (entity.Goal, error) {
	current, err := s.repo.FindGoal(ctx, patch.ID)
	if err != nil {
		return entity.Goal{}, planningError(err)
	}
	if err := s.requireGoalOwner(ctx, current, actorID); err != nil {
		return entity.Goal{}, err
	}
	if patch.Title != "" {
		current.Title = patch.Title
	}
	if patch.Description != "" {
		current.Description = patch.Description
	}
	if patch.TargetDate != nil {
		current.TargetDate = patch.TargetDate
	}
	if patch.Status != "" {
		if !validGoalStatus(patch.Status) {
			return entity.Goal{}, ErrInvalidGoalStatus
		}
		current.Status = patch.Status
	}
	if patch.IsPinnedSet {
		current.IsPinned = patch.IsPinned
	}
	updated, err := s.repo.UpdateGoal(ctx, current)
	return updated, planningError(err)
}

func (s *PlanningService) ArchiveGoal(ctx context.Context, actorID, id string) error {
	goal, err := s.repo.FindGoal(ctx, id)
	if err != nil {
		return planningError(err)
	}
	if err := s.requireGoalOwner(ctx, goal, actorID); err != nil {
		return err
	}
	return planningError(s.repo.ArchiveGoal(ctx, id))
}

func (s *PlanningService) ListTasks(ctx context.Context, actorID string, filter TaskListFilter) ([]entity.Task, int, error) {
	if _, err := s.requireWorkspaceMember(ctx, filter.WorkspaceID, actorID); err != nil {
		return nil, 0, err
	}
	tasks, total, err := s.repo.ListTasks(ctx, planningrepo.TaskFilter(filter))
	return tasks, total, planningError(err)
}

func (s *PlanningService) CreateTask(ctx context.Context, actorID, workspaceID string, task entity.Task) (entity.Task, error) {
	if _, err := s.requireWorkspaceWriter(ctx, workspaceID, actorID); err != nil {
		return entity.Task{}, err
	}
	if task.Priority == "" {
		task.Priority = TaskPriorityMedium
	}
	if !validPriority(task.Priority) {
		return entity.Task{}, ErrInvalidPriority
	}
	task.WorkspaceID = workspaceID
	task.CreatedBy = actorID
	created, err := s.repo.CreateTask(ctx, task)
	return created, planningError(err)
}

func (s *PlanningService) FindTask(ctx context.Context, actorID, id string) (entity.Task, error) {
	task, err := s.repo.FindTask(ctx, id)
	if err != nil {
		return entity.Task{}, planningError(err)
	}
	if _, err := s.requireWorkspaceMember(ctx, task.WorkspaceID, actorID); err != nil {
		return entity.Task{}, err
	}
	return task, nil
}

func (s *PlanningService) UpdateTask(ctx context.Context, actorID string, patch entity.Task) (entity.Task, error) {
	current, err := s.repo.FindTask(ctx, patch.ID)
	if err != nil {
		return entity.Task{}, planningError(err)
	}
	if err := s.requireTaskOwner(ctx, current, actorID); err != nil {
		return entity.Task{}, err
	}
	applyTaskPatch(&current, patch)
	if !validTaskStatus(current.Status) {
		return entity.Task{}, ErrInvalidTaskStatus
	}
	if !validPriority(current.Priority) {
		return entity.Task{}, ErrInvalidPriority
	}
	updated, err := s.repo.UpdateTask(ctx, current)
	return updated, planningError(err)
}

func (s *PlanningService) DeleteTask(ctx context.Context, actorID, id string) error {
	task, err := s.repo.FindTask(ctx, id)
	if err != nil {
		return planningError(err)
	}
	if err := s.requireTaskOwner(ctx, task, actorID); err != nil {
		return err
	}
	return planningError(s.repo.DeleteTask(ctx, id))
}

func (s *PlanningService) MoveTask(ctx context.Context, actorID, id, toStatus string, position int) (entity.TaskMove, error) {
	if !validTaskStatus(toStatus) {
		return entity.TaskMove{}, ErrInvalidTaskStatus
	}
	task, err := s.repo.FindTask(ctx, id)
	if err != nil {
		return entity.TaskMove{}, planningError(err)
	}
	if err := s.requireTaskOwner(ctx, task, actorID); err != nil {
		return entity.TaskMove{}, err
	}
	moved, err := s.repo.MoveTask(ctx, id, toStatus, position)
	return moved, planningError(err)
}

func applyTaskPatch(current *entity.Task, patch entity.Task) {
	if patch.GoalID != nil {
		current.GoalID = patch.GoalID
	}
	if patch.AssigneeID != nil {
		current.AssigneeID = patch.AssigneeID
	}
	if patch.CategoryID != nil {
		current.CategoryID = patch.CategoryID
	}
	if patch.Title != "" {
		current.Title = patch.Title
	}
	if patch.Description != "" {
		current.Description = patch.Description
	}
	if patch.Status != "" {
		current.Status = patch.Status
	}
	if patch.Priority != "" {
		current.Priority = patch.Priority
	}
	if patch.EstimatedMinutes != nil {
		current.EstimatedMinutes = patch.EstimatedMinutes
	}
	if patch.TagIDsSet {
		current.TagIDs = patch.TagIDs
		current.TagIDsSet = true
	}
}

func (s *PlanningService) requireGoalOwner(ctx context.Context, goal entity.Goal, actorID string) error {
	member, err := s.requireWorkspaceMember(ctx, goal.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if member.Role == WorkspaceRoleOwner || member.Role == WorkspaceRoleAdmin || goal.CreatedBy == actorID {
		return nil
	}
	return ErrForbidden
}

func (s *PlanningService) requireTaskOwner(ctx context.Context, task entity.Task, actorID string) error {
	member, err := s.requireWorkspaceMember(ctx, task.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if member.Role == WorkspaceRoleOwner || member.Role == WorkspaceRoleAdmin || task.CreatedBy == actorID || (task.AssigneeID != nil && *task.AssigneeID == actorID) {
		return nil
	}
	return ErrForbidden
}

func (s *PlanningService) requireWorkspaceWriter(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, err := s.requireWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	if member.Role == WorkspaceRoleViewer {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	return member, nil
}

func (s *PlanningService) requireWorkspaceManager(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, err := s.requireWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	if member.Role != WorkspaceRoleOwner && member.Role != WorkspaceRoleAdmin {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	return member, nil
}

func (s *PlanningService) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
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

func validGoalStatus(status string) bool {
	return status == GoalStatusActive || status == GoalStatusArchived || status == GoalStatusCompleted
}

func validTaskStatus(status string) bool {
	return status == TaskStatusBacklog || status == TaskStatusScheduled || status == TaskStatusInProgress || status == TaskStatusDone || status == TaskStatusCancelled
}

func validPriority(priority string) bool {
	return priority == TaskPriorityLow || priority == TaskPriorityMedium || priority == TaskPriorityHigh || priority == TaskPriorityUrgent
}

func planningError(err error) error {
	if errors.Is(err, planningrepo.ErrNotFound) {
		return ErrPlanningNotFound
	}
	if errors.Is(err, planningrepo.ErrConflict) {
		return ErrPlanningConflict
	}
	return err
}
