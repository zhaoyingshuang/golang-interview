---
title: RESTful API 设计
---

## 1. RESTful 规范

REST（Representational State Transfer）的核心思想是：以资源为中心，用 HTTP 方法表达操作语义，用状态码表达结果。资源用名词复数（`/users`、`/orders`），HTTP 方法映射 CRUD：GET 查询、POST 创建、PUT 全量更新、PATCH 部分更新、DELETE 删除。

::: tip 生产建议
严格遵循 RESTful 规范能降低团队沟通成本。但如果业务场景不适合（如批量操作、长耗时任务），不要强行 RESTful，RPC 风格的 POST + action 也是合理选择。一致性比教条更重要。

```text
// RESTful API 设计示例 — 用户管理模块

GET    /api/v1/users          # 获取用户列表（支持分页、过滤）
GET    /api/v1/users/:id      # 获取单个用户
POST   /api/v1/users          # 创建用户
PUT    /api/v1/users/:id      # 全量更新用户
PATCH  /api/v1/users/:id      # 部分更新用户
DELETE /api/v1/users/:id      # 删除用户

// 非标准场景的处理
POST   /api/v1/users/:id/avatar    # 上传头像（非标准资源）
POST   /api/v1/users/:id/activate  # 激活（动作类操作）
POST   /api/v1/batch/users         # 批量操作
```

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    v1 := r.Group("/api/v1")

    users := v1.Group("/users")
    {
        users.GET("",    listUsers)       // 列表
        users.GET("/:id", getUser)        // 详情
        users.POST("",    createUser)      // 创建
        users.PUT("/:id", updateUser)      // 全量更新
        users.PATCH("/:id", patchUser)     // 部分更新
        users.DELETE("/:id", deleteUser)   // 删除
    }

    r.Run(":8080")
}
```

::: warning HTTP 状态码速查
200 成功（GET/PUT/PATCH）、201 创建成功（POST）、204 无内容（DELETE）、400 请求错误、401 未认证、403 无权限、404 不存在、409 冲突（如重复创建）、422 校验失败、429 限流、500 服务器错误。

## 2. 统一响应格式

生产项目必须定义统一的响应结构，前端根据 `code` 判断业务状态，`data` 携带业务数据，`message` 提供人类可读的描述。

::: info 设计原则
HTTP 状态码表达传输层状态（200 成功到达、401 需要认证），业务状态码 `code` 表达业务层状态（0 成功、10001 参数错误）。两者配合使用，不要用 200 包一切。

```go
package response

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// 统一响应结构
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 分页响应
type PageResponse struct {
    Code    int        `json:"code"`
    Message string     `json:"message"`
    Data    PageData   `json:"data"`
}

type PageData struct {
    List     interface{} `json:"list"`
    Total    int64       `json:"total"`
    Page     int         `json:"page"`
    PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, Response{
        Code:    0,
        Message: "created",
        Data:    data,
    })
}

func Fail(c *gin.Context, httpStatus int, code int, msg string) {
    c.JSON(httpStatus, Response{
        Code:    code,
        Message: msg,
    })
}

// 分页成功响应
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
    c.JSON(http.StatusOK, PageResponse{
        Code:    0,
        Message: "success",
        Data: PageData{
            List:     list,
            Total:    total,
            Page:     page,
            PageSize: pageSize,
        },
    })
}
```

## 3. API 文档

Swagger/OpenAPI 是 RESTful API 文档的事实标准。Go 生态中最常用的是 `swaggo/swag`，通过注解自动生成文档。

::: tip 生产建议
API 文档与代码同步维护是关键。注解写在 Handler 函数上方，CI 流程中运行 `swag init` 检查注解完整性。不要在项目后期补文档。

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// @Summary      获取用户列表
// @Description  分页查询用户列表，支持按关键词搜索
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        page      query    int     true  "页码"       minimum(1)
// @Param        page_size query    int     true  "每页数量"    minimum(1) maximum(100)
// @Param        keyword   query    string  false "搜索关键词"
// @Success      200       {object} response.PageResponse
// @Failure      400       {object} response.Response
// @Failure      401       {object} response.Response
// @Router       /api/v1/users [get]
// @Security     BearerAuth
func listUsers(c *gin.Context) {
    // Handler 实现
}

// @Summary      创建用户
// @Description  创建新用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserReq  true  "创建用户请求"
// @Success      201   {object}  response.Response
// @Failure      400   {object}  response.Response
// @Failure      409   {object}  response.Response
// @Router       /api/v1/users [post]
// @Security     BearerAuth
func createUser(c *gin.Context) {
    // Handler 实现
}

type CreateUserReq struct {
    Name  string `json:"name"  example:"张三"`
    Email string `json:"email" example:"zhangsan@example.com"`
    Age   int    `json:"age"   example:"25"`
}
```

```go
// main.go 中注册 Swagger 路由
import (
    _ "myapp/docs"           // swag init 生成的 docs 包
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Go Web API
// @version         1.0
// @description     Go Web 框架 API 文档
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
    r := gin.Default()
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    r.Run(":8080")
}
```

## 4. 最佳实践

**幂等性设计**：GET、PUT、DELETE 天然幂等。POST 非幂等，需要业务层保证（如用唯一约束防止重复创建）。生产建议：创建接口返回资源 ID，前端用 ID 做幂等重试。

**HATEOAS**（Hypermedia as the Engine of Application State）：响应中包含相关资源的链接。大多数国内项目不采用，但理解其思想有助于设计更好的 API。

**Rate Limiting**：通过响应头告知客户端限流状态，推荐使用 `X-RateLimit-Limit`、`X-RateLimit-Remaining`、`X-RateLimit-Reset` 标准 header。

::: warning 面试追问
1. PUT 和 PATCH 的区别？→ PUT 是全量替换（必须传完整资源），PATCH 是部分更新（只传修改的字段）。PUT 幂等，PATCH 不一定幂等。
2. 如何设计 API 版本控制？→ 三种方式：URL 路径 `/api/v1/`（最常用）、请求头 `Accept: application/vnd.api.v1+json`、查询参数 `?v=1`。推荐 URL 路径方式，直观且方便路由。
3. 如何处理长耗时操作？→ 返回 202 Accepted + Location header 指向任务状态查询接口，客户端轮询或 WebSocket 推送结果。

```go
package middleware

import (
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
)

// Rate Limiter 中间件（令牌桶简化版）
func RateLimitMiddleware(rps int) gin.HandlerFunc {
    type bucket struct {
        tokens  float64
        lastTime time.Time
    }
    var (
        mu      sync.Mutex
        buckets = make(map[string]*bucket)
    )

    return func(c *gin.Context) {
        key := c.ClientIP()
        mu.Lock()
        b, ok := buckets[key]
        if !ok {
            b = &bucket{tokens: float64(rps), lastTime: time.Now()}
            buckets[key] = b
        }

        // 补充令牌
        now := time.Now()
        b.tokens += now.Sub(b.lastTime).Seconds() * float64(rps)
        b.lastTime = now
        if b.tokens > float64(rps) {
            b.tokens = float64(rps)
        }

        // 消费令牌
        if b.tokens < 1 {
            mu.Unlock()
            c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rps))
            c.Header("X-RateLimit-Remaining", "0")
            c.Header("Retry-After", "1")
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "code":    42900,
                "message": "too many requests",
            })
            return
        }
        b.tokens--
        remaining := int(b.tokens)
        mu.Unlock()

        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rps))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
        c.Next()
    }
}
```
