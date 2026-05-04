package output

import "time"

type Todo struct {
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

type TodoGetter = Todo

type TodoLister []Todo

type TodoCreater = Todo

type TodoUpdater = Todo

type TodoDeleter struct {
	ID int64
}
