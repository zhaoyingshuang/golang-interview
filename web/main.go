package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

type Topic struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Chapter  string    `json:"chapter"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Code    string `json:"code"`
}

func main() {
	// Static site generation mode
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		outDir := "docs"
		if len(os.Args) > 2 {
			outDir = os.Args[2]
		}
		generateStaticSite(outDir)
		return
	}

	port := "9090"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	staticFS, _ := fs.Sub(staticFiles, "static")
	http.Handle("/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/api/topics", handleTopics)
	http.HandleFunc("/api/topic/", handleTopic)

	fmt.Printf("Go 面试知识点学习平台已启动: http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}
}

func generateStaticSite(outDir string) {
	topics := getAllTopics()
	listData, _ := json.Marshal(getTopicList())
	detailMap := make(map[string]Topic)
	for _, t := range topics {
		detailMap[t.ID] = t
	}
	detailData, _ := json.Marshal(detailMap)

	tpl, _ := staticFiles.ReadFile("static/index.html")
	html := string(tpl)

	inject := fmt.Sprintf("\nconst TOPICS_LIST = %s;\nconst TOPICS_DETAIL = %s;\n", string(listData), string(detailData))

	// Inject data before first <script>
	html = strings.Replace(html, "<script>\nconst chapters", inject+"<script>\nconst chapters", 1)

	// Replace API fetch with static data
	html = strings.Replace(html, "async function init() {", "function init() {", 1)
	html = strings.Replace(html, "async function loadTopic", "function loadTopic", 1)
	html = strings.Replace(html, "const res = await fetch('/api/topics');\n    topics = await res.json();", "    topics = TOPICS_LIST;", 1)
	html = strings.Replace(html, "const res = await fetch(`/api/topic/${id}`);\n    const data = await res.json();", "    const data = TOPICS_DETAIL[id];", 1)

	// Add base tag for GitHub Pages sub-path
	html = strings.Replace(html, `<meta charset="UTF-8">`,
		`<meta charset="UTF-8">`+"\n    "+`<base href="/golang-interview/">`, 1)

	// Fix hash routing for sub-path
	html = strings.Replace(html, "history.replaceState(null, '', '#' + id);",
		"history.replaceState(null, '', location.pathname + '#' + id);", 1)
	html = strings.Replace(html, "if (location.hash) {\n        loadTopic(location.hash.slice(1));\n    }",
		"if (location.hash) { loadTopic(location.hash.slice(1)); }", 1)

	os.MkdirAll(outDir, 0755)
	os.WriteFile(outDir+"/index.html", []byte(html), 0644)
	fmt.Printf("Static site generated in %s/\n", outDir)
}

func handleTopics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getTopicList())
}

func handleTopic(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/topic/")
	for _, t := range getAllTopics() {
		if t.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	http.NotFound(w, r)
}

func getTopicList() []map[string]string {
	topics := getAllTopics()
	result := make([]map[string]string, len(topics))
	for i, t := range topics {
		result[i] = map[string]string{
			"id":      t.ID,
			"title":   t.Title,
			"chapter": t.Chapter,
		}
	}
	return result
}

func getAllTopics() []Topic {
	return []Topic{
		sliceTopic(),
		mapTopic(),
		stringTopic(),
		interfaceTopic(),
		deferTopic(),
		pointerTopic(),
		goroutineTopic(),
		channelTopic(),
		syncTopic(),
		contextTopic(),
		patternTopic(),
		gcTopic(),
		escapeTopic(),
		alignmentTopic(),
		profilingTopic(),
		benchmarkTopic(),
	}
}

func s(title, content, code string) Section {
	return Section{Title: title, Content: content, Code: code}
}

// ==================== 01-basics ====================

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

func goroutineTopic() Topic {
	return Topic{
		ID: "goroutine", Title: "Goroutine 协程", Chapter: "02-concurrency",
		Sections: []Section{
			s("GMP 调度模型",
				"**G**=Goroutine（用户态协程，初始栈 2KB），**M**=Machine（OS 线程），**P**=Processor（逻辑处理器，数量=GOMAXPROCS，默认等于 CPU 核心数）。**为什么需要 P**：Go 早期只有 GM 模型（全局队列），问题：全局锁竞争激烈、缓存不友好。引入 P 后每个 P 有独立队列（256 个 G），减少了锁竞争。**调度流程**：新 G 优先放当前 P 的本地队列 → 满了一半放入全局队列 → M 绑定 P 后从本地队列取 G → 本地空则从全局取或从其他 P 偷（work stealing）→ 系统调用阻塞时 M 释放 P 让其他 M 继续跑。**生产影响**：GOMAXPROCS 默认等于 CPU 核心数，CPU 密集型任务不需要修改；I/O 密集型可以适当增大。",
				`G (Goroutine) - 用户态协程, 初始栈 2KB, 可增长到 1GB
M (Machine)   - OS 线程, 由操作系统调度
P (Processor) - 逻辑处理器, 默认数=CPU核心数

调度流程:
1. 新 G → 当前 P 的本地队列（最多 256 个）
2. 本地队列满 → 前一半放入全局队列
3. M 绑定 P, 从 P 本地队列取 G 执行
4. 本地空 → 从全局取一批 / 从其他 P 偷(work stealing)
5. G 系统调用阻塞 → M 释放 P, P 绑新 M 继续

