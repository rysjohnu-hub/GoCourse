package users

import (
	"database/sql"
	"errors"
	"fmt"
	"Practice4/internal/repository/_postgres"
	"Practice4/pkg/modules"
	"time"
)

type Repository struct {
	db              *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:              db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User
	err := r.db.DB.Select(&users, "SELECT id, name, email, age, city, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	return users, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	err := r.db.DB.Get(&user, "SELECT id, name, email, age, city, created_at FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return &user, nil
}

func (r *Repository) CreateUser(req modules.CreateUserRequest) (*modules.User, error) {
	if req.Name == "" || req.Email == "" {
		return nil, errors.New("name and email are required")
	}

	var id int
	var createdAt time.Time
	
	err := r.db.DB.QueryRow(
		"INSERT INTO users (name, email, age, city, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
		req.Name, req.Email, req.Age, req.City, time.Now(),
	).Scan(&id, &createdAt)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return nil, fmt.Errorf("email %s already exists", req.Email)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &modules.User{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		City:      req.City,
		CreatedAt: createdAt,
	}, nil
}

func (r *Repository) UpdateUser(id int, req modules.UpdateUserRequest) (*modules.User, error) {
	if req.Name == "" || req.Email == "" {
		return nil, errors.New("name and email are required")
	}

	result, err := r.db.DB.Exec(
		"UPDATE users SET name = $1, email = $2, age = $3, city = $4 WHERE id = $5",
		req.Name, req.Email, req.Age, req.City, id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}

	return r.GetUserByID(id)
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	result, err := r.db.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return 0, fmt.Errorf("user with ID %d not found", id)
	}

	return rowsAffected, nil
}