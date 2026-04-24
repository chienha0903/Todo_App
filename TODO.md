# TODO Roadmap Cho Todo App

## Cách đọc file này

- `[x]` đã hoàn thành
- `[-]` đã làm một phần / đang có nền tảng nhưng chưa hoàn thiện
- `[ ]` chưa làm

Mỗi phase gồm:

- `Mục tiêu`
- `Checklist`
- `Deliverable`

---

## Rebaseline Mục tiêu 3 tuần (bắt buộc)


### Tuần 1 (2026-04-27 -> 2026-05-03)
- Chốt contract CRUD + error mapping cơ bản cho gRPC
- Chuẩn hóa README/TODO và checklist test tay
- Khởi tạo skeleton BFF (routing, healthcheck, gọi thử 1 API qua gRPC client)

### Tuần 2 (2026-05-04 -> 2026-05-10)
- Hoàn thiện BFF CRUD flow (HTTP -> BFF -> gRPC todos)
- Bổ sung validate request tại BFF
- Bổ sung logging/interceptor cơ bản
- Hoàn thiện manual test checklist cho cả gRPC và BFF

### Tuần 3 (2026-05-10 -> 2026-05-17)
- Viết unit test cho ValueObject + domain service chính
- Rà soát DB (index, query notes, transaction notes mức cơ bản)
- Refactor các điểm coupling/naming còn lại
- Chuẩn bị demo nội bộ + tài liệu phân tích hệ thống

### Deadline tổng

---

## Phase 0 - Foundation Baseline

**Mục tiêu:** dựng được project chạy end-to-end local, có CRUD cơ bản, có tài liệu run/test đủ để tiếp tục mở rộng.

**Trạng thái:** `Đã hoàn thành phần nền tảng`

### Checklist

- [x] Khởi tạo project Go module
- [x] Tạo cấu trúc service `todos`
- [x] Có `main.go` để boot gRPC server
- [x] Có config load từ env
- [x] Thêm hỗ trợ `.env`
- [x] Tạo `.env.example`
- [x] Tạo `.gitignore`
- [x] Tạo `docker-compose.yml`
- [x] Tạo `services/todos/Dockerfile`
- [x] Chạy được local bằng Docker Compose
- [x] Chạy được local bằng `make run-todos`
- [x] README đã mô tả chức năng, cách chạy và test

### Deliverable

- service chạy ở `localhost:50051`
- postgres chạy ở `localhost:5432`
- project có tài liệu run/test cơ bản

---

## Phase 1 - Architecture and Design

**Mục tiêu:** hiểu và áp dụng clean architecture ở mức đủ rõ để tự phân tích và giải thích luồng code.

**Trạng thái:** `Đã làm được nền tảng`

### 1.1 Clean architecture / layer separation

- [x] Tách `handler`, `domain/service`, `gateway`, `infra`, `config`, `di`
- [x] Có luồng request rõ ràng: `handler -> mapper -> domain service -> gateway -> repo`
- [x] Dùng DI với Wire để khởi tạo dependency
- [x] Tách riêng input/output contracts ở `usecase/todo`
- [-] Rà soát lại dependency direction để tránh phụ thuộc ngược hoàn toàn

### 1.2 Service interface / API design

- [x] Thiết kế interface cho:
  - `TodoCreator`
  - `TodoGetter`
  - `TodoLister`
  - `TodoUpdater`
  - `TodoDeleter`
- [x] Thiết kế gRPC API cho full CRUD
- [x] Mapper input/output giữa proto và application model

### 1.3 Pattern cần học thêm và áp dụng tiếp

- [x] gRPC communication pattern cơ bản
- [ ] gRPC interceptor pattern
- [ ] GraphQL resolver pattern (học khái niệm, chưa cần code ngay)
- [ ] REST API pattern

---

## Phase 2 - API Contract and Integration

**Mục tiêu:** làm chủ phần contract/API, biết sửa proto, generate code, test end-to-end, phân tích lỗi contract.

