# 登录与注册（OAuth2）

本文说明 Blink 的认证方式：**OAuth2 授权码流程**；提供方可为 **第三方（如 Google）** 或 **本服务内置的 OAuth2 授权服务器（自建 IdP，`builtin`）**，无需依赖外部账号体系即可登录。

实现入口：`cmd/main.go`；OAuth 客户端编排：`application/oauth`；自建 IdP：`application/idp`、`infrastructure/interface/http/idp`；邮箱密码注册：`application/auth`。

---

## 两种模式对比

| 模式 | 说明 |
|------|------|
| **第三方 IdP** | 例如 `google`：用户在 Google 登录，本服务用 code 换 token 再拉 userinfo。 |
| **自建 IdP（`builtin`）** | 本进程提供 `/auth/idp/*`（authorize / token / userinfo），浏览器在 **本服务页面** 输入邮箱与密码；仍通过标准 OAuth2 回调到 `/auth/oauth/builtin/callback`，与第三方路径一致。 |

启用 **自建 IdP** 需同时配置 `BLINK_PUBLIC_BASE_URL` 与 `BLINK_OAUTH_CLIENT_SECRET`（见下文）。未配置时不会注册 `builtin` 提供方，也不会挂载 IdP 与 `/auth/register`。

---

## 概念：OAuth「注册」与「登录」

对任意提供方 `provider`（`google` / `builtin` 等）：

- **首次**在某 `provider` 下出现新的 `(provider, provider_subject)` 时：写入 `users` 与 `oauth_identities`（即「注册」）。
- **再次**同一绑定成功时：只更新登录信息并签发会话（即「登录」）。

**邮箱密码注册**（`POST /auth/register`）会创建 `users` 行，并写入 `oauth_identities`（`provider=builtin`，`provider_subject` 为用户 snowflake 的十进制字符串），以便随后走 `builtin` OAuth 时 `CompleteLogin` 能关联到同一用户。

---

## HTTP 路由（OAuth 客户端）

挂载在 **`/auth/oauth`**：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/auth/oauth/{provider}/login` | 生成 `state`、写入 Redis，302 到 IdP 授权页 |
| GET | `/auth/oauth/{provider}/callback` | 校验 `state`、用 `code` 换 token、拉 userinfo、写 Redis 会话、设置 Cookie |

### 查询参数

- **`login`**：`next`（可选）  
  成功后的重定向地址，**仅允许站内相对路径**（以 `/` 开头且非 `//`），否则为 `/`。

- **`callback`**：`code`、`state`（IdP 回调携带）

### 会话 Cookie

- **名称**：`blink_session`  
- **属性**：`HttpOnly`、`SameSite=Lax`、`Path=/`  
- 生产环境建议 HTTPS 下再开启 `Secure`。

---

## 自建 IdP（不依赖第三方）

当环境变量满足 **`BLINK_PUBLIC_BASE_URL` 非空** 且 **`BLINK_OAUTH_CLIENT_SECRET` 非空** 时：

1. 注册 **`builtin`** OAuth 客户端：`AuthURL` / `TokenURL` / `UserInfoURL` 均指向本服务的 `/auth/idp/*`。  
2. 挂载 **授权服务器**：
   - `GET|POST /auth/idp/authorize`：授权页（GET 为登录表单，POST 提交邮箱、密码后签发 `code` 并重定向回 `redirect_uri`）。  
   - `POST /auth/idp/token`：用 `authorization_code` 换 `access_token`。  
   - `GET /auth/idp/userinfo`：`Bearer` 访问，返回 `sub` / `email` / `name`（与现有 `oauthadapter.Provider` 解析逻辑兼容）。  
3. 挂载 **`POST /auth/register`**：JSON `email`、`password`（至少 8 位）、`name`，创建本地账号并建立 `builtin` 绑定。

### 推荐流程（完全自建）

1. `POST /auth/register` 注册。  
2. 浏览器打开：`{BLINK_PUBLIC_BASE_URL}/auth/oauth/builtin/login?next=/`  
3. 经 `/auth/idp/authorize` 输入邮箱密码 → 回调 `/auth/oauth/builtin/callback` → 设置 `blink_session`。

`BLINK_PUBLIC_BASE_URL` 必须与浏览器访问本服务使用的 **协议 + 主机 + 端口** 一致（例如 `http://localhost:8080`），否则服务端用 `oauth2.Config` 回调自身 `TokenURL` / `UserInfoURL` 会失败。

---

