-- =============================================================================
-- 0002_create_post.sql — 社交图、用户名单、动态（帖）
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL.
-- 不在库层声明 FOREIGN KEY；引用关系由 repository 保证。
-- 依赖: 0001_init.sql（users）。
-- 表顺序: user_follows / friendships / user_lists / user_list_members 先于 posts
--         （posts 可引用 user_lists 与 posts 自身）。
-- =============================================================================

-- -----------------------------------------------------------------------------
-- user_follows — 单向关注（粉丝关系）
-- -----------------------------------------------------------------------------
-- 主键 (follower_id, followee_id)：同一对用户至多一行。
-- 取消关注若用软删，再次关注应 UPDATE 将 deleted_at 置 NULL，勿重复 INSERT。
CREATE TABLE IF NOT EXISTS user_follows (
  -- 关注发起方 users.snowflake_id。
  follower_id BIGINT NOT NULL,
  -- 被关注方 users.snowflake_id。
  followee_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (follower_id, followee_id)
);

-- 「谁关注了我」、粉丝列表按 followee_id 查。
CREATE INDEX idx_user_follows_followee_id ON user_follows (followee_id);

-- -----------------------------------------------------------------------------
-- friendships — 好友（申请 → 接受/拒绝）
-- -----------------------------------------------------------------------------
-- 互关可仅用两条 user_follows 在应用层表达；本表面向「好友申请」语义。
-- 主键 (from_user_id, to_user_id)：拒绝或软删后再次申请应 UPDATE 同行，勿重复 INSERT。
CREATE TABLE IF NOT EXISTS friendships (
  -- 申请人 users.snowflake_id。
  from_user_id BIGINT NOT NULL,
  -- 被申请人 users.snowflake_id。
  to_user_id BIGINT NOT NULL,
  -- 状态：见 SCHEMA.md「friendships.status」。
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (from_user_id, to_user_id)
);

-- 待处理申请：按被申请人 + 状态查询。
CREATE INDEX idx_friendships_to_user_id ON friendships (to_user_id);
CREATE INDEX idx_friendships_status ON friendships (status);

-- -----------------------------------------------------------------------------
-- user_lists — 用户自定义名单（类似 Twitter List / 微博分组）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_lists (
  -- 名单 ID，应用下发的 BIGINT（如 snowflake）。
  id BIGINT NOT NULL PRIMARY KEY,
  -- 名单所有者 users.snowflake_id。
  owner_user_id BIGINT NOT NULL,
  -- 名单名称。
  name VARCHAR(128) NOT NULL DEFAULT '',
  -- 名单简介。
  description VARCHAR(512) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

-- 某用户名下所有名单。
CREATE INDEX idx_user_lists_owner_user_id ON user_lists (owner_user_id);

-- -----------------------------------------------------------------------------
-- user_list_members — 名单内用户
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_list_members (
  -- 所属名单 user_lists.id。
  list_id BIGINT NOT NULL,
  -- 成员用户 users.snowflake_id。
  member_user_id BIGINT NOT NULL,
  -- 加入名单时间。
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (list_id, member_user_id)
);

-- 反查「某用户被加入哪些名单」。
CREATE INDEX idx_user_list_members_member_user_id ON user_list_members (member_user_id);

-- -----------------------------------------------------------------------------
-- posts — 动态：类型、回复链、转发/引用、可见范围、正文与图片
-- -----------------------------------------------------------------------------
-- 枚举含义见 platform/db/SCHEMA.md（post_type、visibility、status）。
CREATE TABLE IF NOT EXISTS posts (
  -- 帖子 ID，应用下发的 BIGINT（如 snowflake）。
  id BIGINT NOT NULL PRIMARY KEY,
  -- 作者 users.snowflake_id。
  user_id BIGINT NOT NULL,
  -- 帖子类型：原创 / 回复 / 转发 / 引用；见 SCHEMA.md「posts.post_type」。
  post_type INTEGER NOT NULL DEFAULT 0,
  -- 回复时的直接父帖 posts.id；非回复类型通常为 NULL。
  reply_to_post_id BIGINT NULL,
  -- 被转发或被引用的原帖 posts.id；纯转发可与 post_type=2 共用。
  referenced_post_id BIGINT NULL,
  -- 可见范围；见 SCHEMA.md「posts.visibility」。
  visibility INTEGER NOT NULL DEFAULT 0,
  -- 当 visibility=list_only 时，指定仅对哪个名单可见 user_lists.id；否则 NULL。
  audience_list_id BIGINT NULL,
  -- 正文（纯转发可为空字符串，由产品规则决定）。
  body TEXT NOT NULL DEFAULT '',
  -- 图片列表：JSON 文本数组，如 ["url-or-key", ...]，形状由应用校验。
  images TEXT NOT NULL DEFAULT '[]',
  -- 发布生命周期：草稿 / 已发 / 隐藏等；见 SCHEMA.md「posts.status」。
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

-- 作者时间线、Feed 过滤。
CREATE INDEX idx_posts_user_id ON posts (user_id);
CREATE INDEX idx_posts_status_created_at ON posts (status, created_at);
CREATE INDEX idx_posts_post_type ON posts (post_type);
CREATE INDEX idx_posts_reply_to_post_id ON posts (reply_to_post_id);
CREATE INDEX idx_posts_referenced_post_id ON posts (referenced_post_id);
CREATE INDEX idx_posts_visibility ON posts (visibility);
-- 按名单拉「仅名单可见」的帖。
CREATE INDEX idx_posts_audience_list_id ON posts (audience_list_id);;
