---
title: Linux 常用命令与排查
---

## 1. 进程管理

### ps — 查看进程状态

```bash
# 查看所有进程，显示 PID/TTY/STAT/COMMAND
ps aux

# 查看进程树
ps auxf

# 查找特定进程
ps aux | grep myapp

# 查看指定进程的线程
ps -T -p $(pidof myapp)

# 进程状态码说明
# R: Running       S: Sleeping       D: Disk Sleep (不可中断)
# Z: Zombie        T: Stopped        I: Idle
```

### top / htop — 实时监控

```bash
# top 常用交互命令
# M: 按内存排序    P: 按 CPU 排序    1: 显示每个 CPU 核心
# H: 显示线程      c: 显示完整命令

# 批量模式（适合脚本）
top -b -n 1 -p $(pidof myapp)

# htop 更直观，支持鼠标操作和树形视图
htop -p $(pidof myapp)
```

### strace — 系统调用追踪

```bash
# 追踪 Go 程序的所有系统调用
strace -p $(pidof myapp) -f -tt -T

# 只追踪特定系统调用
strace -p $(pidof myapp) -e trace=network
strace -p $(pidof myapp) -e trace=file
strace -p $(pidof myapp) -e trace=epoll_wait,read,write

# 统计系统调用耗时
strace -p $(pidof myapp) -c
```

::: tip Go 程序排查实战
Go 程序 CPU 飙高时，先用 `top -H` 找到高 CPU 的线程 ID，然后用 `strace -p <tid>` 追踪该线程的系统调用，或者使用 `go tool pprof` 进行更精确的分析。
:::

### kill — 发送信号

```bash
kill -9  $(pidof myapp)   # SIGKILL: 强制终止
kill -15 $(pidof myapp)   # SIGTERM: 优雅退出
kill -USR1 $(pidof myapp) # SIGUSR1: 自定义信号（如重载配置）
kill -HUP  $(pidof myapp) # SIGHUP: 重载配置
```

## 2. 网络排查

### ss / netstat — 连接状态分析

```bash
# ss 比 netstat 更快（直接读内核 socket 数据结构）
ss -tlnp     # 所有 TCP 监听端口
ss -tnp      # 所有 TCP 已建立连接
ss -s        # 连接统计摘要

# 按状态统计 TCP 连接
ss -ant | awk 'NR>1 {count[$1]++} END {for(s in count) print s, count[s]}'
# 输出示例:
# ESTAB 1250
# TIME-WAIT 320
# CLOSE-WAIT 5

# 排查 TIME-WAIT 堆积
ss -ant state time-wait | wc -l
```

::: warning TIME-WAIT 堆积
大量 TIME-WAIT 通常是因为客户端频繁创建短连接。解决方案：(1) 使用连接池；(2) 启用 `SO_REUSEADDR`；(3) 调整 `tcp_tw_reuse`（生产环境慎用）。
:::

### tcpdump — 抓包分析

```bash
# 抓取指定端口的包
tcpdump -i any port 8080 -nn

# 抓取指定 IP 的 TCP SYN 包
tcpdump -i any host 10.0.0.1 and "tcp[tcpflags] & (tcp-syn) != 0"

# 保存到文件用 Wireshark 分析
tcpdump -i any port 8080 -w capture.pcap

# 只看包内容（ASCII）
tcpdump -i any port 8080 -A -s 0
```

### nc — 网络工具

```bash
# 测试 TCP 连通性
nc -zv 10.0.0.1 8080

# 测试 UDP 连通性
nc -zuv 10.0.0.1 8080

# 简易 TCP 服务器
nc -l 9090

# 传输文件
# 接收端: nc -l 9090 > file.txt
# 发送端: nc 10.0.0.1 9090 < file.txt
```

### curl — HTTP 请求

```bash
# 查看完整请求/响应（含 headers）
curl -v http://localhost:8080/api/health

# 只看响应头
curl -I http://localhost:8080/api/health

# 测试接口延迟
curl -o /dev/null -s -w "connect: %{time_connect}s\nttfb: %{time_starttransfer}s\ntotal: %{time_total}s\n" http://localhost:8080/api/health
```

## 3. 磁盘与文件

### df / du — 磁盘使用

```bash
# 查看磁盘使用情况
df -h

# 查看指定目录大小
du -sh /var/log/

# 查看当前目录下各子目录大小
du -h --max-depth=1 /var/log/ | sort -rh

# 查看 inode 使用
df -i
```

### lsof — 文件句柄排查

```bash
# 查看进程打开的所有文件
lsof -p $(pidof myapp)

# 查看进程打开的文件数量
lsof -p $(pidof myapp) | wc -l

# 查看哪个进程占用了指定端口
lsof -i :8080

# 查看已删除但仍被进程占用的文件
lsof | grep deleted

# 查看系统文件描述符限制
ulimit -n
cat /proc/$(pidof myapp)/limits | grep "open files"
```

::: tip 文件句柄泄漏排查
如果 Go 程序出现 `too many open files` 错误：(1) `lsof -p <pid> | wc -l` 查看当前打开数；(2) 对比 `ulimit -n` 的限制；(3) `lsof -p <pid>` 检查是否有大量未关闭的 socket 或文件；(4) 检查代码中 `defer f.Close()` 是否在循环中延迟关闭。
:::

### iostat — IO 性能

```bash
# 查看 IO 统计（每秒刷新）
iostat -xz 1

# 关注指标:
# %util: 设备利用率，接近 100% 说明 IO 饱和
# await: 平均 IO 等待时间（ms）
# svctm: 平均服务时间
# w/s: 每秒写次数
```

## 4. 性能分析

### perf — 性能剖析

