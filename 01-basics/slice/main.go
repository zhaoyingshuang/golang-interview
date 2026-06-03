package main

import "fmt"

// ============================================================
// Slice 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. slice 的底层结构是什么？
// 2. slice 扩容机制是怎样的？
// 3. slice 和 array 的区别？
// 4. slice 作为函数参数会发生什么？
// 5. nil slice 和 empty slice 的区别？
// 6. slice 的 copy 语义是怎样的？

func main() {
	structure()
	capacityAndGrowth()
	sliceAsParameter()
	nilVsEmpty()
	sliceTraps()
	sliceTricks()
}

// ----------------------------------------------------------
// 1. 底层结构
// ----------------------------------------------------------
// runtime/slice.go 中的 sliceHeader:
//
//   type slice struct {
//       array unsafe.Pointer // 指向底层数组的指针
//       len   int            // 当前长度
//       cap   int            // 容量
//   }
//
// slice 本身只是一个包含了指针+长度+容量的结构体（24字节 on 64bit）。
// 真正的数据存储在底层数组中，多个 slice 可以共享同一个底层数组。
//
// 与 array 的关键区别:
//   - array: [3]int 是固定长度，类型的一部分，值类型（赋值会复制整个数组）
//   - slice: []int 是动态长度，引用类型（赋值只复制 slice header）
func structure() {
	fmt.Println("=== 1. 底层结构 ===")

	// 创建 slice 的 4 种方式
	// 方式1: 从字面量创建
	s1 := []int{1, 2, 3}

	// 方式2: make([]T, len, cap)
	s2 := make([]int, 3, 10)

	// 方式3: 从数组切片 arr[low:high:max]
	arr := [5]int{1, 2, 3, 4, 5}
	s3 := arr[1:3] // len=2, cap=4 (从 low 到数组末尾)
	// 注意: s3 和 arr 共享底层数组！修改 s3 会影响 arr

	// 方式4: new — 注意 new 只分配内存，返回 *[]T，len/cap 都是 0
	s4 := new([]int)
	_ = s4 // 实际开发中几乎不用这种方式

	fmt.Printf("s1: len=%d cap=%d\n", len(s1), cap(s1))
	fmt.Printf("s2: len=%d cap=%d\n", len(s2), cap(s2))
	fmt.Printf("s3: len=%d cap=%d, 共享底层数组\n", len(s3), cap(s3))

	// 三索引切片限制容量: arr[low:high:max] → cap = max - low
	s5 := arr[1:3:3] // len=2, cap=2 — 限制了容量，append 时会分配新数组
	fmt.Printf("s5: len=%d cap=%d (三索引切片限制容量)\n", len(s5), cap(s5))
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 扩容机制（Go 1.21+ 之后的实现）
// ----------------------------------------------------------
// runtime/slice.go → growslice 函数
//
// Go 1.18 之前:
//   - 旧容量 < 1024 → 新容量 = 旧容量 * 2
//   - 旧容量 >= 1024 → 新容量 = 旧容量 * 1.25（循环直到 >= 所需容量）
//
// Go 1.18+:
//   - 不再有 1024 的分界线
//   - 使用更平滑的增长公式: newCap = oldCap + (oldCap + 3*256) / 4
//   - 增长因子从 2.0 平滑过渡到 1.25
//
// 最终容量还会做内存对齐（根据元素大小向上取整到合适的内存分配阶级）。
// 所以实际观察到的 cap 可能比公式计算值更大。
func capacityAndGrowth() {
	fmt.Println("=== 2. 扩容机制 ===")

	// 观察扩容过程
	var s []int
	for i := 0; i < 20; i++ {
		fmt.Printf("len=%2d cap=%2d\n", len(s), cap(s))
		s = append(s, i)
	}
	// 你会看到: 0→1→2→4→4→8→8→16→16→32...

	// 面试题: 预分配容量避免扩容
	// 如果知道大概需要多少元素，应该用 make 预分配
	s2 := make([]int, 0, 20) // 一次分配，零次扩容
	fmt.Printf("\n预分配: len=%d cap=%d\n", len(s2), cap(s2))
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Slice 作为函数参数
// ----------------------------------------------------------
// slice 作为参数传递时，复制的是 slice header（指针+len+cap）。
// 底层数组不会被复制，所以函数内可以修改元素。
// 但是如果函数内触发了扩容，调用方不会看到新元素。
func sliceAsParameter() {
	fmt.Println("=== 3. Slice 作为函数参数 ===")

	s := []int{1, 2, 3}
	fmt.Printf("调用前: %v, len=%d, cap=%d\n", s, len(s), cap(s))

	modifySlice(s)
	fmt.Printf("modifySlice 后: %v (元素被修改了)\n", s)

	appendSlice(s)
	fmt.Printf("appendSlice 后: %v (append 不可见！cap=%d)\n", s, cap(s))

	// 如果需要函数内的 append 对调用方可见，有几种方式:
	// 1. 返回新的 slice: s = appendInFunc(s)
	// 2. 传指针: func f(s *[]int)
	// 3. 预分配足够容量，保证不扩容（不推荐，脆弱）
	fmt.Println()
}

func modifySlice(s []int) {
	s[0] = 100 // 修改底层数组，调用方可见
}

func appendSlice(s []int) {
	s = append(s, 4) // 扩容后 s 指向新底层数组，调用方看不到
}

// ----------------------------------------------------------
// 4. nil slice vs empty slice
// ----------------------------------------------------------
// nil slice:   var s []int          → s == nil, len=0, cap=0
// empty slice: s := []int{} 或 s := make([]int, 0)
//              → s != nil, len=0, cap=0
//
// 在使用上它们几乎等价（append/len/cap/range 都正常工作）。
// 区别在于 JSON 编码: nil slice → null, empty slice → []
// 以及 reflect.DeepEqual: nil != empty
func nilVsEmpty() {
	fmt.Println("=== 4. nil slice vs empty slice ===")

	var s1 []int          // nil slice
	s2 := []int{}         // empty slice
	s3 := make([]int, 0)  // empty slice

	fmt.Printf("nil slice:   value=%v, nil=%v, len=%d, cap=%d\n", s1, s1 == nil, len(s1), cap(s1))
	fmt.Printf("empty slice: value=%v, nil=%v, len=%d, cap=%d\n", s2, s2 == nil, len(s2), cap(s2))
	fmt.Printf("make slice:  value=%v, nil=%v, len=%d, cap=%d\n", s3, s3 == nil, len(s3), cap(s3))

	// 两者都可以正常 append
	s1 = append(s1, 1)
	s2 = append(s2, 1)
	fmt.Printf("append 后: nil=%v, nil=%v\n", s1, s2)
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 常见陷阱
// ----------------------------------------------------------
func sliceTraps() {
	fmt.Println("=== 5. 常见陷阱 ===")

	// 陷阱1: slice 内存泄漏
	// 场景: 大 slice 上取一小片，但大 slice 无法被 GC
	{
		// 假设我们有一个很大的 slice，只引用了其中一小部分
		big := make([]byte, 1<<20) // 1MB
		small := big[:10]          // 只需要前10字节

		// 此时 big 的底层数组不会被 GC，因为 small 还在引用它
		_ = small

		// 解决: 用 copy
		small2 := make([]byte, 10)
		copy(small2, big[:10])
		_ = small2
		fmt.Println("陷阱1: 大 slice 上取小片会阻止 GC → 用 copy 解决")
	}

	// 陷阱2: append 共享底层数组
	{
		s1 := make([]int, 3, 6)
		s2 := s1[0:3]

		s2 = append(s2, 100)
		// 如果没有触发扩容，s1 和 s2 共享底层数组
		// 但 s1 的 len 仍然是 3，所以看不到 append 的 100
		// 然而 s2 的第 4 个位置已经被修改了！
		// 如果后续有操作扩展 s1，就会看到意外的值
		fmt.Println("陷阱2: 共享底层数组的 slice，append 可能互相影响")
	}

	// 陷阱3: range 中的值复制
	{
		type Item struct {
			Value int
		}
		items := []Item{{1}, {2}, {3}}
		for _, item := range items {
			item.Value = 0 // 这修改的是副本！
		}
		fmt.Printf("陷阱3: range 值复制: %v (修改无效)\n", items)

		// 正确方式: 用索引
		for i := range items {
			items[i].Value = 0
		}
		fmt.Printf("正确方式: %v\n", items)
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 6. Slice 技巧
// ----------------------------------------------------------
func sliceTricks() {
	fmt.Println("=== 6. Slice 技巧 ===")

	// 删除第 i 个元素（不保序）
	removeFast := func(s []int, i int) []int {
		s[i] = s[len(s)-1]
		return s[:len(s)-1]
	}
	s := []int{1, 2, 3, 4, 5}
	s = removeFast(s, 1) // 删除 index=1
	fmt.Printf("快速删除(不保序): %v\n", s)

	// 删除第 i 个元素（保序）
	remove := func(s []int, i int) []int {
		return append(s[:i], s[i+1:]...)
	}
	s2 := []int{1, 2, 3, 4, 5}
	s2 = remove(s2, 1)
	fmt.Printf("保序删除: %v\n", s2)

	// 插入元素到位置 i
	insert := func(s []int, i int, v int) []int {
		s = append(s[:i+1], s[i:]...) // 向后腾出一个位置
		s[i] = v
		return s
	}
	s3 := []int{1, 2, 4, 5}
	s3 = insert(s3, 2, 3)
	fmt.Printf("插入: %v\n", s3)

	// 去重（需要先排序）
	dedup := func(s []int) []int {
		j := 0
		for i := 1; i < len(s); i++ {
			if s[j] != s[i] {
				j++
				s[j] = s[i]
			}
		}
		return s[:j+1]
	}
	s4 := []int{1, 1, 2, 3, 3, 3, 4}
	s4 = dedup(s4)
	fmt.Printf("去重: %v\n", s4)

	// 过滤（原地）
	filter := func(s []int, keep func(int) bool) []int {
		n := 0
		for _, v := range s {
			if keep(v) {
				s[n] = v
				n++
			}
		}
		return s[:n]
	}
	s5 := []int{1, 2, 3, 4, 5, 6}
	s5 = filter(s5, func(v int) bool { return v%2 == 0 })
	fmt.Printf("过滤偶数: %v\n", s5)
}
