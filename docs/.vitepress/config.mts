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
      { text: 'Web框架', link: '/webframework/router' },
      { text: '数据库', link: '/database/mysql-index' },
      { text: '微服务', link: '/microservice/grpc' },
      { text: '系统设计', link: '/systemdesign/mq' },
      { text: '算法', link: '/algorithm/sorting' },
      { text: '网络', link: '/network/tcp' },
      { text: 'OS', link: '/os/process' },
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
      '/webframework/': [
        {
          text: 'Web 框架',
          items: [
            { text: '路由原理', link: '/webframework/router' },
            { text: '中间件机制', link: '/webframework/middleware' },
            { text: '参数绑定', link: '/webframework/binding' },
            { text: 'RESTful 设计', link: '/webframework/restful' },
            { text: '错误处理', link: '/webframework/errors' },
          ],
        },
      ],
      '/database/': [
        {
          text: '数据库',
          items: [
            { text: 'MySQL 索引', link: '/database/mysql-index' },
            { text: 'MySQL 事务与锁', link: '/database/mysql-tx' },
            { text: 'Redis 缓存策略', link: '/database/redis-cache' },
            { text: 'Redis 分布式锁', link: '/database/redis-struct' },
            { text: 'Go database/sql', link: '/database/gosql' },
            { text: 'GORM 实战', link: '/database/gorm' },
          ],
        },
      ],
      '/microservice/': [
        {
          text: '微服务',
          items: [
            { text: 'gRPC', link: '/microservice/grpc' },
            { text: '服务注册发现', link: '/microservice/registry' },
            { text: '链路追踪', link: '/microservice/tracing' },
            { text: '限流熔断降级', link: '/microservice/circuitbreaker' },
            { text: '配置管理', link: '/microservice/config' },
          ],
        },
      ],
      '/systemdesign/': [
        {
          text: '系统设计',
          items: [
            { text: '消息队列', link: '/systemdesign/mq' },
            { text: '限流设计', link: '/systemdesign/ratelimit' },
            { text: '高可用架构', link: '/systemdesign/highavail' },
            { text: 'CAP 理论', link: '/systemdesign/cap' },
            { text: '一致性算法', link: '/systemdesign/consensus' },
          ],
        },
      ],
      '/algorithm/': [
        {
          text: '算法',
          items: [
            { text: '排序算法', link: '/algorithm/sorting' },
            { text: '二分查找', link: '/algorithm/binary' },
            { text: '哈希表', link: '/algorithm/hash' },
            { text: '树与二叉树', link: '/algorithm/tree' },
            { text: '动态规划', link: '/algorithm/dp' },
            { text: 'Go 标准库算法', link: '/algorithm/gostdlib' },
          ],
        },
      ],
      '/network/': [
        {
          text: '网络',
          items: [
            { text: 'TCP 协议', link: '/network/tcp' },
            { text: 'HTTP 协议', link: '/network/http' },
            { text: 'HTTPS 与 TLS', link: '/network/https' },
            { text: 'Go net/http', link: '/network/nethttp' },
            { text: 'WebSocket', link: '/network/websocket' },
          ],
        },
      ],
      '/os/': [
        {
          text: '操作系统',
          items: [
            { text: '进程与线程', link: '/os/process' },
            { text: '虚拟内存', link: '/os/virtualmem' },
            { text: 'IO 模型', link: '/os/iomodel' },
            { text: '文件系统', link: '/os/filesystem' },
            { text: 'Linux 排查', link: '/os/linuxcmd' },
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
