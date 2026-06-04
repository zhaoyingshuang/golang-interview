package main

import (
	"fmt"
)

// ============================================================
// 消息队列
// ============================================================
//
// 【面试高频问题】
// 1. 消息队列的使用场景？
// 2. 如何保证消息不丢失？
// 3. 消息重复消费怎么处理？
// 4. Kafka 为什么这么快？
// 5. 消息积压怎么处理？

func main() {
	messageModel()
	reliability()
	ordering()
	kafkaPerformance()
	goKafkaExample()
}

// ----------------------------------------------------------
// 1. 消息模型
// ----------------------------------------------------------
func messageModel() {
	fmt.Println("=== 1. 消息模型 ===")
	fmt.Println()
	fmt.Println("点对点 (Queue):")
	fmt.Println("  生产者 → Queue → 消费者 (一对一)")
	fmt.Println("  消息被一个消费者消费后删除")
	fmt.Println("  适用: 任务分发 (订单处理)")
	fmt.Println()
	fmt.Println("发布订阅 (Topic):")
	fmt.Println("  生产者 → Topic → 消费者组1 (一对多)")
	fmt.Println("                   → 消费者组2")
	fmt.Println("  每个消费者组都能收到全量消息")
	fmt.Println("  适用: 事件通知 (订单创建 → 通知库存/物流/积分)")
	fmt.Println()
	fmt.Println("Kafka 核心概念:")
	fmt.Println("  Topic     — 消息主题 (逻辑分类)")
	fmt.Println("  Partition — 分区 (并行度、有序性保证)")
	fmt.Println("  Offset    — 消费位置 (分区内的单调递增偏移)")
	fmt.Println("  Consumer Group — 消费者组 (组内负载均衡)")
	fmt.Println()
	fmt.Println("  Topic: orders")
	fmt.Println("  ┌─── Partition 0: [msg0, msg1, msg2, ...]  → Consumer A")
	fmt.Println("  ├─── Partition 1: [msg0, msg1, msg2, ...]  → Consumer B")
	fmt.Println("  └─── Partition 2: [msg0, msg1, msg2, ...]  → Consumer C")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 消息可靠性
// ----------------------------------------------------------
func reliability() {
	fmt.Println("=== 2. 消息可靠性 ===")
	fmt.Println()
	fmt.Println("三个环节保证不丢失:")
	fmt.Println()
	fmt.Println("  1. 生产者 → Broker (发送确认)")
	fmt.Println("     Kafka: acks=all (等待所有副本确认)")
	fmt.Println("     RabbitMQ: publisher confirm")
	fmt.Println()
	fmt.Println("  2. Broker 持久化")
	fmt.Println("     Kafka: 消息写入磁盘 + 副本复制")
	fmt.Println("     RabbitMQ: 队列 durable + 消息 persistent")
	fmt.Println()
	fmt.Println("  3. 消费者 → 手动 ACK")
	fmt.Println("     处理完成后才确认，失败不确认 → 消息重新投递")
	fmt.Println("     Kafka: enable.auto.commit=false，手动提交 offset")
	fmt.Println()
	fmt.Println("Exactly-Once 语义:")
	fmt.Println("  Kafka 通过以下机制实现:")
	fmt.Println("  1. 幂等生产者 (Idempotent Producer): PID + Sequence Number")
	fmt.Println("  2. 事务 (Transactional): 原子写入多个分区")
	fmt.Println("  3. 消费端: 事务消费 (read-process-write 原子)")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 消息顺序
// ----------------------------------------------------------
func ordering() {
	fmt.Println("=== 3. 消息顺序 ===")
	fmt.Println()
	fmt.Println("Kafka 的顺序保证:")
	fmt.Println("  ✓ Partition 内有序 — 按 offset 严格顺序消费")
	fmt.Println("  ✗ Topic 级别无序 — 不同 Partition 之间无序")
	fmt.Println()
	fmt.Println("保证全局有序:")
	fmt.Println("  方案: 只用一个 Partition (牺牲并行度)")
	fmt.Println("  适用: 需要严格顺序的场景（如数据库 binlog 同步）")
	fmt.Println()
	fmt.Println("保证业务有序:")
	fmt.Println("  方案: 用相同 key 的消息发到同一个 Partition")
	fmt.Println("  例: orderId 作为 key → 同一订单的消息有序")
	fmt.Println()
	fmt.Println("  producer.Send(&kafka.Message{")
	fmt.Println("    Key:   []byte(orderID),  // 相同 key → 同一 partition")
	fmt.Println("    Value: []byte(payload),")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("面试追问: 消费者重平衡时顺序怎么保证?")
	fmt.Println("  答: 重平衡期间会暂停消费，重新分配后从 committed offset 继续")
	fmt.Println("  可能重复消费（上次提交 offset 到当前之间的消息），需要幂等处理")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. Kafka 为什么快
// ----------------------------------------------------------
func kafkaPerformance() {
	fmt.Println("=== 4. Kafka 高性能原理 ===")
	fmt.Println()
	fmt.Println("  1. 顺序写 (Append-Only Log)")
	fmt.Println("     磁盘顺序写 > 随机写 (600MB/s vs 100KB/s)")
	fmt.Println("     消息追加到日志末尾，不修改已有数据")
	fmt.Println()
	fmt.Println("  2. 零拷贝 (Zero Copy / sendfile)")
	fmt.Println("     传统: disk → kernel → user → kernel → nic (4次拷贝)")
	fmt.Println("     零拷贝: disk → kernel → nic (2次拷贝)")
	fmt.Println("     Java: FileChannel.transferTo / Go: syscall.Sendfile")
	fmt.Println()
	fmt.Println("  3. PageCache")
	fmt.Println("     利用操作系统的页缓存，避免 JVM GC 压力")
	fmt.Println("     写入: 先写 PageCache，异步刷盘")
	fmt.Println("     读取: 直接从 PageCache 读 (命中时无磁盘 IO)")
	fmt.Println()
	fmt.Println("  4. 分区并行")
	fmt.Println("     多 Partition → 多 Producer/Consumer 并行")
	fmt.Println("     水平扩展: 增加 Partition → 增加吞吐")
	fmt.Println()
	fmt.Println("  5. 批量操作")
	fmt.Println("     生产者: 微批聚合 (linger.ms / batch.size)")
	fmt.Println("     消费者: 批量拉取 (fetch.min.bytes)")
	fmt.Println("     减少 RPC 次数，提高网络利用率")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. Go Kafka 实战
// ----------------------------------------------------------
func goKafkaExample() {
	fmt.Println("=== 5. Go Kafka 实战 (segmentio/kafka-go) ===")
	fmt.Println()
	fmt.Println("生产者:")
	fmt.Println("  writer := &kafka.Writer{")
	fmt.Println("    Addr:         kafka.TCP(\"localhost:9092\"),")
	fmt.Println("    Topic:        \"orders\",")
	fmt.Println("    Balancer:     &kafka.LeastBytes{},  // 负载均衡")
	fmt.Println("    BatchTimeout: 10 * time.Millisecond,")
	fmt.Println("    RequiredAcks: kafka.RequireAll,     // acks=all")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("  err := writer.WriteMessages(ctx, kafka.Message{")
	fmt.Println("    Key:   []byte(orderID),")
	fmt.Println("    Value: []byte(jsonPayload),")
	fmt.Println("    Headers: []kafka.Header{{")
	fmt.Println("      Key:   \"trace-id\",")
	fmt.Println("      Value: []byte(traceID),")
	fmt.Println("    }},")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("消费者:")
	fmt.Println("  reader := kafka.NewReader(kafka.ReaderConfig{")
	fmt.Println("    Brokers:   []string{\"localhost:9092\"},")
	fmt.Println("    Topic:     \"orders\",")
	fmt.Println("    GroupID:   \"order-processor\",")
	fmt.Println("    MinBytes:  10e3,   // 10KB")
	fmt.Println("    MaxBytes:  10e6,   // 10MB")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("  for {")
	fmt.Println("    msg, err := reader.ReadMessage(ctx)")
	fmt.Println("    // 处理消息...")
	fmt.Println("    if err := process(msg); err == nil {")
	fmt.Println("      reader.CommitMessages(ctx, msg) // 手动提交")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("消息积压处理:")
	fmt.Println("  1. 增加消费者实例 (≤ Partition 数量)")
	fmt.Println("  2. 临时增加 Partition + 消费者")
	fmt.Println("  3. 跳过非关键消息，先追上最新进度")
	fmt.Println("  4. 批量消费，减少单条处理开销")
	fmt.Println()
}
