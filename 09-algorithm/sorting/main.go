package main

import (
	"fmt"
	"sort"
)

// ============================================================
// 排序算法
// ============================================================

func main() {
	bubbleSort()
	insertionSort()
	quickSort()
	mergeSortDemo()
	heapSortDemo()
	goSort()
}

// ----------------------------------------------------------
// 1. 冒泡排序 O(n²)
// ----------------------------------------------------------
func bubbleSort() {
	fmt.Println("=== 1. 冒泡排序 ===")
	arr := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Printf("  原始: %v\n", arr)

	n := len(arr)
	for i := 0; i < n-1; i++ {
		swapped := false
		for j := 0; j < n-1-i; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				swapped = true
			}
		}
		if !swapped {
			break // 优化: 没有交换说明已有序
		}
	}
	fmt.Printf("  结果: %v\n", arr)
	fmt.Println("  复杂度: O(n²), 稳定排序")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 插入排序 O(n²)
// ----------------------------------------------------------
func insertionSort() {
	fmt.Println("=== 2. 插入排序 ===")
	arr := []int{64, 34, 25, 12, 22, 11, 90}

	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}

	fmt.Printf("  结果: %v\n", arr)
	fmt.Println("  复杂度: O(n²), 稳定排序")
	fmt.Println("  优势: 对几乎有序的数据很高效 O(n)")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 快速排序 O(n log n)
// ----------------------------------------------------------
func quickSort() {
	fmt.Println("=== 3. 快速排序 ===")
	arr := []int{64, 34, 25, 12, 22, 11, 90}

	var qsort func([]int)
	qsort = func(a []int) {
		if len(a) < 2 {
			return
		}
		left, right := 0, len(a)-1
		pivot := len(a) / 2
		a[pivot], a[right] = a[right], a[pivot]
		for i := 0; i < right; i++ {
			if a[i] < a[right] {
				a[left], a[i] = a[i], a[left]
				left++
			}
		}
		a[left], a[right] = a[right], a[left]
		qsort(a[:left])
		qsort(a[left+1:])
	}

	qsort(arr)
	fmt.Printf("  结果: %v\n", arr)
	fmt.Println("  复杂度: 平均 O(n log n), 最坏 O(n²)")
	fmt.Println("  最坏场景: 已排序数组 + 首元素作为 pivot")
	fmt.Println("  优化: 三数取中法/随机 pivot")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 归并排序 O(n log n)
// ----------------------------------------------------------
func mergeSortDemo() {
	fmt.Println("=== 4. 归并排序 ===")

	var mergeSort func([]int) []int
	mergeSort = func(a []int) []int {
		if len(a) < 2 {
			return a
		}
		mid := len(a) / 2
		left := mergeSort(a[:mid])
		right := mergeSort(a[mid:])

		result := make([]int, 0, len(left)+len(right))
		i, j := 0, 0
		for i < len(left) && j < len(right) {
			if left[i] <= right[j] {
				result = append(result, left[i])
				i++
			} else {
				result = append(result, right[j])
				j++
			}
		}
		result = append(result, left[i:]...)
		result = append(result, right[j:]...)
		return result
	}

	arr := []int{64, 34, 25, 12, 22, 11, 90}
	sorted := mergeSort(arr)
	fmt.Printf("  结果: %v\n", sorted)
	fmt.Println("  复杂度: O(n log n), 稳定排序")
	fmt.Println("  特点: 需要额外 O(n) 空间")
	fmt.Println("  适用: 大数据排序（外部排序）、链表排序")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 堆排序 O(n log n)
// ----------------------------------------------------------
func heapSortDemo() {
	fmt.Println("=== 5. 堆排序 ===")

	arr := []int{64, 34, 25, 12, 22, 11, 90}
	n := len(arr)

	// 建堆 (从最后一个非叶节点开始)
	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}

	// 逐个提取最大元素
	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heapify(arr, i, 0)
	}

	fmt.Printf("  结果: %v\n", arr)
	fmt.Println("  复杂度: O(n log n), 不稳定排序")
	fmt.Println("  特点: 原地排序，空间 O(1)")
	fmt.Println()
}

func heapify(arr []int, n, i int) {
	largest := i
	left, right := 2*i+1, 2*i+2
	if left < n && arr[left] > arr[largest] {
		largest = left
	}
	if right < n && arr[right] > arr[largest] {
		largest = right
	}
	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, n, largest)
	}
}

// ----------------------------------------------------------
// 6. Go 标准库排序
// ----------------------------------------------------------
func goSort() {
	fmt.Println("=== 6. Go 标准库 sort ===")

	// sort.Slice (不稳定排序)
	nums := []int{64, 34, 25, 12, 22, 11, 90}
	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
	fmt.Printf("  sort.Slice: %v\n", nums)

	// sort.SliceStable (稳定排序)
	type Person struct {
		Name string
		Age  int
	}
	people := []Person{
		{"Alice", 30}, {"Bob", 25}, {"Charlie", 30},
	}
	sort.SliceStable(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Printf("  SliceStable(按年龄): %v\n", people)

	fmt.Println()
	fmt.Println("  Go 1.19+ sort 底层使用 pdqsort (Pattern-Defeating Quicksort):")
	fmt.Println("  - 结合快排/堆排序/插入排序的优点")
	fmt.Println("  - 对常见模式(已排序/逆序/重复元素)做了特殊优化")
	fmt.Println("  - 最坏 O(n log n)，最好 O(n)")
	fmt.Println()
	fmt.Println("  面试追问: 什么时候用稳定排序?")
	fmt.Println("  答: 当需要保持相等元素的原始顺序时")
	fmt.Println("  例: 先按年龄排，再按姓名排 → 第二次排序需稳定")
	fmt.Println()
}
