---
title: Go database/sql 原理
---

## 1. database/sql 接口设计

Go 的 `database/sql` 包定义了一套数据库操作的抽象接口，通过 `database/sql/driver` 包实现具体数据库驱动。

### driver 注册机制

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"  // 注册 MySQL 驱动
)

func main() {
    // driver 注册后通过 DSN 识别驱动
    db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/test?parseTime=true")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}
```

### 核心接口

`database/sql/driver` 定义了以下关键接口：

| 接口 | 职责 |
|------|------|
| `driver.Driver` | 注册驱动，创建连接 |
| `driver.Conn` | 单个数据库连接 |
| `driver.Stmt` | 预处理语句 |
| `driver.Tx` | 事务 |
| `driver.Rows` | 查询结果集 |

::: tip
`sql.Open()` 只是验证参数格式，并不会真正建立连接。真正的连接创建发生在第一次查询时（惰性初始化）。
:::

## 2. 连接池管理

`database/sql` 内置连接池，通过以下参数控制连接池行为：

```go
func setupDB() *sql.DB {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }

    // 最大打开连接数（默认 0 表示无限制）
    db.SetMaxOpenConns(25)

    // 最大空闲连接数（默认 2）
    db.SetMaxIdleConns(10)

    // 连接最大存活时间（必须小于数据库的 wait_timeout）
    db.SetConnMaxLifetime(5 * time.Minute)

    // 空闲连接最大存活时间
    db.SetConnMaxIdleTime(2 * time.Minute)

    return db
}
```

### 连接池工作原理

```go
// 连接池获取连接的流程：
// 1. 检查空闲连接池中是否有可用连接
// 2. 如果有空闲连接，直接返回
// 3. 如果当前连接数 < MaxOpenConns，创建新连接
// 4. 如果达到 MaxOpenConns，等待其他连接释放

// 连接健康检查
err := db.PingContext(ctx) // 验证连接是否存活
```

::: warning
`SetMaxOpenConns` 设置过小会导致高并发时请求排队等待；设置过大会压垮数据库。建议根据数据库的 `max_connections` 和应用实例数计算合理值。
:::

### 监控连接池状态

```go
func monitorPool(ctx context.Context, db *sql.DB) {
    stats := db.Stats()
    log.Printf("MaxOpenConnections: %d", stats.MaxOpenConnections)
    log.Printf("OpenConnections: %d", stats.OpenConnections)
    log.Printf("InUse: %d", stats.InUse)
    log.Printf("Idle: %d", stats.Idle)
    log.Printf("WaitCount: %d", stats.WaitCount)
    log.Printf("WaitDuration: %v", stats.WaitDuration)
}
```

## 3. 预处理语句

### Prepare 原理

预处理语句将 SQL 编译和参数绑定分为两步：

```go
// 预处理查询
func GetUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    stmt, err := db.PrepareContext(ctx, "SELECT id, name, email FROM users WHERE id = ?")
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    var user User
    err = stmt.QueryRowContext(ctx, id).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### 参数绑定与 SQL 注入防护

```go
// 安全：使用参数化查询
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE name = ?", name)

// 危险：字符串拼接（永远不要这样做！）
rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name))
```

::: warning
永远不要通过字符串拼接构造 SQL 语句。参数化查询（`?` 占位符）由驱动层处理参数转义，从根本上防止 SQL 注入。
:::

### 预处理语句在连接池中的行为

```go
// database/sql 内部会对 Prepare 做优化：
// - 首次调用时在某个连接上 Prepare
// - 后续使用时，如果分配到不同连接，会自动重新 Prepare
// - 这意味着频繁 Prepare/Close 有性能开销

// 推荐做法：在循环外 Prepare
stmt, err := db.Prepare("INSERT INTO logs (msg) VALUES (?)")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

for _, msg := range messages {
    _, err = stmt.Exec(msg)
    if err != nil {
        log.Printf("insert error: %v", err)
    }
}
```

## 4. 事务处理

### 基本事务操作

```go
func TransferMoney(ctx context.Context, db *sql.DB, from, to int, amount float64) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    // 使用 defer + 匿名函数确保事务一定被处理
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p) // 重新 panic
        }
    }()

    var balance float64
    err = tx.QueryRowContext(ctx,
        "SELECT balance FROM accounts WHERE id = ? FOR UPDATE", from,
    ).Scan(&balance)
    if err != nil {
        tx.Rollback()
        return err
    }

    if balance < amount {
        tx.Rollback()
        return fmt.Errorf("insufficient balance")
    }

    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from)
    if err != nil {
        tx.Rollback()
        return err
    }

    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to)
    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}
```

