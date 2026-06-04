---
title: IO 模型
---

## 1. 五种 IO 模型

UNIX 环境下有五种经典 IO 模型，以读操作为例：

| 模型 | 用户态 | 内核态 | 特点 |
|------|--------|--------|------|
| 阻塞 IO | 阻塞等待 | 等待数据 + 拷贝 | 最简单，效率低 |
| 非阻塞 IO | 轮询 | 等待数据 + 拷贝 | CPU 空转浪费 |
| IO 多路复用 | 阻塞在 select | 等待数据 + 拷贝 | 一个线程管多个 fd |
| 信号驱动 IO | 不阻塞 | 等待数据完成发信号 | 实现复杂 |
| 异步 IO (AIO) | 不阻塞 | 内核完成一切 | 真正的异步 |

::: info 同步 vs 异步的本质区别
同步 IO：用户线程参与数据从内核态到用户态的拷贝过程。前四种都是同步 IO。异步 IO：内核完成数据拷贝后再通知用户线程（如 Linux `io_submit`，Windows IOCP）。
:::

### 阻塞 IO 示例

```go
package main

import (
    "net"
    "time"
)

func handleConn(conn net.Conn) {
    buf := make([]byte, 1024)
    // Read 会阻塞直到有数据到达
    n, err := conn.Read(buf)
    if err != nil {
        return
    }
    conn.Write(buf[:n])
}

func main() {
    ln, _ := net.Listen("tcp", ":8080")
    for {
        conn, _ := ln.Accept() // 阻塞等待新连接
        go handleConn(conn)     // 为每个连接启动 goroutine
    }
}
```

## 2. IO 多路复用详解

### select

```c
// select 的限制
#define FD_SETSIZE 1024  // 最多监听 1024 个 fd
int select(int nfds, fd_set *readfds, fd_set *writefds,
           fd_set *exceptfds, struct timeval *timeout);
```

- 每次调用需要将 fd 集合从用户态拷贝到内核态
- 内核线性扫描所有 fd，O(n) 复杂度
- 返回后需要遍历所有 fd 找出就绪的

### poll

```c
struct pollfd {
    int fd;         // 文件描述符
    short events;   // 关注的事件
    short revents;  // 返回的事件
};
int poll(struct pollfd *fds, nfds_t nfds, int timeout);
```

- 没有 1024 个 fd 的数量限制
- 使用结构体数组，比 fd_set 更灵活
- 仍然是 O(n) 扫描，大量 fd 时性能下降

### epoll

```c
// 三个核心 API
int epoll_create(int size);              // 创建 epoll 实例
int epoll_ctl(int epfd, int op, ...);    // 注册/修改/删除 fd
int epoll_wait(int epfd, ...);           // 等待事件就绪
```

epoll 使用红黑树管理 fd，注册只操作一次；通过回调机制通知就绪事件，时间复杂度 O(1)。

::: tip epoll 为什么高效
1. **红黑树**：fd 注册/删除 O(log n)，而非每次全量传入
2. **事件回调**：网卡收到数据触发中断，内核将 fd 加入就绪队列，而非轮询
3. **共享内存**：mmap 共享就绪事件，减少内核态-用户态拷贝
:::

### ET (边缘触发) vs LT (水平触发)

| 模式 | 触发条件 | 编程要求 |
|------|----------|----------|
| LT (Level Triggered) | 缓冲区有数据就持续通知 | 简单，可以不一次读完 |
| ET (Edge Triggered) | 只在新数据到达时通知一次 | 必须非阻塞 + 循环读直到 EAGAIN |

```go
// 模拟 ET 模式的处理逻辑
func handleET(conn net.Conn) {
    buf := make([]byte, 1024)
    for {
        n, err := conn.Read(buf)
        if n > 0 {
            process(buf[:n])
        }
        if err != nil {
            break // EAGAIN 或连接关闭
        }
    }
}
```

## 3. Go netpoller

### Go 网络模型的封装

Go 运行时在内部封装了 epoll（Linux）/ kqueue（macOS），提供统一的非阻塞 IO 接口：

1. `net.Listen` / `net.Dial` 创建的 fd 被设置为**非阻塞模式**
2. 将 fd 注册到 runtime 的 poller（epoll 实例）
3. Goroutine 调用 `Read`/`Write` 时，如果没有数据就 `gopark` 挂起
4. epoll 返回就绪事件后，唤醒对应的 Goroutine 继续执行

```go
package main

import (
    "fmt"
    "net"
    "net/http"
)

func main() {
    // Go 的 HTTP 服务器自动使用 netpoller
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "hello from netpoller")
    })

    // 底层流程：
    // 1. net.Listen("tcp", ":8080") → 非阻塞 socket + epoll 注册
    // 2. Accept → 无连接时 gopark，epoll 通知时 goready
    // 3. conn.Read → 无数据时 gopark，数据到达时 goready
    http.ListenAndServe(":8080", nil)
}
```

