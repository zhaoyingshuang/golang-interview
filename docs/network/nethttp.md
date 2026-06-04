---
title: Go net/http 源码解析
---

## 1. Server 核心架构

**ListenAndServe 流程**：
1. `net.Listen("tcp", addr)` — 监听端口
2. `for { Accept() }` — 循环接受连接
3. `go c.serve(ctx)` — 每个连接一个 goroutine

::: tip goroutine-per-conn 模型
Go 不用线程池，因为 goroutine 很轻量（初始栈 2KB），GMP 调度器高效管理百万 goroutine。

请求处理流程：Accept → goroutine → readRequest → ServeMux → Handler → Response
:::

## 2. ServeMux 路由匹配

- 精确匹配优先：`/api/health`
- 最长前缀匹配：`/api/users/` 匹配所有子路径
- `/` 是兜底

```go
// Go 1.22 增强
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
})
```

## 3. Handler 接口

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// HandlerFunc 适配器
type HandlerFunc func(ResponseWriter, *Request)
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
```

## 4. Request/Response

- `r.Body` 只能读一次！中间件读取后需用 `io.NopCloser` 恢复
- `ResponseWriter` 接口：`Header()` / `Write()` / `WriteHeader()`
- `Hijack()` 接口用于 WebSocket 升级

## 5. 优雅关闭

```go
srv := &http.Server{Addr: ":8080"}
go srv.ListenAndServe()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx) // 关闭监听 → 等活跃请求完成 → 超时强制关闭
```
