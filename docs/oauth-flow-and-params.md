# OAuth 登录流程与参数（`next` / `state` / `code`）

本文说明 Blink 中 **OAuth2 授权码模式** 下各参与方、查询参数含义，以及 **`/auth/oauth/{provider}/login`** 上的 **`next`**、全流程中的 **`state`**、**`code`** 分别做什么。实现主要在 `application/oauth`（客户端编排）与 `application/idp`（自建 IdP）。

---

## 0. 为什么有 `/auth/idp/*` 还要 `/auth/oauth/.../callback`？

很多人会先记住自建 IdP 的三条接口，觉得「登录」已经在那边完成了；**callback 不是重复登录，而是 OAuth 里「客户端收码、换票、发本站会话」的必经步骤。**

| 你看到的 URL | 扮演角色 | 一句话 |
|--------------|----------|--------|
| `GET/POST /auth/idp/authorize` | **授权服务器（IdP）** | 浏览器里输入密码；成功后 **302 到 `redirect_uri`**，带上 **`code` + `state`** |
| `POST /auth/idp/token` | **IdP** | 用 **`code` + client_secret** 换 **`access_token`**（通常由 **Blink 服务端**在 callback 里调，不是浏览器直接调） |
| `GET /auth/idp/userinfo` | **IdP** | 用 Bearer token 返回 `sub` / `email` / `name` |
| `GET /auth/oauth/{provider}/login` | **Blink = OAuth 客户端** | 生成 **`state`**、写 Redis、**302 到 authorize**（把用户送进 IdP） |
| `GET /auth/oauth/{provider}/callback` | **Blink = OAuth 客户端** | 浏览器带着 **`code`** 落回本站；Blink **校验 state**、**调 token + userinfo**、写 **Redis 会话**、**`Set-Cookie: blink_session`**、再 **302 到 `next`** |

**记忆法**：`/auth/idp/*` 回答「**谁在 IdP 侧证明身份**」；`/auth/oauth/*` 回答「**Blink 作为接入方，如何把这次授权变成本站已登录**」。没有 callback，浏览器地址栏里会有 `code`，但 **Blink 不会替你换票、也不会下发 `blink_session`**。

---

## 1. 三个角色

| 角色 | 在本项目中的体现 | 职责 |
|------|------------------|------|
| **浏览器** | 用户终端 | 跟随 302 跳转，携带 query，接收 `Set-Cookie` |
| **Blink 应用（OAuth 客户端）** | `/auth/oauth/*` | 发起登录、处理回调、用 `code` 换 token、拉 userinfo、建立 **本站** 会话 `blink_session` |
| **IdP（授权服务器）** | 第三方（如 Google）或自建 **`/auth/idp/*`** | 用户在授权页完成认证，向客户端的 **`redirect_uri`** 颁发 **一次性授权码 `code`** |

对 IdP 而言，「登录成功」= 在重定向 URL 上附上 **`code`（及 `state`）**；对 Blink 而言，还要完成换票、建会话、再按「登录前想去哪」跳转。

---

## 2. `next` 是什么？存在哪？

**出现位置**：仅作为 **`GET /auth/oauth/{provider}/login`** 的查询参数：

```http
GET /auth/oauth/builtin/login?next=/dashboard
```

**含义**：OAuth 整条链走完、会话已建立后，希望浏览器最终落在 **站内哪条路径**（例如 `/`、`/dashboard`）。

**安全处理**（`application/oauth/login.go` 中的 `safeNextURL`）：

- 空或未传 → 视为 **`/`**
- 必须是 **站内相对路径**：以 **`/`** 开头，且不能以 **`//`** 开头（避免开放重定向到外部域名）
- 否则回退为 **`/`**

**如何传到「最后一步」**：  
在调用 **`LoginRedirectURL`** 时，服务端把 **`safeNextURL(next)`** 与 **`provider`** 一起写入 **Redis**，键为随机生成的 **`state`**（见下节）。回调 **`CompleteLogin`** 成功时，从 Redis 取出 **`NextURL`**，作为 **`302 Location`** 的目标。

