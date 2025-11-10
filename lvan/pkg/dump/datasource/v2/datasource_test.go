package v2

import (
    "testing"
)

// TestDatasourceAccept 测试数据源的Accept方法
// 这个测试验证v2访问者模式的正确实现
func TestDatasourceAccept(t *testing.T) {
    // Arrange: 准备测试数据和依赖
    config := NewMySQLConfig("localhost", 3306, "root", "", "testdb", "user")

    // 创建MySQL数据源
    datasource := NewMySQLDatasource(config)

    // 创建测试访问者
    visitor := &TestVisitor{
        visitMySQLCalled: false,
        visitDatasourceCalled: false,
    }

    // Act: 执行被测试的操作
    datasource.Accept(visitor)

    // Assert: 验证结果
    if !visitor.visitMySQLCalled {
        t.Error("访问者的VisitMySQL方法应该被调用")
    }

    // 注意：按照当前的实现，如果是DataVisitor，VisitDatasource不会被调用
    // 这是符合访问者模式的标准行为的
    // 如果需要测试VisitDatasource，应该使用不实现DataVisitor接口的访问者
}

// TestVisitorInterface 验证访问者接口的正确实现
func TestVisitorInterface(t *testing.T) {
    // Arrange: 准备测试数据
    config := NewMySQLConfig("localhost", 3306, "root", "", "testdb", "user")

    datasource := NewMySQLDatasource(config)

    // 测试扩展访问者
    extendedVisitor := &TestExtendedVisitor{
        visitMySQLCalled: false,
    }

    // Act: 执行操作
    datasource.Accept(extendedVisitor)

    // Assert: 验证结果
    if !extendedVisitor.visitMySQLCalled {
        t.Error("扩展访问者的VisitMySQL方法应该被调用")
    }
}

// TestDatasourceMetadata 测试数据源元数据获取
func TestDatasourceMetadata(t *testing.T) {
    // Arrange: 准备测试数据
    config := NewMySQLConfig("localhost", 3306, "root", "", "testdb", "user")

    datasource := NewMySQLDatasource(config)

    // Act: 获取元数据
    metadata := datasource.GetMetadata()

    // Assert: 验证元数据
    if metadata == nil {
        t.Fatal("元数据不应该为空")
    }

    if metadata.GetType() != "mysql" {
        t.Errorf("期望数据源类型为 'mysql'，实际为 '%s'", metadata.GetType())
    }
}


// 测试用的访问者实现
type TestVisitor struct {
    visitDatasourceCalled bool
    visitMySQLCalled      bool
}

func (v *TestVisitor) VisitDatasource(ds Datasource) {
    v.visitDatasourceCalled = true
}

func (v *TestVisitor) VisitMySQL(ds MySQLDatasource) {
    // 记录VisitMySQL被调用
    v.visitMySQLCalled = true
}

func (v *TestVisitor) VisitRedis(ds RedisDatasource) {
    // 默认实现，应该不会被调用
}

// 测试用的扩展访问者实现
type TestExtendedVisitor struct {
    visitMySQLCalled bool
}

func (v *TestExtendedVisitor) VisitDatasource(ds Datasource) {
    // 扩展访问者也应该实现基础方法，确保接口一致性
    // 在实际应用中，这个方法可能会调用特定类型的处理方法
    if mysqlDs, ok := ds.(MySQLDatasource); ok {
        v.VisitMySQL(mysqlDs)
    }
}

func (v *TestExtendedVisitor) VisitMySQL(ds MySQLDatasource) {
    v.visitMySQLCalled = true
}

func (v *TestExtendedVisitor) VisitRedis(ds RedisDatasource) {
    // 扩展访问者应该支持Redis
}