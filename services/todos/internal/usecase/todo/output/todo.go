package output

import "time"

type TodoOutput struct {
	ID          int64
	UserID      int64
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
