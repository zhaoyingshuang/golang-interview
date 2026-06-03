package main

import "fmt"

// ============================================================
// Pointer 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. 值类型和引用类型的区别？
// 2. 什么时候应该用指针？
// 3. Go 有引用传递吗？
// 4. 指针和接口的关系？
// 5. new 和 make 的区别？
// 6. 逃逸分析如何影响指针使用？

func main() {
	valueVsPointer()
	noReferencePassing()
	newVsMake()
	pointerAndInterface()
	methodSet()
	copyCost()
}

// ----------------------------------------------------------
// 1. 值类型 vs 引用类型（更准确地说是"引用语义类型"）
// ----------------------------------------------------------
// 值类型: int, float, bool, string, array, struct
//   - 赋值和传参会复制整个值
//   - 修改副本不影响原始值
//
// 引用语义类型: slice, map, channel, pointer, interface, function
//   - 本质上也是值类型（Go 中一切皆值传递）
//   - 但它们内部包含指针，复制的是"头部"，底层数据仍共享
//   - 修改底层数据会影响原始值（但修改头部不会）

func valueVsPointer() {
	fmt.Println("=== 1. 值类型 vs 引用语义类型 ===")

	// 值类型: 修改副本不影响原始值
	x := 42
	y := x
	y = 100
	fmt.Printf("值类型: x=%d, y=%d (互不影响)\n", x, y)

	// struct 是值类型
	type Point struct{ X, Y int }
	p1 := Point{1, 2}
	p2 := p1
	p2.X = 10
	fmt.Printf("struct: p1=%v, p2=%v (互不影响)\n", p1, p2)

	// slice 是引用语义类型
	s1 := []int{1, 2, 3}
	s2 := s1
	s2[0] = 100
	fmt.Printf("slice: s1=%v, s2=%v (共享底层数组)\n", s1, s2)

	// map 是引用语义类型
	m1 := map[string]int{"a": 1}
	m2 := m1
	m2["a"] = 100
	fmt.Printf("map: m1=%v, m2=%v (共享底层数据)\n", m1, m2)
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Go 没有引用传递
// ----------------------------------------------------------
// Go 中所有参数传递都是值传递（pass by value）。
// 即使传指针，也是复制了指针的值（地址），不是引用传递。
func noReferencePassing() {
	fmt.Println("=== 2. Go 只有值传递 ===")

	x := 42
	fmt.Printf("调用前: x=%d\n", x)

	// 传值: 无法修改原始值
	tryModify(x)
	fmt.Printf("tryModify 后: x=%d (不变)\n", x)

	// 传指针: 可以修改原始值
	tryModifyPtr(&x)
	fmt.Printf("tryModifyPtr 后: x=%d (被修改)\n", x)

	// 但传指针也是值传递！
	// 函数收到的是指针的副本（一个复制的地址）
	// 通过这个副本仍然能访问同一个对象
	fmt.Println("传指针也是值传递：复制的是地址值，不是对象本身")
	fmt.Println()
}

func tryModify(n int) {
	n = 100 // 修改的是副本
}

func tryModifyPtr(n *int) {
	*n = 100  // 通过地址修改原始值
}

// ----------------------------------------------------------
// 3. new vs make
// ----------------------------------------------------------
// new(T):   分配零值内存，返回 *T
// make(T):  只用于 slice/map/channel，返回初始化后的 T（不是指针）
//
// 注意: new 在实际开发中很少使用，因为:
//   - &MyStruct{} 比 new(MyStruct) 更清晰
//   - slice/map/channel 必须用 make 初始化

func newVsMake() {
	fmt.Println("=== 3. new vs make ===")

	// new: 返回指针，零值初始化
	p := new(int)
	fmt.Printf("new(int): value=%d, ptr=%p\n", *p, p)

	type Config struct {
		Name string
		Port int
	}
	cfg := new(Config)
	fmt.Printf("new(Config): %+v (零值)\n", cfg)

	// make: 返回初始化后的值（不是指针）
	s := make([]int, 0, 10)
	m := make(map[string]int, 10)
	ch := make(chan int, 5)
	fmt.Printf("make slice: len=%d cap=%d\n", len(s), cap(s))
	fmt.Printf("make map: len=%d\n", len(m))
	fmt.Printf("make chan: cap=%d\n", cap(ch))
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 指针和接口
// ----------------------------------------------------------
// 值和指针都可以实现接口，但有区别:
//   - 如果方法接收者是值类型 (T)，则 T 和 *T 都能实现接口
//   - 如果方法接收者是指针类型 (*T)，则只有 *T 能实现接口
//
// 原因: Go 可以自动取地址 (T → &T)，但不能自动解引用 (*T → T)
//   但有个限制: 如果值不可寻址（如 map 的元素），就不能自动取地址

type Describer interface {
	Describe() string
}

type Person struct {
	Name string
	Age  int
}

// 值接收者: Person 和 *Person 都可以实现 Describer
func (p Person) Describe() string {
	return fmt.Sprintf("%s (%d)", p.Name, p.Age)
}

func pointerAndInterface() {
	fmt.Println("=== 4. 指针和接口 ===")

	var d Describer

	d = Person{"Alice", 30}      // 值实现接口 ✓
	fmt.Printf("值: %s\n", d.Describe())

	d = &Person{"Bob", 25}       // 指针实现接口 ✓
	fmt.Printf("指针: %s\n", d.Describe())
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 方法集（Method Set）
// ----------------------------------------------------------
// 类型的完整方法集:
//   - T 的方法集: 所有接收者为 T 的方法
//   - *T 的方法集: 所有接收者为 T 或 *T 的方法
//
// 在实际编程中，Go 会自动取地址/解引用来匹配方法，
// 但在接口赋值和类型嵌入时，方法集规则严格生效。

type Animal struct {
	Name string
}

func (a Animal) Speak() string  { return a.Name + " speaks" }
func (a *Animal) ChangeName(n string) { a.Name = n }

func methodSet() {
	fmt.Println("=== 5. 方法集 ===")

	a := Animal{Name: "Dog"}
	ap := &a

	// 值可以调用指针方法（Go 自动取地址）
	a.ChangeName("Cat") // 等价于 (&a).ChangeName("Cat")
	fmt.Printf("值调用指针方法: %s\n", a.Name)

	// 指针可以调用值方法（Go 自动解引用）
	ap.Speak() // 等价于 (*ap).Speak()
	fmt.Println("指针可以调用值方法（自动解引用）")

	// 接口赋值时方法集严格匹配:
	type Changer interface{ ChangeName(string) }

	// var c Changer = a  // 编译错误！Animal 没有 ChangeName（需要指针接收者）
	var c Changer = &a   // ✓ *Animal 有 ChangeName
	_ = c
	fmt.Println("接口赋值时方法集严格匹配: 指针接收者的方法只能用 *T 满足接口")
	fmt.Println()
}

// ----------------------------------------------------------
// 6. 值拷贝开销
// ----------------------------------------------------------
// 什么时候用指针？
//   1. 需要修改原始值
//   2. 结构体较大（避免复制开销）
//   3. 一致性（某个方法需要指针，其他方法也统一用指针）
//
// 什么时候用值？
//   1. 小结构体（<= 64 bytes，复制比指针解引用更快）
//   2. 不可变数据（值传递天然安全）
//   3. map 的 key（必须是可比较的，指针也可以比较但不常用）

type BigStruct struct {
	Data [1024]byte
	Name string
}

func copyCost() {
	fmt.Println("=== 6. 值拷贝开销 ===")

	// 大结构体应该用指针
	big := &BigStruct{Name: "big"}

	// 值传递: 复制 1024+ 字节
	// 指针传递: 只复制 8 字节（一个地址）
	_ = big

	fmt.Println("小结构体用值（<= 64 bytes），大结构体用指针")
	fmt.Println("如果某个方法需要指针接收者，全部方法统一用指针接收者")
}
