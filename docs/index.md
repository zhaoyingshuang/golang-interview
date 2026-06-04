---
layout: home

hero:
  name: Golang 面试知识点
  text: 原理 · 实战 · 深度解析
  tagline: 从底层源码到生产实践，涵盖 16 个核心主题
  actions:
    - theme: brand
      text: 开始学习
      link: /basics/slice
    - theme: alt
      text: GitHub
      link: https://github.com/zhaoyingshuang/golang-interview
---

## 内容概览

### [基础篇](/basics/slice)

Slice、Map、String、Interface、Defer、Pointer 的底层原理与常见陷阱

### [并发篇](/concurrency/goroutine)

GMP 调度模型、Channel、Sync 原语、Context、并发设计模式

### [内存篇](/memory/gc)

三色标记 GC、逃逸分析、内存对齐与 struct 优化

### [性能篇](/performance/profiling)

pprof 性能分析、Benchmark 基准测试与优化实战

### [Web 框架](/webframework/router)

路由树原理、中间件洋葱模型、参数绑定、RESTful 设计、错误处理

### [数据库](/database/mysql-index)

MySQL 索引/事务/锁、Redis 缓存/分布式锁、Go database/sql、GORM

### [微服务](/microservice/grpc)

gRPC、服务注册发现、链路追踪、限流熔断降级、配置管理

### [系统设计](/systemdesign/mq)

消息队列、限流设计、高可用架构、CAP 理论、分布式一致性算法

### [算法](/algorithm/sorting)

排序、二分查找、哈希表、树、动态规划、Go 标准库算法

### [网络](/network/tcp)

TCP/IP、HTTP/2/3、HTTPS/TLS、Go net/http 源码、WebSocket

### [操作系统](/os/process)

进程/线程/Goroutine、虚拟内存、IO 模型、文件系统、Linux 排查

## 项目特色

- **底层源码分析** — 每个知识点追溯到 runtime 源码级别
- **生产实战经验** — 标注真实场景中的使用建议和常见坑点
- **面试追问预演** — 每节附带你可能被追问的问题和回答方向
- **可运行代码** — 所有示例均可 `go run` 运行验证
