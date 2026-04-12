package usecase

import (
	"Practice7/internal/entity"
	"Practice7/internal/usecase/repo"
	"Practice7/utils"
	"fmt"

	"github.com/google/uuid"
)

type UserUseCase struct {
	repo *repo.UserRepository
}

func NewUserUseCase(r *repo.UserRepository) *UserUseCase {
	return &UserUseCase{repo: r}
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if user.Role == "" {
		user.Role = "user"
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	registeredUser, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", err
	}

	sessionID := uuid.New().String()

	return registeredUser, sessionID, nil
}

func (u *UserUseCase) LoginUser(input *entity.LoginUserDTO) (string, error) {
	user, err := u.repo.LoginUser(input.Username)
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}

	if !utils.CheckPassword(user.Password, input.Password) {
		return "", fmt.Errorf("invalid password")
	}

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (u *UserUseCase) GetUserByID(userID string) (*entity.User, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (u *UserUseCase) PromoteUser(userID string, newRole string) (*entity.User, error) {
	user, err := u.repo.PromoteUser(userID, newRole)
	if err != nil {
		return nil, fmt.Errorf("failed to promote user: %w", err)
	}
	return user, nil
}
