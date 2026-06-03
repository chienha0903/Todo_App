package consumer

import (
	"encoding/json"
	"fmt"
)

type UserDeleteRequestedMessage struct {
	EventID   string                     `json:"event_id"`
	EventType string                     `json:"event_type"`
	Payload   UserDeleteRequestedPayload `json:"payload"`
}

type UserDeleteRequestedPayload struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func parseUserDeleteRequested(body []byte) (*UserDeleteRequestedMessage, error) {
	var msg UserDeleteRequestedMessage

	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if msg.Payload.UserID <= 0 {
		return nil, fmt.Errorf("invalid user_id: %d", msg.Payload.UserID)
	}

	return &msg, nil
}
