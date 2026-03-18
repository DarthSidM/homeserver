package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"homeserver/internals/dtos"
	"homeserver/internals/services"
)

type DirectoryHandler struct {
	directoryService services.DirectoryService
	userAuthService  *services.UserAuthService
}

func NewDirectoryHandler(directoryService services.DirectoryService, userAuthService *services.UserAuthService) *DirectoryHandler {
	return &DirectoryHandler{
		directoryService: directoryService,
		userAuthService:  userAuthService,
	}
}

func (h *DirectoryHandler) CreateDirectory(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req dtos.CreateDirectoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	directory, err := h.directoryService.CreateDir(context.Background(), userID, req.Name, req.ParentID)
	if err != nil {
		switch err.Error() {
		case "name cannot be empty", "invalid user id":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create directory"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(dtos.CreateDirectoryResponse{
		ID:        directory.ID,
		Name:      directory.Name,
		ParentID:  directory.ParentID,
		Type:      directory.Type,
		CreatedAt: directory.CreatedAt.Unix(),
	})
}

func (h *DirectoryHandler) authenticatedUserID(c fiber.Ctx) (uuid.UUID, error) {
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
