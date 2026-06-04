package main

import "fmt"

// ============================================================
// TCP 协议
// ============================================================
// 【面试高频】三次握手/四次挥手/滑动窗口/拥塞控制

func main() {
	threeWayHandshake()
	fourWayWave()
	slidingWindow()
	congestionControl()
	interviewQuestions()
}

func threeWayHandshake() {
	fmt.Println("=== 1. 三次握手 ===")
	fmt.Println()
	fmt.Println("客户端         服务端")
	fmt.Println("  |-- SYN(seq=x) -------->|  CLOSED → SYN_SENT")
	fmt.Println("  |                       |  LISTEN → SYN_RCVD")
	fmt.Println("  |<-- SYN+ACK(seq=y,ack=x+1) --|")
	fmt.Println("  |-- ACK(ack=y+1) ------>|  ESTABLISHED")
	fmt.Println("  |                       |  ESTABLISHED")
	fmt.Println()
	fmt.Println("为什么需要三次?")
	fmt.Println("  1. 确认双方的发送和接收能力都正常")
	fmt.Println("  2. 同步初始序列号(ISN, 随机生成防攻击)")
	fmt.Println("  3. 防止历史重复连接初始化(废弃的SYN到达服务端)")
	fmt.Println()
	fmt.Println("SYN Flood 攻击:")
	fmt.Println("  攻击者发大量SYN不完成第三次握手 → 服务端半连接队列满")
	fmt.Println("  防御: SYN Cookie / 限制半连接数 / 防火墙")
	fmt.Println()
}

func fourWayWave() {
	fmt.Println("=== 2. 四次挥手 ===")
	fmt.Println()
	fmt.Println("主动关闭方           被动关闭方")
	fmt.Println("  |-- FIN(seq=u) ---------->|  ESTABLISHED → CLOSE_WAIT")
	fmt.Println("  |                         |")
	fmt.Println("  |<-- ACK(ack=u+1) --------|")
	fmt.Println("  |  FIN_WAIT_1 → FIN_WAIT_2")
	fmt.Println("  |<-- FIN(seq=w) ----------|  LAST_ACK")
	fmt.Println("  |-- ACK(ack=w+1) -------->|  CLOSED")
	fmt.Println("  |  TIME_WAIT (2MSL)")
	fmt.Println("  |  → CLOSED")
	fmt.Println()
	fmt.Println("为什么需要四次?")
	fmt.Println("  TCP全双工, 每个方向需要单独关闭(FIN+ACK)")
	fmt.Println("  被动方可能还有数据要发, FIN和ACK不能合并")
	fmt.Println()
	fmt.Println("TIME_WAIT (2MSL):")
	fmt.Println("  1. 确保最后一个ACK到达(对方会重发FIN)")
	fmt.Println("  2. 等待网络中残留报文消失")
	fmt.Println()
	fmt.Println("TIME_WAIT 过多怎么办?")
	fmt.Println("  1. tcp_tw_reuse=1 (允许复用TIME_WAIT连接)")
	fmt.Println("  2. tcp_max_tw_buckets 调大")
	fmt.Println("  3. 长连接(KeepAlive)")
	fmt.Println("  4. SO_REUSEADDR")
	fmt.Println()
}

func slidingWindow() {
	fmt.Println("=== 3. 滑动窗口与流量控制 ===")
	fmt.Println()
	fmt.Println("滑动窗口: 控制发送方发送速率, 防止接收方处理不过来")
	fmt.Println("  接收方在ACK中携带 window size (rwnd)")
	fmt.Println("  发送方已发送未确认数据 <= rwnd")
	fmt.Println()
	fmt.Println("零窗口探测:")
	fmt.Println("  接收方窗口=0时, 发送方定期发1字节探测包(ZWP)")
	fmt.Println("  防止窗口更新通知丢失导致死锁")
	fmt.Println()
	fmt.Println("糊涂窗口综合征(Silly Window Syndrome):")
	fmt.Println("  接收方缓冲区快满时, 只通告小窗口")
	fmt.Println("  发送方发小包 → 效率低")
	fmt.Println("  解决: Clark方案, 窗口<最小值时不通告")
	fmt.Println()
}

func congestionControl() {
	fmt.Println("=== 4. 拥塞控制 ===")
	fmt.Println()
	fmt.Println("四个阶段:")
	fmt.Println("  慢启动:    cwnd从1开始, 每RTT翻倍(指数增长)")
	fmt.Println("  拥塞避免:  到达ssthresh后, 每RTT加1(线性增长)")
	fmt.Println("  快重传:    收到3个重复ACK立即重传(不等超时)")
	fmt.Println("  快恢复:    ssthresh=cwnd/2, cwnd=ssthresh(不回1)")
	fmt.Println()
	fmt.Println("BBR (Google, Linux 4.9+):")
	fmt.Println("  不基于丢包, 基于带宽和RTT估计")
	fmt.Println("  适合高延迟/有丢包的网络(如跨国链路)")
	fmt.Println()
}

func interviewQuestions() {
	fmt.Println("=== 5. 面试高频问题 ===")
	fmt.Println()
	fmt.Println("Q: TCP粘包问题?")
	fmt.Println("  TCP是字节流协议, 不保留消息边界")
	fmt.Println("  解决: 固定长度/分隔符/长度前缀(最常用)")
	fmt.Println()
	fmt.Println("Q: TCP vs UDP?")
	fmt.Println("  TCP: 可靠/有序/流量控制/连接 → 文件传输/HTTP/WebSocket")
	fmt.Println("  UDP: 无连接/不保证可靠 → DNS/视频流/quic")
	fmt.Println()
	fmt.Println("Q: 为什么握手是三次, 挥手是四次?")
	fmt.Println("  握手: SYN+ACK可以合并(服务端收到SYN立即响应)")
	fmt.Println("  挥手: FIN和ACK不能合并(被动方可能还有数据)")
	fmt.Println()
}
