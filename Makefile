APP_TODOS      = todos
APP_USERS      = users
APP_BFF        = bff
BIN_DIR        = bin
IMAGE_PREFIX  ?= ghcr.io/chienha0903
VERSION       ?= dev
LOCALBIN      ?= $(shell pwd)/bin

GOLANGCI_LINT_VERSION ?= v1.61.0
GOLANGCI_LINT 		  ?= $(LOCALBIN)/golangci-lint
# DB_DSN        ?= postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable
DB_DSN_TODOS  ?= postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable&x-migrations-table=schema_migrations_todos
DB_DSN_USERS  ?= postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable&x-migrations-table=schema_migrations_users
MIGRATIONS_DIR = services/todos/internal/infra/datastore/migrations
MIGRATIONS_DIR_USERS = services/users/internal/infra/datastore/migrations

.PHONY: run-todos run-users run-bff \
        build build-todos build-users build-bff build-worker build-consumer \
        proto mock wire generate \
        tidy fmt fmt-check lint lint-fix vet test ci \
        docker-up docker-down docker-logs docker-build \
        migrate-todos-up migrate-todos-down migrate-todos-version migrate-todos-force migrate-todos-new \
        migrate-users-up migrate-users-down migrate-users-version migrate-users-force migrate-users-new

## Chạy gRPC todos service
run-todos:
	go run ./services/todos/cmd/main.go

## Chạy gRPC users service
run-users:
	go run ./services/users/cmd/main.go

## Chạy BFF HTTP server
run-bff:
	go run ./services/todo-bff/cmd/main.go

## Build tất cả binary (bao gồm worker và consumer)
build: build-todos build-users build-bff build-worker build-consumer

build-todos:
	go build -o $(BIN_DIR)/$(APP_TODOS) ./services/todos/cmd/main.go

build-users:
	go build -o $(BIN_DIR)/$(APP_USERS) ./services/users/cmd/main.go

build-bff:
	go build -o $(BIN_DIR)/$(APP_BFF) ./services/todo-bff/cmd/main.go

build-worker:
	go build -o $(BIN_DIR)/users-worker ./services/users/cmd/worker/main.go

build-consumer:
	go build -o $(BIN_DIR)/todos-consumer ./services/todos/cmd/consumer/main.go

## Generate protobuf Go code
## Cần: brew install protobuf
##       go install google.golang.org/protobuf/cmd/protoc-gens-go@latest
##       go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/todo/todo.proto \
		proto/user/user.proto

## Re-generate mock files
## Cần: go install go.uber.org/mock/mockgen@latest
mock:
	mockgen -source=services/todos/internal/domain/gateway/todo.go -destination=services/todos/internal/domain/gateway/mock/mock_todo.go -package=mock

## Re-generate Wire DI code
## Cần: go install github.com/google/wire/cmd/wire@latest
wire:
	wire gen ./services/todos/internal/di/
	wire gen ./services/users/internal/di/
	wire gen ./services/todo-bff/internal/di/

## Re-generate GraphQL code (phải chạy từ services/todo-bff/)
generate:
	cd services/todo-bff && go run github.com/99designs/gqlgen generate

tidy:
	go mod tidy

fmt:
	gofmt -w .

## Kiểm tra format code – dùng cho CI (không sửa file, chỉ báo lỗi)
fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Các file sau chưa được format đúng chuẩn gofmt:"; \
		gofmt -l .; \
		echo ""; \
		echo "Chạy 'make fmt' để sửa tự động."; \
		exit 1; \
	fi
	@echo "gofmt: OK"

## Kiểm tra linting code – dùng cho CI (không sửa file, chỉ báo lỗi)
##lint: Chạy golangci-lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

##lint-fix: Tự động sửa lỗi linting (nếu golangci-lint hỗ trợ)
lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix ./...

vet:
	go vet ./...

## Chạy toàn bộ unit tests
test:
	go test ./...

## Chạy toàn bộ CI checks: format → vet → test → build
ci: fmt-check vet test build

## Migration todos (cần: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
migrate-todos-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN_TODOS)" up

migrate-todos-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN_TODOS)" down 1

migrate-todos-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN_TODOS)" version

migrate-todos-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN_TODOS)" force $(version)

## Dùng: make migrate-todos-new name=add_tags_table
migrate-todos-new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## Migration cho users service
migrate-users-up:
	migrate -path $(MIGRATIONS_DIR_USERS) -database "$(DB_DSN_USERS)" up

migrate-users-down:
	migrate -path $(MIGRATIONS_DIR_USERS) -database "$(DB_DSN_USERS)" down 1

migrate-users-version:
	migrate -path $(MIGRATIONS_DIR_USERS) -database "$(DB_DSN_USERS)" version

migrate-users-force:
	migrate -path $(MIGRATIONS_DIR_USERS) -database "$(DB_DSN_USERS)" force $(version)

## Dùng: make migrate-users-new name=add_something
migrate-users-new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR_USERS) -seq $(name)

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

## Build tất cả Docker images local (không push lên registry)
## Dùng: make docker-build VERSION=v1.0.0
docker-build:
	docker build -f services/todos/Dockerfile          -t $(IMAGE_PREFIX)/todos:$(VERSION) .
	docker build -f services/users/Dockerfile          -t $(IMAGE_PREFIX)/users:$(VERSION) .
	docker build -f services/todo-bff/Dockerfile       -t $(IMAGE_PREFIX)/todo-bff:$(VERSION) .
	docker build -f services/users/Dockerfile.worker   -t $(IMAGE_PREFIX)/users-worker:$(VERSION) .
	docker build -f services/todos/Dockerfile.consumer -t $(IMAGE_PREFIX)/todos-consumer:$(VERSION) .
