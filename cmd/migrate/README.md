# migrate — 数据库建表与版本迁移 CLI

本目录下的程序会连接指定数据库，按**文件名字典序**执行 `platform/db`（或 `-dir` 指定目录）中的 `*.sql`。每成功执行一个文件，会把**完整文件名**写入表 `schema_migrations.version`；再次运行时会**跳过**已记录的文件，实现幂等升级。

表结构、字段含义见仓库根下 [`platform/db/SCHEMA.md`](../../platform/db/SCHEMA.md)；迁移 SQL 文件在 [`platform/db/`](../../platform/db/)。

---

## 前置条件

- 已安装 **Go**（与 `go.mod` 中版本一致或更高）。
- 在**仓库根目录**执行下文命令（保证默认 `-dir platform/db` 能指到正确路径）。若在其它目录执行，必须显式传入 `-dir /绝对或相对路径/platform/db`。
- **PostgreSQL**：需事先存在目标数据库（例如 `CREATE DATABASE blink;`），且用户具备建表权限。CLI **不会**替你 `CREATE DATABASE`。
- **MySQL**：同上，需已有 database，DSN 里带上库名。

---

## 命令格式

```bash
go run ./cmd/migrate [-driver DRIVER] [-dsn DSN] [-dir DIR]
```

安装为二进制后，可将 `go run ./cmd/migrate` 换成可执行文件路径，参数相同。

```bash
go install ./cmd/migrate
# 确保 $GOPATH/bin 或 $GOBIN 在 PATH 中
migrate -h
```

---

## 参数说明

| 标志 | 默认值 | 说明 |
|------|--------|------|
| `-driver` | `sqlite` | 数据库驱动：`sqlite`、`mysql`、`postgres`（也接受 `sqlite3`、`pg`、`postgresql` 等别名，见代码 `normalizeDriver`）。 |
| `-dsn` | 见下文 SQLite | 数据源连接串，**随驱动变化**，必须符合 `database/sql` 对应驱动的格式。 |
| `-dir` | `platform/db` | 存放 `0000_*.sql`、`0001_*.sql` … 的目录；目录内所有 `.sql` 会参与排序与执行判断。 |

运行时会向 **stderr** 打印一行当前解析后的 `driver`、打码后的 `dsn`（MySQL/PostgreSQL 会遮蔽密码）以及 `dir`，便于排查环境是否一致。

向 **stdout** 打印每个新应用的文件名（`applied …`），结束时打印 `migrations finished (N applied)` 或 `no pending migrations`。

---

## 执行逻辑（摘要）

1. 连接数据库并 `Ping`。
2. 将连接池 `MaxOpenConns` 设为 `1`（减轻 SQLite 多连接锁竞争）。
3. 读取 `schema_migrations` 中已有 `version`；若表尚不存在（首次），视为无任何已执行记录。
4. 对 `-dir` 下所有 `*.sql` 按**文件名**排序（因此 `0000_` 必须排在 `0001_` 前）。
5. 对每个**尚未出现在 `schema_migrations` 的文件**：
   - 读入全文，按顶层分号拆成多条语句（**忽略**注释、`'字符串'` 内的分号，见 `internal/migrator`）。
   - 在**事务**中依次执行这些语句，再 `INSERT INTO schema_migrations (version) VALUES ('文件名')`，提交。
6. 若某文件执行失败，该文件不会写入 `schema_migrations`，需修正 SQL 或环境后重试。

**注意**

- **MySQL**：部分 `CREATE TABLE` 等 DDL 会触发**隐式提交**，与后续 `INSERT schema_migrations` 可能不在同一事务语义内。仓库内 migration 已尽量使用 `IF NOT EXISTS` 等，便于失败后重跑。
- **版本标识**是**文件名**（如 `0002_create_post.sql`），若重命名已发布过的文件，会被当作新 migration 再执行一遍，可能冲突；生产环境应避免改已执行文件名。

---

## SQLite 详细说明

使用纯 Go 驱动 **[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)**，`database/sql` 驱动名为 `sqlite`。

DSN 采用 **URI 形式**（官方推荐），常用查询参数：

| 参数 | 含义 |
|------|------|
| `mode=rwc` | 读写；若库文件不存在则创建。 |
| `cache=shared` | 共享缓存，多连接场景更稳（本 CLI 仍限制单连接）。 |
| `mode=memory` | 内存库，进程结束数据即失，适合本地试跑。 |

### 示例：默认（仓库根）

在当前目录生成（或打开）`blink.db`，并执行 `platform/db` 下全部待执行 migration：

```bash
cd /path/to/blink   # 仓库根
go run ./cmd/migrate
```

等价于：

```bash
go run ./cmd/migrate \
  -driver sqlite \
  -dsn 'file:blink.db?cache=shared&mode=rwc' \
  -dir platform/db
```

