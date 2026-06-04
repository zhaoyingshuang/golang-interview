package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// ============================================================
// 哈希表
// ============================================================

func main() {
	hashFunc()
	chainingHashMap()
	goMapPrinciple()
	syncMapDemo()
}

// ----------------------------------------------------------
// 1. 哈希函数
// ----------------------------------------------------------
func hashFunc() {
	fmt.Println("=== 1. 哈希函数 ===")

	// 简单哈希函数示例
	simpleHash := func(s string, bucketCount int) int {
		h := 0
		for _, c := range s {
			h = (h*31 + int(c)) % bucketCount
		}
		return h
	}

	keys := []string{"hello", "world", "golang", "hash", "map"}
	for _, k := range keys {
		fmt.Printf("  hash(\"%s\") = %d\n", k, simpleHash(k, 16))
	}

	// Go 标准库哈希
	h := fnv.New32a()
	h.Write([]byte("hello"))
	fmt.Printf("  fnv32a(\"hello\") = %d\n", h.Sum32())
	fmt.Println()
	fmt.Println("  好的哈希函数要求:")
	fmt.Println("    1. 均匀分布 — 减少冲突")
	fmt.Println("    2. 计算高效 — O(1)")
	fmt.Println("    3. 确定性 — 相同输入相同输出")
	fmt.Println("    4. Go map 使用 AES 哈希 (如果 CPU 支持 AES-NI)")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 拉链法哈希表
// ----------------------------------------------------------
func chainingHashMap() {
	fmt.Println("=== 2. 拉链法哈希表 ===")

	type KV struct {
		key   string
		value string
		next  *KV
	}

	type HashMap struct {
		buckets []*KV
		size    int
	}

	newHashMap := func(cap int) *HashMap {
		return &HashMap{buckets: make([]*KV, cap)}
	}

	put := func(m *HashMap, key, value string) {
		idx := simpleHashStr(key, len(m.buckets))
		head := m.buckets[idx]
		// 查找是否已存在
		for node := head; node != nil; node = node.next {
			if node.key == key {
				node.value = value
				return
			}
		}
		// 头插法
		m.buckets[idx] = &KV{key: key, value: value, next: head}
		m.size++
	}

	get := func(m *HashMap, key string) (string, bool) {
		idx := simpleHashStr(key, len(m.buckets))
		for node := m.buckets[idx]; node != nil; node = node.next {
			if node.key == key {
				return node.value, true
			}
		}
		return "", false
	}

	m := newHashMap(8)
	put(m, "name", "张三")
	put(m, "age", "25")
	put(m, "city", "北京")

	for _, k := range []string{"name", "age", "city", "email"} {
		if v, ok := get(m, k); ok {
			fmt.Printf("  %s = %s\n", k, v)
		} else {
			fmt.Printf("  %s = (未找到)\n", k)
		}
	}
	fmt.Println()
}

func simpleHashStr(s string, n int) int {
	h := 0
	for _, c := range s {
		h = (h*31 + int(c)) % n
	}
	return h
}

// ----------------------------------------------------------
// 3. Go map 原理
// ----------------------------------------------------------
func goMapPrinciple() {
	fmt.Println("=== 3. Go map 原理 ===")
	fmt.Println()
	fmt.Println("  底层结构 (runtime/map.go):")
	fmt.Println("  type hmap struct {")
	fmt.Println("    count     int            // 元素个数")
	fmt.Println("    B         uint8          // 桶数 = 2^B")
	fmt.Println("    hash0     uint32         // 哈希种子")
	fmt.Println("    buckets   unsafe.Pointer // 桶数组")
	fmt.Println("    oldBuckets unsafe.Pointer // 扩容时的旧桶")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("  每个桶 (bmap) 存储 8 个键值对:")
	fmt.Println("  type bmap struct {")
	fmt.Println("    tophash [8]uint8  // 哈希值高 8 位，用于快速查找")
	fmt.Println("    keys    [8]keyType")
	fmt.Println("    values  [8]valueType")
	fmt.Println("    overflow *bmap    // 溢出桶指针")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("  查找过程:")
	fmt.Println("    1. hash(key) → 低 B 位定位桶")
	fmt.Println("    2. tophash = hash(key) 高 8 位")
	fmt.Println("    3. 遍历桶中的 8 个槽位，比较 tophash")
	fmt.Println("    4. tophash 匹配后再比较完整 key")
	fmt.Println("    5. 当前桶没找到 → 去溢出桶找")
	fmt.Println()
	fmt.Println("  扩容条件:")
	fmt.Println("    负载因子 > 6.5 (元素数 / 桶数) → 翻倍扩容")
	fmt.Println("    溢出桶过多 → 等量扩容(整理)")
	fmt.Println()
	fmt.Println("  面试追问: map 遍历顺序为什么是随机的?")
	fmt.Println("  答: runtime 在遍历开始时随机选一个起始桶位置")
	fmt.Println("  目的: 防止依赖遍历顺序的代码 (Go 官方刻意设计)")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. sync.Map
// ----------------------------------------------------------
func syncMapDemo() {
	fmt.Println("=== 4. sync.Map ===")
	fmt.Println()
	fmt.Println("  Go map 不是线程安全的: 并发读写 → panic")
	fmt.Println()
	fmt.Println("  方案1: sync.RWMutex + map")
	fmt.Println("    type SafeMap struct {")
	fmt.Println("      mu sync.RWMutex")
	fmt.Println("      m  map[string]any")
	fmt.Println("    }")
	fmt.Println("    适用: 写多读多、key 集合稳定")
	fmt.Println()
	fmt.Println("  方案2: sync.Map (标准库)")
	fmt.Println("    var m sync.Map")
	fmt.Println("    m.Store(\"key\", \"value\")")
	fmt.Println("    v, ok := m.Load(\"key\")")
	fmt.Println("    m.Delete(\"key\")")
	fmt.Println("    m.Range(func(k, v any) bool { ... })")
	fmt.Println()
	fmt.Println("  sync.Map 原理 (read + dirty 双存储):")
	fmt.Println("    read  map[any]*entry — 无锁读取")
	fmt.Println("    dirty map[any]*entry — 有锁写入")
	fmt.Println()
	fmt.Println("    读: 先查 read (无锁) → miss → 查 dirty (加锁)")
	fmt.Println("    写: read 中存在 → 原子更新 entry → 不加锁")
	fmt.Println("        read 中不存在 → 加锁写 dirty")
	fmt.Println("    miss 多次 → dirty 提升为 read (整体替换)")
	fmt.Println()
	fmt.Println("  适用场景: 读多写少 (如配置缓存、连接池)")
	fmt.Println("  不适用: 写多、key 频繁增删 → 性能不如 mutex+map")
	fmt.Println()

	// 演示 sync.Map
	var m sync.Map
	m.Store("name", "张三")
	m.Store("age", 25)
	if v, ok := m.Load("name"); ok {
		fmt.Printf("  sync.Map Load: name = %v\n", v)
	}
	fmt.Println()
}
