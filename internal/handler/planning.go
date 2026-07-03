package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	planningdto "timebox-backend/internal/dto/planning"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PlanningHandler struct {
	authService     *service.AuthService
	planningService *service.PlanningService
}

func newPlanningHandler(services *service.Service) *PlanningHandler {
	return &PlanningHandler{authService: services.Auth, planningService: services.Planning}
}

func (h *PlanningHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	workspaces := routeGroup.Group("/workspaces/:wsId")
	workspaces.GET("/categories", h.ListCategories)
	workspaces.POST("/categories", h.CreateCategory)
	workspaces.GET("/goals", h.ListGoals)
	workspaces.POST("/goals", h.CreateGoal)
	workspaces.GET("/tasks", h.ListTasks)
	workspaces.POST("/tasks", h.CreateTask)

	routeGroup.PATCH("/categories/:id", h.UpdateCategory)
	routeGroup.DELETE("/categories/:id", h.DeleteCategory)
	routeGroup.GET("/tags", h.ListTags)
	routeGroup.POST("/tags", h.CreateTag)
	routeGroup.PATCH("/tags/:id", h.UpdateTag)
	routeGroup.DELETE("/tags/:id", h.DeleteTag)
	routeGroup.GET("/goals/:id", h.GetGoal)
	routeGroup.PATCH("/goals/:id", h.UpdateGoal)
	routeGroup.DELETE("/goals/:id", h.ArchiveGoal)
	routeGroup.GET("/tasks/:id", h.GetTask)
	routeGroup.PATCH("/tasks/:id", h.UpdateTask)
	routeGroup.DELETE("/tasks/:id", h.DeleteTask)
	routeGroup.PATCH("/tasks/:id/move", h.MoveTask)
}

func (h *PlanningHandler) ListCategories(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	categories, err := h.planningService.ListCategories(ctx, userID, ctx.Param("wsId"))
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, categoryListResponse(categories), "data fetched", http.StatusOK)
}

func (h *PlanningHandler) CreateCategory(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	category, err := h.planningService.CreateCategory(ctx, userID, ctx.Param("wsId"), entity.Category{Name: req.Name, Color: req.Color})
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, categoryResponse(category, false), "Category created", http.StatusCreated)
}

func (h *PlanningHandler) UpdateCategory(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.Category{ID: ctx.Param("id")}
	if req.Name != nil {
		patch.Name = *req.Name
	}
	if req.Color != nil {
		patch.Color = *req.Color
	}
	category, err := h.planningService.UpdateCategory(ctx, userID, patch)
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, categoryResponse(category, false), "Category updated", http.StatusOK)
}

func (h *PlanningHandler) DeleteCategory(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.planningService.DeleteCategory(ctx, userID, ctx.Param("id")); err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, planningdto.CategoryDeleteResponse{}, "Category deleted or moved to default category", http.StatusOK)
}

func (h *PlanningHandler) ListTags(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	tags, err := h.planningService.ListTags(ctx, userID, ctx.Query("workspace_id"), ctx.Query("q"))
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, tagListResponse(tags), "data fetched", http.StatusOK)
}

func (h *PlanningHandler) CreateTag(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.CreateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	tag, err := h.planningService.CreateTag(ctx, userID, entity.Tag{WorkspaceID: req.WorkspaceID, Name: req.Name})
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, tagResponse(tag), "Tag created", http.StatusCreated)
}

func (h *PlanningHandler) UpdateTag(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.UpdateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.Tag{ID: ctx.Param("id")}
	if req.Name != nil {
		patch.Name = *req.Name
	}
	tag, err := h.planningService.UpdateTag(ctx, userID, patch)
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, tagResponse(tag), "Tag updated", http.StatusOK)
}

func (h *PlanningHandler) DeleteTag(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.planningService.DeleteTag(ctx, userID, ctx.Param("id")); err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Tag deleted", http.StatusOK)
}

func (h *PlanningHandler) ListGoals(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	pagination, ok := pagination(ctx)
	if !ok {
		return
	}
	pinned, pinnedSet := queryBool(ctx.Query("pinned"))
	goals, total, err := h.planningService.ListGoals(ctx, userID, service.GoalListFilter{
		WorkspaceID: ctx.Param("wsId"),
		Q:           ctx.Query("q"),
		Status:      ctx.Query("status"),
		PinnedSet:   pinnedSet,
		Pinned:      pinned,
		Limit:       pagination.Limit,
		Offset:      pagination.Offset(),
	})
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithPagination(ctx, goalListResponse(goals), "data fetched", http.StatusOK, response.NewPagination(pagination.Page, pagination.Limit, total))
}

