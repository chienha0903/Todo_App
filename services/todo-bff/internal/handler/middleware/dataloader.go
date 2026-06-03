package middleware

import (
	"context"
	"net/http"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/loader"
)

type loadersCtxKeyType struct{}

var loadersCtxKey = loadersCtxKeyType{}

// DataLoaderMiddleware injects a fresh set of DataLoaders into each request's context.
// A new Loaders instance is created per request so caches never leak across requests.
func DataLoaderMiddleware(gw gateway.UserGateway) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := loader.NewLoaders(gw)
			ctx := context.WithValue(r.Context(), loadersCtxKey, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLoaders retrieves the DataLoaders from the context.
// Returns nil if the middleware was not applied.
func GetLoaders(ctx context.Context) *loader.Loaders {
	l, _ := ctx.Value(loadersCtxKey).(*loader.Loaders)
	return l
}
