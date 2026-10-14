package server

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"PhotoVault/model/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupImageServer creates a temp SQLite DB, seeds photos on disk and returns
// a running test server for the image handler
func setupImageServer(t *testing.T) (*httptest.Server, string, *gorm.DB) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "image_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Photo{}, &model.Tag{}, &model.PhotoTag{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	// Close the SQLite handle before TempDir cleanup (Windows locks open files)
	if sqlDB, dbErr := db.DB(); dbErr == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	destDir := filepath.Join(tmpDir, "photos")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	now := time.Now()
	photos := []*model.Photo{
		// 1) real large JPEG (2000x1500)
		{ID: 1, DestinationPath: destDir, FileName: "photo1.jpg", RelPath: "photo1.jpg", SizeBytes: 1, Sha256: "s1", BackedUpAt: now, Status: "active", Description: "large jpeg"},
		// 2) real small PNG (300x200) - already within thumb size
		{ID: 2, DestinationPath: destDir, FileName: "photo2.png", RelPath: "photo2.png", SizeBytes: 1, Sha256: "s2", BackedUpAt: now, Status: "active", Description: "small png"},
		// 3) real medium GIF (800x600)
		{ID: 3, DestinationPath: destDir, FileName: "photo3.gif", RelPath: "photo3.gif", SizeBytes: 1, Sha256: "s3", BackedUpAt: now, Status: "active", Description: "gif"},
		// 4) undecodable "image" (garbage bytes with jpg extension)
		{ID: 4, DestinationPath: destDir, FileName: "photo4.jpg", RelPath: "photo4.jpg", SizeBytes: 1, Sha256: "s4", BackedUpAt: now, Status: "active", Description: "broken jpeg"},
		// 5) active record but file missing on disk
		{ID: 5, DestinationPath: destDir, FileName: "photo5.jpg", RelPath: "photo5.jpg", SizeBytes: 1, Sha256: "s5", BackedUpAt: now, Status: "active", Description: "missing file"},
		// 6) not active (deleted)
		{ID: 6, DestinationPath: destDir, FileName: "photo6.png", RelPath: "photo6.png", SizeBytes: 1, Sha256: "s6", BackedUpAt: now, Status: "deleted", Description: "deleted"},
		// 7) traversal attempt
		{ID: 7, DestinationPath: destDir, FileName: "../outside.png", RelPath: "../outside.png", SizeBytes: 1, Sha256: "s7", BackedUpAt: now, Status: "active", Description: "traversal"},
		// 8) legit nested subfolder file
		{ID: 8, DestinationPath: destDir, FileName: "sub/photo8.png", RelPath: "sub/photo8.png", SizeBytes: 1, Sha256: "s8", BackedUpAt: now, Status: "active", Description: "nested"},
	}
	if err := db.Create(&photos).Error; err != nil {
		t.Fatalf("failed to seed photos: %v", err)
	}

	// Real image files (encoded with stdlib encoders)
	if err := writeGradientJPEG(filepath.Join(destDir, "photo1.jpg"), 2000, 1500); err != nil {
		t.Fatalf("failed to create photo1.jpg: %v", err)
	}
	if err := writeSolidPNG(filepath.Join(destDir, "photo2.png"), 300, 200); err != nil {
		t.Fatalf("failed to create photo2.png: %v", err)
	}
	if err := writeGradientGIF(filepath.Join(destDir, "photo3.gif"), 800, 600); err != nil {
		t.Fatalf("failed to create photo3.gif: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "photo4.jpg"), []byte("this-is-not-an-image"), 0o644); err != nil {
		t.Fatalf("failed to create photo4.jpg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "photo6.png"), []byte("png-bytes"), 0o644); err != nil {
		t.Fatalf("failed to create photo6.png: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "outside.png"), []byte("outside"), 0o644); err != nil {
		t.Fatalf("failed to create outside.png: %v", err)
	}
	subDir := filepath.Join(destDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	if err := writeSolidPNG(filepath.Join(subDir, "photo8.png"), 900, 500); err != nil {
		t.Fatalf("failed to create sub/photo8.png: %v", err)
	}

	// Legacy handler receives everything the image routes do not match
	legacy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("legacy"))
	})

	ts := httptest.NewServer(NewHandler(db, legacy))
	t.Cleanup(ts.Close)

	return ts, destDir, db
}

