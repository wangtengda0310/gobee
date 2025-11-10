# 代码迁移指南

## 概述
本文档描述了从现有代码到新架构的迁移策略，确保平滑过渡和向后兼容性。

## 迁移策略概述

### 迁移原则
1. **渐进式迁移**: 小步快跑，避免大规模破坏性变更
2. **向后兼容**: 确保现有功能继续工作
3. **TDD驱动**: 通过测试确保迁移安全
4. **并行支持**: 新旧架构并存，逐步切换

### 迁移阶段
1. **阶段1**: 接口定义和基础框架
2. **阶段2**: 核心功能迁移
3. **阶段3**: 高级功能实现
4. **阶段4**: 清理和优化

## 现有代码分析

### 当前架构问题
```go
// 问题1: 全局变量滥用
var TransImport func(string) []dump.Record
var TransExport func([]dump.Record, ...string) string

// 问题2: 访问者模式实现不清晰
type Acceptor func(visitor Visitor, where string, args ...string)
type Visitor func(dump.Datasource)

// 问题3: 错误处理不一致
if err != nil {
    log.Panic(err)
    return err  // 无效的return
}

// 问题4: 配置管理分散
mysqlCmd.Flags().StringP("host", "h", "localhost", "MySQL 主机名")
// 类似代码分散在多个文件中
```

### 依赖关系
```
cmd/dumper/cmd/
├── root.go      # 核心命令定义
├── mysql.go     # MySQL特定逻辑
├── dump.go      # 导出功能
├── import.go    # 导入功能
└── internal/
    └── action.go # 访问者模式实现

pkg/dump/
├── dump.go      # 核心dump逻辑
├── import.go    # 核心import逻辑
├── mysql.go     # MySQL特定功能
└── [其他文件]
```

## 阶段1: 接口定义和基础框架

### 1.1 创建新的接口定义

#### v2访问者接口
```go
// pkg/datasource/v2/interface.go
package v2

import "context"

// 数据源基础接口
type Datasource interface {
    Accept(visitor Visitor)
    GetMetadata() Metadata
    Close() error
}

// 访问者基础接口
type Visitor interface {
    VisitDatasource(ds Datasource)
}

// 扩展访问者接口
type ExtendedVisitor interface {
    Visitor
    VisitMySQL(ds *MySQLDatasource)
    VisitRedis(ds *RedisDatasource)
}

// 元数据接口
type Metadata interface {
    GetType() string
    GetTables() []string
    GetColumns(table string) []ColumnInfo
}

// MySQL数据源接口
type MySQLDatasource interface {
    Datasource
    GetConnection() *sql.DB
    GetDatabase() string
    GetTable() string
}

// Redis数据源接口
type RedisDatasource interface {
    Datasource
    GetClient() *redis.Client
    GetKeyPattern() string
}
```

#### 配置接口
```go
// pkg/datasource/config/interface.go
package config

// 配置接口
type Config interface {
    Validate() error
    GetType() string
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
}

// Redis配置接口
type RedisConfig interface {
    Config
    GetHost() string
    GetPort() int
    GetPassword() string
    GetDatabase() int
}
```

### 1.2 实现基础工厂

#### 数据源工厂
```go
// pkg/datasource/factory.go
package datasource

import (
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/v2"
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/config"
)

// 工厂接口
type Factory interface {
    CreateMySQL(cfg config.MySQLConfig) (v2.MySQLDatasource, error)
    CreateRedis(cfg config.RedisConfig) (v2.RedisDatasource, error)
}

// 默认工厂实现
type DefaultFactory struct{}

func NewFactory() Factory {
    return &DefaultFactory{}
}

func (f *DefaultFactory) CreateMySQL(cfg config.MySQLConfig) (v2.MySQLDatasource, error) {
    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    // 创建MySQL数据源
    return v2.NewMySQLDatasource(cfg), nil
}

func (f *DefaultFactory) CreateRedis(cfg config.RedisConfig) (v2.RedisDatasource, error) {
    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    // 创建Redis数据源
    return v2.NewRedisDatasource(cfg), nil
}
```

### 1.3 兼容性适配器

