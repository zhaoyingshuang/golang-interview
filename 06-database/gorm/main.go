package main

import "fmt"

// ============================================================
// GORM ORM
// ============================================================
// 【面试高频】零值更新, N+1查询, 软删除, Hook机制

func main() {
	coreConcepts()
	associations()
	hooks()
	pitfalls()
}

func coreConcepts() {
	fmt.Println("=== 1. GORM 核心概念 ===")
	fmt.Println()
	fmt.Println("链式调用原理:")
	fmt.Println("  db.Where().Order().Limit().Find()")
	fmt.Println("  每个方法返回新 *DB(不修改原DB)")
	fmt.Println("  Session 模式: db.Session(&gorm.Session{})")
	fmt.Println()
	fmt.Println("零值更新问题:")
	fmt.Println("  type User { Name string; Age int }")
	fmt.Println("  db.Model(&user).Updates(User{Name: \"new\", Age: 0})")
	fmt.Println("  → Age=0 不会被更新!(GORM认为零值=未设置)")
	fmt.Println()
	fmt.Println("解决方式:")
	fmt.Println("  1. 使用 map: Updates(map[string]any{\"age\": 0})")
	fmt.Println("  2. 使用 Select 指定列: Select(\"Name\", \"Age\").Updates(...)")
	fmt.Println("  3. 使用指针: *int → nil 表示未设置")
	fmt.Println()
}

func associations() {
	fmt.Println("=== 2. 关联关系 ===")
	fmt.Println()
	fmt.Println("预加载 vs Joins:")
	fmt.Println("  Preload: 额外发一条SELECT → 简单但N+1变2条")
	fmt.Println("  Joins:   LEFT JOIN → 一条SQL, 但结果需要处理")
	fmt.Println()
	fmt.Println("N+1 查询问题:")
	fmt.Println("  for _, order := range orders {") // 100个订单
	fmt.Println("      db.Find(&order.Items)              // 每个订单发1条SQL")
	fmt.Println("  }")
	fmt.Println("  → 1+100 = 101 条SQL!")
	fmt.Println()
	fmt.Println("解决:")
	fmt.Println("  db.Preload(\"Items\").Find(&orders)  // 2条SQL")
	fmt.Println("  db.Joins(\"Items\").Find(&orders)     // 1条SQL")
	fmt.Println()
}

func hooks() {
	fmt.Println("=== 3. Hook 与插件 ===")
	fmt.Println()
	fmt.Println("Hook 调用链:")
	fmt.Println("  BeforeSave → BeforeCreate → AfterCreate → AfterSave")
	fmt.Println("  BeforeSave → BeforeUpdate → AfterUpdate → AfterSave")
	fmt.Println()
	fmt.Println("自定义 Hook:")
	fmt.Println("  func (u *User) BeforeCreate(tx *gorm.DB) error {")
	fmt.Println("    if u.Name == \"\" { return errors.New(\"name required\") }")
	fmt.Println("    return nil")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("跳过 Hook:")
	fmt.Println("  db.Session(&gorm.Session{SkipHooks: true}).Create(...)")
	fmt.Println()
}

func pitfalls() {
	fmt.Println("=== 4. 常见陷阱 ===")
	fmt.Println()
	fmt.Println("1. 软删除影响查询:")
	fmt.Println("   db.Unscoped().Where(...).Find()  // 包含已删除记录")
	fmt.Println()
	fmt.Println("2. 批量数据分批处理:")
	fmt.Println("   db.Where(\"id > ?\", 0).FindInBatches(&results, 100, func(tx, batch) {")
	fmt.Println("     // 每100条处理一次, 避免OOM")
	fmt.Println("   })")
	fmt.Println()
	fmt.Println("3. 原生SQL优化:")
	fmt.Println("   复杂查询用 db.Raw() 而非 ORM 链式调用")
	fmt.Println("   批量插入用 db.CreateInBatches()")
	fmt.Println()
	fmt.Println("4. 连接复用:")
	fmt.Println("   gorm.Open 返回的 db 是并发安全的, 全局复用")
	fmt.Println("   不要每次请求创建新连接!")
	fmt.Println()
}
