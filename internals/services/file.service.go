package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"homeserver/internals/dtos"
	"homeserver/internals/models"
	"homeserver/internals/repos"
)

var (
	ErrNoActiveStorage       = errors.New("no active storage available")
	ErrInsufficientDiskSpace = errors.New("insufficient disk space")
	ErrFileNotFound          = errors.New("file not found")
	ErrStorageMissing        = errors.New("file storage not found")
	ErrFileMissingOnDisk     = errors.New("file is missing on disk")
	ErrNodeIsNotAFile        = errors.New("node is not a file")
)

type FileService interface {
	UploadFile(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, originalName string, file multipart.File, size int64) (*models.Node, error)
	DownloadFile(ctx context.Context, userID uuid.UUID, fileID uuid.UUID) (string, string, error)
	GetEditorConfig(ctx context.Context, userID uuid.UUID, fileID uuid.UUID, baseURL string) (*dtos.EditorConfigResponse, error)
	GetNodeByID(ctx context.Context, fileID uuid.UUID) (*models.Node, error)
	ResolveFilePath(ctx context.Context, fileID uuid.UUID) (string, string, error)
	SaveFromOnlyOfficeCallback(ctx context.Context, fileID uuid.UUID, payload []byte) error
}

type fileService struct {
	repo        repos.FileRepository
	storageRepo repos.StorageRepository
	nodeRepo    repos.NodeRepository
}

func NewFileService(repo repos.FileRepository, storageRepo repos.StorageRepository, nodeRepo repos.NodeRepository) FileService {
	return &fileService{repo: repo, storageRepo: storageRepo, nodeRepo: nodeRepo}
}

func (s *fileService) UploadFile(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, originalName string, file multipart.File, size int64) (*models.Node, error) {
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

func (s *fileService) DownloadFile(ctx context.Context, userID uuid.UUID, fileID uuid.UUID) (string, string, error) {
	if userID == uuid.Nil {
		return "", "", errors.New("invalid user id")
	}

	node, err := s.nodeRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return "", "", err
	}
	if node == nil {
		return "", "", ErrFileNotFound
	}

	if !strings.EqualFold(node.Type, "file") {
		return "", "", ErrNodeIsNotAFile
	}

	if node.StorageID == nil {
		return "", "", ErrStorageMissing
	}

	storage, err := s.storageRepo.GetByID(ctx, *node.StorageID)
	if err != nil {
		return "", "", err
	}
	if storage == nil {
		return "", "", ErrStorageMissing
	}

	fileExt := strings.TrimSpace(filepath.Ext(node.Name))
	storedFilePath := filepath.Join(storage.MountPath, node.ID.String()+fileExt)

	if _, err := os.Stat(storedFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", ErrFileMissingOnDisk
		}
		return "", "", err
	}

	return storedFilePath, node.Name, nil
}

func (s *fileService) SaveFromOnlyOfficeCallback(ctx context.Context, fileID uuid.UUID, payload []byte) error {
	var callback dtos.OnlyOfficeSavePayload
	if err := json.Unmarshal(payload, &callback); err != nil {
		return errors.New("invalid onlyoffice callback payload")
	}

	if callback.Status != 2 && callback.Status != 6 {
		return nil
	}

	fileURL := strings.TrimSpace(callback.URL)
	if fileURL == "" {
		return errors.New("invalid onlyoffice callback payload")
	}

	targetPath, _, err := s.ResolveFilePath(ctx, fileID)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch updated onlyoffice file")
	}

	tmpPath := targetPath + ".tmp"
	outFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		outFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := outFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}

func (s *fileService) GetEditorConfig(ctx context.Context, userID uuid.UUID, fileID uuid.UUID, baseURL string) (*dtos.EditorConfigResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	node, err := s.nodeRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrFileNotFound
	}

	if !strings.EqualFold(node.Type, "file") {
		return nil, ErrNodeIsNotAFile
	}

	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("invalid base url")
	}

	resolvedBaseURL, err := resolveOnlyOfficeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	version := strconv.FormatInt(node.UpdatedAt.Unix(), 10)
	fileExt := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(filepath.Ext(node.Name))), ".")
	documentType := resolveDocumentType(fileExt)
	config := &dtos.EditorConfigResponse{
		DocumentType: documentType,
		Document: dtos.EditorDocumentConfig{
			FileType: fileExt,
			Key:      fileID.String() + "-" + version,
			Title:    node.Name,
			URL:      resolvedBaseURL + "/onlyoffice/download/" + fileID.String(),
		},
		EditorConfig: dtos.EditorConfigPayload{
			CallbackURL: resolvedBaseURL + "/onlyoffice/save/" + fileID.String(),
			Mode:        "edit",
		},
	}

	return config, nil
}

func resolveDocumentType(fileType string) string {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "xlsx", "xls", "csv":
		return "cell"
	case "ppt", "pptx":
		return "slide"
	case "pdf":
		return "pdf"
	default:
		return "word"
	}
}

func (s *fileService) GetNodeByID(ctx context.Context, fileID uuid.UUID) (*models.Node, error) {
	return s.nodeRepo.GetByIDDirect(ctx, fileID)
}

func (s *fileService) ResolveFilePath(ctx context.Context, fileID uuid.UUID) (string, string, error) {
	node, err := s.nodeRepo.GetByIDDirect(ctx, fileID)
	if err != nil {
		return "", "", err
	}
	if node == nil {
		return "", "", ErrFileNotFound
	}

	if node.StorageID == nil {
		return "", "", ErrStorageMissing
	}

	storage, err := s.storageRepo.GetByID(ctx, *node.StorageID)
	if err != nil {
		return "", "", err
	}
	if storage == nil {
		return "", "", ErrStorageMissing
	}

	fileExt := strings.TrimSpace(filepath.Ext(node.Name))
	storedFilePath := filepath.Join(storage.MountPath, node.ID.String()+fileExt)

	if _, err := os.Stat(storedFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", ErrFileMissingOnDisk
		}
		return "", "", err
	}

	return storedFilePath, node.Name, nil
}

func resolveOnlyOfficeBaseURL(baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid base url")
	}

	hostname := strings.TrimSpace(parsed.Hostname())
	if hostname != "localhost" && hostname != "127.0.0.1" && hostname != "::1" {
		return strings.TrimRight(baseURL, "/"), nil
	}

	lanIP, err := machineIPv4()
	if err != nil {
		return "", errors.New("invalid base url")
	}

	port := parsed.Port()
	if port != "" {
		parsed.Host = net.JoinHostPort(lanIP, port)
	} else {
		parsed.Host = lanIP
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}

func machineIPv4() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addresses {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}

			if ip == nil {
				continue
			}

			ipv4 := ip.To4()
			if ipv4 == nil || ipv4.IsLoopback() {
				continue
			}

			return ipv4.String(), nil
		}
	}

	return "", errors.New("no active ipv4 address found")
}