// 查看当前状态
runtime.NumCPU()         // CPU 核心数
runtime.GOMAXPROCS(0)    // P 的数量
runtime.NumGoroutine()   // 当前 goroutine 数`),
			s("Goroutine vs 线程",
				"**栈大小**：2KB（可增长到 1GB）vs 1-8MB（固定）。**创建开销**：~0.3µs vs ~10-100µs（goroutine 只需分配栈和几个结构体）。**切换开销**：~100ns（用户态，只保存 ~3 个寄存器）vs ~1-10µs（内核态，需要保存/恢复完整上下文）。**最大数量**：轻松上百万 vs 几千到几万。**通信方式**：channel（CSP 模型）vs 共享内存+锁。**生产影响**：Go 可以在一个 HTTP 请求处理中启动多个 goroutine 并发查询数据库/缓存，而 Java/C++ 做同样的事需要线程池。**面试追问**：goroutine 和 coroutine 的区别？→ goroutine 支持抢占式调度（Go 1.14+），不需要显式 yield。",
				`| 特性         | Goroutine      | OS Thread      |
|-------------|----------------|----------------|
| 栈大小       | 2KB（可增长到1GB）| 1-8MB（固定）   |
| 创建开销     | ~0.3µs         | ~10-100µs      |
| 切换开销     | ~100ns 用户态   | ~1-10µs 内核态  |
| 最大数量     | 百万级          | 几千            |
| 调度方式     | 协作+抢占       | 抢占式          |
| 通信方式     | channel(CSP)   | 共享内存+锁     |

// 实测: 创建 10 万个 goroutine
start := time.Now()
var wg sync.WaitGroup
wg.Add(100000)
for i := 0; i < 100000; i++ {
    go func() { defer wg.Done() }()
}
wg.Wait()
fmt.Println(time.Since(start)) // ~30ms`),
			s("栈增长",
				"Go 1.3+ 使用**连续栈**（contiguous stack）：空间不够时分配 2 倍新栈，复制数据，释放旧栈。解决了分段栈的 \"hot split\" 问题（栈在边界频繁增减导致性能抖动）。GC 时如果栈使用率 < 1/4，缩减为原来的一半。初始 2KB，最大 1GB。**生产影响**：递归函数不需要担心栈溢出（除非超过 1GB）；goroutine 初始只有 2KB，可以轻松创建百万个。**面试追问**：栈增长时地址会变吗？→ 会，因为分配了新的连续内存，Go 会自动更新所有指向旧栈的指针。",
				`// 连续栈机制:
// 空间不足 → 分配 2x 新栈 → 复制数据 → 更新指针 → 释放旧栈
// GC 时使用率 < 25% → 缩减为原来的一半
// 初始: 2KB, 最大: 1GB

// 面试: 栈增长时会暂停 goroutine 吗？
// 会，但只暂停当前 goroutine，不影响其他

// 面试: 什么操作会导致栈增长？
// 函数调用时检查栈空间是否足够
// 递归调用是最常见的触发场景`),
			s("Goroutine 泄漏",
				"**最常见的 goroutine 泄漏原因**：1) 无缓冲 channel 没有接收者/发送者（goroutine 永远阻塞）；2) context 没有 cancel（goroutine 永远不会退出）；3) range channel 没有关闭（消费者永远等待）；4) WaitGroup 计数错误（Wait 永远阻塞）。**生产检测**：runtime.NumGoroutine() 监控数量趋势；pprof goroutine profile 查看阻塞的调用栈；监控告警 goroutine 数量持续增长。**修复模式**：所有 goroutine 都要有明确的退出条件。",
				`// 泄漏场景1: channel 没有接收者
ch := make(chan int)
go func() { ch <- result }()  // 没人接收，永远阻塞
// 修复: 缓冲 channel 或 context 取消

// 泄漏场景2: context 没有 cancel
go func() {
    for {
        select {
        case <-ch:
            // 处理
        // 缺少 case <-ctx.Done()！永远不会退出
        }
    }
}()
// 修复: 加上 ctx.Done() 分支

// 泄漏场景3: range channel 没有关闭
go func() {
    for v := range ch {  // ch 永远不关闭
        process(v)
    }  // 永远到不了这里
}()
// 修复: 发送方 close(ch)

// 检测 goroutine 泄漏
fmt.Printf("goroutines: %d\n", runtime.NumGoroutine())
// 或用 pprof
pprof.Lookup("goroutine").Count()`),
			s("抢占调度",
				"Go 1.14 之前：协作式抢占（编译器在函数入口插入栈检查），紧凑循环（无函数调用）无法被抢占，可能饿死其他 goroutine。Go 1.14+：**基于信号的异步抢占**（SIGURG），信号处理器中设置抢占标志，goroutine 在安全点被挂起。**生产影响**：纯计算循环不再需要手动加 runtime.Gosched()。**调度时机总结**：channel 操作、系统调用、time.Sleep、runtime.Gosched()、GC STW、函数调用栈检查、信号抢占。",
				`// Go 1.14 前: 紧凑循环会饿死其他 goroutine
go func() {
    for i := 0; i < 1e10; i++ {
        _ = i * i  // 无函数调用，无法被抢占！
    }
}()

// Go 1.14+: 基于信号抢占，上面的代码也能被抢占

// 调度时机（面试总结）:
// 1. channel 操作（发送/接收阻塞）
// 2. 系统调用（文件/网络 IO）
// 3. time.Sleep
// 4. runtime.Gosched()（主动让出）
// 5. GC STW
// 6. 函数调用时的栈检查（协作式）
// 7. SIGURG 信号（异步抢占，Go 1.14+）`),
		},
	}
}

