package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"PhotoVault/controller/dto"
	modelGen "PhotoVault/model/model"
	"PhotoVault/repository"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
	".heic": true,
	".heif": true,
	".raw":  true,
	".cr2":  true,
	".nef":  true,
	".arw":  true,
	".dng":  true,
	".svg":  true,
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageExtensions[ext]
}

type FileItem = dto.FileItem
type MoveSummary = dto.MoveSummary
type IntegrityResult = dto.IntegrityResult

type BackupService struct {
	ctx       context.Context
	photoRepo repository.PhotoRepository
	tagRepo   repository.TagRepository
	aiVision  AiVisionService
	source    string
	dest      string
	mu        sync.RWMutex
}

func NewBackupService(photoRepo repository.PhotoRepository, optionalServices ...interface{}) *BackupService {
	s := &BackupService{
		ctx:       context.Background(),
		photoRepo: photoRepo,
	}
	for _, opt := range optionalServices {
		switch v := opt.(type) {
		case repository.TagRepository:
			s.tagRepo = v
		case AiVisionService:
			s.aiVision = v
		}
	}
	return s
}

func (s *BackupService) SetTagRepo(tagRepo repository.TagRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tagRepo = tagRepo
}

func (s *BackupService) SetAiVision(aiVision AiVisionService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aiVision = aiVision
}

// AnalyzePhoto allows analyzing an individual photo directly
func (s *BackupService) AnalyzePhoto(filePath string) (*VisionResult, error) {
	s.mu.RLock()
	ai := s.aiVision
	s.mu.RUnlock()

	if ai == nil {
		ai = NewAiVisionService()
	}
	return ai.AnalyzePhoto(s.getCtx(), filePath)
}

func (s *BackupService) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

func (s *BackupService) getCtx() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

func (s *BackupService) GetSource() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.source
}

func (s *BackupService) GetDest() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dest
}

func (s *BackupService) SetSource(source string) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return errors.New("source cannot be empty")
	}
	info, err := os.Stat(source)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("source path is not a valid directory: %s", source)
	}

	s.mu.Lock()
	s.source = source
	s.mu.Unlock()
	return nil
}

func (s *BackupService) SetDest(dest string) (*IntegrityResult, error) {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return nil, errors.New("destination cannot be empty")
	}
	_ = os.MkdirAll(dest, 0755)

	s.mu.Lock()
	s.dest = dest
	s.mu.Unlock()

	// Feature 5: Immediate integrity scan upon selecting destination
	integrity, err := s.CheckIntegrity(dest)
	if err != nil {
		return nil, err
	}
	return &integrity, nil
}

// SelectSourceDialog opens native OS directory chooser for Source
func (s *BackupService) SelectSourceDialog() (string, error) {
	ctx := s.getCtx()
	dir, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: "Select Source Folder (Flash Drive / Phone)",
	})
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", nil // User cancelled
	}
	err = s.SetSource(dir)
	if err != nil {
		return "", err
	}
	return dir, nil
}

// SelectDestDialog opens native OS directory chooser for Destination and runs immediate integrity scan
func (s *BackupService) SelectDestDialog() (string, *IntegrityResult, error) {
	ctx := s.getCtx()
	dir, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: "Select Destination Backup Folder",
	})
	if err != nil {
		return "", nil, err
	}
	if dir == "" {
		return "", nil, nil // User cancelled
	}
	integrity, err := s.SetDest(dir)
	if err != nil {
		return "", nil, err
	}
	return dir, integrity, nil
}

func listDirectoryFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func listDirectoryFileItems(dirPath string) ([]FileItem, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var items []FileItem
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			size := int64(0)
			mod := ""
			if err == nil {
				size = info.Size()
				mod = info.ModTime().Format("2006-01-02 15:04:05")
			}
			items = append(items, FileItem{
				Filename:  entry.Name(),
				SizeBytes: size,
				ModTime:   mod,
				FullPath:  filepath.Join(dirPath, entry.Name()),
				IsImage:   isImageFile(entry.Name()),
			})
		}
	}
	return items, nil
}

// Feature 4 & 11: ListSource with State Guard
func (s *BackupService) ListSource() ([]string, error) {
	src := s.GetSource()
	if src == "" {
		return nil, errors.New("source is not set! Please set source directory first")
	}
	files, err := listDirectoryFiles(src)
	if err != nil {
		return nil, fmt.Errorf("unable to read source directory: %w", err)
	}

	var images []string
	for _, f := range files {
		if isImageFile(f) {
			images = append(images, f)
		}
	}
	return images, nil
}

