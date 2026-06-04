package main

func goroutineTopic() Topic {
	return Topic{
		ID: "goroutine", Title: "Goroutine 协程", Chapter: "02-concurrency",
		Sections: []Section{
			s("GMP 调度模型",
				"**G**=Goroutine（用户态协程，初始栈 2KB），**M**=Machine（OS 线程），**P**=Processor（逻辑处理器，数量=GOMAXPROCS，默认等于 CPU 核心数）。**为什么需要 P**：Go 早期只有 GM 模型（全局队列），问题：全局锁竞争激烈、缓存不友好。引入 P 后每个 P 有独立队列（256 个 G），减少了锁竞争。**调度流程**：新 G 优先放当前 P 的本地队列 → 满了一半放入全局队列 → M 绑定 P 后从本地队列取 G → 本地空则从全局取或从其他 P 偷（work stealing）→ 系统调用阻塞时 M 释放 P 让其他 M 继续跑。**生产影响**：GOMAXPROCS 默认等于 CPU 核心数，CPU 密集型任务不需要修改；I/O 密集型可以适当增大。",
				`G (Goroutine) - 用户态协程, 初始栈 2KB, 可增长到 1GB
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
runtime.NumGoroutine()   // 当前 goroutine 数`),
			s("Goroutine vs 线程",
				"**栈大小**：2KB（可增长到 1GB）vs 1-8MB（固定）。**创建开销**：~0.3µs vs ~10-100µs（goroutine 只需分配栈和几个结构体）。**切换开销**：~100ns（用户态，只保存 ~3 个寄存器）vs ~1-10µs（内核态，需要保存/恢复完整上下文）。**最大数量**：轻松上百万 vs 几千到几万。**通信方式**：channel（CSP 模型）vs 共享内存+锁。**生产影响**：Go 可以在一个 HTTP 请求处理中启动多个 goroutine 并发查询数据库/缓存，而 Java/C++ 做同样的事需要线程池。**面试追问**：goroutine 和 coroutine 的区别？→ goroutine 支持抢占式调度（Go 1.14+），不需要显式 yield。",
				`| 特性         | Goroutine      | OS Thread      |
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
fmt.Println(time.Since(start)) // ~30ms`),
			s("栈增长",
				"Go 1.3+ 使用**连续栈**（contiguous stack）：空间不够时分配 2 倍新栈，复制数据，释放旧栈。解决了分段栈的 \"hot split\" 问题（栈在边界频繁增减导致性能抖动）。GC 时如果栈使用率 < 1/4，缩减为原来的一半。初始 2KB，最大 1GB。**生产影响**：递归函数不需要担心栈溢出（除非超过 1GB）；goroutine 初始只有 2KB，可以轻松创建百万个。**面试追问**：栈增长时地址会变吗？→ 会，因为分配了新的连续内存，Go 会自动更新所有指向旧栈的指针。",
				`// 连续栈机制:
// 空间不足 → 分配 2x 新栈 → 复制数据 → 更新指针 → 释放旧栈
// GC 时使用率 < 25% → 缩减为原来的一半
// 初始: 2KB, 最大: 1GB

// 面试: 栈增长时会暂停 goroutine 吗？
// 会，但只暂停当前 goroutine，不影响其他

// 面试: 什么操作会导致栈增长？
// 函数调用时检查栈空间是否足够
// 递归调用是最常见的触发场景`),
			s("Goroutine 泄漏",
				"**最常见的 goroutine 泄漏原因**：1) 无缓冲 channel 没有接收者/发送者（goroutine 永远阻塞）；2) context 没有 cancel（goroutine 永远不会退出）；3) range channel 没有关闭（消费者永远等待）；4) WaitGroup 计数错误（Wait 永远阻塞）。**生产检测**：runtime.NumGoroutine() 监控数量趋势；pprof goroutine profile 查看阻塞的调用栈；监控告警 goroutine 数量持续增长。**修复模式**：所有 goroutine 都要有明确的退出条件。",
				`// 泄漏场景1: channel 没有接收者
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
pprof.Lookup("goroutine").Count()`),
			s("抢占调度",
				"Go 1.14 之前：协作式抢占（编译器在函数入口插入栈检查），紧凑循环（无函数调用）无法被抢占，可能饿死其他 goroutine。Go 1.14+：**基于信号的异步抢占**（SIGURG），信号处理器中设置抢占标志，goroutine 在安全点被挂起。**生产影响**：纯计算循环不再需要手动加 runtime.Gosched()。**调度时机总结**：channel 操作、系统调用、time.Sleep、runtime.Gosched()、GC STW、函数调用栈检查、信号抢占。",
				`// Go 1.14 前: 紧凑循环会饿死其他 goroutine
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
// 7. SIGURG 信号（异步抢占，Go 1.14+）`),
		},
	}
}

