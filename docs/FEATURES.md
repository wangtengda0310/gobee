# LVAN Dumper 功能文档

## 项目概述

**LVAN Dumper** 是一个外网数据克隆工具，支持从生产环境导出玩家数据并导入到本地测试环境。

### 核心价值

- 🚀 **快速数据迁移**: 无需手动处理 SQL 文件
- 📦 **格式灵活**: 支持 ZIP、目录、SQL 模板等多种格式
- 🔒 **类型安全**: 使用强类型 Go 语言，编译时错误检查
- 🧪 **可测试性**: 内置 MySQL Mock 框架，无需真实数据库

---

## 功能架构

```
┌──────────────────────────────────────────────────────────────┐
│                      LVAN Dumper                             │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐ │
│  │   MySQL     │      │   Redis     │      │  未来扩展    │ │
│  │  Datasource │      │  Datasource │      │  Datasource │ │
│  └──────┬──────┘      └──────┬──────┘      └──────┬──────┘ │
│         │                    │                    │         │
│         └────────────────────┴────────────────────┘         │
│                            │                                │
│                    ┌───────▼────────┐                       │
│                    │  Dump/Import   │                       │
│                    │    Core Layer  │                       │
│                    └───────┬────────┘                       │
│                            │                                │
│         ┌──────────────────┼──────────────────┐            │
│         │                  │                  │            │
│    ┌────▼────┐       ┌────▼────┐       ┌────▼────┐       │
│    │   ZIP   │       │   DIR   │       │   SQL   │       │
│    │  Format │       │  Format │       │  Tpl    │       │
│    └────┬────┘       └────┬────┘       └────┬────┘       │
│         │                  │                  │            │
│         └──────────────────┼──────────────────┘            │
│                            │                                │
│                    ┌───────▼────────┐                       │
│                    │  File/Console  │                       │
│                    │    Output      │                       │
│                    └────────────────┘                       │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

---

## 核心功能

### 1. 数据导出 (Dump)

#### 功能描述

从数据库导出指定条件的数据，序列化为多种格式。

#### 支持的数据源

| 数据源 | 状态 | 命令 |
|-------|------|------|
| MySQL | ✅ 完成 | `connect mysql dump` |
| Redis | ⚠️ 部分完成 | `connect redis dump` |

#### 导出格式

| 格式 | 扩展名 | 代码位置 | 状态 |
|------|--------|---------|------|
| ZIP 压缩 | .zip | `pkg/dump/type/zip.go` | ✅ |
| 目录结构 | .dir | `pkg/dump/type/dir.go` | ✅ |
| SQL 模板 | .sql | `pkg/dump/write/` | 🚧 开发中 |
| 控制台 | - | `pkg/dump/write/console.go` | ✅ |

#### 使用示例

```bash
# 导出指定 uid 的用户数据到 ZIP
connect mysql dump -h localhost -u root -d gforge -t user uid1 uid2 uid3

# 导出并指定输出格式（ZIP）
connect mysql dump -h localhost -u root -d gforge -t user --in zip uid1

# 导出到目录
connect mysql dump -h localhost -u root -d gforge -t user --in dir uid1

# 导出到控制台（调试用）
connect mysql dump -h localhost -u root -d gforge -t user --in - uid1
```

#### 核心代码引用

- 命令入口: `lvan/cmd/dumper/cmd/dump.go:21-104`
- 导出逻辑: `lvan/pkg/dump/dump.go:12-59`
- 列查询: `lvan/pkg/dump/mysql.go` (Columns 函数)
- 主键查询: `lvan/pkg/dump/mysql.go` (GetPrimaryKeyColumns 函数)

---

### 2. 数据导入 (Import)

#### 功能描述

从导出文件加载数据并导入到目标数据库。

#### 支持的输入格式

| 格式 | 加载器 | 代码位置 | 状态 |
|------|--------|---------|------|
| ZIP 压缩 | Zip() | `pkg/dump/load/zip.go` | ✅ |
| 目录结构 | Dir() | `pkg/dump/load/dir.go` | ✅ |
| 控制台 | Console() | `pkg/dump/load/console.go` | ✅ |

#### 使用示例

```bash
# 从 ZIP 导入
connect mysql import -h localhost -u root -d gforge -t user ./data.zip

# 从目录导入
connect mysql import -h localhost -u root -d gforge -t user ./data_dir

