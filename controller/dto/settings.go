package dto

// ProjectApiKeyEntry represents a single Google Cloud Project and its Gemini API Key.
type ProjectApiKeyEntry struct {
	ID            string `json:"id"`
	ProjectName   string `json:"project_name"`
	ProjectNumber string `json:"project_number"`
	ApiKey        string `json:"api_key"`
	Enabled       bool   `json:"enabled"`
	Status        string `json:"status"`         // "active", "missing_key", "disabled", "cooldown", "daily_limit"
	CooldownCount int    `json:"cooldown_count"` // number of models currently cooling down on this key
}

// AppSettingsDTO represents the full application settings and live AI routing state.
type AppSettingsDTO struct {
	AutoAnalyze       bool                 `json:"auto_analyze"`
	PrimaryModel      string               `json:"primary_model"`
	DuplicateGuard    bool                 `json:"duplicate_guard"`
	CopyWorkers       int                  `json:"copy_workers"`
	ActiveKeyCount    int                  `json:"active_key_count"`
	TotalProjectCount int                  `json:"total_project_count"`
	PrimaryRoutes     int                  `json:"primary_routes"`
	TotalRoutes       int                  `json:"total_routes"`
	AllocatedWorkers  int                  `json:"allocated_workers"`
	BatchSize         int                  `json:"batch_size"`
	EstimatedRPM      int                  `json:"estimated_rpm"`
	EstimatedRPD      int                  `json:"estimated_rpd"`
	Projects          []ProjectApiKeyEntry `json:"projects"`
}
