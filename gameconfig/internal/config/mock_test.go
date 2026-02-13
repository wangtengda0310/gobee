package config

import (
	"testing"
)

// TestMemoryMode 测试 Memory 模式加载
func TestMemoryMode(t *testing.T) {
	type TestItem struct {
		ID    int    `excel:"id"`
		Name  string `excel:"name,required"`
		Value int    `excel:"value,default:0"`
	}

	// 准备 mock 数据
	mockData := [][]string{
		{"id", "name", "value"},
		{"1", "test1", "100"},
		{"2", "test2", "200"},
	}

	// 创建加载器（Memory 模式）
	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	// 加载数据
	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 验证数据
	if len(items) != 2 {
		t.Fatalf("期望 2 条数据，得到 %d", len(items))
	}

	if items[0].ID != 1 {
		t.Errorf("items[0].ID = %d, want 1", items[0].ID)
	}
	if items[0].Name != "test1" {
		t.Errorf("items[0].Name = %s, want test1", items[0].Name)
	}
	if items[0].Value != 100 {
		t.Errorf("items[0].Value = %d, want 100", items[0].Value)
	}

	if items[1].ID != 2 {
		t.Errorf("items[1].ID = %d, want 2", items[1].ID)
	}
	if items[1].Name != "test2" {
		t.Errorf("items[1].Name = %s, want test2", items[1].Name)
	}
}

// TestSetMockData 测试设置 mock 数据
func TestSetMockData(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	// 创建加载器
	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode: ModeMemory,
	})

	// 设置 mock 数据
	mockData := [][]string{
		{"id", "name"},
		{"1", "item1"},
	}
	loader.SetMockData(mockData)

	// 验证数据已设置
	if len(loader.GetMockData()) != 2 {
		t.Errorf("GetMockData() 返回 %d 行，期望 2", len(loader.GetMockData()))
	}

	// 加载数据
	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) != 1 {
		t.Errorf("期望 1 条数据，得到 %d", len(items))
	}
	if items[0].ID != 1 {
		t.Errorf("items[0].ID = %d, want 1", items[0].ID)
	}
}

// TestEmptyMockData 测试空 mock 数据
func TestEmptyMockData(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	// 测试空的 MockData
	loader1 := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: [][]string{},
	})

	items1, err := loader1.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(items1) != 0 {
		t.Errorf("空 mock 数据应返回 0 条，得到 %d", len(items1))
	}

	// 测试 nil MockData
	loader2 := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: nil,
	})

	items2, err := loader2.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(items2) != 0 {
		t.Errorf("nil mock 数据应返回 0 条，得到 %d", len(items2))
	}
}

// TestMockDataMapping 测试数据正确映射到结构体
func TestMockDataMapping(t *testing.T) {
	type ComplexItem struct {
		ID       int     `excel:"id"`
		Name     string  `excel:"name,required"`
		Value     float64 `excel:"value"`
		Enabled   bool    `excel:"enabled"`
		Tags      string  `excel:"tags"`
	}

	// 准备复杂 mock 数据
	mockData := [][]string{
		{"id", "name", "value", "enabled", "tags"},
		{"1", "item1", "99.99", "true", "tag1,tag2"},
		{"2", "item2", "123.45", "false", "tag3"},
	}

	loader := NewLoader[ComplexItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("期望 2 条数据，得到 %d", len(items))
	}

	// 验证第一条数据
	if items[0].ID != 1 {
		t.Errorf("items[0].ID = %d, want 1", items[0].ID)
	}
	if items[0].Name != "item1" {
		t.Errorf("items[0].Name = %s, want item1", items[0].Name)
	}
	if items[0].Value != 99.99 {
		t.Errorf("items[0].Value = %f, want 99.99", items[0].Value)
	}
	if !items[0].Enabled {
		t.Errorf("items[0].Enabled = %v, want true", items[0].Enabled)
	}
	if items[0].Tags != "tag1,tag2" {
		t.Errorf("items[0].Tags = %s, want tag1,tag2", items[0].Tags)
	}
}

