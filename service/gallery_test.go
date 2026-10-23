package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"PhotoVault/controller/dto"
	"PhotoVault/model/model"
	"PhotoVault/repository"

	"gorm.io/gorm"
)

// seedGallery creates photos/tags/links for gallery tests and returns the gallery service + ids
func seedGallery(t *testing.T) (*GalleryService, *gorm.DB, string) {
	t.Helper()
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	destDir, err := os.MkdirTemp("", "gallery_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(destDir) })

	photoRepo := repository.NewPhotoRepository(db)
	_ = photoRepo

	now := time.Now()
	photos := []*model.Photo{
		{ID: 1, DestinationPath: destDir, FileName: "beach.jpg", RelPath: "beach.jpg", SizeBytes: 100, Sha256: "h1", BackedUpAt: now.Add(-3 * time.Hour), Status: "active", Description: "ชายหาด สนุก"},
		{ID: 2, DestinationPath: destDir, FileName: "mountain.jpg", RelPath: "mountain.jpg", SizeBytes: 200, Sha256: "h2", BackedUpAt: now.Add(-2 * time.Hour), Status: "active", Description: "ชายหาด ภูเขา"},
		{ID: 3, DestinationPath: destDir, FileName: "sunset.jpg", RelPath: "sunset.jpg", SizeBytes: 300, Sha256: "h3", BackedUpAt: now.Add(-1 * time.Hour), Status: "active", Description: "พระอาทิตย์ตก"},
		{ID: 4, DestinationPath: destDir, FileName: "full.jpg", RelPath: "full.jpg", SizeBytes: 400, Sha256: "h4", BackedUpAt: now, Status: "active", Description: "ชายหาด พระอาทิตย์ตก เงียบสงบ"},
		{ID: 5, DestinationPath: destDir, FileName: "missing.jpg", RelPath: "missing.jpg", SizeBytes: 500, Sha256: "h5", BackedUpAt: now, Status: "missing", Description: "ชายหาด พระอาทิตย์ตก"},
	}
	if err := db.Create(&photos).Error; err != nil {
		t.Fatalf("failed to seed photos: %v", err)
	}

	sea := &model.Tag{ID: 10, Name: "ทะเล"}
	beach := &model.Tag{ID: 11, Name: "ชายหาด"}
	sun := &model.Tag{ID: 12, Name: "พระอาทิตย์ตก"}
	if err := db.Create([]*model.Tag{sea, beach, sun}).Error; err != nil {
		t.Fatalf("failed to seed tags: %v", err)
	}

	links := []*model.PhotoTag{
		{PhotoID: 1, TagID: 10, Source: "ai"},
		{PhotoID: 1, TagID: 11, Source: "ai"},
		{PhotoID: 3, TagID: 12, Source: "ai"},
		{PhotoID: 4, TagID: 11, Source: "ai"},
		{PhotoID: 4, TagID: 12, Source: "ai"},
	}
	if err := db.Create(&links).Error; err != nil {
		t.Fatalf("failed to seed photo_tags: %v", err)
	}

	svc := NewGalleryService(repository.NewGalleryRepository(db))
	return svc, db, destDir
}

// Feature 15: multi-word description search uses AND semantics
func TestGallerySearchMultipleWords(t *testing.T) {
	svc, _, destDir := seedGallery(t)
	ctx := context.Background()

	result, err := svc.SearchPhotos(destDir, dto.SearchParams{Query: "ชายหาด พระอาทิตย์ตก"})
	if err != nil {
		t.Fatalf("SearchPhotos failed: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 total for AND words, got %d (%+v)", result.Total, result.Items)
	}
	if len(result.Items) != 1 || result.Items[0].FileName != "full.jpg" {
		t.Fatalf("expected full.jpg matched by AND words, got %+v", result.Items)
	}

	// Single word matches all active photos containing it (missing status excluded)
	single, err := svc.SearchPhotos(destDir, dto.SearchParams{Query: "ชายหาด"})
	if err != nil {
		t.Fatalf("SearchPhotos single word failed: %v", err)
	}
	if single.Total != 3 {
		t.Fatalf("expected 3 active photos containing ชายหาด, got %d", single.Total)
	}

	// Empty query returns all active photos (5 records, 1 missing excluded)
	all, err := svc.SearchPhotos(destDir, dto.SearchParams{})
	if err != nil {
		t.Fatalf("SearchPhotos empty query failed: %v", err)
	}
	if all.Total != 4 {
		t.Fatalf("expected 4 active photos, got %d", all.Total)
	}

	// Default paging + ordering: newest backed_up_at first
	if all.Page != 1 || all.PageSize != 40 {
		t.Fatalf("expected default paging 1/40, got %d/%d", all.Page, all.PageSize)
	}
	if all.Items[0].FileName != "full.jpg" {
		t.Fatalf("expected newest photo first, got %s", all.Items[0].FileName)
	}
	if all.TotalPages != 1 {
		t.Fatalf("expected 1 total page, got %d", all.TotalPages)
	}

	// Dest not set -> Thai error
	if _, err := svc.SearchPhotos("", dto.SearchParams{}); err == nil {
		t.Fatalf("expected Thai error when dest is empty")
	}

	_ = ctx
}

// Feature 17: multi-tag search requires ALL selected tags
func TestGallerySearchMultipleTags(t *testing.T) {
	svc, _, destDir := seedGallery(t)

	// Photos with BOTH ทะเล (10) and ชายหาด (11): only photo 1
	result, err := svc.SearchPhotos(destDir, dto.SearchParams{TagIDs: []int32{10, 11}})
	if err != nil {
		t.Fatalf("SearchPhotos by tags failed: %v", err)
	}
	if result.Total != 1 || result.Items[0].FileName != "beach.jpg" {
		t.Fatalf("expected only beach.jpg with both tags, got %+v", result.Items)
	}

	// ชายหาด (11) + พระอาทิตย์ตก (12): only photo 4
	result2, err := svc.SearchPhotos(destDir, dto.SearchParams{TagIDs: []int32{11, 12}})
	if err != nil {
		t.Fatalf("SearchPhotos by tags 2 failed: %v", err)
	}
	if result2.Total != 1 || result2.Items[0].FileName != "full.jpg" {
		t.Fatalf("expected only full.jpg with both tags, got %+v", result2.Items)
	}

	// Duplicate tag IDs are deduped (still "must have all" semantics)
	dup, err := svc.SearchPhotos(destDir, dto.SearchParams{TagIDs: []int32{11, 12, 12, 11, 0, -5}})
	if err != nil {
		t.Fatalf("SearchPhotos with duplicate tag ids failed: %v", err)
	}
	if dup.Total != 1 || dup.Items[0].FileName != "full.jpg" {
		t.Fatalf("expected dedupe to keep same result, got %+v", dup.Items)
	}

	// Combined words + tags
	combined, err := svc.SearchPhotos(destDir, dto.SearchParams{Query: "เงียบสงบ", TagIDs: []int32{11, 12}})
	if err != nil {
		t.Fatalf("SearchPhotos combined failed: %v", err)
	}
	if combined.Total != 1 || combined.Items[0].FileName != "full.jpg" {
		t.Fatalf("expected full.jpg for combined filters, got %+v", combined.Items)
	}
}

// LIKE wildcards (%, _, \) in user input must be matched literally
func TestGallerySearchEscapesWildcards(t *testing.T) {
	svc, db, destDir := seedGallery(t)

	if err := db.Create(&[]*model.Photo{
		{ID: 20, DestinationPath: destDir, FileName: "pct.jpg", RelPath: "pct.jpg", BackedUpAt: time.Now(), Status: "active", Description: "ความคืบหน้า 100% เสร็จสิ้น"},
		{ID: 21, DestinationPath: destDir, FileName: "under.jpg", RelPath: "under.jpg", BackedUpAt: time.Now(), Status: "active", Description: "ไฟล์ abc_def"},
		{ID: 22, DestinationPath: destDir, FileName: "plain.jpg", RelPath: "plain.jpg", BackedUpAt: time.Now(), Status: "active", Description: "ไฟล์ abcXdef"},
	}).Error; err != nil {
		t.Fatalf("failed to seed escape photos: %v", err)
	}

	// "100%" must match literally, not act as a wildcard
	pct, err := svc.SearchPhotos(destDir, dto.SearchParams{Query: "100%"})
	if err != nil {
		t.Fatalf("SearchPhotos percent failed: %v", err)
	}
	if pct.Total != 1 || pct.Items[0].FileName != "pct.jpg" {
		t.Fatalf("expected only pct.jpg for literal '100%%', got %+v", pct.Items)
	}

	// "abc_" must match literally, not act as single-char wildcard
	under, err := svc.SearchPhotos(destDir, dto.SearchParams{Query: "abc_"})
	if err != nil {
		t.Fatalf("SearchPhotos underscore failed: %v", err)
	}
	if under.Total != 1 || under.Items[0].FileName != "under.jpg" {
		t.Fatalf("expected only under.jpg for literal 'abc_', got %+v", under.Items)
	}
}

// Level 2: AI Query Expansion finds relevant photos even when exact phrase is missing
func TestGallerySearchAIQueryExpansion(t *testing.T) {
	svc, db, destDir := seedGallery(t)

	// Add photos: one has normal bread, one has burnt bread, one is unrelated
	now := time.Now()
	testPhotos := []*model.Photo{
		{DestinationPath: destDir, FileName: "toast.jpg", RelPath: "toast.jpg", SizeBytes: 150, Sha256: "h_toast", BackedUpAt: now.Add(-10 * time.Minute), Status: "active", Description: "ขนมปังปิ้งเนยหอมกรุ่น"},
		{DestinationPath: destDir, FileName: "burnt_bread.jpg", RelPath: "burnt_bread.jpg", SizeBytes: 160, Sha256: "h_burnt", BackedUpAt: now.Add(-20 * time.Minute), Status: "active", Description: "ขนมปังไหม้เกรียมสีดำ"},
	}
	if err := db.Create(&testPhotos).Error; err != nil {
		t.Fatalf("failed to insert test photos: %v", err)
	}

	// 1. Exact search without expansion for "ขนมปังไหม้" finds only the burnt bread
	exact, err := svc.SearchPhotos(destDir, dto.SearchParams{
		Query: "ขนมปังไหม้",
	})
	if err != nil {
		t.Fatalf("exact search failed: %v", err)
	}
	if exact.Total != 1 || exact.Items[0].FileName != "burnt_bread.jpg" {
		t.Fatalf("expected 1 exact match (burnt_bread.jpg), got %d (%+v)", exact.Total, exact.Items)
	}

	// 2. With AI Query Expansion: user searched "ขนมปังไหม้"
	// Both burnt bread and normal bread match, with burnt bread ranked #1!
	expanded, err := svc.SearchPhotos(destDir, dto.SearchParams{
		Query:         "ขนมปังไหม้",
		ExpandedTerms: []string{"ขนมปังไหม้", "ขนมปังเกรียม", "ขนมปังปิ้ง", "ขนมปัง", "toast"},
	})
	if err != nil {
		t.Fatalf("expanded search failed: %v", err)
	}
	if expanded.Total != 2 {
		t.Fatalf("expected 2 photos found via AI expansion, got %d", expanded.Total)
	}
	// Burnt bread has higher score than normal bread -> must be ranked first!
	if expanded.Items[0].FileName != "burnt_bread.jpg" {
		t.Fatalf("expected burnt_bread.jpg to rank #1 due to higher relevance score, got %s", expanded.Items[0].FileName)
	}
	if expanded.Items[1].FileName != "toast.jpg" {
		t.Fatalf("expected toast.jpg to rank #2 as fallback match, got %s", expanded.Items[1].FileName)
	}

	// 3. Scenario where NO burnt bread exists: delete burnt_bread.jpg
	_ = db.Where("file_name = ?", "burnt_bread.jpg").Delete(&model.Photo{})

	fallback, err := svc.SearchPhotos(destDir, dto.SearchParams{
		Query:         "ขนมปังไหม้",
		ExpandedTerms: []string{"ขนมปังไหม้", "ขนมปังเกรียม", "ขนมปังปิ้ง", "ขนมปัง", "toast"},
	})
	if err != nil {
		t.Fatalf("fallback search failed: %v", err)
	}
	// Even though "ขนมปังไหม้" has 0 exact matches, the user STILL gets toast.jpg!
	if fallback.Total != 1 || fallback.Items[0].FileName != "toast.jpg" {
		t.Fatalf("expected fallback to find toast.jpg, got total=%d (%+v)", fallback.Total, fallback.Items)
	}
}

// Feature 16: Tag Cloud counts active photos per tag, ordered by count DESC, limit respected
func TestGalleryTagCloud(t *testing.T) {
	svc, _, destDir := seedGallery(t)

	cloud, err := svc.GetTagCloud(destDir, 0)
	if err != nil {
		t.Fatalf("GetTagCloud failed: %v", err)
	}
	byName := make(map[string]int, len(cloud))
	for _, item := range cloud {
		byName[item.Name] = item.Count
	}
	if byName["ชายหาด"] != 2 || byName["พระอาทิตย์ตก"] != 2 || byName["ทะเล"] != 1 {
		t.Fatalf("unexpected tag counts: %+v", cloud)
	}
	if len(cloud) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(cloud))
	}
	// Highest count first
	if cloud[0].Count < cloud[len(cloud)-1].Count {
		t.Fatalf("expected counts ordered DESC, got %+v", cloud)
	}

	// Limit + defaults
	limited, err := svc.GetTagCloud(destDir, 2)
	if err != nil {
		t.Fatalf("GetTagCloud limited failed: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("expected 2 tags with limit 2, got %d", len(limited))
	}
	if _, err := svc.GetTagCloud("", 0); err == nil {
		t.Fatalf("expected Thai error for empty dest")
	}
}

