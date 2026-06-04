# Golang 面试知识点深度解析

每个目录包含可运行的代码示例和详细的原理分析。

## 目录

### [01-basics 基础篇](./01-basics/)
| 主题 | 核心知识点 |
|------|-----------|
| [slice](./01-basics/slice/) | 底层结构、扩容机制、copy语义、性能陷阱 |
| [map](./01-basics/map/) | 哈希实现、扩容策略、并发安全、遍历随机性 |
| [string](./01-basics/string/) | 底层结构、byte/rune区别、拼接优化 |
| [interface](./01-basics/interface/) | eface/iface、nil陷阱、类型断言、鸭子类型 |
| [defer](./01-basics/defer/) | 执行顺序、参数求值时机、open/coded优化 |
| [pointer](./01-basics/pointer/) | 值/指针语义、逃逸分析基础 |

### [02-concurrency 并发篇](./02-concurrency/)
| 主题 | 核心知识点 |
|------|-----------|
| [goroutine](./02-concurrency/goroutine/) | GMP调度模型、goroutine泄漏、GOSCHED |
| [channel](./02-concurrency/channel/) | 底层hchan结构、收发规则、常见模式 |
| [sync](./02-concurrency/syncpkg/) | Mutex/RWMutex/WaitGroup/Once/Pool/Map |
| [context](./02-concurrency/contextx/) | 取消传播、超时控制、值传递 |
| [pattern](./02-concurrency/pattern/) | fan-in/fan-out、pipeline、worker pool、errgroup |

### [03-memory 内存篇](./03-memory/)
| 主题 | 核心知识点 |
|------|-----------|
| [gc](./03-memory/gc/) | 三色标记、混合写屏障、GC调优、GOGC |
| [escape](./03-memory/escape/) | 逃逸分析规则、栈堆分配、性能影响 |
| [alignment](./03-memory/alignment/) | 内存对齐规则、struct布局优化 |

### [04-performance 性能篇](./04-performance/)
| 主题 | 核心知识点 |
|------|-----------|
| [profiling](./04-performance/profiling/) | pprof CPU/内存/goroutine/锁分析 |
| [benchmark](./04-performance/benchmark/) | 基准测试、benchstat、优化案例 |

### [05-webframework Web 框架](./05-webframework/)
| 主题 | 核心知识点 |
|------|-----------|
| [router](./05-webframework/router/) | 压缩基数树、路由匹配优先级、路由分组 |
| [middleware](./05-webframework/middleware/) | 洋葱模型、c.Next/c.Abort、常用中间件实战 |
| [binding](./05-webframework/binding/) | ShouldBind系列、自定义校验器、JSON零值问题 |
| [restful](./05-webframework/restful/) | RESTful规范、统一响应格式、API版本控制、幂等性 |
| [errors](./05-webframework/errors/) | errors.Is/As、自定义错误类型、错误码体系 |

### [06-database 数据库](./06-database/)
| 主题 | 核心知识点 |
|------|-----------|
| [mysql-index](./06-database/mysql-index/) | B+树、聚簇索引/二级索引、覆盖索引、索引失效 |
| [mysql-tx](./06-database/mysql-tx/) | ACID、隔离级别/MVCC、行锁/间隙锁/临键锁、死锁 |
| [redis-cache](./06-database/redis-cache/) | 缓存模式、穿透/击穿/雪崩、淘汰策略、持久化 |
| [redis-struct](./06-database/redis-struct/) | 分布式锁、Redis集群、Pipeline/Lua脚本、大Key/热Key |
| [gosql](./06-database/gosql/) | driver注册、连接池管理、预处理、事务处理 |
| [gorm](./06-database/gorm/) | CRUD最佳实践、关联预加载、批量操作、N+1问题 |

### [07-microservice 微服务](./07-microservice/)
| 主题 | 核心知识点 |
|------|-----------|
| [grpc](./07-microservice/grpc/) | Protobuf编码、四种通信模式、拦截器、负载均衡 |
| [registry](./07-microservice/registry/) | 服务注册模式、注册中心对比、健康检查、优雅下线 |
| [tracing](./07-microservice/tracing/) | OpenTelemetry、Context传播、采样策略 |
| [circuitbreaker](./07-microservice/circuitbreaker/) | 令牌桶/滑动窗口、熔断器、分布式限流、自适应限流 |
| [config](./07-microservice/config/) | Viper、配置中心、多环境管理、敏感信息 |

### [08-systemdesign 系统设计](./08-systemdesign/)
| 主题 | 核心知识点 |
|------|-----------|
| [mq](./08-systemdesign/mq/) | 消息模型、Kafka高性能原理、消息可靠性、Go Kafka实战 |
| [ratelimit](./08-systemdesign/ratelimit/) | 四种限流算法、分布式限流、自适应BBR |
| [highavail](./08-systemdesign/highavail/) | SLA/SLO/SLI、冗余设计、优雅上下线、多活架构 |
| [cap](./08-systemdesign/cap/) | CAP定理、BASE理论、一致性模型、分布式ID |
| [consensus](./08-systemdesign/consensus/) | Raft选举/日志复制、Paxos、hashicorp/raft |

### [09-algorithm 算法](./09-algorithm/)
| 主题 | 核心知识点 |
|------|-----------|
| [sorting](./09-algorithm/sorting/) | 冒泡/插入/快排/归并/堆排序、Go pdqsort |
| [binary](./09-algorithm/binary/) | 基础二分、变体二分、浮点数二分、旋转数组 |
| [hash](./09-algorithm/hash/) | 哈希函数、冲突解决、Go map原理、sync.Map |
| [tree](./09-algorithm/tree/) | 遍历、BST、红黑树、堆/优先队列、TopK |
| [dp](./09-algorithm/dp/) | 线性DP、背包问题、LCS/编辑距离、状态转移推导 |
| [gostdlib](./09-algorithm/gostdlib/) | sort/sort.Search、container/heap、位运算技巧 |

### [10-network 网络](./10-network/)
| 主题 | 核心知识点 |
|------|-----------|
| [tcp](./10-network/tcp/) | 三次握手/四次挥手、拥塞控制、TIME_WAIT、粘包 |
| [http](./10-network/http/) | HTTP/1.1队头阻塞、HTTP/2多路复用、HTTP/3 QUIC |
| [https](./10-network/https/) | TLS握手流程、证书体系、mTLS、Go TLS编程 |
| [nethttp](./10-network/nethttp/) | goroutine-per-conn、ServeMux、Handler接口、优雅关闭 |
| [websocket](./10-network/websocket/) | 帧协议、Go WebSocket实战、并发写保护、Hub模式 |

### [11-os 操作系统](./11-os/)
| 主题 | 核心知识点 |
|------|-----------|
| [process](./11-os/process/) | 进程/线程/协程对比、GMP模型、信号处理 |
| [virtualmem](./11-os/virtualmem/) | TCMalloc思想、Go内存分配器、栈扩缩容、mmap |
| [iomodel](./11-os/iomodel/) | 五种IO模型、epoll原理、Go netpoller、零拷贝 |
| [filesystem](./11-os/filesystem/) | inode、文件锁、fsnotify、日志轮转 |
| [linuxcmd](./11-os/linuxcmd/) | 进程/网络/磁盘排查、pprof火焰图、OOM/CPU/死锁排查 |

## 运行方式

```bash
# 运行任意示例
go run ./01-basics/slice/main.go

# 运行基准测试
go test -bench=. ./04-performance/benchmark/

# 查看逃逸分析
go build -gcflags="-m" ./03-memory/escape/
```
