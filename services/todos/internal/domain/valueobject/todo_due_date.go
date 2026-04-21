package valueobject

import (
	"time"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type TodoDueDate struct {
	value time.Time
}

func NewTodoDueDate(value time.Time) (TodoDueDate, error) {
	if value.IsZero() {
		return TodoDueDate{}, errors.New(errors.REASON_INVALID_PARAMETER, "Due date cannot be empty")
	}
	return TodoDueDate{value: value}, nil
}

func (d TodoDueDate) Value() time.Time {
	return d.value
}