package main

import (
	"container/heap"
	"fmt"
)

// ============================================================
// 树与二叉树
// ============================================================

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func bstInsert(root *TreeNode, val int) *TreeNode {
	if root == nil {
		return &TreeNode{val, nil, nil}
	}
	if val < root.Val {
		root.Left = bstInsert(root.Left, val)
	} else if val > root.Val {
		root.Right = bstInsert(root.Right, val)
	}
	return root
}

func main() {
	traversals()
	bstDemo()
	heapDemo()
	treeSerialization()
}

// ----------------------------------------------------------
// 1. 二叉树遍历
// ----------------------------------------------------------
func traversals() {
	fmt.Println("=== 1. 二叉树遍历 ===")

	//     1
	//    / \
	//   2   3
	//  / \
	// 4   5
	root := &TreeNode{1,
		&TreeNode{2, &TreeNode{4, nil, nil}, &TreeNode{5, nil, nil}},
		&TreeNode{3, nil, nil},
	}

	// 前序遍历 (根左右)
	var preorder func(*TreeNode)
	preorder = func(n *TreeNode) {
		if n == nil {
			return
		}
		fmt.Printf("%d ", n.Val)
		preorder(n.Left)
		preorder(n.Right)
	}
	fmt.Print("  前序: ")
	preorder(root)
	fmt.Println()

	// 中序遍历 (左根右)
	var inorder func(*TreeNode)
	inorder = func(n *TreeNode) {
		if n == nil {
			return
		}
		inorder(n.Left)
		fmt.Printf("%d ", n.Val)
		inorder(n.Right)
	}
	fmt.Print("  中序: ")
	inorder(root)
	fmt.Println()

	// 后序遍历 (左右根)
	var postorder func(*TreeNode)
	postorder = func(n *TreeNode) {
		if n == nil {
			return
		}
		postorder(n.Left)
		postorder(n.Right)
		fmt.Printf("%d ", n.Val)
	}
	fmt.Print("  后序: ")
	postorder(root)
	fmt.Println()

	// 层序遍历 (BFS)
	fmt.Print("  层序: ")
	queue := []*TreeNode{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		fmt.Printf("%d ", node.Val)
		if node.Left != nil {
			queue = append(queue, node.Left)
		}
		if node.Right != nil {
			queue = append(queue, node.Right)
		}
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("  非递归中序遍历 (用栈):")
	fmt.Println("    stack := []*TreeNode{}")
	fmt.Println("    for curr != nil || len(stack) > 0 {")
	fmt.Println("      for curr != nil { stack = append(stack, curr); curr = curr.Left }")
	fmt.Println("      curr = stack[len(stack)-1]; stack = stack[:len(stack)-1]")
	fmt.Println("      visit(curr); curr = curr.Right")
	fmt.Println("    }")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 二叉搜索树
// ----------------------------------------------------------
func bstDemo() {
	fmt.Println("=== 2. 二叉搜索树 (BST) ===")
	fmt.Println()
	fmt.Println("  性质: 左 < 根 < 右")
	fmt.Println("  中序遍历 = 有序序列")
	fmt.Println()
	fmt.Println("  查找: O(log n) 平均, O(n) 最坏(退化为链表)")
	fmt.Println("  插入: 找到位置，挂上新节点")
	fmt.Println("  删除:")
	fmt.Println("    - 叶节点: 直接删除")
	fmt.Println("    - 一个子节点: 子节点替代")
	fmt.Println("    - 两个子节点: 找后继(或前驱)替代")
	fmt.Println()

	var root *TreeNode
	for _, v := range []int{5, 3, 7, 1, 4, 6, 8} {
		root = bstInsert(root, v)
	}
	fmt.Print("  BST 中序: ")
	var inorder func(*TreeNode)
	inorder = func(n *TreeNode) {
		if n == nil {
			return
		}
		inorder(n.Left)
		fmt.Printf("%d ", n.Val)
		inorder(n.Right)
	}
	inorder(root)
	fmt.Println()
	fmt.Println()
	fmt.Println("  面试追问: 为什么 Go map 不用红黑树?")
	fmt.Println("  答: 哈希表平均 O(1) 查找优于红黑树 O(log n)")
	fmt.Println("  Go 选择哈希表 + 优化桶设计，在大多数场景下更快")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 堆与优先队列
// ----------------------------------------------------------
func heapDemo() {
	fmt.Println("=== 3. 堆与优先队列 ===")

	h := &IntHeap{5, 3, 7, 1, 4}
	heap.Init(h)
	heap.Push(h, 2)

	fmt.Print("  弹出最小值: ")
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h).(int))
	}
	fmt.Println()
	fmt.Println("  container/heap 接口:")
	fmt.Println("    type Interface interface {")
	fmt.Println("      sort.Interface")
	fmt.Println("      Push(x any)")
	fmt.Println("      Pop() any")
	fmt.Println("    }")
	fmt.Println()

	fmt.Println("  TopK 问题 (小顶堆维护前K大):")
	nums := []int{3, 2, 1, 5, 6, 4}
	k := 2
	minHeap := &IntHeap{}
	for _, n := range nums {
		heap.Push(minHeap, n)
		if minHeap.Len() > k {
			heap.Pop(minHeap)
		}
	}
	fmt.Printf("  前 %d 大: ", k)
	for minHeap.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(minHeap).(int))
	}
	fmt.Println()
	fmt.Println()
}

// IntHeap implements heap.Interface
type IntHeap []int

// IntHeap 实现 heap.Interface
func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *IntHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *IntHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// ----------------------------------------------------------
// 4. 二叉树序列化
// ----------------------------------------------------------
func treeSerialization() {
	fmt.Println("=== 4. 二叉树序列化 ===")
	fmt.Println()
	fmt.Println("  前序序列化:")
	fmt.Println("    serialize(root):")
	fmt.Println("      if root == nil: return \"#\"")
	fmt.Println("      return str(root.Val) + \",\" + serialize(root.Left) + \",\" + serialize(root.Right)")
	fmt.Println()
	fmt.Println("    deserialize(data):")
	fmt.Println("      vals := strings.Split(data, \",\")")
	fmt.Println("      return buildTree(&vals)")
	fmt.Println()
	fmt.Println("  面试追问: 如何判断两棵树是否相同?")
	fmt.Println("  答: 同时前序遍历，每个节点值都相等")
	fmt.Println()
	fmt.Println("  红黑树核心性质:")
	fmt.Println("    1. 节点是红色或黑色")
	fmt.Println("    2. 根节点是黑色")
	fmt.Println("    3. 叶节点(nil)是黑色")
	fmt.Println("    4. 红色节点的子节点必须是黑色")
	fmt.Println("    5. 从任一节点到其叶节点的路径包含相同数目黑节点")
	fmt.Println("  → 保证最长路径不超过最短路径的 2 倍 → O(log n)")
	fmt.Println()
}
