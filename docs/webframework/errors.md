---
title: 错误处理
---

## 1. Go 错误处理哲学

Go 的错误处理哲学与异常模型（Java/Python）截然不同：**错误是值（Values），不是控制流机制**。`error` 接口只有一个 `Error() string` 方法，任何实现了该方法的类型都是 error。函数通过返回值传递错误，调用方通过 `if err != nil` 检查。

Go 1.13 引入的 `errors.Is` 和 `errors.As` 解决了错误包装后的类型判断问题。`fmt.Errorf("...: %w", err)` 用 `%w` 动词包装错误，`errors.Is` 沿着错误链逐层解包比较，`errors.As` 提取链中特定类型的错误。

::: tip 生产建议
永远不要用 `_` 忽略 error。即使暂时不处理，也打印日志。在 Web 项目中，error 应该在 Handler 层统一转换为 HTTP 响应，不要把内部错误信息直接暴露给客户端。

```go
package main

import (
    "errors"
    "fmt"
)

// 自定义错误类型
type BizError struct {
    Code    int    // 业务错误码
    Message string // 错误描述
    Cause   error  // 原始错误
}

func (e *BizError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *BizError) Unwrap() error {
    return e.Cause
}

// 预定义业务错误
var (
    ErrUserNotFound = &BizError{Code: 10001, Message: "用户不存在"}
    ErrInvalidParam = &BizError{Code: 10002, Message: "参数错误"}
)

// 使用 errors.Is 和 errors.As
func demo() {
    err := getUserFromDB(42)
    if err != nil {
        // errors.Is — 判断是否是特定错误（沿链查找）
        if errors.Is(err, ErrUserNotFound) {
            fmt.Println("用户不存在")
        }

        // errors.As — 提取特定类型
        var bizErr *BizError
        if errors.As(err, &bizErr) {
            fmt.Printf("业务错误: code=%d, msg=%s\n", bizErr.Code, bizErr.Message)
        }
    }
}

func getUserFromDB(id int) error {
    // 包装底层错误
    return fmt.Errorf("query user %d: %w", id, ErrUserNotFound)
}
```

## 2. Web 框架错误处理

Gin 框架的错误处理分为两层：**中间件层**（全局 Recovery 捕获 panic）和 **Handler 层**（业务错误转 HTTP 响应）。生产环境需要一个统一的错误处理中间件，将业务错误自动映射为 HTTP 响应。

::: warning 生产注意
Gin 默认的 Recovery 中间件只写纯文本响应，不符合 API 的 JSON 格式要求。务必替换为自定义 Recovery 中间件。

```go
package main

import (
    "errors"
    "log"
    "net/http"
    "runtime/debug"
    "github.com/gin-gonic/gin"
)

// 统一错误响应中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next() // 先执行后续中间件和 Handler

        // 检查是否有错误（由 Handler 通过 c.Set 设置）
        if errVal, exists := c.Get("error"); exists {
            err := errVal.(error)

            var bizErr *BizError
            if errors.As(err, &bizErr) {
                // 业务错误 → 对应的 HTTP 状态码
                httpStatus := bizErrToHTTPStatus(bizErr)
                c.JSON(httpStatus, gin.H{
                    "code":    bizErr.Code,
                    "message": bizErr.Message,
                })
                return
            }

            // 未知错误 → 500
            log.Printf("unexpected error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "code":    50000,
                "message": "服务内部错误",
            })
        }
    }
}

// 自定义 Recovery 中间件
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("panic recovered: %v\n%s", r, debug.Stack())
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "code":    50001,
                    "message": "服务内部错误",
                })
            }
        }()
        c.Next()
    }
}

func bizErrToHTTPStatus(err *BizError) int {
    switch err.Code {
    case 10001:
        return http.StatusNotFound
    case 10002:
        return http.StatusBadRequest
    case 10003:
        return http.StatusUnauthorized
    case 10004:
        return http.StatusForbidden
    default:
        return http.StatusInternalServerError
    }
}

// Handler 中设置错误
func getUser(c *gin.Context) {
    user, err := userService.GetByID(c.Param("id"))
    if err != nil {
        c.Set("error", err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "data": user,
    })
}
```

## 3. 错误码体系

生产项目需要一套分层的错误码体系。错误码应有固定规则，方便按前缀定位模块和严重程度。

::: info 错误码设计规范
建议 5 位数字编码：前 2 位为模块编号（10=用户、20=订单、30=支付），后 3 位为具体错误编号。0 表示成功，非 0 表示失败。