// TestMockDataTypeConversion 测试类型转换
func TestMockDataTypeConversion(t *testing.T) {
	type TestItem struct {
		IntVal    int     `excel:"int_val"`
		FloatVal   float64 `excel:"float_val"`
		BoolVal    bool    `excel:"bool_val"`
		StringVal  string  `excel:"string_val"`
	}

	mockData := [][]string{
		{"int_val", "float_val", "bool_val", "string_val"},
		{"42", "3.14", "true", "hello"},
		{"-10", "-2.5", "false", "world"},
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 验证类型转换
	if items[0].IntVal != 42 {
		t.Errorf("items[0].IntVal = %d, want 42", items[0].IntVal)
	}
	if items[0].FloatVal != 3.14 {
		t.Errorf("items[0].FloatVal = %f, want 3.14", items[0].FloatVal)
	}
	if !items[0].BoolVal {
		t.Errorf("items[0].BoolVal = %v, want true", items[0].BoolVal)
	}
	if items[0].StringVal != "hello" {
		t.Errorf("items[0].StringVal = %s, want hello", items[0].StringVal)
	}

	if items[1].IntVal != -10 {
		t.Errorf("items[1].IntVal = %d, want -10", items[1].IntVal)
	}
	if items[1].FloatVal != -2.5 {
		t.Errorf("items[1].FloatVal = %f, want -2.5", items[1].FloatVal)
	}
	if items[1].BoolVal {
		t.Errorf("items[1].BoolVal = %v, want false", items[1].BoolVal)
	}
}

// TestMockDataWithDefaultValues 测试默认值处理
func TestMockDataWithDefaultValues(t *testing.T) {
	type TestItem struct {
		ID      int    `excel:"id"`
		Name    string `excel:"name,required"`
		Value   int    `excel:"value,default:100"`
		Enabled bool   `excel:"enabled,default:true"`
	}

	// mock 数据中缺少某些字段
	mockData := [][]string{
		{"id", "name", "value", "enabled"},
		{"1", "item1", "", ""},  // 空值应使用默认值
		{"2", "item2", "50", ""},  // value 有值，enabled 使用默认值
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 验证默认值
	if items[0].Value != 100 {
		t.Errorf("items[0].Value = %d, want 100 (default)", items[0].Value)
	}
	if !items[0].Enabled {
		t.Errorf("items[0].Enabled = %v, want true (default)", items[0].Enabled)
	}

	if items[1].Value != 50 {
		t.Errorf("items[1].Value = %d, want 50", items[1].Value)
	}
	if !items[1].Enabled {
		t.Errorf("items[1].Enabled = %v, want true (default)", items[1].Enabled)
	}
}

// TestMockDataWithSetMockData 测试使用 SetMockData 方法
func TestMockDataWithSetMockData(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	// 创建加载器（不提供 MockData）
	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode: ModeMemory,
	})

	// 初始状态：没有数据
	items1, _ := loader.Load()
	if len(items1) != 0 {
		t.Errorf("初始状态应返回 0 条，得到 %d", len(items1))
	}

	// 使用 SetMockData 设置数据
	mockData := [][]string{
		{"id", "name"},
		{"1", "first"},
		{"2", "second"},
	}
	loader.SetMockData(mockData)

	// 加载新数据
	items2, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items2) != 2 {
		t.Errorf("期望 2 条数据，得到 %d", len(items2))
	}
	if items2[0].Name != "first" {
		t.Errorf("items2[0].Name = %s, want first", items2[0].Name)
	}
	if items2[1].Name != "second" {
		t.Errorf("items2[1].Name = %s, want second", items2[1].Name)
	}

	// 更新数据
	newMockData := [][]string{
		{"id", "name"},
		{"3", "third"},
	}
	loader.SetMockData(newMockData)

	items3, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items3) != 1 {
		t.Errorf("更新后期望 1 条数据，得到 %d", len(items3))
	}
	if items3[0].ID != 3 {
		t.Errorf("items3[0].ID = %d, want 3", items3[0].ID)
	}
}

