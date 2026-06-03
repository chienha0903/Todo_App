package loader

import (
	"context"
	"time"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	userinput "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/input"
	useroutput "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/output"
)

type Loaders struct {
	UserByID *dataloader.Loader[int64, *useroutput.User]
}

func NewLoaders(gw gateway.UserGateway) *Loaders {
	return &Loaders{
		UserByID: dataloader.NewBatchedLoader(
			newUserBatchFn(gw),
			dataloader.WithWait[int64, *useroutput.User](2*time.Millisecond),
			dataloader.WithBatchCapacity[int64, *useroutput.User](100),
		),
	}
}

func newUserBatchFn(gw gateway.UserGateway) dataloader.BatchFunc[int64, *useroutput.User] {
	return func(ctx context.Context, keys []int64) []*dataloader.Result[*useroutput.User] {
		results := make([]*dataloader.Result[*useroutput.User], len(keys))

		users, err := gw.GetUsersByIDs(ctx, &userinput.GetUsersByIDs{IDs: keys})
		if err != nil {
			for i := range keys {
				results[i] = &dataloader.Result[*useroutput.User]{Error: err}
			}
			return results
		}

		userMap := make(map[int64]*useroutput.User, len(users))
		for _, u := range users {
			userMap[u.ID] = u
		}

		for i, id := range keys {
			results[i] = &dataloader.Result[*useroutput.User]{Data: userMap[id]}
		}

		return results
	}
}
