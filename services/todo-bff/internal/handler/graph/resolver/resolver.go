package resolver

import (
	"time"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
)

// This file will not be regenerated automatically.

type Resolver struct {
	todoClient todopb.TodoServiceClient
	timeout    time.Duration
}

func NewResolver(todoClient todopb.TodoServiceClient, timeout time.Duration) *Resolver {
	return &Resolver{todoClient: todoClient, timeout: timeout}
}