### 设置隔离级别

```go
// 设置事务隔离级别
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
    ReadOnly:  false,
})
```

### 嵌套事务（Savepoint）

```go
func NestedTransaction(ctx context.Context, db *sql.DB) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 主事务操作
    _, err = tx.ExecContext(ctx, "INSERT INTO orders (user_id) VALUES (1)")
    if err != nil {
        return err
    }

    // 创建 Savepoint
    _, err = tx.ExecContext(ctx, "SAVEPOINT sp1")
    if err != nil {
        return err
    }

    // 子操作（可能失败）
    _, err = tx.ExecContext(ctx, "INSERT INTO order_items (order_id, product_id) VALUES (1, 999)")
    if err != nil {
        // 回滚到 Savepoint，不影响主事务
        tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT sp1")
    }

    // 释放 Savepoint
    tx.ExecContext(ctx, "RELEASE SAVEPOINT sp1")

    return tx.Commit()
}
```

## 5. 常见陷阱

### 连接泄漏

```go
// 错误示例：忘记关闭 Rows
func LeakingQuery(db *sql.DB) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return
    }
    // 没有 defer rows.Close() -> 连接泄漏！
    for rows.Next() {
        // ...
    }
}

// 正确写法
func CorrectQuery(db *sql.DB) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return
    }
    defer rows.Close() // 必须关闭！
    for rows.Next() {
        // ...
    }
}
```

::: warning
`Query()` 和 `QueryRow()` 返回的 `Rows` 必须调用 `Close()` 释放连接。`QueryRow()` 虽然只返回一行，但其内部使用的 `Rows` 也需要通过 `Scan()` 触发关闭。
:::

### NULL 值处理

```go
import "database/sql"

type User struct {
    ID   int
    Name sql.NullString  // 处理可能为 NULL 的字符串
    Age  sql.NullInt64   // 处理可能为 NULL 的整数
}

func QueryWithNull(db *sql.DB) {
    var u User
    row := db.QueryRow("SELECT id, name, age FROM users WHERE id = 1")
    err := row.Scan(&u.ID, &u.Name, &u.Age)
    if err != nil {
        log.Fatal(err)
    }

    // 检查是否为 NULL
    if u.Name.Valid {
        fmt.Println("Name:", u.Name.String)
    } else {
        fmt.Println("Name is NULL")
    }
}

// 或者使用指针类型
type User2 struct {
    ID   int
    Name *string // nil 表示 NULL
    Age  *int    // nil 表示 NULL
}
```

### 忽略错误

```go
// 错误示例：忽略 rows.Err()
func BadIterate(db *sql.DB) {
    rows, _ := db.Query("SELECT * FROM users")
    defer rows.Close()
    for rows.Next() {
        // ...
    }
    // 必须检查 rows.Err()！循环可能因为错误而提前退出
}

// 正确写法
func GoodIterate(db *sql.DB) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    for rows.Next() {
        // ...
    }
    if err = rows.Err(); err != nil {
        log.Fatal(err) // 检查迭代过程中的错误
    }
}
```

::: warning 面试高频问题
1. **sql.Open 和 sql.Ping 的区别？**
   - `sql.Open` 只创建 DB 对象并验证 DSN 格式，不会建立实际连接。`sql.Ping` 会真正建立连接并验证连通性。

2. **database/sql 的连接池是如何工作的？**
   - 内部维护空闲连接池和活跃连接计数。获取连接时优先从空闲池取，不够时新建（不超过 MaxOpenConns），满了则等待。使用完毕归还空闲池或关闭。

3. **为什么 QueryRow 也需要 Scan？**
   - `QueryRow` 返回的 `Row` 内部持有 `Rows`，只有调用 `Scan()` 才会触发 `Rows.Close()` 释放连接。不调用 `Scan()` 会导致连接泄漏。

4. **如何在 Go 中处理数据库 NULL 值？**
   - 使用 `sql.NullString`/`sql.NullInt64` 等类型，或使用指针类型（`*string`/`*int`），nil 表示 NULL。

5. **database/sql 中事务的连接是如何管理的？**
   - 事务 (`sql.Tx`) 绑定到单个连接，事务内的所有操作都使用同一连接。事务 Commit 或 Rollback 后连接归还连接池。
:::
