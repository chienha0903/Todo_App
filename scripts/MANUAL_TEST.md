# Postman Test Checklist — Todo API

**Endpoint:** `POST http://localhost:8080/graphql` → Body → GraphQL

---

## Test Case 01: Create Todo

**Query:**
```graphql
mutation CreateTodo($input: CreateTodoInput!) {
  createTodo(input: $input) {
    id userId title description status priority dueDate createdAt updatedAt
  }
}
```

### 01a — Đầy đủ fields

**Variables:**
```json
{
  "input": {
    "userId": 1,
    "title": "Hoàn thành báo cáo tháng 5",
    "description": "Tổng hợp số liệu Q2 và viết nhận xét",
    "priority": "HIGH",
    "dueDate": "2026-05-31T17:00:00Z"
  }
}
```

**Expected Response:**
```json
{
  "data": {
    "createTodo": {
      "id": "1",
      "userId": 1,
      "title": "Hoàn thành báo cáo tháng 5",
      "description": "Tổng hợp số liệu Q2 và viết nhận xét",
      "status": "PENDING",
      "priority": "HIGH",
      "dueDate": "2026-05-31T17:00:00Z",
      "createdAt": "<timestamp>",
      "updatedAt": "<timestamp>"
    }
  }
}
```

- [ ] `status` = `"PENDING"` (mặc định, không set được lúc tạo)
- [ ] `id` không rỗng → **lưu lại dùng cho các test case tiếp theo**
- [ ] `createdAt` và `updatedAt` gần bằng thời điểm gọi API

### 01b — Không có dueDate

**Variables:**
```json
{
  "input": {
    "userId": 1,
    "title": "Nghiên cứu GraphQL",
    "description": "Đọc spec và làm thử project nhỏ",
    "priority": "MEDIUM"
  }
}
```

**Expected Response:**
```json
{
  "data": {
    "createTodo": {
      "id": "2",
      "userId": 1,
      "title": "Nghiên cứu GraphQL",
      "description": "Đọc spec và làm thử project nhỏ",
      "status": "PENDING",
      "priority": "MEDIUM",
      "dueDate": null,
      "createdAt": "<timestamp>",
      "updatedAt": "<timestamp>"
    }
  }
}
```

- [ ] `dueDate` = `null` (không phải string rỗng `""`)

---

## Test Case 02: Get Todo

**Query:**
```graphql
query GetTodo($id: ID!) {
  todo(id: $id) {
    id userId title description status priority dueDate createdAt updatedAt
  }
}
```

### 02a — Todo tồn tại

**Variables:**
```json
{ "id": "1" }
```

**Expected Response:**
```json
{
  "data": {
    "todo": {
      "id": "1",
      "userId": 1,
      "title": "Hoàn thành báo cáo tháng 5",
      "description": "Tổng hợp số liệu Q2 và viết nhận xét",
      "status": "PENDING",
      "priority": "HIGH",
      "dueDate": "2026-05-31T17:00:00Z",
      "createdAt": "<timestamp>",
      "updatedAt": "<timestamp>"
    }
  }
}
```

- [ ] Tất cả field khớp với dữ liệu lúc tạo ở Test Case 01
- [ ] Không có field `errors`

### 02b — Chỉ lấy một số fields

**Query:**
```graphql
query GetTodo($id: ID!) {
  todo(id: $id) {
    id title status
  }
}
```

**Variables:**
```json
{ "id": "1" }
```

**Expected Response:**
```json
{
  "data": {
    "todo": {
      "id": "1",
      "title": "Hoàn thành báo cáo tháng 5",
      "status": "PENDING"
    }
  }
}
```

- [ ] Response chỉ có đúng 3 field, không có `description`, `priority`, v.v.

---

## Test Case 03: List Todos

> Tạo ít nhất 3 todo cho `userId: 1` trước khi chạy.

**Query:**
```graphql
query ListTodos($userId: Int!, $page: Int, $pageSize: Int) {
  todos(userId: $userId, page: $page, pageSize: $pageSize) {
    items { id title status priority }
    total page pageSize hasNext
  }
}
```

### 03a — Mặc định (không truyền page/pageSize)

**Variables:**
```json
{ "userId": 1 }
```

**Expected Response:**
```json
{
  "data": {
    "todos": {
      "items": [
        { "id": "1", "title": "Hoàn thành báo cáo tháng 5", "status": "PENDING", "priority": "HIGH" },
        { "id": "2", "title": "Nghiên cứu GraphQL", "status": "PENDING", "priority": "MEDIUM" },
        { "id": "3", "title": "Todo thứ 3", "status": "PENDING", "priority": "LOW" }
      ],
      "total": 3,
      "page": 1,
      "pageSize": 20,
      "hasNext": false
    }
  }
}
```

