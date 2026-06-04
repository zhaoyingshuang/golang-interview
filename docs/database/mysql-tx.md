---
title: MySQL 事务与锁
---

## 1. ACID 特性与实现

事务的四大特性（ACID）在 InnoDB 中通过不同机制实现：

| 特性 | 实现机制 |
|------|---------|
| 原子性（Atomicity） | undo log（回滚日志） |
| 一致性（Consistency） | 数据库约束 + 原子性 + 隔离性 |
| 隔离性（Isolation） | 锁机制 + MVCC |
| 持久性（Durability） | redo log（重做日志） |

### 三大日志协作

```sql
-- 查看日志相关配置
SHOW VARIABLES LIKE 'innodb_log_file_size';
SHOW VARIABLES LIKE 'innodb_log_buffer_size';
SHOW VARIABLES LIKE 'sync_binlog';
```

**redo log**：物理日志，记录数据页的物理修改。WAL（Write-Ahead Logging）策略，先写日志再写磁盘，保证崩溃恢复能力。

**undo log**：逻辑日志，记录数据修改前的值，用于事务回滚和 MVCC 的快照读。

**binlog**：逻辑日志，记录所有 DDL 和 DML 操作，用于主从复制和数据恢复。

::: tip
两阶段提交：事务提交时先写 redo log（prepare），再写 binlog，最后写 redo log（commit），保证 redo log 和 binlog 的一致性。
:::

## 2. 隔离级别与 MVCC

### 四大隔离级别

| 隔离级别 | 脏读 | 不可重复读 | 幻读 |
|---------|------|-----------|------|
| Read Uncommitted | 有 | 有 | 有 |
| Read Committed (RC) | 无 | 有 | 有 |
| Repeatable Read (RR) | 无 | 无 | 部分避免 |
| Serializable | 无 | 无 | 无 |

```sql
-- 查看当前隔离级别
SELECT @@transaction_isolation;

-- 设置隔离级别
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
```

### MVCC 原理

MVCC（多版本并发控制）通过 **隐藏列 + undo log 版本链 + ReadView** 实现：

- **隐藏列**：每行数据包含 `DB_TRX_ID`（最后修改的事务 ID）和 `DB_ROLL_PTR`（回滚指针）
- **版本链**：通过 `DB_ROLL_PTR` 将数据的多个版本串联
- **ReadView**：事务开始时创建的"快照"，决定能看到哪个版本

::: info
RC 级别每次 SELECT 都创建新的 ReadView，RR 级别只在第一次 SELECT 时创建 ReadView。
:::

## 3. 锁机制

### 锁类型

| 锁类型 | 描述 | 加锁对象 |
|-------|------|---------|
| 行锁（Record Lock） | 锁定单行记录 | 索引记录 |
| 间隙锁（Gap Lock） | 锁定索引记录间的间隙 | 索引间隙 |
| 临键锁（Next-Key Lock） | 行锁 + 间隙锁 | 索引记录及其间隙 |
| 表锁（Table Lock） | 锁定整张表 | 整个表 |

```sql
-- 查看当前锁信息
SELECT * FROM performance_schema.data_locks;

-- 查看锁等待
SELECT * FROM performance_schema.data_lock_waits;
```

### 加锁规则（RR 级别）

```sql
-- 示例表
CREATE TABLE t (id INT PRIMARY KEY, c INT, KEY(c));
INSERT INTO t VALUES (5,5), (10,10), (15,15), (20,20);

-- 加行锁（唯一索引等值查询命中）
SELECT * FROM t WHERE id = 10 FOR UPDATE;
-- 锁：id=10 的行锁

-- 加临键锁（普通索引等值查询未完全命中）
SELECT * FROM t WHERE c = 10 FOR UPDATE;
-- 锁：c 在 (5,10] 的临键锁 + (10,15) 的间隙锁

-- 加间隙锁（等值查询未命中）
SELECT * FROM t WHERE c = 12 FOR UPDATE;
-- 锁：c 在 (10,15) 的间隙锁
```

::: warning
InnoDB 的行锁是加在索引上的，如果没有用到索引，行锁会退化为表锁。
:::

## 4. 死锁检测与处理

### 死锁场景

```sql
-- 事务 A
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1; -- 锁住 id=1

-- 事务 B
BEGIN;
UPDATE accounts SET balance = balance + 50 WHERE id = 2;  -- 锁住 id=2

-- 事务 A（等待 id=2 的锁）
UPDATE accounts SET balance = balance + 100 WHERE id = 2;

-- 事务 B（等待 id=1 的锁，死锁产生）
UPDATE accounts SET balance = balance - 50 WHERE id = 1;
```

### 死锁检测与配置

```sql
-- 查看死锁检测开关（默认开启）
SHOW VARIABLES LIKE 'innodb_deadlock_detect';

-- 查看最近一次死锁信息
SHOW ENGINE INNODB STATUS;

-- 设置锁等待超时（默认 50 秒）
SHOW VARIABLES LIKE 'innodb_lock_wait_timeout';
```

::: warning
当死锁检测开启时，InnoDB 会自动检测死锁并回滚代价最小的事务。在高并发场景下，死锁检测本身可能带来性能开销，可以考虑关闭并通过锁超时机制处理。
:::

## 5. 面试常见问题

### 快照读 vs 当前读

```sql
-- 快照读（普通 SELECT，使用 MVCC）
SELECT * FROM users WHERE id = 1;

-- 当前读（加锁读，读取最新已提交数据）
SELECT * FROM users WHERE id = 1 FOR UPDATE;     -- 排他锁
SELECT * FROM users WHERE id = 1 LOCK IN SHARE MODE; -- 共享锁
UPDATE users SET name = 'test' WHERE id = 1;     -- 排他锁
DELETE FROM users WHERE id = 1;                   -- 排他锁
```

### RR 级别下的幻读问题

RR 级别通过 MVCC 解决了快照读的幻读，但当前读仍可能出现幻读：

```sql
-- 事务 A
BEGIN;
SELECT * FROM t WHERE c BETWEEN 10 AND 20; -- 快照读，返回 2 行

-- 事务 B
INSERT INTO t VALUES (12, 12); -- 插入成功
COMMIT;

-- 事务 A
UPDATE t SET c = 100 WHERE c = 12; -- 当前读，能更新成功（因为看到了新行）
SELECT * FROM t WHERE c BETWEEN 10 AND 20; -- 再次快照读，看到 3 行（幻读）
```

::: warning 面试高频问题
1. **redo log 和 binlog 的区别是什么？**
   - redo log 是 InnoDB 引擎层的物理日志，循环写入，用于崩溃恢复；binlog 是 Server 层的逻辑日志，追加写入，用于主从复制和数据恢复。

2. **MVCC 在 RC 和 RR 级别下的区别？**
   - RC 每次快照读都创建新 ReadView，能看到其他事务已提交的最新数据；RR 只在首次快照读时创建 ReadView，后续读都使用同一个快照。

3. **InnoDB 的行锁是加在数据行上还是索引上？**
   - 加在索引记录上。如果没有使用索引，行锁会退化为表锁。

4. **如何避免死锁？**
   - 按固定顺序访问表和行；保持事务简短；合理使用索引避免锁升级；设置合适的锁等待超时。

5. **RR 级别一定能避免幻读吗？**
   - 不一定。MVCC 只解决快照读的幻读问题，当前读（FOR UPDATE/LOCK IN SHARE MODE）需要依赖临键锁来避免幻读。如果快照读和当前读混用，仍可能出现幻读。
:::