```go
package errors

// 错误码常量（5 位数字）
const (
    // 通用错误 00xxx
    CodeSuccess       = 0
    CodeInternalError = 50000
    CodeParamError    = 40000
    CodeUnauthorized  = 40100
    CodeForbidden     = 40300
    CodeNotFound      = 40400
    CodeTooManyReqs   = 42900

    // 用户模块 10xxx
    CodeUserNotFound    = 10001
    CodeUserExists      = 10002
    CodeUserDisabled    = 10003
    CodePasswordWrong   = 10004
    CodeTokenExpired    = 10005
    CodeTokenInvalid    = 10006

    // 订单模块 20xxx
    CodeOrderNotFound   = 20001
    CodeOrderCancelled  = 20002
    CodeOrderPaid       = 20003

    // 支付模块 30xxx
    CodePayFailed       = 30001
    CodePayTimeout      = 30002
    CodeRefundFailed    = 30003
)

// 错误码消息映射
var codeMessages = map[int]string{
    CodeSuccess:       "成功",
    CodeInternalError: "服务内部错误",
    CodeParamError:    "参数错误",
    CodeUnauthorized:  "未认证",
    CodeForbidden:     "无权限",
    CodeNotFound:      "资源不存在",
    CodeUserNotFound:  "用户不存在",
    CodeUserExists:    "用户已存在",
    CodePasswordWrong: "密码错误",
    CodeTokenExpired:  "Token 已过期",
    CodeTokenInvalid:  "Token 无效",
    CodeOrderNotFound: "订单不存在",
    CodePayFailed:     "支付失败",
    CodePayTimeout:    "支付超时",
}

// New 创建业务错误
func New(code int, cause ...error) *BizError {
    msg, ok := codeMessages[code]
    if !ok {
        msg = "未知错误"
    }
    bizErr := &BizError{Code: code, Message: msg}
    if len(cause) > 0 {
        bizErr.Cause = cause[0]
    }
    return bizErr
}

// 使用示例
func (s *userService) GetByID(id string) (*User, error) {
    if id == "" {
        return nil, New(CodeParamError)
    }
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, New(CodeUserNotFound, err)
    }
    return user, nil
}
```

## 4. 日志与监控

错误处理的最后一步是**可观测性**：结构化日志记录错误上下文，错误上报到监控系统，配置告警策略。

::: tip 生产建议
日志使用结构化格式（JSON），包含 requestID、userID、错误码、耗时等字段。日志级别：ERROR（需要人工介入）、WARN（可自动恢复）、INFO（关键业务事件）。不要在日志中输出敏感信息（密码、Token）。

```go
package logger

import (
    "context"
    "log/slog"
    "os"
)

// 结构化日志初始化
func InitLogger() *slog.Logger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
    return slog.New(handler)
}

// 带上下文的错误日志
func LogError(ctx context.Context, msg string, err error, args ...any) {
    args = append(args,
        "error", err.Error(),
        "requestID", ctx.Value("requestID"),
        "userID", ctx.Value("userID"),
    )
    slog.Error(msg, args...)
}

// 在中间件中使用
// func LatencyMiddleware() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         start := time.Now()
//         c.Next()
//         duration := time.Since(start)
//         status := c.Writer.Status()
//
//         slog.Info("request completed",
//             "method", c.Request.Method,
//             "path", c.Request.URL.Path,
//             "status", status,
//             "duration_ms", duration.Milliseconds(),
//             "requestID", c.GetString("requestID"),
//             "clientIP", c.ClientIP(),
//         )
//
//         // 5xx 错误告警
//         if status >= 500 {
//             LogError(c.Request.Context(), "server error", nil,
//                 "status", status,
//                 "path", c.Request.URL.Path,
//             )
//         }
//     }
// }
```

::: warning 告警策略
生产环境的错误告警分级：P0（5xx 错误率 > 1%，立即通知）、P1（特定业务错误激增，5 分钟内通知）、P2（慢请求增多，工作时间处理）。使用 Prometheus + AlertManager 或云厂商的监控服务。

```go
// Prometheus 错误指标采集
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP 请求总数
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // 业务错误计数
    bizErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "biz_errors_total",
            Help: "Total number of business errors",
        },
        []string{"code", "module"},
    )

    // 请求耗时直方图
    httpDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

// 在中间件中采集指标
// func MetricsMiddleware() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         start := time.Now()
//         c.Next()
//         duration := time.Since(start).Seconds()
//         status := strconv.Itoa(c.Writer.Status())
//         httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
//         httpDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
//     }
// }
```

::: warning 面试追问
1. Go 为什么不用 try-catch？→ 错误是值的设计让控制流更清晰，避免异常跳转带来的心智负担。`if err != nil` 虽然冗长但易于审查。
2. `errors.Is` 和 `==` 判断的区别？→ `==` 只判断顶层错误，`errors.Is` 会沿 Unwrap 链逐层查找。包装后的错误用 `==` 永远为 false。
3. 如何防止 goroutine 中的 panic 导致整个进程崩溃？→ 每个 goroutine 入口处加 `defer func() { recover() }()`，或使用 `errgroup` 管理错误传播。
