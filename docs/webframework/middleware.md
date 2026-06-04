---
title: 中间件机制
---

## 1. 中间件原理

Go Web 框架的中间件本质是**函数链（Function Chain）**，采用洋葱模型（Onion Model）：请求从外层中间件逐层进入核心 Handler，响应从核心逐层返回。每个中间件可以在调用 `c.Next()` 前做预处理（如认证、日志），在 `c.Next()` 后做后处理（如记录耗时、写入响应头）。

核心机制依赖于 `context.Context` 的传递。Gin 的 `gin.Context` 封装了请求/响应读写器和中间件执行索引，通过递增索引实现函数链的顺序执行。

::: info 洋葱模型执行顺序
请求 → MiddlewareA(前) → MiddlewareB(前) → Handler → MiddlewareB(后) → MiddlewareA(后) → 响应

```go
// 中间件本质是一个返回 HandlerFunc 的函数（或直接是 HandlerFunc）
// 简化版洋葱模型实现
type HandlerFunc func(*Context)

type Context struct {
    handlers []HandlerFunc   // 中间件 + 业务 Handler 链
    index    int             // 当前执行到第几个
}

func (c *Context) Next() {
    c.index++
    for c.index < len(c.handlers) {
        c.handlers[c.index](c)
        c.index++
    }
}

// 自定义中间件示例
func TimingMiddleware() HandlerFunc {
    return func(c *Context) {
        start := time.Now()           // 前置处理
        c.Next()                       // 调用后续中间件/Handler
        duration := time.Since(start)  // 后置处理
        log.Printf("request took %v", duration)
    }
}
```

## 2. Gin 中间件实现

Gin 中间件的类型签名与普通 Handler 完全一致：`func(*gin.Context)`。区别在于中间件会调用 `c.Next()` 将控制权传递给链中的下一个函数。`c.Abort()` 可终止链的执行（用于认证失败等场景）。

::: tip 生产建议
中间件应保持无状态，所有请求相关数据通过 `c.Set`/`c.Get` 传递。避免在中间件中定义包级变量存储请求数据，否则会导致并发安全问题。

```go
package main

import (
    "log"
    "time"
    "github.com/gin-gonic/gin"
)

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{
                "code":    40100,
                "message": "missing token",
            })
            return // 必须 return，防止继续执行后续代码
        }

        // 解析 token，将用户信息存入 context
        userID, err := parseToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{
                "code":    40101,
                "message": "invalid token",
            })
            return
        }

        c.Set("userID", userID) // 向下游传递数据
        c.Next()                // 继续执行后续中间件/Handler
    }
}

// 请求耗时中间件（演示 c.Next() 前后处理）
func LatencyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next() // 执行后续处理

        latency := time.Since(start)
        status := c.Writer.Status()
        log.Printf("[%d] %s %v", status, path, latency)
    }
}
```

## 3. 常用中间件实战

生产环境中常用的中间件包括：Recovery（panic 恢复）、Logger（请求日志）、CORS（跨域）、Auth（JWT 认证）、RequestID（请求追踪）、Timeout（超时控制）。

::: warning 面试追问
Recovery 中间件为什么必须放在第一个？→ 因为它用 `recover()` 捕获后续所有中间件和 Handler 的 panic。如果放在其他中间件之后，前面的中间件 panic 无法被捕获。

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// Recovery 中间件（Gin 内置，原理如下）
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic recovered: %v", err)
                c.AbortWithStatusJSON(500, gin.H{"error": "internal error"})
            }
        }()
        c.Next()
    }
}

// CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}

// RequestID 中间件
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        id := uuid.New().String()
        c.Set("requestID", id)
        c.Header("X-Request-ID", id)
        c.Next()
    }
}

// Timeout 中间件
func TimeoutMiddleware(duration time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
        defer cancel()

        c.Request = c.Request.WithContext(ctx)

        done := make(chan struct{})
        go func() {
            c.Next()
            close(done)
        }()

        select {
        case <-done:
            return
        case <-ctx.Done():
            c.AbortWithStatusJSON(504, gin.H{"error": "request timeout"})
        }
    }
}

