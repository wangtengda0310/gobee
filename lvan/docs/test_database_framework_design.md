# 测试数据库框架设计文档

## 概述
本文档记录了 LVAN Dumper 项目中测试数据库框架的详细设计，该框架基于 dolthub/go-mysql-server 和 mini redis 提供内存数据库服务，支持 YAML 配置驱动，为后续提升为独立项目做准备。

## 设计背景

### 需求背景
1. **统一测试环境**：为 v1 和 v2 两个版本的架构提供一致的测试环境
2. **真实数据库体验**：使用真实的 MySQL 和 Redis 协议，而非简单的 Mock
3. **配置驱动**：通过 YAML 配置文件定义数据库结构和测试数据
4. **独立化设计**：框架设计为可独立的项目，便于复用和维护

### 技术选型决策
- **MySQL**: 使用 `github.com/dolthub/go-mysql-server` 提供内存 MySQL 服务
- **Redis**: 使用 `github.com/alicebob/miniredis/v2` 提供内存 Redis 服务
- **配置格式**: YAML 格式，支持复杂的结构和数据定义
- **错误处理**: 所有错误直接 panic，保持代码简洁

## 架构设计

### 目录结构
```
pkg/
├── testdb/                     # 独立的测试数据库框架
│   ├── README.md               # 使用说明
│   ├── interface.go            # 接口定义
│   ├── config.go               # 配置结构
│   ├── factory.go              # 工厂函数
│   ├── mysql/
│   │   ├── server.go           # MySQL服务器创建
│   │   ├── config_loader.go    # YAML配置加载
│   │   └── schema_builder.go   # 表结构构建
│   ├── redis/
│   │   ├── server.go           # Redis服务器创建
│   │   ├── config_loader.go    # YAML配置加载
│   │   └── data_builder.go     # 数据构建
│   └── configs/                # 默认配置文件
│       ├── mysql_test.yaml     # MySQL测试配置
│       └── redis_test.yaml     # Redis测试配置
└── datasource/
    └── tests/
        └── testdb_adapter.go   # 测试框架适配器
```

### 核心接口设计

#### 测试数据库基础接口
```go
type TestDatabase interface {
    Start() error
    Stop() error
    GetConnection() interface{}
    Clear() error
    LoadConfig(configPath string) error
}
```

#### MySQL 测试数据库接口
```go
type MySQLTestDatabase interface {
    TestDatabase
    GetMySQLConnection() *sql.DB
    CreateTable(schema *TableSchema) error
    InsertData(tableName string, data []map[string]interface{}) error
    Query(query string, args ...interface{}) ([]map[string][]byte, error)
}
```

#### Redis 测试数据库接口
```go
type RedisTestDatabase interface {
    TestDatabase
    GetRedisClient() *redis.Client
    SetData(key string, value interface{}, ttl time.Duration) error
    SetHashData(key string, fields map[string]interface{}) error
    GetData(key string) (interface{}, error)
    GetKeys(pattern string) ([]string, error)
}
```

## YAML 配置设计

### MySQL 配置结构

#### 配置示例
```yaml
mysql:
  server:
    port: 3307
    database: testdb
    user: root
    password: ""

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
          data: "SGVsbG8gV29ybGQhIQ=="  # Base64编码
        - uid: 2
          accountid: user002
          data: "0x48656c6c6f576f726c642121"  # 十六进制

  variables:
    test_user_count: 100
    default_balance: 1000.00
```

#### 配置特点
- **完整的表结构定义**：支持列类型、约束、索引等
- **多种数据类型支持**：INT、VARCHAR、BLOB、JSON、TIMESTAMP 等
- **灵活的索引定义**：支持唯一索引和复合索引
- **默认值和约束**：支持列默认值和非空约束
- **测试数据预定义**：可以在配置中直接定义测试数据
- **变量系统**：支持自定义变量，便于配置复用

### Redis 配置结构

#### 配置示例
```yaml
redis:
  server:
    port: 6380
    db: 0

  data:
    # 字符串类型数据
    strings:
      - key: "user:1"
        value: '{"id":1,"name":"alice"}'
        ttl: 3600

    # Hash类型数据
    hashes:
      - key: "user:1:profile"
        fields:
          id: "1"
          name: "alice"
          email: "alice@example.com"
        ttl: 3600

    # List类型数据
    lists:
      - key: "user:1:orders"
        values:
          - '{"order_id":1001,"amount":50.00}'
          - '{"order_id":1002,"amount":75.25}'
        ttl: 0

    # Set类型数据
    sets:
      - key: "user:1:tags"
        members:
          - "premium"
          - "verified"
        ttl: 0

    # Sorted Set类型数据
    sorted_sets:
      - key: "leaderboard:score"
        members:
          - member: "user:1"
            score: 1500.5
          - member: "user:2"
            score: 2200.75
        ttl: 0

  variables:
    default_ttl: 3600
    max_users: 1000
```

