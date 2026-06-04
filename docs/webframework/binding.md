---
title: 参数绑定与校验
---

## 1. ShouldBind 系列

Gin 提供 `ShouldBind` 系列方法，根据请求的 `Content-Type` 自动选择绑定器。底层使用 `encoding/json`、`encoding/xml`、`go-playground/validator` 等标准库和第三方库。

核心方法：`ShouldBindJSON`、`ShouldBindXML`、`ShouldBindQuery`、`ShouldBindUri`。`ShouldBind` 根据 Content-Type 自动推断。`MustBind` 系列内部调用 `ShouldBind`，失败时自动写 400 响应——生产环境不推荐，因为无法自定义错误格式。

::: tip 生产建议
始终使用 `ShouldBind` 系列，手动处理错误返回统一格式。为每个接口定义独立的请求结构体，用 struct tag 标注绑定来源和校验规则。

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// JSON Body 绑定
type CreateUserReq struct {
    Name  string `json:"name" binding:"required,min=2,max=50"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"required,gte=1,lte=150"`
}

// Query 参数绑定
type ListUsersReq struct {
    Page     int    `form:"page" binding:"required,gte=1"`
    PageSize int    `form:"page_size" binding:"required,gte=1,lte=100"`
    Keyword  string `form:"keyword" binding:"omitempty,max=100"`
    Status   string `form:"status" binding:"omitempty,oneof=active inactive"`
}

// URI 路径参数绑定
type GetUserReq struct {
    ID int64 `uri:"id" binding:"required,gt=0"`
}

func main() {
    r := gin.Default()

    // JSON Body
    r.POST("/users", func(c *gin.Context) {
        var req CreateUserReq
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "code":    40000,
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusOK, gin.H{"data": req})
    })

    // Query 参数
    r.GET("/users", func(c *gin.Context) {
        var req ListUsersReq
        if err := c.ShouldBindQuery(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "code":    40000,
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusOK, gin.H{"data": req})
    })

    // URI 路径参数
    r.GET("/users/:id", func(c *gin.Context) {
        var req GetUserReq
        if err := c.ShouldBindUri(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "code":    40000,
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusOK, gin.H{"user_id": req.ID})
    })

    r.Run(":8080")
}
```

## 2. 自定义校验器

Gin 默认使用 `go-playground/validator/v10`，支持丰富的内置校验规则。当内置规则不够用时，可以注册自定义校验函数。

::: info 常用校验规则
`required`（必填）、`omitempty`（为空时跳过后续校验）、`min/max`（字符串长度/数值范围）、`oneof`（枚举值）、`email`、`url`、`datetime=2006-01-02`（时间格式）、`dive`（校验切片/map 元素）。

```go
package main

import (
    "fmt"
    "reflect"
    "regexp"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
)

type RegisterReq struct {
    Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
    Password string `json:"password" binding:"required,strongpwd"`
    Phone    string `json:"phone" binding:"required,cnphone"`
}

func main() {
    r := gin.Default()

    // 注册自定义校验规则
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        // 强密码校验: 至少8位，包含大小写字母和数字
        _ = v.RegisterValidation("strongpwd", func(fl validator.FieldLevel) bool {
            pwd := fl.Field().String()
            if len(pwd) < 8 {
                return false
            }
            hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pwd)
            hasLower := regexp.MustCompile(`[a-z]`).MatchString(pwd)
            hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pwd)
            return hasUpper && hasLower && hasDigit
        })

        // 中国手机号校验
        _ = v.RegisterValidation("cnphone", func(fl validator.FieldLevel) bool {
            phone := fl.Field().String()
            matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
            return matched
        })
    }

    r.POST("/register", func(c *gin.Context) {
        var req RegisterReq
        if err := c.ShouldBindJSON(&req); err != nil {
            // 提取校验错误信息
            errs := err.(validator.ValidationErrors)
            for _, e := range errs {
                fmt.Printf("field: %s, rule: %s, value: %v\n",
                    e.Field(), e.Tag(), e.Value())
            }
            c.JSON(400, gin.H{"code": 40000, "message": "参数校验失败"})
            return
        }
        c.JSON(200, gin.H{"data": req})
    })

    r.Run(":8080")
}
```

::: warning 面试追问
如何实现校验错误信息国际化？→ 注册 `RegisterTranslation` 和 `RegisterValidation`，将 field name 映射为中文，错误消息模板支持翻译。使用 `go-playground/locales` 和 `go-playground/universal-translator`。

## 3. 自定义绑定

Gin 的 `Binding` 接口允许自定义绑定逻辑。当需要处理非标准格式（如 Protobuf、MsgPack、自定义 Header 绑定）时，实现 `Binding` 接口即可。

```go
package main

