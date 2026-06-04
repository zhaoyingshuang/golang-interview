package main

func sliceTopic() Topic {
	return Topic{
		ID: "slice", Title: "Slice 切片", Chapter: "01-basics",
		Sections: []Section{
			s("底层结构",
				"Slice 底层是 runtime/slice.go 中的 sliceHeader 结构体，只包含三个字段：array（指向底层数组的指针）、len（当前长度）、cap（容量）。Slice 本身只是一个 24 字节的结构体（64位系统），真正的数据存储在底层数组中。**与 array 的关键区别**：array 的长度是类型的一部分（[3]int 和 [4]int 是不同类型），赋值会复制整个数组；slice 是引用语义类型，赋值只复制 header。**使用场景**：array 适用于长度固定的场景（如 IPv4 地址 [4]byte、SHA256 哈希 [32]byte）；slice 是 Go 中最常用的数据结构，几乎一切列表都用 slice。创建方式有 4 种：字面量、make、从数组切片、三索引切片。三索引切片 arr[low:high:max] 可限制容量，防止 append 时影响原数组。",
				`type slice struct {
    array unsafe.Pointer // 指向底层数组的指针
    len   int            // 当前长度
    cap   int            // 容量
}

// 创建 slice 的 4 种方式
s1 := []int{1, 2, 3}            // 字面量
s2 := make([]int, 3, 10)        // make，cap=10
s3 := arr[1:3]                  // 从数组切片，cap=4
s5 := arr[1:3:3]                // 三索引切片，cap=2（限制容量）

// 面试追问: slice 占多少内存？
// 24 bytes (64位系统): 指针8 + len8 + cap8`),
			s("扩容机制",
				"Go 1.18+ 使用更平滑的增长公式：newCap = oldCap + (oldCap + 3*256) / 4，增长因子从 2.0 平滑过渡到 1.25，不再有 1024 的分界线。最终容量还会做内存对齐（根据元素大小向上取整到合适的内存分配阶级），所以实际 cap 可能比公式计算值更大。**为什么需要扩容**：append 时如果 cap 不够，必须分配新底层数组、复制旧数据、指向新数组。这个过程有 CPU 和内存开销。**生产建议**：如果知道大概需要多少元素，用 make([]T, 0, n) 预分配。不确定时可以分两步：先收集确定元素，再用 append。**面试追问**：Go 1.18 前后的扩容策略有什么区别？→ 1.18 前有 1024 分界线，之后用连续平滑公式。",
				`var s []int
for i := 0; i < 20; i++ {
    fmt.Printf("len=%2d cap=%2d\n", len(s), cap(s))
    s = append(s, i)
}
// 输出: 0→1→2→4→4→8→8→16→16→32

// 预分配避免扩容（生产推荐）
s2 := make([]int, 0, 20) // 一次分配，零次扩容

// 面试追问: 扩容时会发生什么？
// 1. 分配新的底层数组
// 2. 复制旧数据到新数组
// 3. slice 指向新数组
// 4. 旧数组等待 GC`),
			s("作为函数参数",
				"Slice 作为参数传递时，复制的是 slice header（指针+len+cap，共 24 字节），底层数组不会被复制，所以函数内可以修改元素。但如果函数内触发了扩容，调用方不会看到新元素——因为扩容后 slice 指向了新的底层数组。**这是最常见的面试坑之一**。**生产场景**：API handler 中从数据库查询结果 []User 传给 service 层处理，service 可以修改 User 字段，但如果 append 了新元素，handler 看不到。**解决方案**：1) 返回新 slice（最常用）；2) 传 *[]T 指针（少见）；3) 预分配足够容量（脆弱，不推荐）。",
				`func modifySlice(s []int) {
    s[0] = 100  // 修改底层数组，调用方可见
}

func appendSlice(s []int) {
    s = append(s, 4) // 扩容 → 新底层数组，调用方看不到！
}

// 推荐写法: 返回新 slice
func safeAppend(s []int, v int) []int {
    return append(s, v)
}

// 调用方
s = safeAppend(s, 4) // 接收返回值`),
			s("常见陷阱",
				"**陷阱1: 内存泄漏**。大 slice 上取一小片，底层数组仍被引用无法 GC。生产场景：读取大文件后只取 header 部分，但整个文件内容都在内存中。解决：用 copy。**陷阱2: 共享底层数组**。两个 slice 共享底层数组，一个 append 后另一个可能读到意外数据。生产场景：从请求体解析出多个字段引用同一个大 buffer。**陷阱3: range 值复制**。for range 中的变量是副本，修改无效。生产场景：批量更新 struct 字段。**面试高频追问**：如何检测 slice 内存泄漏？→ pprof heap profile，对比两次快照看 inuse_space 是否持续增长。",
				`// 陷阱1: 内存泄漏（生产常见！）
big := make([]byte, 1<<20) // 1MB
small := big[:10]          // 1MB 都无法被 GC！
// 解决: copy
small2 := make([]byte, 10)
copy(small2, big[:10])

// 陷阱2: 共享底层数组
s1 := make([]int, 3, 6)
s2 := s1[0:3]
s2 = append(s2, 100)  // s1 的 cap 区域被修改

// 陷阱3: range 值复制
for _, item := range items {
    item.Value = 0  // 修改的是副本！无效
}
// 正确: 用索引
for i := range items {
    items[i].Value = 0
}`),
			s("Slice 技巧",
				"掌握这些技巧在面试和日常开发中都能写出更优雅的代码。**快速删除**（不保序 O(1)）：把最后一个元素移到被删位置，适用于顺序无关的场景（如从连接池中移除）。**保序删除** O(n)：用 append 覆盖。**插入元素**：先 append 腾位再赋值。**原地过滤**：双指针法，零分配。**去重**：先排序再用双指针。**生产场景**：WebSocket 连接管理（快速删除断开的连接）、消息队列消费者（原地过滤已处理的消息）。",
				`// 删除第 i 个元素（不保序 O(1)）- 连接池场景
s[i] = s[len(s)-1]
s = s[:len(s)-1]

// 保序删除 O(n) - 有序列表场景
s = append(s[:i], s[i+1:]...)

// 插入元素到位置 i
s = append(s[:i+1], s[i:]...)
s[i] = v

// 原地过滤（零分配）- 消息处理场景
n := 0
for _, v := range s {
    if keep(v) { s[n] = v; n++ }
}
s = s[:n]

// 去重（先排序）
sort.Ints(s)
j := 0
for i := 1; i < len(s); i++ {
    if s[j] != s[i] { j++; s[j] = s[i] }
}
s = s[:j+1]`),
		},
	}
}

