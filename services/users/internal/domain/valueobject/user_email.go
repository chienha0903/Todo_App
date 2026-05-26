package valueobject

import (
	"regexp"
	"strings"

	"github.com/chienha0903/Todo_App/pkg/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return Email{}, errors.NewInvalidParameter("email cannot be empty")
	}

	if !emailRegex.MatchString(value) {
		return Email{}, errors.NewInvalidParameter("invalid email format")
	}

	return Email{value: value}, nil
}

func (e Email) Value() string {
	return e.value
}
