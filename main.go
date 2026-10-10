package main

import (
	"PhotoVault/repository"
	"PhotoVault/service"
	"PhotoVault/utils"
	"embed"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func createThumbnailHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/thumbnail") {
			filePath := r.URL.Query().Get("path")
			if filePath == "" {
				http.Error(w, "missing path", http.StatusBadRequest)
				return
			}

			// Try high-performance thumbnail generator (handles JPG, PNG, HEIC, RAW)
			thumbBytes, err := utils.GetThumbnailBytes(filePath, 360)
			if err == nil && len(thumbBytes) > 0 {
				w.Header().Set("Content-Type", "image/jpeg")
				w.Header().Set("Cache-Control", "public, max-age=86400")
				_, _ = w.Write(thumbBytes)
				return
			}

			// Fallback: serve raw file if standard image
			if _, statErr := os.Stat(filePath); statErr == nil {
				http.ServeFile(w, r, filePath)
				return
			}

			http.NotFound(w, r)
			return
		}

		http.NotFound(w, r)
	})
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
	aiVisionService := service.NewAiVisionService()
	backupService := service.NewBackupService(photoRepo, tagRepo, aiVisionService)

	// 3. Create App instance with injected BackupService
	app := NewApp(backupService)

	// 4. Run Wails application
	err = wails.Run(&options.App{
		Title:  "PhotoVault",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: createThumbnailHandler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			backupService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
