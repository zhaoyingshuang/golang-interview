---
title: Channel 通道
---

## 1. 底层结构 hchan

Channel 底层是 runtime/chan.go 中的 hchan 结构体：一个**带锁的环形队列**。sendx/recvx 是环形队列的读写指针，recvq/sendq 是阻塞等待的 goroutine 链表（sudog 结构）。所有操作（send/recv/close）都需要加锁（runtime 级别的 mutex）。

**无缓冲 channel**：dataqsiz=0，没有缓冲区，发送和接收必须同时就绪。

**有缓冲 channel**：dataqsiz>0，环形缓冲区，发送方可以异步写入。

::: tip 生产选择
无缓冲用于同步信号（done、quit）；有缓冲用于数据传递（任务队列、结果收集）。

```go
type hchan struct {
    qcount   uint           // 队列中元素数
    dataqsiz uint           // 环形队列容量（缓冲大小）
    buf      unsafe.Pointer // 环形队列指针
    closed   uint32         // 是否已关闭（0=未关闭，1=已关闭）
    sendx    uint           // 发送索引
    recvx    uint           // 接收索引
    recvq    waitq          // 等待接收的 goroutine 队列
    sendq    waitq          // 等待发送的 goroutine 队列
    lock     mutex          // 互斥锁（所有操作都要加锁）
}

// 面试: channel 发送一个数据的完整流程？
// 1. 加锁
// 2. 如果 closed → 解锁, panic
// 3. 如果 recvq 有等待者 → 直接发送(不经缓冲区)
// 4. 如果缓冲区有空位 → 放入缓冲区
// 5. 否则 → 阻塞当前 goroutine, 加入 sendq
// 6. 解锁
```

## 2. 发送和接收规则

发送优先级：1) recvq 有等待者 → 直接发给它（不经缓冲区）；2) 缓冲区有空位 → 放入缓冲区；3) 否则阻塞。接收优先级：1) sendq 有等待者 → 直接取或从缓冲取；2) 缓冲区有数据 → 取；3) 否则阻塞。

::: warning 面试必记
向已关闭 channel 发送 → panic；关闭已关闭的 channel → panic；从已关闭 channel 接收 → 返回缓冲区剩余数据，之后返回零值+false。
:::

**生产模式**：close(ch) 用于广播通知（所有接收者都能收到零值）。

```go
// 关闭后的行为（面试必考）
ch := make(chan int, 2)
ch <- 1
ch <- 2
close(ch)

// 继续接收: 先取完缓冲区
v, ok := <-ch  // v=1, ok=true
v, ok = <-ch   // v=2, ok=true
v, ok = <-ch   // v=0, ok=false（缓冲区空了）

// ch <- 3      // panic: send on closed channel
// close(ch)    // panic: close of closed channel

// 生产: close 用于广播退出信号
quit := make(chan struct{})
go func() {
    defer close(quit) // 所有接收者都会收到
    doWork()
}()
<-quit // 等待完成
```

## 3. select 行为

select 的规则：1) 所有 case 会随机排序（避免某个 case 饥饿）；2) 按顺序评估哪些 case 可以立即执行；3) 多个 case 就绪时**随机选一个**（公平性）；4) 没有就绪且有 default → 执行 default（非阻塞）；5) 没有就绪且无 default → 阻塞。

**空 select**：select{} 永远阻塞。

::: tip 使用场景
非阻塞收发（default 模式）、超时控制（time.After）、多路复用（多个 channel 同时监听）。

::: warning 面试追问
select 中 case 的执行顺序？→ 多个就绪时随机，不是按书写顺序。

```go
// 非阻塞接收
select {
case v := <-ch:
    fmt.Println(v)
default:
    fmt.Println("无数据")
}

// 超时控制（生产常用）
select {
case result := <-ch:
    return result
case <-time.After(5 * time.Second):
    return errors.New("timeout")
}

// 多路复用（生产: 多个数据源）
select {
case msg := <-ch1:
    handleType1(msg)
case msg := <-ch2:
    handleType2(msg)
case <-ctx.Done():
    return ctx.Err()
}

// 面试: 空 select 的作用？
select {} // 永远阻塞，常用于 main 中防退出
```

## 4. Channel vs Mutex

Go 的哲学："Don't communicate by sharing memory; share memory by communicating." 但不是所有场景都该用 channel。
:::

**Channel 适合**：传递数据所有权（生产者-消费者）、等待/通知（done signal）、超时控制、pipeline。

**Mutex 适合**：保护共享缓存（LRU cache）、计数器、配置读写、并发安全地操作复杂数据结构。

**生产选择原则**：goroutine 之间传递数据用 channel；多个 goroutine 读写同一块数据用 mutex。不要为了"Go 风格"硬用 channel——简单场景 mutex 更清晰。

```go
// Channel: 传递数据所有权
func producer(ch chan<- int) {
    for i := 0; i < 100; i++ {
        ch <- i  // 数据所有权转移给消费者
    }
    close(ch)
}

// Mutex: 保护共享状态
type SafeCache struct {
    mu   sync.RWMutex
    data map[string]string
}
func (c *SafeCache) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

// 生产: 常见错误 - 用 channel 做本该用 mutex 的事
// ❌ 通过 channel 传递"请修改这个变量"的消息
// ✓ 直接用 mutex 保护变量
```
