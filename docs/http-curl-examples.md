# HTTP 接口 curl 示例

下文默认服务地址为 **`http://127.0.0.1:11110`**（与默认 `BLINK_HTTP_ADDR=:11110` 一致）。若不同，请替换变量 **`$BASE`**。

**`POST /auth/register`** 默认即挂载，无需额外环境变量。

自建 IdP（`/auth/idp/*`、OAuth **`builtin`**）需已配置 **`BLINK_PUBLIC_BASE_URL`**、**`BLINK_OAUTH_CLIENT_SECRET`** 等，否则这些路径返回 **404**。

---

## 公共变量（按需 export）

```bash
export BASE=http://127.0.0.1:11110
export CLIENT_ID="${BLINK_OAUTH_CLIENT_ID:-blink}"
export CLIENT_SECRET='你的BLINK_OAUTH_CLIENT_SECRET'   # 勿提交到仓库
# 须与 IdP 白名单完全一致（默认等于下面这个）
export REDIRECT_URI="${BASE}/auth/oauth/builtin/callback"
```

`REDIRECT_URI` 在 query / form 里若含特殊字符，请用 curl 的 **`--data-urlencode`** 或对 query 做 URL 编码。

---

## 健康检查

### `GET /health`（JSON）

```bash
curl -sS "${BASE}/health"
```

### `GET /healthz`（纯文本 `ok`）

```bash
curl -sS "${BASE}/healthz"
```

---

## `POST /auth/register`（JSON）

**Content-Type**：`application/json`

**Body 字段**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `email` | string | 是 | 邮箱 |
| `password` | string | 是 | 至少 8 位 |
| `name` | string | 是 | 展示名 |

**成功**：`201`，JSON；若启用会话，还会 `Set-Cookie: blink_session=...`。

```bash
curl -sS -X POST "${BASE}/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123","name":"Demo"}' \
  -w '\nHTTP_CODE:%{http_code}\n'
```

**常见错误**：`400`（体非法、密码过短、邮箱无效）、`409`（邮箱已注册），正文多为 **纯文本**。

---

## `GET /auth/oauth/{provider}/login`

**路径参数**：`provider`，自建为 **`builtin`**；第三方如 **`google`**。

**Query**：

| 参数 | 必填 | 说明 |
|------|------|------|
| `next` | 否 | 登录成功后站内跳转路径，须以 `/` 开头且非 `//`，否则视为 `/` |

**行为**：`302`，`Location` 为 IdP 授权页 URL（内置 IdP 即带参数的 `/auth/idp/authorize?...`），并携带 **`state`**（已写入 Redis，回调时会校验）。

```bash
curl -sS -D - -o /dev/null "${BASE}/auth/oauth/builtin/login?next=/home"
```

从响应头 **`Location:`** 里取出完整 URL；其中 **`state=`** 后面的值在完整 OAuth 流程里必须原样带到 authorize 与 callback（见文末「串联流程」）。

---

## `GET /auth/oauth/{provider}/callback`

**Query（必填）**：

| 参数 | 说明 |
|------|------|
| `code` | 授权码（由 IdP 在 `redirect_uri` 上附加） |
| `state` | 必须与 **`/login` 跳转里带的 state** 一致且在 Redis 中仍有效 |

**行为**：成功时 `302`，`Set-Cookie: blink_session`，`Location` 为之前 `next` 保存的路径。

不建议手填 `code`；应用从浏览器重定向或按下面「串联流程」获取。

```bash
# 仅作形状说明；code/state 须为真实有效值
curl -sS -D - -o /dev/null \
  "${BASE}/auth/oauth/builtin/callback?code=REPLACE_CODE&state=REPLACE_STATE"
```

---

## `GET /auth/idp/authorize`（授权页 HTML）

**Query（必填）**：

| 参数 | 说明 |
|------|------|
| `client_id` | 与 `CLIENT_ID` 一致 |
| `redirect_uri` | 与 IdP 白名单 **逐字一致**（一般为 `REDIRECT_URI`） |
| `response_type` | 固定 **`code`** |
| `state` | 不透明串；从 **`builtin/login` 的 Location** 里取，或自测时与后续 callback 一致（若走完整 OAuth 客户端流程，须与 login 下发的 state 相同） |

```bash
# redirect_uri 在 query 中必须编码
curl -sS -G "${BASE}/auth/idp/authorize" \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  -d "response_type=code" \
  -d "state=YOUR_STATE" \
  -H 'Accept: text/html'
```

**成功**：`200`，`Content-Type: text/html`。

---

## `POST /auth/idp/authorize`（提交登录表单）

**Content-Type**：`application/x-www-form-urlencoded`

**Body 字段**：

| 字段 | 必填 | 说明 |
|------|------|------|
| `client_id` | 是 | 同 `CLIENT_ID` |
| `redirect_uri` | 是 | 同 `REDIRECT_URI`，须与白名单一致 |
| `state` | 是 | 与 GET 阶段一致 |
| `response_type` | 是 | `code` |
| `email` | 是 | 已注册用户的邮箱 |
| `password` | 是 | 对应密码 |

**成功**：`302`，`Location` 为  
`{redirect_uri}?code={授权码}&state={原 state}`。

```bash
curl -sS -D - -o /dev/null -X POST "${BASE}/auth/idp/authorize" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  --data-urlencode "state=YOUR_STATE" \
  -d "response_type=code" \
  --data-urlencode "email=demo@example.com" \
  --data-urlencode "password=password123"
```

查看响应头 **`Location:`** 可得到 **`code`**（仅一次性、短时有效）。

