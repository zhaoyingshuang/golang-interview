package main

import (
	"fmt"
	"strconv"
	"sync"
)

// ============================================================
// Benchmark 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. 如何写 benchmark？
// 2. b.N 的含义？Go 怎么决定运行多少次？
// 3. 如何避免编译器优化影响 benchmark？
// 4. benchstat 的使用？
// 5. 如何 benchmark 分配次数？

func main() {
	fmt.Println("Benchmark 用法:")
	fmt.Println()
	fmt.Println("  运行: go test -bench=. -benchmem")
	fmt.Println("  对比: benchstat old.txt new.txt")
	fmt.Println()
	fmt.Println("详细 benchmark 代码见 benchmark_test.go")
	fmt.Println()

	// 运行时演示: 字符串拼接性能对比
	compareStringConcat()
}

func compareStringConcat() {
	fmt.Println("=== 字符串拼接性能对比 ===")
	parts := []string{"a", "b", "c", "d", "e"}

	// 方式1: +
	result := ""
	for _, p := range parts {
		result += p
	}
	fmt.Printf("+ 拼接: %s\n", result)

	// 方式2: strconv (比 fmt.Sprintf 快很多)
	s := strconv.Itoa(42)
	fmt.Printf("strconv.Itoa: %s (推荐，比 Sprintf 快 5-10x)\n", s)
}

// ----------------------------------------------------------
// 以下函数在 benchmark_test.go 中使用
// ----------------------------------------------------------

func heavyComputation(n int) int {
	return n * n
}

var globalResult int

var bufPool = sync.Pool{
	New: func() any {
		return new([]byte)
	},
}
