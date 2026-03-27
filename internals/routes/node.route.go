package routes

import (
	"homeserver/internals/handlers"
	"homeserver/internals/middlerwares"
	"homeserver/internals/repos"
	"homeserver/internals/services"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupNodeRoutes(app *fiber.App, db *gorm.DB) {
	userRepo := repos.NewUserRepository(db)
	userService := services.NewUserAuthService(userRepo)
	nodeRepo := repos.NewNodeRepository(db)
	nodeService := services.NewNodeService(nodeRepo)
	nodeHandler := handlers.NewNodeHandler(nodeService, userService)

	nodes := app.Group("/nodes", middlerwares.AuthMiddleware())
	nodes.Get("/", nodeHandler.ListNodes)
	nodes.Get("/search", nodeHandler.SearchNodes)
	nodes.Post("/favourites", nodeHandler.MarkFavouriteNode)
	nodes.Get("/favourites", nodeHandler.GetFavouriteNodes)
	nodes.Patch("/:id", nodeHandler.RenameNode)
	nodes.Delete("/:id", nodeHandler.DeleteNode)
}
