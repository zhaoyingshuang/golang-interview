package main

import (
	"fmt"
	"net/http"
)

// ============================================================
// RESTful API 设计规范
// ============================================================
//
// 【面试高频问题】
// 1. RESTful 的核心原则是什么？
// 2. PUT 和 PATCH 的区别？
// 3. 如何设计 API 版本控制？
// 4. HTTP 状态码 301 和 302 的区别？
// 5. 如何设计统一的错误响应格式？

func main() {
	restfulPrinciples()
	statusCodes()
	versioning()
	responseFormat()
	idempotent()
}

// ----------------------------------------------------------
// 1. RESTful 核心原则
// ----------------------------------------------------------
// REST = Representational State Transfer
//
// 核心约束:
//   1. 资源(Resource) — 用 URL 标识: /users, /users/:id
//   2. 统一接口 — 用 HTTP 方法表达操作:
//      GET    → 查询（幂等、安全）
//      POST   → 创建（非幂等）
//      PUT    → 全量更新（幂等）
//      PATCH  → 部分更新
//      DELETE → 删除（幂等）
//   3. 无状态 — 每个请求包含所有必要信息
//   4. HATEOAS — 响应中包含相关资源链接（实际很少用）
//
// URL 设计规范:
//   ✓ /users          → 用户集合
//   ✓ /users/123      → 单个用户
//   ✓ /users/123/orders → 用户的订单
//   ✗ /getUsers       → 不要用动词
//   ✗ /user/list      → 不要用嵌套的 CRUD
func restfulPrinciples() {
	fmt.Println("=== 1. RESTful 核心原则 ===")

	routes := []struct {
		Method  string
		Path    string
		Desc    string
		Handler string
	}{
		{"GET", "/users", "获取用户列表", "ListUsers"},
		{"GET", "/users/:id", "获取单个用户", "GetUser"},
		{"POST", "/users", "创建用户", "CreateUser"},
		{"PUT", "/users/:id", "全量更新用户", "UpdateUser"},
		{"PATCH", "/users/:id", "部分更新用户", "PatchUser"},
		{"DELETE", "/users/:id", "删除用户", "DeleteUser"},
		{"GET", "/users/:id/orders", "获取用户的订单", "ListUserOrders"},
	}

	fmt.Println("RESTful 路由设计:")
	for _, r := range routes {
		fmt.Printf("  %-8s %-25s → %s (%s)\n", r.Method, r.Path, r.Handler, r.Desc)
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 2. HTTP 状态码使用
// ----------------------------------------------------------
func statusCodes() {
	fmt.Println("=== 2. HTTP 状态码 ===")

	codes := []struct {
		Code int
		Desc string
		Use  string
	}{
		{200, "OK", "GET 成功、PUT/PATCH 成功"},
		{201, "Created", "POST 创建成功，应返回 Location 头"},
		{204, "No Content", "DELETE 成功，无返回体"},
		{301, "Moved Permanently", "永久重定向，SEO 权重转移"},
		{302, "Found", "临时重定向，SEO 权重不转移"},
		{304, "Not Modified", "协商缓存命中"},
		{400, "Bad Request", "参数校验失败、请求格式错误"},
		{401, "Unauthorized", "未认证（未登录）"},
		{403, "Forbidden", "已认证但无权限"},
		{404, "Not Found", "资源不存在"},
		{409, "Conflict", "资源冲突（如重复创建）"},
		{422, "Unprocessable Entity", "语义错误（业务校验失败）"},
		{429, "Too Many Requests", "限流触发"},
		{500, "Internal Server Error", "服务端内部错误"},
		{502, "Bad Gateway", "网关上游错误"},
		{503, "Service Unavailable", "服务不可用（过载/维护）"},
		{504, "Gateway Timeout", "网关上游超时"},
	}

	fmt.Println("常用状态码:")
	for _, c := range codes {
		fmt.Printf("  %d %-25s → %s\n", c.Code, c.Desc, c.Use)
	}

	// 面试追问: 401 vs 403?
	fmt.Println()
	fmt.Println("面试重点: 401 vs 403")
	fmt.Println("  401 = 不知道你是谁（未登录）")
	fmt.Println("  403 = 知道你是谁，但你没权限")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. API 版本控制
// ----------------------------------------------------------
func versioning() {
	fmt.Println("=== 3. API 版本控制 ===")

	approaches := []struct {
		Name   string
		Example string
		Pros   string
		Cons   string
	}{
		{
			"URL 路径",
			"/api/v1/users",
			"直观，浏览器可直接访问",
			"版本切换需改 URL",
		},
		{
			"请求头",
			"Accept: application/vnd.api.v1+json",
			"URL 不变，RESTful 纯粹",
			"调试不便",
		},
		{
			"Query 参数",
			"/api/users?version=1",
			"简单",
			"容易被忽略",
		},
	}

	for _, a := range approaches {
		fmt.Printf("  %s: %s\n", a.Name, a.Example)
		fmt.Printf("    优点: %s\n", a.Pros)
		fmt.Printf("    缺点: %s\n", a.Cons)
	}
	fmt.Println()
	fmt.Println("  推荐: URL 路径版本（最常用、最直观）")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 统一响应格式
// ----------------------------------------------------------
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details any         `json:"details,omitempty"`
}

type PagedData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func responseFormat() {
	fmt.Println("=== 4. 统一响应格式 ===")

	// 成功响应
	success := Response{
		Code:    0,
		Message: "success",
		Data: map[string]any{
			"id":   123,
			"name": "张三",
		},
	}
	fmt.Println("成功响应:")
	printJSON(success)

	// 分页响应
	paged := Response{
		Code:    0,
		Message: "success",
		Data: PagedData{
			List:     []string{"user1", "user2"},
			Total:    100,
			Page:     1,
			PageSize: 20,
		},
	}
	fmt.Println("分页响应:")
	printJSON(paged)

	// 错误响应
	err := ErrorResponse{
		Code:    40001,
		Message: "参数校验失败",
		Details: map[string]string{
			"email": "邮箱格式不正确",
			"age":   "必须大于0",
		},
	}
	fmt.Println("错误响应:")
	printJSON(err)
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 幂等性设计
// ----------------------------------------------------------
// 幂等性: 同一操作执行一次和多次效果相同
//
// 天然幂等: GET, PUT, DELETE
// 非幂等: POST (每次创建一个新资源)
//
// 如何让 POST 也幂等:
//   1. 客户端生成幂等键 (Idempotency-Key)
//   2. 服务端根据 key 去重
//   3. 重复请求返回之前的结果
func idempotent() {
	fmt.Println("=== 5. 幂等性设计 ===")

	// 模拟幂等键去重
	type IdempotentStore struct {
		results map[string]any
	}

	store := &IdempotentStore{results: make(map[string]any)}

	processOrder := func(idempotencyKey string, orderID string) any {
		// 检查是否已处理
		if result, ok := store.results[idempotencyKey]; ok {
			fmt.Printf("  [幂等] 重复请求 key=%s, 返回缓存结果\n", idempotencyKey)
			return result
		}

		// 处理业务逻辑
		result := map[string]any{"order_id": orderID, "status": "created"}
		store.results[idempotencyKey] = result
		fmt.Printf("  [新建] 处理请求 key=%s, orderID=%s\n", idempotencyKey, orderID)
		return result
	}

	// 第一次请求
	processOrder("idem-abc-123", "ORD-001")
	// 重复请求（网络重试）
	processOrder("idem-abc-123", "ORD-001")
	// 不同请求
	processOrder("idem-def-456", "ORD-002")

	fmt.Println()
	fmt.Println("幂等键通常放在 HTTP Header:")
	fmt.Println("  Idempotency-Key: <uuid>")
	fmt.Println()
}

func printJSON(v any) {
	fmt.Printf("  %+v\n", v)
}

var _ = http.StatusOK // 避免未使用导入
