-- =============================================================================
-- 0010_admin_audit_logs.sql — 管理后台操作审计
-- =============================================================================

CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id BIGINT NOT NULL PRIMARY KEY,
  actor_id BIGINT NOT NULL,
  action VARCHAR(64) NOT NULL,
  target_type VARCHAR(32) NOT NULL DEFAULT '',
  target_id BIGINT NULL,
  detail TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_created ON admin_audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_actor ON admin_audit_logs (actor_id, created_at DESC);
