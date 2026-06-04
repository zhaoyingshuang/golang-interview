---
title: Go 标准库算法
---

## 1. sort 包

Go 1.19+ 底层使用 **pdqsort**，对已排序/逆序/重复元素都有特殊优化。

```go
// 不稳定排序
sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })

// 稳定排序 (保持相等元素的原始顺序)
sort.SliceStable(people, func(i, j int) bool { return people[i].Age < people[j].Age })

// 自定义排序 (实现 sort.Interface)
type ByAge []Person
func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
sort.Sort(ByAge(people))
```

## 2. 二分查找

```go
// 查找第一个 >= target 的位置
idx := sort.SearchInts(arr, target)

// 通用二分
idx := sort.Search(len(arr), func(i int) bool { return arr[i] >= target })
```

::: tip 使用模式
`sort.Search(n, func(i int) bool { return 条件(i) })` 返回第一个使条件为 true 的索引。都不满足返回 n。
:::

## 3. container 包

- **heap**：实现 `heap.Interface`（`Len/Less/Swap/Push/Pop`）即可。用于 TopK、合并 K 有序链表、Dijkstra。
- **list**：双向链表。
- **ring**：环形链表。

## 4. strings/bytes

- **strings.Builder**：高效字符串拼接，零拷贝 `String()`。比 `+` 拼接快很多。
- **strings.Count/Contains/Index/Split/Join**
- **bytes.Buffer**：可读写的字节缓冲，`String()` 会拷贝。

## 5. 位运算技巧

| 技巧 | 实现 |
|------|------|
| 判断奇偶 | `n & 1` |
| 乘/除 2 | `n << 1` / `n >> 1` |
| 交换两数 | `a^=b; b^=a; a^=b` |
| 只出现一次 | 全部异或 |
| 判断 2 的幂 | `n > 0 && n&(n-1) == 0` |

::: warning 面试追问
Go sort 底层是什么算法？→ pdqsort (Pattern-Defeating Quicksort)，Go 1.19+ 引入。
:::
