package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewTodoCreater,
	NewTodoGetter,
	NewTodoLister,
	NewTodoUpdater,
	NewTodoDeleter,
)
