# 本地运行 Blink

## 依赖

- **Go**：与 `go.mod` 中版本一致（当前为 1.26+）。
- **Redis**：默认连接 `127.0.0.1:6379`，启动前需可用（进程会 `PING` 检查）。
- **SQLite**：默认数据库文件为仓库根目录下的 `./data/blink.db`（需可写；首次运行前可先创建目录）。

## 启动 API 服务

在**仓库根目录**执行：

```bash
mkdir -p data
go run ./cmd
```

等价写法：

```bash
go run -C . ./cmd
```

常用环境变量（均有默认值，可按需覆盖）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `BLINK_HTTP_ADDR` | `:11110` | 监听地址 |
| `BLINK_DATABASE_DSN` | `file:./data/blink.db?...` | SQLite DSN（相对 **当前工作目录**） |
| `BLINK_REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址 |
| `BLINK_MIGRATIONS_DIR` | `platform/db` | SQL 迁移目录（相对仓库根） |
| `BLINK_SNOWFLAKE_NODE` | `1` | Snowflake 节点 ID，`0..1023` |

示例（换端口，例如避免与本机其它服务冲突）：

```bash
BLINK_HTTP_ADDR=:8080 go run ./cmd
```

启动成功后，日志中会出现类似：`listening on :11110 ...`。

### 健康检查

- `GET /healthz`：返回纯文本 `ok`（探活）。
- `GET /health`：OpenAPI 约定，返回 JSON：`{"status":"ok"}`。

示例（按你实际监听的地址与端口修改）：

```bash
curl -sS http://127.0.0.1:11110/healthz
curl -sS http://127.0.0.1:11110/health
```

若本机某端口在 **IPv4** 上已被占用，而服务实际只在你访问的地址上监听，可改用 `BLINK_HTTP_ADDR` 指定空闲端口，或向当前服务实际绑定的地址发起请求（例如部分环境下需使用 `http://[::1]:11110/...`）。

## 数据库迁移（可选）

单独执行迁移（同样在仓库根目录、`BLINK_DATABASE_DSN` 与主程序一致）：

```bash
go run ./cmd/migrate
```

## 与 VS Code / Cursor 调试

使用仓库内 [`.vscode/launch.json`](../.vscode/launch.json) 中的 **「Blink: API Server」** 配置即可在 IDE 中启动（`cwd` 已为仓库根目录，便于 `./data` 等相对路径生效）。

## 已在本环境验证过的命令

以下命令在具备 Redis、且于仓库根目录执行的前提下可用于确认服务正常：

```bash
mkdir -p data
go run ./cmd
# 另开终端：
curl -sS http://127.0.0.1:11110/healthz
curl -sS http://127.0.0.1:11110/health
```

（若默认端口被其他进程占用，请改用 `BLINK_HTTP_ADDR` 或能访问到 Gin 监听地址的 URL。）

更多接口的 **curl 参数与串联示例** 见 [`http-curl-examples.md`](http-curl-examples.md)。
