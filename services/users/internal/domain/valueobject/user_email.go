package valueobject

import (
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Email{}, errors.NewInvalidParameter("Email cannot be empty")
	}
	// You can add more complex email validation here if needed
	return Email{value: value}, nil
}

func (e Email) Value() string {
	return e.value
}
