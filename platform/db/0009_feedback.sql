-- =============================================================================
-- 0009_feedback.sql — 用户意见反馈会话
-- =============================================================================

CREATE TABLE IF NOT EXISTS feedback_threads (
  id BIGINT NOT NULL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'open',
  user_reply_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_message_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_feedback_threads_user_last ON feedback_threads (user_id, last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_threads_last ON feedback_threads (last_message_at DESC);

CREATE TABLE IF NOT EXISTS feedback_messages (
  id BIGINT NOT NULL PRIMARY KEY,
  feedback_id BIGINT NOT NULL,
  sender_id BIGINT NOT NULL,
  sender_type VARCHAR(16) NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_feedback_messages_feedback_created ON feedback_messages (feedback_id, created_at ASC);
