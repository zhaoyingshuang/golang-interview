package main

import "fmt"

// ============================================================
// Linux 常用命令与排查
// ============================================================

func main() {
	processCommands()
	networkCommands()
	diskCommands()
	performanceAnalysis()
	troubleshooting()
}

// ----------------------------------------------------------
// 1. 进程管理命令
// ----------------------------------------------------------
func processCommands() {
	fmt.Println("=== 1. 进程管理命令 ===")
	fmt.Println()
	fmt.Println("  ps — 查看进程:")
	fmt.Println("    ps aux            — 所有进程详细信息")
	fmt.Println("    ps -ef            — 标准格式显示")
	fmt.Println("    ps -p <PID> -o pid,ppid,cmd  — 指定进程")
	fmt.Println()
	fmt.Println("  top/htop — 实时监控:")
	fmt.Println("    top -p <PID>      — 监控指定进程")
	fmt.Println("    htop              — 交互式 (需安装)")
	fmt.Println("    关键指标: %CPU, %MEM, RES, VIRT, S(状态)")
	fmt.Println()
	fmt.Println("  kill — 发送信号:")
	fmt.Println("    kill <PID>        — SIGTERM (优雅终止)")
	fmt.Println("    kill -9 <PID>     — SIGKILL (强制终止)")
	fmt.Println("    kill -HUP <PID>   — SIGHUP (重载配置)")
	fmt.Println()
	fmt.Println("  strace — 系统调用追踪:")
	fmt.Println("    strace -p <PID>              — 追踪运行中进程")
	fmt.Println("    strace -c -p <PID>           — 统计系统调用")
	fmt.Println("    strace -e trace=network -p <PID>  — 只追踪网络调用")
	fmt.Println("    Go 排查: 程序卡住时 strace 看卡在哪个 syscall")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 网络排查命令
// ----------------------------------------------------------
func networkCommands() {
	fmt.Println("=== 2. 网络排查命令 ===")
	fmt.Println()
	fmt.Println("  ss — Socket 统计 (替代 netstat):")
	fmt.Println("    ss -tlnp           — 查看监听端口 (TCP)")
	fmt.Println("    ss -tn             — 查看已建立连接")
	fmt.Println("    ss -s              — 连接统计摘要")
	fmt.Println("    关键状态: ESTAB/TIME_WAIT/CLOSE_WAIT/SYN_RECV")
	fmt.Println()
	fmt.Println("  tcpdump — 抓包:")
	fmt.Println("    tcpdump -i eth0 port 8080         — 抓指定端口")
	fmt.Println("    tcpdump -i eth0 host 10.0.0.1     — 抓指定 IP")
	fmt.Println("    tcpdump -w dump.pcap -i eth0      — 保存到文件")
	fmt.Println("    tcpdump -r dump.pcap              — 读取文件")
	fmt.Println("    tcpdump -i eth0 -s 0 -A port 80   — 显示 HTTP 内容")
	fmt.Println()
	fmt.Println("  nc — 网络测试:")
	fmt.Println("    nc -zv host port     — 测试端口连通性")
	fmt.Println("    nc -l 8080           — 启动 TCP 服务")
	fmt.Println("    nc host 8080         — 连接 TCP 服务")
	fmt.Println()
	fmt.Println("  curl — HTTP 请求:")
	fmt.Println("    curl -v URL          — 详细输出 (含请求/响应头)")
	fmt.Println("    curl -w '%{time_total}' -o /dev/null -s URL  — 测响应时间")
	fmt.Println("    curl -X POST -H 'Content-Type: application/json' -d '{}' URL")
	fmt.Println()
	fmt.Println("  连接状态分析:")
	fmt.Println("    TIME_WAIT 过多:")
	fmt.Println("      ss -ant state time-wait | wc -l")
	fmt.Println("      解决: 增大 tcp_max_tw_buckets / 开启 tw_reuse")
	fmt.Println()
	fmt.Println("    CLOSE_WAIT 过多 (程序 bug!):")
	fmt.Println("      ss -ant state close-wait | wc -l")
	fmt.Println("      原因: 服务端没有正确关闭连接")
	fmt.Println("      排查: 检查代码中 Response.Body 是否 Close()")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 磁盘与文件命令
// ----------------------------------------------------------
func diskCommands() {
	fmt.Println("=== 3. 磁盘与文件 ===")
	fmt.Println()
	fmt.Println("  df — 磁盘使用:")
	fmt.Println("    df -h              — 人类可读格式")
	fmt.Println("    df -i              — inode 使用情况")
	fmt.Println()
	fmt.Println("  du — 目录大小:")
	fmt.Println("    du -sh /path       — 目录总大小")
	fmt.Println("    du -sh /path/*     — 子目录大小")
	fmt.Println("    du -sh --max-depth=1 /path")
	fmt.Println()
	fmt.Println("  lsof — 打开的文件:")
	fmt.Println("    lsof -p <PID>      — 进程打开的所有文件")
	fmt.Println("    lsof -i :8080      — 使用 8080 端口的进程")
	fmt.Println("    lsof | wc -l       — 系统打开文件总数")
	fmt.Println("    lsof -p <PID> | wc -l  — 进程打开文件数")
	fmt.Println()
	fmt.Println("  文件句柄泄漏排查:")
	fmt.Println("    1. ls -l /proc/<PID>/fd | wc -l  — 进程 fd 数")
	fmt.Println("    2. watch -n 1 'ls /proc/<PID>/fd | wc -l'  — 持续监控")
	fmt.Println("    3. 如果持续增长 → 文件句柄泄漏")
	fmt.Println("    4. 代码检查: os.Open 后是否 defer Close()")
	fmt.Println()
	fmt.Println("  iostat — IO 统计:")
	fmt.Println("    iostat -x 1        — 每秒刷新, 详细模式")
	fmt.Println("    关注: %util (使用率), await (等待时间)")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 性能分析
// ----------------------------------------------------------
func performanceAnalysis() {
	fmt.Println("=== 4. 性能分析 ===")
	fmt.Println()
	fmt.Println("  Go pprof 工作流:")
	fmt.Println("    1. import _ \"net/http/pprof\"")
	fmt.Println("    2. go tool pprof http://localhost:6060/debug/pprof/profile")
	fmt.Println("    3. go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println("    4. (pprof) top 20     — 前 20 热点")
	fmt.Println("    5. (pprof) web        — 生成火焰图")
	fmt.Println("    6. (pprof) list funcName — 查看函数级热点")
	fmt.Println()
	fmt.Println("  火焰图生成:")
	fmt.Println("    go tool pprof -http=:8081 profile.out")
	fmt.Println("    或: go-torch (Uber 开源)")
	fmt.Println()
	fmt.Println("  perf — Linux 性能分析:")
	fmt.Println("    perf top             — 实时热点函数")
	fmt.Println("    perf record -g -p <PID>  — 采集")
	fmt.Println("    perf report          — 分析")
	fmt.Println()
	fmt.Println("  Go 程序 CPU 飙高排查步骤:")
	fmt.Println("    1. top -H -p <PID>  — 找到高 CPU 的线程")
	fmt.Println("    2. pprof profile    — 确认热点函数")
	fmt.Println("    3. strace -c -p <PID>  — 查看系统调用")
	fmt.Println("    4. go tool trace    — 查看调度情况")
	fmt.Println()
	fmt.Println("  Go 程序内存持续增长排查:")
	fmt.Println("    1. pprof heap        — 查看内存分配")
	fmt.Println("    2. 比较两次 heap 快照 (pprof -base)")
	fmt.Println("    3. 检查 goroutine 泄漏: pprof goroutine")
	fmt.Println("    4. 检查全局变量/slice/map 持续增长")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 故障排查流程
// ----------------------------------------------------------
func troubleshooting() {
	fmt.Println("=== 5. 故障排查流程 ===")
	fmt.Println()
	fmt.Println("  OOM 排查:")
	fmt.Println("    1. dmesg | grep -i oom  — 查看 OOM 日志")
	fmt.Println("    2. /var/log/messages    — 系统日志")
	fmt.Println("    3. Go: runtime.SetMemoryLimit() 设置软限制")
	fmt.Println("    4. cgroup: memory.limit_in_bytes 容器内存限制")
	fmt.Println()
	fmt.Println("  CPU 飙高排查:")
	fmt.Println("    1. top -H -p <PID>       — 定位高 CPU 线程")
	fmt.Println("    2. go tool pprof profile — 定位热点函数")
	fmt.Println("    3. 常见原因:")
	fmt.Println("       - 死循环 (for 条件错误)")
	fmt.Println("       - 正则回溯 (ReDoS)")
	fmt.Println("       - JSON 序列化大量数据")
	fmt.Println("       - GC 压力 (频繁分配)")
	fmt.Println()
	fmt.Println("  死锁排查:")
	fmt.Println("    1. pprof goroutine     — 查看阻塞的 goroutine")
	fmt.Println("    2. go tool pprof goroutine.prof")
	fmt.Println("    3. runtime.LockOSThread 导致的 M 泄漏")
	fmt.Println("    4. channel 操作阻塞")
	fmt.Println("    5. sync.Mutex 互锁")
	fmt.Println()
	fmt.Println("  网络超时排查:")
	fmt.Println("    1. ss -tn | grep <port>  — 连接状态")
	fmt.Println("    2. tcpdump 抓包         — 看是否有 RST/超时")
	fmt.Println("    3. ping/traceroute      — 网络连通性")
	fmt.Println("    4. nslookup/dig         — DNS 解析")
	fmt.Println("    5. Go: net/http 设置 Timeout")
	fmt.Println("       client := &http.Client{Timeout: 10 * time.Second}")
	fmt.Println()
}
