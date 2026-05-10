# Triển khai GraphQL vào Todo App

## Mục lục
1. [Kiến trúc tổng thể](#1-kiến-trúc-tổng-thể)
2. [GraphQL fit vào đâu?](#2-graphql-fit-vào-đâu)
3. [Cách suy nghĩ trước khi code](#3-cách-suy-nghĩ-trước-khi-code)
4. [Cấu trúc file thực tế](#4-cấu-trúc-file-thực-tế)
5. [Bước 1 — GraphQL Schema](#bước-1--graphql-schema)
6. [Bước 2 — gqlgen.yml](#bước-2--gqlgenyml)
7. [Bước 3 — Generate code](#bước-3--generate-code)
8. [Bước 4 — Error mapping](#bước-4--error-mapping)
9. [Bước 5 — Mapper](#bước-5--mapper)
10. [Bước 6 — Resolver struct](#bước-6--resolver-struct)
11. [Bước 7 — Implement Resolvers](#bước-7--implement-resolvers)
12. [Bước 8 — Wire DI](#bước-8--wire-di)
13. [Bước 9 — Test](#bước-9--test)
14. [Lỗi thường gặp](#lỗi-thường-gặp)

---

## 1. Kiến trúc tổng thể

```
Client
  │
  ├── REST  → POST /todos, GET /todos/:id ...
  │
  └── GraphQL → POST /graphql
         │
         ▼
      todo-bff (port 8080)
        internal/handler/graph/      ← GraphQL layer (thêm mới)
        internal/handler/http/       ← REST layer (giữ nguyên)
        internal/handler/mapper/     ← Dùng chung cho cả hai
         │
         ▼ gRPC :50051
      todos service
         │
         ▼
      PostgreSQL
```

### Pattern clean architecture của project (`todos` service làm mẫu)

```
handler/      ← Transport layer: nhận request, gọi mapper, gọi usecase
  grpc/
    todo/
      todo_handler.go    ← Handler chính
      error.go           ← Map domain error → gRPC status
    mapper/
      todo.go            ← Map proto ↔ input/output DTO

usecase/todo/
  usecase.go             ← Interfaces
  input/todo.go          ← Input DTOs
  output/todo.go         ← Output DTOs

domain/service/          ← Implements usecase interfaces
domain/gateway/          ← Repository interfaces (port)
infra/datastore/         ← Implements gateway interfaces
```

**BFF (`todo-bff`) là thin layer** — không có domain logic, chỉ translate transport. GraphQL layer ở BFF mirror đúng pattern này:

| `todos` service | `todo-bff` GraphQL |
|---|---|
| `handler/grpc/todo/todo_handler.go` | `handler/graph/resolver/schema.resolvers.go` |
| `handler/grpc/todo/error.go` | `handler/graph/error.go` |
| `handler/grpc/mapper/todo.go` | `handler/mapper/todo.go` |

---

## 2. GraphQL fit vào đâu?

### Vấn đề REST thuần túy giải quyết

**Over-fetching**: Client cần `id`, `title`, `status` nhưng server trả cả 9 fields.

**Under-fetching**: Cần 2 request cho 1 màn hình (`GET /todos` rồi `GET /users/:id`).

**GraphQL**: Client tự khai báo field cần, 1 request duy nhất.

```graphql
# Client chỉ cần 3 fields → server chỉ trả 3 fields
query {
  todos(userId: 1) {
    id
    title
    status
  }
}
```

### Tại sao thêm ở BFF?

1. BFF là aggregation layer — GraphQL phù hợp nhất ở đây
2. Backend (`todos` service + gRPC) **không thay đổi gì**
3. REST endpoints cũ **vẫn giữ nguyên**, GraphQL thêm vào `/graphql`

---

## 3. Cách suy nghĩ trước khi code

### Nguyên tắc: Schema trước, code sau

GraphQL schema = "hợp đồng" với client, giống `.proto` định nghĩa gRPC contract.

```
1. Nhìn vào proto (todo.proto)
   → Có CreateTodo, GetTodo, ListTodos, UpdateTodo, DeleteTodo
   → Todo có id, user_id, title, description, status, priority, due_date...

2. Thiết kế schema
   → Query (đọc) = GetTodo → todo(), ListTodos → todos()
   → Mutation (ghi) = CreateTodo, UpdateTodo, DeleteTodo

3. gqlgen generate → model structs + resolver interfaces (stubs)

4. Implement resolvers
   → Giống hệt REST handler: validate → map → gọi gRPC client → map response
```

### Thư viện: gqlgen

**Schema-first + code generation** — giống pattern protobuf của project. Type-safe, compiler bắt lỗi thay vì panic lúc runtime.

---

## 4. Cấu trúc file thực tế

```
services/todo-bff/
├── gqlgen.yml                                        ← Config generate
├── cmd/main.go
└── internal/
    ├── config/config.go
    ├── di/
    │   ├── wire.go                                   ← Có NewGRPCConn, NewTodoServiceClient
    │   └── wire_gen.go
    └── handler/
        ├── graph/
        │   ├── schema.graphql                        ← Ta viết: "hợp đồng" với client
        │   ├── error.go           (package graphql)  ← Map gRPC error → GraphQL error
        │   ├── generated/
        │   │   └── generated.go                      ← gqlgen tạo, ĐỪNG SỬA
        │   ├── model/
        │   │   └── models_gen.go                     ← gqlgen tạo, ĐỪNG SỬA
        │   └── resolver/
        │       ├── resolver.go                       ← Root Resolver struct + constructor
        │       └── schema.resolvers.go               ← Implement tất cả query + mutation
        ├── http/
        │   └── todo_handler.go                       ← REST handler, không đổi
        └── mapper/
            └── todo.go            (package mapper)   ← Map proto ↔ GraphQL model
```

> **Lưu ý quan trọng**: `error.go` nằm trong thư mục `graph/` nhưng khai báo `package graphql`
> — trong Go, tên package và tên thư mục không bắt buộc phải trùng nhau.
> Khi import `internal/handler/graph`, Go dùng tên package `graphql` để gọi.

---

## Bước 1 — GraphQL Schema

**File: `internal/handler/graph/schema.graphql`**

```graphql
type Todo {
  id:          ID!
  userId:      Int!
  title:       String!
  description: String!
  status:      String!
  priority:    String!
  dueDate:     String
  createdAt:   String!
  updatedAt:   String!
}

type Query {
  todo(id: ID!): Todo
  todos(userId: Int!): [Todo!]!
}

type Mutation {
  createTodo(input: CreateTodoInput!): Todo!
  updateTodo(id: ID!, input: UpdateTodoInput!): Todo!
  deleteTodo(id: ID!): Boolean!
}

input CreateTodoInput {
  userId:      Int!
  title:       String!
  description: String!
  priority:    String!
  dueDate:     String
}

input UpdateTodoInput {
  title:       String
  description: String
  priority:    String
  status:      String
  dueDate:     String
}
```

**Tại sao dùng `String` thay vì enum cho status/priority?**
Proto đang truyền string. Dùng String trước, tránh viết converter không cần thiết.

**Tại sao `ID` thay vì `Int`?**
GraphQL convention: `ID` là opaque identifier dạng string. Todo `id` từ proto là `int64`, ta format sang string trong mapper.

---

## Bước 2 — gqlgen.yml

**File: `gqlgen.yml`** (đặt ở root của `todo-bff`, cùng cấp với `go.mod` của service hoặc ở thư mục chạy lệnh)

```yaml
schema:
  - internal/handler/graph/schema.graphql

exec:
  filename: internal/handler/graph/generated/generated.go
  package: generated

model:
  filename: internal/handler/graph/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: internal/handler/graph/resolver
  package: resolver
  filename_template: "{name}.resolvers.go"

autobind: []
```

**Giải thích `autobind: []`**: Tắt tính năng tự động bind type Go với GraphQL type. Tránh gqlgen map nhầm sang struct ngoài ý muốn.

**Giải thích `layout: follow-schema`**: Tạo 1 resolver file theo từng schema file. Vì chỉ có 1 `schema.graphql` → tạo ra `schema.resolvers.go`.

---

## Bước 3 — Generate code

Chạy từ thư mục `services/todo-bff/`:

```bash
cd services/todo-bff
go run -mod=mod github.com/99designs/gqlgen generate
```

**Phải dùng `-mod=mod`** vì gqlgen cần các tool dependency (`golang.org/x/tools`, v.v.) mà `go mod tidy` có thể đã loại khỏi `go.sum`.

Sau lệnh này, gqlgen tạo ra:

**`internal/handler/graph/model/models_gen.go`** — Go structs từ schema:
```go
// ĐỪNG SỬA file này
type Todo struct {
    ID          string  `json:"id"`
    UserID      int     `json:"userId"`
    Title       string  `json:"title"`
    Description string  `json:"description"`
    Status      string  `json:"status"`
    Priority    string  `json:"priority"`
    DueDate     *string `json:"dueDate,omitempty"`
    CreatedAt   string  `json:"createdAt"`
    UpdatedAt   string  `json:"updatedAt"`
}

type CreateTodoInput struct {
    UserID      int     `json:"userId"`
    Title       string  `json:"title"`
    Description string  `json:"description"`
    Priority    string  `json:"priority"`
    DueDate     *string `json:"dueDate,omitempty"`
}

type UpdateTodoInput struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
    Priority    *string `json:"priority,omitempty"`
    Status      *string `json:"status,omitempty"`
    DueDate     *string `json:"dueDate,omitempty"`
}
```

**`internal/handler/graph/resolver/resolver.go`** — Root Resolver stub (chỉ tạo lần đầu, không bị overwrite):
```go
// File này KHÔNG bị regenerate — ta tự implement
type Resolver struct{}
```

**`internal/handler/graph/resolver/schema.resolvers.go`** — Stubs với `panic("not implemented")`:
```go
func (r *mutationResolver) CreateTodo(...) (*model.Todo, error) {
    panic(fmt.Errorf("not implemented"))
}
// ... các method khác tương tự
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
```

### Hiểu pattern `queryResolver`/`mutationResolver`

gqlgen không đặt method trực tiếp lên `Resolver` mà dùng wrapper types:

```
Resolver (root)
  ├── Query() → &queryResolver{r}   ← implements QueryResolver interface
  │     ├── Todo(...)
  │     └── Todos(...)
  └── Mutation() → &mutationResolver{r}  ← implements MutationResolver interface
        ├── CreateTodo(...)
        ├── UpdateTodo(...)
        └── DeleteTodo(...)
```

`mutationResolver` và `queryResolver` embed `*Resolver` → có thể dùng `r.todoClient`, `r.timeout` qua `r.Resolver.todoClient`.

**Mỗi khi sửa `schema.graphql` → chạy lại lệnh generate.** gqlgen chỉ thêm method mới vào `schema.resolvers.go`, không xoá code đã implement.

---

## Bước 4 — Error mapping

**File: `internal/handler/graph/error.go`**

```go
package graphql  // tên package là graphql, dù thư mục tên là graph

import (
    "context"
    "errors"
    "fmt"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// ToGraphQLError map gRPC error → Go error.
// GraphQL luôn trả HTTP 200, lỗi nằm trong field errors[] của JSON body.
// Pattern giống httpStatusFromGRPCCode trong handler/http/todo_handler.go
// nhưng không map sang HTTP status code.
func ToGraphQLError(err error) error {
    if err == nil {
        return nil
    }

    if errors.Is(err, context.DeadlineExceeded) {
        return errors.New("request timed out")
    }

    st, ok := status.FromError(err)
    if !ok {
        return errors.New("internal server error")
    }

    switch st.Code() {
    case codes.NotFound:
        return fmt.Errorf("not found: %s", st.Message())
    case codes.InvalidArgument:
        return fmt.Errorf("invalid argument: %s", st.Message())
    case codes.Unauthenticated:
        return errors.New("unauthorized")
    case codes.PermissionDenied:
        return errors.New("permission denied")
    case codes.DeadlineExceeded:
        return errors.New("request timed out")
    case codes.Unavailable:
        return errors.New("service unavailable")
    default:
        return errors.New("internal server error")
    }
}
```

---

## Bước 5 — Mapper

**File: `internal/handler/mapper/todo.go`**

```go
package mapper

import (
    "fmt"
    "time"

    todopb "github.com/chienha0903/Todo_App/proto/todo"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
)

// ToModel map proto Todo → GraphQL model Todo.
func ToModel(t *todopb.Todo) *model.Todo {
    if t == nil {
        return nil
    }

    m := &model.Todo{
        ID:          fmt.Sprintf("%d", t.Id),
        UserID:      int(t.UserId),
        Title:       t.Title,
        Description: t.Description,
        Status:      t.Status,
        Priority:    t.Priority,
        CreatedAt:   t.CreatedAt,
        UpdatedAt:   t.UpdatedAt,
    }

    if t.DueDate != "" {
        dueDate := t.DueDate
        m.DueDate = &dueDate
    }

    return m
}

// ToModels map slice proto Todo → slice GraphQL model Todo.
func ToModels(todos []*todopb.Todo) []*model.Todo {
    items := make([]*model.Todo, 0, len(todos))
    for _, t := range todos {
        items = append(items, ToModel(t))
    }
    return items
}

// ParseOptionalDueDate validate và trả string RFC3339 nếu có.
func ParseOptionalDueDate(value *string) (string, error) {
    if value == nil || *value == "" {
        return "", nil
    }
    if _, err := time.Parse(time.RFC3339, *value); err != nil {
        return "", fmt.Errorf("dueDate must be RFC3339 format")
    }
    return *value, nil
}
```

---

## Bước 6 — Resolver struct

Sửa file được gqlgen tạo ra: **`internal/handler/graph/resolver/resolver.go`**

```go
package resolver

import (
    "time"

    todopb "github.com/chienha0903/Todo_App/proto/todo"
)

// This file will not be regenerated automatically.

type Resolver struct {
    todoClient todopb.TodoServiceClient
    timeout    time.Duration
}

func NewResolver(todoClient todopb.TodoServiceClient, timeout time.Duration) *Resolver {
    return &Resolver{todoClient: todoClient, timeout: timeout}
}
```

---

## Bước 7 — Implement Resolvers

Implement vào **`internal/handler/graph/resolver/schema.resolvers.go`** (file gqlgen generate ra, ta thay `panic` bằng code thực):

```go
package resolver

import (
    "context"
    "fmt"
    "strconv"

    todograph "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/generated"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/mapper"
    todopb "github.com/chienha0903/Todo_App/proto/todo"
)

// ── Query ─────────────────────────────────────────────────────

func (r *queryResolver) Todo(ctx context.Context, id string) (*model.Todo, error) {
    todoID, err := parseID(id)
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    resp, err := r.todoClient.GetTodo(ctx, &todopb.GetTodoRequest{Id: todoID})
    if err != nil {
        return nil, todograph.ToGraphQLError(err)
    }

    return mapper.ToModel(resp.GetTodo()), nil
}

func (r *queryResolver) Todos(ctx context.Context, userID int) ([]*model.Todo, error) {
    if userID <= 0 {
        return nil, fmt.Errorf("invalid argument: userId must be a positive integer")
    }

    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    resp, err := r.todoClient.ListTodos(ctx, &todopb.ListTodosRequest{UserId: int64(userID)})
    if err != nil {
        return nil, todograph.ToGraphQLError(err)
    }

    return mapper.ToModels(resp.GetTodos()), nil
}

// ── Mutation ──────────────────────────────────────────────────

func (r *mutationResolver) CreateTodo(ctx context.Context, input model.CreateTodoInput) (*model.Todo, error) {
    dueDate, err := mapper.ParseOptionalDueDate(input.DueDate)
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    resp, err := r.todoClient.CreateTodo(ctx, &todopb.CreateTodoRequest{
        UserId:      int64(input.UserID),
        Title:       input.Title,
        Description: input.Description,
        Priority:    input.Priority,
        DueDate:     dueDate,
    })
    if err != nil {
        return nil, todograph.ToGraphQLError(err)
    }

    return mapper.ToModel(resp.GetTodo()), nil
}

func (r *mutationResolver) UpdateTodo(ctx context.Context, id string, input model.UpdateTodoInput) (*model.Todo, error) {
    todoID, err := parseID(id)
    if err != nil {
        return nil, err
    }

    dueDate, err := mapper.ParseOptionalDueDate(input.DueDate)
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    req := &todopb.UpdateTodoRequest{Id: todoID}
    if input.Title != nil {
        req.Title = *input.Title
    }
    if input.Description != nil {
        req.Description = *input.Description
    }
    if input.Priority != nil {
        req.Priority = *input.Priority
    }
    if input.Status != nil {
        req.Status = *input.Status
    }
    req.DueDate = dueDate

    resp, err := r.todoClient.UpdateTodo(ctx, req)
    if err != nil {
        return nil, todograph.ToGraphQLError(err)
    }

    return mapper.ToModel(resp.GetTodo()), nil
}

func (r *mutationResolver) DeleteTodo(ctx context.Context, id string) (bool, error) {
    todoID, err := parseID(id)
    if err != nil {
        return false, err
    }

    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    _, err = r.todoClient.DeleteTodo(ctx, &todopb.DeleteTodoRequest{Id: todoID})
    if err != nil {
        return false, todograph.ToGraphQLError(err)
    }

    return true, nil
}

// ── Helpers ───────────────────────────────────────────────────

func parseID(id string) (int64, error) {
    parsed, err := strconv.ParseInt(id, 10, 64)
    if err != nil || parsed <= 0 {
        return 0, fmt.Errorf("invalid argument: id must be a positive integer")
    }
    return parsed, nil
}

// ── gqlgen boilerplate (giữ nguyên từ generated) ──────────────

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() generated.QueryResolver       { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
```

### Import alias `todograph`

```go
todograph "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph"
```

Thư mục là `graph` nhưng package name là `graphql` (khai báo trong `error.go`). Dùng alias `todograph` để tránh nhầm với package `graphql` của thư viện gqlgen.

### Import `mapper`

```go
"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/mapper"
```

Đường dẫn là thư mục, không phải file. Package name = `mapper`.

---

## Bước 8 — Wire DI

**File: `internal/di/wire.go`**

Không dùng `client` package — wire trực tiếp `*grpc.ClientConn` và `todopb.TodoServiceClient`:

```go
//go:build wireinject

package di

import (
    nethttp "net/http"
    "time"

    todopb "github.com/chienha0903/Todo_App/proto/todo"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
    handlerhttp "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/http"
    "github.com/google/wire"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func InitHTTPServer(cfg *config.Config) (*nethttp.Server, error) {
    wire.Build(
        NewGRPCConn,
        NewTodoServiceClient,
        NewTodoHandler,
        NewHTTPServer,
    )
    return nil, nil
}

func NewGRPCConn(cfg *config.Config) (*grpc.ClientConn, error) {
    return grpc.NewClient(
        cfg.TodosGRPCAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
}

func NewTodoServiceClient(conn *grpc.ClientConn) todopb.TodoServiceClient {
    return todopb.NewTodoServiceClient(conn)
}

func NewTodoHandler(cfg *config.Config, svc todopb.TodoServiceClient) *handlerhttp.TodoHandler {
    return handlerhttp.NewTodoHandler(svc, cfg.RequestTimeout)
}

func NewHTTPServer(cfg *config.Config, todoHandler *handlerhttp.TodoHandler, conn *grpc.ClientConn) *nethttp.Server {
    server := &nethttp.Server{
        Addr:              ":" + cfg.AppPort,
        Handler:           todoHandler.Routes(),
        ReadHeaderTimeout: 5 * time.Second,
    }
    server.RegisterOnShutdown(func() {
        _ = conn.Close()
    })
    return server
}
```

> GraphQL server chưa được mount vào HTTP server — bước này cần thêm sau khi tạo GraphQL HTTP handler.

---

## Bước 9 — Test

Verify build trước:

```bash
cd services/todo-bff
go build ./...
```

Chạy server và test bằng curl (chưa có GraphQL HTTP handler thì chưa test được qua browser):

```bash
# Query
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ todos(userId: 1) { id title status } }"}'

# Mutation
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"mutation { createTodo(input: { userId: 1, title: \"Test\", description: \"Desc\", priority: \"HIGH\" }) { id title } }"}'
```

Response luôn trả **HTTP 200**. Lỗi nằm trong body:

```json
{
  "data": null,
  "errors": [{ "message": "not found: todo with id 999" }]
}
```

---

## Lỗi thường gặp

### 1. `missing go.sum entry` khi chạy generate

```bash
# Dùng -mod=mod để gqlgen tự update go.sum
go run -mod=mod github.com/99designs/gqlgen generate
```

### 2. Schema thay đổi nhưng resolver không update

Chạy lại generate từ đúng thư mục:
```bash
cd services/todo-bff
go run -mod=mod github.com/99designs/gqlgen generate
```
gqlgen chỉ **thêm** method mới vào `schema.resolvers.go`, không xoá code đã implement.

### 3. Import `internal/handler/graph` — gọi được function nào?

Package name là `graphql` (khai báo trong `error.go`), không phải `graph`. Phải dùng alias:

```go
// Đúng
todograph "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph"
todograph.ToGraphQLError(err)

// Sai — "graph" không phải package name
graph.ToGraphQLError(err)
```

### 4. Import `internal/handler/mapper/todo` — sai

```go
// Sai — todo là tên file, không phải package
"github.com/.../internal/handler/mapper/todo"

// Đúng — import theo đường dẫn thư mục
"github.com/.../internal/handler/mapper"
```

### 5. Method trên `*Resolver` thay vì `*queryResolver`

```go
// Sai — gqlgen không gọi method trực tiếp trên Resolver
func (r *Resolver) Todo(ctx context.Context, ...) (*model.Todo, error) {}

// Đúng — gqlgen dùng wrapper types
func (r *queryResolver) Todo(ctx context.Context, ...) (*model.Todo, error) {}
```

---

## Checklist

- [x] `internal/handler/graph/schema.graphql` — định nghĩa schema
- [x] `gqlgen.yml` — config đúng paths
- [x] `go run -mod=mod github.com/99designs/gqlgen generate` — generate thành công
- [x] `internal/handler/graph/error.go` — `ToGraphQLError` exported
- [x] `internal/handler/mapper/todo.go` — `ToModel`, `ToModels`, `ParseOptionalDueDate`
- [x] `internal/handler/graph/resolver/resolver.go` — `Resolver` struct + `NewResolver`
- [x] `internal/handler/graph/resolver/schema.resolvers.go` — implement hết 5 resolvers
- [x] `internal/di/wire.go` + `wire_gen.go` — không dùng client package
- [x] Thêm GraphQL HTTP handler và mount vào HTTP server
- [x] `go build ./...` không lỗi
