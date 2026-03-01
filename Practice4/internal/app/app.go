package app

import (
	"context"
	"Practice4/internal/handler"
	"Practice4/internal/middleware"
	"Practice4/internal/repository"
	"Practice4/internal/repository/_postgres"
	"Practice4/internal/usecase"
	"Practice4/pkg/modules"
	"log"
	"net/http"
	"os"
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

	log.Printf("Connecting to database: %s:%s@%s:%s/%s",
		dbConfig.Username, "***", dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	a.db = _postgres.NewPGXDialect(a.ctx, dbConfig)
	defer a.db.Close()

	log.Println("✅ Database connected successfully")

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

	log.Println("🚀 Starting the Server on :8080")
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
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "mydb"
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	return &modules.PostgreConfig{
		Host:        host,
		Port:        port,
		Username:    user,
		Password:    password,
		DBName:      dbName,
		SSLMode:     sslMode,
		ExecTimeout: 5 * time.Second,
	}
}