**Trạng thái:** `Đã có full CRUD cơ bản`

### Checklist

- [x] Mở rộng `todo.proto` từ create-only sang full CRUD
- [x] Regenerate `todo.pb.go`
- [x] Regenerate `todo_grpc.pb.go`
- [x] Handler đã implement đủ 5 RPC
- [x] Mapper đã map đủ request/response
- [x] Test được bằng Postman
- [x] Test được bằng `grpcurl`
- [ ] Tạo bộ request mẫu cho tất cả API
- [ ] Tạo checklist expected response cho từng API
- [ ] Tạo danh sách error cases cho từng API
- [ ] Phân loại gRPC status code đúng hơn thay vì tất cả `InvalidArgument`
- [ ] Cân nhắc đổi timestamp string sang `google.protobuf.Timestamp`
- [ ] Cân nhắc dùng `optional`/field mask cho `UpdateTodoRequest`

### Deliverable

- 1 bộ test manual đầy đủ cho CRUD
- 1 bảng mô tả request/response/error cases

---

## Phase 2.5 - BFF Implementation (thiếu và bắt buộc)

**Mục tiêu:** triển khai lớp BFF để team/client gọi HTTP/REST, BFF chịu trách nhiệm orchestration và gọi gRPC service `todos`.

**Trạng thái:** `Thiếu triển khai thực tế`

### Checklist

- [ ] Chốt scope BFF:
  - endpoint HTTP cần expose
  - mapping HTTP status code
  - chuẩn request/response JSON
- [ ] Hoàn thiện cấu trúc service BFF:
  - `cmd/main.go`
  - config
  - transport handler
  - gRPC client tới `todo.v1.TodoService`
- [ ] Implement endpoint:
  - `POST /todos`
  - `GET /todos/:id`
  - `GET /todos?user_id=...`
  - `PUT /todos/:id`
  - `DELETE /todos/:id`
- [ ] Thêm mapper giữa HTTP DTO <-> gRPC DTO
- [ ] Thêm middleware/interceptor tối thiểu:
  - request logging
  - panic recovery
  - timeout
- [ ] Chuẩn hóa error mapping:
  - gRPC `NotFound` -> HTTP 404
  - gRPC `InvalidArgument` -> HTTP 400
  - gRPC `Internal` -> HTTP 500
- [ ] Cập nhật docker-compose (nếu chạy BFF container riêng)
- [ ] Cập nhật README phần run/test cho BFF
- [ ] Viết test manual:
  - Postman collection cho HTTP BFF
  - kiểm tra BFF gọi đúng gRPC todos

### Deliverable

- BFF chạy được local và gọi thành công full CRUD qua gRPC
- Có tài liệu endpoint HTTP + cách test
- Có demo flow: client -> BFF -> todos gRPC -> postgres

---

## Phase 3 - Database Deep Dive

**Mục tiêu:** không chỉ “kết nối DB chạy được”, mà hiểu sâu cách thiết kế schema, transaction, index và hành vi concurrency của PostgreSQL.

**Trạng thái:** `Đã kết nối DB và CRUD được, phần học sâu còn thiếu`

### 3.1 Schema và datatype

- [x] Có bảng `todos`
- [x] Có mapping repo -> entity
- [ ] Bổ sung migration hoặc SQL script rõ ràng cho schema

### 3.2 Index / query performance

- [ ] Xác định các query chính:
  - `GetTodo(id)`
  - `ListTodos(user_id)`
  - `UpdateTodo(id)`
  - `DeleteTodo(id)`
- [ ] Kiểm tra index hiện có ngoài primary key
- [ ] Thêm index phù hợp, ví dụ:
  - index cho `user_id`
  - cân nhắc composite index nếu có filter nhiều hơn sau này
- [ ] Dùng `EXPLAIN` để hiểu query plan
- [ ] Viết note: “khi nào nên thêm index, tradeoff read/write”

### 3.3 Transaction / lock / consistency

