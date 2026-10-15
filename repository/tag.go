package repository

import (
	"context"
	"strings"

	"PhotoVault/model/model"
	"PhotoVault/model/query"

	"gorm.io/gorm"
)

type TagRepository interface {
	FindOrCreateTag(ctx context.Context, name string) (*model.Tag, error)
	FindOrCreateTags(ctx context.Context, names []string) ([]*model.Tag, error)
	AddPhotoTags(ctx context.Context, photoID int32, tags []*model.Tag, source string) error
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
