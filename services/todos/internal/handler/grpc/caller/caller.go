package caller

import "context"

type key struct{}

type Caller struct {
	UserID int64
	Role   string
}

func FromContext(ctx context.Context) (Caller, bool) {
	c, ok := ctx.Value(key{}).(Caller)
	return c, ok
}

func WithContext(ctx context.Context, userID int64, role string) context.Context {
	return context.WithValue(ctx, key{}, Caller{UserID: userID, Role: role})
}
