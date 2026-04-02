-- =============================================================================
-- 0000_schema_migrations.sql — 迁移执行记录（须最先执行）
-- =============================================================================
-- 每成功执行一个 migration 文件，CLI 会向本表插入一行 version = 文件名。
-- 兼容 SQLite / MySQL / PostgreSQL；不使用 FOREIGN KEY。
-- =============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
  -- 已执行的 migration 文件名，如 0001_init.sql；主键保证不重复执行。
  version VARCHAR(255) NOT NULL PRIMARY KEY,
  -- 应用该 migration 的时间。
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
