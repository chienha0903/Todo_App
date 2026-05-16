# Postman Test Checklist — Todo API

**Endpoint:** `POST http://localhost:8080/graphql` → Body → GraphQL

---

## TC-01: Create Todo

**Query:**
```graphql
mutation CreateTodo($input: CreateTodoInput!) {
  createTodo(input: $input) {
    id userId title description status priority dueDate createdAt updatedAt
  }
}
```

**Variables — đầy đủ:**
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

**Variables — không có dueDate:**
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

- [ ] `status` = `"PENDING"` (mặc định)
- [ ] `id` không rỗng → **lưu lại dùng cho TC tiếp theo**
- [ ] `dueDate` = `null` khi không truyền

---

## TC-02: Get Todo

**Query:**
```graphql
query GetTodo($id: ID!) {
  todo(id: $id) {
    id userId title description status priority dueDate createdAt updatedAt
  }
}
```

**Variables:**
```json
{ "id": "1" }
```

- [ ] Data khớp với TC-01
- [ ] Không có field `errors`

---

## TC-03: List Todos

> Tạo ít nhất 3 todo trước khi chạy.

**Query:**
```graphql
query ListTodos($userId: Int!, $page: Int, $pageSize: Int) {
  todos(userId: $userId, page: $page, pageSize: $pageSize) {
    items { id title status priority }
    total page pageSize hasNext
  }
}
```

| Variables | Expected |
|-----------|----------|
| `{ "userId": 1 }` | `page=1`, `pageSize=20`, `hasNext=false` |
| `{ "userId": 1, "page": 1, "pageSize": 2 }` | 2 items, `hasNext=true` |
| `{ "userId": 1, "page": 2, "pageSize": 2 }` | 1 item, `hasNext=false` |
| `{ "userId": 9999 }` | `items=[]`, `total=0` |

- [ ] `total` đúng với số todos đã tạo
- [ ] `hasNext` đúng theo công thức `page * pageSize < total`
- [ ] User không có todo → `items=[]`, không có `errors`

---

## TC-04: Update Todo

**Query:**
```graphql
mutation UpdateTodo($id: ID!, $input: UpdateTodoInput!) {
  updateTodo(id: $id, input: $input) {
    id title description status priority updatedAt
  }
}
```

**Variables — partial update:**
```json
{
  "id": "1",
  "input": {
    "title": "Hoàn thành báo cáo tháng 5 [REVISED]",
    "status": "IN_PROGRESS"
  }
}
```

**Variables — full update:**
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

- [ ] Field không gửi → giữ nguyên giá trị cũ
- [ ] `updatedAt` mới hơn `createdAt`

---

## TC-05: Delete Todo

**Query:**
```graphql
mutation DeleteTodo($id: ID!) {
  deleteTodo(id: $id)
}
```

**Variables:**
```json
{ "id": "3" }
```

- [ ] Response = `true`
- [ ] Gọi GetTodo với id vừa xoá → `NOT_FOUND`

---

## TC-06: Error Cases

**Query dùng chung:** `CreateTodo` từ TC-01, `GetTodo` từ TC-02.

| Case | Variables | Expected `code` |
|------|-----------|----------------|
| Title rỗng | `"title": ""` | `INVALID_ARGUMENT` |
| Title chỉ spaces | `"title": "   "` | `INVALID_ARGUMENT` |
| Priority không hợp lệ | `"priority": "CRITICAL"` | `GRAPHQL_VALIDATION_FAILED` |
| Thiếu `description` | bỏ field | `GRAPHQL_VALIDATION_FAILED` |
| Get id không tồn tại | `"id": "99999"` | `NOT_FOUND` |
| Update id không tồn tại | `"id": "99999"` | `NOT_FOUND` |
| Delete id không tồn tại | `"id": "99999"` | `NOT_FOUND` |
| ID không phải số | `"id": "abc"` | `INVALID_ARGUMENT` |
| Todos service down | `docker compose stop todos` | `UNAVAILABLE` |

- [ ] HTTP status luôn là `200` — kiểm tra lỗi qua field `errors` trong body
- [ ] BFF không crash khi todos service down (recovery middleware)

---

## Checklist tổng hợp

| TC | Mô tả | Pass? |
|----|-------|-------|
| TC-01a | Create đầy đủ fields | `[ ]` |
| TC-01b | Create không có dueDate | `[ ]` |
| TC-02a | Get todo tồn tại | `[ ]` |
| TC-02b | Get chỉ lấy 3 fields | `[ ]` |
| TC-03a | List mặc định | `[ ]` |
| TC-03b | List page 1/2 | `[ ]` |
| TC-03c | List page 2/2 | `[ ]` |
| TC-03d | List user không có todo | `[ ]` |
| TC-04a | Update partial | `[ ]` |
| TC-04b | Update full | `[ ]` |
| TC-05a | Delete | `[ ]` |
| TC-05b | Get sau Delete | `[ ]` |
| TC-06a | Title rỗng | `[ ]` |
| TC-06b | Title chỉ spaces | `[ ]` |
| TC-06c | Priority không hợp lệ | `[ ]` |
| TC-06d | Thiếu description | `[ ]` |
| TC-06e | Get not found | `[ ]` |
| TC-06f | Update not found | `[ ]` |
| TC-06g | Delete not found | `[ ]` |
| TC-06h | ID không phải số | `[ ]` |
| TC-06i | Todos service down | `[ ]` |