import (
    "encoding/csv"
    "io"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
)

// 自定义 CSV 绑定器
type csvBinding struct{}

func (csvBinding) Name() string {
    return "csv"
}

func (csvBinding) Bind(req *http.Request, obj any) error {
    reader := csv.NewReader(req.Body)
    defer req.Body.Close()

    // 读取表头
    headers, err := reader.Read()
    if err != nil {
        return err
    }

    // 读取数据行（简化示例，只取第一行）
    record, err := reader.Read()
    if err != nil && err != io.EOF {
        return err
    }

    // 将 CSV 数据映射到结构体（实际项目可用反射）
    _ = headers
    _ = record
    return nil
}

// 使用方式
func main() {
    r := gin.Default()

    r.POST("/upload", func(c *gin.Context) {
        var data map[string]string
        if err := c.MustBindWith(&data, csvBinding{}); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"data": data})
    })

    r.Run(":8080")
}
```

::: tip 生产建议
大多数场景不需要自定义绑定器。JSON + validator 覆盖 95% 的需求。遇到 CSV/Excel 上传时，直接在 Handler 中手动解析更清晰。

## 4. 常见问题

**问题1: JSON 数字精度丢失**。JSON 的 number 类型在 Go 中默认解析为 `float64`，大整数（如订单号、雪花ID）会丢失精度。解决：使用 `json.Number` 或直接用 `string` tag。

**问题2: 时间格式**。Go 的时间序列化默认 RFC3339，前端可能需要 `yyyy-MM-dd HH:mm:ss`。解决：自定义 `Time` 类型实现 `MarshalJSON`。

**问题3: 嵌套结构体校验**。嵌套结构体需要加 `binding:"required"` 且内部字段需配合 `dive` 标签校验切片元素。

**问题4: 零值绑定问题**。`ShouldBindJSON` 不会绑定零值字段（JSON 中未传的字段保持零值）。区分"未传"和"传了零值"需要用指针类型。

::: warning 面试追问
1. `ShouldBind` 和 `Bind` 的区别？→ `Bind` 在失败时自动写 400 响应并设置 `Content-Type`，`ShouldBind` 只返回 error。
2. 如何区分"字段未传"和"传了零值"？→ 用指针：`Age *int `json:"age"``，未传时为 nil。
3. validator 的 `omitempty` 和 `required` 能同时用吗？→ 不能，`omitempty` 在字段为零值时跳过所有后续校验。

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"
)

// 问题1: 大数字精度 — 使用 string tag
type OrderReq struct {
    OrderID string `json:"order_id" binding:"required"` // string 接收
}

// 问题2: 自定义时间格式
type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
    formatted := time.Time(t).Format("2006-01-02 15:04:05")
    return json.Marshal(formatted)
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
    str := string(data)
    parsed, err := time.Parse(`"2006-01-02 15:04:05"`, str)
    if err != nil {
        return err
    }
    *t = JSONTime(parsed)
    return nil
}

// 问题4: 区分未传和零值
type UpdateUserReq struct {
    Name  *string `json:"name"`   // nil=未传, &""=传了空串
    Age   *int    `json:"age"`    // nil=未传, &0=传了0
    Email *string `json:"email"`  // nil=未传
}

func demo() {
    raw := []byte(`{"name":"","email":null}`)
    var req UpdateUserReq
    json.Unmarshal(raw, &req)
    fmt.Printf("name: %v, age: %v, email: %v\n",
        req.Name, req.Age, req.Email)
    // name: 0xc0001 (), age: <nil> (未传), email: <nil> (传了null)
}
```