func channelTopic() Topic {
	return Topic{
		ID: "channel", Title: "Channel 通道", Chapter: "02-concurrency",
		Sections: []Section{
			s("底层结构 hchan",
				"Channel 底层是 runtime/chan.go 中的 hchan 结构体：一个**带锁的环形队列**。sendx/recvx 是环形队列的读写指针，recvq/sendq 是阻塞等待的 goroutine 链表（sudog 结构）。所有操作（send/recv/close）都需要加锁（runtime 级别的 mutex）。**无缓冲 channel**：dataqsiz=0，没有缓冲区，发送和接收必须同时就绪。**有缓冲 channel**：dataqsiz>0，环形缓冲区，发送方可以异步写入。**生产选择**：无缓冲用于同步信号（done、quit）；有缓冲用于数据传递（任务队列、结果收集）。",
				`type hchan struct {
    qcount   uint           // 队列中元素数
    dataqsiz uint           // 环形队列容量（缓冲大小）
    buf      unsafe.Pointer // 环形队列指针
    closed   uint32         // 是否已关闭（0=未关闭，1=已关闭）
    sendx    uint           // 发送索引
    recvx    uint           // 接收索引
    recvq    waitq          // 等待接收的 goroutine 队列
    sendq    waitq          // 等待发送的 goroutine 队列
    lock     mutex          // 互斥锁（所有操作都要加锁）
}

// 面试: channel 发送一个数据的完整流程？
// 1. 加锁
// 2. 如果 closed → 解锁, panic
// 3. 如果 recvq 有等待者 → 直接发送(不经缓冲区)
// 4. 如果缓冲区有空位 → 放入缓冲区
// 5. 否则 → 阻塞当前 goroutine, 加入 sendq
// 6. 解锁`),
			s("发送和接收规则",
				"发送优先级：1) recvq 有等待者 → 直接发给它（不经缓冲区）；2) 缓冲区有空位 → 放入缓冲区；3) 否则阻塞。接收优先级：1) sendq 有等待者 → 直接取或从缓冲取；2) 缓冲区有数据 → 取；3) 否则阻塞。**面试必记三条**：向已关闭 channel 发送 → panic；关闭已关闭的 channel → panic；从已关闭 channel 接收 → 返回缓冲区剩余数据，之后返回零值+false。**生产模式**：close(ch) 用于广播通知（所有接收者都能收到零值）。",
				`// 关闭后的行为（面试必考）
ch := make(chan int, 2)
ch <- 1
ch <- 2
close(ch)

// 继续接收: 先取完缓冲区
v, ok := <-ch  // v=1, ok=true
v, ok = <-ch   // v=2, ok=true
v, ok = <-ch   // v=0, ok=false（缓冲区空了）

// ch <- 3      // panic: send on closed channel
// close(ch)    // panic: close of closed channel

// 生产: close 用于广播退出信号
quit := make(chan struct{})
go func() {
    defer close(quit) // 所有接收者都会收到
    doWork()
}()
<-quit // 等待完成`),
			s("select 行为",
				"select 的规则：1) 所有 case 会随机排序（避免某个 case 饥饿）；2) 按顺序评估哪些 case 可以立即执行；3) 多个 case 就绪时**随机选一个**（公平性）；4) 没有就绪且有 default → 执行 default（非阻塞）；5) 没有就绪且无 default → 阻塞。**空 select**：select{} 永远阻塞。**生产场景**：非阻塞收发（default 模式）、超时控制（time.After）、多路复用（多个 channel 同时监听）。**面试追问**：select 中 case 的执行顺序？→ 多个就绪时随机，不是按书写顺序。",
				`// 非阻塞接收
select {
case v := <-ch:
    fmt.Println(v)
default:
    fmt.Println("无数据")
}

// 超时控制（生产常用）
select {
case result := <-ch:
    return result
case <-time.After(5 * time.Second):
    return errors.New("timeout")
}

// 多路复用（生产: 多个数据源）
select {
case msg := <-ch1:
    handleType1(msg)
case msg := <-ch2:
    handleType2(msg)
case <-ctx.Done():
    return ctx.Err()
}

// 面试: 空 select 的作用？
select {} // 永远阻塞，常用于 main 中防退出`),
			s("Channel vs Mutex",
				"Go 的哲学：\"Don't communicate by sharing memory; share memory by communicating.\" 但不是所有场景都该用 channel。**Channel 适合**：传递数据所有权（生产者-消费者）、等待/通知（done signal）、超时控制、pipeline。**Mutex 适合**：保护共享缓存（LRU cache）、计数器、配置读写、并发安全地操作复杂数据结构。**生产选择原则**：goroutine 之间传递数据用 channel；多个 goroutine 读写同一块数据用 mutex。不要为了\"Go 风格\"硬用 channel——简单场景 mutex 更清晰。",
				`// Channel: 传递数据所有权
func producer(ch chan<- int) {
    for i := 0; i < 100; i++ {
        ch <- i  // 数据所有权转移给消费者
    }
    close(ch)
}

// Mutex: 保护共享状态
type SafeCache struct {
    mu   sync.RWMutex
    data map[string]string
}
func (c *SafeCache) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

// 生产: 常见错误 - 用 channel 做本该用 mutex 的事
// ❌ 通过 channel 传递"请修改这个变量"的消息
// ✓ 直接用 mutex 保护变量`),
		},
	}
}

