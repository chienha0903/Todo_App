package output

type Todo struct {
	ID          int64
	UserID      int64
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     string
	CreatedAt   string
	UpdatedAt   string
}

type TodoPage struct {
	Items    []*Todo
	Total    int
	Page     int
	PageSize int
	HasNext  bool
}
