package handler

import (
	"errors"
	"net/http"

	workspacedto "timebox-backend/internal/dto/workspace"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"
	"timebox-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	authService      *service.AuthService
	workspaceService *service.WorkspaceService
}

func newWorkspaceHandler(services *service.Service) *WorkspaceHandler {
	return &WorkspaceHandler{authService: services.Auth, workspaceService: services.Workspace}
}

func (h *WorkspaceHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	workspaces := routeGroup.Group("/workspaces")
	workspaces.GET("", h.ListWorkspaces)
	workspaces.POST("", h.CreateWorkspace)
	workspaces.GET("/:wsId", h.GetWorkspace)
	workspaces.PATCH("/:wsId", h.UpdateWorkspace)
	workspaces.POST("/:wsId/invite", h.InviteMember)
	workspaces.GET("/:wsId/members", h.ListMembers)
	workspaces.PATCH("/:wsId/members/:userId", h.UpdateMember)
	workspaces.GET("/:wsId/teams", h.ListTeams)
	workspaces.POST("/:wsId/teams", h.CreateTeam)

	teams := routeGroup.Group("/teams")
	teams.PATCH("/:id", h.UpdateTeam)
	teams.DELETE("/:id", h.DeleteTeam)
}

func (h *WorkspaceHandler) ListWorkspaces(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	pagination, ok := pagination(ctx)
	if !ok {
		return
	}
	workspaces, total, err := h.workspaceService.List(ctx, userID, ctx.Query("q"), ctx.Query("status"), pagination.Page, pagination.Limit)
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithPagination(ctx, workspaceListResponse(workspaces), "data fetched", http.StatusOK, response.NewPagination(pagination.Page, pagination.Limit, total))
}

func (h *WorkspaceHandler) CreateWorkspace(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req workspacedto.CreateWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	workspace, err := h.workspaceService.Create(ctx, userID, entity.Workspace{Name: req.Name, Slug: req.Slug, Timezone: req.Timezone, LogoURL: req.LogoURL})
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, workspaceCreateResponse(workspace), "Workspace created", http.StatusCreated)
}

func (h *WorkspaceHandler) GetWorkspace(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	workspace, err := h.workspaceService.FindByID(ctx, userID, ctx.Param("wsId"))
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, workspaceDetailResponse(workspace), "data fetched", http.StatusOK)
}

func (h *WorkspaceHandler) UpdateWorkspace(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req workspacedto.UpdateWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.Workspace{ID: ctx.Param("wsId")}
	if req.Name != nil {
		patch.Name = *req.Name
	}
	if req.Slug != nil {
		patch.Slug = *req.Slug
	}
	if req.Timezone != nil {
		patch.Timezone = *req.Timezone
	}
	patch.LogoURL = req.LogoURL

	var settingsPatch *service.WorkspaceSettingsPatch
	if req.Settings != nil {
		settingsPatch = &service.WorkspaceSettingsPatch{
			LeaderboardEnabled:            req.Settings.LeaderboardEnabled,
			DefaultPlannerIntervalMinutes: req.Settings.DefaultPlannerIntervalMinutes,
		}
	}
	workspace, err := h.workspaceService.Update(ctx, userID, patch, settingsPatch)
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, workspaceUpdateResponse(workspace), "Workspace updated", http.StatusOK)
}

func (h *WorkspaceHandler) InviteMember(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req workspacedto.InviteWorkspaceMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	invitation, err := h.workspaceService.InviteMember(ctx, userID, ctx.Param("wsId"), entity.WorkspaceInvitation{Email: req.Email, Role: req.Role}, req.TeamIDs)
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, invitationResponse(invitation), "Invitation created", http.StatusCreated)
}

func (h *WorkspaceHandler) ListMembers(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	pagination, ok := pagination(ctx)
	if !ok {
		return
	}
	members, total, err := h.workspaceService.ListMembers(ctx, userID, ctx.Param("wsId"), ctx.Query("role"), ctx.Query("status"), ctx.Query("q"), pagination.Page, pagination.Limit)
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithPagination(ctx, memberListResponse(members), "data fetched", http.StatusOK, response.NewPagination(pagination.Page, pagination.Limit, total))
}

