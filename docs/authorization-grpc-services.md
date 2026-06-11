# Hướng dẫn implement Authorization cho Core gRPC Services

**Phạm vi:** `services/todos` · `services/users`  
**Prerequisite:** JWT auth đang hoạt động tốt ở BFF layer  

---

## 1. Vấn đề hiện tại

### Kiến trúc auth hiện tại

```
Internet
   │
   ▼
┌─────────────────────────────────────────────┐
│  todo-bff  (port 8080)                      │
│  ┌─────────────────────────────────────┐    │
│  │  AuthMiddleware                     │    │
│  │  • Extract Bearer token từ header   │    │
│  │  │  ParseToken(token, JWTSecret)    │    │
│  │  │  Store userID + role in ctx      │    │
│  └─────────────────────────────────────┘    │
│  ┌──────────────┐  ┌──────────────────────┐ │
│  │ gRPC client  │  │  gRPC client         │ │
│  │ (todos)      │  │  (users)             │ │
│  │ ctx truyền   │  │  ctx truyền          │ │
│  │ thẳng, KHÔNG │  │  thẳng, KHÔNG        │ │
│  │ attach token │  │  attach token        │ │
│  └──────┬───────┘  └──────────┬───────────┘ │
└─────────┼────────────────────┼─────────────┘
          │                    │
          ▼                    ▼
┌──────────────────┐  ┌──────────────────────┐
│  todos service   │  │  users service       │
│  (port 50051)    │  │  (port 50052)        │
│                  │  │                      │
│  ❌ Không có     │  │  ❌ Không có         │
│  auth interceptor│  │  auth interceptor    │
│  → Accept mọi   │  │  → Accept mọi       │
│  request         │  │  request             │
└──────────────────┘  └──────────────────────┘
```

### Các lỗ hổng

| # | Lỗ hổng | Impact |
|---|---------|--------|
| 1 | Ai có thể reach port `50051`/`50052` đều gọi được gRPC trực tiếp, không cần JWT | **Critical** |
| 2 | `CreateTodo` nhận `user_id` từ request field — caller tự đặt, backend không verify | **High** |
| 3 | `ListTodos` nhận `user_id` từ request — có thể list todos của người khác | **High** |
| 4 | `GetTodo`, `UpdateTodo`, `DeleteTodo` không check ownership | **High** |
| 5 | users service: `GetUser`, `UpdateUser`, `DeleteUser` không check ownership | **High** |

---

## 2. Kiến trúc mục tiêu

```
Internet
   │
   ▼
┌─────────────────────────────────────────────────────┐
│  todo-bff  (port 8080)                              │
│  ┌──────────────────────────────────────────────┐   │
│  │  AuthMiddleware                              │   │
│  │  ParseToken → store {userID, role, rawToken} │   │
│  └──────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────┐   │
│  │  gRPC client interceptor (forwardAuth)       │   │
│  │  GetToken(ctx) → metadata "authorization"    │   │
│  └──────┬─────────────────────────┬─────────────┘   │
└─────────┼─────────────────────────┼─────────────────┘
          │  metadata: authorization: Bearer <jwt>
          ▼                         ▼
┌──────────────────────┐  ┌──────────────────────────┐
│  todos service       │  │  users service           │
│  ┌────────────────┐  │  │  ┌────────────────────┐  │
│  │ Recovery       │  │  │  │ Recovery           │  │
│  │ Logging        │  │  │  │ Logging            │  │
│  │ ✅ AuthIntercp │  │  │  │ ✅ AuthInterceptor │  │
│  │   ParseToken   │  │  │  │   ParseToken       │  │
│  │   store caller │  │  │  │   store caller     │  │
│  │   in ctx       │  │  │  │   in ctx           │  │
│  └────────────────┘  │  │  └────────────────────┘  │
│  ┌────────────────┐  │  │  ┌────────────────────┐  │
│  │ TodoHandler    │  │  │  │ UserHandler        │  │
│  │ ✅ Check       │  │  │  │ ✅ Check           │  │
│  │   ownership    │  │  │  │   ownership        │  │
│  └────────────────┘  │  │  └────────────────────┘  │
└──────────────────────┘  └──────────────────────────┘
```