::: warning net.Fd() 的隐患
调用 `conn.(*net.TCPConn).File()` 或 `conn.Fd()` 会触发 dup 系统调用，产生新的阻塞 fd。此 fd 不再被 netpoller 管理，必须手动设置为非阻塞并自行管理生命周期。
:::

```go
// 错误示范：Fd() 导致连接泄漏
func badPattern(conn net.Conn) {
    fd, _ := conn.(*net.TCPConn).File()
    // 此时 conn 和 fd 是两个不同的文件描述符
    // conn 关闭不会关闭 fd，fd 关闭不影响 conn
    defer fd.Close() // 必须手动关闭
}
```

## 4. 零拷贝技术

### 传统数据传输的 4 次拷贝

```
磁盘 → 内核页缓存 → 用户缓冲区 → Socket 缓冲区 → 网卡
       DMA拷贝      CPU拷贝        CPU拷贝        DMA拷贝
```

### sendfile

```go
package main

import (
    "net"
    "os"
    "syscall"
)

// 使用 sendfile 零拷贝发送文件
func sendFile(conn net.Conn, filePath string) error {
    f, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer f.Close()

    fi, _ := f.Stat()
    tcpConn, ok := conn.(*net.TCPConn)
    if !ok {
        return syscall.EINVAL
    }

    ffd, _ := f.SyscallConn()
    cfd, _ := tcpConn.SyscallConn()

    return cfd.Write(func(cfd uintptr) bool {
        ffd.Control(func(ffd uintptr) {
            syscall.Sendfile(int(cfd), int(ffd), nil, int(fi.Size()))
        })
        return false
    })
}
```

### splice

splice 在两个 fd 之间移动数据，完全不经过用户态：

```bash
# splice 示意图：磁盘文件 → 管道 → Socket
# 数据只在内核态流转，零用户态拷贝
```

::: info Go 中的零拷贝实践
- `io.Copy` 内部会自动选择最优策略（sendfile 等）
- `net/http` 的 `ServeContent`/`ServeFile` 自动使用 sendfile
- `github.com/cloudwego/netpoll` 提供更高效的零拷贝 API
:::

## 5. 面试常见问题

### epoll 为什么高效

1. **事件驱动**：基于回调通知，而非轮询扫描
2. **O(1) 就绪检测**：只需检查就绪链表是否为空
3. **一次注册，多次使用**：fd 只需 `epoll_ctl` 注册一次
4. **无 fd 数量限制**：取决于 `/proc/sys/fs/file-max`

### Go 网络模型 vs Java NIO

| 维度 | Go netpoller | Java NIO |
|------|-------------|----------|
| 编程模型 | 同步阻塞（对开发者透明） | Reactor 模式 + 非阻塞 |
| 事件循环 | 运行时内部管理 | 需手动实现 Selector Loop |
| 线程模型 | Goroutine per connection | 线程池 + Selector |
| 复杂度 | 低，直接写同步代码 | 高，需处理 SelectionKey |

### 面试追问

**Q1: Reactor 模式是什么？Go 用了 Reactor 吗？**

Reactor 模式由 Reactor 线程监听事件，分发给 Handler 处理。分为单 Reactor、多 Reactor（主从）等变体。Go 的 netpoller 并非经典 Reactor：它把"事件分发"换成了"唤醒 Goroutine"，每个连接对应一个独立的 Goroutine，开发者看到的仍然是同步阻塞的编程模型。

**Q2: 为什么 Go 不用 Linux AIO？**

Linux AIO (io_submit) 有诸多限制：只支持 O_DIRECT、回调模型与 Go 的阻塞 IO 语义不符。Go 1.22+ 开始实验性地支持 `poll()` based IO for files，未来可能引入 `io_uring`。目前 Go 的文件 IO 仍然是阻塞系统调用 + 线程池（runtime 有默认的 sysmon 机制）。

**Q3: 如何实现一个高性能的 TCP 服务器？**

关键点：(1) 设置 `SO_REUSEPORT` 多进程/多 listener 监听同一端口；(2) 调整 `GOMAXPROCS`；(3) 使用 `conn.SetReadBuffer/SetWriteBuffer` 调整内核缓冲区；(4) 连接复用（连接池）；(5) 避免频繁内存分配（`sync.Pool`）。

**Q4: select、poll、epoll 在 C10K 问题上的表现？**

select 受 `FD_SETSIZE` 限制只能管理 1024 个连接；poll 没有数量限制但 O(n) 扫描在万级连接时性能急剧下降；epoll 使用红黑树 + 事件回调，即使 10 万连接也能高效处理，是 C10K 问题的标准解决方案。