func mapTopic() Topic {
	return Topic{
		ID: "map", Title: "Map 哈希表", Chapter: "01-basics",
		Sections: []Section{
			s("底层结构",
				"Map 底层是 runtime/map.go 中的 hmap 结构体。核心设计：每个桶（bmap）最多存 8 个键值对，key 和 value **分开连续存储**（不是交替存储），这样对齐填充更少，内存更紧凑。tophash 存储哈希值高 8 位，用于快速筛选。溢出桶：当一个桶的 8 个槽都满了，会链接一个溢出桶。**为什么用 8**：8 是空间利用率和查找效率的平衡点，tophash 正好 8 字节方便比较。**生产场景**：map 是 Go 中实现路由表（gin/chi）、缓存、配置管理的核心数据结构。**面试追问**：map 的 key 和 value 为什么分开存储？→ 减少内存对齐的 padding 浪费。",
				`type hmap struct {
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
}`),
			s("查找过程",
				"完整查找流程：1) 计算 key 的 hash 值（使用 hash0 种子）→ 2) 用低 B 位确定桶编号 → 3) 用高 8 位（tophash）在桶内快速定位 → 4) tophash 匹配后再完整比较 key → 5) 当前桶没找到沿 overflow 链继续。**面试追问**：tophash 的作用？→ 避免每个 key 都做完整比较，先比 1 字节 tophash，不匹配直接跳过，效率接近 O(1)。**为什么每次运行 map 的 hash 不同**？→ hash0 是随机种子，防止 hash 碰撞攻击（恶意构造大量相同 hash 的 key 导致性能退化）。",
				`// 查找过程:
// hash(key) → 低B位确定桶 → 高8位快速筛选 → 完整比较key

m := make(map[string]int, 10) // hint=10，预分配
m["hello"] = 1
v, ok := m["missing"]  // v=0, ok=false（零值）
delete(m, "hello")

// 面试追问: make(map[string]int, 10) 的 10 是什么？
// 是 hint（提示），不是硬性限制
// runtime 会据此分配足够的桶，减少后续扩容`),
			s("扩容策略",
				"两种扩容：**增量扩容**（负载因子 > 6.5 时触发）：桶数量翻倍，数据逐步迁移。**等量扩容**（溢出桶过多时触发）：桶数量不变，重新排列数据以减少溢出桶，发生在大量增删后。扩容是**渐进式**的：每次写入/删除操作时迁移 1-2 个桶，查找时先查新桶再查旧桶。**为什么是 6.5**：每个桶 8 个槽，6.5 是空间利用率和查找性能的平衡。**生产场景**：高并发写入 map 后大量删除导致内存不释放 → 等量扩容会自动整理。**面试追问**：map 扩容时性能会下降吗？→ 不会，因为渐进式迁移，每次操作只多迁移 1-2 个桶。",
				`// 负载因子 = count / (2^B)
// > 6.5 → 增量扩容（桶翻倍）
// 溢出桶过多 → 等量扩容（整理碎片）

// 面试追问: 怎么知道 map 是否在扩容？
// hmap.oldbuckets != nil 表示正在扩容

// 生产建议: 预分配 hint 减少扩容
m := make(map[string]int, 1000) // 预分配

// 面试追问: 删除很多 key 后内存会释放吗？
// 不会立即释放，但等量扩容时会整理
// 如果需要立即释放，创建新 map 并复制`),
			s("并发安全",
				"Map 本身**不是并发安全的**！同时读写会 fatal error: concurrent map read and map write。这个检测在 runtime 中通过 hmap.flags 实现（不是读写锁），只检测了并发冲突并 panic，没有做互斥保护。**三种解决方案**：1) **sync.RWMutex**（最常用）：简单直接，适合大多数场景。2) **sync.Map**：读多写少场景优化（如缓存），内部用 read/dirty 双 map 实现，读操作无锁。不适合频繁写入。3) **分片 map**：高并发最优（如 N 个 map + hash 分流），被 bigcache 等库采用。**生产选择**：一般用 RWMutex 就够了；读远多于写用 sync.Map；QPS 极高（10w+）考虑分片 map。",
				`// 方案1: RWMutex（推荐，适合大多数场景）
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
}`),
			s("Key 的要求",
				"Key 必须是可比较的（comparable），即支持 == 操作符。**可用的 key**：bool、int、float、string、pointer、channel、interface、array、struct（所有字段都可比较）。**不可用**：slice、map、function。**为什么 slice 不能做 key**：slice 包含指针，但 == 比较的是 header 不是内容，且 slice 是可变的，如果允许做 key，修改后无法查找。**float 的 NaN 陷阱**：NaN != NaN，所以用 NaN 作为 key 存储后无法取出，且可以存无数次。**生产场景**：用 struct 做复合 key（如坐标点、日期+类型组合）。**面试追问**：为什么可比较的要求这么设计？→ map 内部用 == 比较key，不可比较的类型无法做 hash 表的 key。",
				`// struct 做复合 key（生产常用）
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
// m[nan] → 永远找不到（NaN != NaN）`),
		},
	}
}

