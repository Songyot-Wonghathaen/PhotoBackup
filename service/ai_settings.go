package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"PhotoVault/controller/dto"
)

// ConfigurableAiVisionService extends BatchAiVisionService with runtime settings and multi-project API pool management.
type ConfigurableAiVisionService interface {
	BatchAiVisionService
	GetAppSettings() *dto.AppSettingsDTO
	SaveProjectApiKey(projectID, projectName, projectNumber, apiKey string) (*dto.AppSettingsDTO, error)
	ToggleProjectApiKey(projectID string, enabled bool) (*dto.AppSettingsDTO, error)
	DeleteProjectApiKey(projectID string) (*dto.AppSettingsDTO, error)
	UpdateSystemSettings(autoAnalyze bool, primaryModel string, duplicateGuard bool, copyWorkers int) (*dto.AppSettingsDTO, error)
	ResetQuotaCooldowns() (*dto.AppSettingsDTO, error)
	IsAutoAnalyzeEnabled() bool
	IsDuplicateGuardEnabled() bool
	GetCopyWorkers() int
}

type persistedAppSettings struct {
	AutoAnalyze    bool                     `json:"auto_analyze"`
	PrimaryModel   string                   `json:"primary_model"`
	DuplicateGuard bool                     `json:"duplicate_guard"`
	CopyWorkers    int                      `json:"copy_workers"`
	Projects       []dto.ProjectApiKeyEntry `json:"projects"`
}

func settingsFilePath() string {
	cwd, err := os.Getwd()
	if err == nil {
		p := filepath.Join(cwd, "database", "ai_settings.json")
		if _, statErr := os.Stat(filepath.Dir(p)); statErr == nil {
			return p
		}
	}
	return "database/ai_settings.json"
}

// ComputeSmartWorkers automatically allocates optimal concurrency workers based on active API key count.
// - 1 key: 6 workers (distributes across 4 Flash-Lite models + overlap without hitting single-project RPM limit)
// - 2 keys: 8 workers
// - 3 keys: 12 workers
// - 4 keys: 14 workers
// - 5 keys: 16 workers
// - 6-7 keys: 18 workers
// - 8-9 keys: 20 workers
// - 10+ keys: 24 workers (40+ primary Flash-Lite routes, finishes 1,000 photos in ~10-18s)
func ComputeSmartWorkers(activeKeys int) int {
	switch {
	case activeKeys <= 0:
		return 4
	case activeKeys == 1:
		return 6
	case activeKeys == 2:
		return 8
	case activeKeys == 3:
		return 12
	case activeKeys == 4:
		return 14
	case activeKeys == 5:
		return 16
	case activeKeys <= 7:
		return 18
	case activeKeys <= 9:
		return 20
	default:
		return 24
	}
}

func defaultSeedProjects(envKeys []string) []dto.ProjectApiKeyEntry {
	type knownMeta struct {
		name   string
		number string
	}
	known := []knownMeta{
		{name: "Project 1 (Account 1)", number: "285328992732"},
		{name: "Project 2 (Account 1)", number: "917843332776"},
		{name: "Project 3 (Account 1)", number: "335976535875"},
		{name: "Project 4 (Account 1)", number: "163156631856"},
		{name: "Project 5 (Account 1)", number: "853499224099"},
		{name: "project Cross Platform Mobile6", number: "1061737622178"},
		{name: "project Cross Platform Mobile7", number: "898144085662"},
		{name: "project Cross Platform Mobile8", number: "1012454347291"},
		{name: "project Cross Platform Mobile9", number: "369073737348"},
		{name: "project Cross PlatformMobile10", number: "495572832530"},
	}

	var entries []dto.ProjectApiKeyEntry
	usedKey := make(map[string]bool)

	for i, km := range known {
		keyVal := ""
		if i < len(envKeys) && strings.TrimSpace(envKeys[i]) != "" {
			keyVal = strings.TrimSpace(envKeys[i])
		}
		if keyVal != "" {
			usedKey[keyVal] = true
		}
		entries = append(entries, dto.ProjectApiKeyEntry{
			ID:            fmt.Sprintf("proj-%d", i+1),
			ProjectName:   km.name,
			ProjectNumber: km.number,
			ApiKey:        keyVal,
			Enabled:       keyVal != "",
		})
	}

	// Append any extra keys from .env that were not in the first 10 known projects
	for _, ek := range envKeys {
		ek = strings.TrimSpace(ek)
		if ek == "" || usedKey[ek] {
			continue
		}
		usedKey[ek] = true
		placed := false
		for idx := range entries {
			if strings.TrimSpace(entries[idx].ApiKey) == "" {
				entries[idx].ApiKey = ek
				entries[idx].Enabled = true
				placed = true
				break
			}
		}
		if !placed {
			n := len(entries) + 1
			entries = append(entries, dto.ProjectApiKeyEntry{
				ID:            fmt.Sprintf("proj-%d", n),
				ProjectName:   fmt.Sprintf("Project %d", n),
				ProjectNumber: "",
				ApiKey:        ek,
				Enabled:       true,
			})
		}
	}

	return entries
}

