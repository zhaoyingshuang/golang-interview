package main

import (
	"fmt"
)

// ============================================================
// 高可用架构
// ============================================================

func main() {
	slaMetrics()
	redundancy()
	gracefulShutdown()
	multiActive()
	chaosEngineering()
}

// ----------------------------------------------------------
// 1. SLA/SLO/SLI 指标
// ----------------------------------------------------------
func slaMetrics() {
	fmt.Println("=== 1. 高可用指标 ===")
	fmt.Println()
	fmt.Println("  SLA (Service Level Agreement) — 服务等级协议 (对客户的承诺)")
	fmt.Println("  SLO (Service Level Objective) — 服务等级目标 (内部目标)")
	fmt.Println("  SLI (Service Level Indicator) — 服务等级指标 (可衡量的数据)")
	fmt.Println()
	fmt.Println("  常见 SLI:")
	fmt.Println("    可用性   = 成功请求数 / 总请求数")
	fmt.Println("    延迟     = P50/P95/P99 延迟")
	fmt.Println("    错误率   = 5xx 响应数 / 总请求数")
	fmt.Println("    吞吐量   = QPS/TPS")
	fmt.Println()
	fmt.Println("  可用性级别:")
	fmt.Println("    99.9%  → 每月允许停机 43.8 分钟")
	fmt.Println("    99.95% → 每月允许停机 21.9 分钟")
	fmt.Println("    99.99% → 每月允许停机 4.38 分钟")
	fmt.Println("    99.999%→ 每月允许停机 26 秒")
	fmt.Println()
	fmt.Println("  Error Budget (错误预算):")
	fmt.Println("    Error Budget = 1 - SLO")
	fmt.Println("    如 SLO=99.9% → 每月有 0.1% 的错误预算")
	fmt.Println("    用完后停止发布新功能，专注稳定性")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 冗余设计
// ----------------------------------------------------------
func redundancy() {
	fmt.Println("=== 2. 冗余设计 ===")
	fmt.Println()
	fmt.Println("  多副本 (Replication):")
	fmt.Println("    主从复制: 一主多从，主写从读")
	fmt.Println("    主主复制: 双主互为备份，均可写入")
	fmt.Println("    共识复制: Raft/Paxos 保证一致性 (Etcd/Consul)")
	fmt.Println()
	fmt.Println("  无状态服务冗余:")
	fmt.Println("    多实例部署 → 负载均衡 → 任一实例故障自动摘除")
	fmt.Println("    K8s: Deployment + HPA (水平自动扩缩)")
	fmt.Println()
	fmt.Println("  有状态服务冗余:")
	fmt.Println("    MySQL: 主从复制 + MHA Orchestrator 自动切换")
	fmt.Println("    Redis: 哨兵模式/Cluster 模式")
	fmt.Println("    Kafka: 多副本 + ISR 机制")
	fmt.Println()
	fmt.Println("  关键: 健康检查 + 自动故障转移")
	fmt.Println("    K8s: livenessProbe + readinessProbe")
	fmt.Println("    负载均衡器: 被动健康检查 (请求失败后摘除)")
	fmt.Println("    注册中心: 主动健康检查 (定期探测)")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 优雅上下线
// ----------------------------------------------------------
func gracefulShutdown() {
	fmt.Println("=== 3. 优雅上下线 ===")
	fmt.Println()
	fmt.Println("优雅上线:")
	fmt.Println("  1. 启动服务 (监听端口)")
	fmt.Println("  2. 初始化连接 (DB/Redis/MQ)")
	fmt.Println("  3. 预热缓存 (可选)")
	fmt.Println("  4. 健康检查就绪 → 注册到发现服务")
	fmt.Println("  5. 开始接收流量")
	fmt.Println()
	fmt.Println("优雅下线:")
	fmt.Println("  1. 从注册中心注销")
	fmt.Println("  2. 标记为不健康 (不接收新流量)")
	fmt.Println("  3. 等待进行中请求完成 (drain timeout)")
	fmt.Println("  4. 关闭数据库/缓存连接")
	fmt.Println("  5. 停止 HTTP/gRPC 服务")
	fmt.Println()
	fmt.Println("Go 优雅关闭:")
	fmt.Println("  quit := make(chan os.Signal, 1)")
	fmt.Println("  signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)")
	fmt.Println("  <-quit")
	fmt.Println()
	fmt.Println("  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)")
	fmt.Println("  defer cancel()")
	fmt.Println()
	fmt.Println("  // 注销服务")
	fmt.Println("  registry.Deregister(serviceID)")
	fmt.Println()
	fmt.Println("  // 优雅关闭")
	fmt.Println("  if err := srv.Shutdown(ctx); err != nil {")
	fmt.Println("    log.Fatal(\"forced shutdown:\", err)")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 多活架构
// ----------------------------------------------------------
func multiActive() {
	fmt.Println("=== 4. 多活架构 ===")
	fmt.Println()
	fmt.Println("  同城双活:")
	fmt.Println("    两个机房同时提供服务")
	fmt.Println("    数据同步延迟低 (< 1ms)")
	fmt.Println("    DNS/GSLB 做流量调度")
	fmt.Println()
	fmt.Println("  异地多活:")
	fmt.Println("    多个城市部署，就近接入")
	fmt.Println("    按用户 ID 分片路由 (单元化)")
	fmt.Println("    数据同步延迟较高，需处理冲突")
	fmt.Println()
	fmt.Println("  关键挑战:")
	fmt.Println("    1. 数据一致性 — 异步复制有延迟")
	fmt.Println("    2. 全局唯一 ID — Snowflake / Leaf")
	fmt.Println("    3. 分布式事务 — TCC / Saga / 本地消息表")
	fmt.Println("    4. 流量调度 — GSLB + 服务路由")
	fmt.Println()
	fmt.Println("  简化方案: 主备而非多活")
	fmt.Println("    主机房承担所有流量，备机房热备")
	fmt.Println("    故障时 DNS 切换到备用 (RTO ~分钟级)")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 混沌工程
// ----------------------------------------------------------
func chaosEngineering() {
	fmt.Println("=== 5. 混沌工程 (Chaos Engineering) ===")
	fmt.Println()
	fmt.Println("  原理: 主动注入故障，验证系统的容错能力")
	fmt.Println()
	fmt.Println("  常见故障注入:")
	fmt.Println("    - 随机杀进程 (Chaos Monkey)")
	fmt.Println("    - 网络延迟/丢包 (tc/qdisc)")
	fmt.Println("    - 磁盘 IO 延迟")
	fmt.Println("    - CPU/内存压力")
	fmt.Println("    - 依赖服务超时/错误")
	fmt.Println()
	fmt.Println("  工具:")
	fmt.Println("    Chaos Mesh  — K8s 原生混沌工程平台")
	fmt.Println("    Litmus      — CNCF 混沌工程项目")
	fmt.Println("    Gremlin     — 商业混沌工程平台")
	fmt.Println()
	fmt.Println("  Go 服务故障注入示例:")
	fmt.Println("  // 中间件注入延迟")
	fmt.Println("  func ChaosMiddleware(c *gin.Context) {")
	fmt.Println("    if rand.Float32() < 0.1 {  // 10% 概率")
	fmt.Println("      time.Sleep(3 * time.Second) // 注入延迟")
	fmt.Println("    }")
	fmt.Println("    c.Next()")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("面试追问: 如何设计 99.99% 可用系统?")
	fmt.Println("  1. 消除单点 — 多副本 + 自动故障转移")
	fmt.Println("  2. 快速失败 — 超时 + 熔断 + 降级")
	fmt.Println("  3. 限流保护 — 防止雪崩")
	fmt.Println("  4. 可观测性 — 监控 + 告警 + 链路追踪")
	fmt.Println("  5. 混沌验证 — 主动发现隐患")
	fmt.Println()
}
