---
title: Interface 接口
---

## 1. 底层结构: eface 和 iface

**空接口**（interface{}/any）：eface{_type *_type; data unsafe.Pointer}。

**非空接口**（有方法）：iface{tab *itab; data unsafe.Pointer}。itab 缓存了接口方法到实际方法的映射，全局只生成一份（用 map[interfacetype+type]*itab 做缓存）。接口变量占 16 字节（type 指针 + data 指针）。

**data 指针**：小值（<=指针大小）直接内联存储在 data 字段中，大值在堆上分配后 data 指向它。

::: tip 使用场景
fmt.Println 的参数就是 ...any（eface）；io.Reader 是非空接口（iface）。

::: warning 面试追问
接口方法调用比直接调用慢多少？→ 约 20-50ns，因为无法内联。

```go
// eface: 空接口（没有方法）
type eface struct {
    _type *_type           // 类型元数据
    data  unsafe.Pointer   // 数据指针
}

// iface: 非空接口（有方法）
type iface struct {
    tab  *itab             // 接口类型 + 方法表
    data unsafe.Pointer    // 数据指针
}

// itab: 方法查找表（全局缓存）
type itab struct {
    inter *interfacetype   // 接口类型
    _type *_type           // 实际类型
    hash  uint32           // 用于类型断言快速匹配
    fun   [1]uintptr       // 变长: 方法地址数组
}

// 面试追问: 接口变量占多少内存？
// 16 bytes (64位系统): type指针 + data指针
```

## 2. nil 陷阱（经典面试题）

**interface == nil 当且仅当 type 和 data 都是 nil**。这是 Go 面试最常考的陷阱之一。var p *Dog = nil; var s Speaker = p → s != nil！因为 type 是 *Dog（不是 nil），data 是 nil。调用 s.Speak() 会 panic（nil 指针解引用）。

::: tip 使用场景
函数返回 error 接口时，如果内部返回了 (*CustomError)(nil)，调用方检查 err != nil 会得到 true，导致逻辑错误。

::: tip 正确做法
函数需要返回 nil 时，显式 return nil；不要返回具体类型的 nil 指针。

```go
// 真正的 nil interface
var s1 Speaker       // s1 == nil ✓ (type=nil, data=nil)

// 看起来是 nil 其实不是！
var p *Dog = nil
var s2 Speaker = p   // s2 != nil！
// s2 = iface{type:*Dog, data:nil}

// 经典坑: 函数返回 error
func getError() error {
    var err *MyError = nil
    return err  // 返回 iface{type:*MyError, data:nil}
}
e := getError()
fmt.Println(e == nil) // false！

// 正确写法
func getError() error {
    return nil  // 直接返回 nil interface
}
```

## 3. 类型断言与 type switch

类型断言 x.(T) 用于从接口中提取具体类型的值。
:::

**两种形式**：1) d := s.(Dog) — 类型不匹配会 panic；2) d, ok := s.(Dog) — 安全形式，不匹配时 ok=false。

**type switch** 是更优雅的多类型判断：switch v := s.(type) { case Dog: ... }，v 在每个 case 中自动转换为对应类型。

::: tip 使用场景
解析 JSON 时 interface{} 需要类型断言；错误处理中区分不同错误类型；middleware 中提取请求上下文。

::: warning 面试追问
类型断言的性能开销？→ O(1)，直接查 itab.fun 表。

```go
// 安全的类型断言（生产推荐）
if c, ok := s.(*Cat); ok {
    fmt.Println(c.Name)
}

// type switch（多类型判断，优雅）
switch v := s.(type) {
case Dog:
    fmt.Printf("Dog{%s}\n", v.Name)
case *Cat:
    fmt.Printf("*Cat{%s}\n", v.Name)
default:
    fmt.Printf("unknown (%T)\n", v)
}

// 生产: JSON 解析后类型断言
var data any = json.Unmarshal(...)
switch v := data.(type) {
case map[string]any:
    // object
case []any:
    // array
case string:
    // string
}

// 错误类型判断（生产常用）
var netErr net.Error
if errors.As(err, &netErr) && netErr.Timeout() {
    // 处理超时
}
```

## 4. 方法集规则与设计原则

**值接收者 (T)** → T 和 *T 都能实现接口。
:::

**指针接收者 (*T)** → 只有 *T 能实现接口。原因：Go 可以自动取地址（T → &T），但不能自动解引用（*T → T）。有个限制：如果值不可寻址（如 map 元素、字面量），不能自动取地址。

**小接口原则**：Go 标准库推崇接口越小越好——io.Reader 只有 1 个方法，error 只有 1 个方法。

::: tip 生产建议
使用者定义接口，不是实现者（Go 的隐式实现使这成为可能）；接口通常在消费方定义，而不是在提供方。

::: warning 面试追问
什么时候用指针接收者？→ 需要修改状态、结构体较大、一致性。

```go
type Animal struct{ Name string }
func (a Animal) Speak() string  { return a.Name }     // 值接收者
func (a *Animal) ChangeName(n string) { a.Name = n }  // 指针接收者

type Changer interface{ ChangeName(string) }
// var c Changer = Animal{}  // 编译错误！值不能实现指针接收者的方法
var c Changer = &Animal{}   // ✓ 指针可以

// 小接口原则（生产推荐）
type Storer interface {
    Get(key string) (any, error)
}
// 而不是把所有方法塞进一个接口

// 使用者定义接口（Go 最佳实践）
// 在消费方:
type UserGetter interface {
    GetUser(id int) (*User, error)
}
// 而不是在 model 包里定义 UserGetter
```
