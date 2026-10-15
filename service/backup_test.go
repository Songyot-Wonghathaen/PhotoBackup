package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"PhotoVault/model/model"
	"PhotoVault/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir, err := os.MkdirTemp("", "photovault_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect sqlite: %v", err)
	}

	err = db.AutoMigrate(&model.Photo{}, &model.Tag{}, &model.PhotoTag{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// Test 1: State Guard (Feature 11) - Ensure no crash/panic when paths not set
func TestStateGuard(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPhotoRepository(db)
	svc := NewBackupService(repo)

	// ListSource without source set
	_, err := svc.ListSource()
	if err == nil {
		t.Errorf("expected error when source not set, got nil")
	}

	// ListDest without dest set
	_, err = svc.ListDest("")
	if err == nil {
		t.Errorf("expected error when dest not set, got nil")
	}

	// Move without source or dest
	_, err = svc.Move(nil, true)
	if err == nil {
		t.Errorf("expected error when source/dest not set, got nil")
	}

	// CheckIntegrity without dest set
	_, err = svc.CheckIntegrity("")
	if err == nil {
		t.Errorf("expected error when dest not set, got nil")
	}
}

// Test 2: Source selection & Image listing (Feature 4)
func TestSourceListing(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	srcDir, _ := os.MkdirTemp("", "src_test_*")
	defer os.RemoveAll(srcDir)

	_ = os.WriteFile(filepath.Join(srcDir, "photo1.jpg"), []byte("image-data-1"), 0644)
	_ = os.WriteFile(filepath.Join(srcDir, "photo2.png"), []byte("image-data-2"), 0644)
	_ = os.WriteFile(filepath.Join(srcDir, "doc.pdf"), []byte("not-an-image"), 0644)

	repo := repository.NewPhotoRepository(db)
	svc := NewBackupService(repo)

	err := svc.SetSource(srcDir)
	if err != nil {
		t.Fatalf("SetSource failed: %v", err)
	}

	images, err := svc.ListSource()
	if err != nil {
		t.Fatalf("ListSource failed: %v", err)
	}

	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d (%v)", len(images), images)
	}
}

// Test 3: Move with Timer, Concurrency Worker Pool, Duplicate prevention, and GORM Batch (Features 6, 7, 8, 9)
func TestMoveOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	srcDir, _ := os.MkdirTemp("", "src_move_*")
	destDir, _ := os.MkdirTemp("", "dest_move_*")
	defer os.RemoveAll(srcDir)
	defer os.RemoveAll(destDir)

	// Create 5 test files in source
	testFiles := []string{"pic1.jpg", "pic2.png", "pic3.webp", "pic4.jpg", "pic5.png"}
	for _, fname := range testFiles {
		_ = os.WriteFile(filepath.Join(srcDir, fname), []byte("content-"+fname), 0644)
	}

	repo := repository.NewPhotoRepository(db)
	svc := NewBackupService(repo)

	_ = svc.SetSource(srcDir)
	_, _ = svc.SetDest(destDir)

	// First move: All 5 files
	summary, err := svc.Move(nil, true)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	if summary.MovedFiles != 5 {
		t.Errorf("expected 5 moved files, got %d", summary.MovedFiles)
	}
	if summary.DurationFormatted == "" {
		t.Errorf("expected formatted duration, got empty")
	}

	// Verify files exist in dest and deleted from source
	for _, fname := range testFiles {
		if _, err := os.Stat(filepath.Join(destDir, fname)); os.IsNotExist(err) {
			t.Errorf("file %s should exist in destination", fname)
		}
		if _, err := os.Stat(filepath.Join(srcDir, fname)); !os.IsNotExist(err) {
			t.Errorf("file %s should be deleted from source", fname)
		}
	}

	// Verify DB records
	dbPhotos, err := svc.ListDB(destDir)
	if err != nil {
		t.Fatalf("ListDB failed: %v", err)
	}
	if len(dbPhotos) != 5 {
		t.Errorf("expected 5 DB records, got %d", len(dbPhotos))
	}
	for _, p := range dbPhotos {
		if p.Status != "active" {
			t.Errorf("expected status 'active', got %s", p.Status)
		}
		if p.Sha256 == "" {
			t.Errorf("expected sha256 to be populated")
		}
	}

	// Test Duplicate prevention (Feature 7):
	// Recreate a file with same name in source and attempt move again
	_ = os.WriteFile(filepath.Join(srcDir, "pic1.jpg"), []byte("new-content"), 0644)
	summary2, err := svc.Move([]string{"pic1.jpg"}, false)
	if err != nil {
		t.Fatalf("Move duplicate failed: %v", err)
	}
	if summary2.SkippedFiles != 1 {
		t.Errorf("expected 1 skipped duplicate file, got %d", summary2.SkippedFiles)
	}
}

