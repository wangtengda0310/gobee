package v2

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMySQLDatasource_RealConnection 测试真实数据库连接功能
func TestMySQLDatasource_RealConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的测试")
	}

	// 这里需要配置真实的测试数据库连接
	config := NewMySQLConfig(
		"localhost", // 根据实际环境修改
		3306,
		"root",
		"", // 根据实际环境修改密码
		"test", // 测试数据库
		"test_table", // 测试表
	)

	// 这个测试可能因为数据库不存在而失败，这是正常的
	// 主要目的是验证连接逻辑是否正确

	// 测试创建数据源（可能失败）
	t.Run("创建数据源", func(t *testing.T) {
		// 使用panic恢复来处理连接失败的情况
		defer func() {
			if r := recover(); r != nil {
				t.Logf("预期的连接失败: %v", r)
			}
		}()

		datasource := NewMySQLDatasource(config)
		require.NotNil(t, datasource, "应该创建数据源")

		// 如果连接成功，验证基本功能
		assert.True(t, datasource.IsConnected(), "应该已连接")
		assert.NotNil(t, datasource.GetConnection(), "应该有连接对象")

		// 清理
		err := datasource.Close()
		assert.NoError(t, err, "关闭连接应该成功")
	})
}

// TestMySQLDatasource_InvalidConnection 测试无效连接配置
func TestMySQLDatasource_InvalidConnection(t *testing.T) {
	testCases := []struct {
		name        string
		config      MySQLConfig
		expectPanic bool
	}{
		{
			name: "无效主机",
			config: NewMySQLConfig("nonexistent-host", 3306, "user", "pass", "db", "table"),
			expectPanic: true,
		},
		{
			name: "无效端口",
			config: NewMySQLConfig("localhost", 99999, "user", "pass", "db", "table"),
			expectPanic: true,
		},
		{
			name: "无效凭据",
			config: NewMySQLConfig("localhost", 3306, "invalid-user", "invalid-pass", "invalid-db", "table"),
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
				assert.NotPanics(t, func() {
					ds := NewMySQLDatasource(tc.config)
					ds.Close() // 清理
				}, "有效配置不应该panic")
			}
		})
	}
}

// TestMySQLDatasource_ConnectionPool 测试连接池功能
func TestMySQLDatasource_ConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "test_table")

	// 尝试创建数据源
	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到测试数据库，跳过测试: %v", r)
		}
	}()

	datasource = NewMySQLDatasource(config)
	defer datasource.Close()

	// 测试连接统计信息
	t.Run("连接统计信息", func(t *testing.T) {
		// 这个测试需要类型断言来访问具体方法
		if dsImpl, ok := datasource.(*MySQLDatasourceImpl); ok {
			stats := dsImpl.GetConnectionStats()
			assert.NotNil(t, stats, "应该有统计信息")
			assert.True(t, stats["connected"].(bool), "应该已连接")
			assert.Contains(t, stats, "open_connections", "应该包含开放连接数")
			assert.Contains(t, stats, "ping_interval", "应该包含ping间隔")
		}
	})

	// 测试ping间隔设置
	t.Run("ping间隔设置", func(t *testing.T) {
		if dsImpl, ok := datasource.(*MySQLDatasourceImpl); ok {
			originalInterval := dsImpl.GetPingInterval()

			// 设置新的ping间隔
			newInterval := 45 * time.Second
			dsImpl.SetPingInterval(newInterval)

			assert.Equal(t, newInterval, dsImpl.GetPingInterval(), "ping间隔应该更新")

			// 恢复原始间隔
			dsImpl.SetPingInterval(originalInterval)
		}
	})
}

// TestMySQLDatasource_ConnectorIntegration 测试与现有连接函数的集成
func TestMySQLDatasource_ConnectorIntegration(t *testing.T) {
	// 创建配置
	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "test_table")

	// 测试DSN生成
	t.Run("DSN生成", func(t *testing.T) {
		dsn := config.GetDSN()
		assert.NotEmpty(t, dsn, "DSN不应该为空")
		assert.Contains(t, dsn, "tcp(", "应该包含TCP连接")
		assert.Contains(t, dsn, "parseTime=true", "应该包含parseTime参数")
	})

	// 测试配置验证
	t.Run("配置验证", func(t *testing.T) {
		err := config.Validate()
		assert.NoError(t, err, "有效配置应该验证通过")
	})

	// 测试无效配置验证
	t.Run("无效配置验证", func(t *testing.T) {
		invalidConfig := NewMySQLConfig("", 0, "", "", "", "")
		err := invalidConfig.Validate()
		assert.Error(t, err, "无效配置应该验证失败")
	})
}

// TestMySQLDatasource_ConnectionLifecycle 测试连接生命周期
func TestMySQLDatasource_ConnectionLifecycle(t *testing.T) {
	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "test_table")

	// 测试正常生命周期
	t.Run("正常生命周期", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Skipf("无法连接到测试数据库，跳过测试: %v", r)
			}
		}()

		// 创建数据源
		datasource := NewMySQLDatasource(config)
		require.NotNil(t, datasource, "应该创建数据源")

		// 验证连接状态
		assert.True(t, datasource.IsConnected(), "应该已连接")

		// 获取连接
		conn := datasource.GetConnection()
		assert.NotNil(t, conn, "应该有连接对象")

		// 类型断言验证
		if sqlConn, ok := conn.(*sql.DB); ok {
			assert.NotNil(t, sqlConn, "应该是有效的SQL连接")

			// 测试ping
			err := sqlConn.Ping()
			if err != nil {
				t.Logf("数据库ping失败（可能数据库不存在）: %v", err)
			}
		}

		// 关闭连接
		err := datasource.Close()
		assert.NoError(t, err, "关闭连接应该成功")
	})

	// 测试重复关闭
	t.Run("重复关闭", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Skipf("无法连接到测试数据库，跳过测试: %v", r)
			}
		}()

		datasource := NewMySQLDatasource(config)
		require.NotNil(t, datasource)

		// 第一次关闭
		err1 := datasource.Close()
		assert.NoError(t, err1, "第一次关闭应该成功")

		// 第二次关闭（应该安全）
		err2 := datasource.Close()
		assert.NoError(t, err2, "重复关闭应该成功")
	})
}

// TestMySQLDatasource_ConcurrentAccess 测试并发访问安全性
func TestMySQLDatasource_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "test_table")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到测试数据库，跳过测试: %v", r)
		}
	}()

	datasource = NewMySQLDatasource(config)
	defer datasource.Close()

	// 并发测试
	t.Run("并发访问", func(t *testing.T) {
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

				// 测试并发获取连接
				conn := datasource.GetConnection()
				assert.NotNil(t, conn, "Goroutine %d 应该获取到连接", id)

				// 测试并发状态检查
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

// BenchmarkMySQLDatasource_Connection 基准测试连接性能
func BenchmarkMySQLDatasource_Connection(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过基准测试")
	}

	config := NewMySQLConfig("localhost", 3306, "root", "", "test", "test_table")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			b.Skipf("无法连接到测试数据库，跳过基准测试: %v", r)
		}
	}()

	datasource = NewMySQLDatasource(config)
	defer datasource.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn := datasource.GetConnection()
			if conn == nil {
				b.Fatal("获取连接失败")
			}
		}
	})
}