func (s *geminiVisionService) initSettings(envKeys []string) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	path := settingsFilePath()
	var loaded persistedAppSettings
	loadedOk := false

	if data, err := os.ReadFile(path); err == nil {
		if jErr := json.Unmarshal(data, &loaded); jErr == nil && len(loaded.Projects) > 0 {
			loadedOk = true
		}
	}

	if !loadedOk {
		loaded = persistedAppSettings{
			AutoAnalyze:    true,
			PrimaryModel:   "auto",
			DuplicateGuard: true,
			CopyWorkers:    16,
			Projects:       defaultSeedProjects(envKeys),
		}
	} else {
		if loaded.CopyWorkers <= 0 {
			loaded.CopyWorkers = 16
		}
		if loaded.PrimaryModel == "" {
			loaded.PrimaryModel = "auto"
		}
	}

	s.settings = &loaded
	s.syncActiveKeysLocked(false)
	s.saveSettingsFileLocked()
}

func (s *geminiVisionService) saveSettingsFileLocked() {
	if s.settings == nil {
		return
	}
	path := settingsFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err == nil {
		_ = os.WriteFile(path, data, 0644)
	}
}

func (s *geminiVisionService) syncEnvFileLocked(activeKeys []string) {
	if len(activeKeys) == 0 {
		return
	}
	envPath := ".env"
	if _, err := os.Stat(envPath); err != nil {
		// If .env doesn't exist in built binary folder, no need to create unless writable
		return
	}
	content := fmt.Sprintf("DATABASE=database/photo-vault.db\nGEMINI_API_KEY=%s\nGEMINI_API_KEYS=%s\n",
		activeKeys[0],
		strings.Join(activeKeys, ","),
	)
	_ = os.WriteFile(envPath, []byte(content), 0644)
}

func (s *geminiVisionService) syncActiveKeysLocked(updateEnv bool) {
	if s.settings == nil {
		return
	}
	var active []string
	seen := make(map[string]bool)
	for _, p := range s.settings.Projects {
		k := strings.TrimSpace(p.ApiKey)
		if p.Enabled && k != "" && !seen[k] {
			seen[k] = true
			active = append(active, k)
		}
	}
	s.apiKeys = active
	if len(active) > 0 {
		s.apiKey = active[0]
	} else {
		s.apiKey = ""
	}

	// Update model ordering based on PrimaryModel setting
	pm := strings.TrimSpace(s.settings.PrimaryModel)
	var models []string
	seenModel := make(map[string]bool)
	if pm != "" && !strings.EqualFold(pm, "auto") {
		seenModel[pm] = true
		models = append(models, pm)
	}
	for _, dm := range defaultVisionModels {
		if !seenModel[dm] {
			seenModel[dm] = true
			models = append(models, dm)
		}
	}
	s.models = models
	if len(models) > 0 {
		s.modelName = models[0]
	}

	if updateEnv {
		s.syncEnvFileLocked(active)
	}
}

