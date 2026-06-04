package main

import (
	"fmt"
	"os"
)

// ============================================================
// 文件系统
// ============================================================

func main() {
	fsPrinciple()
	goFileOps()
	fileLock()
	fileMonitor()
	tips()
}

// ----------------------------------------------------------
// 1. 文件系统原理
// ----------------------------------------------------------
func fsPrinciple() {
	fmt.Println("=== 1. 文件系统原理 ===")
	fmt.Println()
	fmt.Println("  ext4 文件系统结构:")
	fmt.Println("    Superblock — 文件系统元信息 (块大小/inode 总数)")
	fmt.Println("    Block Group — 块组 (数据分布)")
	fmt.Println("      ├── Block Bitmap (数据块位图)")
	fmt.Println("      ├── Inode Bitmap (inode 位图)")
	fmt.Println("      ├── Inode Table (inode 表)")
	fmt.Println("      └── Data Blocks (数据块)")
	fmt.Println()
	fmt.Println("  Inode:")
	fmt.Println("    存储文件元数据 (权限/大小/时间戳/数据块指针)")
	fmt.Println("    不存储文件名! 文件名存在目录项 (dentry) 中")
	fmt.Println("    df -i 查看 inode 使用情况")
	fmt.Println()
	fmt.Println("  硬链接 vs 软链接:")
	fmt.Println("    硬链接: 相同 inode, 删除原文件不影响硬链接")
	fmt.Println("      ln source.txt hardlink.txt")
	fmt.Println("    软链接: 独立 inode, 存储目标路径, 类似快捷方式")
	fmt.Println("      ln -s source.txt symlink.txt")
	fmt.Println()
	fmt.Println("  ext4 vs xfs:")
	fmt.Println("    ext4: 稳定, 默认, 适合小文件")
	fmt.Println("    xfs:  高性能, 大文件, 并行 IO, CentOS 8 默认")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Go 文件操作
// ----------------------------------------------------------
func goFileOps() {
	fmt.Println("=== 2. Go 文件操作 ===")
	fmt.Println()
	fmt.Println("  基本读写:")
	fmt.Println("    data, err := os.ReadFile(\"config.yaml\")  // 小文件")
	fmt.Println("    err := os.WriteFile(\"out.txt\", data, 0644)")
	fmt.Println()
	fmt.Println("  流式读写 (大文件):")
	fmt.Println("    f, _ := os.Open(\"large.log\")")
	fmt.Println("    defer f.Close()")
	fmt.Println("    scanner := bufio.NewScanner(f)")
	fmt.Println("    for scanner.Scan() { line := scanner.Text() }")
	fmt.Println()
	fmt.Println("  bufio 缓冲:")
	fmt.Println("    bufio.NewReader(f) → 4KB 缓冲区")
	fmt.Println("    减少 read 系统调用次数")
	fmt.Println("    适用于行读/小量多次读取")
	fmt.Println()
	fmt.Println("  mmap 大文件:")
	fmt.Println("    data, _ := syscall.Mmap(int(f.Fd()), 0, size,")
	fmt.Println("      syscall.PROT_READ, syscall.MAP_SHARED)")
	fmt.Println("    // data 就是文件内容，直接通过指针访问")
	fmt.Println("    // 适用于 GB 级文件的随机读写")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 文件锁
// ----------------------------------------------------------
func fileLock() {
	fmt.Println("=== 3. 文件锁 ===")
	fmt.Println()
	fmt.Println("  两种锁:")
	fmt.Println("    1. flock() — 整个文件级别的建议锁")
	fmt.Println("       syscall.Flock(fd, syscall.LOCK_EX)  // 排他锁")
	fmt.Println("       syscall.Flock(fd, syscall.LOCK_SH)  // 共享锁")
	fmt.Println("       syscall.Flock(fd, syscall.LOCK_UN)  // 解锁")
	fmt.Println()
	fmt.Println("    2. fcntl() — 字节范围的记录锁")
	fmt.Println("       可以锁定文件的特定区域")
	fmt.Println("       适用于多进程写同一文件的不同区域")
	fmt.Println()
	fmt.Println("  Go 文件锁示例:")
	fmt.Println("    f, _ := os.OpenFile(\"/tmp/app.lock\", os.O_CREATE|os.O_RDWR, 0644)")
	fmt.Println("    err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)")
	fmt.Println("    if err != nil {")
	fmt.Println("      log.Fatal(\"另一个实例已在运行\")")
	fmt.Println("    }")
	fmt.Println("    defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)")
	fmt.Println()
	fmt.Println("  分布式文件锁:")
	fmt.Println("    单机文件锁只在同一机器生效")
	fmt.Println("    跨机器: Redis SET NX / Etcd / ZooKeeper")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 文件监控
// ----------------------------------------------------------
func fileMonitor() {
	fmt.Println("=== 4. 文件监控 ===")
	fmt.Println()
	fmt.Println("  inotify 原理 (Linux):")
	fmt.Println("    inotify_init()    — 创建实例")
	fmt.Println("    inotify_add_watch() — 注册监控")
	fmt.Println("    read()            — 读取事件")
	fmt.Println("    事件: IN_CREATE/IN_MODIFY/IN_DELETE/IN_MOVED")
	fmt.Println()
	fmt.Println("  Go fsnotify:")
	fmt.Println("    watcher, _ := fsnotify.NewWatcher()")
	fmt.Println("    defer watcher.Close()")
	fmt.Println("    watcher.Add(\"/path/to/dir\")")
	fmt.Println()
	fmt.Println("    for {")
	fmt.Println("      select {")
	fmt.Println("      case event := <-watcher.Events:")
	fmt.Println("        if event.Op&fsnotify.Write == fsnotify.Write {")
	fmt.Println("          log.Printf(\"文件修改: %s\", event.Name)")
	fmt.Println("        }")
	fmt.Println("      case err := <-watcher.Errors:")
	fmt.Println("        log.Println(err)")
	fmt.Println("      }")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("  应用场景:")
	fmt.Println("    配置热更新、日志监控、文件同步")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 最佳实践
// ----------------------------------------------------------
func tips() {
	fmt.Println("=== 5. 文件操作最佳实践 ===")
	fmt.Println()
	fmt.Println("  大文件读取:")
	fmt.Println("    ✗ os.ReadFile — 全部加载到内存")
	fmt.Println("    ✓ bufio.Scanner — 逐行处理")
	fmt.Println("    ✓ mmap — 随机访问大文件")
	fmt.Println()
	fmt.Println("  日志轮转:")
	fmt.Println("    1. lumberjack 库: 自动按大小/时间轮转")
	fmt.Println("       &lumberjack.Logger{")
	fmt.Println("         Filename:   \"app.log\",")
	fmt.Println("         MaxSize:    100, // MB")
	fmt.Println("         MaxBackups: 3,")
	fmt.Println("         MaxAge:     28,  // days")
	fmt.Println("       }")
	fmt.Println("    2. 自己实现: copytruncate / rename+create")
	fmt.Println()
	fmt.Println("  并发安全写入:")
	fmt.Println("    os.File.Write() 在 Linux 上对普通文件是原子 (小于 PIPE_BUF)")
	fmt.Println("    但不保证跨平台，生产环境用 sync.Mutex 或 channel")
	fmt.Println()
	fmt.Println("  面试追问: 如何实现一个简单的日志轮转?")
	fmt.Println("    1. 检查文件大小 > 阈值")
	fmt.Println("    2. rename app.log → app.log.1")
	fmt.Println("    3. 创建新 app.log")
	fmt.Println("    4. 通知写入方重新打开文件 (或使用 fd)")
	fmt.Println()

	_ = os.Getenv // 避免 unused
}
