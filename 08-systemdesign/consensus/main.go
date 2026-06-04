package main

import (
	"fmt"
)

// ============================================================
// 分布式一致性算法
// ============================================================

func main() {
	raftElection()
	raftLogReplication()
	paxosOverview()
	raftInGo()
	compare()
}

// ----------------------------------------------------------
// 1. Raft Leader 选举
// ----------------------------------------------------------
func raftElection() {
	fmt.Println("=== 1. Raft Leader 选举 ===")
	fmt.Println()
	fmt.Println("  角色: Follower → Candidate → Leader")
	fmt.Println()
	fmt.Println("  选举流程:")
	fmt.Println("    1. Follower 启动时有一个选举超时 (150~300ms 随机)")
	fmt.Println("    2. 超时未收到 Leader 心跳 → 转为 Candidate")
	fmt.Println("    3. 自增 Term，投自己一票，向其他节点拉票")
	fmt.Println("    4. 获得多数票 → 成为 Leader")
	fmt.Println("    5. 开始发送心跳维持领导权")
	fmt.Println()
	fmt.Println("  Term (任期):")
	fmt.Println("    每个任期内最多一个 Leader")
	fmt.Println("    Term 像逻辑时钟，过期 Term 的请求会被拒绝")
	fmt.Println()
	fmt.Println("  选举安全性:")
	fmt.Println("    每个节点在一个 Term 内只能投一票 → 保证最多一个 Leader")
	fmt.Println("    随机超时 → 降低多 Candidate 分散选票的概率")
	fmt.Println()
	fmt.Println("  面试追问: 脑裂怎么处理?")
	fmt.Println("    旧 Leader 被网络隔离 → 无法获得多数 ACK → 降级为 Follower")
	fmt.Println("    新 Leader 由多数派选出 → 旧 Leader 的写入无效")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Raft 日志复制
// ----------------------------------------------------------
func raftLogReplication() {
	fmt.Println("=== 2. Raft 日志复制 ===")
	fmt.Println()
	fmt.Println("  写入流程:")
	fmt.Println("    1. 客户端发送写请求到 Leader")
	fmt.Println("    2. Leader 追加到本地日志")
	fmt.Println("    3. 并行发送 AppendEntries RPC 给 Followers")
	fmt.Println("    4. 多数节点确认 → Leader 提交日志")
	fmt.Println("    5. 响应客户端")
	fmt.Println()
	fmt.Println("  Leader:")
	fmt.Println("  log: [1:SET x=1] [2:SET y=2] [3:SET z=3]")
	fmt.Println("    ↑ committed (index 3): 多数节点确认")
	fmt.Println()
	fmt.Println("  Follower A: [1] [2] [3]     — 完全同步")
	fmt.Println("  Follower B: [1] [2] [3]     — 完全同步")
	fmt.Println("  Follower C: [1] [2]         — 落后一个条目")
	fmt.Println()
	fmt.Println("  日志一致性保证:")
	fmt.Println("    如果两个日志条目的 index 和 Term 相同，则:")
	fmt.Println("    - 它们存储的命令相同")
	fmt.Println("    - 它们之前的所有日志也相同")
	fmt.Println()
	fmt.Println("  一致性检查:")
	fmt.Println("    AppendEntries RPC 包含 prevLogIndex 和 prevLogTerm")
	fmt.Println("    Follower 检查匹配 → 不匹配则拒绝 → Leader 回退重试")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Paxos 概览
// ----------------------------------------------------------
func paxosOverview() {
	fmt.Println("=== 3. Paxos 概览 ===")
	fmt.Println()
	fmt.Println("  Basic Paxos 角色:")
	fmt.Println("    Proposer — 提案者 (接受客户端请求)")
	fmt.Println("    Acceptor — 接受者 (投票决定是否接受提案)")
	fmt.Println("    Learner  — 学习者 (了解最终决定的值)")
	fmt.Println()
	fmt.Println("  两阶段提交:")
	fmt.Println("    Phase 1 (Prepare):")
	fmt.Println("      Proposer 发送 Prepare(n) — n 是提案编号")
	fmt.Println("      Acceptor: 如果 n > 已见过的最大编号 → 承诺不接受更小编号")
	fmt.Println()
	fmt.Println("    Phase 2 (Accept):")
	fmt.Println("      Proposer 发送 Accept(n, value)")
	fmt.Println("      多数 Acceptor 接受 → 值被选定 (chosen)")
	fmt.Println()
	fmt.Println("  Multi-Paxos:")
	fmt.Println("    优化: 选出 Leader 后跳过 Phase 1")
	fmt.Println("    Leader 直接发起 Accept (减少一轮 RPC)")
	fmt.Println("    本质上就是 Raft 的前身")
	fmt.Println()
	fmt.Println("  ZAB (ZooKeeper Atomic Broadcast):")
	fmt.Println("    与 Raft 类似，用于 ZooKeeper")
	fmt.Println("    区别: 选举时优先选择数据最新的节点")
	fmt.Println("    模式: 崩溃恢复模式 / 消息广播模式")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. Go 实现 Raft
// ----------------------------------------------------------
func raftInGo() {
	fmt.Println("=== 4. Go Raft 实现 (hashicorp/raft) ===")
	fmt.Println()
	fmt.Println("  核心接口:")
	fmt.Println()
	fmt.Println("  // FSM — 有限状态机 (应用日志到业务状态)")
	fmt.Println("  type FSM interface {")
	fmt.Println("    Apply(*Log) any         // 应用已提交的日志")
	fmt.Println("    Snapshot() (FSMSnapshot, error)  // 生成快照")
	fmt.Println("    Restore(rc io.ReadCloser) error  // 从快照恢复")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("  基本使用:")
	fmt.Println("  config := raft.DefaultConfig()")
	fmt.Println("  config.LocalID = \"node-1\"")
	fmt.Println()
	fmt.Println("  // 创建 Raft 实例")
	fmt.Println("  r, _ := raft.NewRaft(config, fsm, logStore, stableStore,")
	fmt.Println("    snapshotStore, transport)")
	fmt.Println()
	fmt.Println("  // 引导集群 (只需执行一次)")
	fmt.Println("  r.BootstrapCluster(raft.Configuration{")
	fmt.Println("    Servers: []raft.Server{{")
	fmt.Println("      ID:      \"node-1\",")
	fmt.Println("      Address: \"127.0.0.1:8000\",")
	fmt.Println("    }},")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("  // 应用操作 (Leader 节点)")
	fmt.Println("  future := r.Apply([]byte(\"SET x=1\"), 5*time.Second)")
	fmt.Println("  if err := future.Error(); err != nil {")
	fmt.Println("    // 操作失败")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("  日志压缩 (Snapshot):")
	fmt.Println("    日志无限增长 → 定期做 Snapshot 截断旧日志")
	fmt.Println("    Go: r.Snapshot() 触发快照")
	fmt.Println("    快照包含: lastIncludedIndex + lastIncludedTerm + 状态机状态")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 对比总结
// ----------------------------------------------------------
func compare() {
	fmt.Println("=== 5. 一致性算法对比 ===")
	fmt.Println()
	fmt.Println("  ┌──────────┬──────────────┬──────────────┬──────────────┐")
	fmt.Println("  │ 维度     │ Paxos        │ Raft         │ ZAB          │")
	fmt.Println("  ├──────────┼──────────────┼──────────────┼──────────────┤")
	fmt.Println("  │ 可理解性 │ 难           │ 易           │ 中           │")
	fmt.Println("  │ Leader   │ 可无         │ 必须         │ 必须         │")
	fmt.Println("  │ 日志     │ 无序         │ 有序连续     │ 有序连续     │")
	fmt.Println("  │ 实现     │ 难           │ 相对简单     │ 中           │")
	fmt.Println("  │ 应用     │ Chubby/Spanner│ Etcd/Consul │ ZooKeeper    │")
	fmt.Println("  └──────────┴──────────────┴──────────────┴──────────────┘")
	fmt.Println()
	fmt.Println("  面试追问: Raft 和 Paxos 的本质区别?")
	fmt.Println("  答: Raft 通过强 Leader 简化了 Paxos 的复杂性")
	fmt.Println("      - 日志只能从 Leader 流向 Follower (单向)")
	fmt.Println("      - Leader 选举更简单 (随机超时 + 多数投票)")
	fmt.Println("      - 不允许日志空洞 (简化了一致性保证)")
	fmt.Println()
}
