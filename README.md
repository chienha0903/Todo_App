# Todo App

Project Todo App viết bằng Go, kiến trúc microservice: BFF (GraphQL) → Core Service (gRPC) → PostgreSQL, có sẵn Docker Compose để chạy local nhanh.

## Kiến trúc tổng quan

```
Client (Postman/Browser)
  │
  ▼
BFF Service (:8080)          ← GraphQL, gqlgen
  │
  ▼ gRPC
Core Service (:50051)        ← gRPC, domain logic
  │
  ▼
PostgreSQL (:5432)
```

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

## Cấu trúc thư mục

```text
.
├── docker-compose.yml
├── proto/
│   └── todo/todo.proto
├── services/
│   ├── todo-bff/
│   │   ├── Dockerfile
│   │   ├── gqlgen.yml
│   │   ├── cmd/main.go
│   │   └── internal/
│   │       ├── config/
│   │       ├── di/
│   │       ├── domain/
│   │       │   ├── gateway/
│   │       │   └── service/
│   │       ├── handler/
│   │       │   ├── graph/
│   │       │   │   ├── schema.graphql
│   │       │   │   ├── generated/
│   │       │   │   ├── model/
│   │       │   │   └── resolver/
│   │       │   ├── middleware/
│   │       │   └── server/
│   │       ├── infra/todo/
│   │       └── usecase/todo/
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

### 1) Chạy bằng Docker (khuyến nghị)

```bash
make docker-up
```

hoặc:

```bash
docker compose up -d --build
```

Lệnh trên sẽ khởi động 3 container: PostgreSQL, Core Service (gRPC), BFF (GraphQL).

Kiểm tra:

```bash
make docker-ps
make docker-logs-todos
make docker-logs-bff
```

Stop:

```bash
make docker-down
```

Mặc định:

- BFF (GraphQL): `localhost:8080`
- gRPC server: `localhost:50051`
- Postgres: `localhost:5432`
- DB: `todo_db`

### 2) Chạy local không Docker

Đảm bảo có PostgreSQL chạy trước.

Project hỗ trợ đọc biến môi trường theo thứ tự:

1. Biến môi trường hệ thống
2. File `.env` ở root project
3. Default value trong `config.go` (fallback)

Tạo file env local:

```bash
cp .env.example .env
```

Sau đó chỉnh `DB_DSN` nếu cần, rồi chạy từng service:

```bash
# Chạy core service (gRPC)
make run-todos

# Chạy BFF (GraphQL) — cần core service đang chạy
make run-bff
```

Lưu ý:

- `.env` được ignore bởi git, không commit file này.
- Với Docker Compose, biến môi trường đã được khai báo trong `docker-compose.yml`.
- BFF cần core service chạy trước để kết nối gRPC.

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

## API

### GraphQL (BFF)

Endpoint: `POST http://localhost:8080/graphql`

Schema tại `services/todo-bff/internal/handler/graph/schema.graphql`.

Query:

```graphql
# Lấy danh sách todos
query {
  todos(userId: 1, page: 1, pageSize: 10) {
    items { id title description status priority dueDate }
    total
    hasNext
  }
}

# Lấy 1 todo
query {
  todo(id: "1") {
    id title description status priority
  }
}
```

Mutation:

```graphql
# Tạo todo
mutation {
  createTodo(input: {
    userId: 1
    title: "Mua sua"
    description: "Ra sieu thi"
    priority: HIGH
  }) {
    id title status
  }
}

# Cập nhật todo
mutation {
  updateTodo(id: "1", input: {
    title: "Mua sua tuoi"
    status: COMPLETED
  }) {
    id title status updatedAt
  }
}

# Xóa todo
mutation {
  deleteTodo(id: "1")
}
```

### gRPC (Core Service)

Định nghĩa tại `proto/todo/todo.proto`.

Service: `todo.v1.TodoService`

- `CreateTodo(CreateTodoRequest) returns (CreateTodoResponse)`
- `GetTodo(GetTodoRequest) returns (GetTodoResponse)`
- `ListTodos(ListTodosRequest) returns (ListTodosResponse)`
- `UpdateTodo(UpdateTodoRequest) returns (UpdateTodoResponse)`
- `DeleteTodo(DeleteTodoRequest) returns (DeleteTodoResponse)`

---

## Test API

### A) Test BFF bằng Postman (GraphQL)

1. New → HTTP Request
2. Method: `POST`, URL: `http://localhost:8080/graphql`
3. Body → raw → JSON:

```json
{
  "query": "query { todos(userId: 1, page: 1, pageSize: 10) { items { id title } total hasNext } }"
}
```

### B) Test BFF bằng curl

```bash
# List todos
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { todos(userId: 1, page: 1, pageSize: 10) { items { id title } total hasNext } }"}'

# Create todo
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"mutation { createTodo(input: { userId: 1, title: \"Mua sua\", description: \"Ra sieu thi\", priority: HIGH }) { id title status } }"}'

# Health check
curl http://localhost:8080/health
```

### C) Test gRPC bằng grpcurl

```bash
# Liệt kê methods
grpcurl -plaintext localhost:50051 list todo.v1.TodoService

# Create todo
grpcurl -plaintext \
  -d '{"user_id":1,"title":"Mua sua","description":"Ra sieu thi","priority":"HIGH"}' \
  localhost:50051 todo.v1.TodoService/CreateTodo
```

### D) Test gRPC bằng Postman

1. New → gRPC Request
2. URL: `localhost:50051`
3. Chọn method `todo.v1.TodoService/...`
4. Nếu reflection không hiện đủ method, rebuild service:

```bash
docker compose down
docker compose up --build
```

---

## Middleware (BFF)

BFF có 2 middleware hoạt động ở tầng HTTP:

- **Logging middleware**: Log mỗi request với method, path, status, duration, request ID.
- **Recovery middleware**: Bắt panic, log stack trace, trả 500 thay vì crash server.

Log output ra stdout dạng structured text. Truyền header `X-Request-Id` để trace request.

---

## Commands hữu ích

```bash
make proto              # regenerate protobuf code
make wire               # regenerate wire_gen.go
make fmt                # gofmt
make vet                # go vet
make build              # build tất cả app
make run-todos          # chạy core service local
make run-bff            # chạy BFF local
make docker-up          # docker compose up
make docker-down        # docker compose down
make docker-ps          # xem container status
make docker-logs-todos  # log core service
make docker-logs-bff    # log BFF
```

---

## Ghi chú kỹ thuật

- gRPC reflection đã bật, thuận tiện test bằng Postman/grpcurl.
- Mapper đang parse/format thời gian dạng RFC3339 string theo proto hiện tại.
- Nếu thay đổi `todo.proto`, cần chạy lại `make proto`.
- BFF dùng `gqlgen` để generate GraphQL code. Nếu thay đổi `schema.graphql`, chạy `go generate ./...` trong thư mục BFF.
- Biến môi trường `TODOS_GRPC_ADDR` trong BFF trỏ tới core service. Trong Docker dùng `todos:50051` (service name), local dùng `localhost:50051`.