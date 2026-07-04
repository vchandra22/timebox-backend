package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	analyticsdto "timebox-backend/internal/dto/analytics"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	authService      *service.AuthService
	analyticsService *service.AnalyticsService
}

func newAnalyticsHandler(services *service.Service) *AnalyticsHandler {
	return &AnalyticsHandler{authService: services.Auth, analyticsService: services.Analytics}
}

func (h *AnalyticsHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/streaks/me", h.MyStreak)
	routeGroup.GET("/workspaces/:wsId/leaderboard", h.Leaderboard)
	routeGroup.GET("/dashboard/personal", h.PersonalDashboard)
	routeGroup.GET("/dashboard/workspace", h.WorkspaceDashboard)
	routeGroup.GET("/reports/time-allocation", h.TimeAllocationReport)
	routeGroup.GET("/reports/productivity-trend", h.ProductivityTrendReport)
	routeGroup.GET("/reports/team-workload", h.TeamWorkloadReport)
	routeGroup.GET("/search", h.Search)
	routeGroup.GET("/saved-views", h.ListSavedViews)
	routeGroup.POST("/saved-views", h.CreateSavedView)
	routeGroup.PATCH("/saved-views/:id", h.UpdateSavedView)
	routeGroup.DELETE("/saved-views/:id", h.DeleteSavedView)
	routeGroup.GET("/activity-logs", h.ActivityLogs)
}

func (h *AnalyticsHandler) MyStreak(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	streak, err := h.analyticsService.Streak(ctx, userID, ctx.Query("workspace_id"))
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, streakResponse(streak), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) Leaderboard(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	leaderboard, err := h.analyticsService.Leaderboard(ctx, userID, ctx.Param("wsId"), ctx.Query("metric"), ctx.Query("period"), limit)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, leaderboardResponse(leaderboard), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) PersonalDashboard(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	date, ok := analyticsDate(ctx, ctx.Query("date"))
	if !ok {
		return
	}
	dashboard, err := h.analyticsService.PersonalDashboard(ctx, userID, ctx.Query("workspace_id"), date)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, personalDashboardResponse(dashboard), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) WorkspaceDashboard(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	date, ok := analyticsDate(ctx, ctx.Query("date"))
	if !ok {
		return
	}
	dashboard, err := h.analyticsService.WorkspaceDashboard(ctx, userID, ctx.Query("workspace_id"), date)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, workspaceDashboardResponse(dashboard), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) TimeAllocationReport(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	filter, ok := reportFilter(ctx)
	if !ok {
		return
	}
	filter.GroupBy = ctx.DefaultQuery("group_by", "category")
	filter.UserID = ctx.Query("user_id")
	filter.CategoryID = ctx.Query("category_id")
	filter.GoalID = ctx.Query("goal_id")
	report, err := h.analyticsService.TimeAllocationReport(ctx, userID, filter)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, timeAllocationReportResponse(report), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) ProductivityTrendReport(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	filter, ok := reportFilter(ctx)
	if !ok {
		return
	}
	filter.Interval = ctx.DefaultQuery("interval", "day")
	report, err := h.analyticsService.ProductivityTrendReport(ctx, userID, filter)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, productivityTrendResponse(report), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) TeamWorkloadReport(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	filter, ok := reportFilter(ctx)
	if !ok {
		return
	}
	report, err := h.analyticsService.TeamWorkloadReport(ctx, userID, filter)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, teamWorkloadResponse(report), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) Search(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	results, err := h.analyticsService.Search(ctx, userID, ctx.Query("workspace_id"), ctx.Query("q"), splitAnalyticsCSV(ctx.Query("types")), limit)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, searchResponse(results), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) ListSavedViews(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	views, err := h.analyticsService.ListSavedViews(ctx, userID, ctx.Query("workspace_id"), ctx.Query("resource_type"))
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, savedViewListResponse(views), "data fetched", http.StatusOK)
}

func (h *AnalyticsHandler) CreateSavedView(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req analyticsdto.CreateSavedViewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	view, err := h.analyticsService.CreateSavedView(ctx, userID, entity.SavedView{WorkspaceID: req.WorkspaceID, Name: req.Name, ResourceType: req.ResourceType, FilterJSON: req.FilterJSON, Shared: req.Shared})
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, savedViewResponse(view, false), "Saved view created", http.StatusCreated)
}

