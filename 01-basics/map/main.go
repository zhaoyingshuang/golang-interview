package main

import (
	"fmt"
	"hash/maphash"
	"math"
	"sync"
)

// ============================================================
// Map 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. map 的底层实现？(哈希表 + 拉链法)
// 2. map 的扩容策略？(等量扩容 + 增量扩容)
// 3. 为什么 map 的遍历是随机的？
// 4. map 是并发安全的吗？如何实现并发安全的 map？
// 5. map 的 key 有什么要求？
// 6. float 可以作为 key 吗？

func main() {
	structure()
	hashAndBucket()
	expansion()
	iterationRandomness()
	concurrentMap()
	keyRequirements()
}

// ----------------------------------------------------------
// 1. 底层结构
// ----------------------------------------------------------
// runtime/map.go 中的核心结构:
//
//   type hmap struct {
//       count     int            // 元素个数，len(map) 返回的就是它
//       flags     uint8          // 标志位（是否正在写入等）
//       B         uint8          // 桶数量 = 2^B
//       hash0     uint32         // 哈希种子（随机化，防止 hash 碰撞攻击）
//       buckets   unsafe.Pointer // 桶数组指针
//       oldbuckets unsafe.Pointer // 扩容时指向旧桶
//       nevacuate  uintptr       // 扩容进度（已迁移的旧桶编号）
//       extra     *mapextra      // 可选字段，存储 overflow 桶
//   }
//
//   type bmap struct {
//       tophash [8]uint8         // 每个槽位存储 hash 值的高 8 位
//       // 后面紧跟:
//       // keys   [8]keyType     // 8 个 key 连续存储
//       // values [8]valueType   // 8 个 value 连续存储
//       // overflow *bmap        // 溢出桶指针
//   }
//
// 核心设计:
//   - 每个桶最多存 8 个键值对
//   - key 和 value 分开存储（不是 key1/value1/key2/value2），
//     这样对齐填充更少，内存更紧凑
//   - 溢出桶: 当一个桶的 8 个槽都满了，会链接一个溢出桶
//   - tophash 优化: 先比较 hash 高 8 位，不匹配直接跳过，避免完整比较 key

