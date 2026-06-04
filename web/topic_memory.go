package main

func gcTopic() Topic {
	return Topic{
		ID: "gc", Title: "GC 垃圾回收", Chapter: "03-memory",
		Sections: []Section{
			s("三色标记算法",
				"Go 使用**并发三色标记-清除**算法。白色=未标记（将被回收），灰色=已标记但引用未扫描，黑色=已标记且引用已扫描。流程：1) 根对象→灰色 2) 取灰色扫描引用→引用变灰色 3) 当前→黑色 4) 重复直到无灰色 5) 白色=垃圾回收。**根对象**包括：全局变量、当前所有 goroutine 的栈变量。**三色不变式**：保证不会误删活跃对象。**生产影响**：GC 和用户代码并发执行（大部分时间），只有标记开始和结束时极短暂的 STW。",
				`白色 (White)  - 未标记, 将被回收
灰色 (Gray)   - 已标记, 引用未扫描
黑色 (Black)  - 已标记, 引用已扫描

标记过程:
1. STW: 开启写屏障
2. 根对象(stack, global) → 灰色
3. 并发标记(和用户代码并行):
   - 取灰色对象, 扫描其引用
   - 引用的对象 → 灰色
   - 当前对象 → 黑色
   - 重复直到无灰色
4. STW: 关闭写屏障, 清理
5. 所有白色对象 = 垃圾, 回收内存`),
			s("混合写屏障",
				"为什么需要写屏障？三色标记和用户代码并发执行时，可能出现黑色对象指向白色对象的引用（遗漏标记）。**Go 1.5-1.7**：Dijkstra 插入写屏障（A.field=B 时 B 标灰色），需要 STW 重新扫描栈（10-100ms）。**Go 1.8+**：混合写屏障 = Dijkstra + Yuasa。A.field=B 时：1) B 标灰色（Dijkstra）；2) 如果 A 在栈上，原值也标灰色（Yuasa）。结果：**不需要 STW 重新扫描栈**，总 STW < 100µs。**生产影响**：写屏障有 3-5% 的性能开销，但在 GC 期间才启用。",
				`// Go 1.8+ 混合写屏障
// 写操作: A.field = B
1. B 标记为灰色 (Dijkstra 插入写屏障)
2. 如果 A 在栈上, A.field 原值标记为灰色 (Yuasa)

// 为什么栈需要特殊处理？
// 栈上操作极其频繁, 加写屏障开销大
// 混合写屏障让栈上操作几乎零开销

// GC 各阶段:
// 1. 标记准备 (STW ~20µs): 开启写屏障
// 2. 并发标记 (~ms级): 和用户代码并行
// 3. 标记终止 (STW ~40µs): 关闭写屏障
// 4. 并发清除: 回收白色对象`),
			s("GC 触发条件与 GOGC 调优",
				"三种触发：1) **堆内存增长**达到 (1+GOGC/100)×上次存活堆（最常见）；2) **2 分钟**未触发 GC（sysmon 强制）；3) **手动** runtime.GC()。**GOGC 调优**：默认 100（堆翻倍时 GC）；200 更少 GC 但更多内存；50 更多 GC 更少内存。**Go 1.19+ GOMEMLIMIT**：设置运行时总内存软上限，比 GOGC 更直观。**生产建议**：容器环境用 GOMEMLIMIT（如容器限制 512MB → 设 450MB），让 Go 自己决定 GC 时机。",
				`// GOGC 调优
GOGC=100  // 默认, 堆翻倍时 GC（平衡）
GOGC=200  // 更少 GC, 更多内存（CPU密集型）
GOGC=50   // 更多 GC, 更少内存（内存敏感）
GOGC=off  // 禁用 GC（不推荐）

// Go 1.19+ GOMEMLIMIT（推荐）
debug.SetMemoryLimit(450 << 20) // 450MB 软上限
// 配合 GOGC=off → 固定内存上限
// 或 GOGC=默认 → 平衡 GC 和内存

// 生产: K8s 容器
// resources.limits.memory: 512Mi
// GOMEMLIMIT=450MiB (留余量给 Go runtime)
// GOGC=100 (默认即可)`),
			s("观察 GC",
				"**GODEBUG=gctrace=1**：打印每次 GC 详情，包括 STW 时间、堆大小变化、GC 占 CPU 比。**runtime.ReadMemStats**：代码中获取详细内存统计。**pprof**：heap profile 分析内存分配热点。**生产监控**：Prometheus client_golang 自动暴露 GC 指标（go_gc_duration_seconds、go_memstats_alloc_bytes 等）。**面试追问**：如何减少 GC 压力？→ 减少堆分配（逃逸分析、sync.Pool、预分配）。",
				`// 环境变量方式（最快）
// GODEBUG=gctrace=1 go run main.go
// gc 1 @0.003s 0%: 0.018+0.45+0.003 ms clock
// 含义:
// gc 1: 第1次GC
// @0.003s: 程序启动后0.003s
// 0.018ms: STW(标记开始)
// 0.45ms: 并发标记
// 0.003ms: STW(标记结束)
// 4->4->0 MB: GC前->GC后->存活

// 代码方式
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("堆: %.2f MB, GC: %d 次, 暂停: %v\n",
    float64(m.HeapAlloc)/1024/1024,
    m.NumGC,
    time.Duration(m.PauseTotalNs))

// 监控指标（Prometheus）
// go_gc_duration_seconds
// go_memstats_alloc_bytes
// go_goroutines`),
		},
	}
}

