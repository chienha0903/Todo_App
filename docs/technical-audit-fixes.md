# Technical Audit & Fixes — Go Microservices Todo App

**Date:** 2026-06-10  
**Role:** Senior Golang Backend Architect  
**Scope:** Production-readiness audit + concrete fixes  

---

## Tổng quan

Audit tập trung vào **backend/production-readiness** — không viết lại kiến trúc, không thêm Kubernetes, không thay đổi business logic. Mục tiêu: vá các lỗ hổng bảo mật, race condition, và cấu hình thiếu sót trước khi đưa lên production.

---

## Gap Analysis

| # | Category | Issue | Priority | Status |
|---|----------|-------|----------|--------|
| 1 | Security | Authorization bypass: `createTodo` nhận `userID` từ GraphQL input | **P0** | ✅ Fixed |
| 2 | Security | Authorization bypass: `todos` query cho phép xem todos của user khác | **P0** | ✅ Fixed |
| 3 | Correctness | Transaction bypass trong `SoftDeleteByUserID` phá vỡ atomicity | **P0** | ✅ Fixed |
| 4 | Security | JWT_SECRET mặc định `"chien-apvn"` (10 ký tự) — quá yếu | **P1** | ✅ Fixed |
| 5 | Reliability | DB connection pool không giới hạn → PostgreSQL `max_connections` bị exhausted | **P1** | ✅ Fixed |
| 6 | Security | HTTP server thiếu `ReadTimeout`/`WriteTimeout` → slow-loris attack | **P1** | ✅ Fixed |
| 7 | Reliability | Outbox worker retry mãi mãi với event bị lỗi permanent | **P1** | ✅ Fixed |
| 8 | Observability | gRPC server không có health check endpoint | **P1** | ✅ Fixed |
| 9 | CI/CD | CI gate sai: integration test fail không block merge | **P1** | ✅ Fixed |
| 10 | DX | Makefile thiếu `lint`, `test-race`, `test-cover` | **P1** | ✅ Fixed |
| 11 | Config | `.env.example` incomplete — thiếu `JWT_SECRET`, `RABBITMQ_URL`, v.v. | **P1** | ✅ Fixed |
| 12 | Config | `docker-compose.yml` có JWT_SECRET 10 ký tự, sẽ fail sau khi thêm validation | **P1** | ✅ Fixed |
| 13 | Pagination | Không giới hạn `pageSize` → có thể dump toàn bộ database | **P1** | ✅ Fixed |

---

## Chi tiết từng Fix

---

### Fix #1 — P0: Transaction Bypass trong SoftDeleteByUserID

**File:** `services/todos/internal/infra/datastore/todo_command_repo.go`

**Vấn đề:**  
Consumer xử lý event `user.deleted` chạy trong 1 transaction: soft-delete todos + insert vào `processed_events`. Nhưng `SoftDeleteByUserID` gọi `r.db.WithContext(ctx)` trực tiếp thay vì dùng `extractDB(ctx, r.db)`. Kết quả: soft-delete chạy **ngoài transaction**, nếu insert `processed_events` fail sau đó thì soft-delete không được rollback → **data inconsistency**.

**Trước:**
```go
result := r.db.WithContext(ctx).
    Model(&model.Todo{}).
    Where("user_id = ? AND deleted_at IS NULL", userID).
    Updates(map[string]any{...})
```

**Sau:**
```go
result := extractDB(ctx, r.db).WithContext(ctx).
    Model(&model.Todo{}).
    Where("user_id = ? AND deleted_at IS NULL", userID).
    Updates(map[string]any{...})
```

**Tại sao quan trọng:** `extractDB` lấy `*gorm.DB` từ context key `txKey{}` nếu có transaction đang chạy. Không có nó, hai operation chạy trong 2 connection riêng biệt → không atomic.

---

### Fix #2 — P0: Authorization Bypass trong createTodo

**File:** `services/todo-bff/internal/handler/graph/resolver/todo.resolvers.go`

**Vấn đề:**  
Resolver `createTodo` nhận `input.UserID` từ GraphQL input — bất kỳ user đã đăng nhập nào cũng có thể tạo todo cho user khác bằng cách truyền `userID` khác vào input.

**Trước:**
```go
todo, err := r.creater.Create(ctx, &ucin.CreateTodo{
    UserID: int64(input.UserID),  // client-controlled!
    ...
})
```

**Sau:**
```go
callerID, _ := middleware.GetUserID(ctx)
todo, err := r.creater.Create(ctx, &ucin.CreateTodo{
    UserID: callerID,  // từ JWT token đã verify
    ...
})
```

