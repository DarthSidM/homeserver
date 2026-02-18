package main

import (
	database "homeserver/configs"
	"homeserver/internals/routes"
	"homeserver/internals/models"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("could not load .env file")
	}

	port := os.Getenv("PORT")
	log.Printf("running on port: %s", port)

	db, err := database.Connect()
	sqldb, err := db.DB()

	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	defer sqldb.Close()

	if err := db.AutoMigrate(&models.User{}, &models.Storage{}, &models.Node{}); err != nil {
    	log.Fatalf("Could not run migrations: %v", err)
	}


	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hello from homesever")
	})

	routes.SetupUserAuthRoutes(app, db)

	log.Println("server is running")
	app.Listen(":" + port)
}
