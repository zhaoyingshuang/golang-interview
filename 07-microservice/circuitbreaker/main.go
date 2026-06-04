package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================
// 限流熔断降级
// ============================================================
//
// 【面试高频问题】
// 1. 四种限流算法的区别？各适用什么场景？
// 2. 熔断器的三种状态和转换条件？
// 3. 令牌桶和漏桶的区别？
// 4. 分布式限流如何实现？
// 5. Google SRE 熔断算法原理？

func main() {
	tokenBucket()
	slidingWindow()
	circuitBreaker()
	distributedRateLimit()
	adaptiveLimiter()
}

// ----------------------------------------------------------
// 1. 令牌桶算法
// ----------------------------------------------------------
// 原理: 以固定速率向桶中放入令牌，请求需要获取令牌才能通过
//
// 参数:
//   rate     — 每秒放入的令牌数
//   capacity — 桶的最大容量（允许突发）
//
// 特点:
//   - 允许突发流量（桶中有积累的令牌时）
//   - 长期来看速率被限制在 rate
//   - Go 标准库 golang.org/x/time/rate 就是令牌桶实现
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
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// 补充令牌
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func tokenBucket() {
	fmt.Println("=== 1. 令牌桶算法 ===")

	bucket := NewTokenBucket(10, 5) // 每秒10个，桶容量5
	passed := 0
	blocked := 0

	for i := 0; i < 20; i++ {
		if bucket.Allow() {
			passed++
		} else {
			blocked++
		}
	}

	fmt.Printf("  令牌桶 (rate=10/s, capacity=5): 通过=%d, 拒绝=%d\n", passed, blocked)
	fmt.Println("  特点: 允许突发（桶中有积攒令牌时）")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 滑动窗口限流
// ----------------------------------------------------------
// 原理: 维护一个时间窗口内的请求计数，窗口随时间滑动
//
// 优化: 将窗口划分为多个小格子（如 1 秒窗口分为 10 个 100ms 格子）
// 每个格子独立计数，滑动时只需移动格子指针
type SlidingWindow struct {
	mu       sync.Mutex
	window   time.Duration
	limit    int
	requests []time.Time
}

func NewSlidingWindow(window time.Duration, limit int) *SlidingWindow {
	return &SlidingWindow{
		window: window,
		limit:  limit,
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// 清理过期请求
	valid := sw.requests[:0]
	for _, t := range sw.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	sw.requests = valid

	if len(sw.requests) >= sw.limit {
		return false
	}

	sw.requests = append(sw.requests, now)
	return true
}

func slidingWindow() {
	fmt.Println("=== 2. 滑动窗口限流 ===")

	window := NewSlidingWindow(time.Second, 5)
	passed := 0
	for i := 0; i < 10; i++ {
		if window.Allow() {
			passed++
		}
	}
	fmt.Printf("  滑动窗口 (1s内最多5个): 通过=%d/10\n", passed)

	fmt.Println()
	fmt.Println("  算法对比:")
	fmt.Println("  ┌──────────┬────────────┬────────────────┐")
	fmt.Println("  │ 算法     │ 突发流量   │ 适用场景       │")
	fmt.Println("  ├──────────┼────────────┼────────────────┤")
	fmt.Println("  │ 固定窗口 │ 有临界问题 │ 简单限流       │")
	fmt.Println("  │ 滑动窗口 │ 较平滑     │ API 限流       │")
	fmt.Println("  │ 漏桶     │ 不允许     │ 流量整形       │")
	fmt.Println("  │ 令牌桶   │ 允许       │ 允许突发的限流 │")
	fmt.Println("  └──────────┴────────────┴────────────────┘")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 熔断器
// ----------------------------------------------------------
// 三种状态:
//
//   Closed (关闭)  → 正常通过请求，统计失败率
//     失败率 > 阈值 → Open
//
//   Open (打开)    → 直接拒绝请求（快速失败）
//     超时后 → Half-Open
//
//   Half-Open (半开) → 放少量请求探测
//     成功 → Closed (恢复)
//     失败 → Open (继续熔断)
type State int

const (
	Closed    State = 0
	Open      State = 1
	HalfOpen  State = 2
)

type CircuitBreaker struct {
	mu          sync.Mutex
	state       State
	failures    int
	successes   int
	threshold   int
	openTimeout time.Duration
	openedAt    time.Time
}

func NewCircuitBreaker(threshold int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       Closed,
		threshold:   threshold,
		openTimeout: openTimeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Closed:
		return true
	case Open:
		if time.Since(cb.openedAt) > cb.openTimeout {
			cb.state = HalfOpen
			cb.successes = 0
			cb.failures = 0
			return true // 放一个探测请求
		}
		return false
	case HalfOpen:
		return cb.successes < 1 // 只放少量请求
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == HalfOpen {
		cb.successes++
		if cb.successes >= 1 {
			cb.state = Closed
			cb.failures = 0
		}
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.state == HalfOpen || cb.failures >= cb.threshold {
		cb.state = Open
		cb.openedAt = time.Now()
	}
}

func circuitBreaker() {
	fmt.Println("=== 3. 熔断器 ===")

	cb := NewCircuitBreaker(3, 5*time.Second)

	// 模拟正常请求
	fmt.Println("  正常阶段:")
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
		fmt.Printf("    失败 %d/3\n", i+1)
	}

	// 熔断打开
	fmt.Printf("  熔断状态: Allow=%v (应该为 false)\n", cb.Allow())

	fmt.Println()
	fmt.Println("  Google SRE 熆断 (自适应):")
	fmt.Println("    throttling = (requests - K * accepts) / (requests + 1)")
	fmt.Println("    K = 系数(如 1.5), requests = 总请求数, accepts = 成功数")
	fmt.Println("    当后端开始拒绝时，客户端自动降低发送速率")
	fmt.Println("    不需要硬编码阈值，自适应调整")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 分布式限流
// ----------------------------------------------------------
func distributedRateLimit() {
	fmt.Println("=== 4. 分布式限流 ===")
	fmt.Println()
	fmt.Println("Redis + Lua 滑动窗口:")
	fmt.Println()
	fmt.Println("  -- KEYS[1] = 限流key")
	fmt.Println("  -- ARGV[1] = 窗口大小(秒)")
	fmt.Println("  -- ARGV[2] = 最大请求数")
	fmt.Println("  -- ARGV[3] = 当前时间戳(毫秒)")
	fmt.Println("  local key = KEYS[1]")
	fmt.Println("  local window = tonumber(ARGV[1])")
	fmt.Println("  local limit = tonumber(ARGV[2])")
	fmt.Println("  local now = tonumber(ARGV[3])")
	fmt.Println()
	fmt.Println("  -- 移除窗口外的记录")
	fmt.Println("  redis.call('ZREMRANGEBYSCORE', key, 0, now - window * 1000)")
	fmt.Println()
	fmt.Println("  -- 获取当前窗口内的请求数")
	fmt.Println("  local count = redis.call('ZCARD', key)")
	fmt.Println()
	fmt.Println("  if count < limit then")
	fmt.Println("    redis.call('ZADD', key, now, now .. '-' .. math.random())")
	fmt.Println("    redis.call('EXPIRE', key, window)")
	fmt.Println("    return 1  -- 允许")
	fmt.Println("  else")
	fmt.Println("    return 0  -- 拒绝")
	fmt.Println("  end")
	fmt.Println()
	fmt.Println("  面试追问: 分布式限流有什么问题?")
	fmt.Println("    1. 时钟偏移 — 不同节点时间不一致")
	fmt.Println("    2. Redis 延迟 — 每次请求都访问 Redis")
	fmt.Println("    3. 单点问题 — Redis 挂了怎么办")
	fmt.Println("    解决: 本地+分布式二级限流、Token Server 集中式方案")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 自适应限流 (BBR)
// ----------------------------------------------------------
func adaptiveLimiter() {
	fmt.Println("=== 5. 自适应限流 (BBR) ===")
	fmt.Println()
	fmt.Println("TCP BBR 算法思想应用到限流:")
	fmt.Println("  不预设 QPS 阈值，而是根据系统当前负载自适应调整")
	fmt.Println()
	fmt.Println("核心指标:")
	fmt.Println("  inflight  = 正在处理中的请求数")
	fmt.Println("  minRTT    = 最近窗口内的最小延迟")
	fmt.Println("  maxPass   = 窗口内最大通过量")
	fmt.Println()
	fmt.Println("判断条件 (满足则拒绝):")
	fmt.Println("  inflight > maxPass * minRTT * beta")
	fmt.Println("  (即: 当前排队量 > 系统最大吞吐 × 最小延迟 × 安全系数)")
	fmt.Println()
	fmt.Println("Go 实现 (阿里 Sentinel Go):")
	fmt.Println("  import sentinel \"github.com/alibaba/sentinel-golang/api\"")
	fmt.Println("  import \"github.com/alibaba/sentinel-golang/core/flow\"")
	fmt.Println()
	fmt.Println("  _, err := flow.LoadRules([]*flow.Rule{{")
	fmt.Println("    Resource:               \"api:/users\",")
	fmt.Println("    TokenCalculateStrategy: flow.Direct,")
	fmt.Println("    ControlBehavior:        flow.Reject,")
	fmt.Println("    Threshold:              1000,")
	fmt.Println("  }})")
	fmt.Println()
}
