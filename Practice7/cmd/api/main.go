package main

import (
	"Practice7/internal/app"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:12345@localhost:5432/practice7?sslmode=disable"
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8090"
	}

	application, err := app.New(databaseURL)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	defer application.Close()

	if err := application.Run(port); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
