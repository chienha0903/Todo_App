package server

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/generated"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/resolver"
)

func NewHTTPServer(cfg *config.Config, gqlResolver *resolver.Resolver) *nethttp.Server {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: gqlResolver}))
	srv.SetErrorPresenter(resolver.ErrorPresenter)

	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("/graphql", srv)

	return &nethttp.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func healthHandler(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
