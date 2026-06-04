package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================
// 服务注册与发现
// ============================================================
//
// 【面试高频问题】
// 1. 服务注册的两种模式？
// 2. Consul 和 Etcd 的区别？
// 3. 注册中心挂了怎么办？
// 4. 健康检查有哪些方式？
// 5. 如何实现优雅下线？

func main() {
	registryPattern()
	healthCheck()
	consulVsEtcd()
	gracefulShutdown()
}

// ----------------------------------------------------------
// 1. 服务注册模式
// ----------------------------------------------------------
// 两种基本模式:
//
// 客户端发现 (Client-Side Discovery):
//   客户端查询注册中心 → 获取实例列表 → 自行选择 → 直接调用
//   例: Spring Cloud Eureka, Consul DNS
//
// 服务端发现 (Server-Side Discovery):
//   客户端 → 负载均衡器/网关 → 查询注册中心 → 转发到实例
//   例: Nginx + Consul Template, Kubernetes Service
//
// 注册方式:
//   自注册: 服务启动时自行注册到注册中心
//   第三方注册: 通过外部系统(如 Kubernetes)管理注册
func registryPattern() {
	fmt.Println("=== 1. 服务注册模式 ===")

	// 模拟服务注册中心
	type ServiceInstance struct {
		ID      string
		Name    string
		Address string
		Port    int
		Healthy bool
	}

	registry := struct {
		sync.RWMutex
		services map[string][]ServiceInstance
	}{
		services: make(map[string][]ServiceInstance),
	}

	// 注册服务
	register := func(instance ServiceInstance) {
		registry.Lock()
		defer registry.Unlock()
		registry.services[instance.Name] = append(
			registry.services[instance.Name], instance,
		)
		fmt.Printf("  注册: %s → %s:%d\n", instance.Name, instance.Address, instance.Port)
	}

	// 发现服务
	discover := func(serviceName string) []ServiceInstance {
		registry.RLock()
		defer registry.RUnlock()
		instances := registry.services[serviceName]
		healthy := make([]ServiceInstance, 0)
		for _, inst := range instances {
			if inst.Healthy {
				healthy = append(healthy, inst)
			}
		}
		return healthy
	}

	// 模拟注册 3 个实例
	register(ServiceInstance{"user-1", "user-service", "10.0.1.1", 8080, true})
	register(ServiceInstance{"user-2", "user-service", "10.0.1.2", 8080, true})
	register(ServiceInstance{"user-3", "user-service", "10.0.1.3", 8080, true})

	// 模拟服务发现
	instances := discover("user-service")
	fmt.Printf("  发现 %d 个健康实例\n", len(instances))
	for _, inst := range instances {
		fmt.Printf("    → %s:%d\n", inst.Address, inst.Port)
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 健康检查机制
// ----------------------------------------------------------
func healthCheck() {
	fmt.Println("=== 2. 健康检查 ===")

	fmt.Println("常见健康检查方式:")
	fmt.Println()
	fmt.Println("  1. TTL 心跳 (Consul 默认):")
	fmt.Println("     服务定期发送心跳 → 超时未发送 → 标记不健康")
	fmt.Println("     agent.CheckTTL(\"service:web\", TTL: 10s)")
	fmt.Println()
	fmt.Println("  2. HTTP/TCP 主动探测 (Consul/Etcd):")
	fmt.Println("     注册中心定期请求 /health → 非 200 → 不健康")
	fmt.Println("     consul.AgentServiceCheck{HTTP: \"http://:8080/health\", Interval: \"10s\"}")
	fmt.Println()
	fmt.Println("  3. gRPC 健康检查 (标准协议):")
	fmt.Println("     grpc.health.v1.Health/Check")
	fmt.Println("     google.golang.org/grpc/health")
	fmt.Println()

	// 模拟 TTL 健康检查
	type HealthStatus struct {
		LastHeartbeat time.Time
		TTL           time.Duration
	}

	check := HealthStatus{
		LastHeartbeat: time.Now().Add(-5 * time.Second),
		TTL:           10 * time.Second,
	}

	isHealthy := time.Since(check.LastHeartbeat) < check.TTL
	fmt.Printf("  TTL 检查: 最后心跳=%v前, TTL=%v, 健康=%v\n",
		5*time.Second, check.TTL, isHealthy)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 注册中心对比
// ----------------------------------------------------------
func consulVsEtcd() {
	fmt.Println("=== 3. 注册中心对比 ===")

	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "特性", "Consul", "Etcd", "Nacos")
	fmt.Println("  " + strings.Repeat("─", 60))
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "CAP", "CP", "CP", "AP/CP可切换")
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "一致性协议", "Raft", "Raft", "Raft/Distro")
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "健康检查", "TTL/HTTP/TCP", "Lease/HTTP", "心跳/TCP/HTTP")
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "KV存储", "✓", "✓", "✓")
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "服务网格", "Connect", "无", "无")
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "配置管理", "✓", "✓", "✓(强)")
	fmt.Printf("  %-12s %-15s %-15s %-15s\n", "多数据中心", "✓", "需要ETCD集群", "✓")
	fmt.Println()
	fmt.Println("  选型建议:")
	fmt.Println("    Kubernetes 环境 → Etcd (生态融合)")
	fmt.Println("    通用微服务     → Consul (功能全面)")
	fmt.Println("    Java/Spring    → Nacos (阿里生态)")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 优雅下线
// ----------------------------------------------------------
func gracefulShutdown() {
	fmt.Println("=== 4. 优雅下线 ===")

	fmt.Println("优雅下线步骤:")
	fmt.Println("  1. 收到 SIGTERM 信号")
	fmt.Println("  2. 从注册中心注销 (Deregister)")
	fmt.Println("  3. 停止接受新连接")
	fmt.Println("  4. 等待进行中的请求完成 (drain)")
	fmt.Println("  5. 关闭 gRPC/HTTP 连接")
	fmt.Println("  6. 关闭数据库、缓存连接")
	fmt.Println("  7. 进程退出")
	fmt.Println()
	fmt.Println("Go 实现:")
	fmt.Println("  sigCh := make(chan os.Signal, 1)")
	fmt.Println("  signal.Notify(sigCh, syscall.SIGTERM)")
	fmt.Println("  <-sigCh")
	fmt.Println()
	fmt.Println("  // 1. 从注册中心注销")
	fmt.Println("  consulClient.Agent().ServiceDeregister(serviceID)")
	fmt.Println()
	fmt.Println("  // 2. 优雅关闭 HTTP 服务")
	fmt.Println("  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)")
	fmt.Println("  defer cancel()")
	fmt.Println("  srv.Shutdown(ctx)")
	fmt.Println()
	fmt.Println("面试追问: 注册中心挂了怎么办?")
	fmt.Println("  答: 客户端缓存服务列表，注册中心恢复后重新同步")
	fmt.Println("  本地缓存 + 定期刷新 → 即使注册中心短暂不可用，服务仍可通信")
	fmt.Println()
}
