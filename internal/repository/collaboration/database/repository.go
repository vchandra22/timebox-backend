package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"timebox-backend/internal/config"
	"timebox-backend/internal/entity"
	collaborationrepo "timebox-backend/internal/repository/collaboration"
	"timebox-backend/internal/repository/dbexecutor"
)

type Repository struct {
	db         config.PostgreSQL
	dbExecutor *dbexecutor.Executor
}

func NewRepository(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{db: db, dbExecutor: dbExecutor}
}

func (r *Repository) ResourceWorkspace(ctx context.Context, resourceType, resourceID string) (string, error) {
	var workspaceID string
	err := collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &workspaceID, QueryResourceWorkspace, resourceType, resourceID))
	return workspaceID, err
}

func (r *Repository) ListComments(ctx context.Context, resourceType, resourceID string, parentID *string) ([]entity.Comment, error) {
	var rows []CommentRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListComments, resourceType, resourceID, parentID); err != nil {
		return nil, collaborationError(err)
	}
	comments := make([]entity.Comment, 0, len(rows))
	for _, row := range rows {
		comment := row.toEntity()
		comment.Mentions = r.commentMentions(ctx, comment.ID)
		comment.Attachments = r.commentAttachments(ctx, comment.ID)
		comments = append(comments, comment)
	}
	return comments, nil
}

func (r *Repository) CreateComment(ctx context.Context, comment entity.Comment) (entity.Comment, error) {
	tx, err := r.db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return entity.Comment{}, err
	}
	defer tx.Rollback()

	var row CommentRow
	err = collaborationError(r.dbExecutor.Get(ctx, tx, &row, QueryCreateComment, comment.WorkspaceID, comment.ResourceType, comment.ResourceID, comment.ParentID, comment.Body, comment.AuthorID))
	if err != nil {
		return entity.Comment{}, err
	}
	for _, userID := range comment.MentionUserIDs {
		if err := r.dbExecutor.Exec(ctx, tx, QueryInsertMention, row.ID, userID, comment.WorkspaceID); err != nil {
			return entity.Comment{}, collaborationError(err)
		}
	}
	for _, attachmentID := range comment.AttachmentIDs {
		if err := r.dbExecutor.Exec(ctx, tx, QueryMoveAttachmentsToComment, attachmentID, row.ID, comment.WorkspaceID, comment.AuthorID); err != nil {
			return entity.Comment{}, collaborationError(err)
		}
	}
	created := row.toEntity()
	return created, tx.Commit()
}

func (r *Repository) FindComment(ctx context.Context, id string) (entity.Comment, error) {
	var row CommentRow
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindComment, id); err != nil {
		return entity.Comment{}, collaborationError(err)
	}
	comment := row.toEntity()
	comment.Mentions = r.commentMentions(ctx, comment.ID)
	comment.Attachments = r.commentAttachments(ctx, comment.ID)
	return comment, nil
}

func (r *Repository) UpdateComment(ctx context.Context, comment entity.Comment) (entity.Comment, error) {
	var row CommentRow
	err := collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateComment, comment.ID, comment.Body))
	return row.toEntity(), err
}

func (r *Repository) DeleteComment(ctx context.Context, id string) error {
	var deletedID string
	return collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteComment, id))
}

func (r *Repository) GenerateMentionNotifications(ctx context.Context, comment entity.Comment) error {
	return collaborationError(r.dbExecutor.Exec(ctx, r.db.Conn, QueryCreateMentionNotification, comment.ID))
}

func (r *Repository) CreateAttachment(ctx context.Context, attachment entity.Attachment) (entity.Attachment, error) {
	var row AttachmentRow
	err := collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateAttachment, attachment.WorkspaceID, attachment.ResourceType, attachment.ResourceID, attachment.CloudinaryPublicID, attachment.URL, attachment.FileName, attachment.FileType, attachment.FileSize, attachment.UploadedBy))
	return row.toEntity(), err
}

func (r *Repository) ListAttachments(ctx context.Context, resourceType, resourceID string) ([]entity.Attachment, error) {
	var rows []AttachmentRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListAttachments, resourceType, resourceID); err != nil {
		return nil, collaborationError(err)
	}
	attachments := make([]entity.Attachment, 0, len(rows))
	for _, row := range rows {
		attachments = append(attachments, row.toEntity())
	}
	return attachments, nil
}

func (r *Repository) FindAttachment(ctx context.Context, id string) (entity.Attachment, error) {
	var row AttachmentRow
	err := collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindAttachment, id))
	return row.toEntity(), err
}

func (r *Repository) DeleteAttachment(ctx context.Context, id string) error {
	var deletedID string
	return collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteAttachment, id))
}

