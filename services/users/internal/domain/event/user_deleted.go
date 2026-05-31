package event

import (
	"encoding/json"
	"time"
)

const EventTypeUserDeleteRequested = "UserDeleteRequested"

type UserDeleteRequestedPayload struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	DeletedAt time.Time `json:"deleted_at"`
}

func (p *UserDeleteRequestedPayload) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
