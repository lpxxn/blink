-- =============================================================================
-- 0008_app_settings.sql — 应用设置（后台可配置）
-- =============================================================================

CREATE TABLE IF NOT EXISTS app_settings (
  key VARCHAR(128) NOT NULL PRIMARY KEY,
  value TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

