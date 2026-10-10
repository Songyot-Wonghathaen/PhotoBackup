package dto

// VisionResult represents the AI analysis result containing image description and extracted tags
type VisionResult struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
