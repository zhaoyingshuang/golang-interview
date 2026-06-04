package main

import (
	"fmt"
	"net/http"
)

// ============================================================
// Go net/http 源码解析
// ============================================================

func main() {
	serverArchitecture()
	serveMux()
	requestResponse()
	gracefulShutdown()
}

// ----------------------------------------------------------
// 1. Server 核心架构
// ----------------------------------------------------------
func serverArchitecture() {
	fmt.Println("=== 1. Server 核心架构 ===")
	fmt.Println()
	fmt.Println("ListenAndServe 启动流程:")
	fmt.Println("  1. net.Listen(\"tcp\", addr) — 监听端口")
	fmt.Println("  2. for { Accept() } — 循环接受连接")
	fmt.Println("  3. go c.serve(ctx) — 每个连接一个 goroutine")
	fmt.Println()
	fmt.Println("请求处理流程:")
	fmt.Println("  Accept → goroutine → readRequest → ServeMux → Handler → Response")
	fmt.Println()
	fmt.Println("核心模型: goroutine-per-conn")
	fmt.Println("  优点: 简单、自然并发")
	fmt.Println("  缺点: 大量连接时 goroutine 数量多 (但 Go goroutine 很轻量)")
	fmt.Println("  对比: Java NIO 用线程池 + IO 多路复用")
	fmt.Println()
	fmt.Println("面试追问: 为什么 Go 不用线程池?")
	fmt.Println("  答: goroutine 很轻量 (初始栈 2KB)，可以轻松创建百万个")
	fmt.Println("  Go 的调度器 (GMP) 高效管理大量 goroutine")
	fmt.Println("  不需要像 Java 那样用线程池来控制资源")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. ServeMux 路由匹配
// ----------------------------------------------------------
func serveMux() {
	fmt.Println("=== 2. ServeMux 路由匹配 ===")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "users: %s", r.URL.Path)
	})

	fmt.Println("ServeMux 匹配规则:")
	fmt.Println("  1. 精确匹配优先: /api/health")
	fmt.Println("  2. 最长前缀匹配: /api/users/ 匹配所有子路径")
	fmt.Println("  3. / 是兜底: 匹配所有未匹配的路径")
	fmt.Println()
	fmt.Println("Go 1.22 增强的路由:")
	fmt.Println("  mux.HandleFunc(\"GET /users/{id}\", handler)")
	fmt.Println("  - 支持 HTTP 方法")
	fmt.Println("  - 支持路径参数 {id}")
	fmt.Println("  - r.PathValue(\"id\") 获取参数")
	fmt.Println()
	fmt.Println("为什么生产环境不用 ServeMux?")
	fmt.Println("  1. 不支持参数路由 (1.22 之前)")
	fmt.Println("  2. 没有中间件机制")
	fmt.Println("  3. 没有路由分组")
	fmt.Println("  → 使用 Gin/Echo/Chi 等框架")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Request/Response 处理
// ----------------------------------------------------------
func requestResponse() {
	fmt.Println("=== 4. Request/Response 处理 ===")
	fmt.Println()
	fmt.Println("Request 常用操作:")
	fmt.Println("  r.Method       — GET/POST/PUT/DELETE")
	fmt.Println("  r.URL.Path     — 请求路径")
	fmt.Println("  r.URL.Query()  — Query 参数")
	fmt.Println("  r.Header       — 请求头")
	fmt.Println("  r.Body         — 请求体 (io.ReadCloser)")
	fmt.Println("  r.Context()    — 请求上下文")
	fmt.Println("  r.RemoteAddr   — 客户端地址")
	fmt.Println()
	fmt.Println("ResponseWriter 接口:")
	fmt.Println("  type ResponseWriter interface {")
	fmt.Println("    Header() Header       // 写响应头")
	fmt.Println("    Write([]byte) (int, error) // 写响应体")
	fmt.Println("    WriteHeader(int)      // 写状态码")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("注意: r.Body 只能读一次!")
	fmt.Println("  读取后 r.Body 变为 EOF")
	fmt.Println("  中间件需要读取时: 读完后用 io.NopCloser 恢复")
	fmt.Println()
	fmt.Println("Hijack 接口 (WebSocket 用):")
	fmt.Println("  type Hijacker interface {")
	fmt.Println("    Hijack() (net.Conn, *bufio.ReadWriter, error)")
	fmt.Println("  }")
	fmt.Println("  用于接管 HTTP 连接，实现 WebSocket 升级")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 优雅关闭
// ----------------------------------------------------------
// UserHandler 实现 http.Handler
type UserHandler struct{}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "users handler")
}

func gracefulShutdown() {
	fmt.Println("=== 5. 优雅关闭 ===")
	fmt.Println()
	fmt.Println("  srv := &http.Server{Addr: \":8080\"}")
	fmt.Println()
	fmt.Println("  go func() {")
	fmt.Println("    if err := srv.ListenAndServe(); err != http.ErrServerClosed {")
	fmt.Println("      log.Fatal(err)")
	fmt.Println("    }")
	fmt.Println("  }()")
	fmt.Println()
	fmt.Println("  quit := make(chan os.Signal, 1)")
	fmt.Println("  signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)")
	fmt.Println("  <-quit")
	fmt.Println()
	fmt.Println("  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)")
	fmt.Println("  defer cancel()")
	fmt.Println()
	fmt.Println("  if err := srv.Shutdown(ctx); err != nil {")
	fmt.Println("    log.Fatal(\"Server forced to shutdown:\", err)")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("Shutdown 行为:")
	fmt.Println("  1. 关闭监听器 (不再接受新连接)")
	fmt.Println("  2. 等待活跃请求完成 (直到 ctx 超时)")
	fmt.Println("  3. 超时后强制关闭")
	fmt.Println()
}
