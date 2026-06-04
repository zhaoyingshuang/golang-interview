---
title: 虚拟内存与内存管理
---

## 1. 虚拟内存原理

### 核心概念

虚拟内存为每个进程提供独立的地址空间，通过页表映射到物理内存：

- **页 (Page)**：内存管理的最小单位，通常 4KB
- **页表 (Page Table)**：存储虚拟页到物理页帧的映射关系
- **TLB (Translation Lookaside Buffer)**：页表的硬件缓存，加速地址翻译
- **缺页中断 (Page Fault)**：访问未映射的虚拟地址时触发，由内核处理

::: tip 多级页表
64 位系统地址空间巨大，单级页表无法装下。Linux 使用四级页表：PGD → PUD → PMD → PTE → 物理页帧，仅按需分配页表项，节省内存。
:::

### 页面置换算法

当物理内存不足时，内核需要选择页面换出：

| 算法 | 原理 | 优劣 |
|------|------|------|
| OPT | 替换未来最久不使用的页 | 理论最优，无法实现 |
| FIFO | 替换最早进入的页 | 实现简单，有 Belady 异常 |
| LRU | 替换最近最久未使用的页 | 效果好，开销大 |
| Clock | LRU 的近似，环形链表 + 引用位 | Linux 实际使用的方案 |

```bash
# 查看进程内存映射
cat /proc/$(pidof myapp)/maps

# 查看页面大小
getconf PAGESIZE
# 4096

# 查看缺页中断次数
cat /proc/$(pidof myapp)/stat | awk '{print "minor faults:", $10, "major faults:", $12}'
```

## 2. Go 内存分配器

### TCMalloc 思想

Go 内存分配器借鉴了 Google 的 TCMalloc 设计：

- **线程本地缓存**：避免全局锁竞争
- **大小分类 (Size Class)**：将分配请求对齐到预定义的大小类别
- **中央堆 (Central Heap)**：按大小类别管理空闲内存

### 四级分配结构

```
Goroutine → mcache → mcentral → mheap → OS
```

| 层级 | 说明 | 锁粒度 |
|------|------|--------|
| **mspan** | 最小管理单元，包含若干连续页 | - |
| **mcache** | 每个 P 独有的本地缓存 | 无锁 |
| **mcentral** | 全局中央缓存，按 size class 分类 | 按 size class 加锁 |
| **mheap** | 全局堆，管理所有虚拟内存 | 全局锁 |

```go
package main

import (
    "fmt"
    "runtime"
)

func main() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("HeapAlloc:   %d MB\n", m.HeapAlloc/1024/1024)
    fmt.Printf("HeapSys:     %d MB\n", m.HeapSys/1024/1024)
    fmt.Printf("HeapIdle:    %d MB\n", m.HeapIdle/1024/1024)
    fmt.Printf("HeapInuse:   %d MB\n", m.HeapInuse/1024/1024)
    fmt.Printf("HeapReleased:%d MB\n", m.HeapReleased/1024/1024)
    fmt.Printf("Sys:         %d MB\n", m.Sys/1024/1024)
    fmt.Printf("NumGC:       %d\n", m.NumGC)
}
```

::: info HeapReleased
`HeapReleased` 表示已经通过 `madvise(MADV_DONTNEED)` 归还给操作系统的内存。这就是 Go 程序的 `HeapSys` 可能远大于 `HeapAlloc` 的原因。
:::

## 3. 栈与堆

### Go 栈扩缩容原理

Go 使用**连续栈 (Contiguous Stack)** 管理 Goroutine 的栈空间：

1. **初始栈**：每个 Goroutine 启动时分配 2KB 栈
2. **栈扩容**：函数调用前，运行时检查栈空间是否足够。不足时分配 2 倍大小的新栈，拷贝旧栈内容
3. **栈缩容**：GC 期间，如果栈使用率低于 1/4，缩容到一半

::: warning 逃逸分析
如果编译器判断一个变量的生命周期超出函数作用域，会将其分配到堆上。堆分配需要 GC 回收，开销更大。
:::

### 逃逸到堆的常见场景

```go
package main

import "fmt"

// 1. 返回局部变量指针 — 逃逸
func newInt() *int {
    x := 42
    return &x // x 逃逸到堆
}

// 2. 闭包引用 — 逃逸
func counter() func() int {
    n := 0
    return func() int {
        n++ // n 逃逸到堆
        return n
    }
}

// 3. interface 动态派发 — 逃逸
func printVal(v interface{}) {
    fmt.Println(v) // v 逃逸（fmt.Println 参数是 interface）
}

// 4. slice 扩容可能逃逸
func growSlice() {
    s := make([]int, 0)
    for i := 0; i < 10000; i++ {
        s = append(s, i) // 多次扩容，可能逃逸
    }
}

// 不逃逸的场景
func noEscape() int {
    x := 42
    y := x + 1
    return y // 值拷贝，不逃逸
}

func main() {
    p := newInt()
    fmt.Println(*p)

    c := counter()
    fmt.Println(c(), c())
}
```