func syncTopic() Topic {
	return Topic{
		ID: "sync", Title: "Sync 同步原语", Chapter: "02-concurrency",
		Sections: []Section{
			s("Mutex 正常/饥饿模式",
				"**正常模式**：新来的 goroutine 和刚被唤醒的竞争锁（新来的有优势，因为已经在 CPU 上运行）。如果竞争失败，被放到队列前面。**饥饿模式**：某个 goroutine 等待超过 1ms 时触发，锁直接交给队列头部的 goroutine（不竞争），防止饿死。当最后一个等待者获取锁或等待时间 < 1ms，切回正常模式。**不可重入**：Go 的 Mutex 不是可重入锁，同一 goroutine 重复 Lock 会死锁。**生产场景**：保护计数器、缓存、配置。**面试追问**：为什么不支持可重入？→ 容易导致设计问题（函数是否持有锁变得不透明）。",
				`// Mutex 状态位（面试了解）
// state: 0=未锁定, 1=已锁定, 2=已唤醒, 4=饥饿

var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()

// 不可重入！以下会死锁
mu.Lock()
mu.Lock()  // deadlock!

// 生产: defer 释放锁
func (s *Service) Update() {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ...
}

// 生产: TryLock（Go 1.18+，非阻塞）
if mu.TryLock() {
    // 获取成功
    mu.Unlock()
} else {
    // 获取失败
}`),
			s("RWMutex",
				"多个读者可同时持有读锁，写者独占。内部结构：1 个 Mutex + 2 个信号量 + 2 个计数器。readerCount 为负数时表示有写者在等待。**适用场景**：读多写少（如配置热更新、缓存）。**不适用**：读写都频繁（RWMutex 比 Mutex 更重，内部有更多 atomic 操作）。**生产坑**：读锁未释放就加写锁 → 死锁。**面试追问**：RWMutex 的写锁饥饿问题？→ Go 的 RWMutex 有防止写锁饥饿的机制（新读者在有写者等待时会被阻塞）。",
				`var rw sync.RWMutex
data := make(map[string]string)

// 读操作: 多个 goroutine 可以并行读
func Get(key string) string {
    rw.RLock()
    defer rw.RUnlock()
    return data[key]
}

// 写操作: 独占
func Set(key, value string) {
    rw.Lock()
    defer rw.Unlock()
    data[key] = value
}

// 生产坑: 读锁中调用写锁 → 死锁
rw.RLock()
rw.Lock()   // deadlock!
// 解决: 确保读写锁不嵌套`),
			s("WaitGroup",
				"WaitGroup 用于等待一组 goroutine 完成。state1 高 32 位是计数器，低 32 位是等待者数量。**最关键的使用规则**：Add 必须在 go 之前调用（在外部），Done 在 goroutine 内部 defer 调用。如果 Add 放在 goroutine 内部，Wait 可能在所有 Add 之前就返回。**生产场景**：批量并发请求、并行数据处理。**面试追问**：Add 的计数能变负吗？→ 不能，会 panic。",
				`var wg sync.WaitGroup

// 正确用法: Add 在 go 之前
for i := 0; i < 10; i++ {
    wg.Add(1)  // 在 go 之前！
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}
wg.Wait()

// 错误用法: Add 在 goroutine 内部
for i := 0; i < 10; i++ {
    go func() {
        wg.Add(1)    // ❌ 可能 Wait 已经执行了
        defer wg.Done()
    }()
}
wg.Wait()  // 可能在所有 Add 之前就返回

// 生产: 批量并发请求
func fetchAll(urls []string) []Result {
    var wg sync.WaitGroup
    results := make([]Result, len(urls))
    for i, url := range urls {
        wg.Add(1)
        go func(idx int, u string) {
            defer wg.Done()
            results[idx] = fetch(u)
        }(i, url)
    }
    wg.Wait()
    return results
}`),
			s("sync.Once",
				"保证某个操作只执行一次。实现：1) 原子加载 done 标志（快速路径）；2) 加锁；3) double-check done；4) 执行 f()；5) 原子存储 done=1。**Go 1.21+ 提供 OnceValues** 支持返回值。**注意**：如果 f() panic，done 不会被设置为 1，后续调用会再次执行 f()。**生产场景**：初始化数据库连接、加载配置、单例模式。**面试追问**：Once 和 init 的区别？→ init 在包加载时执行，Once 在首次调用时执行（延迟初始化）。",
				`var once sync.Once
var config *Config

func GetConfig() *Config {
    once.Do(func() {
        // 只执行一次（延迟初始化）
        config = loadConfig()
    })
    return config
}

// Go 1.21+ OnceValues（带返回值）
var loadDB = sync.OnceValues(func() (*sql.DB, error) {
    return sql.Open("mysql", dsn)
})
db, err := loadDB()  // 首次调用时初始化，后续调用返回缓存

// 面试: 如果 f() panic 会怎样？
// done 不会被设为 1，下次调用会再执行 f()`),
			s("sync.Pool",
				"对象复用池，减少 GC 压力。每个 P 有本地池（无锁），Get 先从本地取，取不到从其他 P 偷，再取不到调用 New 创建。**关键特性：GC 时会清理池中对象**，所以不适合做连接池！**适用场景**：高频创建和销毁的临时对象（bytes.Buffer、JSON 编码器）。**标准库使用**：fmt.Printf 内部用 Pool 复用 buffer；encoding/json 用 Pool 复用编码器。**面试追问**：Pool 和连接池的区别？→ Pool 的对象会在 GC 时丢失，连接池需要自己管理生命周期。",
				`var bufPool = sync.Pool{
    New: func() any {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func Process(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()          // 重置（但不释放内存）
    defer bufPool.Put(buf)

    buf.Write(data)
    return buf.String()
}

// 生产: JSON 编码复用
var encoderPool = sync.Pool{
    New: func() any { return json.NewEncoder(io.Discard) },
}

// 注意: GC 时池中对象会被清理
// 不适合做数据库连接池！`),
			s("Atomic 操作",
				"底层使用 CPU 原子指令（LOCK 前缀 + CMPXCHG 等），不需要加锁，性能极高。Go 1.19+ 提供 **atomic.Int64、atomic.Bool、atomic.Pointer[T]** 等类型，比旧的 atomic.AddInt64 等函数更易用。**CAS（CompareAndSwap）** 是无锁编程的基础。**适用场景**：计数器、状态标志、无锁队列。**不适用**：复杂操作（多步非原子）、需要等待的场景（用 channel/mutex）。**生产场景**：请求计数、服务状态标记（atomic.Bool）、无锁缓存（atomic.Pointer）。",
				`// atomic.Int64（Go 1.19+，推荐）
var count atomic.Int64
count.Add(1)
v := count.Load()
count.Store(0)

// CAS: 无锁更新
var state atomic.Int64
state.Store(0)  // 0=idle, 1=running
if state.CompareAndSwap(0, 1) {
    // 从 idle 切到 running，只有一方能成功
    defer state.Store(0)
    doWork()
}

// atomic.Pointer[T]（Go 1.19+）
var cache atomic.Pointer[Config]
cache.Store(&Config{Version: 1})
cfg := cache.Load() // 无锁读取

// atomic.Bool
var ready atomic.Bool
ready.Store(true)
if ready.Load() { /* ... */ }`),
		},
	}
}

