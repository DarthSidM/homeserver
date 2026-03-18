package repos

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"homeserver/internals/models"
)

type DirectoryRepository interface {
	CreateDir(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID, userID uuid.UUID) (*models.Node, error)
}

type directoryRepository struct {
	db *gorm.DB
}

func NewDirectoryRepository(db *gorm.DB) DirectoryRepository {
	return &directoryRepository{db: db}
}

func (r *directoryRepository) CreateDir(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID, userID uuid.UUID) (*models.Node, error) {
	dir := &models.Node{
		ID:       id,
		ParentID: parentID,
		UserID:   userID,
		Name:     name,
		Type:     "directory",
	}

	if err := r.db.WithContext(ctx).Create(dir).Error; err != nil {
		return nil, err
	}

	return dir, nil
}
