package output

type User struct {
	ID           int64
	Email        string
	Username     string
	Password string
	Role         string
	CreatedAt    int64
	UpdatedAt    int64
}

type UserGetterOutput = User

type UserPage struct {
	Items    []User
	Total    int32
	Page     int32
	PageSize int32
}

type UserCreaterOutput = User

type UserUpdaterOutput = User

type UserDeleterOutput struct {
	ID int64
}
