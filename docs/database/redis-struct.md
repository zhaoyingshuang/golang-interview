---
title: Redis 高级数据结构与分布式锁
---

## 1. 高级数据结构

### HyperLogLog（基数统计）

用于统计不重复元素的数量，标准误差约 0.81%，每个 Key 只占用 12KB 内存：

```sql
-- 添加元素
PFADD uv:2024-01-15 user1 user2 user3

-- 获取基数估算值
PFCOUNT uv:2024-01-15

-- 合并多个 HyperLogLog
PFMERGE uv:total uv:2024-01-14 uv:2024-01-15
```

::: info
HyperLogLog 适用于 UV 统计、日活统计等不需要精确计数的场景。12KB 固定内存可估算约 2^64 个不同元素。
:::

### Bitmap（位图）

String 类型的位操作，适合存储布尔型数据，节省内存：

```sql
-- 用户签到（以日期为 Key，用户 ID 为偏移量）
SETBIT sign:2024-01-15 1001 1    -- 用户 1001 签到
GETBIT sign:2024-01-15 1001       -- 检查是否签到

-- 统计签到人数
BITCOUNT sign:2024-01-15

-- 连续签到天数（AND 运算）
BITOP AND sign:result sign:2024-01-14 sign:2024-01-15
BITCOUNT sign:result  -- 两天都签到的用户数
```

### Geo（地理位置）

底层使用 ZSet（Sorted Set）实现，使用 GeoHash 编码经纬度：

```sql
-- 添加地理位置
GEOADD stores:beijing 116.397 39.908 "天安门"
GEOADD stores:beijing 116.404 39.915 "王府井"

-- 计算两点距离
GEODIST stores:beijing "天安门" "王府井" km

-- 查找附近 5km 内的地点
GEORADIUS stores:beijing 116.397 39.908 5 km WITHDIST WITHCOORD
```

## 2. 分布式锁

### 基础实现（SET NX EX）

```go
// 加锁
func Lock(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
    return rdb.SetNX(ctx, key, value, ttl).Result()
}

// 释放锁（Lua 脚本保证原子性）
var unlockScript = redis.NewScript(`
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end
`)

func Unlock(ctx context.Context, key, value string) error {
    return unlockScript.Run(ctx, rdb, []string{key}, value).Err()
}
```

::: warning
释放锁必须使用 Lua 脚本保证"判断值 + 删除"的原子性，否则可能出现误删其他客户端的锁。
:::

### Redisson 看门狗机制

Redisson 的看门狗（Watchdog）自动续期机制：

```go
// Redisson 风格的锁续期
func StartWatchdog(ctx context.Context, key, value string, ttl time.Duration) {
    ticker := time.NewTicker(ttl / 3) // 每过 1/3 过期时间续期一次
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            // 续期
            rdb.Expire(ctx, key, ttl)
        case <-ctx.Done():
            return
        }
    }
}
```

### RedLock 算法与争议

RedLock 使用多个独立的 Redis 实例，在大多数节点上加锁成功才算获取锁成功。

::: warning
Martin Kleppmann 对 RedLock 的批评：(1) 系统时钟跳变可能导致锁安全问题；(2) 进程暂停（GC）可能导致锁过期后仍认为持有锁。推荐使用 fencing token 方案或直接使用 ZooKeeper/etcd。
:::

## 3. Redis 集群

### 主从复制

```sql
-- 在从节点上配置主节点
REPLICAOF 192.168.1.100 6379

-- 查看复制状态
INFO replication
```

主从复制是异步的，从节点提供读服务，写操作只在主节点执行。

### 哨兵模式（Sentinel）

```bash
# sentinel.conf
sentinel monitor mymaster 192.168.1.100 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 10000
sentinel parallel-syncs mymaster 1
```

哨兵负责监控主节点健康状态，自动进行故障转移（Failover）。

### Cluster 分片

Redis Cluster 使用 16384 个哈希槽（Hash Slot）进行数据分片：

