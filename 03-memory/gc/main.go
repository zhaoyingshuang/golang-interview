package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// ============================================================
// GC 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. Go 的 GC 算法？(三色标记 + 混合写屏障)
// 2. STW 是什么？Go 如何减少 STW 时间？
// 3. GOGC 的作用？如何调优？
// 4. GC 触发的条件？
// 5. 如何观察 GC 行为？
// 6. Go 1.19+ 的 GOMEMLIMIT？

func main() {
	algorithm()
	triggerConditions()
	barrier()
	tuning()
	observation()
	memoryLimit()
}

// ----------------------------------------------------------
// 1. GC 算法演进
// ----------------------------------------------------------
// Go 1.0:  STW 的标记-清除
// Go 1.1:  并行标记-清除（标记阶段并行，仍然 STW）
// Go 1.5:  三色标记 + 写屏障（并发标记，STW 极短）
// Go 1.8:  混合写屏障（STW < 100µs）
//
// 当前算法: 并发三色标记-清除（Concurrent Tri-color Mark & Sweep）
//
// 三色标记:
//   白色: 未被标记的对象（GC 结束后回收）
//   灰色: 已标记但引用未扫描的对象
//   黑色: 已标记且引用已扫描的对象
//
// 标记过程:
//   1. 初始所有对象为白色
//   2. 根对象（栈变量、全局变量等）标记为灰色
//   3. 取出灰色对象，扫描其引用的对象，标记为灰色
//   4. 当前灰色对象变为黑色
//   5. 重复 3-4 直到没有灰色对象
//   6. 白色对象即为垃圾，回收

