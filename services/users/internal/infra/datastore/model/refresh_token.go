package model

import "time"

type RefreshToken struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64      `gorm:"column:user_id;not null"`
	TokenHash string     `gorm:"column:token_hash;not null;unique"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
