package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type UserRole struct {
	value string
}

func NewUserRole(value string) (UserRole, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return UserRole{}, errors.NewInvalidParameter("User role cannot be empty")
	}
	// You can add more complex role validation here if needed
	return UserRole{value: value}, nil
}

func (r UserRole) Value() string {
	return r.value
}
