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

	// NOTE: FTS5 initialization is deferred until after AutoMigrate
	// to ensure underlying tables (like `nodes`) exist.

	log.Println("connected to sqlite database")
	return db, nil
	// return db, nil
}

// InitializeFTS5 creates the FTS5 virtual table and triggers for nodes.
// Call this after running AutoMigrate so the `nodes` table exists.
func InitializeFTS5(db *gorm.DB) error {
	// Create FTS5 virtual table for full-text search on nodes
	createVirtualTableSQL := `
	CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
		node_id,
		user_id UNINDEXED,
		name
	);
	`
	if err := db.Exec(createVirtualTableSQL).Error; err != nil {
		return err
	}

	// Populate the FTS5 table with existing nodes
	populateSQL := `
	INSERT OR IGNORE INTO nodes_fts(node_id, user_id, name)
	SELECT id, user_id, name FROM nodes
	WHERE deleted_at IS NULL;
	`
	if err := db.Exec(populateSQL).Error; err != nil {
		return err
	}

	// Create triggers to keep FTS5 table synchronized with nodes table
	// Trigger for INSERT
	createTriggerSQL := `
	CREATE TRIGGER IF NOT EXISTS nodes_ai AFTER INSERT ON nodes BEGIN
	  INSERT INTO nodes_fts(node_id, user_id, name) 
	  VALUES (new.id, new.user_id, new.name);
	END;
	`
	if err := db.Exec(createTriggerSQL).Error; err != nil {
		return err
	}

	// Trigger for UPDATE
	updateTriggerSQL := `
	CREATE TRIGGER IF NOT EXISTS nodes_au AFTER UPDATE ON nodes BEGIN
	  DELETE FROM nodes_fts WHERE node_id = old.id;
	  INSERT INTO nodes_fts(node_id, user_id, name) 
	  VALUES (new.id, new.user_id, new.name);
	END;
	`
	if err := db.Exec(updateTriggerSQL).Error; err != nil {
		return err
	}

	// Trigger for DELETE (soft delete)
	deleteTriggerSQL := `
	CREATE TRIGGER IF NOT EXISTS nodes_ad AFTER DELETE ON nodes BEGIN
	  DELETE FROM nodes_fts WHERE node_id = old.id;
	END;
	`
	if err := db.Exec(deleteTriggerSQL).Error; err != nil {
		return err
	}

	return nil
}
