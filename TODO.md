# TODO — Todo App

## Cách đọc

- `[x]` hoàn thành
- `[-]` đang làm / làm được một phần
- `[ ]` chưa làm
- Lịch làm việc: **Thứ 3 và Thứ 6** hàng tuần

---

## Target cuối tháng 5 (31/05/2026)

- BFF layer hoàn chỉnh: middleware logging + recovery, chạy được qua Docker Compose
- Database layer vững: có migration script, có index đúng chỗ, hiểu được query plan
- Error handling + context propagation đi xuyên suốt từ handler xuống repo
- Có 1 design proposal cho async pattern áp dụng vào Todo App
- Toàn bộ manual test được document lại với payload mẫu
- **Core User service hoàn thiện end-to-end: proto → domain → repo → handler → test**

---

## Lịch làm việc

### Buổi 1 — Thứ 3, 12/05 · BFF hoàn thiện
- [x] BFF: thêm logging middleware + panic recovery middleware
- [x] Cập nhật `docker-compose.yml` chạy thêm container BFF
- [x] Cập nhật README phần run/test cho BFF

### Buổi 2 — Thứ 6, 15/05 · Test + Database [-]
- [ ] Viết manual test checklist: create → get → list → update → delete → error cases
- [ ] Lưu sample payload + expected response cho từng API
- [ ] Viết SQL migration script cho bảng `todos`, thêm index `user_id`, chạy `EXPLAIN`

### Buổi 3 — Thứ 3, 19/05 · DB sâu + Error/Context
- [ ] Thêm transaction vào 1 use case cụ thể trong project
- [ ] Kiểm tra và fix các chỗ chưa tận dụng context cancel/timeout
- [ ] Viết guideline phân biệt lỗi trả ra client vs chỉ log nội bộ

### Buổi 4 — Thứ 6, 22/05 · Async + Wrap up
- [ ] Viết design proposal async: chọn 1 bài toán trong Todo App, so sánh sync vs queue vs job
- [ ] Tạo bộ request mẫu đầy đủ cho tất cả API (Postman collection hoặc grpcurl script)
- [ ] Rà soát dependency direction, fix các chỗ còn coupling ngược

### Buổi 5 — Thứ 3, 26/05 · User Service — Foundation
- [ ] Định nghĩa `user.proto`: CreateUser, GetUser, UpdateUser, DeleteUser
- [ ] Domain entity User + Value Objects (email, username, password hash, role)
- [ ] Input/Output DTOs + 3 lớp mapper (gRPC ↔ DTO, Entity ↔ DB Model, Entity ↔ Output)
- [ ] Command/Query repository + SQL migration script cho bảng `users`

### Buổi 6 — Thứ 6, 29/05 · User Service — Hoàn thiện
- [ ] Domain service: UserCreater, UserGetter, UserUpdater, UserDeleter
- [ ] gRPC handler implement đủ 4 RPC + error mapping
- [ ] Wire DI: kết nối toàn bộ dependency chain cho User service
- [ ] Unit test với mockgen + table-driven test
- [ ] BFF: thêm HTTP endpoint cho User + cập nhật README

---

## Đã hoàn thành (tham chiếu)

### Foundation & Infrastructure
- [x] Khởi tạo Go module, cấu trúc project
- [x] Config load từ env + hỗ trợ `.env`
- [x] `docker-compose.yml` với Postgres + healthcheck
- [x] Dockerfile multi-stage cho service todos
- [x] Chạy được local cả Docker và `make run`
- [x] README mô tả chức năng và cách chạy

### Core Todo Service
- [x] Proto định nghĩa full CRUD (Create, Get, List, Update, Delete)
- [x] Handler implement đủ 5 RPC
- [x] Domain entity + Value Objects (title, description, status, priority, due_date)
- [x] Input/Output DTOs tách biệt theo từng use-case
- [x] 3 lớp mapper (gRPC ↔ DTO, Entity ↔ DB Model, Entity ↔ Output)
- [x] Command/Query repository separation
- [x] Error mapping: domain error → gRPC status code
- [x] gRPC interceptor: Recovery + Logging
- [x] Google Wire DI hoàn chỉnh
- [x] Unit test value objects + domain service (mockgen + table-driven test)
- [x] Constructor service trả về struct thay vì interface
- [x] Unwrap error chuẩn Go (`errors.Is`, `errors.As`)

### BFF
- [x] Cấu trúc service BFF: cmd, config, handler, gRPC client
- [x] Implement đủ 5 HTTP endpoint
- [x] HTTP status mapping từ gRPC code
- [x] Timeout per request với context.WithTimeout
- [x] GraphQL resolver cơ bản
