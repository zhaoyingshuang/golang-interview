---
title: 并发模式
---

## 1. Worker Pool

控制并发数量，避免创建过多 goroutine。模式：jobs channel 分发任务，N 个 worker 通过 range jobs 消费，results channel 收集结果。

::: tip 使用场景
批量处理 API 请求、并发下载文件、并行数据处理。
:::

**为什么不用 go + WaitGroup**：Worker Pool 复用 goroutine，避免频繁创建销毁；通过 worker 数量精确控制并发。

::: warning 面试追问
worker 数量设多少？→ CPU 密集型 = CPU 核心数，I/O 密集型可以更多（10-100）。

```go
func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
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
for r := range results { /* 收集结果 */ }
```

## 2. Fan-out / Fan-in

**Fan-out**：一个 channel 分发给多个 goroutine 处理，提高吞吐。
:::

**Fan-in**：多个 channel 合并到一个，统一消费。

**和 Worker Pool 的区别**：Fan-out 的每个 worker 可以有不同的处理逻辑；Worker Pool 的 worker 相同。

::: tip 使用场景
日志处理（多种日志源合并）、数据聚合（从多个 API 获取数据后合并）。

::: warning 面试追问
merge 函数为什么要单独启动 goroutine 关闭 out？→ 要等所有 source channel 关闭后才能关闭 out。

```go
// Fan-out: 多个 worker 并行处理同一个输入
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
}
```

## 3. Pipeline

数据在多个 stage 之间流过，每个 stage 是一个 goroutine + channel。stage1 → stage2 → stage3 → 结果。每个 stage 通过 context 支持取消，任何 stage 出错都能优雅退出。

::: tip 使用场景
数据 ETL（读取 → 转换 → 写入）、图片处理（读取 → 缩放 → 水印 → 保存）、日志分析（读取 → 解析 → 过滤 → 聚合）。
:::

**优势**：每个 stage 可以独立伸缩（不同的 goroutine 数量），解耦清晰。

```go
// Pipeline: gen → square → filter → 结果
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
for v := range p { fmt.Println(v) }
```

## 4. 优雅退出 (Graceful Shutdown)

模式：context/信号控制生命周期 → 收到退出信号后停止接收新请求 → 等待进行中的请求完成 → 关闭资源。用 signal.Notify + context.WithCancel 实现信号监听，用 WaitGroup 或 errgroup 等待所有 goroutine 退出。

::: tip 使用场景
HTTP 服务收到 SIGTERM 后优雅退出（K8s Pod 滚动更新）。

::: warning 面试追问
为什么不直接 os.Exit？→ 进行中的请求会被截断，可能导致数据不一致。

```go
func main() {
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
}
```
