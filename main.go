package main

import (
	"PhotoVault/repository"
	"PhotoVault/service"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

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
	backupService := service.NewBackupService(photoRepo)

	// 3. Create App instance with injected BackupService
	app := NewApp(backupService)

	// 4. Run Wails application
	err = wails.Run(&options.App{
		Title:  "PhotoVault",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
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
