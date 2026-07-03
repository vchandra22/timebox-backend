package collaborationdto

import "time"

type CommentResponse struct {
	ID           string               `json:"id"`
	ResourceType string               `json:"resource_type"`
	ResourceID   string               `json:"resource_id"`
	ParentID     *string              `json:"parent_id,omitempty"`
	Body         string               `json:"body"`
	Mentions     []MentionResponse    `json:"mentions,omitempty"`
	Author       *AuthorResponse      `json:"author,omitempty"`
	Attachments  []AttachmentResponse `json:"attachments,omitempty"`
	EditedAt     *time.Time           `json:"edited_at,omitempty"`
	CreatedAt    *time.Time           `json:"created_at,omitempty"`
}

type MentionResponse struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
}

type AuthorResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type UploadSignatureResponse struct {
	CloudName      string `json:"cloud_name"`
	APIKey         string `json:"api_key"`
	Timestamp      int64  `json:"timestamp"`
	Signature      string `json:"signature"`
	Folder         string `json:"folder"`
	UploadURL      string `json:"upload_url"`
	PublicIDPrefix string `json:"public_id_prefix"`
}

type AttachmentResponse struct {
	ID           string          `json:"id"`
	ResourceType string          `json:"resource_type,omitempty"`
	ResourceID   string          `json:"resource_id,omitempty"`
	URL          string          `json:"url"`
	ThumbnailURL string          `json:"thumbnail_url,omitempty"`
	FileName     string          `json:"file_name"`
	FileType     string          `json:"file_type"`
	FileSize     int64           `json:"file_size"`
	UploadedBy   *AuthorResponse `json:"uploaded_by,omitempty"`
	CreatedAt    *time.Time      `json:"created_at,omitempty"`
}

type AttachmentDeleteResponse struct {
	ID                string `json:"id"`
	DeleteAssetQueued bool   `json:"delete_asset_queued"`
}

type NotificationResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Payload   map[string]any `json:"payload"`
	ReadAt    *time.Time     `json:"read_at"`
	CreatedAt time.Time      `json:"created_at"`
}

type NotificationReadResponse struct {
	ID     string     `json:"id"`
	ReadAt *time.Time `json:"read_at"`
}

type MarkAllReadResponse struct {
	UpdatedCount int `json:"updated_count"`
}

type NotificationPreferenceResponse struct {
	TriggerType           string                       `json:"trigger_type"`
	Channels              NotificationChannelsResponse `json:"channels"`
	ReminderMinutesBefore *int                         `json:"reminder_minutes_before,omitempty"`
	SendTime              *string                      `json:"send_time,omitempty"`
}

type NotificationChannelsResponse struct {
	InApp    bool `json:"in_app"`
	Email    bool `json:"email"`
	Telegram bool `json:"telegram"`
}

type NotificationPreferencesUpdateResponse struct {
	UpdatedCount int `json:"updated_count"`
}
