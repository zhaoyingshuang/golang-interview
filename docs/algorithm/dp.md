---
title: 动态规划
---

## 1. DP 核心思想

动态规划三要素：
1. **最优子结构** — 大问题的最优解包含小问题的最优解
2. **无后效性** — 当前状态确定后，之前如何到达不影响后续决策
3. **状态转移方程** — dp[i] 与 dp[i-1] 的关系

::: tip 解题步骤
1. 定义状态：dp[i] 表示什么
2. 找转移：dp[i] 和 dp[i-1] 的关系
3. 定初始：dp[0] 是什么
4. 考虑空间优化
:::

## 2. 线性 DP

**斐波那契 / 爬楼梯**：`dp[i] = dp[i-1] + dp[i-2]`，空间优化到 O(1)。

**最大子数组和 (Kadane)**：`dp[i] = max(dp[i-1]+nums[i], nums[i])`

```go
maxSum, currSum := nums[0], nums[0]
for i := 1; i < len(nums); i++ {
    if currSum+nums[i] > nums[i] { currSum += nums[i] } else { currSum = nums[i] }
    if currSum > maxSum { maxSum = currSum }
}
```

**最长递增子序列 (LIS)**：`dp[i] = max(dp[j]+1)` for j < i and nums[j] < nums[i]。O(n²) DP，二分优化到 O(n log n)。

## 3. 背包问题

**0-1 背包**：每个物品用一次。逆序遍历避免重复选取。

```go
dp := make([]int, W+1)
for i := 0; i < n; i++ {
    for j := W; j >= weights[i]; j-- {  // 逆序!
        dp[j] = max(dp[j], dp[j-weights[i]]+values[i])
    }
}
```

**完全背包**：物品无限，正序遍历。**多重背包**：二进制拆分优化。

## 4. 字符串 DP

**最长公共子序列 (LCS)**：
- `s1[i]==s2[j]` → `dp[i][j] = dp[i-1][j-1] + 1`
- `s1[i]!=s2[j]` → `dp[i][j] = max(dp[i-1][j], dp[i][j-1])`

**编辑距离**：插入/删除/替换，`dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)`

::: warning 面试追问：如何推导状态转移方程
三步走：定义状态 → 找转移（考虑所有选择）→ 定初始。画表格手推几个例子帮助理解。
:::
