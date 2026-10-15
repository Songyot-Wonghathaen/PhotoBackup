package controller

import (
	"context"

	"PhotoVault/controller/dto"
	modelGen "PhotoVault/model/model"
	"PhotoVault/service"
)

// GalleryController exposes gallery & search features (Features 15-19) to the Wails frontend
type GalleryController struct {
	galleryService *service.GalleryService
}

func NewGalleryController(galleryService *service.GalleryService) *GalleryController {
	return &GalleryController{
		galleryService: galleryService,
	}
}

// SetContext forwards the Wails runtime context to the gallery service
func (c *GalleryController) SetContext(ctx context.Context) {
	if c.galleryService != nil {
		c.galleryService.SetContext(ctx)
	}
}

// SearchPhotos searches photos by description words and/or selected tags (Features 15, 17)
func (c *GalleryController) SearchPhotos(dest string, params dto.SearchParams) (*dto.SearchResult, error) {
	return c.galleryService.SearchPhotos(dest, params)
}

// GetTagCloud returns tags ordered by photo count for the Tag Cloud (Feature 16)
func (c *GalleryController) GetTagCloud(dest string, limit int) ([]dto.TagItem, error) {
	return c.galleryService.GetTagCloud(dest, limit)
}

// GetPhotoDetail returns full detail of one photo (Feature 18)
func (c *GalleryController) GetPhotoDetail(id int32) (*dto.PhotoDetail, error) {
	return c.galleryService.GetPhotoDetail(id)
}

// UpdateDescription edits the description of a photo (Feature 19)
func (c *GalleryController) UpdateDescription(id int32, description string) error {
	return c.galleryService.UpdateDescription(id, description)
}

// AddTag adds a tag to a photo and returns the refreshed photo detail (Feature 19)
func (c *GalleryController) AddTag(photoID int32, name string) (*dto.PhotoDetail, error) {
	return c.galleryService.AddTag(photoID, name)
}

// RemoveTag removes a tag from a photo (Feature 19)
func (c *GalleryController) RemoveTag(photoID int32, tagID int32) error {
	return c.galleryService.RemoveTag(photoID, tagID)
}

// SuggestTags returns tag name suggestions for a typed prefix
func (c *GalleryController) SuggestTags(prefix string, limit int) ([]*modelGen.Tag, error) {
	return c.galleryService.SuggestTags(prefix, limit)
}
