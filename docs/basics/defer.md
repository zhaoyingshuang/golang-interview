---
title: Defer 延迟调用
---

## 1. 执行顺序: LIFO 栈

defer 按后进先出（LIFO）顺序执行。编译时转换为链表，新 defer 插入头部，函数返回时从头部开始执行。

::: tip 使用场景
资源释放顺序很重要——先打开 A 再打开 B，释放时应该先释放 B 再释放 A，defer 天然保证这个顺序。

::: warning 面试追问
如果 defer 和 return 都有逻辑，谁先执行？→ return 的赋值先执行，然后 defer，最后 RET 指令。

```go
fmt.Println("start")
defer fmt.Println("defer 1")  // 最后执行
defer fmt.Println("defer 2")
defer fmt.Println("defer 3")  // 最先执行
fmt.Println("end")
// 输出: start → end → defer 3 → defer 2 → defer 1

// 生产: 资源释放顺序
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close()  // 后开先关

    scanner := bufio.NewScanner(f)
    // ...
    return nil
}
```

## 2. 参数求值时机

**defer 语句声明时参数就被求值了**，不是执行时。但闭包引用的变量在执行时求值。

::: tip 使用场景
在循环中 defer 时经常踩坑——循环变量被闭包捕获，最终都打印最后一个值。Go 1.22 之前需要显式传参解决。

::: warning 面试追问
defer fmt.Println(x) 和 defer func() { fmt.Println(x) }() 的区别？→ 前者捕获 x 的当前值，后者在执行时读取 x 的值。

```go
x := 10
defer fmt.Printf("defer x = %d\n", x)  // 输出 10（声明时求值）
x = 20

// 闭包在执行时求值
y := 10
defer func() {
    fmt.Printf("defer y = %d\n", y)  // 输出 20（执行时求值）
}()
y = 20

// 生产坑: 循环中的 defer
for _, v := range values {
    defer func() {
        process(v)  // Go 1.22前: 全部打印最后一个值
    }()
}
// 修复（Go 1.22前）
for _, v := range values {
    v := v  // 创建新变量
    defer func() { process(v) }()
}
```

## 3. defer 和 return 的执行顺序

完整执行顺序：1) 返回值 = 表达式求值（赋值给返回值变量）→ 2) 执行 defer 函数（LIFO 顺序）→ 3) 执行 RET 指令返回给调用方。关键：defer 可以修改**命名返回值**，因为命名返回值是一个变量，defer 中可以访问和修改它。
:::

**匿名返回值**不受影响（已经赋值给临时变量）。

::: tip 使用场景
函数耗时统计（defer 中记录 time.Since）、panic 恢复并返回错误。

::: warning 面试题
func f() (r int) { defer func() { r = 1 }(); return 0 } 返回什么？→ 返回 1。

```go
// 匿名返回值
func f1() int {
    x := 1
    defer func() { x++ }()
    return x  // 返回 1（x 和返回值无关）
}

// 命名返回值
func f2() (x int) {
    defer func() { x++ }()
    return 1  // 返回 2（defer 修改了返回值 x）
}

// 生产: 函数耗时统计
func timedOperation() (result int, err error) {
    start := time.Now()
    defer func() {
        log.Printf("耗时: %v, err: %v", time.Since(start), err)
    }()
    // ...
    return 42, nil
}

// 生产: panic 恢复并返回错误
func safeCall() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    riskyOperation()
    return nil
}
```

## 4. Open-coded 优化 (Go 1.14+)

Go 1.14 引入 open-coded defer 优化：编译器直接 inline defer 调用，使用 bitmap 记录哪些 defer 需要执行，避免了 deferproc/deferreturn 的调用开销和 _defer 结构体的堆分配。
:::

**触发条件**：defer 数量 ≤ 8、不在循环中、没有 goto、没有 -N 禁用优化。

**性能提升**：从 ~50ns 降到 ~5ns。

::: info 生产影响
大部分 defer 场景（如 Close、Unlock）自动受益，无需改代码。

::: warning 面试追问
循环中的 defer 有什么问题？→ 无法使用 open-coded 优化，仍然走 deferproc 路径，且所有 defer 会堆积到函数返回时才执行（资源延迟释放）。

```go
// 这段代码在 Go 1.14+ 自动使用 open-coded 优化
defer fmt.Println("defer 1")  // 内联，不走 deferproc
defer fmt.Println("defer 2")
defer fmt.Println("defer 3")

// 循环中的 defer → 无法 open-coded（性能差）
for _, f := range files {
    file, _ := os.Open(f)
    defer file.Close()  // 所有文件都打开后才关闭！
}

// 正确: 提取到子函数
for _, f := range files {
    if err := processFile(f); err != nil {
        // handle error
    }
}
func processFile(name string) error {
    f, err := os.Open(name)
    if err != nil { return err }
    defer f.Close()  // 在子函数返回时立即关闭
    // ...
}
```

## 5. recover 捕获 panic

recover() 只能在 defer 函数中**直接调用**才有效。如果 recover() 被嵌套调用（如在 defer 内调用的普通函数中），返回 nil。goroutine 中的 panic 只能自己 recover，否则整个程序崩溃。

::: tip 使用场景
HTTP handler 中 panic 恢复（net/http 默认有 recover）；goroutine 启动时加 recover 防止整个进程崩溃。

::: tip 最佳实践
recover + 命名返回值 模式是 Go 中将 panic 转为 error 的惯用写法。

```go
// 正确: recover 在 defer 中直接调用
defer func() {
    if r := recover(); r != nil {
        log.Printf("recovered: %v\n%s", r, debug.Stack())
    }
}()

// 错误: recover 不在 defer 中直接调用
defer func() {
    doRecover()  // recover 返回 nil，无法捕获！
}()
func doRecover() { recover() }

// 生产: goroutine 中必须自己 recover
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("goroutine panic: %v", r)
        }
    }()
    // 如果不 recover，整个进程会崩溃！
    doWork()
}()

// 生产: HTTP middleware 中 panic 恢复
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```
