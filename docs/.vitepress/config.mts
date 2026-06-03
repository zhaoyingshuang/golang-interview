import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/golang-interview/',
  title: 'Golang 面试知识点',
  description: '从底层原理到实战技巧，深度解析 Go 面试核心知识点',
  lang: 'zh-CN',
  lastUpdated: true,
  sitemap: {
    hostname: 'https://zhaoyingshuang.github.io/golang-interview/',
  },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/golang-interview/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#00add8' }],
  ],
  themeConfig: {
    nav: [
      { text: '首页', link: '/' },
      { text: '基础篇', link: '/basics/slice' },
      { text: '并发篇', link: '/concurrency/goroutine' },
      { text: '内存篇', link: '/memory/gc' },
      { text: '性能篇', link: '/performance/profiling' },
    ],
    sidebar: {
      '/basics/': [
        {
          text: '基础篇',
          items: [
            { text: 'Slice 切片', link: '/basics/slice' },
            { text: 'Map 哈希表', link: '/basics/map' },
            { text: 'String 字符串', link: '/basics/string' },
            { text: 'Interface 接口', link: '/basics/interface' },
            { text: 'Defer 延迟调用', link: '/basics/defer' },
            { text: 'Pointer 指针', link: '/basics/pointer' },
          ],
        },
      ],
      '/concurrency/': [
        {
          text: '并发篇',
          items: [
            { text: 'Goroutine 协程', link: '/concurrency/goroutine' },
            { text: 'Channel 通道', link: '/concurrency/channel' },
            { text: 'Sync 同步原语', link: '/concurrency/sync' },
            { text: 'Context 上下文', link: '/concurrency/context' },
            { text: '并发模式', link: '/concurrency/pattern' },
          ],
        },
      ],
      '/memory/': [
        {
          text: '内存篇',
          items: [
            { text: 'GC 垃圾回收', link: '/memory/gc' },
            { text: '逃逸分析', link: '/memory/escape' },
            { text: '内存对齐', link: '/memory/alignment' },
          ],
        },
      ],
      '/performance/': [
        {
          text: '性能篇',
          items: [
            { text: 'Profiling 性能分析', link: '/performance/profiling' },
            { text: 'Benchmark 基准测试', link: '/performance/benchmark' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/zhaoyingshuang/golang-interview' },
    ],
    footer: {
      message: '基于 CC BY-NC-SA 4.0 协议发布',
    },
    search: {
      provider: 'local',
    },
    lastUpdated: {
      text: '最后更新于',
    },
    outline: {
      label: '页面导航',
    },
    docFooter: {
      prev: '上一篇',
      next: '下一篇',
    },
  },
})
