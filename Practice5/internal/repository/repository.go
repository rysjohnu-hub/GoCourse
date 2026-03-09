package repository

import (
	"Practice5/internal/repository/_postgres"
	"Practice5/internal/repository/_postgres/users"
	"Practice5/pkg/modules"
)

type UserRepository interface {
	GetPaginatedUsers(page int, pageSize int, filters map[string]interface{}, orderBy string) (modules.PaginatedResponse, error)
	GetCommonFriends(userID1 int, userID2 int) ([]modules.User, error)
	GetFriendsOfUser(userID int) ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(req modules.CreateUserRequest) (*modules.User, error)
	AddFriend(userID int, friendID int) error
}

type Repositories struct {
	UserRepository
}

func NewRepositoriesWithDialect(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}