func contextTopic() Topic {
	return Topic{
		ID: "context", Title: "Context 上下文", Chapter: "02-concurrency",
		Sections: []Section{
			s("Context 接口",
				"Context 是一个接口：Deadline() 返回截止时间，Done() 返回取消信号 channel，Err() 返回取消原因，Value() 返回请求作用域的值。4 种内部实现：**emptyCtx**（根 context，永远不会取消）、**cancelCtx**（可取消）、**timerCtx**（带超时，内部包含 cancelCtx）、**valueCtx**（带值，链表结构）。**生产场景**：每个 HTTP 请求都有自己的 context（req.Context()），请求结束后自动取消。**面试追问**：context.Background() 和 context.TODO() 的区别？→ Background 是根 context，TODO 是不确定该用什么时的占位。",
				`type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}

// 4 种实现:
// emptyCtx  → context.Background() / TODO()
// cancelCtx → context.WithCancel()
// timerCtx  → context.WithTimeout() / WithDeadline()
// valueCtx  → context.WithValue()

// 面试: context 作为函数参数的规范？
// 必须是第一个参数，不要放在 struct 中
func Process(ctx context.Context, data string) error {
    // ...
}`),
			s("取消传播",
				"**取消是单向传播的**：父取消 → 子自动取消；子取消 → 不影响父和兄弟。cancel() 的内部流程：1) 设置 err；2) 关闭 done channel；3) 遍历 children 依次取消；4) 从父的 children 中移除自己。**生产场景**：用户取消请求 → 所有下游数据库查询、RPC 调用自动取消。**最佳实践**：cancel() 必须 defer 调用，即使确认会超时也要 cancel（释放内部资源）。",
				`ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 创建子 context
child1, cancel1 := context.WithCancel(ctx)
defer cancel1()
child2, _ := context.WithCancel(ctx)
grandchild, _ := context.WithCancel(child1)

cancel1()
// child1.Err() → context.Canceled
// grandchild.Err() → context.Canceled（子被取消）
// child2.Err() → nil（兄弟不受影响）

// 生产: HTTP handler 中
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // 请求结束后自动取消
    result, err := fetchData(ctx, query)
    // 如果用户断开连接，ctx 自动取消
}`),
			s("WithTimeout/WithDeadline",
				"**WithTimeout** 从当前时间计时，**WithDeadline** 指定绝对时间点。WithTimeout 就是 WithDeadline(ctx, time.Now().Add(timeout))。**必须 defer cancel()**：即使已经超时，cancel 也必须调用以释放内部 timer goroutine，否则会 goroutine 泄漏。**生产场景**：数据库查询超时（通常 3-5s）、RPC 调用超时（通常 1-10s）、HTTP 客户端超时。**面试追问**：超时时间应该设多少？→ 取决于业务，但要小于上游的超时时间。",
				`// 数据库查询超时
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()  // 必须！防止 goroutine 泄漏
rows, err := db.QueryContext(ctx, "SELECT ...")

// RPC 调用超时
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
resp, err := client.Call(ctx, request)

// 面试: 不调用 cancel 会怎样？
// timer goroutine 会一直运行，直到超时
// 如果 timeout 很长（甚至没有），goroutine 就泄漏了`),
			s("WithValue 最佳实践",
				"Key 用**自定义类型**（避免冲突），不要用内建类型（string/int）。**适合传递**：trace ID、auth token、request-scoped logger。**不适合传递**：数据库连接、业务参数、函数必须的参数（应该用函数参数）。**生产反模式**：把所有东西都塞进 context → 应该只放请求级别的元数据。**面试追问**：context.Value 的查找效率？→ O(n)，沿 valueCtx 链表向上查找，层数多了会慢。",
				`// 正确: 自定义 key 类型
type requestIDKey struct{}
type userIDKey struct{}

ctx = context.WithValue(ctx, requestIDKey{}, "req-123")
ctx = context.WithValue(ctx, userIDKey{}, 42)

// 类型安全地获取
id, _ := ctx.Value(requestIDKey{}).(string)
uid, _ := ctx.Value(userIDKey{}).(int)

// 生产: middleware 中设置 trace ID
func TraceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := uuid.New().String()
        ctx := context.WithValue(r.Context(),
            requestIDKey{}, traceID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 面试: 为什么不用 string 做 key？
// 不同包可能用相同的 string，导致冲突
// type myKey string → 也行，但空 struct 更安全`),
		},
	}
}

func patternTopic() Topic {
	return Topic{
		ID: "pattern", Title: "并发模式", Chapter: "02-concurrency",
		Sections: []Section{
			s("Worker Pool",
				"控制并发数量，避免创建过多 goroutine。模式：jobs channel 分发任务，N 个 worker 通过 range jobs 消费，results channel 收集结果。**生产场景**：批量处理 API 请求、并发下载文件、并行数据处理。**为什么不用 go + WaitGroup**：Worker Pool 复用 goroutine，避免频繁创建销毁；通过 worker 数量精确控制并发。**面试追问**：worker 数量设多少？→ CPU 密集型 = CPU 核心数，I/O 密集型可以更多（10-100）。",
				`func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup
    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := range jobs {  // range 直到 close
                results <- process(j)
            }
        }()
    }
    go func() { wg.Wait(); close(results) }()
}

// 使用
jobs := make(chan Job, 100)
results := make(chan Result, 100)
go workerPool(jobs, results, runtime.NumCPU())

for _, j := range allJobs { jobs <- j }
close(jobs)
for r := range results { /* 收集结果 */ }`),
			s("Fan-out / Fan-in",
				"**Fan-out**：一个 channel 分发给多个 goroutine 处理，提高吞吐。**Fan-in**：多个 channel 合并到一个，统一消费。**和 Worker Pool 的区别**：Fan-out 的每个 worker 可以有不同的处理逻辑；Worker Pool 的 worker 相同。**生产场景**：日志处理（多种日志源合并）、数据聚合（从多个 API 获取数据后合并）。**面试追问**：merge 函数为什么要单独启动 goroutine 关闭 out？→ 要等所有 source channel 关闭后才能关闭 out。",
				`// Fan-out: 多个 worker 并行处理同一个输入
func fanOut(in <-chan int, workers int) []<-chan int {
    outs := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        outs[i] = process(in)  // 每个 worker 读同一个 in
    }
    return outs
}

// Fan-in: 合并多个 channel
func merge(chs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, ch := range chs {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c { out <- v }
        }(ch)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}`),
			s("Pipeline",
				"数据在多个 stage 之间流过，每个 stage 是一个 goroutine + channel。stage1 → stage2 → stage3 → 结果。每个 stage 通过 context 支持取消，任何 stage 出错都能优雅退出。**生产场景**：数据 ETL（读取 → 转换 → 写入）、图片处理（读取 → 缩放 → 水印 → 保存）、日志分析（读取 → 解析 → 过滤 → 聚合）。**优势**：每个 stage 可以独立伸缩（不同的 goroutine 数量），解耦清晰。",
				`// Pipeline: gen → square → filter → 结果
func gen(ctx context.Context, values ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, v := range values {
            select {
            case out <- v:
            case <-ctx.Done(): return
            }
        }
    }()
    return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * v:
            case <-ctx.Done(): return
            }
        }
    }()
    return out
}

// 使用: pipeline 组合
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
p := filter(ctx, square(ctx, gen(ctx, 1, 2, 3, 4, 5)))
for v := range p { fmt.Println(v) }`),
			s("优雅退出 (Graceful Shutdown)",
				"模式：context/信号控制生命周期 → 收到退出信号后停止接收新请求 → 等待进行中的请求完成 → 关闭资源。用 signal.Notify + context.WithCancel 实现信号监听，用 WaitGroup 或 errgroup 等待所有 goroutine 退出。**生产场景**：HTTP 服务收到 SIGTERM 后优雅退出（K8s Pod 滚动更新）。**面试追问**：为什么不直接 os.Exit？→ 进行中的请求会被截断，可能导致数据不一致。",
				`func main() {
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: ":8080"}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    <-ctx.Done() // 等待信号
    log.Println("shutting down...")

    // 优雅关闭：5s 超时
    shutdownCtx, cancel := context.WithTimeout(
        context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(shutdownCtx) // 等待进行中的请求完成
    log.Println("server stopped")
}`),
		},
	}
}

