---
title: Redis 缓存策略
---

## 1. Redis 数据结构

Redis 提供了五种基础数据结构，每种都有多种底层编码实现，会根据数据量和特征自动选择最优编码。

### 基础数据结构与底层编码

| 数据结构 | 底层编码 | 适用场景 |
|---------|---------|---------|
| String | int / embstr / raw | 计数器、缓存、分布式锁 |
| Hash | ziplist / hashtable | 对象存储（用户信息等） |
| List | quicklist | 消息队列、最新列表 |
| Set | intset / hashtable | 标签、共同好友 |
| ZSet | ziplist / skiplist + hashtable | 排行榜、延迟队列 |

```sql
-- 查看键的底层编码
OBJECT ENCODING user:1001
```

### Stream 数据结构

Redis 5.0 引入的 Stream 是一个持久化的消息队列，支持消费者组：

```sql
-- 添加消息
XADD orders * user_id 1001 amount 99.9

-- 创建消费者组
XGROUP CREATE orders order_group 0

-- 消费者读取消息
XREADGROUP GROUP order_group consumer1 COUNT 1 BLOCK 5000 STREAMS orders >
```

::: tip
Stream 相比 List 的优势：支持消费者组、消息确认（ACK）、持久化、支持按时间范围查询历史消息。
:::

## 2. 缓存模式

### Cache Aside（旁路缓存）

最常用的缓存模式。应用程序同时与缓存和数据库交互：

```go
// 读取数据
func GetUser(ctx context.Context, id int) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    // 1. 先查缓存
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(val), &user)
        return &user, nil
    }
    // 2. 缓存未命中，查数据库
    user, err := db.GetUserByID(ctx, id)
    if err != nil {
        return nil, err
    }
    // 3. 写入缓存
    data, _ := json.Marshal(user)
    rdb.Set(ctx, key, data, 30*time.Minute)
    return user, nil
}
```

### Read Through / Write Through

缓存层负责与数据库的读写同步，应用层只与缓存交互：

```go
// Write Through：写入缓存时同步更新数据库
func UpdateUser(ctx context.Context, user *User) error {
    key := fmt.Sprintf("user:%d", user.ID)
    data, _ := json.Marshal(user)
    // 缓存层内部负责同步写数据库
    return rdb.Set(ctx, key, data, 0).Err()
}
```

### Write Behind（异步回写）

先将数据写入缓存，异步批量写入数据库，提升写入性能但存在数据丢失风险。

::: warning
Write Behind 模式下，如果缓存节点宕机，尚未落盘的数据会丢失。适用于对数据一致性要求不高但写入性能要求高的场景。
:::

## 3. 缓存问题

### 缓存穿透

查询一个数据库和缓存中都不存在的数据，每次请求都会打到数据库。

```go
// 方案一：缓存空值
func GetUser(ctx context.Context, id int) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        if val == "NULL" {
            return nil, ErrNotFound // 命中空值缓存
        }
        // ...反序列化返回
    }
    user, err := db.GetUserByID(ctx, id)
    if err != nil {
        rdb.Set(ctx, key, "NULL", 5*time.Minute) // 缓存空值，短过期
        return nil, err
    }
    return user, nil
}
```

```go
// 方案二：布隆过滤器（在查询缓存前先判断）
func InitBloomFilter(ctx context.Context) {
    // 将所有合法 ID 加入布隆过滤器
    ids := db.GetAllUserIDs(ctx)
    for _, id := range ids {
        bf.Add(ctx, fmt.Sprintf("user:%d", id))
    }
}
```

### 缓存击穿

某个热点 Key 过期的瞬间，大量请求同时打到数据库。

```go
// 方案：使用 singleflight 合并并发请求
var sg singleflight.Group

func GetUser(ctx context.Context, id int) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    val, err, _ := sg.Do(key, func() (interface{}, error) {
        // 只有第一个请求执行数据库查询
        cached, err := rdb.Get(ctx, key).Result()
        if err == nil {
            return cached, nil
        }
        user, err := db.GetUserByID(ctx, id)
        if err != nil {
            return nil, err
        }
        data, _ := json.Marshal(user)
        rdb.Set(ctx, key, data, 30*time.Minute)
        return string(data), nil
    })
    // ...处理结果
    return parseUser(val)
}
```

