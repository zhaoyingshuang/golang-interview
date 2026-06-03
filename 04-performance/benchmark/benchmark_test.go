package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// ----------------------------------------------------------
// 1. 基本 Benchmark
// ----------------------------------------------------------
// b.N 由 testing 框架自动调整，直到运行时间足够长（默认 >= 1s）

func BenchmarkSliceAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0)
		for j := 0; j < 100; j++ {
			s = append(s, j)
		}
		_ = s
	}
}

func BenchmarkSlicePrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, 100)
		for j := 0; j < 100; j++ {
			s = append(s, j)
		}
		_ = s
	}
}

// ----------------------------------------------------------
// 2. 避免编译器优化
// ----------------------------------------------------------
// 编译器可能会优化掉没有副作用的代码
// 方法: 将结果赋给包级变量

func BenchmarkWithSideEffect(b *testing.B) {
	var result int
	for i := 0; i < b.N; i++ {
		result = heavyComputation(i)
	}
	globalResult = result
}

// ----------------------------------------------------------
// 3. 子 Benchmark
// ----------------------------------------------------------

func BenchmarkStringConcat(b *testing.B) {
	parts := []string{"a", "b", "c", "d", "e"}

	b.Run("plus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := parts[0] + parts[1] + parts[2] + parts[3] + parts[4]
			_ = s
		}
	})

	b.Run("builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buf strings.Builder
			for _, p := range parts {
				buf.WriteString(p)
			}
			_ = buf.String()
		}
	})

	b.Run("join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = strings.Join(parts, "")
		}
	})
}

// ----------------------------------------------------------
// 4. 统计分配次数
// ----------------------------------------------------------

func BenchmarkSprintfAlloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("hello %d", i)
	}
}

func BenchmarkItoaNoAlloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = strconv.Itoa(i)
	}
}

// ----------------------------------------------------------
// 5. sync.Pool 对比
// ----------------------------------------------------------

func BenchmarkWithPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := bufPool.Get().(*[]byte)
		*buf = (*buf)[:0]
		*buf = append(*buf, "hello"...)
		_ = string(*buf)
		bufPool.Put(buf)
	}
}

func BenchmarkWithoutPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 5)
		buf = append(buf, "hello"...)
		_ = string(buf)
	}
}

// ----------------------------------------------------------
// 6. map 预分配
// ----------------------------------------------------------

func BenchmarkMapPrealloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := make(map[int]int, 100)
		for j := 0; j < 100; j++ {
			m[j] = j
		}
	}
}

func BenchmarkMapNoPrealloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := make(map[int]int)
		for j := 0; j < 100; j++ {
			m[j] = j
		}
	}
}

// ----------------------------------------------------------
// 7. bytes.Buffer vs strings.Builder
// ----------------------------------------------------------

func BenchmarkBytesBuffer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		buf.WriteString("hello")
		buf.WriteString(" ")
		buf.WriteString("world")
		_ = buf.String()
	}
}

func BenchmarkStringsBuilder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf strings.Builder
		buf.WriteString("hello")
		buf.WriteString(" ")
		buf.WriteString("world")
		_ = buf.String()
	}
}

// ----------------------------------------------------------
// 使用方式:
//
// 运行所有 benchmark:
//   go test -bench=. -benchmem ./04-performance/benchmark/
//
// 运行特定 benchmark:
//   go test -bench=BenchmarkStringConcat -benchmem ./04-performance/benchmark/
//
// 多次运行用于统计对比:
//   go test -bench=. -count=5 > old.txt
//   # 优化后
//   go test -bench=. -count=5 > new.txt
//   # 对比
//   benchstat old.txt new.txt
// ----------------------------------------------------------
