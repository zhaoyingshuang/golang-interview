---
title: 树与二叉树
---

## 1. 二叉树遍历

```go
// 前序 (根左右)
func preorder(n *TreeNode) {
    if n == nil { return }
    visit(n)
    preorder(n.Left)
    preorder(n.Right)
}

// 中序 (左根右) — BST 中序 = 有序序列
func inorder(n *TreeNode) { /* Left → Visit → Right */ }

// 后序 (左右根) — 用于释放资源
func postorder(n *TreeNode) { /* Left → Right → Visit */ }

// 层序 (BFS) — 用队列
queue := []*TreeNode{root}
for len(queue) > 0 {
    node := queue[0]; queue = queue[1:]
    visit(node)
    if node.Left != nil { queue = append(queue, node.Left) }
    if node.Right != nil { queue = append(queue, node.Right) }
}
```

::: tip 非递归中序遍历
用栈模拟递归：沿左子树压栈 → 弹出访问 → 转右子树。
:::

## 2. 二叉搜索树 (BST)

性质：左 < 根 < 右。中序遍历 = 有序序列。

- 查找/插入：O(log n) 平均，O(n) 最坏（退化为链表）
- 删除：叶节点直接删；一个子节点用子节点替代；两个子节点找后继替代

## 3. 红黑树

5 条性质保证最长路径不超过最短路径的 2 倍 → O(log n)。

1. 节点是红色或黑色
2. 根节点是黑色
3. 叶节点(nil)是黑色
4. 红色节点的子节点必须是黑色
5. 从任一节点到叶节点的路径包含相同数目黑节点

::: warning 面试追问：为什么 Go map 不用红黑树
哈希表平均 O(1) 查找优于红黑树 O(log n)。Go 选择哈希表 + 优化桶设计。
Java HashMap 在链表长度 > 8 时转红黑树，Go 没有这个优化。
:::

## 4. 堆与优先队列

Go `container/heap` 实现最小堆。接口：`Len/Less/Swap/Push/Pop`。

```go
// TopK 问题：用小顶堆维护前 K 大元素
minHeap := &IntHeap{}
for _, n := range nums {
    heap.Push(minHeap, n)
    if minHeap.Len() > k { heap.Pop(minHeap) }
}
```
