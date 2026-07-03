package handler

import (
	"errors"
	"net/http"

	collaborationdto "timebox-backend/internal/dto/collaboration"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CollaborationHandler struct {
	authService          *service.AuthService
	collaborationService *service.CollaborationService
}

func newCollaborationHandler(services *service.Service) *CollaborationHandler {
	return &CollaborationHandler{authService: services.Auth, collaborationService: services.Collaboration}
}

func (h *CollaborationHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/comments", h.ListComments)
	routeGroup.POST("/comments", h.CreateComment)
	routeGroup.PATCH("/comments/:id", h.UpdateComment)
	routeGroup.DELETE("/comments/:id", h.DeleteComment)
	routeGroup.POST("/uploads/signature", h.UploadSignature)
	routeGroup.POST("/attachments", h.CreateAttachment)
	routeGroup.GET("/attachments", h.ListAttachments)
	routeGroup.DELETE("/attachments/:id", h.DeleteAttachment)
	routeGroup.GET("/notifications", h.ListNotifications)
	routeGroup.PATCH("/notifications/:id/read", h.MarkNotificationRead)
	routeGroup.PATCH("/notifications/read-all", h.MarkAllNotificationsRead)
	routeGroup.GET("/notifications/preferences", h.GetNotificationPreferences)
	routeGroup.PATCH("/notifications/preferences", h.UpdateNotificationPreferences)
}

func (h *CollaborationHandler) ListComments(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var parentID *string
	if value := ctx.Query("parent_id"); value != "" {
		parentID = &value
	}
	comments, err := h.collaborationService.ListComments(ctx, userID, ctx.Query("resource_type"), ctx.Query("resource_id"), parentID)
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationCommentListResponse(comments), "data fetched", http.StatusOK)
}

func (h *CollaborationHandler) CreateComment(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	comment, err := h.collaborationService.CreateComment(ctx, userID, entity.Comment{ResourceType: req.ResourceType, ResourceID: req.ResourceID, ParentID: req.ParentID, Body: req.Body, MentionUserIDs: req.MentionUserIDs, AttachmentIDs: req.AttachmentIDs})
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationCommentResponse(comment, false), "Comment created", http.StatusCreated)
}

func (h *CollaborationHandler) UpdateComment(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.UpdateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	comment, err := h.collaborationService.UpdateComment(ctx, userID, entity.Comment{ID: ctx.Param("id"), Body: req.Body})
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationdto.CommentResponse{ID: comment.ID, Body: comment.Body, EditedAt: comment.EditedAt}, "Comment updated", http.StatusOK)
}

func (h *CollaborationHandler) DeleteComment(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	if err := h.collaborationService.DeleteComment(ctx, userID, ctx.Param("id")); err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithoutData(ctx, "Comment deleted", http.StatusOK)
}

func (h *CollaborationHandler) UploadSignature(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.UploadSignatureRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	signature, err := h.collaborationService.GenerateUploadSignature(ctx, userID, entity.Attachment{WorkspaceID: req.WorkspaceID, ResourceType: req.ResourceType, ResourceID: req.ResourceID, FileName: req.FileName, FileType: req.FileType, FileSize: req.FileSize})
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationdto.UploadSignatureResponse{CloudName: signature.CloudName, APIKey: signature.APIKey, Timestamp: signature.Timestamp, Signature: signature.Signature, Folder: signature.Folder, UploadURL: signature.UploadURL, PublicIDPrefix: signature.PublicIDPrefix}, "Upload signature generated", http.StatusOK)
}

func (h *CollaborationHandler) CreateAttachment(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.CreateAttachmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	attachment, err := h.collaborationService.CreateAttachment(ctx, userID, entity.Attachment{WorkspaceID: req.WorkspaceID, ResourceType: req.ResourceType, ResourceID: req.ResourceID, CloudinaryPublicID: req.CloudinaryPublicID, URL: req.URL, FileName: req.FileName, FileType: req.FileType, FileSize: req.FileSize})
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationAttachmentResponse(attachment, true), "Attachment created", http.StatusCreated)
}

func (h *CollaborationHandler) ListAttachments(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	attachments, err := h.collaborationService.ListAttachments(ctx, userID, ctx.Query("resource_type"), ctx.Query("resource_id"))
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationAttachmentListResponse(attachments), "data fetched", http.StatusOK)
}

func (h *CollaborationHandler) DeleteAttachment(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	attachment, err := h.collaborationService.DeleteAttachment(ctx, userID, ctx.Param("id"))
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationdto.AttachmentDeleteResponse{ID: attachment.ID, DeleteAssetQueued: true}, "Attachment deletion queued", http.StatusAccepted)
}

func (h *CollaborationHandler) ListNotifications(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	pagination, ok := pagination(ctx)
	if !ok {
		return
	}
	notifications, total, err := h.collaborationService.ListNotifications(ctx, userID, ctx.Query("status"), ctx.Query("type"), pagination.Page, pagination.Limit)
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithPagination(ctx, collaborationNotificationListResponse(notifications), "data fetched", http.StatusOK, response.NewPagination(pagination.Page, pagination.Limit, total))
}

func (h *CollaborationHandler) MarkNotificationRead(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.MarkNotificationReadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	notification, err := h.collaborationService.MarkNotificationRead(ctx, userID, ctx.Param("id"), req.Read)
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationdto.NotificationReadResponse{ID: notification.ID, ReadAt: notification.ReadAt}, "Notification updated", http.StatusOK)
}

func (h *CollaborationHandler) MarkAllNotificationsRead(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.MarkAllNotificationsReadRequest
	_ = ctx.ShouldBindJSON(&req)
	count, err := h.collaborationService.MarkAllNotificationsRead(ctx, userID, req.Type)
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationdto.MarkAllReadResponse{UpdatedCount: count}, "All notifications marked as read", http.StatusOK)
}

