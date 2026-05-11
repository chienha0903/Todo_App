package input

type CreateTodo struct {
	UserID      int64
	Title       string
	Description string
	Priority    string
	DueDate     string
}

type GetTodo struct {
	ID int64
}

type ListTodos struct {
	UserID int64
}

type UpdateTodo struct {
	ID          int64
	Title       string
	Description string
	Priority    string
	Status      string
	DueDate     string
}

type DeleteTodo struct {
	ID int64
}