**失败**：`400` / `401`，多为纯文本。

---

## `POST /auth/idp/token`（用授权码换访问令牌）

**Content-Type**：`application/x-www-form-urlencoded`

**Body 字段**：

| 字段 | 必填 | 说明 |
|------|------|------|
| `grant_type` | 是 | 固定 **`authorization_code`** |
| `code` | 是 | 上一步 `redirect_uri` 上的 `code` |
| `redirect_uri` | 是 | 须与 authorize 请求里 **完全相同** |
| `client_id` | 是 | 同 `CLIENT_ID` |
| `client_secret` | 是 | 同 `CLIENT_SECRET` |

**成功**：`200`，JSON，例如：

`{"access_token":"...","token_type":"Bearer","expires_in":3600}`（`expires_in` 为秒，随配置变化）。

```bash
curl -sS -X POST "${BASE}/auth/idp/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=authorization_code' \
  --data-urlencode "code=${CODE}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  -d "client_id=${CLIENT_ID}" \
  --data-urlencode "client_secret=${CLIENT_SECRET}"
```

**失败**：`400`，JSON，如 `{"error":"invalid_grant"}`。

---

## `GET /auth/idp/userinfo`

**Header**：

| Header | 说明 |
|--------|------|
| `Authorization` | **`Bearer {access_token}`**（来自 `/auth/idp/token`） |

**成功**：`200`，JSON：`sub`、`email`、`name`。

```bash
curl -sS "${BASE}/auth/idp/userinfo" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**失败**：`401`，多为纯文本。

---

## 串联示例：纯 curl 走通「注册 → IdP 换票 → userinfo」

不经过 `/auth/oauth/.../callback`，只验证 IdP 与注册（适合联调 token/userinfo）：

```bash
export BASE=http://127.0.0.1:11110
export CLIENT_ID=blink
export CLIENT_SECRET='你的密钥'
export REDIRECT_URI="${BASE}/auth/oauth/builtin/callback"

# 1) 注册（若用户已存在可跳过）
curl -sS -X POST "${BASE}/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123","name":"Demo"}'

# 2) POST authorize，从 Location 取出 code（手动复制 CODE 到变量）
STATE=manual-test-state-123
curl -sS -D /tmp/idp.hdr -o /dev/null -X POST "${BASE}/auth/idp/authorize" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  --data-urlencode "state=${STATE}" \
  -d "response_type=code" \
  --data-urlencode "email=demo@example.com" \
  --data-urlencode "password=password123"
grep -i '^Location:' /tmp/idp.hdr
# 假设从 Location 解析出 code 赋给 CODE
export CODE='从Location复制的code'

# 3) token
export ACCESS_TOKEN=$(curl -sS -X POST "${BASE}/auth/idp/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=authorization_code' \
  --data-urlencode "code=${CODE}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  -d "client_id=${CLIENT_ID}" \
  --data-urlencode "client_secret=${CLIENT_SECRET}" | jq -r .access_token)

# 4) userinfo
curl -sS "${BASE}/auth/idp/userinfo" -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

若未安装 `jq`，可肉眼从 JSON 里取 `access_token`。

---

## 串联示例：含 `/auth/oauth/builtin/callback` 的完整浏览器式流程（curl 模拟）

要点：**`state` 必须由 `GET .../login` 生成并存在于 Redis**，authorize 与 callback 必须使用 **同一 `state`**，`code` 来自 authorize POST 之后的 `Location`。

```bash
export BASE=http://127.0.0.1:11110

# 1) 触发 login，拿到 authorize 的完整 URL（内含 state）
LOC=$(curl -sS -D - -o /dev/null "${BASE}/auth/oauth/builtin/login?next=/dashboard" | tr -d '\r' | awk -F': ' 'tolower($1)=="location" {print $2; exit}')
echo "$LOC"

# 2) 用 Python 从 URL 解析 state（也可手工复制）
STATE=$(python3 -c "import urllib.parse; print(urllib.parse.parse_qs(urllib.parse.urlparse('$LOC').query)['state'][0])")

# 3) POST authorize（email/password 换成你的账号）
curl -sS -D /tmp/idp2.hdr -o /dev/null -X POST "${BASE}/auth/idp/authorize" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode "client_id=${CLIENT_ID:-blink}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI:-${BASE}/auth/oauth/builtin/callback}" \
  --data-urlencode "state=${STATE}" \
  -d "response_type=code" \
  --data-urlencode "email=demo@example.com" \
  --data-urlencode "password=password123"

# 4) 从 Location 取 code
grep -i '^Location:' /tmp/idp2.hdr
CODE=$(python3 -c "
import urllib.parse
loc = open('/tmp/idp2.hdr').read().split('Location:')[1].split('\n')[0].strip()
q = urllib.parse.parse_qs(urllib.parse.urlparse(loc).query)
print(q['code'][0])
")

# 5) 回调（会 Set-Cookie；保存 cookie 罐以便后续请求）
curl -sS -c /tmp/blink.cookies -D - -o /dev/null \
  "${BASE}/auth/oauth/builtin/callback?code=${CODE}&state=${STATE}"
```

第 5 步成功时响应头里会有 **`Set-Cookie: blink_session=...`** 与 **`Location: /dashboard`**（或你传入的 `next`）。

---

## 与文档的关系

- 流程说明：[`auth-login-registration.md`](auth-login-registration.md)  
- OpenAPI 字段汇总：[`../api/openapi/openapi.yaml`](../api/openapi/openapi.yaml)
