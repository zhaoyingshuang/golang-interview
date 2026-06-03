---
title: Sync 同步原语
---

## 1. Mutex 正常/饥饿模式

**正常模式**：新来的 goroutine 和刚被唤醒的竞争锁（新来的有优势，因为已经在 CPU 上运行）。如果竞争失败，被放到队列前面。

**饥饿模式**：某个 goroutine 等待超过 1ms 时触发，锁直接交给队列头部的 goroutine（不竞争），防止饿死。当最后一个等待者获取锁或等待时间 < 1ms，切回正常模式。

**不可重入**：Go 的 Mutex 不是可重入锁，同一 goroutine 重复 Lock 会死锁。

::: tip 使用场景
保护计数器、缓存、配置。

::: warning 面试追问
为什么不支持可重入？→ 容易导致设计问题（函数是否持有锁变得不透明）。

```go
// Mutex 状态位（面试了解）
// state: 0=未锁定, 1=已锁定, 2=已唤醒, 4=饥饿

var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()

// 不可重入！以下会死锁
mu.Lock()
mu.Lock()  // deadlock!

// 生产: defer 释放锁
func (s *Service) Update() {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ...
}

// 生产: TryLock（Go 1.18+，非阻塞）
if mu.TryLock() {
    // 获取成功
    mu.Unlock()
} else {
    // 获取失败
}
```

## 2. RWMutex

多个读者可同时持有读锁，写者独占。内部结构：1 个 Mutex + 2 个信号量 + 2 个计数器。readerCount 为负数时表示有写者在等待。
:::

**适用场景**：读多写少（如配置热更新、缓存）。

**不适用**：读写都频繁（RWMutex 比 Mutex 更重，内部有更多 atomic 操作）。

::: danger 生产陷阱
读锁未释放就加写锁 → 死锁。

::: warning 面试追问
RWMutex 的写锁饥饿问题？→ Go 的 RWMutex 有防止写锁饥饿的机制（新读者在有写者等待时会被阻塞）。

```go
var rw sync.RWMutex
data := make(map[string]string)

// 读操作: 多个 goroutine 可以并行读
func Get(key string) string {
    rw.RLock()
    defer rw.RUnlock()
    return data[key]
}

// 写操作: 独占
func Set(key, value string) {
    rw.Lock()
    defer rw.Unlock()
    data[key] = value
}

// 生产坑: 读锁中调用写锁 → 死锁
rw.RLock()
rw.Lock()   // deadlock!
// 解决: 确保读写锁不嵌套
```

## 3. WaitGroup

WaitGroup 用于等待一组 goroutine 完成。state1 高 32 位是计数器，低 32 位是等待者数量。
:::

**最关键的使用规则**：Add 必须在 go 之前调用（在外部），Done 在 goroutine 内部 defer 调用。如果 Add 放在 goroutine 内部，Wait 可能在所有 Add 之前就返回。

::: tip 使用场景
批量并发请求、并行数据处理。

::: warning 面试追问
Add 的计数能变负吗？→ 不能，会 panic。

```go
var wg sync.WaitGroup

// 正确用法: Add 在 go 之前
for i := 0; i < 10; i++ {
    wg.Add(1)  // 在 go 之前！
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}
wg.Wait()

// 错误用法: Add 在 goroutine 内部
for i := 0; i < 10; i++ {
    go func() {
        wg.Add(1)    // ❌ 可能 Wait 已经执行了
        defer wg.Done()
    }()
}
wg.Wait()  // 可能在所有 Add 之前就返回

// 生产: 批量并发请求
func fetchAll(urls []string) []Result {
    var wg sync.WaitGroup
    results := make([]Result, len(urls))
    for i, url := range urls {
        wg.Add(1)
        go func(idx int, u string) {
            defer wg.Done()
            results[idx] = fetch(u)
        }(i, url)
    }
    wg.Wait()
    return results
}
```

## 4. sync.Once

保证某个操作只执行一次。实现：1) 原子加载 done 标志（快速路径）；2) 加锁；3) double-check done；4) 执行 f()；5) 原子存储 done=1。
:::

**Go 1.21+ 提供 OnceValues** 支持返回值。

**注意**：如果 f() panic，done 不会被设置为 1，后续调用会再次执行 f()。

::: tip 使用场景
初始化数据库连接、加载配置、单例模式。

::: warning 面试追问
Once 和 init 的区别？→ init 在包加载时执行，Once 在首次调用时执行（延迟初始化）。

```go
var once sync.Once
var config *Config

func GetConfig() *Config {
    once.Do(func() {
        // 只执行一次（延迟初始化）
        config = loadConfig()
    })
    return config
}

// Go 1.21+ OnceValues（带返回值）
var loadDB = sync.OnceValues(func() (*sql.DB, error) {
    return sql.Open("mysql", dsn)
})
db, err := loadDB()  // 首次调用时初始化，后续调用返回缓存

// 面试: 如果 f() panic 会怎样？
// done 不会被设为 1，下次调用会再执行 f()
```

## 5. sync.Pool

对象复用池，减少 GC 压力。每个 P 有本地池（无锁），Get 先从本地取，取不到从其他 P 偷，再取不到调用 New 创建。
:::

**关键特性：GC 时会清理池中对象**，所以不适合做连接池！

**适用场景**：高频创建和销毁的临时对象（bytes.Buffer、JSON 编码器）。

**标准库使用**：fmt.Printf 内部用 Pool 复用 buffer；encoding/json 用 Pool 复用编码器。

::: warning 面试追问
Pool 和连接池的区别？→ Pool 的对象会在 GC 时丢失，连接池需要自己管理生命周期。

```go
var bufPool = sync.Pool{
    New: func() any {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func Process(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()          // 重置（但不释放内存）
    defer bufPool.Put(buf)

    buf.Write(data)
    return buf.String()
}

// 生产: JSON 编码复用
var encoderPool = sync.Pool{
    New: func() any { return json.NewEncoder(io.Discard) },
}

// 注意: GC 时池中对象会被清理
// 不适合做数据库连接池！
```

## 6. Atomic 操作

底层使用 CPU 原子指令（LOCK 前缀 + CMPXCHG 等），不需要加锁，性能极高。Go 1.19+ 提供 **atomic.Int64、atomic.Bool、atomic.Pointer[T]** 等类型，比旧的 atomic.AddInt64 等函数更易用。
:::

**CAS（CompareAndSwap）** 是无锁编程的基础。

**适用场景**：计数器、状态标志、无锁队列。

**不适用**：复杂操作（多步非原子）、需要等待的场景（用 channel/mutex）。

::: tip 使用场景
请求计数、服务状态标记（atomic.Bool）、无锁缓存（atomic.Pointer）。

```go
// atomic.Int64（Go 1.19+，推荐）
var count atomic.Int64
count.Add(1)
v := count.Load()
count.Store(0)

// CAS: 无锁更新
var state atomic.Int64
state.Store(0)  // 0=idle, 1=running
if state.CompareAndSwap(0, 1) {
    // 从 idle 切到 running，只有一方能成功
    defer state.Store(0)
    doWork()
}

// atomic.Pointer[T]（Go 1.19+）
var cache atomic.Pointer[Config]
cache.Store(&Config{Version: 1})
cfg := cache.Load() // 无锁读取

// atomic.Bool
var ready atomic.Bool
ready.Store(true)
if ready.Load() { /* ... */ }
```
