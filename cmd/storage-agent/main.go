package main

import (
	database "homeserver/configs"
	"homeserver/storage"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting storage-agent...")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded; using environment variables")
	}

	db, _ := database.Connect()

	agent := storage.Agent{
		DB:          db,
		StorageRoot: "/storage",
	}

	agent.Run()
}
