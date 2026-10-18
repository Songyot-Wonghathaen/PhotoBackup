package main

import (
	"PhotoVault/controller"
	"PhotoVault/repository"
	"PhotoVault/service"
	"PhotoVault/utils"

	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func serveImageFile(w http.ResponseWriter, r *http.Request, filePath string, asThumb bool) {
	if filePath == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	if asThumb {
		thumbBytes, err := utils.GetThumbnailBytes(filePath, 360)
		if err == nil && len(thumbBytes) > 0 {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			_, _ = w.Write(thumbBytes)
			return
		}
	}

	if _, statErr := os.Stat(filePath); statErr == nil {
		http.ServeFile(w, r, filePath)
		return
	}

	// Fallback to thumbnail generator (e.g. for HEIC/RAW preview)
	thumbBytes, err := utils.GetThumbnailBytes(filePath, 800)
	if err == nil && len(thumbBytes) > 0 {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(thumbBytes)
		return
	}

	http.NotFound(w, r)
}

func resolvePhotoPathByID(r *http.Request, galleryRepo repository.GalleryRepository, prefix string) string {
	if galleryRepo == nil {
		return ""
	}
	idStr := strings.TrimPrefix(r.URL.Path, prefix)
	idStr = strings.Trim(idStr, "/")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id64 <= 0 {
		return ""
	}
	photo, err := galleryRepo.FindActiveByID(r.Context(), int32(id64))
	if err != nil || photo == nil {
		return ""
	}
	return filepath.Join(photo.DestinationPath, photo.FileName)
}

func createThumbnailHandler(galleryRepo repository.GalleryRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/thumbnail"):
			serveImageFile(w, r, r.URL.Query().Get("path"), true)
			return
		case strings.HasPrefix(r.URL.Path, "/thumb/"):
			serveImageFile(w, r, resolvePhotoPathByID(r, galleryRepo, "/thumb/"), true)
			return
		case strings.HasPrefix(r.URL.Path, "/image/"):
			serveImageFile(w, r, resolvePhotoPathByID(r, galleryRepo, "/image/"), false)
			return
		}
		http.NotFound(w, r)
	})
}

func createThumbnailMiddleware(imgHandler http.Handler) assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/thumbnail") ||
				strings.HasPrefix(r.URL.Path, "/thumb/") ||
				strings.HasPrefix(r.URL.Path, "/image/") {
				imgHandler.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// 1. Connect to Database BEFORE wails.Run
	db, err := repository.NewDbConnection()
	if err != nil {
		log.Println("Error connecting to database:", err.Error())
	} else {
		log.Println("Database connection established successfully.")
	}

	// 2. Initialize Repositories and Services
	photoRepo := repository.NewPhotoRepository(db)
	tagRepo := repository.NewTagRepository(db)
	galleryRepo := repository.NewGalleryRepository(db)
	aiVisionService := service.NewAiVisionService()
	backupService := service.NewBackupService(photoRepo, tagRepo, aiVisionService)
	galleryService := service.NewGalleryService(galleryRepo)

	// 3. Create App instance with injected services
	app := NewApp(backupService, galleryService)
	galleryController := controller.NewGalleryController(galleryService)
	imgHandler := createThumbnailHandler(galleryRepo)

	// 4. Run Wails application
	err = wails.Run(&options.App{
		Title:  "PhotoVault",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Handler:    imgHandler,
			Middleware: createThumbnailMiddleware(imgHandler),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			backupService,
			galleryController,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
