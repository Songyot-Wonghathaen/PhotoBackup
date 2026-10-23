package model

// ModelQuotaRecord holds daily or short-term quota exhaustion info for a model/key pair.
type ModelQuotaRecord struct {
	Model     string `json:"model"`
	KeyPrefix string `json:"key_prefix"`
	Reason    string `json:"reason"`
	ResetAt   string `json:"reset_at"` // RFC3339
}

// ModelQuotaCacheFile represents the persisted quota cache JSON.
type ModelQuotaCacheFile struct {
	Records map[string]ModelQuotaRecord `json:"records"` // keyPrefix|model -> record
}

// BatchItemWire represents an individual photo item decoded from a multi-photo batch response.
type BatchItemWire struct {
	Index       int      `json:"i"`
	Description string   `json:"d"`
	Tags        []string `json:"t"`
	// Fallback full key names if model uses them
	AltIndex       int      `json:"index"`
	AltDescription string   `json:"description"`
	AltTags        []string `json:"tags"`
}

// GeminiPart represents a content part in Gemini API requests and responses.
type GeminiPart struct {
	Text       string            `json:"text,omitempty"`
	Thought    bool              `json:"thought,omitempty"`
	InlineData *GeminiInlineData `json:"inline_data,omitempty"`
}

// GeminiInlineData represents encoded image bytes.
type GeminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GeminiContent represents structured message content.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiGenerationConfig represents generation parameters.
type GeminiGenerationConfig struct {
	ResponseMimeType string  `json:"response_mime_type,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
}

// GeminiRequest represents the root payload sent to Google Gemini generateContent.
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig"`
}

// GeminiResponse represents the root response from Google Gemini generateContent.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []GeminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}
