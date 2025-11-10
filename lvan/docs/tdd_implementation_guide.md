# TDD实施指南

## 概述
本文档定义了LVAN Dumper项目中TDD(测试驱动开发)的实施规范，确保团队协作的一致性和代码质量。

## TDD工作流程

### 红-绿-重构循环

#### 1. 红色阶段 (Red)
- **目标**: 编写一个失败的测试用例
- **原则**: 测试必须失败，且失败原因明确
- **步骤**:
  1. 识别需要实现的功能点
  2. 设计对应的测试用例
  3. 编写测试代码（不包含实现逻辑）
  4. 运行测试确认失败

#### 2. 绿色阶段 (Green)
- **目标**: 编写最少代码使测试通过
- **原则**: 快速通过，不追求完美
- **步骤**:
  1. 编写最小化的实现代码
  2. 解决编译错误
  3. 运行测试确认通过
  4. 验证测试覆盖率

#### 3. 重构阶段 (Refactor)
- **目标**: 优化代码结构，提高质量
- **原则**: 保持测试通过，小步重构
- **步骤**:
  1. 识别代码异味
  2. 小步重构，保持测试通过
  3. 优化设计，消除重复
  4. 验证所有测试仍然通过

### 循环示例
```go
// 红色阶段：编写失败测试
func TestVisitorPattern(t *testing.T) {
    datasource := NewMySQLDatasource(config)
    visitor := &MockVisitor{}

    datasource.Accept(visitor)

    if !visitor.Called {
        t.Error("Visitor not called")
    }
}

// 绿色阶段：最小实现
func (ds *MySQLDatasource) Accept(visitor Visitor) {
    visitor.VisitMySQL(ds) // 简单调用即可通过测试
}

// 重构阶段：优化设计
func (ds *MySQLDatasource) Accept(visitor Visitor) {
    if mysqlVisitor, ok := visitor.(MySQLVisitor); ok {
        mysqlVisitor.VisitMySQL(ds)
    } else {
        visitor.VisitGeneric(ds)
    }
}
```

## 测试用例设计原则

### FIRST原则

#### F - Fast (快速)
- 单个测试运行时间 < 100ms
- 避免网络调用和文件I/O
- 使用内存数据库替代真实数据库

#### I - Isolated (隔离)
- 测试之间相互独立
- 不依赖测试执行顺序
- 每个测试有独立的设置和清理

#### R - Repeatable (可重复)
- 测试结果稳定一致
- 不依赖外部环境
- 使用固定种子生成随机数据

#### S - Self-Validating (自验证)
- 测试自动通过/失败
- 无需人工检查结果
- 明确的断言条件

#### T - Timely (及时)
- 测试在实现之前编写
- 测试驱动代码设计
- 避免过度测试

### 测试分类

#### 单元测试
```go
// 测试单个函数或方法
func TestDatasourceAccept(t *testing.T) {
    ds := NewMySQLDatasource()
    visitor := &TestVisitor{}

    ds.Accept(visitor)

    assert.True(t, visitor.Visited)
}
```

#### 集成测试
```go
// 测试多个组件协作
func TestMySQLDumpIntegration(t *testing.T) {
    factory := testdb.GetTestDatabaseFactory()
    mysqlDB := factory.CreateMySQLFromConfig("test_config.yaml")
    defer mysqlDB.Stop()

    // 测试完整的dump流程
}
```

#### 一致性测试
```go
// 测试v1和v2版本的一致性
func TestV1V2Consistency(t *testing.T) {
    testData := prepareTestData()

    v1Result := testV1Implementation(testData)
    v2Result := testV2Implementation(testData)

    assert.Equal(t, v1Result, v2Result)
}
```

## 重构策略和时机

### 重构触发条件

#### 1. 代码异味检测
- **重复代码**: 相同逻辑出现多次
- **长函数**: 超过20行的函数
- **大类**: 超过200行的类
- **长参数列表**: 超过4个参数
- **特性嫉妒**: 类过度使用其他类的功能

#### 2. 测试覆盖率下降
- 新增功能未覆盖测试
- 测试用例过于复杂
- 测试执行时间增长

#### 3. 设计原则违反
- 单一职责原则违反
- 开闭原则违反
- 依赖倒置原则违反

### 重构技术

#### 提取方法
```go
// 重构前
func (ds *MySQLDatasource) Dump() {
    // 20行复杂逻辑
    conn := ds.getConnection()
    query := "SELECT * FROM table"
    // ... 更多逻辑
}

// 重构后
func (ds *MySQLDatasource) Dump() {
    conn := ds.getConnection()
    query := ds.buildQuery()
    return ds.executeQuery(conn, query)
}

func (ds *MySQLDatasource) buildQuery() string {
    return "SELECT * FROM table"
}
```

#### 提取接口
```go
// 重构前
func processMySQL(ds *MySQLDatasource) {
    // 直接依赖具体类
}

// 重构后
type Datasource interface {
    GetConnection() *sql.DB
    BuildQuery() string
}

func processDatasource(ds Datasource) {
    // 依赖接口，更灵活
}
```