- [ ] `page` mặc định = `1`, `pageSize` mặc định = `20`
- [ ] `hasNext` = `false` khi tổng items ≤ pageSize

### 03b — Page 1, lấy 2 item

**Variables:**
```json
{ "userId": 1, "page": 1, "pageSize": 2 }
```

**Expected Response:**
```json
{
  "data": {
    "todos": {
      "items": [
        { "id": "1", "title": "Hoàn thành báo cáo tháng 5", "status": "PENDING", "priority": "HIGH" },
        { "id": "2", "title": "Nghiên cứu GraphQL", "status": "PENDING", "priority": "MEDIUM" }
      ],
      "total": 3,
      "page": 1,
      "pageSize": 2,
      "hasNext": true
    }
  }
}
```

- [ ] Chỉ có 2 items
- [ ] `hasNext` = `true` vì còn item ở trang sau

### 03c — Page 2

**Variables:**
```json
{ "userId": 1, "page": 2, "pageSize": 2 }
```

**Expected Response:**
```json
{
  "data": {
    "todos": {
      "items": [
        { "id": "3", "title": "Todo thứ 3", "status": "PENDING", "priority": "LOW" }
      ],
      "total": 3,
      "page": 2,
      "pageSize": 2,
      "hasNext": false
    }
  }
}
```

- [ ] Chỉ còn 1 item (trang cuối)
- [ ] `hasNext` = `false`

### 03d — User không có todo

**Variables:**
```json
{ "userId": 9999 }
```

**Expected Response:**
```json
{
  "data": {
    "todos": {
      "items": [],
      "total": 0,
      "page": 1,
      "pageSize": 20,
      "hasNext": false
    }
  }
}
```

- [ ] `items` là array rỗng `[]`, không phải `null`
- [ ] Không có field `errors`

---

## Test Case 04: Update Todo

**Query:**
```graphql
mutation UpdateTodo($id: ID!, $input: UpdateTodoInput!) {
  updateTodo(id: $id, input: $input) {
    id title description status priority updatedAt
  }
}
```

### 04a — Partial update (chỉ một số fields)

**Variables:**
```json
{
  "id": "1",
  "input": {
    "title": "Hoàn thành báo cáo tháng 5 [REVISED]",
    "status": "IN_PROGRESS"
  }
}
```

**Expected Response:**
```json
{
  "data": {
    "updateTodo": {
      "id": "1",
      "title": "Hoàn thành báo cáo tháng 5 [REVISED]",
      "description": "Tổng hợp số liệu Q2 và viết nhận xét",
      "status": "IN_PROGRESS",
      "priority": "HIGH",
      "updatedAt": "<timestamp mới hơn createdAt>"
    }
  }
}
```

- [ ] `title` và `status` đã thay đổi
- [ ] `description` và `priority` **không thay đổi** — không gửi thì giữ nguyên
- [ ] `updatedAt` mới hơn `createdAt`

### 04b — Full update (tất cả fields)

**Variables:**
```json
{
  "id": "1",
  "input": {
    "title": "Báo cáo Q2 hoàn chỉnh",
    "description": "Đã bổ sung phần phân tích trend",
    "priority": "MEDIUM",
    "status": "COMPLETED",
    "dueDate": "2026-06-01T09:00:00Z"
  }
}
```

**Expected Response:**
```json
{
  "data": {
    "updateTodo": {
      "id": "1",
      "title": "Báo cáo Q2 hoàn chỉnh",
      "description": "Đã bổ sung phần phân tích trend",
      "status": "COMPLETED",
      "priority": "MEDIUM",
      "updatedAt": "<timestamp>"
    }
  }
}
```

- [ ] Tất cả 5 field đều được cập nhật đúng

---

## Test Case 05: Delete Todo

**Query:**
```graphql
mutation DeleteTodo($id: ID!) {
  deleteTodo(id: $id)
}
```

### 05a — Xoá thành công

**Variables:**
```json
{ "id": "3" }
```

**Expected Response:**
```json
{
  "data": {
    "deleteTodo": true
  }
}
```

- [ ] Response = `true`, không có field `errors`

### 05b — Get sau Delete (xác nhận đã xoá)

Dùng query `GetTodo`, variables: `{ "id": "3" }`