#### v1适配器
```go
// pkg/datasource/adapter/v1_adapter.go
package adapter

import (
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/v2"
    "github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// v1到v2适配器
type V1Adapter struct {
    v2datasource v2.Datasource
}

func NewV1Adapter(v2ds v2.Datasource) *V1Adapter {
    return &V1Adapter{v2datasource: v2ds}
}

// 适配v1风格的访问
func (a *V1Adapter) Accept(fn func(dump.Datasource)) {
    // 创建v1风格的访问者
    visitor := &V1Visitor{callback: fn}
    a.v2datasource.Accept(visitor)
}

// v1风格访问者
type V1Visitor struct {
    callback func(dump.Datasource)
}

func (v *V1Visitor) VisitDatasource(ds v2.Datasource) {
    // 转换为旧的Datasource接口
    oldDs := &LegacyDatasource{newDs: ds}
    v.callback(oldDs)
}

func (v *V1Visitor) VisitMySQL(ds v2.MySQLDatasource) {
    v.VisitDatasource(ds)
}

func (v *V1Visitor) VisitRedis(ds v2.RedisDatasource) {
    v.VisitDatasource(ds)
}

// 旧的Datasource接口适配
type LegacyDatasource struct {
    newDs v2.Datasource
}

func (l *LegacyDatasource) GetDB() *sql.DB {
    if mysqlDs, ok := l.newDs.(v2.MySQLDatasource); ok {
        return mysqlDs.GetConnection()
    }
    return nil
}

func (l *LegacyDatasource) GetDatabase() string {
    if mysqlDs, ok := l.newDs.(v2.MySQLDatasource); ok {
        return mysqlDs.GetDatabase()
    }
    return ""
}

func (l *LegacyDatasource) GetTable() string {
    if mysqlDs, ok := l.newDs.(v2.MySQLDatasource); ok {
        return mysqlDs.GetTable()
    }
    return ""
}
```

### 1.4 测试基础框架

#### 基础测试套件
```go
// pkg/datasource/tests/suite_test.go
package tests

import (
    "testing"
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/v2"
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/config"
)

// 测试套件接口
type TestSuite interface {
    Setup(t *testing.T)
    Teardown(t *testing.T)
    CreateTestDatasource() v2.Datasource
    CreateTestVisitor() v2.Visitor
}

// 基础测试套件
type BaseTestSuite struct {
    datasource v2.Datasource
    visitor    v2.Visitor
}

func (s *BaseTestSuite) Setup(t *testing.T) {
    // 创建测试数据源
    factory := NewTestFactory()
    config := &TestMySQLConfig{}

    var err error
    s.datasource, err = factory.CreateMySQL(config)
    if err != nil {
        t.Fatalf("创建测试数据源失败: %v", err)
    }

    // 创建测试访问者
    s.visitor = &TestVisitor{}
}

func (s *BaseTestSuite) Teardown(t *testing.T) {
    if s.datasource != nil {
        s.datasource.Close()
    }
}

func (s *BaseTestSuite) CreateTestDatasource() v2.Datasource {
    return s.datasource
}

func (s *BaseTestSuite) CreateTestVisitor() v2.Visitor {
    return s.visitor
}
```

## 阶段2: 核心功能迁移

### 2.1 MySQL数据源迁移

#### v2 MySQL数据源实现
```go
// pkg/datasource/v2/mysql_datasource.go
package v2

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

type MySQLDatasourceImpl struct {
    conn   *sql.DB
    config config.MySQLConfig
    metadata Metadata
}

func NewMySQLDatasource(cfg config.MySQLConfig) MySQLDatasource {
    // 创建连接
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        cfg.GetUser(), cfg.GetPassword(),
        cfg.GetHost(), cfg.GetPort(),
        cfg.GetDatabase())

    conn, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(fmt.Sprintf("连接MySQL失败: %v", err))
    }

    // 验证连接
    if err := conn.Ping(); err != nil {
        panic(fmt.Sprintf("MySQL连接验证失败: %v", err))
    }

    return &MySQLDatasourceImpl{
        conn:   conn,
        config: cfg,
        metadata: NewMySQLMetadata(conn, cfg),
    }
}

func (ds *MySQLDatasourceImpl) Accept(visitor Visitor) {
    if extendedVisitor, ok := visitor.(ExtendedVisitor); ok {
        extendedVisitor.VisitMySQL(ds)
    } else {
        visitor.VisitDatasource(ds)
    }
}

func (ds *MySQLDatasourceImpl) GetMetadata() Metadata {
    return ds.metadata
}

func (ds *MySQLDatasourceImpl) GetConnection() *sql.DB {
    return ds.conn
}

func (ds *MySQLDatasourceImpl) GetDatabase() string {
    return ds.config.GetDatabase()
}

func (ds *MySQLDatasourceImpl) GetTable() string {
    return ds.config.GetTable()
}

func (ds *MySQLDatasourceImpl) Close() error {
    if ds.conn != nil {
        return ds.conn.Close()
    }
    return nil
}
```

