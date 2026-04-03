# 登录与注册（OAuth2）

本文说明 Blink 当前基于 **OAuth2 授权码流程** 的「注册」与「登录」行为、HTTP 接口、存储与安全注意点。实现入口：`cmd/main.go`，应用编排：`application/oauth`，HTTP：`infrastructure/interface/http/oauth`。

## 概念：注册与登录是同一条链路

本服务没有单独的「注册表单」接口。用户在某 OAuth 提供方（如 Google）上**首次**完成授权时，系统会：

1. 在 `users` 表创建一条用户记录；
2. 在 `oauth_identities` 表写入 `(provider, provider_subject) → user_id` 绑定。

同一提供方、同一 `provider_subject` **再次**授权时，不再创建用户，只更新登录信息并签发新会话，即**登录**。

因此：**首次 OAuth 成功 = 注册；后续 OAuth 成功 = 登录**。

## HTTP 路由

服务将 OAuth 路由挂载在 **`/auth/oauth`** 下（见 `cmd/main.go` 的 `r.Mount("/auth/oauth", h.Routes())`）。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/auth/oauth/{provider}/login` | 生成 `state`、写入 Redis，302 重定向到 IdP 授权页 |
| GET | `/auth/oauth/{provider}/callback` | 校验 `state`、用 `code` 换 token、拉取用户信息、建会话、设置 Cookie |

### 查询参数

- **`login`**：`next`（可选）  
  登录成功后的浏览器重定向地址，**仅允许站内相对路径**（必须以 `/` 开头且不能是 `//` 开头），否则回退为 `/`。用于防止开放重定向。

- **`callback`**：由 IdP 回调携带  
  - `code`：授权码（必填）  
  - `state`：与发起登录时一致（必填）

### 会话 Cookie

回调成功后会设置：

- **名称**：`blink_session`  
- **属性**：`HttpOnly`、`SameSite=Lax`、`Path=/`  
- **值**：不透明会话令牌（由服务端生成，实际会话数据在 Redis）

生产环境若全站 HTTPS，建议再为 Cookie 开启 `Secure`（需在代码或反向代理层按需配置）。

## 端到端流程

```mermaid
sequenceDiagram
  participant B as 浏览器
  participant A as Blink
  participant R as Redis
  participant I as IdP (Google 等)
  participant D as 数据库

  B->>A: GET /auth/oauth/{provider}/login?next=/foo
  A->>R: 保存 oauth state
  A->>B: 302 Location = IdP 授权 URL（含 state）
  B->>I: 用户登录并同意授权
  I->>B: 302 redirect_uri?code=&state=
  B->>A: GET /auth/oauth/{provider}/callback?code=&state=
  A->>R: 校验并删除 state（一次性）
  A->>I: code 换 access_token
  A->>I: 拉取 userinfo（sub / email / name）
  alt 首次该 (provider, subject)
    A->>D: INSERT users + oauth_identities
  else 已存在绑定
    A->>D: 读取用户，更新 last_login 等
  end
  A->>R: 写入 session
  A->>B: Set-Cookie blink_session; 302 next
```

## 数据存储

### 关系型数据库（SQLite / 可换 MySQL、PostgreSQL）

- **`users`**：用户主数据（含邮箱、展示名、密码摘要等）。OAuth 专用账号使用随机 bcrypt 占位密码，**不能用于密码登录**（当前未暴露密码登录接口）。  
- **`oauth_identities`**：第三方身份绑定，见迁移 `platform/db/0004_oauth_identities.sql`。  
  唯一约束：`(provider, provider_subject)`。

若 IdP 未返回邮箱，应用层会使用合成邮箱 `oauth.{provider}.{subject}@oauth.local` 以满足 `users.email` 唯一性（见 `application/oauth/login.go` 中 `loginEmail`）。

### Redis

| 用途 | Key 前缀 / 模式 | 说明 |
|------|-------------------|------|
| OAuth CSRF state | `blink:oauth:state:{state}` | 短 TTL（默认 10 分钟），回调时用 **GETDEL** 一次性消费 |
| 登录会话 | `blink:session:{token}` | 值为 JSON（含 `user_id` 等），TTL 与会话有效期一致（默认 7 天） |

实现：`infrastructure/cache/redisstore`。

## 环境变量与本地运行

### 通用

| 变量 | 默认 | 说明 |
|------|------|------|
| `BLINK_HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `BLINK_DATABASE_DSN` | `file:./data/blink.db?...` | SQLite DSN |
| `BLINK_REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址 |
| `BLINK_MIGRATIONS_DIR` | `platform/db` | 迁移 SQL 目录（**需在模块根目录运行**，或设为绝对路径） |
| `BLINK_SNOWFLAKE_NODE` | `1` | Snowflake 节点号 0–1023 |

### Google OAuth（三者同时设置才启用 `google` 提供方）

| 变量 | 说明 |
|------|------|
| `OAUTH_GOOGLE_CLIENT_ID` | Google OAuth 客户端 ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | 客户端密钥 |
| `OAUTH_GOOGLE_REDIRECT_URL` | 必须与 Google 控制台配置的「已授权重定向 URI」**完全一致** |

示例（本地）：

- 授权回调地址：`http://localhost:8080/auth/oauth/google/callback`  
- 则 `OAUTH_GOOGLE_REDIRECT_URL` 必须为该字符串。

用户入口示例：

- 打开：`http://localhost:8080/auth/oauth/google/login?next=/`

## 错误与安全说明

- **未知提供方**：未配置的 `{provider}` 会返回 400。  
- **无效或过期 state**：Redis 中无记录或已消费，拒绝回调（防 CSRF / 重放）。  
- **账号状态**：`users.status` 非「正常」时拒绝登录（见应用层 `ErrUserSuspended`）。  
- **错误细节**：回调失败时 HTTP 层仅返回通用 400，避免向客户端泄露内部原因；排障请查服务端日志。

## 相关代码路径

| 层级 | 路径 |
|------|------|
| 用例 | `application/oauth/login.go`、`application/oauth/port.go` |
| 领域 | `domain/user`、`domain/oauth`（含 `StateStore`）、`domain/session` |
| Redis | `infrastructure/cache/redisstore` |
| SQL | `infrastructure/persistence/sql` |
| OAuth2 + UserInfo | `infrastructure/adapter/oauth2` |
| HTTP | `infrastructure/interface/http/oauth/handler.go` |

## 扩展更多提供方

在 `cmd/main.go` 中向 `providers`  map 注入新的 `OAuth2Provider` 实现（需配置对应 `oauth2.Config` 与 UserInfo URL），并保证回调 URL 与 IdP 控制台配置一致。测试可参考 `application/oauth`、`infrastructure/interface/http/oauth` 下的用例。
