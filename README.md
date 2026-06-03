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

## 运行方式

```bash
# 运行任意示例
go run ./01-basics/slice/main.go

# 运行基准测试
go test -bench=. ./04-performance/benchmark/

# 查看逃逸分析
go build -gcflags="-m" ./03-memory/escape/
```
