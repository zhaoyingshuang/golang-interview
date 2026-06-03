---
title: 逃逸分析
---

## 1. 栈 vs 堆分配

**栈分配**：函数返回后不被引用、编译期确定大小、分配/释放几乎零成本（移动 SP 指针）。

**堆分配**：函数返回后仍被引用、大小编译期不确定、需要 GC 回收。性能差距：栈分配 ~1ns，堆分配 ~50-100ns（mallocgc + GC 开销）。如果每秒百万次堆分配，差距巨大。

::: info 生产影响
热路径上的堆分配是性能瓶颈的常见原因。

::: warning 面试追问
逃逸分析是在编译期还是运行期？→ 编译期，由编译器决定。

```go
// 栈分配 (~1ns): 函数返回后不再使用
x := 42
arr := [100]int{}
p := Point{1, 2}

// 堆分配 (~50-100ns): 函数返回后仍被引用
func newPoint() *Point {
    p := Point{1, 2}
    return &p  // p 逃逸到堆上
}

// 查看: go build -gcflags="-m"
// main.go:3:6: moved to heap: p
```

## 2. 逃逸场景

7 种常见逃逸场景：1) **返回局部变量指针**（最常见）；2) **赋值给 interface**（fmt.Println 参数、any 类型）；3) **闭包捕获变量**（变量生命周期延长）；4) **send 到 channel**（值被其他 goroutine 使用）；5) **slice/map 存储指针**（指向的数据逃逸）；6) **reflect 使用**（编译期无法确定类型）；7) **fmt.Printf 等可变参数**（...any 触发装箱）。

::: tip 生产建议
热路径避免 fmt.Sprintf，用 strconv；避免不必要的指针返回。

```go
// 1. 返回指针
func f() *int { x := 42; return &x }  // x 逃逸

// 2. interface（fmt 系列都会导致逃逸）
fmt.Sprintf("value: %d", x)  // x 逃逸到堆

// 3. 闭包
func counter() func() int {
    n := 0  // n 逃逸
    return func() int { n++; return n }
}

// 4. channel
ch <- data  // data 逃逸

// 5. fmt 可变参数
log.Printf("msg: %s", msg)  // msg 逃逸

// 查看逃逸:
// go build -gcflags="-m"        # 基本
// go build -gcflags="-m -m"     # 详细
```

## 3. 避免逃逸的技巧

1) 小结构体用**值传递**（<=64 bytes 复制比堆分配快）；2) **预分配 slice** 容量（避免 append 触发扩容时的复制）；3) 用 **strconv** 代替 fmt.Sprintf（不触发 interface 装箱）；4) 热路径用**具体类型**代替 interface（避免动态分发和逃逸）；5) **sync.Pool** 复用对象；6) **小函数有利于内联**（内联后编译器能做更好的分析）。

::: tip 使用场景
JSON 序列化中避免 fmt → 用 strconv；高频日志中避免 fmt.Sprintf → 用 strings.Builder。

```go
// 1. strconv 代替 fmt（减少逃逸）
s := strconv.Itoa(42)       // ✓ 零分配
s := fmt.Sprintf("%d", 42)  // ✗ 有分配

// 2. 预分配 slice
s := make([]int, 0, n)  // 避免 append 扩容

// 3. sync.Pool 复用
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}
buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
defer bufPool.Put(buf)

// 4. 值传递小结构体
type Point struct{ X, Y float64 }
func (p Point) Distance() float64 { // 值接收者
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// 5. 避免 interface 导致的逃逸
func process(data []byte) error { // 具体 type
    // 而非 func process(r io.Reader)
}
```