# 导入多个文件
connect mysql import -h localhost -u root -d gforge -t user ./data1.zip ./data2.zip
```

#### 核心代码引用

- 命令入口: `lvan/cmd/dumper/cmd/import.go:18-80`
- 导入逻辑: `lvan/pkg/dump/import.go`
- 连接管理: `lvan/pkg/dump/lifecycle.go`

---

### 2.1 SQL 文件导入 (Import-SQL) 🆕

#### 功能描述

从 `mysql-dump` 导出的 SQL 文件导入数据，转换为 LVAN Dumper 的标准格式。

**设计文档**: [docs/DESIGN_SQL_IMPORT.md](docs/DESIGN_SQL_IMPORT.md)

#### 支持的 SQL 格式

| 格式 | 说明 | 状态 |
|------|------|------|
| 标准 INSERT | `INSERT INTO table VALUES (1,a),(2,b);` | ✅ 计划中 |
| 扩展插入 | 多值 INSERT | ✅ 计划中 |
| mysql-dump 默认输出 | 完整 SQL 文件 | ✅ 计划中 |

#### 导入策略

**Hybrid 方案**（快速解析 + MySQL 回退）:

```
1. 快速解析尝试
   ↓ 失败
2. MySQL/Dolt 回退
   ↓ 导入到临时数据库
3. 使用现有 dump 功能
   ↓ 导出为 ZIP/DIR
4. 清理临时数据
```

#### 使用示例

```bash
# 基础用法 - 导入 SQL 文件
connect mysql import-sql dump.sql

# 指定输出格式为目录
connect mysql import-sql dump.sql --in dir

# 指定数据库连接（用于回退方案）
connect mysql import-sql dump.sql -h localhost -P 3307 -u root

# 保留临时数据库（调试用）
connect mysql import-sql dump.sql --keep-temp

# 设置超时时间
connect mysql import-sql dump.sql --timeout 10m
```

#### 命令选项

| 选项 | 说明 | 默认值 |
|------|------|-------|
| `--temp-db` | 临时数据库名 | 自动生成 |
| `--keep-temp` | 保留临时数据库 | false |
| `--timeout` | 导入超时时间 | 5m |

#### 核心代码引用

- 命令入口: `lvan/cmd/dumper/cmd/import_sql.go` (待创建)
- SQL 解析器: `lvan/pkg/dump/load/sql.go` (待创建)
- 回退方案: 使用现有 `dump` 和 `import` 功能

---

### 3. MySQL 数据源

#### 功能描述

提供 MySQL 数据库的连接、查询和元数据获取功能。

#### 核心能力

```go
// 代码位置: lvan/pkg/dump/mysql.go

// 获取表的所有列名
func Columns(db *sql.DB, database, table string) []string

// 获取表的主键列
func GetPrimaryKeyColumns(db *sql.DB, database, table string) ([]string, error)

// 创建数据库连接
func ConnC(config dump.Config) func(func(dump.Datasource))
```

#### 配置参数

| 参数 | 短选项 | 默认值 | 说明 |
|------|--------|-------|------|
| --host | -h | localhost | MySQL 主机 |
| --port | -P | 3306 | MySQL 端口 |
| --user | -u | root | MySQL 用户名 |
| --password | -p | "" | MySQL 密码 |
| --database | -d | gforge | 数据库名 |
| --table | -t | user | 表名 |

---

### 4. 数据源抽象 v2

#### 功能描述

新一代数据源抽象层，使用访问者模式实现更灵活的数据操作。

#### 架构设计

```
                Datasource (interface)
                       △
                       │
        ┌──────────────┴──────────────┐
        │                             │
   MySQLDatasource              RedisDatasource
        │                             │
        │ Accept(visitor)             │
        ▼                             ▼
   Visitor ───────────────► VisitMySQL(ds)
                            VisitRedis(ds)
```

#### 核心代码引用

- 接口定义: `lvan/pkg/dump/datasource/v2/`
- MySQL 实现: `mysql_datasource.go`
- 访问者接口: `datasource.go`
- 适配器: `mysql_testdb_adapter.go`

---

### 5. 序列化格式

#### ZIP 格式

**目录结构**:
```
database.table.zip
├── {primary_key_value}/
│   ├── column1          (列名文件，内容为字段的值)
│   ├── column2
│   └── ...
├── {primary_key_value}/
│   ├── column1
│   └── ...
```

**代码位置**: `lvan/pkg/dump/type/zip.go:14-69`

#### 目录格式

**目录结构**:
```
database.table/
├── {primary_key_value}/
│   ├── column1
│   ├── column2
│   └── ...
└── ...
```

**代码位置**: `lvan/pkg/dump/type/dir.go`

---

### 6. MySQL Mock 测试框架

#### 功能描述

基于 go-mysql-server 的内存 MySQL 模拟框架，支持无需外部数据库的完整测试。

#### 核心特性

- 🔄 **双模式运行**: 自动检测真实 MySQL，失败时切换到模拟模式
- 📝 **配置驱动**: 支持 YAML 配置文件定义表结构和数据
- 🧪 **真实协议**: 使用 MySQL 协议而非简单 Mock
- 🚀 **快速启动**: 内存模式启动 < 10ms

#### 使用示例

```go
// 设置测试环境
if err := testdb.SetupTestMySQLEnvironment(); err != nil {
    t.Fatalf("设置测试环境失败: %v", err)
}
defer testdb.TeardownTestMySQLEnvironment()

