package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================
// 并发模式深度解析
// ============================================================
//
// 【面试高频问题】
// 1. worker pool 模式？
// 2. fan-in / fan-out 模式？
// 3. pipeline 模式？
// 4. errgroup 的使用？
// 5. 如何限制并发数？
// 6. 如何实现超时控制？

func main() {
	workerPool()
	fanInOut()
	pipeline()
	errgroupPattern()
	rateLimiting()
	teardown()
}

// ----------------------------------------------------------
// 1. Worker Pool 模式
// ----------------------------------------------------------
// 场景: 控制并发数量，避免创建过多 goroutine
func workerPool() {
	fmt.Println("=== 1. Worker Pool ===")

	jobs := make(chan int, 100)
	results := make(chan int, 100)

	// 启动 3 个 worker
	var wg sync.WaitGroup
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range jobs { // range 直到 channel 关闭
				fmt.Printf("  worker %d 处理 job %d\n", id, j)
				results <- j * 2
			}
		}(w)
	}

	// 发送任务
	for j := 0; j < 10; j++ {
		jobs <- j
	}
	close(jobs) // 关闭 jobs，通知 worker 退出

	// 等待 worker 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	for r := range results {
		_ = r
	}
	fmt.Println("Worker Pool 完成")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Fan-out / Fan-in 模式
// ----------------------------------------------------------
// Fan-out: 将一个 channel 分发给多个 goroutine 处理
// Fan-in:  将多个 channel 合并到一个 channel
func fanInOut() {
	fmt.Println("=== 2. Fan-out / Fan-in ===")

	// 生成数据
	generate := func(n int) <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			for i := 0; i < n; i++ {
				ch <- i
			}
		}()
		return ch
	}

	// Fan-out: 多个 goroutine 处理同一个输入 channel
	process := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for v := range in {
				time.Sleep(5 * time.Millisecond) // 模拟耗时
				out <- v * v
			}
		}()
		return out
	}

	// Fan-in: 合并多个 channel
	merge := func(chs ...<-chan int) <-chan int {
		out := make(chan int)
		var wg sync.WaitGroup
		for _, ch := range chs {
			wg.Add(1)
			go func(c <-chan int) {
				defer wg.Done()
				for v := range c {
					out <- v
				}
			}(ch)
		}
		go func() {
			wg.Wait()
			close(out)
		}()
		return out
	}

	// 执行: 1 个生产者 → 3 个 worker → 1 个合并
	in := generate(20)
	c1 := process(in) // fan-out
	c2 := process(in)
	c3 := process(in)

	count := 0
	for range merge(c1, c2, c3) { // fan-in
		count++
	}
	fmt.Printf("Fan-out/fan-in 处理了 %d 个结果\n", count)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Pipeline 模式
// ----------------------------------------------------------
// 数据在多个 stage 之间流过，每个 stage 是一个 goroutine
// stage1 → stage2 → stage3 → ... → 结果
func pipeline() {
	fmt.Println("=== 3. Pipeline ===")

	// Stage 1: 生成数据
	gen := func(ctx context.Context, values ...int) <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			for _, v := range values {
				select {
				case ch <- v:
				case <-ctx.Done():
					return
				}
			}
		}()
		return ch
	}

	// Stage 2: 计算
	square := func(ctx context.Context, in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for v := range in {
				select {
				case out <- v * v:
				case <-ctx.Done():
					return
				}
			}
		}()
		return out
	}

	// Stage 3: 过滤
	filter := func(ctx context.Context, in <-chan int, keep func(int) bool) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for v := range in {
				if keep(v) {
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return out
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := filter(ctx, square(ctx, gen(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)),
		func(v int) bool { return v > 20 })

	for v := range p {
		fmt.Printf("  pipeline 结果: %d\n", v) // 25, 36, 49, 64, 81, 100
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 4. Errgroup 模式
// ----------------------------------------------------------
// golang.org/x/sync/errgroup: 管理一组 goroutine，第一个错误取消其他
type errGroup struct {
	wg     sync.WaitGroup
	err    error
	errOnce sync.Once
	cancel context.CancelFunc
}

func newErrGroup() (*errGroup, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	return &errGroup{cancel: cancel}, ctx
}

func (g *errGroup) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(g.cancelCtx()); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *errGroup) cancelCtx() context.Context {
	ctx, _ := context.WithCancel(context.Background())
	return ctx
}

func (g *errGroup) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

func errgroupPattern() {
	fmt.Println("=== 4. Errgroup ===")

	g, _ := newErrGroup()

	tasks := []string{"task1", "task2", "task3"}
	for _, task := range tasks {
		task := task
		g.Go(func(_ context.Context) error {
			if task == "task2" {
				return fmt.Errorf("%s failed", task)
			}
			fmt.Printf("  %s 成功\n", task)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("errgroup: %v\n", err)
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 限流模式
// ----------------------------------------------------------
func rateLimiting() {
	fmt.Println("=== 5. 限流 ===")

	// 方式1: channel 信号量
	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < 6; i++ {
		wg.Add(1)
		sem <- struct{}{} // 获取令牌
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }() // 释放令牌
			fmt.Printf("  [%v] 任务 %d 开始\n", time.Since(start).Round(time.Millisecond), id)
			time.Sleep(100 * time.Millisecond)
		}(i)
	}
	wg.Wait()
	fmt.Printf("信号量限流完成，耗时: %v\n", time.Since(start).Round(time.Millisecond))

	fmt.Println()
}

// ----------------------------------------------------------
// 6. 优雅退出模式（Teardown）
// ----------------------------------------------------------
func teardown() {
	fmt.Println("=== 6. 优雅退出 ===")

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// 模拟服务
	server := func(ctx context.Context) {
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("  服务: 收到退出信号，清理资源...")
				time.Sleep(20 * time.Millisecond) // 模拟清理
				fmt.Println("  服务: 清理完成")
				return
			case <-ticker.C:
				fmt.Println("  服务: 处理请求")
			}
		}
	}

	server(ctx)
	fmt.Println("优雅退出完成")
}
