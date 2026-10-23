package repository

import (
	"context"
	"strings"

	"PhotoVault/model/model"
	"PhotoVault/model/query"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagRepository interface {
	FindOrCreateTag(ctx context.Context, name string) (*model.Tag, error)
	FindOrCreateTags(ctx context.Context, names []string) ([]*model.Tag, error)
	AddPhotoTags(ctx context.Context, photoID int32, tags []*model.Tag, source string) error
	AddPhotoTagsInBatch(ctx context.Context, photoTagNames map[int32][]string, source string) error
	GetTagsByPhotoID(ctx context.Context, photoID int32) ([]*model.Tag, error)
	GetAllTags(ctx context.Context) ([]*model.Tag, error)
	DeletePhotoTags(ctx context.Context, photoID int32) error
}

type tagRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{
		db: db,
		q:  query.Use(db),
	}
}

// FindOrCreateTag looks up a tag by name (case-insensitive trimmed), creating it if it doesn't exist
func (r *tagRepository) FindOrCreateTag(ctx context.Context, name string) (*model.Tag, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, nil
	}

	t := r.q.Tag
	existing, err := t.WithContext(ctx).Where(t.Name.Eq(cleanName)).First()
	if err == nil && existing != nil {
		return existing, nil
	}

	// Create new tag
	newTag := &model.Tag{
		Name: cleanName,
	}
	err = t.WithContext(ctx).Create(newTag)
	if err != nil {
		// In case of race condition where another goroutine created it
		existing, retryErr := t.WithContext(ctx).Where(t.Name.Eq(cleanName)).First()
		if retryErr == nil && existing != nil {
			return existing, nil
		}
		return nil, err
	}
	return newTag, nil
}

// FindOrCreateTags looks up or creates multiple tags in batch
func (r *tagRepository) FindOrCreateTags(ctx context.Context, names []string) ([]*model.Tag, error) {
	var result []*model.Tag
	seen := make(map[string]bool)

	for _, name := range names {
		clean := strings.TrimSpace(name)
		if clean == "" || seen[strings.ToLower(clean)] {
			continue
		}
		seen[strings.ToLower(clean)] = true

		tag, err := r.FindOrCreateTag(ctx, clean)
		if err != nil {
			return nil, err
		}
		if tag != nil {
			result = append(result, tag)
		}
	}
	return result, nil
}

// AddPhotoTags links a list of tags to a photo_id
func (r *tagRepository) AddPhotoTags(ctx context.Context, photoID int32, tags []*model.Tag, source string) error {
	if photoID <= 0 || len(tags) == 0 {
		return nil
	}
	if source == "" {
		source = "ai"
	}

	pt := r.q.PhotoTag
	for _, tag := range tags {
		if tag == nil || tag.ID <= 0 {
			continue
		}
		// Check if relation already exists
		count, err := pt.WithContext(ctx).
			Where(pt.PhotoID.Eq(photoID)).
			Where(pt.TagID.Eq(tag.ID)).
			Count()
		if err == nil && count > 0 {
			continue
		}

		link := &model.PhotoTag{
			PhotoID: photoID,
			TagID:   tag.ID,
			Source:  source,
		}
		_ = pt.WithContext(ctx).Create(link)
	}
	return nil
}

// AddPhotoTagsInBatch creates/looks up all tags and links them to multiple photos inside a single transaction
func (r *tagRepository) AddPhotoTagsInBatch(ctx context.Context, photoTagNames map[int32][]string, source string) error {
	if len(photoTagNames) == 0 {
		return nil
	}
	if source == "" {
		source = "ai"
	}

	uniqueNamesMap := make(map[string]string)
	for _, names := range photoTagNames {
		for _, n := range names {
			clean := strings.TrimSpace(n)
			if clean != "" {
				uniqueNamesMap[strings.ToLower(clean)] = clean
			}
		}
	}
	if len(uniqueNamesMap) == 0 {
		return nil
	}

	var uniqueNames []string
	for _, clean := range uniqueNamesMap {
		uniqueNames = append(uniqueNames, clean)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Query all existing tags that match
		var existingTags []*model.Tag
		if err := tx.Where("name IN ?", uniqueNames).Find(&existingTags).Error; err != nil {
			return err
		}

		tagIDByLower := make(map[string]int32, len(uniqueNames))
		for _, t := range existingTags {
			lower := strings.ToLower(strings.TrimSpace(t.Name))
			if _, exists := tagIDByLower[lower]; !exists {
				tagIDByLower[lower] = t.ID
			}
		}

		// 2. Only insert genuinely new tags that are not in DB
		var toInsert []*model.Tag
		for _, clean := range uniqueNames {
			lower := strings.ToLower(clean)
			if tagIDByLower[lower] <= 0 {
				toInsert = append(toInsert, &model.Tag{Name: clean})
			}
		}

		if len(toInsert) > 0 {
			if err := tx.Create(&toInsert).Error; err != nil {
				// Retry lookup if another thread inserted
				var refound []*model.Tag
				_ = tx.Where("name IN ?", uniqueNames).Find(&refound)
				for _, t := range refound {
					lower := strings.ToLower(strings.TrimSpace(t.Name))
					if tagIDByLower[lower] <= 0 {
						tagIDByLower[lower] = t.ID
					}
				}
			} else {
				for _, t := range toInsert {
					lower := strings.ToLower(strings.TrimSpace(t.Name))
					tagIDByLower[lower] = t.ID
				}
			}
		}

		// 3. Build PhotoTag links
		var links []*model.PhotoTag
		seenLink := make(map[[2]int32]bool)
		for photoID, names := range photoTagNames {
			if photoID <= 0 {
				continue
			}
			for _, n := range names {
				lower := strings.ToLower(strings.TrimSpace(n))
				tagID := tagIDByLower[lower]
				if tagID <= 0 {
					continue
				}
				key := [2]int32{photoID, tagID}
				if seenLink[key] {
					continue
				}
				seenLink[key] = true
				links = append(links, &model.PhotoTag{
					PhotoID: photoID,
					TagID:   tagID,
					Source:  source,
				})
			}
		}

		if len(links) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(links, 500).Error
	})
}

// GetTagsByPhotoID fetches all tags associated with a specific photo
func (r *tagRepository) GetTagsByPhotoID(ctx context.Context, photoID int32) ([]*model.Tag, error) {
	pt := r.q.PhotoTag
	photoTags, err := pt.WithContext(ctx).
		Where(pt.PhotoID.Eq(photoID)).
		Preload(pt.Tag).
		Find()
	if err != nil {
		return nil, err
	}

	tags := make([]*model.Tag, 0, len(photoTags))
	for _, item := range photoTags {
		if item.Tag.ID > 0 {
			t := item.Tag
			tags = append(tags, &t)
		}
	}
	return tags, nil
}

// GetAllTags returns all tags available in the database
func (r *tagRepository) GetAllTags(ctx context.Context) ([]*model.Tag, error) {
	t := r.q.Tag
	return t.WithContext(ctx).Order(t.Name.Asc()).Find()
}

// DeletePhotoTags removes all tag relations for a given photo
func (r *tagRepository) DeletePhotoTags(ctx context.Context, photoID int32) error {
	pt := r.q.PhotoTag
	_, err := pt.WithContext(ctx).Where(pt.PhotoID.Eq(photoID)).Delete()
	return err
}
