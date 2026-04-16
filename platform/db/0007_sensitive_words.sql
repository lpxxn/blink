-- =============================================================================
-- 0007_sensitive_words.sql — 敏感词表（启用行参与内存快照匹配）
-- =============================================================================
-- 依赖: 无。应用层 snowflake id。
-- =============================================================================

CREATE TABLE IF NOT EXISTS sensitive_words (
  id BIGINT NOT NULL PRIMARY KEY,
  word TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sensitive_words_word ON sensitive_words (word);

CREATE INDEX IF NOT EXISTS idx_sensitive_words_enabled ON sensitive_words (enabled);
