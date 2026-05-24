package entity

import (
	"time"

	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
)

type UserID int64

type User struct {
	UserID       UserID      `json:"user_id"`
	Username     vo.Username `json:"username"`
	Email        vo.Email    `json:"email"`
	Password vo.Password `json:"password"`
	Role         vo.UserRole `json:"role"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
