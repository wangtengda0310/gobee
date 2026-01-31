# LVAN TestDB - MySQL测试数据库框架

## 概述

LVAN TestDB 是一个基于内存模拟的MySQL测试数据库框架，提供MySQL协议支持，无需外部数据库服务器即可进行完整的数据库功能测试。

## 特性

- **双模式支持**: 自动检测真实MySQL服务器，失败时切换到内存模拟模式
- **真实协议支持**: 基于标准MySQL协议，提供真实的数据库体验
- **配置驱动**: 支持YAML配置文件和程序化配置定义表结构和测试数据
- **内存模拟**: 完全在内存中运行，启动快速，清理简单
- **类型安全**: 强类型接口设计，编译时错误检查
- **易于集成**: 简单的API设计，易于集成到现有测试中
- **向后兼容**: 与现有测试框架完全兼容，可逐步迁移

## 快速开始

### 基本使用

```go
package main

import (
    "testing"
    "github.com/wangtengda0310/gobee/lvan/pkg/testdb"
)

func TestYourDatabaseFunction(t *testing.T) {
    // 设置测试环境
    if err := testdb.SetupTestMySQLEnvironment(); err != nil {
        t.Fatalf("设置测试环境失败: %v", err)
    }
    defer testdb.TeardownTestMySQLEnvironment()

    // 获取MySQL测试数据库
    adapter := testdb.GetGlobalMySQLTestDBAdapter()
    mysqlDB := adapter.GetMySQLTestDB()

    // 获取数据库连接
    conn := mysqlDB.GetMySQLConnection()
    if conn.DB != nil {
        defer conn.DB.Close()
    }

    // 执行你的测试逻辑...
}
```

### 使用YAML配置

```go
func TestWithYAMLConfig(t *testing.T) {
    adapter := testdb.GetGlobalMySQLTestDBAdapter()

    // 从YAML配置文件创建测试数据库
    err := adapter.SetupMySQLTestDB("mysql_test.yaml")
    if err != nil {
        t.Fatalf("设置MySQL测试数据库失败: %v", err)
    }
    defer adapter.Cleanup()

    // 测试逻辑...
}
```

## 配置文件格式

### MySQL配置示例

```yaml
mysql:
  server:
    port: 3307
    database: testdb
    user: root
    password: ""
    host: 127.0.0.1

  schemas:
    - name: user
      columns:
        - name: uid
          type: INT
          primary: true
          nullable: false
          auto_increment: true
        - name: accountid
          type: VARCHAR(50)
          nullable: false
          unique: true
        - name: data
          type: BLOB
          nullable: true
      indexes:
        - name: idx_accountid
          columns: [accountid]
          unique: true
      data:
        - uid: 1
          accountid: user001
          data: "SGVsbG8gV29ybGQhIQ=="
          created_at: "2025-01-10 10:00:00"
        - uid: 2
          accountid: user002
          data: "0x48656c6c6f576f726c642121"
          created_at: "2025-01-10 11:00:00"

  variables:
    test_user_count: 100
    default_balance: 1000.00
```

## API 参考

### 工厂接口

```go
type TestDatabaseFactory interface {
    CreateMySQLFromConfig(configPath string) (MySQLTestDatabase, error)
    CreateMySQL(config MySQLConfig) (MySQLTestDatabase, error)
}
```

### MySQL测试数据库接口

```go
type MySQLTestDatabase interface {
    Start() error
    Stop() error
    GetMySQLConnection() *MySQLConnection
    CreateTable(schema *TableSchema) error
    InsertData(tableName string, data []map[string]interface{}) error
    Query(query string, args ...interface{}) ([]map[string][]byte, error)
    Clear() error
    LoadConfig(configPath string) error
}
```

### MySQL连接包装器

```go
type MySQLConnection struct {
    DB *sql.DB  // 真实数据库连接（模拟模式下为nil）
    DSN string    // 数据源连接字符串
}
```

## 与LVAN Dumper集成

### 优化现有测试

