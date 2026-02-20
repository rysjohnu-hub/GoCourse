package app

import (
	"context"
	"Practice3/internal/handler"
	"Practice3/internal/middleware"
	"Practice3/internal/repository"
	"Practice3/internal/repository/_postgres"
	"Practice3/internal/usecase"
	"Practice3/pkg/modules"
	"log"
	"net/http"
	"time"
)

type App struct {
	ctx    context.Context
	db     *_postgres.Dialect
	server *http.Server
}

func NewApp(ctx context.Context) *App {
	return &App{
		ctx: ctx,
	}
}

func (a *App) Run() {
	dbConfig := a.initPostgreConfig()

	a.db = _postgres.NewPGXDialect(a.ctx, dbConfig)
	defer a.db.Close()

	log.Println("Database connected successfully")

	repos := repository.NewRepositoriesWithDialect(a.db)

	userUsecase := usecase.NewUserUsecase(repos.UserRepository)

	userHandler := handler.NewUserHandler(userUsecase)

	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = middleware.LoggingMiddleware(handler)

	mux.HandleFunc("/health", userHandler.HealthCheck)

	mux.HandleFunc("/users", userHandler.GetAllUsers)
	mux.HandleFunc("POST /users", userHandler.CreateUser)

	mux.HandleFunc("GET /users/", a.withAuth(userHandler.GetUserByID))
	mux.HandleFunc("PUT /users/", a.withAuth(userHandler.UpdateUser))
	mux.HandleFunc("PATCH /users/", a.withAuth(userHandler.UpdateUser))
	mux.HandleFunc("DELETE /users/", a.withAuth(userHandler.DeleteUser))

	a.server = &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server starting on :8080")
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func (a *App) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func (a *App) initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "12345",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}