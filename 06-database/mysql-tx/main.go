package main

import "fmt"

// ============================================================
// MySQL 事务与锁
// ============================================================
//
// 【面试高频问题】
// 1. 事务的四大特性(ACID)和四种隔离级别？
// 2. MVCC 的实现原理？
// 3. InnoDB 有哪些锁？行锁/表锁/间隙锁/临键锁？
// 4. 死锁如何检测和避免？
// 5. 乐观锁和悲观锁的区别？

func main() {
	isolationLevels()
	mvcc()
	lockTypes()
	deadlock()
	optimisticVsPessimistic()
	productionTx()
}

// ----------------------------------------------------------
// 1. ACID 与隔离级别
// ----------------------------------------------------------
// ACID:
//   A(原子性): 事务操作要么全部成功要么全部回滚(Undo Log保证)
//   C(一致性): 数据从一个一致状态到另一个一致状态(应用层保证)
//   I(隔离性): 并发事务互不影响(Lock + MVCC保证)
//   D(持久性): 提交后数据不丢失(Redo Log保证)
//
// 四种隔离级别:
//   READ UNCOMMITTED  - 脏读/不可重复读/幻读 都可能
//   READ COMMITTED   - 防止脏读(Oracle/PostgreSQL默认)
//   REPEATABLE READ  - 防止脏读+不可重复读(MySQL默认)
//   SERIALIZABLE     - 全部防止, 性能最差
//
// **面试追问**: MySQL RR级别能解决幻读吗？
//   → 快照读通过MVCC解决, 当前读通过临键锁解决, 但有特例不能完全解决
func isolationLevels() {
	fmt.Println("=== 1. ACID 与隔离级别 ===")
	fmt.Println("隔离级别          脏读  不可重复读  幻读")
	fmt.Println("READ UNCOMMITTED  可能    可能      可能")
	fmt.Println("READ COMMITTED    防止    可能      可能")
	fmt.Println("REPEATABLE READ   防止    防止      可能*")
	fmt.Println("SERIALIZABLE      防止    防止      防止")
	fmt.Println()
	fmt.Println("MySQL 默认 RR, 通过 MVCC + 临键锁基本解决幻读")
	fmt.Println("* RR 下特定场景仍可能出现幻读:")
	fmt.Println("  事务A SELECT(快照读无数据) → 事务B INSERT+COMMIT")
	fmt.Println("  → 事务A UPDATE(当前读命中) → 事务A SELECT(快照读出现数据)")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. MVCC 实现原理
// ----------------------------------------------------------
// Multi-Version Concurrency Control: 多版本并发控制
// 核心组件:
//   1. 隐藏列: DB_TRX_ID(事务ID) + DB_ROLL_PTR(回滚指针)
//   2. Undo Log: 保存数据的历史版本, 形成版本链
//   3. ReadView: 决定哪个版本对当前事务可见
//
// ReadView 四个字段:
//   creator_trx_id: 创建该 ReadView 的事务ID
//   m_ids: 创建时所有活跃事务ID列表
//   min_trx_id: m_ids 中最小的事务ID
//   max_trx_id: 下一个将分配的事务ID
//
// 可见性判断:
//   版本 trx_id == creator_trx_id → 可见(自己修改的)
//   trx_id < min_trx_id → 可见(事务已提交)
//   trx_id >= max_trx_id → 不可见(在ReadView之后开始)
//   min_trx_id <= trx_id < max_trx_id 且不在 m_ids 中 → 可见(已提交)
//
// RC vs RR 的 MVCC 区别:
//   RC: 每次SELECT都创建新ReadView → 能看到其他已提交事务的修改
//   RR: 只在第一次SELECT创建ReadView → 整个事务看到一致的快照
func mvcc() {
	fmt.Println("=== 2. MVCC 实现原理 ===")
	fmt.Println("版本链: 最新数据 → Undo Log旧版本1 → Undo Log旧版本2 → ...")
	fmt.Println()
	fmt.Println("ReadView 可见性规则:")
	fmt.Println("  trx_id == 自己的ID       → 可见(自己改的)")
	fmt.Println("  trx_id < min_trx_id      → 可见(创建前已提交)")
	fmt.Println("  trx_id 在 m_ids 中       → 不可见(未提交)")
	fmt.Println("  trx_id >= max_trx_id     → 不可见(ReadView之后)")
	fmt.Println("  其他                      → 可见(已提交)")
	fmt.Println()
	fmt.Println("RC: 每次 SELECT 新建 ReadView")
	fmt.Println("RR: 只建一次 ReadView → 可重复读")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. InnoDB 锁类型
// ----------------------------------------------------------
// 锁分类:
//   全局锁: FLUSH TABLES WITH READ LOCK (备份用)
//   表级锁: 表锁/元数据锁(MDL)/意向锁/自增锁
//   行级锁: Record Lock(行锁)/Gap Lock(间隙锁)/Next-Key Lock(临键锁)
//
// 临键锁 = 行锁 + 间隙锁, 锁住记录 + 前面的间隙
// 左开右闭区间: (前一条记录, 当前记录]
//
// **面试追问**: 什么时候加行锁, 什么时候加临键锁?
//   等值查询唯一索引: 存在→行锁; 不存在→间隙锁
//   等值查询普通索引: 存在→行锁+间隙锁(两侧); 不存在→间隙锁
//   范围查询: 扫到的都加临键锁
func lockTypes() {
	fmt.Println("=== 3. InnoDB 锁类型 ===")
	fmt.Println("行锁(Record Lock):    锁定索引记录")
	fmt.Println("间隙锁(Gap Lock):     锁定索引记录之间的间隙")
	fmt.Println("临键锁(Next-Key Lock): 行锁+间隙锁 (默认行锁算法)")
	fmt.Println()
	fmt.Println("加锁规则(等值查询唯一索引):")
	fmt.Println("  记录存在 → 行锁")
	fmt.Println("  记录不存在 → 间隙锁(防止幻读)")
	fmt.Println("加锁规则(等值查询普通索引):")
	fmt.Println("  记录存在 → 临键锁(两条边界) + 行锁(满足条件的)")
	fmt.Println("  记录不存在 → 间隙锁")
	fmt.Println("加锁规则(范围查询):")
	fmt.Println("  扫描到的记录加临键锁")
	fmt.Println()
	fmt.Println("意向锁(IS/IX):")
	fmt.Println("  表级锁, 快速判断表中是否有行锁")
	fmt.Println("  事务获取行锁前先获取意向锁")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 死锁
// ----------------------------------------------------------
// 死锁: 两个事务互相等待对方持有的锁
//
// 检测方式:
//   1. 等待图(Wait-For Graph): 每个事务是节点, 等待关系是边, 有环=死锁
//   2. InnoDB 自动检测, 回滚代价最小的事务
//
// 避免:
//   1. 固定加锁顺序(如按ID升序)
//   2. 大事务拆小
//   3. 降低隔离级别
//   4. 添加合理索引(避免锁升级)
//   5. 设置 innodb_lock_wait_timeout
//
// **面试追问**: 如何排查死锁?
//   SHOW ENGINE INNODB STATUS → LATEST DETECTED DEADLOCK
func deadlock() {
	fmt.Println("=== 4. 死锁 ===")
	fmt.Println("经典死锁场景:")
	fmt.Println("  事务A: UPDATE t SET val=1 WHERE id=1 → 获得id=1行锁")
	fmt.Println("  事务B: UPDATE t SET val=2 WHERE id=2 → 获得id=2行锁")
	fmt.Println("  事务A: UPDATE t SET val=3 WHERE id=2 → 等待id=2行锁")
	fmt.Println("  事务B: UPDATE t SET val=4 WHERE id=1 → 等待id=1行锁 → 死锁!")
	fmt.Println()
	fmt.Println("排查: SHOW ENGINE INNODB STATUS")
	fmt.Println("避免: 固定加锁顺序, 大事务拆小, 合理索引, 超时回滚")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 乐观锁与悲观锁
// ----------------------------------------------------------
// 悲观锁: 假设冲突一定发生, 先加锁再操作
//   实现: SELECT ... FOR UPDATE (排他锁)
//   适用: 写多读少, 冲突频繁
//
// 乐观锁: 假设冲突很少发生, 提交时检查
//   实现: version 字段 + CAS
//   适用: 读多写少, 冲突较少
//
// **Go 中的映射**: sync.Mutex = 悲观, atomic.CAS = 乐观
func optimisticVsPessimistic() {
	fmt.Println("=== 5. 乐观锁与悲观锁 ===")
	fmt.Println()
	fmt.Println("悲观锁: SELECT * FROM t WHERE id=1 FOR UPDATE")
	fmt.Println("  → 获得行锁, 其他事务阻塞")
	fmt.Println("  适用: 写多, 冲突频繁")
	fmt.Println()
	fmt.Println("乐观锁: UPDATE t SET val=new, version=version+1")
	fmt.Println("        WHERE id=1 AND version=old_version")
	fmt.Println("  → 影响行数=0则重试, =1则成功")
	fmt.Println("  适用: 读多写少, 冲突少")
	fmt.Println()
	fmt.Println("Go 映射: Mutex=悲观锁, atomic.CAS=乐观锁")
	fmt.Println()
}

// ----------------------------------------------------------
// 6. 生产事务设计
// ----------------------------------------------------------
// 大事务的危害:
//   1. 锁持有时间长 → 并发度低
//   2. Undo Log膨胀 → 占用大量空间
//   3. 主从延迟 → binlog 在事务提交后才发送
//   4. 长事务回滚代价大
//
// 最佳实践:
//   1. 事务尽量短(不要在事务中做RPC/耗时计算)
//   2. 根据业务选择合适的隔离级别
//   3. Go 中使用 context 控制事务超时
//   4. 读写分离: 写走主库, 读走从库
//   5. 分库分表后使用分布式事务(TCC/Saga/消息最终一致性)
func productionTx() {
	fmt.Println("=== 6. 生产事务设计 ===")
	fmt.Println()
	fmt.Println("大事务危害: 锁久/Undo膨胀/主从延迟/回滚代价大")
	fmt.Println()
	fmt.Println("最佳实践:")
	fmt.Println("  1. 事务尽量短(不含RPC/耗时操作)")
	fmt.Println("  2. Go: ctx+超时控制")
	fmt.Println("  3. 读写分离(写主读从)")
	fmt.Println("  4. 分库后用 TCC/Saga/最终一致性")
	fmt.Println()
	fmt.Println("Go 事务模板:")
	fmt.Println("  tx, _ := db.BeginTx(ctx, nil)")
	fmt.Println("  defer tx.Rollback() // 无副作用回滚")
	fmt.Println("  // ... 操作 ...")
	fmt.Println("  if err := tx.Commit(); err != nil { return err }")
}
