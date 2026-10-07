package main

import (
	"context"
	"fmt"

	modelGen "PhotoVault/model/model"
	"PhotoVault/service"
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
