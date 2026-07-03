package handler

import (
	"errors"
	"net/http"
	"time"

	executiondto "timebox-backend/internal/dto/execution"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ExecutionHandler struct {
	authService      *service.AuthService
	executionService *service.ExecutionService
}

func newExecutionHandler(services *service.Service) *ExecutionHandler {
	return &ExecutionHandler{authService: services.Auth, executionService: services.Execution}
}

func (h *ExecutionHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/timeboxes", h.ListTimeboxes)
	routeGroup.POST("/timeboxes", h.CreateTimebox)
	routeGroup.GET("/timeboxes/:id", h.GetTimebox)
	routeGroup.PATCH("/timeboxes/:id", h.UpdateTimebox)
	routeGroup.DELETE("/timeboxes/:id", h.DeleteTimebox)
	routeGroup.GET("/timer/active", h.ActiveTimer)
	routeGroup.POST("/timeboxes/:id/start", h.StartTimer)
	routeGroup.POST("/timeboxes/:id/pause", h.PauseTimer)
	routeGroup.POST("/timeboxes/:id/resume", h.ResumeTimer)
	routeGroup.POST("/timeboxes/:id/complete", h.CompleteTimer)
	routeGroup.POST("/timeboxes/:id/skip", h.SkipTimer)
	routeGroup.GET("/timeboxes/:id/logs", h.ListTimeLogs)
	routeGroup.POST("/timeboxes/:id/logs", h.CreateTimeLog)
}

func (h *ExecutionHandler) ListTimeboxes(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	date, ok := parseOptionalDate(ctx, ctx.Query("date"))
	if !ok {
		return
	}
	startDate, ok := parseOptionalDate(ctx, ctx.Query("start_date"))
	if !ok {
		return
	}
	endDate, ok := parseOptionalDate(ctx, ctx.Query("end_date"))
	if !ok {
		return
	}
	timeboxes, err := h.executionService.ListTimeboxes(ctx, userID, service.TimeboxListFilter{
		WorkspaceID: ctx.Query("workspace_id"),
		Date:        date,
		StartDate:   startDate,
		EndDate:     endDate,
		OwnerID:     ctx.Query("owner_id"),
		Status:      ctx.Query("status"),
		CategoryID:  ctx.Query("category_id"),
		GoalID:      ctx.Query("goal_id"),
	})
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, executionTimeboxListResponse(timeboxes), "data fetched", http.StatusOK)
}

func (h *ExecutionHandler) CreateTimebox(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.CreateTimeboxRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	start, end, ok := parseTimeRange(ctx, req.ScheduledStart, req.ScheduledEnd)
	if !ok {
		return
	}
	timebox, err := h.executionService.CreateTimebox(ctx, userID, entity.Timebox{
		WorkspaceID:    req.WorkspaceID,
		TaskID:         req.TaskID,
		OwnerID:        req.OwnerID,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Description:    req.Description,
		ScheduledStart: start,
		ScheduledEnd:   end,
		IsBuffer:       req.IsBuffer,
	})
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, executionTimeboxResponse(timebox, true, true), "Timebox created", http.StatusCreated)
}

func (h *ExecutionHandler) GetTimebox(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	timebox, err := h.executionService.FindTimebox(ctx, userID, ctx.Param("id"))
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, executionTimeboxResponse(timebox, true, true), "data fetched", http.StatusOK)
}

func (h *ExecutionHandler) UpdateTimebox(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.UpdateTimeboxRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := service.TimeboxPatch{Timebox: entity.Timebox{ID: ctx.Param("id")}}
	if req.TaskID != nil {
		patch.TaskID = req.TaskID
		patch.TaskIDSet = true
	}
	if req.OwnerID != nil {
		patch.OwnerID = *req.OwnerID
		patch.OwnerIDSet = true
	}
	if req.CategoryID != nil {
		patch.CategoryID = req.CategoryID
		patch.CategoryIDSet = true
	}
	if req.Title != nil {
		patch.Title = *req.Title
		patch.TitleSet = true
	}
	if req.Description != nil {
		patch.Description = *req.Description
		patch.DescriptionSet = true
	}
	if req.ScheduledStart != nil {
		start, ok := parseRFC3339(ctx, *req.ScheduledStart)
		if !ok {
			return
		}
		patch.ScheduledStart = start
		patch.ScheduledStartSet = true
	}
	if req.ScheduledEnd != nil {
		end, ok := parseRFC3339(ctx, *req.ScheduledEnd)
		if !ok {
			return
		}
		patch.ScheduledEnd = end
		patch.ScheduledEndSet = true
	}
	if req.Status != nil {
		patch.Status = *req.Status
		patch.StatusSet = true
	}
	if req.IsBuffer != nil {
		patch.IsBuffer = *req.IsBuffer
		patch.IsBufferSet = true
	}
	timebox, err := h.executionService.UpdateTimebox(ctx, userID, patch)
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, executionTimeboxResponse(timebox, false, true), "Timebox updated", http.StatusOK)
}

