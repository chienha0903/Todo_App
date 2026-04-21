package input

import "time"

type CreateTodoInput struct {
	UserID      int64
	Title       string
	Description string
	Priority    string
	DueDate     *time.Time
}
