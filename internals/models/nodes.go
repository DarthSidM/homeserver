package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Node struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ParentID  *uuid.UUID     `gorm:"type:uuid;index"`
	UserID    uuid.UUID      `gorm:"type:uuid;index;not null"`
	Name      string         `gorm:"not null"`
	Type      string         `gorm:"type:varchar(20);not null"` // file | directory
	Size      int64          `gorm:"default:0"`
	StorageID *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
