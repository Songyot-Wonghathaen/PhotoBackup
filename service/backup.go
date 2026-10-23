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
	"sync/atomic"
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
	ctx            context.Context
	photoRepo      repository.PhotoRepository
	tagRepo        repository.TagRepository
	aiVision       AiVisionService
	source         string
	dest           string
	mu             sync.RWMutex
	analyzeCancel  context.CancelFunc
	transferMu     sync.Mutex
	aiAnalyzeMu    sync.Mutex
}

// CancelAnalysis stops any running background AI analysis immediately.
// Called from frontend via Wails IPC when the user dismisses the AI popup.
func (s *BackupService) CancelAnalysis() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.analyzeCancel != nil {
		s.analyzeCancel()
		s.analyzeCancel = nil
	}
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

// AnalyzePhoto allows analyzing an individual photo directly and saving to DB if the photo exists
func (s *BackupService) AnalyzePhoto(filePath string) (*VisionResult, error) {
	s.mu.RLock()
	ai := s.aiVision
	photoRepo := s.photoRepo
	tagRepo := s.tagRepo
	s.mu.RUnlock()

	if ai == nil {
		ai = NewAiVisionService()
	}
	ctx := s.getCtx()
	res, err := ai.AnalyzePhoto(ctx, filePath)
	if err == nil && res != nil {
		tagsStr := "-"
		if len(res.Tags) > 0 {
			tagsStr = strings.Join(res.Tags, ", ")
		}
		descStr := strings.TrimSpace(res.Description)
		if descStr == "" {
			descStr = "(ไม่มีคำอธิบาย)"
		}
		log.Printf("[AI-Vision] %s -> Tags: [%s] | คำอธิบาย: %s", filepath.Base(filePath), tagsStr, descStr)

		// Feature 13: บันทึกคำอธิบายและ Tag ลง DB ทันทีเมื่อวิเคราะห์ภาพสำเร็จ
		if photoRepo != nil {
			dest := filepath.Dir(filePath)
			filename := filepath.Base(filePath)
			photo, findErr := photoRepo.FindPhotoByFile(ctx, dest, filename)
			if findErr == nil && photo != nil && photo.ID > 0 {
				if res.Description != "" {
					_ = photoRepo.UpdatePhotoDescription(ctx, photo.ID, res.Description)
				}
				if tagRepo != nil && len(res.Tags) > 0 {
					_ = tagRepo.AddPhotoTagsInBatch(ctx, map[int32][]string{photo.ID: res.Tags}, "ai")
				}
			}
		}
	}
	return res, err
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
		log.Printf("[BackupService] ⚠️ Integrity check warning for %s: %v", dest, err)
		return &IntegrityResult{IsExactMatch: true}, nil
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

// ListAllDB lists all photo records (active, missing, deleted) for destPath (Feature 20 & 24)
func (s *BackupService) ListAllDB(destPath string) ([]*modelGen.Photo, error) {
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	return s.photoRepo.GetAllPhotosByDest(s.getCtx(), destPath)
}


// Backup copies selected or all files to destination (Feature 6: Backup / Copy - preserves source)
func (s *BackupService) Backup(targets []string, isAll bool) (MoveSummary, error) {
	return s.processTransfer(targets, isAll, false)
}

// Move moves selected or all files to destination (Feature 6: Move - removes source)
func (s *BackupService) Move(targets []string, isAll bool) (MoveSummary, error) {
	return s.processTransfer(targets, isAll, true)
}

func (s *BackupService) processTransfer(targets []string, isAll bool, deleteSource bool) (MoveSummary, error) {
	s.transferMu.Lock()
	defer s.transferMu.Unlock()

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

	s.mu.RLock()
	aiSvc := s.aiVision
	s.mu.RUnlock()

	copyWorkers := 16
	duplicateGuard := true
	autoAnalyze := true
	if cfgAi, ok := aiSvc.(ConfigurableAiVisionService); ok {
		copyWorkers = cfgAi.GetCopyWorkers()
		duplicateGuard = cfgAi.IsDuplicateGuardEnabled()
		autoAnalyze = cfgAi.IsAutoAnalyzeEnabled()
	}

	totalToMove := len(filesToMove)
	numWorkers := copyWorkers
	if numWorkers < 1 {
		numWorkers = 16
	}
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

				var shaHash string
				var size int64
				var err error

				if deleteSource {
					shaHash, size, err = streamMoveFile(srcPath, dstPath, buf)
				} else {
					shaHash, size, err = streamCopyFile(srcPath, dstPath, buf)
				}

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
			if duplicateGuard && res.sha256 != "" && shaMap[res.sha256] {
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

		// Part C (Features 12, 13, 14): AI Vision Analysis runs in background
		// so Move returns immediately after files are copied and DB records inserted.
		s.mu.RLock()
		ai := s.aiVision
		tagRepo := s.tagRepo
		s.mu.RUnlock()

		if autoAnalyze && ai != nil && ai.IsAvailable() {
			analyzeCtx, cancel := context.WithCancel(s.getCtx())
			s.mu.Lock()
			s.analyzeCancel = cancel
			s.mu.Unlock()
			go s.analyzeMovedPhotos(analyzeCtx, dst, movedPhotos, ai, tagRepo)
		}
	}

	return summary, nil
}

// analyzeMovedPhotos runs AI vision analysis in the background in strict FIFO order (never jumps the queue).
func (s *BackupService) analyzeMovedPhotos(ctx context.Context, dst string, photos []*modelGen.Photo, ai AiVisionService, tagRepo repository.TagRepository) {
	s.aiAnalyzeMu.Lock()
	defer s.aiAnalyzeMu.Unlock()

	defer func() {
		s.mu.Lock()
		s.analyzeCancel = nil
		s.mu.Unlock()
	}()

	var validPhotos []*modelGen.Photo
	for _, p := range photos {
		if isImageFile(p.FileName) {
			validPhotos = append(validPhotos, p)
		}
	}
	totalPhotos := len(validPhotos)
	if totalPhotos == 0 {
		return
	}

	startAnalyze := time.Now()

	// Build a filename→ID map using existing IDs from movedPhotos first,
	// then fill from GetAllPhotosByDest so both active and checked photos are found
	idMap := make(map[string]int32, totalPhotos)
	for _, p := range photos {
		if p != nil && p.ID > 0 {
			idMap[strings.ToLower(p.FileName)] = p.ID
		}
	}
	dbPhotos, qErr := s.photoRepo.GetAllPhotosByDest(ctx, dst)
	if qErr == nil {
		for _, dbP := range dbPhotos {
			if dbP != nil && dbP.ID > 0 {
				idMap[strings.ToLower(dbP.FileName)] = dbP.ID
			}
		}
	}

	batchAi, isBatchCapable := ai.(BatchAiVisionService)
	batchSize := 1
	numWorkers := 4
	if isBatchCapable {
		batchSize = batchAi.BatchSize()
		if batchSize < 1 {
			batchSize = 20
		}
		numWorkers = batchAi.WorkerConcurrency()
		if numWorkers < 1 {
			numWorkers = 8
		}
	}

	type photoBatch struct {
		items []*modelGen.Photo
	}

	var batches []photoBatch
	var allBatchPaths [][]string
	for i := 0; i < totalPhotos; i += batchSize {
		end := i + batchSize
		if end > totalPhotos {
			end = totalPhotos
		}
		slice := validPhotos[i:end]
		batches = append(batches, photoBatch{items: slice})
		if isBatchCapable && len(slice) > 1 {
			paths := make([]string, len(slice))
			for j, p := range slice {
				paths[j] = filepath.Join(dst, p.FileName)
			}
			allBatchPaths = append(allBatchPaths, paths)
		}
	}

	if isBatchCapable && len(allBatchPaths) > 0 {
		go batchAi.PrebuildBatchCollages(ctx, allBatchPaths)
	}

	if numWorkers > len(batches) {
		numWorkers = len(batches)
	}

	log.Printf("[AI-Vision] เริ่มต้นวิเคราะห์ความเร็วสูง %d รูป (%d Batches, Batch ละ %d รูป, %d Workers)", totalPhotos, len(batches), batchSize, numWorkers)

	batchCh := make(chan photoBatch, len(batches))
	for _, b := range batches {
		batchCh <- b
	}
	close(batchCh)

	var completedCount int32
	var successBatches int32
	var dbWriteMu sync.Mutex
	var firstErrMu sync.Mutex
	var firstErr error
	var firstErrFile string

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(wID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[AI-Vision] Worker %d กู้คืนจากข้อผิดพลาด: %v", wID, r)
				}
			}()

			// Stagger initial worker start by 90ms per worker so requests flow in a smooth continuous pipeline
			if wID > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(wID*90) * time.Millisecond):
				}
			}

			for batch := range batchCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				paths := make([]string, len(batch.items))
				for i, p := range batch.items {
					paths[i] = filepath.Join(dst, p.FileName)
				}

				var results []*VisionResult
				var err error
				for retry := 0; retry < 4; retry++ {
					select {
					case <-ctx.Done():
						return
					default:
					}

					if isBatchCapable && len(paths) > 1 {
						results, err = batchAi.AnalyzePhotoBatch(ctx, paths)
					} else {
						res, singleErr := ai.AnalyzePhoto(ctx, paths[0])
						if singleErr == nil {
							results = []*VisionResult{res}
							err = nil
						} else {
							err = singleErr
						}
					}

					if err == nil && len(results) > 0 {
						break
					}
					backoff := time.Duration(retry+1) * 400 * time.Millisecond
					log.Printf("[AI-Vision] กำลังลอง Batch ใหม่ (%s) รอบที่ %d/4 (พัก %v)...", batch.items[0].FileName, retry+1, backoff)
					time.Sleep(backoff)
				}

				// Fallback rescue: if a multi-photo batch still failed, split into two half-batches to rescue photos
				if err != nil && isBatchCapable && len(paths) > 2 {
					mid := len(paths) / 2
					res1, err1 := batchAi.AnalyzePhotoBatch(ctx, paths[:mid])
					res2, err2 := batchAi.AnalyzePhotoBatch(ctx, paths[mid:])
					if err1 == nil || err2 == nil {
						results = make([]*VisionResult, len(paths))
						if err1 == nil {
							copy(results[:mid], res1)
						}
						if err2 == nil {
							copy(results[mid:], res2)
						}
						err = nil
					}
				}

				if err != nil {
					log.Printf("[AI-Vision] Batch วิเคราะห์ล้มเหลวถาวร (%s): %v", batch.items[0].FileName, err)
					firstErrMu.Lock()
					if firstErr == nil {
						firstErr = err
						firstErrFile = batch.items[0].FileName
					}
					firstErrMu.Unlock()
				} else {
					atomic.AddInt32(&successBatches, 1)

					descUpdates := make(map[int32]string, len(batch.items))
					tagUpdates := make(map[int32][]string, len(batch.items))

					for i, p := range batch.items {
						if i >= len(results) || results[i] == nil {
							continue
						}
						vr := results[i]

						tagsStr := "-"
						if len(vr.Tags) > 0 {
							tagsStr = strings.Join(vr.Tags, ", ")
						}
						descStr := strings.TrimSpace(vr.Description)
						if descStr == "" {
							descStr = "(ไม่มีคำอธิบาย)"
						}
						log.Printf("[AI-Vision] %s -> Tags: [%s] | คำอธิบาย: %s", p.FileName, tagsStr, descStr)

						realID := idMap[strings.ToLower(p.FileName)]
						if realID <= 0 {
							// Fallback lookup directly
							if found, fErr := s.photoRepo.FindPhotoByFile(ctx, dst, p.FileName); fErr == nil && found != nil && found.ID > 0 {
								realID = found.ID
								idMap[strings.ToLower(p.FileName)] = realID
							}
						}
						if realID <= 0 {
							if vr.Description != "" {
								_ = s.photoRepo.UpdatePhotoDescriptionByFile(ctx, dst, p.FileName, vr.Description)
							}
							continue
						}
						if vr.Description != "" {
							descUpdates[realID] = vr.Description
						}
						if len(vr.Tags) > 0 {
							tagUpdates[realID] = vr.Tags
						}
					}

					if len(descUpdates) > 0 || (tagRepo != nil && len(tagUpdates) > 0) {
						dbWriteMu.Lock()
						if len(descUpdates) > 0 {
							_ = s.photoRepo.UpdatePhotoDescriptionsInBatch(ctx, descUpdates)
						}
						if tagRepo != nil && len(tagUpdates) > 0 {
							_ = tagRepo.AddPhotoTagsInBatch(ctx, tagUpdates, "ai")
						}
						dbWriteMu.Unlock()
					}
				}

				// Always increment completedCount so UI progress never gets stuck below totalPhotos
				done := int(atomic.AddInt32(&completedCount, int32(len(batch.items))))
				lastFile := batch.items[len(batch.items)-1].FileName
				safeEventsEmit(s.ctx, "ai_analysis_progress", map[string]interface{}{
					"current":  done,
					"total":    totalPhotos,
					"filename": lastFile,
				})
			}
		}(w)
	}

	wg.Wait()

	if atomic.LoadInt32(&successBatches) == 0 && firstErr != nil {
		safeEventsEmit(s.ctx, "ai_analysis_error", map[string]interface{}{
			"filename": firstErrFile,
			"error":    firstErr.Error(),
		})
		return
	}

	lastFile := ""
	if len(validPhotos) > 0 {
		lastFile = validPhotos[len(validPhotos)-1].FileName
	}
	safeEventsEmit(s.ctx, "ai_analysis_progress", map[string]interface{}{
		"current":  totalPhotos,
		"total":    totalPhotos,
		"filename": lastFile,
		"done":     true,
	})

	log.Printf("[AI-Vision] วิเคราะห์เสร็จสิ้น %d รูป ใช้เวลาทั้งหมด %v", totalPhotos, time.Since(startAnalyze))
}

