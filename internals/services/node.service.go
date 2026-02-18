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