### Cơ chế: Forward JWT qua gRPC Metadata

```
BFF (HTTP context)           gRPC Metadata              Backend context
──────────────────      ──────────────────────      ──────────────────────
ctx.userID = 42    →    authorization:              ctx.callerID = 42
ctx.role = "USER"  →    Bearer eyJhbGci...          ctx.callerRole = "USER"
ctx.token = "eyJ…"      (validated tại BFF)
```

**Tại sao không dùng service account / mTLS?**  
Không cần thêm infrastructure. JWT đã có ở tất cả services. Mỗi backend service độc lập validate token — không tin tưởng mù quáng vào BFF.

---

## 3. Kế hoạch implement

| Phase | Thay đổi | Files |
|-------|----------|-------|
| 1 | Tạo `pkg/jwt` shared package | `pkg/jwt/jwt.go` |
| 2 | Cập nhật config todos/users thêm JWTSecret | `services/todos/internal/config/config.go` · `services/users/internal/config/config.go` |
| 3 | BFF: store raw token + client interceptor | `middleware/auth.go` · `infra/todo/grpc.go` · `infra/user/grpc.go` |
| 4 | Backend: auth interceptor + context helpers | `todos/handler/grpc/auth_interceptor.go` · `todos/handler/grpc/caller.go` · `users/handler/grpc/auth_interceptor.go` · `users/handler/grpc/caller.go` |
| 5 | Backend: wire auth interceptor vào server | `todos/handler/grpc/server.go` · `users/handler/grpc/server.go` |
| 6 | Backend: ownership check trong handlers | `todos/handler/grpc/todo/todo_handler.go` · `users/handler/grpc/user/user_handler.go` |
| 7 | Config: update docker-compose + env | `docker-compose.yml` · `.env.example` |

---

## Phase 1 — Shared JWT Package

Tạo `pkg/jwt/jwt.go` để tái sử dụng logic parse token ở tất cả services, tránh duplicate code.

```go
// pkg/jwt/jwt.go
package jwt

import (
    gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID int64  `json:"user_id"`
    Role   string `json:"role"`
    gojwt.RegisteredClaims
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
    token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (any, error) {
        if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
            return nil, gojwt.ErrSignatureInvalid
        }
        return []byte(secret), nil
    })
    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, gojwt.ErrTokenInvalidClaims
    }

    return claims, nil
}
```

> **Note:** `services/todo-bff/internal/jwt/jwt.go` và `services/users/internal/infra/jwt/` hiện đang duplicate logic này. Sau khi tạo `pkg/jwt`, refactor chúng để import từ package chung (không bắt buộc ngay).

---

## Phase 2 — Todos Config: Thêm JWTSecret

**File:** `services/todos/internal/config/config.go`

```go
package config

import (
    "errors"
    "os"

    "github.com/joho/godotenv"
)

const minJWTSecretLen = 32

type Config struct {
    AppName     string
    AppPort     string
    AppEnv      string
    DBDSN       string
    JWTSecret   string  // ← THÊM MỚI
    RabbitMQURL string
}

func Load() (*Config, error) {
    _ = godotenv.Load()

    cfg := &Config{
        AppName:     getenv("APP_NAME", "todo-app"),
        AppPort:     getenv("APP_TODO_PORT", "50051"),
        AppEnv:      getenv("APP_ENV", "development"),
        DBDSN:       getenv("DB_TODO_DSN", "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"),
        JWTSecret:   getenv("JWT_SECRET", ""),  // ← THÊM MỚI
        RabbitMQURL: getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
    }

    if len(cfg.JWTSecret) < minJWTSecretLen {
        return nil, errors.New("config: JWT_SECRET must be at least 32 characters")
    }

    return cfg, nil
}
```

