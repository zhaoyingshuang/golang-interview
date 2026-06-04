package main

import (
	"fmt"
	"sort"
)

// ============================================================
// Go 标准库算法
// ============================================================

func main() {
	sortDemo()
	searchDemo()
	heapDemo()
	stringAlgo()
	bitOps()
}

// ----------------------------------------------------------
// 1. sort 包
// ----------------------------------------------------------
func sortDemo() {
	fmt.Println("=== 1. sort 包 ===")

	// sort.Slice (不稳定, 底层 pdqsort)
	nums := []int{5, 3, 1, 4, 2}
	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
	fmt.Printf("  sort.Slice: %v\n", nums)

	// sort.SliceStable (稳定排序)
	type Student struct {
		Name  string
		Score int
	}
	students := []Student{
		{"Alice", 90}, {"Bob", 85}, {"Charlie", 90}, {"David", 85},
	}
	sort.SliceStable(students, func(i, j int) bool {
		return students[i].Score > students[j].Score
	})
	fmt.Printf("  SliceStable(按分数降序): %v\n", students)

	// 自定义排序 (实现 sort.Interface)
	fmt.Println()
	fmt.Println("  sort.Interface 接口:")
	fmt.Println("    type Interface interface {")
	fmt.Println("      Len() int")
	fmt.Println("      Less(i, j int) bool")
	fmt.Println("      Swap(i, j int)")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("  Go 1.21+ 使用 pdqsort (Pattern-Defeating Quicksort):")
	fmt.Println("    已排序 → O(n)")
	fmt.Println("    逆序 → O(n)")
	fmt.Println("    重复元素多 → O(n)")
	fmt.Println("    最坏 → O(n log n) (退化为堆排序)")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 二分查找
// ----------------------------------------------------------
func searchDemo() {
	fmt.Println("=== 2. sort.Search 二分查找 ===")

	arr := []int{1, 3, 5, 7, 9, 11, 13}

	// sort.SearchInts — 查找第一个 >= target 的位置
	idx := sort.SearchInts(arr, 7)
	fmt.Printf("  SearchInts(7) → 索引 %d\n", idx)

	// sort.Search — 通用二分查找
	target := 7
	pos := sort.Search(len(arr), func(i int) bool {
		return arr[i] >= target
	})
	fmt.Printf("  Search(>=7) → 索引 %d, 值=%d\n", pos, arr[pos])

	// 自定义: 查找满足条件的第一个位置
	fmt.Println()
	fmt.Println("  sort.Search 使用模式:")
	fmt.Println("    sort.Search(n, func(i int) bool { return 条件(i) })")
	fmt.Println("    返回第一个使条件为 true 的索引")
	fmt.Println("    如果都不满足 → 返回 n")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. container/heap
// ----------------------------------------------------------
func heapDemo() {
	fmt.Println("=== 3. container/heap ===")
	fmt.Println()
	fmt.Println("  优先队列 (最小堆):")
	fmt.Println("    h := &IntHeap{5, 3, 7, 1}")
	fmt.Println("    heap.Init(h)")
	fmt.Println("    heap.Push(h, 2)")
	fmt.Println("    min := heap.Pop(h) // 1")
	fmt.Println()
	fmt.Println("  常见应用:")
	fmt.Println("    TopK 问题     — 小顶堆维护前 K 大")
	fmt.Println("    合并 K 有序链表 — 每次取最小节点")
	fmt.Println("    任务调度      — 按优先级执行")
	fmt.Println("    Dijkstra 最短路 — 取当前最短距离")
	fmt.Println()
	fmt.Println("  container/list — 双向链表")
	fmt.Println("  container/ring — 环形链表")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 字符串算法
// ----------------------------------------------------------
func stringAlgo() {
	fmt.Println("=== 4. strings 包常用操作 ===")
	fmt.Println()
	fmt.Println("  strings.Builder — 高效字符串拼接")
	fmt.Println("    var b strings.Builder")
	fmt.Println("    for _, s := range parts { b.WriteString(s) }")
	fmt.Println("    result := b.String()")
	fmt.Println("    比 + 拼接快: 底层用 []byte，避免多次分配复制")
	fmt.Println()
	fmt.Println("  strings.Count / Contains / Index")
	fmt.Println("  strings.Split / Join")
	fmt.Println("  strings.TrimSpace / Trim / TrimPrefix")
	fmt.Println()
	fmt.Println("  strings.Builder vs bytes.Buffer:")
	fmt.Println("    Builder: 不可复制(用后不能再用), String() 零拷贝")
	fmt.Println("    Buffer:  可复制, String() 会拷贝")
	fmt.Println("    推荐: 纯字符串拼接用 Builder")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 位运算技巧
// ----------------------------------------------------------
func bitOps() {
	fmt.Println("=== 5. 位运算技巧 ===")

	fmt.Println("  判断奇偶:")
	fmt.Printf("    7 & 1 = %d (奇数)\n", 7&1)
	fmt.Printf("    8 & 1 = %d (偶数)\n", 8&1)

	fmt.Println("\n  乘除 2 的幂:")
	fmt.Printf("    5 << 1 = %d (×2)\n", 5<<1)
	fmt.Printf("    5 >> 1 = %d (÷2)\n", 5>>1)

	fmt.Println("\n  交换两数 (不用临时变量):")
	fmt.Println("    a ^= b; b ^= a; a ^= b")

	fmt.Println("\n  只出现一次的数字 (其他出现两次):")
	nums := []int{2, 3, 2, 5, 3}
	xor := 0
	for _, n := range nums {
		xor ^= n
	}
	fmt.Printf("    %v → %d\n", nums, xor)

	fmt.Println("\n  判断 2 的幂:")
	fmt.Printf("    8 & 7 = %d (是2的幂)\n", 8&7)
	fmt.Printf("    6 & 5 = %d (不是)\n", 6&5)

	fmt.Println()
	fmt.Println("  面试追问: 如何找到只出现一次的两个数?")
	fmt.Println("  答: 先全部异或得到 xor=x^y, 取 xor 最低位 1 (区分位),")
	fmt.Println("     将所有数按该位分组，每组分别异或")
	fmt.Println()
}
