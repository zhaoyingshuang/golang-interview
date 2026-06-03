---
title: Slice 切片
---

## 1. 底层结构

Slice 底层是 runtime/slice.go 中的 sliceHeader 结构体，只包含三个字段：array（指向底层数组的指针）、len（当前长度）、cap（容量）。Slice 本身只是一个 24 字节的结构体（64位系统），真正的数据存储在底层数组中。

**与 array 的关键区别**：array 的长度是类型的一部分（[3]int 和 [4]int 是不同类型），赋值会复制整个数组；slice 是引用语义类型，赋值只复制 header。

::: tip 使用场景
array 适用于长度固定的场景（如 IPv4 地址 [4]byte、SHA256 哈希 [32]byte）；slice 是 Go 中最常用的数据结构，几乎一切列表都用 slice。创建方式有 4 种：字面量、make、从数组切片、三索引切片。三索引切片 arr[low:high:max] 可限制容量，防止 append 时影响原数组。

```go
type slice struct {
    array unsafe.Pointer // 指向底层数组的指针
    len   int            // 当前长度
    cap   int            // 容量
}

// 创建 slice 的 4 种方式
s1 := []int{1, 2, 3}            // 字面量
s2 := make([]int, 3, 10)        // make，cap=10
s3 := arr[1:3]                  // 从数组切片，cap=4
s5 := arr[1:3:3]                // 三索引切片，cap=2（限制容量）

// 面试追问: slice 占多少内存？
// 24 bytes (64位系统): 指针8 + len8 + cap8
```

## 2. 扩容机制

Go 1.18+ 使用更平滑的增长公式：newCap = oldCap + (oldCap + 3*256) / 4，增长因子从 2.0 平滑过渡到 1.25，不再有 1024 的分界线。最终容量还会做内存对齐（根据元素大小向上取整到合适的内存分配阶级），所以实际 cap 可能比公式计算值更大。

::: info 为什么需要扩容
append 时如果 cap 不够，必须分配新底层数组、复制旧数据、指向新数组。这个过程有 CPU 和内存开销。

::: tip 生产建议
如果知道大概需要多少元素，用 make([]T, 0, n) 预分配。不确定时可以分两步：先收集确定元素，再用 append。

::: warning 面试追问
Go 1.18 前后的扩容策略有什么区别？→ 1.18 前有 1024 分界线，之后用连续平滑公式。

```go
var s []int
for i := 0; i < 20; i++ {
    fmt.Printf("len=%2d cap=%2d\n", len(s), cap(s))
    s = append(s, i)
}
// 输出: 0→1→2→4→4→8→8→16→16→32

// 预分配避免扩容（生产推荐）
s2 := make([]int, 0, 20) // 一次分配，零次扩容

// 面试追问: 扩容时会发生什么？
// 1. 分配新的底层数组
// 2. 复制旧数据到新数组
// 3. slice 指向新数组
// 4. 旧数组等待 GC
```

## 3. 作为函数参数

Slice 作为参数传递时，复制的是 slice header（指针+len+cap，共 24 字节），底层数组不会被复制，所以函数内可以修改元素。但如果函数内触发了扩容，调用方不会看到新元素——因为扩容后 slice 指向了新的底层数组。
:::

**这是最常见的面试坑之一**。

::: tip 使用场景
API handler 中从数据库查询结果 []User 传给 service 层处理，service 可以修改 User 字段，但如果 append 了新元素，handler 看不到。
:::

**解决方案**：1) 返回新 slice（最常用）；2) 传 *[]T 指针（少见）；3) 预分配足够容量（脆弱，不推荐）。

```go
func modifySlice(s []int) {
    s[0] = 100  // 修改底层数组，调用方可见
}

func appendSlice(s []int) {
    s = append(s, 4) // 扩容 → 新底层数组，调用方看不到！
}

// 推荐写法: 返回新 slice
func safeAppend(s []int, v int) []int {
    return append(s, v)
}

// 调用方
s = safeAppend(s, 4) // 接收返回值
```

## 4. 常见陷阱

**陷阱1: 内存泄漏**。大 slice 上取一小片，底层数组仍被引用无法 GC。生产场景：读取大文件后只取 header 部分，但整个文件内容都在内存中。解决：用 copy。

**陷阱2: 共享底层数组**。两个 slice 共享底层数组，一个 append 后另一个可能读到意外数据。生产场景：从请求体解析出多个字段引用同一个大 buffer。

**陷阱3: range 值复制**。for range 中的变量是副本，修改无效。生产场景：批量更新 struct 字段。

::: warning 面试高频
如何检测 slice 内存泄漏？→ pprof heap profile，对比两次快照看 inuse_space 是否持续增长。

```go
// 陷阱1: 内存泄漏（生产常见！）
big := make([]byte, 1<<20) // 1MB
small := big[:10]          // 1MB 都无法被 GC！
// 解决: copy
small2 := make([]byte, 10)
copy(small2, big[:10])

// 陷阱2: 共享底层数组
s1 := make([]int, 3, 6)
s2 := s1[0:3]
s2 = append(s2, 100)  // s1 的 cap 区域被修改

// 陷阱3: range 值复制
for _, item := range items {
    item.Value = 0  // 修改的是副本！无效
}
// 正确: 用索引
for i := range items {
    items[i].Value = 0
}
```

## 5. Slice 技巧

掌握这些技巧在面试和日常开发中都能写出更优雅的代码。
:::

**快速删除**（不保序 O(1)）：把最后一个元素移到被删位置，适用于顺序无关的场景（如从连接池中移除）。

**保序删除** O(n)：用 append 覆盖。

**插入元素**：先 append 腾位再赋值。

**原地过滤**：双指针法，零分配。

**去重**：先排序再用双指针。

::: tip 使用场景
WebSocket 连接管理（快速删除断开的连接）、消息队列消费者（原地过滤已处理的消息）。

```go
// 删除第 i 个元素（不保序 O(1)）- 连接池场景
s[i] = s[len(s)-1]
s = s[:len(s)-1]

// 保序删除 O(n) - 有序列表场景
s = append(s[:i], s[i+1:]...)

// 插入元素到位置 i
s = append(s[:i+1], s[i:]...)
s[i] = v

// 原地过滤（零分配）- 消息处理场景
n := 0
for _, v := range s {
    if keep(v) { s[n] = v; n++ }
}
s = s[:n]

// 去重（先排序）
sort.Ints(s)
j := 0
for i := 1; i < len(s); i++ {
    if s[j] != s[i] { j++; s[j] = s[i] }
}
s = s[:j+1]
```