> Users service đã có validation này (từ audit trước). Todos service cần được cập nhật tương tự.

---

## Phase 3 — BFF: Store Raw Token + Client Interceptor

### 3a. Lưu raw token vào context

**File:** `services/todo-bff/internal/handler/middleware/auth.go`

```go
// Thêm context key mới
const (
    ctxUserID contextKey = "user_id"
    ctxRole   contextKey = "role"
    ctxToken  contextKey = "token"  // ← THÊM MỚI
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)

            if token != "" {
                claim, err := userjwt.ParseToken(token, jwtSecret)
                if err == nil {
                    ctx := context.WithValue(r.Context(), ctxUserID, claim.UserID)
                    ctx = context.WithValue(ctx, ctxRole, claim.Role)
                    ctx = context.WithValue(ctx, ctxToken, token)  // ← THÊM MỚI
                    r = r.WithContext(ctx)
                }
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Thêm helper function mới
func GetToken(ctx context.Context) (string, bool) {
    token, ok := ctx.Value(ctxToken).(string)
    return token, ok
}
```

### 3b. Client interceptor cho todos gRPC connection

**File:** `services/todo-bff/internal/infra/todo/grpc.go`

```go
package todo

import (
    "context"
    "fmt"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"  // ← THÊM

    todopb "github.com/chienha0903/Todo_App/proto/todo"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/middleware"  // ← THÊM
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
    "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

// forwardAuthInterceptor đọc JWT token từ HTTP context và inject vào gRPC metadata.
// Backend service sẽ validate lại token này độc lập.
func forwardAuthInterceptor(
    ctx context.Context,
    method string,
    req, reply any,
    cc *grpc.ClientConn,
    invoker grpc.UnaryInvoker,
    opts ...grpc.CallOption,
) error {
    if token, ok := middleware.GetToken(ctx); ok && token != "" {
        ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
    }
    return invoker(ctx, method, req, reply, cc, opts...)
}

func NewGRPCConn(cfg *config.Config) (*ClientConn, func(), error) {
    conn, err := grpc.NewClient(
        cfg.TodosGRPCAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(forwardAuthInterceptor),  // ← THÊM
    )
    if err != nil {
        return nil, nil, fmt.Errorf("grpc dial %s: %w", cfg.TodosGRPCAddr, err)
    }

    return (*ClientConn)(conn), func() { _ = conn.Close() }, nil
}

// ... phần còn lại không đổi
```

### 3c. Client interceptor cho users gRPC connection

**File:** `services/todo-bff/internal/infra/user/grpc.go`

Áp dụng **hoàn toàn tương tự** như `todo/grpc.go`:

```go
// Thêm import:
"google.golang.org/grpc/metadata"
"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/middleware"

// Thêm hàm:
func forwardAuthInterceptor(...) error { /* giống todo/grpc.go */ }

// Trong NewGRPCConn:
grpc.WithUnaryInterceptor(forwardAuthInterceptor),
```

---

## Phase 4 — Backend: Auth Interceptor + Context Helpers

### 4a. Context helpers cho todos service

**File mới:** `services/todos/internal/handler/grpc/caller.go`

```go
package grpc

import "context"

type callerKey struct{}

type Caller struct {
    UserID int64
    Role   string
}

// callerFromContext lấy thông tin caller đã được auth interceptor inject.
func callerFromContext(ctx context.Context) (Caller, bool) {
    c, ok := ctx.Value(callerKey{}).(Caller)
    return c, ok
}

func contextWithCaller(ctx context.Context, userID int64, role string) context.Context {
    return context.WithValue(ctx, callerKey{}, Caller{UserID: userID, Role: role})
}
```

### 4b. Auth interceptor cho todos service

**File mới:** `services/todos/internal/handler/grpc/auth_interceptor.go`

