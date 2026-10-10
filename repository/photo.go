package repository

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"PhotoVault/model/model"
	"PhotoVault/model/query"

	"gorm.io/gorm"
)

type PhotoRepository interface {
	GetActivePhotosByDest(ctx context.Context, dest string) ([]*model.Photo, error)
	GetAllPhotosByDest(ctx context.Context, dest string) ([]*model.Photo, error)
	AddPhoto(ctx context.Context, photo *model.Photo) error
	AddPhotosInBatch(ctx context.Context, photos []*model.Photo) error
	IsPhotoRecorded(ctx context.Context, dest string, filename string) (bool, error)
	IsSha256Recorded(ctx context.Context, dest string, sha256 string) (bool, error)
	MarkPhotoMissing(ctx context.Context, dest string, filename string) error
	MarkPhotoActive(ctx context.Context, dest string, filename string) error
	MarkPhotoDeleted(ctx context.Context, dest string, filename string) (int64, error)
	MarkAllPhotosDeleted(ctx context.Context, dest string) (int64, error)
	UpdatePhotoDescription(ctx context.Context, id int32, description string) error
	UpdatePhotoDescriptionByFile(ctx context.Context, dest string, filename string, description string) error
	CleanPhotos(ctx context.Context) error
}

type photoRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewPhotoRepository(db *gorm.DB) PhotoRepository {
	return &photoRepository{
		db: db,
		q:  query.Use(db),
	}
}

// destConditions generates equivalent path variations (clean, abs, rel, slash, backslash)
// to ensure rock-solid cross-platform queries on both Windows and macOS
func destConditions(dest string) []string {
	condMap := make(map[string]bool)
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return nil
	}
	condMap[dest] = true
	condMap[filepath.Clean(dest)] = true
	condMap[filepath.ToSlash(dest)] = true
	condMap[filepath.FromSlash(dest)] = true

	abs, err := filepath.Abs(dest)
	if err == nil {
		condMap[abs] = true
		condMap[filepath.Clean(abs)] = true
		condMap[filepath.ToSlash(abs)] = true
		condMap[filepath.FromSlash(abs)] = true
	}

	cwd, err := os.Getwd()
	if err == nil {
		rel, err := filepath.Rel(cwd, dest)
		if err == nil {
			condMap[rel] = true
			condMap[filepath.Clean(rel)] = true
			condMap[filepath.ToSlash(rel)] = true
			condMap[filepath.FromSlash(rel)] = true
		}
	}

	var conds []string
	for k := range condMap {
		if k != "" {
			conds = append(conds, k)
		}
	}
	return conds
}

// GetActivePhotosByDest returns all active photos for the destination folder with preloaded tags
func (r *photoRepository) GetActivePhotosByDest(ctx context.Context, dest string) ([]*model.Photo, error) {
	conds := destConditions(dest)
	p := r.q.Photo
	return p.WithContext(ctx).
		Preload(p.PhotoTags.Tag).
		Where(p.DestinationPath.In(conds...)).
		Where(p.Status.Eq("active")).
		Order(p.ID.Asc()).
		Find()
}

// GetAllPhotosByDest returns all photos (active, missing, deleted) for destination history with preloaded tags
func (r *photoRepository) GetAllPhotosByDest(ctx context.Context, dest string) ([]*model.Photo, error) {
	conds := destConditions(dest)
	p := r.q.Photo
	return p.WithContext(ctx).
		Preload(p.PhotoTags.Tag).
		Where(p.DestinationPath.In(conds...)).
		Order(p.ID.Asc()).
		Find()
}

// AddPhoto inserts a single photo record
func (r *photoRepository) AddPhoto(ctx context.Context, photo *model.Photo) error {
	p := r.q.Photo
	return p.WithContext(ctx).Create(photo)
}

// AddPhotosInBatch inserts photo records in batches of 1000 for maximum performance
func (r *photoRepository) AddPhotosInBatch(ctx context.Context, photos []*model.Photo) error {
	if len(photos) == 0 {
		return nil
	}
	p := r.q.Photo
	return p.WithContext(ctx).CreateInBatches(photos, 1000)
}

// IsPhotoRecorded checks whether a file with same filename exists with active or missing status
func (r *photoRepository) IsPhotoRecorded(ctx context.Context, dest string, filename string) (bool, error) {
	conds := destConditions(dest)
	p := r.q.Photo
	count, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.FileName.Eq(filename)).
		Where(p.Status.Neq("deleted")).
		Count()
	return count > 0, err
}

// IsSha256Recorded checks whether a photo with identical content hash exists in destination
func (r *photoRepository) IsSha256Recorded(ctx context.Context, dest string, sha256 string) (bool, error) {
	if sha256 == "" {
		return false, nil
	}
	conds := destConditions(dest)
	p := r.q.Photo
	count, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.Sha256.Eq(sha256)).
		Where(p.Status.Neq("deleted")).
		Count()
	return count > 0, err
}

// MarkPhotoMissing updates status to 'missing' when a DB record is missing from disk
func (r *photoRepository) MarkPhotoMissing(ctx context.Context, dest string, filename string) error {
	conds := destConditions(dest)
	p := r.q.Photo
	_, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.FileName.Eq(filename)).
		Update(p.Status, "missing")
	return err
}

// MarkPhotoActive updates status to 'active'
func (r *photoRepository) MarkPhotoActive(ctx context.Context, dest string, filename string) error {
	conds := destConditions(dest)
	p := r.q.Photo
	_, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.FileName.Eq(filename)).
		Update(p.Status, "active")
	return err
}

// MarkPhotoDeleted soft-deletes a photo record (status = 'deleted') to preserve history
func (r *photoRepository) MarkPhotoDeleted(ctx context.Context, dest string, filename string) (int64, error) {
	conds := destConditions(dest)
	p := r.q.Photo
	info, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.FileName.Eq(filename)).
		Update(p.Status, "deleted")
	return info.RowsAffected, err
}

// MarkAllPhotosDeleted soft-deletes all photos in the destination folder
func (r *photoRepository) MarkAllPhotosDeleted(ctx context.Context, dest string) (int64, error) {
	conds := destConditions(dest)
	p := r.q.Photo
	info, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.Status.Neq("deleted")).
		Update(p.Status, "deleted")
	return info.RowsAffected, err
}

// UpdatePhotoDescription updates the description text of a specific photo by ID
func (r *photoRepository) UpdatePhotoDescription(ctx context.Context, id int32, description string) error {
	p := r.q.Photo
	_, err := p.WithContext(ctx).Where(p.ID.Eq(id)).Update(p.Description, description)
	return err
}

// UpdatePhotoDescriptionByFile updates description by filename+dest (safe when ID not yet known)
func (r *photoRepository) UpdatePhotoDescriptionByFile(ctx context.Context, dest string, filename string, description string) error {
	conds := destConditions(dest)
	p := r.q.Photo
	_, err := p.WithContext(ctx).
		Where(p.DestinationPath.In(conds...)).
		Where(p.FileName.Eq(filename)).
		Update(p.Description, description)
	return err
}

// CleanPhotos truncates/cleans the photos table
func (r *photoRepository) CleanPhotos(ctx context.Context) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Photo{}).Error
}