func channelTopic() Topic {
	return Topic{
		ID: "channel", Title: "Channel 通道", Chapter: "02-concurrency",
		Sections: []Section{
			s("底层结构 hchan",
				"Channel 底层是 runtime/chan.go 中的 hchan 结构体：一个**带锁的环形队列**。sendx/recvx 是环形队列的读写指针，recvq/sendq 是阻塞等待的 goroutine 链表（sudog 结构）。所有操作（send/recv/close）都需要加锁（runtime 级别的 mutex）。**无缓冲 channel**：dataqsiz=0，没有缓冲区，发送和接收必须同时就绪。**有缓冲 channel**：dataqsiz>0，环形缓冲区，发送方可以异步写入。**生产选择**：无缓冲用于同步信号（done、quit）；有缓冲用于数据传递（任务队列、结果收集）。",
				`type hchan struct {
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
// 6. 解锁`),
			s("发送和接收规则",
				"发送优先级：1) recvq 有等待者 → 直接发给它（不经缓冲区）；2) 缓冲区有空位 → 放入缓冲区；3) 否则阻塞。接收优先级：1) sendq 有等待者 → 直接取或从缓冲取；2) 缓冲区有数据 → 取；3) 否则阻塞。**面试必记三条**：向已关闭 channel 发送 → panic；关闭已关闭的 channel → panic；从已关闭 channel 接收 → 返回缓冲区剩余数据，之后返回零值+false。**生产模式**：close(ch) 用于广播通知（所有接收者都能收到零值）。",
				`// 关闭后的行为（面试必考）
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
<-quit // 等待完成`),
			s("select 行为",
				"select 的规则：1) 所有 case 会随机排序（避免某个 case 饥饿）；2) 按顺序评估哪些 case 可以立即执行；3) 多个 case 就绪时**随机选一个**（公平性）；4) 没有就绪且有 default → 执行 default（非阻塞）；5) 没有就绪且无 default → 阻塞。**空 select**：select{} 永远阻塞。**生产场景**：非阻塞收发（default 模式）、超时控制（time.After）、多路复用（多个 channel 同时监听）。**面试追问**：select 中 case 的执行顺序？→ 多个就绪时随机，不是按书写顺序。",
				`// 非阻塞接收
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
select {} // 永远阻塞，常用于 main 中防退出`),
			s("Channel vs Mutex",
				"Go 的哲学：\"Don't communicate by sharing memory; share memory by communicating.\" 但不是所有场景都该用 channel。**Channel 适合**：传递数据所有权（生产者-消费者）、等待/通知（done signal）、超时控制、pipeline。**Mutex 适合**：保护共享缓存（LRU cache）、计数器、配置读写、并发安全地操作复杂数据结构。**生产选择原则**：goroutine 之间传递数据用 channel；多个 goroutine 读写同一块数据用 mutex。不要为了\"Go 风格\"硬用 channel——简单场景 mutex 更清晰。",
				`// Channel: 传递数据所有权
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
// ✓ 直接用 mutex 保护变量`),
		},
	}
}