### 示例：自定义文件路径（相对仓库根）

```bash
go run ./cmd/migrate -dsn 'file:./data/app.db?cache=shared&mode=rwc'
```

需保证 `./data` 已存在，或 SQLite 能创建父目录（视系统与驱动行为而定；生产环境建议先 `mkdir -p data`）。

### 示例：绝对路径

```bash
go run ./cmd/migrate -dsn 'file:/var/lib/blink/app.db?cache=shared&mode=rwc'
```

路径含空格时务必用引号包住整个 `-dsn` 参数。

### 示例：内存数据库（一次性）

```bash
go run ./cmd/migrate -dsn 'file:mem.db?mode=memory&cache=shared'
```

适合 CI 或本地验证 migration 能否跑通；**不要**用于持久化数据。

### 示例：指定迁移目录

从其它工作目录执行时：

```bash
go run ./cmd/migrate \
  -dsn 'file:/tmp/test.db?cache=shared&mode=rwc' \
  -dir /path/to/blink/platform/db
```

---

## PostgreSQL 详细说明

使用 **[github.com/lib/pq](https://pkg.go.dev/github.com/lib/pq)**，驱动名 `postgres`。

DSN 常用 **URL**：

```text
postgres://[用户[:密码]@][主机][:端口][/数据库][?参数]
```

常见查询参数：

| 参数 | 说明 |
|------|------|
| `sslmode=disable` | 关闭 SSL（仅建议开发/内网；生产请用 `require` 等）。 |
| `sslmode=require` | 要求 SSL。 |

密码中含 `@`、`:`、`/` 等字符时，必须进行 **URL 编码**后再放入 DSN。

### 准备数据库

```sql
CREATE DATABASE blink OWNER your_user;
```

确保该用户对 `blink` 库有 `CREATE TABLE` 等权限。

### 示例：本机、密码认证、禁用 SSL

```bash
cd /path/to/blink
go run ./cmd/migrate \
  -driver postgres \
  -dsn 'postgres://blink:your_password@127.0.0.1:5432/blink?sslmode=disable' \
  -dir platform/db
```

### 示例：本机 socket / trust（无密码用户）

将 `USER` 换成实际系统用户名或数据库角色名：

```bash
go run ./cmd/migrate \
  -driver postgres \
  -dsn 'postgres://USER@localhost:5432/blink?sslmode=disable' \
  -dir platform/db
```

### 示例：远程主机与 SSL

```bash
go run ./cmd/migrate \
  -driver postgres \
  -dsn 'postgres://blink:secret@db.example.com:5432/blink?sslmode=require' \
  -dir platform/db
```

### 验证是否已记录 migration

连接 `psql` 后：

```sql
SELECT version, applied_at FROM schema_migrations ORDER BY version;
```

---

## MySQL（可选）

驱动为 `github.com/go-sql-driver/mysql`，DSN 形式示例：

```bash
go run ./cmd/migrate \
  -driver mysql \
  -dsn 'user:password@tcp(127.0.0.1:3306)/blink?parseTime=true' \
  -dir platform/db
```

需事先 `CREATE DATABASE blink`，且用户对该库有 DDL 权限。详见上文「执行逻辑」中关于 DDL 与事务的说明。

---

## 常见问题

**Q：`no such file or directory` 或找不到 sql**

- 确认当前工作目录是否为仓库根，或 `-dir` 是否指向包含 `0000_schema_migrations.sql` 的目录。

**Q：第二次运行仍尝试执行已跑过的文件**

- 检查是否连到了**另一个**数据库文件或另一个 PostgreSQL/MySQL 库（对比 CLI 打印的 `dsn` 打码行与实际连接）。
- 检查 `schema_migrations` 表是否被清空或换库。

**Q：PostgreSQL 报错 `relation "schema_migrations" does not exist`**

- 首次运行应执行 `0000_schema_migrations.sql` 创建该表；若 `0000` 被跳过或失败，需手动排查 `0000` 是否存在于 `-dir` 且排序在最前。

**Q：需要回滚某个 migration**

- 当前 CLI **只支持向前应用**，不提供自动 down。回滚需手写 SQL 或从备份恢复，并视情况手动删除 `schema_migrations` 中对应行（**高风险**，仅建议在明确后果下操作）。

---

## 与 `-h` 的关系

命令行 **`go run ./cmd/migrate -h`** 会打印简短说明与本页路径提示；**完整步骤、注意点与可复制示例以本文档为准**。

---

## 相关代码

| 路径 | 作用 |
|------|------|
| [`main.go`](./main.go) | 解析参数、打开连接、调用 migrator |
| [`internal/migrator`](../../internal/migrator/) | 扫描文件、拆分 SQL、事务执行、读写 `schema_migrations` |
