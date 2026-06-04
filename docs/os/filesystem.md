---
title: 文件系统
---

## 1. 文件系统原理

### inode 与目录项

Linux 文件系统中，每个文件由一个 **inode**（索引节点）唯一标识：

- **inode** 存储文件的元数据：大小、权限、时间戳、数据块位置
- **目录项 (dentry)** 存储文件名到 inode 编号的映射关系
- **数据块 (data block)** 存储文件的实际内容

```
目录项: "main.go" → inode 12345
inode 12345: { mode: 0644, size: 1024, blocks: [4096, 4097] }
数据块 4096: "package main\n..."
```

### 软链接与硬链接

| 特性 | 硬链接 | 软链接 (Symbolic Link) |
|------|--------|----------------------|
| 本质 | 指向相同 inode 的新目录项 | 指向路径名的特殊文件 |
| 删除原文件 | 不影响硬链接 | 软链接失效（dangling） |
| 跨文件系统 | 不可以 | 可以 |
| 链接目录 | 不可以 | 可以 |

```bash
# 创建硬链接
ln original.txt hardlink.txt

# 创建软链接
ln -s original.txt symlink.txt

# 查看 inode 号
ls -li original.txt hardlink.txt symlink.txt
```

### ext4 vs xfs

| 特性 | ext4 | xfs |
|------|------|-----|
| 最大文件 | 16TB | 8EB |
| 最大文件系统 | 1EB | 8EB |
| 延迟分配 | 支持 | 支持 |
| 在线扩缩容 | 支持缩容 | 只支持扩容 |
| 适用场景 | 通用、小文件多 | 大文件、高并发 IO |

::: tip 生产环境选择
Kubernetes 默认使用 ext4 作为 Pod 卷的文件系统。对于大数据、日志存储等大文件场景，xfs 性能更优。etcd 官方推荐使用 xfs 以获得更好的 WAL 写入性能。
:::

## 2. Go 文件操作

### 基础文件读写

```go
package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
)

// 小文件直接读取
func readSmallFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

// 小文件直接写入
func writeSmallFile(path string, data []byte) error {
    return os.WriteFile(path, data, 0644)
}

// 大文件逐行读取（内存友好）
func readLargeFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    // 增大缓冲区以处理长行
    scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
    for scanner.Scan() {
        line := scanner.Text()
        _ = line // 处理每一行
    }
    return scanner.Err()
}

// 带缓冲写入（减少系统调用次数）
func bufferedWrite(path string, lines []string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    w := bufio.NewWriterSize(f, 64*1024) // 64KB 缓冲
    for _, line := range lines {
        fmt.Fprintln(w, line)
    }
    return w.Flush()
}
```

### 使用 io.Copy 高效拷贝

```go
package main

import (
    "io"
    "os"
)

// io.Copy 内部使用 32KB 缓冲，且自动选择 sendfile 等零拷贝策略
func copyFile(src, dst string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    written, err := io.Copy(dstFile, srcFile)
    if err != nil {
        return err
    }
    _ = written
    return nil
}
```

::: info bufio 的性能影响
默认情况下 `os.File.Read` 每次调用都是一次 `read` 系统调用。`bufio.Reader` 在用户态维护一块缓冲区，一次 `read` 读取大量数据填满缓冲区，后续 `Read` 从缓冲区返回，显著减少系统调用次数。
:::

## 3. 文件锁

### flock 与 fcntl

| 特性 | flock | fcntl (POSIX 锁) |
|------|-------|------------------|
| 锁粒度 | 整个文件 | 文件区间（byte-range） |
| 关联对象 | 文件描述符（fd） | (进程, inode) 对 |
| 继承 | fork 子进程继承 | 不继承 |
| NFS 支持 | Linux 2.6.12+ | 支持 |

```go
package main

import (
    "fmt"
    "os"
    "syscall"
)

// 使用 flock 实现进程级互斥
type FileLock struct {
    f *os.File
}

func NewFileLock(path string) (*FileLock, error) {
    f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return nil, err
    }
    return &FileLock{f: f}, nil
}

func (fl *FileLock) Lock() error {
    return syscall.Flock(int(fl.f.Fd()), syscall.LOCK_EX)
}

func (fl *FileLock) TryLock() error {
    err := syscall.Flock(int(fl.f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
    if err != nil {
        return fmt.Errorf("lock is held by another process: %w", err)
    }
    return nil
}

func (fl *FileLock) Unlock() error {
    return syscall.Flock(int(fl.f.Fd()), syscall.LOCK_UN)
}

func (fl *FileLock) Close() error {
    return fl.f.Close()
}

// 使用示例
func main() {
    lock, _ := NewFileLock("/tmp/myapp.lock")
    defer lock.Close()

    if err := lock.TryLock(); err != nil {
        fmt.Println("another instance is running")
        os.Exit(1)
    }
    // 确保单实例运行
    fmt.Println("running...")
    select {}
}
```