func escapeTopic() Topic {
	return Topic{
		ID: "escape", Title: "逃逸分析", Chapter: "03-memory",
		Sections: []Section{
			s("栈 vs 堆分配",
				"**栈分配**：函数返回后不被引用、编译期确定大小、分配/释放几乎零成本（移动 SP 指针）。**堆分配**：函数返回后仍被引用、大小编译期不确定、需要 GC 回收。性能差距：栈分配 ~1ns，堆分配 ~50-100ns（mallocgc + GC 开销）。如果每秒百万次堆分配，差距巨大。**生产影响**：热路径上的堆分配是性能瓶颈的常见原因。**面试追问**：逃逸分析是在编译期还是运行期？→ 编译期，由编译器决定。",
				`// 栈分配 (~1ns): 函数返回后不再使用
x := 42
arr := [100]int{}
p := Point{1, 2}

// 堆分配 (~50-100ns): 函数返回后仍被引用
func newPoint() *Point {
    p := Point{1, 2}
    return &p  // p 逃逸到堆上
}

// 查看: go build -gcflags="-m"
// main.go:3:6: moved to heap: p`),
			s("逃逸场景",
				"7 种常见逃逸场景：1) **返回局部变量指针**（最常见）；2) **赋值给 interface**（fmt.Println 参数、any 类型）；3) **闭包捕获变量**（变量生命周期延长）；4) **send 到 channel**（值被其他 goroutine 使用）；5) **slice/map 存储指针**（指向的数据逃逸）；6) **reflect 使用**（编译期无法确定类型）；7) **fmt.Printf 等可变参数**（...any 触发装箱）。**生产建议**：热路径避免 fmt.Sprintf，用 strconv；避免不必要的指针返回。",
				`// 1. 返回指针
func f() *int { x := 42; return &x }  // x 逃逸

// 2. interface（fmt 系列都会导致逃逸）
fmt.Sprintf("value: %d", x)  // x 逃逸到堆

// 3. 闭包
func counter() func() int {
    n := 0  // n 逃逸
    return func() int { n++; return n }
}

// 4. channel
ch <- data  // data 逃逸

// 5. fmt 可变参数
log.Printf("msg: %s", msg)  // msg 逃逸

// 查看逃逸:
// go build -gcflags="-m"        # 基本
// go build -gcflags="-m -m"     # 详细`),
			s("避免逃逸的技巧",
				"1) 小结构体用**值传递**（<=64 bytes 复制比堆分配快）；2) **预分配 slice** 容量（避免 append 触发扩容时的复制）；3) 用 **strconv** 代替 fmt.Sprintf（不触发 interface 装箱）；4) 热路径用**具体类型**代替 interface（避免动态分发和逃逸）；5) **sync.Pool** 复用对象；6) **小函数有利于内联**（内联后编译器能做更好的分析）。**生产场景**：JSON 序列化中避免 fmt → 用 strconv；高频日志中避免 fmt.Sprintf → 用 strings.Builder。",
				`// 1. strconv 代替 fmt（减少逃逸）
s := strconv.Itoa(42)       // ✓ 零分配
s := fmt.Sprintf("%d", 42)  // ✗ 有分配

// 2. 预分配 slice
s := make([]int, 0, n)  // 避免 append 扩容

// 3. sync.Pool 复用
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}
buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
defer bufPool.Put(buf)

// 4. 值传递小结构体
type Point struct{ X, Y float64 }
func (p Point) Distance() float64 { // 值接收者
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// 5. 避免 interface 导致的逃逸
func process(data []byte) error { // 具体 type
    // 而非 func process(r io.Reader)
}`),
		},
	}
}