```sql
-- 创建集群
redis-cli --cluster create \
  192.168.1.101:6379 192.168.1.102:6379 192.168.1.103:6379 \
  192.168.1.104:6379 192.168.1.105:6379 192.168.1.106:6379 \
  --cluster-replicas 1

-- 查看集群信息
CLUSTER INFO
CLUSTER NODES

-- 查看键所在槽
CLUSTER KEYSLOT user:1001
```

::: tip
Cluster 模式下，Key 的大批量操作建议使用 Hash Tag（如 `{user}.1001`）将相关 Key 路由到同一节点。
:::

## 4. Pipeline 与 Lua 脚本

### Pipeline 批量操作

Pipeline 将多个命令打包一次性发送，减少网络往返（RTT）：

```go
// 使用 Pipeline 批量写入
pipe := rdb.Pipeline()
for i := 0; i < 1000; i++ {
    key := fmt.Sprintf("key:%d", i)
    pipe.Set(ctx, key, i, 0)
}
_, err := pipe.Exec(ctx)
```

### Lua 脚本实现原子操作

```go
// 限流脚本：滑动窗口
var limitScript = redis.NewScript(`
    local key = KEYS[1]
    local limit = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])
    local current = tonumber(ARGV[3])

    redis.call("ZREMRANGEBYSCORE", key, 0, current - window)
    local count = redis.call("ZCARD", key)
    if count < limit then
        redis.call("ZADD", key, current, current)
        redis.call("PEXPIRE", key, window)
        return 1
    end
    return 0
`)

func RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now().UnixMilli()
    result, err := limitScript.Run(ctx, rdb, []string{key}, limit, window.Milliseconds(), now).Int()
    return result == 1, err
}
```

::: info
Lua 脚本在 Redis 中原子执行，不会被其他命令打断。适合实现限流、分布式锁、库存扣减等需要原子性的操作。
:::

## 5. 面试常见问题

### 大 Key 问题

```bash
# 查找大 Key
redis-cli --bigkeys

# 查看指定 Key 的内存占用
MEMORY USAGE user:1001
```

大 Key 的危害：阻塞 Redis（DEL 大 Key 耗时长）、网络传输慢、集群迁移困难。

**处理方案**：拆分大 Key（如 Hash 拆分为多个小 Hash）、使用 `UNLINK` 异步删除。

### 热 Key 处理

```go
// 方案一：本地缓存
var localCache = gcache.New(1000).Expiration(time.Minute).Build()

func GetHotKey(ctx context.Context, key string) (string, error) {
    val, err := localCache.Get(key)
    if err == nil {
        return val.(string), nil
    }
    return rdb.Get(ctx, key).Result()
}
```

### 数据一致性

```go
// 延迟双删策略
func UpdateWithCache(ctx context.Context, id int, data string) error {
    key := fmt.Sprintf("data:%d", id)
    // 1. 删除缓存
    rdb.Del(ctx, key)
    // 2. 更新数据库
    err := db.Update(ctx, id, data)
    if err != nil {
        return err
    }
    // 3. 延迟再次删除缓存
    time.AfterFunc(500*time.Millisecond, func() {
        rdb.Del(context.Background(), key)
    })
    return nil
}
```

::: warning 面试高频问题
1. **Redis Cluster 为什么用 16384 个槽而不是 65536 个？**
   - 16384 个槽足够使用，节点间心跳包携带的位图更小（2KB vs 8KB），节省网络带宽。

2. **大 Key 如何发现和处理？**
   - 使用 `redis-cli --bigkeys` 或 `MEMORY USAGE` 发现；通过拆分、异步删除（UNLINK）处理。

3. **RedLock 有什么争议？你推荐使用什么方案？**
   - 时钟跳变和 GC 暂停可能导致锁安全问题。对强一致性要求高的场景推荐 etcd 或 ZooKeeper。

4. **Lua 脚本有什么注意事项？**
   - 脚本必须简短，长时间运行的脚本会阻塞 Redis；注意脚本超时配置（lua-time-limit）；Cluster 模式下所有 Key 必须在同一节点。

5. **如何保证缓存与数据库的一致性？**
   - 常用方案：延迟双删、基于 binlog 的异步更新（Canal）、设置合理的过期时间作为兜底。
:::
