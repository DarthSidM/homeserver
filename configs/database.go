package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./homeserver.db" // fallback (useful for dev)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error connecting to sqlite db: %v", err)
	}

	// IMPORTANT: improve concurrency
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1) // SQLite best practice
	}

	// Enable WAL mode (critical for performance)
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=NORMAL;")

	log.Println("connected to sqlite database")
	return db, nil
}