APP_TODOS = todos
APP_BFF   = bff
BIN_DIR   = bin

.PHONY: run-todos run-bff build build-todos build-bff proto wire tidy fmt vet \
        docker-up docker-down docker-logs

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

## Re-generate Wire DI code
## Cần: go install github.com/google/wire/cmd/wire@latest
wire:
	wire gen ./services/todos/internal/di/
	wire gen ./services/todo-bff/internal/di/

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

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
