# Blink — 数据库 Schema 说明

本文档与 `platform/db/*.sql` 中的注释一致，描述表、字段语义及建议的应用层枚举。  
SQL 需同时兼容 **SQLite / MySQL / PostgreSQL**；业务访问请走 **repository 层**，勿在 handler/service 里拼接 SQL。

## Migration 顺序

| 文件 | 内容 |
|------|------|
| `0000_schema_migrations.sql` | `schema_migrations` 表（记录已执行的 migration 文件名） |
| `0001_init.sql` | `users`、`sessions` |
| `0002_create_post.sql` | `user_follows`、`friendships`、`user_lists`、`user_list_members`、`posts` |
| `0003_post_replies.sql` | `post_replies`（帖子评论 / 楼中楼） |

### CLI：`cmd/migrate`

在项目根目录执行（默认 SQLite 文件 `blink.db`、迁移目录 `platform/db`）：

```bash
go run ./cmd/migrate
```

完整操作说明见 **[cmd/migrate/README.md](../cmd/migrate/README.md)**（SQLite / PostgreSQL、FAQ）；`go run ./cmd/migrate -h` 为简短帮助。

常用参数：

| 参数 | 说明 |
|------|------|
| `-driver` | `sqlite`（默认）、`mysql`、`postgres` |
| `-dsn` | 驱动对应的连接串（SQLite 默认 `file:blink.db?cache=shared&mode=rwc`） |
| `-dir` | 存放 `*.sql` 的目录（默认 `platform/db`） |

行为：按文件名字典序执行尚未出现在 `schema_migrations.version` 中的 `.sql`；每个文件在事务中执行（多语句按分号拆分，忽略注释与字符串内的分号），成功后插入一行 `version = 文件名`。再次执行同一命令会跳过已记录文件。

**MySQL 说明**：部分 DDL 会隐式提交，与 `INSERT schema_migrations` 未必在同一事务内；请保持 migration 可重复执行（如 `CREATE TABLE IF NOT EXISTS`）以便失败重跑。

---

## schema_migrations

由 `0000_schema_migrations.sql` 创建，供迁移 CLI 使用。

| 列 | 类型 | 说明 |
|----|------|------|
| version | VARCHAR(255) PK | 已执行的 migration 文件名，如 `0001_init.sql`。 |
| applied_at | TIMESTAMP | 执行完成时间。 |

## 公共约定

- **命名**：表与列均为 `snake_case`。
- **时间**：`created_at` / `updated_at` 由应用维护写入；`deleted_at` 非 NULL 表示软删。
- **ID**：除 `sessions.id`、`users.snowflake_id` 外，其余 BIGINT 主键及指向他表的 ID 列一般由应用用 snowflake（或同等策略）生成。
- **枚举**：整型或小字符串枚举在库中仅存值；**含义以本文档 + 应用常量为权威**，避免在 SQL 里写不可移植的 ENUM 类型。
- **引用完整性**：DDL 中**不声明** `FOREIGN KEY`。文档中的「引用 xxx」仅描述列的业务含义，插入/更新/删除时由 **repository** 校验目标行存在且满足软删等策略。

---

## users

| 列 | 类型 | 说明 |
|----|------|------|
| snowflake_id | BIGINT PK | 用户主键，应用生成。 |
| email | VARCHAR(255) | 登录邮箱；唯一索引 `idx_users_email`。 |
| name | VARCHAR(128) | 展示名。 |
| wechat_id | VARCHAR(128) | 微信相关标识，按对接方案填。 |
| phone | VARCHAR(32) | 手机号。 |
| password_hash | VARCHAR(255) | 密码摘要。 |
| password_salt | VARCHAR(255) | 盐；若算法内置盐可置空串。 |
| last_login_at | TIMESTAMP | 最近登录时间。 |
| last_login_ip | VARCHAR(45) | 最近登录 IP。 |
| last_login_device | VARCHAR(512) | 最近登录设备/UA 等。 |
| status | INTEGER | 账号状态，见下表。 |
| role | VARCHAR(32) | 角色字符串，见下表。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 审计与软删。 |

