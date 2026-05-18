package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type TodoPriority string

const (
	TodoPriorityLow    TodoPriority = "LOW"
	TodoPriorityMedium TodoPriority = "MEDIUM"
	TodoPriorityHigh   TodoPriority = "HIGH"
)

func (p TodoPriority) String() string {
	return string(p)
}

func NewTodoPriority(value string) (TodoPriority, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", errors.NewInvalidParameter("Priority cannot be empty")
	}

	priority := TodoPriority(value)
	switch priority {
	case TodoPriorityLow, TodoPriorityMedium, TodoPriorityHigh:
		return priority, nil
	default:
		return "", errors.NewInvalidParameter("Priority is invalid")
	}
}
