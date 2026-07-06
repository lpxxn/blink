-- =============================================================================
-- 0012_reply_likes.sql — 评论点赞
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL.
-- 主键 (user_id, reply_id)：同一用户对同一评论至多一行。
-- 取消点赞若用软删，再次点赞应对同一行 UPDATE 清空 deleted_at，勿重复 INSERT。
-- =============================================================================

CREATE TABLE IF NOT EXISTS reply_likes (
  user_id BIGINT NOT NULL,
  reply_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (user_id, reply_id)
);

CREATE INDEX IF NOT EXISTS idx_reply_likes_reply_id ON reply_likes (reply_id);