func syncTopic() Topic {
	return Topic{
		ID: "sync", Title: "Sync 同步原语", Chapter: "02-concurrency",
		Sections: []Section{
			s("Mutex 正常/饥饿模式",
				"**正常模式**：新来的 goroutine 和刚被唤醒的竞争锁（新来的有优势，因为已经在 CPU 上运行）。如果竞争失败，被放到队列前面。**饥饿模式**：某个 goroutine 等待超过 1ms 时触发，锁直接交给队列头部的 goroutine（不竞争），防止饿死。当最后一个等待者获取锁或等待时间 < 1ms，切回正常模式。**不可重入**：Go 的 Mutex 不是可重入锁，同一 goroutine 重复 Lock 会死锁。**生产场景**：保护计数器、缓存、配置。**面试追问**：为什么不支持可重入？→ 容易导致设计问题（函数是否持有锁变得不透明）。",
				`// Mutex 状态位（面试了解）
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
}`),
			s("RWMutex",
				"多个读者可同时持有读锁，写者独占。内部结构：1 个 Mutex + 2 个信号量 + 2 个计数器。readerCount 为负数时表示有写者在等待。**适用场景**：读多写少（如配置热更新、缓存）。**不适用**：读写都频繁（RWMutex 比 Mutex 更重，内部有更多 atomic 操作）。**生产坑**：读锁未释放就加写锁 → 死锁。**面试追问**：RWMutex 的写锁饥饿问题？→ Go 的 RWMutex 有防止写锁饥饿的机制（新读者在有写者等待时会被阻塞）。",
				`var rw sync.RWMutex
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
// 解决: 确保读写锁不嵌套`),
			s("WaitGroup",
				"WaitGroup 用于等待一组 goroutine 完成。state1 高 32 位是计数器，低 32 位是等待者数量。**最关键的使用规则**：Add 必须在 go 之前调用（在外部），Done 在 goroutine 内部 defer 调用。如果 Add 放在 goroutine 内部，Wait 可能在所有 Add 之前就返回。**生产场景**：批量并发请求、并行数据处理。**面试追问**：Add 的计数能变负吗？→ 不能，会 panic。",
				`var wg sync.WaitGroup

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
}`),
			s("sync.Once",
				"保证某个操作只执行一次。实现：1) 原子加载 done 标志（快速路径）；2) 加锁；3) double-check done；4) 执行 f()；5) 原子存储 done=1。**Go 1.21+ 提供 OnceValues** 支持返回值。**注意**：如果 f() panic，done 不会被设置为 1，后续调用会再次执行 f()。**生产场景**：初始化数据库连接、加载配置、单例模式。**面试追问**：Once 和 init 的区别？→ init 在包加载时执行，Once 在首次调用时执行（延迟初始化）。",
				`var once sync.Once
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
// done 不会被设为 1，下次调用会再执行 f()`),
			s("sync.Pool",
				"对象复用池，减少 GC 压力。每个 P 有本地池（无锁），Get 先从本地取，取不到从其他 P 偷，再取不到调用 New 创建。**关键特性：GC 时会清理池中对象**，所以不适合做连接池！**适用场景**：高频创建和销毁的临时对象（bytes.Buffer、JSON 编码器）。**标准库使用**：fmt.Printf 内部用 Pool 复用 buffer；encoding/json 用 Pool 复用编码器。**面试追问**：Pool 和连接池的区别？→ Pool 的对象会在 GC 时丢失，连接池需要自己管理生命周期。",
				`var bufPool = sync.Pool{
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
// 不适合做数据库连接池！`),
			s("Atomic 操作",
				"底层使用 CPU 原子指令（LOCK 前缀 + CMPXCHG 等），不需要加锁，性能极高。Go 1.19+ 提供 **atomic.Int64、atomic.Bool、atomic.Pointer[T]** 等类型，比旧的 atomic.AddInt64 等函数更易用。**CAS（CompareAndSwap）** 是无锁编程的基础。**适用场景**：计数器、状态标志、无锁队列。**不适用**：复杂操作（多步非原子）、需要等待的场景（用 channel/mutex）。**生产场景**：请求计数、服务状态标记（atomic.Bool）、无锁缓存（atomic.Pointer）。",
				`// atomic.Int64（Go 1.19+，推荐）
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
if ready.Load() { /* ... */ }`),
		},
	}
}

