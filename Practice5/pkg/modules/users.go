package modules

import "time"

type User struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Gender    string    `db:"gender" json:"gender"`
	BirthDate *time.Time `db:"birth_date" json:"birth_date"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Gender    string    `json:"gender"`
	BirthDate *time.Time `json:"birth_date"`
}

type UpdateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Gender    string    `json:"gender"`
	BirthDate *time.Time `json:"birth_date"`
}

type PaginatedResponse struct {
	Data      []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}

type Friendship struct {
	UserID   int `json:"user_id"`
	FriendID int `json:"friend_id"`
}