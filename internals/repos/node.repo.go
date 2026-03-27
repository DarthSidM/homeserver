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
	IsFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (bool, error)
	CreateFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (*models.Favourite, error)
	DeleteFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error
	ListFavouriteNodes(ctx context.Context, userID uuid.UUID) ([]models.Node, error)
	SearchNodes(ctx context.Context, userID uuid.UUID, query string) ([]models.Node, error)
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

func (r *nodeRepository) IsFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Favourite{}).
		Where("user_id = ? AND node_id = ?", userID, nodeID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *nodeRepository) CreateFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (*models.Favourite, error) {
	favourite := &models.Favourite{
		UserID: userID,
		NodeID: nodeID,
	}

	if err := r.db.WithContext(ctx).Create(favourite).Error; err != nil {
		return nil, err
	}

	return favourite, nil
}

func (r *nodeRepository) DeleteFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.Favourite{}, "user_id = ? AND node_id = ?", userID, nodeID).
		Error
}

func (r *nodeRepository) ListFavouriteNodes(ctx context.Context, userID uuid.UUID) ([]models.Node, error) {
	var nodes []models.Node

	err := r.db.WithContext(ctx).
		Model(&models.Node{}).
		Joins("JOIN favourites ON favourites.node_id = nodes.id").
		Where("favourites.user_id = ?", userID).
		Where("nodes.user_id = ?", userID).
		Order("favourites.created_at DESC").
		Find(&nodes).Error

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (r *nodeRepository) SearchNodes(ctx context.Context, userID uuid.UUID, query string) ([]models.Node, error) {
	var nodes []models.Node

	// Get underlying sql.DB connection to bypass GORM's query processing
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}

	// First verify the FTS5 table exists
	var tableExists int
	checkTableSQL := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='nodes_fts'"
	err = sqlDB.QueryRowContext(ctx, checkTableSQL).Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Fall back to LIKE search if FTS5 table doesn't exist
		likeQuery := "%" + query + "%"
		err = r.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Where("deleted_at IS NULL").
			Where("name LIKE ?", likeQuery).
			Order("name ASC").
			Find(&nodes).Error
		return nodes, err
	}

	// Try FTS5 search with direct query
	sql := `
	SELECT n.id, n.parent_id, n.user_id, n.name, n.type, n.size, n.storage_id, n.created_at, n.updated_at, n.deleted_at
	FROM nodes n
	INNER JOIN nodes_fts fts ON n.id = fts.node_id
	WHERE n.user_id = ? 
	AND n.deleted_at IS NULL 
	AND fts MATCH ?
	ORDER BY n.name ASC
	`

	rows, err := sqlDB.QueryContext(ctx, sql, userID.String(), query)
	if err != nil {
		// If FTS5 query fails, fall back to LIKE
		likeQuery := "%" + query + "%"
		err = r.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Where("deleted_at IS NULL").
			Where("name LIKE ?", likeQuery).
			Order("name ASC").
			Find(&nodes).Error
		return nodes, err
	}
	defer rows.Close()

	for rows.Next() {
		var node models.Node
		if err := rows.Scan(
			&node.ID, &node.ParentID, &node.UserID, &node.Name, &node.Type,
			&node.Size, &node.StorageID, &node.CreatedAt, &node.UpdatedAt, &node.DeletedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}
