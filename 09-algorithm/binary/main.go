package main

import (
	"fmt"
	"sort"
)

// ============================================================
// 二分查找
// ============================================================

func main() {
	basicBinarySearch()
	findFirst()
	findLast()
	findInsertPos()
	variants()
}

// ----------------------------------------------------------
// 1. 基础二分查找
// ----------------------------------------------------------
func basicBinarySearch() {
	fmt.Println("=== 1. 基础二分查找 ===")

	search := func(arr []int, target int) int {
		left, right := 0, len(arr)-1
		for left <= right {
			mid := left + (right-left)/2 // 防溢出
			if arr[mid] == target {
				return mid
			} else if arr[mid] < target {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
		return -1
	}

	arr := []int{1, 3, 5, 7, 9, 11, 13, 15}
	fmt.Printf("  数组: %v\n", arr)
	fmt.Printf("  查找 7 → 索引 %d\n", search(arr, 7))
	fmt.Printf("  查找 8 → %d (未找到)\n", search(arr, 8))
	fmt.Println()
	fmt.Println("  关键点: left <= right (闭区间), mid = left + (right-left)/2")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 查找第一个等于目标值的位置
// ----------------------------------------------------------
func findFirst() {
	fmt.Println("=== 2. 查找第一个等于 ===")

	arr := []int{1, 3, 5, 5, 5, 7, 9}
	target := 5

	left, right := 0, len(arr)-1
	result := -1
	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] >= target {
			if arr[mid] == target {
				result = mid
			}
			right = mid - 1 // 继续向左找
		} else {
			left = mid + 1
		}
	}

	fmt.Printf("  数组: %v\n", arr)
	fmt.Printf("  第一个 %d → 索引 %d\n", target, result)
	fmt.Println()

	// Go 标准库方式
	idx := sort.SearchInts(arr, target)
	fmt.Printf("  sort.SearchInts(%d) → 索引 %d\n", target, idx)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 查找最后一个等于目标值的位置
// ----------------------------------------------------------
func findLast() {
	fmt.Println("=== 3. 查找最后一个等于 ===")

	arr := []int{1, 3, 5, 5, 5, 7, 9}
	target := 5

	left, right := 0, len(arr)-1
	result := -1
	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] <= target {
			if arr[mid] == target {
				result = mid
			}
			left = mid + 1 // 继续向右找
		} else {
			right = mid - 1
		}
	}

	fmt.Printf("  数组: %v\n", arr)
	fmt.Printf("  最后一个 %d → 索引 %d\n", target, result)
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 搜索插入位置
// ----------------------------------------------------------
func findInsertPos() {
	fmt.Println("=== 4. 搜索插入位置 ===")

	searchInsert := func(arr []int, target int) int {
		left, right := 0, len(arr)
		for left < right { // 注意: 左闭右开
			mid := left + (right-left)/2
			if arr[mid] < target {
				left = mid + 1
			} else {
				right = mid // 不减一
			}
		}
		return left
	}

	arr := []int{1, 3, 5, 7, 9}
	fmt.Printf("  数组: %v\n", arr)
	fmt.Printf("  插入 4 → 位置 %d\n", searchInsert(arr, 4))
	fmt.Printf("  插入 0 → 位置 %d\n", searchInsert(arr, 0))
	fmt.Printf("  插入 10 → 位置 %d\n", searchInsert(arr, 10))
	fmt.Println()
	fmt.Println("  Go 标准库: sort.Search(n, func(i int) bool { return arr[i] >= target })")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 变体: 旋转数组 + 求平方根
// ----------------------------------------------------------
func variants() {
	fmt.Println("=== 5. 二分变体 ===")

	// 旋转数组查找
	fmt.Println("旋转数组查找最小值:")
	rotated := []int{4, 5, 6, 7, 0, 1, 2}
	findMin := func(arr []int) int {
		left, right := 0, len(arr)-1
		for left < right {
			mid := left + (right-left)/2
			if arr[mid] > arr[right] {
				left = mid + 1
			} else {
				right = mid
			}
		}
		return arr[left]
	}
	fmt.Printf("  %v → 最小值 %d\n", rotated, findMin(rotated))

	// 浮点数二分求平方根
	fmt.Println("\n浮点数二分求平方根:")
	sqrt := func(x float64) float64 {
		if x < 0 {
			return -1
		}
		left, right := 0.0, x
		if x < 1 {
			right = 1
		}
		for right-left > 1e-8 {
			mid := (left + right) / 2
			if mid*mid < x {
				left = mid
			} else {
				right = mid
			}
		}
		return (left + right) / 2
	}
	fmt.Printf("  sqrt(2) ≈ %.6f\n", sqrt(2))
	fmt.Printf("  sqrt(9) ≈ %.6f\n", sqrt(9))
	fmt.Println()
	fmt.Println("  面试追问: 二分如何避免死循环?")
	fmt.Println("  答: 确保 left/right 在每轮都变化:")
	fmt.Println("    left = mid + 1 或 right = mid - 1 (闭区间)")
	fmt.Println("    left = mid + 1 或 right = mid (左闭右开)")
	fmt.Println()
}