```go
package grpc

import (
    "context"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"

    jwtpkg "github.com/chienha0903/Todo_App/pkg/jwt"
)

// UnaryAuthInterceptor validate JWT từ gRPC metadata và inject Caller vào context.
// Nếu token không hợp lệ hoặc thiếu → trả codes.Unauthenticated.
func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req any,
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (any, error) {
        token, err := extractTokenFromMD(ctx)
        if err != nil {
            return nil, err
        }

        claims, err := jwtpkg.ParseToken(token, jwtSecret)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
        }

        ctx = contextWithCaller(ctx, claims.UserID, claims.Role)
        return handler(ctx, req)
    }
}

func extractTokenFromMD(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", status.Error(codes.Unauthenticated, "missing metadata")
    }

    values := md.Get("authorization")
    if len(values) == 0 || values[0] == "" {
        return "", status.Error(codes.Unauthenticated, "missing authorization header")
    }

    token := strings.TrimPrefix(values[0], "Bearer ")
    if token == values[0] { // không có prefix "Bearer "
        return "", status.Error(codes.Unauthenticated, "authorization must be Bearer token")
    }

    return token, nil
}
```

### 4c. Context helpers + Auth interceptor cho users service

**File mới:** `services/users/internal/handler/grpc/caller.go`

```go
package grpc

import "context"

type callerKey struct{}

type Caller struct {
    UserID int64
    Role   string
}

func callerFromContext(ctx context.Context) (Caller, bool) {
    c, ok := ctx.Value(callerKey{}).(Caller)
    return c, ok
}

func contextWithCaller(ctx context.Context, userID int64, role string) context.Context {
    return context.WithValue(ctx, callerKey{}, Caller{UserID: userID, Role: role})
}
```

**File mới:** `services/users/internal/handler/grpc/auth_interceptor.go`

```go
// Nội dung giống hệt todos/handler/grpc/auth_interceptor.go
// Chỉ khác package import path
package grpc

import (
    "context"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"

    jwtpkg "github.com/chienha0903/Todo_App/pkg/jwt"
)

func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        token, err := extractTokenFromMD(ctx)
        if err != nil {
            return nil, err
        }

        claims, err := jwtpkg.ParseToken(token, jwtSecret)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
        }

        ctx = contextWithCaller(ctx, claims.UserID, claims.Role)
        return handler(ctx, req)
    }
}

func extractTokenFromMD(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", status.Error(codes.Unauthenticated, "missing metadata")
    }

    values := md.Get("authorization")
    if len(values) == 0 || values[0] == "" {
        return "", status.Error(codes.Unauthenticated, "missing authorization header")
    }

    token := strings.TrimPrefix(values[0], "Bearer ")
    if token == values[0] {
        return "", status.Error(codes.Unauthenticated, "authorization must be Bearer token")
    }

    return token, nil
}
```

---

## Phase 5 — Wire Auth Interceptor vào gRPC Server

### 5a. todos gRPC server

**File:** `services/todos/internal/handler/grpc/server.go`

```go
package grpc

import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
    "google.golang.org/grpc/reflection"

    todopb "github.com/chienha0903/Todo_App/proto/todo"
    "github.com/chienha0903/Todo_App/services/todos/internal/config"
    todohandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/todo"
)

func NewGRPCServer(cfg *config.Config, h *todohandler.TodoHandler) *grpc.Server {
    srv := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            UnaryRecoveryInterceptor,               // 1. catch panics
            UnaryLoggingInterceptor,                // 2. log all requests (kể cả auth fail)
            UnaryAuthInterceptor(cfg.JWTSecret),    // 3. validate JWT ← THÊM MỚI
        ),
    )

    todopb.RegisterTodoServiceServer(srv, h)
    grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
    reflection.Register(srv)

    return srv
}
```

> **Quan trọng:** `NewGRPCServer` cần thêm tham số `cfg *config.Config`. Cần cập nhật Wire DI injection.

### 5b. Wire DI — truyền config vào NewGRPCServer

**File:** `services/todos/internal/di/wire.go`