### users.status（建议）

| 值 | 含义 |
|----|------|
| 0 | 未激活 / 停用 |
| 1 | 正常 |
| 2 | 封禁 |

### users.role（建议）

在应用内用常量约束，例如：`user`、`admin`。若需多角色，后续可拆 `user_roles` 关联表。

---

## sessions

| 列 | 类型 | 说明 |
|----|------|------|
| id | VARCHAR(128) PK | 会话令牌，不透明、高熵。 |
| user_id | BIGINT，引用 users.snowflake_id | 所属用户。 |
| expires_at | TIMESTAMP | 过期时间。 |
| ip_address | VARCHAR(45) | 创建会话时 IP。 |
| user_agent | VARCHAR(512) | 创建会话时 UA。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 审计与软删/吊销。 |

---

## user_follows

单向关注：`follower_id` 关注 `followee_id`。

| 列 | 类型 | 说明 |
|----|------|------|
| follower_id | BIGINT | 关注者，引用 users.snowflake_id。 |
| followee_id | BIGINT | 被关注者，引用 users.snowflake_id。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 关注时间与软删（取关）。 |

**注意**：主键为 `(follower_id, followee_id)`。若取关使用软删，再次关注应对**同一行** `UPDATE` 清空 `deleted_at`，不要 `INSERT` 新行。

---

## friendships

好友申请与结果；方向为 `from_user_id` → `to_user_id`。

| 列 | 类型 | 说明 |
|----|------|------|
| from_user_id | BIGINT | 申请人。 |
| to_user_id | BIGINT | 被申请人。 |
| status | INTEGER | 见下表。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 时间与软删。 |

### friendships.status（建议）

| 值 | 含义 |
|----|------|
| 0 | 待处理 |
| 1 | 已同意 |
| 2 | 已拒绝 |

再次发起申请时优先复用同一主键行并更新 `status` / `deleted_at`。

---

## user_lists

用户创建的名单（类似 Twitter List）。

| 列 | 类型 | 说明 |
|----|------|------|
| id | BIGINT PK | 名单 ID，应用生成。 |
| owner_user_id | BIGINT | 所有者，引用 users.snowflake_id。 |
| name | VARCHAR(128) | 名称。 |
| description | VARCHAR(512) | 描述。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 审计与软删。 |

---

## user_list_members

名单与成员的关联。

| 列 | 类型 | 说明 |
|----|------|------|
| list_id | BIGINT | 引用 user_lists.id。 |
| member_user_id | BIGINT | 引用 users.snowflake_id。 |
| created_at | TIMESTAMP | 加入时间。 |

主键 `(list_id, member_user_id)`。

---

## posts

| 列 | 类型 | 说明 |
|----|------|------|
| id | BIGINT PK | 帖子 ID，应用生成。 |
| user_id | BIGINT | 作者，引用 users.snowflake_id。 |
| post_type | INTEGER | 类型，见下表。 |
| reply_to_post_id | BIGINT NULL | 回复时的**直接父帖**。 |
| referenced_post_id | BIGINT NULL | 被转发/被引用的**目标帖**。 |
| visibility | INTEGER | 可见范围，见下表。 |
| audience_list_id | BIGINT NULL | `visibility=list_only` 时指向 `user_lists.id`。 |
| body | TEXT | 正文。 |
| images | TEXT | JSON 数组字符串，元素为 URL 或对象存储 key。 |
| status | INTEGER | 发布状态，见下表。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 审计与软删。 |

### posts.post_type（建议）

| 值 | 含义 |
|----|------|
| 0 | 原创 |
| 1 | 回复（通常配合 `reply_to_post_id`） |
| 2 | 转发 / 纯转推（通常配合 `referenced_post_id`） |
| 3 | 引用 / 带评转发（`body` + `referenced_post_id`） |

