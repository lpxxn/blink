-- =============================================================================
-- 0001_init.sql — users, sessions
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL (no vendor-only DDL).
-- 不在库层声明 FOREIGN KEY；引用关系由 repository 保证。
-- Run order: before 0002_create_post.sql (posts reference users).
-- =============================================================================

-- -----------------------------------------------------------------------------
-- users — 账户与登录资料（主键为应用下发的 snowflake）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
  -- 分布式唯一 ID，由应用生成；同时作为主键。
  snowflake_id BIGINT NOT NULL PRIMARY KEY,
  -- 登录邮箱；全表唯一（含已软删行时若需复用邮箱，需在业务层处理）。
  email VARCHAR(255) NOT NULL,
  -- 展示名 / 昵称。
  name VARCHAR(128) NOT NULL DEFAULT '',
  -- 微信侧标识（openid、unionid 或业务自定义 ID，按对接方案填写）。
  wechat_id VARCHAR(128) NOT NULL DEFAULT '',
  -- 手机号，建议存 E.164 或统一清洗后的字符串。
  phone VARCHAR(32) NOT NULL DEFAULT '',
  -- 密码摘要（如 bcrypt/argon2 输出）；算法与版本由应用约定。
  password_hash VARCHAR(255) NOT NULL,
  -- 独立盐；若算法把盐嵌入 hash，可留空字符串。
  password_salt VARCHAR(255) NOT NULL DEFAULT '',
  -- 最近一次登录成功时间。
  last_login_at TIMESTAMP NULL,
  -- 最近一次登录 IP（IPv6 最长 45 字符）。
  last_login_ip VARCHAR(45) NULL,
  -- 最近一次登录设备描述（如 UA 截断、客户端上报的 device id）。
  last_login_device VARCHAR(512) NULL,
  -- 账号状态：见 platform/db/SCHEMA.md「users.status」。
  status INTEGER NOT NULL DEFAULT 1,
  -- 角色：见 SCHEMA.md「users.role」；单角色字符串，多角色可后续拆关联表。
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  -- 创建时间。
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- 最后更新时间（应用写入；库层不做 ON UPDATE 以保证多库一致）。
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- 软删时间；非 NULL 表示已删除。
  deleted_at TIMESTAMP NULL
);

-- 邮箱唯一，用于注册与登录查重。
CREATE UNIQUE INDEX idx_users_email ON users (email);

-- -----------------------------------------------------------------------------
-- sessions — 登录会话（opaque id，可放 cookie / header）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sessions (
  -- 会话令牌 ID，不透明字符串，由应用生成（建议高熵、足够长）。
  id VARCHAR(128) NOT NULL PRIMARY KEY,
  -- 所属用户，对应 users.snowflake_id。
  user_id BIGINT NOT NULL,
  -- 过期时间；过期会话可由定时任务清理或校验时拒绝。
  expires_at TIMESTAMP NOT NULL,
  -- 创建会话时的客户端 IP。
  ip_address VARCHAR(45) NULL,
  -- 创建会话时的 User-Agent 或等价字符串。
  user_agent VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

-- 按用户列会话、按过期时间清理或扫描。
CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