func gcTopic() Topic {
	return Topic{
		ID: "gc", Title: "GC 垃圾回收", Chapter: "03-memory",
		Sections: []Section{
			s("三色标记算法",
				"Go 使用**并发三色标记-清除**算法。白色=未标记（将被回收），灰色=已标记但引用未扫描，黑色=已标记且引用已扫描。流程：1) 根对象→灰色 2) 取灰色扫描引用→引用变灰色 3) 当前→黑色 4) 重复直到无灰色 5) 白色=垃圾回收。**根对象**包括：全局变量、当前所有 goroutine 的栈变量。**三色不变式**：保证不会误删活跃对象。**生产影响**：GC 和用户代码并发执行（大部分时间），只有标记开始和结束时极短暂的 STW。",
				`白色 (White)  - 未标记, 将被回收
灰色 (Gray)   - 已标记, 引用未扫描
黑色 (Black)  - 已标记, 引用已扫描

标记过程:
1. STW: 开启写屏障
2. 根对象(stack, global) → 灰色
3. 并发标记(和用户代码并行):
   - 取灰色对象, 扫描其引用
   - 引用的对象 → 灰色
   - 当前对象 → 黑色
   - 重复直到无灰色
4. STW: 关闭写屏障, 清理
5. 所有白色对象 = 垃圾, 回收内存`),
			s("混合写屏障",
				"为什么需要写屏障？三色标记和用户代码并发执行时，可能出现黑色对象指向白色对象的引用（遗漏标记）。**Go 1.5-1.7**：Dijkstra 插入写屏障（A.field=B 时 B 标灰色），需要 STW 重新扫描栈（10-100ms）。**Go 1.8+**：混合写屏障 = Dijkstra + Yuasa。A.field=B 时：1) B 标灰色（Dijkstra）；2) 如果 A 在栈上，原值也标灰色（Yuasa）。结果：**不需要 STW 重新扫描栈**，总 STW < 100µs。**生产影响**：写屏障有 3-5% 的性能开销，但在 GC 期间才启用。",
				`// Go 1.8+ 混合写屏障
// 写操作: A.field = B
1. B 标记为灰色 (Dijkstra 插入写屏障)
2. 如果 A 在栈上, A.field 原值标记为灰色 (Yuasa)

// 为什么栈需要特殊处理？
// 栈上操作极其频繁, 加写屏障开销大
// 混合写屏障让栈上操作几乎零开销

// GC 各阶段:
// 1. 标记准备 (STW ~20µs): 开启写屏障
// 2. 并发标记 (~ms级): 和用户代码并行
// 3. 标记终止 (STW ~40µs): 关闭写屏障
// 4. 并发清除: 回收白色对象`),
			s("GC 触发条件与 GOGC 调优",
				"三种触发：1) **堆内存增长**达到 (1+GOGC/100)×上次存活堆（最常见）；2) **2 分钟**未触发 GC（sysmon 强制）；3) **手动** runtime.GC()。**GOGC 调优**：默认 100（堆翻倍时 GC）；200 更少 GC 但更多内存；50 更多 GC 更少内存。**Go 1.19+ GOMEMLIMIT**：设置运行时总内存软上限，比 GOGC 更直观。**生产建议**：容器环境用 GOMEMLIMIT（如容器限制 512MB → 设 450MB），让 Go 自己决定 GC 时机。",
				`// GOGC 调优
GOGC=100  // 默认, 堆翻倍时 GC（平衡）
GOGC=200  // 更少 GC, 更多内存（CPU密集型）
GOGC=50   // 更多 GC, 更少内存（内存敏感）
GOGC=off  // 禁用 GC（不推荐）

// Go 1.19+ GOMEMLIMIT（推荐）
debug.SetMemoryLimit(450 << 20) // 450MB 软上限
// 配合 GOGC=off → 固定内存上限
// 或 GOGC=默认 → 平衡 GC 和内存

// 生产: K8s 容器
// resources.limits.memory: 512Mi
// GOMEMLIMIT=450MiB (留余量给 Go runtime)
// GOGC=100 (默认即可)`),
			s("观察 GC",
				"**GODEBUG=gctrace=1**：打印每次 GC 详情，包括 STW 时间、堆大小变化、GC 占 CPU 比。**runtime.ReadMemStats**：代码中获取详细内存统计。**pprof**：heap profile 分析内存分配热点。**生产监控**：Prometheus client_golang 自动暴露 GC 指标（go_gc_duration_seconds、go_memstats_alloc_bytes 等）。**面试追问**：如何减少 GC 压力？→ 减少堆分配（逃逸分析、sync.Pool、预分配）。",
				`// 环境变量方式（最快）
// GODEBUG=gctrace=1 go run main.go
// gc 1 @0.003s 0%: 0.018+0.45+0.003 ms clock
// 含义:
// gc 1: 第1次GC
// @0.003s: 程序启动后0.003s
// 0.018ms: STW(标记开始)
// 0.45ms: 并发标记
// 0.003ms: STW(标记结束)
// 4->4->0 MB: GC前->GC后->存活

// 代码方式
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("堆: %.2f MB, GC: %d 次, 暂停: %v\n",
    float64(m.HeapAlloc)/1024/1024,
    m.NumGC,
    time.Duration(m.PauseTotalNs))

// 监控指标（Prometheus）
// go_gc_duration_seconds
// go_memstats_alloc_bytes
// go_goroutines`),
		},
	}
}

