package main

import "fmt"

// ============================================================
// HTTP 协议演进
// ============================================================

func main() {
	http11()
	http2()
	http3()
	statusCodes()
	cookieSessionToken()
}

func http11() {
	fmt.Println("=== 1. HTTP/1.1 ===")
	fmt.Println()
	fmt.Println("核心改进(相比1.0):")
	fmt.Println("  1. 持久连接(Keep-Alive): 一个TCP连接发多个请求")
	fmt.Println("  2. 管线化(Pipelining): 不等响应就发下一个请求")
	fmt.Println("  3. Host头: 支持虚拟主机")
	fmt.Println("  4. Chunked传输: 不需要Content-Length")
	fmt.Println()
	fmt.Println("队头阻塞(HOL):")
	fmt.Println("  管线化响应必须按序返回, 前一个慢了后面全等")
	fmt.Println("  → 浏览器限制6-8个TCP连接来缓解")
	fmt.Println()
}

func http2() {
	fmt.Println("=== 2. HTTP/2 ===")
	fmt.Println()
	fmt.Println("核心特性:")
	fmt.Println("  1. 多路复用: 一个TCP连接上并行多个流(Stream)")
	fmt.Println("     每个请求/响应是一个流, 帧是最小通信单位")
	fmt.Println("  2. 头部压缩(HPACK): 静态表+动态表+Huffman编码")
	fmt.Println("     常见header(如method:GET)用1字节索引表示")
	fmt.Println("  3. 服务器推送: 服务端主动推资源(如CSS/JS)")
	fmt.Println("  4. 二进制协议: 解析更快, 不再是文本")
	fmt.Println()
	fmt.Println("仍存在TCP层队头阻塞:")
	fmt.Println("  HTTP/2解决了应用层HOL, 但TCP丢包会阻塞所有流")
	fmt.Println("  → 一个包丢失, 所有流都得等重传")
	fmt.Println()
}

func http3() {
	fmt.Println("=== 3. HTTP/3 (QUIC) ===")
	fmt.Println()
	fmt.Println("基于UDP, 解决HTTP/2的TCP层队头阻塞")
	fmt.Println()
	fmt.Println("QUIC 核心改进:")
	fmt.Println("  1. 独立流: 每个流独立可靠传输, 丢包只影响该流")
	fmt.Println("  2. 0-RTT: 首次1-RTT, 后续0-RTT(缓存服务端信息)")
	fmt.Println("  3. 连接迁移: 用Connection ID而非四元组标识连接")
	fmt.Println("     手机WiFi切4G不断连!")
	fmt.Println("  4. 内置TLS 1.3: 加密是协议的一部分")
	fmt.Println()
	fmt.Println("Go 支持: golang.org/x/net/quic (实验性)")
	fmt.Println()
}

func statusCodes() {
	fmt.Println("=== 4. 状态码 ===")
	fmt.Println()
	fmt.Println("2xx 成功: 200 OK, 201 Created, 204 No Content")
	fmt.Println("3xx 重定向: 301 永久, 302 临时, 304 未修改(缓存)")
	fmt.Println("4xx 客户端错误:")
	fmt.Println("  400 Bad Request, 401 Unauthorized, 403 Forbidden")
	fmt.Println("  404 Not Found, 409 Conflict, 429 Too Many Requests")
	fmt.Println("5xx 服务端错误:")
	fmt.Println("  500 Internal, 502 Bad Gateway, 503 Unavailable")
	fmt.Println()
	fmt.Println("缓存相关:")
	fmt.Println("  Cache-Control: max-age/ no-cache/no-store")
	fmt.Println("  ETag: 资源指纹, If-None-Match → 304")
	fmt.Println("  Last-Modified: If-Modified-Since → 304")
	fmt.Println()
}

func cookieSessionToken() {
	fmt.Println("=== 5. Cookie/Session/Token ===")
	fmt.Println()
	fmt.Println("Cookie: 服务端设置, 浏览器自动携带")
	fmt.Println("  缺点: CSRF风险, 大小限制(4KB), 跨域限制")
	fmt.Println()
	fmt.Println("Session: 服务端存储, SessionID通过Cookie传递")
	fmt.Println("  缺点: 服务端内存/Redis, 分布式共享问题")
	fmt.Println()
	fmt.Println("Token(JWT): 无状态, 自包含(Header.Payload.Signature)")
	fmt.Println("  优点: 无服务端存储, 跨服务验证, 移动端友好")
	fmt.Println("  缺点: 无法主动失效(需黑名单), Payload不加密")
	fmt.Println()
	fmt.Println("JWT 安全:")
	fmt.Println("  1. 短过期时间 + Refresh Token")
	fmt.Println("  2. HTTPS传输")
	fmt.Println("  3. 不存敏感信息(Payload可base64解码)")
	fmt.Println()
}
