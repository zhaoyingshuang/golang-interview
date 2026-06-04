---
title: GORM 实战
---

## 1. GORM 架构

GORM 采用插件化架构设计，核心组件包括 ConnPool、Callbacks 和 Serializer。

### 初始化与连接

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
    dsn := "user:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // 打印 SQL 日志
    })
    if err != nil {
        log.Fatal(err)
    }

    // 获取底层 sql.DB 进行连接池配置
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    return db
}
```

### 插件体系

```go
// 注册回调插件
db.Callback().Create().Before("gorm:create").Register("my_plugin:before_create", func(db *gorm.DB) {
    log.Println("before create callback")
})

// 使用内置插件
import "gorm.io/plugin/dbresolver"

db.Use(dbresolver.Register(dbresolver.Config{
    Sources:  []gorm.Dialector{mysql.Open(primaryDSN)},  // 写库
    Replicas: []gorm.Dialector{mysql.Open(replicaDSN)},  // 读库
    Policy:   dbresolver.RandomPolicy{},
}))
```

::: tip
GORM 的 Callback 机制允许在 CRUD 操作前后插入自定义逻辑，类似于中间件模式。可以用于审计日志、参数校验等。
:::

## 2. CRUD 操作

### 模型定义

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:100;not null;index"`
    Email     string         `gorm:"size:200;uniqueIndex"`
    Age       int            `gorm:"default:18"`
    Status    string         `gorm:"size:20;default:active"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除
}
```

### Create

```go
// 创建单条记录
user := User{Name: "张三", Email: "zhangsan@example.com", Age: 25}
result := db.Create(&user)
fmt.Println(result.Error)        // 错误信息
fmt.Println(result.RowsAffected) // 影响行数
fmt.Println(user.ID)             // 自动填充自增 ID

// 批量创建（建议使用 CreateInBatches）
users := []User{
    {Name: "李四", Email: "lisi@example.com"},
    {Name: "王五", Email: "wangwu@example.com"},
}
db.CreateInBatches(users, 100) // 每批 100 条
```

### Read

```go
// 单条查询
var user User
db.First(&user, 1)                             // 按主键查询
db.First(&user, "email = ?", "test@example.com") // 按条件查询

// 多条查询
var users []User
db.Where("age > ?", 18).Find(&users)

// 条件构造
db.Where("name LIKE ?", "%张%").
    Where("age BETWEEN ? AND ?", 20, 30).
    Order("created_at DESC").
    Limit(10).
    Offset(0).
    Find(&users)

// 统计
var count int64
db.Model(&User{}).Where("status = ?", "active").Count(&count)
```

### Update

```go
// 更新单个字段
db.Model(&user).Update("name", "新名字")

// 更新多个字段（struct 方式，注意零值问题）
db.Model(&user).Updates(User{Name: "新名字", Age: 30})

// 更新多个字段（map 方式，可以更新零值）
db.Model(&user).Updates(map[string]interface{}{
    "name":   "新名字",
    "age":    0,      // map 方式可以更新为零值
    "status": "",
})

// 条件更新
db.Model(&User{}).Where("age < ?", 18).Update("status", "minor")
```

### Delete

```go
// 软删除（模型包含 DeletedAt 字段时）
db.Delete(&user)                   // UPDATE users SET deleted_at=NOW WHERE id=1
db.Where("name = ?", "张三").Delete(&User{})

// 查询包含软删除的记录
db.Unscoped().Where("name = ?", "张三").Find(&users)

// 永久删除
db.Unscoped().Delete(&user)
```

::: warning
使用 struct 进行 `Updates` 时，零值字段（如 `Age: 0`、`Name: ""`）不会被更新。如果需要更新零值，必须使用 `map[string]interface{}` 或使用 `Select` 指定字段。
:::

## 3. 关联关系

### 关系定义

```go
type User struct {
    ID       uint
    Name     string
    Profile  Profile   `gorm:"belongsTo"`           // 属于（一对一）
    Orders   []Order   `gorm:"foreignKey:UserID"`   // 拥有（一对多）
    Roles    []Role    `gorm:"many2many:user_roles"` // 多对多
}

type Profile struct {
    ID     uint
    UserID uint   // 外键
    Bio    string
}

type Order struct {
    ID     uint
    UserID uint   // 外键
    Amount float64
    Items  []Item `gorm:"foreignKey:OrderID"`
}

type Role struct {
    ID   uint
    Name string
}
```

### 预加载（Preload）

```go
// 预加载关联（解决 N+1 问题）
var users []User
db.Preload("Orders").Find(&users)
// SQL: SELECT * FROM users;
// SQL: SELECT * FROM orders WHERE user_id IN (1,2,3,...);

// 嵌套预加载
db.Preload("Orders.Items").Find(&users)

// 条件预加载
db.Preload("Orders", "status = ?", "paid").Find(&users)

// Joins 预加载（使用 JOIN 而非分两条查询）
db.Joins("Profile").Find(&users)
```