// Feature 19: AddTag in one transaction, duplicates produce no extra rows
func TestGalleryAddTagNoDuplicate(t *testing.T) {
	svc, db, _ := seedGallery(t)

	// First add: creates the tag link with source 'user'
	detail, err := svc.AddTag(2, "  พระอาทิตย์ตก  ")
	if err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if detail == nil || detail.FileName != "mountain.jpg" {
		t.Fatalf("expected refreshed detail for mountain.jpg, got %+v", detail)
	}
	found := false
	for _, tag := range detail.Tags {
		if tag.Name == "พระอาทิตย์ตก" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected พระอาทิตย์ตก in photo tags, got %+v", detail.Tags)
	}

	// Second add of the SAME tag: no duplicate row
	if _, err := svc.AddTag(2, "พระอาทิตย์ยก"); err != nil {
		t.Fatalf("AddTag second failed: %v", err)
	}
	if _, err := svc.AddTag(2, "พระอาทิตย์ตก"); err != nil {
		t.Fatalf("AddTag third failed: %v", err)
	}

	var count int64
	if err := db.Model(&model.PhotoTag{}).
		Where("photo_id = ? AND tag_id IN (SELECT id FROM tags WHERE name = ?)", 2, "พระอาทิตย์ตก").
		Count(&count).Error; err != nil {
		t.Fatalf("count photo_tags failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 photo_tags row (no duplicate), got %d", count)
	}

	var source string
	if err := db.Model(&model.PhotoTag{}).
		Where("photo_id = ? AND tag_id IN (SELECT id FROM tags WHERE name = ?)", 2, "พระอาทิตย์ตก").
		Pluck("source", &source).Error; err != nil {
		t.Fatalf("pluck source failed: %v", err)
	}
	if source != "user" {
		t.Fatalf("expected source 'user', got %s", source)
	}

	// Whitespace collapsing + reuse of existing tag (no new tags row)
	var tagCountBefore int64
	_ = db.Model(&model.Tag{}).Count(&tagCountBefore)
	if _, err := svc.AddTag(2, "ทะเล"); err != nil {
		t.Fatalf("AddTag existing failed: %v", err)
	}
	var tagCountAfter int64
	_ = db.Model(&model.Tag{}).Count(&tagCountAfter)
	if tagCountAfter != tagCountBefore {
		t.Fatalf("expected no new tag row for existing tag, before=%d after=%d", tagCountBefore, tagCountAfter)
	}

	// Validation errors in Thai
	if _, err := svc.AddTag(2, "   "); err == nil {
		t.Fatalf("expected Thai error for blank tag name")
	}
	longName := make([]rune, 51)
	for i := range longName {
		longName[i] = 'a'
	}
	if _, err := svc.AddTag(2, string(longName)); err == nil {
		t.Fatalf("expected Thai error for too long tag name")
	}
	if _, err := svc.AddTag(0, "x"); err == nil {
		t.Fatalf("expected Thai error for invalid photo id")
	}
}

