package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"PhotoVault/controller/dto"
	modelGen "PhotoVault/model/model"
	"PhotoVault/model/query"
	"PhotoVault/repository"

	"gorm.io/gorm"
)

const (
	defaultGalleryPageSize = 40
	maxGalleryPageSize     = 100
	defaultTagCloudLimit   = 50
	maxTagCloudLimit       = 200
	defaultSuggestLimit    = 10
	maxSuggestLimit        = 50
	maxDescriptionRunes    = 2000
	maxTagNameRunes        = 50
)

type GalleryService struct {
	ctx         context.Context
	galleryRepo repository.GalleryRepository
	mu          sync.RWMutex
}

func NewGalleryService(galleryRepo repository.GalleryRepository) *GalleryService {
	return &GalleryService{
		ctx:         context.Background(),
		galleryRepo: galleryRepo,
	}
}

func (s *GalleryService) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

func (s *GalleryService) getCtx() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

// uniqueTagIDs drops non-positive and duplicated tag IDs
func uniqueTagIDs(ids []int32) []int32 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int32]bool, len(ids))
	result := make([]int32, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

// SearchPhotos searches active photos by description words (AND) and/or tags (must have all)
// Feature 15: description search | Feature 17: multi-tag filter
func (s *GalleryService) SearchPhotos(dest string, params dto.SearchParams) (*dto.SearchResult, error) {
	if strings.TrimSpace(dest) == "" {
		return nil, errors.New("กรุณาเลือกโฟลเดอร์ปลายทางก่อนค้นหา")
	}
	params.Query = strings.TrimSpace(params.Query)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = defaultGalleryPageSize
	}
	if params.PageSize > maxGalleryPageSize {
		params.PageSize = maxGalleryPageSize
	}
	params.TagIDs = uniqueTagIDs(params.TagIDs)
	return s.galleryRepo.Search(s.getCtx(), dest, params)
}

// GetTagCloud returns tags ordered by photo count (Feature 16)
func (s *GalleryService) GetTagCloud(dest string, limit int) ([]dto.TagItem, error) {
	if strings.TrimSpace(dest) == "" {
		return nil, errors.New("กรุณาเลือกโฟลเดอร์ปลายทางก่อนดึง Tag Cloud")
	}
	if limit <= 0 {
		limit = defaultTagCloudLimit
	}
	if limit > maxTagCloudLimit {
		limit = maxTagCloudLimit
	}
	return s.galleryRepo.TagCloud(s.getCtx(), dest, limit)
}

// GetPhotoDetail returns full detail of one photo (Feature 18)
func (s *GalleryService) GetPhotoDetail(id int32) (*dto.PhotoDetail, error) {
	if id <= 0 {
		return nil, errors.New("รหัสรูปภาพไม่ถูกต้อง")
	}
	detail, err := s.galleryRepo.GetByID(s.getCtx(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ไม่พบรูปภาพที่ระบุในระบบ")
		}
		return nil, err
	}
	return detail, nil
}

// UpdateDescription edits the description of a photo (blank allowed, max 2000 runes) (Feature 19)
func (s *GalleryService) UpdateDescription(id int32, description string) error {
	if id <= 0 {
		return errors.New("รหัสรูปภาพไม่ถูกต้อง")
	}
	description = strings.TrimSpace(description)
	if runeCount := utf8.RuneCountInString(description); runeCount > maxDescriptionRunes {
		return fmt.Errorf("คำอธิบายต้องไม่เกิน %d ตัวอักษร (ขณะนี้ %d ตัว)", maxDescriptionRunes, runeCount)
	}
	return s.galleryRepo.UpdateDescription(s.getCtx(), nil, id, description)
}

// AddTag attaches a user tag to a photo inside a single transaction:
// find or create the tag, skip duplicates (photo_tags has no PK),
// insert with source 'user', then return the photo's latest tags (Feature 19)
func (s *GalleryService) AddTag(photoID int32, name string) (*dto.PhotoDetail, error) {
	if photoID <= 0 {
		return nil, errors.New("รหัสรูปภาพไม่ถูกต้อง")
	}

	// Trim and collapse repeated whitespace inside the tag name
	cleanName := strings.Join(strings.Fields(name), " ")
	if cleanName == "" {
		return nil, errors.New("กรุณาระบุชื่อ Tag")
	}
	if runeCount := utf8.RuneCountInString(cleanName); runeCount > maxTagNameRunes {
		return nil, fmt.Errorf("ชื่อ Tag ต้องไม่เกิน %d ตัวอักษร (ขณะนี้ %d ตัว)", maxTagNameRunes, runeCount)
	}

	err := s.galleryRepo.WithTx(s.getCtx(), func(tx *query.Query) error {
		tag, findErr := s.galleryRepo.FindTagByName(s.getCtx(), tx, cleanName)
		if findErr != nil {
			if !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return findErr
			}
			created, createErr := s.galleryRepo.CreateTag(s.getCtx(), tx, cleanName)
			if createErr != nil {
				return createErr
			}
			tag = created
		}
		if tag == nil || tag.ID <= 0 {
			return errors.New("ไม่สามารถเตรียม Tag สำหรับรูปภาพนี้ได้")
		}

		exists, existErr := s.galleryRepo.PhotoTagExists(s.getCtx(), tx, photoID, tag.ID)
		if existErr != nil {
			return existErr
		}
		if exists {
			// Photo already carries this tag: do not insert a duplicate row
			return nil
		}
		return s.galleryRepo.CreatePhotoTag(s.getCtx(), tx, photoID, tag.ID, "user")
	})
	if err != nil {
		return nil, err
	}

	return s.galleryRepo.GetByID(s.getCtx(), photoID)
}

// RemoveTag removes a tag from a photo (Feature 19)
func (s *GalleryService) RemoveTag(photoID int32, tagID int32) error {
	if photoID <= 0 {
		return errors.New("รหัสรูปภาพไม่ถูกต้อง")
	}
	if tagID <= 0 {
		return errors.New("รหัส Tag ไม่ถูกต้อง")
	}
	return s.galleryRepo.DeletePhotoTag(s.getCtx(), nil, photoID, tagID)
}

// SuggestTags returns tag name suggestions for the typed prefix
func (s *GalleryService) SuggestTags(prefix string, limit int) ([]*modelGen.Tag, error) {
	prefix = strings.TrimSpace(prefix)
	if limit <= 0 {
		limit = defaultSuggestLimit
	}
	if limit > maxSuggestLimit {
		limit = maxSuggestLimit
	}
	return s.galleryRepo.SuggestTags(s.getCtx(), prefix, limit)
}