func writeGradientJPEG(path string, w, h int) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: uint8((x + y) * 255 / (w + h)),
				A: 255,
			})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
}

func writeSolidPNG(path string, w, h int) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 40, G: 120, B: 200, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func writeGradientGIF(path string, w, h int) error {
	img := image.NewPaletted(image.Rect(0, 0, w, h), color.Palette{
		color.RGBA{R: 10, G: 10, B: 10, A: 255},
		color.RGBA{R: 200, G: 60, B: 60, A: 255},
		color.RGBA{R: 60, G: 200, B: 120, A: 255},
	})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetColorIndex(x, y, uint8((x+y)%3))
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.Encode(f, img, nil)
}

// snapshotCacheFiles lists the current thumbnail cache files
func snapshotCacheFiles(t *testing.T) map[string]bool {
	t.Helper()
	dir, err := thumbCacheDir()
	if err != nil {
		t.Fatalf("thumbCacheDir failed: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(entries))
	for _, e := range entries {
		set[e.Name()] = true
	}
	return set
}

// cleanupNewCacheFiles removes only cache files created during the test
func cleanupNewCacheFiles(t *testing.T, before map[string]bool) {
	t.Helper()
	dir, err := thumbCacheDir()
	if err != nil {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !before[e.Name()] {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func get(t *testing.T, ts *httptest.Server, path string) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("reading body of %s failed: %v", path, err)
	}
	return resp, body
}

// TestImageServesThreeRealOriginals verifies GET /image/{id} with 3 real images
func TestImageServesThreeRealOriginals(t *testing.T) {
	ts, destDir, _ := setupImageServer(t)

	cases := []struct {
		id          string
		file        string
		contentType string
	}{
		{"1", "photo1.jpg", "image/jpeg"},
		{"2", "photo2.png", "image/png"},
		{"3", "photo3.gif", "image/gif"},
	}

	for _, c := range cases {
		resp, body := get(t, ts, "/image/"+c.id)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for /image/%s, got %d", c.id, resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != c.contentType {
			t.Fatalf("expected Content-Type %s for /image/%s, got %s", c.contentType, c.id, ct)
		}
		original, err := os.ReadFile(filepath.Join(destDir, c.file))
		if err != nil {
			t.Fatalf("failed to read original %s: %v", c.file, err)
		}
		if !bytes.Equal(body, original) {
			t.Fatalf("expected original bytes for /image/%s, got %d bytes (want %d)", c.id, len(body), len(original))
		}
	}
}

// TestThumbShrinksCachesAndFallsBack verifies GET /thumb/{id} behavior
func TestThumbShrinksCachesAndFallsBack(t *testing.T) {
	before := snapshotCacheFiles(t)
	defer cleanupNewCacheFiles(t, before)

	ts, destDir, _ := setupImageServer(t)

	// 1) Large JPEG 2000x1500 -> shrunk to 400x300 JPEG, cached
	resp, body := get(t, ts, "/thumb/1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /thumb/1, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("expected image/jpeg for /thumb/1, got %s", ct)
	}
	thumb, err := jpeg.Decode(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to decode thumbnail jpeg: %v", err)
	}
	if b := thumb.Bounds(); b.Dx() != 400 || b.Dy() != 300 {
		t.Fatalf("expected 400x300 thumbnail, got %dx%d", b.Dx(), b.Dy())
	}

	// Cache file was written
	after := snapshotCacheFiles(t)
	created := 0
	for name := range after {
		if !before[name] {
			created++
		}
	}
	if created == 0 {
		t.Fatalf("expected at least one new cache file in thumbCacheDir")
	}

	// Cache is actually served: replace the expected cache file with marker
	// bytes and verify the handler returns them without re-shrinking
	info, err := os.Stat(filepath.Join(destDir, "photo1.jpg"))
	if err != nil {
		t.Fatalf("failed to stat source: %v", err)
	}
	cachePath, ok := thumbCachePath(1, info)
	if !ok {
		t.Fatalf("expected cache path for photo 1")
	}
	marker := []byte("CACHED-THUMB-MARKER")
	if err := os.WriteFile(cachePath, marker, 0o644); err != nil {
		t.Fatalf("failed to write marker cache: %v", err)
	}
	resp2, body2 := get(t, ts, "/thumb/1")
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected cached 200 for /thumb/1, got %d", resp2.StatusCode)
	}
	if !bytes.Equal(body2, marker) {
		t.Fatalf("expected marker bytes from cache, got %d bytes", len(body2))
	}

	// 2) Small PNG 300x200 -> original passthrough (within 400px)
	resp3, body3 := get(t, ts, "/thumb/2")
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /thumb/2, got %d", resp3.StatusCode)
	}
	if ct := resp3.Header.Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected image/png passthrough for /thumb/2, got %s", ct)
	}
	small, err := png.Decode(bytes.NewReader(body3))
	if err != nil {
		t.Fatalf("failed to decode passthrough png: %v", err)
	}
	if b := small.Bounds(); b.Dx() != 300 || b.Dy() != 200 {
		t.Fatalf("expected 300x200 passthrough, got %dx%d", b.Dx(), b.Dy())
	}

	// 3) GIF 800x600 -> shrunk to 400x300 JPEG
	resp4, body4 := get(t, ts, "/thumb/3")
	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /thumb/3, got %d", resp4.StatusCode)
	}
	if ct := resp4.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("expected image/jpeg for /thumb/3, got %s", ct)
	}
	gifThumb, err := jpeg.Decode(bytes.NewReader(body4))
	if err != nil {
		t.Fatalf("failed to decode gif thumbnail: %v", err)
	}
	if b := gifThumb.Bounds(); b.Dx() != 400 || b.Dy() != 300 {
		t.Fatalf("expected 400x300 thumbnail for gif, got %dx%d", b.Dx(), b.Dy())
	}
}

