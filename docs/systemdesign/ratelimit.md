---
title: 限流设计
---

## 1. 限流算法详解

### 固定窗口计数器

将时间划分为固定窗口，在每个窗口内统计请求计数。实现简单但存在临界点突刺问题：窗口边界处可能出现 2 倍阈值的流量。

```go
package ratelimit

import (
    "sync"
    "time"
)

type FixedWindowLimiter struct {
    mu       sync.Mutex
    limit    int
    window   time.Duration
    count    int
    lastTime time.Time
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindowLimiter {
    return &FixedWindowLimiter{
        limit:    limit,
        window:   window,
        lastTime: time.Now(),
    }
}

func (l *FixedWindowLimiter) Allow() bool {
    l.mu.Lock()
    defer l.mu.Unlock()

    now := time.Now()
    if now.Sub(l.lastTime) >= l.window {
        l.count = 0
        l.lastTime = now
    }

    if l.count < l.limit {
        l.count++
        return true
    }
    return false
}
```

### 滑动窗口计数器

滑动窗口将固定窗口进一步细分为多个小格子，通过滑动统计来消除临界突刺。窗口越细，精度越高，但内存开销也越大。

数学原理：设窗口大小为 W，细分为 N 个格子，每个格子大小为 W/N。当前窗口的请求总数 = 最近 N 个格子的计数之和。

### 漏桶算法

请求像水一样注入漏桶，漏桶以固定速率漏水。超出桶容量的请求直接丢弃。漏桶的输出速率恒定，适合需要均匀处理请求的场景。

```go
package ratelimit

import "time"

type LeakyBucket struct {
    rate     float64       // 每秒漏出速率
    capacity float64       // 桶容量
    water    float64       // 当前水位
    lastLeak time.Time     // 上次漏水时间
}

func NewLeakyBucket(rate, capacity float64) *LeakyBucket {
    return &LeakyBucket{
        rate:     rate,
        capacity: capacity,
        lastLeak: time.Now(),
    }
}

func (b *LeakyBucket) Allow() bool {
    now := time.Now()
    elapsed := now.Sub(b.lastLeak).Seconds()

    // 先漏水
    b.water -= elapsed * b.rate
    if b.water < 0 {
        b.water = 0
    }
    b.lastLeak = now

    // 再加水
    if b.water+1 <= b.capacity {
        b.water++
        return true
    }
    return false
}
```

### 令牌桶算法

令牌桶以固定速率生成令牌，请求需要获取令牌才能通过。允许一定程度的突发流量（桶内有累积令牌时），是工业界最常用的限流算法。

```go
package ratelimit

import "time"

type TokenBucket struct {
    rate       float64    // 每秒生成令牌数
    capacity   float64    // 桶容量（最大令牌数）
    tokens     float64    // 当前令牌数
    lastRefill time.Time  // 上次补充令牌时间
}

func NewTokenBucket(rate, capacity float64) *TokenBucket {
    return &TokenBucket{
        rate:       rate,
        capacity:   capacity,
        tokens:     capacity, // 初始满桶
        lastRefill: time.Now(),
    }
}

func (b *TokenBucket) Allow() bool {
    now := time.Now()
    elapsed := now.Sub(b.lastRefill).Seconds()

    // 补充令牌
    b.tokens += elapsed * b.rate
    if b.tokens > b.capacity {
        b.tokens = b.capacity
    }
    b.lastRefill = now

    // 消耗令牌
    if b.tokens >= 1 {
        b.tokens--
        return true
    }
    return false
}
```

::: tip
令牌桶 vs 漏桶：令牌桶允许突发流量（积攒的令牌可以一次性消耗），适合大多数业务场景；漏桶输出均匀，适合对下游服务保护严格的场景。
:::

## 2. 分布式限流

### Redis + Lua 滑动窗口

单机限流无法满足分布式场景，需要借助 Redis 实现全局限流。使用 Lua 脚本保证原子性：

```go
package ratelimit

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisSlidingWindow struct {
    client *redis.Client
    limit  int
    window time.Duration
}

func NewRedisSlidingWindow(client *redis.Client, limit int, window time.Duration) *RedisSlidingWindow {
    return &RedisSlidingWindow{
        client: client,
        limit:  limit,
        window: window,
    }
}

var slidingWindowScript = redis.NewScript(`
    local key = KEYS[1]
    local now = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])
    local limit = tonumber(ARGV[3])
    local member = ARGV[4]

    redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
    local count = redis.call('ZCARD', key)
    if count < limit then
        redis.call('ZADD', key, now, member)
        redis.call('PEXPIRE', key, window)
        return 1
    end
    return 0
