package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type UserRole string

const (
	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"
)

func (r UserRole) String() string {
	return string(r)
}

func NewUserRole(value string) (UserRole, error) {
	value = strings.ToUpper(strings.TrimSpace(value))

	if value == "" {
		return "", errors.NewInvalidParameter("role cannot be empty")
	}

	role := UserRole(value)
	switch role {
	case UserRoleUser, UserRoleAdmin:
		return role, nil
	default:
		return "", errors.NewInvalidParameter("role must be USER or ADMIN")
	}
}
