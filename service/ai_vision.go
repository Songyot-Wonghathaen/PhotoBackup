package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"PhotoVault/controller/dto"
	"PhotoVault/model"
	"PhotoVault/utils"

	"github.com/joho/godotenv"
)

type VisionResult = dto.VisionResult
type modelQuotaRecord = model.ModelQuotaRecord
type modelQuotaCacheFile = model.ModelQuotaCacheFile
type batchItemWire = model.BatchItemWire
type geminiRequest = model.GeminiRequest
type geminiResponse = model.GeminiResponse
type geminiContent = model.GeminiContent
type geminiPart = model.GeminiPart
type geminiInlineData = model.GeminiInlineData
type geminiGenerationConfig = model.GeminiGenerationConfig

type AiVisionService interface {
	AnalyzePhoto(ctx context.Context, filePath string) (*VisionResult, error)
	IsAvailable() bool
	ExpandSearchQuery(ctx context.Context, query string) ([]string, error)
}

// BatchAiVisionService analyzes multiple photos per request.
type BatchAiVisionService interface {
	AiVisionService
	AnalyzePhotoBatch(ctx context.Context, filePaths []string) ([]*VisionResult, error)
	PrebuildBatchCollages(ctx context.Context, batches [][]string)
	WorkerConcurrency() int
	BatchSize() int
}

var defaultVisionModels = []string{
	"gemini-3.5-flash-lite",
	"gemini-3.1-flash-lite",
	"gemini-3-flash-preview",
	"gemini-3.8-flash",
	"gemini-3.6-flash",
}

func isPrimaryFlashLiteModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "flash-lite")
}

type geminiVisionService struct {
	apiKey           string
	apiKeys          []string
	httpClient       *http.Client
	modelName        string
	models           []string
	counter          uint64
	cooldownMu       sync.RWMutex
	cooldowns        map[string]time.Time
	lastUsed         map[string]time.Time
	lastKeyUsed      map[string]time.Time
	sharedKeyBucket  map[string]bool
	lastGlobalSlot   time.Time
	smoothBurstCount int
	inFlight         map[string]int
	lastRateLog      map[string]time.Time
	cachePath        string
	collageCache     sync.Map
	settingsMu       sync.RWMutex
	settings         *persistedAppSettings
}

func quotaCachePath() string {
	cwd, err := os.Getwd()
	if err == nil && cwd != "/" && cwd != "\\" {
		p := filepath.Join(cwd, "database", "ai_model_quota.json")
		if _, statErr := os.Stat(filepath.Dir(p)); statErr == nil {
			return p
		}
		if mkErr := os.MkdirAll(filepath.Dir(p), 0755); mkErr == nil {
			return p
		}
	}
	if cfgDir, cfgErr := os.UserConfigDir(); cfgErr == nil {
		appDir := filepath.Join(cfgDir, "PhotoVault", "database")
		_ = os.MkdirAll(appDir, 0755)
		return filepath.Join(appDir, "ai_model_quota.json")
	}
	return "database/ai_model_quota.json"
}

func nextDailyResetTime() time.Time {
	// Google AI Studio daily quota resets at 00:00 UTC (07:00 Bangkok) or 00:00 PST (15:00 Bangkok)
	now := time.Now().UTC()
	reset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, time.UTC)
	if reset.Sub(now) < 2*time.Hour {
		reset = reset.Add(24 * time.Hour)
	}
	return reset
}

func (s *geminiVisionService) loadQuotaCache() {
	s.cooldownMu.Lock()
	defer s.cooldownMu.Unlock()

	data, err := os.ReadFile(s.cachePath)
	if err != nil {
		return
	}
	var f modelQuotaCacheFile
	if err := json.Unmarshal(data, &f); err != nil || f.Records == nil {
		return
	}

	now := time.Now()
	restored := 0
	for k, rec := range f.Records {
		resetTime, tErr := time.Parse(time.RFC3339, rec.ResetAt)
		if tErr == nil && resetTime.Sub(now) > 30*time.Minute {
			s.cooldowns[k] = resetTime
			restored++
		}
	}
	if restored > 0 {
		log.Printf("[AI-Vision] โหลดประวัติโควตา: มี %d เส้นทางกำลังพักใช้งานเพื่อรอรีเซ็ตรายวัน", restored)
	}
}

