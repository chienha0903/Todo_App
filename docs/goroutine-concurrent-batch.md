# Goroutine: Concurrent Batch Processing trong Outbox Worker

## Mục lục
1. [Vấn đề cụ thể trong project](#1-vấn-đề-cụ-thể-trong-project)
2. [Phân tích tại sao nên dùng goroutine](#2-phân-tích-tại-sao-nên-dùng-goroutine)
3. [Kiến thức nền tảng](#3-kiến-thức-nền-tảng)
4. [Hướng dẫn implement từng bước](#4-hướng-dẫn-implement-từng-bước)
5. [Code hoàn chỉnh và giải thích](#5-code-hoàn-chỉnh-và-giải-thích)
6. [Trade-offs và các lưu ý quan trọng](#6-trade-offs-và-các-lưu-ý-quan-trọng)
7. [Các pattern goroutine phổ biến khác](#7-các-pattern-goroutine-phổ-biến-khác)

---

## 1. Vấn đề cụ thể trong project

**File:** `services/users/cmd/worker/main.go`

### Code gốc (sequential)

```go
ticker := time.NewTicker(5 * time.Second)

// Initial run — chạy tuần tự
if err := publisher.PublishBatch(ctx); err != nil {   // bước 1, chờ xong
    slog.Error("initial publish batch failed", "error", err)
}
if err := invalidator.ProcessBatch(ctx); err != nil { // bước 2, mới bắt đầu
    slog.Error("initial cache invalidation batch failed", "error", err)
}

for {
    select {
    case <-ctx.Done():
        return nil
    case <-ticker.C:
        if err := publisher.PublishBatch(ctx); err != nil {   // bước 1
            slog.Error("publish batch failed", "error", err)
        }
        if err := invalidator.ProcessBatch(ctx); err != nil { // bước 2
            slog.Error("cache invalidation batch failed", "error", err)
        }
    }
}
```

### Vấn đề là gì?

Giả sử mỗi hàm mất ~300ms (network I/O):

```
Ticker fires tại T=0
├── PublishBatch:  T=0   → T=300ms  (đợi RabbitMQ + DB)
└── ProcessBatch:         T=300ms → T=600ms (đợi Redis + DB)
Tick xong tại: T=600ms
```

Hai hàm này **hoàn toàn độc lập** về dữ liệu:

| | `PublishBatch` | `ProcessBatch` |
|-|----------------|----------------|
| Query outbox | `GetUnpublishedEvents()` (no filter) | `GetUnpublishedEventsByAggregateType("cache", ...)` |
| Target | RabbitMQ exchange | Redis token store |
| DB rows | `aggregate_type != 'cache'` | `aggregate_type = 'cache'` |
| Conflict | Không | Không |

Vì vậy hoàn toàn có thể chạy song song.

---

## 2. Phân tích tại sao nên dùng goroutine

### 2.1 Đây là I/O-bound workload, không phải CPU-bound

```
PublishBatch flow:
  [Go code] → [DB query 5-50ms] → [JSON marshal <1ms] → [RabbitMQ publish 5-20ms] × N events
                  ↑ blocking wait              ↑ blocking wait

ProcessBatch flow:
  [Go code] → [DB query 5-50ms] → [Redis DEL 1-5ms] × M events
                  ↑ blocking wait    ↑ blocking wait
```

Trong I/O-bound workload, goroutine đặc biệt hiệu quả vì:
- Khi goroutine đang chờ I/O (blocking syscall), **Go runtime scheduler** tự động switch sang goroutine khác
- Không lãng phí CPU cycles trong lúc chờ network

### 2.2 Không có shared mutable state

Đây là điều kiện tiên quyết để chạy song song an toàn:

```
PublishBatch đọc/ghi:
  - outbox_events WHERE status='pending' AND aggregate_type != 'cache'  (riêng)
  - rabbitmq channel                                                     (riêng)

ProcessBatch đọc/ghi:
  - outbox_events WHERE aggregate_type = 'cache'                        (riêng)
  - redis client                                                         (riêng)
```

Không có row nào bị cả hai hàm cùng xử lý → không cần lock, không có race condition.

### 2.3 Lợi ích về thời gian

```
Sequential (hiện tại):
  T=0 ─────────────────────────────────────────── T=600ms
       [   PublishBatch 300ms   ][  ProcessBatch 300ms  ]

Concurrent (sau khi optimize):
  T=0 ─────────────────── T=300ms
       [   PublishBatch   ]
       [   ProcessBatch   ]
```

Thời gian giảm từ `T(pub) + T(proc)` xuống còn `max(T(pub), T(proc))`.
Với ticker 5s, mỗi tick tiết kiệm ~300ms — batch có thể chạy nhiều lần hơn mà không ảnh hưởng throughput.

---

## 3. Kiến thức nền tảng

### 3.1 Goroutine là gì?

Goroutine là **lightweight thread** được quản lý bởi Go runtime (không phải OS thread):

```go
go func() {       // tạo goroutine mới, return ngay
    doWork()      // chạy concurrently
}()
// tiếp tục chạy ngay, không chờ doWork xong
```

| | OS Thread | Goroutine |
|-|-----------|-----------|
| Stack size ban đầu | 1-8 MB | 2-8 KB |
| Tạo mới | ~1ms | ~1µs |
| Switch cost | Kernel context switch | User-space, cực nhanh |
| Số lượng thực tế | Hàng nghìn | Hàng triệu |

### 3.2 `sync.WaitGroup` — đồng bộ goroutines

`WaitGroup` là cơ chế để goroutine cha **chờ** các goroutine con hoàn thành:

```go
var wg sync.WaitGroup

wg.Add(2)          // báo sẽ có 2 goroutines cần chờ

go func() {
    defer wg.Done()  // báo goroutine này xong, giảm counter xuống 1
    doTaskA()
}()

go func() {
    defer wg.Done()  // báo goroutine này xong, giảm counter xuống 0
    doTaskB()
}()

wg.Wait()           // block cho đến khi counter = 0 (cả 2 xong)
```

**Quy tắc quan trọng:**
- `Add(n)` phải gọi **trước** khi `go` spawns goroutine
- `Done()` phải luôn được gọi, ngay cả khi có lỗi → dùng `defer`
- Không tái sử dụng WaitGroup khi các goroutines chưa Done hết

### 3.3 Tại sao KHÔNG dùng `errgroup.WithContext` ở đây?

`errgroup.WithContext` cancel context khi **bất kỳ** goroutine nào return error:

```go
// KHÔNG DÙNG cái này cho use case này:
g, gCtx := errgroup.WithContext(ctx)
g.Go(func() error { return publisher.PublishBatch(gCtx) })
g.Go(func() error { return invalidator.ProcessBatch(gCtx) })
g.Wait()

// Vấn đề: nếu PublishBatch lỗi (RabbitMQ down),
// context bị cancel → ProcessBatch bị interrupt dù Redis vẫn hoạt động tốt
```

Hai công việc này **độc lập về lỗi** — một cái fail không nên ảnh hưởng cái kia.
Dùng `sync.WaitGroup` + error handling riêng là đúng hơn.

### 3.4 Context và Cancellation

```go
// Context được pass từ signal.NotifyContext (nhận SIGTERM/Interrupt)
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
```

Khi process nhận SIGTERM, `ctx.Done()` được close. Cả `PublishBatch` và `ProcessBatch` đều nhận cùng context → cả hai tự dừng gracefully khi shutdown.

```go
// Bên trong PublishBatch / ProcessBatch thường có:
select {
case <-ctx.Done():
    return ctx.Err()   // graceful stop
// hoặc operations nhận ctx và tự cancel
}
```

### 3.5 Goroutine leak — lỗi phổ biến cần tránh

```go
// BUG: goroutine leak — goroutine chạy mãi không dừng
go func() {
    for {
        doWork()      // nếu ctx không được check, goroutine không bao giờ dừng
    }
}()

// ĐÚNG: goroutine tôn trọng context cancellation
go func() {
    defer wg.Done()
    if err := publisher.PublishBatch(ctx); err != nil {  // PublishBatch check ctx nội bộ
        slog.Error(...)
    }
    // goroutine kết thúc tự nhiên sau khi PublishBatch return
}()
```

Trong pattern này goroutine sẽ kết thúc sau khi `PublishBatch` return → không có leak.

### 3.6 `defer wg.Done()` — tại sao quan trọng?

```go
// KHÔNG AN TOÀN:
go func() {
    err := doWork()
    if err != nil {
        return              // wg.Done() không được gọi → wg.Wait() block mãi
    }
    wg.Done()
}()

// AN TOÀN:
go func() {
    defer wg.Done()        // luôn gọi dù return bất kỳ đâu, dù panic
    err := doWork()
    if err != nil {
        slog.Error(...)
        return             // wg.Done() vẫn được gọi qua defer
    }
}()
```

---

## 4. Hướng dẫn implement từng bước

### Bước 1: Xác định các hàm độc lập

Kiểm tra 3 điều kiện:
- [ ] Chúng có đọc/ghi cùng data không? → **Không** (khác aggregate_type)
- [ ] Output của cái này có phải input của cái kia không? → **Không**
- [ ] Thứ tự thực thi có quan trọng không? → **Không**

Nếu cả 3 đều KHÔNG → an toàn để chạy song song.

### Bước 2: Tạo helper function

Extract logic concurrent vào một hàm riêng để có thể tái sử dụng:

```go
// runBatch chạy publisher và invalidator đồng thời.
// Mỗi lỗi được log độc lập, không ảnh hưởng nhau.
func runBatch(ctx context.Context, publisher *worker.OutboxPublisher, invalidator *worker.CacheInvalidator) {
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        if err := publisher.PublishBatch(ctx); err != nil {
            slog.Error("publish batch failed", "error", err)
        }
    }()

    go func() {
        defer wg.Done()
        if err := invalidator.ProcessBatch(ctx); err != nil {
            slog.Error("cache invalidation batch failed", "error", err)
        }
    }()

    wg.Wait()
}
```

### Bước 3: Thêm import `sync`

```go
import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "os/signal"
    "sync"          // ← thêm dòng này
    "syscall"
    "time"
    // ... các import khác
)
```

### Bước 4: Thay thế sequential calls bằng `runBatch`

```go
// Xóa cái này:
if err := publisher.PublishBatch(ctx); err != nil {
    slog.Error("initial publish batch failed", "error", err)
}
if err := invalidator.ProcessBatch(ctx); err != nil {
    slog.Error("initial cache invalidation batch failed", "error", err)
}

// Thay bằng:
runBatch(ctx, publisher, invalidator)
```

```go
// Trong loop, xóa cái này:
case <-ticker.C:
    if err := publisher.PublishBatch(ctx); err != nil {
        slog.Error("publish batch failed", "error", err)
    }
    if err := invalidator.ProcessBatch(ctx); err != nil {
        slog.Error("cache invalidation batch failed", "error", err)
    }

// Thay bằng:
case <-ticker.C:
    runBatch(ctx, publisher, invalidator)
```

### Bước 5: Verify không có race condition

Chạy Go race detector:

```bash
go run -race ./services/users/cmd/worker/main.go
```

Hoặc build với race flag:

```bash
go build -race -o bin/users-worker ./services/users/cmd/worker/
```

Nếu có race condition, race detector sẽ log ra ngay khi chạy.

---

## 5. Code hoàn chỉnh và giải thích

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/chienha0903/Todo_App/pkg/rabbitmq"
    "github.com/chienha0903/Todo_App/services/users/internal/config"
    "github.com/chienha0903/Todo_App/services/users/internal/infra/datastore"
    redisstore "github.com/chienha0903/Todo_App/services/users/internal/infra/redis"
    "github.com/chienha0903/Todo_App/services/users/internal/worker"
)

func main() {
    slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
    if err := run(); err != nil {
        slog.Error("worker failed", "error", err)
        os.Exit(1)
    }
}

func run() error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // ... setup DB, RabbitMQ, Redis (không đổi) ...

    publisher := worker.NewOutboxPublisher(outboxQueryGW, outboxCmdGW, conn)
    invalidator := worker.NewCacheInvalidator(outboxQueryGW, outboxCmdGW, tokenCmdGW)

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    slog.Info("outbox worker started", "poll_interval", "5s", "mode", "concurrent")

    // Initial run: cả hai chạy song song ngay khi start
    runBatch(ctx, publisher, invalidator)

    for {
        select {
        case <-ctx.Done():
            slog.Info("outbox worker shutting down")
            return nil
        case <-ticker.C:
            // Mỗi tick: cả hai batch chạy song song
            runBatch(ctx, publisher, invalidator)
        }
    }
}

// runBatch chạy PublishBatch và ProcessBatch đồng thời.
// Hai operations này độc lập hoàn toàn (khác aggregate_type trong outbox table)
// nên safe để chạy song song. Mỗi lỗi được log riêng, không cancel nhau.
func runBatch(ctx context.Context, publisher *worker.OutboxPublisher, invalidator *worker.CacheInvalidator) {
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        if err := publisher.PublishBatch(ctx); err != nil {
            slog.Error("publish batch failed", "error", err)
        }
    }()

    go func() {
        defer wg.Done()
        if err := invalidator.ProcessBatch(ctx); err != nil {
            slog.Error("cache invalidation batch failed", "error", err)
        }
    }()

    wg.Wait() // block cho đến khi cả hai xong
}
```

### Execution flow sau khi optimize

```
Ticker fires tại T=0
├── goroutine 1: PublishBatch(ctx)
│     ├── DB query: GetUnpublishedEvents()        [5ms]
│     ├── JSON marshal × N                        [<1ms]
│     └── RabbitMQ PublishWithContext × N         [20ms]
│   Total: ~26ms ────────────────────────── DONE
│
└── goroutine 2: ProcessBatch(ctx)
      ├── DB query: GetUnpublishedEventsByType()  [5ms]
      └── Redis DeleteByUserID × M               [3ms]
    Total: ~8ms ──────────────────────────── DONE

wg.Wait() unblocks tại T=26ms (max của hai goroutines)
Tiết kiệm: ~8ms so với ~34ms sequential
```

---

## 6. Trade-offs và các lưu ý quan trọng

### ✅ Lý do approach này tối ưu

| Tiêu chí | Đánh giá |
|----------|----------|
| Độ phức tạp | Thấp — chỉ thêm `sync.WaitGroup` |
| An toàn race condition | Đảm bảo — không share state |
| Graceful shutdown | Đảm bảo — cả hai nhận cùng `ctx` |
| Error handling | Tốt — mỗi lỗi log riêng, không ảnh hưởng nhau |
| Thêm dependency | Không — `sync` là stdlib |

### ⚠️ Khi nào KHÔNG dùng pattern này

**1. Khi hai hàm cùng đọc/ghi cùng rows:**
```go
// NGUY HIỂM nếu không có proper locking:
go func() { repo.GetAndLockRow(id); processA(id) }()
go func() { repo.GetAndLockRow(id); processB(id) }()
// → deadlock hoặc race condition
```

**2. Khi output của A là input của B:**
```go
// SAI khi chạy song song:
go func() { result = stepA() }()   // B cần đợi A xong
go func() { stepB(result) }()      // nhưng chạy song song → result chưa sẵn
wg.Wait()
```

**3. Khi resources bị giới hạn:**
Nếu DB connection pool size = 5 và mỗi hàm cần 3 connections → concurrent sẽ gây timeout.
Cần điều chỉnh `MaxOpenConns` trong DB config nếu tăng concurrency.

### 🔍 Monitoring sau khi deploy

Thêm log để quan sát hiệu quả:

```go
func runBatch(ctx context.Context, ...) {
    start := time.Now()
    var wg sync.WaitGroup
    wg.Add(2)
    // ... goroutines ...
    wg.Wait()
    slog.Info("batch completed", "duration_ms", time.Since(start).Milliseconds())
}
```

---

## 7. Các pattern goroutine phổ biến khác

### Pattern 1: Fan-out với worker pool (khi có nhiều items)

Dùng khi cần xử lý N items song song với giới hạn concurrency:

```go
const maxWorkers = 5
sem := make(chan struct{}, maxWorkers)  // semaphore

var wg sync.WaitGroup
for _, item := range items {
    item := item  // capture loop variable (Go < 1.22)
    sem <- struct{}{}  // acquire slot
    wg.Add(1)
    go func() {
        defer func() {
            <-sem  // release slot
            wg.Done()
        }()
        process(item)
    }()
}
wg.Wait()
```

### Pattern 2: `errgroup` (khi muốn cancel on first error)

Dùng khi các operations phụ thuộc nhau và muốn dừng nếu bất kỳ cái nào fail:

```go
import "golang.org/x/sync/errgroup"

g, gCtx := errgroup.WithContext(ctx)
g.Go(func() error { return stepA(gCtx) })
g.Go(func() error { return stepB(gCtx) })
if err := g.Wait(); err != nil {
    // một trong hai lỗi, cả hai đã bị cancel
    return err
}
```

### Pattern 3: Pipeline (khi dữ liệu cần chạy qua nhiều stages)

```go
func pipeline(ctx context.Context, source <-chan Item) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for item := range source {
            select {
            case <-ctx.Done(): return
            case out <- process(item):
            }
        }
    }()
    return out
}
```

### Pattern 4: Ticker worker (pattern đang dùng trong project)

```go
ticker := time.NewTicker(interval)
defer ticker.Stop()

for {
    select {
    case <-ctx.Done():
        return nil
    case <-ticker.C:
        runBatch(ctx, ...)  // blocking — đảm bảo không chồng chéo giữa các ticks
    }
}
```

> **Lưu ý:** Trong pattern ticker này, `runBatch` là blocking (có `wg.Wait()`). Điều này đảm bảo nếu một batch mất hơn 5s, tick tiếp theo sẽ bị bỏ qua thay vì chồng chéo — hành vi đúng cho outbox worker.

---

## Tổng kết

| | Trước | Sau |
|-|-------|-----|
| Thời gian mỗi tick | `T(pub) + T(proc)` | `max(T(pub), T(proc))` |
| Độ phức tạp code | Thấp | Thấp (chỉ thêm WaitGroup) |
| Race condition risk | Không | Không |
| Dependency mới | Không | Không |
| Graceful shutdown | Đúng | Đúng |

**Rule of thumb:** Nếu hai hàm (1) không share mutable state, (2) không phụ thuộc output của nhau, (3) đều là I/O-bound → chạy song song với `sync.WaitGroup` là lựa chọn tối ưu nhất.