func (s *BackupService) ListSourceFiles() ([]FileItem, error) {
	src := s.GetSource()
	if src == "" {
		return nil, errors.New("source is not set! Please set source directory first")
	}
	items, err := listDirectoryFileItems(src)
	if err != nil {
		return nil, fmt.Errorf("unable to read source directory: %w", err)
	}
	return items, nil
}

// Feature 10 & 11: ListDest with State Guard
func (s *BackupService) ListDest(destPath string) ([]string, error) {
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	return listDirectoryFiles(destPath)
}

func (s *BackupService) ListDestFiles(destPath string) ([]FileItem, error) {
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	return listDirectoryFileItems(destPath)
}

// ListDB lists active records from database for destPath
func (s *BackupService) ListDB(destPath string) ([]*modelGen.Photo, error) {
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	return s.photoRepo.GetActivePhotosByDest(s.getCtx(), destPath)
}

// Feature 6, 7, 8, 9, 11: Move with Worker Pool, timer, cross-drive copy, duplicate check, batch insert
func (s *BackupService) Move(targets []string, isAll bool) (MoveSummary, error) {
	src := s.GetSource()
	dst := s.GetDest()

	// Feature 11: State Guard
	if src == "" {
		return MoveSummary{}, errors.New("source is not set! Please set source directory first")
	}
	if dst == "" {
		return MoveSummary{}, errors.New("destination is not set! Please set destination directory first")
	}

	_ = os.MkdirAll(dst, 0755)

	var filesToProcess []string
	if isAll {
		srcFiles, err := listDirectoryFiles(src)
		if err != nil {
			return MoveSummary{}, fmt.Errorf("unable to read source directory: %w", err)
		}
		for _, f := range srcFiles {
			if isImageFile(f) {
				filesToProcess = append(filesToProcess, f)
			}
		}
	} else {
		for _, t := range targets {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				filesToProcess = append(filesToProcess, trimmed)
			}
		}
	}

	total := len(filesToProcess)
	summary := MoveSummary{
		TotalFiles:  total,
		MovedList:   []string{},
		SkippedList: []string{},
		FailedList:  []string{},
	}

	if total == 0 {
		return summary, nil
	}

	// Feature 7: Duplicate prevention (checks disk and DB)
	existingDestFiles, _ := listDirectoryFiles(dst)
	diskMap := make(map[string]bool, len(existingDestFiles))
	for _, f := range existingDestFiles {
		diskMap[strings.ToLower(f)] = true
	}

	dbPhotos, _ := s.photoRepo.GetActivePhotosByDest(s.getCtx(), dst)
	dbMap := make(map[string]bool, len(dbPhotos))
	shaMap := make(map[string]bool, len(dbPhotos))
	for _, p := range dbPhotos {
		dbMap[strings.ToLower(p.FileName)] = true
		if p.Sha256 != "" {
			shaMap[p.Sha256] = true
		}
	}

	var filesToMove []string
	for _, fname := range filesToProcess {
		lower := strings.ToLower(fname)
		if diskMap[lower] || dbMap[lower] {
			summary.SkippedList = append(summary.SkippedList, fname)
			summary.SkippedFiles++
			continue
		}
		filesToMove = append(filesToMove, fname)
	}

	if len(filesToMove) == 0 {
		return summary, nil
	}

	totalToMove := len(filesToMove)
	numWorkers := 16
	if numWorkers > totalToMove {
		numWorkers = totalToMove
	}

	jobs := make(chan string, totalToMove)
	for _, fname := range filesToMove {
		jobs <- fname
	}
	close(jobs)

	type moveResult struct {
		filename  string
		sha256    string
		sizeBytes int64
		err       error
	}

	resultChan := make(chan moveResult, totalToMove)

	// Feature 8: Timer Start
	startTime := time.Now()

	// Feature 9: Goroutine Worker Pool
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 128*1024) // 128 KB high throughput buffer

			for fname := range jobs {
				srcPath := filepath.Join(src, fname)
				dstPath := filepath.Join(dst, fname)

				shaHash, size, err := streamCopyAndRemoveFile(srcPath, dstPath, buf)
				resultChan <- moveResult{
					filename:  fname,
					sha256:    shaHash,
					sizeBytes: size,
					err:       err,
				}
			}
		}()
	}

	wg.Wait()
	close(resultChan)

	// Feature 8: Timer End
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	summary.DurationMs = duration.Milliseconds()
	summary.DurationFormatted = fmt.Sprintf("%.2f ms", float64(duration.Microseconds())/1000.0)

	var movedPhotos []*modelGen.Photo
	now := time.Now()

	for res := range resultChan {
		if res.err != nil {
			summary.FailedFiles++
			summary.FailedList = append(summary.FailedList, fmt.Sprintf("%s (%v)", res.filename, res.err))
		} else {
			// Check if duplicate SHA256 was found in DB
			if res.sha256 != "" && shaMap[res.sha256] {
				// Duplicate content already exists in DB
				summary.SkippedFiles++
				summary.SkippedList = append(summary.SkippedList, fmt.Sprintf("%s (duplicate content sha256)", res.filename))
				continue
			}

			summary.MovedFiles++
			summary.MovedList = append(summary.MovedList, res.filename)

			movedPhotos = append(movedPhotos, &modelGen.Photo{
				DestinationPath: dst,
				FileName:        res.filename,
				RelPath:         res.filename,
				SizeBytes:       int32(res.sizeBytes),
				Sha256:          res.sha256,
				BackedUpAt:      now,
				Status:          "active",
				Description:     "",
			})
		}
	}

	// Feature 9: Batch Insert into GORM SQLite (photos table)
	if len(movedPhotos) > 0 {
		_ = s.photoRepo.AddPhotosInBatch(s.getCtx(), movedPhotos)

		// Part C (Features 12, 13, 14): AI Vision Analysis on Backup
		s.mu.RLock()
		ai := s.aiVision
		tagRepo := s.tagRepo
		s.mu.RUnlock()

	if ai != nil && ai.IsAvailable() {
			for idx, p := range movedPhotos {
				if !isImageFile(p.FileName) {
					continue
				}
				fullDst := filepath.Join(dst, p.FileName)
				visionRes, err := ai.AnalyzePhoto(s.getCtx(), fullDst)
				if err != nil {
					// Feature 14: Handle failure gracefully, leave description blank and continue
					continue
				}

				// Feature 13: Save Description and Tags to DB
				if visionRes != nil {
					if visionRes.Description != "" {
						// Use filename+dest query (p.ID is 0 after CreateInBatches)
						dbErr := s.photoRepo.UpdatePhotoDescriptionByFile(s.getCtx(), dst, p.FileName, visionRes.Description)
						if dbErr != nil {
							log.Printf("[AI-Vision] ⚠️ บันทึก description ล้มเหลวสำหรับ %s: %v", p.FileName, dbErr)
						} else {
							log.Printf("[AI-Vision] 💾 บันทึก description ลง DB สำเร็จ: %s", p.FileName)
						}
					}

					if tagRepo != nil && len(visionRes.Tags) > 0 {
						// Fetch the real DB record to get actual ID for photo_tags
						dbPhotos, qErr := s.photoRepo.GetActivePhotosByDest(s.getCtx(), dst)
						var realID int32
						if qErr == nil {
							for _, dbP := range dbPhotos {
								if strings.EqualFold(dbP.FileName, p.FileName) {
									realID = dbP.ID
									break
								}
							}
						}
						if realID > 0 {
							createdTags, tagErr := tagRepo.FindOrCreateTags(s.getCtx(), visionRes.Tags)
							if tagErr == nil && len(createdTags) > 0 {
								_ = tagRepo.AddPhotoTags(s.getCtx(), realID, createdTags, "ai")
								log.Printf("[AI-Vision] 🏷️ บันทึก %d tags ลง DB สำเร็จ: %s", len(createdTags), p.FileName)
							}
						} else {
							log.Printf("[AI-Vision] ⚠️ ไม่พบ ID ใน DB สำหรับ %s — ข้าม tag insert", p.FileName)
						}
					}
				}

	// Emit progress event to Frontend
				if s.ctx != nil {
					runtime.EventsEmit(s.ctx, "ai_analysis_progress", map[string]interface{}{
						"current":  idx + 1,
						"total":    len(movedPhotos),
						"filename": p.FileName,
					})
				}

				// Polite pause to stay smoothly within the free tier rate limit
				// 2 seconds = ~30 requests/minute, safely under 15 req/min free tier limit
				if idx < len(movedPhotos)-1 {
					time.Sleep(2 * time.Second)
				}
			}
		}
	}

	return summary, nil
}