func (s *geminiVisionService) saveQuotaCacheLocked() {
	if s.cachePath == "" {
		return
	}
	records := make(map[string]modelQuotaRecord)
	now := time.Now()
	for k, until := range s.cooldowns {
		// Only persist long daily RPD cooldowns (> 30m), never short RPM cooldowns
		if until.Sub(now) > 30*time.Minute {
			parts := strings.SplitN(k, "|", 2)
			kp := ""
			m := k
			if len(parts) == 2 {
				kp = parts[0]
				m = parts[1]
			}
			records[k] = modelQuotaRecord{
				Model:     m,
				KeyPrefix: kp,
				Reason:    "Daily Quota Exceeded (RPD)",
				ResetAt:   until.Format(time.RFC3339),
			}
		}
	}
	_ = os.MkdirAll(filepath.Dir(s.cachePath), 0755)
	data, err := json.MarshalIndent(modelQuotaCacheFile{Records: records}, "", "  ")
	if err == nil {
		_ = os.WriteFile(s.cachePath, data, 0644)
	}
}

func loadKeysFromEnv() []string {
	_ = godotenv.Overload()
	_ = godotenv.Overload("../.env")
	if home, err := os.UserHomeDir(); err == nil {
		_ = godotenv.Load(filepath.Join(home, ".env"))
	}

	var keys []string
	seenKey := make(map[string]bool)
	rawKeys := strings.Join([]string{
		os.Getenv("GEMINI_API_KEYS"),
		os.Getenv("GEMINI_API_KEY"),
		os.Getenv("GOOGLE_API_KEY"),
	}, ",")
	for _, part := range strings.Split(rawKeys, ",") {
		k := strings.TrimSpace(part)
		k = strings.Trim(k, "\"'")
		if k != "" && !seenKey[k] {
			seenKey[k] = true
			keys = append(keys, k)
		}
	}

	for _, emb := range getEmbeddedDefaultKeys() {
		emb = strings.TrimSpace(emb)
		if emb != "" && !seenKey[emb] {
			seenKey[emb] = true
			keys = append(keys, emb)
		}
	}

	return keys
}

func NewAiVisionService() AiVisionService {
	keys := loadKeysFromEnv()

	primaryKey := ""
	if len(keys) > 0 {
		primaryKey = keys[0]
	}

	var models []string
	seenModel := make(map[string]bool)
	rawModels := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if rawModels != "" && !strings.EqualFold(rawModels, "auto") {
		for _, m := range strings.Split(rawModels, ",") {
			m = strings.TrimSpace(m)
			m = strings.Trim(m, "\"'")
			if m != "" && !seenModel[m] {
				seenModel[m] = true
				models = append(models, m)
			}
		}
	}
	for _, dm := range defaultVisionModels {
		if !seenModel[dm] {
			seenModel[dm] = true
			models = append(models, dm)
		}
	}

	primaryModel := "gemini-3.5-flash-lite"
	if len(models) > 0 {
		primaryModel = models[0]
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		IdleConnTimeout:     90 * time.Second,
	}

	svc := &geminiVisionService{
		apiKey:  primaryKey,
		apiKeys: keys,
		httpClient: &http.Client{
			Timeout:   20 * time.Second,
			Transport: transport,
		},
		modelName:       primaryModel,
		models:          models,
		cooldowns:       make(map[string]time.Time),
		lastUsed:        make(map[string]time.Time),
		lastKeyUsed:     make(map[string]time.Time),
		sharedKeyBucket: make(map[string]bool),
		inFlight:        make(map[string]int),
		lastRateLog:     make(map[string]time.Time),
		cachePath:       quotaCachePath(),
	}
	svc.loadQuotaCache()
	svc.initSettings(keys)
	activeCount := len(svc.getKeys())
	log.Printf("[AI-Vision] โหลด Gemini API Keys สำเร็จจำนวน %d Keys (%d เส้นทางความเร็วสูง, %d Workers) พร้อมระบบ Smooth Pacing ต่อเนื่อง",
		activeCount, activeCount*2, svc.WorkerConcurrency())
	return svc
}

func (s *geminiVisionService) getKeys() []string {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	if len(s.apiKeys) > 0 {
		out := make([]string, len(s.apiKeys))
		copy(out, s.apiKeys)
		return out
	}
	if s.apiKey != "" {
		return []string{s.apiKey}
	}
	return nil
}

func (s *geminiVisionService) getModels() []string {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	if len(s.models) > 0 {
		out := make([]string, len(s.models))
		copy(out, s.models)
		return out
	}
	if s.modelName != "" {
		return []string{s.modelName}
	}
	return defaultVisionModels
}

func (s *geminiVisionService) IsAvailable() bool {
	return len(s.getKeys()) > 0
}

func (s *geminiVisionService) WorkerConcurrency() int {
	return ComputeSmartWorkers(len(s.getKeys()))
}

func (s *geminiVisionService) BatchSize() int {
	return 25
}

