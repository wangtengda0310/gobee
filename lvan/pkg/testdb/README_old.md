# LVAN TestDB - MySQL测试数据库框架

## 概述

LVAN TestDB 是一个基于 `go-mysql-server` 的MySQL测试数据库框架，提供真实的MySQL协议支持，支持内存模拟模式和真实数据库连接两种模式，无需外部数据库服务器即可进行完整的数据库功能测试。

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
    adapter := testdb.GetGlobalTestDBAdapter()

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
```

### Redis配置示例

```yaml
redis:
  server:
    port: 6380
    host: 127.0.0.1
    db: 0

  data:
    strings:
      - key: "user:1"
        value: '{"id":1,"name":"alice"}'
        ttl: 3600
    hashes:
      - key: "user:1:profile"
        fields:
          id: "1"
          name: "alice"
          email: "alice@example.com"
        ttl: 3600
```

## API 参考

### 工厂接口

```go
type TestDatabaseFactory interface {
    CreateMySQLFromConfig(configPath string) (MySQLTestDatabase, error)
    CreateRedisFromConfig(configPath string) (RedisTestDatabase, error)
    CreateMySQL(config MySQLConfig) (MySQLTestDatabase, error)
    CreateRedis(config RedisConfig) (RedisTestDatabase, error)
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
}
```

### Redis测试数据库接口

```go
type RedisTestDatabase interface {
    Start() error
    Stop() error
    GetRedisClient() *redis.Client
    SetData(key string, value interface{}, ttl time.Duration) error
    SetHashData(key string, fields map[string]interface{}) error
    GetData(key string) (interface{}, error)
    GetKeys(pattern string) ([]string, error)
    Clear() error
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
    if err := SetupTestEnvironment(); err == nil {
        defer TeardownTestEnvironment()
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

### 1. 测试环境管理

```go
func TestSuite(t *testing.T) {
    // 在测试开始前设置环境
    if err := SetupTestEnvironment(); err != nil {
        t.Fatalf("设置测试环境失败: %v", err)
    }
    defer TeardownTestEnvironment()

    // 运行多个子测试
    t.Run("测试1", testFunction1)
    t.Run("测试2", testFunction2)
}
```

### 2. 数据隔离

```go
func TestWithCleanState(t *testing.T) {
    adapter := GetGlobalTestDBAdapter()
    mysqlDB := adapter.GetMySQLTestDB()

    // 清理数据确保测试独立性
    defer mysqlDB.Clear()

    // 测试逻辑...
}
```

### 3. 配置驱动测试

```go
func TestMultipleScenarios(t *testing.T) {
    scenarios := []string{
        "basic_test.yaml",
        "with_blobs.yaml",
        "large_dataset.yaml",
    }

    for _, scenario := range scenarios {
        t.Run(scenario, func(t *testing.T) {
            adapter := GetGlobalTestDBAdapter()
            err := adapter.SetupMySQLTestDB(scenario)
            require.NoError(t, err)
            defer adapter.Cleanup()

            // 场景特定测试...
        })
    }
}
```

## 故障排除

### 常见问题

1. **端口冲突**: 确保端口 3307 (MySQL) 和 6380 (Redis) 没有被占用
2. **依赖缺失**: 运行 `go mod tidy` 确保所有依赖已安装
3. **配置文件错误**: 检查 YAML 配置文件语法和路径

### 调试技巧

```go
// 启用详细日志
func TestWithDebugging(t *testing.T) {
    if err := SetupTestEnvironment(); err != nil {
        t.Logf("调试信息: %v", err)
    }
    defer TeardownTestEnvironment()

    // 检查连接状态
    adapter := GetGlobalTestDBAdapter()
    mysqlDB := adapter.GetMySQLTestDB()
    if mysqlDB != nil {
        conn := mysqlDB.GetMySQLConnection()
        if conn != nil {
            t.Logf("数据库连接状态: %v", conn.DB.Ping() == nil)
        }
    }
}
```

## 性能考虑

- **启动时间**: 模拟服务器启动通常需要 100-200ms
- **内存使用**: 内存数据库完全在内存中运行，注意大数据集测试
- **并发测试**: 支持并发测试，但每个测试应该使用独立的数据

## 扩展性

框架设计为易于扩展：

1. **新数据库类型**: 实现 `TestDatabase` 接口
2. **新配置格式**: 扩展配置加载器
3. **新数据类型**: 扩展类型映射系统

## 许可证

本框架是 LVAN Dumper 项目的一部分，遵循项目许可证。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个测试框架。