// streamCopyFile streams the file content to destination and computes SHA-256 simultaneously,
// preserving the original source file (Copy instead of Move).
func streamCopyFile(srcPath, dstPath string, buffer []byte) (string, int64, error) {
	fi, err := os.Stat(srcPath)
	if err != nil {
		return "", 0, err
	}
	size := fi.Size()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", 0, err
	}
	defer dstFile.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(dstFile, hasher)

	_, copyErr := io.CopyBuffer(writer, srcFile, buffer)
	if copyErr != nil {
		dstFile.Close()
		_ = os.Remove(dstPath)
		return "", 0, copyErr
	}

	return hex.EncodeToString(hasher.Sum(nil)), size, nil
}

// streamMoveFile implements Feature 6 (Move):
// Tries os.Rename first. If across drives, fallback to stream copy buffer and remove source file.
func streamMoveFile(srcPath, dstPath string, buffer []byte) (string, int64, error) {
	fi, err := os.Stat(srcPath)
	if err != nil {
		return "", 0, err
	}
	size := fi.Size()

	// Attempt fast Rename first (works on same drive/filesystem)
	renameErr := os.Rename(srcPath, dstPath)
	if renameErr == nil {
		hash := fastFileHash(dstPath)
		return hash, size, nil
	}

	// Fallback to cross-drive stream copy buffer and remove source
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", 0, err
	}

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		srcFile.Close()
		return "", 0, err
	}

	hasher := sha256.New()
	writer := io.MultiWriter(dstFile, hasher)

	_, copyErr := io.CopyBuffer(writer, srcFile, buffer)
	dstFile.Close()

	if copyErr != nil {
		srcFile.Close()
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
	start := time.Now()
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return IntegrityResult{}, errors.New("destination is not set! Please set destination directory first")
	}

	allPhotos, err := s.photoRepo.GetAllPhotosByDest(s.getCtx(), destPath)
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
	var toActive []string
	var toMissing []string
	matched := 0

	for _, p := range allPhotos {
		if p.Status == "deleted" {
			// Intentionally deleted photos are not counted as missing
			continue
		}
		if diskMap[strings.ToLower(p.FileName)] {
			matched++
			if p.Status == "missing" {
				toActive = append(toActive, p.FileName)
			}
		} else {
			missingInDisk = append(missingInDisk, p.FileName)
			if p.Status != "missing" {
				toMissing = append(toMissing, p.FileName)
			}
		}
	}

	// High Performance: Atomic Batch Status Synchronization in SQLite
	if len(toActive) > 0 {
		_ = s.photoRepo.MarkPhotosActiveInBatch(s.getCtx(), destPath, toActive)
	}
	if len(toMissing) > 0 {
		_ = s.photoRepo.MarkPhotosMissingInBatch(s.getCtx(), destPath, toMissing)
	}

	duration := time.Since(start)
	ms := duration.Milliseconds()
	formattedDuration := fmt.Sprintf("%.1f ms", float64(duration.Microseconds())/1000.0)

	result := IntegrityResult{
		TotalDB:           len(allPhotos),
		TotalDisk:         len(diskFiles),
		MissingInDisk:     missingInDisk,
		MatchedCount:      matched,
		IsExactMatch:      len(missingInDisk) == 0,
		DurationMs:        ms,
		DurationFormatted: formattedDuration,
	}

	if len(missingInDisk) > 0 {
		result.AlertMessage = fmt.Sprintf("Missing %d file(s) detected: recorded in DB but missing from disk!", len(missingInDisk))
	} else {
		result.AlertMessage = "Integrity scan passed: All DB records match disk files."
	}

	return result, nil
}

