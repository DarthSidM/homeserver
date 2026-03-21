package handlers

import (
	"context"
	"errors"
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
		switch err.Error() {
		case "file name cannot be empty", "invalid user id":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upload file"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        createdFile.ID,
		"name":      createdFile.Name,
		"parent_id": createdFile.ParentID,
		"type":      createdFile.Type,
		"size":      createdFile.Size,
	})
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
