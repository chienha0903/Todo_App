ALTER TABLE todos ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_todos_deleted_at ON todos (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_todos_user_id_not_deleted ON todos (user_id) WHERE deleted_at IS NULL;
