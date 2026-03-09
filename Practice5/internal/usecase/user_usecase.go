package usecase

import (
	"fmt"
	"Practice5/internal/repository"
	"Practice5/pkg/modules"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
}

func (u *UserUsecase) GetPaginatedUsers(
	page int,
	pageSize int,
	filters map[string]interface{},
	orderBy string,
) (modules.PaginatedResponse, error) {
	result, err := u.repo.GetPaginatedUsers(page, pageSize, filters, orderBy)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("usecase: %w", err)
	}
	return result, nil
}

func (u *UserUsecase) GetCommonFriends(userID1 int, userID2 int) ([]modules.User, error) {
	friends, err := u.repo.GetCommonFriends(userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("usecase: %w", err)
	}
	return friends, nil
}

func (u *UserUsecase) GetFriendsOfUser(userID int) ([]modules.User, error) {
	friends, err := u.repo.GetFriendsOfUser(userID)
	if err != nil {
		return nil, fmt.Errorf("usecase: %w", err)
	}
	return friends, nil
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	user, err := u.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("usecase: %w", err)
	}
	return user, nil
}

func (u *UserUsecase) CreateUser(req modules.CreateUserRequest) (*modules.User, error) {
	user, err := u.repo.CreateUser(req)
	if err != nil {
		return nil, fmt.Errorf("usecase: %w", err)
	}
	return user, nil
}

func (u *UserUsecase) AddFriend(userID int, friendID int) error {
	err := u.repo.AddFriend(userID, friendID)
	if err != nil {
		return fmt.Errorf("usecase: %w", err)
	}
	return nil
}