package model

import "PhotoVault/controller/dto"

// PersistedAppSettings represents the application configuration saved to database/ai_settings.json.
type PersistedAppSettings struct {
	AutoAnalyze    bool                     `json:"auto_analyze"`
	PrimaryModel   string                   `json:"primary_model"`
	DuplicateGuard bool                     `json:"duplicate_guard"`
	CopyWorkers    int                      `json:"copy_workers"`
	Projects       []dto.ProjectApiKeyEntry `json:"projects"`
}