#### 引入参数对象
```go
// 重构前
func connect(host string, port int, user string, password string, database string) error {
    // 参数过多
}

// 重构后
type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
}

func connect(config DatabaseConfig) error {
    // 参数清晰
}
```

## 代码质量标准

### 命名规范

#### 测试命名
```go
// 格式: Test[功能名][测试场景]
func TestVisitorPatternAcceptCalled(t *testing.T) {
    // 清晰表达测试意图
}

func TestMySQLDatasourceConnectWithInvalidCredentials(t *testing.T) {
    // 明确测试条件和预期
}
```

#### 变量和函数命名
```go
// 清晰的命名
type MySQLDatasource struct{}
type RedisDatasource struct{}

func (ds *MySQLDatasource) Accept(visitor Visitor) {}
func (ds *RedisDatasource) Accept(visitor Visitor) {}

// 避免模糊命名
type Data struct{} // 太模糊
type Record struct{} // 更清晰
```

### 注释标准

#### 公共API注释
```go
// Accept 接受访问者对数据源进行操作
// 实现访问者模式，将操作逻辑从数据源中分离
func (ds *MySQLDatasource) Accept(visitor Visitor) {
    // 实现逻辑
}
```

#### 复杂逻辑注释
```go
func (ds *MySQLDatasource) handleTransaction(tx *sql.Tx) error {
    // 开启事务确保数据一致性
    // 所有操作必须在同一个事务中完成
    // 如果任何步骤失败，整个事务回滚
}
```

### 错误处理标准

#### 统一panic策略
```go
func (ds *MySQLDatasource) Connect() error {
    conn, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(fmt.Sprintf("连接数据库失败: %v", err))
    }
    return nil
}
```

#### 资源清理
```go
func (ds *MySQLDatasource) Close() error {
    if ds.conn != nil {
        if err := ds.conn.Close(); err != nil {
            panic(fmt.Sprintf("关闭数据库连接失败: %v", err))
        }
    }
    return nil
}
```

## 测试工具和最佳实践

### 测试数据生成

#### 固定种子随机数据
```go
func TestWithRandomData(t *testing.T) {
    generator := NewTestDataGenerator(42) // 固定种子
    testData := generator.GenerateRecords(10)

    // 测试逻辑
}
```

#### 测试构建器模式
```go
type DatasourceBuilder struct {
    config MySQLConfig
}

func NewDatasourceBuilder() *DatasourceBuilder {
    return &DatasourceBuilder{
        config: MySQLConfig{
            Host: "localhost",
            Port: 3306,
            User: "root",
            Database: "test",
        },
    }
}

func (b *DatasourceBuilder) WithHost(host string) *DatasourceBuilder {
    b.config.Host = host
    return b
}

func (b *DatasourceBuilder) Build() *MySQLDatasource {
    return NewMySQLDatasource(b.config)
}

// 使用
ds := NewDatasourceBuilder().WithHost("testhost").Build()
```

### Mock和Stub

#### 接口Mock
```go
type MockVisitor struct {
    Called bool
    Error  error
}

func (m *MockVisitor) VisitMySQL(ds *MySQLDatasource) {
    m.Called = true
    if m.Error != nil {
        panic(m.Error)
    }
}
```

#### 测试替身
```go
type TestDatasource struct {
    data []Record
}

func (t *TestDatasource) Accept(visitor Visitor) {
    if mysqlVisitor, ok := visitor.(MySQLVisitor); ok {
        mysqlVisitor.VisitMySQL(t)
    }
}
```

## 持续集成要求

### 测试执行
```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行性能基准测试
go test -bench=. ./...
```

### 代码质量检查
```bash
# 代码格式化
go fmt ./...

# 静态分析
go vet ./...

# 依赖安全检查
go list -json -m all | nancy sleuth
```

## 常见陷阱和解决方案

### 陷阱1：测试过于复杂
**问题**: 测试代码比业务代码还复杂
**解决**: 遵循FIRST原则，简化测试逻辑

### 陷阱2：过度Mock
**问题**: Mock过多，测试失去意义
**解决**: 优先使用真实实现，必要时使用测试数据库

### 陷阱3：测试脆弱
**问题**: 测试容易因无关变更失败
**解决**: 关注行为而非实现，使用稳定的接口

### 陷阱4：重构恐惧
**问题**: 害怕重构破坏测试
**解决**: 高测试覆盖率，小步重构

## 总结

TDD是一种开发方法论，需要团队全员参与和持续实践。通过遵循本指南，我们可以：

1. **提高代码质量**：测试驱动的代码更加健壮
2. **降低维护成本**：充分的测试减少回归风险
3. **促进设计思考**：测试先行推动更好的设计
4. **增强开发信心**：快速重构而不破坏功能

记住TDD的核心价值：**通过测试驱动设计，通过重构保持质量**。

---

*文档版本: v1.0*
*最后更新: 2025-01-07*