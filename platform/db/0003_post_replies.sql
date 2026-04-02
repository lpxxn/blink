-- =============================================================================
-- 0003_post_replies.sql — 帖子下的回复/评论（支持楼中楼）
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL.
-- 不在库层声明 FOREIGN KEY；引用关系由 repository 保证。
-- 依赖: 0002_create_post.sql（posts）。
--
-- 与 posts.reply_to_post_id / post_type=回复 的区别：
--   post_replies = 挂在某条 post 下的评论流（微博评论、Reddit 评论式）。
--   posts 的回复类型 = 时间线上「一条动态回复另一条动态」（Twitter 回复推文式）。
--   产品可只启用其一，或两者并存，由应用层约定。
-- =============================================================================

-- -----------------------------------------------------------------------------
-- post_replies — 对帖子的回复；可嵌套回复他人的回复
-- -----------------------------------------------------------------------------
-- post_id：始终指向「根帖」posts.id，便于按帖拉取整棵评论树。
-- parent_reply_id：NULL = 直接回复帖子；非 NULL = 回复某条 post_replies（楼中楼）。
-- 业务约束：若 parent_reply_id 非空，父行的 post_id 须与本行 post_id 相同（库外校验）。
CREATE TABLE IF NOT EXISTS post_replies (
  -- 回复 ID，应用下发的 BIGINT（如 snowflake）。
  id BIGINT NOT NULL PRIMARY KEY,
  -- 所属根帖 posts.id（所有层级回复都填同一个 post_id）。
  post_id BIGINT NOT NULL,
  -- 回复者 users.snowflake_id。
  user_id BIGINT NOT NULL,
  -- 若回复另一条回复则填父回复 post_replies.id；若直接回复帖子则为 NULL。
  parent_reply_id BIGINT NULL,
  -- 回复正文。
  body TEXT NOT NULL DEFAULT '',
  -- 展示/审核状态：见 platform/db/SCHEMA.md「post_replies.status」。
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

-- 某帖下回复：等值 post_id 与按时间排序、分页（左前缀覆盖仅 post_id 的查询）。
CREATE INDEX idx_post_replies_post_id_created_at ON post_replies (post_id, created_at);
-- 某条回复下的子回复（楼中楼展开）。
CREATE INDEX idx_post_replies_parent_reply_id ON post_replies (parent_reply_id);
-- 某用户发表的回复。
CREATE INDEX idx_post_replies_user_id ON post_replies (user_id);
