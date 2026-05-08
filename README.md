# Blink

Blink 是一个轻量的微博/动态（microblog）后端项目：提供注册/登录、帖子流、评论、图片上传、站内通知，以及一个简洁的超级管理员后台（静态 HTML）。后端采用 Go，并按 DDD 分层组织代码，接口契约以 OpenAPI 维护。

## 功能

- **认证与会话**
  - 邮箱密码注册：`POST /auth/register`（始终可用）
  - OAuth2 授权码登录：第三方（如 Google）或自建 IdP（`builtin`）
  - 会话 Cookie：`blink_session`（Redis 存储会话）
- **社交内容**
  - 分类：`GET /api/categories`（启动时空表会写入内置分类）
  - 帖子流：`GET /api/posts`、`GET /api/posts/{id}`、`GET /api/me/posts`
  - 评论/楼中楼：`GET /api/posts/{id}/replies`、发布评论（需登录）
  - 图片上传：`POST /api/uploads`（multipart 字段 `file`），默认存储到 `data/uploads`，通过 `/uploads/...` 访问
- **站内通知（异步）**
  - Watermill + Redis Stream：业务成功后发布事件，消费者写入 `notifications` 表（可在未来独立成 worker）
- **内容治理与后台**
  - 敏感词：命中则拒绝发布/评论；后台 CRUD 后通过 Redis Stream 广播刷新
  - 超级管理员 JSON API：`/admin/api/*`（`users.role=super_admin`）
  - 静态页面：`/web/*.html`（无 React 等框架，保持简单）
- **工程约定**
  - Snowflake ID 在 JSON 中统一用**字符串**传输，避免 JS 精度问题

## 快速开始（本地运行）

需要 **Go** 和 **Redis**（默认 `127.0.0.1:6379`）。在仓库根目录执行：

```bash
mkdir -p data
go run ./cmd
```

启动后可用健康检查确认：

```bash
curl -sS http://127.0.0.1:11110/healthz
curl -sS http://127.0.0.1:11110/health
```

## 配置（环境变量）

常用环境变量（均有默认值，可按需覆盖）：

- **`BLINK_HTTP_ADDR`**：监听地址（默认 `:11110`）
- **`BLINK_DATABASE_DSN`**：数据库 DSN（默认 SQLite，文件在 `./data/blink.db`）
- **`BLINK_REDIS_ADDR`**：Redis 地址（默认 `127.0.0.1:6379`）
- **`BLINK_MIGRATIONS_DIR`**：迁移目录（默认 `platform/db`）
- **`BLINK_UPLOAD_DIR`**：上传目录（默认 `data/uploads`）
- **`BLINK_BOOTSTRAP_SUPER_ADMIN_EMAIL`**：启动时将指定邮箱用户提升为 `super_admin`（幂等）

自建 IdP（`builtin`）只在同时配置以下变量时启用（未配置时相关路由不挂载，不影响 `POST /auth/register`）：

- **`BLINK_PUBLIC_BASE_URL`**：对外基址，如 `http://localhost:11110`（无尾部 `/`）
- **`BLINK_OAUTH_CLIENT_SECRET`**：第一方客户端密钥（生产请用强随机）

更完整列表与解释见 `docs/`。

## 数据库迁移

默认 SQLite 迁移：

```bash
go run ./cmd/migrate
```

该迁移 CLI 也支持 `postgres` / `mysql`，详见：
- `platform/db/SCHEMA.md`
- `cmd/migrate/README.md`

## 文档与接口契约

- **本地运行与健康检查**：`docs/run-local.md`
- **架构（DDD 分层、HTTP vs 通知流）**：`docs/architecture.md`
- **登录/注册与 OAuth 流程**：`docs/auth-login-registration.md`
- **帖子流与管理后台**：`docs/social-feed-and-admin.md`
- **站内通知（Watermill + Redis Stream）**：`docs/watermill-notifications.md`
- **HTTP curl 示例**：`docs/http-curl-examples.md`
- **OpenAPI**：`api/openapi/openapi.yaml`
  - 若修改了 OpenAPI，并需要重新生成代码：见 `docs/oapi-codegen.md`（会生成/更新 `api/gen/apigen.gen.go`）

## 目录结构（速览）

- `cmd/`：入口（API Server、migrate 等）
- `domain/`：领域模型、仓储接口、领域事件
- `application/`：用例服务（post/admin/auth/notification/...）
- `infrastructure/`：HTTP（Gin）、GORM 持久化、Redis/Watermill、OAuth 适配
- `platform/db/`：SQL migrations
- `web/`：静态 HTML（管理/页面资源）

## Roadmap（实现中）

- 前端计划：Flutter 客户端（仓库内会逐步补齐对应实现与文档）
  - 规划文档：`docs/flutter-client-plan.md`
