package service

import (
	"testing"
	"time"

	"timebox-backend/internal/entity"
)

func TestExecutionTimerClock(t *testing.T) {
	startedAt := time.Date(2026, 7, 4, 9, 30, 0, 0, time.UTC)
	now := startedAt.Add(15 * time.Minute)
	svc := newExecutionService(nil, nil, nil)

	state := svc.timerWithClock(entity.TimerState{
		Status:         TimerStatusRunning,
		StartedAt:      startedAt,
		PlannedMinutes: 120,
	}, now)

	if state.ElapsedSeconds != 900 {
		t.Fatalf("ElapsedSeconds = %d, want 900", state.ElapsedSeconds)
	}
	if state.RemainingSeconds != 6300 {
		t.Fatalf("RemainingSeconds = %d, want 6300", state.RemainingSeconds)
	}
}
