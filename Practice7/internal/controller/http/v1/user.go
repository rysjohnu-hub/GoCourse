package v1

import (
	"net/http"

	"Practice7/internal/entity"
	"Practice7/internal/usecase"
	"Practice7/utils"

	"github.com/gin-gonic/gin"
)

type userRoutes struct {
	useCase *usecase.UserUseCase
}

func NewUserRoutes(handler *gin.RouterGroup, uc *usecase.UserUseCase) {
	r := &userRoutes{useCase: uc}

	h := handler.Group("/users")
	{
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)

		protected := h.Group("/")
		protected.Use(utils.JWTAuthMiddleware())
		protected.Use(utils.RateLimiterMiddleware())
		{
			protected.GET("/me", r.GetMe)
			protected.PATCH("/promote/:id", utils.RoleMiddleware("admin"), r.PromoteUser)
		}
	}
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
	var input entity.CreateUserDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &entity.User{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
		Role:     "user",
	}

	if input.Role != "" {
		user.Role = input.Role
	}

	createdUser, sessionID, err := r.useCase.RegisterUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "User registered successfully",
		"session_id": sessionID,
		"user": gin.H{
			"id":       createdUser.ID,
			"username": createdUser.Username,
			"email":    createdUser.Email,
			"role":     createdUser.Role,
			"verified": createdUser.Verified,
		},
	})
}

func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := r.useCase.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

func (r *userRoutes) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	user, err := r.useCase.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"verified": user.Verified,
	})
}

func (r *userRoutes) PromoteUser(c *gin.Context) {
	userID := c.Param("id")

	var input entity.PromoteUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := r.useCase.PromoteUser(userID, input.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User promoted successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"verified": user.Verified,
		},
	})
}
