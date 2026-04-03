-- =============================================================================
-- 0004_oauth_identities.sql — OAuth2 提供方与本地用户绑定
-- =============================================================================
-- 每个 (provider, provider_subject) 唯一对应一个用户；首次授权即「注册」。
-- =============================================================================

CREATE TABLE IF NOT EXISTS oauth_identities (
  snowflake_id BIGINT NOT NULL PRIMARY KEY,
  provider VARCHAR(32) NOT NULL,
  provider_subject VARCHAR(255) NOT NULL,
  user_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX idx_oauth_provider_subject ON oauth_identities (provider, provider_subject);
CREATE INDEX idx_oauth_user_id ON oauth_identities (user_id);
