package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Favourite struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	NodeID    uuid.UUID `gorm:"type:uuid;index;not null"`
	CreatedAt time.Time
}

func (f *Favourite) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}

func (Favourite) TableName() string {
	return "favourites"
}
