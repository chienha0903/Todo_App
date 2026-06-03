package input

type CreateUserInput struct {
	Email    string
	Username string
	Password string
	Role     string
}

type GetUserInput struct {
	ID int64
}

type GetUsersByIDsInput struct {
	IDs []int64
}

type ListUsersInput struct {
	Page     int32
	PageSize int32
}

type UpdateUserInput struct {
	ID       int64
	Email    string
	Username string
	Password string
	Role     string
}

type DeleteUserInput struct {
	ID int64
}

type UserLoginInput struct {
	Email    string
	Password string
}

type UserRefreshTokenInput struct {
	RefreshToken string
}

type ChangePasswordInput struct {
	UserID          int64
	CurrentPassword string
	NewPassword     string
}