- [ ] Học transaction trong PostgreSQL
- [ ] Học lock row / table ở mức thực hành
- [ ] Thử implement 1 use case có transaction
- [ ] Hiểu và note lại:
  - ACID
  - MVCC
  - isolation level
- [ ] Áp dụng vào một case thực tế trong demo project

### Deliverable

- Schema note + query note + index note

---

## Phase 4 - Concurrency, Context, Error Handling

**Mục tiêu:** hiểu cách Go xử lý context, goroutine, cancellation và cách đưa nó vào server/backend đúng cách.

**Trạng thái:** `Mới ở mức nền tảng`

### 4.1 Context & cancellation

- [-] Đang truyền `context.Context` xuyên suốt handler -> service -> repo
- [ ] Kiểm tra nơi nào thực sự tận dụng cancel/timeout
- [ ] Thêm timeout/timeout config cho DB hoặc request
- [ ] Viết note “context đi từ đâu tới đâu”

### 4.2 Error handling

- [x] Đang có validate từ value object
- [-] Có map lỗi ra gRPC nhưng còn đơn giản
- [ ] Chuẩn hóa domain error -> grpc status mapping:
  - invalid argument
  - not found
  - internal
- [ ] Thêm logging có cấu trúc khi lỗi xảy ra
- [ ] Viết guideline: lỗi nào trả ra client, lỗi nào chỉ log nội bộ

### 4.3 Goroutine / concurrency

- [ ] Học rõ khi nào nên dùng goroutine
- [ ] Tạo 1 use case demo nhỏ có concurrent work hợp lý
- [ ] Tránh dùng goroutine chỉ để “thử cho biết”
- [ ] Biết các rủi ro:
  - data race
  - leak goroutine
  - orphan goroutine khi request cancel

### 4.4 Middleware / interceptor

- [ ] Tìm hiểu gRPC interceptor unary
- [ ] Thêm logging interceptor
- [ ] Thêm request timing interceptor
- [ ] Nếu phù hợp, thêm recovery interceptor

### Deliverable

- 1 flow có timeout/logging rõ
- 1 file note về context + error + interceptor

---

## Phase 5 - Async Pattern

**Mục tiêu:** hiểu khi nào hệ thống nên xử lý đồng bộ, khi nào nên chuyển sang async/event-driven.

**Trạng thái:** `Chưa bắt đầu`

### Checklist

- [ ] Học message queue cơ bản:
  - Kafka
  - RabbitMQ
- [ ] Học event-driven:
  - SNS / SQS hoặc pattern tương đương
- [ ] Học job nền / cronjob
- [ ] Chọn 1 bài toán thực tế trong Todo App để demo async, ví dụ:
  - gửi notification khi todo gần tới hạn
  - đồng bộ audit log
  - daily cleanup/report
- [ ] Viết 1 note so sánh:
  - sync call
  - async queue
  - scheduled job

### Deliverable

- 1 demo async nhỏ hoặc ít nhất 1 design proposal rõ ràng


## Phase 6 - Unit Test, Manual Test, Refactor

**Mục tiêu:** từ code chạy được sang code có khả năng kiểm chứng, biết phân tích case và giảm regression.

**Trạng thái:** `Manual test có nền tảng, unit test gần như chưa có`

### 6.1 Unit test

- [ ] Viết test cho value objects:
  - title
  - description
  - priority
  - status
  - due_date
- [ ] Viết test cho domain service:
  - happy case
  - invalid input
  - not found
  - repo/gateway error
- [ ] Mock command/query gateway
- [ ] Assert đầy đủ input/output/error

### 6.2 Manual test

- [x] Đã có cách test bằng Postman
- [x] Đã có cách test bằng grpcurl
- [ ] Viết manual test checklist theo thứ tự:
  - create
  - get
  - list
  - update
  - delete
  - invalid cases
- [ ] Lưu sample payload và expected result

### Deliverable

- Có test cho layer quan trọng nhất
- Có checklist manual test đủ happy case + error case

---
