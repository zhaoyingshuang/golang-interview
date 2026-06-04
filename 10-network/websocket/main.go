package main

import (
	"fmt"
)

// ============================================================
// WebSocket 协议
// ============================================================

func main() {
	handshake()
	frameProtocol()
	goWebSocket()
	patterns()
	compare()
}

// ----------------------------------------------------------
// 1. WebSocket 握手
// ----------------------------------------------------------
func handshake() {
	fmt.Println("=== 1. WebSocket 握手 ===")
	fmt.Println()
	fmt.Println("客户端请求 (HTTP Upgrade):")
	fmt.Println("  GET /chat HTTP/1.1")
	fmt.Println("  Host: example.com")
	fmt.Println("  Upgrade: websocket")
	fmt.Println("  Connection: Upgrade")
	fmt.Println("  Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==")
	fmt.Println("  Sec-WebSocket-Version: 13")
	fmt.Println()
	fmt.Println("服务端响应 (101 Switching Protocols):")
	fmt.Println("  HTTP/1.1 101 Switching Protocols")
	fmt.Println("  Upgrade: websocket")
	fmt.Println("  Connection: Upgrade")
	fmt.Println("  Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=")
	fmt.Println()
	fmt.Println("Accept 计算过程:")
	fmt.Println("  key = \"dGhlIHNhbXBsZSBub25jZQ==\"")
	fmt.Println("  guid = \"258EAFA5-E914-47DA-95CA-C5AB0DC85B11\"")
	fmt.Println("  accept = base64(sha1(key + guid))")
	fmt.Println()
	fmt.Println("Go 计算 Accept:")
	fmt.Println("  import \"github.com/coder/websocket\"")
	fmt.Println("  // 或手动计算:")
	fmt.Println("  h := sha1.Sum([]byte(key + \"258EAFA5-E914-47DA-95CA-C5AB0DC85B11\"))")
	fmt.Println("  accept := base64.StdEncoding.EncodeToString(h[:])")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 帧协议
// ----------------------------------------------------------
func frameProtocol() {
	fmt.Println("=== 2. WebSocket 帧协议 ===")
	fmt.Println()
	fmt.Println("帧格式 (RFC 6455):")
	fmt.Println()
	fmt.Println("   0                   1                   2                   3")
	fmt.Println("   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1")
	fmt.Println("  +-+-+-+-+-------+-+-------------+-------------------------------+")
	fmt.Println("  |F|R|R|R| opcode|M| Payload len |    Extended payload length    |")
	fmt.Println("  |I|S|S|S| (4)   |A|     (7)     |            (16/64)            |")
	fmt.Println("  |N|V|V|V|       |S|             |   (if payload len==126/127)   |")
	fmt.Println("  | |1|2|3|       |K|             |                               |")
	fmt.Println("  +-+-+-+-+-------+-+-------------+-------------------------------+")
	fmt.Println()
	fmt.Println("Opcode:")
	fmt.Println("  0x0 — Continuation (分片帧)")
	fmt.Println("  0x1 — Text (UTF-8 文本)")
	fmt.Println("  0x2 — Binary (二进制数据)")
	fmt.Println("  0x8 — Close (关闭连接)")
	fmt.Println("  0x9 — Ping (心跳)")
	fmt.Println("  0xA — Pong (心跳回复)")
	fmt.Println()
	fmt.Println("Payload Length:")
	fmt.Println("  0-125      → 7 bit 直接表示")
	fmt.Println("  126        → 后续 2 字节 (uint16)")
	fmt.Println("  127        → 后续 8 字节 (uint64)")
	fmt.Println()
	fmt.Println("Mask (客户端→服务端必须 masking):")
	fmt.Println("  masking-key (4 字节) XOR payload")
	fmt.Println("  安全考虑: 防止缓存投毒攻击")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Go WebSocket 实战
// ----------------------------------------------------------
func goWebSocket() {
	fmt.Println("=== 3. Go WebSocket 实战 ===")
	fmt.Println()
	fmt.Println("推荐库: nhooyr.io/websocket (现 github.com/coder/websocket)")
	fmt.Println()
	fmt.Println("服务端:")
	fmt.Println("  func handleWS(w http.ResponseWriter, r *http.Request) {")
	fmt.Println("    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{")
	fmt.Println("      OriginPatterns: []string{\"*\"},  // CORS")
	fmt.Println("    })")
	fmt.Println("    if err != nil { return }")
	fmt.Println("    defer conn.Close(websocket.StatusNormalClosure, \"\")")
	fmt.Println()
	fmt.Println("    for {")
	fmt.Println("      _, msg, err := conn.Read(r.Context())")
	fmt.Println("      if err != nil { break }")
	fmt.Println("      // 处理消息")
	fmt.Println("      conn.Write(r.Context(), websocket.MessageText, response)")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("并发写保护:")
	fmt.Println("  WebSocket 连接不是线程安全的!")
	fmt.Println("  多 goroutine 同时写会损坏帧")
	fmt.Println()
	fmt.Println("  解决方案:")
	fmt.Println("  type SafeConn struct {")
	fmt.Println("    mu   sync.Mutex")
	fmt.Println("    conn *websocket.Conn")
	fmt.Println("  }")
	fmt.Println("  func (c *SafeConn) Write(ctx context.Context, typ MessageType, data []byte) error {")
	fmt.Println("    c.mu.Lock()")
	fmt.Println("    defer c.mu.Unlock()")
	fmt.Println("    return c.conn.Write(ctx, typ, data)")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 实际应用模式
// ----------------------------------------------------------
func patterns() {
	fmt.Println("=== 4. 应用模式 ===")
	fmt.Println()
	fmt.Println("心跳保活:")
	fmt.Println("  // 服务端")
	fmt.Println("  ctx, cancel := context.WithTimeout(ctx, 30*time.Second)")
	fmt.Println("  defer cancel()")
	fmt.Println("  conn.Ping(ctx)  // Go 库自动回复 Pong")
	fmt.Println()
	fmt.Println("  // 客户端")
	fmt.Println("  go func() {")
	fmt.Println("    ticker := time.NewTicker(20 * time.Second)")
	fmt.Println("    for range ticker.C {")
	fmt.Println("      conn.Ping(ctx)")
	fmt.Println("    }")
	fmt.Println("  }()")
	fmt.Println()
	fmt.Println("重连策略:")
	fmt.Println("  func connectWithRetry(url string) (*websocket.Conn, error) {")
	fmt.Println("    backoff := time.Second")
	fmt.Println("    for {")
	fmt.Println("      conn, _, err := websocket.Dial(ctx, url, nil)")
	fmt.Println("      if err == nil { return conn, nil }")
	fmt.Println("      time.Sleep(backoff)")
	fmt.Println("      backoff = min(backoff*2, 30*time.Second) // 指数退避")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("Hub 模式 (聊天室):")
	fmt.Println("  type Hub struct {")
	fmt.Println("    clients    map[*Client]bool")
	fmt.Println("    broadcast   chan []byte")
	fmt.Println("    register    chan *Client")
	fmt.Println("    unregister  chan *Client")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. WebSocket vs SSE vs 长轮询
// ----------------------------------------------------------
func compare() {
	fmt.Println("=== 5. 实时通信方案对比 ===")
	fmt.Println()
	fmt.Println("  ┌──────────┬──────────────┬──────────────┬──────────────┐")
	fmt.Println("  │ 维度     │ WebSocket    │ SSE          │ 长轮询       │")
	fmt.Println("  ├──────────┼──────────────┼──────────────┼──────────────┤")
	fmt.Println("  │ 方向     │ 双向         │ 服务端→客户端│ 服务端→客户端│")
	fmt.Println("  │ 协议     │ ws/wss       │ HTTP/1.1     │ HTTP/1.1     │")
	fmt.Println("  │ 数据格式 │ 文本/二进制  │ 文本         │ 任意         │")
	fmt.Println("  │ 连接数   │ 1            │ 1            │ 每次请求新建 │")
	fmt.Println("  │ 浏览器   │ 全部支持     │ 全部支持     │ 全部支持     │")
	fmt.Println("  │ 代理兼容 │ 可能有问题   │ 好           │ 好           │")
	fmt.Println("  │ 适用     │ 聊天/游戏    │ 通知/股票    │ 兼容性方案   │")
	fmt.Println("  └──────────┴──────────────┴──────────────┴──────────────┘")
	fmt.Println()
	fmt.Println("面试追问: WebSocket 最大连接数?")
	fmt.Println("  理论上无上限，实际受限于:")
	fmt.Println("  1. 系统文件描述符限制 (ulimit -n)")
	fmt.Println("  2. 内存 (每个连接约 10KB-100KB)")
	fmt.Println("  3. CPU (消息处理)")
	fmt.Println("  单机通常支持 10K-100K 并发连接")
	fmt.Println()
}
