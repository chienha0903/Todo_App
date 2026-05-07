package model

import "time"

type Todo struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64      `gorm:"column:user_id;not null"`
	Title       string     `gorm:"column:title;not null"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status;not null"`
	Priority    string     `gorm:"column:priority;not null"`
	DueDate     *time.Time `gorm:"column:due_date"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Todo) TableName() string {
	return "todos"
}
