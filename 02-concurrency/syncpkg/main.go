package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ============================================================
// Sync 原语深度解析
// ============================================================
//
// 【面试高频问题】
// 1. Mutex 的实现？正常模式和饥饿模式？
// 2. RWMutex 的实现？为什么读写锁更重？
// 3. WaitGroup 的实现？
// 4. sync.Once 的实现？为什么能保证只执行一次？
// 5. sync.Pool 的用途和原理？
// 6. atomic 的使用场景？

func main() {
	mutex()
	rwMutex()
	waitGroup()
	once()
	pool()
	atomicOps()
	mapType()
}

// ----------------------------------------------------------
// 1. Mutex
// ----------------------------------------------------------
// runtime/sema.go + sync/mutex.go
//
// 实现演进:
//   Go 1.8 之前: 纯正常模式（FIFO 等待队列）
//   Go 1.9+: 正常模式 + 饥饿模式
//
// 正常模式:
//   - 新来的 goroutine 和刚被唤醒的 goroutine 竞争锁
//   - 新来的 goroutine 有优势（已经在 CPU 上运行）
//   - 刚唤醒的 goroutine 可能竞争失败，被放到队列前面
//   - 如果一个 goroutine 等待超过 1ms，切换到饥饿模式
//
// 饥饿模式:
//   - 锁直接交给队列头部的 goroutine（不竞争）
//   - 新来的 goroutine 直接排队，不尝试获取锁
//   - 当最后一个等待者获取锁，或等待时间 < 1ms，切回正常模式
//
// 关键字段:
//   type Mutex struct {
//       state int32  // 0=未锁定, 1=已锁定, 2=已唤醒, 4=饥饿
//       sema  uint32 // 信号量，用于阻塞/唤醒 goroutine
//   }

func mutex() {
	fmt.Println("=== 1. Mutex ===")

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
	fmt.Printf("并发计数: %d\n", counter)

	// 不可重入！
	// Go 的 Mutex 不是可重入锁（不支持同一个 goroutine 重复 Lock）
	// 以下代码会死锁:
	//   mu.Lock()
	//   mu.Lock() // 死锁！

	fmt.Println("Mutex 要点:")
	fmt.Println("  - 正常模式: 新 goroutine 和等待者竞争（有利于吞吐）")
	fmt.Println("  - 饥饿模式: 等待超过 1ms 时直接交给队列头部（防止饿死）")
	fmt.Println("  - 不可重入: 同一 goroutine 重复 Lock 会死锁")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. RWMutex
// ----------------------------------------------------------
// 读写锁:
//   - 多个读者可以同时持有锁
//   - 写者独占锁
//   - 写者获取锁时，新读者也要等待
//
// 实现:
//   type RWMutex struct {
//       w           Mutex   // 写锁（也用于阻止新读者）
//       writerSem   uint32  // 写者等待信号量
//       readerSem   uint32  // 读者等待信号量
//       readerCount int32   // 当前读者数（负数表示有写者在等待）
//       readerWait  int32   // 写者等待当前读者完成的数量
//   }
//
// 写者获取锁的过程:
//   1. 获取 w（Mutex），阻止后续写者
//   2. readerCount -= rwmutexMaxReaders（使其变为负数，通知新读者等待）
//   3. 如果有活跃读者，等待 readerSem
//
// 读者获取锁的过程:
//   1. readerCount++，如果 readerCount < 0（有写者在等），阻塞在 readerSem
//   2. 读取数据
//   3. readerCount--，如果是最后一个读者且有写者在等，唤醒写者
//
// 为什么 RWMutex 更重？
//   - 内部包含一个 Mutex + 两个信号量 + 两个计数器
//   - 读者获取/释放需要 atomic 操作
//   - 如果读写比例不高，RWMutex 可能比 Mutex 慢

func rwMutex() {
	fmt.Println("=== 2. RWMutex ===")

	var rw sync.RWMutex
	data := map[string]int{"a": 1, "b": 2}
	var wg sync.WaitGroup

	// 多个读者并发
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rw.RLock()
			fmt.Printf("  读者 %d: data=%v\n", id, data)
			rw.RUnlock()
		}(i)
	}

	// 写者
	wg.Add(1)
	go func() {
		defer wg.Done()
		rw.Lock()
		data["c"] = 3
		rw.Unlock()
		fmt.Println("  写者: 添加 c=3")
	}()

	wg.Wait()
	fmt.Println("RWMutex 适用: 读多写少，且临界区较长")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. WaitGroup
