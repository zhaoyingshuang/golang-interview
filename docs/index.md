---
layout: home

hero:
  name: Golang 面试知识点
  text: 原理 · 实战 · 深度解析
  tagline: 从底层源码到生产实践，涵盖 16 个核心主题
  actions:
    - theme: brand
      text: 开始学习
      link: /basics/slice
    - theme: alt
      text: GitHub
      link: https://github.com/zhaoyingshuang/golang-interview
---

## 内容概览

### [基础篇](/basics/slice)

Slice、Map、String、Interface、Defer、Pointer 的底层原理与常见陷阱

### [并发篇](/concurrency/goroutine)

GMP 调度模型、Channel、Sync 原语、Context、并发设计模式

### [内存篇](/memory/gc)

三色标记 GC、逃逸分析、内存对齐与 struct 优化

### [性能篇](/performance/profiling)

pprof 性能分析、Benchmark 基准测试与优化实战

## 项目特色

- **底层源码分析** — 每个知识点追溯到 runtime 源码级别
- **生产实战经验** — 标注真实场景中的使用建议和常见坑点
- **面试追问预演** — 每节附带你可能被追问的问题和回答方向
- **可运行代码** — 所有示例均可 `go run` 运行验证
