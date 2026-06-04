package main

import (
	"fmt"
)

// ============================================================
// 分布式链路追踪
// ============================================================
//
// 【面试高频问题】
// 1. 什么是 Trace/Span？它们的关系？
// 2. W3C Trace Context 如何传播？
// 3. Go 如何接入 OpenTelemetry？
// 4. 采样策略有哪些？
// 5. 异步任务如何传递 TraceID？

func main() {
	otelConcepts()
	contextPropagation()
	goIntegration()
	samplingStrategy()
	asyncTracing()
}

// ----------------------------------------------------------
// 1. OpenTelemetry 核心概念
// ----------------------------------------------------------
// Trace: 一次完整的请求链路，由一个 TraceID 标识
// Span:  链路中的一个操作，由一个 SpanID 标识
//
// 一个 Trace 包含多个 Span，形成一棵树:
//
//   Trace (TraceID: abc123)
//   ├── Span: HTTP GET /api/users (SpanID: s1, ParentID: root)
//   │   ├── Span: DB Query SELECT (SpanID: s2, ParentID: s1)
//   │   └── Span: Redis GET (SpanID: s3, ParentID: s1)
//   └── Span: HTTP Response 200 (SpanID: s4, ParentID: root)
//
// Span 包含:
//   - TraceID / SpanID / ParentSpanID
//   - Operation Name (操作名)
//   - Start Time / End Time
//   - Attributes (属性键值对)
//   - Events (事件)
//   - Status (状态: OK/Error)
//   - Links (关联其他 Span)
func otelConcepts() {
	fmt.Println("=== 1. OpenTelemetry 核心概念 ===")
	fmt.Println()
	fmt.Println("  Trace (链路) — 一次完整请求，唯一 TraceID")
	fmt.Println("  Span (跨度)  — 一个操作，唯一 SpanID")
	fmt.Println("  关系: Trace 包含多个 Span，形成树结构")
	fmt.Println()
	fmt.Println("  示例链路:")
	fmt.Println("  TraceID: abc123")
	fmt.Println("  [s1: HTTP GET /api/users] ← root span")
	fmt.Println("    ├── [s2: DB SELECT users] ← child of s1")
	fmt.Println("    └── [s3: Redis GET cache] ← child of s1")
	fmt.Println()
	fmt.Println("  Span 数据结构:")
	fmt.Println("  type Span struct {")
	fmt.Println("    TraceID    string            // 链路ID")
	fmt.Println("    SpanID     string            // 当前跨度ID")
	fmt.Println("    ParentID   string            // 父跨度ID")
	fmt.Println("    Name       string            // 操作名")
	fmt.Println("    StartTime  time.Time")
	fmt.Println("    EndTime    time.Time")
	fmt.Println("    Attributes map[string]any    // 属性")
	fmt.Println("    Events     []Event           // 事件")
	fmt.Println("    Status     Status            // OK/Error")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Context 传播
// ----------------------------------------------------------
// W3C Trace Context 标准 (traceparent 头):
//
//   traceparent: {version}-{trace-id}-{parent-id}-{trace-flags}
//   例: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
//
//   version:     00 (固定)
//   trace-id:    32 hex digits (16 bytes)
//   parent-id:   16 hex digits (8 bytes)
//   trace-flags: 02 hex digits (sampled=01, not-sampled=00)
//
// 传播方式:
//   HTTP: traceparent / tracestate 请求头
//   gRPC: metadata 中的 traceparent key
//   消息队列: 消息的 header/property
func contextPropagation() {
	fmt.Println("=== 2. Context 传播 ===")
	fmt.Println()
	fmt.Println("W3C Trace Context (HTTP Header):")
	fmt.Println("  traceparent: 00-<trace-id>-<span-id>-<flags>")
	fmt.Println("  例: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	fmt.Println()
	fmt.Println("传播流程:")
	fmt.Println("  1. 网关收到请求 → 生成 TraceID → 注入 Header")
	fmt.Println("  2. Service A 收到请求 → 提取 TraceID → 生成子 Span")
	fmt.Println("  3. Service A 调用 Service B → 注入 TraceID 到 gRPC metadata")
	fmt.Println("  4. Service B 提取 → 继续传播")
	fmt.Println()
	fmt.Println("Go HTTP 传播:")
	fmt.Println("  // 注入 (客户端)")
	fmt.Println("  propagation := otel.GetTextMapPropagator()")
	fmt.Println("  propagation.Inject(ctx, propagation.HeaderCarrier(req.Header))")
	fmt.Println()
	fmt.Println("  // 提取 (服务端)")
	fmt.Println("  ctx = propagation.Extract(ctx, propagation.HeaderCarrier(r.Header))")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Go 接入 OpenTelemetry
// ----------------------------------------------------------
func goIntegration() {
	fmt.Println("=== 3. Go 接入 OpenTelemetry ===")
	fmt.Println()
	fmt.Println("初始化 TracerProvider:")
	fmt.Println("  exporter, _ := otlptracegrpc.New(ctx,")
	fmt.Println("    otlptracegrpc.WithEndpoint(\"localhost:4317\"),")
	fmt.Println("  )")
	fmt.Println("  tp := sdktrace.NewTracerProvider(")
	fmt.Println("    sdktrace.WithBatcher(exporter),")
	fmt.Println("    sdktrace.WithResource(resource.NewWithAttributes(")
	fmt.Println("      semconv.SchemaURL,")
	fmt.Println("      attribute.String(\"service.name\", \"user-service\"),")
	fmt.Println("    )),")
	fmt.Println("  )")
	fmt.Println("  otel.SetTracerProvider(tp)")
	fmt.Println()
	fmt.Println("gRPC 拦截器:")
	fmt.Println("  // 客户端")
	fmt.Println("  grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor())")
	fmt.Println()
	fmt.Println("  // 服务端")
	fmt.Println("  grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor())")
	fmt.Println()
	fmt.Println("HTTP 中间件:")
	fmt.Println("  handler := otelhttp.NewHandler(http.HandlerFunc(handle), \"api/users\")")
	fmt.Println()
	fmt.Println("手动创建 Span:")
	fmt.Println("  ctx, span := otel.Tracer(\"my-app\").Start(ctx, \"processOrder\")")
	fmt.Println("  defer span.End()")
	fmt.Println("  span.SetAttributes(attribute.Int64(\"orderID\", orderID))")
	fmt.Println("  span.AddEvent(\"order_validated\")")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 采样策略
// ----------------------------------------------------------
func samplingStrategy() {
	fmt.Println("=== 4. 采样策略 ===")
	fmt.Println()
	fmt.Println("为什么需要采样:")
	fmt.Println("  高流量系统全量采集成本太高（存储、带宽、处理）")
	fmt.Println("  需要在可观测性和成本之间取平衡")
	fmt.Println()
	fmt.Println("常见策略:")
	fmt.Println()
	fmt.Println("  1. 概率采样 (ProbabilitySampler):")
	fmt.Println("    每条链路有固定概率被采样")
	fmt.Println("    sdktrace.TraceIDRatioBased(0.1) // 10%")
	fmt.Println("    优点: 简单   缺点: 可能错过异常链路")
	fmt.Println()
	fmt.Println("  2. 头部采样 (ParentBasedSampler):")
	fmt.Println("    在链路开始时决定是否采样")
	fmt.Println("    一旦决定，整条链路所有 Span 都被采集")
	fmt.Println("    大多数系统的默认策略")
	fmt.Println()
	fmt.Println("  3. 尾部采样 (TailBasedSampler):")
	fmt.Println("    等链路结束后再决定是否采样")
	fmt.Println("    可以基于: 错误率 > 阈值、延迟 > P99、特定属性")
	fmt.Println("    优点: 不漏关键链路   缺点: 需要缓存所有 Span")
	fmt.Println()
	fmt.Println("  推荐组合:")
	fmt.Println("    ParentBased(RateLimiting(100/s)) + TailBased(错误优先)")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 异步任务链路传递
// ----------------------------------------------------------
func asyncTracing() {
	fmt.Println("=== 5. 异步任务链路传递 ===")
	fmt.Println()
	fmt.Println("问题: goroutine/消息队列 中如何延续链路?")
	fmt.Println()
	fmt.Println("方案1: 传递 Context (goroutine)")
	fmt.Println("  ctx, span := tracer.Start(ctx, \"async-task\")")
	fmt.Println("  go func(ctx context.Context) {")
	fmt.Println("    defer span.End()")
	fmt.Println("    // 这里的 ctx 携带了 TraceID")
	fmt.Println("    _, childSpan := tracer.Start(ctx, \"sub-task\")")
	fmt.Println("    defer childSpan.End()")
	fmt.Println("  }(ctx)")
	fmt.Println()
	fmt.Println("方案2: 消息队列传播")
	fmt.Println("  // 生产者: 注入 TraceContext 到消息 Header")
	fmt.Println("  propagator.Inject(ctx, &msg.Headers)")
	fmt.Println()
	fmt.Println("  // 消费者: 从消息 Header 提取 TraceContext")
	fmt.Println("  ctx = propagator.Extract(context.Background(), &msg.Headers)")
	fmt.Println("  ctx, span := tracer.Start(ctx, \"process-message\")")
	fmt.Println()
	fmt.Println("面试追问: Context 传递丢失的常见原因?")
	fmt.Println("  1. goroutine 没有传 ctx 参数，用了 context.Background()")
	fmt.Println("  2. 消息队列序列化时没有保留 trace header")
	fmt.Println("  3. 跨服务调用没有配置 propagator")
	fmt.Println()
}