查看逃逸分析结果：

```bash
go build -gcflags="-m -m" main.go 2>&1 | grep escape
```

## 4. 内存映射 (mmap)

### mmap 原理

`mmap` 将文件或设备映射到进程的虚拟地址空间，访问内存等同于读写文件：

- 避免内核态和用户态之间的数据拷贝
- 按需调页（lazy allocation），不立即占用物理内存
- 适合大文件读写和共享内存场景

```go
package main

import (
    "fmt"
    "os"
    "syscall"
    "unsafe"
)

func main() {
    f, err := os.OpenFile("test.dat", os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    // 预分配文件大小
    const size = 4096 * 100 // 400 KB
    f.Truncate(size)

    // mmap 映射
    data, err := syscall.Mmap(
        int(f.Fd()),
        0,
        size,
        syscall.PROT_READ|syscall.PROT_WRITE,
        syscall.MAP_SHARED,
    )
    if err != nil {
        panic(err)
    }
    defer syscall.Munmap(data)

    // 直接操作内存 = 操作文件
    for i := 0; i < 100; i++ {
        data[i] = byte(i)
    }

    // 使用 unsafe 访问结构体
    type Record struct {
        ID   uint32
        Name [28]byte
    }
    records := (*[size / 32]Record)(unsafe.Pointer(&data[0]))
    records[0].ID = 42
    copy(records[0].Name[:], "hello mmap")

    fmt.Printf("record: ID=%d, Name=%s\n", records[0].ID, records[0].Name[:])
}
```

::: warning 注意事项
- `mmap` 的文件偏移必须是页大小的整数倍
- 映射区域大小建议是页大小的倍数
- 使用 `unsafe.Pointer` 操作映射内存时需格外小心
- 写入后需要 `Munmap` 或 `Msync` 确保数据落盘
:::

## 5. 面试常见问题

### Go 程序内存占用为什么比 RSS 大

Go 运行时通过 `mmap` 从操作系统申请大块虚拟内存（`HeapSys`），但实际使用的物理页（RSS）只占其中一部分。`HeapSys - HeapInuse = HeapIdle` 是空闲的虚拟内存，而 `HeapReleased` 是已经通过 `madvise(MADV_DONTNEED)` 归还给 OS 的部分。

```go
// 使用 runtime/debug 强制归还内存
import "runtime/debug"

func releaseMemory() {
    debug.FreeOSMemory()
}
```

### madvise 的作用

`madvise` 系统调用用于向内核建议内存区域的使用模式：

- `MADV_DONTNEED`：告知内核这些页面不再需要，内核可立即回收物理页
- `MADV_FREE`：标记为可释放，内核在内存紧张时回收（延迟回收，性能更好）
- `MADV_HUGEPAGE`：建议使用透明大页（THP）

Go 1.12+ 默认使用 `MADV_FREE`（Linux 4.5+），内存压力增大时再转为 `MADV_DONTNEED`。

### 面试追问

**Q1: Go 的内存分配为什么快？**

mcache 是每个 P 独有的本地缓存，分配小对象时直接从 mcache 取，无需加锁。只有 mcache 为空时才去 mcentral 获取，减少了全局锁竞争。

**Q2: 什么是内存对齐？Go 中如何影响结构体大小？**

CPU 访问对齐的内存地址更快。Go 编译器自动对齐字段，可通过调整字段顺序减小结构体大小：

```go
// 24 字节（未优化）
type Bad struct {
    a bool   // 1 + 7 padding
    b int64  // 8
    c int32  // 4 + 4 padding
    d bool   // 1 + 7 padding
}

// 16 字节（优化后）
type Good struct {
    b int64  // 8
    c int32  // 4
    a bool   // 1
    d bool   // 1 + 2 padding
}
```

**Q3: 如何分析 Go 程序的内存使用？**

1. `runtime.ReadMemStats` 获取详细内存统计
2. `go tool pprof http://localhost:6060/debug/pprof/heap` 分析堆分配
3. `go tool pprof -alloc_objects` 查看分配对象数
4. `GODEBUG=gctrace=1` 打印 GC 日志分析回收频率

**Q4: 逃逸分析有什么用？如何减少堆分配？**

逃逸分析让编译器决定变量分配在栈还是堆上。栈分配零成本（函数返回自动回收），堆分配需要 GC。减少逃逸的方法：(1) 避免返回局部变量指针；(2) 使用值类型而非指针；(3) 预分配 slice/map 大小；(4) 使用 `sync.Pool` 复用对象。
