package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type TodoTitle struct {
	value string
}

func NewTodoTitle(value string) (TodoTitle, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return TodoTitle{}, errors.New(errors.REASON_INVALID_PARAMETER, "Title cannot be empty")
	}
	return TodoTitle{value: value}, nil
}

func (t TodoTitle) Value() string {
	return t.value
}