func stringTopic() Topic {
	return Topic{
		ID: "string", Title: "String 字符串", Chapter: "01-basics",
		Sections: []Section{
			s("底层结构",
				"String 底层是 StringHeader{Data unsafe.Pointer; Len int}，本质是**只读**的字节切片。\"你好\" 在 UTF-8 下占 6 字节但只有 2 个 rune。字符串字面量存储在二进制文件的 .rodata 段，运行时字符串赋值只复制 StringHeader（16 字节），不复制底层数据。**和 slice 的区别**：string 没有 cap 字段（因为不可变，不需要扩容）。**生产场景**：string 的零拷贝特性使其非常适合传递大量文本数据（如 HTTP 响应体、文件内容），因为赋值只复制 16 字节 header。**面试追问**：string 和 []byte 的转换有开销吗？→ 有，每次转换都会复制底层数据（编译器会优化一些简单场景）。",
				`type StringHeader struct {
    Data unsafe.Pointer // 指向底层字节数组
    Len  int            // 字节长度（不是字符数！）
}

s := "Hello, 世界"
len(s)                         // 13 (每个中文 3 字节)
utf8.RuneCountInString(s)      // 9 (rune 字符数)

// string 赋值只复制 header，不复制数据
s2 := s  // 只复制 16 bytes

// []byte(string) 和 string([]byte) 都会复制数据
b := []byte(s)  // 复制 13 bytes
s3 := string(b) // 又复制 13 bytes`),
			s("byte vs rune",
				"**byte = uint8**：表示一个 UTF-8 编码的字节。**rune = int32**：表示一个 Unicode 码点。\"中\" 的 UTF-8 编码是 0xE4 0xB8 0xAD → 3 个 byte，1 个 rune。**生产场景**：处理用户输入时必须用 rune 计算字符数（如限制用户名长度）；处理网络协议/文件时用 byte 操作原始数据。for range 遍历 string 时按 rune 迭代，index 是字节位置。**面试追问**：string 可以包含无效 UTF-8 吗？→ 可以，string 就是字节序列，for range 遇到无效 UTF-8 会产出 U+FFFD。",
				`s := "中A"

// 字节遍历（网络/文件 IO 场景）
for i := 0; i < len(s); i++ {
    fmt.Printf("%02x ", s[i])  // e4 b8 ad 41
}

// rune 遍历（用户输入处理场景）
for i, r := range s {
    // i 是字节位置，r 是 rune
    fmt.Printf("byte=%d rune=%U(%c)\n", i, r, r)
}

// rune 切片：支持按字符索引
runes := []rune(s)
fmt.Printf("第1个字符: %c\n", runes[0]) // 中

// 用户名长度限制（必须用 rune 计数）
if utf8.RuneCountInString(name) > 20 {
    return errors.New("name too long")
}`),
			s("拼接性能",
				"性能排序（从快到慢）：strings.Builder > strings.Join > bytes.Buffer > + 运算符 > fmt.Sprintf。**strings.Builder 最优的原因**：String() 方法通过 unsafe 直接转换底层 []byte，不需要复制。但 String() 后不能再修改（否则 panic）。**+ 运算符的问题**：每次拼接都分配新字符串，编译器会优化固定数量的 +，但循环中的 += 每次都分配。**生产建议**：循环拼接用 Builder（并 b.Grow 预分配）；少量拼接用 + 即可；数字转字符串用 strconv.Itoa（比 fmt.Sprintf 快 5-10 倍）。",
				`// 推荐: strings.Builder（循环拼接）
var b strings.Builder
b.Grow(len(parts) * 10) // 预分配
for _, p := range parts {
    b.WriteString(p)
}
result := b.String()

// 少量拼接直接用 +（编译器会优化）
s := "hello" + " " + "world" // 编译期合并

// 数字转字符串
s := strconv.Itoa(42)           // 推荐，零分配
s := fmt.Sprintf("%d", 42)      // 不推荐，有分配

// Builder 复用（高频场景）
var buf strings.Builder
buf.Grow(256)
for _, line := range lines {
    buf.Reset() // 重置但不释放内存
    buf.WriteString(line)
    process(buf.String())
}`),
			s("不可变性与内存泄漏",
				"String 是不可变的：不能通过下标修改（s[0] = 'H' 编译错误）。这个设计使得 string 天然并发安全，且可以被安全地作为 map key。**Substring 内存泄漏**：对大字符串取子串不会复制底层数据（只创建新 header），如果长期持有子串，大字符串无法被 GC。**生产场景**：解析日志文件时只取某一行，但整行引用了整个文件内容。解决：用 strings.Clone()（Go 1.18+）或 string([]byte(sub)) 强制复制。**面试追问**：string 真的完全不可变吗？→ 通过 unsafe 可以修改（修改 .rodata 段），但极其危险，可能影响所有引用相同字面量的字符串。",
				`// substring 内存泄漏（生产常见！）
big := strings.Repeat("x", 1<<20) // 1MB
section := big[:50]               // 1MB 都无法被 GC

// 解决方案
section = strings.Clone(section)  // Go 1.18+，只复制 50 字节
// 或者
section = string([]byte(big[:50])) // 强制复制

// string 和 []byte 零拷贝转换（极端性能优化，不推荐）
// unsafe 方式，修改 b 会影响 s！
s := "hello"
b := unsafe.Slice(unsafe.StringData(s), len(s))

// 面试追问: 字符串可以比较吗？
// 可以，按字节逐个比较
// "hello" == "hello" → true
// "Hello" == "hello" → false`),
		},
	}
}

