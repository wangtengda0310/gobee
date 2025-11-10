package v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullIntegration_ConnectionToExport 测试从连接到导出的完整流程
func TestFullIntegration_ConnectionToExport(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过完整集成测试")
	}

	// 创建测试配置
	config := NewMySQLConfig(
		"localhost",     // 根据实际环境修改
		3306,
		"root",
		"",             // 根据实际环境修改密码
		"test",         // 测试数据库
		"users",        // 测试表
	)

	t.Run("完整数据流程", func(t *testing.T) {
		// 步骤1: 创建数据源
		var datasource MySQLDatasource
		defer func() {
			if datasource != nil {
				datasource.Close()
			}
		}()

		// 尝试创建数据源（可能因数据库不存在而失败）
		defer func() {
			if r := recover(); r != nil {
				t.Skipf("无法连接到测试数据库，跳过完整集成测试: %v", r)
			}
		}()

		datasource = NewMySQLDatasource(config)
		require.NotNil(t, datasource, "应该创建数据源")

		// 步骤2: 验证连接状态
		assert.True(t, datasource.IsConnected(), "应该已连接")
		assert.NotNil(t, datasource.GetConnection(), "应该有连接对象")

		// 步骤3: 获取连接统计信息
		if dsImpl, ok := datasource.(*MySQLDatasourceImpl); ok {
			stats := dsImpl.GetConnectionStats()
			assert.True(t, stats["connected"].(bool), "统计显示应该已连接")
			t.Logf("连接统计: %+v", stats)
		}

		// 步骤4: 创建导出访问者
		options := &MySQLOptions{
			Where:        "1=1", // 基本条件
			Fields:       []string{"*"}, // 所有字段
			Limit:        5,     // 限制5条记录
			OutputFormat: "console",
		}

		visitor := NewMySQLExportVisitor(options)
		require.NotNil(t, visitor, "应该创建导出访问者")

		// 步骤5: 执行导出
		startTime := time.Now()
		datasource.Accept(visitor)
		elapsed := time.Since(startTime)

		// 步骤6: 验证导出结果
		results := visitor.GetResults()
		stats := visitor.GetStats()

		require.NotNil(t, results, "应该有导出结果")
		require.NotNil(t, stats, "应该有统计信息")

		// 验证统计信息
		assert.GreaterOrEqual(t, stats.ExportedRows, int64(0), "导出行数应该大于等于0")
		assert.GreaterOrEqual(t, stats.TotalRows, int64(0), "总行数应该大于等于0")
		assert.Equal(t, stats.ExportedRows, stats.TotalRows, "导出行数应该等于总行数")
		assert.NotEmpty(t, stats.StartTime, "应该有开始时间")
		assert.NotEmpty(t, stats.EndTime, "应该有结束时间")

		t.Logf("导出统计: 总行数=%d, 导出行数=%d, 耗时=%v, 数据大小=%d bytes",
			stats.TotalRows, stats.ExportedRows, elapsed, stats.ProcessedSize)

		// 验证结果数据
		if stats.ExportedRows > 0 {
			assert.Len(t, results, int(stats.ExportedRows), "结果数量应该与统计一致")

			// 验证第一个记录的结构
			firstRecord := results[0]
			assert.NotNil(t, firstRecord, "第一个记录不应该为空")

			// 检查元数据
			metadata := firstRecord.GetMetadata("query")
			assert.NotEmpty(t, metadata, "应该有查询元数据")
			t.Logf("查询元数据: %v", metadata)

			source := firstRecord.GetMetadata("source")
			assert.Equal(t, "mysql", source, "源应该是mysql")

			operation := firstRecord.GetMetadata("operation")
			assert.Equal(t, "export", operation, "操作应该是export")
		}
	})
}

// TestFullIntegration_QueryValidation 测试查询验证功能
func TestFullIntegration_QueryValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过查询验证测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "users")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到测试数据库，跳过查询验证测试: %v", r)
		}
		if datasource != nil {
			datasource.Close()
		}
	}()

	datasource = NewMySQLDatasource(config)

	t.Run("查询构建验证", func(t *testing.T) {
		// 测试不同类型的查询构建
		testQueries := []struct {
			name     string
			options  *MySQLOptions
			expected []string // 查询中应该包含的关键词
		}{
			{
				name: "基本查询",
				options: &MySQLOptions{
					Fields: []string{"id", "name"},
					Limit:  10,
				},
				expected: []string{"SELECT id, name", "LIMIT 10"},
			},
			{
				name: "条件查询",
				options: &MySQLOptions{
					Fields: []string{"*"},
					Where:  "age > 18",
					Limit:  5,
				},
				expected: []string{"SELECT *", "WHERE age > 18", "LIMIT 5"},
			},
			{
				name: "ID列表查询",
				options: &MySQLOptions{
					IDs:   []string{"1", "2", "3"},
					Limit: 100,
				},
				expected: []string{"WHERE id IN ('1', '2', '3')", "LIMIT 100"},
			},
		}

		for _, tq := range testQueries {
			t.Run(tq.name, func(t *testing.T) {
				visitor := NewMySQLExportVisitor(tq.options)
				query := visitor.(*mysqlExportVisitorImpl).BuildQuery(
					datasource.GetDatabase(),
					datasource.GetTable(),
				)

				for _, expected := range tq.expected {
					assert.Contains(t, query, expected, "查询应该包含: %s", expected)
				}

				t.Logf("构建的查询: %s", query)
			})
		}
	})
}

