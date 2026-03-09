package users

import (
	"database/sql"
	"errors"
	"fmt"
	"Practice5/internal/repository/_postgres"
	"Practice5/pkg/modules"
	"strings"
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

func (r *Repository) GetPaginatedUsers(
	page int,
	pageSize int,
	filters map[string]interface{},
	orderBy string,
) (modules.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	allowedColumns := []string{"id", "name", "email", "gender", "birth_date", "created_at"}
	validOrderBy := false
	for _, col := range allowedColumns {
		if orderBy == col {
			validOrderBy = true
			break
		}
	}
	if !validOrderBy {
		orderBy = "id"
	}

	whereConditions := []string{}
	args := []interface{}{}
	argCount := 1

	if id, ok := filters["id"]; ok {
		whereConditions = append(whereConditions, fmt.Sprintf("id = $%d", argCount))
		args = append(args, id)
		argCount++
	}

	if name, ok := filters["name"]; ok {
		whereConditions = append(whereConditions, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+fmt.Sprintf("%v", name)+"%")
		argCount++
	}

	if email, ok := filters["email"]; ok {
		whereConditions = append(whereConditions, fmt.Sprintf("email ILIKE $%d", argCount))
		args = append(args, "%"+fmt.Sprintf("%v", email)+"%")
		argCount++
	}

	if gender, ok := filters["gender"]; ok {
		whereConditions = append(whereConditions, fmt.Sprintf("gender = $%d", argCount))
		args = append(args, gender)
		argCount++
	}

	if birthDate, ok := filters["birth_date"]; ok {
		whereConditions = append(whereConditions, fmt.Sprintf("birth_date = $%d", argCount))
		args = append(args, birthDate)
		argCount++
	}

	whereClause := "1=1"
	if len(whereConditions) > 0 {
		whereClause = strings.Join(whereConditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var totalCount int
	err := r.db.DB.Get(&totalCount, countQuery, args...)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("failed to count users: %w", err)
	}

	args = append(args, pageSize)
	args = append(args, offset)

	query := fmt.Sprintf(
		"SELECT id, name, email, gender, birth_date, created_at FROM users WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d",
		whereClause,
		orderBy,
		argCount,
		argCount+1,
	)

	var users []modules.User
	err = r.db.DB.Select(&users, query, args...)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("failed to fetch users: %w", err)
	}

	return modules.PaginatedResponse{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) GetCommonFriends(userID1 int, userID2 int) ([]modules.User, error) {
	if userID1 == userID2 {
		return nil, errors.New("cannot get common friends of the same user")
	}

	query := `
		SELECT DISTINCT u.id, u.name, u.email, u.gender, u.birth_date, u.created_at
		FROM users u
		WHERE u.id IN (
			-- Друзья первого пользователя
			SELECT friend_id FROM user_friends WHERE user_id = $1
			INTERSECT
			-- Друзья второго пользователя
			SELECT friend_id FROM user_friends WHERE user_id = $2
		)
		ORDER BY u.id
	`

	var commonFriends []modules.User
	err := r.db.DB.Select(&commonFriends, query, userID1, userID2)
	if err != nil {
		if err == sql.ErrNoRows {
			return []modules.User{}, nil
		}
		return nil, fmt.Errorf("failed to fetch common friends: %w", err)
	}

	return commonFriends, nil
}

func (r *Repository) GetFriendsOfUser(userID int) ([]modules.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.gender, u.birth_date, u.created_at
		FROM users u
		INNER JOIN user_friends uf ON u.id = uf.friend_id
		WHERE uf.user_id = $1
		ORDER BY u.id
	`

	var friends []modules.User
	err := r.db.DB.Select(&friends, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []modules.User{}, nil
		}
		return nil, fmt.Errorf("failed to fetch friends: %w", err)
	}

	return friends, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	err := r.db.DB.Get(&user, "SELECT id, name, email, gender, birth_date, created_at FROM users WHERE id = $1", id)
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
		"INSERT INTO users (name, email, gender, birth_date, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
		req.Name, req.Email, req.Gender, req.BirthDate, time.Now(),
	).Scan(&id, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &modules.User{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Gender:    req.Gender,
		BirthDate: req.BirthDate,
		CreatedAt: createdAt,
	}, nil
}

func (r *Repository) AddFriend(userID int, friendID int) error {
	if userID == friendID {
		return errors.New("cannot be friend with yourself")
	}

	var user1Exists, user2Exists bool
	err := r.db.DB.Get(&user1Exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID)
	if err != nil || !user1Exists {
		return fmt.Errorf("user %d not found", userID)
	}

	err = r.db.DB.Get(&user2Exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", friendID)
	if err != nil || !user2Exists {
		return fmt.Errorf("user %d not found", friendID)
	}

	_, err = r.db.DB.Exec(
		"INSERT INTO user_friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, friendID,
	)
	if err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	_, err = r.db.DB.Exec(
		"INSERT INTO user_friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		friendID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	return nil
}
