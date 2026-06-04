---
title: 高可用架构
---

## 1. 高可用指标

### SLA / SLO / SLI

| 概念 | 全称 | 说明 |
|------|------|------|
| SLA | Service Level Agreement | 服务等级协议，与用户的正式承诺，通常包含违约赔偿条款 |
| SLO | Service Level Objective | 服务等级目标，内部设定的可用性目标，如 99.95% |
| SLI | Service Level Indicator | 服务等级指标，衡量可用性的具体指标，如请求成功率、延迟 P99 |

### 可用性计算

| 可用性 | 年度不可用时间 | 日均不可用时间 | 典型系统 |
|--------|---------------|---------------|---------|
| 99% | 3 天 15 小时 | 14 分钟 | 内部工具 |
| 99.9% | 8 小时 45 分钟 | 1 分钟 26 秒 | 一般业务 |
| 99.99% | 52 分钟 | 8 秒 | 核心交易 |
| 99.999% | 5 分钟 | 0.8 秒 | 基础设施 |

::: tip
可用性从 99.9% 提升到 99.99%，成本可能增加 10 倍。需要根据业务价值和成本做权衡，不是所有系统都需要 99.99%。
:::

组合可用性计算公式（串联系统）：

```
系统可用性 = 组件A可用性 × 组件B可用性 × ... × 组件N可用性
```

并联系统（冗余）可用性：

```
系统可用性 = 1 - (1 - 组件可用性)^N
```

## 2. 冗余设计

### 多副本

数据冗余是高可用的基础。常见副本策略：

- **主从复制**：一主多从，写操作由主节点处理，读操作可分散到从节点
- **多主复制**：多个节点同时接受写入，需解决写冲突
- **共识复制**：基于 Raft/Paxos 协议，保证强一致性

### 主从切换

主节点故障时需要自动切换到从节点，关键流程：

1. **故障检测**：心跳超时或健康检查失败
2. **选主**：从健康从节点中选出新主（优先选择数据最新的从节点）
3. **流量切换**：更新 VIP 或修改 DNS，将流量切到新主
4. **通知**：通知应用层更新连接

```go
package ha

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

// HealthCheck 健康检查器
type HealthCheck struct {
    endpoint  string
    interval  time.Duration
    timeout   time.Duration
    retries   int
    unhealthy int // 连续不健康次数
}

func NewHealthCheck(endpoint string) *HealthCheck {
    return &HealthCheck{
        endpoint: endpoint,
        interval: 5 * time.Second,
        timeout:  3 * time.Second,
        retries:  3,
    }
}

func (h *HealthCheck) Run(ctx context.Context, onUnhealthy func()) {
    ticker := time.NewTicker(h.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if !h.check() {
                h.unhealthy++
                if h.unhealthy >= h.retries {
                    log.Printf("[HA] 端点 %s 连续 %d 次不健康，触发切换\n",
                        h.endpoint, h.unhealthy)
                    onUnhealthy()
                    h.unhealthy = 0
                }
            } else {
                h.unhealthy = 0
            }
        }
    }
}

func (h *HealthCheck) check() bool {
    // 实际实现：HTTP 请求或 TCP 连接检测
    return true
}
```

### 多活架构

多活架构分为不同层级：

- **同城多活**：机房之间延迟低（<5ms），可做强一致同步
- **异地多活**：机房之间延迟高（>30ms），通常做最终一致性
- **单元化架构**：按用户维度分片，每个单元独立部署，互不影响

## 3. 容灾方案

### 同城双活

两个机房同时提供服务，数据库主库在其中一个机房：

- 正常时：两个机房分担流量，从库就近读取
- 故障时：快速切换主库，另一个机房接管全量流量
- RTO（恢复时间目标）：分钟级
- RPO（恢复点目标）：秒级（异步复制）或零（同步复制）

### 异地多活

异地多活是最高级别的容灾方案，但也是最复杂的：

1. 数据分片：按用户 ID 或地域将数据分散到不同机房
2. 异步复制：核心数据跨机房异步同步
3. 路由层：根据用户所在地域路由到最近机房
4. 冲突处理：多写场景下的数据冲突解决策略

::: warning
异地多活不是银弹。数据一致性、运维复杂度、成本都是巨大的挑战。大多数业务用同城双活 + 异地冷备就足够了。
:::

### 灾备切换

灾备切换的关键考虑：

- **切换条件**：自动切换 vs 人工决策（核心系统建议人工确认）
- **数据一致性**：切换前确保数据同步完成
- **回滚方案**：切换失败后如何回退
- **流量切换方式**：DNS（分钟级）、VIP（秒级）、应用层路由（毫秒级）

## 4. 优雅上下线

### 优雅关闭

优雅关闭的核心是确保正在处理的请求完成后再退出：

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    server := &http.Server{Addr: ":8080"}

    // 注册路由
    http.HandleFunc("/api/orders", handleOrder)

    // 启动信号监听
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        log.Println("服务启动，监听 :8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal("服务异常退出:", err)
        }
    }()

    // 等待终止信号
    sig := <-quit
    log.Printf("收到信号 %v，开始优雅关闭...\n", sig)

    // 设置关闭超时
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 1. 停止接受新请求
    // 2. 等待已有请求处理完成
    // 3. 释放资源
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("优雅关闭超时，强制退出: %v\n", err)
    }

    log.Println("服务已关闭")
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
    // 业务处理
    w.WriteHeader(http.StatusOK)
}
```

### 健康检查

Kubernetes 环境下的健康检查配置：

```go
// 就绪探针：服务准备好接收流量
http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
    if isReady() {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
})

// 存活探针：服务进程健康
http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})
```

### 流量排空

上线前需要排空当前实例的流量：

1. 从负载均衡器摘除节点（标记为不可用）
2. 等待正在处理的请求完成（检查 inflight 计数器）
3. 执行优雅关闭
4. 启动新版本
5. 健康检查通过后重新加入负载均衡

::: info
Kubernetes 通过 Pod 的 `preStop` 钩子和就绪探针实现流量排空。`preStop` 中建议 sleep 一小段时间，让 kube-proxy 有时间更新 iptables 规则。
:::

## 5. 面试常见问题

### 如何设计一个 99.99% 可用的系统？

1. **消除单点**：所有组件至少双副本部署
2. **故障隔离**：一个组件故障不影响其他组件（熔断、隔离舱）
3. **快速检测**：秒级监控告警，自动故障发现
4. **快速恢复**：自动切换、自动重启、自动扩容
5. **容量规划**：预留 30% 以上的冗余容量
6. **变更管理**：灰度发布、快速回滚能力

### 故障演练（Chaos Engineering）

故障演练是验证系统高可用能力的有效手段：

- **网络故障**：模拟网络延迟、丢包、分区
- **节点故障**：随机 kill 进程或重启机器
- **依赖故障**：模拟下游服务超时或返回错误
- **资源耗尽**：模拟 CPU、内存、磁盘满

常用工具：Chaos Mesh、LitmusChaos、Gremlin

## 面试追问

- 什么是故障域？如何通过故障域划分提高可用性？
- 熔断器（Circuit Breaker）的三种状态是什么？如何配置阈值？
- 如何实现灰度发布（金丝雀发布）？需要注意什么？
- 限流、降级、熔断的区别和联系？
- 如何设计一个监控系统来快速发现和定位故障？
- 你参与过的最长故障是什么？复盘学到了什么？
