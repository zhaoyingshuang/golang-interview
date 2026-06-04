package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================
// 限流设计
// ============================================================

func main() {
	fixedWindow()
	tokenBucketDemo()
	slidingWindowDemo()
	leakyBucket()
	distributedRateLimit()
}

// ----------------------------------------------------------
// 1. 固定窗口
// ----------------------------------------------------------
type FixedWindowLimiter struct {
	mu       sync.Mutex
	count    int
	limit    int
	window   time.Duration
	resetAt  time.Time
}

func (fw *FixedWindowLimiter) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	if now.After(fw.resetAt) {
		fw.count = 0
		fw.resetAt = now.Add(fw.window)
	}

	if fw.count < fw.limit {
		fw.count++
		return true
	}
	return false
}

func fixedWindow() {
	fmt.Println("=== 1. 固定窗口算法 ===")
	limiter := &FixedWindowLimiter{limit: 5, window: time.Second, resetAt: time.Now().Add(time.Second)}

	passed := 0
	for i := 0; i < 8; i++ {
		if limiter.Allow() {
			passed++
		}
	}
	fmt.Printf("  通过 %d/8 (限制: 5/s)\n", passed)
	fmt.Println("  问题: 窗口边界可能通过 2x 流量 (00:59 的 5 个 + 01:01 的 5 个)")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 令牌桶
// ----------------------------------------------------------
type TokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64
	rate     float64
	lastTime time.Time
}

func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		tokens:   capacity,
		capacity: capacity,
		rate:     rate,
		lastTime: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.tokens += now.Sub(tb.lastTime).Seconds() * tb.rate
	tb.lastTime = now
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func tokenBucketDemo() {
	fmt.Println("=== 2. 令牌桶算法 ===")
	bucket := NewTokenBucket(5, 10)
	passed := 0
	for i := 0; i < 15; i++ {
		if bucket.Allow() {
			passed++
		}
	}
	fmt.Printf("  通过 %d/15 (rate=5/s, capacity=10)\n", passed)
	fmt.Println("  特点: 允许突发 (桶中有积攒令牌)")
	fmt.Println("  Go 标准库: golang.org/x/time/rate")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 滑动窗口
// ----------------------------------------------------------
type SlidingWindowLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	limit    int
	requests []time.Time
}

func (sw *SlidingWindowLimiter) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	valid := sw.requests[:0]
	for _, t := range sw.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	sw.requests = valid

	if len(sw.requests) < sw.limit {
		sw.requests = append(sw.requests, now)
		return true
	}
	return false
}

func slidingWindowDemo() {
	fmt.Println("=== 3. 滑动窗口算法 ===")
	limiter := &SlidingWindowLimiter{window: time.Second, limit: 5}
	passed := 0
	for i := 0; i < 8; i++ {
		if limiter.Allow() {
			passed++
		}
	}
	fmt.Printf("  通过 %d/8 (限制: 5/s)\n", passed)
	fmt.Println("  特点: 没有固定窗口的边界问题，更平滑")
	fmt.Println("  优化: 分成多个小格子，用环形数组实现 O(1) 操作")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 漏桶
// ----------------------------------------------------------
func leakyBucket() {
	fmt.Println("=== 4. 漏桶算法 ===")
	fmt.Println("  原理: 请求进入桶中，以固定速率流出处理")
	fmt.Println("  桶满则拒绝新请求")
	fmt.Println()
	fmt.Println("  与令牌桶的区别:")
	fmt.Println("  ┌──────────┬─────────────────┬─────────────────┐")
	fmt.Println("  │ 维度     │ 令牌桶          │ 漏桶            │")
	fmt.Println("  ├──────────┼─────────────────┼─────────────────┤")
	fmt.Println("  │ 突发流量 │ 允许            │ 不允许          │")
	fmt.Println("  │ 速率     │ 可变(≤rate)     │ 恒定            │")
	fmt.Println("  │ 场景     │ API限流         │ 流量整形        │")
	fmt.Println("  └──────────┴─────────────────┴─────────────────┘")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 分布式限流
// ----------------------------------------------------------
func distributedRateLimit() {
	fmt.Println("=== 5. 分布式限流 ===")
	fmt.Println()
	fmt.Println("方案1: Redis + Lua 脚本")
	fmt.Println("  优点: 精确   缺点: 每次请求都有网络延迟")
	fmt.Println()
	fmt.Println("方案2: Token Server (集中式)")
	fmt.Println("  令牌服务器定期批量发放令牌给各节点")
	fmt.Println("  节点本地消耗令牌，减少 Redis 调用")
	fmt.Println("  代表: Sentinel Cluster Flow Control")
	fmt.Println()
	fmt.Println("方案3: 本地 + 分布式二级限流")
	fmt.Println("  全局限流: QPS / 节点数 = 单节点配额")
	fmt.Println("  本地执行: 每个节点独立执行令牌桶")
	fmt.Println("  优点: 零延迟   缺点: 不精确(节点数动态变化)")
	fmt.Println()
	fmt.Println("方案4: 滑动日志 (精确但成本高)")
	fmt.Println("  Redis Sorted Set 存储每个请求的时间戳")
	fmt.Println("  ZREMRANGEBYSCORE + ZCARD + ZADD")
	fmt.Println("  优点: 精确   缺点: 内存消耗大")
	fmt.Println()
	fmt.Println("面试追问: 限流被触发后怎么处理?")
	fmt.Println("  1. 返回 429 + Retry-After 头")
	fmt.Println("  2. 排队等待 (适用于可以延迟的场景)")
	fmt.Println("  3. 降级 (返回缓存数据/默认值)")
	fmt.Println("  4. 转发 (路由到备用服务)")
	fmt.Println()
}