具体组合由产品在应用层校验（例如纯转推是否允许空 `body`）。

### posts.visibility（建议）

| 值 | 含义 |
|----|------|
| 0 | 公开 |
| 1 | 仅粉丝 |
| 2 | 仅好友（与 `friendships` / 双向关注策略一致，由应用定义） |
| 3 | 仅名单（须设置 `audience_list_id`） |

### posts.status（建议）

| 值 | 含义 |
|----|------|
| 0 | 草稿 |
| 1 | 已发布 |
| 2 | 隐藏（不公开时间线但仍保留数据，可按需调整语义） |

---

## post_replies

挂在 **`posts` 根帖**下的评论；支持对评论再回复（`parent_reply_id` 自关联）。

与 **`posts.reply_to_post_id` + `post_type=1`** 的差异：

| 机制 | 语义 |
|------|------|
| `post_replies` | 评论流：用户对一条动态留言，可嵌套回复。 |
| `posts` 回复类型 | 时间线上独立动态，结构上「回复另一条动态」（类 Twitter）。 |

二者可并存或只实现其一，由产品决定；若并存，请在应用层明确各自入口与展示。

| 列 | 类型 | 说明 |
|----|------|------|
| id | BIGINT PK | 回复 ID，应用生成。 |
| post_id | BIGINT | 根帖 `posts.id`；**每一层**回复都指向同一根帖。 |
| user_id | BIGINT | 作者，引用 users.snowflake_id。 |
| parent_reply_id | BIGINT NULL | 父回复 `post_replies.id`；NULL 表示直接回复帖子。 |
| body | TEXT | 正文。 |
| status | INTEGER | 见下表。 |
| created_at / updated_at / deleted_at | TIMESTAMP | 审计与软删。 |

**业务约束**：当 `parent_reply_id` 非空时，父行的 `post_id` 必须与本行 `post_id` 一致（建议在写入前于 repository/service 校验；跨库可移植的 CHECK 难以统一表达）。

### post_replies.status（建议）

| 值 | 含义 |
|----|------|
| 0 | 正常展示 |
| 1 | 隐藏 / 审核下架（仍保留数据） |

---

## 索引一览（便于排查性能）

| 索引名 | 表 | 用途 |
|--------|-----|------|
| idx_users_email | users | 邮箱唯一 |
| idx_sessions_user_id | sessions | 按用户查会话 |
| idx_sessions_expires_at | sessions | 过期清理 |
| idx_user_follows_followee_id | user_follows | 粉丝列表 |
| idx_friendships_to_user_id | friendships | 收到的申请 |
| idx_friendships_status | friendships | 按状态筛选 |
| idx_user_lists_owner_user_id | user_lists | 某用户的名单列表 |
| idx_user_list_members_member_user_id | user_list_members | 成员反查名单 |
| idx_posts_user_id | posts | 用户时间线 |
| idx_posts_status_created_at | posts | 按状态与时间排序 |
| idx_posts_post_type | posts | 按类型筛选 |
| idx_posts_reply_to_post_id | posts | 拉取回复串 |
| idx_posts_referenced_post_id | posts | 反查转发/引用 |
| idx_posts_visibility | posts | 按可见性筛选 |
| idx_posts_audience_list_id | posts | 名单可见帖 |
| idx_post_replies_post_id_created_at | post_replies | 按帖查评论、时间排序与分页 |
| idx_post_replies_parent_reply_id | post_replies | 子回复列表 |
| idx_post_replies_user_id | post_replies | 用户回复记录 |

---

## 与 SQL 注释的关系

- 字段级说明以 **`.sql` 文件内 `--` 注释** 为准，便于迁移时随版本审阅。
- **枚举取值与业务规则** 在本文档集中列出，便于产品与后端对齐；若变更枚举，请同步改 SQL 头注释中的 SCHEMA 引用与本文档。
