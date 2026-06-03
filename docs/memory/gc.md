---
title: GC 垃圾回收
---

## 1. 三色标记算法

Go 使用**并发三色标记-清除**算法。白色=未标记（将被回收），灰色=已标记但引用未扫描，黑色=已标记且引用已扫描。流程：1) 根对象→灰色 2) 取灰色扫描引用→引用变灰色 3) 当前→黑色 4) 重复直到无灰色 5) 白色=垃圾回收。

**根对象**包括：全局变量、当前所有 goroutine 的栈变量。

**三色不变式**：保证不会误删活跃对象。

::: info 生产影响
GC 和用户代码并发执行（大部分时间），只有标记开始和结束时极短暂的 STW。

```go
白色 (White)  - 未标记, 将被回收
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
5. 所有白色对象 = 垃圾, 回收内存
```

## 2. 混合写屏障

为什么需要写屏障？三色标记和用户代码并发执行时，可能出现黑色对象指向白色对象的引用（遗漏标记）。
:::

**Go 1.5-1.7**：Dijkstra 插入写屏障（A.field=B 时 B 标灰色），需要 STW 重新扫描栈（10-100ms）。

**Go 1.8+**：混合写屏障 = Dijkstra + Yuasa。A.field=B 时：1) B 标灰色（Dijkstra）；2) 如果 A 在栈上，原值也标灰色（Yuasa）。结果：**不需要 STW 重新扫描栈**，总 STW < 100µs。

::: info 生产影响
写屏障有 3-5% 的性能开销，但在 GC 期间才启用。

```go
// Go 1.8+ 混合写屏障
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
// 4. 并发清除: 回收白色对象
```

## 3. GC 触发条件与 GOGC 调优

三种触发：1) **堆内存增长**达到 (1+GOGC/100)×上次存活堆（最常见）；2) **2 分钟**未触发 GC（sysmon 强制）；3) **手动** runtime.GC()。
:::

**GOGC 调优**：默认 100（堆翻倍时 GC）；200 更少 GC 但更多内存；50 更多 GC 更少内存。

**Go 1.19+ GOMEMLIMIT**：设置运行时总内存软上限，比 GOGC 更直观。

::: tip 生产建议
容器环境用 GOMEMLIMIT（如容器限制 512MB → 设 450MB），让 Go 自己决定 GC 时机。

```go
// GOGC 调优
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
// GOGC=100 (默认即可)
```

## 4. 观察 GC

**GODEBUG=gctrace=1**：打印每次 GC 详情，包括 STW 时间、堆大小变化、GC 占 CPU 比。
:::

**runtime.ReadMemStats**：代码中获取详细内存统计。

**pprof**：heap profile 分析内存分配热点。

**生产监控**：Prometheus client_golang 自动暴露 GC 指标（go_gc_duration_seconds、go_memstats_alloc_bytes 等）。

::: warning 面试追问
如何减少 GC 压力？→ 减少堆分配（逃逸分析、sync.Pool、预分配）。

```go
// 环境变量方式（最快）
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
// go_goroutines
```