```bash
# 采样 CPU profile（30 秒）
perf record -p $(pidof myapp) -g -- sleep 30

# 查看报告
perf report

# 生成火焰图
perf script | stackcollapse-perf.pl | flamegraph.pl > flame.svg
# 或使用 FlameGraph 工具:
git clone https://github.com/brendangregg/FlameGraph
perf script | FlameGraph/stackcollapse-perf.pl | FlameGraph/flamegraph.pl > cpu.svg
```

### Go 程序 CPU 分析工作流

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
    "runtime/pprof"
    "os"
)

func main() {
    // 方式1: 内置 pprof HTTP 服务器
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // 方式2: 写入文件
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // 业务代码...
    doWork()
}

func doWork() {
    // 模拟工作负载
}
```

```bash
# 采集 CPU profile（30 秒）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 采集内存 profile
go tool pprof http://localhost:6060/debug/pprof/heap

# 在 pprof 交互界面
(pprof) top 20          # 查看 top 20 热点函数
(pprof) web             # 生成调用图（需要 graphviz）
(pprof) list funcName   # 查看函数级别的逐行分析
(pprof) svg             # 导出 SVG 图

# 对比两个时间点的 profile
go tool pprof -base cpu1.prof cpu2.prof
```

### Go 程序内存分析

```bash
# 查看当前内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看累计分配对象数（比默认的 inuse_objects 更易发现泄漏）
go tool pprof -alloc_objects http://localhost:6060/debug/pprof/heap

# 查看正在使用的对象
go tool pprof -inuse_objects http://localhost:6060/debug/pprof/heap

# GC 追踪
GODEBUG=gctrace=1 ./myapp
# 输出示例:
# gc 1 @0.003s 5%: 0.018+0.52+0.004 ms clock, 0.14+0.25+0.046 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
```

## 5. 面试常见问题

### 线上 OOM 如何排查

```bash
# 1. 查看系统日志
dmesg | grep -i "out of memory" | tail -20
dmesg | grep -i "killed process"

# 2. 查看 cgroup 内存限制（容器环境）
cat /sys/fs/cgroup/memory/memory.usage_in_bytes
cat /sys/fs/cgroup/memory/memory.limit_in_bytes

# 3. Go pprof 分析内存
go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top
(pprof) web

# 4. 生成 core dump（Go 1.16+）
GOTRACEBACK=crash ./myapp
# 崩溃后用 Go tool 分析
go tool pprof myapp core
```

::: info Go OOM 常见原因
1. Goroutine 泄漏导致栈内存持续增长
2. 全局 map/slice 无限增长
3. 大量 `[]byte` 字符串转换导致内存翻倍
4. `defer` 在循环中使用导致资源延迟释放
5. `sync.Pool` 使用不当，缓存了过多大对象
:::

### CPU 飙高排查步骤

```bash
# 1. 定位高 CPU 进程
top -c    # 找到 PID

# 2. 定位高 CPU 线程
top -Hp <PID>    # 找到线程 TID

# 3. 转换线程 TID 为十六进制
printf "%x\n" <TID>

# 4. 查看线程堆栈
jstack <PID> | grep <HEX_TID> -A 50   # Java
go tool pprof http://localhost:6060/debug/pprof/profile  # Go

# 5. 或使用 perf 追踪
perf top -p <PID>
```

### 死锁排查方法

```bash
# Go 死锁排查
# 1. 使用 pprof goroutine 查看阻塞的 goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) traces   # 查看调用栈

# 2. 查看 goroutine 数量趋势
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1

# 3. Go 运行时会自动检测死锁（仅限所有 goroutine 都阻塞的情况）
# fatal error: all goroutines are asleep - deadlock!
```

```go
// 使用 go-deadlock 库检测死锁
import "github.com/sasha-s/go-deadlock"

func init() {
    deadlock.Opts.DumpAllGoroutines = true
    deadlock.Opts.OnPotentialDeadlock = func() {
        log.Println("potential deadlock detected!")
    }
}

// 替换 sync.Mutex 为 deadlock.Mutex
type SafeMap struct {
    mu   deadlock.Mutex
    data map[string]string
}
```

### 面试追问

**Q1: 如何排查 Go 程序的 Goroutine 泄漏？**

1. `curl /debug/pprof/goroutine?debug=1` 查看所有 goroutine 栈
2. 对比两次采样的 goroutine 数量是否持续增长
3. 重点检查 channel 操作（无缓冲 channel 未收发、select 无 default 分支）
4. 使用 `goleak` 在测试中检测：`defer goleak.VerifyTestMain(m)`

**Q2: 容器中 top 显示的内存为什么不准确？**

`top` 显示的是宿主机的 CPU 和内存信息，容器内看到的进程信息可能包含宿主机所有进程。在 Kubernetes 中应使用 `kubectl top pod` 或查看 cgroup 指标（`/sys/fs/cgroup/memory/`）。Go 1.19+ 的 `GODEBUG=cgocheck=0` 和 cgroup v2 支持改善了容器内的内存统计。

**Q3: 如何快速定位 Go 程序的慢请求？**

1. 在 HTTP 中间件记录请求耗时，超过阈值打印调用栈
2. 使用 `go tool trace` 可视化分析请求生命周期
3. OpenTelemetry 分布式追踪（Jaeger/Zipkin）
4. `net/http/pprof` 的 profile endpoint 对比正常/异常时段

**Q4: 线上问题排查的一般方法论是什么？**

遵循 "先止损、再定位、后修复" 原则：(1) 监控告警确认异常；(2) `top/df/ss` 快速定位资源瓶颈；(3) `strace/perf/tcpdump` 深入分析；(4) `pprof` 定位代码热点；(5) 修复并灰度验证。关键是用好 `pprof` 和 `trace`，这是 Go 程序排查的核心武器。