**Expected Response:**
```json
{
  "data": {
    "todo": null
  },
  "errors": [
    {
      "message": "todo not found",
      "extensions": { "code": "NOT_FOUND" }
    }
  ]
}
```

- [ ] `data.todo` = `null`
- [ ] `errors[0].extensions.code` = `"NOT_FOUND"`

---

## Test Case 06: Error Cases

### 06a — Title rỗng

**Variables (dùng query CreateTodo):**
```json
{
  "input": { "userId": 1, "title": "", "description": "Valid", "priority": "LOW" }
}
```

**Expected Response:**
```json
{
  "data": null,
  "errors": [{ "message": "Title cannot be empty", "extensions": { "code": "INVALID_ARGUMENT" } }]
}
```

### 06b — Title chỉ có spaces

**Variables:**
```json
{
  "input": { "userId": 1, "title": "   ", "description": "Valid", "priority": "LOW" }
}
```

**Expected Response:** Giống 06a — `"Title cannot be empty"`

### 06c — Priority không hợp lệ

**Variables:**
```json
{
  "input": { "userId": 1, "title": "Test", "description": "Test", "priority": "CRITICAL" }
}
```

**Expected Response:**
```json
{
  "errors": [{ "message": "CRITICAL is not a valid TodoPriority", "extensions": { "code": "GRAPHQL_VALIDATION_FAILED" } }]
}
```

- [ ] Lỗi bắt ngay ở schema — không có field `data`

### 06d — Thiếu description

**Variables:**
```json
{
  "input": { "userId": 1, "title": "Chỉ có title", "priority": "LOW" }
}
```

**Expected Response:**
```json
{
  "errors": [{ "message": "Field \"CreateTodoInput.description\" of required type \"String!\" was not provided." }]
}
```

### 06e — Get id không tồn tại

**Variables (dùng query GetTodo):** `{ "id": "99999" }`

**Expected Response:**
```json
{
  "data": { "todo": null },
  "errors": [{ "message": "todo not found", "extensions": { "code": "NOT_FOUND" } }]
}
```

### 06f — Update id không tồn tại

**Variables (dùng query UpdateTodo):**
```json
{ "id": "99999", "input": { "title": "Test" } }
```

**Expected Response:**
```json
{
  "data": null,
  "errors": [{ "message": "todo not found", "extensions": { "code": "NOT_FOUND" } }]
}
```

### 06g — Delete id không tồn tại

**Variables (dùng query DeleteTodo):** `{ "id": "99999" }`

**Expected Response:**
```json
{
  "data": null,
  "errors": [{ "message": "todo not found", "extensions": { "code": "NOT_FOUND" } }]
}
```

### 06h — ID không phải số

**Variables (dùng query GetTodo):** `{ "id": "abc" }`

**Expected Response:**
```json
{
  "data": null,
  "errors": [{ "message": "invalid id", "extensions": { "code": "INVALID_ARGUMENT" } }]
}
```

### 06i — Todos service down

```powershell
docker compose stop todos
```

Gửi bất kỳ request nào, ví dụ GetTodo.

**Expected Response (sau ~5 giây):**
```json
{
  "data": null,
  "errors": [{ "message": "service unavailable", "extensions": { "code": "UNAVAILABLE" } }]
}
```

- [ ] BFF không crash (recovery middleware bảo vệ)
- [ ] Khởi động lại: `docker compose start todos`

---

## Checklist tổng hợp

| Test Case | Mô tả | Pass? |
|-----------|-------|-------|
| 01a | Create đầy đủ fields | `[ ]` |
| 01b | Create không có dueDate | `[ ]` |
| 02a | Get todo tồn tại | `[ ]` |
| 02b | Get chỉ lấy một số fields | `[ ]` |
| 03a | List mặc định | `[ ]` |
| 03b | List page 1/2 | `[ ]` |
| 03c | List page 2/2 | `[ ]` |
| 03d | List user không có todo | `[ ]` |
| 04a | Update partial | `[ ]` |
| 04b | Update full | `[ ]` |
| 05a | Delete | `[ ]` |
| 05b | Get sau Delete | `[ ]` |
| 06a | Title rỗng | `[ ]` |
| 06b | Title chỉ spaces | `[ ]` |
| 06c | Priority không hợp lệ | `[ ]` |
| 06d | Thiếu description | `[ ]` |
| 06e | Get not found | `[ ]` |
| 06f | Update not found | `[ ]` |
| 06g | Delete not found | `[ ]` |
| 06h | ID không phải số | `[ ]` |
| 06i | Todos service down | `[ ]` |
