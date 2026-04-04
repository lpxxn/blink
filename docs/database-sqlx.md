# 数据库访问：sqlx

本仓库在 **`infrastructure/persistence/sql`** 使用 [jmoiron/sqlx](https://github.com/jmoiron/sqlx) 作为对标准库 `database/sql` 的扩展：在**不引入 ORM** 的前提下，用 `GetContext` / `Select` 等能力把查询结果映射到带 `db` 标签的结构体，并保持与现有 SQL、迁移工具兼容。

---

## 为何使用 sqlx

| 能力 | 说明 |
|------|------|
| **兼容 `database/sql`** | `*sqlx.DB` 包装 `*sql.DB`，`ExecContext`、`QueryRowContext` 等用法不变。 |
| **按列扫描到结构体** | `GetContext`、`Select` 根据 `db` 标签映射列名，减少手写 `Scan` 顺序错误。 |
| **命名参数（可选）** | 后续可对复杂 SQL 使用 `NamedExec` / `NamedQuery`（当前仓储以位置参数为主）。 |
| **多驱动绑定** | sqlx 按驱动名选择 `?`、`$1` 等占位符风格，便于同一套 SQL 在 SQLite / PostgreSQL / MySQL 上迭代。 |

领域层（`domain/`）**不出现** `db` 标签或 sqlx 类型；持久化层定义 `userRow`、`oauthRow` 等内部结构，再转换为领域实体。

---

## 依赖与模块

- Go 模块：`github.com/jmoiron/sqlx`（见根目录 `go.mod`）。
- SQLite 驱动：`_ "modernc.org/sqlite"`，驱动名 **`sqlite`**（与 `cmd/migrate` 一致）。
- 其他驱动（迁移 CLI 已引用）：MySQL `github.com/go-sql-driver/mysql`，PostgreSQL `github.com/lib/pq`。

安装或升级示例：

```bash
go get github.com/jmoiron/sqlx@latest
go mod tidy
```

---

## 连接与生命周期

### 主进程（`cmd/main.go`）

1. 使用 **`sql.Open("sqlite", dsn)`** 得到 `*sql.DB`（迁移器 `internal/migrator` 仍接收 `*sql.DB`）。
2. 使用 **`sqlx.NewDb(sqldb, "sqlite")`** 得到 `*sqlx.DB`，供仓储注入。
3. 对 `sqlx.DB` 调用 **`Ping()`** 校验连通性。
4. **`migrator.Run(sqldb, "sqlite", migDir)`** 在**底层** `*sql.DB` 上执行，与是否使用 sqlx 无关。

要点：**关闭连接**应对底层 `sqldb` 调用 `Close()`（`defer sqldb.Close()`）；`sqlx.NewDb` 不拥有独立连接池，与传入的 `*sql.DB` 共享。

### DSN 与环境变量

- 默认：`BLINK_DATABASE_DSN`，见 `cmd/main.go`。
- 迁移目录：`BLINK_MIGRATIONS_DIR`（默认 `platform/db`）。

---

## 仓储实现约定

### 类型

- `UserRepository`、`OAuthRepository` 字段为 **`DB *sqlx.DB`**。
- 查询单行：优先 **`GetContext(ctx, &row, query, args...)`**。
- 写操作：继续使用 **`ExecContext`**（与标准库一致）。

### 行模型与领域模型

- 在 `infrastructure/persistence/sql` 内定义 `userRow`、`oauthRow`，字段带 **`db:"column_name"`**，与 `platform/db/*.sql` 中列名一致。
- 提供 **`toDomain()`**（或等价函数）映射为 `domain/user.User`、`domain/oauth.Identity`，避免在领域结构体上挂持久化标签。

### SQL 与可移植性

- DDL 与迁移脚本仍放在 **`platform/db`**，由 `cmd/migrate` 顺序执行。
- 应用内 SQL 当前使用 **`?`** 占位符（SQLite / MySQL 风格）。若切换到 PostgreSQL，需将仓储 SQL 改为 **`$1, $2, ...`** 或使用 sqlx 的命名查询由驱动绑定；sqlx 的 `BindDriver` 可辅助统一策略（见 [sqlx README](https://github.com/jmoiron/sqlx)）。

---

## 迁移（与 sqlx 的关系）

- **`internal/migrator`** 仅依赖 **`database/sql`**，签名 `Run(db *sql.DB, driver, dir string)`。
- **不改变**迁移文件格式与执行顺序。
- `cmd/migrate` 仍可用纯 `sql.Open` + `migrator.Run`；主服务在迁移完成后用 **`sqlx.NewDb`** 包装同一 `*sql.DB` 即可。

---

## 测试

- **`internal/testutil.OpenSQLiteMemory`**：内存 SQLite + 跑全量 `platform/db` 迁移，返回 **`*sqlx.DB`**，供 `application/*`、`infrastructure/interface/http/*` 测试注入仓储。

示例：

```go
db := testutil.OpenSQLiteMemory(t)
repo := &sqlrepo.UserRepository{DB: db}
```

---

## 常见问题

**Q：能否改成 `sqlx.Connect`？**  
可以。`sqlx.Connect` 内部 `Open` + `Ping`，返回 `*sqlx.DB`；若仍需调用 `migrator.Run`，需拿到底层 `*sql.DB`。当前代码显式 `sql.Open` + `sqlx.NewDb`，便于一眼看出迁移与 ORM 扩展的分界。

**Q：sqlx 会缓存 prepared statement 吗？**  
`GetContext` 等会按 sqlx 行为处理；高频路径可再考虑 `Preparex`。当前仓储以简单 CRUD 为主，未强制预编译。

**Q：领域层能否直接依赖 sqlx？**  
不应。仓储接口在 `domain`，实现在 `infrastructure/persistence/sql`，符合仓库 DDD 分层规则。

---

## 相关路径

| 路径 | 说明 |
|------|------|
| `infrastructure/persistence/sql/user_repository.go` | 用户表访问（sqlx + `userRow`） |
| `infrastructure/persistence/sql/oauth_repository.go` | OAuth 绑定表访问 |
| `internal/testutil/sqlite.go` | 测试用内存库 + sqlx |
| `cmd/main.go` | 打开 DB、`sqlx.NewDb`、迁移、注入仓储 |
| `internal/migrator` | 仅 `database/sql` |
| `platform/db` | SQL 迁移脚本 |

更多认证与注册流程见 [`docs/auth-login-registration.md`](auth-login-registration.md)。