### 2.2 访问者实现迁移

#### Dump访问者
```go
// pkg/datasource/v2/dump_visitor.go
package v2

import (
    "fmt"
    "github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

type DumpVisitor struct {
    where   string
    ids     []string
    options DumpOptions
    results []dump.Record
}

type DumpOptions struct {
    OutputFormat string
    OutputPath   string
}

func NewDumpVisitor(where string, ids []string, options DumpOptions) *DumpVisitor {
    return &DumpVisitor{
        where:   where,
        ids:     ids,
        options: options,
    }
}

func (v *DumpVisitor) VisitMySQL(ds MySQLDatasource) {
    conn := ds.GetConnection()
    database := ds.GetDatabase()
    table := ds.GetTable()

    // 获取列信息
    columns := dump.Columns(conn, database, table)

    // 执行dump
    records := dump.Dump(conn, database, table, columns, v.where, v.ids...)
    v.results = records

    // 处理输出
    v.handleOutput()
}

func (v *DumpVisitor) VisitRedis(ds RedisDatasource) {
    // Redis特定的dump逻辑
    panic("Redis dump功能尚未实现")
}

func (v *DumpVisitor) VisitDatasource(ds Datasource) {
    panic(fmt.Sprintf("不支持的数据源类型: %T", ds))
}

func (v *DumpVisitor) handleOutput() {
    // 根据选项处理输出
    switch v.options.OutputFormat {
    case "zip":
        v.outputToZip()
    case "dir":
        v.outputToDir()
    case "sql-tpl":
        v.outputToSQLTemplate()
    default:
        v.outputToConsole()
    }
}

func (v *DumpVisitor) GetResults() []dump.Record {
    return v.results
}
```

### 2.3 命令行集成迁移

#### 重构MySQL命令
```go
// cmd/dumper/cmd/mysql_v2.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource"
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/v2"
    "github.com/wangtengda0310/gobee/lvan/pkg/datasource/config"
)

// 新的MySQL命令 (v2架构)
var mysqlV2Cmd = &cobra.Command{
    Use:   "mysql-v2",
    Short: "MySQL操作 (v2架构)",
    Long:  `使用新架构的MySQL数据源操作`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // 验证配置
        cfg := buildMySQLConfig(cmd)
        if err := cfg.Validate(); err != nil {
            return fmt.Errorf("配置验证失败: %v", err)
        }
        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("MySQL v2命令已准备")
    },
}

func init() {
    rootCmd.AddCommand(mysqlV2Cmd)
    setupMySQLV2Flags(mysqlV2Cmd)
}

func setupMySQLV2Flags(cmd *cobra.Command) {
    // 复用现有的MySQL参数
    cmd.Flags().StringP("host", "h", "localhost", "MySQL 主机名")
    cmd.Flags().Uint16P("port", "P", 3306, "MySQL 端口号")
    cmd.Flags().StringP("user", "u", "root", "MySQL 用户名")
    cmd.Flags().StringP("password", "p", "", "MySQL 密码")
    cmd.Flags().StringP("database", "d", "gforge", "数据库名")
    cmd.Flags().StringP("table", "t", "user", "表名")
}

func buildMySQLConfig(cmd *cobra.Command) config.MySQLConfig {
    host, _ := cmd.Flags().GetString("host")
    port, _ := cmd.Flags().GetUint16("port")
    user, _ := cmd.Flags().GetString("user")
    password, _ := cmd.Flags().GetString("password")
    database, _ := cmd.Flags().GetString("database")
    table, _ := cmd.Flags().GetString("table")

    return config.NewMySQLConfig(host, int(port), user, password, database, table)
}
```

## 阶段3: 高级功能实现