func extractApiKeysFromInput(raw string) []string {
	// Look for standard Google/Vertex API key patterns first (AQ.Ab8... or AIza...)
	re := regexp.MustCompile(`(AQ\.[A-Za-z0-9_\-]+|AIza[A-Za-z0-9_\-]+)`)
	matches := re.FindAllString(raw, -1)
	if len(matches) > 0 {
		var res []string
		seen := make(map[string]bool)
		for _, m := range matches {
			m = strings.TrimSpace(m)
			if m != "" && !seen[m] {
				seen[m] = true
				res = append(res, m)
			}
		}
		return res
	}

	// Fallback: split by comma, newline, or semicolon
	var res []string
	seen := make(map[string]bool)
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == ';'
	}) {
		k := strings.TrimSpace(part)
		k = strings.Trim(k, "\"'")
		if k != "" && !seen[k] {
			seen[k] = true
			res = append(res, k)
		}
	}
	return res
}

func (s *geminiVisionService) buildSettingsDTOLocked() *dto.AppSettingsDTO {
	if s.settings == nil {
		return &dto.AppSettingsDTO{}
	}

	s.cooldownMu.RLock()
	now := time.Now()
	cooldownCopy := make(map[string]time.Time, len(s.cooldowns))
	for k, v := range s.cooldowns {
		if now.Before(v) {
			cooldownCopy[k] = v
		}
	}
	s.cooldownMu.RUnlock()

	projectsDTO := make([]dto.ProjectApiKeyEntry, len(s.settings.Projects))
	activeKeys := 0
	seenActive := make(map[string]bool)

	for i, p := range s.settings.Projects {
		entry := p
		k := strings.TrimSpace(entry.ApiKey)
		if k == "" {
			entry.Status = "missing_key"
			entry.Enabled = false
		} else if !entry.Enabled {
			entry.Status = "disabled"
		} else {
			if !seenActive[k] {
				seenActive[k] = true
				activeKeys++
			}
			// Check cooldown status across models for this key
			cdCount := 0
			dailyCount := 0
			for _, m := range defaultVisionModels {
				if until, ok := cooldownCopy[k+"|"+m]; ok {
					cdCount++
					if until.Sub(now) > 30*time.Minute {
						dailyCount++
					}
				}
			}
			entry.CooldownCount = cdCount
			if dailyCount >= 4 {
				entry.Status = "daily_limit"
			} else if cdCount > 0 {
				entry.Status = "cooldown"
			} else {
				entry.Status = "active"
			}
		}
		projectsDTO[i] = entry
	}

	primaryRoutes := activeKeys * 4 // 4 Flash-Lite models per project
	totalRoutes := activeKeys * len(defaultVisionModels)
	allocatedWorkers := ComputeSmartWorkers(activeKeys)

	return &dto.AppSettingsDTO{
		AutoAnalyze:       s.settings.AutoAnalyze,
		PrimaryModel:      s.settings.PrimaryModel,
		DuplicateGuard:    s.settings.DuplicateGuard,
		CopyWorkers:       s.settings.CopyWorkers,
		ActiveKeyCount:    activeKeys,
		TotalProjectCount: len(projectsDTO),
		PrimaryRoutes:     primaryRoutes,
		TotalRoutes:       totalRoutes,
		AllocatedWorkers:  allocatedWorkers,
		BatchSize:         s.BatchSize(),
		EstimatedRPM:      activeKeys * 60,   // 4 Flash-Lite models * 15 RPM per project
		EstimatedRPD:      activeKeys * 2000, // 4 Flash-Lite models * 500 RPD per project
		Projects:          projectsDTO,
	}
}

func (s *geminiVisionService) GetAppSettings() *dto.AppSettingsDTO {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	return s.buildSettingsDTOLocked()
}

func (s *geminiVisionService) clearKeyCooldowns(apiKey string) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return
	}
	s.cooldownMu.Lock()
	defer s.cooldownMu.Unlock()
	prefix := apiKey + "|"
	changed := false
	for k := range s.cooldowns {
		if strings.HasPrefix(k, prefix) {
			delete(s.cooldowns, k)
			changed = true
		}
	}
	if changed {
		s.saveQuotaCacheLocked()
	}
}

