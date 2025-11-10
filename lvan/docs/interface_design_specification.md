# 接口设计规范

## 概述
本文档定义了LVAN Dumper项目中接口设计的标准规范，确保接口的一致性、可扩展性和可维护性。

## v1和v2接口对比

### v1: 函数式接口设计

#### 核心接口
```go
// 数据源接口
type Datasource interface {
    Accept(fn func(DatasourceAccessor))
}

// 数据源访问器
type DatasourceAccessor interface {
    GetMetadata() Metadata
    Dump(where string, ids []string) [][]map[string][]byte
    Import(data [][]map[string][]byte)
    Type() string
}
```

#### 特点
- **函数式风格**: 通过函数回调访问数据
- **简单直接**: 接口方法少，学习成本低
- **类型灵活性**: 运行时类型检查
- **易于理解**: 符合Go语言习惯

#### 使用示例
```go
datasource.Accept(func(accessor DatasourceAccessor) {
    metadata := accessor.GetMetadata()
    records := accessor.Dump("uid", []string{"1", "2"})
    // 处理records
})
```

### v2: 访问者模式接口设计

#### 核心接口
```go
// 数据源接口
type Datasource interface {
    Accept(visitor Visitor)
}

// 访问者接口
type Visitor interface {
    VisitMySQL(ds *MySQLDatasource)
    VisitRedis(ds *RedisDatasource)
}

// 数据源基类
type BaseDatasource struct {
    metadata Metadata
}

func (b *BaseDatasource) GetMetadata() Metadata {
    return b.metadata
}
```

#### 特点
- **类型安全**: 编译时类型检查
- **扩展性强**: 添加新数据源类型容易
- **双重分派**: 动态行为绑定
- **模式明确**: 符合标准设计模式

#### 使用示例
```go
visitor := &DumpVisitor{where: "uid", ids: []string{"1", "2"}}
datasource.Accept(visitor)
```

## 接口选择指南

### 何时选择v1函数式接口
- **简单场景**: 数据源类型固定
- **快速开发**: 学习成本和开发速度优先
- **类型灵活性**: 需要运行时类型检查
- **团队经验**: 团队对函数式编程更熟悉

### 何时选择v2访问者模式
- **复杂场景**: 多种数据源类型
- **类型安全**: 编译时检查很重要
- **长期维护**: 代码需要长期演进
- **扩展需求**: 预期会添加新的数据源类型

## 访问者模式实现规范

### 接口定义标准

#### 数据源接口
```go
// 基础数据源接口
type Datasource interface {
    Accept(visitor Visitor)
    GetMetadata() Metadata
    Close() error
}

// 支持元数据的数据源
type DatasourceWithMetadata interface {
    Datasource
    GetColumns(table string) []ColumnInfo
    GetPrimaryKeys(table string) []string
}
```

#### 访问者接口
```go
// 基础访问者接口
type Visitor interface {
    VisitDatasource(ds Datasource)
}

// 扩展访问者接口
type ExtendedVisitor interface {
    Visitor
    VisitMySQL(ds *MySQLDatasource)
    VisitRedis(ds *RedisDatasource)
    VisitPostgreSQL(ds *PostgreSQLDatasource)
}

// 功能专用访问者
type DumpVisitor interface {
    ExtendedVisitor
    SetDumpOptions(options DumpOptions)
    GetResults() []Record
}

type ImportVisitor interface {
    ExtendedVisitor
    SetImportData(data []Record)
    GetImportResults() ImportResult
}
```

### 实现标准

#### 数据源实现
```go
// MySQL数据源实现
type MySQLDatasource struct {
    BaseDatasource
    conn   *sql.DB
    config MySQLConfig
}

func (ds *MySQLDatasource) Accept(visitor Visitor) {
    if extendedVisitor, ok := visitor.(ExtendedVisitor); ok {
        extendedVisitor.VisitMySQL(ds)
    } else {
        visitor.VisitDatasource(ds)
    }
}

func (ds *MySQLDatasource) VisitMySQL(visitor MySQLVisitor) {
    // 特定于MySQL的访问逻辑
}
```

#### 访问者实现
```go
// Dump访问者实现
type DumpVisitor struct {
    where   string
    ids     []string
    options DumpOptions
    results []Record
}

func (v *DumpVisitor) VisitMySQL(ds *MySQLDatasource) {
    // MySQL特定的dump逻辑
    columns := ds.GetColumns(ds.config.Table)
    v.results = ds.Dump(ds.config.Database, ds.config.Table, columns, v.where, v.ids...)
}

func (v *DumpVisitor) VisitRedis(ds *RedisDatasource) {
    // Redis特定的dump逻辑
    keys := ds.GetKeys(v.where)
    v.results = ds.GetValues(keys)
}

func (v *DumpVisitor) GetResults() []Record {
    return v.results
}
```

## 错误处理标准

### 统一panic策略
```go
// 所有接口方法遇到错误时直接panic
func (ds *MySQLDatasource) Connect() error {
    conn, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(fmt.Sprintf("连接MySQL失败: %v", err))
    }
    ds.conn = conn
    return nil
}

func (v *DumpVisitor) VisitMySQL(ds *MySQLDatasource) {
    defer func() {
        if r := recover(); r != nil {
            panic(fmt.Sprintf("MySQL访问失败: %v", r))
        }
    }()

    // 访问逻辑
}
```

### 错误信息格式
```go
// 标准错误信息格式
panic(fmt.Sprintf("[组件名] 操作失败: %v", err))

// 示例
panic(fmt.Sprintf("[MySQL连接] 连接失败: %v", err))
panic(fmt.Sprintf("[文件写入] 写入失败: %v", err))
```

