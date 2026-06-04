package main

import (
	"fmt"
)

// ============================================================
// IO 模型
// ============================================================

func main() {
	fiveIOModels()
	ioMultiplexing()
	goNetpoller()
	zeroCopy()
	reactorPattern()
}

// ----------------------------------------------------------
// 1. 五种 IO 模型
// ----------------------------------------------------------
func fiveIOModels() {
	fmt.Println("=== 1. 五种 IO 模型 ===")
	fmt.Println()
	fmt.Println("  1. 阻塞 IO (Blocking IO)")
	fmt.Println("     线程调用 read() → 阻塞直到数据就绪")
	fmt.Println("     Go: net.Listen.Accept() 默认阻塞 (但 goroutine 很轻)")
	fmt.Println()
	fmt.Println("  2. 非阻塞 IO (Non-blocking IO)")
	fmt.Println("     调用 read() → 立即返回 EWOULDBLOCK")
	fmt.Println("     需要轮询检查 (浪费 CPU)")
	fmt.Println()
	fmt.Println("  3. IO 多路复用 (IO Multiplexing)")
	fmt.Println("     select/poll/epoll 监听多个 fd")
	fmt.Println("     有 fd 就绪时返回 → 处理")
	fmt.Println("     Go netpoller 底层使用 epoll/kqueue")
	fmt.Println()
	fmt.Println("  4. 信号驱动 IO (Signal-driven IO)")
	fmt.Println("     fd 就绪时内核发送 SIGIO 信号")
	fmt.Println("     实际使用较少")
	fmt.Println()
	fmt.Println("  5. 异步 IO (AIO)")
	fmt.Println("     内核完成整个操作后通知应用")
	fmt.Println("     Linux AIO 不完善, Windows IOCP 成熟")
	fmt.Println()
	fmt.Println("  Go 的选择:")
	fmt.Println("     netpoller (epoll/kqueue) + goroutine-per-conn")
	fmt.Println("     对开发者呈现阻塞模型，底层使用非阻塞 IO + epoll")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. IO 多路复用详解
// ----------------------------------------------------------
func ioMultiplexing() {
	fmt.Println("=== 2. IO 多路复用 ===")
	fmt.Println()
	fmt.Println("  select:")
	fmt.Println("    监听 fd 数量: 1024 (FD_SETSIZE)")
	fmt.Println("    每次调用需要传入全部 fd → 内核遍历 → O(n)")
	fmt.Println("    返回后需要遍历所有 fd 找就绪的 → O(n)")
	fmt.Println()
	fmt.Println("  poll:")
	fmt.Println("    无数量限制 (链表)")
	fmt.Println("    但仍然是 O(n) 遍历")
	fmt.Println()
	fmt.Println("  epoll:")
	fmt.Println("    1. epoll_create() — 创建 epoll 实例")
	fmt.Println("    2. epoll_ctl()    — 添加/修改/删除 fd (红黑树)")
	fmt.Println("    3. epoll_wait()   — 等待就绪事件 (只返回就绪 fd)")
	fmt.Println()
	fmt.Println("    优点:")
	fmt.Println("      fd 数量无限制 (红黑树管理)")
	fmt.Println("      只返回就绪的 fd → O(就绪fd数)")
	fmt.Println("      注册一次，持续监听 (不需要每次传入)")
	fmt.Println()
	fmt.Println("    两种模式:")
	fmt.Println("      LT (Level Triggered, 默认):")
	fmt.Println("        fd 就绪 → 通知 → 不处理 → 下次还通知")
	fmt.Println("      ET (Edge Triggered):")
	fmt.Println("        fd 状态变化时通知一次 → 必须一次读完所有数据")
	fmt.Println("        更高效但编程更复杂")
	fmt.Println()
	fmt.Println("  ┌──────────┬──────────┬──────────┬──────────┐")
	fmt.Println("  │ 维度     │ select   │ poll     │ epoll    │")
	fmt.Println("  ├──────────┼──────────┼──────────┼──────────┤")
	fmt.Println("  │ fd 数量  │ 1024     │ 无限制   │ 无限制   │")
	fmt.Println("  │ 复杂度   │ O(n)     │ O(n)     │ O(1)*    │")
	fmt.Println("  │ 数据结构 │ bitmap   │ 链表     │ 红黑树   │")
	fmt.Println("  │ 触发模式 │ LT       │ LT       │ LT/ET    │")
	fmt.Println("  └──────────┴──────────┴──────────┴──────────┘")
	fmt.Println("  * O(1) 指 epoll_wait 的就绪通知，不含回调处理")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Go netpoller
// ----------------------------------------------------------
func goNetpoller() {
	fmt.Println("=== 3. Go netpoller ===")
	fmt.Println()
	fmt.Println("  架构:")
	fmt.Println("    goroutine → net.Conn.Read() → gopark (挂起)")
	fmt.Println("    epoll 监听 fd 就绪 → netpoll → goready (唤醒 goroutine)")
	fmt.Println()
	fmt.Println("  工作流程:")
	fmt.Println("    1. net.Listen() → 创建非阻塞 socket → epoll_ctl(ADD)")
	fmt.Println("    2. Accept() → 非阻塞 accept → 无连接时 gopark")
	fmt.Println("    3. 有新连接 → epoll 触发 → goroutine 被唤醒")
	fmt.Println("    4. conn.Read() → 非阻塞 read → 无数据时 gopark")
	fmt.Println("    5. 有数据 → epoll 触发 → goroutine 被唤醒")
	fmt.Println()
	fmt.Println("  netpoll 的调用时机:")
	fmt.Println("    在 schedule() 中定期调用 netpoll()")
	fmt.Println("    检查是否有就绪的网络 fd → 唤醒对应 goroutine")
	fmt.Println()
	fmt.Println("  net/Fd() 的隐患:")
	fmt.Println("    conn.File() 会将 fd 设为阻塞模式!")
	fmt.Println("    因为返回的是原始 os.File，不受 netpoller 管理")
	fmt.Println("    可能阻塞 M 线程，影响调度")
	fmt.Println("    结论: 不要在生产中使用 conn.File()")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 零拷贝技术
// ----------------------------------------------------------
func zeroCopy() {
	fmt.Println("=== 4. 零拷贝技术 ===")
	fmt.Println()
	fmt.Println("  传统文件传输 (4 次拷贝):")
	fmt.Println("    disk → kernel buffer → user buffer → kernel socket buffer → nic")
	fmt.Println()
	fmt.Println("  sendfile (2 次拷贝):")
	fmt.Println("    disk → kernel buffer → nic (通过 DMA)")
	fmt.Println("    Go: syscall.Sendfile(outFd, inFd, offset, count)")
	fmt.Println()
	fmt.Println("  mmap:")
	fmt.Println("    文件映射到内存 → 直接读写")
	fmt.Println("    1 次拷贝: disk → page cache → CPU 直接访问")
	fmt.Println()
	fmt.Println("  splice:")
	fmt.Println("    管道 + sendfile → 零拷贝转发")
	fmt.Println("    适用于代理/网关场景")
	fmt.Println()
	fmt.Println("  Go 中的零拷贝:")
	fmt.Println("    io.Copy → 内部检测是否可用 sendfile")
	fmt.Println("    http.ServeContent → 自动使用 sendfile")
	fmt.Println()
	fmt.Println("  面试追问: 为什么零拷贝重要?")
	fmt.Println("    减少用户态/内核态数据拷贝 → 降低 CPU 占用")
	fmt.Println("    在高吞吐网络场景 (CDN/代理) 中尤其重要")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. Reactor 模式
// ----------------------------------------------------------
func reactorPattern() {
	fmt.Println("=== 5. Reactor 模式 ===")
	fmt.Println()
	fmt.Println("  单 Reactor 单线程:")
	fmt.Println("    accept/read/write/业务处理 都在一个线程")
	fmt.Println("    Redis 使用这种模式 (Redis 6.0 前单线程)")
	fmt.Println()
	fmt.Println("  单 Reactor 多线程:")
	fmt.Println("    一个线程处理 accept/read/write")
	fmt.Println("    业务处理交给线程池")
	fmt.Println()
	fmt.Println("  主从 Reactor 多线程 (Netty/Go netpoller):")
	fmt.Println("    Main Reactor: accept 新连接")
	fmt.Println("    Sub Reactor: 处理已建立连接的 read/write")
	fmt.Println("    Worker Pool: 业务处理")
	fmt.Println()
	fmt.Println("  Go 模型 vs Reactor:")
	fmt.Println("    Go 不是严格意义上的 Reactor")
	fmt.Println("    但 goroutine-per-conn + netpoller 效果等价于:")
	fmt.Println("    - 每个 goroutine 是一个独立的 handler")
	fmt.Println("    - netpoller 是 Reactor (epoll)")
	fmt.Println("    - M:N 调度器替代了线程池")
	fmt.Println()
	fmt.Println("  面试追问: Go 网络模型 vs Java NIO?")
	fmt.Println("    Java NIO: Reactor 模式 + 线程池, 需要手动管理")
	fmt.Println("    Go: goroutine-per-conn, 自动管理, 开发体验更好")
	fmt.Println("    性能: 两者都基于 epoll, 差距不大")
	fmt.Println()
}
