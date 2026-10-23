package repository

import (
	"log"
	"os"
	"path/filepath"

	"PhotoVault/model/model"

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

	// Automatically ensure all core tables and columns exist
	if migrateErr := database.AutoMigrate(&model.Photo{}, &model.Tag{}, &model.PhotoTag{}); migrateErr != nil {
		log.Println("AutoMigrate warning:", migrateErr)
	}

	// Merge any existing duplicate tags, re-link photo_tags, and create unique constraint
	database.Exec(`
		UPDATE photo_tags 
		SET tag_id = (SELECT MIN(t2.id) FROM tags t2 WHERE LOWER(TRIM(t2.name)) = LOWER(TRIM((SELECT t1.name FROM tags t1 WHERE t1.id = photo_tags.tag_id))))
		WHERE EXISTS (SELECT 1 FROM tags t1 WHERE t1.id = photo_tags.tag_id);
	`)
	database.Exec(`
		DELETE FROM photo_tags 
		WHERE rowid NOT IN (
			SELECT MIN(rowid) FROM photo_tags GROUP BY photo_id, tag_id
		);
	`)
	database.Exec(`
		DELETE FROM tags 
		WHERE id NOT IN (
			SELECT MIN(id) FROM tags GROUP BY LOWER(TRIM(name))
		);
	`)
	database.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name_unique ON tags (LOWER(TRIM(name)));")

	db = database
	return db, nil
}