---

### Fix #3 — P0: Authorization Bypass trong todos Query

**File:** `services/todo-bff/internal/handler/graph/resolver/todo.resolvers.go`

**Vấn đề:**  
Query `todos` nhận `userID` như parameter — bất kỳ user nào cũng có thể liệt kê todos của user khác.

**Sau:**
```go
callerID, _ := middleware.GetUserID(ctx)
role, _ := middleware.GetRole(ctx)
if role != "ADMIN" {
    userID = int(callerID)  // non-ADMIN chỉ thấy todos của mình
}
```

Đồng thời thêm pagination cap để tránh dump toàn bộ data:
```go
const maxPageSize = 100
if ps > maxPageSize {
    ps = maxPageSize
}
```

---

### Fix #4 — P1: JWT Secret Validation

**Files:**  
- `services/users/internal/config/config.go`  
- `services/todo-bff/internal/config/config.go`

**Vấn đề:**  
Default JWT_SECRET là `"chien-apvn"` (10 ký tự). HS256 cần secret ≥256 bit (32 bytes) để an toàn. Secret ngắn dễ bị brute-force offline.

**Sau:**
```go
const minJWTSecretLen = 32

cfg := &Config{
    JWTSecret: getenv("JWT_SECRET", ""),  // empty default — buộc phải set
    ...
}

if len(cfg.JWTSecret) < minJWTSecretLen {
    return nil, errors.New("config: JWT_SECRET must be at least 32 characters")
}
```

Service sẽ **refuse to start** nếu không có JWT_SECRET đủ mạnh.

---

### Fix #5 — P1: Database Connection Pool

**Files:**  
- `services/todos/internal/infra/datastore/db.go`  
- `services/users/internal/infra/datastore/db.go`

**Vấn đề:**  
GORM mặc định không giới hạn connection pool. Dưới load cao, service mở hàng trăm connection → PostgreSQL default `max_connections = 100` bị exhausted → toàn bộ hệ thống bị từ chối kết nối.

**Sau:**
```go
sqlDB.SetMaxOpenConns(25)           // tối đa 25 connection active
sqlDB.SetMaxIdleConns(5)            // giữ 5 connection idle
sqlDB.SetConnMaxLifetime(5 * time.Minute)   // connection sống tối đa 5 phút
sqlDB.SetConnMaxIdleTime(1 * time.Minute)   // idle connection hết sau 1 phút
```

---

### Fix #6 — P1: HTTP Server Timeouts

**File:** `services/todo-bff/internal/handler/server/server.go`