func escapeTopic() Topic {
	return Topic{
		ID: "escape", Title: "逃逸分析", Chapter: "03-memory",
		Sections: []Section{
			s("栈 vs 堆分配",
				"**栈分配**：函数返回后不被引用、编译期确定大小、分配/释放几乎零成本（移动 SP 指针）。**堆分配**：函数返回后仍被引用、大小编译期不确定、需要 GC 回收。性能差距：栈分配 ~1ns，堆分配 ~50-100ns（mallocgc + GC 开销）。如果每秒百万次堆分配，差距巨大。**生产影响**：热路径上的堆分配是性能瓶颈的常见原因。**面试追问**：逃逸分析是在编译期还是运行期？→ 编译期，由编译器决定。",
				`// 栈分配 (~1ns): 函数返回后不再使用
x := 42
arr := [100]int{}
p := Point{1, 2}

// 堆分配 (~50-100ns): 函数返回后仍被引用
func newPoint() *Point {
    p := Point{1, 2}
    return &p  // p 逃逸到堆上
}

// 查看: go build -gcflags="-m"
// main.go:3:6: moved to heap: p`),
			s("逃逸场景",
				"7 种常见逃逸场景：1) **返回局部变量指针**（最常见）；2) **赋值给 interface**（fmt.Println 参数、any 类型）；3) **闭包捕获变量**（变量生命周期延长）；4) **send 到 channel**（值被其他 goroutine 使用）；5) **slice/map 存储指针**（指向的数据逃逸）；6) **reflect 使用**（编译期无法确定类型）；7) **fmt.Printf 等可变参数**（...any 触发装箱）。**生产建议**：热路径避免 fmt.Sprintf，用 strconv；避免不必要的指针返回。",
				`// 1. 返回指针
func f() *int { x := 42; return &x }  // x 逃逸

// 2. interface（fmt 系列都会导致逃逸）
fmt.Sprintf("value: %d", x)  // x 逃逸到堆

// 3. 闭包
func counter() func() int {
    n := 0  // n 逃逸
    return func() int { n++; return n }
}

// 4. channel
ch <- data  // data 逃逸

// 5. fmt 可变参数
log.Printf("msg: %s", msg)  // msg 逃逸

// 查看逃逸:
// go build -gcflags="-m"        # 基本
// go build -gcflags="-m -m"     # 详细`),
			s("避免逃逸的技巧",
				"1) 小结构体用**值传递**（<=64 bytes 复制比堆分配快）；2) **预分配 slice** 容量（避免 append 触发扩容时的复制）；3) 用 **strconv** 代替 fmt.Sprintf（不触发 interface 装箱）；4) 热路径用**具体类型**代替 interface（避免动态分发和逃逸）；5) **sync.Pool** 复用对象；6) **小函数有利于内联**（内联后编译器能做更好的分析）。**生产场景**：JSON 序列化中避免 fmt → 用 strconv；高频日志中避免 fmt.Sprintf → 用 strings.Builder。",
				`// 1. strconv 代替 fmt（减少逃逸）
s := strconv.Itoa(42)       // ✓ 零分配
s := fmt.Sprintf("%d", 42)  // ✗ 有分配

// 2. 预分配 slice
s := make([]int, 0, n)  // 避免 append 扩容

// 3. sync.Pool 复用
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}
buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
defer bufPool.Put(buf)

// 4. 值传递小结构体
type Point struct{ X, Y float64 }
func (p Point) Distance() float64 { // 值接收者
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// 5. 避免 interface 导致的逃逸
func process(data []byte) error { // 具体 type
    // 而非 func process(r io.Reader)
}`),
		},
	}
}

func alignmentTopic() Topic {
	return Topic{
		ID: "alignment", Title: "内存对齐", Chapter: "03-memory",
		Sections: []Section{
			s("对齐规则",
				"三条规则：1) 字段按类型的对齐系数对齐（不足的补 padding）；2) struct 整体大小必须是最大对齐系数的倍数；3) 基本类型对齐系数 = 其大小（64位系统：bool=1, int32=4, int64=8）。**为什么需要内存对齐**：CPU 访问对齐的内存更快（一次总线操作），不对齐可能触发硬件异常（某些架构）。**生产影响**：百万个 struct 的微优化可能节省数 MB 内存。**面试追问**：unsafe.Alignof 和 unsafe.Sizeof 的区别？→ Alignof 返回对齐系数，Sizeof 返回实际大小。",
				`// Bad: 24 bytes（浪费 8 bytes padding）
type Bad struct {
    A bool    // 1 byte + 7 bytes padding
    B int64   // 8 bytes
    C int32   // 4 bytes + 4 bytes padding
}
fmt.Println(unsafe.Sizeof(Bad{}))  // 24

// Good: 16 bytes（零 padding）
type Good struct {
    B int64   // 8 bytes
    C int32   // 4 bytes
    A bool    // 1 byte + 3 bytes padding
}
fmt.Println(unsafe.Sizeof(Good{}))  // 16

// 节省 33% 内存！百万个实例省 8MB`),
			s("字段顺序优化",
				"原则：**按字段大小从大到小排列**，减少 padding。使用 **fieldalignment** 工具自动检测和修复（go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest）。**生产场景**：高并发服务中大量传输的 struct（如 HTTP 请求/响应体、数据库模型）。**面试追问**：编译器会自动优化字段顺序吗？→ 不会，Go 编译器保持字段声明顺序。",
				`// Bad: 24 bytes
type UserBad struct {
    Active  bool     // 1+7
    Balance float64  // 8
    Age     int32    // 4+4
}

// Good: 16 bytes (省 33%)
type UserGood struct {
    Balance float64  // 8
    Age     int32    // 4
    Active  bool     // 1+3
}

// 工具自动修复:
// fieldalignment -fix ./...

// 生产: 百万用户列表
// Bad:  24MB → Good: 16MB (省 8MB)
// 且缓存命中率更高（更紧凑 = 更多数据在缓存行中)`),
			s("atomic 对齐与零大小类型",
				"64 位原子操作要求 **8 字节对齐**，32 位系统上不对齐会 panic。atomic 变量应放在 struct **第一个字段**。**零大小类型 struct{}**：大小为 0，不占内存。用于实现 set（map[T]struct{}）和信号 channel（chan struct{}）。**生产场景**：原子计数器放 struct 首字段；连接管理用 map[string]struct{} 代替 map[string]bool。",
				`// atomic 变量放第一个字段
type Counter struct {
    count int64  // 第一个字段，保证 8 字节对齐
    flag  bool
}

// 32 位系统上以下代码可能 panic！
type BadCounter struct {
    flag  bool
    count int64  // 可能不对齐到 8 字节
}

// 零大小类型: set 实现
type Set struct {
    m map[string]struct{}
}
func (s *Set) Add(v string) { s.m[v] = struct{}{} }
func (s *Set) Has(v string) bool {
    _, ok := s.m[v]; return ok
}

// 信号 channel（不传数据）
done := make(chan struct{})
close(done)  // 广播通知`),
		},
	}
}

