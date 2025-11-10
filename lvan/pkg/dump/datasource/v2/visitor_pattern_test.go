package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVisitorPattern_BasicAcceptance tests the basic visitor pattern acceptance flow
func TestVisitorPattern_BasicAcceptance(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)
	visitor := &MockVisitor{}

	// Act
	datasource.Accept(visitor)

	// Assert
	assert.True(t, visitor.visited, "访问者应该被调用")
	assert.Equal(t, "mysql", visitor.visitedDatasourceType, "应该访问MySQL数据源")
	assert.NotNil(t, visitor.visitedDatasource, "应该传递数据源实例")
}

// TestVisitorPattern_DataVisitorDispatch tests proper dispatch to DataVisitor
func TestVisitorPattern_DataVisitorDispatch(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)
	dataVisitor := &MockDataVisitor{}

	// Act
	datasource.Accept(dataVisitor)

	// Assert
	assert.True(t, dataVisitor.visitMySQLCalled, "应该调用VisitMySQL方法")
	assert.False(t, dataVisitor.visitRedisCalled, "不应该调用VisitRedis方法")
	assert.False(t, dataVisitor.visitDatasourceCalled, "不应该调用通用VisitDatasource方法")
	assert.Equal(t, datasource, dataVisitor.visitedMySQLDatasource, "应该传递正确的MySQL数据源")
}

// TestVisitorPattern_FallbackToBaseVisitor tests fallback to base visitor when DataVisitor is not implemented
func TestVisitorPattern_FallbackToBaseVisitor(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)
	baseVisitor := &MockBaseVisitor{}

	// Act
	datasource.Accept(baseVisitor)

	// Assert
	assert.True(t, baseVisitor.visited, "应该调用基础访问者")
	assert.Equal(t, datasource, baseVisitor.visitedDatasource, "应该传递数据源实例")
}

// TestMySQLDatasource_GetMetadata tests metadata retrieval
func TestMySQLDatasource_GetMetadata(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)

	// Act
	metadata := datasource.GetMetadata()

	// Assert
	require.NotNil(t, metadata, "元数据不应该为空")
	assert.Equal(t, "mysql", metadata.GetType(), "数据源类型应该是mysql")
	tables := metadata.GetTables()
	assert.NotEmpty(t, tables, "应该包含表列表")
	assert.Contains(t, tables, "user", "应该包含user表")
}

// TestMySQLDatasource_ConfigurationAccess tests configuration access methods
func TestMySQLDatasource_ConfigurationAccess(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("test-host", 3307, "test-user", "test-password", "test-db", "test-table")
	datasource := NewMySQLDatasource(config)

	// Act & Assert
	assert.Equal(t, "test-host", datasource.GetHost(), "应该返回正确的主机名")
	assert.Equal(t, 3307, datasource.GetPort(), "应该返回正确的端口")
	assert.Equal(t, "test-user", datasource.GetUser(), "应该返回正确的用户名")
	assert.Equal(t, "test-password", datasource.GetPassword(), "应该返回正确的密码")
	assert.Equal(t, "test-db", datasource.GetDatabase(), "应该返回正确的数据库名")
	assert.Equal(t, "test-table", datasource.GetTable(), "应该返回正确的表名")
}

// Mock implementations for testing

// MockVisitor 基础模拟访问者
type MockVisitor struct {
	visited                bool
	visitedDatasource      Datasource
	visitedDatasourceType  string
}

func (m *MockVisitor) VisitDatasource(ds Datasource) {
	m.visited = true
	m.visitedDatasource = ds
	m.visitedDatasourceType = ds.GetMetadata().GetType()
}

// MockDataVisitor 数据特定模拟访问者
type MockDataVisitor struct {
	visitMySQLCalled      bool
	visitRedisCalled      bool
	visitDatasourceCalled bool
	visitedMySQLDatasource MySQLDatasource
	visitedRedisDatasource RedisDatasource
	visitedDatasource     Datasource
}

func (m *MockDataVisitor) VisitMySQL(ds MySQLDatasource) {
	m.visitMySQLCalled = true
	m.visitedMySQLDatasource = ds
}

func (m *MockDataVisitor) VisitRedis(ds RedisDatasource) {
	m.visitRedisCalled = true
	m.visitedRedisDatasource = ds
}

func (m *MockDataVisitor) VisitDatasource(ds Datasource) {
	m.visitDatasourceCalled = true
	m.visitedDatasource = ds
}

// MockBaseVisitor 基础模拟访问者（不实现DataVisitor）
type MockBaseVisitor struct {
	visited         bool
	visitedDatasource Datasource
}

func (m *MockBaseVisitor) VisitDatasource(ds Datasource) {
	m.visited = true
	m.visitedDatasource = ds
}

// TestVisitorPattern_InvalidVisitorPanic tests that panic occurs when visitor doesn't handle the datasource properly
func TestVisitorPattern_InvalidVisitorPanic(t *testing.T) {
	// Arrange
	config := NewMySQLConfig("localhost", 3306, "test", "password", "testdb", "testtable")
	datasource := NewMySQLDatasource(config)
	invalidVisitor := &InvalidVisitor{}

	// Act & Assert
	assert.Panics(t, func() {
		datasource.Accept(invalidVisitor)
	}, "当访问者无法处理数据源时应该panic")
}

// InvalidVisitor 无效访问者，用于测试panic情况
type InvalidVisitor struct{}

func (m *InvalidVisitor) VisitDatasource(ds Datasource) {
	panic("无效访问者无法处理任何数据源")
}