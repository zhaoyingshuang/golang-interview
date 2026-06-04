package main

import (
	"fmt"
)

// ============================================================
// 配置管理
// ============================================================
//
// 【面试高频问题】
// 1. Go 项目配置管理有哪些方案？
// 2. Viper 的核心功能？如何实现热更新？
// 3. 敏感信息如何管理？
// 4. 配置中心的高可用怎么保证？

func main() {
	configPatterns()
	viperUsage()
	configCenter()
	bestPractices()
}

// ----------------------------------------------------------
// 1. 配置管理方案
// ----------------------------------------------------------
func configPatterns() {
	fmt.Println("=== 1. 配置管理方案 ===")
	fmt.Println()
	fmt.Println("  ┌──────────────┬──────────────────────────────────┐")
	fmt.Println("  │ 方案         │ 适用场景                         │")
	fmt.Println("  ├──────────────┼──────────────────────────────────┤")
	fmt.Println("  │ 环境变量     │ 12-Factor、Docker/K8s            │")
	fmt.Println("  │ 配置文件     │ 单机部署、本地开发               │")
	fmt.Println("  │ Viper        │ 综合方案、多格式支持             │")
	fmt.Println("  │ 配置中心     │ 微服务、需要热更新               │")
	fmt.Println("  └──────────────┴──────────────────────────────────┘")
	fmt.Println()
	fmt.Println("环境变量 (os.Getenv):")
	fmt.Println("  dbHost := os.Getenv(\"DB_HOST\")")
	fmt.Println("  dbPort := os.Getenv(\"DB_PORT\")")
	fmt.Println("  问题: 复杂结构难以表达，没有类型转换")
	fmt.Println()
	fmt.Println("结构体映射 (推荐):")
	fmt.Println("  type Config struct {")
	fmt.Println("    Server ServerConfig `yaml:\"server\"`")
	fmt.Println("    DB     DBConfig     `yaml:\"db\"`")
	fmt.Println("    Redis  RedisConfig  `yaml:\"redis\"`")
	fmt.Println("  }")
	fmt.Println("  type ServerConfig struct {")
	fmt.Println("    Port int    `yaml:\"port\"`")
	fmt.Println("    Mode string `yaml:\"mode\"`")
	fmt.Println("  }")
	fmt.Println("  type DBConfig struct {")
	fmt.Println("    Host     string `yaml:\"host\"`")
	fmt.Println("    Port     int    `yaml:\"port\"`")
	fmt.Println("    Database string `yaml:\"database\"`")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. Viper 使用
// ----------------------------------------------------------
func viperUsage() {
	fmt.Println("=== 2. Viper 配置管理 ===")
	fmt.Println()
	fmt.Println("Viper 核心功能:")
	fmt.Println("  1. 多格式支持: JSON/YAML/TOML/HCL/ENV/Java properties")
	fmt.Println("  2. 设置默认值")
	fmt.Println("  3. 优先级覆盖: 默认值 → 配置文件 → 环境变量 → 命令行参数")
	fmt.Println("  4. 热更新 Watch")
	fmt.Println("  5. 远程配置中心 (Etcd/Consul)")
	fmt.Println()
	fmt.Println("基本使用:")
	fmt.Println("  v := viper.New()")
	fmt.Println()
	fmt.Println("  // 设置默认值")
	fmt.Println("  v.SetDefault(\"server.port\", 8080)")
	fmt.Println()
	fmt.Println("  // 读取配置文件")
	fmt.Println("  v.SetConfigName(\"config\")     // config.yaml")
	fmt.Println("  v.SetConfigType(\"yaml\")")
	fmt.Println("  v.AddConfigPath(\"/etc/app\")")
	fmt.Println("  v.AddConfigPath(\".\")")
	fmt.Println("  v.ReadInConfig()")
	fmt.Println()
	fmt.Println("  // 环境变量覆盖")
	fmt.Println("  v.AutomaticEnv()")
	fmt.Println("  v.SetEnvPrefix(\"APP\")  // APP_SERVER_PORT")
	fmt.Println()
	fmt.Println("  // 反序列化到结构体")
	fmt.Println("  var cfg Config")
	fmt.Println("  v.Unmarshal(&cfg)")
	fmt.Println()
	fmt.Println("热更新:")
	fmt.Println("  v.OnConfigChange(func(e fsnotify.Event) {")
	fmt.Println("    log.Printf(\"配置变更: %s\", e.Name)")
	fmt.Println("    v.Unmarshal(&cfg)  // 重新解析")
	fmt.Println("  })")
	fmt.Println("  v.WatchConfig()")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. 配置中心
// ----------------------------------------------------------
func configCenter() {
	fmt.Println("=== 3. 配置中心 ===")
	fmt.Println()
	fmt.Println("常见配置中心:")
	fmt.Println()
	fmt.Println("  ┌──────────┬────────────┬─────────────────────┐")
	fmt.Println("  │ 名称     │ 协议       │ 特点                │")
	fmt.Println("  ├──────────┼────────────┼─────────────────────┤")
	fmt.Println("  │ Nacos    │ gRPC/HTTP  │ 阿里系、功能全      │")
	fmt.Println("  │ Apollo   │ HTTP       │ 携程、配置审计完善  │")
	fmt.Println("  │ Consul   │ HTTP/gRPC  │ HashiCorp、KV + 注册│")
	fmt.Println("  │ Etcd     │ gRPC       │ CNCF、K8s 生态      │")
	fmt.Println("  └──────────┴────────────┴─────────────────────┘")
	fmt.Println()
	fmt.Println("Go 接入 Nacos 示例:")
	fmt.Println("  client, _ := clients.CreateConfigClient(")
	fmt.Println("    vo.NacosClientParam{ServerConfigs: serverConfigs},")
	fmt.Println("  )")
	fmt.Println()
	fmt.Println("  // 获取配置")
	fmt.Println("  content, _ := client.GetConfig(vo.ConfigParam{")
	fmt.Println("    DataId: \"user-service.yaml\",")
	fmt.Println("    Group:  \"DEFAULT_GROUP\",")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("  // 监听配置变更")
	fmt.Println("  client.ListenConfig(vo.ConfigParam{")
	fmt.Println("    DataId:   \"user-service.yaml\",")
	fmt.Println("    OnChange: func(namespace, group, dataId, data string) {")
	fmt.Println("      log.Printf(\"配置变更: %s\", dataId)")
	fmt.Println("      // 重新加载配置...")
	fmt.Println("    },")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("高可用:")
	fmt.Println("  1. 配置中心集群部署 (3-5 节点)")
	fmt.Println("  2. 客户端本地缓存 (配置中心不可用时使用)")
	fmt.Println("  3. 多级缓存: 内存 → 本地文件 → 远程配置中心")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. 配置最佳实践
// ----------------------------------------------------------
func bestPractices() {
	fmt.Println("=== 4. 配置最佳实践 ===")
	fmt.Println()
	fmt.Println("敏感信息管理:")
	fmt.Println("  ✗ 不要将密钥写入配置文件或代码")
	fmt.Println("  ✗ 不要将密钥提交到 Git")
	fmt.Println("  ✓ 使用环境变量传递密钥")
	fmt.Println("  ✓ 使用 Secret 管理工具 (Vault/AWS Secrets Manager)")
	fmt.Println("  ✓ K8s Secret / Docker Secrets")
	fmt.Println()
	fmt.Println("多环境管理:")
	fmt.Println("  config/")
	fmt.Println("  ├── config.yaml          # 默认配置")
	fmt.Println("  ├── config-dev.yaml      # 开发环境覆盖")
	fmt.Println("  ├── config-staging.yaml  # 预发环境覆盖")
	fmt.Println("  └── config-prod.yaml     # 生产环境覆盖")
	fmt.Println()
	fmt.Println("  Viper 合并:")
	fmt.Println("  v.SetConfigName(\"config\")       // 基础配置")
	fmt.Println("  v.MergeInConfig()")
	fmt.Println("  v.SetConfigName(\"config-\" + env) // 环境覆盖")
	fmt.Println("  v.MergeInConfig()")
	fmt.Println()
	fmt.Println("配置校验:")
	fmt.Println("  type Config struct {")
	fmt.Println("    Server ServerConfig `validate:\"required\"`")
	fmt.Println("  }")
	fmt.Println("  validate := validator.New()")
	fmt.Println("  if err := validate.Struct(cfg); err != nil {")
	fmt.Println("    log.Fatal(\"配置校验失败:\", err)")
	fmt.Println("  }")
	fmt.Println()
}
