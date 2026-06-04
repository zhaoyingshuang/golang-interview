---
title: 哈希表
---

## 1. 哈希函数设计

好的哈希函数要求：均匀分布、计算高效 O(1)、确定性。Go map 使用 AES 哈希（如果 CPU 支持 AES-NI）。

## 2. 冲突解决

**拉链法**：每个桶维护一个链表。Go map 使用溢出桶机制。

**开放寻址法**：冲突时探测下一个位置。适用于负载因子低的场景。

::: info Go map 的溢出桶
每个桶 (bmap) 存储 8 个键值对。满了后通过 `overflow` 指针链接溢出桶。
查找过程：hash(key) 低 B 位定位桶 → 比较 tophash（高 8 位）→ 比较完整 key → 溢出桶继续。
:::

## 3. Go map 扩容

- **负载因子 > 6.5** → 翻倍扩容（新桶数量 × 2）
- **溢出桶过多** → 等量扩容（整理，减少溢出桶）
- **渐进式迁移**：扩容不一次性完成，每次写入操作迁移 1-2 个桶

::: warning 面试高频
Go map 遍历顺序为什么随机？→ runtime 在遍历开始时随机选一个起始桶位置，防止依赖遍历顺序的代码。

map 不是线程安全的！并发读写会 panic。
:::

## 4. sync.Map

双存储结构：`read` (无锁读) + `dirty` (有锁写)

- 读：先查 read（无锁）→ miss → 查 dirty（加锁）
- 写：read 中存在 → 原子更新 → 不加锁
- miss 多次 → dirty 提升为 read

适用：**读多写少**（配置缓存、连接池）。写多场景不如 `sync.RWMutex + map`。

```go
var m sync.Map
m.Store("key", "value")
v, ok := m.Load("key")
m.Range(func(k, v any) bool { return true })
```
