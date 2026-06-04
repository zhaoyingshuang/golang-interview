package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

// ============================================================
// 虚拟内存与内存管理
// ============================================================

func main() {
	virtualMemory()
	goMemoryAllocator()
	stackAndHeap()
	mmapDemo()
	tcmallocPrinciple()
}

// ----------------------------------------------------------
// 1. 虚拟内存原理
// ----------------------------------------------------------
func virtualMemory() {
	fmt.Println("=== 1. 虚拟内存原理 ===")
	fmt.Println()
	fmt.Println("  虚拟地址空间 (64位, 实际使用 48 位):")
	fmt.Println("    用户空间: 0x0000000000000000 ~ 0x00007FFFFFFFFFFF (128TB)")
	fmt.Println("    内核空间: 0xFFFF800000000000 ~ 0xFFFFFFFFFFFFFFFF (128TB)")
	fmt.Println()
	fmt.Println("  页表映射:")
	fmt.Println("    虚拟地址 → 页表 → 物理地址")
	fmt.Println("    页大小: 4KB (标准) / 2MB (大页) / 1GB (巨页)")
	fmt.Println("    多级页表: PML4 → PDPT → PD → PT → Page")
	fmt.Println()
	fmt.Println("  TLB (Translation Lookaside Buffer):")
	fmt.Println("    页表的 CPU 缓存，加速地址翻译")
	fmt.Println("    TLB miss → 需要多次内存访问查页表 → 慢!")
	fmt.Println("    大页 → 减少 TLB miss → 提高性能")
	fmt.Println()
	fmt.Println("  页面置换算法:")
	fmt.Println("    LRU (Least Recently Used) — 最近最少使用")
	fmt.Println("    Clock (近似 LRU) — 时钟算法")
	fmt.Println("    LFU (Least Frequently Used) — 最少使用频率")
	fmt.Println("    Linux 使用: 改进版 Clock (多链表)")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Go 内存分配器 (TCMalloc 思想)
// ----------------------------------------------------------
func goMemoryAllocator() {
	fmt.Println("=== 2. Go 内存分配器 ===")
	fmt.Println()
	fmt.Println("  层级结构:")
	fmt.Println("    mcache  → 每 P 一个，无锁分配")
	fmt.Println("    mcentral → 全局，按 size class 管理")
	fmt.Println("    mheap   → 全局堆，向 OS 申请内存")
	fmt.Println()
	fmt.Println("  分配流程:")
	fmt.Println("    1. size class 查表确定大小类别 (67 种)")
	fmt.Println("    2. 从当前 P 的 mcache 分配 (无锁)")
	fmt.Println("    3. mcache 不够 → 从 mcentral 获取一批")
	fmt.Println("    4. mcentral 不够 → 从 mheap 获取")
	fmt.Println("    5. mheap 不够 → 向 OS 申请 (mmap)")
	fmt.Println()
	fmt.Println("  Size Class (部分):")
	fmt.Println("    class  bytes  objects  waste%")
	fmt.Println("      1       8      512    0.00")
	fmt.Println("      2      16      256    0.00")
	fmt.Println("      3      24      170    1.18")
	fmt.Println("      4      32      128    0.00")
	fmt.Println("      5      48       85    1.18")
	fmt.Println("      ...")
	fmt.Println("     67    32768       1    0.00")
	fmt.Println()
	fmt.Println("  小对象 (< 32KB): mcache → mspan")
	fmt.Println("  大对象 (≥ 32KB): 直接从 mheap 分配")
	fmt.Println()
	fmt.Println("  mspan:")
	fmt.Println("    连续的页组成，包含多个相同大小的对象")
	fmt.Println("    使用位图 (bitmap) 标记对象是否已分配")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 栈与堆
// ----------------------------------------------------------
func stackAndHeap() {
	fmt.Println("=== 3. 栈与堆 ===")
	fmt.Println()
	fmt.Println("  Go 栈管理:")
	fmt.Println("    初始: 2KB (goroutine)")
	fmt.Println("    扩容: 翻倍增长 (连续栈)")
	fmt.Println("    缩容: GC 时检测，缩到实际使用大小")
	fmt.Println()
	fmt.Println("  栈扩容过程:")
	fmt.Println("    1. 函数入口检查栈空间 (morestack)")
	fmt.Println("    2. 空间不够 → 分配新栈 (2x)")
	fmt.Println("    3. 复制旧栈数据到新栈")
	fmt.Println("    4. 调整所有指针指向新栈 (指向旧栈的指针需要更新)")
	fmt.Println("    5. 继续执行")
	fmt.Println()
	fmt.Println("  逃逸到堆的场景:")
	fmt.Println("    1. 返回局部变量指针")
	fmt.Println("    2. 闭包捕获变量")
	fmt.Println("    3. 接口类型赋值")
	fmt.Println("    4. 栈空间不足")
	fmt.Println("    5. 发送到 channel 的数据")
	fmt.Println()

	// 演示逃逸
	escapeExample()
	fmt.Println()

	fmt.Println("  查看逃逸分析:")
	fmt.Println("    go build -gcflags=\"-m\" main.go")
	fmt.Println("    go build -gcflags=\"-m -m\" main.go  # 更详细")
	fmt.Println()
}

// 逃逸到堆 — 返回局部变量指针
func escapeExample() *int {
	x := 42
	return &x // x 逃逸到堆
}

// ----------------------------------------------------------
// 4. 内存映射 mmap
// ----------------------------------------------------------
func mmapDemo() {
	fmt.Println("=== 4. 内存映射 (mmap) ===")
	fmt.Println()
	fmt.Println("  mmap 原理:")
	fmt.Println("    将文件映射到进程的虚拟地址空间")
	fmt.Println("    对映射区域的读写 = 对文件的操作")
	fmt.Println("    无需 read/write 系统调用")
	fmt.Println()
	fmt.Println("  Go 使用:")
	fmt.Println("    data, err := syscall.Mmap(fd, offset, length,")
	fmt.Println("      syscall.PROT_READ|syscall.PROT_WRITE,")
	fmt.Println("      syscall.MAP_SHARED)")
	fmt.Println("    defer syscall.Munmap(data)")
	fmt.Println()
	fmt.Println("  应用场景:")
	fmt.Println("    1. 大文件读写 (避免内核态拷贝)")
	fmt.Println("    2. 共享内存 (IPC)")
	fmt.Println("    3. Go runtime: 堆内存分配用 mmap")
	fmt.Println()
	fmt.Println("  madvise:")
	fmt.Println("    告知内核内存的使用模式")
	fmt.Println("    MADV_DONTNEED — 内存可释放 (Go GC 用)")
	fmt.Println("    MADV_HUGEPAGE — 使用大页")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. TCMalloc 原理
// ----------------------------------------------------------
func tcmallocPrinciple() {
	fmt.Println("=== 5. TCMalloc 原理 ===")
	fmt.Println()
	fmt.Println("  Thread-Caching Malloc (Google):")
	fmt.Println("    Go 的内存分配器基于 TCMalloc 思想")
	fmt.Println()
	fmt.Println("  核心思想:")
	fmt.Println("    1. 线程本地缓存 → 减少全局锁竞争")
	fmt.Println("       Go: P 的 mcache (每个 P 无锁分配)")
	fmt.Println()
	fmt.Println("    2. 按大小分级 → 减少碎片")
	fmt.Println("       Go: 67 种 size class")
	fmt.Println("       小对象: 从 size class 对应的 span 分配")
	fmt.Println("       大对象: 直接从 heap 分配")
	fmt.Println()
	fmt.Println("    3. 中央缓存 → 线程缓存不够时补充")
	fmt.Println("       Go: mcentral → mcache 批量转移")
	fmt.Println()
	fmt.Println("  对比 ptmalloc (glibc):")
	fmt.Println("    ptmalloc: 多 arena，每个 arena 有锁")
	fmt.Println("    TCMalloc: 线程本地缓存，无锁分配")
	fmt.Println("    Go: P 本地缓存，无锁分配")
	fmt.Println()
	fmt.Println("  面试追问: Go 程序 VIRT 为什么比 RSS 大很多?")
	fmt.Println("    VIRT: 虚拟地址空间 (mmap 预留)")
	fmt.Println("    RSS:  实际使用的物理内存")
	fmt.Println("    Go runtime 预先 mmap 大量虚拟地址空间")
	fmt.Println("    但只有在实际使用时才占用物理内存")
	fmt.Println()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  当前进程内存: Alloc=%dMB, Sys=%dMB, HeapAlloc=%dMB\n",
		m.Alloc/1024/1024, m.Sys/1024/1024, m.HeapAlloc/1024/1024)

	_ = unsafe.Sizeof(0) // 避免未使用
}
