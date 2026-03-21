package repos

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"homeserver/internals/models"
)

type FileRepository interface {
	CreateFile(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID, userID uuid.UUID, size int64) (*models.Node, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID, userID uuid.UUID, size int64) (*models.Node, error) {
	node := &models.Node{
		ID:       id,
		ParentID: parentID,
		UserID:   userID,
		Name:     name,
		Type:     "file",
		Size:     size,
	}

	if err := r.db.WithContext(ctx).Create(node).Error; err != nil {
		return nil, err
	}

	return node, nil
}
