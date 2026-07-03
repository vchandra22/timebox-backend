package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"timebox-backend/internal/entity"
	collaborationrepo "timebox-backend/internal/repository/collaboration"
	workspacerepo "timebox-backend/internal/repository/workspace"
)

const (
	NotificationTypeMention         = "mention"
	NotificationTypeTimeboxReminder = "timebox_reminder"
	NotificationTypeDailySummary    = "daily_summary"

	maxUploadSize = 20 * 1024 * 1024
)

var (
	ErrCollaborationNotFound = errors.New("collaboration resource not found")
	ErrCollaborationConflict = errors.New("collaboration resource conflict")
	ErrInvalidResourceType   = errors.New("invalid resource type")
	ErrInvalidUpload         = errors.New("invalid upload")
	ErrCloudinaryNotReady    = errors.New("cloudinary config missing")
)

type CollaborationService struct {
	repo          collaborationrepo.Repository
	workspaceRepo workspacerepo.Repository
	options       CollaborationOptions
}

type CollaborationOptions struct {
	CloudName string
	APIKey    string
	APISecret string
}

func newCollaborationService(repo collaborationrepo.Repository, workspaceRepo workspacerepo.Repository, options CollaborationOptions) *CollaborationService {
	return &CollaborationService{repo: repo, workspaceRepo: workspaceRepo, options: options}
}

func (s *CollaborationService) ListComments(ctx context.Context, actorID, resourceType, resourceID string, parentID *string) ([]entity.Comment, error) {
	workspaceID, err := s.resourceWorkspace(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	comments, err := s.repo.ListComments(ctx, resourceType, resourceID, parentID)
	return comments, collaborationError(err)
}

func (s *CollaborationService) CreateComment(ctx context.Context, actorID string, comment entity.Comment) (entity.Comment, error) {
	workspaceID, err := s.resourceWorkspace(ctx, comment.ResourceType, comment.ResourceID)
	if err != nil {
		return entity.Comment{}, err
	}
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.Comment{}, err
	}
	comment.WorkspaceID = workspaceID
	comment.AuthorID = actorID
	created, err := s.repo.CreateComment(ctx, comment)
	if err != nil {
		return entity.Comment{}, collaborationError(err)
	}
	_ = s.repo.GenerateMentionNotifications(ctx, created)
	return created, nil
}

func (s *CollaborationService) UpdateComment(ctx context.Context, actorID string, patch entity.Comment) (entity.Comment, error) {
	current, err := s.repo.FindComment(ctx, patch.ID)
	if err != nil {
		return entity.Comment{}, collaborationError(err)
	}
	if err := s.requireCommentAuthor(ctx, current, actorID); err != nil {
		return entity.Comment{}, err
	}
	current.Body = patch.Body
	updated, err := s.repo.UpdateComment(ctx, current)
	return updated, collaborationError(err)
}

func (s *CollaborationService) DeleteComment(ctx context.Context, actorID, id string) error {
	comment, err := s.repo.FindComment(ctx, id)
	if err != nil {
		return collaborationError(err)
	}
	if err := s.requireCommentAuthor(ctx, comment, actorID); err != nil {
		return err
	}
	return collaborationError(s.repo.DeleteComment(ctx, id))
}

func (s *CollaborationService) GenerateUploadSignature(ctx context.Context, actorID string, req entity.Attachment) (entity.UploadSignature, error) {
	if s.options.CloudName == "" || s.options.APIKey == "" || s.options.APISecret == "" {
		return entity.UploadSignature{}, ErrCloudinaryNotReady
	}
	if err := validateUpload(req.FileType, req.FileSize); err != nil {
		return entity.UploadSignature{}, err
	}
	workspaceID, err := s.resourceWorkspace(ctx, req.ResourceType, req.ResourceID)
	if err != nil {
		return entity.UploadSignature{}, err
	}
	if workspaceID != req.WorkspaceID {
		return entity.UploadSignature{}, ErrForbidden
	}
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.UploadSignature{}, err
	}
	timestamp := time.Now().Unix()
	folder := fmt.Sprintf("timebox-space/development/%s/%s", req.WorkspaceID, req.ResourceType)
	publicIDPrefix := fmt.Sprintf("%s/%s", req.ResourceType, req.ResourceID)
	signature := s.cloudinarySignature(map[string]string{
		"folder":    folder,
		"public_id": publicIDPrefix,
		"timestamp": fmt.Sprint(timestamp),
	})
	return entity.UploadSignature{CloudName: s.options.CloudName, APIKey: s.options.APIKey, Timestamp: timestamp, Signature: signature, Folder: folder, UploadURL: fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/auto/upload", s.options.CloudName), PublicIDPrefix: publicIDPrefix}, nil
}

func (s *CollaborationService) CreateAttachment(ctx context.Context, actorID string, attachment entity.Attachment) (entity.Attachment, error) {
	if err := validateUpload(attachment.FileType, attachment.FileSize); err != nil {
		return entity.Attachment{}, err
	}
	workspaceID, err := s.resourceWorkspace(ctx, attachment.ResourceType, attachment.ResourceID)
	if err != nil {
		return entity.Attachment{}, err
	}
	if workspaceID != attachment.WorkspaceID {
		return entity.Attachment{}, ErrForbidden
	}
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return entity.Attachment{}, err
	}
	attachment.UploadedBy = actorID
	created, err := s.repo.CreateAttachment(ctx, attachment)
	return created, collaborationError(err)
}