// streamCopyAndRemoveFile implements Feature 6:
// Tries os.Rename first. If across drives, fallback to stream copy buffer and delete source.
func streamCopyAndRemoveFile(srcPath, dstPath string, buffer []byte) (string, int64, error) {
	fi, err := os.Stat(srcPath)
	if err != nil {
		return "", 0, err
	}
	size := fi.Size()

	// Attempt fast Rename first
	renameErr := os.Rename(srcPath, dstPath)
	if renameErr == nil {
		hash := fastFileHash(dstPath)
		return hash, size, nil
	}

	// Fallback to cross-drive stream copy buffer
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", 0, err
	}

	hasher := sha256.New()
	writer := io.MultiWriter(dstFile, hasher)

	_, copyErr := io.CopyBuffer(writer, srcFile, buffer)
	dstFile.Close()

	if copyErr != nil {
		_ = os.Remove(dstPath)
		return "", 0, copyErr
	}

	srcFile.Close()
	_ = os.Remove(srcPath)

	return hex.EncodeToString(hasher.Sum(nil)), size, nil
}

func fastFileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	hasher := sha256.New()
	buf := make([]byte, 64*1024)
	_, _ = io.CopyBuffer(hasher, f, buf)
	return hex.EncodeToString(hasher.Sum(nil))
}

