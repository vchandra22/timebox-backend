package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAnalyticsValidateReportRejectsInvalidRange(t *testing.T) {
	service := newAnalyticsService(nil, nil)

	err := service.validateReport(context.Background(), "user-id", "workspace-id", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC))

	if !errors.Is(err, ErrInvalidAnalyticsRange) {
		t.Fatalf("expected ErrInvalidAnalyticsRange, got %v", err)
	}
}

func TestAnalyticsRejectsMissingWorkspace(t *testing.T) {
	service := newAnalyticsService(nil, nil)

	_, err := service.Streak(context.Background(), "user-id", "")

	if !errors.Is(err, ErrInvalidAnalyticsFilter) {
		t.Fatalf("expected ErrInvalidAnalyticsFilter, got %v", err)
	}
}