func isDailyQuotaError(code int, msg string) bool {
	if code != 429 && code != 403 {
		return false
	}
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "per-minute") || strings.Contains(lower, "per minute") || strings.Contains(lower, "perminute") {
		return false
	}
	// Only treat as Daily RPD if the message DOES NOT say PerMinute and explicitly mentions PerDay or retry in hours
	if strings.Contains(lower, "perminute") {
		return false
	}
	if idx := strings.Index(lower, "retry in "); idx >= 0 {
		rest := lower[idx+len("retry in "):]
		if end := strings.IndexByte(rest, '.'); end > 0 {
			rest = rest[:end]
		}
		if strings.Contains(rest, "h") && !strings.Contains(lower, "perminute") {
			return true
		}
		if !strings.Contains(lower, "limit: 0") && !strings.Contains(lower, "limit: 20") {
			return false
		}
	}
	return strings.Contains(lower, "per-day") ||
		strings.Contains(lower, "per day") ||
		strings.Contains(lower, "perday") ||
		strings.Contains(lower, "limit: 20") ||
		strings.Contains(lower, "daily")
}

// minModelInterval returns per-route cooldown interval.
func minModelInterval(model string) time.Duration {
	lower := strings.ToLower(model)
	switch {
	case strings.Contains(lower, "flash-lite"):
		return 9600 * time.Millisecond
	default:
		return 12500 * time.Millisecond
	}
}

// modelTierBias prioritizes Flash-Lite models over fallback models.
func modelTierBias(model string) time.Duration {
	lower := strings.ToLower(model)
	switch {
	case lower == "gemini-3.5-flash-lite":
		return 0
	case strings.Contains(lower, "flash-lite"):
		return 120 * time.Millisecond
	default:
		return 30 * time.Second
	}
}

// globalSmoothGap returns delay between global request launches.
func globalSmoothGap(keyCount int) time.Duration {
	switch {
	case keyCount <= 1:
		return 1800 * time.Millisecond
	case keyCount <= 3:
		return 950 * time.Millisecond
	case keyCount <= 5:
		return 650 * time.Millisecond
	default:
		return 380 * time.Millisecond
	}
}

func (s *geminiVisionService) markDailyQuotaExhausted(key, model string) {
	resetAt := nextDailyResetTime()
	s.cooldownMu.Lock()
	if s.cooldowns == nil {
		s.cooldowns = make(map[string]time.Time)
	}
	s.cooldowns[key+"|"+model] = resetAt
	s.saveQuotaCacheLocked()
	s.cooldownMu.Unlock()

	locBangkok, _ := time.LoadLocation("Asia/Bangkok")
	bkkTime := resetAt.In(locBangkok).Format("15:04:05 (วันที่ 02/01)")
	log.Printf("[AI-Vision] โมเดล %s ติดลิมิตรายวัน (RPD) สลับไปใช้เส้นทางอื่นอัตโนมัติ (รีเซ็ต %s)", model, bkkTime)
}

func (s *geminiVisionService) markRpmCooldown(key, model string, d time.Duration) {
	now := time.Now()
	s.cooldownMu.Lock()
	if s.cooldowns == nil {
		s.cooldowns = make(map[string]time.Time)
	}
	if s.lastRateLog == nil {
		s.lastRateLog = make(map[string]time.Time)
	}
	if s.sharedKeyBucket == nil {
		s.sharedKeyBucket = make(map[string]bool)
	}
	id := key + "|" + model
	until := now.Add(d)
	if existing, ok := s.cooldowns[id]; !ok || until.After(existing) {
		s.cooldowns[id] = until
	}
	// Mark key as shared-bucket if secondary model hits 429 right after primary model.
	if model != "gemini-3.5-flash-lite" {
		if lu35, ok := s.lastUsed[key+"|gemini-3.5-flash-lite"]; ok && now.Sub(lu35) < 9500*time.Millisecond {
			s.sharedKeyBucket[key] = true
		}
	}
	shouldLog := false
	if last, ok := s.lastRateLog[model]; !ok || now.Sub(last) > 6*time.Second {
		s.lastRateLog[model] = now
		shouldLog = true
	}
	s.cooldownMu.Unlock()

	if shouldLog {
		log.Printf("[AI-Vision] สลับเส้นทางอัตโนมัติ (%s พัก %v เพื่อรักษาความเร็วต่อเนื่อง)", model, d)
	}
}

func (s *geminiVisionService) releaseEndpoint(key, model string) {
	s.cooldownMu.Lock()
	if s.inFlight != nil {
		id := key + "|" + model
		if s.inFlight[id] > 0 {
			s.inFlight[id]--
		}
	}
	s.cooldownMu.Unlock()
}

