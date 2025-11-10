# LVAN Dumper v2架构

v2架构是LVAN Dumper项目的现代化重构版本，采用TDD（测试驱动开发）方法开发，提供类型安全、高性能、易扩展的数据导出解决方案。

## 🎯 核心特性

### 类型安全设计
- 强类型配置接口替代`interface{}`
- 编译时错误检查
- 类型推断和转换机制

### 访问者模式优化
- 分层接口设计：`Visitor → DataVisitor`
- 自动资源管理
- 扩展性支持

### 统一数据结构
- 多类型字段支持
- 元数据管理
- 序列化/反序列化

### 工厂模式
- 配置驱动的数据源创建
- 智能缓存机制
- 连接池管理

## 📁 架构结构

```
pkg/datasource/v2/
├── README.md                   # 本文档
├── interface.go                # 核心接口定义
├── config.go                   # 配置接口和实现
├── config_test.go              # 配置测试
├── record.go                   # 统一数据记录结构
├── record_test.go              # Record测试
├── mysql_datasource.go         # MySQL数据源实现
├── mysql_export_visitor.go     # MySQL导出访问者
├── mysql_export_visitor_test.go # 导出功能测试
├── factory.go                  # 数据源工厂
├── factory_test.go             # 工厂测试
└── datasource_test.go          # 数据源基础测试
```

## 🚀 快速开始

### 1. 创建数据源工厂

```go
import "github.com/wangtengda0310/gobee/lvan/pkg/datasource/v2"

// 创建工厂
factory := v2.NewDatasourceFactory()
```

### 2. 配置MySQL数据源

```go
// 创建MySQL配置
config := v2.NewMySQLConfig(
    "localhost",    // 主机
    3306,           // 端口
    "username",     // 用户名
    "password",     // 密码
    "database",     // 数据库
    "table",        // 表名
)

// 验证配置
if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

### 3. 创建数据源

```go
// 通过工厂创建数据源
datasource, err := factory.CreateMySQL(config)
if err != nil {
    log.Fatal(err)
}
```

### 4. 配置导出选项

```go
options := &v2.MySQLOptions{
    Where:       "age > 18",
    IDs:         []string{"1", "2", "3"},
    Fields:      []string{"id", "name", "email"},
    Limit:       1000,
    OutputFormat: "zip",
}
```

### 5. 执行导出

```go
// 创建导出访问者
visitor := v2.NewMySQLExportVisitor(options)

// 执行导出
datasource.Accept(visitor)

// 获取结果
results := visitor.GetResults()
stats := visitor.GetStats()
```

## 🔧 核心组件

### 配置接口

```go
// 基础配置接口
type Config interface {
    GetType() string
    Validate() error
    Clone() Config
    ToMap() map[string]interface{}
}

// MySQL配置接口
type MySQLConfig interface {
    Config
    GetHost() string
    GetPort() int
    GetUser() string
    GetPassword() string
    GetDatabase() string
    GetTable() string
    GetDSN() string
}
```

### 数据记录

```go
// 统一数据记录接口
type Record interface {
    // 字段操作
    SetField(name string, value interface{})
    GetField(name string) (interface{}, error)
    GetString(name string) (string, error)
    GetInt(name string) (int, error)
    GetBool(name string) (bool, error)

    // 元数据操作
    SetMetadata(key, value string)
    GetMetadata(key string) string

    // 序列化操作
    ToMap() map[string]interface{}
    Clone() Record
}
```

### 访问者模式

```go
// 数据访问者接口
type DataVisitor interface {
    Visitor
    VisitMySQL(ds MySQLDatasource)
    VisitRedis(ds RedisDatasource)
}

// MySQL导出访问者
type MySQLExportVisitor interface {
    DataVisitor
    GetOptions() *MySQLOptions
    GetResults() []Record
    GetStats() *ExportStats
}
```

### 工厂模式

```go
// 数据源工厂接口
type DatasourceFactory interface {
    CreateMySQL(config MySQLConfig) (MySQLDatasource, error)
    CreateRedis(config RedisConfig) (RedisDatasource, error)
    Create(config Config) (Datasource, error)
    GetStats() *FactoryStats
    ClearCache() error
}
```

## 📊 支持的输出格式

- **ZIP**: 压缩包格式，支持多文件导出
- **DIR**: 目录格式，每个表一个文件
- **SQL-TPL**: SQL模板格式，支持自定义模板
- **CSV**: CSV文件格式
- **JSON**: JSON文件格式

## 🧪 测试覆盖

v2架构采用TDD开发，拥有完整的测试覆盖：

```bash
# 运行所有测试
go test ./pkg/datasource/v2 -v

# 运行特定组件测试
go test ./pkg/datasource/v2 -run TestConfig
go test ./pkg/datasource/v2 -run TestRecord
go test ./pkg/datasource/v2 -run TestFactory
go test ./pkg/datasource/v2 -run TestMySQLExportVisitor
```

### 测试统计

- **总测试数量**: 85+ 个测试用例
- **覆盖范围**: 接口、配置、数据结构、导出功能、工厂模式
- **测试类型**: 单元测试、集成测试、边界测试、性能测试
- **执行时间**: < 1秒

## 🔍 示例代码

查看 `examples/v2_integration_example.go` 了解完整的使用示例。

```bash
# 运行集成示例
go run examples/v2_integration_example.go
```

## 📈 性能特性

### 缓存机制
- 智能配置缓存
- 数据源实例复用
- 内存优化管理

### 连接池
- 自动连接管理
- 连接健康检查
- 超时处理

### 批量处理
- 流式数据处理
- 内存控制
- 并发支持

## 🛡️ 安全特性

### 敏感信息保护
- 密码自动脱敏
- 安全连接字符串生成
- 审计日志支持

### 输入验证
- 配置参数验证
- SQL注入防护
- 类型安全检查

## 🔄 v1 vs v2 对比

| 特性 | v1 | v2 |
|------|----|----|
| 类型安全 | ❌ interface{} | ✅ 强类型接口 |
| 测试覆盖 | 部分 | ✅ 完整TDD |
| 配置管理 | 分散 | ✅ 统一配置系统 |
| 错误处理 | 不一致 | ✅ 标准化错误处理 |
| 资源管理 | 手动 | ✅ 自动资源管理 |
| 扩展性 | 有限 | ✅ 插件化架构 |
| 缓存机制 | 无 | ✅ 智能缓存 |
| 监控统计 | 基础 | ✅ 详细统计信息 |

## 🚧 迁移指南

详细的迁移指南请参考：`docs/code_migration_guide.md`

### 基本迁移步骤

1. 安装v2依赖
2. 更新配置创建方式
3. 使用工厂模式创建数据源
4. 采用访问者模式处理数据
5. 运行兼容性测试

## 📚 相关文档

- [架构决策记录](../../docs/architecture_decisions.md)
- [TDD实施指南](../../docs/tdd_implementation_guide.md)
- [接口设计规范](../../docs/interface_design_specification.md)
- [开发环境配置](../../docs/development_setup.md)
- [代码迁移指南](../../docs/code_migration_guide.md)

## 🤝 贡献指南

v2架构遵循严格的TDD开发流程：

1. 先写测试（红阶段）
2. 实现最小功能（绿阶段）
3. 重构优化（重构阶段）

所有新功能必须包含完整的测试用例。

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 LICENSE 文件。