// 获取数据库连接
adapter := testdb.GetGlobalMySQLTestDBAdapter()
conn := adapter.GetMySQLTestDB().GetMySQLConnection()

// 执行测试...
```

**详细文档**: `lvan/pkg/testdb/README.md`

---

## 命令行接口

### 命令结构

```
connect <datasource> <action> [flags] <args>

datasource: mysql | redis
action: dump | import
```

### 全局选项

| 选项 | 说明 | 默认值 |
|------|------|-------|
| --config | 配置文件路径 | $HOME/.dumper.yaml |
| --in | 输出格式 | zip |

### Dump 选项

| 选项 | 说明 | 默认值 |
|------|------|-------|
| --output, -o | 输出路径 | 自动生成 |
| --where, -w | 查询列名 | uid |
| --simulate | 模拟模式 | false |

### Import 选项

| 选项 | 说明 | 默认值 |
|------|------|-------|
| (无特定选项) | 输入文件作为位置参数 | - |

---

## 配置文件

### YAML 配置示例

```yaml
# ~/.dumper.yaml
host: localhost
port: 3306
user: root
password: ""
database: gforge
table: user
in: zip  # zip | dir | sql-tpl | -
output: ./output
```

---

## 测试

### 测试框架

| 组件 | 说明 |
|------|------|
| 单元测试 | Go testing + testify |
| Mock 框架 | go-mysql-server (内存模式) |
| **集成测试** | **Dolt (轻量级 Git 版本数据库)** |

### 测试脚本

| 脚本 | 用途 |
|------|------|
| `tests/setup_test_db.sh` | 创建 Dolt 测试数据库 |
| `tests/setup_test_db.bat` | Windows 版本 |
| `tests/e2e_test.sh` | 端到端测试 |

### 测试数据

Dolt 测试数据库覆盖所有 MySQL 数据类型：

```sql
CREATE TABLE user (
    uid INT PRIMARY KEY AUTO_INCREMENT,
    accountid VARCHAR(50) NOT NULL UNIQUE,
    data BLOB,                    -- 🔑 Protobuf 数据
    int_val INT,
    float_val FLOAT,
    blob_val BLOB,                -- 🔑 二进制数据
    text_val TEXT,
    json_val JSON,
    -- ... 其他数据类型
);
```

**测试重点**：
- 🔑 **BLOB 字段完整性**（protobuf 数据）
- 🔄 完整的 dump → import 流程
- 📦 ZIP 和 DIR 两种传输方式

### 运行测试

```bash
# 准备测试数据库
bash tests/setup_test_db.sh

# 运行端到端测试
bash tests/e2e_test.sh

# 单元测试
cd lvan && go test ./pkg/dump/... -v

# CLI 手动测试
go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user \
  test_user_001
```

详细测试用例见: [docs/REGRESSION_TEST.md](docs/REGRESSION_TEST.md)

---

## 依赖组件

### 核心依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| spf13/cobra | v1.10.1 | CLI 框架 |
| spf13/viper | v1.21.0 | 配置管理 |
| go-sql-driver/mysql | v1.9.3 | MySQL 驱动 |
| go-spring/spring-core | v1.2.5 | 依赖注入 |
| vmihailenco/msgpack | v5.4.1 | 二进制序列化 |
| stretchr/testify | v1.11.1 | 测试断言 |

---

## 开发路线图

### 已完成 ✅

- [x] MySQL dump 基础功能
- [x] MySQL import 基础功能
- [x] ZIP 序列化格式
- [x] 目录序列化格式
- [x] MySQL Mock 测试框架
- [x] v2 数据源抽象
- [x] **Dolt 测试数据库脚本**
- [x] **端到端测试脚本**
- [x] **回归测试文档完善**
- [x] **BLOB 字段专项测试**

### 进行中 🚧

- [ ] SQL 模板导出格式
- [ ] Redis dump/import 完善

### 计划中 📋

- [ ] **SQL 文件导入功能 (import-sql)** 🆕
  - 设计文档: [docs/DESIGN_SQL_IMPORT.md](docs/DESIGN_SQL_IMPORT.md)
  - 方案: Hybrid (快速解析 + MySQL 回退)
  - 优先级: 高
- [ ] PostgreSQL 数据源支持
- [ ] MongoDB 数据源支持
- [ ] 数据加密功能
- [ ] 增量导出/导入
- [ ] 并行导出优化
- [ ] Web 管理界面

---

## 贡献指南

1. 遵循现有代码风格
2. 为新功能添加测试
3. 更新相关文档
4. 提交 PR 前运行 `go test ./...`

---

*维护者: LVAN Dumper 开发团队*
*最后更新: 2025-01-31*