// acquireEndpoint selects the next available (key, model) slot.
func (s *geminiVisionService) acquireEndpoint(ctx context.Context, tried map[string]bool) (string, string, error) {
	keys := s.getKeys()
	models := s.getModels()
	if len(keys) == 0 || len(models) == 0 {
		return "", "", errors.New("no API key available")
	}

	seq := int(atomic.AddUint64(&s.counter, 1) - 1)

	s.cooldownMu.Lock()
	if s.cooldowns == nil {
		s.cooldowns = make(map[string]time.Time)
	}
	if s.lastUsed == nil {
		s.lastUsed = make(map[string]time.Time)
	}
	if s.lastKeyUsed == nil {
		s.lastKeyUsed = make(map[string]time.Time)
	}
	if s.sharedKeyBucket == nil {
		s.sharedKeyBucket = make(map[string]bool)
	}
	if s.inFlight == nil {
		s.inFlight = make(map[string]int)
	}

	now := time.Now()

	// Check if any Flash-Lite route is available.
	hasAvailableFlashLite := false
	for _, k := range keys {
		for _, m := range models {
			if !isPrimaryFlashLiteModel(m) {
				continue
			}
			id := k + "|" + m
			if tried != nil && tried[id] {
				continue
			}
			if until, ok := s.cooldowns[id]; ok && until.Sub(now) > 25*time.Minute {
				continue
			}
			hasAvailableFlashLite = true
			break
		}
		if hasAvailableFlashLite {
			break
		}
	}

	bestKey := ""
	bestModel := ""
	var bestSlotTime time.Time
	var bestScoreTime time.Time

	totalCombos := len(keys) * len(models)
	for offset := 0; offset < totalCombos; offset++ {
		idx := (seq + offset) % totalCombos
		k := keys[idx%len(keys)]
		m := models[(idx/len(keys))%len(models)]
		id := k + "|" + m

		if tried != nil && tried[id] {
			continue
		}
		if hasAvailableFlashLite && !isPrimaryFlashLiteModel(m) {
			continue
		}

		slotTime := now
		if until, ok := s.cooldowns[id]; ok {
			if now.After(until) {
				delete(s.cooldowns, id)
			} else if until.Sub(now) > 25*time.Minute {
				continue
			} else if until.After(slotTime) {
				slotTime = until
			}
		}

		interval := minModelInterval(m)

		if lu, ok := s.lastUsed[id]; ok {
			nextAllowed := lu.Add(interval)
			if nextAllowed.After(slotTime) {
				slotTime = nextAllowed
			}
		}

		if s.sharedKeyBucket[k] {
			if lku, ok := s.lastKeyUsed[k]; ok {
				nextKeyAllowed := lku.Add(interval)
				if nextKeyAllowed.After(slotTime) {
					slotTime = nextKeyAllowed
				}
			}
		} else if lku, ok := s.lastKeyUsed[k]; ok {
			minKeyGap := lku.Add(5200 * time.Millisecond)
			if minKeyGap.After(slotTime) {
				slotTime = minKeyGap
			}
		}

		activeReqs := s.inFlight[id]
		scoreTime := slotTime.
			Add(time.Duration(activeReqs) * 5000 * time.Millisecond).
			Add(modelTierBias(m))

		if bestKey == "" || scoreTime.Before(bestScoreTime) {
			bestKey = k
			bestModel = m
			bestSlotTime = slotTime
			bestScoreTime = scoreTime
		}
	}

	if bestKey == "" {
		bestKey = keys[seq%len(keys)]
		bestModel = models[0]
		bestSlotTime = now
	}

	// Stagger launches after the initial 3 requests.
	if now.Sub(s.lastGlobalSlot) > 3*time.Second {
		s.smoothBurstCount = 0
	}
	s.smoothBurstCount++
	if s.smoothBurstCount > 3 {
		minGlobalNext := s.lastGlobalSlot.Add(globalSmoothGap(len(keys)))
		if minGlobalNext.After(bestSlotTime) {
			bestSlotTime = minGlobalNext
		}
	}
	s.lastGlobalSlot = bestSlotTime

	bestID := bestKey + "|" + bestModel
	s.lastUsed[bestID] = bestSlotTime
	s.lastKeyUsed[bestKey] = bestSlotTime
	s.inFlight[bestID]++
	s.cooldownMu.Unlock()

	waitDur := time.Until(bestSlotTime)
	if waitDur > 5*time.Millisecond {
		if waitDur > 15*time.Second {
			waitDur = 15 * time.Second
		}
		timer := time.NewTimer(waitDur)
		select {
		case <-ctx.Done():
			timer.Stop()
			s.releaseEndpoint(bestKey, bestModel)
			return "", "", ctx.Err()
		case <-timer.C:
		}
	}

	return bestKey, bestModel, nil
}

