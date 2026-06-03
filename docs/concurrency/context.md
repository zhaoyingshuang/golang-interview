---
title: Context 上下文
---

## 1. Context 接口

Context 是一个接口：Deadline() 返回截止时间，Done() 返回取消信号 channel，Err() 返回取消原因，Value() 返回请求作用域的值。4 种内部实现：**emptyCtx**（根 context，永远不会取消）、**cancelCtx**（可取消）、**timerCtx**（带超时，内部包含 cancelCtx）、**valueCtx**（带值，链表结构）。

::: tip 使用场景
每个 HTTP 请求都有自己的 context（req.Context()），请求结束后自动取消。

::: warning 面试追问
context.Background() 和 context.TODO() 的区别？→ Background 是根 context，TODO 是不确定该用什么时的占位。

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}

// 4 种实现:
// emptyCtx  → context.Background() / TODO()
// cancelCtx → context.WithCancel()
// timerCtx  → context.WithTimeout() / WithDeadline()
// valueCtx  → context.WithValue()

// 面试: context 作为函数参数的规范？
// 必须是第一个参数，不要放在 struct 中
func Process(ctx context.Context, data string) error {
    // ...
}
```

## 2. 取消传播

**取消是单向传播的**：父取消 → 子自动取消；子取消 → 不影响父和兄弟。cancel() 的内部流程：1) 设置 err；2) 关闭 done channel；3) 遍历 children 依次取消；4) 从父的 children 中移除自己。

::: tip 使用场景
用户取消请求 → 所有下游数据库查询、RPC 调用自动取消。

::: tip 最佳实践
cancel() 必须 defer 调用，即使确认会超时也要 cancel（释放内部资源）。

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 创建子 context
child1, cancel1 := context.WithCancel(ctx)
defer cancel1()
child2, _ := context.WithCancel(ctx)
grandchild, _ := context.WithCancel(child1)

cancel1()
// child1.Err() → context.Canceled
// grandchild.Err() → context.Canceled（子被取消）
// child2.Err() → nil（兄弟不受影响）

// 生产: HTTP handler 中
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // 请求结束后自动取消
    result, err := fetchData(ctx, query)
    // 如果用户断开连接，ctx 自动取消
}
```

## 3. WithTimeout/WithDeadline

**WithTimeout** 从当前时间计时，**WithDeadline** 指定绝对时间点。WithTimeout 就是 WithDeadline(ctx, time.Now().Add(timeout))。
:::

**必须 defer cancel()**：即使已经超时，cancel 也必须调用以释放内部 timer goroutine，否则会 goroutine 泄漏。

::: tip 使用场景
数据库查询超时（通常 3-5s）、RPC 调用超时（通常 1-10s）、HTTP 客户端超时。

::: warning 面试追问
超时时间应该设多少？→ 取决于业务，但要小于上游的超时时间。

```go
// 数据库查询超时
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()  // 必须！防止 goroutine 泄漏
rows, err := db.QueryContext(ctx, "SELECT ...")

// RPC 调用超时
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
resp, err := client.Call(ctx, request)

// 面试: 不调用 cancel 会怎样？
// timer goroutine 会一直运行，直到超时
// 如果 timeout 很长（甚至没有），goroutine 就泄漏了
```

## 4. WithValue 最佳实践

Key 用**自定义类型**（避免冲突），不要用内建类型（string/int）。
:::

**适合传递**：trace ID、auth token、request-scoped logger。

**不适合传递**：数据库连接、业务参数、函数必须的参数（应该用函数参数）。

**生产反模式**：把所有东西都塞进 context → 应该只放请求级别的元数据。

::: warning 面试追问
context.Value 的查找效率？→ O(n)，沿 valueCtx 链表向上查找，层数多了会慢。

```go
// 正确: 自定义 key 类型
type requestIDKey struct{}
type userIDKey struct{}

ctx = context.WithValue(ctx, requestIDKey{}, "req-123")
ctx = context.WithValue(ctx, userIDKey{}, 42)

// 类型安全地获取
id, _ := ctx.Value(requestIDKey{}).(string)
uid, _ := ctx.Value(userIDKey{}).(int)

// 生产: middleware 中设置 trace ID
func TraceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := uuid.New().String()
        ctx := context.WithValue(r.Context(),
            requestIDKey{}, traceID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 面试: 为什么不用 string 做 key？
// 不同包可能用相同的 string，导致冲突
// type myKey string → 也行，但空 struct 更安全
```
