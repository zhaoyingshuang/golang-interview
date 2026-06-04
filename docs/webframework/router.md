---
title: Gin/Chi 路由原理
---

## 1. 路由树原理

Go Web 框架的路由匹配核心是高效的数据结构。标准库 `net/http` 的 `ServeMux` 使用简单的前缀匹配，而 Gin 借鉴了 `httprouter` 的**压缩基数树（Compressed Radix Tree）**，将查找复杂度降到 O(k)，其中 k 是路径长度而非路由数量。

**Trie 树**每个节点只存一个字符，路径越深层级越多。**基数树（Radix Tree / Patricia Trie）**将只有一个子节点的路径压缩为一条边，减少节点数量。httprouter 在此基础上进一步优化：将 HTTP 方法各自维护一棵独立的基数树，节点内部保存完整的路径片段，实现一次遍历完成匹配。

::: info 压缩基数树 vs Trie
Trie 树查找路径 `/api/v1/users` 需要逐字符比较，而压缩基数树将 `/api/v1/` 作为单条边存储，匹配时直接比较字符串前缀，减少内存分配和比较次数。

```go
// Gin 内部路由树结构（简化版）
type methodTree struct {
    method string
    root   *node
}

type node struct {
    path      string      // 当前节点存储的路径片段
    indices   string      // 子节点的首字符索引
    children  []*node     // 子节点
    handlers  HandlersChain // 该节点绑定的处理函数链
    priority  uint32      // 优先级（用于平衡）
    nType     nodeType    // 节点类型: static/param/catchAll
    wildChild bool        // 是否有通配子节点
}

// 路由注册后内部树形结构示例
// GET /api/v1/users
// GET /api/v1/users/:id
// GET /api/v1/posts
//
// 压缩存储:
// root → /api/v1/ → users (static)
//                  → users/ → :id (param)
//                  → posts (static)
```

## 2. 路由注册与匹配

Gin 支持三种路由类型：**静态路由**（精确匹配）、**参数路由**（`:param`，匹配单个路径段）、**通配路由**（`*wildcard`，匹配剩余所有路径）。

匹配优先级：静态路由 > 参数路由 > 通配路由。同一层级不允许注册冲突的参数路由（如 `:id` 和 `:name`）。

::: tip 使用场景
参数路由用于资源 ID（`/users/:id`），通配路由用于文件服务（`/files/*filepath`）。RESTful API 设计中参数路由最常用。

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    // 静态路由 — 精确匹配
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 参数路由 — 匹配单个路径段
    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id") // 获取路径参数
        c.JSON(200, gin.H{"user_id": id})
    })

    // 通配路由 — 匹配剩余路径
    r.GET("/files/*filepath", func(c *gin.Context) {
        fp := c.Param("filepath") // 值以 "/" 开头
        c.JSON(200, gin.H{"path": fp})
    })

    // 路由优先级示例
    r.GET("/users/list", listUsers)   // 静态路由优先
    r.GET("/users/:id",  getUser)     // 参数路由次之
    r.GET("/users/*any", fallback)    // 通配路由最后

    r.Run(":8080")
}
```

::: warning 路由冲突
以下注册会 panic：`/users/:id` 和 `/users/:name` 冲突（同层级不能有两个参数路由）。但 `/users/:id` 和 `/users/profile` 不冲突（静态优先）。

## 3. 路由分组

路由分组（RouterGroup）是 Gin 组织大型项目的核心机制。`Group` 返回一个新的 `RouterGroup`，继承父组的路径前缀和中间件，并支持链式调用和嵌套。

::: info 生产建议
实际项目中按业务模块分组：`/api/v1/users`、`/api/v1/orders`。每个模块注册各自的中间件（认证、限流），保持代码结构清晰。

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    // 基础分组：添加公共前缀
    api := r.Group("/api/v1")
    {
        // 公开接口 — 无需认证
        public := api.Group("/public")
        {
            public.GET("/health", healthHandler)
            public.POST("/login", loginHandler)
        }

        // 需要认证的接口
        auth := api.Group("/")
        auth.Use(AuthMiddleware()) // 中间件继承
        {
            users := auth.Group("/users")
            {
                users.GET("",     listUsers)
                users.GET("/:id", getUser)
                users.POST("",    createUser)
            }

            // 嵌套分组 — 订单模块
            orders := auth.Group("/orders")
            orders.Use(RateLimitMiddleware()) // 模块专属中间件
            {
                orders.GET("",     listOrders)
                orders.GET("/:id", getOrder)
            }
        }
    }

    r.Run(":8080")
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Next()
    }
}

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 限流逻辑
        c.Next()
    }
}
```

## 4. 性能对比

不同框架的路由实现性能差异显著。Gin（httprouter 基数树）的查找是 O(k)，Chi（基于标准库的改进）也是线性查找但实现更轻量，标准库 `http.ServeMux`（Go 1.22 前）只支持简单前缀匹配。

::: tip 生产建议
路由性能在大多数 Web 应用中不是瓶颈（单次查找在百纳秒级）。选框架时更应关注：中间件生态、社区活跃度、代码可维护性。Go 1.22 增强了标准库路由（支持方法和路径参数），简单服务可考虑不用第三方框架。

```go
// Go 1.22+ 标准库路由增强（支持方法和路径参数）
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintf(w, "User: %s", id)
})
```

```text
// 基准测试参考（GHz CPU, 100 条路由）
// 框架          ns/op    allocs/op
// Gin           ~650     0
// Chi           ~800     0
// Stdlib 1.22   ~550     0  (路由数少时最优)
// Echo          ~700     0
```

## 5. 常见陷阱

**陷阱1: 路由冲突导致 panic**。Gin 启动时检测路由冲突，直接 panic。生产环境务必在启动阶段就暴露问题，不要用 `recover` 吞掉。

**陷阱2: 尾斜杠不一致**。`/users/` 和 `/users` 是两条不同的路由。客户端请求 `/users` 不会匹配 `/users/`。建议统一策略：中间件重定向或注册时统一。

**陷阱3: 参数路由吞掉静态路由**。注册了 `/users/:id` 后，`GET /users/list` 如果在 `:id` 之后注册，静态路由仍优先；但如果注册顺序颠倒或框架实现不同，可能出现意外行为。

::: warning 面试追问
1. Gin 路由树的查找时间复杂度是多少？→ O(k)，k 是请求路径长度。
2. 为什么 Gin 不支持正则路由？→ 正则匹配需要回溯，无法保证 O(k)。如需正则可在 Handler 内校验。
3. 如何实现路由的热更新？→ Gin 不原生支持，可用 `gin.Engine` 指针替换或使用 `fsnotify` 监听配置变化重建路由。

```go
// 尾斜杠重定向中间件
func TrailingSlashRedirect() gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        if len(path) > 1 && strings.HasSuffix(path, "/") {
            c.Redirect(http.StatusMovedPermanently, strings.TrimRight(path, "/"))
            c.Abort()
            return
        }
        c.Next()
    }
}

// 参数校验 — 在 Handler 内做正则校验
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    if _, err := strconv.ParseInt(id, 10, 64); err != nil {
        c.JSON(400, gin.H{"error": "invalid user id"})
        return
    }
    // 业务逻辑...
})
```
