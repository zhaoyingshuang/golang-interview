---
title: 排序算法
---

## 1. 排序算法概览

| 算法 | 平均 | 最坏 | 空间 | 稳定 |
|------|------|------|------|------|
| 冒泡排序 | O(n²) | O(n²) | O(1) | ✓ |
| 插入排序 | O(n²) | O(n²) | O(1) | ✓ |
| 选择排序 | O(n²) | O(n²) | O(1) | ✗ |
| 快速排序 | O(n log n) | O(n²) | O(log n) | ✗ |
| 归并排序 | O(n log n) | O(n log n) | O(n) | ✓ |
| 堆排序 | O(n log n) | O(n log n) | O(1) | ✗ |

::: tip 稳定排序的应用
需要保持相等元素的原始顺序时（如先按年龄排，再按姓名排 → 第二次排序需稳定）。
:::

## 2. O(n²) 排序

**冒泡排序**：相邻元素比较交换。优化：一轮无交换则已有序。

**插入排序**：将元素插入已排序部分的正确位置。对几乎有序的数据很高效 O(n)。

**选择排序**：每轮选出最小元素放到前面。不稳定。

## 3. O(n log n) 排序

**快速排序**：选 pivot 分区，递归排序。最坏 O(n²)（已排序数组 + 首元素 pivot）。优化：三数取中/随机 pivot。

```go
func quickSort(arr []int) {
    if len(arr) < 2 { return }
    left, right := 0, len(arr)-1
    pivot := len(arr) / 2
    arr[pivot], arr[right] = arr[right], arr[pivot]
    for i := 0; i < right; i++ {
        if arr[i] < arr[right] {
            arr[left], arr[i] = arr[i], arr[left]
            left++
        }
    }
    arr[left], arr[right] = arr[right], arr[left]
    quickSort(arr[:left])
    quickSort(arr[left+1:])
}
```

**归并排序**：分治 + 合并。稳定，适合链表排序和外部排序。需要额外 O(n) 空间。

## 4. Go 标准库 sort

Go 1.19+ 使用 **pdqsort** (Pattern-Defeating Quicksort)：
- 已排序 → O(n)
- 逆序 → O(n)
- 重复元素多 → O(n)
- 最坏 → O(n log n)（退化为堆排序）

```go
sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
sort.SliceStable(people, func(i, j int) bool { return people[i].Age < people[j].Age })
```

::: warning 面试追问
Go sort 底层是什么算法？→ pdqsort（Go 1.19+），结合快排/堆排序/插入排序的优点。
:::
