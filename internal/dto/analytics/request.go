package analyticsdto

type CreateSavedViewRequest struct {
	WorkspaceID  string         `json:"workspace_id" binding:"required"`
	Name         string         `json:"name" binding:"required"`
	ResourceType string         `json:"resource_type" binding:"required"`
	FilterJSON   map[string]any `json:"filter_json"`
	Shared       bool           `json:"shared"`
}

type UpdateSavedViewRequest struct {
	Name       *string        `json:"name"`
	FilterJSON map[string]any `json:"filter_json"`
	Shared     *bool          `json:"shared"`
}
