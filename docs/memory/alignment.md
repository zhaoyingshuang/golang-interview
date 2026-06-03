---
title: 内存对齐
---

## 1. 对齐规则

三条规则：1) 字段按类型的对齐系数对齐（不足的补 padding）；2) struct 整体大小必须是最大对齐系数的倍数；3) 基本类型对齐系数 = 其大小（64位系统：bool=1, int32=4, int64=8）。

::: info 为什么需要内存对齐
CPU 访问对齐的内存更快（一次总线操作），不对齐可能触发硬件异常（某些架构）。

::: info 生产影响
百万个 struct 的微优化可能节省数 MB 内存。

::: warning 面试追问
unsafe.Alignof 和 unsafe.Sizeof 的区别？→ Alignof 返回对齐系数，Sizeof 返回实际大小。

```go
// Bad: 24 bytes（浪费 8 bytes padding）
type Bad struct {
    A bool    // 1 byte + 7 bytes padding
    B int64   // 8 bytes
    C int32   // 4 bytes + 4 bytes padding
}
fmt.Println(unsafe.Sizeof(Bad{}))  // 24

// Good: 16 bytes（零 padding）
type Good struct {
    B int64   // 8 bytes
    C int32   // 4 bytes
    A bool    // 1 byte + 3 bytes padding
}
fmt.Println(unsafe.Sizeof(Good{}))  // 16

// 节省 33% 内存！百万个实例省 8MB
```

## 2. 字段顺序优化

原则：**按字段大小从大到小排列**，减少 padding。使用 **fieldalignment** 工具自动检测和修复（go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest）。

::: tip 使用场景
高并发服务中大量传输的 struct（如 HTTP 请求/响应体、数据库模型）。

::: warning 面试追问
编译器会自动优化字段顺序吗？→ 不会，Go 编译器保持字段声明顺序。

```go
// Bad: 24 bytes
type UserBad struct {
    Active  bool     // 1+7
    Balance float64  // 8
    Age     int32    // 4+4
}

// Good: 16 bytes (省 33%)
type UserGood struct {
    Balance float64  // 8
    Age     int32    // 4
    Active  bool     // 1+3
}

// 工具自动修复:
// fieldalignment -fix ./...

// 生产: 百万用户列表
// Bad:  24MB → Good: 16MB (省 8MB)
// 且缓存命中率更高（更紧凑 = 更多数据在缓存行中)
```

## 3. atomic 对齐与零大小类型

64 位原子操作要求 **8 字节对齐**，32 位系统上不对齐会 panic。atomic 变量应放在 struct **第一个字段**。
:::

**零大小类型 struct{}**：大小为 0，不占内存。用于实现 set（map[T]struct{}）和信号 channel（chan struct{}）。

::: tip 使用场景
原子计数器放 struct 首字段；连接管理用 map[string]struct{} 代替 map[string]bool。

```go
// atomic 变量放第一个字段
type Counter struct {
    count int64  // 第一个字段，保证 8 字节对齐
    flag  bool
}

// 32 位系统上以下代码可能 panic！
type BadCounter struct {
    flag  bool
    count int64  // 可能不对齐到 8 字节
}

// 零大小类型: set 实现
type Set struct {
    m map[string]struct{}
}
func (s *Set) Add(v string) { s.m[v] = struct{}{} }
func (s *Set) Has(v string) bool {
    _, ok := s.m[v]; return ok
}

// 信号 channel（不传数据）
done := make(chan struct{})
close(done)  // 广播通知
```
