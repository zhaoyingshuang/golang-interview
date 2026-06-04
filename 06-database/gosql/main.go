package main

import "fmt"

// ============================================================
// Go database/sql
// ============================================================
// 【面试高频】连接池原理, 预处理, 事务, sql.Scan

func main() {
	connectionPool()
	preparedStmt()
	transactions()
	scanMechanism()
	productionConfig()
}

func connectionPool() {
	fmt.Println("=== 1. 连接池原理 ===")
	fmt.Println()
	fmt.Println("database/sql DB 结构管理连接池:")
	fmt.Println("  MaxOpenConns: 最大连接数(默认无限制)")
	fmt.Println("  MaxIdleConns: 最大空闲连接(默认2)")
	fmt.Println("  ConnMaxLifetime: 连接最大存活时间(默认无限制)")
	fmt.Println()
	fmt.Println("获取连接流程:")
	fmt.Println("  1. 检查空闲连接池 → 有则复用")
	fmt.Println("  2. 当前连接数 < MaxOpenConns → 创建新连接")
	fmt.Println("  3. 等待其他连接释放(阻塞直到获取或超时)")
	fmt.Println()
	fmt.Println("归还连接:")
	fmt.Println("  空闲数 < MaxIdleConns → 放回池中")
	fmt.Println("  空闲数 >= MaxIdleConns → 关闭连接")
	fmt.Println()
	fmt.Println("**面试追问**: 为什么需要 MaxIdleConns?")
	fmt.Println("  避免过多空闲连接占用DB资源")
	fmt.Println("  但也不能太小(否则频繁创建连接)")
	fmt.Println()
}

func preparedStmt() {
	fmt.Println("=== 2. 预处理语句 ===")
	fmt.Println()
	fmt.Println("Prepare 流程:")
	fmt.Println("  1. 发送 SQL 模板到DB(server端预编译)")
	fmt.Println("  2. DB 返回 statement ID")
	fmt.Println("  3. 执行时只发送参数(二进制协议, 非拼接SQL)")
	fmt.Println()
	fmt.Println("优点:")
	fmt.Println("  1. 防SQL注入(参数不参与SQL解析)")
	fmt.Println("  2. 重复执行更快(只传参数)")
	fmt.Println()
	fmt.Println("Go 中的连接绑定问题:")
	fmt.Println("  Prepare 返回 Stmt 绑定到特定连接")
	fmt.Println("  执行时需要同一个连接 → 连接池效率降低")
	fmt.Println("  database/sql 的优化: 自动在各连接上缓存 Stmt")
	fmt.Println()
	fmt.Println("最佳实践:")
	fmt.Println("  单次查询: db.QueryContext(ctx, sql, args...) → 自动预处理")
	fmt.Println("  批量操作: stmt = db.Prepare → 多次 Exec → stmt.Close")
	fmt.Println()
}

func transactions() {
	fmt.Println("=== 3. 事务处理 ===")
	fmt.Println()
	fmt.Println("Go 事务机制:")
	fmt.Println("  tx, err := db.BeginTx(ctx, nil)")
	fmt.Println("  // tx 内所有操作使用同一个连接!")
	fmt.Println("  tx.Exec / tx.Query / tx.QueryRow")
	fmt.Println("  tx.Commit() 或 tx.Rollback()")
	fmt.Println()
	fmt.Println("推荐模板:")
	fmt.Println("  tx, _ := db.BeginTx(ctx, nil)")
	fmt.Println("  defer tx.Rollback() // 提交后Rollback无操作")
	fmt.Println("  // ... 操作 ...")
	fmt.Println("  if err := tx.Commit(); err != nil { return err }")
	fmt.Println()
	fmt.Println("context 超时控制:")
	fmt.Println("  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)")
	fmt.Println("  tx, _ := db.BeginTx(ctx, nil) // 超时自动Rollback")
	fmt.Println()
	fmt.Println("**面试追问**: tx 内用 db.Query 会怎样?")
	fmt.Println("  → 不会在事务内! 会获取新连接, 不受事务保护")
	fmt.Println("  → 必须用 tx.Query/tx.Exec")
	fmt.Println()
}

func scanMechanism() {
	fmt.Println("=== 4. sql.Scan 原理 ===")
	fmt.Println()
	fmt.Println("Rows.Scan 实现原理:")
	fmt.Println("  1. 从MySQL协议包中解析列数据(二进制/文本协议)")
	fmt.Println("  2. 根据目标类型进行类型转换")
	fmt.Println("  3. 写入传入的指针")
	fmt.Println()
	fmt.Println("NULL 处理:")
	fmt.Println("  使用 sql.NullString/sql.NullInt64 等")
	fmt.Println("  或指针类型: *string/*int64 (NULL → nil)")
	fmt.Println()
	fmt.Println("自定义 Scanner:")
	fmt.Println("  type Scanner interface { Scan(src any) error }")
	fmt.Println("  实现此接口可自定义类型转换")
	fmt.Println()
	fmt.Println("常见陷阱:")
	fmt.Println("  1. SELECT * 列顺序变化 → Scan 参数不匹配")
	fmt.Println("  2. UNSIGNED BIGINT → uint64 需要 ✗, 用 string 接收")
	fmt.Println("  3. 时间类型 → 用 time.Time 接收, 注意时区")
	fmt.Println()
}

func productionConfig() {
	fmt.Println("=== 5. 生产配置 ===")
	fmt.Println()
	fmt.Println("连接池参数(经验值):")
	fmt.Println("  MaxOpenConns:    CPU核心数 * 2 + 磁盘数 (或DB max_connections/应用实例数)")
	fmt.Println("  MaxIdleConns:    = MaxOpenConns (避免频繁创建)")
	fmt.Println("  ConnMaxLifetime: < DB wait_timeout (默认8h) → 建议30min")
	fmt.Println("  ConnMaxIdleTime: 15min (清理长时间空闲连接)")
	fmt.Println()
	fmt.Println("Go 1.15+ Config 初始化:")
	fmt.Println("  sql.OpenDB(mysql.Config{Conn: conn})")
	fmt.Println()
	fmt.Println("监控指标:")
	fmt.Println("  db.Stats().OpenConnections / Idle / InUse / WaitCount / WaitDuration")
	fmt.Println()
}
