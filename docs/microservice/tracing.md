---
title: 分布式链路追踪
---

## 1. OpenTelemetry 核心概念

- **Trace**：一次完整请求链路，唯一 TraceID
- **Span**：链路中的一个操作，唯一 SpanID，包含 ParentSpanID 形成树结构
- **Context**：在服务间传播的追踪上下文

```
Trace (TraceID: abc123)
├── Span: HTTP GET /api/users (SpanID: s1)
│   ├── Span: DB SELECT users (SpanID: s2, ParentID: s1)
│   └── Span: Redis GET cache (SpanID: s3, ParentID: s1)
└── Span: HTTP Response 200 (SpanID: s4)
```

::: tip Span 包含的数据
TraceID / SpanID / ParentSpanID、操作名、开始/结束时间、Attributes（属性）、Events（事件）、Status（OK/Error）、Links（关联 Span）
:::

## 2. Context 传播

W3C Trace Context 标准（`traceparent` 头）：

```
traceparent: 00-<trace-id-32hex>-<span-id-16hex>-<flags-2hex>
例: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

```go
// 注入（客户端发送）
propagation := otel.GetTextMapPropagator()
propagation.Inject(ctx, propagation.HeaderCarrier(req.Header))

// 提取（服务端接收）
ctx = propagation.Extract(ctx, propagation.HeaderCarrier(r.Header))
```

::: warning Context 传播丢失的常见原因
1. goroutine 用了 `context.Background()` 而非传入 ctx
2. 消息队列序列化时没保留 trace header
3. 跨服务调用没配置 propagator
:::

## 3. Go 接入 OpenTelemetry

```go
// 初始化 TracerProvider
exporter, _ := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("localhost:4317"),
)
tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
    sdktrace.WithResource(resource.NewWithAttributes(
        semconv.SchemaURL,
        attribute.String("service.name", "user-service"),
    )),
)
otel.SetTracerProvider(tp)

// gRPC 拦截器
grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor())

// HTTP 中间件
handler := otelhttp.NewHandler(http.HandlerFunc(handle), "api/users")
```

## 4. 采样策略

| 策略 | 原理 | 优点 | 缺点 |
|------|------|------|------|
| 概率采样 | 固定比例 | 简单 | 可能错过异常 |
| 头部采样 | 链路开始时决定 | 一致性好 | 无法按结果采样 |
| 尾部采样 | 链路结束后决定 | 不漏关键链路 | 需缓存所有 Span |

::: info 推荐
头部采样 `ParentBased(RateLimiting(100/s))` + 尾部采样 `错误优先` 组合使用。
:::

## 5. 异步任务链路传递

```go
// goroutine 传递 ctx
ctx, span := tracer.Start(ctx, "async-task")
go func(ctx context.Context) {
    defer span.End()
    _, childSpan := tracer.Start(ctx, "sub-task")
    defer childSpan.End()
}(ctx)

// 消息队列传播
// 生产者: propagator.Inject(ctx, &msg.Headers)
// 消费者: ctx = propagator.Extract(context.Background(), &msg.Headers)
```
