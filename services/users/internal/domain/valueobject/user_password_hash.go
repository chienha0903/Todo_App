package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type PasswordHash struct {
	value string
}

func NewPasswordHash(value string) (PasswordHash, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return PasswordHash{}, errors.NewInvalidParameter("Password hash cannot be empty")
	}
	return PasswordHash{value: value}, nil
}

func (p PasswordHash) Value() string {
	return p.value
}