func (h *PlanningHandler) CreateGoal(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.CreateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	targetDate, ok := parseDate(ctx, req.TargetDate)
	if !ok {
		return
	}
	goal, err := h.planningService.CreateGoal(ctx, userID, ctx.Param("wsId"), entity.Goal{Title: req.Title, Description: req.Description, TargetDate: targetDate, IsPinned: req.IsPinned})
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, goalCreateResponse(goal), "Goal created", http.StatusCreated)
}

func (h *PlanningHandler) GetGoal(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	goal, err := h.planningService.FindGoal(ctx, userID, ctx.Param("id"))
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, goalDetailResponse(goal), "data fetched", http.StatusOK)
}

func (h *PlanningHandler) UpdateGoal(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.UpdateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.Goal{ID: ctx.Param("id")}
	if req.Title != nil {
		patch.Title = *req.Title
	}
	if req.Description != nil {
		patch.Description = *req.Description
	}
	if req.TargetDate != nil {
		targetDate, ok := parseDate(ctx, *req.TargetDate)
		if !ok {
			return
		}
		patch.TargetDate = targetDate
	}
	if req.Status != nil {
		patch.Status = *req.Status
	}
	if req.IsPinned != nil {
		patch.IsPinned = *req.IsPinned
		patch.IsPinnedSet = true
	}
	goal, err := h.planningService.UpdateGoal(ctx, userID, patch)
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, goalUpdateResponse(goal), "Goal updated", http.StatusOK)
}

func (h *PlanningHandler) ArchiveGoal(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.planningService.ArchiveGoal(ctx, userID, ctx.Param("id")); err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Goal archived", http.StatusOK)
}

func (h *PlanningHandler) ListTasks(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	pagination, ok := pagination(ctx)
	if !ok {
		return
	}
	tasks, total, err := h.planningService.ListTasks(ctx, userID, service.TaskListFilter{
		WorkspaceID: ctx.Param("wsId"),
		Q:           ctx.Query("q"),
		Status:      ctx.Query("status"),
		Priority:    ctx.Query("priority"),
		GoalID:      ctx.Query("goal_id"),
		AssigneeID:  ctx.Query("assignee_id"),
		CategoryID:  ctx.Query("category_id"),
		TagIDs:      splitCSV(ctx.Query("tag_ids")),
		IncludeDone: ctx.Query("include_done") == "true",
		Sort:        ctx.Query("sort"),
		Order:       ctx.Query("order"),
		Limit:       pagination.Limit,
		Offset:      pagination.Offset(),
	})
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithPagination(ctx, taskListResponse(tasks), "data fetched", http.StatusOK, response.NewPagination(pagination.Page, pagination.Limit, total))
}

func (h *PlanningHandler) CreateTask(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	task, err := h.planningService.CreateTask(ctx, userID, ctx.Param("wsId"), entity.Task{
		GoalID:           req.GoalID,
		AssigneeID:       req.AssigneeID,
		CategoryID:       req.CategoryID,
		Title:            req.Title,
		Description:      req.Description,
		Priority:         req.Priority,
		EstimatedMinutes: req.EstimatedMinutes,
		TagIDs:           req.TagIDs,
		Checklist:        checklistEntity(req.Checklist),
	})
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, taskCreateResponse(task), "Task created", http.StatusCreated)
}

func (h *PlanningHandler) GetTask(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	task, err := h.planningService.FindTask(ctx, userID, ctx.Param("id"))
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, taskDetailResponse(task), "data fetched", http.StatusOK)
}

func (h *PlanningHandler) UpdateTask(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.Task{ID: ctx.Param("id"), GoalID: req.GoalID, AssigneeID: req.AssigneeID, CategoryID: req.CategoryID, EstimatedMinutes: req.EstimatedMinutes}
	if req.Title != nil {
		patch.Title = *req.Title
	}
	if req.Description != nil {
		patch.Description = *req.Description
	}
	if req.Status != nil {
		patch.Status = *req.Status
	}
	if req.Priority != nil {
		patch.Priority = *req.Priority
	}
	if req.TagIDs != nil {
		patch.TagIDs = *req.TagIDs
		patch.TagIDsSet = true
	}
	task, err := h.planningService.UpdateTask(ctx, userID, patch)
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, taskUpdateResponse(task), "Task updated", http.StatusOK)
}

