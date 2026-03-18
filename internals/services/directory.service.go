package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"homeserver/internals/models"
	"homeserver/internals/repos"
)

type DirectoryService interface {
	CreateDir(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (*models.Node, error)
}

type directoryService struct {
	repo repos.DirectoryRepository
}

func NewDirectoryService(repo repos.DirectoryRepository) DirectoryService {
	return &directoryService{repo: repo}
}

func (s *directoryService) CreateDir(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (*models.Node, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	directoryID := uuid.New()
	return s.repo.CreateDir(ctx, directoryID, name, parentID, userID)
}
