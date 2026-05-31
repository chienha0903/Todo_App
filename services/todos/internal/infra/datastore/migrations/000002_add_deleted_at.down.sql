DROP INDEX IF EXISTS idx_todos_user_id_not_deleted;
DROP INDEX IF EXISTS idx_todos_deleted_at;

ALTER TABLE todos DROP COLUMN IF EXISTS deleted_at;
