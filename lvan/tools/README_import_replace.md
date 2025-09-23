# MySQL 数据导入和替换脚本

## 功能特性

- 创建临时库用于数据处理
- 将指定SQL文件导入临时库
- 根据参数替换指定列的旧值为新值
- 将更新后的数据导入目标库对应表
- 执行参数指定的后续SQL
- 自动清理临时库

## 使用方法

```bash
./import_replace.sh [选项]
```

## 选项说明

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

## 工作流程

1. 创建临时库
2. 将指定SQL文件导入临时库
3. 根据参数替换指定列的旧值为新值
4. 将更新后的数据导入目标库对应表
5. 执行参数指定的后续SQL
6. 清理临时库

## 使用示例

### 基本用法

```bash
# 导入数据并替换列值
./import_replace.sh -h localhost -u root -p password -d mydb -t mytable -c status -o old -n new -f data.sql
```

### 使用自定义替换SQL

```bash
# 使用自定义SQL进行替换
./import_replace.sh -h localhost -u root -p password -d mydb -t mytable -r "UPDATE mytable SET status='new' WHERE status='old'" -f data.sql
```

### 执行后续SQL

```bash
# 导入数据并执行后续SQL
./import_replace.sh -h localhost -u root -p password -d mydb -t mytable -f data.sql -s "UPDATE mytable SET updated_at=NOW()"
```

## 注意事项

- 脚本会自动创建和删除临时库，请确保MySQL用户有足够的权限
- 脚本假设SQL文件中的表结构与目标库中的表结构一致
- 如果目标表不存在，脚本会尝试根据临时库中的表结构创建目标表