// SaveProjectApiKey adds a new project + API key, or updates an existing project with a new API key.
// Also supports pasting multiple API keys at once.
func (s *geminiVisionService) SaveProjectApiKey(projectID, projectName, projectNumber, apiKey string) (*dto.AppSettingsDTO, error) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	projectID = strings.TrimSpace(projectID)
	projectName = strings.TrimSpace(projectName)
	projectNumber = strings.TrimSpace(projectNumber)
	keys := extractApiKeysFromInput(apiKey)

	if len(keys) == 0 && (projectID == "" || projectID == "new") && projectName == "" && projectNumber == "" {
		return nil, errors.New("กรุณาระบุ API Key หรือข้อมูลโปรเจกต์")
	}

	// Helper to update or create a single project entry
	applySingle := func(targetID, pName, pNum, keyVal string) {
		if keyVal != "" {
			s.clearKeyCooldowns(keyVal)
		}

		// 1. Try matching by explicit ID
		if targetID != "" && targetID != "new" {
			for i := range s.settings.Projects {
				if s.settings.Projects[i].ID == targetID {
					if pName != "" {
						s.settings.Projects[i].ProjectName = pName
					}
					if pNum != "" {
						s.settings.Projects[i].ProjectNumber = pNum
					}
					if keyVal != "" {
						s.settings.Projects[i].ApiKey = keyVal
						s.settings.Projects[i].Enabled = true
					}
					return
				}
			}
		}

		// 2. Try matching by ProjectNumber if provided
		if pNum != "" {
			for i := range s.settings.Projects {
				if s.settings.Projects[i].ProjectNumber == pNum {
					if pName != "" {
						s.settings.Projects[i].ProjectName = pName
					}
					if keyVal != "" {
						s.settings.Projects[i].ApiKey = keyVal
						s.settings.Projects[i].Enabled = true
					}
					return
				}
			}
		}

		// 3. Try matching if this exact API key already exists
		if keyVal != "" {
			for i := range s.settings.Projects {
				if strings.TrimSpace(s.settings.Projects[i].ApiKey) == keyVal {
					if pName != "" {
						s.settings.Projects[i].ProjectName = pName
					}
					if pNum != "" {
						s.settings.Projects[i].ProjectNumber = pNum
					}
					s.settings.Projects[i].Enabled = true
					return
				}
			}
		}

		// 4. Create new project entry
		nextNum := len(s.settings.Projects) + 1
		newID := fmt.Sprintf("proj-%d-%d", nextNum, time.Now().UnixMilli()%10000)
		if pName == "" {
			if pNum != "" {
				pName = fmt.Sprintf("Project %d (%s)", nextNum, pNum)
			} else {
				pName = fmt.Sprintf("Project %d", nextNum)
			}
		}
		s.settings.Projects = append(s.settings.Projects, dto.ProjectApiKeyEntry{
			ID:            newID,
			ProjectName:   pName,
			ProjectNumber: pNum,
			ApiKey:        keyVal,
			Enabled:       keyVal != "",
		})
	}

	if len(keys) <= 1 {
		singleKey := ""
		if len(keys) == 1 {
			singleKey = keys[0]
		}
		applySingle(projectID, projectName, projectNumber, singleKey)
	} else {
		// Multiple keys pasted at once!
		for idx, k := range keys {
			if idx == 0 && projectID != "" && projectID != "new" {
				applySingle(projectID, projectName, projectNumber, k)
				continue
			}
			// Fill any existing project slot that is missing a key first
			filledEmptySlot := false
			for i := range s.settings.Projects {
				if strings.TrimSpace(s.settings.Projects[i].ApiKey) == "" {
					s.settings.Projects[i].ApiKey = k
					s.settings.Projects[i].Enabled = true
					s.clearKeyCooldowns(k)
					filledEmptySlot = true
					break
				}
			}
			if !filledEmptySlot {
				applySingle("new", "", "", k)
			}
		}
	}

	s.syncActiveKeysLocked(true)
	s.saveSettingsFileLocked()
	dtoRes := s.buildSettingsDTOLocked()
	log.Printf("[AI-Settings] อัปเดตรายการ API Keys สำเร็จ: เปิดใช้งาน %d โปรเจกต์ (%d เส้นทางหลัก, %d AI Workers)",
		dtoRes.ActiveKeyCount, dtoRes.PrimaryRoutes, dtoRes.AllocatedWorkers)
	return dtoRes, nil
}