// Delete removes files from destination disk and sets status = 'deleted' in DB (Feature 23 & 25)
func (s *BackupService) Delete(targets []string, isAll bool) (int, error) {
	return s.DeleteFromDest(s.GetDest(), targets, isAll)
}

// DeleteFromDest removes files from the given destination disk and sets status = 'deleted' in DB (Feature 23 & 25)
// Uses Concurrency Worker Pool (16 goroutines) for parallel disk deletion and atomic batch status update in DB.
func (s *BackupService) DeleteFromDest(destPath string, targets []string, isAll bool) (int, error) {
	if destPath == "" {
		destPath = s.GetDest()
	}
	if destPath == "" {
		return 0, errors.New("destination is not set! Please set destination directory first")
	}

	if isAll {
		diskFiles, _ := listDirectoryFiles(destPath)
		var deletedCount atomic.Int64
		var wg sync.WaitGroup
		sem := make(chan struct{}, 16)

		for _, f := range diskFiles {
			wg.Add(1)
			go func(fileName string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				p := filepath.Join(destPath, fileName)
				if err := os.Remove(p); err == nil {
					deletedCount.Add(1)
				}
			}(f)
		}
		wg.Wait()

		// Soft-delete in DB (update status = 'deleted' to preserve history)
		_, _ = s.photoRepo.MarkAllPhotosDeleted(s.getCtx(), destPath)
		return int(deletedCount.Load()), nil
	}

	// Filter valid non-empty targets
	var validTargets []string
	for _, fname := range targets {
		fname = strings.TrimSpace(fname)
		if fname != "" {
			validTargets = append(validTargets, fname)
		}
	}

	if len(validTargets) == 0 {
		return 0, nil
	}

	var deletedCount atomic.Int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)

	for _, fname := range validTargets {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			p := filepath.Join(destPath, fileName)
			if err := os.Remove(p); err == nil {
				deletedCount.Add(1)
			}
		}(fname)
	}
	wg.Wait()

	// Atomic batch update status = 'deleted' in DB
	_, _ = s.photoRepo.MarkPhotosDeletedInBatch(s.getCtx(), destPath, validTargets)
	return int(deletedCount.Load()), nil
}


