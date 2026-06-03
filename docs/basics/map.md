---
title: Map 哈希表
---

## 1. 底层结构

Map 底层是 runtime/map.go 中的 hmap 结构体。核心设计：每个桶（bmap）最多存 8 个键值对，key 和 value **分开连续存储**（不是交替存储），这样对齐填充更少，内存更紧凑。tophash 存储哈希值高 8 位，用于快速筛选。溢出桶：当一个桶的 8 个槽都满了，会链接一个溢出桶。

**为什么用 8**：8 是空间利用率和查找效率的平衡点，tophash 正好 8 字节方便比较。

::: tip 使用场景
map 是 Go 中实现路由表（gin/chi）、缓存、配置管理的核心数据结构。

::: warning 面试追问
map 的 key 和 value 为什么分开存储？→ 减少内存对齐的 padding 浪费。

```go
type hmap struct {
    count     int            // 元素个数, len(map) 返回它
    B         uint8          // 桶数量 = 2^B
    hash0     uint32         // 哈希种子（每次创建 map 随机生成）
    buckets   unsafe.Pointer // 桶数组指针
    oldbuckets unsafe.Pointer // 扩容时指向旧桶
    nevacuate  uintptr       // 扩容进度（已迁移的旧桶编号）
}

type bmap struct {
    tophash [8]uint8  // hash 高 8 位，快速比较
    // 后面紧跟:
    // keys   [8]keyType     // 8个key连续存储
    // values [8]valueType   // 8个value连续存储
    // overflow *bmap        // 溢出桶指针
}
```

## 2. 查找过程

完整查找流程：1) 计算 key 的 hash 值（使用 hash0 种子）→ 2) 用低 B 位确定桶编号 → 3) 用高 8 位（tophash）在桶内快速定位 → 4) tophash 匹配后再完整比较 key → 5) 当前桶没找到沿 overflow 链继续。

::: warning 面试追问
tophash 的作用？→ 避免每个 key 都做完整比较，先比 1 字节 tophash，不匹配直接跳过，效率接近 O(1)。
:::

**为什么每次运行 map 的 hash 不同**？→ hash0 是随机种子，防止 hash 碰撞攻击（恶意构造大量相同 hash 的 key 导致性能退化）。

```go
// 查找过程:
// hash(key) → 低B位确定桶 → 高8位快速筛选 → 完整比较key

m := make(map[string]int, 10) // hint=10，预分配
m["hello"] = 1
v, ok := m["missing"]  // v=0, ok=false（零值）
delete(m, "hello")

// 面试追问: make(map[string]int, 10) 的 10 是什么？
// 是 hint（提示），不是硬性限制
// runtime 会据此分配足够的桶，减少后续扩容
```

## 3. 扩容策略

两种扩容：**增量扩容**（负载因子 > 6.5 时触发）：桶数量翻倍，数据逐步迁移。

**等量扩容**（溢出桶过多时触发）：桶数量不变，重新排列数据以减少溢出桶，发生在大量增删后。扩容是**渐进式**的：每次写入/删除操作时迁移 1-2 个桶，查找时先查新桶再查旧桶。

**为什么是 6.5**：每个桶 8 个槽，6.5 是空间利用率和查找性能的平衡。

::: tip 使用场景
高并发写入 map 后大量删除导致内存不释放 → 等量扩容会自动整理。

::: warning 面试追问
map 扩容时性能会下降吗？→ 不会，因为渐进式迁移，每次操作只多迁移 1-2 个桶。

```go
// 负载因子 = count / (2^B)
// > 6.5 → 增量扩容（桶翻倍）
// 溢出桶过多 → 等量扩容（整理碎片）

// 面试追问: 怎么知道 map 是否在扩容？
// hmap.oldbuckets != nil 表示正在扩容

// 生产建议: 预分配 hint 减少扩容
m := make(map[string]int, 1000) // 预分配

// 面试追问: 删除很多 key 后内存会释放吗？
// 不会立即释放，但等量扩容时会整理
// 如果需要立即释放，创建新 map 并复制
```

## 4. 并发安全

Map 本身**不是并发安全的**！同时读写会 fatal error: concurrent map read and map write。这个检测在 runtime 中通过 hmap.flags 实现（不是读写锁），只检测了并发冲突并 panic，没有做互斥保护。
:::

**三种解决方案**：1) **sync.RWMutex**（最常用）：简单直接，适合大多数场景。2) **sync.Map**：读多写少场景优化（如缓存），内部用 read/dirty 双 map 实现，读操作无锁。不适合频繁写入。3) **分片 map**：高并发最优（如 N 个 map + hash 分流），被 bigcache 等库采用。

::: tip 生产选择
一般用 RWMutex 就够了；读远多于写用 sync.Map；QPS 极高（10w+）考虑分片 map。

```go
// 方案1: RWMutex（推荐，适合大多数场景）
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}
func (sm *SafeMap) Get(key string) (int, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    v, ok := sm.m[key]
    return v, ok
}

// 方案2: sync.Map（适合读多写少，如缓存）
var m sync.Map
m.Store("key", "value")
v, ok := m.Load("key")
m.Range(func(k, v any) bool { return true })

// 方案3: 分片 map（超高并发）
type ShardedMap struct {
    shards [64]struct {
        sync.RWMutex
        m map[string]int
    }
}
```

## 5. Key 的要求

Key 必须是可比较的（comparable），即支持 == 操作符。
:::

**可用的 key**：bool、int、float、string、pointer、channel、interface、array、struct（所有字段都可比较）。

**不可用**：slice、map、function。

**为什么 slice 不能做 key**：slice 包含指针，但 == 比较的是 header 不是内容，且 slice 是可变的，如果允许做 key，修改后无法查找。

**float 的 NaN 陷阱**：NaN != NaN，所以用 NaN 作为 key 存储后无法取出，且可以存无数次。

::: tip 使用场景
用 struct 做复合 key（如坐标点、日期+类型组合）。

::: warning 面试追问
为什么可比较的要求这么设计？→ map 内部用 == 比较key，不可比较的类型无法做 hash 表的 key。

```go
// struct 做复合 key（生产常用）
type CacheKey struct {
    UserID int
    Type   string
}
cache := map[CacheKey]string{
    {1, "profile"}: "cached_data",
}

// 面试: 以下哪种可以作为 map key？
// [3]int   → 可以（array，值类型）
// []int    → 不可以（slice）
// *int     → 可以（pointer，比较地址）
// any      → 可以（interface，比较动态值）

// NaN 陷阱
nan := math.NaN()
m := map[float64]int{}
m[nan] = 1
m[nan] = 2
// len(m) == 2！每次 NaN 都是"新" key
// m[nan] → 永远找不到（NaN != NaN）
```
