package valueObject

import (
	"strings"

	"github.com/chienha0903/Todo_App/internal/domain/errors"
)

type TodoTitle struct {
	value string `json:"value"`
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