package routes

import (
	"homeserver/internals/handlers"
	"homeserver/internals/repos"
	"homeserver/internals/services"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupUserAuthRoutes(app *fiber.App, db *gorm.DB) {
	userRepo := repos.NewUserRepository(db)
	userService := services.NewUserAuthService(userRepo)
	userHandler := handlers.NewUserAuthHandler(userService)

	auth := app.Group("/auth")
	auth.Post("/signup", userHandler.Signup)
	auth.Post("/login", userHandler.Login)
}