因此：**`next` 只在第一次请求里出现在 URL 上**；之后靠 **`state` + Redis** 记住「成功后要去哪」。（实现见 `safeNextURL` 与 `LoginRedirectURL`。）

---

## 3. `state` 是什么？

**在标准 OAuth2 里**：`state` 是客户端生成、经 IdP **原样回传** 的不透明字符串，用于 **防 CSRF**（防止攻击者把别人的 `code` 绑到你的会话上）。

**在 Blink 里**它还承担：

1. **Redis 键**：`LoginRedirectURL` 生成随机 `state`，写入 Redis（`RedirectState`：`Provider` + `NextURL`），TTL 见 `StateTTL`。
2. **回调校验**：`GET /auth/oauth/{provider}/callback?code=...&state=...` 中 **`Consume(state)`** 必须成功且 **`provider`** 一致，否则拒绝。

若 **`state`** 从未经过 **`/auth/oauth/.../login`**（例如手工写死 `state` 只调 IdP），Redis 中无记录，**走 callback 会失败**。

---

## 4. `code` 是什么？

**授权码**：由 IdP 在重定向到 **`redirect_uri`** 时通过 query 提供 **`code`**，**一次性、短时有效**，只能由 **持有 `client_secret` 的服务端** 到 **`POST /auth/idp/token`**（或第三方 Token 端点）换取 **`access_token`**。

---

## 5. 自建 IdP（`builtin`）时浏览器完整路径

前提：已配置 `BLINK_PUBLIC_BASE_URL`、`BLINK_OAUTH_CLIENT_SECRET` 等，使 **`builtin`** 与 **`/auth/idp/*`** 可用。

| 步骤 | 行为 |
|------|------|
| 1 | 用户访问 `{BASE}/auth/oauth/builtin/login?next=/dashboard`（示例） |
| 2 | 服务端生成 **`state`**，Redis 保存 `{ Provider: builtin, NextURL: /dashboard }` |
| 3 | **302** 到 IdP 授权页，例如 `/auth/idp/authorize?...&state=【随机 state】` |
| 4 | 用户在 authorize 页提交邮箱密码（或第三方则在 Google 登录） |
| 5 | IdP **302** 到 **`redirect_uri`**（默认 `{BASE}/auth/oauth/builtin/callback`），带上 **`code`** 与 **`state`** |
| 6 | **`callback`**：`Consume(state)` → 用 `code` 换 token → userinfo → 建 Redis 会话 → **`Set-Cookie: blink_session`** → **302 到 `/dashboard`**（即当初的 **`next`**） |

---

## 6. 与「只测 IdP」的 curl 手工流程的区别

文档中 **`STATE=manual-test-state-123`** 直接 **`POST /auth/idp/authorize`** 时：

- IdP 仍会 **302**，`Location` 中含 **`code`** 与 **`state`**。
- 该 **`state` 未经过 `/auth/oauth/.../login`**，Redis **无** 对应记录。
- 若再访问 **`/auth/oauth/builtin/callback?code=...&state=...`**，**`Consume(state)` 会失败**。

**若只想验证 IdP 与 token**：  
从 `Location` 取出 **`code`**，再 **`POST /auth/idp/token`** 即可，**不必**走 callback。

**若要验证含 `next` 的整段登录**：  
必须从 **`/auth/oauth/builtin/login?next=...`** 进入，使用服务端生成的 **`state`**。

---

## 7. 参数速查表

| 参数 | 典型出现位置 | 作用 |
|------|----------------|------|
| **`next`** | `GET /auth/oauth/{provider}/login?next=` | 登录成功后站内跳转路径；经 `safeNextUrl` 后与 `provider` 一起存入 Redis（键为 `state`） |
| **`state`** | login 生成 → 进 IdP URL → 回到 callback | 防 CSRF；在 Blink 中兼作 **Redis 键**，用于取回 **`next`** 与 **`provider`** |
| **`code`** | IdP 重定向到 `redirect_uri` 的 query | 一次性授权码，用于换 `access_token` 并完成会话建立 |

---

## 相关文档

- [登录与注册（总览）](auth-login-registration.md)
- [HTTP curl 示例](http-curl-examples.md)
