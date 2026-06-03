package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================
// Channel 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. channel 的底层实现？
// 2. 有缓冲和无缓冲 channel 的区别？
// 3. channel 的发送和接收规则？
// 4. select 的行为？
// 5. channel 的常见使用模式？
// 6. channel 和 mutex 该选哪个？

func main() {
	hchanStructure()
	sendRecvRules()
	selectBehavior()
	patterns()
	channelVsMutex()
}

// ----------------------------------------------------------
// 1. 底层结构
// ----------------------------------------------------------
// runtime/chan.go:
//
//   type hchan struct {
//       qcount   uint           // 队列中的元素数量
//       dataqsiz uint           // 环形队列容量（缓冲大小）
//       buf      unsafe.Pointer // 环形队列指针
//       elemsize uint16         // 元素大小
//       closed   uint32         // 是否已关闭
//       timer    *timer         // Go 1.23+ timer 集成
//       elemtype *_type         // 元素类型
//       sendx    uint           // 发送索引
//       recvx    uint           // 接收索引
//       recvq    waitq          // 等待接收的 goroutine 队列
//       sendq    waitq          // 等待发送的 goroutine 队列
//       lock     mutex          // 互斥锁
//   }
//
// 核心要点:
//   - channel 内部就是一个带锁的环形队列
//   - sendx/recvx 是环形队列的读写指针
//   - recvq/sendq 是阻塞等待的 goroutine 链表（sudog 结构）
//   - 所有操作（send/recv/close）都需要加锁
//   - 锁是 runtime 级别的 mutex（不是 sync.Mutex）

func hchanStructure() {
	fmt.Println("=== 1. 底层结构 ===")

	// 无缓冲 channel: dataqsiz = 0
	unbuffered := make(chan int)

	// 有缓冲 channel: dataqsiz = 3
	buffered := make(chan int, 3)

	_ = unbuffered
	_ = buffered

	fmt.Println("hchan 核心字段:")
	fmt.Println("  qcount: 队列中元素数")
	fmt.Println("  dataqsiz: 环形队列容量")
	fmt.Println("  sendx/recvx: 读写指针")
	fmt.Println("  recvq/sendq: 等待的 goroutine 队列")
	fmt.Println("  lock: 互斥锁（所有操作都需要）")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 发送和接收规则
// ----------------------------------------------------------
// 发送 (ch <- x):
//   Case 1: recvq 中有等待的 goroutine → 直接把值发给它（不经过缓冲区）
//   Case 2: 缓冲区有空位 → 放入缓冲区
//   Case 3: 缓冲区满或无缓冲 → 阻塞当前 goroutine，加入 sendq
//
// 接收 (x <- ch):
//   Case 1: sendq 中有等待的 goroutine → 从缓冲区取（如果有）或直接取
//   Case 2: 缓冲区有数据 → 从缓冲区取
//   Case 3: 缓冲区空且无发送者 → 阻塞当前 goroutine，加入 recvq
//
// 关闭 (close(ch)):
//   - 设置 closed = 1
//   - 唤醒 recvq 中所有 goroutine（返回零值）
//   - 唤醒 sendq 中所有 goroutine（panic!）
//
// 面试必记:
//   - 向已关闭的 channel 发送 → panic
//   - 关闭已关闭的 channel → panic
//   - 从已关闭的 channel 接收 → 返回零值 + false
//   - 关闭 nil channel → panic

func sendRecvRules() {
	fmt.Println("=== 2. 发送和接收规则 ===")

	// 无缓冲 channel: 发送和接收必须同时就绪
	ch := make(chan string)

	go func() {
		ch <- "hello" // 阻塞，直到有人接收
		fmt.Println("发送完成")
	}()

	msg := <-ch // 阻塞，直到有人发送
	fmt.Printf("收到: %s\n", msg)
	time.Sleep(time.Millisecond) // 等 goroutine 打印

	// 关闭后的行为
	ch2 := make(chan int, 3)
	ch2 <- 1
	ch2 <- 2
	close(ch2)

	// 从关闭的 channel 接收，直到缓冲区为空
	for {
		v, ok := <-ch2
		if !ok {
			fmt.Println("channel 已关闭且为空")
			break
		}
		fmt.Printf("从关闭 channel 接收: %d\n", v)
	}

	// 重复关闭 → panic
	// close(ch2) // panic: close of closed channel

	// 向关闭 channel 发送 → panic
	// ch2 <- 3 // panic: send on closed channel

	fmt.Println()
}

// ----------------------------------------------------------
// 3. select 行为
// ----------------------------------------------------------
// select 的规则:
//   1. 所有 case 会随机排序（避免饥饿）
//   2. 按顺序评估所有 case（发送/接收是否可以立即进行）
//   3. 如果有多个 case 就绪，随机选一个执行
//   4. 如果没有 case 就绪:
//      a. 有 default → 执行 default
//      b. 没有 default → 阻塞，直到某个 case 就绪
//   5. 空 select（无 case 无 default）→ 永远阻塞

func selectBehavior() {
	fmt.Println("=== 3. select 行为 ===")

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	ch1 <- "one"
	ch2 <- "two"

	// 多个 case 就绪时随机选择
	for i := 0; i < 4; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("  ch1: %s\n", msg)
		case msg := <-ch2:
			fmt.Printf("  ch2: %s\n", msg)
		default:
			fmt.Println("  default (无数据)")
		}
	}

	// 非阻塞接收（default 模式）
	select {
	case v := <-ch1:
		fmt.Printf("  非阻塞接收: %s\n", v)
	default:
		fmt.Println("  无数据可接收")
	}

	// select 实现 timeout
	timeoutCh := make(chan int)
	go func() {
		time.Sleep(50 * time.Millisecond)
		timeoutCh <- 42
	}()

	select {
	case v := <-timeoutCh:
		fmt.Printf("  正常接收: %d\n", v)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  超时！")
	}

	// 空 select: 永远阻塞
	// select {} // 用于让 main goroutine 永不退出

	fmt.Println()
}

