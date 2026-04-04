package handlers

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"homeserver/internals/services"
)

type FileHandler struct {
	fileService     services.FileService
	userAuthService *services.UserAuthService
}

func NewFileHandler(fileService services.FileService, userAuthService *services.UserAuthService) *FileHandler {
	return &FileHandler{
		fileService:     fileService,
		userAuthService: userAuthService,
	}
}

func (h *FileHandler) UploadFile(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var parentID *uuid.UUID
	parentIDParam := strings.TrimSpace(c.Params("parentID"))
	if parentIDParam != "" && !strings.EqualFold(parentIDParam, "null") {
		parsedParentID, err := uuid.Parse(parentIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid parent id"})
		}
		parentID = &parsedParentID
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file"})
	}
	defer src.Close()

	createdFile, err := h.fileService.UploadFile(context.Background(), userID, parentID, fileHeader.Filename, src, fileHeader.Size)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoActiveStorage), errors.Is(err, services.ErrInsufficientDiskSpace):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "file name cannot be empty", err.Error() == "invalid user id":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upload file"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         createdFile.ID,
		"name":       createdFile.Name,
		"parent_id":  createdFile.ParentID,
		"type":       createdFile.Type,
		"size":       createdFile.Size,
		"storage_id": createdFile.StorageID,
	})
}

func (h *FileHandler) DownloadFile(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	fileIDParam := strings.TrimSpace(c.Params("fileID"))
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file id"})
	}

	filePath, fileName, err := h.fileService.DownloadFile(context.Background(), userID, fileID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrFileNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrNodeIsNotAFile):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrStorageMissing), errors.Is(err, services.ErrFileMissingOnDisk):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to download file"})
		}
	}

	c.Attachment(fileName)
	return c.SendFile(filePath)
}

func (h *FileHandler) OnlyOfficeSave(c fiber.Ctx) error {
	fileIDParam := strings.TrimSpace(c.Params("id"))
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file id"})
	}

	err = h.fileService.SaveFromOnlyOfficeCallback(context.Background(), fileID, c.Body())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": 0})
}

func (h *FileHandler) GetEditorConfig(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	fileIDParam := strings.TrimSpace(c.Params("id"))
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file id"})
	}

	config, err := h.fileService.GetEditorConfig(context.Background(), userID, fileID, c.BaseURL())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrFileNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrNodeIsNotAFile):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "invalid base url":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate editor config"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(config)
}

func (h *FileHandler) OnlyOfficeDownload(c fiber.Ctx) error {
	fileIDParam := strings.TrimSpace(c.Params("id"))
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file id"})
	}

	node, err := h.fileService.GetNodeByID(context.Background(), fileID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get file"})
	}
	if node == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
	}

	if !strings.EqualFold(node.Type, "file") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "node is not a file"})
	}

	filePath, fileName, err := h.fileService.ResolveFilePath(context.Background(), fileID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrStorageMissing), errors.Is(err, services.ErrFileMissingOnDisk):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to resolve file"})
		}
	}
	// // ✅ CORS FIX
	// c.Set("Access-Control-Allow-Origin", "*")
	// c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	// c.Set("Access-Control-Allow-Headers", "*")
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	if ext != "" {
		c.Type(ext)
	}
	c.Set("Content-Disposition", "inline; filename=\""+fileName+"\"")

	return c.SendFile(filePath)
}

func (h *FileHandler) authenticatedUserID(c fiber.Ctx) (uuid.UUID, error) {
	username, ok := c.Locals("username").(string)
	if !ok || strings.TrimSpace(username) == "" {
		return uuid.Nil, errors.New("unauthorized")
	}

	user, err := h.userAuthService.FindUserByUsername(username)
	if err != nil || user == nil {
		return uuid.Nil, errors.New("unauthorized")
	}

	return user.UserID, nil
}