### 3.1 SQL模板支持

#### SQL模板访问者
```go
// pkg/datasource/v2/sql_template_visitor.go
package v2

import (
    "bytes"
    "text/template"
)

type SQLTemplateVisitor struct {
    templatePath string
    variables   map[string]interface{}
    output      string
}

func NewSQLTemplateVisitor(templatePath string, variables map[string]interface{}) *SQLTemplateVisitor {
    return &SQLTemplateVisitor{
        templatePath: templatePath,
        variables:   variables,
    }
}

func (v *SQLTemplateVisitor) VisitMySQL(ds MySQLDatasource) {
    // 读取模板文件
    templateContent, err := ioutil.ReadFile(v.templatePath)
    if err != nil {
        panic(fmt.Sprintf("读取SQL模板失败: %v", err))
    }

    // 解析模板
    tmpl, err := template.New("sql").Parse(string(templateContent))
    if err != nil {
        panic(fmt.Sprintf("解析SQL模板失败: %v", err))
    }

    // 准备模板数据
    data := v.prepareTemplateData(ds)

    // 执行模板
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        panic(fmt.Sprintf("执行SQL模板失败: %v", err))
    }

    v.output = buf.String()
}

func (v *SQLTemplateVisitor) prepareTemplateData(ds MySQLDatasource) map[string]interface{} {
    return map[string]interface{}{
        "Database": ds.GetDatabase(),
        "Table":    ds.GetTable(),
        "Metadata": ds.GetMetadata(),
        "Variables": v.variables,
        "Timestamp": time.Now(),
    }
}

func (v *SQLTemplateVisitor) GetOutput() string {
    return v.output
}
```

### 3.2 Redis支持

#### Redis数据源实现
```go
// pkg/datasource/v2/redis_datasource.go
package v2

import (
    "context"
    "github.com/go-redis/redis/v8"
)

type RedisDatasourceImpl struct {
    client *redis.Client
    config config.RedisConfig
    metadata Metadata
}

func NewRedisDatasource(cfg config.RedisConfig) RedisDatasource {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.GetHost(), cfg.GetPort()),
        Password: cfg.GetPassword(),
        DB:       cfg.GetDatabase(),
    })

    // 验证连接
    ctx := context.Background()
    if err := rdb.Ping(ctx).Err(); err != nil {
        panic(fmt.Sprintf("Redis连接验证失败: %v", err))
    }

    return &RedisDatasourceImpl{
        client: rdb,
        config: cfg,
        metadata: NewRedisMetadata(rdb, cfg),
    }
}

func (ds *RedisDatasourceImpl) Accept(visitor Visitor) {
    if extendedVisitor, ok := visitor.(ExtendedVisitor); ok {
        extendedVisitor.VisitRedis(ds)
    } else {
        visitor.VisitDatasource(ds)
    }
}

func (ds *RedisDatasourceImpl) GetClient() *redis.Client {
    return ds.client
}

func (ds *RedisDatasourceImpl) GetKeyPattern() string {
    return ds.config.GetKeyPattern()
}

// 其他方法实现...
```

## 阶段4: 清理和优化

### 4.1 移除旧代码

#### 逐步废弃
```go
// cmd/dumper/cmd/legacy/mysql.go
package legacy

import "github.com/spf13/cobra"

// 添加废弃标记
// Deprecated: 使用mysql-v2命令替代
var mysqlCmd = &cobra.Command{
    Use:        "mysql [deprecated]",
    Short:      "MySQL操作 (已废弃，请使用mysql-v2)",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return fmt.Errorf("mysql命令已废弃，请使用mysql-v2命令")
    },
}
```

### 4.2 性能优化

#### 连接池优化
```go
// pkg/datasource/v2/pool.go
package v2

import "sync"

type DatasourcePool struct {
    mu    sync.RWMutex
    pools map[string]interface{}
}

func NewDatasourcePool() *DatasourcePool {
    return &DatasourcePool{
        pools: make(map[string]interface{}),
    }
}

func (p *DatasourcePool) GetMySQL(cfg config.MySQLConfig) MySQLDatasource {
    key := fmt.Sprintf("mysql:%s:%s:%d", cfg.GetHost(), cfg.GetDatabase(), cfg.GetPort())

    p.mu.RLock()
    if ds, exists := p.pools[key]; exists {
        p.mu.RUnlock()
        return ds.(MySQLDatasource)
    }
    p.mu.RUnlock()

    p.mu.Lock()
    defer p.mu.Unlock()

    // 双重检查
    if ds, exists := p.pools[key]; exists {
        return ds.(MySQLDatasource)
    }

    ds := NewMySQLDatasource(cfg)
    p.pools[key] = ds
    return ds
}
```

