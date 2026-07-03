package database

import "time"

type CommentRow struct {
	ID           string     `db:"id"`
	WorkspaceID  string     `db:"workspace_id"`
	ResourceType string     `db:"resource_type"`
	ResourceID   string     `db:"resource_id"`
	ParentID     *string    `db:"parent_id"`
	Body         string     `db:"body"`
	AuthorID     string     `db:"author_id"`
	AuthorName   string     `db:"author_name"`
	EditedAt     *time.Time `db:"edited_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type MentionRow struct {
	UserID   string `db:"user_id"`
	FullName string `db:"full_name"`
}

type AttachmentRow struct {
	ID                 string    `db:"id"`
	WorkspaceID        string    `db:"workspace_id"`
	ResourceType       string    `db:"resource_type"`
	ResourceID         string    `db:"resource_id"`
	CloudinaryPublicID string    `db:"cloudinary_public_id"`
	URL                string    `db:"url"`
	FileName           string    `db:"file_name"`
	FileType           string    `db:"file_type"`
	FileSize           int64     `db:"file_size"`
	UploadedBy         string    `db:"uploaded_by"`
	UploadedByName     string    `db:"uploaded_by_name"`
	CreatedAt          time.Time `db:"created_at"`
}

type NotificationRow struct {
	ID          string     `db:"id"`
	UserID      string     `db:"user_id"`
	WorkspaceID *string    `db:"workspace_id"`
	Type        string     `db:"type"`
	Title       string     `db:"title"`
	Body        string     `db:"body"`
	Payload     []byte     `db:"payload"`
	ReadAt      *time.Time `db:"read_at"`
	CreatedAt   time.Time  `db:"created_at"`
}

type PreferenceRow struct {
	TriggerType           string  `db:"trigger_type"`
	InAppEnabled          bool    `db:"in_app_enabled"`
	EmailEnabled          bool    `db:"email_enabled"`
	TelegramEnabled       bool    `db:"telegram_enabled"`
	ReminderMinutesBefore *int    `db:"reminder_minutes_before"`
	SendTime              *string `db:"send_time"`
}
