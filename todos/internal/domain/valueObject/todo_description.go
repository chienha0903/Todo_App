package valueObject

import (
	"strings"

	"github.com/chienha0903/Todo_App/internal/domain/errors"
)

type TodoDescription struct {
	value string `json:"value"`
}

func NewTodoDescription(value string) (TodoDescription, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return TodoDescription{}, errors.New(errors.REASON_INVALID_PARAMETER, "Description cannot be empty")
	}
	return TodoDescription{value: value}, nil
}

func (d TodoDescription) Value() string {
	return d.value
}