func (h *ExecutionHandler) DeleteTimebox(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.executionService.DeleteTimebox(ctx, userID, ctx.Param("id")); err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Timebox deleted", http.StatusOK)
}

func (h *ExecutionHandler) ActiveTimer(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	state, err := h.executionService.ActiveTimer(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrTimerNotRunning) {
			response.WithoutData(ctx, "No active timer", http.StatusOK)
			return
		}
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timerStateResponse(state), "data fetched", http.StatusOK)
}

func (h *ExecutionHandler) StartTimer(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.StartTimerRequest
	_ = ctx.ShouldBindJSON(&req)
	state, err := h.executionService.StartTimer(ctx, userID, ctx.Param("id"))
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timerStateResponse(state), "Timer started", http.StatusOK)
}

func (h *ExecutionHandler) PauseTimer(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.PauseTimerRequest
	_ = ctx.ShouldBindJSON(&req)
	state, err := h.executionService.PauseTimer(ctx, userID, ctx.Param("id"), req.Reason)
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timerStateResponse(state), "Timer paused", http.StatusOK)
}

func (h *ExecutionHandler) ResumeTimer(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	state, err := h.executionService.ResumeTimer(ctx, userID, ctx.Param("id"))
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	result := timerStateResponse(state)
	now := state.ServerTime
	result.ResumedAt = &now
	response.WithData(ctx, result, "Timer resumed", http.StatusOK)
}

func (h *ExecutionHandler) CompleteTimer(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.CompleteTimerRequest
	_ = ctx.ShouldBindJSON(&req)
	completion, err := h.executionService.CompleteTimer(ctx, userID, ctx.Param("id"), req.Note)
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timerCompletionResponse(completion), "Timebox completed", http.StatusOK)
}

func (h *ExecutionHandler) SkipTimer(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.SkipTimerRequest
	_ = ctx.ShouldBindJSON(&req)
	completion, err := h.executionService.SkipTimer(ctx, userID, ctx.Param("id"), req.Reason)
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timerCompletionResponse(completion), "Timebox skipped", http.StatusOK)
}

func (h *ExecutionHandler) ListTimeLogs(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	logs, err := h.executionService.ListTimeLogs(ctx, userID, ctx.Param("id"))
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timeLogListResponse(logs), "data fetched", http.StatusOK)
}

func (h *ExecutionHandler) CreateTimeLog(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req executiondto.CreateTimeLogRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	start, end, ok := parseTimeRange(ctx, req.StartedAt, req.EndedAt)
	if !ok {
		return
	}
	var note *string
	if req.Note != "" {
		note = &req.Note
	}
	log, err := h.executionService.CreateManualTimeLog(ctx, userID, ctx.Param("id"), entity.TimeLog{StartedAt: start, EndedAt: &end, Source: req.Source, Note: note})
	if err != nil {
		writeExecutionError(ctx, err)
		return
	}
	response.WithData(ctx, timeLogResponse(log), "Time log created", http.StatusCreated)
}

func (h *ExecutionHandler) currentUserID(ctx *gin.Context) (string, bool) {
	userID, err := h.authService.ValidateAccessToken(bearerToken(ctx))
	if err != nil {
		writeAuthError(ctx, err)
		return "", false
	}
	return userID, true
}

func writeExecutionError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrExecutionNotFound):
		response.Error(ctx, "not found", "resource not found", http.StatusNotFound)
	case errors.Is(err, service.ErrTimerAlreadyRunning):
		response.Error(ctx, "conflict", "timer already running", http.StatusConflict)
	case errors.Is(err, service.ErrTimerNotRunning):
		response.Error(ctx, "validation error", "timebox timer not running", http.StatusUnprocessableEntity)
	case errors.Is(err, service.ErrForbidden):
		response.Error(ctx, "forbidden", "forbidden", http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidTimeRange), errors.Is(err, service.ErrInvalidTimeboxStatus), errors.Is(err, service.ErrInvalidTimeLogSource):
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
	default:
		response.Error(ctx, "internal server error", "execution request failed", http.StatusInternalServerError)
	}
}

func parseOptionalDate(ctx *gin.Context, value string) (*time.Time, bool) {
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
		return nil, false
	}
	return &parsed, true
}

func parseTimeRange(ctx *gin.Context, startValue, endValue string) (time.Time, time.Time, bool) {
	start, ok := parseRFC3339(ctx, startValue)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	end, ok := parseRFC3339(ctx, endValue)
	return start, end, ok
}

func parseRFC3339(ctx *gin.Context, value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
		return time.Time{}, false
	}
	return parsed, true
}

