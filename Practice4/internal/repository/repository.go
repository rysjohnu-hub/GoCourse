package repository

import (
	"Practice4/internal/repository/_postgres"
	"Practice4/internal/repository/_postgres/users"
	"Practice4/pkg/modules"
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(req modules.CreateUserRequest) (*modules.User, error)
	UpdateUser(id int, req modules.UpdateUserRequest) (*modules.User, error)
	DeleteUser(id int) (int64, error)
}

type Repositories struct {
	UserRepository
}

func NewRepositoriesWithDialect(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}