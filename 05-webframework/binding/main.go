package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ============================================================
// 参数绑定与校验
// ============================================================
//
// 【面试高频问题】
// 1. Gin 的 ShouldBind 和 MustBind 有什么区别？
// 2. 如何自定义校验规则？
// 3. JSON 绑定时的零值问题怎么处理？
// 4. binding:"-" 和 binding:"omitempty" 的区别？

func main() {
	jsonBinding()
	queryBinding()
	customValidation()
	zeroValueProblem()
}

// ----------------------------------------------------------
// 1. JSON 绑定原理
// ----------------------------------------------------------
// Gin 的 ShouldBindJSON 底层:
//   1. 从 c.Request.Body 读取 JSON
//   2. 调用 json.Unmarshal 反序列化
//   3. 调用 validator.Validate 进行校验
//
// ShouldBind vs MustBind:
//   - ShouldBind: 校验失败返回 error，需要自行处理
//   - MustBind: 校验失败自动返回 400，不可控（不推荐）
func jsonBinding() {
	fmt.Println("=== 1. JSON 绑定 ===")

	type CreateUserReq struct {
		Name  string `json:"name" binding:"required,min=2,max=50"`
		Email string `json:"email" binding:"required,email"`
		Age   int    `json:"age" binding:"required,gte=1,lte=150"`
	}

	// 模拟 JSON 请求体
	jsonBody := `{"name":"张三","email":"zhangsan@example.com","age":25}`

	var req CreateUserReq
	if err := json.Unmarshal([]byte(jsonBody), &req); err != nil {
		fmt.Println("绑定失败:", err)
		return
	}
	fmt.Printf("绑定成功: %+v\n", req)

	// 模拟校验失败
	invalidJSON := `{"name":"","email":"invalid","age":200}`
	var req2 CreateUserReq
	if err := json.Unmarshal([]byte(invalidJSON), &req2); err != nil {
		fmt.Println("JSON 解析失败:", err)
		return
	}
	fmt.Printf("解析成功但校验失败: name=%q (空), email=%q, age=%d (超范围)\n",
		req2.Name, req2.Email, req2.Age)
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Query 参数绑定
// ----------------------------------------------------------
// Gin 支持多种绑定方式:
//   - ShouldBindJSON    → Body (application/json)
//   - ShouldBindXML     → Body (application/xml)
//   - ShouldBindQuery   → URL Query (?name=foo&age=20)
//   - ShouldBindUri     → URI Path (/users/:id)
//   - ShouldBind        → 自动根据 Content-Type 选择
func queryBinding() {
	fmt.Println("=== 2. Query 参数绑定 ===")

	type ListUsersReq struct {
		Page     int    `form:"page" binding:"required,min=1"`
		PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
		Keyword  string `form:"keyword"`
		Sort     string `form:"sort" binding:"omitempty,oneof=asc desc"`
	}

	// 模拟解析 query 参数
	query := "page=2&page_size=20&keyword=张&sort=desc"
	params := parseQuery(query)

	req := ListUsersReq{
		Page:     mustAtoi(params["page"]),
		PageSize: mustAtoi(params["page_size"]),
		Keyword:  params["keyword"],
		Sort:     params["sort"],
	}
	fmt.Printf("Query 绑定: %+v\n", req)

	// URI 绑定示例
	type GetUserReq struct {
		ID uint `uri:"id" binding:"required"`
	}
	fmt.Println("URI 绑定: GET /users/:id → GetUserReq{ID: 123}")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 自定义校验器
// ----------------------------------------------------------
// Gin 使用 go-playground/validator 库
// 注册自定义校验:
//
//   if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
//       v.RegisterValidation("mobile", validateMobile)
//   }
func customValidation() {
	fmt.Println("=== 3. 自定义校验器 ===")

	// 自定义手机号校验
	validateMobile := func(mobile string) bool {
		if len(mobile) != 11 {
			return false
		}
		if !strings.HasPrefix(mobile, "1") {
			return false
		}
		for _, c := range mobile {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}

	testCases := []string{
		"13800138000", // 有效
		"23800138000", // 无效: 不以 1 开头
		"1380013800",  // 无效: 长度不够
		"1380013800a", // 无效: 含非数字
	}

	for _, tc := range testCases {
		fmt.Printf("  %s → %v\n", tc, validateMobile(tc))
	}
	fmt.Println()

	// 自定义错误信息（通过 RegisterTranslation）
	fmt.Println("自定义错误信息:")
	fmt.Println("  validator.RegisterTranslation(\"mobile\", trans, ")
	fmt.Println("    func(ut ut.Translator) error { return ut.Add(\"mobile\", \"手机号格式不正确\", true) },")
	fmt.Println("    func(ut ut.Translator, fe validator.FieldError) string { ... })")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. JSON 零值问题
// ----------------------------------------------------------
// 问题: JSON 中省略字段和传零值无法区分
//
//   {"age": 0}     → age 是 0（明确传了 0）
//   {}             → age 也是 0（没传，默认值）
//
// 解决方案:
//   1. 使用指针: Age *int `json:"age"` → nil 表示未传
//   2. 使用 omitempty: 忽略零值（但不能区分"未传"和"传了零值"）
//   3. 使用特殊类型: sql.NullInt64、自定义 Optional 类型
func zeroValueProblem() {
	fmt.Println("=== 4. JSON 零值问题 ===")

	type UpdateUserReq struct {
		Name *string `json:"name"` // 指针: nil 表示未传
		Age  *int    `json:"age"`  // 指针: nil 表示未传
	}

	// 场景1: 传了 age=0
	json1 := `{"age":0}`
	var req1 UpdateUserReq
	json.Unmarshal([]byte(json1), &req1)
	fmt.Printf("  {\"age\":0} → Age=%v (明确传了0)\n", req1.Age)

	// 场景2: 没传 age
	json2 := `{}`
	var req2 UpdateUserReq
	json.Unmarshal([]byte(json2), &req2)
	fmt.Printf("  {}       → Age=%v (未传，nil)\n", req2.Age)

	// 通过指针可以区分
	if req1.Age != nil {
		fmt.Printf("  req1 更新 age 为 %d\n", *req1.Age)
	} else {
		fmt.Println("  req1 跳过 age 更新")
	}
	if req2.Age != nil {
		fmt.Printf("  req2 更新 age 为 %d\n", *req2.Age)
	} else {
		fmt.Println("  req2 跳过 age 更新")
	}
	fmt.Println()

	// 面试追问: omitempty 的坑
	fmt.Println("binding:\"omitempty\" 的坑:")
	fmt.Println("  omitempty 会跳过零值校验")
	fmt.Println("  type Req struct {")
	fmt.Println("    Status int `binding:\"omitempty,min=1\"`") // 传 0 时不校验
	fmt.Println("  }")
	fmt.Println("  → 如果需要区分零值，用指针 + required")
	fmt.Println()
}

// --- 辅助函数 ---

func parseQuery(query string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(query, "&") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// 避免未使用导入
var _ = reflect.TypeOf(0)
