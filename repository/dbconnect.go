package repository

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB

func NewDbConnection() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}
	_ = godotenv.Load()
	dbPath := os.Getenv("DATABASE")
	if dbPath == "" {
		dbPath = "database/photo-vault.db"
	}
	if !filepath.IsAbs(dbPath) {
		dir, err := os.Getwd()
		if err == nil {
			dbPath = filepath.Join(dir, dbPath)
		}
	}
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)
	log.Println("Database path:", dbPath)
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Println("Error connecting to database:", err)
		return nil, err
	}
	database.Exec("PRAGMA journal_mode = WAL;")
	database.Exec("PRAGMA synchronous = NORMAL;")
	database.Exec("PRAGMA temp_store = MEMORY;")
	database.Exec("PRAGMA cache_size = -64000;")

	db = database
	return db, nil
}