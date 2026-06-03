---
title: Benchmark 基准测试
---

## 1. 基本用法

函数签名 func BenchmarkXxx(b *testing.B)，循环 b.N 次。b.N 由框架自动调整（从 1 开始，逐步增大），直到运行时间足够长（默认 >= 1s）结果可信。运行：go test -bench=. -benchmem。

**输出解读**：671.4 ns/op（每次操作耗时）、2040 B/op（每次分配字节数）、8 allocs/op（每次分配次数）。

::: tip 使用场景
优化前后对比、竞品方案选型、CI 性能回归检测。

```go
func BenchmarkSliceAppend(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}

func BenchmarkSlicePrealloc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 100)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}
// go test -bench=. -benchmem
// Append:   671 ns/op  2040 B/op  8 allocs/op
// Prealloc:  73 ns/op     0 B/op  0 allocs/op
```

## 2. 避免编译器优化与子 Benchmark

编译器可能优化掉没有副作用的代码。解决：将结果赋给**包级变量**。Go 1.24+ 可用 b.Loop() 更简洁。
:::

**子 Benchmark**：用 b.Run 创建子测试，对比多种实现方案。

**b.ReportAllocs()**：即使不用 -benchmem 也会报告分配信息。

```go
var globalResult int

func BenchmarkSafe(b *testing.B) {
    var result int
    for i := 0; i < b.N; i++ {
        result = heavyComputation(i)
    }
    globalResult = result  // 防止优化
}

// Go 1.24+: b.Loop()
func BenchmarkLoop(b *testing.B) {
    for b.Loop() {
        heavyComputation(0)
    }
}

// 子 Benchmark（对比多种实现）
func BenchmarkConcat(b *testing.B) {
    parts := []string{"a", "b", "c", "d", "e"}
    b.Run("plus", func(b *testing.B) { /* ... */ })
    b.Run("builder", func(b *testing.B) { /* ... */ })
    b.Run("join", func(b *testing.B) { /* ... */ })
}
```

## 3. benchstat 对比与实战技巧

**benchstat** 用于统计显著性对比。优化前后各运行多次（-count=5），用 benchstat 对比，输出 delta 百分比和置信区间。

**实战技巧**：1) -count=5 多次运行取中位数；2) -benchtime=5s 增加采样时间；3) -run=^$ 跳过单元测试加速；4) 关闭其他程序减少噪声。

::: tip 使用场景
PR 提交前跑 benchmark 确认没有性能退化。

```go
# 完整的 benchmark 对比流程

# 优化前（运行 5 次取统计）
go test -bench=BenchmarkConcat -count=5 -benchmem > old.txt

# 优化后
go test -bench=BenchmarkConcat -count=5 -benchmem > new.txt

# 对比
benchstat old.txt new.txt
# 输出:
# name         old time/op  new time/op  delta
# Concat/plus    120ns ± 2%    45ns ± 1%  -62.50%
# Concat/builder  80ns ± 1%    40ns ± 2%  -50.00%

# 常用参数:
# -benchmem     显示分配信息
# -count=5      多次运行
# -benchtime=5s 增加采样时间
# -run=^$       跳过单元测试
```
