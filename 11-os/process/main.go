package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// ============================================================
// 进程与线程
// ============================================================

func main() {
	processVsThread()
	goRoutineModel()
	ipcDemo()
	processManagement()
	signalDemo()
}

// ----------------------------------------------------------
// 1. 进程/线程/协程对比
// ----------------------------------------------------------
func processVsThread() {
	fmt.Println("=== 1. 进程/线程/协程对比 ===")
	fmt.Println()
	fmt.Println("  ┌──────────┬──────────────┬──────────────┬──────────────┐")
	fmt.Println("  │ 维度     │ 进程         │ 线程         │ Goroutine    │")
	fmt.Println("  ├──────────┼──────────────┼──────────────┼──────────────┤")
	fmt.Println("  │ 调度     │ OS 内核      │ OS 内核      │ Go 运行时    │")
	fmt.Println("  │ 切换开销 │ 大 (~μs)     │ 中           │ 小 (~ns)     │")
	fmt.Println("  │ 内存     │ 独立地址空间 │ 共享地址空间 │ 共享地址空间 │")
	fmt.Println("  │ 栈大小   │ MB 级        │ MB 级(固定)  │ 2KB(可扩缩)  │")
	fmt.Println("  │ 创建成本 │ 高 (fork)    │ 中           │ 极低         │")
	fmt.Println("  │ 通信     │ IPC          │ 共享内存     │ Channel      │")
	fmt.Println("  └──────────┴──────────────┴──────────────┴──────────────┘")
	fmt.Println()
	fmt.Println("  面试追问: Goroutine 为什么轻量?")
	fmt.Println("  答: 1. 用户态调度，不需要内核态切换")
	fmt.Println("      2. 栈从 2KB 起始，按需增长 (连续栈)")
	fmt.Println("      3. 创建只需分配栈空间和 G 结构体")
	fmt.Println("      4. GMP 模型高效调度")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. GMP 调度模型
// ----------------------------------------------------------
func goRoutineModel() {
	fmt.Println("=== 2. GMP 调度模型 ===")
	fmt.Println()
	fmt.Println("  G — Goroutine (用户协程)")
	fmt.Println("  M — Machine (OS 线程)")
	fmt.Println("  P — Processor (逻辑处理器，GOMAXPROCS 个)")
	fmt.Println()
	fmt.Println("  调度流程:")
	fmt.Println("    1. P 的本地队列有 G → M 绑定 P 执行 G")
	fmt.Println("    2. 本地队列空 → 从全局队列获取")
	fmt.Println("    3. 全局也空 → 从其他 P 偷一半 (work stealing)")
	fmt.Println()
	fmt.Println("  关键机制:")
	fmt.Println("    - Hand Off: G 进行系统调用时, M 让出 P 给其他 M")
	fmt.Println("    - Work Stealing: 空闲 P 从其他 P 的本地队列偷 G")
	fmt.Println("    - 抢占式调度: 基于信号的异步抢占 (Go 1.14+)")
	fmt.Println()
	fmt.Println("  GOMAXPROCS:")
	fmt.Println("    默认 = CPU 核心数")
	fmt.Println("    runtime.GOMAXPROCS(n) 设置 P 的数量")
	fmt.Println("    P 的数量决定了并行度 (同时执行的 Goroutine 数)")
	fmt.Println()
	fmt.Println("  Goroutine 栈管理:")
	fmt.Println("    初始: 2KB")
	fmt.Println("    增长: 翻倍扩容 (2KB→4KB→8KB→...→1GB)")
	fmt.Println("    缩容: GC 时检测，收缩到实际使用大小")
	fmt.Println("    实现: 连续栈 (contiguous stack)")
	fmt.Println("      扩容: 分配新栈 → 复制数据 → 调整指针")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 进程间通信
// ----------------------------------------------------------
func ipcDemo() {
	fmt.Println("=== 3. 进程间通信 ===")
	fmt.Println()
	fmt.Println("  1. 管道 (Pipe):")
	fmt.Println("     父子进程通信, 半双工")
	fmt.Println("     Go: os.Pipe()")
	fmt.Println()
	fmt.Println("  2. 信号 (Signal):")
	fmt.Println("     异步通知进程")
	fmt.Println("     Go: os/signal.Notify()")
	fmt.Println()
	fmt.Println("  3. 共享内存 + 锁:")
	fmt.Println("     Go: syscall.Mmap() + sync.Mutex")
	fmt.Println()
	fmt.Println("  4. Unix Domain Socket:")
	fmt.Println("     同机器进程通信, 比 TCP 快")
	fmt.Println("     Go: net.Listen(\"unix\", \"/tmp/app.sock\")")
	fmt.Println()
	fmt.Println("  5. 环境变量/命令行参数:")
	fmt.Println("     简单数据传递")
	fmt.Println("     Go: os.Getenv() / os.Args")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 进程管理
// ----------------------------------------------------------
func processManagement() {
	fmt.Println("=== 4. Go 进程管理 ===")

	// 启动子进程
	fmt.Println("  启动子进程:")
	fmt.Println("    cmd := exec.Command(\"ls\", \"-la\")")
	fmt.Println("    output, err := cmd.Output()")
	fmt.Println("    cmd.Run()  // 等待完成")
	fmt.Println("    cmd.Start() // 不等待")
	fmt.Println("    cmd.Wait()  // 等待完成")
	fmt.Println()

	// 获取进程信息
	fmt.Println("  进程信息:")
	fmt.Printf("    PID: %d\n", os.Getpid())
	fmt.Printf("    PPID: %d\n", os.Getppid())
	fmt.Printf("    UID: %d\n", os.Getuid())
	fmt.Println()

	// 演示 exec (实际不执行, 只展示)
	_ = exec.Command("echo", "hello")

	fmt.Println("  僵尸进程 (Zombie):")
	fmt.Println("    子进程退出但父进程未 Wait → 变为僵尸")
	fmt.Println("    解决: 父进程必须 Wait 或使用 SIGCHLD 处理")
	fmt.Println("    Go: cmd.Process.Wait() 回收子进程")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 信号处理
// ----------------------------------------------------------
func signalDemo() {
	fmt.Println("=== 5. 信号处理 ===")
	fmt.Println()
	fmt.Println("  常用信号:")
	fmt.Println("    SIGINT  (2)  — Ctrl+C 中断")
	fmt.Println("    SIGTERM (15) — 优雅终止 (kill 默认)")
	fmt.Println("    SIGKILL (9)  — 强制终止 (不可捕获)")
	fmt.Println("    SIGHUP  (1)  — 挂断 (常用于重载配置)")
	fmt.Println()
	fmt.Println("  Go 信号处理:")
	fmt.Println("    sigCh := make(chan os.Signal, 1)")
	fmt.Println("    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)")
	fmt.Println("    sig := <-sigCh")
	fmt.Println("    log.Printf(\"收到信号: %v\", sig)")
	fmt.Println()
	fmt.Println("  优雅关闭流程:")
	fmt.Println("    1. signal.Notify 监听 SIGTERM")
	fmt.Println("    2. 收到信号 → 开始优雅关闭")
	fmt.Println("    3. 停止接受新请求")
	fmt.Println("    4. 等待进行中请求完成 (超时强杀)")
	fmt.Println("    5. 关闭资源 (DB/Redis/MQ)")
	fmt.Println("    6. os.Exit(0)")
	fmt.Println()

	// 演示 (不实际等待)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("  收到 SIGTERM, 开始优雅关闭...")
	}()

	_ = time.Now // 避免 unused
}