// TestThumbFallbackForUndecodable verifies undecodable images fall back to the original
func TestThumbFallbackForUndecodable(t *testing.T) {
	before := snapshotCacheFiles(t)
	defer cleanupNewCacheFiles(t, before)

	ts, destDir, _ := setupImageServer(t)

	resp, body := get(t, ts, "/thumb/4")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 fallback for /thumb/4, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("expected image/jpeg content type for fallback, got %s", ct)
	}
	original, err := os.ReadFile(filepath.Join(destDir, "photo4.jpg"))
	if err != nil {
		t.Fatalf("failed to read original: %v", err)
	}
	if !bytes.Equal(body, original) {
		t.Fatalf("expected original bytes as fallback, got %d bytes (want %d)", len(body), len(original))
	}
}

// TestNotFoundAndTraversal verifies 404 handling and traversal protection
func TestNotFoundAndTraversal(t *testing.T) {
	ts, _, _ := setupImageServer(t)

	notFound := []struct {
		name string
		path string
	}{
		{"unknown id", "/image/99999"},
		{"inactive photo", "/image/6"},
		{"inactive thumb", "/thumb/6"},
		{"missing file", "/image/5"},
		{"missing file thumb", "/thumb/5"},
		{"non numeric id", "/image/abc"},
		{"empty id", "/image/"},
		{"nested id path", "/image/1/2"},
		{"negative id", "/image/-1"},
		{"traversal", "/image/7"},
		{"traversal thumb", "/thumb/7"},
	}
	for _, c := range notFound {
		resp, _ := get(t, ts, c.path)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 for %s (%s), got %d", c.name, c.path, resp.StatusCode)
		}
	}

	// Nested legitimate path inside destination
	resp, _ := get(t, ts, "/image/8")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for nested /image/8, got %d", resp.StatusCode)
	}
}

// TestMethodNotAllowedAndFallback verifies method guard and legacy delegation
func TestMethodNotAllowedAndFallback(t *testing.T) {
	ts, _, _ := setupImageServer(t)

	resp, err := http.Post(ts.URL+"/image/1", "text/plain", nil)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for POST /image/1, got %d", resp.StatusCode)
	}

	// Unknown routes fall through to the legacy handler
	resp2, body2 := get(t, ts, "/api/thumbnail?path=x")
	if resp2.StatusCode != http.StatusTeapot {
		t.Fatalf("expected legacy handler (418) for /api/thumbnail, got %d", resp2.StatusCode)
	}
	if string(body2) != "legacy" {
		t.Fatalf("expected legacy body, got %q", string(body2))
	}
}
