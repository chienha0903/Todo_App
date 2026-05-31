package valueobject

import "github.com/chienha0903/Todo_App/pkg/errors"

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusDeleting UserStatus = "DELETING"
	UserStatusDeleted  UserStatus = "DELETED"
)

func (s UserStatus) String() string   { return string(s) }
func (s UserStatus) IsActive() bool   { return s == UserStatusActive }
func (s UserStatus) IsDeleting() bool { return s == UserStatusDeleting }
func (s UserStatus) IsDeleted() bool  { return s == UserStatusDeleted }

func NewUserStatus(value string) (UserStatus, error) {
	switch UserStatus(value) {
	case UserStatusActive, UserStatusDeleting, UserStatusDeleted:
		return UserStatus(value), nil
	default:
		return "", errors.NewInvalidParameter("status must be ACTIVE, DELETING or DELETED")
	}
}
