---
title: Profiling 性能分析
---

## 1. CPU Profiling

原理：操作系统每秒发送 100 次 SIGPROF 信号（100Hz），每次记录所有线程的调用栈。采样结束后统计每个函数出现在调用栈中的次数，出现越多 = 占用 CPU 越多。

**使用方式**：pprof.StartCPUProfile(f) 开始，pprof.StopCPUProfile() 结束。

**分析命令**：top20 查看热点函数，web 生成火焰图（需要 graphviz），list funcName 查看行级耗时。

::: tip 使用场景
接口 RT 突然变慢 → CPU profile 找热点函数。

```go
f, _ := os.Create("cpu.prof")
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
// go tool pprof -http=:8080 cpu.prof
```

## 2. Memory Profiling

两种内存 profile：**allocs**（累计分配统计）和 **heap**（当前存活对象的分配统计）。用 pprof.WriteHeapProfile 写入。
:::

**分析维度**：-inuse_space（正在使用的内存）、-inuse_objects（正在使用的对象数）、-alloc_space（累计分配量）、-alloc_objects（累计分配次数）。

::: tip 使用场景
内存持续增长 → 对比两次 heap profile 找泄漏点。

::: warning 面试追问
allocs 和 heap 的区别？→ allocs 是累计值（从程序启动），heap 是当前值。

```go
f, _ := os.Create("mem.prof")
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
// 2. 看 inuse_space 增长的函数
```

## 3. 在线 Profiling

**net/http/pprof** 自动注册 /debug/pprof/ 路由，无需额外代码（import _ "net/http/pprof"）。支持 CPU、heap、goroutine、thread、block、mutex 等 profile。
:::

**goroutine 泄漏排查**：查看 goroutine profile 中阻塞的调用栈，找到泄漏的 goroutine。

::: tip 生产建议
只在内部端口暴露 pprof，不要对外暴露。

::: warning 面试追问
block 和 mutex profile 需要额外设置？→ 是，需要 runtime.SetBlockProfileRate 和 runtime.SetMutexProfileFraction。

```go
import _ "net/http/pprof"
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
}()
```