func algorithm() {
	fmt.Println("=== 1. 三色标记算法 ===")
	fmt.Println("白色: 未标记（将被回收）")
	fmt.Println("灰色: 已标记但引用未扫描")
	fmt.Println("黑色: 已标记且引用已扫描")
	fmt.Println()
	fmt.Println("标记流程:")
	fmt.Println("  1. 根对象 → 灰色")
	fmt.Println("  2. 取灰色对象，扫描引用 → 引用变灰色")
	fmt.Println("  3. 当前对象 → 黑色")
	fmt.Println("  4. 重复直到无灰色对象")
	fmt.Println("  5. 白色对象 = 垃圾，回收")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. GC 触发条件
// ----------------------------------------------------------
// 三种触发方式:
//
// 1. 堆内存增长触发（最常见）
//    当堆内存达到上一次 GC 后存活堆内存的 1 + GOGC/100 倍时触发
//    默认 GOGC=100，即堆内存翻倍时触发
//    例: 上次 GC 后存活 100MB，堆增长到 200MB 时触发
//
// 2. 定时触发
//    runtime/mproc.go: forcegcperiod = 2 分钟
//    如果 2 分钟没有触发 GC，sysmon 会强制触发
//
// 3. 手动触发
//    runtime.GC() — 阻塞直到 GC 完成

func triggerConditions() {
	fmt.Println("=== 2. GC 触发条件 ===")

	fmt.Println("1. 堆内存达到 (1 + GOGC/100) × 上次存活堆内存")
	fmt.Println("2. 距上次 GC 超过 2 分钟")
	fmt.Println("3. 手动调用 runtime.GC()")
	fmt.Println()

	// 观察当前 GC 状态
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("GC 次数: %d\n", stats.NumGC)
	if stats.NumGC > 0 {
		fmt.Printf("最近 GC: %v 前\n", time.Since(stats.LastGC))
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 写屏障（Write Barrier）
// ----------------------------------------------------------
// 为什么需要写屏障？
//   三色标记和用户代码并发执行，可能出现:
//   - 用户代码修改了黑色对象的引用，指向白色对象
//   - 这个白色对象不会被标记，但它是活跃的
//   → 对象丢失（遗漏标记）
//
// 两种写屏障:
//
// Dijkstra 插入写屏障（Go 1.5-1.7）:
//   A.field = B 时，把 B 标记为灰色
//   保证: 黑色对象不会直接指向白色对象
//   问题: 栈上写操作也需要写屏障（开销大），所以 Go 对栈无写屏障
//         需要 STW 重新扫描栈（~10-100ms）
//
// Yuasa 删除写屏障:
//   A.field = B 时，把 A.field 原来的值标记为灰色
//   保证: 被删除引用的白色对象不会被遗漏
//
// Go 1.8+ 混合写屏障:
//   = Dijkstra + Yuasa
//   1. A.field = B 时，把 B 标记为灰色（Dijkstra）
//   2. 如果 A 在栈上，还要把 A.field 原来的值标记为灰色（Yuasa）
//   好处: 不需要 STW 重新扫描栈！STW 时间降到 < 100µs

func barrier() {
	fmt.Println("=== 3. 写屏障 ===")
	fmt.Println("Go 1.5-1.7: Dijkstra 插入写屏障")
	fmt.Println("  → 栈需要 STW 重新扫描（10-100ms）")
	fmt.Println("Go 1.8+: 混合写屏障（Dijkstra + Yuasa）")
	fmt.Println("  → 不需要重新扫描栈，STW < 100µs")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. GOGC 调优
// ----------------------------------------------------------
// GOGC 环境变量（默认值 100）:
//   GC 触发阈值 = 上次 GC 存活堆内存 × (1 + GOGC/100)
//
//   GOGC=100（默认）: 堆翻倍时 GC，内存使用 ~2x 存活数据
//   GOGC=200:        堆增长 3 倍时 GC，更少 GC 但更多内存
//   GOGC=50:         堆增长 1.5 倍时 GC，更多 GC 但更少内存
//   GOGC=off:        禁用 GC
//
// 调优原则:
//   - CPU 密集型 → 增大 GOGC（减少 GC 次数）
//   - 内存敏感型 → 减小 GOGC（减少内存使用）
//   - 也可以用 debug.SetGCPercent() 动态调整

func tuning() {
	fmt.Println("=== 4. GOGC 调优 ===")

	// 当前 GOGC
	fmt.Printf("当前 GOGC: %d%%\n", debug.SetGCPercent(100))
	fmt.Println()
	fmt.Println("GOGC=100: 堆翻倍时 GC（默认，平衡）")
	fmt.Println("GOGC=200: 更少 GC，更多内存")
	fmt.Println("GOGC=50:  更多 GC，更少内存")
	fmt.Println("GOGC=off: 禁用 GC（不推荐）")

	// 手动触发 GC
	fmt.Println("\n手动触发 GC:")
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	fmt.Printf("  GC 后堆大小: %.2f MB\n", float64(m1.HeapAlloc)/1024/1024)

	// 分配一些内存
	_ = make([]byte, 10<<20) // 10MB
	runtime.GC()
	runtime.ReadMemStats(&m2)
	fmt.Printf("  分配 10MB 并 GC 后: %.2f MB\n", float64(m2.HeapAlloc)/1024/1024)

	fmt.Println()
}

// ----------------------------------------------------------
// 5. 观察 GC
// ----------------------------------------------------------
func observation() {
	fmt.Println("=== 5. 观察 GC ===")

	// 方式1: runtime.ReadMemStats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("堆分配: %.2f MB\n", float64(m.HeapAlloc)/1024/1024)
	fmt.Printf("系统分配: %.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Printf("GC 次数: %d\n", m.NumGC)
	fmt.Printf("GC 总暂停: %v\n", time.Duration(m.PauseTotalNs))
	fmt.Printf("最近 GC 暂停: %v\n", time.Duration(m.PauseNs[(m.NumGC+255)%256]))

	// 方式2: debug.ReadGCStats
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("GC 次数: %d\n", stats.NumGC)

	// 方式3: 环境变量 GODEBUG=gctrace=1
	// 运行时会打印每次 GC 的详细信息:
	//   gc 1 @0.003s 0%: 0.018+0.45+0.003 ms clock, 0.14+0.21/0.39/0.043+0.024 ms cpu, 4->4->0 MB, 5 MB goal, 8 P
	// 含义:
	//   gc 1: 第 1 次 GC
	//   @0.003s: 程序启动后 0.003s
	//   0%: GC 占用 CPU 百分比
	//   0.018+0.45+0.003 ms: STW(标记开始)+并发标记+STW(标记结束)
	//   4->4->0 MB: GC前堆大小→GC后堆大小→存活堆大小
	//   5 MB goal: 下次 GC 触发阈值
	//   8 P: P 的数量

	fmt.Println("\nGODEBUG=gctrace=1 可打印每次 GC 详情")
	fmt.Println("pprof: go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println()
}

// ----------------------------------------------------------
// 6. GOMEMLIMIT (Go 1.19+)
// ----------------------------------------------------------
// GOGC 的问题: 只控制 GC 频率，不限制堆的最大大小
//   如果存活数据很多（比如大缓存），GC 可能不会及时触发
//
// GOMEMLIMIT: 设置运行时的整体内存软上限
//   - 当运行时总内存接近限制时，更积极地触发 GC
//   - 可以和 GOGC 配合使用
//   - debug.SetMemoryLimit() 动态设置
//
// 推荐:
//   GOGC=off + GOMEMLIMIT=合理值 → 固定内存上限（适合容器环境）
//   GOGC=默认 + GOMEMLIMIT=容器限制 → 平衡 GC 和内存

func memoryLimit() {
	fmt.Println("=== 6. GOMEMLIMIT (Go 1.19+) ===")

	// 设置内存限制
	limit := debug.SetMemoryLimit(1 << 30) // 1GB
	fmt.Printf("内存限制: %d bytes (%.0f MB)\n", limit, float64(limit)/1024/1024)

	fmt.Println("GOMEMLIMIT 优势:")
	fmt.Println("  - 限制运行时总内存（不只是堆）")
	fmt.Println("  - 容器环境中防止 OOM kill")
	fmt.Println("  - 可与 GOGC 配合使用")
	fmt.Println()
	fmt.Println("推荐组合:")
	fmt.Println("  容器限制 512MB → GOMEMLIMIT=450MB (留余量)")
	fmt.Println("  GOGC=默认 → 正常 GC 频率")
	fmt.Println()
}