## 数据存储

### 数据库

- **`users`**：含邮箱、展示名、`password_hash`（bcrypt）等。仅通过第三方 OAuth 创建的用户使用随机占位密码，不能用于 IdP 密码登录。  
- **`oauth_identities`**：`platform/db/0004_oauth_identities.sql`，`(provider, provider_subject)` 唯一。

应用层通过 [sqlx](https://github.com/jmoiron/sqlx) 访问上述表，说明见 [`docs/database-sqlx.md`](database-sqlx.md)。

### Redis

| 用途 | Key 前缀 | 说明 |
|------|-----------|------|
| OAuth CSRF state | `blink:oauth:state:{state}` | 短 TTL，回调 **GETDEL** 消费 |
| 应用会话 | `blink:session:{token}` | 登录会话 JSON |
| IdP 授权码 | `blink:idp:code:{code}` | 自建 IdP，一次性 |
| IdP access token | `blink:idp:access:{token}` | 自建 IdP，换 userinfo |

---

## 环境变量

### 通用

| 变量 | 默认 | 说明 |
|------|------|------|
| `BLINK_HTTP_ADDR` | `:8080` | 监听地址 |
| `BLINK_DATABASE_DSN` | `file:./data/blink.db?...` | SQLite |
| `BLINK_REDIS_ADDR` | `127.0.0.1:6379` | Redis |
| `BLINK_MIGRATIONS_DIR` | `platform/db` | 迁移目录（建议在模块根目录运行） |
| `BLINK_SNOWFLAKE_NODE` | `1` | Snowflake 节点 0–1023 |

### 自建 IdP + `builtin` 客户端（同时设置才启用）

| 变量 | 说明 |
|------|------|
| `BLINK_PUBLIC_BASE_URL` | 对外基址，**无尾部斜杠**，如 `http://localhost:8080` |
| `BLINK_OAUTH_CLIENT_SECRET` | 第一方客户端密钥（请在生产中设为强随机） |
| `BLINK_OAUTH_CLIENT_ID` | 默认 `blink` |
| `BLINK_OAUTH_REDIRECT_URL` | 默认 `{BLINK_PUBLIC_BASE_URL}/auth/oauth/builtin/callback`，须与授权请求中的 `redirect_uri` 完全一致（IdP 白名单） |

### Google（可选，三者齐全才启用 `google`）

| 变量 | 说明 |
|------|------|
| `OAUTH_GOOGLE_CLIENT_ID` | Google 客户端 ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | 客户端密钥 |
| `OAUTH_GOOGLE_REDIRECT_URL` | 须与 Google 控制台「已授权重定向 URI」一致 |

---

## 端到端流程（示意）

```mermaid
sequenceDiagram
  participant B as 浏览器
  participant A as Blink
  participant R as Redis
  participant I as IdP（Google 或 /auth/idp）
  participant D as 数据库

  B->>A: GET /auth/oauth/{provider}/login?next=/foo
  A->>R: 保存 oauth state
  A->>B: 302 Location = IdP 授权 URL（含 state）
  B->>I: 用户登录（或自建页输入密码）
  I->>B: 302 redirect_uri?code=&state=
  B->>A: GET /auth/oauth/{provider}/callback?code=&state=
  A->>R: 校验并删除 state
  A->>I: code 换 access_token
  A->>I: 拉取 userinfo
  alt 首次 (provider, subject)
    A->>D: INSERT users + oauth_identities
  else 已存在
    A->>D: 读用户，更新 last_login
  end
  A->>R: 写入 blink_session 数据
  A->>B: Set-Cookie blink_session; 302 next
```

---

## 错误与安全

- 未知 `provider`、无效 `state`、账号非活跃等会返回 4xx；回调错误不暴露内部细节。  
- 自建 IdP 的 `redirect_uri` 必须在服务端白名单内（当前为配置的 `BLINK_OAUTH_REDIRECT_URL`）。  
- `BLINK_OAUTH_CLIENT_SECRET` 勿提交到仓库。

---

## 相关代码路径

| 层级 | 路径 |
|------|------|
| OAuth 登录用例 | `application/oauth/login.go` |
| 邮箱注册 | `application/auth/register.go` |
| 自建 IdP 用例 | `application/idp/service.go` |
| IdP / 注册 HTTP | `infrastructure/interface/http/idp`、`infrastructure/interface/http/auth` |
| Redis | `infrastructure/cache/redisstore` |
| SQL | `infrastructure/persistence/sql` |
