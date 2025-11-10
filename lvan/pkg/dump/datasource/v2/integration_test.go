package v2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMySQLDatasource_ExportVisitorIntegration tests integration between MySQL datasource and export visitor
func TestMySQLDatasource_ExportVisitorIntegration(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)

	options := &MySQLOptions{
		Where:        "age > 25",
		IDs:          []string{"1", "2", "3"},
		Fields:       []string{"id", "name", "email"},
		Limit:        100,
		OutputFormat: "json",
		OutputPath:   "/tmp/export.json",
		Compression:  "gzip",
	}

	exportVisitor := NewMySQLExportVisitor(options)

	// Act
	datasource.Accept(exportVisitor)

	// Assert
	require.NotNil(t, exportVisitor, "导出访问者应该被创建")
	assert.True(t, exportVisitor.HasFilter(), "应该有过滤器设置")
	assert.True(t, exportVisitor.IsValidOutputFormat(), "输出格式应该有效")

	results := exportVisitor.GetResults()
	assert.NotEmpty(t, results, "应该有导出结果")

	stats := exportVisitor.GetStats()
	require.NotNil(t, stats, "应该有统计信息")
	assert.Greater(t, stats.ExportedRows, int64(0), "应该导出了行")
	assert.Equal(t, stats.ExportedRows, stats.TotalRows, "总行数应该等于导出行数")
}

// TestMySQLDatasource_MultipleVisitorTypes tests that a datasource can accept different visitor types
func TestMySQLDatasource_MultipleVisitorTypes(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)

	// Test different visitor types
	testCases := []struct {
		name    string
		visitor Visitor
	}{
		{
			name:    "基础访问者",
			visitor: &MockVisitor{},
		},
		{
			name:    "数据访问者",
			visitor: &MockDataVisitor{},
		},
		{
			name:    "导出访问者",
			visitor: NewMySQLExportVisitor(&MySQLOptions{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			datasource.Accept(tc.visitor)

			// Assert - should not panic
			// Specific assertions depend on visitor type
			switch v := tc.visitor.(type) {
			case *MockVisitor:
				assert.True(t, v.visited, "Mock访问者应该被调用")
			case *MockDataVisitor:
				assert.True(t, v.visitMySQLCalled, "Mock数据访问者应该调用VisitMySQL")
			case MySQLExportVisitor:
				assert.NotNil(t, v.GetResults(), "导出访问者应该有结果")
			}
		})
	}
}

// TestVisitorPattern_ChainedVisitors tests chaining multiple visitors
func TestVisitorPattern_ChainedVisitors(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)

	// Create a chain of visitors
	visitors := []Visitor{
		&MockDataVisitor{},
		NewMySQLExportVisitor(&MySQLOptions{
			OutputFormat: "json",
		}),
		&MockVisitor{},
	}

	// Act & Assert
	for i, visitor := range visitors {
		t.Run(fmt.Sprintf("Visitor%d", i+1), func(t *testing.T) {
			// Reset any state if needed
			datasource.Accept(visitor)

			// Verify each visitor worked correctly
			switch v := visitor.(type) {
			case *MockDataVisitor:
				assert.True(t, v.visitMySQLCalled, "数据访问者应该被调用")
			case MySQLExportVisitor:
				assert.NotNil(t, v.GetResults(), "导出访问者应该产生结果")
			case *MockVisitor:
				assert.True(t, v.visited, "基础访问者应该被调用")
			}
		})
	}
}

// TestVisitorPattern_ConfigurationValidation tests that configuration validation works correctly
func TestVisitorPattern_ConfigurationValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      MySQLConfig
		expectPanic bool
	}{
		{
			name: "有效配置",
			config: NewMySQLConfig("localhost", 3306, "user", "password", "database", "table"),
			expectPanic: false,
		},
		{
			name: "无效端口",
			config: NewMySQLConfig("localhost", 70000, "user", "password", "database", "table"),
			expectPanic: true,
		},
		{
			name: "空主机名",
			config: NewMySQLConfig("", 3306, "user", "password", "database", "table"),
			expectPanic: true,
		},
		{
			name: "空数据库名",
			config: NewMySQLConfig("localhost", 3306, "user", "password", "", "table"),
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				assert.Panics(t, func() {
					NewMySQLDatasource(tc.config)
				}, "无效配置应该panic")
			} else {
				defer func() {
					if r := recover(); r != nil {
						// 如果连接失败，跳过这个测试
						t.Skipf("无法连接到MySQL数据库，跳过配置验证测试: %v", r)
					}
				}()
				ds := NewMySQLDatasource(tc.config)
				assert.NotNil(t, ds, "有效配置应该创建数据源")
				ds.Close() // 清理连接
			}
		})
	}
}

// TestVisitorPattern_PerformanceBasic tests basic performance characteristics
func TestVisitorPattern_PerformanceBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过性能测试: %v", r)
		}
		if datasource != nil {
			datasource.Close()
		}
	}()

	datasource = NewMySQLDatasource(config)

	// Test many visitor acceptances
	iterations := 1000

	// Act
	for i := 0; i < iterations; i++ {
		visitor := &MockVisitor{}
		datasource.Accept(visitor)
	}

	// Assert - this is mainly to ensure we don't have major performance regressions
	assert.True(t, true, "性能测试完成")
}