`)

func (r *RedisSlidingWindow) Allow(ctx context.Context, key string) (bool, error) {
    now := time.Now().UnixMilli()
    member := fmt.Sprintf("%d:%d", now, time.Now().UnixNano())
    result, err := slidingWindowScript.Run(ctx, r.client,
        []string{key},
        now, r.window.Milliseconds(), r.limit, member,
    ).Int64()
    if err != nil {
        return false, err
    }
    return result == 1, nil
}
```

### Token Server 集中式限流

Token Server 是独立的限流服务，所有节点向其申请令牌。优势是限流精确，劣势是存在单点瓶颈和网络开销。Sentinel 的集群限流就采用此模式。

::: warning
分布式限流需要考虑 Redis 本身的可用性。Redis 不可用时通常采用降级策略：切换为本地限流，使用较宽松的阈值。
:::

## 3. 自适应限流

### Google BBR 算法原理

BBR（Bottleneck Bandwidth and RTT）原本是 TCP 拥塞控制算法，被 Google 应用于服务限流。核心思想是根据系统的实际处理能力动态调整限流阈值，而非静态配置。

关键指标：

- **minRTT**：最近时间窗口内的最小响应时间
- **maxPass**：窗口内最大通过请求数
- **inFlight**：当前正在处理中的请求数

限流公式：当 `inFlight >= maxPass * minRT` 时触发限流。

### 基于系统负载的动态限流

```go
package ratelimit

import (
    "runtime"
    "sync/atomic"
    "time"
)

type AdaptiveLimiter struct {
    maxInFlight int64
    inFlight    int64
    cpuThreshold float64 // CPU 使用率阈值
}

func NewAdaptiveLimiter(maxInFlight int64, cpuThreshold float64) *AdaptiveLimiter {
    return &AdaptiveLimiter{
        maxInFlight:  maxInFlight,
        cpuThreshold: cpuThreshold,
    }
}

func (l *AdaptiveLimiter) Allow() bool {
    // 检查 CPU 负载
    if l.cpuUsage() > l.cpuThreshold {
        return false
    }
    // 检查并发数
    current := atomic.LoadInt64(&l.inFlight)
    if current >= l.maxInFlight {
        return false
    }
    atomic.AddInt64(&l.inFlight, 1)
    return true
}

func (l *AdaptiveLimiter) Done() {
    atomic.AddInt64(&l.inFlight, -1)
}

func (l *AdaptiveLimiter) cpuUsage() float64 {
    // 简化实现，生产环境可用 /proc/stat 或 gopsutil
    numGoroutine := runtime.NumGoroutine()
    numCPU := runtime.NumCPU()
    return float64(numGoroutine) / float64(numCPU) * 0.5
}
```

## 4. 限流粒度与策略

### 多级限流架构

实际生产中通常采用多级限流：

1. **全局限流**：保护整个系统，防止单点故障扩散
2. **用户级限流**：防止单个用户过度消费资源（如每用户 100 次/分钟）
3. **IP 级限流**：防御恶意请求和爬虫
4. **API 级限流**：不同接口设置不同阈值（核心接口更宽松）

```
请求 → 网关全局限流 → IP 限流 → 用户限流 → API 级限流 → 服务处理
```

::: info
限流 Key 的设计很关键。通常使用 `limit:{type}:{id}` 格式，如 `limit:user:12345`、`limit:ip:1.2.3.4`、`limit:api:/v1/orders`。
:::

## 5. 面试常见问题

### 限流被触发后怎么处理？

- **直接拒绝**：返回 429 Too Many Requests，适用于非核心接口
- **排队等待**：请求进入队列，超时则丢弃，适用于秒杀场景
- **降级处理**：返回缓存数据或默认值，适用于读接口
- **弹性扩容**：触发自动扩容，适用于可预测的流量高峰

### 限流与降级如何配合？

限流是第一道防线，降级是第二道防线。当限流仍然无法缓解压力时，逐步降级非核心功能：

1. 关闭推荐、评论等非核心功能
2. 返回静态页面或缓存数据
3. 暂停后台任务和定时任务
4. 限制写入，只允许读取

### 如何避免误杀？

- 设置合理的限流阈值（通常为正常流量的 1.5~2 倍）
- 使用滑动窗口代替固定窗口，减少临界突刺
- 实现限流预热：新上线的限流规则先以宽松阈值运行
- 提供白名单机制：内部服务调用、健康检查等不受限流影响

## 面试追问

- 固定窗口和滑动窗口的区别？滑动窗口的精度和内存开销如何权衡？
- 令牌桶和漏桶各自适合什么场景？能否组合使用？
- 分布式限流中 Redis 挂了怎么办？降级策略是什么？
- 如何设计一个支持百万 QPS 的限流系统？
- Sentinel 和 Hystrix 的限流机制有什么区别？
- 如何对限流系统本身做压测和容量规划？
