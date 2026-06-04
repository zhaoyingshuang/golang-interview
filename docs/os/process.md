---
title: 进程与线程
---

## 1. 进程、线程与协程对比

### 基本概念

| 维度 | 进程 (Process) | 线程 (Thread) | 协程 (Coroutine) |
|------|----------------|---------------|-------------------|
| 调度方式 | 内核调度 | 内核调度 | 用户态调度 |
| 上下文切换开销 | 大（需切换页表/TLB） | 中（共享地址空间） | 小（只保存寄存器） |
| 内存模型 | 独立地址空间 | 共享地址空间 | 共享地址空间 |
| 创建开销 | fork 开销大 | pthread_create 较大 | 极小（几 KB 栈） |
| 通信方式 | IPC（管道/共享内存等） | 共享内存 + 锁 | 直接共享变量 |

::: tip 核心区别
协程的关键在于**用户态调度**：切换不需要陷入内核态，由运行时自行决定何时让出执行权。Go 的 Goroutine 就是一种协程的实现。
:::

### Go 中的并发模型

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func main() {
    // 启动多个 Goroutine
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            fmt.Printf("goroutine %d running on thread %d\n",
                id, getThreadID())
        }(i)
    }
    wg.Wait()
    fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}

func getThreadID() int {
    // runtime 不直接暴露线程 ID，这里用 NumCPU 示意
    return runtime.NumCPU()
}
```

## 2. Go Goroutine 与 OS 线程

### GMP 调度模型

GMP 模型是 Go 运行时的核心调度设计：

- **G (Goroutine)**：轻量级用户态线程，初始栈仅 2KB
- **M (Machine)**：操作系统线程，由 OS 调度
- **P (Processor)**：逻辑处理器，数量默认等于 CPU 核心数（`GOMAXPROCS`）

调度流程：

1. 每个 P 持有一个本地 G 队列（local run queue，最多 256 个 G）
2. M 绑定 P 后，从 P 的本地队列取 G 执行
3. 本地队列为空时，从全局队列（global run queue）或其它 P 偷取（work stealing）
4. G 执行系统调用阻塞 M 时，P 解绑并交给其它空闲 M 继续运行

::: info M:N 调度
Go 使用 M:N 调度模型：M 个 Goroutine 映射到 N 个 OS 线程上。优势在于少量线程即可承载大量 Goroutine，减少内核调度开销。
:::

### Goroutine 栈管理

```go
package main

import "runtime"

func deepRecursion(n int) int {
    if n <= 0 {
        // 打印当前栈大小
        var buf [1 << 20]byte
        n := runtime.Stack(buf[:], false)
        println(string(buf[:n]))
        return 0
    }
    var pad [1024]byte // 每层分配 1KB
    _ = pad
    return deepRecursion(n - 1)
}

func main() {
    // Goroutine 初始栈 2KB，按需增长到 1GB
    go deepRecursion(10000)
}
```

::: warning 连续栈 (Contiguous Stack)
Go 1.4+ 使用连续栈替代分段栈。当栈空间不足时，分配一块翻倍大小的新栈并拷贝旧栈内容。这种方式避免了分段栈的"热分裂"(hot split) 问题。
:::

## 3. 进程间通信

### 管道 (Pipe)

```go
package main

import (
    "fmt"
    "io"
    "os"
    "os/exec"
)

func main() {
    // 创建管道连接父子进程
    cmd := exec.Command("grep", "hello")
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()

    cmd.Start()

    // 向子进程 stdin 写入数据
    go func() {
        io.WriteString(stdin, "hello world\ngoodbye\nhello go\n")
        stdin.Close()
    }()

    // 读取子进程 stdout
    buf := make([]byte, 1024)
    n, _ := stdout.Read(buf)
    fmt.Printf("filtered: %s", buf[:n])

    cmd.Wait()
}
```

### 信号 (Signal)

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    sigCh := make(chan os.Signal, 1)
    // 监听 SIGINT 和 SIGTERM
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        sig := <-sigCh
        fmt.Printf("received signal: %v, shutting down...\n", sig)
        os.Exit(0)
    }()

    select {} // 阻塞主线程
}
```

