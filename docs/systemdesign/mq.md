---
title: 消息队列
---

## 1. 消息模型

### 点对点与发布订阅

消息队列有两种核心模型：

- **点对点（Point-to-Point）**：一条消息只能被一个消费者消费，消费后从队列中移除。适用于任务分发场景。
- **发布订阅（Pub/Sub）**：一条消息可以被多个订阅者消费，每个订阅者都能收到完整消息副本。适用于事件通知场景。

### Topic / Partition / Consumer Group

以 Kafka 为例的核心概念：

| 概念 | 说明 |
|------|------|
| Topic | 消息的逻辑分类，生产者向 Topic 发送消息 |
| Partition | Topic 的物理分片，每个 Partition 是一个有序日志 |
| Consumer Group | 消费者组，组内消费者共同分担 Partition 的消费 |
| Offset | 消费者在 Partition 中的消费位置 |

::: tip
同一个 Consumer Group 内，一个 Partition 只能被一个消费者消费。因此消费者数量不应超过 Partition 数量，否则多余的消费者将处于空闲状态。
:::

## 2. 消息可靠性

### 生产者确认

生产者发送消息后需要等待 Broker 确认，以确保消息成功写入：

- **同步发送**：阻塞等待 ACK，可靠性最高但吞吐量最低
- **异步发送**：提供回调函数，兼顾可靠性与性能
- **重试机制**：网络异常时自动重试，需注意幂等性

### 消费者 ACK

消费者处理完消息后必须显式确认（ACK），否则 Broker 会重新投递：

- **自动提交**：消费者拉取消息后自动提交 Offset，可能丢消息
- **手动提交**：处理完成后手动提交，确保至少一次消费

### 持久化

消息持久化策略决定了消息的可靠性级别：

- 异步刷盘：性能高，可能丢少量消息
- 同步刷盘：性能低，可靠性高
- 多副本同步：最高可靠性，消息写入多个副本才算成功

### Exactly-Once 语义

::: info
三种消息语义：At Most Once（最多一次）、At Least Once（至少一次）、Exactly Once（精确一次）。实际生产中通常实现 At Least Once + 幂等消费，等价于 Exactly Once。
:::

实现 Exactly-Once 的关键手段：

1. 生产者幂等：为每条消息分配唯一 ID（Producer ID + Sequence Number）
2. 事务消息：一批消息要么全部成功，要么全部失败
3. 消费者幂等：业务层去重，通过唯一键保证

## 3. 消息顺序

### Partition 内有序

Kafka 保证同一 Partition 内消息的顺序性。发送到同一 Partition 的消息会按照发送顺序存储和消费。

保证同一业务 key 的消息进入同一 Partition：

```go
// 通过 Key 的 Hash 决定 Partition，相同 Key 的消息一定进入同一 Partition
msg := &kafka.Message{
    TopicPartition: kafka.TopicPartition{Topic: &topic},
    Key:   []byte(orderID), // 相同订单 ID 保证进入同一 Partition
    Value: []byte(payload),
}
```

### 全局有序的代价

全局有序要求所有消息进入同一个 Partition，代价极大：

- 吞吐量严重下降（丧失并行能力）
- 无法横向扩展
- 单点故障风险

::: warning
除非业务强制要求全局有序（如 binlog 同步），否则应优先使用 Partition 内有序。大部分业务场景只需保证同一实体的操作有序即可。
:::

## 4. 常见消息队列对比

| 特性 | Kafka | RabbitMQ | RocketMQ | Pulsar |
|------|-------|----------|----------|--------|
| 定位 | 分布式流平台 | 传统消息代理 | 金融级消息 | 云原生消息 |
| 吞吐量 | 极高（百万级） | 万级 | 十万级 | 极高 |
| 延迟 | ms 级 | us 级 | ms 级 | ms 级 |
| 消息可靠性 | 高（副本） | 高（镜像队列） | 极高（同步刷盘） | 高（BookKeeper） |
| 顺序消息 | Partition 级 | 队列级 | 队列级 | Partition 级 |
| 事务消息 | 支持 | 不支持 | 支持 | 支持 |
| 适用场景 | 日志/大数据 | 业务解耦 | 电商/金融 | 多租户/SaaS |

## 5. Go 消息队列实战

### Kafka 生产者（segmentio/kafka-go）

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

func main() {
    w := &kafka.Writer{
        Addr:         kafka.TCP("localhost:9092"),
        Topic:        "orders",
        Balancer:     &kafka.Hash{}, // 相同 Key 路由到同一 Partition
        BatchTimeout: 10 * time.Millisecond,
        RequiredAcks: kafka.RequireAll, // 等待所有副本确认
        MaxAttempts:  3,
    }
    defer w.Close()

    err := w.WriteMessages(context.Background(),
        kafka.Message{
            Key:   []byte("order-123"),
            Value: []byte(`{"event":"created","amount":99.9}`),
        },
    )
    if err != nil {
        log.Fatal("写入消息失败:", err)
    }
}
```

### Kafka 消费者与幂等处理

```go
package main

import (
    "context"
    "log"

    "github.com/segmentio/kafka-go"
)

func main() {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  []string{"localhost:9092"},
        Topic:    "orders",
        GroupID:  "order-service",
        MinBytes: 10e3,
        MaxBytes: 10e6,
    })
    defer r.Close()

    // 幂等消费：使用消息 Key 作为去重 ID
    processed := make(map[string]bool)

    for {
        m, err := r.ReadMessage(context.Background())
        if err != nil {
            log.Println("读取消息失败:", err)
            continue
        }

        msgKey := string(m.Key)
        if processed[msgKey] {
            log.Printf("重复消息 %s，跳过处理\n", msgKey)
            continue
        }

        // 处理业务逻辑
        if err := processOrder(m.Value); err != nil {
            log.Printf("处理消息 %s 失败: %v\n", msgKey, err)
            continue
        }

        processed[msgKey] = true
        log.Printf("成功处理消息: %s\n", msgKey)
    }
}

func processOrder(data []byte) error {
    // 业务处理逻辑
    return nil
}
```

::: warning
生产环境中的幂等处理不应使用内存 Map，应使用 Redis 或数据库唯一索引来保证去重，避免进程重启后状态丢失。
:::

## 6. 面试常见问题

### 消息积压如何处理？

1. 紧急扩容消费者实例（需先增加 Partition 数量）
2. 临时消费者方案：新建临时 Topic + 临时消费者，快速转发消息
3. 排查积压根因：消费者处理慢、下游服务超时、数据库瓶颈
4. 长期方案：优化消费逻辑、批量消费、异步处理

### 消息丢失如何排查？

按链路逐环节排查：

- **生产端**：是否使用同步发送或回调确认？是否忽略了发送错误？
- **Broker 端**：副本数是否足够？同步刷盘还是异步刷盘？
- **消费端**：是否先提交 Offset 再处理？是否开启了自动提交？

### 重复消费如何处理？

1. 消费端幂等：业务唯一键 + 数据库唯一索引
2. 消费端去重：Redis SETNX 记录已处理的消息 ID
3. 乐观锁：更新时携带版本号，避免重复扣减

## 面试追问

- 如何保证消息的 Exactly-Once 语义？生产端和消费端分别怎么做？
- Kafka 的 Consumer Group Rebalance 机制是什么？会导致什么问题？
- 如何设计一个延迟消息队列？RocketMQ 延迟级别的实现原理？
- 消息队列如何实现消息回溯（重新消费历史消息）？
- Kafka 为什么这么快？零拷贝、页缓存、顺序写磁盘各自的作用？
- 如何监控消息队列的 Lag（积压量）？告警阈值怎么设置？