```go
// Trong ProviderSet, thêm:
wire.Build(
    config.Load,
    // ...
    grpc.NewGRPCServer,  // wire tự inject cfg vì cfg đã có trong graph
)
```

**File:** `services/todos/internal/di/wire_gen.go`  
_(chạy `wire gen ./services/todos/internal/di/` để regenerate)_

### 5c. users gRPC server

**File:** `services/users/internal/handler/grpc/server.go`

```go
func NewGRPCServer(cfg *config.Config, h *userhandler.UserHandler) *grpc.Server {
    srv := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            UnaryRecoveryInterceptor,
            UnaryLoggingInterceptor,
            UnaryAuthInterceptor(cfg.JWTSecret),  // ← THÊM MỚI
        ),
    )

    userpb.RegisterUserserviceServer(srv, h)
    grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
    reflection.Register(srv)

    return srv
}
```

---

## Phase 6 — Ownership Checks trong Handlers

### 6a. todos handler

**File:** `services/todos/internal/handler/grpc/todo/todo_handler.go`

```go
func (h *TodoHandler) CreateTodo(ctx context.Context, req *todopb.CreateTodoRequest) (*todopb.CreateTodoResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    // Enforce: chỉ được tạo todo cho chính mình (trừ ADMIN)
    if caller.Role != "ADMIN" && req.UserId != caller.UserID {
        return nil, status.Error(codes.PermissionDenied, "cannot create todo for another user")
    }

    in, err := mapper.ToCreateTodoInput(req)
    if err != nil {
        return nil, toGRPCError(err)
    }

    out, err := h.creater.Create(ctx, in)
    if err != nil {
        return nil, toGRPCError(err)
    }

    return &todopb.CreateTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) ListTodos(ctx context.Context, req *todopb.ListTodosRequest) (*todopb.ListTodosResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    // Enforce: chỉ list todos của mình (trừ ADMIN)
    if caller.Role != "ADMIN" {
        req.UserId = caller.UserID
    }

    page, err := h.lister.List(ctx, mapper.ToListTodosInput(req))
    if err != nil {
        return nil, toGRPCError(err)
    }

    return &todopb.ListTodosResponse{
        Todos:    mapper.ToProtoTodos(page.Items),
        Total:    page.Total,
        Page:     page.Page,
        PageSize: page.PageSize,
    }, nil
}

func (h *TodoHandler) UpdateTodo(ctx context.Context, req *todopb.UpdateTodoRequest) (*todopb.UpdateTodoResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    // Kiểm tra quyền sở hữu: lấy todo trước, so sánh user_id
    if caller.Role != "ADMIN" {
        existing, err := h.getter.Get(ctx, &ucin.GetTodo{ID: req.Id})
        if err != nil {
            return nil, toGRPCError(err)
        }
        if existing.UserID != caller.UserID {
            return nil, status.Error(codes.PermissionDenied, "cannot update another user's todo")
        }
    }

    in, err := mapper.ToUpdateTodoInput(req)
    if err != nil {
        return nil, toGRPCError(err)
    }

    out, err := h.updater.Update(ctx, in)
    if err != nil {
        return nil, toGRPCError(err)
    }

    return &todopb.UpdateTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) DeleteTodo(ctx context.Context, req *todopb.DeleteTodoRequest) (*todopb.DeleteTodoResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    if caller.Role != "ADMIN" {
        existing, err := h.getter.Get(ctx, &ucin.GetTodo{ID: req.Id})
        if err != nil {
            return nil, toGRPCError(err)
        }
        if existing.UserID != caller.UserID {
            return nil, status.Error(codes.PermissionDenied, "cannot delete another user's todo")
        }
    }

    if err := h.deleter.Delete(ctx, mapper.ToDeleteTodoInput(req)); err != nil {
        return nil, toGRPCError(err)
    }

    return &todopb.DeleteTodoResponse{}, nil
}
```

