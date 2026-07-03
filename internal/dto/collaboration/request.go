package collaborationdto

type CreateCommentRequest struct {
	ResourceType   string   `json:"resource_type" binding:"required"`
	ResourceID     string   `json:"resource_id" binding:"required"`
	ParentID       *string  `json:"parent_id"`
	Body           string   `json:"body" binding:"required"`
	MentionUserIDs []string `json:"mention_user_ids"`
	AttachmentIDs  []string `json:"attachment_ids"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type UploadSignatureRequest struct {
	WorkspaceID  string `json:"workspace_id" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	ResourceID   string `json:"resource_id" binding:"required"`
	FileName     string `json:"file_name" binding:"required"`
	FileType     string `json:"file_type" binding:"required"`
	FileSize     int64  `json:"file_size" binding:"required"`
}

type CreateAttachmentRequest struct {
	WorkspaceID        string `json:"workspace_id" binding:"required"`
	ResourceType       string `json:"resource_type" binding:"required"`
	ResourceID         string `json:"resource_id" binding:"required"`
	CloudinaryPublicID string `json:"cloudinary_public_id" binding:"required"`
	URL                string `json:"url" binding:"required"`
	FileName           string `json:"file_name" binding:"required"`
	FileType           string `json:"file_type" binding:"required"`
	FileSize           int64  `json:"file_size" binding:"required"`
}

type MarkNotificationReadRequest struct {
	Read bool `json:"read"`
}

type MarkAllNotificationsReadRequest struct {
	Type *string `json:"type"`
}

type UpdateNotificationPreferencesRequest struct {
	Preferences []NotificationPreferenceRequest `json:"preferences"`
}

type NotificationPreferenceRequest struct {
	TriggerType           string                      `json:"trigger_type" binding:"required"`
	Channels              NotificationChannelsRequest `json:"channels"`
	ReminderMinutesBefore *int                        `json:"reminder_minutes_before"`
	SendTime              *string                     `json:"send_time"`
}

type NotificationChannelsRequest struct {
	InApp    bool `json:"in_app"`
	Email    bool `json:"email"`
	Telegram bool `json:"telegram"`
}
