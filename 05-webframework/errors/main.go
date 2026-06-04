package main

import (
	"errors"
	"fmt"
)

// ============================================================
// 错误处理规范
// ============================================================
//
// 【面试高频问题】
// 1. Go 的 error 和 panic 的使用场景？
// 2. errors.Is 和 errors.As 的区别？
// 3. 如何设计业务错误码体系？
// 4. 自定义错误类型有哪些方式？
// 5. 错误链（Error Chain）是什么？

func main() {
	errorBasics()
	customError()
	errorChain()
	errorCodeSystem()
	panicVsError()
}

// ----------------------------------------------------------
// 1. Go 错误处理基础
// ----------------------------------------------------------
// Go 的 error 是一个接口: type error interface { Error() string }
//
// 核心原则:
//   - error 用于可预期的错误（业务错误、IO 失败）
//   - panic 用于不可恢复的错误（空指针、数组越界）
//   - recover 在 defer 中捕获 panic
func errorBasics() {
	fmt.Println("=== 1. 错误处理基础 ===")

	// 基本错误创建
	err1 := errors.New("something went wrong")
	err2 := fmt.Errorf("user %d not found", 42)
	fmt.Printf("  err1: %v\n", err1)
	fmt.Printf("  err2: %v\n", err2)
	fmt.Println()

	// errors.Is — 判断错误链中是否包含特定错误
	ErrNotFound := errors.New("not found")
	wrappedErr := fmt.Errorf("query failed: %w", ErrNotFound)
	fmt.Printf("  errors.Is(wrappedErr, ErrNotFound) = %v\n", errors.Is(wrappedErr, ErrNotFound))

	// errors.As — 从错误链中提取特定类型的错误
	type TimeoutError struct{ Duration int }
	var timeoutErr TimeoutError
	otherErr := fmt.Errorf("operation failed: %w", TimeoutError{Duration: 30})
	fmt.Printf("  errors.As match: %v\n", errors.As(otherErr, &timeoutErr))
	fmt.Printf("  extracted duration: %d\n", timeoutErr.Duration)
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 自定义错误类型
// ----------------------------------------------------------
// 三种方式:
//   1. sentinel error: var ErrNotFound = errors.New("not found")
//   2. 自定义 struct: type NotFoundError struct { ... }
//   3. fmt.Errorf + %w: 包装底层错误添加上下文
func customError() {
	fmt.Println("=== 2. 自定义错误类型 ===")

	// 方式1: Sentinel Error（哨兵错误）
	var (
		ErrUserNotFound = errors.New("user not found")
		ErrDuplicate    = errors.New("duplicate entry")
	)

	// 使用 errors.Is 判断
	err := fmt.Errorf("create user failed: %w", ErrDuplicate)
	if errors.Is(err, ErrDuplicate) {
		fmt.Println("  是重复错误")
	}
	_ = ErrUserNotFound

	// 方式2: 自定义错误类型
	type BizError struct {
		Code    int
		Message string
	}

	// 实现 error 接口
	// func (e *BizError) Error() string { return fmt.Sprintf("[%d] %s", e.Code, e.Message) }
	bizErr := &BizError{Code: 40001, Message: "余额不足"}
	fmt.Printf("  自定义错误: code=%d, msg=%s\n", bizErr.Code, bizErr.Message)

	// 方式3: 用 fmt.Errorf 包装上下文
	baseErr := errors.New("connection refused")
	wrapped := fmt.Errorf("dial db host=%s port=%d: %w", "localhost", 3306, baseErr)
	fmt.Printf("  包装错误: %v\n", wrapped)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 错误链 (Error Chain)
// ----------------------------------------------------------
// Go 1.13+ 的 %w 动词支持错误链:
//
//   err1 := errors.New("base error")
//   err2 := fmt.Errorf("layer2: %w", err1)
//   err3 := fmt.Errorf("layer3: %w", err2)
//
//   errors.Is(err3, err1)  → true  (沿链查找)
//   errors.Unwrap(err3)    → err2  (解包一层)
//
// 面试追问: errors.Is 和 == 的区别？
// 答: == 只比较最外层，errors.Is 会遍历整个错误链。
func errorChain() {
	fmt.Println("=== 3. 错误链 ===")

	baseErr := errors.New("disk full")
	layer2 := fmt.Errorf("write log failed: %w", baseErr)
	layer3 := fmt.Errorf("request processing failed: %w", layer2)

	fmt.Printf("  层3: %v\n", layer3)
	fmt.Printf("  Unwrap → 层2: %v\n", errors.Unwrap(layer3))
	fmt.Printf("  Unwrap → 层1: %v\n", errors.Unwrap(errors.Unwrap(layer3)))
	fmt.Printf("  errors.Is(layer3, baseErr) = %v\n", errors.Is(layer3, baseErr))
	fmt.Println()

	// 多错误包装 (Go 1.20+)
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	combined := errors.Join(err1, err2)
	fmt.Printf("  Join: %v\n", combined)
	fmt.Printf("  errors.Is(combined, err1) = %v\n", errors.Is(combined, err1))
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 业务错误码体系
// ----------------------------------------------------------
// 设计原则:
//   - 错误码唯一，可快速定位问题
//   - 分层设计: 模块 + 具体错误
//   - 区分系统错误和业务错误
func errorCodeSystem() {
	fmt.Println("=== 4. 业务错误码体系 ===")

	// 错误码设计示例
	// 格式: AABBBB (AA=模块, BBBB=具体错误)
	//
	// 10xxxx — 用户模块
	//   100001 — 用户不存在
	//   100002 — 密码错误
	//   100003 — 用户已禁用
	//
	// 20xxxx — 订单模块
	//   200001 — 订单不存在
	//   200002 — 库存不足
	//   200003 — 订单状态不允许此操作
	//
	// 50xxxx — 系统错误
	//   500001 — 内部服务错误
	//   500002 — 数据库错误
	//   500003 — 第三方服务超时

	type Code struct {
		HTTPStatus int
		Code       int
		Message    string
	}

	codes := map[string]Code{
		"UserNotFound":     {404, 100001, "用户不存在"},
		"PasswordWrong":    {401, 100002, "密码错误"},
		"InsufficientStock": {409, 200002, "库存不足"},
		"InternalError":    {500, 500001, "内部服务错误"},
	}

	for name, c := range codes {
		fmt.Printf("  %s: HTTP=%d, Code=%d, Msg=%s\n", name, c.HTTPStatus, c.Code, c.Message)
	}

	// Web 框架中的错误处理中间件
	fmt.Println()
	fmt.Println("错误处理中间件:")
	fmt.Println("  func ErrorHandler(c *gin.Context) {")
	fmt.Println("    c.Next()")
	fmt.Println("    if len(c.Errors) > 0 {")
	fmt.Println("      err := c.Errors.Last().Err")
	fmt.Println("      switch e := err.(type) {")
	fmt.Println("      case *BizError:")
	fmt.Println("        c.JSON(e.HTTPStatus, ErrorResponse{Code: e.Code, Message: e.Message})")
	fmt.Println("      default:")
	fmt.Println("        c.JSON(500, ErrorResponse{Code: 500001, Message: \"内部错误\"})")
	fmt.Println("      }")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. Panic vs Error 使用场景
// ----------------------------------------------------------
func panicVsError() {
	fmt.Println("=== 5. Panic vs Error ===")

	fmt.Println("使用 error (可恢复的错误):")
	fmt.Println("  - 文件不存在")
	fmt.Println("  - 网络超时")
	fmt.Println("  - 参数校验失败")
	fmt.Println("  - 业务规则违反")
	fmt.Println()
	fmt.Println("使用 panic (不可恢复的严重错误):")
	fmt.Println("  - 数组越界（编程错误）")
	fmt.Println("  - 空指针解引用（编程错误）")
	fmt.Println("  - 初始化失败（配置缺失）")
	fmt.Println("  - 不变式被破坏")
	fmt.Println()
	fmt.Println("Recovery 中间件模式:")
	fmt.Println("  func Recovery() gin.HandlerFunc {")
	fmt.Println("    return func(c *gin.Context) {")
	fmt.Println("      defer func() {")
	fmt.Println("        if r := recover(); r != nil {")
	fmt.Println("          log.Printf(\"panic: %v\\n%s\", r, debug.Stack())")
	fmt.Println("          c.AbortWithStatusJSON(500, H{\"error\": \"Internal Server Error\"})")
	fmt.Println("        }")
	fmt.Println("      }()")
	fmt.Println("      c.Next()")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
}