// ----------------------------------------------------------
// type WaitGroup struct {
//     noCopy noCopy
//     state1 uint64  // 高 32 位: 计数器, 低 32 位: 等待者数量
//     state2 uint32  // 信号量
// }
//
// Add(delta): 计数器 += delta
// Done():    计数器 -= 1（等价于 Add(-1)）
// Wait():    等待计数器归零
//
// 注意事项:
//   - Add 的计数不能变负（panic）
//   - Add 应该在 goroutine 外部调用（在 go 之前）
//   - Wait 不能并发调用（虽然实际上可以）
//   - WaitGroup 不能被复制（传指针或用全局变量）

func waitGroup() {
	fmt.Println("=== 3. WaitGroup ===")

	var wg sync.WaitGroup

	// 正确用法: Add 在 goroutine 之前调用
	for i := 0; i < 5; i++ {
		wg.Add(1) // 在这里 Add，不要在 goroutine 内部
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  worker %d done\n", id)
		}(i)
	}
	wg.Wait()
	fmt.Println("所有 worker 完成")

	// 常见错误: Add 放在 goroutine 内部
	// for i := 0; i < 5; i++ {
	//     go func() {
	//         wg.Add(1)    // 错误！可能 wg.Wait() 已经执行了
	//         defer wg.Done()
	//     }()
	// }
	// wg.Wait() // 可能在所有 Add 之前就返回了

	fmt.Println()
}

// ----------------------------------------------------------
// 4. sync.Once
// ----------------------------------------------------------
// type Once struct {
//     done uint32     // 原子标志位
//     m    Mutex      // 互斥锁
// }
//
// 实现原理:
//   1. 原子加载 done，如果为 1，直接返回（快速路径）
//   2. 加锁
//   3. 再次检查 done（double-checking）
//   4. 执行 f()
//   5. 原子存储 done = 1（在 f 执行成功后）
//   6. 释放锁
//
// 关键: done 的赋值在 f() 执行成功之后，
//       所以即使 f() panic，done 也不会被设置为 1。
//       但注意: 如果 f() panic，后续调用会再次执行 f()。
//
// Go 1.21+ sync.OnceValues: 返回值的 once

func once() {
	fmt.Println("=== 4. sync.Once ===")

	var once sync.Once
	callCount := 0

	for i := 0; i < 5; i++ {
		once.Do(func() {
			callCount++
			fmt.Printf("  只执行一次！count=%d\n", callCount)
		})
	}
	fmt.Printf("最终调用次数: %d\n", callCount)

	// Go 1.21+ OnceValues
	// loadConfig := sync.OnceValues(func() (*Config, error) {
	//     return &Config{Port: 8080}, nil
	// })
	// cfg, err := loadConfig()

	fmt.Println()
}

// ----------------------------------------------------------
// 5. sync.Pool
// ----------------------------------------------------------
// 用途: 对象复用，减少 GC 压力
//
// 原理:
//   - 每个 P 有一个本地池（poolLocal）
//   - Get: 先从本地池取，取不到从其他 P 偷，再取不到调用 New 创建
//   - Put: 放入本地池
//   - GC 时会清理池中对象！所以不适合做连接池
//
// 适用场景:
//   - 高频创建和销毁的临时对象（如 bytes.Buffer）
//   - 标准库中的使用: fmt 包、encoding/json、http 包
//
// 不适用:
//   - 连接池（GC 会清理，不适合持久连接）
//   - 需要精确数量控制的对象池

