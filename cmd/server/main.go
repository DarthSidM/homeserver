package main

import (
	database "homeserver/configs"
	"homeserver/internals/models"
	"homeserver/internals/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env", "cmd/server/.env")
	if err != nil {
		log.Println("could not load .env file from known paths")
	}

	port := os.Getenv("PORT")
	log.Printf("running on port: %s", port)

	db, err := database.Connect()
	sqldb, err := db.DB()

	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	defer sqldb.Close()

	if err := db.AutoMigrate(&models.User{}, &models.Storage{}, &models.Node{}, &models.Favourite{}); err != nil {
		log.Fatalf("Could not run migrations: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hello from homesever")
	})

	routes.SetupUserAuthRoutes(app, db)
	routes.SetupNodeRoutes(app, db)
	routes.SetupDirectoryRoutes(app, db)
	routes.SetupFileRoutes(app, db)

	log.Println("server is running")
	app.Listen(":" + port)
}