#### 配置特点
- **支持所有 Redis 数据类型**：String、Hash、List、Set、Sorted Set
- **TTL 支持**：每个数据项都可以设置过期时间
- **灵活的值格式**：支持字符串、JSON 等多种格式
- **批量数据定义**：可以一次性定义多个键值对
- **命名空间友好**：支持复杂的键命名模式

## 核心组件实现

### 工厂模式设计

```go
type TestDatabaseFactory interface {
    CreateMySQLFromConfig(configPath string) (MySQLTestDatabase, error)
    CreateRedisFromConfig(configPath string) (RedisTestDatabase, error)
    CreateMySQL(config MySQLConfig) (MySQLTestDatabase, error)
    CreateRedis(config RedisConfig) (RedisTestDatabase, error)
}
```

#### 设计优势
- **统一创建接口**：通过工厂统一管理数据库实例创建
- **配置文件支持**：支持从 YAML 配置文件直接创建实例
- **程序化创建**：也支持通过代码直接创建实例
- **全局实例**：提供全局工厂实例，简化使用

### MySQL 服务器实现

#### 核心特性
- **内存数据库**：基于 dolthub/go-mysql-server 的内存实现
- **完整的 MySQL 协议**：支持标准 MySQL 连接和 SQL 语法
- **动态表创建**：根据 YAML 配置动态创建表结构
- **数据初始化**：自动插入配置中定义的测试数据
- **连接管理**：自动管理连接生命周期

#### 实现要点
```go
func (m *MySQLTestDatabaseImpl) Start() error {
    // 1. 验证配置
    if err := ValidateMySQLConfig(m.config); err != nil {
        return err
    }

    // 2. 创建内存数据库
    m.db = memory.NewDatabase(m.config.Server.Database)

    // 3. 启动 MySQL 服务器
    m.server = createMySQLServer(m.db, m.config.Server)

    // 4. 建立连接
    m.conn = createConnection(m.config.Server)

    // 5. 创建表结构
    if err := m.createSchemas(); err != nil {
        return err
    }

    // 6. 插入测试数据
    if err := m.insertTestData(); err != nil {
        return err
    }

    return nil
}
```

### Redis 服务器实现

#### 核心特性
- **内存 Redis 服务**：基于 miniredis 的完整 Redis 实现
- **支持所有数据类型**：String、Hash、List、Set、Sorted Set
- **TTL 支持**：完整的过期时间管理
- **数据持久化**：支持从配置批量导入数据
- **协议兼容**：完全兼容 Redis 客户端库

#### 数据类型支持
```go
func (r *RedisTestDatabaseImpl) loadData() error {
    // 加载字符串数据
    for _, str := range r.config.Data.Strings {
        r.client.Set(ctx, str.Key, str.Value, str.TTL)
    }

    // 加载 Hash 数据
    for _, hash := range r.config.Data.Hashes {
        r.client.HSet(ctx, hash.Key, hash.Fields)
        if hash.TTL > 0 {
            r.client.Expire(ctx, hash.Key, hash.TTL)
        }
    }

    // ... 其他数据类型
}
```

## 配置加载机制

### 配置文件查找策略
```go
func LoadMySQLConfig(configPath string) (MySQLConfig, error) {
    // 1. 如果是绝对路径，直接使用
    if filepath.IsAbs(configPath) {
        return loadFromFile(configPath)
    }

    // 2. 尝试相对路径查找
    searchPaths := []string{
        configPath,
        filepath.Join("configs", configPath),
        filepath.Join("pkg/testdb", configPath),
        filepath.Join("pkg/testdb", "configs", configPath),
    }

    for _, path := range searchPaths {
        if fileExists(path) {
            return loadFromFile(path)
        }
    }

    return MySQLConfig{}, fmt.Errorf("配置文件未找到: %s", configPath)
}
```

### 配置验证机制
```go
func ValidateMySQLConfig(config MySQLConfig) error {
    // 端口验证
    if config.Server.Port <= 0 || config.Server.Port > 65535 {
        return fmt.Errorf("无效的端口号: %d", config.Server.Port)
    }

    // 表结构验证
    for _, schema := range config.Schemas {
        // 检查主键
        hasPrimary := false
        for _, col := range schema.Columns {
            if col.Primary {
                hasPrimary = true
                break
            }
        }

        if !hasPrimary {
            return fmt.Errorf("表 %s 需要一个主键", schema.Name)
        }
    }

    return nil
}
```

## 使用方式

