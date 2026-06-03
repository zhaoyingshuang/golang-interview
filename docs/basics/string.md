---
title: String 字符串
---

## 1. 底层结构

String 底层是 StringHeader{Data unsafe.Pointer; Len int}，本质是**只读**的字节切片。"你好" 在 UTF-8 下占 6 字节但只有 2 个 rune。字符串字面量存储在二进制文件的 .rodata 段，运行时字符串赋值只复制 StringHeader（16 字节），不复制底层数据。

**和 slice 的区别**：string 没有 cap 字段（因为不可变，不需要扩容）。

::: tip 使用场景
string 的零拷贝特性使其非常适合传递大量文本数据（如 HTTP 响应体、文件内容），因为赋值只复制 16 字节 header。

::: warning 面试追问
string 和 []byte 的转换有开销吗？→ 有，每次转换都会复制底层数据（编译器会优化一些简单场景）。

```go
type StringHeader struct {
    Data unsafe.Pointer // 指向底层字节数组
    Len  int            // 字节长度（不是字符数！）
}

s := "Hello, 世界"
len(s)                         // 13 (每个中文 3 字节)
utf8.RuneCountInString(s)      // 9 (rune 字符数)

// string 赋值只复制 header，不复制数据
s2 := s  // 只复制 16 bytes

// []byte(string) 和 string([]byte) 都会复制数据
b := []byte(s)  // 复制 13 bytes
s3 := string(b) // 又复制 13 bytes
```

## 2. byte vs rune

**byte = uint8**：表示一个 UTF-8 编码的字节。
:::

**rune = int32**：表示一个 Unicode 码点。"中" 的 UTF-8 编码是 0xE4 0xB8 0xAD → 3 个 byte，1 个 rune。

::: tip 使用场景
处理用户输入时必须用 rune 计算字符数（如限制用户名长度）；处理网络协议/文件时用 byte 操作原始数据。for range 遍历 string 时按 rune 迭代，index 是字节位置。

::: warning 面试追问
string 可以包含无效 UTF-8 吗？→ 可以，string 就是字节序列，for range 遇到无效 UTF-8 会产出 U+FFFD。

```go
s := "中A"

// 字节遍历（网络/文件 IO 场景）
for i := 0; i < len(s); i++ {
    fmt.Printf("%02x ", s[i])  // e4 b8 ad 41
}

// rune 遍历（用户输入处理场景）
for i, r := range s {
    // i 是字节位置，r 是 rune
    fmt.Printf("byte=%d rune=%U(%c)\n", i, r, r)
}

// rune 切片：支持按字符索引
runes := []rune(s)
fmt.Printf("第1个字符: %c\n", runes[0]) // 中

// 用户名长度限制（必须用 rune 计数）
if utf8.RuneCountInString(name) > 20 {
    return errors.New("name too long")
}
```

## 3. 拼接性能

性能排序（从快到慢）：strings.Builder > strings.Join > bytes.Buffer > + 运算符 > fmt.Sprintf。
:::

**strings.Builder 最优的原因**：String() 方法通过 unsafe 直接转换底层 []byte，不需要复制。但 String() 后不能再修改（否则 panic）。

**+ 运算符的问题**：每次拼接都分配新字符串，编译器会优化固定数量的 +，但循环中的 += 每次都分配。

::: tip 生产建议
循环拼接用 Builder（并 b.Grow 预分配）；少量拼接用 + 即可；数字转字符串用 strconv.Itoa（比 fmt.Sprintf 快 5-10 倍）。

```go
// 推荐: strings.Builder（循环拼接）
var b strings.Builder
b.Grow(len(parts) * 10) // 预分配
for _, p := range parts {
    b.WriteString(p)
}
result := b.String()

// 少量拼接直接用 +（编译器会优化）
s := "hello" + " " + "world" // 编译期合并

// 数字转字符串
s := strconv.Itoa(42)           // 推荐，零分配
s := fmt.Sprintf("%d", 42)      // 不推荐，有分配

// Builder 复用（高频场景）
var buf strings.Builder
buf.Grow(256)
for _, line := range lines {
    buf.Reset() // 重置但不释放内存
    buf.WriteString(line)
    process(buf.String())
}
```

## 4. 不可变性与内存泄漏

String 是不可变的：不能通过下标修改（s[0] = 'H' 编译错误）。这个设计使得 string 天然并发安全，且可以被安全地作为 map key。
:::

**Substring 内存泄漏**：对大字符串取子串不会复制底层数据（只创建新 header），如果长期持有子串，大字符串无法被 GC。

::: tip 使用场景
解析日志文件时只取某一行，但整行引用了整个文件内容。解决：用 strings.Clone()（Go 1.18+）或 string([]byte(sub)) 强制复制。

::: warning 面试追问
string 真的完全不可变吗？→ 通过 unsafe 可以修改（修改 .rodata 段），但极其危险，可能影响所有引用相同字面量的字符串。

```go
// substring 内存泄漏（生产常见！）
big := strings.Repeat("x", 1<<20) // 1MB
section := big[:50]               // 1MB 都无法被 GC

// 解决方案
section = strings.Clone(section)  // Go 1.18+，只复制 50 字节
// 或者
section = string([]byte(big[:50])) // 强制复制

// string 和 []byte 零拷贝转换（极端性能优化，不推荐）
// unsafe 方式，修改 b 会影响 s！
s := "hello"
b := unsafe.Slice(unsafe.StringData(s), len(s))

// 面试追问: 字符串可以比较吗？
// 可以，按字节逐个比较
// "hello" == "hello" → true
// "Hello" == "hello" → false
```