```go
// 原始测试
func TestDatasourceAccept(t *testing.T) {
    config := NewMySQLConfig("localhost", 3306, "root", "", "testdb", "user")

    var datasource MySQLDatasource
    defer func() {
        if r := recover(); r != nil {
            t.Skipf("无法连接到MySQL数据库，跳过测试: %v", r)
        }
    }()

    datasource = NewMySQLDatasource(config)
    // 测试逻辑...
}

// 优化后的测试
func TestDatasourceAccept(t *testing.T) {
    // 首先尝试使用模拟服务器
    if err := SetupTestMySQLEnvironment(); err == nil {
        defer TeardownTestMySQLEnvironment()
        testDatasourceAcceptWithMockServer(t)
        return
    }

    // 回退到原始测试
    testDatasourceAcceptWithRealDatabase(t)
}

func testDatasourceAcceptWithMockServer(t *testing.T) {
    config := NewMySQLConfig("127.0.0.1", 3307, "root", "", "testdb", "user")
    datasource := NewMySQLDatasource(config)
    defer datasource.Close()

    // 测试逻辑...
}
```

## 最佳实践

### 测试环境管理

```go
func TestSuite(t *testing.T) {
    // 在测试开始前设置环境
    if err := SetupTestMySQLEnvironment(); err != nil {
        t.Fatalf("设置测试环境失败: %v", err)
    }
    defer TeardownTestMySQLEnvironment()

    // 运行多个子测试
    t.Run("测试1", testFunction1)
    t.Run("测试2", testFunction2)
}
```

### 数据隔离

```go
func TestWithCleanState(t *testing.T) {
    adapter := GetGlobalMySQLTestDBAdapter()
    mysqlDB := adapter.GetMySQLTestDB()

    // 清理数据确保测试独立性
    defer mysqlDB.Clear()

    // 测试逻辑...
}
```

## 技术架构

### 双模式设计

1. **真实模式**: 连接真实MySQL服务器时使用
2. **模拟模式**: 连接失败时使用内存模拟
3. **自动切换**: 框架自动检测并选择合适的模式

### 配置驱动架构

- **YAML配置**: 支持复杂的表结构和数据定义
- **程序化配置**: 支持代码创建的配置对象
- **灵活验证**: 完善的配置验证和错误处理

### 模拟模式特性

- **内存存储**: 完全在内存中运行
- **SQL解析**: 支持基本SQL语句解析
- **类型映射**: 支持多种MySQL数据类型
- **数据持久化**: 支持表结构创建和测试数据插入

## 故障排除

### 常见问题

1. **连接失败**: 模拟模式下自动处理，不影响测试
2. **配置错误**: 提供详细的错误信息
3. **SQL解析错误**: 支持基本的SELECT、INSERT、UPDATE、DELETE语句

### 调试技巧

```go
func TestWithDebugging(t *testing.T) {
    if err := SetupTestMySQLEnvironment(); err != nil {
        t.Logf("调试信息: %v", err)
    }
    defer TeardownTestMySQLEnvironment()

    // 检查连接状态
    adapter := GetGlobalMySQLTestDBAdapter()
    mysqlDB := adapter.GetMySQLTestDB()
    if mysqlDB != nil {
        conn := mysqlDB.GetMySQLConnection()
        t.Logf("连接状态: %v", conn.DSN == "")
    }
}
```

## 性能考虑

- **启动时间**: 模拟模式启动通常 < 10ms
- **内存使用**: 内存数据库完全在内存中运行
- **并发测试**: 支持并发测试

## 总结

LVAN TestDB为LVAN Dumper项目提供了：

1. **统一的测试环境**: 为v2数据源测试提供一致的测试基础
2. **无外部依赖**: 无需MySQL服务器即可进行完整测试
3. **真实体验**: 使用MySQL协议而非简单Mock
4. **易于使用**: 简单的API设计，降低学习成本
5. **高质量**: 完善的测试覆盖和错误处理

这个框架专注于MySQL测试，为项目提供可靠的测试基础设施。

---

*文档版本：v2.0*
*最后更新：2025-01-10*
*维护者：LVAN Dumper 开发团队*