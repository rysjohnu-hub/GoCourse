package main

import (
	"log"
	"net/http"
	"task-service/internal/handlers"
	"task-service/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", middleware.AuthMiddleware(handlers.TaskHandler))

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}