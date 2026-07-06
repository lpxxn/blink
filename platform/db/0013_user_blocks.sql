-- =============================================================================
-- 0013_user_blocks.sql — 用户拉黑
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL.
-- blocker_id 拉黑 blocked_id；主键 (blocker_id, blocked_id)。
-- 取消拉黑若用软删，再次拉黑应对同一行 UPDATE 清空 deleted_at，勿重复 INSERT。
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_blocks (
  blocker_id BIGINT NOT NULL,
  blocked_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (blocker_id, blocked_id)
);

CREATE INDEX IF NOT EXISTS idx_user_blocks_blocked_id ON user_blocks (blocked_id);
