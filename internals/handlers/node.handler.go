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

type NodeHandler struct {
	nodeService     services.NodeService
	userAuthService *services.UserAuthService
}

func NewNodeHandler(nodeService services.NodeService, userAuthService *services.UserAuthService) *NodeHandler {
	return &NodeHandler{
		nodeService:     nodeService,
		userAuthService: userAuthService,
	}
}

func (h *NodeHandler) ListNodes(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var parentID *uuid.UUID
	parentIDParam := strings.TrimSpace(c.Query("parent_id"))
	if parentIDParam != "" {
		parsedParentID, err := uuid.Parse(parentIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid parent_id"})
		}
		parentID = &parsedParentID
	}

	nodes, err := h.nodeService.ListNodes(context.Background(), userID, parentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list nodes"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"nodes": nodes})
}

func (h *NodeHandler) RenameNode(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	nodeIDParam := strings.TrimSpace(c.Params("id"))
	nodeID, err := uuid.Parse(nodeIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid node id"})
	}

	var req dtos.RenameNodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	node, err := h.nodeService.RenameNode(context.Background(), userID, nodeID, req.NewName)
	if err != nil {
		switch {
		case err.Error() == "name cannot be empty":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "node not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case strings.Contains(err.Error(), "already exists"):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to rename node"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(dtos.RenameNodeResponse{
		ID:      node.ID.String(),
		Name:    node.Name,
		Message: "node renamed successfully",
	})
}

func (h *NodeHandler) DeleteNode(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	nodeIDParam := strings.TrimSpace(c.Params("id"))
	nodeID, err := uuid.Parse(nodeIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid node id"})
	}

	err = h.nodeService.DeleteNode(context.Background(), userID, nodeID)
	if err != nil {
		if err.Error() == "node not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete node"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "node deleted successfully"})
}

func (h *NodeHandler) MarkFavouriteNode(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	nodeIDParam := strings.TrimSpace(c.Query("node_id"))
	if nodeIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "node_id is required"})
	}

	nodeID, err := uuid.Parse(nodeIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid node_id"})
	}

	isMarked, err := h.nodeService.MarkFavouriteNode(context.Background(), userID, nodeID)
	if err != nil {
		switch err.Error() {
		case "node not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case "invalid user id":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to mark favourite node"})
		}
	}

	var message string
	if isMarked {
		message = "node marked as favourite"
	} else {
		message = "node removed from favourites"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": message})
}

func (h *NodeHandler) GetFavouriteNodes(c fiber.Ctx) error {
	userID, err := h.authenticatedUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	nodes, err := h.nodeService.GetFavouriteNodes(context.Background(), userID)
	if err != nil {
		if err.Error() == "invalid user id" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get favourite nodes"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"nodes": nodes})
}

func (h *NodeHandler) authenticatedUserID(c fiber.Ctx) (uuid.UUID, error) {
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