func (s *BackupService) CleanAll() error {
	s.mu.Lock()
	s.source = ""
	s.dest = ""
	s.mu.Unlock()

	return s.photoRepo.CleanPhotos(s.getCtx())
}

// SearchPhotosByDescription searches photos by AI-generated description text
func (s *BackupService) SearchPhotosByDescription(dest string, query string, limit int, offset int) ([]*modelGen.Photo, error) {
	if dest == "" {
		dest = s.GetDest()
	}
	if dest == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	if limit <= 0 {
		limit = 50
	}
	return s.photoRepo.SearchPhotosByDescription(s.getCtx(), dest, query, limit, offset)
}

// SearchPhotosByTag searches photos by tag name
func (s *BackupService) SearchPhotosByTag(dest string, tagName string, limit int, offset int) ([]*modelGen.Photo, error) {
	if dest == "" {
		dest = s.GetDest()
	}
	if dest == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	if limit <= 0 {
		limit = 50
	}
	return s.photoRepo.SearchPhotosByTag(s.getCtx(), dest, tagName, limit, offset)
}

// GetTagsWithCount returns all tags with their photo counts for the destination
func (s *BackupService) GetTagsWithCount(dest string) ([]repository.TagCount, error) {
	if dest == "" {
		dest = s.GetDest()
	}
	if dest == "" {
		return nil, errors.New("destination is not set! Please set destination directory first")
	}
	return s.photoRepo.GetTagsWithCount(s.getCtx(), dest)
}

