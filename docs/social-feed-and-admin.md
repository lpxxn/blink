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
| `BLINK_SENSITIVE_WORDS` | 可选，英文逗号分隔的敏感词子串；未配置则不匹配任何词（帖子、评论均 **视为无敏感词**）。命中时：帖子写入 `moderation_flag=1` 与 `moderation_note`；评论接口返回 400。 |

仍需：**Redis**（会话）、**数据库迁移**（含 `platform/db/0005_post_categories_moderation.sql`）。

## 超级管理员

- 库中 `users.role` 为 **`super_admin`** 时可访问 `/admin/api/*`（需有效 `blink_session` 或 `Authorization: Bearer`）。
- 授予方式：SQL 更新 `role`，或配置 `BLINK_BOOTSTRAP_SUPER_ADMIN_EMAIL` 后重启进程。
- 不可在接口中将自己的角色从 `super_admin` 改为其他值（会返回 403）。

## 默认分类

进程启动时若 `categories` 表为空，会插入若干内置分类（slug：`general`、`tech`、`life`、`fun`），ID 由当前 snowflake 节点生成。

## 审核字段

`posts.moderation_flag`：`0` 正常（公开流可见）、`1` 标记违规、`2` 管理下架。公开流仅展示 `0` 且已发布、未软删的原创公开帖。

新建/编辑帖子时会对正文跑 **敏感词子串检测**（见 `application/moderation`）：无命中则 **`moderation_flag=0`（审核通过）**；有命中则自动标为 `1` 并在 `moderation_note` 中记录 `sensitive_hit: ...`。评论若命中敏感词则 **拒绝提交**（HTTP 400）。
