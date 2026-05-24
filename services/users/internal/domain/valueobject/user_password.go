package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type Password struct {
	value string
}

func NewPassword(value string) (Password, error) {
	value = strings.TrimSpace(value)
	if len(value) < 8 {
		return Password{}, errors.New("password must be at least 8 characters long")
	}
	return Password{value: value}, nil
}

func (p Password) Value() string {
	return p.value
}
