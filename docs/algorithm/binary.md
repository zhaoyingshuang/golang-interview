---
title: 二分查找
---

## 1. 基础二分查找

```go
func search(arr []int, target int) int {
    left, right := 0, len(arr)-1
    for left <= right {  // 注意: <= (闭区间)
        mid := left + (right-left)/2  // 防溢出
        if arr[mid] == target { return mid }
        if arr[mid] < target { left = mid + 1 }
        if arr[mid] > target { right = mid - 1 }
    }
    return -1
}
```

::: tip 关键点
- `left <= right`（闭区间）
- `mid = left + (right-left)/2`（防止 int 溢出）
- `left = mid + 1` / `right = mid - 1`（确保每轮都在缩小范围）
:::

## 2. 变体二分

**查找第一个等于**：找到后继续向左搜索 `right = mid - 1`

**查找最后一个等于**：找到后继续向右搜索 `left = mid + 1`

**查找第一个大于等于**：`arr[mid] >= target` 时记录并右缩

**查找最后一个小于等于**：`arr[mid] <= target` 时记录并左扩

```go
// Go 标准库: 查找第一个 >= target 的位置
idx := sort.SearchInts(arr, target)
// 或通用:
idx := sort.Search(len(arr), func(i int) bool { return arr[i] >= target })
```

## 3. 浮点数二分

```go
sqrt := func(x float64) float64 {
    left, right := 0.0, x
    if x < 1 { right = 1 }
    for right-left > 1e-8 {
        mid := (left + right) / 2
        if mid*mid < x { left = mid } else { right = mid }
    }
    return (left + right) / 2
}
```

## 4. 实际应用

- **旋转数组查找最小值**：`arr[mid] > arr[right]` → 最小值在右半
- **搜索插入位置**：左闭右开 `[left, right)`
- **求平方根**：浮点数二分 / 牛顿迭代

::: warning 面试追问：二分如何避免死循环
确保 left/right 在每轮都变化：闭区间用 `mid ± 1`，左闭右开用 `left = mid + 1` / `right = mid`。
:::