// Feature 19: UpdateDescription rules (trim, empty allowed, max 2000 runes)
func TestGalleryUpdateDescription(t *testing.T) {
	svc, db, destDir := seedGallery(t)
	_ = destDir

	if err := svc.UpdateDescription(1, "  คำอธิบายใหม่  "); err != nil {
		t.Fatalf("UpdateDescription failed: %v", err)
	}
	var photo model.Photo
	if err := db.First(&photo, 1).Error; err != nil {
		t.Fatalf("reload photo failed: %v", err)
	}
	if photo.Description != "คำอธิบายใหม่" {
		t.Fatalf("expected trimmed description, got %q", photo.Description)
	}

	// Empty description is allowed
	if err := svc.UpdateDescription(1, "   "); err != nil {
		t.Fatalf("UpdateDescription blank failed: %v", err)
	}
	_ = db.First(&photo, 1)
	if photo.Description != "" {
		t.Fatalf("expected empty description, got %q", photo.Description)
	}

	// Over 2000 runes -> Thai error
	tooLong := ""
	for i := 0; i < 2001; i++ {
		tooLong += "ก"
	}
	if err := svc.UpdateDescription(1, tooLong); err == nil {
		t.Fatalf("expected Thai error for description over 2000 runes")
	}
	if err := svc.UpdateDescription(0, "x"); err == nil {
		t.Fatalf("expected Thai error for invalid photo id")
	}
}