### 基本使用
```go
// 从配置文件创建
mysqlDB, err := testdb.CreateMySQLFromConfig("mysql_test.yaml")
if err != nil {
    panic(err)
}
defer mysqlDB.Stop()

// 获取连接进行操作
conn := mysqlDB.GetMySQLConnection()
rows, err := conn.Query("SELECT * FROM user WHERE uid = ?", 1)
```

### 在测试中使用
```go
func TestUserOperations(t *testing.T) {
    // 设置测试数据库
    mysqlDB, err := testdb.CreateMySQLFromConfig("mysql_test.yaml")
    if err != nil {
        t.Fatalf("创建测试数据库失败: %v", err)
    }
    defer mysqlDB.Stop()

    // 获取连接
    conn := mysqlDB.GetMySQLConnection()

    // 执行测试逻辑
    result := testUserCreation(t, conn)

    // 清理数据（可选）
    if err := mysqlDB.Clear(); err != nil {
        t.Errorf("清理测试数据失败: %v", err)
    }
}
```

### 批量测试场景
```go
func TestMultipleScenarios(t *testing.T) {
    scenarios := []string{
        "mysql_basic_test.yaml",
        "mysql_with_blobs.yaml",
        "mysql_large_dataset.yaml",
    }

    for _, scenario := range scenarios {
        t.Run(scenario, func(t *testing.T) {
            mysqlDB, err := testdb.CreateMySQLFromConfig(scenario)
            if err != nil {
                t.Fatalf("创建测试数据库失败: %v", err)
            }
            defer mysqlDB.Stop()

            // 执行特定场景的测试
            runScenarioTest(t, mysqlDB)
        })
    }
}
```

## 项目独立化规划

### 当前设计优势
1. **清晰的接口抽象**：便于扩展其他数据库类型
2. **配置驱动架构**：易于维护和定制
3. **模块化设计**：各个组件职责明确
4. **完整的生命周期管理**：从创建到销毁的完整支持

### 独立化准备
```
testdb/ (独立项目)
├── cmd/
│   └── testdb-cli/           # 命令行工具
├── pkg/
│   ├── mysql/               # MySQL 支持
│   ├── redis/               # Redis 支持
│   ├── postgresql/          # PostgreSQL 支持（未来）
│   ├── mongodb/             # MongoDB 支持（未来）
│   └── core/                # 核心接口和工厂
├── configs/                 # 配置模板
├── examples/                # 使用示例
└── docs/                    # 文档
```

### 功能扩展规划
1. **更多数据库支持**：PostgreSQL、MongoDB、SQLite 等
2. **Docker 集成**：支持 Docker 容器中的数据库
3. **云服务支持**：支持云数据库服务
4. **数据生成器**：自动生成测试数据的工具
5. **CLI 工具**：命令行管理工具
6. **Web 界面**：可视化的数据库管理界面

## 性能考虑

### 内存使用优化
- **按需启动**：数据库实例在使用时才启动
- **及时清理**：测试结束后自动清理资源
- **连接池管理**：复用数据库连接

### 启动时间优化
- **端口自动分配**：避免端口冲突
- **异步启动**：服务器启动不阻塞主流程
- **健康检查**：确保服务可用后再使用

### 数据操作优化
- **批量操作**：支持批量插入数据
- **事务支持**：支持事务操作确保数据一致性
- **索引优化**：根据配置自动创建索引

## 最佳实践

### 配置文件组织
```
configs/
├── mysql/
│   ├── basic.yaml          # 基础表结构
│   ├── with_blobs.yaml     # 包含 Blob 数据
│   ├── large_dataset.yaml  # 大数据集
│   └── performance.yaml    # 性能测试专用
└── redis/
    ├── basic.yaml
    ├── complex_types.yaml
    └── ttl_tests.yaml
```

### 测试数据管理
- **使用固定种子**：确保测试数据可重现
- **数据大小合理**：避免测试数据过大影响性能
- **多样化场景**：覆盖不同的使用场景
- **定期清理**：避免测试数据累积

### 错误处理
- **快速失败**：配置错误立即暴露
- **详细日志**：提供足够的错误信息
- **资源清理**：确保异常情况下资源正确释放

## 总结

该测试数据库框架提供了：

1. **统一的测试环境**：为不同架构版本提供一致的测试基础
2. **真实的数据库体验**：使用真实协议而非简单 Mock
3. **灵活的配置驱动**：通过 YAML 文件轻松定义测试场景
4. **良好的扩展性**：为后续独立化和功能扩展做好准备
5. **简洁的使用方式**：最小化学习成本，提高开发效率

这个框架不仅满足当前 LVAN Dumper 项目的测试需求，还为未来独立化发展奠定了坚实基础。

---

*文档版本：v1.0*
*最后更新：2025-01-07*
*维护者：LVAN Dumper 开发团队*