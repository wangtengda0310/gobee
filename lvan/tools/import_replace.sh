#!/bin/bash

# MySQL 数据导入和替换脚本
# 用法: ./import_replace.sh [选项]

# 默认参数
HOST="localhost"
USER="root"
PASSWORD=""
DATABASE=""
TABLE=""
COLUMN=""
OLD_VALUE=""
NEW_VALUE=""
SQL_FILE=""
TEMP_DATABASE="temp_import"
REPLACE_SQL=""
POST_SQL=""

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --host HOST          MySQL 主机名 (默认: localhost)"
    echo "  -u, --user USER          MySQL 用户名 (默认: root)"
    echo "  -p, --password PASSWORD  MySQL 密码 (空密码时可不提供)"
    echo "  -d, --database DATABASE  目标数据库名"
    echo "  -t, --table TABLE        目标表名"
    echo "  -c, --column COLUMN      要替换值的列名"
    echo "  -o, --old OLD_VALUE      要被替换的旧值"
    echo "  -n, --new NEW_VALUE      替换后的新值"
    echo "  -f, --file SQL_FILE      要导入的SQL文件路径"
    echo "  -r, --replace REPLACE_SQL  替换操作的SQL语句"
    echo "  -s, --post POST_SQL      导入后的后续SQL语句"
    echo "  --help                   显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -h localhost -u root -p password -d mydb -t mytable -c status -o old -n new -f data.sql"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            HOST="$2"
            shift
            shift
            ;;
        -u|--user)
            USER="$2"
            shift
            shift
            ;;
        -p|--password)
            # 检查是否有密码值参数
            if [[ -n "$2" && "$2" != -* ]]; then
                PASSWORD="$2"
                shift
            else
                # 没有密码值参数，使用空密码
                PASSWORD=""
            fi
            shift
            ;;
        -d|--database)
            DATABASE="$2"
            shift
            shift
            ;;
        -t|--table)
            TABLE="$2"
            shift
            shift
            ;;
        -c|--column)
            COLUMN="$2"
            shift
            shift
            ;;
        -o|--old)
            OLD_VALUE="$2"
            shift
            shift
            ;;
        -n|--new)
            NEW_VALUE="$2"
            shift
            shift
            ;;
        -f|--file)
            SQL_FILE="$2"
            shift
            shift
            ;;
        -r|--replace)
            REPLACE_SQL="$2"
            shift
            shift
            ;;
        -s|--post)
            POST_SQL="$2"
            shift
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证必需参数
if [[ -z "$DATABASE" ]]; then
    echo "错误: 必须指定目标数据库名 (-d, --database)"
    show_help
    exit 1
fi

if [[ -z "$TABLE" ]]; then
    echo "错误: 必须指定目标表名 (-t, --table)"
    show_help
    exit 1
fi

if [[ -z "$SQL_FILE" ]]; then
    echo "错误: 必须指定要导入的SQL文件 (-f, --file)"
    show_help
    exit 1
fi

if [[ ! -f "$SQL_FILE" ]]; then
    echo "错误: 指定的SQL文件不存在: $SQL_FILE"
    exit 1
fi

# 构建MySQL命令
if command -v mysql &> /dev/null; then
    MYSQL_CMD="mysql"
elif command -v mariadb &> /dev/null; then
    MYSQL_CMD="mariadb"
else
    echo "错误: 未找到 mysql 或 mariadb 命令!"
    exit 1
fi

# 构建基本MySQL命令
MYSQL_CMD="$MYSQL_CMD -h $HOST -u $USER"
if [[ -n "$PASSWORD" ]]; then
    MYSQL_CMD="$MYSQL_CMD -p$PASSWORD"
fi

# 步骤1: 创建临时库
echo "步骤1: 创建临时库 $TEMP_DATABASE"
$MYSQL_CMD -e "DROP DATABASE IF EXISTS $TEMP_DATABASE; CREATE DATABASE $TEMP_DATABASE;"
if [[ $? -ne 0 ]]; then
    echo "错误: 创建临时库失败!"
    exit 1
fi

# 步骤2: 将指定文件导入临时库
echo "步骤2: 将指定文件导入临时库"
$MYSQL_CMD $TEMP_DATABASE < "$SQL_FILE"
if [[ $? -ne 0 ]]; then
    echo "错误: 导入SQL文件到临时库失败!"
    exit 1
fi

# 步骤3: 根据参数指定column替换旧值为新值
if [[ -n "$COLUMN" && -n "$OLD_VALUE" && -n "$NEW_VALUE" ]]; then
    echo "步骤3: 替换 $TABLE 表中 $COLUMN 列的值，将 '$OLD_VALUE' 替换为 '$NEW_VALUE'"
    $MYSQL_CMD $TEMP_DATABASE -e "UPDATE $TABLE SET $COLUMN = '$NEW_VALUE' WHERE $COLUMN = '$OLD_VALUE';"
    if [[ $? -ne 0 ]]; then
        echo "错误: 替换值失败!"
        exit 1
    fi
elif [[ -n "$REPLACE_SQL" ]]; then
    echo "步骤3: 执行自定义替换SQL"
    $MYSQL_CMD $TEMP_DATABASE -e "$REPLACE_SQL"
    if [[ $? -ne 0 ]]; then
        echo "错误: 执行自定义替换SQL失败!"
        exit 1
    fi
else
    echo "步骤3: 跳过值替换步骤（未提供替换参数）"
fi

# 步骤4: 将更新后的数据导入指定库中的对应表
echo "步骤4: 将更新后的数据导入目标库 $DATABASE.$TABLE"
# 首先确保目标表存在
$MYSQL_CMD $DATABASE -e "CREATE TABLE IF NOT EXISTS $TABLE LIKE $TEMP_DATABASE.$TABLE;"
if [[ $? -ne 0 ]]; then
    echo "错误: 创建目标表失败!"
    exit 1
fi

# 然后将数据从临时库导入目标库
$MYSQL_CMD $DATABASE -e "INSERT INTO $TABLE SELECT * FROM $TEMP_DATABASE.$TABLE;"
if [[ $? -ne 0 ]]; then
    echo "错误: 导入数据到目标库失败!"
    exit 1
fi

# 步骤5: 执行参数指定的后续SQL
if [[ -n "$POST_SQL" ]]; then
    echo "步骤5: 执行后续SQL"
    $MYSQL_CMD $DATABASE -e "$POST_SQL"
    if [[ $? -ne 0 ]]; then
        echo "错误: 执行后续SQL失败!"
        exit 1
    fi
else
    echo "步骤5: 跳过后续SQL执行（未提供后续SQL参数）"
fi

# 清理临时库
echo "清理: 删除临时库 $TEMP_DATABASE"
$MYSQL_CMD -e "DROP DATABASE $TEMP_DATABASE;"
if [[ $? -ne 0 ]]; then
    echo "警告: 删除临时库失败!"
fi

echo "所有步骤完成!"