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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		return nil, err
	}
	dbPath := os.Getenv("DATABASE")
	dir, err := os.Getwd()
	dbPath = filepath.Join(dir, dbPath)
	log.Println("Database path:", dbPath)
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Println("Error connecting to database:", err)
		return nil, err
	}
	db = database
	return db, nil
}