func profilingTopic() Topic {
	return Topic{
		ID: "profiling", Title: "Profiling 性能分析", Chapter: "04-performance",
		Sections: []Section{
			s("CPU Profiling",
				"原理：操作系统每秒发送 100 次 SIGPROF 信号（100Hz），每次记录所有线程的调用栈。采样结束后统计每个函数出现在调用栈中的次数，出现越多 = 占用 CPU 越多。**使用方式**：pprof.StartCPUProfile(f) 开始，pprof.StopCPUProfile() 结束。**分析命令**：top20 查看热点函数，web 生成火焰图（需要 graphviz），list funcName 查看行级耗时。**生产场景**：接口 RT 突然变慢 → CPU profile 找热点函数。",
				`f, _ := os.Create("cpu.prof")
defer f.Close()
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// 你的代码...
busyWork()

// 分析:
// go tool pprof cpu.prof
// (pprof) top20          # 热点函数
// (pprof) web            # 火焰图
// (pprof) list myFunc    # 行级耗时
// (pprof) peek myFunc    # 调用者和被调用者

// Web UI（推荐）:
// go tool pprof -http=:8080 cpu.prof`),
			s("Memory Profiling",
				"两种内存 profile：**allocs**（累计分配统计）和 **heap**（当前存活对象的分配统计）。用 pprof.WriteHeapProfile 写入。**分析维度**：-inuse_space（正在使用的内存）、-inuse_objects（正在使用的对象数）、-alloc_space（累计分配量）、-alloc_objects（累计分配次数）。**生产场景**：内存持续增长 → 对比两次 heap profile 找泄漏点。**面试追问**：allocs 和 heap 的区别？→ allocs 是累计值（从程序启动），heap 是当前值。",
				`f, _ := os.Create("mem.prof")
defer f.Close()
pprof.WriteHeapProfile(f)

// 分析:
// go tool pprof mem.prof
// (pprof) top20 -inuse_space   # 当前占用内存最多
// (pprof) top20 -alloc_space   # 累计分配最多
// (pprof) web
// (pprof) list myFunc

// 内存泄漏排查:
// 1. 两次 heap profile 对比
// go tool pprof -base mem1.prof mem2.prof
// 2. 看 inuse_space 增长的函数`),
			s("在线 Profiling",
				"**net/http/pprof** 自动注册 /debug/pprof/ 路由，无需额外代码（import _ \"net/http/pprof\"）。支持 CPU、heap、goroutine、thread、block、mutex 等 profile。**goroutine 泄漏排查**：查看 goroutine profile 中阻塞的调用栈，找到泄漏的 goroutine。**生产建议**：只在内部端口暴露 pprof，不要对外暴露。**面试追问**：block 和 mutex profile 需要额外设置？→ 是，需要 runtime.SetBlockProfileRate 和 runtime.SetMutexProfileFraction。",
				`import _ "net/http/pprof"
go http.ListenAndServe(":6060", nil)

// 在线 profiling:
// CPU (30秒采样):
// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
// Heap:
// go tool pprof http://localhost:6060/debug/pprof/heap
// Goroutine:
// go tool pprof http://localhost:6060/debug/pprof/goroutine

// goroutine 泄漏排查:
// 1. 看数量: curl localhost:6060/debug/pprof/goroutine?debug=1
// 2. 找到阻塞的调用栈
// 3. 定位泄漏原因

// 生产: 只在内部分析端口暴露
go func() {
    listener, _ := net.Listen("tcp", "127.0.0.1:6060")
    http.Serve(listener, nil)
}()`),
		},
	}
}

func benchmarkTopic() Topic {
	return Topic{
		ID: "benchmark", Title: "Benchmark 基准测试", Chapter: "04-performance",
		Sections: []Section{
			s("基本用法",
				"函数签名 func BenchmarkXxx(b *testing.B)，循环 b.N 次。b.N 由框架自动调整（从 1 开始，逐步增大），直到运行时间足够长（默认 >= 1s）结果可信。运行：go test -bench=. -benchmem。**输出解读**：671.4 ns/op（每次操作耗时）、2040 B/op（每次分配字节数）、8 allocs/op（每次分配次数）。**生产场景**：优化前后对比、竞品方案选型、CI 性能回归检测。",
				`func BenchmarkSliceAppend(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}

func BenchmarkSlicePrealloc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 100)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}
// go test -bench=. -benchmem
// Append:   671 ns/op  2040 B/op  8 allocs/op
// Prealloc:  73 ns/op     0 B/op  0 allocs/op`),
			s("避免编译器优化与子 Benchmark",
				"编译器可能优化掉没有副作用的代码。解决：将结果赋给**包级变量**。Go 1.24+ 可用 b.Loop() 更简洁。**子 Benchmark**：用 b.Run 创建子测试，对比多种实现方案。**b.ReportAllocs()**：即使不用 -benchmem 也会报告分配信息。",
				`var globalResult int

func BenchmarkSafe(b *testing.B) {
    var result int
    for i := 0; i < b.N; i++ {
        result = heavyComputation(i)
    }
    globalResult = result  // 防止优化
}

// Go 1.24+: b.Loop()
func BenchmarkLoop(b *testing.B) {
    for b.Loop() {
        heavyComputation(0)
    }
}

// 子 Benchmark（对比多种实现）
func BenchmarkConcat(b *testing.B) {
    parts := []string{"a", "b", "c", "d", "e"}
    b.Run("plus", func(b *testing.B) { /* ... */ })
    b.Run("builder", func(b *testing.B) { /* ... */ })
    b.Run("join", func(b *testing.B) { /* ... */ })
}`),
			s("benchstat 对比与实战技巧",
				"**benchstat** 用于统计显著性对比。优化前后各运行多次（-count=5），用 benchstat 对比，输出 delta 百分比和置信区间。**实战技巧**：1) -count=5 多次运行取中位数；2) -benchtime=5s 增加采样时间；3) -run=^$ 跳过单元测试加速；4) 关闭其他程序减少噪声。**生产场景**：PR 提交前跑 benchmark 确认没有性能退化。",
				`# 完整的 benchmark 对比流程

# 优化前（运行 5 次取统计）
go test -bench=BenchmarkConcat -count=5 -benchmem > old.txt

# 优化后
go test -bench=BenchmarkConcat -count=5 -benchmem > new.txt

# 对比
benchstat old.txt new.txt
# 输出:
# name         old time/op  new time/op  delta
# Concat/plus    120ns ± 2%    45ns ± 1%  -62.50%
# Concat/builder  80ns ± 1%    40ns ± 2%  -50.00%

# 常用参数:
# -benchmem     显示分配信息
# -count=5      多次运行
# -benchtime=5s 增加采样时间
# -run=^$       跳过单元测试`),
		},
	}
}
