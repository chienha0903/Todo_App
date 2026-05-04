package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type TodoPriority string

const (
	TODO_PRIORITY_LOW    TodoPriority = "LOW"
	TODO_PRIORITY_MEDIUM TodoPriority = "MEDIUM"
	TODO_PRIORITY_HIGH   TodoPriority = "HIGH"
)

func (p TodoPriority) String() string {
	return string(p)
}

func NewTodoPriority(value string) (TodoPriority, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", errors.NewAppError(errors.ReasonInvalidParameter, "Priority cannot be empty")
	}

	priority := TodoPriority(value)
	switch priority {
	case TODO_PRIORITY_LOW, TODO_PRIORITY_MEDIUM, TODO_PRIORITY_HIGH:
		return priority, nil
	default:
		return "", errors.NewAppError(errors.ReasonInvalidParameter, "Priority is invalid")
	}
}
