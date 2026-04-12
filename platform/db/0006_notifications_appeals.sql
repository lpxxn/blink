-- =============================================================================
-- 0006_notifications_appeals.sql — 站内消息 + 帖子申诉
-- =============================================================================
-- Dialect: SQLite / MySQL / PostgreSQL portable patterns.
-- =============================================================================

-- posts: 作者申诉 / 申请复核（管理员下架 moderation_removed 后）
ALTER TABLE posts ADD COLUMN appeal_body TEXT NOT NULL DEFAULT '';
ALTER TABLE posts ADD COLUMN appeal_status INTEGER NOT NULL DEFAULT 0;
-- appeal_status: 0 无 1 待处理 3 管理员驳回（可再次发起）

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT NOT NULL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type VARCHAR(32) NOT NULL,
  title VARCHAR(512) NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  ref_post_id BIGINT NULL,
  ref_reply_id BIGINT NULL,
  read_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, read_at);
