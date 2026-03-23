package repos

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"homeserver/internals/models"
)

var ErrStorageUpdateRejected = errors.New("storage update rejected")

type StorageRepository interface {
	GetAllActive(ctx context.Context) ([]models.Storage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Storage, error)
	UpdateUsedSpaceByID(ctx context.Context, id uuid.UUID, delta int64) error
}

type storageRepository struct {
	db *gorm.DB
}

func NewStorageRepository(db *gorm.DB) StorageRepository {
	return &storageRepository{db: db}
}

func (r *storageRepository) GetAllActive(ctx context.Context) ([]models.Storage, error) {
	var storages []models.Storage

	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("created_at ASC, id ASC").
		Find(&storages).Error
	if err != nil {
		return nil, err
	}

	return storages, nil
}

func (r *storageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Storage, error) {
	var storage models.Storage

	err := r.db.WithContext(ctx).First(&storage, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &storage, nil
}

func (r *storageRepository) UpdateUsedSpaceByID(ctx context.Context, id uuid.UUID, delta int64) error {
	query := r.db.WithContext(ctx).
		Model(&models.Storage{}).
		Where("id = ?", id).
		Where("status = ?", "active")

	if delta >= 0 {
		query = query.Where("(total_space - used_space) >= ?", delta)
	} else {
		query = query.Where("used_space >= ?", -delta)
	}

	result := query.Update("used_space", gorm.Expr("used_space + ?", delta))
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrStorageUpdateRejected
	}

	return nil
}