func (r *Repository) ListNotifications(ctx context.Context, filter collaborationrepo.NotificationFilter) ([]entity.Notification, int, error) {
	status := filter.Status
	if status == "" {
		status = "unread"
	}
	var total int
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &total, QueryCountNotifications, filter.UserID, status, filter.Type); err != nil {
		return nil, 0, collaborationError(err)
	}
	var rows []NotificationRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListNotifications, filter.UserID, status, filter.Type, filter.Limit, filter.Offset); err != nil {
		return nil, 0, collaborationError(err)
	}
	notifications := make([]entity.Notification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, row.toEntity())
	}
	return notifications, total, nil
}

func (r *Repository) MarkNotificationRead(ctx context.Context, userID, id string, read bool) (entity.Notification, error) {
	var row NotificationRow
	err := collaborationError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryMarkNotificationRead, userID, id, read))
	return row.toEntity(), err
}

func (r *Repository) MarkAllNotificationsRead(ctx context.Context, userID string, notificationType *string) (int, error) {
	result, err := r.db.Conn.ExecContext(ctx, QueryMarkAllNotificationsRead, userID, notificationType)
	if err != nil {
		return 0, collaborationError(err)
	}
	count, err := result.RowsAffected()
	return int(count), err
}

func (r *Repository) GetPreferences(ctx context.Context, userID string) ([]entity.NotificationPreference, error) {
	var rows []PreferenceRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryGetPreferences, userID); err != nil {
		return nil, collaborationError(err)
	}
	preferences := make([]entity.NotificationPreference, 0, len(rows))
	for _, row := range rows {
		preferences = append(preferences, row.toEntity())
	}
	return preferences, nil
}

func (r *Repository) UpdatePreferences(ctx context.Context, userID string, preferences []entity.NotificationPreference) (int, error) {
	for _, preference := range preferences {
		if err := r.dbExecutor.Exec(ctx, r.db.Conn, QueryUpsertPreference, userID, preference.TriggerType, preference.InAppEnabled, preference.EmailEnabled, preference.ReminderMinutesBefore, preference.SendTime); err != nil {
			return 0, collaborationError(err)
		}
	}
	return len(preferences), nil
}

func (r *Repository) commentMentions(ctx context.Context, commentID string) []entity.Mention {
	var rows []MentionRow
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryListMentions, commentID); err != nil {
		return nil
	}
	mentions := make([]entity.Mention, 0, len(rows))
	for _, row := range rows {
		mentions = append(mentions, row.toEntity())
	}
	return mentions
}

func (r *Repository) commentAttachments(ctx context.Context, commentID string) []entity.Attachment {
	attachments, _ := r.ListAttachments(ctx, "comment", commentID)
	return attachments
}

func collaborationError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return collaborationrepo.ErrNotFound
	}
	return err
}

func (r CommentRow) toEntity() entity.Comment {
	return entity.Comment{ID: r.ID, WorkspaceID: r.WorkspaceID, ResourceType: r.ResourceType, ResourceID: r.ResourceID, ParentID: r.ParentID, Body: r.Body, AuthorID: r.AuthorID, AuthorName: r.AuthorName, EditedAt: r.EditedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (r MentionRow) toEntity() entity.Mention {
	return entity.Mention{UserID: r.UserID, FullName: r.FullName}
}

func (r AttachmentRow) toEntity() entity.Attachment {
	return entity.Attachment{ID: r.ID, WorkspaceID: r.WorkspaceID, ResourceType: r.ResourceType, ResourceID: r.ResourceID, CloudinaryPublicID: r.CloudinaryPublicID, URL: r.URL, ThumbnailURL: thumbnailURL(r.URL), FileName: r.FileName, FileType: r.FileType, FileSize: r.FileSize, UploadedBy: r.UploadedBy, UploadedByName: r.UploadedByName, CreatedAt: r.CreatedAt}
}

func (r NotificationRow) toEntity() entity.Notification {
	payload := map[string]any{}
	_ = json.Unmarshal(r.Payload, &payload)
	return entity.Notification{ID: r.ID, UserID: r.UserID, WorkspaceID: r.WorkspaceID, Type: r.Type, Title: r.Title, Body: r.Body, Payload: payload, ReadAt: r.ReadAt, CreatedAt: r.CreatedAt}
}

func (r PreferenceRow) toEntity() entity.NotificationPreference {
	return entity.NotificationPreference{TriggerType: r.TriggerType, InAppEnabled: r.InAppEnabled, EmailEnabled: r.EmailEnabled, TelegramEnabled: r.TelegramEnabled, ReminderMinutesBefore: r.ReminderMinutesBefore, SendTime: r.SendTime}
}

func thumbnailURL(url string) string {
	return url
}