func alignmentTopic() Topic {
	return Topic{
		ID: "alignment", Title: "内存对齐", Chapter: "03-memory",
		Sections: []Section{
			s("对齐规则",
				"三条规则：1) 字段按类型的对齐系数对齐（不足的补 padding）；2) struct 整体大小必须是最大对齐系数的倍数；3) 基本类型对齐系数 = 其大小（64位系统：bool=1, int32=4, int64=8）。**为什么需要内存对齐**：CPU 访问对齐的内存更快（一次总线操作），不对齐可能触发硬件异常（某些架构）。**生产影响**：百万个 struct 的微优化可能节省数 MB 内存。**面试追问**：unsafe.Alignof 和 unsafe.Sizeof 的区别？→ Alignof 返回对齐系数，Sizeof 返回实际大小。",
				`// Bad: 24 bytes（浪费 8 bytes padding）
type Bad struct {
    A bool    // 1 byte + 7 bytes padding
    B int64   // 8 bytes
    C int32   // 4 bytes + 4 bytes padding
}
fmt.Println(unsafe.Sizeof(Bad{}))  // 24

// Good: 16 bytes（零 padding）
type Good struct {
    B int64   // 8 bytes
    C int32   // 4 bytes
    A bool    // 1 byte + 3 bytes padding
}
fmt.Println(unsafe.Sizeof(Good{}))  // 16

// 节省 33% 内存！百万个实例省 8MB`),
			s("字段顺序优化",
				"原则：**按字段大小从大到小排列**，减少 padding。使用 **fieldalignment** 工具自动检测和修复（go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest）。**生产场景**：高并发服务中大量传输的 struct（如 HTTP 请求/响应体、数据库模型）。**面试追问**：编译器会自动优化字段顺序吗？→ 不会，Go 编译器保持字段声明顺序。",
				`// Bad: 24 bytes
type UserBad struct {
    Active  bool     // 1+7
    Balance float64  // 8
    Age     int32    // 4+4
}

// Good: 16 bytes (省 33%)
type UserGood struct {
    Balance float64  // 8
    Age     int32    // 4
    Active  bool     // 1+3
}

// 工具自动修复:
// fieldalignment -fix ./...

// 生产: 百万用户列表
// Bad:  24MB → Good: 16MB (省 8MB)
// 且缓存命中率更高（更紧凑 = 更多数据在缓存行中)`),
			s("atomic 对齐与零大小类型",
				"64 位原子操作要求 **8 字节对齐**，32 位系统上不对齐会 panic。atomic 变量应放在 struct **第一个字段**。**零大小类型 struct{}**：大小为 0，不占内存。用于实现 set（map[T]struct{}）和信号 channel（chan struct{}）。**生产场景**：原子计数器放 struct 首字段；连接管理用 map[string]struct{} 代替 map[string]bool。",
				`// atomic 变量放第一个字段
type Counter struct {
    count int64  // 第一个字段，保证 8 字节对齐
    flag  bool
}

// 32 位系统上以下代码可能 panic！
type BadCounter struct {
    flag  bool
    count int64  // 可能不对齐到 8 字节
}

// 零大小类型: set 实现
type Set struct {
    m map[string]struct{}
}
func (s *Set) Add(v string) { s.m[v] = struct{}{} }
func (s *Set) Has(v string) bool {
    _, ok := s.m[v]; return ok
}

// 信号 channel（不传数据）
done := make(chan struct{})
close(done)  // 广播通知`),
		},
	}
}

