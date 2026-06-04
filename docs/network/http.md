---
title: HTTP 协议
---

## 1. HTTP/1.1

- **持久连接**：`Connection: keep-alive`，一个 TCP 连接发多个请求
- **管线化**：不等响应就发下一个请求（实际很少用，有队头阻塞）
- **Content-Length / Transfer-Encoding: chunked**：定长用 Content-Length，不定长用分块传输

::: warning 队头阻塞 (Head-of-Line Blocking)
HTTP/1.1 的管线化仍然有队头阻塞：一个慢响应会阻塞后续所有请求。浏览器通常限制每个域名 6 个并发连接来缓解。
:::

## 2. HTTP/2

- **二进制帧**：请求/响应被拆分为二进制帧（HEADERS/DATA 帧）
- **多路复用**：一个 TCP 连接上并行多个流（Stream），无队头阻塞
- **头部压缩 HPACK**：静态表 + 动态表 + 哈夫曼编码
- **服务器推送**：服务端主动推送关联资源（如 CSS/JS）
- **流优先级**：客户端可指定流的权重和依赖关系

## 3. HTTP/3 (QUIC)

基于 **UDP + QUIC** 协议：
- **0-RTT 连接**：首次 1-RTT，后续 0-RTT 恢复
- **解决 TCP 队头阻塞**：每个 Stream 独立可靠传输
- **连接迁移**：用 Connection ID 而非四元组标识连接（WiFi 切蜂窝不断连）
- **内置 TLS 1.3**：握手和加密合二为一

## 4. 状态码详解

| 状态码 | 含义 | 典型使用 |
|--------|------|----------|
| 200 | OK | GET/PUT/PATCH 成功 |
| 201 | Created | POST 创建成功 |
| 204 | No Content | DELETE 成功 |
| 301 | 永久重定向 | URL 永久变更，SEO 权重转移 |
| 302 | 临时重定向 | 临时跳转 |
| 304 | Not Modified | 协商缓存命中 |
| 400 | Bad Request | 参数校验失败 |
| 401 | Unauthorized | 未认证（未登录） |
| 403 | Forbidden | 已认证但无权限 |
| 404 | Not Found | 资源不存在 |
| 429 | Too Many Requests | 限流触发 |
| 500 | Internal Server Error | 服务端错误 |
| 502 | Bad Gateway | 网关上游错误 |
| 503 | Service Unavailable | 服务不可用 |
| 504 | Gateway Timeout | 网关超时 |

::: tip 面试重点：401 vs 403
401 = 不知道你是谁（未登录）。403 = 知道你是谁，但你没权限。
:::
