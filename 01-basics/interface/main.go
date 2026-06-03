package main

import (
	"fmt"
	"reflect"
)

// ============================================================
// Interface 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. interface 的底层实现？（eface 和 iface）
// 2. nil interface 是什么？为什么 "nil != nil"？
// 3. 类型断言和类型开关（type switch）？
// 4. 空接口 interface{} 和 any 的区别？
// 5. 接口的隐式实现（鸭子类型）？
// 6. iface 的 itab 是怎么缓存和复用的？

func main() {
	efaceAndIface()
	nilTrap()
	typeAssertion()
	dynamicDispatch()
	composition()
	emptyInterface()
}

// ----------------------------------------------------------
// 1. 底层结构: eface 和 iface
// ----------------------------------------------------------
// runtime/runtime2.go:
//
// 空接口（没有方法）:
//   type eface struct {
//       _type *_type           // 类型元数据
//       data  unsafe.Pointer   // 数据指针
//   }
//
// 非空接口（有方法）:
//   type iface struct {
//       tab  *itab             // 接口类型信息 + 实际类型的方法表
//       data unsafe.Pointer    // 数据指针
//   }
//
// itab 结构:
//   type itab struct {
//       inter *interfacetype   // 接口类型
//       _type *_type           // 实际类型
//       hash  uint32           // _type.hash 的拷贝，用于类型断言快速匹配
//       _     [4]byte
//       fun   [1]uintptr       // 变长数组，存储接口方法对应的实际方法地址
//   }
//
// 关键点:
//   - interface 变量占 16 字节（64位系统）：一个 type 指针 + 一个 data 指针
//   - itab 是全局缓存的：每种 (接口类型, 实际类型) 组合只生成一份
//   - data 指针指向堆上的数据（如果值可以被内联到指针中，则直接存储）

type Speaker interface {
	Speak() string
}

type Dog struct {
	Name string
}

func (d Dog) Speak() string { return d.Name + ": Wang!" }

type Cat struct {
	Name string
}

func (c *Cat) Speak() string { return c.Name + ": Miao!" }

func efaceAndIface() {
	fmt.Println("=== 1. eface 和 iface ===")

	// eface: 空接口
	var e interface{} = "hello"
	fmt.Printf("空接口: type=%T, value=%v\n", e, e)

	// iface: 非空接口
	var s Speaker = Dog{Name: "Wangcai"}
	fmt.Printf("非空接口: type=%T, value=%v\n", s, s)

	// 接口值的比较:
	// 两个接口值相等 ⟺ type 相同 且 data 相同（使用 == 比较 data）
	s1 := Speaker(Dog{"A"})
	s2 := Speaker(Dog{"A"})
	fmt.Printf("相同内容的接口比较: %v\n", s1 == s2) // true

	s3 := Speaker(&Cat{"A"})
	s4 := Speaker(&Cat{"A"})
	fmt.Printf("相同内容的指针接口比较: %v\n", s3 == s4) // false（指针不同）
	fmt.Println()
}

// ----------------------------------------------------------
// 2. nil 陷阱（经典面试题）
// ----------------------------------------------------------
// interface 的 nil 判断:
//   只有当 type 和 data 都是 nil 时，interface == nil 才为 true。
//
// 最常见的坑:
//   var p *Dog = nil
//   var s Speaker = p
//   fmt.Println(s == nil) // false！
//
// 为什么？因为 s 的 type = *Dog（不是 nil），data = nil。
// 只有两个字段都是 nil 才算 nil interface。

func nilTrap() {
	fmt.Println("=== 2. nil 陷阱 ===")

	// 真正的 nil interface
	var s1 Speaker
	fmt.Printf("var s1 Speaker: nil=%v, type=%T\n", s1 == nil, s1)

	// 看起来是 nil，其实不是
	var p *Dog = nil
	var s2 Speaker = p
	fmt.Printf("var s2 Speaker = (*Dog)(nil): nil=%v, type=%T\n", s2 == nil, s2)
	// s2 != nil 因为 type 字段是 *Dog，不是 nil

	// 调用 s2.Speak() 会怎样？
	// s2.Speak() // panic: nil pointer dereference
	// 因为 data 是 nil，Speak 方法中的 c.Name 会解引用 nil 指针

	// 正确的 nil 检查方式:
	// 方式1: 不要把可能为 nil 的具体类型赋给接口
	// 方式2: 在赋值前检查
	//   if p != nil { s = p }
	// 方式3: 函数返回接口时，显式返回 nil（不要返回 *T 类型的 nil）

	// 另一个例子: 函数返回接口
	result := getSpeaker()
	fmt.Printf("函数返回: nil=%v, type=%T\n", result == nil, result)
	fmt.Println()
}

