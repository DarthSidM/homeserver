package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"homeserver/internals/models"
	"homeserver/internals/repos"
)

type NodeService interface {
	ListNodes(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]models.Node, error)
	RenameNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, newName string) (*models.Node, error)
	DeleteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error
	MarkFavouriteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (bool, error)
	GetFavouriteNodes(ctx context.Context, userID uuid.UUID) ([]models.Node, error)
	SearchNodes(ctx context.Context, userID uuid.UUID, query string) ([]models.Node, error)
}

type nodeService struct {
	repo repos.NodeRepository
}

func NewNodeService(repo repos.NodeRepository) NodeService {
	return &nodeService{repo: repo}
}

func (s *nodeService) ListNodes(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]models.Node, error) {
	return s.repo.ListByParentID(ctx, userID, parentID)
}

func (s *nodeService) RenameNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, newName string) (*models.Node, error) {

	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, errors.New("name cannot be empty")
	}

	node, err := s.repo.GetByID(ctx, userID, nodeID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found")
	}

	if node.Name == newName {
		return node, nil
	}

	exists, err := s.repo.ExistsByNameAndParent(
		ctx,
		userID,
		newName,
		node.ParentID,
	)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("a node with this name already exists in this directory")
	}

	if err := s.repo.UpdateName(ctx, userID, nodeID, newName); err != nil {
		return nil, err
	}

	node.Name = newName
	return node, nil
}

func (s *nodeService) DeleteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error {

	node, err := s.repo.GetByID(ctx, userID, nodeID)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("node not found")
	}

	if node.Type == "file" {
		return s.repo.SoftDelete(ctx, userID, nodeID)
	}

	// directory
	return s.repo.SoftDeleteSubtree(ctx, userID, nodeID)
}

func (s *nodeService) MarkFavouriteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, errors.New("invalid user id")
	}

	node, err := s.repo.GetByID(ctx, userID, nodeID)
	if err != nil {
		return false, err
	}
	if node == nil {
		return false, errors.New("node not found")
	}

	isFavourite, err := s.repo.IsFavourite(ctx, userID, nodeID)
	if err != nil {
		return false, err
	}

	if isFavourite {
		return false, s.repo.DeleteFavourite(ctx, userID, nodeID)
	}

	_, err = s.repo.CreateFavourite(ctx, userID, nodeID)
	return true, err
}

func (s *nodeService) GetFavouriteNodes(ctx context.Context, userID uuid.UUID) ([]models.Node, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	return s.repo.ListFavouriteNodes(ctx, userID)
}

func (s *nodeService) SearchNodes(ctx context.Context, userID uuid.UUID, query string) ([]models.Node, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, errors.New("search query cannot be empty")
	}

	return s.repo.SearchNodes(ctx, userID, trimmedQuery)
}