func profilingTopic() Topic {
	return Topic{
		ID: "profiling", Title: "Profiling 性能分析", Chapter: "04-performance",
		Sections: []Section{
			s("CPU Profiling",
				"原理：操作系统每秒发送 100 次 SIGPROF 信号（100Hz），每次记录所有线程的调用栈。采样结束后统计每个函数出现在调用栈中的次数，出现越多 = 占用 CPU 越多。**使用方式**：pprof.StartCPUProfile(f) 开始，pprof.StopCPUProfile() 结束。**分析命令**：top20 查看热点函数，web 生成火焰图（需要 graphviz），list funcName 查看行级耗时。**生产场景**：接口 RT 突然变慢 → CPU profile 找热点函数。",
				`f, _ := os.Create("cpu.prof")
defer f.Close()
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// 你的代码...
busyWork()

// 分析:
// go tool pprof cpu.prof
// (pprof) top20          # 热点函数
// (pprof) web            # 火焰图
// (pprof) list myFunc    # 行级耗时
// (pprof) peek myFunc    # 调用者和被调用者

// Web UI（推荐）:
// go tool pprof -http=:8080 cpu.prof`),
			s("Memory Profiling",
				"两种内存 profile：**allocs**（累计分配统计）和 **heap**（当前存活对象的分配统计）。用 pprof.WriteHeapProfile 写入。**分析维度**：-inuse_space（正在使用的内存）、-inuse_objects（正在使用的对象数）、-alloc_space（累计分配量）、-alloc_objects（累计分配次数）。**生产场景**：内存持续增长 → 对比两次 heap profile 找泄漏点。**面试追问**：allocs 和 heap 的区别？→ allocs 是累计值（从程序启动），heap 是当前值。",
				`f, _ := os.Create("mem.prof")
defer f.Close()
pprof.WriteHeapProfile(f)

// 分析:
// go tool pprof mem.prof
// (pprof) top20 -inuse_space   # 当前占用内存最多
// (pprof) top20 -alloc_space   # 累计分配最多
// (pprof) web
// (pprof) list myFunc

// 内存泄漏排查:
// 1. 两次 heap profile 对比
// go tool pprof -base mem1.prof mem2.prof
// 2. 看 inuse_space 增长的函数`),
			s("在线 Profiling",
				"**net/http/pprof** 自动注册 /debug/pprof/ 路由，无需额外代码（import _ \"net/http/pprof\"）。支持 CPU、heap、goroutine、thread、block、mutex 等 profile。**goroutine 泄漏排查**：查看 goroutine profile 中阻塞的调用栈，找到泄漏的 goroutine。**生产建议**：只在内部端口暴露 pprof，不要对外暴露。**面试追问**：block 和 mutex profile 需要额外设置？→ 是，需要 runtime.SetBlockProfileRate 和 runtime.SetMutexProfileFraction。",
				`import _ "net/http/pprof"
go http.ListenAndServe(":6060", nil)

// 在线 profiling:
// CPU (30秒采样):
// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
// Heap:
// go tool pprof http://localhost:6060/debug/pprof/heap
// Goroutine:
// go tool pprof http://localhost:6060/debug/pprof/goroutine

// goroutine 泄漏排查:
// 1. 看数量: curl localhost:6060/debug/pprof/goroutine?debug=1
// 2. 找到阻塞的调用栈
// 3. 定位泄漏原因

// 生产: 只在内部分析端口暴露
go func() {
    listener, _ := net.Listen("tcp", "127.0.0.1:6060")
    http.Serve(listener, nil)
}()`),
		},
	}
}

