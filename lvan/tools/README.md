# MySQL 表数据导出和导入替换脚本

## 目录

- [数据导出脚本 (dump.sh)](#数据导出脚本-dumpsh)
- [数据导入和替换脚本 (import_replace.sh)](#数据导入和替换脚本-import_replacesh)

---

## 数据导出脚本 (dump.sh)

### 功能特性

- 按指定ID列表导出MySQL/MariaDB表数据
- 支持自定义主机、用户、密码、数据库和表名
- 自动生成SQL文件，文件名包含表名和ID信息
- 支持模拟模式，无需连接实际数据库即可生成测试文件
- 支持MySQL风格的stdin输入密码方式
- 参数验证和错误处理

### 使用方法

```bash
./dump.sh [选项] [ID列表]
```

### 选项说明

- `-h, --host HOST`：MySQL主机名（默认：localhost）
- `-u, --user USER`：MySQL用户名（默认：root）
- `-p, --password PASSWORD`：MySQL密码（可选，支持MySQL风格的stdin输入方式）
- `-d, --database DATABASE`：数据库名（默认：gforge）
- `-t, --table TABLE`：表名（默认：user）
- `-w, --where COLUMN`：WHERE条件列名（默认：uid）
- `--simulate`：模拟模式，不实际连接数据库
- `--help`：显示帮助信息

### 使用示例

#### 基本用法

```bash
# 导出account表中ID为1,2,3的记录
./dump.sh -h localhost -u root -p password -d gforge -t account 1 2 3

# 导出user表中ID为1,2,3的记录
./dump.sh -h localhost -u root -p password -d gforge -t user 1 2 3

# 指定条件导出user表中ID为1,2,3的记录
./dump.sh -h localhost -u root -p password -d gforge -t user -w uid 1 2 3

# 使用默认数据库和表名导出user表中ID为1,2,3的记录
./dump.sh -h localhost -u root -p password 1 2 3
```

#### MySQL风格的stdin输入密码方式

脚本支持MySQL风格的stdin输入密码方式，即只指定`-p`选项而不提供密码值：

```bash
# 使用stdin输入密码
./dump.sh -h localhost -u root -p -d gforge -t user 1 2 3

# 或者使用管道输入密码
echo "password" | ./dump.sh -h localhost -u root -p -d gforge -t user 1 2 3
```

当使用`-p`选项但不提供密码值时，脚本会将密码输入委托给mysqldump/mariadb-dump命令，该命令会从stdin读取密码。

#### 模拟模式

```bash
# 使用模拟模式生成测试文件
./dump.sh --simulate -d testdb -t users 1 2 3

# 模拟模式下导出整个表
./dump.sh --simulate -d testdb -t users
```

### 输出文件

脚本会根据表名和ID列表自动生成SQL文件：

- 导出特定ID：`表名_列名_ID列表.sql`（例如：`user_uid_1,2,3.sql`）
- 导出整个表：`表名_all.sql`（例如：`user_all.sql`）

### 模拟模式

在模拟模式下，脚本不会连接实际数据库，而是生成包含模拟数据的SQL文件，用于测试目的。

### 错误处理

脚本包含以下验证机制：

- 验证ID列表必须为数字
- 检查mysqldump/mariadb-dump命令是否存在
- 为主机名和用户名提供默认值（主机名默认为localhost，用户名默认为root）

### 依赖

- bash
- mysqldump 或 mariadb-dump

---

## 数据导入和替换脚本 (import_replace.sh)

### 功能特性

- 创建临时库用于数据处理
- 将指定SQL文件导入临时库
- 根据参数替换指定列的旧值为新值
- 将更新后的数据导入目标库对应表
- 执行参数指定的后续SQL
- 自动清理临时库

### 使用方法

```bash
./import_replace.sh [选项]
```

### 选项说明

- `-h, --host HOST`：MySQL主机名（默认：localhost）
- `-u, --user USER`：MySQL用户名（默认：root）
- `-p, --password PASSWORD`：MySQL密码（可选）
- `-d, --database DATABASE`：目标数据库名（必需）
- `-t, --table TABLE`：目标表名（必需）
- `-c, --column COLUMN`：要替换值的列名
- `-o, --old OLD_VALUE`：要被替换的旧值
- `-n, --new NEW_VALUE`：替换后的新值
- `-f, --file SQL_FILE`：要导入的SQL文件路径（必需）
- `-r, --replace REPLACE_SQL`：替换操作的SQL语句（可选，替代-c/-o/-n参数）
- `-s, --post POST_SQL`：导入后的后续SQL语句（可选）
- `--help`：显示帮助信息

### 工作流程

1. 创建临时库
2. 将指定SQL文件导入临时库
3. 根据参数替换指定列的旧值为新值
4. 将更新后的数据导入目标库对应表
5. 执行参数指定的后续SQL
6. 清理临时库

### 使用示例

#### 基本用法

```bash
# 导入数据并替换列值
./import_replace.sh -h localhost -u root -p password -d mydb -t mytable -c status -o old -n new -f data.sql
```

#### 使用自定义替换SQL

```bash
# 使用自定义SQL进行替换
./import_replace.sh -h localhost -u root -p password -d mydb -t mytable -r "UPDATE mytable SET status='new' WHERE status='old'" -f data.sql
```

#### 执行后续SQL

```bash
# 导入数据并执行后续SQL
./import_replace.sh -h localhost -u root -p password -d mydb -t mytable -f data.sql -s "UPDATE mytable SET updated_at=NOW()"
```

### 注意事项

- 脚本会自动创建和删除临时库，请确保MySQL用户有足够的权限
- 脚本假设SQL文件中的表结构与目标库中的表结构一致
- 如果目标表不存在，脚本会尝试根据临时库中的表结构创建目标表