func (h *PlanningHandler) DeleteTask(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.planningService.DeleteTask(ctx, userID, ctx.Param("id")); err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Task deleted", http.StatusOK)
}

func (h *PlanningHandler) MoveTask(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req planningdto.MoveTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	moved, err := h.planningService.MoveTask(ctx, userID, ctx.Param("id"), req.ToStatus, req.Position)
	if err != nil {
		writePlanningError(ctx, err)
		return
	}
	response.WithData(ctx, planningdto.TaskMoveResponse{ID: moved.ID, FromStatus: moved.FromStatus, ToStatus: moved.ToStatus, Position: moved.Position, UpdatedAt: moved.UpdatedAt}, "Task moved", http.StatusOK)
}

func (h *PlanningHandler) currentUserID(ctx *gin.Context) (string, bool) {
	userID, err := h.authService.ValidateAccessToken(bearerToken(ctx))
	if err != nil {
		writeAuthError(ctx, err)
		return "", false
	}
	return userID, true
}

func writePlanningError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPlanningNotFound):
		response.Error(ctx, "not found", "resource not found", http.StatusNotFound)
	case errors.Is(err, service.ErrPlanningConflict):
		response.Error(ctx, "conflict", "resource already exists", http.StatusConflict)
	case errors.Is(err, service.ErrForbidden):
		response.Error(ctx, "forbidden", "forbidden", http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidGoalStatus), errors.Is(err, service.ErrInvalidTaskStatus), errors.Is(err, service.ErrInvalidPriority):
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
	default:
		response.Error(ctx, "internal server error", "planning request failed", http.StatusInternalServerError)
	}
}

