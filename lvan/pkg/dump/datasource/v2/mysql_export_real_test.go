package v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMySQLExportVisitor_RealDatabase 测试真实数据库导出功能
func TestMySQLExportVisitor_RealDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的测试")
	}

	// 配置测试数据库连接
	config := NewMySQLConfig(
		"localhost",     // 根据实际环境修改
		3306,
		"root",
		"",             // 根据实际环境修改密码
		"test_db",      // 测试数据库
		"test_table",   // 测试表
	)

	// 尝试创建数据源
	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到测试数据库，跳过测试: %v", r)
		}
		if datasource != nil {
			datasource.Close()
		}
	}()

	datasource = NewMySQLDatasource(config)

	// 测试导出功能
	t.Run("基本导出功能", func(t *testing.T) {
		options := &MySQLOptions{
			Where:        "1=1", // 基本条件，获取所有数据
			Fields:       []string{"*"}, // 所有字段
			Limit:        10,      // 限制10条记录
			OutputFormat: "console", // 控制台输出
		}

		visitor := NewMySQLExportVisitor(options)
		require.NotNil(t, visitor, "应该创建导出访问者")

		// 执行导出
		datasource.Accept(visitor)

		// 验证结果
		results := visitor.GetResults()
		stats := visitor.GetStats()

		assert.NotNil(t, results, "应该有导出结果")
		assert.NotNil(t, stats, "应该有统计信息")
		assert.True(t, stats.ExportedRows >= 0, "导出行数应该大于等于0")

		if stats.ExportedRows > 0 {
			t.Logf("成功导出 %d 行数据", stats.ExportedRows)
		} else {
			t.Log("导出行为空，可能表不存在或没有数据")
		}
	})

	// 测试条件导出
	t.Run("条件导出", func(t *testing.T) {
		options := &MySQLOptions{
			Where:        "id > 0", // 假设有id字段
			IDs:          []string{"1", "2", "3"},
			Fields:       []string{"id", "name"}, // 假设这些字段存在
			Limit:        5,
			OutputFormat: "console",
		}

		visitor := NewMySQLExportVisitor(options)
		datasource.Accept(visitor)

		results := visitor.GetResults()
		stats := visitor.GetStats()

		assert.NotNil(t, results, "应该有导出结果")
		assert.NotNil(t, stats, "应该有统计信息")

		// 验证查询构建
		builtQuery := visitor.(*mysqlExportVisitorImpl).BuildQuery(
			datasource.GetDatabase(),
			datasource.GetTable(),
		)
		assert.Contains(t, builtQuery, "WHERE", "查询应该包含WHERE条件")
		assert.Contains(t, builtQuery, "LIMIT 5", "查询应该包含LIMIT")
	})
}

// TestMySQLExportVisitor_QueryBuilding 测试查询构建逻辑
func TestMySQLExportVisitor_QueryBuilding(t *testing.T) {
	testCases := []struct {
		name     string
		options  *MySQLOptions
		expected string
	}{
		{
			name: "基本查询",
			options: &MySQLOptions{
				Fields: []string{"id", "name"},
				Limit:  10,
			},
			expected: "SELECT id, name FROM test_db.test_table LIMIT 10",
		},
		{
			name: "带WHERE条件",
			options: &MySQLOptions{
				Fields: []string{"*"},
				Where:  "age > 18",
				Limit:  5,
			},
			expected: "SELECT * FROM test_db.test_table WHERE age > 18 LIMIT 5",
		},
		{
			name: "带ID列表",
			options: &MySQLOptions{
				Fields: []string{"id", "email"},
				IDs:    []string{"1", "2", "3"},
				Limit:  100,
			},
			expected: "SELECT id, email FROM test_db.test_table WHERE id IN ('1', '2', '3') LIMIT 100",
		},
		{
			name: "复合条件",
			options: &MySQLOptions{
				Fields: []string{"id", "name", "email"},
				Where:  "status = 'active'",
				IDs:    []string{"1", "5"},
				Limit:  50,
			},
			expected: "SELECT id, name, email FROM test_db.test_table WHERE status = 'active' AND id IN ('1', '5') LIMIT 50",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			visitor := NewMySQLExportVisitor(tc.options)
			query := visitor.(*mysqlExportVisitorImpl).BuildQuery("test_db", "test_table")
			assert.Equal(t, tc.expected, query, "查询构建应该正确")

			// 使用visitor变量避免编译警告
			_ = visitor.GetStats()
		})
	}
}

