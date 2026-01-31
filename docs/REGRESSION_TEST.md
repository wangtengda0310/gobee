# LVAN Dumper 回归测试文档

## 概述

本文档记录 LVAN Dumper 项目的回归测试用例，确保功能修改不会破坏现有行为。

**测试重点**：
- ✅ 覆盖所有 MySQL 数据类型
- 🔍 **重点测试 BLOB 字段**（生产环境存储 protobuf 数据）
- 🔄 完整的 dump → import 流程验证
- 📦 ZIP 和 DIR 两种传输方式

---

## 测试环境

| 组件 | 配置 |
|------|------|
| Go 版本 | 1.25.0 |
| 测试框架 | testing + testify |
| Mock 框架 | go-mysql-server (内存模式) |
| 测试数据库 | **Dolt** (轻量级 Git 版本数据库) |
| Dolt 端口 | 3307 |
| 运行方式 | `go run` (非 go build) |

---

## 一、环境准备

### 1.1 使用 Dolt 准备测试数据

**脚本位置**: `tests/setup_test_db.sh` (Linux/Mac) / `tests/setup_test_db.bat` (Windows)

**Dolt 安装验证**:
```bash
dolt --version
```

**启动测试数据库**:
```bash
# Windows
tests\setup_test_db.bat

# Linux/Mac
bash tests/setup_test_db.sh
```

**测试数据 Schema** (覆盖所有 MySQL 类型):

```sql
CREATE TABLE user (
    uid INT PRIMARY KEY AUTO_INCREMENT,
    accountid VARCHAR(50) NOT NULL UNIQUE,
    username VARCHAR(100),
    data BLOB,                    -- 🔑 BLOB: protobuf 数据
    tiny_val TINYINT,
    small_val SMALLINT,
    medium_val MEDIUMINT,
    int_val INT,
    big_val BIGINT,
    float_val FLOAT,
    double_val DOUBLE,
    decimal_val DECIMAL(10,2),
    bool_val BOOLEAN,
    date_val DATE,
    time_val TIME,
    datetime_val DATETIME,
    timestamp_val TIMESTAMP,
    year_val YEAR,
    char_val CHAR(10),
    varchar_val VARCHAR(255),
    text_val TEXT,
    blob_val BLOB,                -- 🔑 BLOB: 二进制数据
    json_val JSON,
    enum_val ENUM('A', 'B', 'C'),
    set_val SET('X', 'Y', 'Z'),
    ctime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    mtime TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**测试数据记录**:

| accountid | 说明 | data (BLOB) | blob_val | 特殊点 |
|-----------|------|-------------|----------|--------|
| test_user_001 | 完整数据 | protobuf | UTF-8 二进制 | 所有字段有值 |
| test_user_002 | 边界值 | NULL | NULL | 最小/负数值 |
| test_user_003 | NULL 测试 | NULL | NULL | 仅必填字段 |
| test_user_004 | 特殊字符 | 二进制 | 二进制 | 引号/转义字符 |
| test_user_005 | 大 BLOB | - | 1KB 数据 | 大块二进制 |

---

## 二、CLI 命令回归测试

### 2.1 MySQL Dump 命令

**代码位置**: `lvan/cmd/dumper/cmd/dump.go:34-56`

#### 测试用例组 A: 基础功能

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| A01 | `go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P 3307 -u root -d lvan_dumper_test -t user test_user_001` | 生成 `lvan_dumper_test.user.zip` | 检查文件存在 |
| A02 | 同上，多个 uid | ZIP 包含 2 个目录 | 解压验证 |
| A03 | 不存在的 uid | 空 ZIP 或 0 记录 | 检查日志 |
| A04 | `--in dir` | 生成目录结构 | 检查目录 |

**执行示例**:
```bash
cd E:\wangtengda\lvan_dumper

# ZIP 格式导出
go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user \
  test_user_001 test_user_002

# 预期输出: lvan_dumper_test.user.zip
```

#### 测试用例组 B: 输出格式

| 用例ID | --in 参数 | 预期输出 | 验证方法 |
|--------|----------|---------|---------|
| B01 | `zip` (默认) | `database.table.zip` | ZIP 文件 |
| B02 | `dir` | `database.table/` 目录 | 目录结构 |
| B03 | `-` (控制台) | JSON 输出到 stdout | 检查输出 |

**执行示例**:
```bash
# ZIP 格式
go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P 3307 -u root -d lvan_dumper_test -t user --in zip test_user_001

# 目录格式
go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P 3307 -u root -d lvan_dumper_test -t user --in dir test_user_001

