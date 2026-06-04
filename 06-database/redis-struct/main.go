package main

import "fmt"

// ============================================================
// Redis 数据结构
// ============================================================
//
// 【面试高频问题】
// 1. Redis 五大基本类型的底层实现？
// 2. SDS 和 C 字符串的区别？
// 3. 跳表(skiplist)的原理？为什么不用红黑树？
// 4. Redis 持久化方式？RDB/AOF 的区别？
// 5. Redis 为什么这么快？

func main() {
	fiveTypes()
	sds()
	skiplist()
	persistence()
	whyFast()
}

// ----------------------------------------------------------
// 1. 五大基本类型底层实现
// ----------------------------------------------------------
// string → SDS (Simple Dynamic String)
// list   → quicklist (ziplist + 双向链表), Redis 7.0+ listpack
// hash   → ziplist(小)/hashtable(大)
// set    → intset(纯整数小集合)/hashtable
// zset   → ziplist(小)/skiplist + hashtable(大)
//
// **面试追问**: 什么时候从 ziplist 升级为 hashtable?
//   hash-max-ziplist-entries (默认512)
//   hash-max-ziplist-value (默认64字节)
func fiveTypes() {
	fmt.Println("=== 1. 五大基本类型底层实现 ===")
	fmt.Println("string  → SDS (Simple Dynamic String)")
	fmt.Println("list    → quicklist (ziplist + 双向链表)")
	fmt.Println("hash    → listpack(小) / hashtable(大)")
	fmt.Println("set     → intset(纯整数) / hashtable(通用)")
	fmt.Println("zset    → listpack(小) / skiplist + hashtable(大)")
	fmt.Println()
	fmt.Println("ziplist → listpack (Redis 7.0) 解决级联更新问题")
	fmt.Println("转换阈值: hash-max-ziplist-entries=512, hash-max-ziplist-value=64")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. SDS (Simple Dynamic String)
// ----------------------------------------------------------
// struct sdshdr {
//     len   uint32  // 已用长度(O(1)获取)
//     alloc uint32  // 总分配空间
//     flags uint8   // 类型(sdshdr5/8/16/32/64)
//     buf[] byte    // 数据缓冲区(以\0结尾,兼容C字符串)
// }
//
// SDS vs C字符串:
//   C: O(n)获取长度, 二进制不安全(遇\0截断), 缓冲区溢出风险
//   SDS: O(1)长度, 二进制安全, 空间预分配+惰性释放
//
// 空间预分配:
//   len < 1MB → alloc = 2 * len
//   len >= 1MB → alloc = len + 1MB
func sds() {
	fmt.Println("=== 2. SDS vs C 字符串 ===")
	fmt.Println()
	fmt.Println("SDS 结构: len + alloc + flags + buf[]")
	fmt.Println()
	fmt.Println("对比:")
	fmt.Println("  获取长度:    C=O(n)     SDS=O(1)")
	fmt.Println("  二进制安全:  C=否(NUL)  SDS=是(len控制)")
	fmt.Println("  缓冲区溢出:  C=可能     SDS=自动扩容")
	fmt.Println("  内存重分配:  C=N次      SDS=预分配+惰性释放")
	fmt.Println()
	fmt.Println("空间预分配:")
	fmt.Println("  len < 1MB  → alloc = 2 * len")
	fmt.Println("  len >= 1MB → alloc = len + 1MB")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 跳表(skiplist)
// ----------------------------------------------------------
// 用于 ZSET 的有序实现
//
// 结构: 多层链表, 每层是下层的子集
//   L3:  1 → ─── → ─── → 9
//   L2:  1 → 3 → ─── → 9
//   L1:  1 → 3 → 5 → 7 → 9
//
// 查找: 从最高层开始, 每层二分, 时间复杂度 O(logN)
// 插入: 随机层数(概率1/2升一层), 平均 1.33 个指针/节点
//
// 为什么选 skiplet 不选红黑树?
//   1. 实现简单(≈200行 vs 红黑树≈500行)
//   2. 范围查询更高效(链表顺序遍历)
//   3. 内存灵活(可通过概率控制空间)
//   4. 并发友好(局部调整 vs 红黑树全局旋转)
//
// **面试追问**: zset 为什么用 skiplist + hashtable 双结构?
//   skiplist: 支持范围查询和排名操作
//   hashtable: O(1) 精确查找score
func skiplist() {
	fmt.Println("=== 3. 跳表(skiplist) ===")
	fmt.Println()
	fmt.Println("L3:  1 ───────── 9")
	fmt.Println("L2:  1 ── 3 ──── 9")
	fmt.Println("L1:  1 ── 3 ── 5 ── 7 ── 9")
	fmt.Println()
	fmt.Println("查找: O(logN), 从最高层开始逐层下降")
	fmt.Println("插入: 随机层数(1/2概率升层), 平均1.33个指针/节点")
	fmt.Println()
	fmt.Println("为什么不用红黑树?")
	fmt.Println("  1. 实现简单(~200行 vs ~500行)")
	fmt.Println("  2. 范围查询高效(链表顺序遍历)")
	fmt.Println("  3. 内存可控(概率控制)")
	fmt.Println("  4. 并发友好(局部调整)")
	fmt.Println()
	fmt.Println("zset = skiplist(范围查询+排名) + hashtable(O(1)查找)")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 持久化
// ----------------------------------------------------------
// RDB (Redis Database):
//   - 全量快照, fork子进程+COW(Copy-On-Write)
//   - 优点: 恢复快, 文件紧凑
//   - 缺点: 非实时(可能丢失数据), fork耗时(内存大时)
//   - 触发: save/bgsave/自动触发(配置条件)
//
// AOF (Append Only File):
//   - 写后日志, 记录每条写命令
//   - 三种刷盘策略: always/everysec/no
//   - AOF重写: fork子进程, 基于当前状态生成最短命令序列
//   - 优点: 数据更安全, 可读
//   - 缺点: 文件大, 恢复慢
//
// 混合持久化 (Redis 4.0+):
//   AOF重写时前半段RDB格式+后半段AOF格式
//   兼顾恢复速度和数据安全
func persistence() {
	fmt.Println("=== 4. 持久化 ===")
	fmt.Println()
	fmt.Println("RDB: 全量快照(fork+COW)")
	fmt.Println("  优点: 恢复快, 文件小")
	fmt.Println("  缺点: 非实时, fork耗时(大内存)")
	fmt.Println()
	fmt.Println("AOF: 写后日志(记录写命令)")
	fmt.Println("  always: 每条写都刷盘(最安全,最慢)")
	fmt.Println("  everysec: 每秒刷盘(推荐,最多丢1秒)")
	fmt.Println("  no: 由OS决定(最快,可能丢数据)")
	fmt.Println()
	fmt.Println("混合持久化(Redis 4.0+):")
	fmt.Println("  AOF重写 = RDB格式前半段 + AOF格式后半段")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. Redis 为什么快
// ----------------------------------------------------------
// 1. 基于内存操作: 内存访问 ~100ns vs 磁盘 ~10ms
// 2. 单线程 Reactor: 无锁竞争, 无上下文切换
// 3. IO多路复用: epoll/kqueue, 单线程处理大量连接
// 4. 高效数据结构: SDS/ziplist/skiplist/intset
// 5. 单线程避免了: 锁开销/线程切换/死锁
//
// **面试追问**: 单线程怎么利用多核?
//   1. Redis 6.0+ 多IO线程(网络读写多线程, 命令执行单线程)
//   2. 部署多实例, 客户端分片
//   3. Redis Cluster 分片
func whyFast() {
	fmt.Println("=== 5. Redis 为什么快 ===")
	fmt.Println()
	fmt.Println("1. 内存操作: ~100ns vs 磁盘 ~10ms")
	fmt.Println("2. 单线程: 无锁/无切换/无死锁")
	fmt.Println("3. IO多路复用: epoll 单线程处理万级连接")
	fmt.Println("4. 高效数据结构: SDS/ziplist/skiplist")
	fmt.Println()
	fmt.Println("单线程怎么利用多核?")
	fmt.Println("  - Redis 6.0+: 多IO线程(网络多线程, 命令单线程)")
	fmt.Println("  - 多实例 + 客户端分片")
	fmt.Println("  - Redis Cluster 分片")
}
