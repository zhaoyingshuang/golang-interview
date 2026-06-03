package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

// ============================================================
// Profiling 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. 如何进行 CPU profiling？
// 2. 如何分析内存泄漏？
// 3. 如何排查 goroutine 泄漏？
// 4. pprof 的使用方式？
// 5. 火焰图怎么生成和分析？
// 6. 线上服务如何持续 profiling？

func main() {
	cpuProfile()
	memProfile()
	goroutineProfile()
	basicProfilingWorkflow()
}

// ----------------------------------------------------------
// 1. CPU Profiling
// ----------------------------------------------------------
// 原理: SIGPROF 信号（默认 100Hz，即每秒 100 次）
//   每次信号到来时，记录当前所有线程的调用栈
//   采样结束后，统计每个函数出现在调用栈中的次数
//   出现越多 = 占用 CPU 越多

func cpuProfile() {
	fmt.Println("=== 1. CPU Profiling ===")

	// 开始 CPU profiling
	f, err := os.CreateTemp("", "cpu.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}

	// 模拟 CPU 密集型工作
	busyWork()

	pprof.StopCPUProfile()

	fmt.Printf("CPU profile 写入: %s\n", f.Name())
	fmt.Println("分析: go tool pprof <file>")
	fmt.Println("  top20       - 查看耗时最多的函数")
	fmt.Println("  web         - 生成火焰图（需要 graphviz）")
	fmt.Println("  list <func> - 查看函数内每行代码的耗时")
	fmt.Println()
}

func busyWork() {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := 0
			for i := 0; i < 10_000_000; i++ {
				n += i
			}
			_ = n
		}()
	}
	wg.Wait()
}

// ----------------------------------------------------------
// 2. Memory Profiling
// ----------------------------------------------------------
// 两种内存 profile:
//   allocs: 累计分配统计（从程序启动到采样点）
//   heap:   当前存活对象的分配统计
//
// runtime/pprof.WriteHeapProfile 写的是 allocs profile
// 查看当前内存: 用 heap profile

func memProfile() {
	fmt.Println("=== 2. Memory Profiling ===")

	// 分配一些内存
	_ = make([]byte, 10<<20) // 10 MB

	f, err := os.CreateTemp("", "mem.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}

	fmt.Printf("Memory profile 写入: %s\n", f.Name())
	fmt.Println("分析: go tool pprof <file>")
	fmt.Println("  top20         - 查看分配最多的函数")
	fmt.Println("  top20 -inuse  - 查看正在使用的内存")
	fmt.Println("  top20 -alloc  - 查看累计分配")
	fmt.Println("  web           - 生成调用图")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Goroutine Profiling
// ----------------------------------------------------------
// 查看当前所有 goroutine 的调用栈
// 排查 goroutine 泄漏的利器

func goroutineProfile() {
	fmt.Println("=== 3. Goroutine Profiling ===")

	// 启动一些 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Second)
		}(i)
	}

	p := pprof.Lookup("goroutine")
	fmt.Printf("当前 goroutine 数: %d\n", p.Count())

	// 写入 goroutine profile
	f, err := os.CreateTemp("", "goroutine.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p.WriteTo(f, 1) // debug=1 显示调用栈

	fmt.Printf("Goroutine profile 写入: %s\n", f.Name())

	wg.Wait()

	fmt.Println("\n其他 runtime profile:")
	fmt.Println("  pprof.Lookup(\"thread\")     - OS 线程")
	fmt.Println("  pprof.Lookup(\"block\")      - 阻塞操作（需先 runtime.SetBlockProfileRate）")
	fmt.Println("  pprof.Lookup(\"mutex\")      - 锁竞争（需先 runtime.SetMutexProfileFraction）")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 完整 Profiling 工作流
// ----------------------------------------------------------
func basicProfilingWorkflow() {
	fmt.Println("=== 4. Profiling 工作流 ===")

	fmt.Println("方式1: 代码内嵌 (如上所示)")
	fmt.Println()
	fmt.Println("方式2: net/http/pprof (在线服务)")
	fmt.Println("  import _ \"net/http/pprof\"")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/goroutine")
	fmt.Println()
	fmt.Println("方式3: go test -cpuprofile -memprofile")
	fmt.Println("  go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof")
	fmt.Println("  go tool pprof cpu.prof")
	fmt.Println()
	fmt.Println("方式4: runtime/pprof.StartCPUProfile / WriteHeapProfile")
	fmt.Println()
	fmt.Println("火焰图生成:")
	fmt.Println("  go tool pprof -http=:8080 cpu.prof    # Web UI (推荐)")
	fmt.Println("  go tool pprof -png cpu.prof > flame.png # 静态图")
	fmt.Println()
	fmt.Println("常见分析场景:")
	fmt.Println("  CPU 瓶颈:  看 top20 + list <func> 找热点")
	fmt.Println("  内存泄漏:  对比两次 heap profile，看 inuse_space 增长")
	fmt.Println("  goroutine泄漏: 看调用栈中阻塞在哪里")
	fmt.Println("  锁竞争:   看 mutex profile，找到竞争最多的锁")
}