::: warning 分布式环境
`flock` 和 `fcntl` 只在单机有效。分布式场景需要使用分布式锁（如 etcd、Redis RedLock），或分布式文件锁（如 NFS 的 `fcntl` 锁，但可靠性依赖 NFS 实现）。
:::

## 4. 文件监控

### inotify 原理

Linux 内核 2.6.13+ 提供的文件系统事件监控机制：

```bash
# 查看 inotify 上限
cat /proc/sys/fs/inotify/max_user_watches
# 通常 524288 (512K)

# 调整上限
echo 524288 | sudo tee /proc/sys/fs/inotify/max_user_watches
```

### Go 文件变更监听 (fsnotify)

```go
package main

import (
    "fmt"
    "log"

    "github.com/fsnotify/fsnotify"
)

func main() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    // 监控目录
    err = watcher.Add("/var/log/myapp")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("watching /var/log/myapp ...")

    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            fmt.Printf("event: %s %s\n", event.Name, event.Op)
            if event.Op&fsnotify.Write == fsnotify.Write {
                fmt.Printf("modified file: %s\n", event.Name)
            }
            if event.Op&fsnotify.Create == fsnotify.Create {
                fmt.Printf("new file: %s\n", event.Name)
                // 新文件需要显式加入监控
                watcher.Add(event.Name)
            }

        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            fmt.Printf("watcher error: %v\n", err)
        }
    }
}
```

::: tip inotify 的注意事项
- 监控大量文件/目录时需要调大 `max_user_watches`
- 监控的是 inode，文件被替换（如 `mv` 覆盖）可能导致监控丢失
- Docker 容器中挂载卷的 inotify 事件可能不完整，需要特殊处理
:::

## 5. 面试常见问题

### 大文件读取优化

```go
// 方案1：分块读取
func readInChunks(path string, chunkSize int) error {
    f, _ := os.Open(path)
    defer f.Close()

    buf := make([]byte, chunkSize)
    for {
        n, err := f.Read(buf)
        if n > 0 {
            process(buf[:n])
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
    }
    return nil
}

// 方案2：mmap 映射（见虚拟内存章节）
// 方案3：并发读取，Seek 到不同偏移量并行处理
```

### 日志轮转实现

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type RotatingLogger struct {
    dir      string
    baseName string
    maxSize  int64
    file     *os.File
    written  int64
}

func (l *RotatingLogger) Write(p []byte) (int, error) {
    if l.file == nil || l.written+int64(len(p)) > l.maxSize {
        l.rotate()
    }
    n, err := l.file.Write(p)
    l.written += int64(n)
    return n, err
}

func (l *RotatingLogger) rotate() {
    if l.file != nil {
        l.file.Close()
        // 重命名旧文件
        oldName := l.file.Name()
        ts := time.Now().Format("20060102-150405")
        os.Rename(oldName, filepath.Join(l.dir, l.baseName+"."+ts))
    }

    name := filepath.Join(l.dir, l.baseName)
    f, _ := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    l.file = f
    l.written = 0
    fmt.Println("log rotated to", name)
}
```

### 面试追问

**Q1: 如何实现并发安全的文件写入？**

(1) 使用 `sync.Mutex` 保护写入操作；(2) 使用 channel 将写入请求序列化到单个 goroutine（actor 模型）；(3) 使用 `O_APPEND` 标志打开文件，内核保证每次 write 是原子追加（但超过页大小的写入不保证原子性）。

**Q2: 为什么 `rm` 删除文件后磁盘空间没有释放？**

如果有进程仍然持有已删除文件的 fd（通过 `open` 打开后未关闭），文件的 inode 不会被回收，磁盘空间不会释放。用 `lsof | grep deleted` 可以找到这些进程，关闭对应进程或 fd 即可释放空间。

**Q3: inode 用完了怎么办？**

每个文件系统有固定的 inode 数量（`mkfs` 时确定），即使磁盘有空间也无法创建新文件。查看 `df -i` 可确认。解决方法：(1) 删除无用文件；(2) 重新格式化增加 inode 数量；(3) 使用更大 inode 密度的文件系统。

**Q4: 如何监控配置文件变更并热重载？**

使用 `fsnotify` 监控配置文件，收到 `Write` 事件后延迟一小段时间（防抖）再重新加载配置。注意 SIGHUP 信号也是常见的重载触发方式（如 nginx -s reload），Go 程序可监听 `syscall.SIGHUP` 实现类似机制。