func interfaceTopic() Topic {
	return Topic{
		ID: "interface", Title: "Interface 接口", Chapter: "01-basics",
		Sections: []Section{
			s("底层结构: eface 和 iface",
				"**空接口**（interface{}/any）：eface{_type *_type; data unsafe.Pointer}。**非空接口**（有方法）：iface{tab *itab; data unsafe.Pointer}。itab 缓存了接口方法到实际方法的映射，全局只生成一份（用 map[interfacetype+type]*itab 做缓存）。接口变量占 16 字节（type 指针 + data 指针）。**data 指针**：小值（<=指针大小）直接内联存储在 data 字段中，大值在堆上分配后 data 指向它。**生产场景**：fmt.Println 的参数就是 ...any（eface）；io.Reader 是非空接口（iface）。**面试追问**：接口方法调用比直接调用慢多少？→ 约 20-50ns，因为无法内联。",
				`// eface: 空接口（没有方法）
type eface struct {
    _type *_type           // 类型元数据
    data  unsafe.Pointer   // 数据指针
}

// iface: 非空接口（有方法）
type iface struct {
    tab  *itab             // 接口类型 + 方法表
    data unsafe.Pointer    // 数据指针
}

// itab: 方法查找表（全局缓存）
type itab struct {
    inter *interfacetype   // 接口类型
    _type *_type           // 实际类型
    hash  uint32           // 用于类型断言快速匹配
    fun   [1]uintptr       // 变长: 方法地址数组
}

// 面试追问: 接口变量占多少内存？
// 16 bytes (64位系统): type指针 + data指针`),
			s("nil 陷阱（经典面试题）",
				"**interface == nil 当且仅当 type 和 data 都是 nil**。这是 Go 面试最常考的陷阱之一。var p *Dog = nil; var s Speaker = p → s != nil！因为 type 是 *Dog（不是 nil），data 是 nil。调用 s.Speak() 会 panic（nil 指针解引用）。**生产场景**：函数返回 error 接口时，如果内部返回了 (*CustomError)(nil)，调用方检查 err != nil 会得到 true，导致逻辑错误。**正确做法**：函数需要返回 nil 时，显式 return nil；不要返回具体类型的 nil 指针。",
				`// 真正的 nil interface
var s1 Speaker       // s1 == nil ✓ (type=nil, data=nil)

// 看起来是 nil 其实不是！
var p *Dog = nil
var s2 Speaker = p   // s2 != nil！
// s2 = iface{type:*Dog, data:nil}

// 经典坑: 函数返回 error
func getError() error {
    var err *MyError = nil
    return err  // 返回 iface{type:*MyError, data:nil}
}
e := getError()
fmt.Println(e == nil) // false！

// 正确写法
func getError() error {
    return nil  // 直接返回 nil interface
}`),
			s("类型断言与 type switch",
				"类型断言 x.(T) 用于从接口中提取具体类型的值。**两种形式**：1) d := s.(Dog) — 类型不匹配会 panic；2) d, ok := s.(Dog) — 安全形式，不匹配时 ok=false。**type switch** 是更优雅的多类型判断：switch v := s.(type) { case Dog: ... }，v 在每个 case 中自动转换为对应类型。**生产场景**：解析 JSON 时 interface{} 需要类型断言；错误处理中区分不同错误类型；middleware 中提取请求上下文。**面试追问**：类型断言的性能开销？→ O(1)，直接查 itab.fun 表。",
				`// 安全的类型断言（生产推荐）
if c, ok := s.(*Cat); ok {
    fmt.Println(c.Name)
}

// type switch（多类型判断，优雅）
switch v := s.(type) {
case Dog:
    fmt.Printf("Dog{%s}\n", v.Name)
case *Cat:
    fmt.Printf("*Cat{%s}\n", v.Name)
default:
    fmt.Printf("unknown (%T)\n", v)
}

// 生产: JSON 解析后类型断言
var data any = json.Unmarshal(...)
switch v := data.(type) {
case map[string]any:
    // object
case []any:
    // array
case string:
    // string
}

// 错误类型判断（生产常用）
var netErr net.Error
if errors.As(err, &netErr) && netErr.Timeout() {
    // 处理超时
}`),
			s("方法集规则与设计原则",
				"**值接收者 (T)** → T 和 *T 都能实现接口。**指针接收者 (*T)** → 只有 *T 能实现接口。原因：Go 可以自动取地址（T → &T），但不能自动解引用（*T → T）。有个限制：如果值不可寻址（如 map 元素、字面量），不能自动取地址。**小接口原则**：Go 标准库推崇接口越小越好——io.Reader 只有 1 个方法，error 只有 1 个方法。**生产建议**：使用者定义接口，不是实现者（Go 的隐式实现使这成为可能）；接口通常在消费方定义，而不是在提供方。**面试追问**：什么时候用指针接收者？→ 需要修改状态、结构体较大、一致性。",
				`type Animal struct{ Name string }
func (a Animal) Speak() string  { return a.Name }     // 值接收者
func (a *Animal) ChangeName(n string) { a.Name = n }  // 指针接收者

type Changer interface{ ChangeName(string) }
// var c Changer = Animal{}  // 编译错误！值不能实现指针接收者的方法
var c Changer = &Animal{}   // ✓ 指针可以

// 小接口原则（生产推荐）
type Storer interface {
    Get(key string) (any, error)
}
// 而不是把所有方法塞进一个接口

// 使用者定义接口（Go 最佳实践）
// 在消费方:
type UserGetter interface {
    GetUser(id int) (*User, error)
}
// 而不是在 model 包里定义 UserGetter`),
		},
	}
}