func (s *geminiVisionService) ToggleProjectApiKey(projectID string, enabled bool) (*dto.AppSettingsDTO, error) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	projectID = strings.TrimSpace(projectID)
	found := false
	for i := range s.settings.Projects {
		if s.settings.Projects[i].ID == projectID {
			if enabled && strings.TrimSpace(s.settings.Projects[i].ApiKey) == "" {
				return nil, errors.New("โปรเจกต์นี้ยังไม่มี API Key กรุณาใส่ API Key ก่อนเปิดใช้งาน")
			}
			s.settings.Projects[i].Enabled = enabled
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("ไม่พบโปรเจกต์ที่ระบุ")
	}

	s.syncActiveKeysLocked(true)
	s.saveSettingsFileLocked()
	dtoRes := s.buildSettingsDTOLocked()
	log.Printf("[AI-Settings] เปลี่ยนสถานะโปรเจกต์ %s -> enabled=%v (พร้อมใช้งาน %d โปรเจกต์, %d Workers)",
		projectID, enabled, dtoRes.ActiveKeyCount, dtoRes.AllocatedWorkers)
	return dtoRes, nil
}

func (s *geminiVisionService) DeleteProjectApiKey(projectID string) (*dto.AppSettingsDTO, error) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	projectID = strings.TrimSpace(projectID)
	var next []dto.ProjectApiKeyEntry
	for _, p := range s.settings.Projects {
		if p.ID != projectID {
			next = append(next, p)
		}
	}
	s.settings.Projects = next
	s.syncActiveKeysLocked(true)
	s.saveSettingsFileLocked()
	dtoRes := s.buildSettingsDTOLocked()
	log.Printf("[AI-Settings] ลบโปรเจกต์ %s สำเร็จ (คงเหลือที่เปิดใช้งาน %d โปรเจกต์, %d Workers)",
		projectID, dtoRes.ActiveKeyCount, dtoRes.AllocatedWorkers)
	return dtoRes, nil
}

func (s *geminiVisionService) UpdateSystemSettings(autoAnalyze bool, primaryModel string, duplicateGuard bool, copyWorkers int) (*dto.AppSettingsDTO, error) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	if copyWorkers < 1 {
		copyWorkers = 4
	}
	if copyWorkers > 64 {
		copyWorkers = 64
	}
	if strings.TrimSpace(primaryModel) == "" {
		primaryModel = "auto"
	}

	s.settings.AutoAnalyze = autoAnalyze
	s.settings.PrimaryModel = strings.TrimSpace(primaryModel)
	s.settings.DuplicateGuard = duplicateGuard
	s.settings.CopyWorkers = copyWorkers

	s.syncActiveKeysLocked(false)
	s.saveSettingsFileLocked()
	dtoRes := s.buildSettingsDTOLocked()
	log.Printf("[AI-Settings] บันทึกการตั้งค่าระบบ: autoAnalyze=%v, model=%s, duplicateGuard=%v, copyWorkers=%d",
		autoAnalyze, primaryModel, duplicateGuard, copyWorkers)
	return dtoRes, nil
}

func (s *geminiVisionService) ResetQuotaCooldowns() (*dto.AppSettingsDTO, error) {
	s.cooldownMu.Lock()
	s.cooldowns = make(map[string]time.Time)
	s.saveQuotaCacheLocked()
	s.cooldownMu.Unlock()

	log.Printf("[AI-Settings] รีเซ็ตสถานะพักโควตา (Cooldown/RPD) ของทุกโปรเจกต์และโมเดลเรียบร้อยแล้ว")
	return s.GetAppSettings(), nil
}

func (s *geminiVisionService) IsAutoAnalyzeEnabled() bool {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	if s.settings == nil {
		return true
	}
	return s.settings.AutoAnalyze
}

func (s *geminiVisionService) IsDuplicateGuardEnabled() bool {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	if s.settings == nil {
		return true
	}
	return s.settings.DuplicateGuard
}

func (s *geminiVisionService) GetCopyWorkers() int {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	if s.settings == nil || s.settings.CopyWorkers <= 0 {
		return 16
	}
	return s.settings.CopyWorkers
}
