package todo

import (
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

// IsOverdue is true when the todo has a due date, is not completed, and the due date is in the past.
func IsOverdue(t *entity.Todo) bool {
	if t == nil || t.DueDate == nil {
		return false
	}
	if t.Status == vo.TODO_STATUS_COMPLETED {
		return false
	}
	return time.Now().After(t.DueDate.Value())
}
