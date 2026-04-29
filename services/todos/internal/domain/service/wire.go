package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewTodoCreator,
	NewTodoGetter,
	NewTodoLister,
	NewTodoUpdater,
	NewTodoDeleter,
)
