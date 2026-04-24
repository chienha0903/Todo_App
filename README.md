# Todo App

Project Todo App viết bằng Go, giao tiếp qua gRPC, lưu dữ liệu vào PostgreSQL, có sẵn Docker Compose để chạy local nhanh.

## Chức năng hiện có

Service `todos` hiện hỗ trợ đầy đủ CRUD:

- `CreateTodo`
- `GetTodo`
- `ListTodos`
- `UpdateTodo`
- `DeleteTodo`

Các rule nghiệp vụ chính:

- Validate dữ liệu bằng Value Object (`title`, `description`, `priority`, `status`, `due_date`)
- `CreateTodo` mặc định `status = PENDING`
- `UpdateTodo` chỉ update field nào được truyền vào (chuỗi rỗng xem như không update field đó)

---

## Kiến trúc tổng quan

Luồng request:

`gRPC Handler` -> `Mapper` -> `Domain Service` -> `Gateway` -> `PostgreSQL`

Các layer chính:

- **Handler (`internal/handler/grpc`)**  
  Nhận request gRPC, map input/output, trả về gRPC response.

- **Usecase contracts (`internal/usecase/todo`)**  
  Chứa interface contract input/output giữa handler và domain service.

- **Domain Service (`internal/domain/service`)**  
  Chứa logic nghiệp vụ: validate VO, áp quy tắc domain, gọi gateway.

- **Gateway (`internal/domain/gateway`)**  
  Interface cho command/query persistence.

- **Infra Datastore (`internal/infra/datastore`)**  
  Triển khai gateway bằng PostgreSQL (`pgxpool`).

- **DI (`internal/di`, Wire)**  
  Nối dependencies để tạo `*grpc.Server`.

---

## Cấu trúc thư mục quan trọng

```text
.
├── docker-compose.yml
├── proto/
│   └── todo/todo.proto
├── services/
│   └── todos/
│       ├── Dockerfile
│       ├── cmd/main.go
│       └── internal/
│           ├── config/
│           ├── di/
│           ├── domain/
│           │   ├── entity/
│           │   ├── gateway/
│           │   ├── service/
│           │   └── valueobject/
│           ├── handler/grpc/
│           ├── infra/datastore/
│           └── usecase/todo/
└── Makefile
```

---

## Yêu cầu môi trường

- Go (theo `go.mod`)
- Docker + Docker Compose
- `protoc` (nếu muốn regenerate protobuf)
- plugins:
  - `protoc-gen-go`
  - `protoc-gen-go-grpc`
- Wire (nếu muốn regenerate DI code)

Cài nhanh:

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/google/wire/cmd/wire@latest
```

---

## Cách chạy project

## 1) Chạy bằng Docker (khuyến nghị)

```bash
make docker-up
```

hoặc:

```bash
docker compose up -d --build
```

Kiểm tra:

```bash
make docker-ps
make docker-logs-todos
```

Stop:

```bash
make docker-down
```

Mặc định:

- gRPC server: `localhost:50051`
- Postgres: `localhost:5432`
- DB: `todo_db`

## 2) Chạy local không Docker (service todos)

Đảm bảo có PostgreSQL chạy trước.

Project hỗ trợ đọc biến môi trường theo thứ tự:

1. Biến môi trường hệ thống
2. File `.env` ở root project
3. Default value trong `config.go` (fallback)

Tạo file env local:

```bash
cp .env.example .env
```

Sau đó chỉnh `DB_DSN` nếu cần, rồi chạy:

```bash
make run-todos
```

Lưu ý:

- `.env` được ignore bởi git, không commit file này.
- Với Docker Compose, biến môi trường đã được khai báo trong `docker-compose.yml`.

---

## Database

Bảng chính: `todos`

Các cột đang dùng:

- `id`, `user_id`
- `title`, `description`
- `status`, `priority`
- `due_date`
- `created_at`, `updated_at`

---

## API gRPC

Định nghĩa tại `proto/todo/todo.proto`.

Service: `todo.v1.TodoService`

- `CreateTodo(CreateTodoRequest) returns (CreateTodoResponse)`
- `GetTodo(GetTodoRequest) returns (GetTodoResponse)`
- `ListTodos(ListTodosRequest) returns (ListTodosResponse)`
- `UpdateTodo(UpdateTodoRequest) returns (UpdateTodoResponse)`
- `DeleteTodo(DeleteTodoRequest) returns (DeleteTodoResponse)`

---

## Test API

## A) Test bằng Postman (gRPC)

1. New -> gRPC Request  
2. URL: `localhost:50051`  
3. Chọn method `todo.v1.TodoService/...`  
4. Nếu reflection không hiện đủ method, rebuild service:

```bash
docker compose down
docker compose up --build
```

## B) Test bằng grpcurl

Liệt kê methods:

```bash
grpcurl -plaintext localhost:50051 list todo.v1.TodoService
```

Ví dụ create:

```bash
grpcurl -plaintext \
  -d '{"user_id":1,"title":"Mua sua","description":"Ra sieu thi","priority":"HIGH"}' \
  localhost:50051 todo.v1.TodoService/CreateTodo
```

---

## Commands hữu ích

```bash
make proto        # regenerate protobuf code
make wire         # regenerate wire_gen.go
make fmt          # gofmt
make vet          # go vet
make build        # build tất cả app
```

---

## Ghi chú kỹ thuật

- gRPC reflection đã bật, thuận tiện test bằng Postman/grpcurl.
- Mapper đang parse/format thời gian dạng RFC3339 string theo proto hiện tại.
- Nếu thay đổi `todo.proto`, cần chạy lại `make proto`.
