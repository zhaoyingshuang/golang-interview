---
title: Goroutine 协程
---

## 1. GMP 调度模型

**G**=Goroutine（用户态协程，初始栈 2KB），**M**=Machine（OS 线程），**P**=Processor（逻辑处理器，数量=GOMAXPROCS，默认等于 CPU 核心数）。

::: info 为什么需要 P
Go 早期只有 GM 模型（全局队列），问题：全局锁竞争激烈、缓存不友好。引入 P 后每个 P 有独立队列（256 个 G），减少了锁竞争。
:::

**调度流程**：新 G 优先放当前 P 的本地队列 → 满了一半放入全局队列 → M 绑定 P 后从本地队列取 G → 本地空则从全局取或从其他 P 偷（work stealing）→ 系统调用阻塞时 M 释放 P 让其他 M 继续跑。

::: info 生产影响
GOMAXPROCS 默认等于 CPU 核心数，CPU 密集型任务不需要修改；I/O 密集型可以适当增大。

```go
G (Goroutine) - 用户态协程, 初始栈 2KB, 可增长到 1GB
M (Machine)   - OS 线程, 由操作系统调度
P (Processor) - 逻辑处理器, 默认数=CPU核心数

调度流程:
1. 新 G → 当前 P 的本地队列（最多 256 个）
2. 本地队列满 → 前一半放入全局队列
3. M 绑定 P, 从 P 本地队列取 G 执行
4. 本地空 → 从全局取一批 / 从其他 P 偷(work stealing)
5. G 系统调用阻塞 → M 释放 P, P 绑新 M 继续

// 查看当前状态
runtime.NumCPU()         // CPU 核心数
runtime.GOMAXPROCS(0)    // P 的数量
runtime.NumGoroutine()   // 当前 goroutine 数
```

## 2. Goroutine vs 线程

**栈大小**：2KB（可增长到 1GB）vs 1-8MB（固定）。
:::

**创建开销**：~0.3µs vs ~10-100µs（goroutine 只需分配栈和几个结构体）。

**切换开销**：~100ns（用户态，只保存 ~3 个寄存器）vs ~1-10µs（内核态，需要保存/恢复完整上下文）。

**最大数量**：轻松上百万 vs 几千到几万。

**通信方式**：channel（CSP 模型）vs 共享内存+锁。

::: info 生产影响
Go 可以在一个 HTTP 请求处理中启动多个 goroutine 并发查询数据库/缓存，而 Java/C++ 做同样的事需要线程池。

::: warning 面试追问
goroutine 和 coroutine 的区别？→ goroutine 支持抢占式调度（Go 1.14+），不需要显式 yield。

```go
| 特性         | Goroutine      | OS Thread      |
|-------------|----------------|----------------|
| 栈大小       | 2KB（可增长到1GB）| 1-8MB（固定）   |
| 创建开销     | ~0.3µs         | ~10-100µs      |
| 切换开销     | ~100ns 用户态   | ~1-10µs 内核态  |
| 最大数量     | 百万级          | 几千            |
| 调度方式     | 协作+抢占       | 抢占式          |
| 通信方式     | channel(CSP)   | 共享内存+锁     |

// 实测: 创建 10 万个 goroutine
start := time.Now()
var wg sync.WaitGroup
wg.Add(100000)
for i := 0; i < 100000; i++ {
    go func() { defer wg.Done() }()
}
wg.Wait()
fmt.Println(time.Since(start)) // ~30ms
```

## 3. 栈增长

Go 1.3+ 使用**连续栈**（contiguous stack）：空间不够时分配 2 倍新栈，复制数据，释放旧栈。解决了分段栈的 "hot split" 问题（栈在边界频繁增减导致性能抖动）。GC 时如果栈使用率 < 1/4，缩减为原来的一半。初始 2KB，最大 1GB。

::: info 生产影响
递归函数不需要担心栈溢出（除非超过 1GB）；goroutine 初始只有 2KB，可以轻松创建百万个。

::: warning 面试追问
栈增长时地址会变吗？→ 会，因为分配了新的连续内存，Go 会自动更新所有指向旧栈的指针。

```go
// 连续栈机制:
// 空间不足 → 分配 2x 新栈 → 复制数据 → 更新指针 → 释放旧栈
// GC 时使用率 < 25% → 缩减为原来的一半
// 初始: 2KB, 最大: 1GB

// 面试: 栈增长时会暂停 goroutine 吗？
// 会，但只暂停当前 goroutine，不影响其他

// 面试: 什么操作会导致栈增长？
// 函数调用时检查栈空间是否足够
// 递归调用是最常见的触发场景
```

## 4. Goroutine 泄漏

**最常见的 goroutine 泄漏原因**：1) 无缓冲 channel 没有接收者/发送者（goroutine 永远阻塞）；2) context 没有 cancel（goroutine 永远不会退出）；3) range channel 没有关闭（消费者永远等待）；4) WaitGroup 计数错误（Wait 永远阻塞）。
:::

**生产检测**：runtime.NumGoroutine() 监控数量趋势；pprof goroutine profile 查看阻塞的调用栈；监控告警 goroutine 数量持续增长。

**修复模式**：所有 goroutine 都要有明确的退出条件。

```go
// 泄漏场景1: channel 没有接收者
ch := make(chan int)
go func() { ch <- result }()  // 没人接收，永远阻塞
// 修复: 缓冲 channel 或 context 取消

// 泄漏场景2: context 没有 cancel
go func() {
    for {
        select {
        case <-ch:
            // 处理
        // 缺少 case <-ctx.Done()！永远不会退出
        }
    }
}()
// 修复: 加上 ctx.Done() 分支

// 泄漏场景3: range channel 没有关闭
go func() {
    for v := range ch {  // ch 永远不关闭
        process(v)
    }  // 永远到不了这里
}()
// 修复: 发送方 close(ch)

// 检测 goroutine 泄漏
fmt.Printf("goroutines: %d\n", runtime.NumGoroutine())
// 或用 pprof
pprof.Lookup("goroutine").Count()
```

## 5. 抢占调度

Go 1.14 之前：协作式抢占（编译器在函数入口插入栈检查），紧凑循环（无函数调用）无法被抢占，可能饿死其他 goroutine。Go 1.14+：**基于信号的异步抢占**（SIGURG），信号处理器中设置抢占标志，goroutine 在安全点被挂起。

::: info 生产影响
纯计算循环不再需要手动加 runtime.Gosched()。
:::

**调度时机总结**：channel 操作、系统调用、time.Sleep、runtime.Gosched()、GC STW、函数调用栈检查、信号抢占。

```go
// Go 1.14 前: 紧凑循环会饿死其他 goroutine
go func() {
    for i := 0; i < 1e10; i++ {
        _ = i * i  // 无函数调用，无法被抢占！
    }
}()

// Go 1.14+: 基于信号抢占，上面的代码也能被抢占

// 调度时机（面试总结）:
// 1. channel 操作（发送/接收阻塞）
// 2. 系统调用（文件/网络 IO）
// 3. time.Sleep
// 4. runtime.Gosched()（主动让出）
// 5. GC STW
// 6. 函数调用时的栈检查（协作式）
// 7. SIGURG 信号（异步抢占，Go 1.14+）
```
