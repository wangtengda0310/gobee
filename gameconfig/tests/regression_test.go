package tests

import (
	"os"
	"testing"

	"github.com/wangtengda0310/gobee/gameconfig/internal/config"
	pkgconfig "github.com/wangtengda0310/gobee/gameconfig/pkg/config"
)

// 回归测试：防止已修复的 bug 再次出现
//
// 维护规范：
// 1. 每个修复的 bug 都应有对应的回归测试
// 2. 测试命名：TestRegression_<Issue编号>_<描述>
// 3. 在测试注释中记录：问题、原因、解决方案、相关提交
// 4. 本文档即是测试文档，无需额外的 .md 文件
//
// 快速查看所有回归测试：
//   grep "^func TestRegression" regression_test.go
//
// 运行回归测试：
//   go test -v ./tests/... -run TestRegression

// TestRegression_Issue1_ConcurrentLoad 测试并发加载时的竞争条件
//
// 问题描述：多个 goroutine 同时调用 Load() 时可能产生数据竞争，导致 panic 或数据不一致
//
// 复现步骤：
//   loader := config.NewLoader[Equipment](path, "sheet", opts)
//   for i := 0; i < 10; i++ {
//       go loader.Load()  // 数据竞争
//   }
//
// 根本原因：缓存未使用互斥锁保护
//
// 解决方案：添加 sync.RWMutex 保护缓存访问
//
// 修复日期：2025-02-15
//
// 相关提交：6aa9c81 feat: 添加并发安全保护和测试
//
// 测试覆盖：验证 10 个 goroutine 并发加载不会产生竞争
func TestRegression_Issue1_ConcurrentLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试")
	}

	excelFile := createIntegrationTestExcel(t)
	defer os.Remove(excelFile)

	loader := pkgconfig.NewLoader[Equipment](excelFile, "Equipment", pkgconfig.LoadOptions{
		Mode:      config.ModeExcel,
		HeaderRow: 0,
	})

	// 并发加载
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := loader.Load()
			if err != nil {
				t.Errorf("并发加载失败: %v", err)
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestRegression_Issue2_EmptySheetName 测试空 Sheet 名称处理
//
// 问题描述：Sheet 名称参数为空字符串时，程序 panic 而非返回友好错误
//
// 复现步骤：
//   loader := config.NewLoader[Equipment](path, "", opts)
//   loader.Load()  // panic
//
// 解决方案：添加空字符串检查，返回明确的错误信息
//
// 修复日期：2025-02-15
//
// 测试覆盖：验证空 Sheet 名称返回错误而非 panic
func TestRegression_Issue2_EmptySheetName(t *testing.T) {
	excelFile := createIntegrationTestExcel(t)
	defer os.Remove(excelFile)

	loader := pkgconfig.NewLoader[Equipment](excelFile, "", pkgconfig.LoadOptions{
		Mode: config.ModeExcel,
	})

	_, err := loader.Load()
	if err == nil {
		t.Error("期望返回错误，但得到 nil")
	}
}

// TestRegression_Issue3_SetMockData_Concurrent 测试 SetMockData 并发安全
//
// 问题描述：多个 goroutine 同时调用 SetMockData() 时产生数据竞争
//
// 复现步骤：
//   for i := 0; i < 10; i++ {
//       go loader.SetMockData(data)  // 竞争条件
//   }
//
// 解决方案：使用 sync.RWMutex 保护 mockData 字段
//
// 修复日期：2025-02-15
//
// 相关提交：6aa9c81 feat: 添加并发安全保护和测试
//
// 测试覆盖：验证 10 个 goroutine 并发调用 SetMockData 不会产生竞争
func TestRegression_Issue3_SetMockData_Concurrent(t *testing.T) {
	loader := pkgconfig.NewLoader[Equipment]("", "test", pkgconfig.LoadOptions{
		Mode: config.ModeMemory,
	})

	mockData := [][]string{
		{"id", "name"},
		{"1", "test"},
	}

	// 并发设置
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			loader.SetMockData(mockData)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证最终状态
	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}
	if len(items) == 0 {
		t.Error("期望有数据，但得到空")
	}
}

// TestRegression_Issue4_ConditionalField_DependencyOrder 测试条件字段依赖顺序
//
// 问题描述：条件字段的依赖字段必须在条件字段之前定义，否则 panic 或无法正确解析
//
// 错误示例：
//   type WrongOrder struct {
//       Attack int `excel:"attack,when:type=1"`  // type 还未定义
//       Type   int `excel:"type"`
//   }
//
// 解决方案：添加依赖检查，在字段映射前验证依赖关系，返回友好错误
//
// 修复日期：2025-02-15
//
// 相关提交：f4a61e9 feat: 实现条件字段解析功能 (Phase 3-4 完成)
//
// 测试覆盖：验证条件字段依赖顺序检查
func TestRegression_Issue4_ConditionalField_DependencyOrder(t *testing.T) {
	// 错误的顺序：条件字段在前
	type WrongOrder struct {
		Attack int `excel:"attack,when:type=1"`  // type 还没定义
		Type   int `excel:"type"`
	}

	excelFile := createIntegrationTestExcel(t)
	defer os.Remove(excelFile)

	loader := pkgconfig.NewLoader[WrongOrder](excelFile, "Equipment", pkgconfig.LoadOptions{
		Mode: config.ModeExcel,
	})

	_, err := loader.Load()
	if err == nil {
		t.Error("期望返回依赖错误，但得到 nil")
	}
}

// TestRegression_Issue5_MockData_EmptyData 测试空 Mock 数据处理
//
// 问题描述：Mock 数据为空数组时，程序可能产生异常或返回不一致结果
//
// 复现步骤：
//   loader.SetMockData([][]string{})
//   loader.Load()  // 可能 panic
//
// 解决方案：正确处理空数据情况，返回空结果而非 panic
//
// 修复日期：2025-02-15
//
// 测试覆盖：验证空 Mock 数据返回空结果且不 panic
func TestRegression_Issue5_MockData_EmptyData(t *testing.T) {
	loader := pkgconfig.NewLoader[Equipment]("", "test", pkgconfig.LoadOptions{
		Mode: config.ModeMemory,
	})

	// 空的 Mock 数据
	loader.SetMockData([][]string{})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("期望空数据，得到 %d 条", len(items))
	}
}
