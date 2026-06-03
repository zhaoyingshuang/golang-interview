package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================================
// Goroutine 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. goroutine 和线程的区别？
// 2. GMP 调度模型是什么？
// 3. goroutine 泄漏的原因和检测？
// 4. goroutine 的栈大小？如何增长？
// 5. runtime.GOMAXPROCS 的作用？
// 6. goroutine 什么时候会切换（让出 CPU）？

func main() {
	gmpModel()
	goroutineVsThread()
	stackGrowth()
	contextSwitch()
	goroutineLeak()
	gosched()
}

// ----------------------------------------------------------
// 1. GMP 调度模型
// ----------------------------------------------------------
// G - Goroutine: 用户态协程，初始栈 2KB（Go 1.4+），可增长到 1GB
// M - Machine: 操作系统线程，由 OS 调度
// P - Processor: 逻辑处理器，数量默认等于 CPU 核心数（GOMAXPROCS）
//
// 调度流程:
//   1. 每个 P 有一个本地队列（local run queue），最多 256 个 G
//   2. 创建新 G 时优先放入当前 P 的本地队列
//   3. 本地队列满时，把前一半 G 放入全局队列（global run queue）
//   4. M 绑定 P 后，从 P 的本地队列取 G 执行
//   5. 本地队列空时:
//      a. 从全局队列取一批 G
//      b. 从其他 P 的本地队列偷一半（work stealing）
//   6. 如果 G 发生系统调用阻塞，M 会释放 P，让其他 M 接管 P 继续执行
//
// 为什么需要 P？
//   Go 早期只有 GM 模型（全局队列），问题:
//   - 全局锁竞争激烈
//   - 没有局部性（G 可能在不同 M 上执行，缓存不友好）
//   引入 P 后，每个 P 有独立队列，减少了锁竞争

