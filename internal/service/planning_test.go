package service

import (
	"context"
	"errors"
	"testing"

	"timebox-backend/internal/entity"
	planningrepo "timebox-backend/internal/repository/planning"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

func TestPlanningTaskPermission(t *testing.T) {
	planningRepo := &planningRepoStub{
		task: entity.Task{
			ID:          "task-1",
			WorkspaceID: "ws-1",
			CreatedBy:   "owner-1",
			Title:       "Task",
			Status:      TaskStatusBacklog,
			Priority:    TaskPriorityMedium,
		},
	}
	workspaceRepo := &planningWorkspaceRepoStub{members: map[string]entity.WorkspaceMember{
		"ws-1:viewer-1": {WorkspaceID: "ws-1", UserID: "viewer-1", Role: WorkspaceRoleViewer, Status: WorkspaceMemberActive},
	}}
	svc := newPlanningService(planningRepo, workspaceRepo)

	_, err := svc.UpdateTask(context.Background(), "viewer-1", entity.Task{ID: "task-1", Status: TaskStatusDone})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("UpdateTask err = %v, want ErrForbidden", err)
	}
	if _, err := svc.CreateTask(context.Background(), "viewer-1", "ws-1", entity.Task{Title: "Task", Priority: "later"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("CreateTask err = %v, want ErrForbidden before priority validation", err)
	}
}

type planningRepoStub struct {
	task entity.Task
}

func (r *planningRepoStub) ListCategories(context.Context, string) ([]entity.Category, error) {
	return nil, nil
}
func (r *planningRepoStub) CreateCategory(context.Context, entity.Category) (entity.Category, error) {
	return entity.Category{}, nil
}
func (r *planningRepoStub) FindCategory(context.Context, string) (entity.Category, error) {
	return entity.Category{}, nil
}
func (r *planningRepoStub) UpdateCategory(context.Context, entity.Category) (entity.Category, error) {
	return entity.Category{}, nil
}
func (r *planningRepoStub) DeleteCategory(context.Context, string) error { return nil }
func (r *planningRepoStub) ListTags(context.Context, string, string) ([]entity.Tag, error) {
	return nil, nil
}
func (r *planningRepoStub) CreateTag(context.Context, entity.Tag) (entity.Tag, error) {
	return entity.Tag{}, nil
}
func (r *planningRepoStub) FindTag(context.Context, string) (entity.Tag, error) {
	return entity.Tag{}, nil
}
func (r *planningRepoStub) UpdateTag(context.Context, entity.Tag) (entity.Tag, error) {
	return entity.Tag{}, nil
}
func (r *planningRepoStub) DeleteTag(context.Context, string) error { return nil }
func (r *planningRepoStub) ListGoals(context.Context, planningrepo.GoalFilter) ([]entity.Goal, int, error) {
	return nil, 0, nil
}
func (r *planningRepoStub) CreateGoal(context.Context, entity.Goal) (entity.Goal, error) {
	return entity.Goal{}, nil
}
func (r *planningRepoStub) FindGoal(context.Context, string) (entity.Goal, error) {
	return entity.Goal{}, nil
}
func (r *planningRepoStub) UpdateGoal(context.Context, entity.Goal) (entity.Goal, error) {
	return entity.Goal{}, nil
}
func (r *planningRepoStub) ArchiveGoal(context.Context, string) error { return nil }
func (r *planningRepoStub) ListTasks(context.Context, planningrepo.TaskFilter) ([]entity.Task, int, error) {
	return nil, 0, nil
}
func (r *planningRepoStub) CreateTask(context.Context, entity.Task) (entity.Task, error) {
	return entity.Task{}, nil
}
func (r *planningRepoStub) FindTask(context.Context, string) (entity.Task, error) {
	return r.task, nil
}
func (r *planningRepoStub) UpdateTask(context.Context, entity.Task) (entity.Task, error) {
	return entity.Task{}, nil
}
func (r *planningRepoStub) DeleteTask(context.Context, string) error { return nil }
func (r *planningRepoStub) MoveTask(context.Context, string, string, int) (entity.TaskMove, error) {
	return entity.TaskMove{}, nil
}

type planningWorkspaceRepoStub struct {
	members map[string]entity.WorkspaceMember
}

func (r *planningWorkspaceRepoStub) Create(context.Context, entity.Workspace) (entity.Workspace, error) {
	return entity.Workspace{}, nil
}
func (r *planningWorkspaceRepoStub) List(context.Context, workspacerepo.ListFilter) ([]entity.Workspace, int, error) {
	return nil, 0, nil
}
func (r *planningWorkspaceRepoStub) FindByID(context.Context, string) (entity.Workspace, error) {
	return entity.Workspace{}, nil
}
func (r *planningWorkspaceRepoStub) Update(context.Context, entity.Workspace) (entity.Workspace, error) {
	return entity.Workspace{}, nil
}
func (r *planningWorkspaceRepoStub) FindMember(_ context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, ok := r.members[workspaceID+":"+userID]
	if !ok {
		return entity.WorkspaceMember{}, workspacerepo.ErrNotFound
	}
	return member, nil
}
func (r *planningWorkspaceRepoStub) ListMembers(context.Context, workspacerepo.ListFilter) ([]entity.WorkspaceMember, int, error) {
	return nil, 0, nil
}
func (r *planningWorkspaceRepoStub) InviteMember(context.Context, entity.WorkspaceInvitation, []string) (entity.WorkspaceInvitation, error) {
	return entity.WorkspaceInvitation{}, nil
}
func (r *planningWorkspaceRepoStub) UpdateMember(context.Context, entity.WorkspaceMember) (entity.WorkspaceMember, error) {
	return entity.WorkspaceMember{}, nil
}
func (r *planningWorkspaceRepoStub) ListTeams(context.Context, string) ([]entity.Team, error) {
	return nil, nil
}
func (r *planningWorkspaceRepoStub) CreateTeam(context.Context, entity.Team) (entity.Team, error) {
	return entity.Team{}, nil
}
func (r *planningWorkspaceRepoStub) FindTeam(context.Context, string) (entity.Team, error) {
	return entity.Team{}, nil
}
func (r *planningWorkspaceRepoStub) UpdateTeam(context.Context, entity.Team) (entity.Team, error) {
	return entity.Team{}, nil
}
func (r *planningWorkspaceRepoStub) DeleteTeam(context.Context, string) error { return nil }
