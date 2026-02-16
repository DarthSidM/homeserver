package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username string    `gorm:"uniqueIndex;not null"`
	Password string    `gorm:"not null"`
	Name     string    `gorm:"not null"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UserID == uuid.Nil {
		u.UserID = uuid.New()
	}
	return
}