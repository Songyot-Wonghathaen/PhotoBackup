package service

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"PhotoVault/model/model"
	"PhotoVault/repository"
)

// Helper to create a dummy test image
func createSampleImage(t *testing.T, dir, filename string) string {
	filePath := filepath.Join(dir, filename)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 255, G: 120, B: 0, A: 255})
		}
	}
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create image file: %v", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatalf("failed to encode jpeg: %v", err)
	}
	return filePath
}

// Test TagRepository operations (Part C Feature 13)
func TestTagRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := repository.NewTagRepository(db)
	photoRepo := repository.NewPhotoRepository(db)

	// 1. Create photo
	photo := &model.Photo{
		DestinationPath: "dest",
		FileName:        "test.jpg",
		RelPath:         "test.jpg",
		SizeBytes:       1024,
		Status:          "active",
		Description:     "ภาพทดสอบสีส้ม",
	}
	err := photoRepo.AddPhoto(ctx, photo)
	if err != nil {
		t.Fatalf("failed to add photo: %v", err)
	}

	// 2. FindOrCreateTags
	tags, err := tagRepo.FindOrCreateTags(ctx, []string{"แมว", "สัตว์เลี้ยง", "แมว"})
	if err != nil {
		t.Fatalf("failed to create tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 unique tags, got %d", len(tags))
	}

	// 3. Add PhotoTags
	err = tagRepo.AddPhotoTags(ctx, photo.ID, tags, "ai")
	if err != nil {
		t.Fatalf("failed to link photo tags: %v", err)
	}

	// 4. Retrieve tags by photo ID
	linkedTags, err := tagRepo.GetTagsByPhotoID(ctx, photo.ID)
	if err != nil {
		t.Fatalf("failed to get tags: %v", err)
	}
	if len(linkedTags) != 2 {
		t.Fatalf("expected 2 linked tags, got %d", len(linkedTags))
	}
}

// Test Feature 14: AI Failure Fallback (continues backup and leaves description empty)
func TestAiFailureFallback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir, err := os.MkdirTemp("", "photovault_ai_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.MkdirAll(dstDir, 0755)

	createSampleImage(t, srcDir, "cat.jpg")

	photoRepo := repository.NewPhotoRepository(db)
	tagRepo := repository.NewTagRepository(db)

	// Create AI service with invalid key to guarantee fallback behavior
	failingAi := &geminiVisionService{
		apiKey:    "INVALID_TEST_KEY_FOR_FALLBACK",
		modelName: "gemini-2.0-flash",
	}

	svc := NewBackupService(photoRepo, tagRepo, failingAi)
	_ = svc.SetSource(srcDir)
	_, _ = svc.SetDest(dstDir)

	// Execute Move
	summary, err := svc.Move([]string{"cat.jpg"}, false)
	if err != nil {
		t.Fatalf("backup should not fail even if AI fails: %v", err)
	}
	if summary.MovedFiles != 1 {
		t.Fatalf("expected 1 moved file, got %d", summary.MovedFiles)
	}

	// Verify photo is saved with empty description (Feature 14 requirement)
	photos, err := photoRepo.GetActivePhotosByDest(context.Background(), dstDir)
	if err != nil || len(photos) != 1 {
		t.Fatalf("expected 1 photo in DB, got %d", len(photos))
	}
	if photos[0].Description != "" {
		t.Errorf("expected empty description on AI failure, got %q", photos[0].Description)
	}
}

