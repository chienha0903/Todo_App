package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type Username struct {
	value string
}

func NewUsername(value string) (Username, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Username{}, errors.NewInvalidParameter("Username cannot be empty")
	}
	return Username{value: value}, nil
}

func (u Username) Value() string {
	return u.value
}