### 缓存雪崩

大量缓存 Key 在同一时刻过期，或 Redis 节点宕机。

```go
// 方案一：过期时间加随机偏移
func SetWithJitter(ctx context.Context, key string, val interface{}, baseTTL time.Duration) {
    jitter := time.Duration(rand.Intn(300)) * time.Second
    rdb.Set(ctx, key, val, baseTTL+jitter)
}
```

::: info
缓存雪崩的解决方案：(1) 过期时间加随机偏移；(2) 使用 Redis 集群保证高可用；(3) 多级缓存（本地缓存 + Redis）；(4) 限流降级。
:::

## 4. 过期与淘汰策略

### TTL 设置建议

```sql
-- 设置带过期时间的 Key
SET session:token123 "user_data" EX 1800

-- 查看剩余过期时间（秒）
TTL session:token123

-- 查看剩余过期时间（毫秒）
PTTL session:token123
```

### 淘汰策略

| 策略 | 说明 |
|------|------|
| noeviction | 不淘汰，内存满时拒绝写入（默认） |
| allkeys-lru | 所有键中淘汰最久未使用的 |
| allkeys-lfu | 所有键中淘汰使用频率最低的 |
| allkeys-random | 所有键中随机淘汰 |
| volatile-lru | 设置了过期时间的键中淘汰最久未使用的 |
| volatile-lfu | 设置了过期时间的键中淘汰频率最低的 |
| volatile-random | 设置了过期时间的键中随机淘汰 |
| volatile-ttl | 淘汰 TTL 最短的键 |

```sql
-- 配置淘汰策略
CONFIG SET maxmemory-policy allkeys-lru

-- 查看当前内存使用情况
INFO memory
```

::: tip
推荐使用 `allkeys-lru`（通用缓存）或 `volatile-lru`（Redis 同时用于缓存和持久化存储时）。
:::

## 5. Redis 持久化

### RDB（快照）

在指定时间间隔内将内存数据集快照写入磁盘：

```sql
-- 手动触发 RDB
SAVE        -- 同步阻塞
BGSAVE      -- 后台异步

-- 配置自动快照策略
CONFIG SET save "900 1 300 10 60 10000"
-- 表示：900秒内有1次修改 / 300秒内有10次修改 / 60秒内有10000次修改
```

### AOF（追加日志）

将每个写命令追加到文件末尾：

```sql
-- 开启 AOF
CONFIG SET appendonly yes

-- AOF 同步策略
CONFIG SET appendfsync everysec  -- 推荐：每秒同步
-- always：每条命令同步（最安全，性能最差）
-- no：由操作系统决定同步时机（性能最好，安全性最低）
```

### 混合持久化（RDB + AOF）

Redis 4.0 引入，AOF 重写时将前半部分用 RDB 格式写入，后半部分用 AOF 格式：

```sql
-- 开启混合持久化
CONFIG SET aof-use-rdb-preamble yes
```

::: warning 面试高频问题
1. **RDB 和 AOF 各自的优缺点？**
   - RDB：恢复速度快、文件紧凑，但可能丢失最后一次快照后的数据。AOF：数据安全性高（最多丢失 1 秒），但文件体积大、恢复速度慢。

2. **缓存穿透、击穿、雪崩的区别和解决方案？**
   - 穿透：查不存在的数据 -> 布隆过滤器/缓存空值。击穿：热点 Key 过期 -> 互斥锁/singleflight。雪崩：大量 Key 同时过期 -> 随机过期时间/集群/限流。

3. **Redis 的过期删除策略是什么？**
   - 惰性删除：访问 Key 时检查是否过期。定期删除：每 100ms 随机检查一批设置过期时间的 Key，删除已过期的。

4. **如何选择淘汰策略？**
   - 纯缓存场景用 `allkeys-lru`；Redis 有持久化数据时用 `volatile-lru`；热点数据多用 `allkeys-lfu`。
:::
