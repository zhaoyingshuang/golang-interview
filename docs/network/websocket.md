---
title: WebSocket 协议
---

## 1. WebSocket 握手

```
客户端: GET /chat HTTP/1.1
        Upgrade: websocket
        Connection: Upgrade
        Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==

服务端: HTTP/1.1 101 Switching Protocols
        Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

Accept 计算：`base64(sha1(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))`

## 2. 帧协议

- **Opcode**：0x0 Continuation / 0x1 Text / 0x2 Binary / 0x8 Close / 0x9 Ping / 0xA Pong
- **Payload Length**：0-125 直接 / 126 后续 2 字节 / 127 后续 8 字节
- **Mask**：客户端→服务端必须 masking（XOR），防止缓存投毒

## 3. Go WebSocket 实战

```go
import "github.com/coder/websocket"

// 服务端
func handleWS(w http.ResponseWriter, r *http.Request) {
    conn, _ := websocket.Accept(w, r, &websocket.AcceptOptions{
        OriginPatterns: []string{"*"},
    })
    defer conn.Close(websocket.StatusNormalClosure, "")

    for {
        _, msg, err := conn.Read(r.Context())
        if err != nil { break }
        conn.Write(r.Context(), websocket.MessageText, response)
    }
}
```

::: warning 并发写保护
WebSocket 连接不是线程安全的！多 goroutine 同时写会损坏帧。用 `sync.Mutex` 保护写入。
:::

## 4. 应用模式

- **心跳保活**：定期 Ping，20-30 秒间隔
- **重连策略**：指数退避，最大 30 秒
- **Hub 模式**：聊天室用 Hub 管理所有 Client 的连接和消息广播

## 5. WebSocket vs SSE vs 长轮询

| 维度 | WebSocket | SSE | 长轮询 |
|------|-----------|-----|--------|
| 方向 | 双向 | 服务端→客户端 | 服务端→客户端 |
| 数据 | 文本/二进制 | 文本 | 任意 |
| 适用 | 聊天/游戏 | 通知/股票 | 兼容方案 |

::: tip 面试追问
WebSocket 最大连接数？受文件描述符限制（`ulimit -n`）、内存、CPU 约束，单机通常 10K-100K。
:::
