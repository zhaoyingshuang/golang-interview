---
title: 分布式一致性算法
---

## 1. Raft Leader 选举

角色：Follower → Candidate → Leader

选举流程：
1. Follower 选举超时（150~300ms 随机）→ 转为 Candidate
2. 自增 Term，投自己一票，向其他节点拉票
3. 获得多数票 → 成为 Leader
4. 开始发送心跳维持领导权

::: tip 选举安全性
每个节点在一个 Term 内只能投一票 → 保证最多一个 Leader。随机超时降低多 Candidate 分散选票的概率。

面试追问：脑裂怎么处理？→ 旧 Leader 被隔离后无法获得多数 ACK，降级为 Follower。新 Leader 由多数派选出。
:::

## 2. Raft 日志复制

1. 客户端写请求到 Leader
2. Leader 追加到本地日志
3. 并行发送 AppendEntries RPC 给 Followers
4. 多数节点确认 → Leader 提交日志
5. 响应客户端

**一致性检查**：AppendEntries 包含 `prevLogIndex + prevLogTerm`，Follower 检查匹配 → 不匹配拒绝 → Leader 回退重试。

## 3. Paxos 概览

Basic Paxos 两阶段：
- **Phase 1 (Prepare)**：Proposer 发送提案编号 n，Acceptor 承诺不接受更小编号
- **Phase 2 (Accept)**：Proposer 发送 Accept(n, value)，多数接受 → 值被选定

Multi-Paxos 优化：选出 Leader 后跳过 Phase 1。Raft 本质上是简化版的 Multi-Paxos。

## 4. Go 实现 Raft (hashicorp/raft)

```go
// FSM — 有限状态机接口
type FSM interface {
    Apply(*Log) any                        // 应用已提交的日志
    Snapshot() (FSMSnapshot, error)        // 生成快照
    Restore(rc io.ReadCloser) error        // 从快照恢复
}

// 使用
r, _ := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
future := r.Apply([]byte("SET x=1"), 5*time.Second)
```

::: info 日志压缩 (Snapshot)
日志无限增长 → 定期做 Snapshot 截断旧日志。快照包含 lastIncludedIndex + lastIncludedTerm + 状态机状态。
:::

## 5. 算法对比

| 维度 | Paxos | Raft | ZAB |
|------|-------|------|-----|
| 可理解性 | 难 | 易 | 中 |
| Leader | 可无 | 必须 | 必须 |
| 应用 | Chubby/Spanner | Etcd/Consul | ZooKeeper |

Raft 通过强 Leader 简化了复杂性：日志单向流动、简单的选举机制、不允许日志空洞。