func executionTimeboxListResponse(timeboxes []entity.Timebox) []executiondto.TimeboxResponse {
	result := make([]executiondto.TimeboxResponse, 0, len(timeboxes))
	for _, timebox := range timeboxes {
		result = append(result, executionTimeboxResponse(timebox, false, false))
	}
	return result
}

func executionTimeboxResponse(timebox entity.Timebox, includeCreated bool, includeUpdated bool) executiondto.TimeboxResponse {
	item := executiondto.TimeboxResponse{
		ID:             timebox.ID,
		WorkspaceID:    timebox.WorkspaceID,
		TaskID:         timebox.TaskID,
		OwnerID:        timebox.OwnerID,
		CategoryID:     timebox.CategoryID,
		Title:          timebox.Title,
		Description:    timebox.Description,
		ScheduledStart: timebox.ScheduledStart,
		ScheduledEnd:   timebox.ScheduledEnd,
		PlannedMinutes: timebox.PlannedMinutes,
		ActualMinutes:  timebox.ActualMinutes,
		Status:         timebox.Status,
		IsBuffer:       timebox.IsBuffer,
		Owner:          executionUserSummaryResponse(timebox.Owner),
		Category:       executionCategorySummaryResponse(timebox.Category),
		Task:           executionTaskSummaryResponse(timebox.Task),
		Warnings:       executionWarningResponse(timebox.Warnings),
	}
	if includeCreated {
		createdAt := timebox.CreatedAt
		item.CreatedAt = &createdAt
	}
	if includeUpdated {
		updatedAt := timebox.UpdatedAt
		item.UpdatedAt = &updatedAt
	}
	return item
}

func timerStateResponse(state entity.TimerState) executiondto.TimerStateResponse {
	return executiondto.TimerStateResponse{
		TimeboxID:        state.TimeboxID,
		Status:           state.Status,
		StartedAt:        state.StartedAt,
		PausedAt:         state.PausedAt,
		PlannedMinutes:   state.PlannedMinutes,
		ElapsedSeconds:   state.ElapsedSeconds,
		RemainingSeconds: state.RemainingSeconds,
		ServerTime:       state.ServerTime,
		Timebox:          executionOptionalTimeboxResponse(state.Timebox),
	}
}

func executionOptionalTimeboxResponse(timebox *entity.Timebox) *executiondto.TimeboxResponse {
	if timebox == nil {
		return nil
	}
	item := executionTimeboxResponse(*timebox, false, false)
	return &item
}

func timerCompletionResponse(completion entity.TimerCompletion) executiondto.TimerCompletionResponse {
	return executiondto.TimerCompletionResponse{
		TimeboxID:       completion.TimeboxID,
		Status:          completion.Status,
		CompletedAt:     completion.CompletedAt,
		SkippedAt:       completion.SkippedAt,
		PlannedMinutes:  completion.PlannedMinutes,
		ActualMinutes:   completion.ActualMinutes,
		VarianceMinutes: completion.VarianceMinutes,
		StreakUpdated:   completion.StreakUpdated,
	}
}

func timeLogListResponse(logs []entity.TimeLog) []executiondto.TimeLogResponse {
	result := make([]executiondto.TimeLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, timeLogResponse(log))
	}
	return result
}

func timeLogResponse(log entity.TimeLog) executiondto.TimeLogResponse {
	return executiondto.TimeLogResponse{ID: log.ID, TimeboxID: log.TimeboxID, StartedAt: log.StartedAt, EndedAt: log.EndedAt, DurationSeconds: log.DurationSeconds, Source: log.Source, Note: log.Note, CreatedBy: log.CreatedBy}
}

func executionUserSummaryResponse(user *entity.UserSummary) *executiondto.UserSummaryResponse {
	if user == nil {
		return nil
	}
	return &executiondto.UserSummaryResponse{ID: user.ID, FullName: user.FullName, AvatarURL: user.AvatarURL}
}

func executionCategorySummaryResponse(category *entity.CategorySummary) *executiondto.CategorySummaryResp {
	if category == nil {
		return nil
	}
	return &executiondto.CategorySummaryResp{ID: category.ID, Name: category.Name, Color: category.Color}
}

func executionTaskSummaryResponse(task *entity.TaskSummary) *executiondto.TaskSummaryResponse {
	if task == nil {
		return nil
	}
	return &executiondto.TaskSummaryResponse{ID: task.ID, Title: task.Title}
}

func executionWarningResponse(warnings []entity.Warning) []executiondto.WarningResponse {
	result := make([]executiondto.WarningResponse, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, executiondto.WarningResponse{Code: warning.Code, Message: warning.Message, ConflictTimeboxID: warning.ConflictTimeboxID})
	}
	return result
}
