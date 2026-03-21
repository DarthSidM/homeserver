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

type FileService interface {
	UploadFile(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, originalName string, file multipart.File, size int64) (*models.Node, error)
}

type fileService struct {
	repo repos.FileRepository
}

func NewFileService(repo repos.FileRepository) FileService {
	return &fileService{repo: repo}
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

	fileID := uuid.New()
	fileExt := strings.TrimSpace(filepath.Ext(originalName))

	baseDir := "/storage"
	storageDir := filepath.Join(baseDir, "default")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return nil, err
	}

	storedFilePath := filepath.Join(storageDir, fileID.String()+fileExt)

	dst, err := os.Create(storedFilePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	return s.repo.CreateFile(ctx, fileID, originalName, parentID, userID, size)
}
