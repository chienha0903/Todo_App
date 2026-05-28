package output

type User struct {
	ID        int64
	Email     string
	Username  string
	Role      string
	CreatedAt string
	UpdatedAt string
}

type UserPage struct {
	Items    []*User
	Total    int32
	Page     int32
	PageSize int32
}