func deferTopic() Topic {
	return Topic{
		ID: "defer", Title: "Defer 延迟调用", Chapter: "01-basics",
		Sections: []Section{
			s("执行顺序: LIFO 栈",
				"defer 按后进先出（LIFO）顺序执行。编译时转换为链表，新 defer 插入头部，函数返回时从头部开始执行。**生产场景**：资源释放顺序很重要——先打开 A 再打开 B，释放时应该先释放 B 再释放 A，defer 天然保证这个顺序。**面试追问**：如果 defer 和 return 都有逻辑，谁先执行？→ return 的赋值先执行，然后 defer，最后 RET 指令。",
				`fmt.Println("start")
defer fmt.Println("defer 1")  // 最后执行
defer fmt.Println("defer 2")
defer fmt.Println("defer 3")  // 最先执行
fmt.Println("end")
// 输出: start → end → defer 3 → defer 2 → defer 1

// 生产: 资源释放顺序
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close()  // 后开先关

    scanner := bufio.NewScanner(f)
    // ...
    return nil
}`),
			s("参数求值时机",
				"**defer 语句声明时参数就被求值了**，不是执行时。但闭包引用的变量在执行时求值。**生产场景**：在循环中 defer 时经常踩坑——循环变量被闭包捕获，最终都打印最后一个值。Go 1.22 之前需要显式传参解决。**面试追问**：defer fmt.Println(x) 和 defer func() { fmt.Println(x) }() 的区别？→ 前者捕获 x 的当前值，后者在执行时读取 x 的值。",
				`x := 10
defer fmt.Printf("defer x = %d\n", x)  // 输出 10（声明时求值）
x = 20

// 闭包在执行时求值
y := 10
defer func() {
    fmt.Printf("defer y = %d\n", y)  // 输出 20（执行时求值）
}()
y = 20

// 生产坑: 循环中的 defer
for _, v := range values {
    defer func() {
        process(v)  // Go 1.22前: 全部打印最后一个值
    }()
}
// 修复（Go 1.22前）
for _, v := range values {
    v := v  // 创建新变量
    defer func() { process(v) }()
}`),
			s("defer 和 return 的执行顺序",
				"完整执行顺序：1) 返回值 = 表达式求值（赋值给返回值变量）→ 2) 执行 defer 函数（LIFO 顺序）→ 3) 执行 RET 指令返回给调用方。关键：defer 可以修改**命名返回值**，因为命名返回值是一个变量，defer 中可以访问和修改它。**匿名返回值**不受影响（已经赋值给临时变量）。**生产场景**：函数耗时统计（defer 中记录 time.Since）、panic 恢复并返回错误。**面试经典题**：func f() (r int) { defer func() { r = 1 }(); return 0 } 返回什么？→ 返回 1。",
				`// 匿名返回值
func f1() int {
    x := 1
    defer func() { x++ }()
    return x  // 返回 1（x 和返回值无关）
}

// 命名返回值
func f2() (x int) {
    defer func() { x++ }()
    return 1  // 返回 2（defer 修改了返回值 x）
}

// 生产: 函数耗时统计
func timedOperation() (result int, err error) {
    start := time.Now()
    defer func() {
        log.Printf("耗时: %v, err: %v", time.Since(start), err)
    }()
    // ...
    return 42, nil
}

// 生产: panic 恢复并返回错误
func safeCall() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    riskyOperation()
    return nil
}`),
			s("Open-coded 优化 (Go 1.14+)",
				"Go 1.14 引入 open-coded defer 优化：编译器直接 inline defer 调用，使用 bitmap 记录哪些 defer 需要执行，避免了 deferproc/deferreturn 的调用开销和 _defer 结构体的堆分配。**触发条件**：defer 数量 ≤ 8、不在循环中、没有 goto、没有 -N 禁用优化。**性能提升**：从 ~50ns 降到 ~5ns。**生产影响**：大部分 defer 场景（如 Close、Unlock）自动受益，无需改代码。**面试追问**：循环中的 defer 有什么问题？→ 无法使用 open-coded 优化，仍然走 deferproc 路径，且所有 defer 会堆积到函数返回时才执行（资源延迟释放）。",
				`// 这段代码在 Go 1.14+ 自动使用 open-coded 优化
defer fmt.Println("defer 1")  // 内联，不走 deferproc
defer fmt.Println("defer 2")
defer fmt.Println("defer 3")

// 循环中的 defer → 无法 open-coded（性能差）
for _, f := range files {
    file, _ := os.Open(f)
    defer file.Close()  // 所有文件都打开后才关闭！
}

// 正确: 提取到子函数
for _, f := range files {
    if err := processFile(f); err != nil {
        // handle error
    }
}
func processFile(name string) error {
    f, err := os.Open(name)
    if err != nil { return err }
    defer f.Close()  // 在子函数返回时立即关闭
    // ...
}`),
			s("recover 捕获 panic",
				"recover() 只能在 defer 函数中**直接调用**才有效。如果 recover() 被嵌套调用（如在 defer 内调用的普通函数中），返回 nil。goroutine 中的 panic 只能自己 recover，否则整个程序崩溃。**生产场景**：HTTP handler 中 panic 恢复（net/http 默认有 recover）；goroutine 启动时加 recover 防止整个进程崩溃。**最佳实践**：recover + 命名返回值 模式是 Go 中将 panic 转为 error 的惯用写法。",
				`// 正确: recover 在 defer 中直接调用
defer func() {
    if r := recover(); r != nil {
        log.Printf("recovered: %v\n%s", r, debug.Stack())
    }
}()

// 错误: recover 不在 defer 中直接调用
defer func() {
    doRecover()  // recover 返回 nil，无法捕获！
}()
func doRecover() { recover() }

// 生产: goroutine 中必须自己 recover
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("goroutine panic: %v", r)
        }
    }()
    // 如果不 recover，整个进程会崩溃！
    doWork()
}()

// 生产: HTTP middleware 中 panic 恢复
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}`),
		},
	}
}