// Feature 19: RemoveTag removes the link; Feature 18: GetPhotoDetail + SuggestTags
func TestGalleryRemoveTagAndDetail(t *testing.T) {
	svc, db, _ := seedGallery(t)

	if err := svc.RemoveTag(0, 10); err == nil {
		t.Fatalf("expected Thai error for invalid photo id")
	}
	if err := svc.RemoveTag(1, 0); err == nil {
		t.Fatalf("expected Thai error for invalid tag id")
	}

	if err := svc.RemoveTag(1, 10); err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}
	var count int64
	_ = db.Model(&model.PhotoTag{}).Where("photo_id = ? AND tag_id = ?", 1, 10).Count(&count)
	if count != 0 {
		t.Fatalf("expected photo_tags link removed, got %d", count)
	}

	detail, err := svc.GetPhotoDetail(1)
	if err != nil {
		t.Fatalf("GetPhotoDetail failed: %v", err)
	}
	if detail.ID != 1 || detail.FileName != "beach.jpg" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
	for _, tag := range detail.Tags {
		if tag.ID == 10 {
			t.Fatalf("tag ทะเล should be removed from detail, got %+v", detail.Tags)
		}
	}
	if detail.ThumbURL == "" || detail.FullPath == "" || detail.BackedUpAt == "" {
		t.Fatalf("expected full path/thumb/backed_at in detail, got %+v", detail)
	}

	// Missing photo -> Thai error
	if _, err := svc.GetPhotoDetail(9999); err == nil {
		t.Fatalf("expected Thai error for missing photo")
	}

	// SuggestTags prefix
	suggest, err := svc.SuggestTags("ช", 5)
	if err != nil {
		t.Fatalf("SuggestTags failed: %v", err)
	}
	if len(suggest) == 0 {
		t.Fatalf("expected prefix suggestions")
	}
	for _, tag := range suggest {
		if !strings.HasPrefix(tag.Name, "ช") {
			t.Fatalf("expected suggestions starting with ช, got %s", tag.Name)
		}
	}

	// Escaped prefix behaves literally
	if _, err := svc.SuggestTags("%", 5); err != nil {
		t.Fatalf("SuggestTags escaped prefix failed: %v", err)
	}
}

// Feature 18: gallery repository direct check (Preload PhotoTags.Tag)
func TestGalleryRepoGetByIDPreloadsTags(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	photo := &model.Photo{DestinationPath: "D", FileName: "a.jpg", RelPath: "a.jpg", BackedUpAt: time.Now(), Status: "active", Description: "x"}
	_ = db.Create(photo).Error
	tag := &model.Tag{Name: "ทดสอบ"}
	_ = db.Create(tag).Error
	_ = db.Create(&model.PhotoTag{PhotoID: photo.ID, TagID: tag.ID, Source: "ai"}).Error

	repo := repository.NewGalleryRepository(db)
	detail, err := repo.GetByID(context.Background(), photo.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if len(detail.Tags) != 1 || detail.Tags[0].Name != "ทดสอบ" {
		t.Fatalf("expected preloaded tag, got %+v", detail.Tags)
	}
}