func (h *AnalyticsHandler) UpdateSavedView(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req analyticsdto.UpdateSavedViewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.SavedView{ID: ctx.Param("id"), FilterJSON: req.FilterJSON}
	if req.Name != nil {
		patch.Name = *req.Name
	}
	if req.Shared != nil {
		patch.Shared = *req.Shared
		patch.SharedSet = true
	}
	view, err := h.analyticsService.UpdateSavedView(ctx, userID, patch)
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithData(ctx, savedViewResponse(view, true), "Saved view updated", http.StatusOK)
}

func (h *AnalyticsHandler) DeleteSavedView(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.analyticsService.DeleteSavedView(ctx, userID, ctx.Param("id")); err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Saved view deleted", http.StatusOK)
}

func (h *AnalyticsHandler) ActivityLogs(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	pagination, ok := pagination(ctx)
	if !ok {
		return
	}
	startDate, ok := optionalAnalyticsDate(ctx, ctx.Query("start_date"))
	if !ok {
		return
	}
	endDate, ok := optionalAnalyticsDate(ctx, ctx.Query("end_date"))
	if !ok {
		return
	}
	logs, total, err := h.analyticsService.ActivityLogs(ctx, userID, service.ActivityLogFilter{WorkspaceID: ctx.Query("workspace_id"), ActorID: ctx.Query("actor_id"), Action: ctx.Query("action"), ResourceType: ctx.Query("resource_type"), ResourceID: ctx.Query("resource_id"), StartDate: startDate, EndDate: endDate, Limit: pagination.Limit, Offset: pagination.Offset()})
	if err != nil {
		writeAnalyticsError(ctx, err)
		return
	}
	response.WithPagination(ctx, activityLogListResponse(logs), "data fetched", http.StatusOK, response.NewPagination(pagination.Page, pagination.Limit, total))
}

func (h *AnalyticsHandler) currentUserID(ctx *gin.Context) (string, bool) {
	userID, err := h.authService.ValidateAccessToken(bearerToken(ctx))
	if err != nil {
		writeAuthError(ctx, err)
		return "", false
	}
	return userID, true
}

func writeAnalyticsError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAnalyticsNotFound):
		response.Error(ctx, "not found", "resource not found", http.StatusNotFound)
	case errors.Is(err, service.ErrForbidden):
		response.Error(ctx, "forbidden", "forbidden", http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidAnalyticsRange), errors.Is(err, service.ErrInvalidAnalyticsFilter):
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
	default:
		response.Error(ctx, "internal server error", "analytics request failed", http.StatusInternalServerError)
	}
}

func analyticsDate(ctx *gin.Context, value string) (time.Time, bool) {
	if value == "" {
		return time.Now(), true
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
		return time.Time{}, false
	}
	return parsed, true
}

func optionalAnalyticsDate(ctx *gin.Context, value string) (*time.Time, bool) {
	if value == "" {
		return nil, true
	}
	parsed, ok := analyticsDate(ctx, value)
	return &parsed, ok
}

func reportFilter(ctx *gin.Context) (service.ReportFilter, bool) {
	startDate, ok := requiredAnalyticsDate(ctx, ctx.Query("start_date"))
	if !ok {
		return service.ReportFilter{}, false
	}
	endDate, ok := requiredAnalyticsDate(ctx, ctx.Query("end_date"))
	if !ok {
		return service.ReportFilter{}, false
	}
	return service.ReportFilter{WorkspaceID: ctx.Query("workspace_id"), StartDate: startDate, EndDate: endDate}, true
}

func requiredAnalyticsDate(ctx *gin.Context, value string) (time.Time, bool) {
	if value == "" {
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
		return time.Time{}, false
	}
	return analyticsDate(ctx, value)
}

func splitAnalyticsCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func streakResponse(streak entity.Streak) analyticsdto.StreakResponse {
	var lastDate *string
	if streak.LastCompletedDate != nil {
		formatted := streak.LastCompletedDate.Format("2006-01-02")
		lastDate = &formatted
	}
	return analyticsdto.StreakResponse{WorkspaceID: streak.WorkspaceID, CurrentStreak: streak.CurrentStreak, LongestStreak: streak.LongestStreak, LastCompletedDate: lastDate, Badges: nil}
}