func (s *BackupService) getConfigurableAi() ConfigurableAiVisionService {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.aiVision == nil {
		s.aiVision = NewAiVisionService()
	}
	if cfg, ok := s.aiVision.(ConfigurableAiVisionService); ok {
		return cfg
	}
	return nil
}

// GetAppSettings returns current system settings and multi-project API pool status
func (s *BackupService) GetAppSettings() (*dto.AppSettingsDTO, error) {
	cfg := s.getConfigurableAi()
	if cfg == nil {
		return nil, errors.New("AI Vision service is not configurable")
	}
	return cfg.GetAppSettings(), nil
}

// SaveProjectApiKey adds a new project + API key, or updates an existing project with a new API key
func (s *BackupService) SaveProjectApiKey(projectID, projectName, projectNumber, apiKey string) (*dto.AppSettingsDTO, error) {
	cfg := s.getConfigurableAi()
	if cfg == nil {
		return nil, errors.New("AI Vision service is not configurable")
	}
	return cfg.SaveProjectApiKey(projectID, projectName, projectNumber, apiKey)
}

// ToggleProjectApiKey enables or disables a specific project's API key in the pool
func (s *BackupService) ToggleProjectApiKey(projectID string, enabled bool) (*dto.AppSettingsDTO, error) {
	cfg := s.getConfigurableAi()
	if cfg == nil {
		return nil, errors.New("AI Vision service is not configurable")
	}
	return cfg.ToggleProjectApiKey(projectID, enabled)
}

