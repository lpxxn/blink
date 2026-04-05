# 数据库访问：GORM

本仓库在 **`infrastructure/persistence/gormdb`** 使用 [GORM](https://gorm.io/) 作为持久化实现：通过 **模型（Model）** 映射 `platform/db` 中的表，仓储实现 `domain` 中定义的接口，并在模型与领域实体之间做显式转换。

SQLite 驱动采用纯 Go 的 **[github.com/glebarez/sqlite](https://github.com/glebarez/sqlite)**（底层基于 `modernc.org/sqlite`），与 `cmd/migrate` 使用的 `modernc.org/sqlite` 在语义上一致，便于本地与 CI 不依赖 CGO。

---

## 依赖

| 模块 | 作用 |
|------|------|
| `gorm.io/gorm` | ORM 核心 |
| `github.com/glebarez/sqlite` | GORM 的 SQLite Dialector（纯 Go） |

安装示例：

```bash
go get gorm.io/gorm@latest
go get github.com/glebarez/sqlite@latest
```

---

## 连接与迁移

### 主进程（`cmd/main.go`）

1. **`gorm.Open(glsqlite.Open(dsn), &gorm.Config{})`** 得到 `*gorm.DB`。
2. **`db.DB()`** 取出底层 `*sql.DB`，用于：
   - **`migrator.Run(sqldb, "sqlite", migDir)`**（`internal/migrator` 仍只依赖 `database/sql`）；
   - **`Ping()`**、**`SetMaxOpenConns(1)`**（SQLite 文件库推荐单连接，减少锁竞争）。
3. **`defer sqldb.Close()`** 关闭底层连接池。

### DSN

- 环境变量 **`BLINK_DATABASE_DSN`**，默认见 `cmd/main.go`（含 busy_timeout、WAL 等 pragma）。

### 与迁移的关系

- **表结构**仍以 **`platform/db/*.sql`** 为权威，由 **`cmd/migrate`** 执行；应用启动时主进程也会对同一库跑一遍迁移（未执行的文件才会应用）。
- 本仓库**未**在启动时调用 `AutoMigrate` 替代 SQL 迁移；若日后引入，需与现有迁移策略统一，避免双头维护。

---

## 模型与领域边界

- **`UserModel`**、**`OAuthIdentityModel`** 定义在 `gormdb/models.go`，带 `gorm` 标签，并实现 **`TableName()`** 与真实表名一致（`users`、`oauth_identities`）。
- **`DeletedAt gorm.DeletedAt`** 与表中 **`deleted_at`** 对齐；查询默认排除软删行（与原先 `deleted_at IS NULL` 条件一致）。
- **`domain/user.User`**、**`domain/oauth.Identity`** **不**引用 GORM；仓储内通过 **`domainToUserModel` / `userModelToDomain`** 等函数转换。

---

## 仓储

| 类型 | 文件 | 说明 |
|------|------|------|
| `UserRepository` | `user_repository.go` | `Create`、`FindByEmail`、`GetByID`、`UpdateLastLogin` |
| `OAuthRepository` | `oauth_repository.go` | `Create`、`FindByProviderSubject` |

所有数据库操作使用 **`WithContext(ctx)`** 传递请求上下文。

---

## 测试

- **`internal/testutil.OpenSQLiteMemory(t)`** 返回 **`*gorm.DB`**：内存 SQLite + 执行 `platform/db` 全量迁移，供 `application/*` 与 HTTP 测试注入仓储。

---

## 切换到其他数据库（说明）

当前 SQL 迁移与 GORM 模型按 **SQLite** 习惯编写（如部分驱动下整型/时间戳行为）。若迁移到 PostgreSQL / MySQL：

1. 在 **`platform/db`** 中维护对应方言的迁移或统一策略；
2. 为 GORM 更换 Dialector（如 `gorm.io/driver/postgres`）；
3. 复查 **`Updates` map**、**软删**、**唯一索引** 在各驱动下的行为。

---

## 相关路径

| 路径 | 说明 |
|------|------|
| `infrastructure/persistence/gormdb/` | 模型 + 仓储 |
| `cmd/main.go` | `gorm.Open`、迁移、注入 |
| `internal/testutil/sqlite.go` | 测试用 GORM + 内存库 |
| `internal/migrator` | 仅 `database/sql` |
| `platform/db` | DDL 迁移脚本 |

认证与注册流程见 [`docs/auth-login-registration.md`](auth-login-registration.md)。
