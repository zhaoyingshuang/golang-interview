package main

import "fmt"

// ============================================================
// Redis 缓存策略
// ============================================================
// 【面试高频】缓存击穿/穿透/雪崩, 分布式锁, 缓存一致性

func main() {
	cachePatterns()
	cacheFailures()
	distributedLock()
	redisCluster()
	productionCache()
}

func cachePatterns() {
	fmt.Println("=== 1. 缓存模式 ===")
	fmt.Println()
	fmt.Println("Cache Aside (最常用, 推荐):")
	fmt.Println("  读: 先查缓存→命中返回→未命中查DB→写回缓存")
	fmt.Println("  写: 先更新DB→再删缓存(不是更新缓存!)")
	fmt.Println()
	fmt.Println("为什么删缓存而不是更新?")
	fmt.Println("  1. 复杂缓存(多表join)更新代价高")
	fmt.Println("  2. 并发写可能导致缓存与DB不一致")
	fmt.Println("  3. 懒加载: 只缓存热点数据")
	fmt.Println()
	fmt.Println("Write Through: 缓存负责同步写DB(缓存更新+DB更新原子操作)")
	fmt.Println("Write Behind: 先写缓存, 异步批量写DB(可能丢数据)")
	fmt.Println()
}

func cacheFailures() {
	fmt.Println("=== 2. 缓存击穿/穿透/雪崩 ===")
	fmt.Println()
	fmt.Println("缓存穿透: 查不存在的key, 请求直达DB")
	fmt.Println("  解决: 1)布隆过滤器 2)缓存空值(短TTL) 3)接口校验")
	fmt.Println()
	fmt.Println("缓存击穿: 热点key过期, 大量请求瞬间压到DB")
	fmt.Println("  解决: 1)互斥锁(SETNX) 2)逻辑过期(不设TTL)")
	fmt.Println("  互斥锁实现: SET key value NX EX timeout")
	fmt.Println()
	fmt.Println("缓存雪崩: 大量key同时过期/Redis宕机")
	fmt.Println("  解决: 1)随机过期时间 2)多级缓存 3)熔断降级 4)Redis集群")
	fmt.Println()
	fmt.Println("Go 互斥锁防止击穿示例:")
	fmt.Println("  if redis.SetNX(lockKey, 1, 10s):")
	fmt.Println("    data = db.Query()")
	fmt.Println("    redis.Set(key, data, ttl)")
	fmt.Println("    redis.Del(lockKey)")
	fmt.Println()
}

func distributedLock() {
	fmt.Println("=== 3. 分布式锁 ===")
	fmt.Println()
	fmt.Println("基础实现:")
	fmt.Println("  SET lock_key unique_value NX EX 30")
	fmt.Println("  释放: Lua脚本原子性检查+删除")
	fmt.Println()
	fmt.Println("Redisson 原理:")
	fmt.Println("  1. 自动续期(看门狗): 每10s续期到30s")
	fmt.Println("  2. 可重入: Hash结构 field=thread_id, value=重入次数")
	fmt.Println("  3. 等待订阅: 发布订阅等待锁释放通知")
	fmt.Println()
	fmt.Println("RedLock (多节点):")
	fmt.Println("  向N个独立Redis实例加锁, 超过半数成功=获取锁")
	fmt.Println("  争议: 时钟漂移/GC暂停可能导致不安全")
	fmt.Println()
	fmt.Println("Go 实现: github.com/go-redsync/redsync")
	fmt.Println("  适合: 简单场景, 非强一致性要求")
	fmt.Println()
}

func redisCluster() {
	fmt.Println("=== 4. Redis 集群 ===")
	fmt.Println()
	fmt.Println("主从复制: 全量同步(RDB)+增量同步(Repl Backlog)")
	fmt.Println("  从库只读, 适合读扩展")
	fmt.Println()
	fmt.Println("哨兵(Sentinel):")
	fmt.Println("  监控+故障转移, 选举新主库(优先级+偏移量)")
	fmt.Println("  适合: 自动故障转移, 读写分离")
	fmt.Println()
	fmt.Println("Cluster (16384个哈希槽):")
	fmt.Println("  slot = CRC16(key) % 16384")
	fmt.Println("  每个节点负责一部分槽")
	fmt.Println("  gossip协议通信, 故障自动转移")
	fmt.Println("  适合: 大规模数据分片")
	fmt.Println()
	fmt.Println("选型: 数据量<单机内存 → 主从+哨兵; 更大 → Cluster")
	fmt.Println()
}

func productionCache() {
	fmt.Println("=== 5. 生产缓存设计 ===")
	fmt.Println()
	fmt.Println("缓存一致性方案:")
	fmt.Println("  1. 延迟双删: 删缓存→更新DB→延迟删缓存")
	fmt.Println("  2. 监听binlog(Canal): DB变更→删缓存(最终一致性)")
	fmt.Println("  3. 读时修复: 读缓存miss→查DB→写缓存+设短TTL")
	fmt.Println()
	fmt.Println("热key处理:")
	fmt.Println("  1. 本地缓存+Redis二级缓存")
	fmt.Println("  2. 热key分散: key加后缀, 分到多个key")
	fmt.Println("  3. 提前加载: 预测+预热")
	fmt.Println()
	fmt.Println("大key拆分:")
	fmt.Println("  hash: 拆分为多个小hash(按分片规则)")
	fmt.Println("  list: 拆分为多个list(按时间/范围)")
	fmt.Println("  检测: redis-cli --bigkeys / memory usage key")
	fmt.Println()
}