# 控制台输出
go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P 3307 -u root -d lvan_dumper_test -t user --in - test_user_001
```

#### 测试用例组 C: ZIP 结构验证

**ZIP 内部结构**:
```
lvan_dumper_test.user.zip
├── test_user_001/
│   ├── uid              # 内容: "1"
│   ├── accountid        # 内容: "test_user_001"
│   ├── username         # 内容: "测试用户001"
│   ├── data             # 🔑 BLOB: protobuf 二进制
│   ├── blob_val         # 🔑 BLOB: UTF-8 二进制
│   ├── int_val          # 内容: "2147483647"
│   └── ...              # 其他字段
├── test_user_002/
│   └── ...
```

**验证脚本**:
```bash
# 解压检查
unzip -l lvan_dumper_test.user.zip

# 检查 BLOB 字段
unzip -p lvan_dumper_test.user.zip test_user_001/data | xxd | head
# 预期: protobuf 二进制数据
```

---

### 2.2 MySQL Import 命令

**代码位置**: `lvan/cmd/dumper/cmd/import.go:35-62`

#### 测试用例组 D: 导入功能

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| D01 | `go run lvan/cmd/dumper/main.go mysql import -h 127.0.0.1 -P 3307 -u root -d lvan_dumper_test -t user ./data.zip` | 数据导入成功 | 查询验证 |
| D02 | 目录格式导入 | 同上 | 查询验证 |
| D03 | 重复主键 | 根据策略处理 | 检查行为 |

**执行示例**:
```bash
# 导入 ZIP
go run lvan/cmd/dumper/main.go mysql import \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user \
  ./lvan_dumper_test.user.zip

