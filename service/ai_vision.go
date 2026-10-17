package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"PhotoVault/controller/dto"
	"PhotoVault/utils"

	"github.com/joho/godotenv"
)

type VisionResult = dto.VisionResult

type AiVisionService interface {
	AnalyzePhoto(ctx context.Context, filePath string) (*VisionResult, error)
	AnalyzePhotosBatch(ctx context.Context, filePaths []string) ([]*VisionResult, error)
	IsAvailable() bool
}

const analyzeBatchSize = 8

type geminiVisionService struct {
	apiKey     string
	httpClient *http.Client
	modelName  string
}

func NewAiVisionService() AiVisionService {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	key := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	// Clean surrounding quotes if any
	key = strings.Trim(key, "\"'")

	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if model == "" {
		model = "gemini-3.1-flash-lite"
	}

	return &geminiVisionService{
		apiKey: key,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		modelName: model,
	}
}

func (s *geminiVisionService) IsAvailable() bool {
	return s.apiKey != ""
}

func (s *geminiVisionService) AnalyzePhoto(ctx context.Context, filePath string) (*VisionResult, error) {
	filename := filepath.Base(filePath)

	if !s.IsAvailable() {
		err := errors.New("GEMINI_API_KEY is not configured in .env or environment")
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] ข้ามการวิเคราะห์ %s: %v", filename, err)
		return nil, err
	}

	start := time.Now()
	log.Printf("[AI-Vision] 🚀 เริ่มต้นการวิเคราะห์ภาพ: %s", filename)

	// Step 1: Prepare image bytes (use thumbnail generation to optimize latency and bandwidth)
	var imgBytes []byte
	mimeType := "image/jpeg"

	thumbBytes, err := utils.GetThumbnailBytes(filePath, 256)
	if err == nil && len(thumbBytes) > 0 {
		imgBytes = thumbBytes
		log.Printf("[AI-Vision] 📦 สร้าง Thumbnail สำเร็จ: %.1f KB (256px, ลดขนาดเพื่อส่ง AI)", float64(len(imgBytes))/1024.0)
	} else {
		// Fallback to reading file directly
		rawBytes, readErr := os.ReadFile(filePath)
		if readErr != nil {
			log.Printf("[AI-Vision] ⚠️ [FALLBACK] ไม่สามารถอ่านไฟล์ภาพ %s: %v", filename, readErr)
			return nil, fmt.Errorf("read image file failed: %w", readErr)
		}
		imgBytes = rawBytes

		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".png":
			mimeType = "image/png"
		case ".webp":
			mimeType = "image/webp"
		default:
			mimeType = "image/jpeg"
		}
		log.Printf("[AI-Vision] 📦 โหลดภาพต้นฉบับ: %.1f KB", float64(len(imgBytes))/1024.0)
	}

	// Step 2: Encode to Base64
	encoded := base64.StdEncoding.EncodeToString(imgBytes)

	// Step 3: Construct Prompt with Structured Output Request
	prompt := "Analyze this photo carefully. Provide: " +
		"1) A concise natural description of the image content in Thai (1-2 sentences). " +
		"2) An array of 3-6 relevant, specific keyword tags in Thai (nouns, activities, objects, or scenery). " +
		"Respond strictly with valid JSON having exactly two keys: 'description' (string) and 'tags' (array of strings)."

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
					{
						InlineData: &geminiInlineData{
							MimeType: mimeType,
							Data:     encoded,
						},
					},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.2,
		},
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.modelName, s.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	log.Printf("[AI-Vision] 🌐 กำลังส่งคำขอไปยัง Google Gemini (%s)...", s.modelName)

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] การเชื่อมต่อ Gemini API ล้มเหลวสำหรับ %s: %v", filename, err)
		return nil, fmt.Errorf("gemini api network error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] ไม่สามารถ parse response จาก Gemini: %s", string(respBody))
		return nil, fmt.Errorf("unmarshal gemini response error: %w", err)
	}

	if geminiResp.Error != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] Gemini API ส่ง Error (%d): %s", geminiResp.Error.Code, geminiResp.Error.Message)
		return nil, fmt.Errorf("gemini api error [%d]: %s", geminiResp.Error.Code, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no content candidate returned from gemini")
	}

	rawText := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	// Clean markdown block if present (e.g., ```json ... ```)
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	var result VisionResult
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] JSON format ไม่ถูกต้อง: %s", rawText)
		return nil, fmt.Errorf("parse structured json error: %w", err)
	}

	duration := time.Since(start)
	log.Printf("[AI-Vision] ✅ วิเคราะห์สำเร็จใน %v!", duration)
	log.Printf("            ├── คำอธิบาย: %q", result.Description)
	log.Printf("            └── แท็ก (%d): %v", len(result.Tags), result.Tags)

	return &result, nil
}

