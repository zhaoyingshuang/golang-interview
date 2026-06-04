package main

import (
	"fmt"
	"log"
	"time"
)

// ============================================================
// Web 框架中间件机制
// ============================================================
//
// 【面试高频问题】
// 1. 中间件的洋葱模型是什么？
// 2. Gin 的 c.Next() 和 c.Abort() 原理？
// 3. 如何实现跨中间件数据传递？
// 4. 中间件的执行顺序是怎样的？
// 5. 如何实现超时中间件？

func main() {
	onionModel()
	simulateGinMiddleware()
	dataPassing()
	commonMiddleware()
}

// ----------------------------------------------------------
// 1. 洋葱模型原理
// ----------------------------------------------------------
// 中间件的核心思想是"洋葱模型"：
//
//   请求 → [Middleware1 Before] → [Middleware2 Before] → [Handler]
//   响应 ← [Middleware1 After ] ← [Middleware2 After ] ← [Handler]
//
// 本质是一个递归调用链:
//
//   func Middleware1(next Handler) Handler {
//       return func(c *Context) {
//           // Before
//           next(c)
//           // After
//       }
//   }
//
// 面试追问: 为什么叫洋葱模型？
// 答: 因为请求从外层进入，穿过所有中间件到达 handler，然后响应从内层穿出，
//     像穿透洋葱的每一层。
func onionModel() {
	fmt.Println("=== 1. 洋葱模型 ===")

	// 用闭包链模拟洋葱模型
	type Handler func()

	middleware := func(name string, wrap func()) func(Handler) Handler {
		return func(next Handler) Handler {
			return func() {
				fmt.Printf("  [%s] Before\n", name)
				if wrap != nil {
					wrap()
				}
				next()
				fmt.Printf("  [%s] After\n", name)
			}
		}
	}

	// 构建中间件链
	var final Handler = func() {
		fmt.Println("  [Handler] 处理请求")
	}

	m1 := middleware("Logger", nil)
	m2 := middleware("Auth", nil)
	m3 := middleware("Recovery", nil)

	// 链式包装: m1(m2(m3(final)))
	chained := m1(m2(m3(final)))
	chained()
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 模拟 Gin 中间件实现
// ----------------------------------------------------------
// Gin 的中间件核心:
//
//   type Context struct {
//       handlers []HandlerFunc  // 中间件 + 业务handler
//       index    int            // 当前执行到第几个
//   }
//
//   func (c *Context) Next() {
//       c.index++
//       for c.index < len(c.handlers) {
//           c.handlers[c.index](c)
//           c.index++
//       }
//   }
//
//   func (c *Context) Abort() {
//       c.index = abortIndex  // 设为一个很大的值，跳过后续 handler
//   }
func simulateGinMiddleware() {
	fmt.Println("=== 2. Gin 中间件实现模拟 ===")

	type Context struct {
		index    int
		handlers []func(*Context)
		aborted  bool
	}

	next := func(c *Context) {
		c.index++
		for c.index < len(c.handlers) && !c.aborted {
			c.handlers[c.index](c)
			c.index++
		}
	}

	abort := func(c *Context) {
		c.aborted = true
	}

	// 模拟注册中间件 + handler
	handlers := []func(*Context){
		// 中间件1: Logger
		func(c *Context) {
			start := time.Now()
			fmt.Println("  [Logger] 开始")
			next(c)
			fmt.Printf("  [Logger] 完成 (%v)\n", time.Since(start))
		},
		// 中间件2: Auth
		func(c *Context) {
			fmt.Println("  [Auth] 检查认证")
			// 模拟认证失败场景
			// abort(c)
			// return
			next(c)
		},
		// 业务 Handler
		func(c *Context) {
			fmt.Println("  [Handler] 返回用户数据")
		},
	}

	c := &Context{index: 0, handlers: handlers}
	c.handlers[0](c) // 从第一个中间件开始
	fmt.Println()

	// 演示 Abort
	fmt.Println("--- 演示 Abort (Auth 失败) ---")
	handlers2 := []func(*Context){
		func(c *Context) {
			fmt.Println("  [Logger] 开始")
			next(c)
			fmt.Println("  [Logger] 完成")
		},
		func(c *Context) {
			fmt.Println("  [Auth] 认证失败!")
			abort(c) // 后续 handler 不会执行
		},
		func(c *Context) {
			fmt.Println("  [Handler] 这不会执行")
		},
	}

	c2 := &Context{index: 0, handlers: handlers2}
	c2.handlers[0](c2)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 跨中间件数据传递
// ----------------------------------------------------------
// Gin 中通过 c.Set(key, value) / c.Get(key) 传递数据
// 底层是 map[string]any
//
// 注意: Context 是会被复用的！请求结束后必须清理自定义数据。
// 实际上 Gin 的 Context 有 sync.Pool 复用机制。
func dataPassing() {
	fmt.Println("=== 3. 跨中间件数据传递 ===")

	type Context struct {
		keys map[string]any
	}

	set := func(c *Context, key string, value any) {
		if c.keys == nil {
			c.keys = make(map[string]any)
		}
		c.keys[key] = value
	}

	get := func(c *Context, key string) (any, bool) {
		v, ok := c.keys[key]
		return v, ok
	}

	c := &Context{}
	set(c, "userID", 42)
	set(c, "requestID", "req-abc-123")

	if uid, ok := get(c, "userID"); ok {
		fmt.Printf("  userID = %v\n", uid)
	}
	if rid, ok := get(c, "requestID"); ok {
		fmt.Printf("  requestID = %v\n", rid)
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 常用中间件模式
// ----------------------------------------------------------
func commonMiddleware() {
	fmt.Println("=== 4. 常用中间件模式 ===")

	// Recovery 中间件
	fmt.Println("Recovery 中间件:")
	fmt.Println("  defer func() {")
	fmt.Println("    if err := recover(); err != nil {")
	fmt.Println("      log.Printf(\"panic recovered: %v\", err)")
	fmt.Println("      c.JSON(500, gin.H{\"error\": \"Internal Server Error\"})")
	fmt.Println("    }")
	fmt.Println("  }()")
	fmt.Println()

	// CORS 中间件
	fmt.Println("CORS 中间件:")
	fmt.Println("  c.Header(\"Access-Control-Allow-Origin\", \"*\")")
	fmt.Println("  c.Header(\"Access-Control-Allow-Methods\", \"GET,POST,PUT,DELETE\")")
	fmt.Println("  c.Header(\"Access-Control-Allow-Headers\", \"Content-Type,Authorization\")")
	fmt.Println("  if c.Request.Method == \"OPTIONS\" {")
	fmt.Println("    c.AbortWithStatus(204)")
	fmt.Println("    return")
	fmt.Println("  }")
	fmt.Println()

	// RequestID 中间件
	fmt.Println("RequestID 中间件:")
	fmt.Println("  requestID := uuid.NewString()")
	fmt.Println("  c.Set(\"requestID\", requestID)")
	fmt.Println("  c.Header(\"X-Request-ID\", requestID)")
	fmt.Println()

	// 超时中间件
	fmt.Println("超时中间件:")
	fmt.Println("  ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)")
	fmt.Println("  defer cancel()")
	fmt.Println("  c.Request = c.Request.WithContext(ctx)")
	fmt.Println("  done := make(chan struct{})")
	fmt.Println("  go func() { c.Next(); close(done) }()")
	fmt.Println("  select {")
	fmt.Println("  case <-done: return")
	fmt.Println("  case <-ctx.Done(): c.JSON(504, gin.H{\"error\": \"timeout\"})")
	fmt.Println("  }")

	_ = log.Println // 避免未使用导入
}
