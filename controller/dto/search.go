package dto

// SearchParams carries gallery search filters for Features 15, 16, 17
type SearchParams struct {
	Query    string `json:"query"`
	TagIDs   []int32 `json:"tag_ids"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// PhotoListItem represents a photo card in gallery search results
type PhotoListItem struct {
	ID          int32    `json:"id"`
	FileName    string   `json:"file_name"`
	FullPath    string   `json:"full_path"`
	ThumbURL    string   `json:"thumb_url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	BackedUpAt  string   `json:"backed_up_at"`
}

// SearchResult contains a page of photo cards plus paging metadata
type SearchResult struct {
	Items      []PhotoListItem `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// TagCount represents a tag name with its photo count
type TagCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// TagItem represents a tag with its ID and photo count (clickable in Tag Cloud).
// Source is only populated on the photo detail endpoint ('ai' or 'user').
type TagItem struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Source string `json:"source"`
}

// PhotoDetail contains full information of a single photo for the detail page
type PhotoDetail struct {
	ID              int32     `json:"id"`
	FileName        string    `json:"file_name"`
	DestinationPath string    `json:"destination_path"`
	FullPath        string    `json:"full_path"`
	ThumbURL        string    `json:"thumb_url"`
	Description     string    `json:"description"`
	Tags            []TagItem `json:"tags"`
	SizeBytes       int32     `json:"size_bytes"`
	Sha256          string    `json:"sha256"`
	Status          string    `json:"status"`
	BackedUpAt      string    `json:"backed_up_at"`
}