func main() {
    r := gin.New()
    r.Use(RecoveryMiddleware())       // 1. panic 恢复（必须第一）
    r.Use(RequestIDMiddleware())      // 2. 请求ID
    r.Use(LatencyMiddleware())        // 3. 日志
    r.Use(CORSMiddleware())           // 4. 跨域
    r.Use(AuthMiddleware())           // 5. 认证

    r.Run(":8080")
}
```

## 4. 中间件执行顺序

Gin 支持三个层级的中间件注册：全局（`r.Use`）、路由组（`group.Use`）、单路由（`r.GET(path, mw, handler)`）。执行顺序按注册顺序组成一条链，依次执行。

::: info 执行顺序规则
全局中间件 → 路由组中间件（按嵌套层级从外到内）→ 单路由中间件 → 业务 Handler。`c.Next()` 调用后进入下一个中间件，最后一个执行完后回到前一个的 `c.Next()` 之后。

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()

    // 全局中间件
    r.Use(func(c *gin.Context) {
        fmt.Println("[1] 全局中间件 - 前")
        c.Next()
        fmt.Println("[1] 全局中间件 - 后")
    })

    api := r.Group("/api")
    api.Use(func(c *gin.Context) {
        fmt.Println("[2] 路由组中间件 - 前")
        c.Next()
        fmt.Println("[2] 路由组中间件 - 后")
    })

    // 单路由中间件
    api.GET("/users",
        func(c *gin.Context) {
            fmt.Println("[3] 单路由中间件 - 前")
            c.Next()
            fmt.Println("[3] 单路由中间件 - 后")
        },
        func(c *gin.Context) {
            fmt.Println("[4] 业务 Handler")
            c.JSON(200, gin.H{"users": []string{"alice", "bob"}})
        },
    )

    r.Run(":8080")
}

// 请求 GET /api/users 的输出:
// [1] 全局中间件 - 前
// [2] 路由组中间件 - 前
// [3] 单路由中间件 - 前
// [4] 业务 Handler
// [3] 单路由中间件 - 后
// [2] 路由组中间件 - 后
// [1] 全局中间件 - 后
```

## 5. 面试常见问题

**Q1: 中间件之间如何传递数据？** 通过 `c.Set(key, value)` 和 `c.Get(key)` — 底层是 `sync.RWMutex` 保护的 `map[string]any`。类型安全的取值用 `c.MustGet` 或断言。

**Q2: 如何跳过后续中间件？** `c.Abort()` 设置 `index = abortIndex`（一个足够大的值），`c.Next()` 不再调用后续函数。注意 `c.Abort()` 不会 return，必须手动 return。

**Q3: gin.Context 可以复用吗？** 可以。Gin 使用 `sync.Pool` 复用 Context 对象。因此**绝不能在 goroutine 中持有 `*gin.Context` 引用**，必须在启动 goroutine 前用 `c.Copy()` 创建副本。

::: warning Context 泄漏
在异步场景（如将任务投递到消息队列）中使用 `c.Copy()`。直接传递 `*gin.Context` 到 goroutine 会导致数据竞争。

```go
// Q1: 中间件数据传递
func SetUserMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("userID", 42)
        c.Set("role", "admin")
        c.Next()
    }
}

// Handler 中取值（注意类型断言）
func handler(c *gin.Context) {
    // 方式一: Get + 断言
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(500, gin.H{"error": "userID not found"})
        return
    }
    uid := userID.(int) // 类型断言

    // 方式二: 类型安全的便捷方法
    uid2, _ := c.Get("userID")

    c.JSON(200, gin.H{"user_id": uid, "uid2": uid2})
}

// Q3: Context 安全传递
func asyncHandler(c *gin.Context) {
    // 错误: 直接传 c 到 goroutine
    // go processAsync(c) // 数据竞争!

    // 正确: 使用 c.Copy()
    cCopy := c.Copy()
    go func() {
        // 安全使用 cCopy
        _ = cCopy.Request.URL.Path
    }()

    c.JSON(202, gin.H{"status": "processing"})
}
```
