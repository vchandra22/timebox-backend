package collaboration

import (
	"context"
	"errors"

	"timebox-backend/internal/entity"
)

var (
	ErrNotFound = errors.New("collaboration resource not found")
	ErrConflict = errors.New("collaboration resource conflict")
)

type NotificationFilter struct {
	UserID string
	Status string
	Type   string
	Limit  int
	Offset int
}

type Repository interface {
	ResourceWorkspace(ctx context.Context, resourceType, resourceID string) (string, error)
	ListComments(ctx context.Context, resourceType, resourceID string, parentID *string) ([]entity.Comment, error)
	CreateComment(ctx context.Context, comment entity.Comment) (entity.Comment, error)
	FindComment(ctx context.Context, id string) (entity.Comment, error)
	UpdateComment(ctx context.Context, comment entity.Comment) (entity.Comment, error)
	DeleteComment(ctx context.Context, id string) error
	GenerateMentionNotifications(ctx context.Context, comment entity.Comment) error
	CreateAttachment(ctx context.Context, attachment entity.Attachment) (entity.Attachment, error)
	ListAttachments(ctx context.Context, resourceType, resourceID string) ([]entity.Attachment, error)
	FindAttachment(ctx context.Context, id string) (entity.Attachment, error)
	DeleteAttachment(ctx context.Context, id string) error
	ListNotifications(ctx context.Context, filter NotificationFilter) ([]entity.Notification, int, error)
	MarkNotificationRead(ctx context.Context, userID, id string, read bool) (entity.Notification, error)
	MarkAllNotificationsRead(ctx context.Context, userID string, notificationType *string) (int, error)
	GetPreferences(ctx context.Context, userID string) ([]entity.NotificationPreference, error)
	UpdatePreferences(ctx context.Context, userID string, preferences []entity.NotificationPreference) (int, error)
}
