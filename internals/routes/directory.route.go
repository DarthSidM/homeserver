package routes

import (
	"homeserver/internals/handlers"
	"homeserver/internals/middlerwares"
	"homeserver/internals/repos"
	"homeserver/internals/services"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupDirectoryRoutes(app *fiber.App, db *gorm.DB) {
	userRepo := repos.NewUserRepository(db)
	userService := services.NewUserAuthService(userRepo)
	directoryRepo := repos.NewDirectoryRepository(db)
	directoryService := services.NewDirectoryService(directoryRepo)
	directoryHandler := handlers.NewDirectoryHandler(directoryService, userService)

	directories := app.Group("/directories", middlerwares.AuthMiddleware())
	directories.Post("/", directoryHandler.CreateDirectory)
}
