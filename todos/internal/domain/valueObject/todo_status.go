package valueObject

import (
	"strings"

	"github.com/chienha0903/Todo_App/internal/domain/errors"
)

type TodoStatus string

const (
	TODO_STATUS_PENDING TodoStatus = "PENDING"
	TODO_STATUS_IN_PROGRESS TodoStatus = "IN_PROGRESS"
	TODO_STATUS_COMPLETED TodoStatus = "COMPLETED"
)

func (s TodoStatus) String() string {
	return string(s)
}

func NewTodoStatus(value string) (TodoStatus, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return "", errors.New(errors.REASON_INVALID_PARAMETER, "Status cannot be empty")
	}

	status := TodoStatus(normalized)
	switch status {
	case TODO_STATUS_PENDING, TODO_STATUS_IN_PROGRESS, TODO_STATUS_COMPLETED:
		return status, nil
	default:
		return "", errors.New(errors.REASON_INVALID_PARAMETER, "Status is invalid")
	}
}