func gmpModel() {
	fmt.Println("=== 1. GMP 调度模型 ===")

	fmt.Printf("CPU 核心数: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d (P 的数量)\n", runtime.GOMAXPROCS(0))
	fmt.Printf("当前 goroutine 数: %d\n", runtime.NumGoroutine())

	// 调度时机（goroutine 让出 CPU 的时刻）:
	//   1. channel 操作阻塞
	//   2. 系统调用（文件/网络 IO）
	//   3. time.Sleep
	//   4. runtime.Gosched() 主动让出
	//   5. 函数调用时栈检查点（抢占式调度的辅助）
	//   6. runtime.GC() 触发 STW

	fmt.Println("\nGMP 关键点:")
	fmt.Println("  - 本地队列 256 个 G，满了放全局队列")
	fmt.Println("  - 本地空 → 从全局取 / 从其他 P 偷 (work stealing)")
	fmt.Println("  - 系统调用阻塞 → M 释放 P，P 绑定新 M 继续跑")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Goroutine vs 线程
// ----------------------------------------------------------
// | 特性           | Goroutine          | OS Thread          |
// |----------------|-------------------|--------------------|
// | 栈大小          | 2KB（可增长到1GB）  | 1-8MB（固定）       |
// | 创建开销        | ~0.3µs            | ~10-100µs         |
// | 切换开销        | ~100ns（用户态）    | ~1-10µs（内核态）   |
// | 调度方式        | 协作+抢占          | 抢占式             |
// | 最大数量        | 轻松上百万         | 几千到几万          |
// | 通信方式        | channel（CSP）     | 共享内存+锁        |

func goroutineVsThread() {
	fmt.Println("=== 2. Goroutine vs 线程 ===")

	start := time.Now()
	var wg sync.WaitGroup
	const n = 100000

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("创建 %d 个 goroutine 耗时: %v\n", n, time.Since(start))
	fmt.Printf("当前 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 栈增长（分段栈 → 连续栈）
// ----------------------------------------------------------
// Go 1.3 之前: 分段栈（segmented stack）
//   - 栈空间不够时，分配新段，通过链表连接
//   - 问题: "hot split" — 栈在边界频繁增减导致性能抖动
//
// Go 1.3+: 连续栈（contiguous stack）
//   - 栈空间不够时，分配一个 2 倍大的新栈
//   - 把旧栈数据复制到新栈
//   - 释放旧栈
//   - 栈缩容: GC 时如果栈使用率 < 1/4，缩减为原来的一半
//
// 栈初始大小: 2KB (Go 1.4+)
// 最大栈大小: 1GB (64位系统默认)

func stackGrowth() {
	fmt.Println("=== 3. 栈增长 ===")

	var stackDepth func(n int) int
	stackDepth = func(n int) int {
		if n <= 0 {
			// 当前栈使用情况
			buf := make([]byte, 1)
			return len(buf)
		}
		return stackDepth(n - 1)
	}

	// 深度递归，观察栈是否增长
	fmt.Printf("递归 1000 层正常执行（栈自动增长）\n")
	stackDepth(1000)
	fmt.Println("连续栈: 空间不足时分配 2 倍新栈并复制数据")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 上下文切换
// ----------------------------------------------------------
// goroutine 的上下文切换只涉及 ~3 个寄存器（PC/SP/DX），
// 不需要经过内核态，所以比线程快很多。

func contextSwitch() {
	fmt.Println("=== 4. 上下文切换 ===")

	const iterations = 100000
	var wg sync.WaitGroup
	ch := make(chan struct{})

	// 测量 goroutine 切换开销
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			ch <- struct{}{}
		}
	}()

	start := time.Now()
	for i := 0; i < iterations; i++ {
		<-ch
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("%d 次 channel 切换耗时: %v (%.0f ns/op)\n",
		iterations, elapsed, float64(elapsed.Nanoseconds())/iterations)
	fmt.Println()
}

// ----------------------------------------------------------
// 5. Goroutine 泄漏
// ----------------------------------------------------------
// goroutine 泄漏: 创建了 goroutine 但它永远无法退出。
// 常见原因:
//   1. channel 没有发送者/接收者，goroutine 永远阻塞
//   2. 忘记调用 cancel()（context 未取消）
//   3. 死锁
//   4. 无限循环没有退出条件
//
// 检测方式:
//   1. runtime.NumGoroutine()
//   2. runtime/pprof 的 goroutine profile
//   3. github.com/google/pprof 的 web UI

func goroutineLeak() {
	fmt.Println("=== 5. Goroutine 泄漏 ===")

	// 泄漏示例1: channel 阻塞
	leakyChan := func() <-chan int {
		ch := make(chan int)
		go func() {
			result := expensiveComputation()
			ch <- result // 如果没人接收，永远阻塞！
		}()
		return ch
	}

	// 修复: 使用缓冲 channel 或 context
	fixedChan := func() <-chan int {
		ch := make(chan int, 1) // 缓冲为 1，不会阻塞
		go func() {
			result := expensiveComputation()
			ch <- result
		}()
		return ch
	}

	_ = leakyChan
	_ = fixedChan

	// 泄漏示例2: 生产者没有关闭 channel
	leakyProducer := func() {
		ch := make(chan int)
		go func() {
			for i := 0; i < 100; i++ {
				ch <- i
			}
			// 忘记 close(ch)！消费者的 range 会永远阻塞
		}()
		// 只消费了 50 个就退出，生产者的后续发送永远阻塞
		for i := 0; i < 50; i++ {
			<-ch
		}
	}

	_ = leakyProducer

	fmt.Println("常见泄漏场景:")
	fmt.Println("  1. 无缓冲 channel 没有接收者")
	fmt.Println("  2. context 没有 cancel")
	fmt.Println("  3. range channel 没有关闭")
	fmt.Println("  4. WaitGroup 计数不正确")
	fmt.Println()
}

func expensiveComputation() int {
	return 42
}

// ----------------------------------------------------------
// 6. Gosched 和抢占
// ----------------------------------------------------------
// Go 1.14 之前: 协作式抢占
//   - 编译器在函数入口插入栈检查（stack check）
//   - 如果栈不够或需要抢占，调用 runtime.morestack()
//   - 问题: 紧凑循环（无函数调用）无法被抢占
//
// Go 1.14+: 基于信号的异步抢占
//   - 发送 SIGURG 信号给运行时间过长的 goroutine
//   - 信号处理器中设置抢占标志
//   - goroutine 在安全点被挂起
//   - 解决了紧凑循环饿死其他 goroutine 的问题

func gosched() {
	fmt.Println("=== 6. 抢占调度 ===")

	// runtime.Gosched(): 主动让出 CPU
	// 很少需要手动调用，调度器会自动处理

	done := make(chan bool)
	go func() {
		// 紧凑循环：Go 1.14+ 可以被信号抢占
		for i := 0; i < 1e8; i++ {
			// 纯计算，没有函数调用
			_ = i
		}
		done <- true
	}()

	fmt.Println("Go 1.14+ 基于信号的异步抢占解决了紧凑循环饿死问题")
	<-done
	fmt.Println()
}