### Unix Domain Socket

```go
package main

import (
    "net"
    "os"
)

func server() {
    addr := "/tmp/unixdomain.sock"
    os.Remove(addr) // 清理残留文件
    l, _ := net.Listen("unix", addr)
    defer l.Close()

    conn, _ := l.Accept()
    buf := make([]byte, 512)
    n, _ := conn.Read(buf)
    println("server got:", string(buf[:n]))
}

func client() {
    conn, _ := net.Dial("unix", "/tmp/unixdomain.sock")
    conn.Write([]byte("hello from client"))
    conn.Close()
}
```

## 4. 进程管理

### os/exec 执行子进程

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "time"
)

func main() {
    // 带超时的命令执行
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "sleep", "10")
    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        fmt.Printf("command failed: %v, stderr: %s\n", err, stderr.String())
    }
}
```

### 子进程生命周期管理

```go
package main

import (
    "context"
    "log"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    cmd := exec.CommandContext(ctx, "./my-worker")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    // 确保子进程在父进程退出时被回收
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh
        log.Println("forwarding signal to child process")
        cmd.Process.Signal(syscall.SIGTERM)
        cancel()
    }()

    cmd.Wait()
}
```

::: warning 僵尸进程
子进程退出后，如果父进程未调用 `wait`，子进程会变成僵尸进程（Z 状态）。Go 的 `exec.Cmd.Wait()` 内部会调用 `wait4` 回收子进程。如果使用 `os.StartProcess` 而不调用 `Wait`，就会产生僵尸进程。
:::

## 5. 面试常见问题

### Goroutine 泄漏排查

```go
// 典型泄漏场景：channel 未关闭，goroutine 永远阻塞
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch // 永远不会收到数据
        println(val)
    }()
    // 函数返回后，ch 无人写入，goroutine 泄漏
}

// 修复方案：使用 context 超时控制
func fixed(ctx context.Context) {
    ch := make(chan int, 1)
    go func() {
        select {
        case val := <-ch:
            println(val)
        case <-ctx.Done():
            return // 超时退出
        }
    }()
}
```

排查工具：

```bash
# 使用 runtime/pprof 查看 goroutine 数量
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 使用 go tool trace 分析
go tool trace trace.out
```

### 面试追问

**Q1: Goroutine 和线程的区别是什么？为什么 Goroutine 更轻量？**

Goroutine 初始栈仅 2KB（线程通常 1-8MB），创建和切换在用户态完成（无需内核参与），且使用 M:N 调度模型，少量线程即可运行百万级 Goroutine。

**Q2: 什么是僵尸进程？如何避免？**

子进程退出后，其进程描述符保留在内核中直到父进程调用 `wait`。避免方法：(1) 父进程调用 `wait`/`waitpid`；(2) 父进程 `fork` 后立刻退出，让子进程被 `init` 接管（double fork 技巧）；(3) 捕获 `SIGCHLD` 信号并在处理函数中调用 `wait`。

**Q3: Go 的线程数有限制吗？**

Go 运行时默认最多创建 `GOMAXPROCS` 个活跃线程（即 P 的数量），但阻塞在系统调用中的 M 不受此限制。可通过 `runtime/debug.SetMaxThreads` 设置上限（默认 10000）。如果超过限制，程序会崩溃。

**Q4: 如何排查 Goroutine 泄漏？**

1. `runtime.NumGoroutine()` 监控数量趋势
2. `net/http/pprof` 的 `/debug/pprof/goroutine` 查看栈追踪
3. `go tool pprof` 分析 goroutine profile
4. 使用 `goleak` 测试框架在单元测试中检测泄漏