**Vấn đề:**  
Server chỉ có `ReadHeaderTimeout: 5s`. Thiếu `ReadTimeout` và `WriteTimeout` → dễ bị [Slowloris attack](https://en.wikipedia.org/wiki/Slowloris_(computer_security)) hoặc slow client giữ connection mãi, exhausting goroutines.

**Sau:**
```go
return &nethttp.Server{
    Addr:              ":" + cfg.AppPort,
    Handler:           h,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      30 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```

---

### Fix #7 — P1: Outbox Worker Retry Cap

**File:** `services/users/internal/infra/datastore/outbox_repo.go`

**Vấn đề:**  
`GetUnpublishedEvents` chỉ filter `published_at IS NULL`. Một event với payload lỗi (malformed JSON, RabbitMQ reject) sẽ bị retry **mãi mãi** mỗi 5 giây, làm log spam và tốn CPU.

**Trước:**
```go
Where("published_at IS NULL")
```

**Sau:**
```go
Where("published_at IS NULL AND retry_count < 5")
```

Event sau 5 lần thất bại sẽ bị bỏ qua. Cần thêm alerting/monitoring trên `retry_count >= 5` để ops team biết có event bị stuck.

---

### Fix #8 — P1: gRPC Health Check

**Files:**  
- `services/todos/internal/handler/grpc/server.go`  
- `services/users/internal/handler/grpc/server.go`

**Vấn đề:**  
Không có gRPC health check endpoint. Load balancer, Docker health check, Kubernetes readiness probe không thể biết service có healthy không.

**Sau:**
```go
import (
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)

grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
```

Sau đó có thể dùng `grpc_health_probe` hoặc `grpcurl` để check:
```bash
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

---

### Fix #9 — P1: CI Gate Fix

**File:** `.github/workflows/ci.yml`

**Vấn đề:**  
Job `ci` (Build & Test) chỉ `needs: lint`. Job `integration-test` chạy song song nhưng không block `ci`. Kết quả: integration test fail **không ngăn merge được**.

**Trước:**
```yaml
ci:
  needs: lint
```

**Sau:**
```yaml
ci:
  needs: [lint, integration-test]
```

Pipeline flow mới:
```
lint ──┬──> integration-test ──> ci (Build & Test)
       └──────────────────────────────────────────/
```

---

### Fix #10 — P1: Makefile Targets

**File:** `Makefile`

**Thêm 3 targets:**

```makefile
# Chạy với race detector — phát hiện data race
test-race:
    go test -race ./...

# Chạy với coverage report
test-cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Chạy golangci-lint
lint:
    golangci-lint run ./...
```

`ci` target cũng được cập nhật để bao gồm `lint`:
```makefile
ci: fmt-check vet lint test build
```

---

### Fix #11 — P1: .env.example Hoàn chỉnh

**File:** `.env.example`

**Trước:** Chỉ có 7 biến cơ bản, thiếu `JWT_SECRET`, `RABBITMQ_URL`, `DB_TODO_DSN`, `DB_USERS_DSN`.

**Sau:** Đầy đủ tất cả biến cho cả 3 service:
```env
# Application
APP_NAME=
APP_ENV=development

# BFF
BFF_PORT=8080
TODOS_GRPC_ADDR=localhost:50051
USERS_GRPC_ADDR=localhost:50052
REQUEST_TIMEOUT=5s

# JWT (must be at least 32 characters)
JWT_SECRET=

# Todos service
APP_TODOS_PORT=50051
DB_TODO_DSN=postgres://...

# Users service
APP_USERS_PORT=50052
DB_USERS_DSN=postgres://...

# RabbitMQ
RABBITMQ_URL=amqp://...
```

---

### Fix #12 — P1: docker-compose.yml JWT_SECRET

**File:** `docker-compose.yml`

**Vấn đề:**  
Sau Fix #4 (JWT validation), `users` service sẽ fail to start với `JWT_SECRET: "chien-apvn"` (10 ký tự). `todo-bff` cũng không có `JWT_SECRET` trong environment.

**Sau:**
- `users` service: `JWT_SECRET: "chien-apvn-super-secret-key-2024"` (32 ký tự)
- `todo-bff` service: thêm `JWT_SECRET: "chien-apvn-super-secret-key-2024"`

> ⚠️ **Production note:** Đây là secret cho development. Trên production phải dùng secret manager (AWS Secrets Manager, HashiCorp Vault, v.v.) và rotate định kỳ.

---

## Những gì chưa làm (P2 / Out of Scope)

| Issue | Lý do chưa làm |
|-------|----------------|
| Refresh token rotation không atomic | Cần refactor transaction scope trong usecase — có thể thay đổi behavior |
| `GetTodo` dùng `FOR UPDATE` cho read-only | Cần test kỹ trước khi đổi sang `FOR SHARE` |
| Không có TTL/cleanup cho `processed_events` / `outbox_events` / `refresh_tokens` | Cần scheduled job hoặc pg_cron — thêm infrastructure |
| x-request-id propagation BFF → gRPC metadata | Cần thêm grpc interceptor + middleware coordination |
| Domain layer import infra/jwt trực tiếp (Clean Architecture violation) | Cần interface extraction — refactor lớn |
| Worker restart không có exponential backoff | Cần refactor worker loop |
| Không có metrics endpoint (Prometheus) | Thêm infrastructure mới |

---

## Cách verify

```bash
# Build tất cả service
make build

# Unit tests
make test

# Unit tests với race detector
make test-race

# Coverage report
make test-cover

# Lint
make lint

# Docker Compose (cần Docker Desktop đang chạy)
docker compose up -d
docker compose ps

# Kiểm tra health gRPC sau khi services up
# grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
# grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check
```

---

## Files đã thay đổi

```
services/todos/internal/infra/datastore/todo_command_repo.go   # Fix #1
services/todo-bff/internal/handler/graph/resolver/todo.resolvers.go  # Fix #2 #3
services/users/internal/config/config.go                       # Fix #4
services/todo-bff/internal/config/config.go                    # Fix #4
services/todos/internal/infra/datastore/db.go                  # Fix #5
services/users/internal/infra/datastore/db.go                  # Fix #5
services/todo-bff/internal/handler/server/server.go            # Fix #6
services/users/internal/infra/datastore/outbox_repo.go         # Fix #7
services/todos/internal/handler/grpc/server.go                 # Fix #8
services/users/internal/handler/grpc/server.go                 # Fix #8
.github/workflows/ci.yml                                       # Fix #9
Makefile                                                        # Fix #10
.env.example                                                    # Fix #11
docker-compose.yml                                             # Fix #12
```
