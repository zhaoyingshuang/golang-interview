package main

import (
	"fmt"
	"strings"
)

// ============================================================
// gRPC 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. gRPC 和 REST 的区别？各自适用场景？
// 2. Protobuf 编码原理？为什么比 JSON 快？
// 3. gRPC 的四种通信模式？
// 4. gRPC 拦截器原理？如何实现中间件？
// 5. HTTP/2 多路复用在 gRPC 中的作用？

func main() {
	protobufEncoding()
	communicationPatterns()
	interceptorPattern()
	loadBalancing()
	grpcVsRest()
}

// ----------------------------------------------------------
// 1. Protobuf 编码原理
// ----------------------------------------------------------
// Protobuf 使用二进制编码，核心概念:
//
//   字段编码 = (field_number << 3) | wire_type
//   wire_type:
//     0 = Varint      (int32, int64, uint32, uint64, sint32, sint64, bool, enum)
//     1 = 64-bit      (fixed64, sfixed64, double)
//     2 = Length-delimited (string, bytes, embedded messages, packed repeated)
//     5 = 32-bit      (fixed32, sfixed32, float)
//
//   Varint 编码: 每个字节的最高位(MSB)表示是否还有后续字节
//     1 → 0x01 (1 byte)
//     300 → 0xAC 0x02 (2 bytes: 10101100 00000010)
//
//   sint32/sint64 使用 ZigZag 编码处理负数:
//     0 → 0, -1 → 1, 1 → 2, -2 → 3, ...
//     公式: (n << 1) ^ (n >> 31)
//
// 为什么比 JSON 快:
//   1. 二进制，无需文本解析
//   2. 字段用编号而非名称，更紧凑
//   3. 不需要引号、逗号、括号
//   4. 生成的代码直接操作内存，无反射
func protobufEncoding() {
	fmt.Println("=== 1. Protobuf 编码原理 ===")

	// 模拟 Varint 编码
	encodeVarint := func(n uint64) []byte {
		var buf []byte
		for n > 0x7f {
			buf = append(buf, byte(n)&0x7f|0x80)
			n >>= 7
		}
		buf = append(buf, byte(n))
		return buf
	}

	// 模拟 ZigZag 编码
	zigzag := func(n int64) uint64 {
		return uint64((n << 1) ^ (n >> 63))
	}

	fmt.Println("Varint 编码示例:")
	for _, n := range []uint64{1, 127, 128, 300, 150} {
		encoded := encodeVarint(n)
		fmt.Printf("  %d → %v (%d bytes)\n", n, encoded, len(encoded))
	}

	fmt.Println("\nZigZag 编码示例:")
	for _, n := range []int64{0, -1, 1, -2, 2} {
		fmt.Printf("  %d → %d\n", n, zigzag(n))
	}

	fmt.Println("\nProtobuf 示例 .proto:")
	fmt.Println("  syntax = \"proto3\";")
	fmt.Println("  package api;")
	fmt.Println("  service UserService {")
	fmt.Println("    rpc GetUser(GetUserReq) returns (GetUserResp);")
	fmt.Println("  }")
	fmt.Println("  message GetUserReq { int64 id = 1; }")
	fmt.Println("  message GetUserResp { string name = 1; int32 age = 2; }")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. gRPC 四种通信模式
// ----------------------------------------------------------
func communicationPatterns() {
	fmt.Println("=== 2. gRPC 四种通信模式 ===")

	patterns := []struct {
		Name      string
		Desc      string
		UseCase   string
	}{
		{
			"Unary RPC",
			"一请求一响应（类似 HTTP）",
			"GetUser, CreateOrder",
		},
		{
			"Server Streaming",
			"一请求，服务端流式返回多个响应",
			"大量数据查询、实时日志、股票行情",
		},
		{
			"Client Streaming",
			"客户端流式发送，服务端一个响应",
			"文件上传、批量数据导入",
		},
		{
			"Bidirectional Streaming",
			"双方都可以独立流式发送",
			"聊天、实时协作",
		},
	}

	for _, p := range patterns {
		fmt.Printf("  %s:\n", p.Name)
		fmt.Printf("    原理: %s\n", p.Desc)
		fmt.Printf("    场景: %s\n\n", p.UseCase)
	}

	fmt.Println("Go 代码示例:")
	fmt.Println("  // Unary")
	fmt.Println("  resp, err := client.GetUser(ctx, &pb.GetUserReq{Id: 1})")
	fmt.Println()
	fmt.Println("  // Server Streaming")
	fmt.Println("  stream, _ := client.ListUsers(ctx, &pb.ListReq{})")
	fmt.Println("  for { resp, err := stream.Recv(); if err == io.EOF { break } }")
	fmt.Println()
	fmt.Println("  // Client Streaming")
	fmt.Println("  stream, _ := client.UploadFile(ctx)")
	fmt.Println("  stream.Send(&pb.Chunk{Data: chunk})")
	fmt.Println("  resp, _ := stream.CloseAndRecv()")
	fmt.Println()
	fmt.Println("  // Bidirectional Streaming")
	fmt.Println("  stream, _ := client.Chat(ctx)")
	fmt.Println("  go func() { stream.Send(msg); stream.CloseSend() }()")
	fmt.Println("  for { resp, err := stream.Recv() }")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 拦截器机制
// ----------------------------------------------------------
// gRPC 拦截器分两种:
//   - UnaryInterceptor: 拦截 Unary RPC
//   - StreamInterceptor: 拦截 Streaming RPC
//
// 原理类似 Gin 中间件的洋葱模型
func interceptorPattern() {
	fmt.Println("=== 3. 拦截器机制 ===")

	fmt.Println("Unary 拦截器:")
	fmt.Println("  func UnaryInterceptor(")
	fmt.Println("    ctx context.Context,")
	fmt.Println("    req any,")
	fmt.Println("    info *grpc.UnaryServerInfo,")
	fmt.Println("    handler grpc.UnaryHandler,")
	fmt.Println("  ) (any, error) {")
	fmt.Println("    // Before: 日志、认证、限流")
	fmt.Println("    start := time.Now()")
	fmt.Println("    resp, err := handler(ctx, req) // 调用实际方法")
	fmt.Println("    log.Printf(\"%s %v\", info.FullMethod, time.Since(start))")
	fmt.Println("    return resp, err")
	fmt.Println("  }")
	fmt.Println()

	fmt.Println("链式拦截器:")
	fmt.Println("  // grpc.ChainUnaryInterceptor(")
	fmt.Println("  //   RecoveryInterceptor,")
	fmt.Println("  //   LoggingInterceptor,")
	fmt.Println("  //   AuthInterceptor,")
	fmt.Println("  // )")
	fmt.Println("  // 执行顺序: Recovery → Logging → Auth → Handler")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 负载均衡与连接管理
// ----------------------------------------------------------
func loadBalancing() {
	fmt.Println("=== 4. 负载均衡与连接管理 ===")

	fmt.Println("gRPC 负载均衡方案:")
	fmt.Println()
	fmt.Println("  1. 客户端负载均衡 (常用):")
	fmt.Println("     resolver → 发现服务实例")
	fmt.Println("     balancer → 选择实例 (round_robin/custom)")
	fmt.Println()
	fmt.Println("     resolver, _ := resolver.NewBuilder()")
	fmt.Println("     conn, _ := grpc.Dial(")
	fmt.Println("       \"dns:///my-service:8080\",")
	fmt.Println("       grpc.WithDefaultServiceConfig(")
	fmt.Println("         `{\"loadBalancingPolicy\":\"round_robin\"}`),")
	fmt.Println("       `),")
	fmt.Println("     )")
	fmt.Println()
	fmt.Println("  2. 代理负载均衡 (Proxy/LB):")
	fmt.Println("     客户端 → LB (Nginx/Envoy) → 后端")
	fmt.Println("     注意: 需要 keepalive 防止连接断开")
	fmt.Println()
	fmt.Println("Keepalive 配置:")
	fmt.Println("  keepalive.ServerParameters{")
	fmt.Println("    MaxConnectionIdle:     15 * time.Minute,")
	fmt.Println("    MaxConnectionAge:      30 * time.Minute,")
	fmt.Println("    MaxConnectionAgeGrace: 10 * time.Second,")
	fmt.Println("    Time:                  5 * time.Second,")
	fmt.Println("    Timeout:               1 * time.Second,")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. gRPC vs REST 对比
// ----------------------------------------------------------
func grpcVsRest() {
	fmt.Println("=== 5. gRPC vs REST ===")

	comparisons := []struct {
		Dimension string
		GRPC      string
		REST      string
	}{
		{"协议", "HTTP/2 (二进制)", "HTTP/1.1 (文本)"},
		{"数据格式", "Protobuf", "JSON/XML"},
		{"代码生成", "内建 (protoc)", "需第三方 (Swagger/OpenAPI)"},
		{"流式", "四种模式", "SSE/WebSocket (非标准)"},
		{"性能", "高 (二进制+HTTP/2)", "中 (文本+HTTP/1.1)"},
		{"浏览器支持", "gRPC-Web (有限)", "原生支持"},
		{"调试", "需工具 (grpcurl)", "curl 即可"},
		{"适用场景", "微服务间通信", "对外 API、前端调用"},
	}

	fmt.Printf("  %-12s %-30s %s\n", "维度", "gRPC", "REST")
	fmt.Println("  " + strings.Repeat("─", 70))
	for _, c := range comparisons {
		fmt.Printf("  %-12s %-30s %s\n", c.Dimension, c.GRPC, c.REST)
	}
	fmt.Println()
	fmt.Println("  面试追问: 什么时候用 gRPC？")
	fmt.Println("  答: 内部微服务间通信、对性能敏感的场景、需要流式通信的场景")
	fmt.Println("  对外 API 和前端调用仍然推荐 REST")
	fmt.Println()
}
