package entity

import (
	"time"

	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

type TodoID int64
type UserID int64
type TodoListID int64

type Todo struct {
	ID          TodoID             `json:"id"`
	UserID      UserID             `json:"user_id"`
	TodoListID  TodoListID         `json:"todo_list_id"`
	Title       vo.TodoTitle       `json:"title"`
	Description vo.TodoDescription `json:"description"`
	Status      vo.TodoStatus      `json:"status"`
	Priority    vo.TodoPriority    `json:"priority"`
	DueDate     *vo.TodoDueDate    `json:"due_date,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func IsOverdue(t *Todo) bool {
	if t == nil || t.DueDate == nil {
		return false
	}
	if t.Status == vo.TODO_STATUS_COMPLETED {
		return false
	}
	return time.Now().After(t.DueDate.Value())
}