// CheckIntegrity implements Feature 5 & 21:
// Compares database records with physical files on disk, updates 'missing' status in DB if not found
func (s *BackupService) CheckIntegrity(destPath string) (IntegrityResult, error) {
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return IntegrityResult{}, errors.New("destination is not set! Please set destination directory first")
	}

	dbPhotos, err := s.photoRepo.GetActivePhotosByDest(s.getCtx(), destPath)
	if err != nil {
		return IntegrityResult{}, err
	}

	diskFiles, err := listDirectoryFiles(destPath)
	if err != nil && !os.IsNotExist(err) {
		return IntegrityResult{}, err
	}

	diskMap := make(map[string]bool, len(diskFiles))
	for _, f := range diskFiles {
		diskMap[strings.ToLower(f)] = true
	}

	var missingInDisk []string
	matched := 0

	for _, p := range dbPhotos {
		if diskMap[strings.ToLower(p.FileName)] {
			matched++
		} else {
			missingInDisk = append(missingInDisk, p.FileName)
			// Update status in DB to 'missing'
			_ = s.photoRepo.MarkPhotoMissing(s.getCtx(), destPath, p.FileName)
		}
	}

	result := IntegrityResult{
		TotalDB:       len(dbPhotos),
		TotalDisk:     len(diskFiles),
		MissingInDisk: missingInDisk,
		MatchedCount:  matched,
		IsExactMatch:  len(missingInDisk) == 0,
	}

	if len(missingInDisk) > 0 {
		result.AlertMessage = fmt.Sprintf("Missing %d file(s) detected: recorded in DB but missing from disk!", len(missingInDisk))
	} else {
		result.AlertMessage = "Integrity scan passed: All DB records match disk files."
	}

	return result, nil
}

// Delete removes files from destination disk and sets status = 'deleted' in DB
func (s *BackupService) Delete(targets []string, isAll bool) (int, error) {
	dst := s.GetDest()
	if dst == "" {
		return 0, errors.New("destination is not set! Please set destination directory first")
	}

	deletedCount := 0

	if isAll {
		diskFiles, _ := listDirectoryFiles(dst)
		for _, f := range diskFiles {
			p := filepath.Join(dst, f)
			if err := os.Remove(p); err == nil {
				deletedCount++
			}
		}
		// Soft-delete in DB (update status = 'deleted' to preserve history)
		_, _ = s.photoRepo.MarkAllPhotosDeleted(s.getCtx(), dst)
		return deletedCount, nil
	}

	for _, fname := range targets {
		fname = strings.TrimSpace(fname)
		if fname == "" {
			continue
		}
		p := filepath.Join(dst, fname)
		if err := os.Remove(p); err == nil {
			deletedCount++
		}
		_, _ = s.photoRepo.MarkPhotoDeleted(s.getCtx(), dst, fname)
	}

	return deletedCount, nil
}

func (s *BackupService) CleanAll() error {
	s.mu.Lock()
	s.source = ""
	s.dest = ""
	s.mu.Unlock()

	return s.photoRepo.CleanPhotos(s.getCtx())
}