func benchmarkTopic() Topic {
	return Topic{
		ID: "benchmark", Title: "Benchmark 基准测试", Chapter: "04-performance",
		Sections: []Section{
			s("基本用法",
				"函数签名 func BenchmarkXxx(b *testing.B)，循环 b.N 次。b.N 由框架自动调整（从 1 开始，逐步增大），直到运行时间足够长（默认 >= 1s）结果可信。运行：go test -bench=. -benchmem。**输出解读**：671.4 ns/op（每次操作耗时）、2040 B/op（每次分配字节数）、8 allocs/op（每次分配次数）。**生产场景**：优化前后对比、竞品方案选型、CI 性能回归检测。",
				`func BenchmarkSliceAppend(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}

func BenchmarkSlicePrealloc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 100)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}
// go test -bench=. -benchmem
// Append:   671 ns/op  2040 B/op  8 allocs/op
// Prealloc:  73 ns/op     0 B/op  0 allocs/op`),
			s("避免编译器优化与子 Benchmark",
				"编译器可能优化掉没有副作用的代码。解决：将结果赋给**包级变量**。Go 1.24+ 可用 b.Loop() 更简洁。**子 Benchmark**：用 b.Run 创建子测试，对比多种实现方案。**b.ReportAllocs()**：即使不用 -benchmem 也会报告分配信息。",
				`var globalResult int

func BenchmarkSafe(b *testing.B) {
    var result int
    for i := 0; i < b.N; i++ {
        result = heavyComputation(i)
    }
    globalResult = result  // 防止优化
}

// Go 1.24+: b.Loop()
func BenchmarkLoop(b *testing.B) {
    for b.Loop() {
        heavyComputation(0)
    }
}

// 子 Benchmark（对比多种实现）
func BenchmarkConcat(b *testing.B) {
    parts := []string{"a", "b", "c", "d", "e"}
    b.Run("plus", func(b *testing.B) { /* ... */ })
    b.Run("builder", func(b *testing.B) { /* ... */ })
    b.Run("join", func(b *testing.B) { /* ... */ })
}`),
			s("benchstat 对比与实战技巧",
				"**benchstat** 用于统计显著性对比。优化前后各运行多次（-count=5），用 benchstat 对比，输出 delta 百分比和置信区间。**实战技巧**：1) -count=5 多次运行取中位数；2) -benchtime=5s 增加采样时间；3) -run=^$ 跳过单元测试加速；4) 关闭其他程序减少噪声。**生产场景**：PR 提交前跑 benchmark 确认没有性能退化。",
				`# 完整的 benchmark 对比流程

# 优化前（运行 5 次取统计）
go test -bench=BenchmarkConcat -count=5 -benchmem > old.txt

# 优化后
go test -bench=BenchmarkConcat -count=5 -benchmem > new.txt

# 对比
benchstat old.txt new.txt
# 输出:
# name         old time/op  new time/op  delta
# Concat/plus    120ns ± 2%    45ns ± 1%  -62.50%
# Concat/builder  80ns ± 1%    40ns ± 2%  -50.00%

# 常用参数:
# -benchmem     显示分配信息
# -count=5      多次运行
# -benchtime=5s 增加采样时间
# -run=^$       跳过单元测试`),
		},
	}
}

