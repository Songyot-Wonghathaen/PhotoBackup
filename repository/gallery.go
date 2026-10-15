package repository

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"PhotoVault/controller/dto"
	"PhotoVault/model/model"
	"PhotoVault/model/query"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// GalleryRepository handles gallery search, tag cloud and photo detail data access (Features 15-19)
type GalleryRepository interface {
	Search(ctx context.Context, dest string, params dto.SearchParams) (*dto.SearchResult, error)
	TagCloud(ctx context.Context, dest string, limit int) ([]dto.TagItem, error)
	GetByID(ctx context.Context, id int32) (*dto.PhotoDetail, error)
	FindActiveByID(ctx context.Context, id int32) (*model.Photo, error)
	WithTx(ctx context.Context, fn func(tx *query.Query) error) error
	UpdateDescription(ctx context.Context, tx *query.Query, id int32, description string) error
	FindTagByName(ctx context.Context, tx *query.Query, name string) (*model.Tag, error)
	CreateTag(ctx context.Context, tx *query.Query, name string) (*model.Tag, error)
	PhotoTagExists(ctx context.Context, tx *query.Query, photoID int32, tagID int32) (bool, error)
	CreatePhotoTag(ctx context.Context, tx *query.Query, photoID int32, tagID int32, source string) error
	DeletePhotoTag(ctx context.Context, tx *query.Query, photoID int32, tagID int32) error
	SuggestTags(ctx context.Context, prefix string, limit int) ([]*model.Tag, error)
}

type galleryRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewGalleryRepository(db *gorm.DB) GalleryRepository {
	return &galleryRepository{
		db: db,
		q:  query.Use(db),
	}
}

// queryOr returns the transactional query when running inside a transaction,
// otherwise the repository default query
func (r *galleryRepository) queryOr(tx *query.Query) *query.Query {
	if tx != nil {
		return tx
	}
	return r.q
}

// escapeLikePattern escapes LIKE wildcards (backslash, percent, underscore)
// so user input is matched literally when combined with ESCAPE '\'
func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)
	return replacer.Replace(value)
}

// likeEscaped builds a parameter-bound "column LIKE ? ESCAPE '\'" condition.
// The column SQL is a static identifier; the pattern travels as a bound parameter.
func likeEscaped(columnSQL string, pattern string) field.Expr {
	return field.NewUnsafeFieldRaw(columnSQL+" LIKE ? ESCAPE '\\'", pattern)
}