// ----------------------------------------------------------
// 4. 常见使用模式
// ----------------------------------------------------------
func patterns() {
	fmt.Println("=== 4. 常见模式 ===")

	// 模式1: 信号量（done channel）
	done := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(done) // 用 close 而不是发送，可以有多个接收者
	}()
	<-done
	fmt.Println("模式1: done channel (close 广播)")

	// 模式2: 限流器（带缓冲 channel）
	sem := make(chan struct{}, 3) // 最多 3 个并发
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}        // 获取令牌
			defer func() { <-sem }() // 释放令牌
			fmt.Printf("  worker %d working\n", id)
		}(i)
	}
	wg.Wait()
	fmt.Println("模式2: 信号量限流")

	// 模式3: 单次执行（once via channel）
	result := make(chan int, 1)
	go func() {
		// 模拟耗时操作
		result <- 42
		close(result)
	}()
	// 第一个接收者得到结果
	if v, ok := <-result; ok {
		fmt.Printf("模式3: 单次结果 %d\n", v)
	}

	// 模式4: 通知退出
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				fmt.Println("模式4: 收到退出信号")
				return
			default:
				// do work
			}
		}
	}()
	close(quit)
	time.Sleep(time.Millisecond)

	fmt.Println()
}

// ----------------------------------------------------------
// 5. Channel vs Mutex
// ----------------------------------------------------------
// Channel: 用于 goroutine 之间通信和同步
//   - "Don't communicate by sharing memory; share memory by communicating"
//   - 更适合: 生产者-消费者、信号通知、超时控制
//
// Mutex: 用于保护共享资源
//   - 更适合: cache、计数器、配置等共享状态的读写
//
// 选择原则:
//   - 传递数据所有权 → channel
//   - 保护共享状态 → mutex
//   - 等待/通知 → channel
//   - 并发安全地读写 map → sync.Map 或 mutex

func channelVsMutex() {
	fmt.Println("=== 5. Channel vs Mutex ===")

	// Mutex 方式: 共享计数器
	var mu sync.Mutex
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("Mutex 计数器: %d\n", counter)

	// Channel 方式: 传递计数请求
	ch3 := make(chan int, 1000)
	go func() {
		count := 0
		for range ch3 {
			count++
		}
		fmt.Printf("Channel 计数器: %d\n", count)
	}()

	for i := 0; i < 1000; i++ {
		ch3 <- 1
	}
	close(ch3)
	time.Sleep(time.Millisecond)

	fmt.Println("\n选择: 传递数据所有权用 channel，保护共享状态用 mutex")
	fmt.Println()
}
