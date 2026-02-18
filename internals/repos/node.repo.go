package repos

import (
	"context"
	"errors"
	"homeserver/internals/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NodeRepository interface {
	ListByParentID(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]models.Node, error)
	GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*models.Node, error)
	ExistsByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (bool, error)
	UpdateName(ctx context.Context, userID uuid.UUID, id uuid.UUID, newName string) error
	SoftDelete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	SoftDeleteSubtree(ctx context.Context, userID uuid.UUID, rootID uuid.UUID) error
}

type nodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) ListByParentID(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]models.Node, error) {

	var nodes []models.Node

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("parent_id IS NOT DISTINCT FROM ?", parentID).
		Order("type DESC, name ASC") // directories first if desired

	if err := query.Find(&nodes).Error; err != nil {
		return nil, err
	}

	return nodes, nil
}

func (r *nodeRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*models.Node, error) {

	var node models.Node

	if err := r.db.WithContext(ctx).
		First(&node, "id = ? AND user_id = ?", id, userID).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &node, nil
}

func (r *nodeRepository) ExistsByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Node{}).
		Where("user_id = ?", userID).
		Where("name = ?", name).
		Where("parent_id IS NOT DISTINCT FROM ?", parentID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *nodeRepository) UpdateName(ctx context.Context, userID uuid.UUID, id uuid.UUID, newName string) error {

	return r.db.WithContext(ctx).
		Model(&models.Node{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("name", newName).Error
}

func (r *nodeRepository) SoftDelete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {

	return r.db.WithContext(ctx).
		Delete(&models.Node{}, "id = ? AND user_id = ?", id, userID).Error
}

func (r *nodeRepository) SoftDeleteSubtree(ctx context.Context, userID uuid.UUID, rootID uuid.UUID) error {

	query := `
	WITH RECURSIVE subtree AS (
	    SELECT id FROM nodes WHERE id = ? AND user_id = ?
	    UNION ALL
	    SELECT n.id
	    FROM nodes n
	    INNER JOIN subtree s ON n.parent_id = s.id
	    WHERE n.user_id = ?
	)
	UPDATE nodes
	SET deleted_at = ?
	WHERE user_id = ? AND id IN (SELECT id FROM subtree);
	`
	return r.db.WithContext(ctx).
		Exec(query, rootID, userID, userID, time.Now(), userID).
		Error
}