## 扩展性指导原则

### 添加新数据源类型

#### 1. 定义数据源结构
```go
type RedisDatasource struct {
    BaseDatasource
    client *redis.Client
    config RedisConfig
}
```

#### 2. 实现Datasource接口
```go
func (ds *RedisDatasource) Accept(visitor Visitor) {
    if extendedVisitor, ok := visitor.(ExtendedVisitor); ok {
        extendedVisitor.VisitRedis(ds)
    } else {
        visitor.VisitDatasource(ds)
    }
}
```

#### 3. 扩展Visitor接口
```go
type ExtendedVisitor interface {
    Visitor
    VisitMySQL(ds *MySQLDatasource)
    VisitRedis(ds *RedisDatasource)  // 新增方法
}
```

#### 4. 更新现有访问者
```go
func (v *DumpVisitor) VisitRedis(ds *RedisDatasource) {
    // Redis特定的dump实现
}
```

### 添加新操作类型

#### 1. 定义专用访问者
```go
type ValidateVisitor interface {
    ExtendedVisitor
    GetValidationResults() ValidationResults
}
```

#### 2. 实现访问者
```go
type MySQLValidateVisitor struct {
    rules []ValidationRule
    results ValidationResults
}

func (v *MySQLValidateVisitor) VisitMySQL(ds *MySQLDatasource) {
    // MySQL特定的验证逻辑
}
```

## 类型安全保证

### 接口断言
```go
// 安全的类型断言
func (ds *MySQLDatasource) Accept(visitor Visitor) {
    if extendedVisitor, ok := visitor.(ExtendedVisitor); ok {
        extendedVisitor.VisitMySQL(ds)
        return
    }

    // 降级处理或panic
    panic("访问者不支持MySQL数据源")
}
```

### 编译时检查
```go
// 使用接口约束确保类型安全
type DatasourceProcessor[T Datasource] interface {
    Process(ds T) error
}

// 编译时确保类型匹配
type MySQLProcessor struct{}

func (p *MySQLProcessor) Process(ds *MySQLDatasource) error {
    return p.processMySQL(ds)
}
```

## 性能考虑

### 接口调用优化
```go
// 避免不必要的类型断言
type optimizedVisitor struct {
    MySQLHandler func(*MySQLDatasource)
    RedisHandler func(*RedisDatasource)
    DefaultHandler func(Datasource)
}

func (v *optimizedVisitor) VisitMySQL(ds *MySQLDatasource) {
    if v.MySQLHandler != nil {
        v.MySQLHandler(ds)
    }
}
```

### 内存管理
```go
// 及时释放资源
func (v *DumpVisitor) VisitMySQL(ds *MySQLDatasource) {
    defer func() {
        v.results = nil // 释放内存
    }()

    // 访问逻辑
}
```

## 测试接口标准

### Mock接口
```go
// 测试用的Mock数据源
type MockDatasource struct {
    metadata Metadata
    data     []Record
}

func (m *MockDatasource) Accept(visitor Visitor) {
    visitor.VisitDatasource(m)
}

func (m *MockDatasource) GetMetadata() Metadata {
    return m.metadata
}
```

### 测试访问者
```go
// 测试用的Mock访问者
type MockVisitor struct {
    Visited bool
    Error   error
}

func (m *MockVisitor) VisitMySQL(ds *MySQLDatasource) {
    m.Visited = true
    if m.Error != nil {
        panic(m.Error)
    }
}
```

### 接口测试
```go
func TestDatasourceAccept(t *testing.T) {
    ds := &MockDatasource{}
    visitor := &MockVisitor{}

    ds.Accept(visitor)

    assert.True(t, visitor.Visited)
}
```

## 文档标准

### 接口文档
```go
// Accept 接受访问者对数据源进行操作
//
// 参数:
//   - visitor: 访问者对象，用于执行具体操作
//
// 使用示例:
//   ds := NewMySQLDatasource(config)
//   visitor := &DumpVisitor{}
//   ds.Accept(visitor)
//
// 注意:
//   - 访问者必须实现相应的Visit方法
//   - 错误会以panic形式抛出
func (ds *MySQLDatasource) Accept(visitor Visitor) {
    // 实现逻辑
}
```

### 示例代码
```go
// Example_Datasource_Accent 展示如何使用Accept方法
func Example_Datasource_Accept() {
    ds := NewMySQLDatasource(MySQLConfig{
        Host: "localhost",
        Port: 3306,
        Database: "test",
    })

    visitor := &DumpVisitor{
        where: "uid",
        ids:   []string{"1", "2"},
    }

    ds.Accept(visitor)
    fmt.Printf("导出了%d条记录", len(visitor.GetResults()))

    // Output: 导出了2条记录
}
```

## 总结

接口设计是LVAN Dumper项目架构的核心。通过遵循本规范，我们可以确保：

1. **一致性**: 所有接口遵循统一的设计原则
2. **可维护性**: 清晰的接口定义便于维护和扩展
3. **类型安全**: 充分利用Go的类型系统
4. **性能**: 高效的接口调用和内存管理
5. **测试友好**: 易于Mock和测试的接口设计

选择合适的接口设计模式(v1或v2)应该基于具体的业务需求、团队经验和长期维护考虑。在项目初期，可以从v1开始，根据复杂度逐步迁移到v2。

---

*文档版本: v1.0*
*最后更新: 2025-01-07*