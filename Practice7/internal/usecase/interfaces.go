package usecase

import "Practice7/internal/entity"

type UserRepository interface {
	RegisterUser(user *entity.User) (*entity.User, error)
	LoginUser(username string) (*entity.User, error)
	GetUserByID(userID string) (*entity.User, error)
	PromoteUser(userID string, newRole string) (*entity.User, error)
}