// TestMockDataWithVersionRow 测试带版本行的 mock 数据
func TestMockDataWithVersionRow(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	mockData := [][]string{
		{"__version__", "2"},
		{"__changes__", "新增了 quality 列"},
		{"id", "name"},
		{"1", "item1"},
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 应正确解析数据，跳过版本行和变更说明行
	if len(items) != 1 {
		t.Errorf("期望 1 条数据，得到 %d", len(items))
	}
	if items[0].ID != 1 {
		t.Errorf("items[0].ID = %d, want 1", items[0].ID)
	}
}

// TestClearMockData 测试清空 mock 数据
func TestClearMockData(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode: ModeMemory,
	})

	// 设置数据
	mockData := [][]string{
		{"id", "name"},
		{"1", "item1"},
	}
	loader.SetMockData(mockData)

	items1, _ := loader.Load()
	if len(items1) != 1 {
		t.Errorf("期望 1 条数据，得到 %d", len(items1))
	}

	// 清空数据
	loader.SetMockData(nil)
	items2, _ := loader.Load()
	if len(items2) != 0 {
		t.Errorf("清空后应返回 0 条数据，得到 %d", len(items2))
	}

	// 也可以使用空数组清空
	loader.SetMockData([][]string{})
	items3, _ := loader.Load()
	if len(items3) != 0 {
		t.Errorf("使用空数组清空后应返回 0 条数据，得到 %d", len(items3))
	}
}

// TestMockDataWithRequiredField 测试必填字段验证
func TestMockDataWithRequiredField(t *testing.T) {
	type TestItem struct {
		ID    int    `excel:"id"`
		Name  string `excel:"name,required"`
		Value int    `excel:"value,default:0"`
	}

	// 正常数据
	mockData := [][]string{
		{"id", "name", "value"},
		{"1", "item1", "100"},
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) != 1 {
		t.Errorf("期望 1 条数据，得到 %d", len(items))
	}

	// 测试缺少必填字段的情况
	invalidMockData := [][]string{
		{"id", "name", "value"},
		{"2", "", "50"},  // name 为空
	}

	loader2 := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: invalidMockData,
	})

	_, err = loader2.Load()
	if err == nil {
		t.Error("缺少必填字段应返回错误")
	}
}

// TestMockDataIndependence 测试不同 Loader 实例的数据独立性
func TestMockDataIndependence(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	// 创建两个独立的加载器
	loader1 := NewLoader[TestItem]("", "test1", LoadOptions{Mode: ModeMemory})
	loader2 := NewLoader[TestItem]("", "test2", LoadOptions{Mode: ModeMemory})

	// 为 loader1 设置数据
	data1 := [][]string{
		{"id", "name"},
		{"1", "first"},
	}
	loader1.SetMockData(data1)

	// 为 loader2 设置数据
	data2 := [][]string{
		{"id", "name"},
		{"2", "second"},
	}
	loader2.SetMockData(data2)

	// 验证两个加载器的数据独立
	items1, _ := loader1.Load()
	items2, _ := loader2.Load()

	if len(items1) != 1 || items1[0].Name != "first" {
		t.Errorf("loader1 数据错误: got %v", items1)
	}
	if len(items2) != 1 || items2[0].Name != "second" {
		t.Errorf("loader2 数据错误: got %v", items2)
	}
}

// TestMemoryModeWithCustomHeaderRow 测试 Memory 模式下自定义表头行
func TestMemoryModeWithCustomHeaderRow(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	mockData := [][]string{
		{"metadata", "value"},
		{"id", "name"},
		{"1", "test"},
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:      ModeMemory,
		MockData:   mockData,
		HeaderRow: 1,  // 表头在第二行（索引 1）
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) != 1 {
		t.Errorf("期望 1 条数据，得到 %d", len(items))
	}
	if items[0].ID != 1 {
		t.Errorf("items[0].ID = %d, want 1", items[0].ID)
	}
}

// TestMemoryModeReload 测试 Memory 模式下的重新加载
func TestMemoryModeReload(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	mockData := [][]string{
		{"id", "name"},
		{"1", "first"},
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: mockData,
	})

	// 第一次加载
	items1, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 重新加载
	items2, err := loader.Reload()
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// 两次加载结果应该相同
	if len(items1) != len(items2) {
		t.Errorf("Reload() 返回 %d 条，期望 %d", len(items2), len(items1))
	}
	if items1[0].ID != items2[0].ID {
		t.Errorf("Reload() 后 ID 不一致: %d vs %d", items1[0].ID, items2[0].ID)
	}
}