func (s *geminiVisionService) callGeminiWithFallback(ctx context.Context, prompt string, mimeType string, encodedImage string) (string, error) {
	keys := s.getKeys()
	if len(keys) == 0 {
		return "", errors.New("no API key available")
	}

	maxAttempts := 6
	if len(keys) == 1 {
		maxAttempts = 4
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	tried := make(map[string]bool, maxAttempts)
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		apiKey, modelName, acqErr := s.acquireEndpoint(ctx, tried)
		if acqErr != nil {
			return "", acqErr
		}
		tried[apiKey+"|"+modelName] = true

		genCfg := geminiGenerationConfig{
			Temperature: 0.2,
		}
		// Gemma models on v1beta do not accept response_mime_type="application/json"
		if !strings.HasPrefix(strings.ToLower(modelName), "gemma-") {
			genCfg.ResponseMimeType = "application/json"
		}

		reqParts := []geminiPart{
			{Text: prompt},
		}
		if encodedImage != "" {
			reqParts = append(reqParts, geminiPart{
				InlineData: &geminiInlineData{
					MimeType: mimeType,
					Data:     encodedImage,
				},
			})
		}

		reqPayload := geminiRequest{
			Contents: []geminiContent{
				{
					Parts: reqParts,
				},
			},
			GenerationConfig: genCfg,
		}

		jsonBytes, err := json.Marshal(reqPayload)
		if err != nil {
			s.releaseEndpoint(apiKey, modelName)
			return "", fmt.Errorf("marshal request error: %w", err)
		}

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
		if err != nil {
			s.releaseEndpoint(apiKey, modelName)
			return "", fmt.Errorf("create http request error: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		reqStart := time.Now()
		s.cooldownMu.Lock()
		s.lastUsed[apiKey+"|"+modelName] = reqStart
		s.lastKeyUsed[apiKey] = reqStart
		s.cooldownMu.Unlock()

		resp, err := client.Do(httpReq)
		if err != nil {
			s.releaseEndpoint(apiKey, modelName)
			lastErr = fmt.Errorf("gemini api network error (%s): %w", modelName, err)
			s.markRpmCooldown(apiKey, modelName, 3*time.Second)
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		s.releaseEndpoint(apiKey, modelName)

		if readErr != nil {
			lastErr = fmt.Errorf("read response body error: %w", readErr)
			continue
		}

		var geminiResp geminiResponse
		if err := json.Unmarshal(respBody, &geminiResp); err != nil {
			lastErr = fmt.Errorf("unmarshal gemini response error (%s): %w", modelName, err)
			s.markRpmCooldown(apiKey, modelName, 3*time.Second)
			continue
		}

		if geminiResp.Error != nil {
			code := geminiResp.Error.Code
			msg := geminiResp.Error.Message
			lastErr = fmt.Errorf("gemini api error [%d] (%s): %s", code, modelName, msg)

			if isDailyQuotaError(code, msg) {
				s.markDailyQuotaExhausted(apiKey, modelName)
				continue
			}
			if code == 429 || code == 503 {
				s.markRpmCooldown(apiKey, modelName, 5*time.Second)
				continue
			}
			if code == 404 || code == 400 {
				s.markRpmCooldown(apiKey, modelName, 15*time.Minute)
				continue
			}
			if len(keys) == 1 {
				return "", lastErr
			}
			s.markRpmCooldown(apiKey, modelName, 10*time.Minute)
			continue
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("no content candidate returned from %s", modelName)
			s.markRpmCooldown(apiKey, modelName, 3*time.Second)
			continue
		}

		parts := geminiResp.Candidates[0].Content.Parts
		rawText := ""
		// Always select the last non-thought part so chain-of-thought reasoning is never mistaken for JSON output
		for i := len(parts) - 1; i >= 0; i-- {
			if !parts[i].Thought && strings.TrimSpace(parts[i].Text) != "" {
				rawText = strings.TrimSpace(parts[i].Text)
				break
			}
		}
		if rawText == "" {
			rawText = strings.TrimSpace(parts[len(parts)-1].Text)
		}
		rawText = cleanJSONText(rawText)
		return rawText, nil
	}

	return "", lastErr
}

func cleanJSONText(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```JSON")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

func (s *geminiVisionService) AnalyzePhoto(ctx context.Context, filePath string) (*VisionResult, error) {
	filename := filepath.Base(filePath)

	if !s.IsAvailable() {
		err := errors.New("GEMINI_API_KEY is not configured in .env or environment")
		log.Printf("[AI-Vision] [FALLBACK] ข้ามการวิเคราะห์ %s: %v", filename, err)
		return nil, err
	}

	var imgBytes []byte
	mimeType := "image/jpeg"

	thumbBytes, err := utils.GetThumbnailBytes(filePath, 256)
	if err == nil && len(thumbBytes) > 0 {
		imgBytes = thumbBytes
	} else {
		rawBytes, readErr := os.ReadFile(filePath)
		if readErr != nil {
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
	}

	encoded := base64.StdEncoding.EncodeToString(imgBytes)
	prompt := "Analyze this photo carefully. Provide: " +
		"1) A concise natural description of the image content in Thai (1-2 sentences). " +
		"2) An array of 3-5 relevant, specific keyword tags in Thai (nouns, activities, objects, or scenery). " +
		"Respond strictly with valid JSON having exactly two keys: 'description' (string) and 'tags' (array of strings)."

	rawText, err := s.callGeminiWithFallback(ctx, prompt, mimeType, encoded)
	if err != nil {
		return nil, err
	}

	var result VisionResult
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		return nil, fmt.Errorf("parse structured json error: %w", err)
	}
	return &result, nil
}

func batchCacheKey(filePaths []string) string {
	if len(filePaths) == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%s:%s", len(filePaths), filePaths[0], filePaths[len(filePaths)-1])
}

// PrebuildBatchCollages pre-encodes grid collages in background goroutines.
func (s *geminiVisionService) PrebuildBatchCollages(ctx context.Context, batches [][]string) {
	ch := make(chan []string, len(batches))
	for _, b := range batches {
		if len(b) > 1 {
			ch <- b
		}
	}
	close(ch)

	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for paths := range ch {
				select {
				case <-ctx.Done():
					return
				default:
				}
				key := batchCacheKey(paths)
				if _, exists := s.collageCache.Load(key); !exists {
					if data, err := buildGridCollageJPEG(paths, 220); err == nil {
						s.collageCache.Store(key, data)
					}
				}
			}
		}()
	}
	wg.Wait()
}

// AnalyzePhotoBatch combines photos into a grid collage and analyzes them in one request.
func (s *geminiVisionService) AnalyzePhotoBatch(ctx context.Context, filePaths []string) ([]*VisionResult, error) {
	n := len(filePaths)
	if n == 0 {
		return nil, nil
	}
	if n == 1 {
		res, err := s.AnalyzePhoto(ctx, filePaths[0])
		if err != nil {
			return nil, err
		}
		return []*VisionResult{res}, nil
	}

	if !s.IsAvailable() {
		return nil, errors.New("GEMINI_API_KEY is not configured in .env or environment")
	}

	cacheKey := batchCacheKey(filePaths)
	var collageBytes []byte
	if cached, ok := s.collageCache.Load(cacheKey); ok {
		collageBytes = cached.([]byte)
	} else {
		var err error
		collageBytes, err = buildGridCollageJPEG(filePaths, 220)
		if err != nil {
			return nil, fmt.Errorf("failed to build grid collage: %w", err)
		}
		s.collageCache.Store(cacheKey, collageBytes)
	}

	prompt := fmt.Sprintf(
		"Grid collage of %d photos labeled [0] to [%d] at the top-left of each tile. "+
			"For EVERY tile 0..%d, inspect all visual details (subject, action, objects, setting) and return ONLY a JSON array of %d objects:\n"+
			"[{\"i\":0,\"d\":\"คำอธิบายภาพภาษาไทยกระชับตรงจุด 1 ประโยค\",\"t\":[\"แท็ก1\",\"แท็ก2\",\"แท็ก3\"]}]",
		n, n-1, n-1, n,
	)

	encoded := base64.StdEncoding.EncodeToString(collageBytes)
	rawText, err := s.callGeminiWithFallback(ctx, prompt, "image/jpeg", encoded)
	if err != nil {
		return nil, err
	}
	s.collageCache.Delete(cacheKey)

	items, err := parseBatchJsonItems(rawText)
	if err != nil {
		return nil, fmt.Errorf("parse batch json error: %w", err)
	}

	results := make([]*VisionResult, n)
	for pos, item := range items {
		idx := item.Index
		if idx == 0 && pos > 0 && item.AltIndex > 0 {
			idx = item.AltIndex
		}
		if idx < 0 || idx >= n {
			if pos < n && results[pos] == nil {
				idx = pos
			} else {
				continue
			}
		}
		desc := strings.TrimSpace(item.Description)
		if desc == "" {
			desc = strings.TrimSpace(item.AltDescription)
		}
		tags := item.Tags
		if len(tags) == 0 {
			tags = item.AltTags
		}
		results[idx] = &VisionResult{
			Description: desc,
			Tags:        tags,
		}
	}

	return results, nil
}

// parseBatchJsonItems parses batch JSON items and falls back to object-by-object extraction.
func parseBatchJsonItems(raw string) ([]batchItemWire, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, errors.New("empty response text")
	}

	var items []batchItemWire
	if err := json.Unmarshal([]byte(s), &items); err == nil && len(items) > 0 {
		return items, nil
	}

	var recovered []batchItemWire
	inString := false
	escaped := false
	depth := 0
	objStart := -1

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			if depth == 0 {
				objStart = i
			}
			depth++
		} else if ch == '}' {
			if depth > 0 {
				depth--
				if depth == 0 && objStart >= 0 {
					objStr := s[objStart : i+1]
					var single batchItemWire
					if uErr := json.Unmarshal([]byte(objStr), &single); uErr == nil {
						if single.Description != "" || single.AltDescription != "" || len(single.Tags) > 0 || len(single.AltTags) > 0 {
							recovered = append(recovered, single)
						}
					}
					objStart = -1
				}
			}
		}
	}

	if len(recovered) > 0 {
		return recovered, nil
	}

	return nil, errors.New("no valid JSON objects found in batch response")
}

// buildGridCollageJPEG builds a numbered JPEG grid collage from image paths.
func buildGridCollageJPEG(filePaths []string, cellDim int) ([]byte, error) {
	n := len(filePaths)
	cols := 5
	if n < cols {
		cols = n
	}
	rows := (n + cols - 1) / cols

	canvasW := cols * cellDim
	canvasH := rows * cellDim
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	// Decode tiles concurrently for speed
	decoded := make([]image.Image, n)
	var wg sync.WaitGroup
	for i, fp := range filePaths {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			defer func() {
				_ = recover()
			}()
			decoded[idx] = decodeImageFast(path, cellDim)
		}(i, fp)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		col := i % cols
		row := i / cols
		ox := col * cellDim
		oy := row * cellDim

		srcImg := decoded[i]
		if srcImg != nil {
			drawScaledTile(canvas, ox, oy, cellDim, cellDim, srcImg)
		}
		drawIndexBadge(canvas, ox+3, oy+3, i)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, canvas, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeImageFast(filePath string, targetDim int) image.Image {
	thumbBytes, err := utils.GetThumbnailBytes(filePath, targetDim)
	if err == nil && len(thumbBytes) > 0 {
		if img, _, decErr := image.Decode(bytes.NewReader(thumbBytes)); decErr == nil {
			return img
		}
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return img
}

// drawScaledTile scales src into dst at (ox, oy, w, h).
func drawScaledTile(dst *image.RGBA, ox, oy, w, h int, src image.Image) {
	sb := src.Bounds()
	sw := sb.Dx()
	sh := sb.Dy()
	if sw <= 0 || sh <= 0 {
		return
	}
	dxStep := sw / (w * 2)
	if dxStep < 1 {
		dxStep = 1
	}
	dyStep := sh / (h * 2)
	if dyStep < 1 {
		dyStep = 1
	}
	for y := 0; y < h; y++ {
		sy0 := sb.Min.Y + (y*sh)/h
		sy1 := sy0 + dyStep
		if sy1 >= sb.Max.Y {
			sy1 = sb.Max.Y - 1
		}
		for x := 0; x < w; x++ {
			sx0 := sb.Min.X + (x*sw)/w
			sx1 := sx0 + dxStep
			if sx1 >= sb.Max.X {
				sx1 = sb.Max.X - 1
			}
			r00, g00, b00, _ := src.At(sx0, sy0).RGBA()
			r10, g10, b10, _ := src.At(sx1, sy0).RGBA()
			r01, g01, b01, _ := src.At(sx0, sy1).RGBA()
			r11, g11, b11, _ := src.At(sx1, sy1).RGBA()
			dst.SetRGBA(ox+x, oy+y, color.RGBA{
				R: uint8(((r00 + r10 + r01 + r11) >> 2) >> 8),
				G: uint8(((g00 + g10 + g01 + g11) >> 2) >> 8),
				B: uint8(((b00 + b10 + b01 + b11) >> 2) >> 8),
				A: 255,
			})
		}
	}
}

// 3x5 bitmap digits for crisp index badge rendering without external font dependencies
var digitBitmaps = [10][5]uint8{
	{0b111, 0b101, 0b101, 0b101, 0b111}, // 0
	{0b010, 0b110, 0b010, 0b010, 0b111}, // 1
	{0b111, 0b001, 0b111, 0b100, 0b111}, // 2
	{0b111, 0b001, 0b111, 0b001, 0b111}, // 3
	{0b101, 0b101, 0b111, 0b001, 0b001}, // 4
	{0b111, 0b100, 0b111, 0b001, 0b111}, // 5
	{0b111, 0b100, 0b111, 0b101, 0b111}, // 6
	{0b111, 0b001, 0b010, 0b010, 0b010}, // 7
	{0b111, 0b101, 0b111, 0b101, 0b111}, // 8
	{0b111, 0b101, 0b111, 0b001, 0b111}, // 9
}

func drawIndexBadge(dst *image.RGBA, x, y, index int) {
	label := fmt.Sprintf("%d", index)
	scale := 2
	charW := 3 * scale
	charH := 5 * scale
	gap := 1 * scale
	pad := 3

	badgeW := len(label)*(charW+gap) - gap + pad*2
	badgeH := charH + pad*2

	bgRect := image.Rect(x, y, x+badgeW, y+badgeH)
	draw.Draw(dst, bgRect, &image.Uniform{color.RGBA{0, 0, 0, 220}}, image.Point{}, draw.Src)

	white := color.RGBA{255, 255, 255, 255}
	cx := x + pad
	cy := y + pad
	for _, ch := range label {
		if ch >= '0' && ch <= '9' {
			bm := digitBitmaps[ch-'0']
			for r := 0; r < 5; r++ {
				for c := 0; c < 3; c++ {
					if (bm[r]>>(2-c))&1 == 1 {
						for dy := 0; dy < scale; dy++ {
							for dx := 0; dx < scale; dx++ {
								dst.SetRGBA(cx+c*scale+dx, cy+r*scale+dy, white)
							}
						}
					}
				}
			}
		}
		cx += charW + gap
	}
}

// queryExpansionCache caches expanded search terms in memory to keep search snappy
var queryExpansionCache sync.Map // map[string][]string

// ExpandSearchQuery uses Gemini to expand a search query (e.g. "ขนมปังไหม้")
// into specific and broader related terms in Thai to enable high-quality ranked searching.
func (s *geminiVisionService) ExpandSearchQuery(ctx context.Context, query string) ([]string, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, nil
	}

	// 1. Fast in-memory cache lookup
	cacheKey := strings.ToLower(trimmed)
	if cached, ok := queryExpansionCache.Load(cacheKey); ok {
		if terms, ok := cached.([]string); ok && len(terms) > 0 {
			return terms, nil
		}
	}

	if !s.IsAvailable() {
		return []string{trimmed}, nil
	}

	prompt := fmt.Sprintf(`You are a search query expansion assistant for an image gallery in Thai.
User query: "%s"

Task:
1. Break down the query into relevant search keywords in Thai (and English if relevant).
2. The first item MUST be the original query "%s".
3. Add close synonyms or specific variations (e.g. for "ขนมปังไหม้": "ขนมปังเกรียม", "ขนมปังดำ").
4. Add broader base concepts/categories (e.g. "ขนมปังปิ้ง", "ขนมปัง", "toast") so that if specific photos are not found, relevant general photos can still be returned.
5. Return ONLY a valid JSON array of strings (e.g. ["term1", "term2", "term3"]). Maximum 6 items. No explanation.`, trimmed, trimmed)

	rawJSON, err := s.callGeminiWithFallback(ctx, prompt, "", "")
	if err != nil {
		log.Printf("[AI-Vision] ExpandSearchQuery fallback for '%s': %v", trimmed, err)
		return []string{trimmed}, nil
	}

	var terms []string
	if err := json.Unmarshal([]byte(rawJSON), &terms); err != nil {
		start := strings.Index(rawJSON, "[")
		end := strings.LastIndex(rawJSON, "]")
		if start >= 0 && end > start {
			_ = json.Unmarshal([]byte(rawJSON[start:end+1]), &terms)
		}
	}

	cleanedTerms := make([]string, 0, len(terms)+1)
	seen := make(map[string]bool)

	// Ensure original query is the very first term
	cleanedTerms = append(cleanedTerms, trimmed)
	seen[strings.ToLower(trimmed)] = true

	for _, t := range terms {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		lower := strings.ToLower(t)
		if seen[lower] {
			continue
		}
		seen[lower] = true
		cleanedTerms = append(cleanedTerms, t)
	}

	if len(cleanedTerms) > 0 {
		queryExpansionCache.Store(cacheKey, cleanedTerms)
		return cleanedTerms, nil
	}

	return []string{trimmed}, nil
}

