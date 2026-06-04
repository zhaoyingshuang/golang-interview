---
title: gRPC 深度解析
---

## 1. Protobuf 编码原理

Protobuf (Protocol Buffers) 使用二进制编码，核心是变长整数 (Varint) + 字段编号，比 JSON 更紧凑、解析更快。

**字段编码格式**: `(field_number << 3) | wire_type`

Wire Type 含义：
- 0 = Varint（int32/int64/uint32/uint64/bool/enum）
- 1 = 64-bit（fixed64/sfixed64/double）
- 2 = Length-delimited（string/bytes/嵌套消息/packed repeated）
- 5 = 32-bit（fixed32/sfixed32/float）

::: tip 为什么比 JSON 快
1. 二进制格式，无需文本解析（无引号/逗号/括号）
2. 字段用编号而非名称（`field_number=1` vs `"username"`）
3. 生成的代码直接操作内存，无反射开销
4. Varint 编码对小数值只占 1 字节

```protobuf
syntax = "proto3";
package api;

service UserService {
  rpc GetUser(GetUserReq) returns (GetUserResp);
  rpc ListUsers(ListReq) returns (stream User);
}

message GetUserReq {
  int64 id = 1;    // 字段编号1, 编码: (1<<3)|0 = 0x08
}

message GetUserResp {
  string name = 1;
  int32  age  = 2;
}
```

## 2. 四种通信模式

| 模式 | 请求/响应 | 适用场景 |
|------|-----------|----------|
| Unary | 一请求一响应 | GetUser, CreateOrder |
| Server Streaming | 一请求，流式多响应 | 大数据查询、实时日志 |
| Client Streaming | 流式多请求，一响应 | 文件上传、批量导入 |
| Bidirectional | 双方独立流式 | 聊天、实时协作 |

::: info Unary 示例
```go
resp, err := client.GetUser(ctx, &pb.GetUserReq{Id: 1})
```

```go
// Server Streaming
stream, _ := client.ListUsers(ctx, &pb.ListReq{})
for {
    resp, err := stream.Recv()
    if err == io.EOF { break }
    // 处理 resp
}
```

## 3. 拦截器机制

gRPC 拦截器类似 Gin 中间件，分为 UnaryInterceptor 和 StreamInterceptor。

::: warning 面试高频
拦截器链的执行顺序是注册顺序，`handler()` 之后的代码在响应返回时执行（洋葱模型）。

```go
func LoggingInterceptor(ctx context.Context, req any,
    info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
    start := time.Now()
    resp, err := handler(ctx, req) // 调用实际方法
    log.Printf("%s %v err=%v", info.FullMethod, time.Since(start), err)
    return resp, err
}

// 注册拦截器链
grpc.ChainUnaryInterceptor(
    RecoveryInterceptor,
    LoggingInterceptor,
    AuthInterceptor,
)
```

## 4. 负载均衡与连接管理

gRPC 客户端负载均衡通过 Resolver + Balancer 实现：
- Resolver：服务发现（DNS/自定义）
- Balancer：选择实例（round_robin/custom）

```go
conn, _ := grpc.Dial(
    "dns:///my-service:8080",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

::: tip Keepalive 配置
防止连接被中间设备（LB/防火墙）静默断开：
```go
keepalive.ServerParameters{
    MaxConnectionIdle:     15 * time.Minute,
    MaxConnectionAge:      30 * time.Minute,
    Time:                  5 * time.Second,  // ping 间隔
    Timeout:               1 * time.Second,   // ping 超时
}
```

## 5. gRPC vs REST

| 维度 | gRPC | REST |
|------|------|------|
| 协议 | HTTP/2 二进制 | HTTP/1.1 文本 |
| 数据格式 | Protobuf | JSON/XML |
| 流式支持 | 四种模式 | SSE/WebSocket |
| 浏览器 | gRPC-Web (有限) | 原生 |
| 调试 | grpcurl | curl |
| 适用场景 | 微服务间通信 | 对外 API |

::: warning 面试追问
什么时候用 gRPC？→ 内部微服务间通信、性能敏感场景、需要流式通信。对外 API 和前端调用仍推荐 REST。
:::