func pointerTopic() Topic {
	return Topic{
		ID: "pointer", Title: "Pointer 指针", Chapter: "01-basics",
		Sections: []Section{
			s("值类型 vs 引用语义类型",
				"**值类型**（int/bool/string/array/struct）：赋值复制整个值，修改副本不影响原始值。**引用语义类型**（slice/map/channel/pointer/interface/function）：本质也是值传递，但内部包含指针，复制的是头部（24-16 字节），底层数据共享。**Go 中一切皆值传递，没有引用传递**。传指针也是值传递——复制的是地址值。**生产场景**：理解这个区别才能正确判断函数调用是否会修改原始数据。**面试追问**：string 是值类型还是引用类型？→ 值类型（赋值复制 header），但底层有指针（数据共享）。",
				`// 值类型: 互不影响
p1 := Point{1, 2}
p2 := p1    // 复制整个 struct
p2.X = 10   // p1.X 仍是 1

// slice: 共享底层数组
s1 := []int{1, 2, 3}
s2 := s1    // 只复制 header（指针+len+cap）
s2[0] = 100 // s1[0] 也变成 100

// 面试: 以下哪种是引用传递？
// func f(s []int)   → 值传递（复制 slice header）
// func f(m map[string]int) → 值传递（复制 map 指针）
// func f(p *int)    → 值传递（复制指针值）
// Go 中没有引用传递！`),
			s("new vs make",
				"**new(T)**：分配零值内存，返回 *T。分配在堆上，会被 GC 管理。**make(T)**：只用于 slice/map/channel，返回初始化后的 T（不是指针）。**为什么 make 不返回指针**：这三个类型本身就是引用语义的，内部已经包含指针，不需要再包一层。**生产建议**：new 在实际开发中很少使用，&MyStruct{} 比 new(MyStruct) 更清晰。make 是必须的——slice/map/channel 不 make 就用会 panic。**面试追问**：make([]int, 0) 和 make([]int, 0, 0) 的区别？→ 没区别，第三个参数是 cap。",
				`p := new(int)        // *int, 值为 0（零值）
cfg := new(Config)  // *Config, 所有字段零值

// new 很少用，以下等价且更清晰：
cfg := &Config{}

s := make([]int, 0, 10)   // slice（返回 []int，不是 *[]int）
m := make(map[string]int) // map（返回 map[string]int）
ch := make(chan int, 5)   // channel（返回 chan int）

// 面试: 以下会 panic 吗？
var s []int
s[0] = 1  // panic！nil slice 不能写入
// 需要先 make
s = make([]int, 1)
s[0] = 1  // OK`),
			s("何时用指针",
				"**用指针**：1) 需要修改原始值（最常见）；2) 结构体较大（> 64 bytes，避免复制开销）；3) 一致性（某个方法需要指针，其他也统一用指针）。**用值**：1) 小结构体（<= 64 bytes，复制比指针解引用+逃逸更快）；2) 不可变数据（值传递天然安全）；3) 需要 map key 或做比较。**64 bytes 分界线的理由**：缓存行通常 64 bytes，小结构体复制在缓存行内完成，比指针追快。**生产场景**：方法的接收者选择是团队规范的重要部分。**面试追问**：如果 receiver 是指针，赋给接口时必须用指针？→ 是的。",
				`// 大结构体 → 指针（避免复制 1KB）
type BigConfig struct {
    Data [1024]byte
    Name string
}
func process(b *BigConfig) { b.Name = "updated" }

// 小结构体 → 值（复制更快，不逃逸）
type Point struct{ X, Y float64 }
func distance(p Point) float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// 一致性原则: 如果一个方法用指针，全部用指针
type User struct { Name string }
func (u *User) SetName(n string) { u.Name = n }  // 需要修改
func (u *User) GetName() string { return u.Name } // 也用指针（一致性）

// 面试: 以下哪种更好？
// func (u User) GetName() string   // 小 struct，值接收者也行
// func (u *User) GetName() string  // 但如果 SetName 用指针，统一用指针`),
		},
	}
}

