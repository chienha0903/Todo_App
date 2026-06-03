package input

type GetUser struct {
	ID int64
}

type GetUsersByIDs struct {
	IDs []int64
}

type ListUsers struct {
	Page     int32
	PageSize int32
}

type UpdateUser struct {
	ID       int64
	Email    string
	Username string
	Password string
	Role     string
}

type DeleteUser struct {
	ID int64
}

type RefreshToken struct {
	RefreshToken string
}

type ChangePassword struct {
	UserID          int64
	CurrentPassword string
	NewPassword     string
}