// dedupeIDs removes non-positive and duplicated tag IDs
func dedupeIDs(ids []int32) []int32 {
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

// thumbnailURL builds the ID-based thumbnail endpoint URL. Using the photo ID
// lets the Go asset handler resolve and validate the file in both Wails dev and
// production modes; the legacy /api/thumbnail endpoint is handled by Nuxt in
// dev mode and returns HTML instead of image bytes.
func thumbnailURL(photoID int32) string {
	return fmt.Sprintf("/thumb/%d", photoID)
}

func photoTagNames(photo *model.Photo) []string {
	names := make([]string, 0, len(photo.PhotoTags))
	seen := make(map[string]bool, len(photo.PhotoTags))
	for _, link := range photo.PhotoTags {
		name := strings.TrimSpace(link.Tag.Name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

func toPhotoListItem(photo *model.Photo) dto.PhotoListItem {
	fullPath := filepath.Join(photo.DestinationPath, photo.FileName)
	return dto.PhotoListItem{
		ID:          photo.ID,
		FileName:    photo.FileName,
		FullPath:    fullPath,
		ThumbURL:    thumbnailURL(photo.ID),
		Description: photo.Description,
		Tags:        photoTagNames(photo),
		BackedUpAt:  photo.BackedUpAt.Format("2006-01-02 15:04:05"),
	}
}

// Search finds active photos by description words (AND) and/or selected tags
// (must have every selected tag), ordered by backed_up_at DESC, id DESC with paging.
// Feature 15: description search | Feature 17: multi-tag filter
func (r *galleryRepository) Search(ctx context.Context, dest string, params dto.SearchParams) (*dto.SearchResult, error) {
	conds := destConditions(dest)

	// Normalise paging (service validates first; this keeps the repo safe standalone)
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 40
	}

	if len(conds) == 0 {
		return &dto.SearchResult{
			Items:    []dto.PhotoListItem{},
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	// Feature 15: whitespace-separated words are combined with AND
	terms := strings.Fields(params.Query)

	// Feature 17: "must have every selected tag" filtering
	tagIDs := dedupeIDs(params.TagIDs)

	p := r.q.Photo
	pt := r.q.PhotoTag

	// Subquery: photo IDs that carry ALL selected tags
	var tagSub query.IPhotoTagDo
	if len(tagIDs) > 0 {
		tagSub = pt.WithContext(ctx).
			Select(pt.PhotoID).
			Where(pt.TagID.In(tagIDs...)).
			Group(pt.PhotoID).
			Having(field.NewUnsafeFieldRaw("COUNT(DISTINCT tag_id) = ?", len(tagIDs)))
	}

	applyFilters := func(do query.IPhotoDo) query.IPhotoDo {
		do = do.Where(p.DestinationPath.In(conds...)).
			Where(p.Status.Eq("active"))
		for _, term := range terms {
			do = do.Where(likeEscaped(model.TableNamePhoto+".description", "%"+escapeLikePattern(term)+"%"))
		}
		if tagSub != nil {
			do = do.Where(gen.Columns{p.ID}.In(tagSub))
		}
		return do
	}

	total, err := applyFilters(p.WithContext(ctx)).Count()
	if err != nil {
		return nil, err
	}

	photos, err := applyFilters(p.WithContext(ctx)).
		Preload(p.PhotoTags.Tag).
		Order(p.BackedUpAt.Desc(), p.ID.Desc()).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find()
	if err != nil {
		return nil, err
	}

	items := make([]dto.PhotoListItem, 0, len(photos))
	for _, photo := range photos {
		items = append(items, toPhotoListItem(photo))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &dto.SearchResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// TagCloud returns tags used by active photos in the destination,
// ordered by photo count DESC (Feature 16)
func (r *galleryRepository) TagCloud(ctx context.Context, dest string, limit int) ([]dto.TagItem, error) {
	conds := destConditions(dest)
	if limit <= 0 {
		limit = 50
	}
	if len(conds) == 0 {
		return []dto.TagItem{}, nil
	}

	p := r.q.Photo
	pt := r.q.PhotoTag
	t := r.q.Tag

	// Subquery: active photo IDs in this destination
	photoIDs := p.WithContext(ctx).
		Select(p.ID).
		Where(p.DestinationPath.In(conds...)).
		Where(p.Status.Eq("active"))

	rows, err := t.WithContext(ctx).
		Select(t.ID, t.Name, pt.PhotoID.Count().As("photo_count")).
		LeftJoin(pt, pt.TagID.EqCol(t.ID)).
		Where(gen.Columns{pt.PhotoID}.In(photoIDs)).
		Group(t.ID, t.Name).
		Order(pt.PhotoID.Count().Desc(), t.Name.Asc()).
		Limit(limit).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.TagItem, 0)
	for rows.Next() {
		var id int32
		var name string
		var count int64
		if err := rows.Scan(&id, &name, &count); err != nil {
			return nil, err
		}
		items = append(items, dto.TagItem{ID: id, Name: name, Count: int(count)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetByID returns full photo detail with preloaded tags (Feature 18)
func (r *galleryRepository) GetByID(ctx context.Context, id int32) (*dto.PhotoDetail, error) {
	p := r.q.Photo
	photo, err := p.WithContext(ctx).
		Preload(p.PhotoTags.Tag).
		Where(p.ID.Eq(id)).
		First()
	if err != nil {
		return nil, err
	}
	return r.toPhotoDetail(ctx, photo)
}

// FindActiveByID returns the active photo row by ID (used by the image server)
func (r *galleryRepository) FindActiveByID(ctx context.Context, id int32) (*model.Photo, error) {
	p := r.q.Photo
	return p.WithContext(ctx).
		Where(p.ID.Eq(id)).
		Where(p.Status.Eq("active")).
		First()
}

func (r *galleryRepository) toPhotoDetail(ctx context.Context, photo *model.Photo) (*dto.PhotoDetail, error) {
	fullPath := filepath.Join(photo.DestinationPath, photo.FileName)

	tagItems := make([]dto.TagItem, 0, len(photo.PhotoTags))
	tagIDs := make([]int32, 0, len(photo.PhotoTags))
	for _, link := range photo.PhotoTags {
		name := strings.TrimSpace(link.Tag.Name)
		if link.TagID <= 0 || name == "" {
			continue
		}
		tagItems = append(tagItems, dto.TagItem{ID: link.TagID, Name: name, Source: link.Source})
		tagIDs = append(tagIDs, link.TagID)
	}

	// Populate how many active photos share each tag
	if len(tagIDs) > 0 {
		counts, err := r.countActivePhotosPerTag(ctx, tagIDs)
		if err != nil {
			return nil, err
		}
		for i := range tagItems {
			tagItems[i].Count = counts[tagItems[i].ID]
		}
	}

	return &dto.PhotoDetail{
		ID:              photo.ID,
		FileName:        photo.FileName,
		DestinationPath: photo.DestinationPath,
		FullPath:        fullPath,
		ThumbURL:        thumbnailURL(photo.ID),
		Description:     photo.Description,
		Tags:            tagItems,
		SizeBytes:       photo.SizeBytes,
		Sha256:          photo.Sha256,
		Status:          photo.Status,
		BackedUpAt:      photo.BackedUpAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// countActivePhotosPerTag returns how many active photos carry each of the given tags
func (r *galleryRepository) countActivePhotosPerTag(ctx context.Context, tagIDs []int32) (map[int32]int, error) {
	p := r.q.Photo
	pt := r.q.PhotoTag

	// Subquery: all active photo IDs
	photoIDs := p.WithContext(ctx).
		Select(p.ID).
		Where(p.Status.Eq("active"))

	rows, err := pt.WithContext(ctx).
		Select(pt.TagID, pt.PhotoID.Count().As("photo_count")).
		Where(pt.TagID.In(tagIDs...)).
		Where(gen.Columns{pt.PhotoID}.In(photoIDs)).
		Group(pt.TagID).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int32]int, len(tagIDs))
	for rows.Next() {
		var id int32
		var count int64
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		counts[id] = int(count)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

// WithTx runs fn inside a single database transaction
func (r *galleryRepository) WithTx(ctx context.Context, fn func(tx *query.Query) error) error {
	return r.q.Transaction(fn)
}

// UpdateDescription updates the description text of a photo by ID (Feature 19)
func (r *galleryRepository) UpdateDescription(ctx context.Context, tx *query.Query, id int32, description string) error {
	q := r.queryOr(tx)
	p := q.Photo
	_, err := p.WithContext(ctx).
		Where(p.ID.Eq(id)).
		Update(p.Description, description)
	return err
}

// FindTagByName looks up a tag by exact name (lowest ID wins when duplicated)
func (r *galleryRepository) FindTagByName(ctx context.Context, tx *query.Query, name string) (*model.Tag, error) {
	q := r.queryOr(tx)
	t := q.Tag
	return t.WithContext(ctx).
		Where(t.Name.Eq(name)).
		Order(t.ID.Asc()).
		First()
}

// CreateTag inserts a new tag row
func (r *galleryRepository) CreateTag(ctx context.Context, tx *query.Query, name string) (*model.Tag, error) {
	q := r.queryOr(tx)
	t := q.Tag
	tag := &model.Tag{Name: name}
	if err := t.WithContext(ctx).Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// PhotoTagExists reports whether a photo-tag link already exists (photo_tags has no PK)
func (r *galleryRepository) PhotoTagExists(ctx context.Context, tx *query.Query, photoID int32, tagID int32) (bool, error) {
	q := r.queryOr(tx)
	pt := q.PhotoTag
	count, err := pt.WithContext(ctx).
		Where(pt.PhotoID.Eq(photoID)).
		Where(pt.TagID.Eq(tagID)).
		Count()
	return count > 0, err
}

// CreatePhotoTag links a tag to a photo with the given source ('ai' or 'user')
func (r *galleryRepository) CreatePhotoTag(ctx context.Context, tx *query.Query, photoID int32, tagID int32, source string) error {
	if photoID <= 0 || tagID <= 0 {
		return nil
	}
	if source == "" {
		source = "ai"
	}
	q := r.queryOr(tx)
	pt := q.PhotoTag
	return pt.WithContext(ctx).Create(&model.PhotoTag{
		PhotoID: photoID,
		TagID:   tagID,
		Source:  source,
	})
}

// DeletePhotoTag removes the link between a photo and a tag (Feature 19)
func (r *galleryRepository) DeletePhotoTag(ctx context.Context, tx *query.Query, photoID int32, tagID int32) error {
	q := r.queryOr(tx)
	pt := q.PhotoTag
	_, err := pt.WithContext(ctx).
		Where(pt.PhotoID.Eq(photoID)).
		Where(pt.TagID.Eq(tagID)).
		Delete()
	return err
}

// SuggestTags returns tag names starting with the given prefix (literal prefix match)
func (r *galleryRepository) SuggestTags(ctx context.Context, prefix string, limit int) ([]*model.Tag, error) {
	if limit <= 0 {
		limit = 10
	}
	t := r.q.Tag
	q := t.WithContext(ctx)
	if trimmed := strings.TrimSpace(prefix); trimmed != "" {
		q = q.Where(likeEscaped(model.TableNameTag+".name", escapeLikePattern(trimmed)+"%"))
	}
	return q.Order(t.Name.Asc()).Limit(limit).Find()
}
