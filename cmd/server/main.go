package main

import (
	database "homeserver/configs"
	"homeserver/internals/models"
	"homeserver/internals/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/joho/godotenv"

	"io/fs"
	web "homeserver/web"
)

func main() {
	paths := []string{".env", "cmd/server/.env"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				log.Printf("loaded env from %s", p)
				break
			}
		}
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

	// Initialize FTS5 virtual tables and triggers after migrations
	if err := database.InitializeFTS5(db); err != nil {
		log.Fatalf("Could not initialize FTS5: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:3000"},
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
	
	frontendApp := fiber.New();

	frontendFS, err := fs.Sub(web.Files, "dist")
	if err != nil {
		log.Fatal(err)
	}

	frontendApp.Use("/", static.New("", static.Config{
		FS: frontendFS,
		Browse: false,
	}))

	// SPA fallback
	frontendApp.Use(func(c fiber.Ctx) error {
		index, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			return err
		}

		c.Type("html")
		return c.Send(index)
	})

	// running the frontend and the backend servers concurrently
	go func() {
		log.Println("frontend running on :3000")

		if err := frontendApp.Listen(":3000"); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("backend running on :%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