### 4.3 文档更新

#### 迁移文档
```markdown
# 迁移指南

## 从v1迁移到v2

### 命令变更
```bash
# 旧命令 (已废弃)
connect mysql dump --where uid 1 2 3

# 新命令
connect mysql-v2 dump --condition-column uid 1 2 3
```

### API变更
```go
// 旧代码
datasource.Accept(func(accessor DatasourceAccessor) {
    records := accessor.Dump("uid", []string{"1", "2"})
})

// 新代码
visitor := v2.NewDumpVisitor("uid", []string{"1", "2"}, options)
datasource.Accept(visitor)
```
```

## 兼容性保证

### 功能兼容性
```go
// pkg/datasource/compatibility/compat.go
package compatibility

// 兼容性包装器
type CompatibilityLayer struct {
    v2Factory datasource.Factory
}

func NewCompatibilityLayer() *CompatibilityLayer {
    return &CompatibilityLayer{
        v2Factory: datasource.NewFactory(),
    }
}

// 提供v1风格的接口
func (c *CompatibilityLayer) CreateMySQLLegacy(host string, port int, user, password, database, table string) dump.Datasource {
    config := config.NewMySQLConfig(host, port, user, password, database, table)
    v2ds, err := c.v2Factory.CreateMySQL(config)
    if err != nil {
        panic(err)
    }

    return adapter.NewV1Adapter(v2ds)
}
```

### 配置兼容性
```yaml
# config/compatibility.yaml
# 支持旧配置格式
legacy:
  mysql:
    host: localhost
    port: 3306
    user: root
    password: ""
    database: gforge
    table: user

# 新配置格式
v2:
  datasources:
    mysql:
      type: mysql
      config:
        host: localhost
        port: 3306
        user: root
        password: ""
        database: gforge
        table: user
```

## 回滚策略

### 版本切换
```go
// cmd/dumper/cmd/version.go
package cmd

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "显示版本信息",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("LVAN Dumper %s\n", Version)
        fmt.Printf("架构版本: %s\n", getArchitectureVersion())
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}

// 环境变量控制架构版本
func getArchitectureVersion() string {
    if env := os.Getenv("LVAN_ARCH_VERSION"); env != "" {
        return env
    }
    return "v2"
}
```

### 快速回滚
```bash
# 回滚到v1架构
export LVAN_ARCH_VERSION=v1
./dumper mysql dump --where uid 1 2 3

# 使用v2架构 (默认)
./dumper mysql-v2 dump --condition-column uid 1 2 3
```

## 测试策略

### 兼容性测试
```go
// pkg/datasource/tests/compatibility_test.go
package tests

func TestV1V2Compatibility(t *testing.T) {
    // 创建相同的数据
    testData := prepareTestData()

    // v1实现
    v1Result := testV1Implementation(testData)

    // v2实现
    v2Result := testV2Implementation(testData)

    // 验证一致性
    assert.Equal(t, v1Result, v2Result)
}
```

### 性能测试
```go
func BenchmarkV1V2Performance(b *testing.B) {
    testData := prepareBenchmarkData()

    b.Run("v1", func(b *testing.B) {
        benchmarkV1Implementation(b, testData)
    })

    b.Run("v2", func(b *testing.B) {
        benchmarkV2Implementation(b, testData)
    })
}
```

## 总结

通过这个迁移指南，我们可以：

1. **平滑过渡**: 通过适配器保持向后兼容
2. **风险控制**: 分阶段迁移，每个阶段都可以独立测试和回滚
3. **质量保证**: 通过TDD确保迁移的安全性
4. **性能优化**: 在新架构基础上进行性能改进

关键成功因素：
- 充分的测试覆盖
- 渐进式迁移策略
- 清晰的文档和沟通
- 持续的监控和反馈

---

*文档版本: v1.0*
*最后更新: 2025-01-07*