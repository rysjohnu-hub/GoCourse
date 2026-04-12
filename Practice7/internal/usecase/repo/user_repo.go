package repo

import (
	"fmt"

	"Practice7/internal/entity"
	"Practice7/pkg/postgres"
)

type UserRepository struct {
	db *postgres.Postgres
}

func NewUserRepository(db *postgres.Postgres) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) RegisterUser(user *entity.User) (*entity.User, error) {
	if err := r.db.Conn.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) LoginUser(username string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Conn.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(userID string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Conn.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) PromoteUser(userID string, newRole string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Conn.Model(&user).Where("id = ?", userID).Update("role", newRole).Error; err != nil {
		return nil, fmt.Errorf("failed to promote user: %w", err)
	}

	if err := r.db.Conn.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	return &user, nil
}