> **Import cần thêm:**  
> ```go
> grpcpkg "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc"
> ucin "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
> "google.golang.org/grpc/codes"
> "google.golang.org/grpc/status"
> ```

### 6b. users handler — các endpoint nhạy cảm

**File:** `services/users/internal/handler/grpc/user/user_handler.go`

```go
// GetUser: user chỉ xem được profile của mình
func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    if caller.Role != "ADMIN" && req.Id != caller.UserID {
        return nil, status.Error(codes.PermissionDenied, "cannot access another user's profile")
    }

    // ... logic hiện tại
}

// UpdateUser: user chỉ update được profile của mình
func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    if caller.Role != "ADMIN" && req.Id != caller.UserID {
        return nil, status.Error(codes.PermissionDenied, "cannot update another user's profile")
    }

    // ... logic hiện tại
}

// DeleteUser: chỉ ADMIN mới được xóa user
func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    if caller.Role != "ADMIN" {
        return nil, status.Error(codes.PermissionDenied, "only admin can delete users")
    }

    // ... logic hiện tại
}

// Login, RefreshToken — ĐÂY là các endpoint PUBLIC, không cần auth check
// Chúng được gọi trước khi có token → interceptor cần bỏ qua chúng

// ChangePassword: verify đây là user của chính mình
func (h *UserHandler) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error) {
    caller, ok := grpcpkg.CallerFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing caller identity")
    }

    if req.UserId != caller.UserID {
        return nil, status.Error(codes.PermissionDenied, "cannot change another user's password")
    }

    // ... logic hiện tại
}
```

---

## Phase 4 (bổ sung) — Xử lý Public Endpoints trong users service

`Login` và `RefreshToken` là **public** — không có JWT khi gọi. Auth interceptor sẽ reject chúng.

**Giải pháp: Skip list trong interceptor**

```go
// services/users/internal/handler/grpc/auth_interceptor.go

// publicMethods là các gRPC methods không yêu cầu authentication.
var publicMethods = map[string]bool{
    "/user.Userservice/Login":        true,
    "/user.Userservice/RefreshToken": true,
    "/user.Userservice/Register":     true, // nếu có
}

func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        // Bỏ qua auth cho public endpoints
        if publicMethods[info.FullMethod] {
            return handler(ctx, req)
        }

        token, err := extractTokenFromMD(ctx)
        if err != nil {
            return nil, err
        }

        claims, err := jwtpkg.ParseToken(token, jwtSecret)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
        }

        ctx = contextWithCaller(ctx, claims.UserID, claims.Role)
        return handler(ctx, req)
    }
}
```

> **Kiểm tra method name:** Dùng `grpcurl -plaintext localhost:50052 list user.Userservice` để xem exact method names, hoặc xem trong `proto/user/user.proto`.

---

## Phase 7 — Cập nhật docker-compose + .env.example

### docker-compose.yml

```yaml
todos:
  environment:
    JWT_SECRET: "chien-apvn-super-secret-key-2024"  # ← THÊM MỚI
    # ... biến hiện có
```

### .env.example

```env
# JWT (dùng chung cho tất cả services — phải ≥ 32 ký tự)
JWT_SECRET=your-super-secret-key-minimum-32-chars
```

---

## Thứ tự thực thi + Kiểm tra

### Build & test sau khi implement

```bash
# 1. Tạo pkg/jwt
# 2. Cập nhật tất cả files theo hướng dẫn
# 3. Regenerate Wire DI
wire gen ./services/todos/internal/di/
wire gen ./services/users/internal/di/

# 4. Build
make build

# 5. Unit test (bao gồm -race)
make test-race

# 6. Lint
make lint
```

### Kiểm tra thủ công sau khi deploy

