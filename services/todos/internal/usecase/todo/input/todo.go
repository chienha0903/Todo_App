package input

import "time"

type CreateTodoInput struct {
	UserID      int64
	Title       string
	Description string
	Priority    string
	DueDate     *time.Time
}

type GetTodoInput struct {
	ID int64
}

type ListTodosInput struct {
	UserID int64
}

type UpdateTodoInput struct {
	ID          int64
	Title       string
	Description string
	Priority    string
	Status      string
	DueDate     *time.Time
}

type DeleteTodoInput struct {
	ID int64
}
