package main

import (
	"PhotoVault/controller/dto"
	modelGen "PhotoVault/model/model"
	"PhotoVault/service"
	"PhotoVault/utils"
	"context"
	"encoding/base64"
	"fmt"
	"sync"
)

// App struct
type App struct {
	ctx           context.Context
	backupService *service.BackupService
}

// NewApp creates a new App application struct
func NewApp(backupService *service.BackupService) *App {
	return &App{
		backupService: backupService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.backupService != nil {
		a.backupService.SetContext(ctx)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// --- Feature B Methods Exposing to Frontend ---

// SelectSourceFolder opens native directory dialog for source (Flash Drive / Phone)
func (a *App) SelectSourceFolder() (string, error) {
	return a.backupService.SelectSourceDialog()
}

// SetSourcePath sets source folder manually
func (a *App) SetSourcePath(path string) error {
	return a.backupService.SetSource(path)
}

// GetSourcePath returns currently selected source folder
func (a *App) GetSourcePath() string {
	return a.backupService.GetSource()
}

// ListSourcePhotos returns list of image files in source (Feature 4 & 11)
func (a *App) ListSourcePhotos() ([]string, error) {
	return a.backupService.ListSource()
}

// ListSourceFileItems returns detailed list of items in source (Feature 4 & 11)
func (a *App) ListSourceFileItems() ([]service.FileItem, error) {
	return a.backupService.ListSourceFiles()
}

// SelectDestFolder opens native directory dialog for destination and scans missing files (Feature 5)
func (a *App) SelectDestFolder() (*service.IntegrityResult, error) {
	_, result, err := a.backupService.SelectDestDialog()
	return result, err
}

// SetDestPath sets destination folder manually and scans missing files (Feature 5)
func (a *App) SetDestPath(path string) (*service.IntegrityResult, error) {
	return a.backupService.SetDest(path)
}

// GetDestPath returns currently selected destination folder
func (a *App) GetDestPath() string {
	return a.backupService.GetDest()
}

// CheckIntegrity compares DB records with disk files, detects missing files (Feature 5 & 21)
func (a *App) CheckIntegrity(destPath string) (service.IntegrityResult, error) {
	return a.backupService.CheckIntegrity(destPath)
}

// MovePhotos moves selected files or all files to destination (Feature 6, 7, 8, 9, 11)
func (a *App) MovePhotos(targets []string, isAll bool) (service.MoveSummary, error) {
	return a.backupService.Move(targets, isAll)
}

// ListDestPhotos lists actual files on disk in destination (Feature 10 & 11)
func (a *App) ListDestPhotos(destPath string) ([]string, error) {
	return a.backupService.ListDest(destPath)
}

// ListDestFileItems lists detailed file items on disk in destination (Feature 10 & 11)
func (a *App) ListDestFileItems(destPath string) ([]service.FileItem, error) {
	return a.backupService.ListDestFiles(destPath)
}

// ListDBPhotos lists photos recorded in database for destination
func (a *App) ListDBPhotos(destPath string) ([]*modelGen.Photo, error) {
	return a.backupService.ListDB(destPath)
}

// DeletePhotos deletes files from destination and marks status as 'deleted' in DB (Feature 23 & 25)
func (a *App) DeletePhotos(targets []string, isAll bool) (int, error) {
	return a.backupService.Delete(targets, isAll)
}

// GetPhotoThumbnail returns base64 data URL for a single photo
func (a *App) GetPhotoThumbnail(fullPath string) (string, error) {
	bytes, err := utils.GetThumbnailBytes(fullPath, 280)
	if err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(bytes), nil
}

// GetPhotoThumbnails returns base64 data URLs for a batch of photo paths concurrently
func (a *App) GetPhotoThumbnails(paths []string) map[string]string {
	result := make(map[string]string)
	if len(paths) == 0 {
		return result
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			bytes, err := utils.GetThumbnailBytes(path, 280)
			if err == nil && len(bytes) > 0 {
				b64 := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(bytes)
				mu.Lock()
				result[path] = b64
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()
	return result
}

// AnalyzePhoto triggers AI vision analysis directly for any photo file
func (a *App) AnalyzePhoto(fullPath string) (*dto.VisionResult, error) {
	return a.backupService.AnalyzePhoto(fullPath)
}