func pool() {
	fmt.Println("=== 5. sync.Pool ===")

	pool := &sync.Pool{
		New: func() any {
			fmt.Println("  创建新对象")
			return make([]byte, 1024)
		},
	}

	// 首次 Get，触发 New
	buf := pool.Get().([]byte)
	fmt.Printf("  获取对象: len=%d\n", len(buf))

	// 放回池中
	pool.Put(buf)

	// 再次 Get，从池中复用
	buf2 := pool.Get().([]byte)
	fmt.Printf("  复用对象: len=%d\n", len(buf2))

	fmt.Println("注意: GC 时池中对象会被清理，不适合做连接池")
	fmt.Println()
}

// ----------------------------------------------------------
// 6. Atomic 操作
// ----------------------------------------------------------
// sync/atomic 提供的原子操作:
//   - Add (T)    : 原子加减
//   - CompareAndSwap (T): CAS 操作
//   - Swap (T)   : 原子交换
//   - Load (T)   : 原子读取
//   - Store (T)  : 原子写入
//
// 底层: 直接使用 CPU 的原子指令（LOCK 前缀 + CMPXCHG 等）
//
// Go 1.19+ atomic.Int64, atomic.Bool 等类型更易用

func atomicOps() {
	fmt.Println("=== 6. Atomic 操作 ===")

	// 原子计数器
	var count atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count.Add(1)
		}()
	}
	wg.Wait()
	fmt.Printf("原子计数: %d\n", count.Load())

	// CAS 操作（无锁更新）
	var value atomic.Int64
	value.Store(10)
	swapped := value.CompareAndSwap(10, 20)
	fmt.Printf("CAS(10→20): swapped=%v, value=%d\n", swapped, value.Load())

	swapped = value.CompareAndSwap(10, 30) // 10 已经不存在了
	fmt.Printf("CAS(10→30): swapped=%v, value=%d\n", swapped, value.Load())

	// atomic.Bool
	var flag atomic.Bool
	flag.Store(true)
	fmt.Printf("flag: %v\n", flag.Load())

	// atomic.Pointer[T] (Go 1.19+)
	var p atomic.Pointer[string]
	s := "hello"
	p.Store(&s)
	fmt.Printf("atomic pointer: %s\n", *p.Load())

	fmt.Println()
}

// ----------------------------------------------------------
// 7. sync.Map
// ----------------------------------------------------------
// 并发安全的 map，适用于:
//   1. key 一旦写入很少变化
//   2. 多个 goroutine 读写不同的 key
//
// 不适用:
//   - 频繁写入的场景（比 Mutex + map 慢）
//
// 原理:
//   - 读操作: 先读 read map（原子操作，无锁）
//   - 写操作: 写入 dirty map（加锁）
//   - 当 read map 未命中次数达到阈值，将 dirty 提升为 read
//   - 使用 readonly 结构包装 map，实现无锁读

func mapType() {
	fmt.Println("=== 7. sync.Map ===")

	var m sync.Map

	// 写入
	m.Store("name", "Go")
	m.Store("version", "1.22")

	// 读取
	v, ok := m.Load("name")
	fmt.Printf("Load: %v, ok=%v\n", v, ok)

	// LoadOrStore: 存在就返回，不存在就存入
	actual, loaded := m.LoadOrStore("name", "Rust")
	fmt.Printf("LoadOrStore: actual=%v, loaded=%v\n", actual, loaded)

	// Range 遍历
	m.Range(func(key, value any) bool {
		fmt.Printf("  %v = %v\n", key, value)
		return true
	})

	// 删除
	m.Delete("version")
	_, ok = m.Load("version")
	fmt.Printf("Delete 后: ok=%v\n", ok)

	fmt.Println("sync.Map 适用: 读多写少、不同 goroutine 读写不同 key")
	fmt.Println()
}
