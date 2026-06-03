package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// ============================================================
// String 深度解析
// ============================================================
//
// 【面试高频问题】
// 1. string 的底层结构？
// 2. string 是不可变的吗？能不能修改？
// 3. byte 和 rune 的区别？
// 4. 字符串拼接的几种方式及性能对比？
// 5. 字符串的遍历方式有哪些？
// 6. substring 操作会导致内存泄漏吗？

func main() {
	structure()
	immutability()
	byteAndRune()
	iteration()
	concatenation()
	substringLeak()
}

// ----------------------------------------------------------
// 1. 底层结构
// ----------------------------------------------------------
// reflect/string.go:
//
//   type StringHeader struct {
//       Data unsafe.Pointer // 指向底层字节数组
//       Len  int            // 字节长度（不是字符数！）
//   }
//
// string 本质是 {指针, 长度} 的只读字节切片。
// "你好" → UTF-8 编码占 6 字节，但只有 2 个 rune。
func structure() {
	fmt.Println("=== 1. 底层结构 ===")

	s := "Hello, 世界"
	fmt.Printf("string: %q\n", s)
	fmt.Printf("字节数(len): %d\n", len(s))           // 13 (每个中文 3 字节)
	fmt.Printf("字符数(RuneCountInString): %d\n", utf8.RuneCountInString(s)) // 9

	// 字符串的字面量在编译期就确定了，存储在二进制文件的 .rodata 段
	// 运行时字符串赋值只复制 StringHeader（16字节），不复制底层数据

	// 空字符串
	s1 := ""         // 真正的空字符串
	s2 := string("") // 也是空字符串
	fmt.Printf("空字符串: %q %q, 相等=%v\n", s1, s2, s1 == s2)
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 不可变性
// ----------------------------------------------------------
// string 的底层 byte slice 是只读的。
// 不能通过下标修改: s[0] = 'h' → 编译错误
//
// 但可以通过 unsafe 强行修改（极其危险，不要在生产环境使用）:
func immutability() {
	fmt.Println("=== 2. 不可变性 ===")

	s := "hello"
	// s[0] = 'H' // 编译错误: cannot assign to s[0]

	// 通过 unsafe 修改（仅作演示，实际禁止使用）
	bytes := unsafe.Slice(unsafe.StringData(s), len(s))
	bytes[0] = 'H'
	fmt.Printf("unsafe 修改后: %s (危险！可能修改了 .rodata 段)\n", s)
	// 注意: 这可能影响所有引用相同字面量的字符串！

	// 正确方式: 转为 []byte 修改后转回 string
	s2 := "world"
	b := []byte(s2)
	b[0] = 'W'
	s2 = string(b)
	fmt.Printf("正确方式: %s\n", s2)
	fmt.Println()
}

// ----------------------------------------------------------
// 3. byte vs rune
// ----------------------------------------------------------
// byte = uint8，表示一个 UTF-8 编码的字节
// rune = int32，表示一个 Unicode 码点
//
// "中" 的 UTF-8 编码: 0xE4 0xB8 0xAD → 3 个 byte，1 个 rune
// "A"  的 UTF-8 编码: 0x41            → 1 个 byte，1 个 rune
func byteAndRune() {
	fmt.Println("=== 3. byte vs rune ===")

	s := "中A"

	// 遍历字节
	fmt.Print("字节遍历: ")
	for i := 0; i < len(s); i++ {
		fmt.Printf("%02x ", s[i])
	}
	fmt.Println()

	// 遍历 rune
	fmt.Print("rune 遍历: ")
	for _, r := range s {
		fmt.Printf("%U(%c) ", r, r)
	}
	fmt.Println()

	// rune 和 byte 的转换
	r := '中'
	fmt.Printf("rune '中': value=%d, hex=%U, bytes=%x\n", r, r, []byte(string(r)))

	// string 和 []rune 的转换
	runes := []rune(s)
	fmt.Printf("[]rune: %U (长度=%d)\n", runes, len(runes))
	fmt.Printf("[]byte: %x (长度=%d)\n", []byte(s), len(s))
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 遍历方式
// ----------------------------------------------------------
func iteration() {
	fmt.Println("=== 4. 遍历方式 ===")

	s := "Go语言"

	// 方式1: for range — 按 rune 遍历，index 是字节位置
	fmt.Println("for range (rune遍历):")
	for i, r := range s {
		fmt.Printf("  byte_offset=%d, rune=%c\n", i, r)
	}

	// 方式2: for + []rune — 按 rune 遍历，index 是 rune 索引
	fmt.Println("[]rune 遍历 (rune索引):")
	runes := []rune(s)
	for i, r := range runes {
		fmt.Printf("  rune_index=%d, rune=%c\n", i, r)
	}

	// 方式3: for + byte — 按字节遍历
	fmt.Println("字节遍历:")
	for i := 0; i < len(s); i++ {
		fmt.Printf("  byte[%d]=%02x\n", i, s[i])
	}
	fmt.Println()
}

// ----------------------------------------------------------
// 5. 字符串拼接性能
// ----------------------------------------------------------
// 性能从差到好:
//   1. + 拼接 (每次都分配新字符串)
//   2. fmt.Sprintf (最慢，需要解析格式字符串)
//   3. strings.Builder (推荐，内部使用 []byte，只分配一次)
//   4. bytes.Buffer (和 Builder 差不多，但 Builder 不需要复制)
//
// strings.Builder 比 bytes.Buffer 快的原因:
// Builder 的 String() 方法直接返回底层 []byte 转成的 string（通过 unsafe），
// 不需要复制。而 Buffer 的 String() 会复制一份。
// 但 Builder 在 String() 后不能再修改（panic）。

func concatenation() {
	fmt.Println("=== 5. 字符串拼接 ===")

	parts := []string{"Hello", ",", " ", "世界", "!"}

	// 方式1: +
	s1 := parts[0] + parts[1] + parts[2] + parts[3] + parts[4]
	fmt.Printf("+ 拼接: %s\n", s1)

	// 方式2: strings.Join
	s2 := strings.Join(parts, "")
	fmt.Printf("Join: %s\n", s2)

	// 方式3: strings.Builder（推荐）
	var b strings.Builder
	b.Grow(20) // 预分配，避免扩容
	for _, p := range parts {
		b.WriteString(p)
	}
	s3 := b.String()
	fmt.Printf("Builder: %s\n", s3)

	// 方式4: fmt.Sprintf（最灵活但最慢）
	s4 := fmt.Sprintf("%s%s%s%s%s", parts[0], parts[1], parts[2], parts[3], parts[4])
	fmt.Printf("Sprintf: %s\n", s4)

	// 方式5: bytes.Buffer
	var buf bytes.Buffer
	for _, p := range parts {
		buf.WriteString(p)
	}
	s5 := buf.String()
	fmt.Printf("Buffer: %s\n", s5)

	// 类型转换的开销:
	// []byte(string) 和 string([]byte) 都会复制数据。
	// 如果只是临时需要，可以用 unsafe 零拷贝（但不安全，不推荐）。
	fmt.Println()

	// 数字转字符串:
	// strconv.Itoa 比 fmt.Sprintf 快 5-10 倍
	n := 42
	_ = strconv.Itoa(n)        // 推荐
	_ = fmt.Sprintf("%d", n)   // 不推荐（性能敏感场景）
	fmt.Println("数字转字符串用 strconv.Itoa，不要用 fmt.Sprintf")
}

// ----------------------------------------------------------
// 6. substring 内存泄漏
// ----------------------------------------------------------
// 和 slice 类似，string 的切片操作不会复制底层数据。
// 如果对一个很大的字符串取一小段并长期持有，大字符串的内存无法被 GC。
func substringLeak() {
	fmt.Println("=== 6. Substring 内存泄漏 ===")

	// 模拟: 解析一个很大的配置文件，只需要其中一小段
	bigConfig := strings.Repeat("x", 1<<20) // 1MB
	section := bigConfig[:50]               // 只需要前 50 字节

	// 此时 bigConfig 的整个 1MB 都无法被 GC（即使不再使用 bigConfig 变量）
	// 因为 section 的 StringHeader.Data 仍指向 bigConfig 的底层数组起始位置

	// 解决: 强制复制
	section = strings.Clone(section) // Go 1.18+
	// 或: section = string([]byte(bigConfig[:50]))

	_ = section
	fmt.Println("大字符串上取子串会阻止 GC → 用 strings.Clone() 解决")
	fmt.Println()
}