// DeleteProjectApiKey deletes a project entry from the pool
func (s *BackupService) DeleteProjectApiKey(projectID string) (*dto.AppSettingsDTO, error) {
	cfg := s.getConfigurableAi()
	if cfg == nil {
		return nil, errors.New("AI Vision service is not configurable")
	}
	return cfg.DeleteProjectApiKey(projectID)
}

// UpdateSystemSettings updates general backup and AI settings
func (s *BackupService) UpdateSystemSettings(autoAnalyze bool, primaryModel string, duplicateGuard bool, copyWorkers int) (*dto.AppSettingsDTO, error) {
	cfg := s.getConfigurableAi()
	if cfg == nil {
		return nil, errors.New("AI Vision service is not configurable")
	}
	return cfg.UpdateSystemSettings(autoAnalyze, primaryModel, duplicateGuard, copyWorkers)
}

// ResetAiQuotaCooldowns clears all temporary RPM and daily RPD cooldowns across all projects
func (s *BackupService) ResetAiQuotaCooldowns() (*dto.AppSettingsDTO, error) {
	cfg := s.getConfigurableAi()
	if cfg == nil {
		return nil, errors.New("AI Vision service is not configurable")
	}
	return cfg.ResetQuotaCooldowns()
}

func safeEventsEmit(ctx context.Context, eventName string, data ...interface{}) {
	if ctx == nil {
		return
	}
	if ctx.Value("frontend") == nil && ctx.Value("events") == nil && ctx.Value("app") == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	runtime.EventsEmit(ctx, eventName, data...)
}