func getSpeaker() Speaker {
	var p *Dog = nil
	return p // 返回的是 iface{type:*Dog, data:nil}，不是 nil！
}

// ----------------------------------------------------------
// 3. 类型断言
// ----------------------------------------------------------
func typeAssertion() {
	fmt.Println("=== 3. 类型断言 ===")

	var s Speaker = Dog{Name: "Wangcai"}

	// 类型断言: x.(T)
	// 如果 s 的动态类型是 Dog，提取值；否则 panic
	d := s.(Dog)
	fmt.Printf("类型断言: %v\n", d)

	// comma-ok 模式（安全的类型断言）
	if c, ok := s.(*Cat); ok {
		fmt.Printf("是 *Cat: %v\n", c)
	} else {
		fmt.Println("不是 *Cat")
	}

	// type switch（更优雅的多类型判断）
	whoAmI(s)
	whoAmI(&Cat{Name: "Kitty"})

	fmt.Println()
}

func whoAmI(s Speaker) {
	switch v := s.(type) {
	case Dog:
		fmt.Printf("  type switch: Dog{%s}\n", v.Name)
	case *Cat:
		fmt.Printf("  type switch: *Cat{%s}\n", v.Name)
	default:
		fmt.Printf("  type switch: unknown (%T)\n", v)
	}
}

// ----------------------------------------------------------
// 4. 动态分发
// ----------------------------------------------------------
// Go 的接口方法调用是动态分发的:
//   s.Speak()
// 编译器会在 itab.fun 中查找方法地址，然后间接调用。
// 这比直接调用（静态分发）慢约 20-50ns。
//
// 但 Go 的 itab 是全局缓存的，查找开销很小（只是一次数组索引访问）。

type Animal interface {
	Sound() string
	Move() string
}

type Bird struct {
	Name string
}

func (b Bird) Sound() string { return b.Name + ": Tweet!" }
func (b Bird) Move() string  { return b.Name + " flies" }

func dynamicDispatch() {
	fmt.Println("=== 4. 动态分发 ===")

	var a Animal = Bird{Name: "Tweety"}
	fmt.Printf("动态调用 Sound: %s\n", a.Sound())
	fmt.Printf("动态调用 Move: %s\n", a.Move())

	// 面试要点:
	// 接口方法调用 ≈ 间接函数调用（通过 itab.fun 查表）
	// 内联优化: 接口方法调用无法内联（编译期不知道具体类型）
	// 所以热路径上应该避免使用接口，直接使用具体类型
	fmt.Println("注意: 接口方法无法内联，热路径避免使用接口")
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 接口组合
// ----------------------------------------------------------
// Go 的接口组合: 通过嵌入其他接口来组合新接口。
// io.Reader + io.Writer = io.ReadWriter
// 这是一种"组合优于继承"的设计。

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}

func composition() {
	fmt.Println("=== 5. 接口组合 ===")

	// 接口的类型集（Go 1.18+ 泛型约束）
	// interface 既可以作为值类型，也可以作为约束:
	//   type Number interface {
	//       int | float64 | float32
	//   }

	// 小接口原则:
	// Go 标准库推崇"接口越小越好"
	// io.Reader 只有 1 个方法
	// fmt.Stringer 只有 1 个方法
	// error 只有 1 个方法
	// 好处: 更容易实现，更容易组合，更容易测试
	fmt.Println("最佳实践: 接口越小越好（1-3个方法）")
	fmt.Println("使用者定义接口，而不是实现者（Go 的隐式实现使这成为可能）")
	fmt.Println()
}

// ----------------------------------------------------------
// 6. 空接口与泛型
// ----------------------------------------------------------
func emptyInterface() {
	fmt.Println("=== 6. 空接口 ===")

	// Go 1.18+: any 是 interface{} 的别名
	var x any = 42
	fmt.Printf("any value: %v, type: %T\n", x, x)

	x = "hello"
	fmt.Printf("any value: %v, type: %T\n", x, x)

	// 反射获取信息
	v := reflect.ValueOf(x)
	fmt.Printf("reflect: kind=%v, value=%v\n", v.Kind(), v)

	// Go 1.18+ 泛型 vs 空接口:
	// 泛型在编译期确定类型，空接口在运行时确定
	// 泛型有类型安全，空接口需要类型断言
	// 泛型值不装箱（不一定），空接口值总是装箱
	fmt.Println()
	fmt.Println("Go 1.18+ 泛型优于空接口: 编译期类型安全，无运行时开销")
}