// TestMySQLExportVisitor_ErrorHandling 测试错误处理
func TestMySQLExportVisitor_ErrorHandling(t *testing.T) {
	// 测试无效配置
	t.Run("无效配置处理", func(t *testing.T) {
		// 这个测试主要验证配置验证逻辑
		options := &MySQLOptions{
			Where:        "", // 空条件是允许的
			Fields:       []string{}, // 空字段列表应该使用*
			Limit:        -1,     // 负数limit应该被忽略或重置
			OutputFormat: "invalid_format", // 无效格式应该有默认处理
		}

		visitor := NewMySQLExportVisitor(options)
		require.NotNil(t, visitor, "应该创建访问者")

		// 验证配置被正确处理
		assert.False(t, visitor.HasFilter(), "没有过滤器时应该返回false")
		assert.False(t, visitor.IsValidOutputFormat(), "无效格式应该返回false")
	})

	// 测试Redis访问（应该panic）
	t.Run("Redis访问错误处理", func(t *testing.T) {
		options := &MySQLOptions{}
		visitor := NewMySQLExportVisitor(options)

		// 创建一个模拟的Redis数据源
		redisConfig := NewRedisConfig("localhost", 6379, "", 0, "*")
		redisDatasource := NewRedisDatasource(redisConfig)

		assert.Panics(t, func() {
			redisDatasource.Accept(visitor)
		}, "MySQL导出访问者不支持Redis数据源应该panic")

		// 使用visitor变量避免编译警告
		_ = visitor.GetStats()
	})
}

// TestMySQLExportVisitor_Statistics 测试统计功能
func TestMySQLExportVisitor_Statistics(t *testing.T) {
	options := &MySQLOptions{
		Where:        "id > 0",
		IDs:          []string{"1", "2"},
		Limit:        10,
		OutputFormat: "console",
	}

	visitor := NewMySQLExportVisitor(options)

	// 初始统计
	stats := visitor.GetStats()
	require.NotNil(t, stats, "应该有统计信息")
	assert.Equal(t, int64(0), stats.ExportedRows, "初始导出行数应该为0")
	assert.Equal(t, int64(0), stats.TotalRows, "初始总行数应该为0")
	assert.Equal(t, int64(0), stats.ErrorRows, "初始错误行数应该为0")
	assert.Equal(t, int64(0), stats.ProcessedSize, "初始处理大小应该为0")

	// 验证统计结构
	assert.NotEmpty(t, stats.StartTime, "统计应该包含开始时间")
	assert.Equal(t, "", stats.EndTime, "结束时间初始应该为空")
}

// TestMySQLExportVisitor_Performance 测试性能相关功能
func TestMySQLExportVisitor_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	// 测试大量数据处理
	t.Run("大量数据处理", func(t *testing.T) {
		options := &MySQLOptions{
			Where:        "1=1",
			Limit:        1000, // 较大的limit
			OutputFormat: "console",
		}

		start := time.Now()
		visitor := NewMySQLExportVisitor(options)
		elapsed := time.Since(start)

		assert.Less(t, elapsed, 100*time.Millisecond, "创建访问者应该很快")

		// 使用visitor变量避免编译警告
		_ = visitor.GetOptions()

		// 如果有真实数据库连接，可以测试实际查询性能
		// 这里主要测试创建时间
	})
}

// TestMySQLExportVisitor_OutputFormats 测试输出格式
func TestMySQLExportVisitor_OutputFormats(t *testing.T) {
	supportedFormats := []string{"zip", "dir", "sql-tpl", "csv", "json"}

	for _, format := range supportedFormats {
		t.Run("输出格式_"+format, func(t *testing.T) {
			options := &MySQLOptions{
				OutputFormat: format,
				OutputPath:   "/tmp/test." + format,
			}

			visitor := NewMySQLExportVisitor(options)
			assert.True(t, visitor.IsValidOutputFormat(), "%s 格式应该被支持", format)
		})
	}

	// 测试不支持的格式
	t.Run("不支持的输出格式", func(t *testing.T) {
		options := &MySQLOptions{
			OutputFormat: "unsupported_format",
		}

		visitor := NewMySQLExportVisitor(options)
		assert.False(t, visitor.IsValidOutputFormat(), "不支持的格式应该返回false")
	})
}