func (h *WorkspaceHandler) UpdateMember(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req workspacedto.UpdateWorkspaceMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.WorkspaceMember{}
	if req.Role != nil {
		patch.Role = *req.Role
	}
	if req.Status != nil {
		patch.Status = *req.Status
	}
	if req.TeamIDs != nil {
		patch.TeamIDs = *req.TeamIDs
		patch.TeamIDsSet = true
	}
	member, err := h.workspaceService.UpdateMember(ctx, userID, ctx.Param("wsId"), ctx.Param("userId"), patch)
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, memberUpdateResponse(member), "Member updated", http.StatusOK)
}

func (h *WorkspaceHandler) ListTeams(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	teams, err := h.workspaceService.ListTeams(ctx, userID, ctx.Param("wsId"))
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, teamListResponse(teams), "data fetched", http.StatusOK)
}

func (h *WorkspaceHandler) CreateTeam(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req workspacedto.CreateTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	team, err := h.workspaceService.CreateTeam(ctx, userID, ctx.Param("wsId"), entity.Team{Name: req.Name, Description: req.Description})
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, teamCreateResponse(team), "Team created", http.StatusCreated)
}

func (h *WorkspaceHandler) UpdateTeam(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req workspacedto.UpdateTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	patch := entity.Team{ID: ctx.Param("id")}
	if req.Name != nil {
		patch.Name = *req.Name
	}
	if req.Description != nil {
		patch.Description = *req.Description
	}
	team, err := h.workspaceService.UpdateTeam(ctx, userID, patch)
	if err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithData(ctx, teamUpdateResponse(team), "Team updated", http.StatusOK)
}

func (h *WorkspaceHandler) DeleteTeam(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.workspaceService.DeleteTeam(ctx, userID, ctx.Param("id")); err != nil {
		writeWorkspaceError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Team deleted", http.StatusOK)
}

func (h *WorkspaceHandler) currentUserID(ctx *gin.Context) (string, bool) {
	userID, err := h.authService.ValidateAccessToken(bearerToken(ctx))
	if err != nil {
		writeAuthError(ctx, err)
		return "", false
	}
	return userID, true
}

func pagination(ctx *gin.Context) (utils.PaginationFilter, bool) {
	filter, err := utils.NewPaginationFilter(ctx.Query("page"), ctx.Query("limit"))
	if err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return utils.PaginationFilter{}, false
	}
	return filter, true
}

func writeWorkspaceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWorkspaceNotFound):
		response.Error(ctx, "not found", "workspace not found", http.StatusNotFound)
	case errors.Is(err, service.ErrWorkspaceSlugUsed):
		response.Error(ctx, "conflict", "slug already exists", http.StatusConflict)
	case errors.Is(err, service.ErrForbidden):
		response.Error(ctx, "forbidden", "forbidden", http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidRole), errors.Is(err, service.ErrInvalidStatus), errors.Is(err, service.ErrInvalidTimezone):
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
	default:
		response.Error(ctx, "internal server error", "workspace request failed", http.StatusInternalServerError)
	}
}

func workspaceListResponse(workspaces []entity.Workspace) []workspacedto.WorkspaceResponse {
	result := make([]workspacedto.WorkspaceResponse, 0, len(workspaces))
	for _, workspace := range workspaces {
		createdAt := workspace.CreatedAt
		result = append(result, workspacedto.WorkspaceResponse{
			ID:          workspace.ID,
			Name:        workspace.Name,
			Slug:        workspace.Slug,
			LogoURL:     workspace.LogoURL,
			Timezone:    workspace.Timezone,
			Role:        workspace.Role,
			MemberCount: workspace.MemberCount,
			Status:      workspace.Status,
			CreatedAt:   &createdAt,
		})
	}
	return result
}

func workspaceCreateResponse(workspace entity.Workspace) workspacedto.WorkspaceResponse {
	createdAt := workspace.CreatedAt
	return workspacedto.WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Slug:      workspace.Slug,
		LogoURL:   workspace.LogoURL,
		Timezone:  workspace.Timezone,
		OwnerID:   workspace.OwnerID,
		Status:    workspace.Status,
		CreatedAt: &createdAt,
	}
}

