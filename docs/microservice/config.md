---
title: 配置管理
---

## 1. 配置方案对比

| 方案 | 适用场景 | 优点 | 缺点 |
|------|----------|------|------|
| 环境变量 | Docker/K8s | 简单 | 复杂结构难表达 |
| 配置文件 | 本地开发 | 直观 | 不支持热更新 |
| Viper | 综合方案 | 多格式+热更新 | 有学习成本 |
| 配置中心 | 微服务 | 热更新+版本管理 | 运维复杂 |

## 2. Viper 核心功能

优先级：默认值 → 配置文件 → 环境变量 → 命令行参数（后者覆盖前者）

```go
v := viper.New()
v.SetDefault("server.port", 8080)
v.SetConfigName("config")
v.SetConfigType("yaml")
v.AddConfigPath(".")
v.ReadInConfig()

// 环境变量覆盖: APP_SERVER_PORT
v.SetEnvPrefix("APP")
v.AutomaticEnv()

// 反序列化到结构体
var cfg Config
v.Unmarshal(&cfg)

// 热更新
v.OnConfigChange(func(e fsnotify.Event) {
    log.Printf("配置变更: %s", e.Name)
    v.Unmarshal(&cfg)
})
v.WatchConfig()
```

## 3. 配置中心

| 名称 | 特点 |
|------|------|
| Nacos | 阿里系、功能全面 |
| Apollo | 携程、配置审计完善 |
| Consul | HashiCorp、KV + 服务发现 |
| Etcd | CNCF、K8s 生态 |

::: tip 高可用保证
1. 配置中心集群部署（3-5 节点）
2. 客户端本地缓存（配置中心不可用时使用）
3. 多级缓存：内存 → 本地文件 → 远程配置中心
:::

## 4. 最佳实践

- 敏感信息：用 Vault/K8s Secret，不要写入配置文件或 Git
- 多环境：`config.yaml` + `config-{env}.yaml` 合并覆盖
- 配置校验：用 `validator` 在启动时校验必填项
- 结构体映射：定义强类型 Config struct，避免 `v.Get("xxx")` 散落各处

```go
type Config struct {
    Server ServerConfig `yaml:"server" validate:"required"`
    DB     DBConfig     `yaml:"db"     validate:"required"`
}
```
