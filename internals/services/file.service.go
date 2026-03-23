package services

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"homeserver/internals/models"
	"homeserver/internals/repos"
)

var (
	ErrNoActiveStorage       = errors.New("no active storage available")
	ErrInsufficientDiskSpace = errors.New("insufficient disk space")
)

type FileService interface {
	UploadFile(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, originalName string, file multipart.File, size int64) (*models.Node, error)
}

type fileService struct {
	repo        repos.FileRepository
	storageRepo repos.StorageRepository
}

func NewFileService(repo repos.FileRepository, storageRepo repos.StorageRepository) FileService {
	return &fileService{repo: repo, storageRepo: storageRepo}
}

func (s *fileService) UploadFile(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, originalName string, file multipart.File, size int64) (*models.Node, error) {

	// defer file.Close()

	originalName = strings.TrimSpace(originalName)
	if originalName == "" {
		return nil, errors.New("file name cannot be empty")
	}

	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	if size < 0 {
		return nil, ErrInsufficientDiskSpace
	}

	selectedStorage, err := s.selectStorage(ctx, size)
	if err != nil {
		return nil, err
	}

	fileID := uuid.New()
	fileExt := strings.TrimSpace(filepath.Ext(originalName))

	if err := s.storageRepo.UpdateUsedSpaceByID(ctx, selectedStorage.ID, size); err != nil {
		if errors.Is(err, repos.ErrStorageUpdateRejected) {
			return nil, ErrInsufficientDiskSpace
		}
		return nil, err
	}

	storageDir := selectedStorage.MountPath

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		_ = s.storageRepo.UpdateUsedSpaceByID(ctx, selectedStorage.ID, -size)
		return nil, err
	}

	storedFilePath := filepath.Join(storageDir, fileID.String()+fileExt)

	dst, err := os.Create(storedFilePath)
	if err != nil {
		_ = s.storageRepo.UpdateUsedSpaceByID(ctx, selectedStorage.ID, -size)
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		_ = s.storageRepo.UpdateUsedSpaceByID(ctx, selectedStorage.ID, -size)
		_ = os.Remove(storedFilePath)
		return nil, err
	}

	node, err := s.repo.CreateFile(ctx, fileID, selectedStorage.ID, originalName, parentID, userID, size)
	if err != nil {
		_ = s.storageRepo.UpdateUsedSpaceByID(ctx, selectedStorage.ID, -size)
		_ = os.Remove(storedFilePath)
		return nil, err
	}

	return node, nil
}

func (s *fileService) selectStorage(ctx context.Context, size int64) (*models.Storage, error) {
	activeDisks, err := s.storageRepo.GetAllActive(ctx)
	if err != nil {
		return nil, err
	}

	if len(activeDisks) == 0 {
		return nil, ErrNoActiveStorage
	}

	for i := range activeDisks {
		available := activeDisks[i].TotalSpace - activeDisks[i].UsedSpace
		if available >= size {
			return &activeDisks[i], nil
		}
	}

	return nil, ErrInsufficientDiskSpace
}
