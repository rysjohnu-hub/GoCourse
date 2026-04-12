package app

import (
	v1 "Practice7/internal/controller/http/v1"
	"Practice7/internal/entity"
	"Practice7/internal/usecase"
	"Practice7/internal/usecase/repo"
	"Practice7/pkg/postgres"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type App struct {
	db     *postgres.Postgres
	engine *gin.Engine
}

func New(databaseURL string) (*App, error) {
	db, err := postgres.New(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Conn.AutoMigrate(&entity.User{}); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database connected successfully")

	engine := gin.Default()

	userRepo := repo.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)

	v1.NewRouter(engine, userUseCase)

	return &App{
		db:     db,
		engine: engine,
	}, nil
}

func (a *App) Run(port string) error {
	log.Printf("🚀 Starting server on port %s", port)
	return a.engine.Run(":" + port)
}

func (a *App) Close() error {
	return a.db.Close()
}
