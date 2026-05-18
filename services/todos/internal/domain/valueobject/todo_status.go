package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "PENDING"
	TodoStatusInProgress TodoStatus = "IN_PROGRESS"
	TodoStatusCompleted  TodoStatus = "COMPLETED"
)

func (s TodoStatus) String() string {
	return string(s)
}

func NewTodoStatus(value string) (TodoStatus, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return "", errors.NewInvalidParameter("Status cannot be empty")
	}

	status := TodoStatus(normalized)
	switch status {
	case TodoStatusPending, TodoStatusInProgress, TodoStatusCompleted:
		return status, nil
	default:
		return "", errors.NewInvalidParameter("Status is invalid")
	}
}