func contextTopic() Topic {
	return Topic{
		ID: "context", Title: "Context 上下文", Chapter: "02-concurrency",
		Sections: []Section{
			s("Context 接口",
				"Context 是一个接口：Deadline() 返回截止时间，Done() 返回取消信号 channel，Err() 返回取消原因，Value() 返回请求作用域的值。4 种内部实现：**emptyCtx**（根 context，永远不会取消）、**cancelCtx**（可取消）、**timerCtx**（带超时，内部包含 cancelCtx）、**valueCtx**（带值，链表结构）。**生产场景**：每个 HTTP 请求都有自己的 context（req.Context()），请求结束后自动取消。**面试追问**：context.Background() 和 context.TODO() 的区别？→ Background 是根 context，TODO 是不确定该用什么时的占位。",
				`type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}

// 4 种实现:
// emptyCtx  → context.Background() / TODO()
// cancelCtx → context.WithCancel()
// timerCtx  → context.WithTimeout() / WithDeadline()
// valueCtx  → context.WithValue()

// 面试: context 作为函数参数的规范？
// 必须是第一个参数，不要放在 struct 中
func Process(ctx context.Context, data string) error {
    // ...
}`),
			s("取消传播",
				"**取消是单向传播的**：父取消 → 子自动取消；子取消 → 不影响父和兄弟。cancel() 的内部流程：1) 设置 err；2) 关闭 done channel；3) 遍历 children 依次取消；4) 从父的 children 中移除自己。**生产场景**：用户取消请求 → 所有下游数据库查询、RPC 调用自动取消。**最佳实践**：cancel() 必须 defer 调用，即使确认会超时也要 cancel（释放内部资源）。",
				`ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 创建子 context
child1, cancel1 := context.WithCancel(ctx)
defer cancel1()
child2, _ := context.WithCancel(ctx)
grandchild, _ := context.WithCancel(child1)

cancel1()
// child1.Err() → context.Canceled
// grandchild.Err() → context.Canceled（子被取消）
// child2.Err() → nil（兄弟不受影响）

// 生产: HTTP handler 中
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // 请求结束后自动取消
    result, err := fetchData(ctx, query)
    // 如果用户断开连接，ctx 自动取消
}`),
			s("WithTimeout/WithDeadline",
				"**WithTimeout** 从当前时间计时，**WithDeadline** 指定绝对时间点。WithTimeout 就是 WithDeadline(ctx, time.Now().Add(timeout))。**必须 defer cancel()**：即使已经超时，cancel 也必须调用以释放内部 timer goroutine，否则会 goroutine 泄漏。**生产场景**：数据库查询超时（通常 3-5s）、RPC 调用超时（通常 1-10s）、HTTP 客户端超时。**面试追问**：超时时间应该设多少？→ 取决于业务，但要小于上游的超时时间。",
				`// 数据库查询超时
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()  // 必须！防止 goroutine 泄漏
rows, err := db.QueryContext(ctx, "SELECT ...")

// RPC 调用超时
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
resp, err := client.Call(ctx, request)

// 面试: 不调用 cancel 会怎样？
// timer goroutine 会一直运行，直到超时
// 如果 timeout 很长（甚至没有），goroutine 就泄漏了`),
			s("WithValue 最佳实践",
				"Key 用**自定义类型**（避免冲突），不要用内建类型（string/int）。**适合传递**：trace ID、auth token、request-scoped logger。**不适合传递**：数据库连接、业务参数、函数必须的参数（应该用函数参数）。**生产反模式**：把所有东西都塞进 context → 应该只放请求级别的元数据。**面试追问**：context.Value 的查找效率？→ O(n)，沿 valueCtx 链表向上查找，层数多了会慢。",
				`// 正确: 自定义 key 类型
type requestIDKey struct{}
type userIDKey struct{}

ctx = context.WithValue(ctx, requestIDKey{}, "req-123")
ctx = context.WithValue(ctx, userIDKey{}, 42)

// 类型安全地获取
id, _ := ctx.Value(requestIDKey{}).(string)
uid, _ := ctx.Value(userIDKey{}).(int)

// 生产: middleware 中设置 trace ID
func TraceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := uuid.New().String()
        ctx := context.WithValue(r.Context(),
            requestIDKey{}, traceID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 面试: 为什么不用 string 做 key？
// 不同包可能用相同的 string，导致冲突
// type myKey string → 也行，但空 struct 更安全`),
		},
	}
}