### 关联操作

```go
// 创建时关联
user := User{
    Name: "张三",
    Profile: Profile{Bio: "Hello"},
    Orders: []Order{{Amount: 99.9}},
}
db.Create(&user) // 同时创建关联记录

// 添加关联
db.Model(&user).Association("Roles").Append([]Role{{Name: "admin"}})

// 替换关联
db.Model(&user).Association("Roles").Replace([]Role{role1, role2})

// 删除关联（只删除关联关系，不删除记录）
db.Model(&user).Association("Roles").Delete(role1)

// 清空关联
db.Model(&user).Association("Roles").Clear()
```

## 4. 性能优化

### 批量操作

```go
// 批量创建（避免逐条插入）
db.CreateInBatches(users, 100)

// 批量更新（使用 Case 语句）
db.Model(&User{}).Where("id IN ?", ids).Updates(map[string]interface{}{
    "status": gorm.Expr("CASE WHEN age >= 18 THEN 'adult' ELSE 'minor' END"),
})
```

### Select 指定字段

```go
// 只查询需要的字段，减少数据传输
var users []User
db.Select("id", "name").Find(&users)

// 查询特定字段到自定义结构
type UserBrief struct {
    ID   uint
    Name string
}
var briefs []UserBrief
db.Model(&User{}).Select("id", "name").Find(&briefs)
```

### Debug SQL

```go
// 单次查询 Debug
db.Debug().Where("name = ?", "张三").First(&user)

// 全局 Debug
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

// 自定义慢查询日志
newLogger := logger.New(
    log.New(os.Stdout, "\r\n", log.LstdFlags),
    logger.Config{
        SlowThreshold: 200 * time.Millisecond, // 慢查询阈值
        LogLevel:      logger.Warn,
    },
)
db.Logger = newLogger
```

::: info
生产环境建议设置 `SlowThreshold` 为 200ms 或更低，只记录慢查询，避免日志量过大。
:::

## 5. 常见陷阱

### 零值更新问题

```go
// 问题：struct 的零值字段会被忽略
db.Model(&user).Updates(User{Name: "新名字", Age: 0})
// SQL: UPDATE users SET name='新名字', updated_at=... WHERE id=1
// Age=0 被忽略了！

// 方案一：使用 map
db.Model(&user).Updates(map[string]interface{}{"name": "新名字", "age": 0})

// 方案二：使用 Select 指定字段
db.Model(&user).Select("Name", "Age").Updates(User{Name: "新名字", Age: 0})

// 方案三：Select("*") 更新所有字段
db.Model(&user).Select("*").Updates(User{Name: "新名字", Age: 0})
```

### 软删除查询

```go
// 软删除的记录默认不会出现在查询结果中
db.Where("name = ?", "张三").Find(&users) // 不包含已软删除的

// 包含软删除记录
db.Unscoped().Where("name = ?", "张三").Find(&users)

// 注意：关联查询也需要处理软删除
db.Preload("Orders").Unscoped().Find(&users)
```

### N+1 问题

```go
// N+1 问题：循环中查询关联
var users []User
db.Find(&users)
for _, u := range users {
    var orders []Order
    db.Where("user_id = ?", u.ID).Find(&orders) // 每个用户一次查询！
}

// 解决方案：使用 Preload
var users []User
db.Preload("Orders").Find(&users) // 只需 2 条 SQL

// 或者使用 Joins
db.Joins("Orders").Find(&users)
```

::: warning 面试高频问题
1. **GORM 的 Updates 使用 struct 和 map 有什么区别？**
   - struct 会忽略零值字段（`Age: 0`、`Name: ""` 不更新）；map 会更新所有指定字段，包括零值。推荐使用 map 或 `Select` 明确指定要更新的字段。

2. **如何解决 GORM 的 N+1 问题？**
   - 使用 `Preload` 预加载关联（生成 IN 查询）；使用 `Joins` 进行 JOIN 查询；避免在循环中查询数据库。

3. **GORM 的软删除是如何实现的？**
   - 模型包含 `gorm.DeletedAt` 字段时，`Delete` 操作会将其设为当前时间而非真正删除。查询时自动添加 `WHERE deleted_at IS NULL` 条件。使用 `Unscoped()` 可以查询包含软删除的记录。

4. **如何查看 GORM 生成的 SQL？**
   - 使用 `db.Debug()` 开启单次 Debug；通过 `Logger` 配置打印 SQL；自定义 Logger 设置 `SlowThreshold` 记录慢查询。

5. **GORM 的 ConnPool 和底层 sql.DB 是什么关系？**
   - GORM 的 `DB.DB()` 返回底层的 `*sql.DB`，连接池参数（MaxOpenConns 等）需要在 `*sql.DB` 上设置。GORM 的 CRUD 操作最终通过 `*sql.DB` 的连接池执行。
:::
