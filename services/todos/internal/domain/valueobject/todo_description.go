package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type TodoDescription struct {
	value string
}

func NewTodoDescription(value string) (TodoDescription, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return TodoDescription{}, errors.NewAppError(errors.ReasonInvalidParameter, "Description cannot be empty")
	}
	return TodoDescription{value: value}, nil
}

func (d TodoDescription) Value() string {
	return d.value
}