// TestFullIntegration_ErrorHandling 测试错误处理流程
func TestFullIntegration_ErrorHandling(t *testing.T) {
	t.Run("连接错误处理", func(t *testing.T) {
		// 创建无效配置
		invalidConfig := NewMySQLConfig("nonexistent-host", 99999, "invalid", "invalid", "invalid", "invalid")

		// 应该panic创建数据源
		assert.Panics(t, func() {
			NewMySQLDatasource(invalidConfig)
		}, "无效配置应该panic")
	})

	t.Run("导出错误处理", func(t *testing.T) {
		// 创建有效配置但可能没有真实数据库
		config := NewMySQLConfig("localhost", 3306, "root", "", "nonexistent_db", "nonexistent_table")

		var datasource MySQLDatasource
		defer func() {
			if r := recover(); r != nil {
				// 预期的连接失败
				t.Logf("预期的连接失败: %v", r)
			}
			if datasource != nil {
				datasource.Close()
			}
		}()

		// 尝试创建数据源
		datasource = NewMySQLDatasource(config)

		// 如果连接成功，继续测试导出错误处理
		if datasource != nil && datasource.IsConnected() {
			options := &MySQLOptions{
				Where:        "1=1",
				OutputFormat: "console",
			}

			visitor := NewMySQLExportVisitor(options)

			// 执行导出（可能会因表不存在而失败）
			assert.NotPanics(t, func() {
				datasource.Accept(visitor)
			}, "导出应该能够优雅地处理错误")

			// 检查错误统计
			stats := visitor.GetStats()
			if stats.ErrorRows > 0 {
				t.Logf("检测到错误行数: %d", stats.ErrorRows)
			}
		}
	})
}

// TestFullIntegration_Performance 测试性能表现
func TestFullIntegration_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "users")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到测试数据库，跳过性能测试: %v", r)
		}
		if datasource != nil {
			datasource.Close()
		}
	}()

	datasource = NewMySQLDatasource(config)

	t.Run("连接性能", func(t *testing.T) {
		// 测试连接创建性能
		iterations := 10
		totalTime := time.Duration(0)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			ds := NewMySQLDatasource(config)
			ds.Close()
			totalTime += time.Since(start)
		}

		avgTime := totalTime / time.Duration(iterations)
		t.Logf("平均连接创建时间: %v", avgTime)
		assert.Less(t, avgTime, 1*time.Second, "连接创建应该在1秒内完成")
	})

	t.Run("导出性能", func(t *testing.T) {
		options := &MySQLOptions{
			Where:        "1=1",
			Limit:        100, // 较大的数据集
			OutputFormat: "console",
		}

		// 测试多次导出的性能
		iterations := 5
		totalTime := time.Duration(0)
		totalRows := int64(0)

		for i := 0; i < iterations; i++ {
			visitor := NewMySQLExportVisitor(options)

			start := time.Now()
			datasource.Accept(visitor)
			elapsed := time.Since(start)

			totalTime += elapsed
			totalRows += visitor.GetStats().ExportedRows
		}

		avgTime := totalTime / time.Duration(iterations)
		t.Logf("平均导出时间: %v, 平均行数: %d, 吞吐量: %.2f rows/sec",
			avgTime, int(totalRows)/iterations, float64(totalRows)/avgTime.Seconds())
	})
}

// TestFullIntegration_ConcurrentOperations 测试并发操作
func TestFullIntegration_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "users")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到测试数据库，跳过并发测试: %v", r)
		}
		if datasource != nil {
			datasource.Close()
		}
	}()

	datasource = NewMySQLDatasource(config)

	t.Run("并发连接访问", func(t *testing.T) {
		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Goroutine %d panic: %v", id, r)
					}
					done <- true
				}()

				// 并发获取连接
				conn := datasource.GetConnection()
				assert.NotNil(t, conn, "Goroutine %d 应该获取到连接", id)

				// 并发检查连接状态
				connected := datasource.IsConnected()
				assert.True(t, connected, "Goroutine %d 应该显示已连接", id)
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < concurrency; i++ {
			select {
			case <-done:
				// 正常完成
			case <-time.After(5 * time.Second):
				t.Fatal("并发测试超时")
			}
		}
	})
}

// BenchmarkFullIntegration_EndToEnd 端到端基准测试
func BenchmarkFullIntegration_EndToEnd(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过基准测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "users")

	var datasource MySQLDatasource
	defer func() {
		if datasource != nil {
			datasource.Close()
		}
	}()

	// 创建数据源（在基准测试外完成）
	defer func() {
		if r := recover(); r != nil {
			b.Skipf("无法连接到测试数据库，跳过基准测试: %v", r)
		}
	}()

	datasource = NewMySQLDatasource(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			options := &MySQLOptions{
				Where:        "1=1",
				Limit:        10,
				OutputFormat: "console",
			}

			visitor := NewMySQLExportVisitor(options)
			datasource.Accept(visitor)

			// 防止编译器优化掉结果
			if len(visitor.GetResults()) == 0 {
				b.Error("导出结果为空")
			}
		}
	})
}