package entity

import "time"

type TodoID int64
type UserID int64
type TodolistID int64
type TodoStatus string
type TodoPriority string

const (
	TODO_STATUS_PENDING TodoStatus = "PENDING"
	TODO_STATUS_IN_PROGRESS TodoStatus = "IN_PROGRESS"
	TODO_STATUS_COMPLETED TodoStatus = "COMPLETED"
)

const (
	TODO_PRIORITY_LOW TodoPriority = "LOW"
	TODO_PRIORITY_MEDIUM TodoPriority = "MEDIUM"
	TODO_PRIORITY_HIGH TodoPriority = "HIGH"
)
type Todo struct {
	ID TodoID `json:"id"`
	UserID UserID `json:"user_id"`
	TodolistID TodolistID `json:"todolist_id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Status TodoStatus `json:"status"`
	Priority TodoPriority `json:"priority"`
	DueDate time.Time `json:"due_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Todo) IsOverdue() bool {
	if t.DueDate == nil || t.Status != TODO_STATUS_COMPLETED{
		return false
	}
	return time.Now().After(*t.DueDate)
}