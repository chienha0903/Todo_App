-- =============================================================================
-- Migration 001: Create todos table + index on user_id
-- Database : todo_db (PostgreSQL 16)
-- Author   : chienha0903
-- Date     : 2026-05-16
-- =============================================================================

-- ========================= UP (chạy khi migrate) =============================

BEGIN;

-- 1. Tạo bảng todos
CREATE TABLE IF NOT EXISTS todos (
    id          BIGSERIAL       PRIMARY KEY,
    user_id     BIGINT          NOT NULL,
    title       VARCHAR(255)    NOT NULL,
    description TEXT,
    status      VARCHAR(50)     NOT NULL DEFAULT 'pending',
    priority    VARCHAR(50)     NOT NULL DEFAULT 'medium',
    due_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 2. Thêm index trên user_id (query chính: lấy todos theo user)
CREATE INDEX IF NOT EXISTS idx_todos_user_id
    ON todos (user_id);

-- 3. Index kết hợp: lọc theo user + sắp xếp mới nhất trước (dùng trong GetTodos)
CREATE INDEX IF NOT EXISTS idx_todos_user_id_created_at
    ON todos (user_id, created_at DESC);

-- 4. Index kết hợp: lọc theo user + status (dùng khi filter theo trạng thái)
CREATE INDEX IF NOT EXISTS idx_todos_user_id_status
    ON todos (user_id, status);

COMMIT;


-- ========================= VERIFY (kiểm tra sau migrate) =====================

-- Xem cấu trúc bảng
-- \d todos

-- Xem danh sách index đã tạo
-- SELECT indexname, indexdef
-- FROM   pg_indexes
-- WHERE  tablename = 'todos'
-- ORDER  BY indexname;


-- ========================= EXPLAIN ANALYZE (đo hiệu năng) ====================

-- Phân tích query lấy tất cả todos của user (dùng trong GetTodos)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id, user_id, title, status, priority, created_at
FROM   todos
WHERE  user_id = 1
ORDER  BY created_at DESC
LIMIT  20 OFFSET 0;

-- Phân tích query COUNT (dùng để tính total page)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT COUNT(*)
FROM   todos
WHERE  user_id = 1;

-- Phân tích query lấy single todo (dùng trong GetTodo)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT *
FROM   todos
WHERE  id = 1;

-- Phân tích query lọc theo user + status
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id, title, status, priority
FROM   todos
WHERE  user_id = 1
  AND  status = 'pending';


-- ========================= DOWN (rollback nếu cần) ===========================

-- DROP INDEX IF EXISTS idx_todos_user_id_status;
-- DROP INDEX IF EXISTS idx_todos_user_id_created_at;
-- DROP INDEX IF EXISTS idx_todos_user_id;
-- DROP TABLE IF EXISTS todos;
