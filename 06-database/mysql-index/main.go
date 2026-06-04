package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("========== MySQL 索引 ==========")
	fmt.Println()

	bPlusTree()
	clusteredAndSecondaryIndex()
	indexOptimization()
	explainExecutionPlan()
	indexFailureScenarios()
}

// ============================================================
// 1. B+树结构
// B+树 vs B树, 为什么数据库选B+, 查询过程(根->叶), 3层B+树存2000万行
// ============================================================

func bPlusTree() {
	fmt.Println("【1. B+树结构】")
	fmt.Println(strings.Repeat("-", 50))

	// B+树 vs B树的关键区别
	fmt.Println("B+树 vs B树的关键区别:")
	fmt.Println("  1. B+树所有数据存储在叶子节点, 内部节点只存键值(索引)")
	fmt.Println("  2. B+树叶子节点通过双向链表连接, 支持高效范围查询")
	fmt.Println("  3. B+树内部节点更小, 同样页大小能存更多键, 树更矮")
	fmt.Println("  4. B+树查询性能稳定, 每次都要走到叶子节点, 路径长度一致")
	fmt.Println()

	// 为什么数据库选择B+树而不是B树
	fmt.Println("为什么数据库选择B+树:")
	fmt.Println("  - 磁盘IO是瓶颈, B+树更矮意味着更少的磁盘IO")
	fmt.Println("  - 范围查询是数据库最常见的操作, 叶子链表完美支持")
	fmt.Println("  - 页大小固定(通常16KB), 内部节点只存键可存更多条目")
	fmt.Println()

	// 模拟B+树查询过程: 根 -> 中间 -> 叶子
	fmt.Println("B+树查询过程 (从根到叶):")

	// 模拟一个3层B+树
	type BPlusTreeNode struct {
		Keys     []int
		Children []*BPlusTreeNode // 内部节点的子节点
		Values   []string         // 仅叶子节点存储数据
		Next     *BPlusTreeNode   // 叶子节点的链表指针
		IsLeaf   bool
	}

	// 构建一个简单的B+树示例
	// 叶子节点
	leaf1 := &BPlusTreeNode{
		Keys:   []int{1, 3, 5},
		Values: []string{"row1", "row3", "row5"},
		IsLeaf: true,
	}
	leaf2 := &BPlusTreeNode{
		Keys:   []int{7, 9, 11},
		Values: []string{"row7", "row9", "row11"},
		IsLeaf: true,
	}
	leaf3 := &BPlusTreeNode{
		Keys:   []int{13, 15, 17},
		Values: []string{"row13", "row15", "row17"},
		IsLeaf: true,
	}

	// 叶子节点链表连接
	leaf1.Next = leaf2
	leaf2.Next = leaf3

	// 根节点
	root := &BPlusTreeNode{
		Keys:     []int{7, 13},
		Children: []*BPlusTreeNode{leaf1, leaf2, leaf3},
		IsLeaf:   false,
	}

	// 模拟查询key=9的过程
	searchKey := 9
	fmt.Printf("  查询 key=%d 的过程:\n", searchKey)

	// 第1层: 根节点
	fmt.Printf("  [第1层-根节点] Keys: %v\n", root.Keys)
	childIndex := 0
	for i, k := range root.Keys {
		if searchKey >= k {
			childIndex = i + 1
		}
	}
	fmt.Printf("  -> key=%d >= %d? 是, 走第%d个子节点\n", searchKey, root.Keys[len(root.Keys)-1], childIndex+1)

	// 第2层: 叶子节点
	leaf := root.Children[childIndex]
	fmt.Printf("  [第2层-叶子节点] Keys: %v\n", leaf.Keys)

	// 在叶子节点中二分查找
	found := false
	for i, k := range leaf.Keys {
		if k == searchKey {
			fmt.Printf("  -> 找到! key=%d, value=%s\n", searchKey, leaf.Values[i])
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("  -> 未找到 key=%d\n", searchKey)
	}
	fmt.Println()

	// 3层B+树存储2000万行的计算
	fmt.Println("3层B+树如何存2000万行:")
	fmt.Println("  假设:")
	fmt.Println("    - InnoDB页大小 = 16KB")
	fmt.Println("    - 主键为BIGINT = 8字节")
	fmt.Println("    - 页指针 = 6字节")
	fmt.Println()
	fmt.Println("  内部节点(非叶子):")
	fmt.Println("    每个键值+指针 = 8 + 6 = 14字节")
	fmt.Println("    每页可存约 16KB/14B ≈ 1170个键")
	fmt.Println("    每页有 1170+1 = 1171 个子节点指针")
	fmt.Println()
	fmt.Println("  叶子节点:")
	fmt.Println("    假设每行数据约1KB, 每页可存约16行")
	fmt.Println()
	fmt.Println("  容量计算:")
	rootCapacity := 1171
	leafCapacity := 16
	level2Capacity := rootCapacity * leafCapacity
	level3Capacity := rootCapacity * rootCapacity * leafCapacity
	fmt.Printf("    1层B+树: %d 行\n", leafCapacity)
	fmt.Printf("    2层B+树: %d × %d = %d 行\n", rootCapacity, leafCapacity, level2Capacity)
	fmt.Printf("    3层B+树: %d × %d × %d = %d 行 (约%d万)\n",
		rootCapacity, rootCapacity, leafCapacity, level3Capacity, level3Capacity/10000)
	fmt.Println("    结论: 3层B+树最多只需3次磁盘IO即可找到任何一行数据")
	fmt.Println()
}

// ============================================================
// 2. 聚簇索引与二级索引
// InnoDB 聚簇索引=主键, 二级索引回表, 覆盖索引避免回表
// ============================================================

func clusteredAndSecondaryIndex() {
	fmt.Println("【2. 聚簇索引与二级索引】")
	fmt.Println(strings.Repeat("-", 50))

	// 聚簇索引: 数据和主键索引存储在一起
	fmt.Println("聚簇索引 (Clustered Index):")
	fmt.Println("  - InnoDB中主键索引就是聚簇索引")
	fmt.Println("  - 叶子节点直接存储完整的行数据")
	fmt.Println("  - 一张表只能有一个聚簇索引")
	fmt.Println("  - 数据按主键顺序物理存储(所以叫'聚簇')")
	fmt.Println("  - 如果没有主键, InnoDB会选择唯一非空索引, 或自动生成ROWID")
	fmt.Println()

	// 模拟聚簇索引结构
	type RowData struct {
		ID   int
		Name string
		Age  int
	}

	// 聚簇索引: 主键ID -> 完整行数据
	clusteredIndex := map[int]RowData{
		1: {ID: 1, Name: "Alice", Age: 25},
		2: {ID: 2, Name: "Bob", Age: 30},
		3: {ID: 3, Name: "Charlie", Age: 28},
		5: {ID: 5, Name: "Eve", Age: 22},
	}
	fmt.Println("  聚簇索引示例 (主键ID -> 行数据):")
	for k, v := range clusteredIndex {
		fmt.Printf("    ID=%d -> {name: %s, age: %d}\n", k, v.Name, v.Age)
	}
	fmt.Println()

	// 二级索引: 存储索引列值 -> 主键ID
	fmt.Println("二级索引 (Secondary Index):")
	fmt.Println("  - 非主键索引都是二级索引")
	fmt.Println("  - 叶子节点存储: 索引列值 + 主键ID (不是行数据)")
	fmt.Println("  - 查询时需要'回表': 先查二级索引得到主键, 再查聚簇索引得到行数据")
	fmt.Println()

	// 模拟二级索引: Name -> ID
	secondaryIndex := map[string]int{
		"Alice":   1,
		"Bob":     2,
		"Charlie": 3,
		"Eve":     5,
	}
	fmt.Println("  二级索引示例 (name -> 主键ID):")
	for k, v := range secondaryIndex {
		fmt.Printf("    name=%s -> ID=%d\n", k, v)
	}
	fmt.Println()

	// 模拟回表过程
	fmt.Println("  回表查询过程: SELECT * FROM users WHERE name = 'Bob'")
	fmt.Println("    步骤1: 查二级索引 name='Bob' -> 得到主键ID=2")
	fmt.Println("    步骤2: 查聚簇索引 ID=2 -> 得到完整行数据")
	fmt.Println("    这就是'回表', 多了一次B+树查找")
	fmt.Println()

	// 覆盖索引
	fmt.Println("覆盖索引 (Covering Index):")
	fmt.Println("  - 如果查询的列都在索引中, 就不需要回表")
	fmt.Println("  - 例: 索引(name), 查询 SELECT id, name FROM users WHERE name='Bob'")
	fmt.Println("  - 二级索引叶子已经包含name和id, 无需回表!")
	fmt.Println("  - Explain中Extra列显示 'Using index' 表示使用了覆盖索引")
	fmt.Println()

	// 模拟覆盖索引判断
	type IndexInfo struct {
		Name    string
		Columns []string
	}
	type QueryInfo struct {
		SelectColumns []string
		WhereColumns  []string
	}

	index := IndexInfo{Name: "idx_name", Columns: []string{"name"}}
	query := QueryInfo{SelectColumns: []string{"id", "name"}, WhereColumns: []string{"name"}}

	// 检查是否覆盖索引: 查询需要的所有列是否都在索引中
	allNeeded := make(map[string]bool)
	for _, col := range query.SelectColumns {
		allNeeded[col] = true
	}
	for _, col := range query.WhereColumns {
		allNeeded[col] = true
	}

	isCovering := true
	for col := range allNeeded {
		found := false
		for _, idxCol := range index.Columns {
			if col == idxCol || col == "id" { // 二级索引自动包含主键
				found = true
				break
			}
		}
		if !found {
			isCovering = false
		}
	}

	if isCovering {
		fmt.Println("  查询 SELECT id, name FROM users WHERE name='Bob' 使用覆盖索引, 无需回表!")
	} else {
		fmt.Println("  该查询需要回表")
	}
	fmt.Println()
}

// ============================================================
// 3. 索引优化策略
// 最左前缀匹配, 索引下推(ICP), 索引合并, 函数索引
// ============================================================

func indexOptimization() {
	fmt.Println("【3. 索引优化策略】")
	fmt.Println(strings.Repeat("-", 50))

	// 最左前缀匹配
	fmt.Println("最左前缀匹配原则:")
	fmt.Println("  联合索引 (a, b, c) 实际上相当于创建了三个索引:")
	fmt.Println("    - (a)")
	fmt.Println("    - (a, b)")
	fmt.Println("    - (a, b, c)")
	fmt.Println("  WHERE条件必须从最左列开始, 不能跳过")
	fmt.Println()

	// 模拟联合索引匹配判断
	type CombinedIndex struct {
		Columns []string
	}

	queries := []struct {
		where   string
		canUse  []string // 能用到的索引前缀
		explain string
	}{
		{"a = 1", []string{"a"}, "能用索引第1列"},
		{"a = 1 AND b = 2", []string{"a", "b"}, "能用索引前2列"},
		{"a = 1 AND b = 2 AND c = 3", []string{"a", "b", "c"}, "能用全部3列"},
		{"b = 2", []string{}, "跳过了a, 无法使用索引"},
		{"b = 2 AND c = 3", []string{}, "跳过了a, 无法使用索引"},
		{"a = 1 AND c = 3", []string{"a"}, "只能用到a列 (跳过b, c用不上)"},
	}

	idx := CombinedIndex{Columns: []string{"a", "b", "c"}}
	fmt.Printf("  联合索引: %v\n\n", idx.Columns)
	for _, q := range queries {
		fmt.Printf("  WHERE %-25s -> %s\n", q.where, q.explain)
		if len(q.canUse) > 0 {
			fmt.Printf("     使用索引前缀: (%s)\n", strings.Join(q.canUse, ", "))
		}
	}
	fmt.Println()

	// 索引下推 (Index Condition Pushdown, ICP)
	fmt.Println("索引下推 (ICP, Index Condition Pushdown):")
	fmt.Println("  MySQL 5.6+引入的优化")
	fmt.Println("  没有ICP时:")
	fmt.Println("    存储引擎根据索引找到行 -> 回表 -> Server层判断WHERE条件")
	fmt.Println("  有ICP时:")
	fmt.Println("    存储引擎在索引中就判断能用的条件, 不满足的不回表")
	fmt.Println()
	fmt.Println("  示例: 索引(name, age), 查询 WHERE name LIKE '张%' AND age > 20")
	fmt.Println("  无ICP: 先用name前缀找所有'张%'开头的行, 逐个回表, 再过滤age")
	fmt.Println("  有ICP: 在索引中就检查age>20, 不满足的不回表, 减少回表次数")
	fmt.Println()

	// 索引合并
	fmt.Println("索引合并 (Index Merge):")
	fmt.Println("  当WHERE中有多个条件分别命中不同索引时, MySQL可以合并结果")
	fmt.Println("  类型:")
	fmt.Println("    - Index Merge Intersection: AND条件取交集")
	fmt.Println("    - Index Merge Union: OR条件取并集")
	fmt.Println("    - Sort-Union: 先排序再取并集")
	fmt.Println()
	fmt.Println("  示例:")
	fmt.Println("    索引: idx_a(a), idx_b(b)")
	fmt.Println("    SELECT * FROM t WHERE a=1 OR b=2")
	fmt.Println("    -> 分别用idx_a和idx_b查找, 合并结果集")
	fmt.Println()

	// 函数索引
	fmt.Println("函数索引:")
	fmt.Println("  MySQL 8.0+ 支持函数索引 (Functional Index)")
	fmt.Println("  之前: WHERE UPPER(name) = 'ALICE' 会导致索引失效")
	fmt.Println("  现在: CREATE INDEX idx_name_upper ON t ((UPPER(name)))")
	fmt.Println("  也可以用生成列 (Generated Column) + 普通索引替代")
	fmt.Println("    ALTER TABLE t ADD COLUMN name_upper VARCHAR(100)")
	fmt.Println("      GENERATED ALWAYS AS (UPPER(name)) STORED;")
	fmt.Println("    CREATE INDEX idx_name_upper ON t(name_upper);")
	fmt.Println()
}

// ============================================================
// 4. Explain 执行计划
// type/key/rows/Extra 字段解读, 优化案例
// ============================================================

func explainExecutionPlan() {
	fmt.Println("【4. Explain 执行计划】")
	fmt.Println(strings.Repeat("-", 50))

	// type 字段 - 访问类型, 从好到差
	fmt.Println("type 字段 (访问类型, 从优到差):")
	accessTypes := []struct {
		name  string
		desc  string
		score int
	}{
		{"system", "表中只有一行 (const的特例)", 10},
		{"const", "通过索引一次就找到 (主键/唯一索引)", 9},
		{"eq_ref", "关联查询中, 对每行使用唯一索引", 8},
		{"ref", "使用非唯一索引查找", 7},
		{"range", "索引范围扫描 (BETWEEN, >, <, IN)", 6},
		{"index", "全索引扫描 (扫描整棵索引树)", 4},
		{"ALL", "全表扫描 (最差)", 1},
	}
	for _, at := range accessTypes {
		fmt.Printf("  %-10s - %s [性能评分: %d/10]\n", at.name, at.desc, at.score)
	}
	fmt.Println()

	// key/key_len 字段
	fmt.Println("key/key_len 字段:")
	fmt.Println("  key: 实际使用的索引名")
	fmt.Println("  key_len: 使用索引的字节数, 可以判断联合索引用了几列")
	fmt.Println("  计算示例:")
	fmt.Println("    BIGINT (8字节) -> key_len = 8")
	fmt.Println("    VARCHAR(100) utf8mb4 (允许NULL) -> key_len = 100*4+2+1 = 403")
	fmt.Println()

	// rows 字段
	fmt.Println("rows 字段:")
	fmt.Println("  预估需要扫描的行数 (不是精确值)")
	fmt.Println("  乘以 filtered% 得到实际返回行数估算")
	fmt.Println()

	// Extra 字段常见值
	fmt.Println("Extra 字段 (重要信息):")
	extraValues := []struct {
		value   string
		desc    string
		isGood  bool
	}{
		{"Using index", "覆盖索引, 无需回表", true},
		{"Using where", "Server层过滤 (存储引擎返回的行需要再过滤)", false},
		{"Using index condition", "索引下推(ICP)", true},
		{"Using temporary", "使用临时表 (常出现在GROUP BY/DISTINCT)", false},
		{"Using filesort", "文件排序 (未使用索引排序)", false},
		{"Using join buffer", "关联查询无索引可用, 使用连接缓冲区", false},
	}
	for _, ev := range extraValues {
		status := "[好]"
		if !ev.isGood {
			status = "[需关注]"
		}
		fmt.Printf("  %-25s %s %s\n", ev.value, status, ev.desc)
	}
	fmt.Println()

	// 优化案例
	fmt.Println("优化案例:")
	fmt.Println("  案例1: type=ALL, 无索引")
	fmt.Println("    EXPLAIN SELECT * FROM orders WHERE user_id = 123")
	fmt.Println("    -> 添加索引: CREATE INDEX idx_user_id ON orders(user_id)")
	fmt.Println("    -> type变为ref, rows大幅减少")
	fmt.Println()
	fmt.Println("  案例2: Extra=Using filesort")
	fmt.Println("    EXPLAIN SELECT * FROM orders WHERE user_id=123 ORDER BY created_at")
	fmt.Println("    -> 联合索引: CREATE INDEX idx_uid_time ON orders(user_id, created_at)")
	fmt.Println("    -> Extra中Using filesort消失, 排序走索引")
	fmt.Println()
}

// ============================================================
// 5. 索引失效场景
// 函数计算/隐式转换/LIKE前缀通配符/OR条件/范围查询后续列
// ============================================================

func indexFailureScenarios() {
	fmt.Println("【5. 索引失效场景】")
	fmt.Println(strings.Repeat("-", 50))

	// 场景1: 函数计算
	fmt.Println("场景1: 对索引列使用函数或计算")
	fmt.Println("  -- 索引失效 (对列使用了函数):")
	fmt.Println("  SELECT * FROM users WHERE YEAR(created_at) = 2024")
	fmt.Println("  SELECT * FROM users WHERE id + 1 = 10")
	fmt.Println()
	fmt.Println("  -- 索引有效 (改写为范围查询):")
	fmt.Println("  SELECT * FROM users WHERE created_at >= '2024-01-01'")
	fmt.Println("    AND created_at < '2025-01-01'")
	fmt.Println("  SELECT * FROM users WHERE id = 9")
	fmt.Println()

	// 场景2: 隐式类型转换
	fmt.Println("场景2: 隐式类型转换")
	fmt.Println("  -- 索引列是VARCHAR, 传入整数导致隐式转换:")
	fmt.Println("  SELECT * FROM users WHERE phone = 13800138000  -- 失效!")
	fmt.Println("  -- MySQL会将phone列转为数字比较, 相当于 CAST(phone AS SIGNED)")
	fmt.Println("  -- 正确做法: 传入字符串")
	fmt.Println("  SELECT * FROM users WHERE phone = '13800138000'  -- 有效")
	fmt.Println()

	// 模拟隐式转换问题
	type ColumnDef struct {
		Name     string
		Type     string
		Value    string
		ParamType string
	}
	column := ColumnDef{Name: "phone", Type: "varchar", Value: "13800138000", ParamType: "int"}
	if column.Type == "varchar" && column.ParamType == "int" {
		fmt.Printf("  警告: 列 %s 是 %s 类型, 但传入的是 %s 类型, 索引将失效!\n",
			column.Name, column.Type, column.ParamType)
	}
	fmt.Println()

	// 场景3: LIKE前缀通配符
	fmt.Println("场景3: LIKE 前缀通配符")
	fmt.Println("  -- 索引失效 (%在前):")
	fmt.Println("  SELECT * FROM users WHERE name LIKE '%张'")
	fmt.Println("  SELECT * FROM users WHERE name LIKE '%张%'")
	fmt.Println()
	fmt.Println("  -- 索引有效 (%在后):")
	fmt.Println("  SELECT * FROM users WHERE name LIKE '张%'")
	fmt.Println("  -- 原因: B+树按前缀排序, 前缀不确定则无法利用索引有序性")
	fmt.Println()

	// 场景4: OR条件
	fmt.Println("场景4: OR条件导致索引失效")
	fmt.Println("  -- 当OR连接的列没有全部建立索引:")
	fmt.Println("  SELECT * FROM users WHERE name = '张三' OR age = 25")
	fmt.Println("  -- 如果只有 idx_name(name), age没有索引, 则整个条件无法走索引")
	fmt.Println()
	fmt.Println("  -- 解决方案1: 为OR两侧的列都建立索引 (索引合并)")
	fmt.Println("  -- 解决方案2: 改为UNION ALL")
	fmt.Println("  SELECT * FROM users WHERE name = '张三'")
	fmt.Println("  UNION ALL")
	fmt.Println("  SELECT * FROM users WHERE age = 25 AND name != '张三'")
	fmt.Println()

	// 场景5: 范围查询后的列
	fmt.Println("场景5: 联合索引中范围查询后的列无法使用索引")
	fmt.Println("  -- 联合索引 (a, b, c)")
	fmt.Println("  SELECT * FROM t WHERE a = 1 AND b > 10 AND c = 5")
	fmt.Println("  -- a可以用索引, b可以用索引(范围扫描), 但c无法继续使用索引")
	fmt.Println("  -- 原因: b的范围查询导致c在索引中不再是有序的")
	fmt.Println()
	fmt.Println("  -- 优化: 调整联合索引列顺序为 (a, c, b)")
	fmt.Println("  -- 这样 a和c都可以用到等值查询, b最后做范围扫描")
	fmt.Println()

	// 总结: 索引失效判断规则
	fmt.Println("索引失效快速判断口诀:")
	rules := []string{
		"1. 全值匹配我最爱, 最左前缀要遵守",
		"2. 带头大哥不能死, 中间兄弟不能断",
		"3. 索引列上少计算, 范围之后全失效",
		"4. LIKE百分写最右, 覆盖索引不写星",
		"5. 不等空值还有OR, 索引失效要少用",
		"6. VARCHAR引号不可省, 隐式转换索引亡",
	}
	for _, rule := range rules {
		fmt.Printf("  %s\n", rule)
	}
	fmt.Println()
}