// Test 4: Integrity Check (Feature 5 & 21) - Detects missing files and updates status to 'missing'
func TestIntegrityCheck(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	destDir, _ := os.MkdirTemp("", "dest_integ_*")
	defer os.RemoveAll(destDir)

	repo := repository.NewPhotoRepository(db)
	svc := NewBackupService(repo)

	// Add photo record directly to DB
	photo := &model.Photo{
		DestinationPath: destDir,
		FileName:        "missing_file.jpg",
		RelPath:         "missing_file.jpg",
		SizeBytes:       100,
		Sha256:          "abc123hash",
		BackedUpAt:      time.Now(),
		Status:          "active",
	}
	_ = repo.AddPhoto(context.Background(), photo)

	// Run integrity check on destDir (file does not exist on disk)
	result, err := svc.CheckIntegrity(destDir)
	if err != nil {
		t.Fatalf("CheckIntegrity failed: %v", err)
	}

	if result.IsExactMatch {
		t.Errorf("expected integrity check to fail due to missing file")
	}
	if len(result.MissingInDisk) != 1 || result.MissingInDisk[0] != "missing_file.jpg" {
		t.Errorf("expected missing_file.jpg in MissingInDisk, got %v", result.MissingInDisk)
	}

	// Verify status was updated to 'missing' in DB
	allPhotos, _ := repo.GetAllPhotosByDest(context.Background(), destDir)
	if len(allPhotos) == 1 && allPhotos[0].Status != "missing" {
		t.Errorf("expected photo status in DB to be 'missing', got %s", allPhotos[0].Status)
	}
}

// Test 5: Soft Delete with DB status update (Feature 23 & 25)
func TestDeleteOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	destDir, _ := os.MkdirTemp("", "dest_del_*")
	defer os.RemoveAll(destDir)

	// Create physical file
	_ = os.WriteFile(filepath.Join(destDir, "to_delete.png"), []byte("data"), 0644)

	repo := repository.NewPhotoRepository(db)
	svc := NewBackupService(repo)

	photo := &model.Photo{
		DestinationPath: destDir,
		FileName:        "to_delete.png",
		RelPath:         "to_delete.png",
		SizeBytes:       4,
		Sha256:          "hash",
		BackedUpAt:      time.Now(),
		Status:          "active",
	}
	_ = repo.AddPhoto(context.Background(), photo)
	_, _ = svc.SetDest(destDir)

	// Delete file
	delCount, err := svc.Delete([]string{"to_delete.png"}, false)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if delCount != 1 {
		t.Errorf("expected 1 file deleted, got %d", delCount)
	}

	// Verify file is removed from disk
	if _, err := os.Stat(filepath.Join(destDir, "to_delete.png")); !os.IsNotExist(err) {
		t.Errorf("file should be removed from disk")
	}

	// Verify status is 'deleted' in DB (row is preserved for history)
	allPhotos, _ := repo.GetAllPhotosByDest(context.Background(), destDir)
	if len(allPhotos) == 1 && allPhotos[0].Status != "deleted" {
		t.Errorf("expected DB status to be 'deleted', got %s", allPhotos[0].Status)
	}
}
