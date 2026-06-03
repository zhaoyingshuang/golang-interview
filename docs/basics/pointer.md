---
title: Pointer 指针
---

## 1. 值类型 vs 引用语义类型

**值类型**（int/bool/string/array/struct）：赋值复制整个值，修改副本不影响原始值。

**引用语义类型**（slice/map/channel/pointer/interface/function）：本质也是值传递，但内部包含指针，复制的是头部（24-16 字节），底层数据共享。

**Go 中一切皆值传递，没有引用传递**。传指针也是值传递——复制的是地址值。

::: tip 使用场景
理解这个区别才能正确判断函数调用是否会修改原始数据。

::: warning 面试追问
string 是值类型还是引用类型？→ 值类型（赋值复制 header），但底层有指针（数据共享）。

```go
// 值类型: 互不影响
p1 := Point{1, 2}
p2 := p1    // 复制整个 struct
p2.X = 10   // p1.X 仍是 1

// slice: 共享底层数组
s1 := []int{1, 2, 3}
s2 := s1    // 只复制 header（指针+len+cap）
s2[0] = 100 // s1[0] 也变成 100

// 面试: 以下哪种是引用传递？
// func f(s []int)   → 值传递（复制 slice header）
// func f(m map[string]int) → 值传递（复制 map 指针）
// func f(p *int)    → 值传递（复制指针值）
// Go 中没有引用传递！
```

## 2. new vs make

**new(T)**：分配零值内存，返回 *T。分配在堆上，会被 GC 管理。
:::

**make(T)**：只用于 slice/map/channel，返回初始化后的 T（不是指针）。

**为什么 make 不返回指针**：这三个类型本身就是引用语义的，内部已经包含指针，不需要再包一层。

::: tip 生产建议
new 在实际开发中很少使用，&MyStruct{} 比 new(MyStruct) 更清晰。make 是必须的——slice/map/channel 不 make 就用会 panic。

::: warning 面试追问
make([]int, 0) 和 make([]int, 0, 0) 的区别？→ 没区别，第三个参数是 cap。

```go
p := new(int)        // *int, 值为 0（零值）
cfg := new(Config)  // *Config, 所有字段零值

// new 很少用，以下等价且更清晰：
cfg := &Config{}

s := make([]int, 0, 10)   // slice（返回 []int，不是 *[]int）
m := make(map[string]int) // map（返回 map[string]int）
ch := make(chan int, 5)   // channel（返回 chan int）

// 面试: 以下会 panic 吗？
var s []int
s[0] = 1  // panic！nil slice 不能写入
// 需要先 make
s = make([]int, 1)
s[0] = 1  // OK
```

## 3. 何时用指针

**用指针**：1) 需要修改原始值（最常见）；2) 结构体较大（> 64 bytes，避免复制开销）；3) 一致性（某个方法需要指针，其他也统一用指针）。
:::

**用值**：1) 小结构体（<= 64 bytes，复制比指针解引用+逃逸更快）；2) 不可变数据（值传递天然安全）；3) 需要 map key 或做比较。

**64 bytes 分界线的理由**：缓存行通常 64 bytes，小结构体复制在缓存行内完成，比指针追快。

::: tip 使用场景
方法的接收者选择是团队规范的重要部分。

::: warning 面试追问
如果 receiver 是指针，赋给接口时必须用指针？→ 是的。

```go
// 大结构体 → 指针（避免复制 1KB）
type BigConfig struct {
    Data [1024]byte
    Name string
}
func process(b *BigConfig) { b.Name = "updated" }

// 小结构体 → 值（复制更快，不逃逸）
type Point struct{ X, Y float64 }
func distance(p Point) float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// 一致性原则: 如果一个方法用指针，全部用指针
type User struct { Name string }
func (u *User) SetName(n string) { u.Name = n }  // 需要修改
func (u *User) GetName() string { return u.Name } // 也用指针（一致性）

// 面试: 以下哪种更好？
// func (u User) GetName() string   // 小 struct，值接收者也行
// func (u *User) GetName() string  // 但如果 SetName 用指针，统一用指针
```
