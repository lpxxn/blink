-- =============================================================================
-- 0011_post_likes.sql — 帖子点赞
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL.
-- 主键 (user_id, post_id)：同一用户对同一帖至多一行。
-- 取消点赞若用软删，再次点赞应对同一行 UPDATE 清空 deleted_at，勿重复 INSERT。
-- =============================================================================

CREATE TABLE IF NOT EXISTS post_likes (
  user_id BIGINT NOT NULL,
  post_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (user_id, post_id)
);

CREATE INDEX IF NOT EXISTS idx_post_likes_post_id ON post_likes (post_id);
