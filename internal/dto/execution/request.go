package executiondto

type CreateTimeboxRequest struct {
	WorkspaceID    string   `json:"workspace_id" binding:"required"`
	TaskID         *string  `json:"task_id"`
	OwnerID        string   `json:"owner_id" binding:"required"`
	CategoryID     *string  `json:"category_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ScheduledStart string   `json:"scheduled_start" binding:"required"`
	ScheduledEnd   string   `json:"scheduled_end" binding:"required"`
	IsBuffer       bool     `json:"is_buffer"`
	ParticipantIDs []string `json:"participant_ids"`
}

type UpdateTimeboxRequest struct {
	TaskID         *string `json:"task_id"`
	OwnerID        *string `json:"owner_id"`
	CategoryID     *string `json:"category_id"`
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	ScheduledStart *string `json:"scheduled_start"`
	ScheduledEnd   *string `json:"scheduled_end"`
	Status         *string `json:"status"`
	IsBuffer       *bool   `json:"is_buffer"`
}

type StartTimerRequest struct {
	StartMode string `json:"start_mode"`
}

type PauseTimerRequest struct {
	Reason string `json:"reason"`
}

type CompleteTimerRequest struct {
	Note string `json:"note"`
}

type SkipTimerRequest struct {
	Reason string `json:"reason"`
}

type CreateTimeLogRequest struct {
	StartedAt string `json:"started_at" binding:"required"`
	EndedAt   string `json:"ended_at" binding:"required"`
	Source    string `json:"source"`
	Note      string `json:"note"`
}
