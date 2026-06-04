package main

import (
	"fmt"
	"net/http"
)

// ============================================================
// Web 框架路由原理
// ============================================================
//
// 【面试高频问题】
// 1. Gin 的路由树是什么结构？查找时间复杂度？
// 2. 静态路由和参数路由的优先级？
// 3. 路由分组是如何实现的？
// 4. httprouter 的压缩基数树(Radix Tree)原理？
// 5. 如何处理路由冲突？

func main() {
	radixTreeConcept()
	stdMuxDemo()
	priorityDemo()
	groupDemo()
}

// ----------------------------------------------------------
// 1. 压缩基数树 (Radix Tree) 概念
// ----------------------------------------------------------
// Gin 底层使用 httprouter 的压缩基数树（Compressed Radix Tree / Prefix Tree）
//
// 特点:
//   - 每条路径的公共前缀只存储一次
//   - 查找时间复杂度 O(k)，k 是路径长度（不是节点数）
//   - 支持 :param（参数捕获）和 *wildcard（通配符）
//
// 例如注册以下路由:
//   /user/profile
//   /user/settings
//   /user/:id
//   /user/:id/posts
//
// 树结构（简化）:
//
//   /user/
//   ├── profile     (静态节点)
//   ├── settings    (静态节点)
//   └── :id         (参数节点)
//       └── /posts  (静态节点)
//
// httprouter 为每种 HTTP 方法维护一棵独立的树（GET/POST/PUT...）
// 所以 GET /user 和 POST /user 是不同树上的节点
func radixTreeConcept() {
	fmt.Println("=== 1. 压缩基数树概念 ===")
	fmt.Println("httprouter 使用压缩基数树（Compressed Radix Tree）")
	fmt.Println("查找复杂度: O(k)，k 为路径长度")
	fmt.Println("每种 HTTP 方法维护一棵独立树")
	fmt.Println("支持 :param 参数捕获和 *wildcard 通配符")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 标准库 net/http ServeMux 路由
// ----------------------------------------------------------
// 标准库 ServeMux 的路由规则:
//   - 固定路径: /api/users  只匹配精确路径
//   - 以 / 结尾的路径: /api/  匹配 /api/* 所有子路径
//   - 最长匹配原则: /api/users/ 优先于 /api/
//   - 不支持参数路由，不支持正则
func stdMuxDemo() {
	fmt.Println("=== 2. 标准库 ServeMux ===")

	mux := http.NewServeMux()

	// 精确匹配 — 只匹配 /api/health
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "health: exact match, path=%s\n", r.URL.Path)
	})

	// 以 / 结尾 — 匹配所有子路径（最长匹配）
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "users: prefix match, path=%s\n", r.URL.Path)
	})

	// 根路径
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "root: fallback, path=%s\n", r.URL.Path)
	})

	fmt.Println("ServeMux 注册完成 (不启动服务器)")
	fmt.Println("规则: 最长路径匹配，/结尾匹配子路径")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 路由优先级
// ----------------------------------------------------------
// Gin/httprouter 的优先级规则:
//
//   1. 静态路由 > 参数路由 > 通配路由
//      /user/profile  优先于  /user/:id  优先于  /user/*file
//
//   2. 同层级不允许冲突
//      /user/:id  和  /user/:name  → 冲突！(同一位置两个参数)
//      /user/:id  和  /user/profile → 不冲突（静态优先）
//
//   3. 通配符只能在路径末尾
//      /static/*filepath  ✓
//      /static/*filepath/css  ✗
//
// 面试追问: 为什么 httprouter 不支持同一位置多个参数？
// 答: 因为路由查找是确定性的——给定 URL，必须唯一确定匹配哪条路由。
//     如果允许 /user/:id 和 /user/:name 同时存在，就无法区分。
func priorityDemo() {
	fmt.Println("=== 3. 路由优先级 ===")
	fmt.Println("优先级: 静态路由 > 参数路由(:param) > 通配路由(*wildcard)")
	fmt.Println()
	fmt.Println("合法组合:")
	fmt.Println("  /user/profile     → 静态")
	fmt.Println("  /user/:id         → 参数")
	fmt.Println("  /user/*filepath   → 通配")
	fmt.Println()
	fmt.Println("冲突示例:")
	fmt.Println("  /user/:id 和 /user/:name → 冲突!")
	fmt.Println("  /user/:id/posts 和 /user/:id/comments → 合法(参数后不同路径)")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 路由分组原理
// ----------------------------------------------------------
// Gin 的 RouterGroup 实现原理:
//
//   type RouterGroup struct {
//       Handlers []HandlerFunc  // 中间件链
//       basePath string         // 前缀路径
//       engine   *Engine        // 所属引擎
//   }
//
// Group() 方法只是创建一个新的 RouterGroup，继承 basePath 和 Handlers:
//
//   r := gin.Default()
//   api := r.Group("/api")                    // basePath = "/api"
//   v1 := api.Group("/v1")                    // basePath = "/api/v1"
//   users := v1.Group("/users", authMiddleware) // basePath = "/api/v1/users", 附加中间件
//
// 实际注册路由时，会将 basePath + 路由路径拼接，并将所有中间件合并:
//   users.GET("/:id", getUser)
//   等价于: r.GET("/api/v1/users/:id", authMiddleware, getUser)
func groupDemo() {
	fmt.Println("=== 4. 路由分组原理 ===")

	// 模拟 Gin RouterGroup 的核心逻辑
	type Middleware func()
	type Route struct {
		Method      string
		Path        string
		Middlewares []Middleware
	}

	var routes []Route

	// 模拟分组注册
	registerRoute := func(basePath string, middlewares []Middleware, method, path string, handler Middleware) {
		fullPath := basePath + path
		allMiddlewares := append(middlewares, handler)
		routes = append(routes, Route{
			Method:      method,
			Path:        fullPath,
			Middlewares: allMiddlewares,
		})
	}

	// 模拟: api := r.Group("/api", loggerMiddleware)
	var loggerMW Middleware = func() { fmt.Println("  [logger]") }
	var authMW Middleware = func() { fmt.Println("  [auth]") }

	// api/v1/users 分组
	apiMiddlewares := []Middleware{loggerMW}
	registerRoute("/api", apiMiddlewares, "GET", "/v1/users", func() { fmt.Println("  [listUsers]") })

	// api/v1/admin 分组（附加 auth 中间件）
	adminMiddlewares := append([]Middleware{loggerMW}, authMW)
	registerRoute("/api", adminMiddlewares, "GET", "/v1/admin/dashboard", func() { fmt.Println("  [dashboard]") })

	fmt.Println("注册的路由:")
	for _, r := range routes {
		fmt.Printf("  %s %s (%d handlers)\n", r.Method, r.Path, len(r.Middlewares))
	}
	fmt.Println()
}