func (h *CollaborationHandler) GetNotificationPreferences(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	preferences, err := h.collaborationService.GetPreferences(ctx, userID)
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationPreferenceListResponse(preferences), "data fetched", http.StatusOK)
}

func (h *CollaborationHandler) UpdateNotificationPreferences(ctx *gin.Context) {
	userID, ok := h.currentUserID(ctx)
	if !ok {
		return
	}
	var req collaborationdto.UpdateNotificationPreferencesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}
	count, err := h.collaborationService.UpdatePreferences(ctx, userID, collaborationPreferenceEntity(req.Preferences))
	if err != nil {
		writeCollaborationError(ctx, err)
		return
	}
	response.WithData(ctx, collaborationdto.NotificationPreferencesUpdateResponse{UpdatedCount: count}, "Notification preferences updated", http.StatusOK)
}

func (h *CollaborationHandler) currentUserID(ctx *gin.Context) (string, bool) {
	userID, err := h.authService.ValidateAccessToken(bearerToken(ctx))
	if err != nil {
		writeAuthError(ctx, err)
		return "", false
	}
	return userID, true
}

func writeCollaborationError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCollaborationNotFound):
		response.Error(ctx, "not found", "resource not found", http.StatusNotFound)
	case errors.Is(err, service.ErrCollaborationConflict):
		response.Error(ctx, "conflict", "resource conflict", http.StatusConflict)
	case errors.Is(err, service.ErrForbidden):
		response.Error(ctx, "forbidden", "forbidden", http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidResourceType), errors.Is(err, service.ErrInvalidUpload), errors.Is(err, service.ErrCloudinaryNotReady):
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
	default:
		response.Error(ctx, "internal server error", "collaboration request failed", http.StatusInternalServerError)
	}
}

func collaborationCommentListResponse(comments []entity.Comment) []collaborationdto.CommentResponse {
	result := make([]collaborationdto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		result = append(result, collaborationCommentResponse(comment, true))
	}
	return result
}

func collaborationCommentResponse(comment entity.Comment, includeRelations bool) collaborationdto.CommentResponse {
	createdAt := comment.CreatedAt
	item := collaborationdto.CommentResponse{ID: comment.ID, ResourceType: comment.ResourceType, ResourceID: comment.ResourceID, ParentID: comment.ParentID, Body: comment.Body, EditedAt: comment.EditedAt, CreatedAt: &createdAt}
	if includeRelations {
		item.Author = &collaborationdto.AuthorResponse{ID: comment.AuthorID, FullName: comment.AuthorName}
		item.Mentions = collaborationMentionResponse(comment.Mentions)
		item.Attachments = collaborationAttachmentListResponse(comment.Attachments)
	}
	return item
}

func collaborationMentionResponse(mentions []entity.Mention) []collaborationdto.MentionResponse {
	result := make([]collaborationdto.MentionResponse, 0, len(mentions))
	for _, mention := range mentions {
		result = append(result, collaborationdto.MentionResponse{UserID: mention.UserID, FullName: mention.FullName})
	}
	return result
}

func collaborationAttachmentListResponse(attachments []entity.Attachment) []collaborationdto.AttachmentResponse {
	result := make([]collaborationdto.AttachmentResponse, 0, len(attachments))
	for _, attachment := range attachments {
		result = append(result, collaborationAttachmentResponse(attachment, true))
	}
	return result
}

func collaborationAttachmentResponse(attachment entity.Attachment, includeCreated bool) collaborationdto.AttachmentResponse {
	item := collaborationdto.AttachmentResponse{ID: attachment.ID, ResourceType: attachment.ResourceType, ResourceID: attachment.ResourceID, URL: attachment.URL, ThumbnailURL: attachment.ThumbnailURL, FileName: attachment.FileName, FileType: attachment.FileType, FileSize: attachment.FileSize, UploadedBy: &collaborationdto.AuthorResponse{ID: attachment.UploadedBy, FullName: attachment.UploadedByName}}
	if includeCreated {
		createdAt := attachment.CreatedAt
		item.CreatedAt = &createdAt
	}
	return item
}

func collaborationNotificationListResponse(notifications []entity.Notification) []collaborationdto.NotificationResponse {
	result := make([]collaborationdto.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		result = append(result, collaborationdto.NotificationResponse{ID: notification.ID, Type: notification.Type, Title: notification.Title, Body: notification.Body, Payload: notification.Payload, ReadAt: notification.ReadAt, CreatedAt: notification.CreatedAt})
	}
	return result
}

func collaborationPreferenceListResponse(preferences []entity.NotificationPreference) []collaborationdto.NotificationPreferenceResponse {
	result := make([]collaborationdto.NotificationPreferenceResponse, 0, len(preferences))
	for _, preference := range preferences {
		result = append(result, collaborationdto.NotificationPreferenceResponse{TriggerType: preference.TriggerType, Channels: collaborationdto.NotificationChannelsResponse{InApp: preference.InAppEnabled, Email: preference.EmailEnabled, Telegram: false}, ReminderMinutesBefore: preference.ReminderMinutesBefore, SendTime: preference.SendTime})
	}
	return result
}

func collaborationPreferenceEntity(requests []collaborationdto.NotificationPreferenceRequest) []entity.NotificationPreference {
	result := make([]entity.NotificationPreference, 0, len(requests))
	for _, req := range requests {
		result = append(result, entity.NotificationPreference{TriggerType: req.TriggerType, InAppEnabled: req.Channels.InApp, EmailEnabled: req.Channels.Email, TelegramEnabled: false, ReminderMinutesBefore: req.ReminderMinutesBefore, SendTime: req.SendTime})
	}
	return result
}
