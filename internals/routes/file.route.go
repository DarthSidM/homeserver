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
	nodeRepo := repos.NewNodeRepository(db)
	fileService := services.NewFileService(fileRepo, storageRepo, nodeRepo)
	fileHandler := handlers.NewFileHandler(fileService, userService)

	files := app.Group("/files", middlerwares.AuthMiddleware())
	files.Post("/upload", fileHandler.UploadFile)
	files.Post("/upload/:parentID", fileHandler.UploadFile)
	files.Get("/download/:fileID", fileHandler.DownloadFile)
	files.Get("/:id/editor-config", fileHandler.GetEditorConfig)

	app.Get("/onlyoffice/download/:id", fileHandler.OnlyOfficeDownload)
	app.Post("/onlyoffice/save/:id", fileHandler.OnlyOfficeSave)
}