func leaderboardResponse(leaderboard entity.Leaderboard) analyticsdto.LeaderboardResponse {
	items := make([]analyticsdto.LeaderboardItemResponse, 0, len(leaderboard.Items))
	for _, item := range leaderboard.Items {
		items = append(items, analyticsdto.LeaderboardItemResponse{Rank: item.Rank, User: analyticsUserResponse(item.User), Value: item.Value})
	}
	return analyticsdto.LeaderboardResponse{Metric: leaderboard.Metric, Period: leaderboard.Period, Items: items}
}

func personalDashboardResponse(d entity.PersonalDashboard) analyticsdto.PersonalDashboardResponse {
	return analyticsdto.PersonalDashboardResponse{Date: d.Date, Timezone: d.Timezone, ActiveTimer: nil, TodayTimeline: timelineResponse(d.TodayTimeline), Summary: dashboardSummaryResponse(d.Summary), CategoryDistribution: categoryDistributionResponse(d.CategoryDistribution), OverrunOrMissed: timelineResponse(d.OverrunOrMissed)}
}

func workspaceDashboardResponse(d entity.WorkspaceDashboard) analyticsdto.WorkspaceDashboardResponse {
	items := make([]analyticsdto.TeamTodayResponse, 0, len(d.TeamToday))
	for _, item := range d.TeamToday {
		items = append(items, analyticsdto.TeamTodayResponse{User: analyticsUserResponse(item.User), PlannedMinutes: item.PlannedMinutes, ActualMinutes: item.ActualMinutes, CompletionRate: item.CompletionRate})
	}
	return analyticsdto.WorkspaceDashboardResponse{Date: d.Date, Summary: dashboardSummaryResponse(d.Summary), TeamToday: items, CategoryDistribution: categoryDistributionResponse(d.CategoryDistribution), Trend: trendResponse(d.Trend)}
}

func dashboardSummaryResponse(s entity.DashboardSummary) analyticsdto.DashboardSummaryResponse {
	return analyticsdto.DashboardSummaryResponse{PlannedMinutes: s.PlannedMinutes, ActualMinutes: s.ActualMinutes, CompletedTimeboxes: s.CompletedTimeboxes, TotalTimeboxes: s.TotalTimeboxes, CompletionRate: s.CompletionRate, CurrentStreak: s.CurrentStreak, TeamFocusMinutesToday: s.TeamFocusMinutesToday, TeamFocusMinutesWeek: s.TeamFocusMinutesWeek, ActiveMembersNow: s.ActiveMembersNow, OverrunRiskMembers: s.OverrunRiskMembers}
}

func timelineResponse(items []entity.TimelineItem) []analyticsdto.TimelineItemResponse {
	result := make([]analyticsdto.TimelineItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, analyticsdto.TimelineItemResponse{ID: item.ID, Title: item.Title, ScheduledStart: item.ScheduledStart, ScheduledEnd: item.ScheduledEnd, Status: item.Status})
	}
	return result
}

func categoryDistributionResponse(items []entity.CategoryDistribution) []analyticsdto.CategoryDistributionResponse {
	result := make([]analyticsdto.CategoryDistributionResponse, 0, len(items))
	for _, item := range items {
		result = append(result, analyticsdto.CategoryDistributionResponse{CategoryID: item.CategoryID, CategoryName: item.CategoryName, PlannedMinutes: item.PlannedMinutes, ActualMinutes: item.ActualMinutes})
	}
	return result
}

func timeAllocationReportResponse(report entity.TimeAllocationReport) analyticsdto.TimeAllocationReportResponse {
	return analyticsdto.TimeAllocationReportResponse{Range: analyticsdto.ReportRangeResponse{StartDate: report.Range.StartDate, EndDate: report.Range.EndDate}, GroupBy: report.GroupBy, Summary: analyticsdto.ReportSummaryResponse{PlannedMinutes: report.Summary.PlannedMinutes, ActualMinutes: report.Summary.ActualMinutes, VarianceMinutes: report.Summary.VarianceMinutes}, Items: reportItemsResponse(report.Items)}
}

