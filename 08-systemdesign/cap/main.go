package main

import (
	"fmt"
)

// ============================================================
// CAP 理论与分布式基础
// ============================================================

func main() {
	capTheory()
	baseTheory()
	consistencyModels()
	distributedID()
	practicalTradeoffs()
}

// ----------------------------------------------------------
// 1. CAP 定理
// ----------------------------------------------------------
func capTheory() {
	fmt.Println("=== 1. CAP 定理 ===")
	fmt.Println()
	fmt.Println("  C — Consistency (一致性): 所有节点看到相同的数据")
	fmt.Println("  A — Availability (可用性): 每个请求都能得到响应")
	fmt.Println("  P — Partition Tolerance (分区容错): 网络分区时系统继续运行")
	fmt.Println()
	fmt.Println("  核心结论: 网络分区不可避免，只能在 C 和 A 之间选择")
	fmt.Println()
	fmt.Println("  ┌──────────┬──────────────────┬───────────────────────┐")
	fmt.Println("  │ 选择     │ 牺牲            │ 典型系统              │")
	fmt.Println("  ├──────────┼──────────────────┼───────────────────────┤")
	fmt.Println("  │ CP       │ 可用性 A         │ Etcd, ZooKeeper, HBase│")
	fmt.Println("  │ AP       │ 强一致性 C       │ Cassandra, DynamoDB   │")
	fmt.Println("  │ CA       │ 分区容错 P       │ 单机 MySQL (无分布式) │")
	fmt.Println("  └──────────┴──────────────────┴───────────────────────┘")
	fmt.Println()
	fmt.Println("  注意: CA 只存在于没有网络分区的理想情况，")
	fmt.Println("  在分布式系统中 P 是必选项，所以实际只有 CP 和 AP")
	fmt.Println()
	fmt.Println("  面试追问: CAP 在实际中不是非此即彼?")
	fmt.Println("  答: 是的。CAP 是在分区发生时的选择，正常情况下 CA 兼顾。")
	fmt.Println("  大部分系统在不同操作上可以有不同的选择 (如读 AP，写 CP)")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. BASE 理论
// ----------------------------------------------------------
func baseTheory() {
	fmt.Println("=== 2. BASE 理论 ===")
	fmt.Println()
	fmt.Println("  BA — Basically Available (基本可用)")
	fmt.Println("    系统出现故障时，允许响应时间增加或功能降级")
	fmt.Println("    例: 双11 大促时，商品详情显示简化版")
	fmt.Println()
	fmt.Println("  S — Soft State (软状态)")
	fmt.Println("    允许系统中的数据存在中间状态")
	fmt.Println("    例: 订单状态从「已支付」到「已发货」之间有延迟")
	fmt.Println()
	fmt.Println("  E — Eventually Consistent (最终一致性)")
	fmt.Println("    系统保证在没有新更新的情况下，最终所有副本数据一致")
	fmt.Println("    例: DNS 更新后，全球生效需要 48 小时")
	fmt.Println()
	fmt.Println("  BASE 是对 CAP 中 AP 方向的补充，")
	fmt.Println("  也是大多数互联网系统的实际选择")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 一致性模型
// ----------------------------------------------------------
func consistencyModels() {
	fmt.Println("=== 3. 一致性模型 ===")
	fmt.Println()
	fmt.Println("  强一致性 (Strong Consistency / Linearizable)")
	fmt.Println("    任何时刻读到的都是最新写入的值")
	fmt.Println("    代价: 高延迟（需要多数派确认）")
	fmt.Println("    实现: Raft/Paxos/ZAB")
	fmt.Println()
	fmt.Println("  最终一致性 (Eventual Consistency)")
	fmt.Println("    经过一段时间后，所有副本最终一致")
	fmt.Println("    中间可能读到旧数据 (stale read)")
	fmt.Println("    实现: 异步复制 (MySQL 主从)")
	fmt.Println()
	fmt.Println("  因果一致性 (Causal Consistency)")
	fmt.Println("    有因果关系的操作顺序一致")
	fmt.Println("    例: 「发帖」→「评论」，所有人先看到帖子再看到评论")
	fmt.Println()
	fmt.Println("  读己之写 (Read Your Writes)")
	fmt.Println("    写入后立即能读到自己的写入")
	fmt.Println("    实现: 写主读主 / 写后路由到主")
	fmt.Println()
	fmt.Println("  单调读 (Monotonic Read)")
	fmt.Println("    不会读到比之前更旧的数据")
	fmt.Println("    实现: 绑定到同一副本 / 读主")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 分布式 ID 生成
// ----------------------------------------------------------
func distributedID() {
	fmt.Println("=== 4. 分布式 ID 生成 ===")
	fmt.Println()
	fmt.Println("  要求: 全局唯一、趋势递增、高性能、高可用")
	fmt.Println()
	fmt.Println("  ┌──────────┬────────────────────┬───────────────┐")
	fmt.Println("  │ 方案     │ 原理               │ 特点          │")
	fmt.Println("  ├──────────┼────────────────────┼───────────────┤")
	fmt.Println("  │ UUID     │ 随机生成 128 bit   │ 无序、长度长  │")
	fmt.Println("  │ Snowflake│ 时间+机器+序列号   │ 趋势递增      │")
	fmt.Println("  │ Leaf     │ 号段模式/Snowflake │ 美团、双缓冲  │")
	fmt.Println("  │ Redis    │ INCR 自增          │ 简单、依赖Redis│")
	fmt.Println("  │ 数据库   │ auto_increment     │ 简单、性能低  │")
	fmt.Println("  └──────────┴────────────────────┴───────────────┘")
	fmt.Println()
	fmt.Println("  Snowflake 结构 (64 bit):")
	fmt.Println("  ┌─ 1bit ─┬────── 41bit 时间 ──────┬─ 10bit 机器 ─┬─ 12bit 序列 ─┐")
	fmt.Println("  │  0     │ 毫秒级时间戳(约69年)    │ 机器ID(1024) │ 序列(4096/ms)│")
	fmt.Println("  └────────┴────────────────────────┴──────────────┴──────────────┘")
	fmt.Println()
	fmt.Println("  Go 实现 Snowflake:")
	fmt.Println("  type Snowflake struct {")
	fmt.Println("    mu        sync.Mutex")
	fmt.Println("    epoch     int64  // 起始时间戳")
	fmt.Println("    machineID int64  // 机器ID (10bit)")
	fmt.Println("    sequence  int64  // 序列号 (12bit)")
	fmt.Println("    lastTime  int64  // 上次生成时间")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("  时钟回拨问题:")
	fmt.Println("    NTP 同步可能导致时钟回拨")
	fmt.Println("    解决: 等待追平 / 拒绝生成 / 使用历史最大时间")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 实际取舍
// ----------------------------------------------------------
func practicalTradeoffs() {
	fmt.Println("=== 5. 实际系统中的取舍 ===")
	fmt.Println()
	fmt.Println("  金融系统 → CP (宁可不可用也不能数据错)")
	fmt.Println("    转账、交易必须强一致，使用分布式事务 (TCC/Saga)")
	fmt.Println()
	fmt.Println("  电商系统 → 混合策略")
	fmt.Println("    下单/支付: CP (强一致)")
	fmt.Println("    商品搜索: AP (最终一致)")
	fmt.Println("    库存: CP (宁可少卖不能超卖)")
	fmt.Println()
	fmt.Println("  社交系统 → AP")
	fmt.Println("    帖子/评论: 最终一致 (延迟几秒可接受)")
	fmt.Println("    点赞计数: 最终一致 (允许短暂不准)")
	fmt.Println()
	fmt.Println("  核心原则: 不是所有数据都需要强一致")
	fmt.Println("  根据业务特点选择合适的一致性级别")
	fmt.Println()
}