func workspaceDetailResponse(workspace entity.Workspace) workspacedto.WorkspaceResponse {
	createdAt := workspace.CreatedAt
	return workspacedto.WorkspaceResponse{
		ID:          workspace.ID,
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		LogoURL:     workspace.LogoURL,
		Timezone:    workspace.Timezone,
		Owner:       &workspacedto.WorkspaceOwnerResponse{ID: workspace.OwnerID, FullName: workspace.OwnerName},
		Settings:    settingsResponse(workspace.Settings),
		MemberCount: workspace.MemberCount,
		CreatedAt:   &createdAt,
	}
}

func workspaceUpdateResponse(workspace entity.Workspace) workspacedto.WorkspaceResponse {
	updatedAt := workspace.UpdatedAt
	return workspacedto.WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Slug:      workspace.Slug,
		LogoURL:   workspace.LogoURL,
		Timezone:  workspace.Timezone,
		Settings:  settingsResponse(workspace.Settings),
		UpdatedAt: &updatedAt,
	}
}

func settingsResponse(settings entity.WorkspaceSettings) *workspacedto.WorkspaceSettingsResponse {
	return &workspacedto.WorkspaceSettingsResponse{
		LeaderboardEnabled:            settings.LeaderboardEnabled,
		DefaultPlannerIntervalMinutes: settings.DefaultPlannerIntervalMinutes,
	}
}

func invitationResponse(invitation entity.WorkspaceInvitation) workspacedto.WorkspaceInvitationResponse {
	return workspacedto.WorkspaceInvitationResponse{
		InviteID:  invitation.ID,
		Email:     invitation.Email,
		Role:      invitation.Role,
		Status:    invitation.Status,
		ExpiresAt: invitation.ExpiresAt,
	}
}

func memberListResponse(members []entity.WorkspaceMember) []workspacedto.WorkspaceMemberResponse {
	result := make([]workspacedto.WorkspaceMemberResponse, 0, len(members))
	for _, member := range members {
		result = append(result, workspacedto.WorkspaceMemberResponse{
			UserID:    member.UserID,
			FullName:  member.FullName,
			Email:     member.Email,
			AvatarURL: member.AvatarURL,
			Role:      member.Role,
			Status:    member.Status,
			Teams:     teamSummaryResponse(member.Teams),
			JoinedAt:  member.JoinedAt,
		})
	}
	return result
}

func memberUpdateResponse(member entity.WorkspaceMember) workspacedto.WorkspaceMemberUpdateResponse {
	return workspacedto.WorkspaceMemberUpdateResponse{
		UserID:    member.UserID,
		Role:      member.Role,
		Status:    member.Status,
		TeamIDs:   member.TeamIDs,
		UpdatedAt: member.UpdatedAt,
	}
}

func teamListResponse(teams []entity.Team) []workspacedto.TeamResponse {
	result := make([]workspacedto.TeamResponse, 0, len(teams))
	for _, team := range teams {
		createdAt := team.CreatedAt
		result = append(result, workspacedto.TeamResponse{
			ID:          team.ID,
			WorkspaceID: team.WorkspaceID,
			Name:        team.Name,
			Description: team.Description,
			MemberCount: team.MemberCount,
			CreatedAt:   &createdAt,
		})
	}
	return result
}

func teamCreateResponse(team entity.Team) workspacedto.TeamResponse {
	createdAt := team.CreatedAt
	return workspacedto.TeamResponse{ID: team.ID, WorkspaceID: team.WorkspaceID, Name: team.Name, Description: team.Description, CreatedAt: &createdAt}
}

func teamUpdateResponse(team entity.Team) workspacedto.TeamResponse {
	updatedAt := team.UpdatedAt
	return workspacedto.TeamResponse{ID: team.ID, Name: team.Name, Description: team.Description, UpdatedAt: &updatedAt}
}

func teamSummaryResponse(teams []entity.TeamSummary) []workspacedto.TeamSummaryResponse {
	result := make([]workspacedto.TeamSummaryResponse, 0, len(teams))
	for _, team := range teams {
		result = append(result, workspacedto.TeamSummaryResponse{ID: team.ID, Name: team.Name})
	}
	return result
}
