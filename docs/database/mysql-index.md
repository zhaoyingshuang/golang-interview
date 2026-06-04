---
title: MySQL 索引原理
---

## 1. B+ 树结构原理

MySQL InnoDB 引擎使用 B+ 树作为索引的默认数据结构。B+ 树是 B 树的变体，具有以下关键特性：

- **非叶子节点只存储键值**，不存储数据，使得每个节点能容纳更多键值，降低树的高度
- **叶子节点通过指针相连**，支持高效的范围查询和排序操作
- **所有数据都存储在叶子节点**，查询性能稳定，每次查找都走根到叶子的路径

```sql
-- 查看表的索引信息
SHOW INDEX FROM users;

-- 查看 InnoDB 页结构信息
SELECT * FROM information_schema.INNODB_TABLES
WHERE NAME = 'test/users';
```

### 为什么不用其他数据结构？

| 数据结构 | 劣势 |
|---------|------|
| B 树 | 非叶子节点存储数据，节点扇出小，树更高，IO 次数多 |
| 红黑树 | 二叉树高度过大，IO 次数远高于 B+ 树 |
| Hash | 不支持范围查询和排序，无法利用索引排序 |
| 跳表 | 层级过深，磁盘 IO 效率不如 B+ 树 |

::: tip
B+ 树的一个节点通常对应一个磁盘页（默认 16KB），三层的 B+ 树大约可以存储 2000 万行记录。
:::

## 2. 聚簇索引与二级索引

### 聚簇索引

InnoDB 的聚簇索引将数据行与主键索引存储在同一棵 B+ 树中。表中的数据按照主键顺序物理存储，因此一张表只能有一个聚簇索引。

### 二级索引与回表查询

二级索引的叶子节点存储的是主键值，而非数据行。通过二级索引查找数据时，需要先在二级索引中找到主键值，再回到聚簇索引中查找完整数据，这个过程称为**回表**。

```sql
-- 创建二级索引
CREATE INDEX idx_user_email ON users(email);

-- 该查询会触发回表
SELECT * FROM users WHERE email = 'test@example.com';
```

### 覆盖索引优化

当查询的列全部包含在索引中时，无需回表，这就是覆盖索引。

```sql
-- 创建联合索引实现覆盖索引
CREATE INDEX idx_user_cover ON users(email, name, age);

-- 该查询不需要回表（Extra: Using index）
SELECT email, name, age FROM users WHERE email = 'test@example.com';
```

::: info
通过 `EXPLAIN` 看到 `Extra` 字段出现 `Using index` 时，说明使用了覆盖索引，避免了回表操作。
:::

## 3. 索引优化实战

### 最左前缀原则

联合索引 `(a, b, c)` 相当于创建了 `(a)`, `(a, b)`, `(a, b, c)` 三个索引。查询条件必须从最左列开始匹配。

```sql
-- 联合索引
CREATE INDEX idx_abc ON orders(user_id, status, created_at);

-- 能命中索引
SELECT * FROM orders WHERE user_id = 1 AND status = 'paid';
SELECT * FROM orders WHERE user_id = 1;

-- 无法命中索引（缺少最左列 user_id）
SELECT * FROM orders WHERE status = 'paid' AND created_at > '2024-01-01';
```

### 索引下推（Index Condition Pushdown, ICP）

MySQL 5.6 引入的优化。在联合索引中，即使某些列无法用于索引查找，也可以在索引层面进行条件过滤，减少回表次数。

```sql
-- 联合索引 (name, age)
-- MySQL 5.6 之前：通过 name 找到所有行 -> 回表 -> 再过滤 age
-- ICP 之后：通过 name 找到索引项 -> 在索引中过滤 age -> 回表
SELECT * FROM users WHERE name LIKE '张%' AND age = 25;
```

::: tip
ICP 在 `EXPLAIN` 的 `Extra` 字段中显示为 `Using index condition`。
:::

## 4. Explain 执行计划分析

```sql
EXPLAIN SELECT * FROM users WHERE email = 'test@example.com';
```

### type 字段（从优到差）

| type | 含义 |
|------|------|
| system | 表中只有一行 |
| const | 通过主键/唯一索引匹配一行 |
| eq_ref | 唯一索引扫描，每次关联匹配一行 |
| ref | 非唯一索引扫描 |
| range | 索引范围扫描 |
| index | 全索引扫描 |
| ALL | 全表扫描 |

### Extra 字段关键值

| Extra 值 | 含义 |
|----------|------|
| Using index | 覆盖索引，无需回表 |
| Using index condition | 索引下推 |
| Using where | Server 层过滤 |
| Using filesort | 额外排序，需优化 |
| Using temporary | 使用临时表，需优化 |

::: warning
当 `type` 为 `ALL` 或 `Extra` 出现 `Using filesort` / `Using temporary` 时，需要重点优化查询或添加索引。
:::

## 5. 常见索引失效场景

### 函数操作

```sql
-- 索引失效：在索引列上使用函数
SELECT * FROM users WHERE YEAR(created_at) = 2024;

-- 优化：改写为范围查询
SELECT * FROM users WHERE created_at >= '2024-01-01' AND created_at < '2025-01-01';
```

### 隐式类型转换

```sql
-- 索引失效：varchar 列用数字查询（隐式转换）
SELECT * FROM users WHERE phone = 13800138000;

-- 优化：使用字符串类型
SELECT * FROM users WHERE phone = '13800138000';
```

### OR 条件

```sql
-- 索引可能失效：OR 连接不同列
SELECT * FROM users WHERE name = '张三' OR email = 'test@example.com';

-- 优化：使用 UNION ALL
SELECT * FROM users WHERE name = '张三'
UNION ALL
SELECT * FROM users WHERE email = 'test@example.com';
```

### LIKE 前缀通配符

```sql
-- 索引失效：前缀通配符
SELECT * FROM users WHERE name LIKE '%三';

-- 可命中索引：后缀通配符
SELECT * FROM users WHERE name LIKE '张%';
```

## 面试追问

::: warning 面试高频问题
1. **为什么 MySQL 使用 B+ 树而不是 B 树？**
   - B+ 树非叶子节点不存数据，单页存更多键值，树更矮，IO 更少；叶子节点链表相连，范围查询效率高。

2. **什么是最左前缀原则？联合索引 (a,b,c) 查询条件为 a=1 AND c=3 能否命中索引？**
   - 只能命中 a 列的索引部分，c 无法利用索引（跳过了 b）。

3. **什么是索引下推（ICP）？它解决了什么问题？**
   - 在索引遍历过程中对索引中包含的字段执行条件判断，直接在存储引擎层过滤掉不满足条件的记录，减少回表次数。

4. **EXPLAIN 中 type=index 和 type=ALL 有什么区别？**
   - index 是全索引扫描（遍历索引树），ALL 是全表扫描（遍历数据文件）。index 通常比 ALL 好，但两者都是需要优化的信号。

5. **覆盖索引和索引下推有什么区别？**
   - 覆盖索引是完全不需要回表，所有需要的数据都在索引中；索引下推是在索引层面预过滤，减少回表次数但仍然需要回表。
:::
