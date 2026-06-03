package main

import (
	"fmt"
	"unsafe"
)

// ============================================================
// 内存对齐深度解析
// ============================================================
//
// 【面试高频问题】
// 1. 什么是内存对齐？为什么需要？
// 2. Go 的对齐规则？
// 3. 如何优化 struct 的字段顺序？
// 4. atomic 操作的对齐要求？
// 5. 64 位原子操作在 32 位系统上的问题？

func main() {
	basicRules()
	structOptimization()
	atomicAlignment()
	zeroSize()
	practical()
}

// ----------------------------------------------------------
// 1. 基本规则
// ----------------------------------------------------------
// 对齐规则:
//   1. 结构体的字段必须按照其类型的对齐系数对齐
//   2. 结构体的整体大小必须是最大对齐系数的倍数
//   3. 基本类型的对齐系数 = 其大小
//
// 常见类型的大小和对齐系数（64位系统）:
//   类型         大小    对齐系数
//   bool         1       1
//   int8/uint8   1       1
//   int16/uint16 2       2
//   int32/uint32 4       4
//   int/uint     8       8
//   int64/uint64 8       8
//   float32      4       4
//   float64      8       8
//   pointer      8       8
//   string       16      8 (Data+Len)
//   slice        24      8 (Data+Len+Cap)
//   interface    16      8

func basicRules() {
	fmt.Println("=== 1. 基本对齐规则 ===")

	type S1 struct {
		A bool    // 1 byte, 对齐到 1
		B int64   // 8 bytes, 对齐到 8 → 需要 7 bytes padding
		C int32   // 4 bytes, 对齐到 4
		// 总大小: 1 + 7(padding) + 8 + 4 + 4(padding) = 24
	}

	type S2 struct {
		A int64   // 8 bytes, 对齐到 8
		B int32   // 4 bytes, 对齐到 4
		C bool    // 1 byte,  对齐到 1
		// 总大小: 8 + 4 + 1 + 3(padding) = 16
	}

	fmt.Printf("S1 (不好): size=%d\n", unsafe.Sizeof(S1{})) // 24
	fmt.Printf("S2 (优化): size=%d\n", unsafe.Sizeof(S2{})) // 16

	fmt.Println("S1 的字段布局:")
	fmt.Printf("  A (bool):  offset=%d, size=%d\n", unsafe.Offsetof(S1{}.A), unsafe.Sizeof(S1{}.A))
	fmt.Printf("  B (int64): offset=%d, size=%d\n", unsafe.Offsetof(S1{}.B), unsafe.Sizeof(S1{}.B))
	fmt.Printf("  C (int32): offset=%d, size=%d\n", unsafe.Offsetof(S1{}.C), unsafe.Sizeof(S1{}.C))
	fmt.Println()
}

// ----------------------------------------------------------
// 2. struct 字段顺序优化
// ----------------------------------------------------------
// 原则: 按字段大小从大到小排列，减少 padding
//
// 优化前:
//   type Bad struct {
//       a bool    // 1 + 7 padding
//       b float64 // 8
//       c int32   // 4
//       // + 4 padding
//   }
//   总大小: 24
//
// 优化后:
//   type Good struct {
//       b float64 // 8
//       c int32   // 4
//       a bool    // 1 + 3 padding
//   }
//   总大小: 16

func structOptimization() {
	fmt.Println("=== 2. 字段顺序优化 ===")

	type Bad struct {
		a bool
		b float64
		c int32
		d int16
		e byte
	}
	// 1+7(pad)+8+4+2+1+1(pad) = 24

	type Good struct {
		b float64 // 8
		c int32   // 4
		d int16   // 2
		a bool    // 1
		e byte    // 1
	}
	// 8+4+2+1+1 = 16

	fmt.Printf("Bad:  size=%d\n", unsafe.Sizeof(Bad{}))  // 24
	fmt.Printf("Good: size=%d\n", unsafe.Sizeof(Good{})) // 16

	// 节省了 33% 的内存！如果有百万个实例，差距巨大

	// 工具: fieldalignment
	// go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	// fieldalignment -fix ./...

	fmt.Println("优化原则: 字段按大小从大到小排列")
	fmt.Println("工具: go install golang.org/x/tools/.../fieldalignment@latest")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 64 位原子操作的对齐要求
// ----------------------------------------------------------
// 在 64 位系统上，64 位原子操作要求 8 字节对齐
// 在 32 位系统上，如果不对齐到 8 字节，atomic 操作会 panic！
//
// 所以 atomic 变量应该放在 struct 的第一个字段

type AtomicCounter struct {
	count int64 // 必须是第一个字段（保证 8 字节对齐）
	flag  bool
}

type BadAtomicCounter struct {
	flag  bool
	count int64 // 在 32 位系统上可能不满足 8 字节对齐！
}

func atomicAlignment() {
	fmt.Println("=== 3. 原子操作对齐 ===")

	ac := AtomicCounter{}
	fmt.Printf("AtomicCounter.count offset=%d (对齐到 8)\n",
		unsafe.Offsetof(ac.count))

	bac := BadAtomicCounter{}
	fmt.Printf("BadAtomicCounter.count offset=%d (在 32 位系统可能不对齐)\n",
		unsafe.Offsetof(bac.count))

	fmt.Println("规则: atomic 64 位变量放在 struct 的第一个字段")
	fmt.Println("32 位系统上不对齐会 panic")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 零大小类型
// ----------------------------------------------------------
// struct{} 的大小为 0，不占用内存
// 用途: 信号 channel (chan struct{})，set 实现

func zeroSize() {
	fmt.Println("=== 4. 零大小类型 ===")

	var s struct{}
	fmt.Printf("struct{} 大小: %d\n", unsafe.Sizeof(s))

	// 用 map 实现 set
	set := make(map[string]struct{})
	set["a"] = struct{}{}
	set["b"] = struct{}{}

	if _, ok := set["a"]; ok {
		fmt.Println("set 包含 'a'")
	}

	// 对比 map[string]bool: 每个 value 多用 1 byte
	// 如果有百万个 key，节省 ~1MB

	fmt.Println("用途: map[T]struct{} 做 set, chan struct{} 做信号")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 实际案例
// ----------------------------------------------------------
func practical() {
	fmt.Println("=== 5. 实际案例 ===")

	// 案例1: slice header 的布局
	type SliceHeader struct {
		Data unsafe.Pointer
		Len  int
		Cap  int
	}
	fmt.Printf("slice header: size=%d\n", unsafe.Sizeof(SliceHeader{})) // 24

	// 案例2: string header 的布局
	type StringHeader struct {
		Data unsafe.Pointer
		Len  int
	}
	fmt.Printf("string header: size=%d\n", unsafe.Sizeof(StringHeader{})) // 16

	// 案例3: interface 的布局
	type IfaceHeader struct {
		Type unsafe.Pointer
		Data unsafe.Pointer
	}
	fmt.Printf("interface header: size=%d\n", unsafe.Sizeof(IfaceHeader{})) // 16

	fmt.Println()
	fmt.Println("面试要点:")
	fmt.Println("  - 字段按大小降序排列减少 padding")
	fmt.Println("  - atomic 64 位变量放第一个字段")
	fmt.Println("  - 用 struct{} 做 set 和信号")
	fmt.Println("  - unsafe.Sizeof 返回编译期大小（不含 slice 底层数组）")
}
