package service

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

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
