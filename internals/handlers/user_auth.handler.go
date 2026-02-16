package handlers

import (
	"homeserver/internals/dtos"
	"homeserver/internals/services"
	
	"github.com/gofiber/fiber/v3"
)

type UserAuthHandler struct {
	service *services.UserAuthService
}

func NewUserAuthHandler(service *services.UserAuthService) *UserAuthHandler {
	return &UserAuthHandler{service: service}
}

func (h *UserAuthHandler) Signup(c fiber.Ctx) error {
	var req dtos.SignupRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create user with hashed password
	user, err := h.service.CreateUser(req.Username, req.Name, req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Issue JWT token
	token, err := h.service.IssueJWT(user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to issue token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dtos.SignupResponse{
		Token: token,
	})
}

func (h *UserAuthHandler) Login(c fiber.Ctx) error {
	var req dtos.LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Find user by username
	user, err := h.service.FindUserByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Check password
	if !h.service.CheckPassword(user.Password, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Issue JWT token
	token, err := h.service.IssueJWT(user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to issue token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dtos.LoginResponse{
		Token: token,
	})
}
