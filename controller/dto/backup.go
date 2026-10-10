package dto

// FileItem represents a file entry with metadata
type FileItem struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	ModTime   string `json:"mod_time"`
	FullPath  string `json:"full_path"`
	IsImage   bool   `json:"is_image"`
}

// MoveSummary contains summary statistics and file lists for the backup move operation
type MoveSummary struct {
	TotalFiles        int      `json:"total_files"`
	MovedFiles        int      `json:"moved_files"`
	SkippedFiles      int      `json:"skipped_files"`
	FailedFiles       int      `json:"failed_files"`
	DurationMs        int64    `json:"duration_ms"`
	DurationFormatted string   `json:"duration_formatted"`
	MovedList         []string `json:"moved_list"`
	SkippedList       []string `json:"skipped_list"`
	FailedList        []string `json:"failed_list"`
}

// IntegrityResult contains scanning results comparing database records with files on disk
type IntegrityResult struct {
	TotalDB       int      `json:"total_db"`
	TotalDisk     int      `json:"total_disk"`
	MissingInDisk []string `json:"missing_in_disk"`
	MatchedCount  int      `json:"matched_count"`
	IsExactMatch  bool     `json:"is_exact_match"`
	AlertMessage  string   `json:"alert_message"`
}
