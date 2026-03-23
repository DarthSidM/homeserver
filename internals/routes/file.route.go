package routes

import (
	"homeserver/internals/handlers"
	"homeserver/internals/middlerwares"
	"homeserver/internals/repos"
	"homeserver/internals/services"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupFileRoutes(app *fiber.App, db *gorm.DB) {
	userRepo := repos.NewUserRepository(db)
	userService := services.NewUserAuthService(userRepo)
	fileRepo := repos.NewFileRepository(db)
	storageRepo := repos.NewStorageRepository(db)
	fileService := services.NewFileService(fileRepo, storageRepo)
	fileHandler := handlers.NewFileHandler(fileService, userService)

	files := app.Group("/files", middlerwares.AuthMiddleware())
	files.Post("/upload", fileHandler.UploadFile)
	files.Post("/upload/:parentID", fileHandler.UploadFile)
}
