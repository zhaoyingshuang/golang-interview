---
title: 限流熔断降级
---

## 1. 四种限流算法

| 算法 | 突发流量 | 复杂度 | 适用场景 |
|------|----------|--------|----------|
| 固定窗口 | 有临界问题 | O(1) | 简单限流 |
| 滑动窗口 | 较平滑 | O(n) | API 限流 |
| 漏桶 | 不允许 | O(1) | 流量整形 |
| 令牌桶 | 允许 | O(1) | 允许突发的限流 |

::: tip 令牌桶 vs 漏桶
令牌桶允许突发（桶中有积攒令牌时），适合 API 限流。漏桶以固定速率输出，适合流量整形。Go 标准库 `golang.org/x/time/rate` 是令牌桶实现。

```go
limiter := rate.NewLimiter(100, 10) // 100/s, 桶容量10
if !limiter.Allow() {
    http.Error(w, "rate limit", http.StatusTooManyRequests)
    return
}
```
:::

## 2. 熔断器模式

三种状态：**Closed**（正常通过）→ **Open**（直接拒绝）→ **Half-Open**（放少量探测）

```
                失败率 > 阈值
  Closed ────────────────────→ Open
    ↑                            │
    │ 探测成功                    │ 超时后
    │                            ↓
    └─────────── Half-Open ←─────┘
```

::: warning Google SRE 熔断（自适应）
不硬编码阈值：`throttling = (requests - K * accepts) / (requests + 1)`。当后端开始拒绝时，客户端自动降低发送速率。

```go
// gobreaker 熔断器
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "user-service",
    MaxRequests: 3,               // Half-Open 时放 3 个探测
    Interval:    10 * time.Second, // 统计周期
    Timeout:     30 * time.Second, // Open → Half-Open 等待时间
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
})
result, err := cb.Execute(func() (any, error) {
    return callService()
})
```
:::

## 3. 降级策略

- **返回默认值**：非核心功能降级（推荐列表 → 返回热门数据）
- **缓存兜底**：服务不可用时返回过期缓存
- **异步化**：同步调用失败 → 写 MQ 异步处理

## 4. 分布式限流

```lua
-- Redis + Lua 滑动窗口限流
local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
redis.call('ZREMRANGEBYSCORE', key, 0, now - window * 1000)
local count = redis.call('ZCARD', key)
if count < limit then
    redis.call('ZADD', key, now, now .. '-' .. math.random())
    redis.call('EXPIRE', key, window)
    return 1
else
    return 0
end
```

::: warning 面试追问
分布式限流问题：时钟偏移、Redis 延迟、单点问题。解决：本地+分布式二级限流、Token Server 集中式方案。
:::
