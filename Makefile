APP_TODOS      = todos
APP_BFF        = bff
BIN_DIR        = bin
DB_DSN        ?= postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable
MIGRATIONS_DIR = services/todos/internal/infra/datastore/migrations

.PHONY: run-todos run-bff build build-todos build-bff proto mock wire generate tidy fmt vet \
        docker-up docker-down docker-logs \
        migrate-up migrate-down migrate-version migrate-force migrate-new

## Chạy gRPC todos service
run-todos:
	go run ./services/todos/cmd/main.go

## Chạy BFF HTTP server
run-bff:
	go run ./services/todo-bff/cmd/main.go

## Build tất cả
build: build-todos build-bff

build-todos:
	go build -o $(BIN_DIR)/$(APP_TODOS) ./services/todos/cmd/main.go

build-bff:
	go build -o $(BIN_DIR)/$(APP_BFF) ./services/todo-bff/cmd/main.go

## Generate protobuf Go code
## Cần: brew install protobuf
##       go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
##       go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/todo/todo.proto

## Re-generate mock files
## Cần: go install go.uber.org/mock/mockgen@latest
mock:
	mockgen -source=services/todos/internal/domain/gateway/todo.go -destination=services/todos/internal/domain/gateway/mock/mock_todo.go -package=mock

## Re-generate Wire DI code
## Cần: go install github.com/google/wire/cmd/wire@latest
wire:
	wire gen ./services/todos/internal/di/
	wire gen ./services/todo-bff/internal/di/

## Re-generate GraphQL code (phải chạy từ services/todo-bff/)
generate:
	cd services/todo-bff && go run github.com/99designs/gqlgen generate

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

## Migration (cần: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down 1

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" version

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" force $(version)

## Dùng: make migrate-new name=add_tags_table
migrate-new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## Docker
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-logs-todos:
	docker compose logs -f todos

docker-ps:
	docker compose ps
