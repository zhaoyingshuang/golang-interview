package main

import (
	"fmt"
	"sync"
)

// ============================================================
// Defer 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. defer 的执行顺序？
// 2. defer 中参数的求值时机？
// 3. defer 和 return 的执行顺序？
// 4. defer 的性能开销？
// 5. defer 在循环中使用的问题？
// 6. defer 如何影响命名返回值？

func main() {
	executionOrder()
	argumentEvaluation()
	deferAndReturn()
	namedReturn()
	loopDefer()
	deferRecover()
	openCodedDefer()
}

// ----------------------------------------------------------
// 1. 执行顺序（LIFO 栈）
// ----------------------------------------------------------
// defer 按照后进先出（LIFO）的顺序执行。
// 原因: defer 在编译时会被转换成链表，新 defer 插入链表头部，
// 函数返回时从链表头部开始执行。
func executionOrder() {
	fmt.Println("=== 1. 执行顺序 ===")

	fmt.Println("start")
	defer fmt.Println("defer 1")
	defer fmt.Println("defer 2")
	defer fmt.Println("defer 3")
	fmt.Println("end")
	// 输出: start → end → defer 3 → defer 2 → defer 1
}

// ----------------------------------------------------------
// 2. 参数求值时机
// ----------------------------------------------------------
// defer 语句声明时，参数就已经被求值了！
// 不是在 defer 实际执行时才求值。
func argumentEvaluation() {
	fmt.Println("\n=== 2. 参数求值时机 ===")

	x := 10
	defer fmt.Printf("defer x = %d (声明时求值)\n", x)

	x = 20
	fmt.Printf("x = %d (修改后)\n", x)

	// 输出: defer x = 10 (声明时求值)
	// 不是 20！

	// 闭包的情况不同:
	y := 10
	defer func() {
		fmt.Printf("defer 闭包 y = %d (执行时求值)\n", y)
	}()
	y = 20
	// 输出: defer 闭包 y = 20 (执行时求值)
}

// ----------------------------------------------------------
// 3. defer 和 return 的执行顺序
// ----------------------------------------------------------
// 函数返回的过程（带 defer）:
//   1. 返回值 = 表达式求值（赋值给返回值）
//   2. 执行 defer 函数（LIFO 顺序）
//   3. 执行 RET 指令，返回给调用方
//
// 所以 defer 可以修改返回值（如果是命名返回值）。
func deferAndReturn() {
	fmt.Println("\n=== 3. defer 和 return 顺序 ===")

	fmt.Printf("func1 返回: %d\n", func1())
	fmt.Printf("func2 返回: %d\n", func2())
}

func func1() int {
	// 执行顺序:
	// 1. x = 1 (返回值赋值)
	// 2. defer: x++ → x = 2 (但返回值已经是 1 的拷贝)
	// 3. return 1
	x := 1
	defer func() {
		x++
		fmt.Printf("  func1 defer: x = %d\n", x)
	}()
	return x // 返回 1，不是 2
}

func func2() (x int) {
	// 命名返回值 x
	// 执行顺序:
	// 1. x = 1 (返回值赋值)
	// 2. defer: x++ → x = 2 (修改的就是返回值 x)
	// 3. return x → 返回 2
	defer func() {
		x++
		fmt.Printf("  func2 defer: x = %d\n", x)
	}()
	return 1 // 返回 2！
}

// ----------------------------------------------------------
// 4. 命名返回值陷阱
// ----------------------------------------------------------
func namedReturn() {
	fmt.Println("\n=== 4. 命名返回值 ===")

	// 经典面试题: 以下函数返回什么？
	result := namedReturnFunc()
	fmt.Printf("namedReturnFunc 返回: %d\n", result) // 1, 不是 0

	// 另一个经典: recover 只在 defer 中直接调用才有效
	fmt.Printf("recoverFunc 返回: %v\n", recoverFunc())
}

func namedReturnFunc() (r int) {
	defer func() {
		r = 1 // 修改命名返回值
	}()
	return 0 // 看起来返回 0，实际返回 1
}

func recoverFunc() (err error) {
	// 命名返回值 + recover 模式
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	panic("something wrong")
}

// ----------------------------------------------------------
// 5. 循环中的 defer
// ----------------------------------------------------------
// 常见错误: 在循环中使用 defer，导致资源延迟释放。
// defer 要等到函数返回才执行，不是循环迭代结束。
func loopDefer() {
	fmt.Println("\n=== 5. 循环中的 defer ===")

	// 错误用法:
	// for _, file := range files {
	//     f, _ := os.Open(file)
	//     defer f.Close() // 所有文件都打开后才关闭！
	// }
	// 如果循环 1000 次，会同时打开 1000 个文件

	// 正确用法: 提取到子函数中
	process := func(i int) {
		// defer 在这个函数返回时执行
		defer func() {
			fmt.Printf("  关闭资源 %d\n", i)
		}()
		fmt.Printf("  处理 %d\n", i)
	}

	for i := 0; i < 3; i++ {
		process(i) // 每次迭代都会释放资源
	}
	fmt.Println("循环结束")
}

// ----------------------------------------------------------
// 6. defer + recover 捕获 panic
// ----------------------------------------------------------
// recover() 只能在 defer 函数中直接调用才有效。
// 如果 recover() 被嵌套调用（如在一个普通函数中调用），返回 nil。
func deferRecover() {
	fmt.Println("\n=== 6. recover ===")

	// 正确: recover 在 defer 中直接调用
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  恢复 panic: %v\n", r)
			}
		}()
		panic("oops!")
	}()

	// 错误: recover 不在 defer 中直接调用
	wrongRecover()

	// goroutine 中的 panic 只能自己 recover
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  goroutine 恢复: %v\n", r)
			}
		}()
		panic("goroutine panic")
	}()
	wg.Wait()
}

func wrongRecover() {
	// recover() 在 deferred 函数的嵌套函数中调用，无效
	defer func() {
		doRecover() // recover 返回 nil，无法捕获 panic
	}()
}

func doRecover() {
	r := recover()
	fmt.Printf("  嵌套 recover: %v (nil，没有捕获到)\n", r)
}

// ----------------------------------------------------------
// 7. Open-coded defer (Go 1.14+)
// ----------------------------------------------------------
// Go 1.14 引入了 open-coded defer 优化:
//
// 传统 defer:
//   编译器插入 runtime.deferproc() 和 runtime.deferreturn()
//   每次 defer 都要分配 _defer 结构体（堆分配），链入 goroutine 的 defer 链表
//   函数返回时遍历链表执行
//
// Open-coded defer:
//   编译器直接在函数体内 inline defer 调用
//   使用一个 bitmap 记录哪些 defer 需要执行
//   避免了 deferproc/deferreturn 的调用开销
//
// 条件:
//   - 函数内 defer 数量 <= 8
//   - defer 不在循环中
//   - 没有 goto 跳过 defer
//   - 编译参数 -N（禁用优化）时关闭
//
// 性能提升: 简单场景下 defer 开销从 ~50ns 降到 ~5ns

func openCodedDefer() {
	fmt.Println("\n=== 7. Open-coded defer ===")

	// 这段代码在 Go 1.14+ 会使用 open-coded 优化
	// 不再有 deferproc/deferreturn 的开销
	defer fmt.Println("  open-coded defer 1")
	defer fmt.Println("  open-coded defer 2")
	defer fmt.Println("  open-coded defer 3")

	fmt.Println("  函数体执行")
	fmt.Println("Go 1.14+ 的 open-coded defer 大幅降低了 defer 开销")
}
