package v1

import (
	"Practice7/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(engine *gin.Engine, uc *usecase.UserUseCase) {
	apiV1 := engine.Group("/v1")
	{
		NewUserRoutes(apiV1, uc)
	}
}
