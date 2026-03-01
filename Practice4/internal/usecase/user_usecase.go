package usecase

import (
	"fmt"
	"Practice4/internal/repository"
	"Practice4/pkg/modules"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
}

func (u *UserUsecase) GetAllUsers() ([]modules.User, error) {
	users, err := u.repo.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("usecase: failed to get users: %w", err)
	}
	return users, nil
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

func (u *UserUsecase) UpdateUser(id int, req modules.UpdateUserRequest) (*modules.User, error) {
	user, err := u.repo.UpdateUser(id, req)
	if err != nil {
		return nil, fmt.Errorf("usecase: %w", err)
	}
	return user, nil
}

func (u *UserUsecase) DeleteUser(id int) (string, error) {
	rowsAffected, err := u.repo.DeleteUser(id)
	if err != nil {
		return "", fmt.Errorf("usecase: %w", err)
	}
	return fmt.Sprintf("User deleted successfully. Rows affected: %d", rowsAffected), nil
}