// batchResultItem mirrors VisionResult but with an explicit index for ordering
type batchResultItem struct {
	Index       int      `json:"index"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// AnalyzePhotosBatch sends multiple images in a single Gemini request and returns
// one VisionResult per input path, in the same order as filePaths.
//
// If the response is missing items or fails to parse, the method falls back to
// analyzing those specific photos one-by-one so no image is silently dropped.
func (s *geminiVisionService) AnalyzePhotosBatch(ctx context.Context, filePaths []string) ([]*VisionResult, error) {
	n := len(filePaths)
	if n == 0 {
		return nil, nil
	}

	if !s.IsAvailable() {
		err := errors.New("GEMINI_API_KEY is not configured in .env or environment")
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] ข้าม batch: %v", err)
		return nil, err
	}

	start := time.Now()
	log.Printf("[AI-Vision] 🚀 เริ่ม batch วิเคราะห์ %d รูป (ส่งทีเดียว)", n)

	// Step 1: Build parts — prompt first, then one image part per photo in order
	prompt := fmt.Sprintf(
		"You are given %d photos, listed IN ORDER as image parts 1..%d. "+
			"For EACH photo, provide a concise natural description in Thai (1-2 sentences) "+
			"and 3-6 relevant Thai tags (nouns, activities, objects, scenery). "+
			"Respond strictly with a JSON ARRAY of exactly %d objects, each shaped as "+
			"{\"index\": <1-based int>, \"description\": \"...\", \"tags\": [\"...\"]}, "+
			"ordered to match the input order. Do not skip any photo.", n, n, n)

	parts := []geminiPart{{Text: prompt}}
	type imgEntry struct {
		path     string
		mimeType string
		data     string
		ok       bool
	}
	entries := make([]imgEntry, n)

	for i, p := range filePaths {
		thumbBytes, err := utils.GetThumbnailBytes(p, 256)
		mimeType := "image/jpeg"
		var imgBytes []byte

		if err == nil && len(thumbBytes) > 0 {
			imgBytes = thumbBytes
		} else {
			rawBytes, readErr := os.ReadFile(p)
			if readErr != nil {
				log.Printf("[AI-Vision] ⚠️ อ่านไฟล์ %s ไม่ได้ ข้ามรูปนี้: %v", filepath.Base(p), readErr)
				entries[i] = imgEntry{path: p}
				parts = append(parts, geminiPart{Text: fmt.Sprintf("[Photo %d: image could not be read]", i+1)})
				continue
			}
			imgBytes = rawBytes
			ext := strings.ToLower(filepath.Ext(p))
			switch ext {
			case ".png":
				mimeType = "image/png"
			case ".webp":
				mimeType = "image/webp"
			default:
				mimeType = "image/jpeg"
			}
		}

		entries[i] = imgEntry{
			path:     p,
			mimeType: mimeType,
			data:     base64.StdEncoding.EncodeToString(imgBytes),
			ok:       true,
		}
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{
				MimeType: mimeType,
				Data:     entries[i].data,
			},
		})
	}

	// Step 2: Build & send request
	reqPayload := geminiRequest{
		Contents: []geminiContent{{Parts: parts}},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:       0.2,
		},
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal batch request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.modelName, s.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("create batch http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] batch network error: %v — ถอยไปทีละรูป", err)
		return s.fallbackPerPhoto(ctx, filePaths), nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.fallbackPerPhoto(ctx, filePaths), nil
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] ไม่สามารถ parse batch response: %s", string(respBody))
		return s.fallbackPerPhoto(ctx, filePaths), nil
	}
	if geminiResp.Error != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] batch Gemini error (%d): %s — ถอยไปทีละรูป",
			geminiResp.Error.Code, geminiResp.Error.Message)
		return s.fallbackPerPhoto(ctx, filePaths), nil
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return s.fallbackPerPhoto(ctx, filePaths), nil
	}

	rawText := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	var items []batchResultItem
	if err := json.Unmarshal([]byte(rawText), &items); err != nil {
		log.Printf("[AI-Vision] ⚠️ [FALLBACK] ไม่สามารถ parse batch JSON array: %s", rawText)
		return s.fallbackPerPhoto(ctx, filePaths), nil
	}

	// Step 3: Map results back by index, fall back for any missing photo
	results := make([]*VisionResult, n)
	covered := make([]bool, n)
	for _, it := range items {
		idx := it.Index - 1 // convert 1-based → 0-based
		if idx < 0 || idx >= n {
			continue
		}
		if !entries[idx].ok {
			// AI returned an entry for a photo we couldn't read — skip, leave nil
			covered[idx] = true
			continue
		}
		results[idx] = &VisionResult{Description: it.Description, Tags: it.Tags}
		covered[idx] = true
	}

	// Fall back to single-photo analysis for any uncovered image
	var missingIdx []int
	for i := 0; i < n; i++ {
		if !covered[i] && entries[i].ok {
			missingIdx = append(missingIdx, i)
		}
	}
	for _, i := range missingIdx {
		log.Printf("[AI-Vision] ℹ️ รูปลำดับ %d ไม่อยู่ในผลลัพธ์ batch — วิเคราะห์เดี่ยว", i+1)
		r, err := s.AnalyzePhoto(ctx, filePaths[i])
		if err == nil && r != nil {
			results[i] = r
		}
	}

	duration := time.Since(start)
	coveredCount := 0
	for _, r := range results {
		if r != nil {
			coveredCount++
		}
	}
	log.Printf("[AI-Vision] ✅ batch วิเคราะห์เสร็จใน %v — สำเร็จ %d/%d รูป", duration, coveredCount, n)

	return results, nil
}

// fallbackPerPhoto analyzes each path one-by-one, returning results aligned with filePaths.
// Used when the batch request fails so callers still get per-path output.
func (s *geminiVisionService) fallbackPerPhoto(ctx context.Context, filePaths []string) []*VisionResult {
	results := make([]*VisionResult, len(filePaths))
	for i, p := range filePaths {
		r, err := s.AnalyzePhoto(ctx, p)
		if err == nil && r != nil {
			results[i] = r
		}
	}
	return results
}
