package entity

import "time"

type Comment struct {
	ID             string
	WorkspaceID    string
	ResourceType   string
	ResourceID     string
	ParentID       *string
	Body           string
	AuthorID       string
	AuthorName     string
	Mentions       []Mention
	Attachments    []Attachment
	MentionUserIDs []string
	AttachmentIDs  []string
	EditedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Mention struct {
	UserID   string
	FullName string
}

type Attachment struct {
	ID                 string
	WorkspaceID        string
	ResourceType       string
	ResourceID         string
	CloudinaryPublicID string
	URL                string
	ThumbnailURL       string
	FileName           string
	FileType           string
	FileSize           int64
	UploadedBy         string
	UploadedByName     string
	CreatedAt          time.Time
}

type UploadSignature struct {
	CloudName      string
	APIKey         string
	Timestamp      int64
	Signature      string
	Folder         string
	UploadURL      string
	PublicIDPrefix string
}

type Notification struct {
	ID          string
	UserID      string
	WorkspaceID *string
	Type        string
	Title       string
	Body        string
	Payload     map[string]any
	ReadAt      *time.Time
	CreatedAt   time.Time
}

type NotificationPreference struct {
	TriggerType           string
	InAppEnabled          bool
	EmailEnabled          bool
	TelegramEnabled       bool
	ReminderMinutesBefore *int
	SendTime              *string
}
