package service

import (
	"context"
	"errors"
	"time"

	"timebox-backend/internal/entity"
	executionrepo "timebox-backend/internal/repository/execution"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

const (
	TimeboxStatusPlanned   = "planned"
	TimeboxStatusRunning   = "running"
	TimeboxStatusPaused    = "paused"
	TimeboxStatusCompleted = "completed"
	TimeboxStatusSkipped   = "skipped"
	TimeboxStatusCancelled = "cancelled"

	TimerStatusRunning = "running"
	TimerStatusPaused  = "paused"

	TimeLogSourceTimer  = "timer"
	TimeLogSourceManual = "manual"
)

var (
	ErrExecutionNotFound    = errors.New("execution resource not found")
	ErrTimerAlreadyRunning  = errors.New("timer already running")
	ErrTimerNotRunning      = errors.New("timebox timer not running")
	ErrInvalidTimeRange     = errors.New("invalid time range")
	ErrInvalidTimeboxStatus = errors.New("invalid timebox status")
	ErrInvalidTimeLogSource = errors.New("invalid time log source")
)

type ExecutionService struct {
	repo          executionrepo.Repository
	timerRepo     executionrepo.TimerRepository
	workspaceRepo workspacerepo.Repository
}

type TimeboxListFilter struct {
	WorkspaceID string
	Date        *time.Time
	StartDate   *time.Time
	EndDate     *time.Time
	OwnerID     string
	Status      string
	CategoryID  string
	GoalID      string
}

type TimeboxPatch struct {
	Timebox
	TaskIDSet         bool
	OwnerIDSet        bool
	CategoryIDSet     bool
	TitleSet          bool
	DescriptionSet    bool
	ScheduledStartSet bool
	ScheduledEndSet   bool
	StatusSet         bool
	IsBufferSet       bool
}

type Timebox = entity.Timebox

func newExecutionService(repo executionrepo.Repository, timerRepo executionrepo.TimerRepository, workspaceRepo workspacerepo.Repository) *ExecutionService {
	return &ExecutionService{repo: repo, timerRepo: timerRepo, workspaceRepo: workspaceRepo}
}

func (s *ExecutionService) ListTimeboxes(ctx context.Context, actorID string, filter TimeboxListFilter) ([]entity.Timebox, error) {
	if _, err := s.requireWorkspaceMember(ctx, filter.WorkspaceID, actorID); err != nil {
		return nil, err
	}
	timeboxes, err := s.repo.ListTimeboxes(ctx, executionrepo.TimeboxFilter(filter))
	return timeboxes, executionError(err)
}

func (s *ExecutionService) CreateTimebox(ctx context.Context, actorID string, timebox entity.Timebox) (entity.Timebox, error) {
	if err := validRange(timebox.ScheduledStart, timebox.ScheduledEnd); err != nil {
		return entity.Timebox{}, err
	}
	if _, err := s.requireWorkspaceWriter(ctx, timebox.WorkspaceID, actorID); err != nil {
		return entity.Timebox{}, err
	}
	created, err := s.repo.CreateTimebox(ctx, timebox)
	return created, executionError(err)
}

func (s *ExecutionService) FindTimebox(ctx context.Context, actorID, id string) (entity.Timebox, error) {
	timebox, err := s.repo.FindTimebox(ctx, id)
	if err != nil {
		return entity.Timebox{}, executionError(err)
	}
	if _, err := s.requireWorkspaceMember(ctx, timebox.WorkspaceID, actorID); err != nil {
		return entity.Timebox{}, err
	}
	return timebox, nil
}

func (s *ExecutionService) UpdateTimebox(ctx context.Context, actorID string, patch TimeboxPatch) (entity.Timebox, error) {
	current, err := s.repo.FindTimebox(ctx, patch.ID)
	if err != nil {
		return entity.Timebox{}, executionError(err)
	}
	if err := s.requireTimeboxOwner(ctx, current, actorID); err != nil {
		return entity.Timebox{}, err
	}
	applyTimeboxPatch(&current, patch)
	if err := validRange(current.ScheduledStart, current.ScheduledEnd); err != nil {
		return entity.Timebox{}, err
	}
	if !validTimeboxStatus(current.Status) {
		return entity.Timebox{}, ErrInvalidTimeboxStatus
	}
	updated, err := s.repo.UpdateTimebox(ctx, current)
	return updated, executionError(err)
}

func (s *ExecutionService) DeleteTimebox(ctx context.Context, actorID, id string) error {
	timebox, err := s.repo.FindTimebox(ctx, id)
	if err != nil {
		return executionError(err)
	}
	if err := s.requireTimeboxOwner(ctx, timebox, actorID); err != nil {
		return err
	}
	return executionError(s.repo.DeleteTimebox(ctx, id))
}

func (s *ExecutionService) ActiveTimer(ctx context.Context, actorID string) (entity.TimerState, error) {
	state, err := s.timerRepo.Get(ctx, actorID)
	if err != nil {
		return entity.TimerState{}, executionError(err)
	}
	timebox, err := s.repo.FindTimebox(ctx, state.TimeboxID)
	if err == nil {
		state.Timebox = &timebox
	}
	return s.timerWithClock(state, time.Now()), nil
}

func (s *ExecutionService) StartTimer(ctx context.Context, actorID, timeboxID string) (entity.TimerState, error) {
	if existing, err := s.timerRepo.Get(ctx, actorID); err == nil && existing.TimeboxID != "" {
		return entity.TimerState{}, ErrTimerAlreadyRunning
	}
	timebox, err := s.repo.FindTimebox(ctx, timeboxID)
	if err != nil {
		return entity.TimerState{}, executionError(err)
	}
	if err := s.requireTimeboxOwner(ctx, timebox, actorID); err != nil {
		return entity.TimerState{}, err
	}
	now := time.Now()
	if _, err := s.repo.CreateTimeLog(ctx, entity.TimeLog{TimeboxID: timebox.ID, StartedAt: now, Source: TimeLogSourceTimer, CreatedBy: actorID}); err != nil {
		return entity.TimerState{}, executionError(err)
	}
	updated, err := s.repo.UpdateTimeboxStatus(ctx, timebox.ID, TimeboxStatusRunning, timebox.ActualMinutes)
	if err != nil {
		return entity.TimerState{}, executionError(err)
	}
	state := entity.TimerState{TimeboxID: timebox.ID, UserID: actorID, Status: TimerStatusRunning, StartedAt: now, PlannedMinutes: updated.PlannedMinutes, Timebox: &updated}
	state = s.timerWithClock(state, now)
	return state, s.timerRepo.Save(ctx, state)
}

func (s *ExecutionService) PauseTimer(ctx context.Context, actorID, timeboxID string, reason string) (entity.TimerState, error) {
	state, err := s.timerForTimebox(ctx, actorID, timeboxID)
	if err != nil {
		return entity.TimerState{}, err
	}
	if state.Status != TimerStatusRunning {
		return entity.TimerState{}, ErrTimerNotRunning
	}
	now := time.Now()
	var note *string
	if reason != "" {
		note = &reason
	}
	if err := s.repo.CloseRunningLog(ctx, timeboxID, now, note); err != nil {
		return entity.TimerState{}, executionError(err)
	}
	seconds, err := s.repo.SumTimeLogSeconds(ctx, timeboxID)
	if err != nil {
		return entity.TimerState{}, executionError(err)
	}
	state.Status = TimerStatusPaused
	state.PausedAt = &now
	state.ElapsedSeconds = seconds
	state = s.timerWithClock(state, now)
	if _, err := s.repo.UpdateTimeboxStatus(ctx, timeboxID, TimeboxStatusPaused, seconds/60); err != nil {
		return entity.TimerState{}, executionError(err)
	}
	return state, s.timerRepo.Save(ctx, state)
}

func (s *ExecutionService) ResumeTimer(ctx context.Context, actorID, timeboxID string) (entity.TimerState, error) {
	state, err := s.timerForTimebox(ctx, actorID, timeboxID)
	if err != nil {
		return entity.TimerState{}, err
	}
	if state.Status != TimerStatusPaused {
		return entity.TimerState{}, ErrTimerNotRunning
	}
	now := time.Now()
	if _, err := s.repo.CreateTimeLog(ctx, entity.TimeLog{TimeboxID: timeboxID, StartedAt: now, Source: TimeLogSourceTimer, CreatedBy: actorID}); err != nil {
		return entity.TimerState{}, executionError(err)
	}
	state.Status = TimerStatusRunning
	state.PausedAt = nil
	state = s.timerWithClock(state, now)
	if _, err := s.repo.UpdateTimeboxStatus(ctx, timeboxID, TimeboxStatusRunning, state.ElapsedSeconds/60); err != nil {
		return entity.TimerState{}, executionError(err)
	}
	return state, s.timerRepo.Save(ctx, state)
}

func (s *ExecutionService) CompleteTimer(ctx context.Context, actorID, timeboxID string, note string) (entity.TimerCompletion, error) {
	timebox, err := s.repo.FindTimebox(ctx, timeboxID)
	if err != nil {
		return entity.TimerCompletion{}, executionError(err)
	}
	if err := s.requireTimeboxOwner(ctx, timebox, actorID); err != nil {
		return entity.TimerCompletion{}, err
	}
	now := time.Now()
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	_ = s.repo.CloseRunningLog(ctx, timeboxID, now, notePtr)
	seconds, err := s.repo.SumTimeLogSeconds(ctx, timeboxID)
	if err != nil {
		return entity.TimerCompletion{}, executionError(err)
	}
	actualMinutes := seconds / 60
	updated, err := s.repo.UpdateTimeboxStatus(ctx, timeboxID, TimeboxStatusCompleted, actualMinutes)
	if err != nil {
		return entity.TimerCompletion{}, executionError(err)
	}
	_ = s.timerRepo.Delete(ctx, actorID)
	return entity.TimerCompletion{TimeboxID: timeboxID, Status: updated.Status, CompletedAt: now, PlannedMinutes: updated.PlannedMinutes, ActualMinutes: updated.ActualMinutes, VarianceMinutes: updated.ActualMinutes - updated.PlannedMinutes, StreakUpdated: false}, nil
}

func (s *ExecutionService) SkipTimer(ctx context.Context, actorID, timeboxID string, reason string) (entity.TimerCompletion, error) {
	timebox, err := s.repo.FindTimebox(ctx, timeboxID)
	if err != nil {
		return entity.TimerCompletion{}, executionError(err)
	}
	if err := s.requireTimeboxOwner(ctx, timebox, actorID); err != nil {
		return entity.TimerCompletion{}, err
	}
	now := time.Now()
	var note *string
	if reason != "" {
		note = &reason
	}
	_ = s.repo.CloseRunningLog(ctx, timeboxID, now, note)
	updated, err := s.repo.UpdateTimeboxStatus(ctx, timeboxID, TimeboxStatusSkipped, timebox.ActualMinutes)
	if err != nil {
		return entity.TimerCompletion{}, executionError(err)
	}
	_ = s.timerRepo.Delete(ctx, actorID)
	return entity.TimerCompletion{TimeboxID: timeboxID, Status: updated.Status, SkippedAt: now}, nil
}

func (s *ExecutionService) ListTimeLogs(ctx context.Context, actorID, timeboxID string) ([]entity.TimeLog, error) {
	timebox, err := s.repo.FindTimebox(ctx, timeboxID)
	if err != nil {
		return nil, executionError(err)
	}
	if _, err := s.requireWorkspaceMember(ctx, timebox.WorkspaceID, actorID); err != nil {
		return nil, err
	}
	logs, err := s.repo.ListTimeLogs(ctx, timeboxID)
	return logs, executionError(err)
}

func (s *ExecutionService) CreateManualTimeLog(ctx context.Context, actorID, timeboxID string, log entity.TimeLog) (entity.TimeLog, error) {
	timebox, err := s.repo.FindTimebox(ctx, timeboxID)
	if err != nil {
		return entity.TimeLog{}, executionError(err)
	}
	if err := s.requireTimeboxOwner(ctx, timebox, actorID); err != nil {
		return entity.TimeLog{}, err
	}
	if err := validRange(log.StartedAt, *log.EndedAt); err != nil {
		return entity.TimeLog{}, err
	}
	if log.Source == "" {
		log.Source = TimeLogSourceManual
	}
	if log.Source != TimeLogSourceManual && log.Source != TimeLogSourceTimer {
		return entity.TimeLog{}, ErrInvalidTimeLogSource
	}
	log.TimeboxID = timeboxID
	log.CreatedBy = actorID
	created, err := s.repo.CreateTimeLog(ctx, log)
	if err != nil {
		return entity.TimeLog{}, executionError(err)
	}
	seconds, err := s.repo.SumTimeLogSeconds(ctx, timeboxID)
	if err == nil {
		_, _ = s.repo.UpdateTimeboxStatus(ctx, timeboxID, timebox.Status, seconds/60)
	}
	return created, nil
}

func (s *ExecutionService) timerForTimebox(ctx context.Context, actorID, timeboxID string) (entity.TimerState, error) {
	state, err := s.timerRepo.Get(ctx, actorID)
	if err != nil {
		return entity.TimerState{}, executionError(err)
	}
	if state.TimeboxID != timeboxID {
		return entity.TimerState{}, ErrTimerNotRunning
	}
	return state, nil
}

func (s *ExecutionService) timerWithClock(state entity.TimerState, now time.Time) entity.TimerState {
	state.ServerTime = now
	elapsed := state.ElapsedSeconds
	if state.Status == TimerStatusRunning {
		elapsed += int(now.Sub(state.StartedAt).Seconds())
	}
	state.ElapsedSeconds = elapsed
	state.RemainingSeconds = state.PlannedMinutes*60 - elapsed
	if state.RemainingSeconds < 0 {
		state.RemainingSeconds = 0
	}
	return state
}

func applyTimeboxPatch(current *entity.Timebox, patch TimeboxPatch) {
	if patch.TaskIDSet {
		current.TaskID = patch.TaskID
	}
	if patch.OwnerIDSet {
		current.OwnerID = patch.OwnerID
	}
	if patch.CategoryIDSet {
		current.CategoryID = patch.CategoryID
	}
	if patch.TitleSet {
		current.Title = patch.Title
	}
	if patch.DescriptionSet {
		current.Description = patch.Description
	}
	if patch.ScheduledStartSet {
		current.ScheduledStart = patch.ScheduledStart
	}
	if patch.ScheduledEndSet {
		current.ScheduledEnd = patch.ScheduledEnd
	}
	if patch.StatusSet {
		current.Status = patch.Status
	}
	if patch.IsBufferSet {
		current.IsBuffer = patch.IsBuffer
	}
}

func (s *ExecutionService) requireTimeboxOwner(ctx context.Context, timebox entity.Timebox, actorID string) error {
	member, err := s.requireWorkspaceMember(ctx, timebox.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if member.Role == WorkspaceRoleOwner || member.Role == WorkspaceRoleAdmin || timebox.OwnerID == actorID {
		return nil
	}
	return ErrForbidden
}

func (s *ExecutionService) requireWorkspaceWriter(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
	member, err := s.requireWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		return entity.WorkspaceMember{}, err
	}
	if member.Role == WorkspaceRoleViewer {
		return entity.WorkspaceMember{}, ErrForbidden
	}
	return member, nil
}

func (s *ExecutionService) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
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

func validRange(start, end time.Time) error {
	if !end.After(start) {
		return ErrInvalidTimeRange
	}
	return nil
}

func validTimeboxStatus(status string) bool {
	return status == TimeboxStatusPlanned || status == TimeboxStatusRunning || status == TimeboxStatusPaused || status == TimeboxStatusCompleted || status == TimeboxStatusSkipped || status == TimeboxStatusCancelled
}

func executionError(err error) error {
	if errors.Is(err, executionrepo.ErrNotFound) {
		return ErrExecutionNotFound
	}
	if errors.Is(err, executionrepo.ErrTimerNotFound) {
		return ErrTimerNotRunning
	}
	if errors.Is(err, executionrepo.ErrTimerAlreadyRunning) {
		return ErrTimerAlreadyRunning
	}
	return err
}
