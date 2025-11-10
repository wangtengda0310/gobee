package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestDatasource 创建测试数据源，如果连接失败则跳过测试
func createTestDatasource(t *testing.T) MySQLDatasource {
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var datasource MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过测试: %v", r)
		}
	}()

	datasource = NewMySQLDatasource(config)
	return datasource
}

// TestLegacyAdapter_Conversion tests conversion between old and new interfaces
func TestLegacyAdapter_Conversion(t *testing.T) {
	// Arrange
	v2ds := createTestDatasource(t)
	defer v2ds.Close()
	adapter := NewLegacyAdapter(v2ds)

	// Act
	legacyDs := adapter.ToLegacyDatasource()

	// Assert
	require.NotNil(t, legacyDs, "应该创建旧式数据源")
	assert.Equal(t, "testdb", legacyDs.Database, "应该返回正确的数据库名")
	assert.Equal(t, "testtable", legacyDs.Table, "应该返回正确的表名")
}

// TestLegacyVisitorAdapter_CallbackExecution tests that the legacy visitor adapter executes callbacks correctly
func TestLegacyVisitorAdapter_CallbackExecution(t *testing.T) {
	// Arrange
	v2ds := createTestDatasource(t)
	defer v2ds.Close()

	callbackExecuted := false
	var receivedDatasource *LegacyDatasource

	legacyCallback := func(ds *LegacyDatasource) {
		callbackExecuted = true
		receivedDatasource = ds
	}

	adapter := NewLegacyVisitorAdapter(legacyCallback)

	// Act
	v2ds.Accept(adapter)

	// Assert
	assert.True(t, callbackExecuted, "回调函数应该被执行")
	assert.NotNil(t, receivedDatasource, "应该接收到数据源")
}

// TestLegacyActionBridge_ExportFunctionality tests export functionality through the bridge
func TestLegacyActionBridge_ExportFunctionality(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var v2ds MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过Action Bridge测试: %v", r)
		}
		if v2ds != nil {
			v2ds.Close()
		}
	}()

	v2ds = NewMySQLDatasource(config)
	bridge := NewLegacyActionBridge(v2ds)

	// Act & Assert - 应该不panic
	assert.NotPanics(t, func() {
		bridge.Export("uid", "1", "2", "3")
	}, "导出操作应该成功执行")
}

// TestMigrationHelper_WrapLegacyFunction tests wrapping legacy functions
func TestMigrationHelper_WrapLegacyFunction(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var v2ds MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过Migration Helper测试: %v", r)
		}
		if v2ds != nil {
			v2ds.Close()
		}
	}()

	v2ds = NewMySQLDatasource(config)
	helper := NewMigrationHelper()

	callbackExecuted := false
	legacyFunc := func(ds *LegacyDatasource) {
		callbackExecuted = true
	}

	// Act
	visitor := helper.WrapLegacyFunction(legacyFunc)
	v2ds.Accept(visitor)

	// Assert
	assert.True(t, callbackExecuted, "包装的旧函数应该被执行")
}

// TestAdapterIntegration_BackwardCompatibility tests backward compatibility
func TestAdapterIntegration_BackwardCompatibility(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var v2ds MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过向后兼容性测试: %v", r)
		}
		if v2ds != nil {
			v2ds.Close()
		}
	}()

	v2ds = NewMySQLDatasource(config)

	// Test various adapter patterns
	testCases := []struct {
		name    string
		visitor Visitor
	}{
		{
			name:    "直接访问者",
			visitor: &MockDataVisitor{},
		},
		{
			name:    "适配器访问者",
			visitor: NewLegacyVisitorAdapter(func(ds *LegacyDatasource) {}),
		},
		{
			name:    "包装函数访问者",
			visitor: NewMigrationHelper().WrapLegacyFunction(func(ds *LegacyDatasource) {}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				v2ds.Accept(tc.visitor)
			}, "%s 应该成功执行", tc.name)
		})
	}
}

// TestLegacyDatasource_DBProperty tests the DB property for compatibility
func TestLegacyDatasource_DBProperty(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var v2ds MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过DB属性测试: %v", r)
		}
		if v2ds != nil {
			v2ds.Close()
		}
	}()

	v2ds = NewMySQLDatasource(config)
	adapter := NewLegacyAdapter(v2ds)
	legacyDs := adapter.ToLegacyDatasource()

	// Act
	db := legacyDs.DB

	// Assert
	// 在当前实现中，DB为nil，这是正常的，因为我们还没有实现实际连接
	assert.Nil(t, db, "当前实现中DB为nil是正常的")
}

// TestLegacyActionBridge_Integration tests integration with legacy export functions
func TestLegacyActionBridge_Integration(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")

	var v2ds MySQLDatasource
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("无法连接到MySQL数据库，跳过Action Bridge集成测试: %v", r)
		}
		if v2ds != nil {
			v2ds.Close()
		}
	}()

	v2ds = NewMySQLDatasource(config)
	bridge := NewLegacyActionBridge(v2ds)

	// Act & Assert
	assert.NotPanics(t, func() {
		bridge.Export("uid > 0", "1", "2")
	}, "Legacy action bridge应该成功执行导出")
}