---
title: 服务注册与发现
---

## 1. 两种发现模式

**客户端发现 (Client-Side Discovery)**：客户端查询注册中心获取实例列表，自行选择并直接调用。代表：Spring Cloud Eureka、Consul DNS。

**服务端发现 (Server-Side Discovery)**：客户端请求负载均衡器/网关，由其查询注册中心并转发。代表：Nginx + Consul Template、Kubernetes Service。

::: tip 注册方式
- **自注册**：服务启动时自行注册（简单但耦合）
- **第三方注册**：通过 K8s/Registrator 等外部系统管理注册（解耦但复杂）
:::

## 2. 注册中心对比

| 特性 | Consul | Etcd | Nacos |
|------|--------|------|-------|
| CAP | CP | CP | AP/CP 可切换 |
| 一致性协议 | Raft | Raft | Raft/Distro |
| 健康检查 | TTL/HTTP/TCP | Lease/HTTP | 心跳/TCP/HTTP |
| 多数据中心 | ✓ | 需集群 | ✓ |
| 配置管理 | ✓ | ✓ | ✓(强) |

**选型建议**：K8s 环境 → Etcd；通用微服务 → Consul；Java/Spring → Nacos。

## 3. 健康检查机制

::: info 常见方式
1. **TTL 心跳**（Consul 默认）：服务定期发送心跳，超时未发送 → 标记不健康
2. **HTTP/TCP 主动探测**：注册中心定期请求 `/health`，非 200 → 不健康
3. **gRPC 健康检查**：标准协议 `grpc.health.v1.Health/Check`
:::

```go
// Consul 健康检查配置
consul.AgentServiceCheck{
    HTTP:     "http://localhost:8080/health",
    Interval: "10s",
    Timeout:  "3s",
}
```

## 4. 负载均衡策略

- **Round Robin**：轮询，最简单
- **Weighted Round Robin**：加权轮询，适合异构实例
- **一致性哈希**：相同 key 路由到同一实例（会话保持）
- **自适应**：基于延迟/错误率动态调整

## 5. 优雅下线与容灾

::: warning 面试高频
注册中心挂了怎么办？→ 客户端缓存服务列表，恢复后重新同步。本地缓存 + 定期刷新保证短时间可用。

```go
// 优雅下线流程
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGTERM)
<-sigCh
// 1. 注销服务
consulClient.Agent().ServiceDeregister(serviceID)
// 2. 优雅关闭 HTTP
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
srv.Shutdown(ctx)
```
:::