```bash
# Khởi động hệ thống
docker compose up -d
docker compose ps

# Test 1: Gọi gRPC trực tiếp không có token → expect UNAUTHENTICATED
grpcurl -plaintext localhost:50051 todo.TodoService/ListTodos

# Test 2: Gọi với token hợp lệ qua BFF → expect 200 OK
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer <valid-jwt>" \
  -H "Content-Type: application/json" \
  -d '{"query": "{ todos(userID: 1, page: 1, pageSize: 10) { items { id title } } }"}'

# Test 3: Gọi với token hết hạn → expect UNAUTHENTICATED
# Test 4: User thường cố update todo của người khác → expect PERMISSION_DENIED
# Test 5: Login (public endpoint) vẫn hoạt động không cần token
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { login(email: \"test@test.com\", password: \"pass\") { token } }"}'
```

---

## Lưu ý quan trọng

### Circular dependency risk

`todo_handler.go` sẽ import package `grpc` (cùng service) để dùng `callerFromContext`. Đây là **circular import** vì `grpc` package đã import `todo` package.

**Giải pháp:** Move context helpers ra package riêng:

```
services/todos/internal/handler/grpc/
├── server.go
├── interceptor.go
├── auth_interceptor.go
├── caller/              ← package riêng
│   └── caller.go        ← CallerFromContext, contextWithCaller
└── todo/
    └── todo_handler.go  ← import "caller" package, không import "grpc" package
```

```go
// services/todos/internal/handler/grpc/caller/caller.go
package caller

import "context"

type key struct{}

type Caller struct {
    UserID int64
    Role   string
}

func FromContext(ctx context.Context) (Caller, bool) { ... }
func WithContext(ctx context.Context, userID int64, role string) context.Context { ... }
```

### JWT Secret management

Tất cả services dùng chung một `JWT_SECRET`. Điều này có nghĩa:
- Rotate secret → phải deploy lại cùng lúc tất cả services
- Nếu secret bị leak → toàn bộ hệ thống bị compromise

**Long-term solution:** Dùng asymmetric JWT (RS256):
- `users` service giữ private key, ký token
- `todos` và `bff` giữ public key, chỉ verify

```
JWT_PRIVATE_KEY (chỉ users service)
JWT_PUBLIC_KEY  (todos + bff — chỉ dùng để verify)
```

### Network isolation

Authorization trong code chỉ là **defence in depth**, không thay thế network security:

```yaml
# docker-compose.yml — thêm network isolation
networks:
  frontend:      # todo-bff ↔ internet
  backend:       # todos + users — không expose ra ngoài

services:
  todos:
    networks: [backend]
    # Không có ports: expose ra host → chỉ bff mới reach được
  users:
    networks: [backend]
  todo-bff:
    networks: [frontend, backend]
    ports: ["8080:8080"]
```

---

## Tóm tắt Flow hoàn chỉnh

```
1. Client gọi: POST /graphql
   Authorization: Bearer eyJhbGci...

2. BFF AuthMiddleware:
   ParseToken(token) → {userID: 42, role: "USER"}
   ctx.userID = 42
   ctx.role   = "USER"
   ctx.token  = "eyJhbGci..."  ← LƯU RAW TOKEN

3. BFF GraphQL Resolver:
   callerID, _ = middleware.GetUserID(ctx)  // = 42
   gọi grpcGateway.CreateTodo(ctx, ...)

4. BFF gRPC Client interceptor (forwardAuth):
   token = middleware.GetToken(ctx)  // = "eyJhbGci..."
   metadata["authorization"] = "Bearer eyJhbGci..."

5. todos gRPC Server interceptors (theo thứ tự):
   RecoveryInterceptor → pass through
   LoggingInterceptor  → start timer
   AuthInterceptor:
     extractToken(ctx) → "eyJhbGci..."
     ParseToken(token, JWTSecret) → {userID: 42, role: "USER"}
     ctx.caller = Caller{42, "USER"}

6. TodoHandler.CreateTodo:
   caller = CallerFromContext(ctx)  // = {42, "USER"}
   req.UserId == caller.UserID? → OK
   creater.Create(ctx, input)

7. Response trả về → Logging ghi duration + status
```
