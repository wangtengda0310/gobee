# Dumper 命令行工具

## 简介
Dumper 是一个用于数据库导出和导入的命令行工具。

## 使用方法

### 基本命令
```bash
# 导出数据
./dumper dump

# 导入数据
./dumper import
```

### 配置参数

#### 1. 通过命令行标志设置参数
```bash
go run main.go dump mysql -h 101.34.211.79 -P 32533 -p p_mysql -d gforge_u01_alpha2 1

# 设置import-dir参数
./dumper import --import-dir="gforge_u01_alpha2.user/0"

# 设置数据库连接参数
./dumper import -h localhost -P 3306 -u root -p password -d gforge -t user --import-dir="gforge_u01_alpha2.user/0"
```

#### 2. 通过配置文件设置参数
创建配置文件 `.dumper.yaml`:
```yaml
host: localhost
port: 3306
user: root
password: ""
database: gforge
table: user
import:
  dir: "gforge_u01_alpha2.user/0"
```

#### 3. 通过环境变量设置参数
```bash
export DUMP_HOST=localhost
export DUMP_PORT=3306
export DUMP_USER=root
export DUMP_PASSWORD=password
export DUMP_DATABASE=gforge
export DUMP_TABLE=user
export DUMP_IMPORT_DIR="gforge_u01_alpha2.user/0"
```

参数优先级: 命令行标志 > 配置文件 > 环境变量 > 默认值