# 验证导入
dolt sql "SELECT uid, accountid, HEX(data) as data_hex FROM user WHERE accountid='test_user_001';"
```

---

### 2.3 SQL 文件导入 (Import-SQL) 🆕

**代码位置**: `lvan/cmd/dumper/cmd/import_sql.go` (待创建)

#### 测试用例组 E: SQL 文件导入

| 用例ID | 场景 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| E01 | 标准 INSERT SQL | 转换为 ZIP | 检查 ZIP 文件 |
| E02 | 扩展插入 SQL | 转换为 ZIP | 检查记录数 |
| E03 | 多表 SQL 文件 | 生成多个 ZIP | 检查输出文件 |
| E04 | 包含 CREATE TABLE | 正确处理表结构 | 验证表存在 |
| E05 | 快速解析失败 | 回退到 MySQL | 检查日志 |
| E06 | 大 SQL 文件 | 在超时内完成 | 记录时间 |

#### 测试数据准备

```bash
# 创建测试 SQL 文件
cat > tests/data/test_dump.sql << 'EOF'
-- MySQL dump 10.13
--
CREATE TABLE `user` (
  `uid` int(11) NOT NULL,
  `accountid` varchar(50) DEFAULT NULL,
  `username` varchar(100) DEFAULT NULL,
  `data` blob DEFAULT NULL,
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `user` WRITE;
INSERT INTO `user` VALUES (1,'test_user_001','测试用户001',_binary '08019601010A0774657374696E67');
INSERT INTO `user` VALUES (2,'test_user_002','边界值测试',NULL);
INSERT INTO `user` VALUES (3,'test_user_003','NULL测试',NULL);
UNLOCK TABLES;
EOF
```

#### 执行示例

```bash
cd E:\wangtengda\lvan_dumper

# 基础导入测试
go run lvan/cmd/dumper/main.go mysql import-sql \
  tests/data/test_dump.sql

# 预期输出: user.zip

# 验证 ZIP 内容
unzip -l user.zip

# 验证数据一致性
dolt sql "
SELECT
    HEX(o.data) = HEX(i.data) as blob_match,
    o.username = i.username as name_match
FROM (SELECT * FROM user WHERE uid <= 3) o
LEFT JOIN (SELECT * FROM user_import WHERE uid <= 3) i ON o.uid = i.uid;
"
```

#### 扩展插入测试

```bash
# 创建扩展插入 SQL 文件
cat > tests/data/test_extended.sql << 'EOF'
CREATE TABLE `user` (
  `uid` int(11) NOT NULL,
  `accountid` varchar(50) DEFAULT NULL,
  `data` blob DEFAULT NULL,
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB;

INSERT INTO `user` VALUES
  (1,'user001',_binary'data1'),
  (2,'user002',_binary'data2'),
  (3,'user003',NULL);
EOF

# 导入测试
go run lvan/cmd/dumper/main.go mysql import-sql \
  tests/data/test_extended.sql

# 预期: 3 条记录正确解析
```

#### 性能测试

```bash
# 创建大 SQL 文件 (1000 条记录)
dolt sql "
INSERT INTO user (uid, accountid, data)
SELECT
    seq,
    CONCAT('user_', seq),
    UNHEX(REPEAT(HEX(seq), 100))
FROM (
    SELECT ROW_NUMBER() OVER () as seq
    FROM information_schema.columns
    CROSS JOIN information_schema.tables
    LIMIT 1000
) t;
"

# 导出为 SQL
dolt sql --batch --result-format=single > tests/data/large_dump.sql << EOF
SELECT CONCAT('INSERT INTO \`user\` VALUES (',
    uid, ',''', accountid, ''',_binary''',
    HEX(data), ''');') as stmt
FROM user;
EOF

# 测试导入性能
time go run lvan/cmd/dumper/main.go mysql import-sql \
  tests/data/large_dump.sql

# 预期: 1000 条记录在 30 秒内完成
```

---

## 三、BLOB 字段专项测试 🔑

**为什么重要**: 生产环境使用 BLOB 字段存储 protobuf 数据，必须保证克隆操作不损坏二进制数据。

### 3.1 Protobuf BLOB 测试

**测试数据**: 模拟 protobuf 序列化后的二进制

```python
# 生成测试 protobuf 数据
import struct
data = struct.pack('<B', 0x08)  # field 1, varint
data += struct.pack('<B', 0x96)  # value 150
data += struct.pack('<B', 0x0A)  # field 1, length-delimited
data += struct.pack('<B', 0x07)  # length 7
data += b'testing'               # string value
```

**测试用例**:

| 用例ID | 场景 | 验证方法 |
|--------|------|---------|
| BLOB01 | 导出 protobuf BLOB | HEX(data) = 原始值 |
| BLOB02 | ZIP 存储后导出 | 二进制一致 |
| BLOB03 | 大 BLOB (1KB) | 完整传输 |
| BLOB04 | UTF-8 二进制 | 字节级一致 |

**验证脚本**:
```bash
# 1. 导出前记录原始 HEX
dolt sql "SELECT accountid, HEX(data) as original FROM user WHERE accountid='test_user_001';"

# 2. 执行导出
go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P 3307 -u root -d lvan_dumper_test -t user test_user_001

# 3. 检查 ZIP 中的 BLOB
unzip -p lvan_dumper_test.user.zip test_user_001/data | xxd

# 4. 导入到新表并验证 HEX 一致性
# （需要创建测试表用于验证）
```

---

## 四、端到端测试

### 4.1 完整流程测试

**场景**: dump → ZIP → import → 数据一致性验证

```bash
#!/bin/bash
# e2e_test.sh

# 1. 准备测试数据库
bash tests/setup_test_db.sh

# 2. 导出数据
go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user \
  test_user_001 test_user_002 test_user_003 test_user_004

# 3. 创建导入测试表
dolt sql "CREATE TABLE user_import LIKE user;"

# 4. 导入数据
go run lvan/cmd/dumper/main.go mysql import \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user_import \
  ./lvan_dumper_test.user.zip

# 5. 数据一致性验证
dolt sql "
SELECT
    o.uid,
    o.accountid,
    HEX(o.data) = HEX(i.data) as blob_match,
    o.int_val = i.int_val as int_match,
    o.blob_val = i.blob_val as blob_val_match
FROM user o
LEFT JOIN user_import i ON o.accountid = i.accountid;
"

# 预期: 所有 match 列为 TRUE
```

### 4.2 数据一致性验证清单

| 字段类型 | 验证方法 | 预期结果 |
|---------|---------|---------|
| INT/BIGINT | 值比较 | 完全相等 |
| VARCHAR/TEXT | 字符串比较 | 完全相等 |
| BLOB | **HEX 比较** | 🔑 字节级一致 |
| DATE/DATETIME | 格式化比较 | 完全相等 |
| FLOAT/DOUBLE | 精度比较 | 误差 < 1e-6 |
| JSON | 结构比较 | 完全相等 |
| NULL | NULL 判断 | 一致 |

---

## 五、边界条件测试

### 5.1 空值处理

| 用例ID | 场景 | 预期行为 |
|--------|------|---------|
| E01 | 空 uid 列表 | 无操作或错误提示 |
| E02 | 不存在的 uid | 返回空结果 |
| E03 | NULL BLOB | 正确处理 NULL |
| E04 | 空字符串 | 区分空字符串和 NULL |

### 5.2 特殊字符处理

| 用例ID | 场景 | 预期行为 |
|--------|------|---------|
| E05 | 包含引号的字符串 | 正确转义 |
| E06 | 包含换行符的文本 | 正确转义 |
| E07 | 二进制 NULL 字节 | 完整保留 |
| E08 | UTF-8 多字节字符 | 字节级一致 |

### 5.3 大数据量

| 用例ID | 场景 | 预期行为 |
|--------|------|---------|
| E09 | 1KB BLOB | 完整传输 |
| E10 | 1MB BLOB | 完整传输（待测试） |
| E11 | 1000 条记录 | 合理时间内完成 |
| E12 | 混合大小记录 | 正确处理 |

---

## 六、单元测试（Go 代码）

### 6.1 已实现的单元测试

| 测试文件 | 测试内容 | 状态 |
|---------|---------|------|
| `pkg/dump/type/zip_test.go` | ZIP 序列化 | ✅ |
| `pkg/dump/type/dir_test.go` | 目录序列化 | ✅ |
| `pkg/dump/load/zip_test.go` | ZIP 加载 | ✅ |
| `pkg/testdb/mysql_test.go` | Mock 框架 | ✅ |
| `pkg/dump/datasource/v2/mock_integration_test.go` | v2 数据源集成 | ✅ |

### 6.2 运行单元测试

```bash
cd E:\wangtengda\lvan_dumper\lvan

# 所有测试
go test ./pkg/dump/... -v

# 带覆盖率
go test ./pkg/dump/... -cover

# 特定包
go test ./pkg/dump/type/ -run TestZip -v
```

---

## 七、运行回归测试

### 7.1 完整回归测试流程

```bash
#!/bin/bash
# run_regression.sh

set -e

echo "=== LVAN Dumper 回归测试 ==="

# 1. 环境检查
echo "检查 Dolt..."
dolt --version || { echo "❌ Dolt 未安装"; exit 1; }

# 2. 启动测试数据库
echo "启动测试数据库..."
bash tests/setup_test_db.sh

# 3. 单元测试
echo "运行单元测试..."
cd lvan
go test ./pkg/dump/... -v
cd ..

# 4. CLI 功能测试
echo "测试 MySQL Dump..."
go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user \
  test_user_001 test_user_002

# 5. 验证输出
if [ -f "lvan_dumper_test.user.zip" ]; then
    echo "✅ Dump 测试通过"
else
    echo "❌ Dump 测试失败"
    exit 1
fi

# 6. 端到端测试
echo "运行端到端测试..."
bash tests/e2e_test.sh

echo "=== 回归测试完成 ==="
```

### 7.2 快速验证命令

```bash
# 单独验证 BLOB 完整性
dolt sql "SELECT accountid, HEX(data) FROM user WHERE accountid='test_user_001';"
# 然后导出并检查 ZIP 中的 data 文件

# 验证导出的所有字段
unzip -l lvan_dumper_test.user.zip
```

---

## 八、已知问题与限制

| 问题ID | 描述 | 严重程度 | 状态 |
|-------|------|---------|------|
| K001 | Redis dump/import 未实现 | 低 | 📋 计划中 |
| K002 | 超大 BLOB (>1MB) 未测试 | 中 | 📊 待测 |
| K003 | 并发导出未测试 | 低 | 📋 待测 |

---

## 九、测试覆盖矩阵

| 功能 | 单元测试 | 集成测试 | E2E 测试 | BLOB 测试 |
|------|---------|---------|---------|---------|
| MySQL Dump | ✅ | ✅ | ✅ | ✅ |
| MySQL Import | ✅ | ✅ | ✅ | ✅ |
| ZIP 格式 | ✅ | ✅ | ✅ | ✅ |
| DIR 格式 | ✅ | ✅ | ⚠️ | ⚠️ |
| Redis | ❌ | ❌ | ❌ | ❌ |

---

## 十、故障排除

### 问题: Dolt SQL Server 无法连接

```bash
# 检查端口占用
netstat -an | grep 3307

# 停止旧的 Dolt 进程
pkill -f "dolt sql-server"

# 重新启动
dolt sql-server --port=3307 --user=root --password=""
```

### 问题: BLOB 数据不一致

```bash
# 使用 HEX 比较验证
dolt sql "SELECT HEX(data) FROM user WHERE uid=1;"
# 检查导出的 ZIP 中对应文件的 HEX 值
```

---

*维护者: LVAN Dumper 开发团队*
*最后更新: 2025-01-31*
*Dolt 版本要求: >= 1.0.0*
