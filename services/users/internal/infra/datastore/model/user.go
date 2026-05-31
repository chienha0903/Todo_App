package model

import "time"

type User struct {
	UserID    int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Username  string     `gorm:"column:username;not null"`
	Email     string     `gorm:"column:email;not null;unique"`
	Password  string     `gorm:"column:password_hash;not null"`
	Role      string     `gorm:"column:role;not null;default:USER"`
	Status    string     `gorm:"column:status;not null;default:ACTIVE"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (User) TableName() string {
	return "users"
}
