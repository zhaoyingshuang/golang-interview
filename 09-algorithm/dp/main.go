package main

import "fmt"

// ============================================================
// 动态规划
// ============================================================

func main() {
	fibonacci()
	climbStairs()
	lis()
	maxSubArray()
	knapsack()
	lcs()
}

// ----------------------------------------------------------
// 1. 斐波那契数列
// ----------------------------------------------------------
func fibonacci() {
	fmt.Println("=== 1. 斐波那契数列 ===")

	// 递归 (指数级，不推荐)
	_ = func(n int) int {
		if n <= 1 {
			return n
		}
		return 0 // fib(n-1) + fib(n-2) — 省略实际递归
	}

	// DP — 自底向上
	fib := func(n int) int {
		if n <= 1 {
			return n
		}
		dp := make([]int, n+1)
		dp[0], dp[1] = 0, 1
		for i := 2; i <= n; i++ {
			dp[i] = dp[i-1] + dp[i-2]
		}
		return dp[n]
	}

	// 空间优化: 只用两个变量
	fibOpt := func(n int) int {
		if n <= 1 {
			return n
		}
		prev, curr := 0, 1
		for i := 2; i <= n; i++ {
			prev, curr = curr, prev+curr
		}
		return curr
	}

	fmt.Printf("  fib(10) = %d\n", fib(10))
	fmt.Printf("  fib(10) = %d (空间优化)\n", fibOpt(10))
	fmt.Println("  复杂度: O(n) 时间, O(1) 空间")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 爬楼梯
// ----------------------------------------------------------
func climbStairs() {
	fmt.Println("=== 2. 爬楼梯 ===")
	fmt.Println("  问题: 每次 1 或 2 步，到第 n 阶有多少种方法?")
	fmt.Println("  状态转移: dp[i] = dp[i-1] + dp[i-2]")

	n := 10
	prev, curr := 1, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	fmt.Printf("  爬到第 %d 阶: %d 种方法\n", n, curr)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 最长递增子序列 (LIS)
// ----------------------------------------------------------
func lis() {
	fmt.Println("=== 3. 最长递增子序列 ===")

	nums := []int{10, 9, 2, 5, 3, 7, 101, 18}

	// O(n²) DP
	n := len(nums)
	dp := make([]int, n)
	for i := range dp {
		dp[i] = 1
	}
	for i := 1; i < n; i++ {
		for j := 0; j < i; j++ {
			if nums[j] < nums[i] && dp[j]+1 > dp[i] {
				dp[i] = dp[j] + 1
			}
		}
	}

	maxLen := 0
	for _, v := range dp {
		if v > maxLen {
			maxLen = v
		}
	}
	fmt.Printf("  nums: %v\n", nums)
	fmt.Printf("  LIS 长度: %d\n", maxLen)
	fmt.Println("  状态转移: dp[i] = max(dp[j]+1) for j<i and nums[j]<nums[i]")
	fmt.Println("  优化: 二分查找 → O(n log n)")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 最大子数组和 (Kadane)
// ----------------------------------------------------------
func maxSubArray() {
	fmt.Println("=== 4. 最大子数组和 ===")

	nums := []int{-2, 1, -3, 4, -1, 2, 1, -5, 4}

	maxSum := nums[0]
	currSum := nums[0]
	for i := 1; i < len(nums); i++ {
		if currSum+nums[i] > nums[i] {
			currSum += nums[i]
		} else {
			currSum = nums[i]
		}
		if currSum > maxSum {
			maxSum = currSum
		}
	}

	fmt.Printf("  nums: %v\n", nums)
	fmt.Printf("  最大子数组和: %d\n", maxSum)
	fmt.Println("  状态转移: dp[i] = max(dp[i-1]+nums[i], nums[i])")
	fmt.Println("  即: 要么续接前面的，要么从当前重新开始")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 0-1 背包
// ----------------------------------------------------------
func knapsack() {
	fmt.Println("=== 5. 0-1 背包 ===")

	weights := []int{1, 3, 4, 5}
	values := []int{1, 4, 5, 7}
	W := 7

	// dp[j] = 容量为 j 时的最大价值
	dp := make([]int, W+1)
	for i := 0; i < len(weights); i++ {
		for j := W; j >= weights[i]; j-- { // 逆序! 避免重复选取
			if dp[j-weights[i]]+values[i] > dp[j] {
				dp[j] = dp[j-weights[i]] + values[i]
			}
		}
	}

	fmt.Printf("  重量: %v\n", weights)
	fmt.Printf("  价值: %v\n", values)
	fmt.Printf("  背包容量 %d, 最大价值: %d\n", W, dp[W])
	fmt.Println()
	fmt.Println("  逆序遍历的原因:")
	fmt.Println("    正序: dp[j-w[i]] 可能已经被当前物品更新过 → 重复选取")
	fmt.Println("    逆序: dp[j-w[i]] 还是上一层的结果 → 每个物品只用一次")
	fmt.Println()
	fmt.Println("  完全背包 (物品无限): 正序遍历即可")
	fmt.Println("  多重背包 (物品有限): 二进制拆分优化")
	fmt.Println()
}

// ----------------------------------------------------------
// 6. 最长公共子序列 (LCS)
// ----------------------------------------------------------
func lcs() {
	fmt.Println("=== 6. 最长公共子序列 ===")

	s1, s2 := "abcde", "ace"

	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	fmt.Printf("  s1=%q, s2=%q\n", s1, s2)
	fmt.Printf("  LCS 长度: %d\n", dp[m][n])
	fmt.Println()
	fmt.Println("  状态转移:")
	fmt.Println("    s1[i]==s2[j] → dp[i][j] = dp[i-1][j-1] + 1")
	fmt.Println("    s1[i]!=s2[j] → dp[i][j] = max(dp[i-1][j], dp[i][j-1])")
	fmt.Println()
	fmt.Println("  面试追问: 如何推导状态转移方程?")
	fmt.Println("  答: 三步走:")
	fmt.Println("    1. 定义状态: dp[i] 表示什么")
	fmt.Println("    2. 找转移: dp[i] 和 dp[i-1] 的关系")
	fmt.Println("    3. 定初始: dp[0] 是什么")
	fmt.Println()
}
