---
title: TCP 协议
---

## 1. 三次握手与四次挥手

**三次握手**：
1. Client → SYN → Server
2. Server → SYN+ACK → Client
3. Client → ACK → Server

::: tip 为什么需要三次
防止已失效的连接请求到达服务器。两次握手无法确认客户端收到 SYN+ACK。

**四次挥手**：
1. Client → FIN → Server
2. Server → ACK → Client
3. Server → FIN → Client
4. Client → ACK → Server

为什么需要四次？因为 TCP 全双工，每个方向需要单独关闭。Server 收到 FIN 后可能还有数据要发。
:::

## 2. 可靠传输机制

- **序号/确认号**：保证有序、不丢
- **超时重传**：RTO 动态计算（基于 RTT）
- **快速重传**：收到 3 个重复 ACK → 立即重传（不等超时）
- **选择性确认 (SACK)**：告知发送方哪些数据已收到，只重传缺失部分

## 3. 流量控制与拥塞控制

**流量控制**：滑动窗口，接收方通过窗口大小告知发送方能接受的数据量。

**拥塞控制**：慢启动（指数增长）→ 拥塞避免（线性增长）→ 快速重传 → 快速恢复。BBR 算法基于带宽和 RTT 而非丢包。

## 4. TCP Go 编程

```go
// 服务端
listener, _ := net.Listen("tcp", ":8080")
for {
    conn, _ := listener.Accept()
    go handleConn(conn)
}

// TCP_NODELAY — 禁用 Nagle 算法，减少延迟
tcpConn, _ := net.DialTCP("tcp", nil, tcpAddr)
tcpConn.SetNoDelay(true)

// SO_REUSEPORT — 多进程监听同一端口
lc := net.ListenConfig{Control: func(network, address string, c syscall.RawConn) error {
    return c.Control(func(fd uintptr) { syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1) })
}}
```

::: warning 面试高频
TIME_WAIT 过多怎么办？→ 增大 `tcp_max_tw_buckets`、开启 `tw_reuse`、使用长连接。
粘包问题？→ TCP 是字节流协议，需要应用层定义消息边界（长度前缀/分隔符）。
:::
