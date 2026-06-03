package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================
// Context 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. context 的作用？
// 2. context 的类型和继承关系？
// 3. WithCancel/WithTimeout/WithDeadline 的区别？
// 4. context.WithValue 的正确用法？
// 5. context 为什么是接口而不是 struct？
// 6. context 的取消传播机制？

func main() {
	cancelContext()
	timeoutContext()
	valueContext()
	propagation()
	bestPractices()
}

// ----------------------------------------------------------
// 1. WithCancel — 取消传播
// ----------------------------------------------------------
// context.Context 接口:
//   type Context interface {
//       Deadline() (deadline time.Time, ok bool) // 返回截止时间
//       Done() <-chan struct{}                   // 返回关闭信号 channel
//       Err() error                              // 返回取消原因
//       Value(key any) any                       // 返回请求作用域的值
//   }
//
// 内部实现有 4 种:
//   - emptyCtx: 根 context（background/todo），永远不会取消
//   - cancelCtx: 可取消的 context
//   - timerCtx: 带超时的 context（内部包含 cancelCtx）
//   - valueCtx: 带值的 context（链表结构）

func cancelContext() {
	fmt.Println("=== 1. WithCancel ===")

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// 启动 3 个 worker，都能感知取消
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(ctx, id)
		}(i)
	}

	time.Sleep(50 * time.Millisecond)
	fmt.Println("主 goroutine: 取消所有 worker")
	cancel() // 取消所有子 context

	wg.Wait()
	fmt.Println()
}

func worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  worker %d: 退出, err=%v\n", id, ctx.Err())
			return
		default:
			fmt.Printf("  worker %d: 工作中...\n", id)
			time.Sleep(20 * time.Millisecond)
		}
	}
}

// ----------------------------------------------------------
// 2. WithTimeout / WithDeadline
// ----------------------------------------------------------
// WithTimeout(ctx, timeout):  从当前时间开始计时 timeout
// WithDeadline(ctx, deadline): 指定一个绝对时间点
//
// 内部实现: WithTimeout 就是 WithDeadline(ctx, time.Now().Add(timeout))
//
// 取消传播:
//   - 父 context 取消 → 子 context 自动取消
//   - 子 context 取消 → 不影响父 context
//   - 多个子 context 互不影响

func timeoutContext() {
	fmt.Println("=== 2. WithTimeout / WithDeadline ===")

	// WithTimeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel() // 必须调用 cancel！即使已经超时，也要释放资源

	select {
	case <-time.After(200 * time.Millisecond):
		fmt.Println("操作完成")
	case <-ctx.Done():
		fmt.Printf("超时: %v\n", ctx.Err()) // context deadline exceeded
	}

	// WithDeadline
	deadline := time.Now().Add(50 * time.Millisecond)
	ctx2, cancel2 := context.WithDeadline(context.Background(), deadline)
	defer cancel2()

	<-ctx2.Done()
	fmt.Printf("到达截止时间: %v\n", ctx2.Err())

	// 为什么 cancel() 即使超时也要调用？
	// WithTimeout/WithDeadline 内部创建了一个 timer goroutine，
	// cancel() 会停止 timer 并释放资源，避免 goroutine 泄漏。

	fmt.Println()
}

// ----------------------------------------------------------
// 3. WithValue — 请求作用域的值传递
// ----------------------------------------------------------
// 实现原理: valueCtx 是一个链表
//   type valueCtx struct {
//       Context   // 父 context
//       key, val any
//   }
//
// Value(key) 的查找过程: 从当前 ctx 沿链表向上查找，直到找到或到达根节点
//
// 关键规则:
//   - key 必须是可比较的
//   - 不要用内建类型（string/int）作为 key，避免冲突
//   - 应该定义自己的 key 类型

func valueContext() {
	fmt.Println("=== 3. WithValue ===")

	// 正确的 key 定义方式
	type requestIDKey struct{}
	type userIDKey struct{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey{}, "req-123")
	ctx = context.WithValue(ctx, userIDKey{}, 42)

	// 获取值（类型安全）
	reqID, ok := ctx.Value(requestIDKey{}).(string)
	fmt.Printf("requestID: %s, ok=%v\n", reqID, ok)

	userID, ok := ctx.Value(userIDKey{}).(int)
	fmt.Printf("userID: %d, ok=%v\n", userID, ok)

	// 错误用法:
	// ctx = context.WithValue(ctx, "user", 42)           // 用 string 作为 key
	// ctx = context.WithValue(ctx, "user", "different")   // 另一个地方也用 "user"，冲突！

	fmt.Println("WithValue 要点:")
	fmt.Println("  - key 用自定义类型（避免冲突）")
	fmt.Println("  - 不要用于传递可选参数")
	fmt.Println("  - 不要用于传递业务逻辑必须的参数（应该用函数参数）")
	fmt.Println("  - 适合传递请求级别的元数据: trace ID, auth token 等")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 取消传播机制
// ----------------------------------------------------------
// cancelCtx 的内部结构:
//   type cancelCtx struct {
//       Context                      // 父 context
//       mu       sync.Mutex
//       done     atomic.Value        // chan struct{}（懒初始化）
//       children map[canceler]struct{} // 子 context 集合
//       err      error               // 取消原因
//   }
//
// 取消传播:
//   cancel() 被调用时:
//   1. 设置 err
//   2. 关闭 done channel（通知所有等待者）
//   3. 遍历 children，依次取消
//   4. 从父 context 的 children 中移除自己

func propagation() {
	fmt.Println("=== 4. 取消传播 ===")

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// 创建两层子 context
	child1, cancel1 := context.WithCancel(rootCtx)
	defer cancel1()

	child2, cancel2 := context.WithCancel(rootCtx)
	defer cancel2()

	grandchild, cancel3 := context.WithCancel(child1)
	defer cancel3()

	_ = child2
	_ = grandchild

	// 取消 child1 会影响 grandchild，不影响 child2
	cancel1()

	fmt.Printf("child1: %v\n", child1.Err())
	fmt.Printf("grandchild (child1 的子): %v\n", grandchild.Err())
	fmt.Printf("child2 (child1 的兄弟): %v\n", child2.Err())

	fmt.Println("规则: 父取消 → 子取消，子取消 → 不影响父和兄弟")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 最佳实践
// ----------------------------------------------------------
func bestPractices() {
	fmt.Println("=== 5. 最佳实践 ===")

	fmt.Println("1. context 作为函数第一个参数，不要放在 struct 中")
	fmt.Println("   func Process(ctx context.Context, data string) error")
	fmt.Println()
	fmt.Println("2. 不要传递 nil context，用 context.TODO()")
	fmt.Println()
	fmt.Println("3. WithTimeout/WithDeadline 的 cancel 一定要 defer")
	fmt.Println("   ctx, cancel := context.WithTimeout(ctx, timeout)")
	fmt.Println("   defer cancel() // 防止 goroutine 泄漏")
	fmt.Println()
	fmt.Println("4. 不要用 WithValue 传递业务逻辑参数")
	fmt.Println("   好: trace ID, auth token, request-scoped logger")
	fmt.Println("   坏: 数据库连接, 配置参数, 函数返回值")
	fmt.Println()
	fmt.Println("5. 短生命周期操作用 WithTimeout")
	fmt.Println("   长生命周期操作用 WithCancel + 手动取消")
	fmt.Println()
	fmt.Println("6. HTTP 请求的 context:")
	fmt.Println("   req.Context() 在请求结束后自动取消")
	fmt.Println("   handler 中应该使用 req.Context()，不是 context.Background()")
}