// Test Live AI Vision analysis with real GEMINI_API_KEY from .env
func TestLiveAiVision(t *testing.T) {
	ai := NewAiVisionService()
	if !ai.IsAvailable() {
		t.Skip("Skipping live AI test: GEMINI_API_KEY is not set")
	}

	tmpDir, err := os.MkdirTemp("", "photovault_live_ai_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	imgFile := createSampleImage(t, tmpDir, "orange_square.jpg")

	result, err := ai.AnalyzePhoto(context.Background(), imgFile)
	if err != nil {
		t.Logf("Note: Live AI request returned error (e.g. rate limit or invalid key): %v", err)
		return
	}

	t.Logf("Live AI Result:\n  Description: %s\n  Tags: %v", result.Description, result.Tags)
	if result.Description == "" {
		t.Errorf("expected non-empty description")
	}
}

func TestGridCollage1000Images(t *testing.T) {
	datasetDir := `C:\Users\keenz\Downloads\set_1000_image`
	entries, err := os.ReadDir(datasetDir)
	if err != nil || len(entries) == 0 {
		t.Skipf("Skipping 1000-image collage test: %v", err)
	}

	var paths []string
	for _, e := range entries {
		if !e.IsDir() && isImageFile(e.Name()) {
			paths = append(paths, filepath.Join(datasetDir, e.Name()))
		}
	}

	batchSize := 20
	numBatches := (len(paths) + batchSize - 1) / batchSize
	totalCollageBytes := 0

	for i := 0; i < len(paths); i += batchSize {
		end := i + batchSize
		if end > len(paths) {
			end = len(paths)
		}
		collage, err := buildGridCollageJPEG(paths[i:end], 160)
		if err != nil {
			t.Fatalf("failed to build collage for batch %d: %v", i/batchSize, err)
		}
		totalCollageBytes += len(collage)
	}

	t.Logf("Successfully built %d Grid Collages for %d images (avg %.1f KB per 20-image collage)",
		numBatches, len(paths), float64(totalCollageBytes)/float64(numBatches)/1024.0)

	if os.Getenv("RUN_LIVE_BATCH_TEST") == "1" {
		ai := NewAiVisionService()
		if batchAi, ok := ai.(BatchAiVisionService); ok && ai.IsAvailable() {
			sample20 := paths[:20]
			res, err := batchAi.AnalyzePhotoBatch(context.Background(), sample20)
			if err != nil {
				t.Logf("Live Batch 20 returned error: %v", err)
				return
			}
			for i := 0; i < 5 && i < len(res); i++ {
				if res[i] != nil {
					t.Logf("Photo [%d] %s -> Desc: %s | Tags: %v", i, filepath.Base(sample20[i]), res[i].Description, res[i].Tags)
				}
			}
		}
	}
}

func TestListModels(t *testing.T) {
	datasetDir := `C:\Users\keenz\Downloads\set_1000_image`
	entries, err := os.ReadDir(datasetDir)
	if err != nil || len(entries) < 25 {
		datasetDir = `D:\backup`
		entries, err = os.ReadDir(datasetDir)
	}
	if err != nil || len(entries) < 25 {
		t.Skipf("Dataset not found: %v", err)
	}

	var sample25 []string
	for _, e := range entries {
		if !e.IsDir() && isImageFile(e.Name()) {
			sample25 = append(sample25, filepath.Join(datasetDir, e.Name()))
			if len(sample25) == 25 {
				break
			}
		}
	}

	startCollage := time.Now()
	collageBytes, err := buildGridCollageJPEG(sample25, 220)
	if err != nil {
		t.Fatalf("buildGridCollageJPEG error: %v", err)
	}
	t.Logf("Built 25-photo collage (220x220 tile, 1100x1100 px) in %v, size: %.1f KB", time.Since(startCollage), float64(len(collageBytes))/1024.0)

	ai := NewAiVisionService().(*geminiVisionService)
	ai.models = []string{"gemini-3.5-flash-lite", "gemini-3.1-flash-lite"}
	ai.cooldowns = make(map[string]time.Time)

	startAPI := time.Now()
	res, err := ai.AnalyzePhotoBatch(context.Background(), sample25)
	if err != nil {
		t.Fatalf("AnalyzePhotoBatch failed: %v", err)
	}
	t.Logf("Gemini analyzed 25 photos in %v!", time.Since(startAPI))

	validCount := 0
	for i, r := range res {
		if r != nil && r.Description != "" {
			validCount++
			t.Logf("[%02d] %-28s -> Tags: %v | คำอธิบาย: %s", i, filepath.Base(sample25[i]), r.Tags, r.Description)
		} else {
			t.Logf("[%02d] %-28s -> MISSING!", i, filepath.Base(sample25[i]))
		}
	}
	t.Logf("Total valid analyzed in single batch: %d/25", validCount)
}

func TestSmartProjectAllocation(t *testing.T) {
	if got := ComputeSmartWorkers(1); got != 4 {
		t.Fatalf("expected 4 workers for 1 key, got %d", got)
	}
	if got := ComputeSmartWorkers(5); got != 12 {
		t.Fatalf("expected 12 workers for 5 keys, got %d", got)
	}
	if got := ComputeSmartWorkers(10); got != 16 {
		t.Fatalf("expected 16 workers for 10 keys, got %d", got)
	}

	keys := extractApiKeysFromInput("AQ.Ab8RN6KeyOne12345, AQ.Ab8RN6KeyTwo67890")
	if len(keys) != 2 {
		t.Fatalf("expected 2 extracted keys, got %d: %v", len(keys), keys)
	}

	// Verify resilient JSON recovery against extra ']' or prose brackets '[0]'
	malformed := `Here is tile [0]: [{"i":0,"d":"ภูเขาหิมะ","t":["ภูเขา","หิมะ"]},{"i":1,"d":"ทะเลทราย","t":["ทราย"]}}]]`
	items, err := parseBatchJsonItems(malformed)
	if err != nil || len(items) != 2 {
		t.Fatalf("expected 2 recovered items from malformed JSON, got %d (err=%v)", len(items), err)
	}
}