func parseDate(ctx *gin.Context, value string) (*time.Time, bool) {
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

func queryBool(value string) (bool, bool) {
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	return parsed, err == nil
}

func splitCSV(value string) []string {
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

func categoryListResponse(categories []entity.Category) []planningdto.CategoryResponse {
	result := make([]planningdto.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		result = append(result, categoryResponse(category, true))
	}
	return result
}

func categoryResponse(category entity.Category, includeWorkspace bool) planningdto.CategoryResponse {
	createdAt := category.CreatedAt
	response := planningdto.CategoryResponse{ID: category.ID, Name: category.Name, Color: category.Color, IsDefault: category.IsDefault, CreatedAt: &createdAt}
	if includeWorkspace {
		response.WorkspaceID = category.WorkspaceID
	}
	return response
}

func tagListResponse(tags []entity.Tag) []planningdto.TagResponse {
	result := make([]planningdto.TagResponse, 0, len(tags))
	for _, tag := range tags {
		result = append(result, tagResponse(tag))
	}
	return result
}

func tagResponse(tag entity.Tag) planningdto.TagResponse {
	return planningdto.TagResponse{ID: tag.ID, WorkspaceID: tag.WorkspaceID, Name: tag.Name}
}

func goalListResponse(goals []entity.Goal) []planningdto.GoalResponse {
	result := make([]planningdto.GoalResponse, 0, len(goals))
	for _, goal := range goals {
		createdAt := goal.CreatedAt
		result = append(result, planningdto.GoalResponse{ID: goal.ID, WorkspaceID: goal.WorkspaceID, Title: goal.Title, Description: goal.Description, TargetDate: dateString(goal.TargetDate), Status: goal.Status, IsPinned: goal.IsPinned, ProgressPercent: goal.ProgressPercent, CreatedAt: &createdAt})
	}
	return result
}

func goalCreateResponse(goal entity.Goal) planningdto.GoalResponse {
	createdAt := goal.CreatedAt
	return planningdto.GoalResponse{ID: goal.ID, WorkspaceID: goal.WorkspaceID, Title: goal.Title, Description: goal.Description, TargetDate: dateString(goal.TargetDate), Status: goal.Status, IsPinned: goal.IsPinned, CreatedAt: &createdAt}
}

func goalDetailResponse(goal entity.Goal) planningdto.GoalResponse {
	createdAt := goal.CreatedAt
	updatedAt := goal.UpdatedAt
	return planningdto.GoalResponse{
		ID:          goal.ID,
		WorkspaceID: goal.WorkspaceID,
		Title:       goal.Title,
		Description: goal.Description,
		TargetDate:  dateString(goal.TargetDate),
		Status:      goal.Status,
		IsPinned:    goal.IsPinned,
		Progress: &planningdto.GoalProgressResp{
			PlannedMinutes:     goal.PlannedMinutes,
			ActualMinutes:      goal.ActualMinutes,
			CompletedTimeboxes: goal.CompletedBlocks,
			ProgressPercent:    goal.ProgressPercent,
		},
		CreatedBy: &planningdto.CreatedByResp{ID: goal.CreatedBy, FullName: goal.CreatedByName},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}

func goalUpdateResponse(goal entity.Goal) planningdto.GoalResponse {
	updatedAt := goal.UpdatedAt
	return planningdto.GoalResponse{ID: goal.ID, Title: goal.Title, Description: goal.Description, TargetDate: dateString(goal.TargetDate), Status: goal.Status, IsPinned: goal.IsPinned, UpdatedAt: &updatedAt}
}

func taskListResponse(tasks []entity.Task) []planningdto.TaskResponse {
	result := make([]planningdto.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		createdAt := task.CreatedAt
		updatedAt := task.UpdatedAt
		item := taskBaseResponse(task)
		item.CreatedAt = &createdAt
		item.UpdatedAt = &updatedAt
		result = append(result, item)
	}
	return result
}

func taskCreateResponse(task entity.Task) planningdto.TaskResponse {
	createdAt := task.CreatedAt
	item := taskBaseResponse(task)
	item.AssigneeID = task.AssigneeID
	item.CreatedAt = &createdAt
	return item
}

func taskDetailResponse(task entity.Task) planningdto.TaskResponse {
	createdAt := task.CreatedAt
	updatedAt := task.UpdatedAt
	item := taskBaseResponse(task)
	item.Checklist = checklistResponse(task.Checklist)
	item.TimeboxesCount = task.TimeboxesCount
	item.CreatedAt = &createdAt
	item.UpdatedAt = &updatedAt
	return item
}

func taskUpdateResponse(task entity.Task) planningdto.TaskResponse {
	updatedAt := task.UpdatedAt
	return planningdto.TaskResponse{ID: task.ID, Title: task.Title, Status: task.Status, Priority: task.Priority, EstimatedMinutes: task.EstimatedMinutes, UpdatedAt: &updatedAt}
}

func taskBaseResponse(task entity.Task) planningdto.TaskResponse {
	return planningdto.TaskResponse{
		ID:               task.ID,
		WorkspaceID:      task.WorkspaceID,
		GoalID:           task.GoalID,
		CategoryID:       task.CategoryID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		Priority:         task.Priority,
		EstimatedMinutes: task.EstimatedMinutes,
		Position:         task.Position,
		Assignee:         userSummaryResponse(task.Assignee),
		Goal:             goalSummaryResponse(task.Goal),
		Tags:             tagListResponse(task.Tags),
	}
}

func userSummaryResponse(user *entity.UserSummary) *planningdto.UserSummaryResponse {
	if user == nil {
		return nil
	}
	return &planningdto.UserSummaryResponse{ID: user.ID, FullName: user.FullName, AvatarURL: user.AvatarURL}
}

func goalSummaryResponse(goal *entity.GoalSummary) *planningdto.GoalSummaryResponse {
	if goal == nil {
		return nil
	}
	return &planningdto.GoalSummaryResponse{ID: goal.ID, Title: goal.Title}
}

func checklistEntity(items []planningdto.TaskChecklistCreateReq) []entity.TaskChecklist {
	result := make([]entity.TaskChecklist, 0, len(items))
	for _, item := range items {
		result = append(result, entity.TaskChecklist{Title: item.Title})
	}
	return result
}

func checklistResponse(items []entity.TaskChecklist) []planningdto.TaskChecklistResponse {
	result := make([]planningdto.TaskChecklistResponse, 0, len(items))
	for _, item := range items {
		result = append(result, planningdto.TaskChecklistResponse{ID: item.ID, Title: item.Title, IsDone: item.IsDone, Position: item.Position})
	}
	return result
}

func dateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}