func (s *CollaborationService) ListAttachments(ctx context.Context, actorID, resourceType, resourceID string) ([]entity.Attachment, error) {
	workspaceID, err := s.resourceWorkspace(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWorkspaceMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	attachments, err := s.repo.ListAttachments(ctx, resourceType, resourceID)
	return attachments, collaborationError(err)
}

func (s *CollaborationService) DeleteAttachment(ctx context.Context, actorID, id string) (entity.Attachment, error) {
	attachment, err := s.repo.FindAttachment(ctx, id)
	if err != nil {
		return entity.Attachment{}, collaborationError(err)
	}
	member, err := s.requireWorkspaceMember(ctx, attachment.WorkspaceID, actorID)
	if err != nil {
		return entity.Attachment{}, err
	}
	if attachment.UploadedBy != actorID && member.Role != WorkspaceRoleOwner && member.Role != WorkspaceRoleAdmin {
		return entity.Attachment{}, ErrForbidden
	}
	return attachment, collaborationError(s.repo.DeleteAttachment(ctx, id))
}

func (s *CollaborationService) ListNotifications(ctx context.Context, userID, status, notificationType string, page, limit int) ([]entity.Notification, int, error) {
	notifications, total, err := s.repo.ListNotifications(ctx, collaborationrepo.NotificationFilter{UserID: userID, Status: status, Type: notificationType, Limit: limit, Offset: (page - 1) * limit})
	return notifications, total, collaborationError(err)
}

func (s *CollaborationService) MarkNotificationRead(ctx context.Context, userID, id string, read bool) (entity.Notification, error) {
	notification, err := s.repo.MarkNotificationRead(ctx, userID, id, read)
	return notification, collaborationError(err)
}

func (s *CollaborationService) MarkAllNotificationsRead(ctx context.Context, userID string, notificationType *string) (int, error) {
	count, err := s.repo.MarkAllNotificationsRead(ctx, userID, notificationType)
	return count, collaborationError(err)
}

func (s *CollaborationService) GetPreferences(ctx context.Context, userID string) ([]entity.NotificationPreference, error) {
	preferences, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, collaborationError(err)
	}
	if len(preferences) > 0 {
		return preferences, nil
	}
	return defaultNotificationPreferences(), nil
}

func (s *CollaborationService) UpdatePreferences(ctx context.Context, userID string, preferences []entity.NotificationPreference) (int, error) {
	for i := range preferences {
		preferences[i].TelegramEnabled = false
	}
	count, err := s.repo.UpdatePreferences(ctx, userID, preferences)
	return count, collaborationError(err)
}

func (s *CollaborationService) resourceWorkspace(ctx context.Context, resourceType, resourceID string) (string, error) {
	if !validResourceType(resourceType) {
		return "", ErrInvalidResourceType
	}
	workspaceID, err := s.repo.ResourceWorkspace(ctx, resourceType, resourceID)
	return workspaceID, collaborationError(err)
}

func (s *CollaborationService) requireCommentAuthor(ctx context.Context, comment entity.Comment, actorID string) error {
	member, err := s.requireWorkspaceMember(ctx, comment.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if comment.AuthorID == actorID || member.Role == WorkspaceRoleOwner || member.Role == WorkspaceRoleAdmin {
		return nil
	}
	return ErrForbidden
}

func (s *CollaborationService) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) (entity.WorkspaceMember, error) {
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

func (s *CollaborationService) cloudinarySignature(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	sum := sha1.Sum([]byte(strings.Join(parts, "&") + s.options.APISecret))
	return hex.EncodeToString(sum[:])
}

func validateUpload(fileType string, fileSize int64) error {
	if fileSize < 1 || fileSize > maxUploadSize {
		return ErrInvalidUpload
	}
	switch fileType {
	case "image/png", "image/jpeg", "image/webp", "application/pdf":
		return nil
	default:
		return ErrInvalidUpload
	}
}

func validResourceType(resourceType string) bool {
	return resourceType == "goal" || resourceType == "task" || resourceType == "timebox" || resourceType == "comment"
}

func defaultNotificationPreferences() []entity.NotificationPreference {
	ten := 10
	sendTime := "20:00"
	return []entity.NotificationPreference{
		{TriggerType: NotificationTypeMention, InAppEnabled: true, EmailEnabled: false, TelegramEnabled: false},
		{TriggerType: NotificationTypeTimeboxReminder, InAppEnabled: true, EmailEnabled: false, TelegramEnabled: false, ReminderMinutesBefore: &ten},
		{TriggerType: NotificationTypeDailySummary, InAppEnabled: true, EmailEnabled: false, TelegramEnabled: false, SendTime: &sendTime},
	}
}

func collaborationError(err error) error {
	if errors.Is(err, collaborationrepo.ErrNotFound) {
		return ErrCollaborationNotFound
	}
	if errors.Is(err, collaborationrepo.ErrConflict) {
		return ErrCollaborationConflict
	}
	return err
}