func reportItemsResponse(items []entity.ReportItem) []analyticsdto.ReportItemResponse {
	result := make([]analyticsdto.ReportItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, analyticsdto.ReportItemResponse{Key: item.Key, Label: item.Label, PlannedMinutes: item.PlannedMinutes, ActualMinutes: item.ActualMinutes, VarianceMinutes: item.VarianceMinutes, Percentage: item.Percentage})
	}
	return result
}

func productivityTrendResponse(report entity.ProductivityTrendReport) analyticsdto.ProductivityTrendResponse {
	return analyticsdto.ProductivityTrendResponse{Interval: report.Interval, Items: trendResponse(report.Items)}
}

func trendResponse(items []entity.TrendItem) []analyticsdto.TrendItemResponse {
	result := make([]analyticsdto.TrendItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, analyticsdto.TrendItemResponse{Date: item.Date, PlannedMinutes: item.PlannedMinutes, ActualMinutes: item.ActualMinutes, CompletedTimeboxes: item.CompletedTimeboxes, TotalTimeboxes: item.TotalTimeboxes, CompletionRate: item.CompletionRate})
	}
	return result
}

func teamWorkloadResponse(report entity.TeamWorkloadReport) analyticsdto.TeamWorkloadResponse {
	items := make([]analyticsdto.TeamWorkloadItemResponse, 0, len(report.Items))
	for _, item := range report.Items {
		items = append(items, analyticsdto.TeamWorkloadItemResponse{User: analyticsUserResponse(item.User), PlannedMinutes: item.PlannedMinutes, ActualMinutes: item.ActualMinutes, CompletedTimeboxes: item.CompletedTimeboxes, OverrunCount: item.OverrunCount, RiskLevel: item.RiskLevel})
	}
	return analyticsdto.TeamWorkloadResponse{Range: analyticsdto.ReportRangeResponse{StartDate: report.Range.StartDate, EndDate: report.Range.EndDate}, Items: items}
}

func searchResponse(results entity.SearchResults) analyticsdto.SearchResponse {
	return analyticsdto.SearchResponse{Query: results.Query, Results: map[string][]analyticsdto.SearchItemResponse{"tasks": searchItemsResponse(results.Tasks), "timeboxes": searchItemsResponse(results.Timeboxes), "goals": searchItemsResponse(results.Goals), "comments": searchItemsResponse(results.Comments), "attachments": searchItemsResponse(results.Attachments)}}
}

func searchItemsResponse(items []entity.SearchResult) []analyticsdto.SearchItemResponse {
	result := make([]analyticsdto.SearchItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, analyticsdto.SearchItemResponse{ID: item.ID, Title: item.Title, Snippet: item.Snippet})
	}
	return result
}

func savedViewListResponse(views []entity.SavedView) []analyticsdto.SavedViewResponse {
	result := make([]analyticsdto.SavedViewResponse, 0, len(views))
	for _, view := range views {
		result = append(result, savedViewResponse(view, true))
	}
	return result
}

func savedViewResponse(view entity.SavedView, includeFilter bool) analyticsdto.SavedViewResponse {
	response := analyticsdto.SavedViewResponse{ID: view.ID, Name: view.Name, ResourceType: view.ResourceType, Shared: view.Shared}
	if includeFilter {
		response.FilterJSON = view.FilterJSON
	}
	return response
}

func activityLogListResponse(logs []entity.ActivityLog) []analyticsdto.ActivityLogResponse {
	result := make([]analyticsdto.ActivityLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, analyticsdto.ActivityLogResponse{ID: log.ID, WorkspaceID: log.WorkspaceID, Actor: optionalAnalyticsUserResponse(log.Actor), Action: log.Action, ResourceType: log.ResourceType, ResourceID: log.ResourceID, OldValue: log.OldValue, NewValue: log.NewValue, IPAddress: log.IPAddress, UserAgent: log.UserAgent, CreatedAt: log.CreatedAt})
	}
	return result
}

func analyticsUserResponse(user entity.UserSummary) analyticsdto.UserSummaryResponse {
	return analyticsdto.UserSummaryResponse{ID: user.ID, FullName: user.FullName, AvatarURL: user.AvatarURL}
}

func optionalAnalyticsUserResponse(user *entity.UserSummary) *analyticsdto.UserSummaryResponse {
	if user == nil {
		return nil
	}
	response := analyticsUserResponse(*user)
	return &response
}