func patternTopic() Topic {
	return Topic{
		ID: "pattern", Title: "并发模式", Chapter: "02-concurrency",
		Sections: []Section{
			s("Worker Pool",
				"控制并发数量，避免创建过多 goroutine。模式：jobs channel 分发任务，N 个 worker 通过 range jobs 消费，results channel 收集结果。**生产场景**：批量处理 API 请求、并发下载文件、并行数据处理。**为什么不用 go + WaitGroup**：Worker Pool 复用 goroutine，避免频繁创建销毁；通过 worker 数量精确控制并发。**面试追问**：worker 数量设多少？→ CPU 密集型 = CPU 核心数，I/O 密集型可以更多（10-100）。",
				`func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup
    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := range jobs {  // range 直到 close
                results <- process(j)
            }
        }()
    }
    go func() { wg.Wait(); close(results) }()
}

// 使用
jobs := make(chan Job, 100)
results := make(chan Result, 100)
go workerPool(jobs, results, runtime.NumCPU())

for _, j := range allJobs { jobs <- j }
close(jobs)
for r := range results { /* 收集结果 */ }`),
			s("Fan-out / Fan-in",
				"**Fan-out**：一个 channel 分发给多个 goroutine 处理，提高吞吐。**Fan-in**：多个 channel 合并到一个，统一消费。**和 Worker Pool 的区别**：Fan-out 的每个 worker 可以有不同的处理逻辑；Worker Pool 的 worker 相同。**生产场景**：日志处理（多种日志源合并）、数据聚合（从多个 API 获取数据后合并）。**面试追问**：merge 函数为什么要单独启动 goroutine 关闭 out？→ 要等所有 source channel 关闭后才能关闭 out。",
				`// Fan-out: 多个 worker 并行处理同一个输入
func fanOut(in <-chan int, workers int) []<-chan int {
    outs := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        outs[i] = process(in)  // 每个 worker 读同一个 in
    }
    return outs
}

// Fan-in: 合并多个 channel
func merge(chs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, ch := range chs {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c { out <- v }
        }(ch)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}`),
			s("Pipeline",
				"数据在多个 stage 之间流过，每个 stage 是一个 goroutine + channel。stage1 → stage2 → stage3 → 结果。每个 stage 通过 context 支持取消，任何 stage 出错都能优雅退出。**生产场景**：数据 ETL（读取 → 转换 → 写入）、图片处理（读取 → 缩放 → 水印 → 保存）、日志分析（读取 → 解析 → 过滤 → 聚合）。**优势**：每个 stage 可以独立伸缩（不同的 goroutine 数量），解耦清晰。",
				`// Pipeline: gen → square → filter → 结果
func gen(ctx context.Context, values ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, v := range values {
            select {
            case out <- v:
            case <-ctx.Done(): return
            }
        }
    }()
    return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * v:
            case <-ctx.Done(): return
            }
        }
    }()
    return out
}

// 使用: pipeline 组合
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
p := filter(ctx, square(ctx, gen(ctx, 1, 2, 3, 4, 5)))
for v := range p { fmt.Println(v) }`),
			s("优雅退出 (Graceful Shutdown)",
				"模式：context/信号控制生命周期 → 收到退出信号后停止接收新请求 → 等待进行中的请求完成 → 关闭资源。用 signal.Notify + context.WithCancel 实现信号监听，用 WaitGroup 或 errgroup 等待所有 goroutine 退出。**生产场景**：HTTP 服务收到 SIGTERM 后优雅退出（K8s Pod 滚动更新）。**面试追问**：为什么不直接 os.Exit？→ 进行中的请求会被截断，可能导致数据不一致。",
				`func main() {
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: ":8080"}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    <-ctx.Done() // 等待信号
    log.Println("shutting down...")

    // 优雅关闭：5s 超时
    shutdownCtx, cancel := context.WithTimeout(
        context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(shutdownCtx) // 等待进行中的请求完成
    log.Println("server stopped")
}`),
		},
	}
}

