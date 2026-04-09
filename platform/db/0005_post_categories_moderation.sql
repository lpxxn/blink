-- =============================================================================
-- 0005_post_categories_moderation.sql — 帖子分类与审核字段
-- =============================================================================
-- Dialect: portable SQL for SQLite, MySQL, and PostgreSQL.
-- 依赖: 0002_create_post.sql（posts）。
-- =============================================================================

-- -----------------------------------------------------------------------------
-- categories — 帖子主题分类（由应用种子写入 ID）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
  id BIGINT NOT NULL PRIMARY KEY,
  slug VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug ON categories (slug);
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories (sort_order, id);

-- -----------------------------------------------------------------------------
-- posts — 扩展：分类与审核（含义见 SCHEMA.md）
-- -----------------------------------------------------------------------------
ALTER TABLE posts ADD COLUMN category_id BIGINT NULL;
CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts (category_id);

ALTER TABLE posts ADD COLUMN moderation_flag INTEGER NOT NULL DEFAULT 0;
ALTER TABLE posts ADD COLUMN moderation_note TEXT NOT NULL DEFAULT '';