func structure() {
	fmt.Println("=== 1. 底层结构 ===")

	m := make(map[string]int, 10) // 预分配 hint=10，减少扩容
	m["hello"] = 1
	m["world"] = 2
	fmt.Printf("map: %v, len=%d\n", m, len(m))

	// 检查 key 是否存在
	v, ok := m["missing"]
	fmt.Printf("key 不存在: value=%d, ok=%v\n", v, ok)

	// 删除 key
	delete(m, "hello")
	fmt.Printf("delete 后: %v\n", m)
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 哈希与桶
// ----------------------------------------------------------
// 查找过程:
//   1. 计算 key 的 hash 值（使用 hash0 种子）
//   2. 用 hash 值的低位（低 B 位）确定桶编号
//   3. 用 hash 值的高 8 位（tophash）在桶内快速定位
//   4. tophash 匹配后，再做完整的 key 比较
//   5. 当前桶没找到，沿 overflow 链继续查找
func hashAndBucket() {
	fmt.Println("=== 2. 哈希与桶 ===")

	// Go 的 map 使用的哈希算法取决于 key 类型:
	// - string/aes 启用时: 使用 AES 指令加速
	// - 其他: 使用运行时实现的哈希函数

	// 演示 hash 种子的随机性
	var h maphash.Hash
	h.Write([]byte("hello"))
	hash1 := h.Sum64()
	h.Reset()
	h.Write([]byte("hello"))
	hash2 := h.Sum64()
	fmt.Printf("同一段代码不同 map 实例的 hash 不同（每次运行也不同）\n")
	fmt.Printf("hash1=%016x hash2=%016x 相等=%v\n", hash1, hash2, hash1 == hash2)

	// tophash 的作用: 快速筛选
	// tophash = hash >> (64 - 8)，即高 8 位
	tophash := byte(hash1 >> 56)
	fmt.Printf("tophash = 0x%02x (用于桶内快速比较)\n", tophash)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 扩容策略
// ----------------------------------------------------------
// 触发扩容的条件:
//   1. 负载因子 > 6.5（平均每个桶超过 6.5 个元素）→ 增量扩容（B+1）
//   2. 溢出桶过多 → 等量扩容（B 不变，重新排列以减少溢出桶）
//
// 增量扩容:
//   - 新桶数量 = 旧桶数量 * 2
//   - 数据从旧桶逐步迁移到新桶（每次写入/删除操作时迁移 1-2 个桶）
//   - 查找时先查新桶，再查旧桶
//
// 等量扩容:
//   - 大量增删导致溢出桶过多，但实际数据不多
//   - 新桶数量 = 旧桶数量（不增长）
//   - 重新排列数据，减少溢出桶
//
// 重要: 扩容不是一次完成的，而是渐进式的（避免一次性大延迟）。
func expansion() {
	fmt.Println("=== 3. 扩容策略 ===")

	// 演示负载因子
	// 负载因子 = count / (2^B)，Go 选择的阈值是 6.5
	// 每个桶最多 8 个槽，6.5 意味着在空间利用率和性能之间取平衡
	fmt.Printf("负载因子阈值: 6.5 (每个桶平均 6.5 个元素时扩容)\n")

	// 为什么不选更高的阈值？→ 溢出桶太多，查找变慢
	// 为什么不选更低的阈值？→ 内存浪费

	// 溢出桶过多的判断（Go 源码中的条件）:
	//   如果 B <= 15: overflow 桶数 > 2^B 时触发等量扩容
	//   如果 B > 15:  overflow 桶数 > 2^15 时触发等量扩容
	fmt.Println("溢出桶过多 → 等量扩容（压缩溢出桶）")
	fmt.Println("负载因子 > 6.5 → 增量扩容（桶数量翻倍）")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 遍历随机性
// ----------------------------------------------------------
// 为什么 for range map 的顺序是随机的？
//
// 不是因为 map 真的随机排列了元素。
// map 的内部存储是确定的，但 Go 故意在遍历起始位置加了随机偏移。
// runtime/map.go → mapiterinit():
//   r := uintptr(fastrand())
//   it.startBucket = r & bucketMask(h.B)     // 随机起始桶
//   it.offset = uint8(r >> h.B & (bucketCnt - 1)) // 随机桶内偏移
//
// 原因: 防止开发者依赖遍历顺序（哈希表本身就不保证顺序），
// 如果 Go 不加随机性，代码可能碰巧正确运行，但换一个 Go 版本或平台就可能出 bug。
func iterationRandomness() {
	fmt.Println("=== 4. 遍历随机性 ===")

	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	fmt.Print("第一次遍历: ")
	for k, v := range m {
		fmt.Printf("%s=%d ", k, v)
	}
	fmt.Print("\n第二次遍历: ")
	for k, v := range m {
		fmt.Printf("%s=%d ", k, v)
	}
	fmt.Println("\n(每次运行结果不同，Go 故意为之)")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 并发安全
// ----------------------------------------------------------
// map 本身不是并发安全的！
// 同时读写会导致: fatal error: concurrent map read and map write
// 或者: fatal error: concurrent map writes
//
// 这个检测在 runtime 中通过 hmap.flags 实现（不是读写锁！）。
// 只检测了并发冲突并 panic，没有做互斥保护。
//
// 并发安全方案:
//   1. sync.RWMutex — 简单直接
//   2. sync.Map — 读多写少场景优化
//   3. 分片 map — 高并发场景最优

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

func (sm *SafeMap) Set(key string, value int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

func concurrentMap() {
	fmt.Println("=== 5. 并发安全 ===")

	sm := &SafeMap{m: make(map[string]int)}

	// 使用 sync.Map 的场景:
	// 1. key 一旦写入很少变更（读多写少）
	// 2. 多个 goroutine 读写不同的 key（无竞争）

	var smap sync.Map
	smap.Store("key1", "value1")
	v, ok := smap.Load("key1")
	fmt.Printf("sync.Map: value=%v, ok=%v\n", v, ok)

	// LoadOrStore: 原子性的"读取或存储"
	actual, loaded := smap.LoadOrStore("key1", "new_value")
	fmt.Printf("LoadOrStore: actual=%v, loaded=%v (已存在，不会覆盖)\n", actual, loaded)

	_ = sm
	fmt.Println()
}

// ----------------------------------------------------------
// 6. Key 的要求
// ----------------------------------------------------------
// map 的 key 必须是可比较的（comparable），即支持 == 操作符。
//
// 可用的 key 类型:
//   bool, int, float, complex, string, pointer, channel, interface,
//   array, struct（如果所有字段都可比较）
//
// 不可用的 key 类型:
//   slice, map, function（不可比较）
//
// float 作为 key 的陷阱:
//   NaN != NaN，所以如果用 NaN 作为 key，存储后无法取出！
//   而且 NaN 作为 key 可以存无数次（每次查找都不匹配）。

func keyRequirements() {
	fmt.Println("=== 6. Key 的要求 ===")

	// struct 作为 key（所有字段都可比较即可）
	type Point struct{ X, Y int }
	m := map[Point]string{
		{1, 2}: "A",
		{3, 4}: "B",
	}
	fmt.Printf("struct 作为 key: %v\n", m)

	// NaN 陷阱
	nan := math.NaN()
	nanMap := map[float64]int{}
	nanMap[nan] = 1
	nanMap[nan] = 2
	nanMap[nan] = 3
	fmt.Printf("NaN 作为 key, len=%d (每次都是\"新\"key！)\n", len(nanMap))

	v, ok := nanMap[nan]
	fmt.Printf("用 NaN 查找: value=%v, ok=%v (永远找不到！)\n", v, ok)
	fmt.Println()
}
