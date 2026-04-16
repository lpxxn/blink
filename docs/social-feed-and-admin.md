# 帖子流、评论与超级管理后台

本仓库在原有 OAuth / 注册 / Redis 会话之上，提供：

- **公开 API**：`GET /api/categories`、`GET /api/posts`（按 `category_id` 或 `uncategorized=1` 筛选）、`GET /api/posts/{id}`（可选登录，作者可看自己的草稿/非公开帖）、`GET /api/posts/{id}/replies`。
- **需登录**：发帖与改删、`GET /api/me/posts`、评论、图片上传 `POST /api/uploads`（multipart 字段 `file`）。
- **静态页**：`/web/*.html`（`r.Static("/web", "./web")`）。
- **JSON 与浏览器精度**：帖子、分类、评论、用户等 snowflake ID 在 JSON 中均为 **带引号的字符串**（Go 侧 `json:",string"`），`next_cursor` 亦为字符串，避免 JavaScript `Number` 丢精度。
- **上传文件**：默认目录 `data/uploads`（`BLINK_UPLOAD_DIR`），通过 `GET /uploads/...` 访问。

## 环境变量

| 变量 | 说明 |
|------|------|
| `BLINK_UPLOAD_DIR` | 上传存储目录，默认 `data/uploads` |
| `BLINK_BOOTSTRAP_SUPER_ADMIN_EMAIL` | 若该邮箱用户存在且当前角色为 `user`，启动时提升为 `super_admin`（幂等；已为非 `user` 则跳过） |
| `BLINK_SENSITIVE_WORDS_POLL_INTERVAL` | 可选，如 `3m`：周期性从数据库全量重载敏感词内存快照（消息丢失或仅 API 节点时的兜底）。 |
| `BLINK_DISABLE_SENSITIVE_WORDS_CONSUMER` | 可选，设为非空表示禁用敏感词变更广播消费（不建议；多实例会依赖轮询兜底）。 |

敏感词存于表 **`sensitive_words`**，进程内维护快照；管理端 `POST/PATCH/DELETE /admin/api/sensitive_words` 写库后会 **Reload** 并通过 Redis Stream `blink.moderation.sensitive_words` 通知其他实例刷新。正式发布帖子若正文命中敏感词则 **拒绝**（HTTP 400）；草稿可保存。评论命中敏感词同样 **拒绝提交**。

仍需：**Redis**（会话、Watermill Stream）、**数据库迁移**（含 `platform/db/0007_sensitive_words.sql` 等）。

## 超级管理员

- 库中 `users.role` 为 **`super_admin`** 时可访问 `/admin/api/*`（需有效 `blink_session` 或 `Authorization: Bearer`）。
- 授予方式：SQL 更新 `role`，或配置 `BLINK_BOOTSTRAP_SUPER_ADMIN_EMAIL` 后重启进程。
- 不可在接口中将自己的角色从 `super_admin` 改为其他值（会返回 403）。

## 默认分类

进程启动时若 `categories` 表为空，会插入若干内置分类（slug：`general`、`tech`、`life`、`fun`），ID 由当前 snowflake 节点生成。

## 审核字段

`posts.moderation_flag`：`0` 正常（公开流可见）、`1` 标记违规、`2` 管理下架。公开流仅展示 `0` 且已发布、未软删的原创公开帖。

**正式发布**的帖子在创建/编辑时会检测敏感词：命中则接口 **失败**（不写入已发布内容）。**草稿**可含敏感词直至用户改为发布。评论若命中敏感词则 **拒绝提交**（HTTP 400）。管理员可对单条评论 `PATCH /admin/api/replies/{id}` body `{"hidden":true}`，该条及其所有子回复均设为隐藏。
