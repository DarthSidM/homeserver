package models

import (
	"time"

	"github.com/google/uuid"
)

type Storage struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name       string
	MountPath  string `gorm:"uniqueIndex"`
	DiskUUID   string `gorm:"uniqueIndex"`
	TotalSpace int64
	UsedSpace  int64
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
