# GORM 事务与仓储（`*gorm.DB`）

## 核心结论

在 GORM v2 里，**根连接**和**事务会话**的类型都是 **`*gorm.DB`**。`db.Transaction(func(tx *gorm.DB) error { ... })` 回调里的 `tx` 不是另一种类型，而是绑定到当前事务的 `*gorm.DB`。

因此仓储结构体里的字段 **`DB *gorm.DB`** 既可以注入**根 DB**，也可以注入 **`tx`**。事务内只要保证**所有写操作都用同一个 `tx` 构造出来的仓储**，就会落在同一事务里。

---

## 推荐写法：`Transaction` / `WithTransaction`

### 1. 直接使用 GORM

```go
import "gorm.io/gorm"

err := db.Transaction(func(tx *gorm.DB) error {
    users := &gormdb.UserRepository{DB: tx}
    oauth := &gormdb.OAuthRepository{DB: tx}

    if err := users.Create(ctx, user); err != nil {
        return err
    }
    if err := oauth.Create(ctx, identity); err != nil {
        return err
    }
    return nil // 提交
})
// err != nil 时整段已回滚
```

### 2. 使用本仓库封装（与上面等价）

`infrastructure/persistence/gormdb/transaction.go`：

```go
err := gormdb.WithTransaction(db, func(tx *gorm.DB) error {
    // 同样用 {DB: tx} 构造仓储
    ...
})
```

---

## 不要做的事

| 错误做法 | 后果 |
|----------|------|
| 事务里仍用 `UserRepository{DB: rootDB}` | 写落在事务外或另一连接，**无法与 `tx` 一起提交/回滚** |
| 一部分用 `tx`、一部分用根 `db` | 同一业务原子性被破坏 |
| 在 `tx` 上开 goroutine 各用各的 `*gorm.DB` | 除非显式传递同一个 `tx`，否则易错 |

---

## 应用层 / 用例层怎么接

常见两种风格：

1. **用例接收 `*gorm.DB`**（根库），内部调用 `WithTransaction(db, func(tx *gorm.DB) error { ... })`，在闭包里 new 仓储。
2. **工厂函数**：`func NewUserRepository(db *gorm.DB) *UserRepository`，事务里传入 `tx` 即可。

若希望仓储**始终**从同一入口拿 DB，可再包一层：

```go
type UnitOfWork struct {
    db *gorm.DB
}

func (u *UnitOfWork) InTx(fn func(tx *gorm.DB) error) error {
    return gormdb.WithTransaction(u.db, fn)
}
```

---

## 可运行示例（仓库内）

- **`infrastructure/persistence/gormdb/transaction_test.go`**  
  - 提交：同一事务内插入 user + oauth，提交后用根 `gdb` 能读到。  
  - 回滚：插入 user 后返回错误，确认库中无该用户。  
  - 辅助函数接收 `*gorm.DB`：演示把 `tx` 传给子函数，子函数内仍用 `{DB: db}`。

- **`infrastructure/persistence/gormdb/transaction.go`**：`WithTransaction` 封装。

---

## 手动 `Begin` / `Commit` / `Rollback`（少用）

```go
tx := db.Begin()
if tx.Error != nil {
    return tx.Error
}
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

users := &gormdb.UserRepository{DB: tx}
if err := users.Create(ctx, u); err != nil {
    tx.Rollback()
    return err
}
return tx.Commit().Error
```

优先用 **`Transaction` / `WithTransaction`**，由 GORM 统一处理提交与回滚，代码更短、更不易漏 `Rollback`。

---

## 延伸阅读

- [GORM 事务文档](https://gorm.io/docs/transactions.html)  
- 本仓库 ORM 总览：[`docs/database-gorm.md`](database-gorm.md)
