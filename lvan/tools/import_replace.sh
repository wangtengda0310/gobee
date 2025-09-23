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
    echo "说明:"
    echo "  本脚本会执行以下操作:"
    echo "  1. 检查目标库和目标表是否存在"
    echo "  2. 检查临时库是否存在，如存在则自动清理"
    echo "  3. 创建临时库并导入SQL文件数据"
    echo "  4. 根据参数执行值替换操作"
    echo "  5. 将临时库中的数据追加到目标库表中"
    echo "  6. 执行后续SQL语句（如有）"
    echo "  7. 清理临时库"
    echo ""
    echo "注意:"
    echo "  - 数据导入是追加操作，不是替换操作"
    echo "  - 如果目标表中已有数据，执行脚本会导致数据重复"
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

# 步骤1: 检查目标库中目标表是否存在
echo "步骤1: 检查目标库 $DATABASE 中目标表 $TABLE 是否存在"
TABLE_EXISTS=$($MYSQL_CMD -N -s -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$DATABASE' AND table_name = '$TABLE';")
if [[ $? -ne 0 || "$TABLE_EXISTS" -ne 1 ]]; then
    echo "错误: 目标库 $DATABASE 中不存在目标表 $TABLE!"
    exit 1
fi

# 检查临时库是否已存在
TEMP_DB_EXISTS=$($MYSQL_CMD -N -s -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '$TEMP_DATABASE';")
if [[ $? -ne 0 ]]; then
    echo "错误: 检查临时库是否存在时发生错误!"
    exit 1
fi

if [[ "$TEMP_DB_EXISTS" -eq 1 ]]; then
    // 临时库存在，检查临时表是否存在
    TEMP_TABLE_EXISTS=$($MYSQL_CMD -N -s -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$TEMP_DATABASE' AND table_name = '$TABLE';")
    if [[ $? -ne 0 ]]; then
        echo "错误: 检查临时库中表是否存在时发生错误!"
        exit 1
    fi
    
    // 优化：当临时库存在但临时表不存在时，允许脚本继续执行，自动清理临时库
    echo "警告: 临时库 $TEMP_DATABASE 已存在，将自动清理..."
    $MYSQL_CMD -e "DROP DATABASE $TEMP_DATABASE;"
    if [[ $? -ne 0 ]]; then
        echo "错误: 清理已存在的临时库失败!"
        exit 1
    fi
fi

# 步骤2: 创建临时库
echo "步骤2: 创建临时库 $TEMP_DATABASE"
$MYSQL_CMD -e "CREATE DATABASE $TEMP_DATABASE;"
if [[ $? -ne 0 ]]; then
    echo "错误: 创建临时库失败!"
    exit 1
fi

# 步骤3: 将指定文件导入临时库
echo "步骤3: 将指定文件导入临时库"
$MYSQL_CMD $TEMP_DATABASE < "$SQL_FILE"
if [[ $? -ne 0 ]]; then
    echo "错误: 导入SQL文件到临时库失败!"
    exit 1
fi

# 步骤4: 根据参数指定column替换旧值为新值
if [[ -n "$COLUMN" && -n "$OLD_VALUE" && -n "$NEW_VALUE" ]]; then
    echo "步骤4: 替换临时库 $TEMP_DATABASE.$TABLE 表中 $COLUMN 列的值，将 '$OLD_VALUE' 替换为 '$NEW_VALUE'"
    $MYSQL_CMD $TEMP_DATABASE -e "UPDATE $TABLE SET $COLUMN = '$NEW_VALUE' WHERE $COLUMN = '$OLD_VALUE';"
    if [[ $? -ne 0 ]]; then
        echo "错误: 替换值失败!"
        exit 1
    fi
elif [[ -n "$REPLACE_SQL" ]]; then
    echo "步骤4: 执行自定义替换SQL（临时库：$TEMP_DATABASE）"
    $MYSQL_CMD $TEMP_DATABASE -e "$REPLACE_SQL"
    if [[ $? -ne 0 ]]; then
        echo "错误: 执行自定义替换SQL失败!"
        exit 1
    fi
else
    echo "步骤4: 跳过值替换步骤（未提供替换参数）"
fi

// 步骤5: 检查目标表中是否存在与临时表中相同的数据，避免重复导入
echo "步骤5: 检查数据冲突并导入数据到目标库 $DATABASE.$TABLE"

// 获取临时表中的记录数
TEMP_COUNT=$($MYSQL_CMD -N -s -e "SELECT COUNT(*) FROM $TEMP_DATABASE.$TABLE;")
if [[ $? -ne 0 ]]; then
    echo "错误: 获取临时表记录数失败!"
    exit 1
fi

if [[ "$TEMP_COUNT" -eq 0 ]]; then
    echo "警告: 临时表中没有数据，跳过数据导入步骤"
else
    // 检查目标表是否为空，如果不为空则提示用户
    TARGET_COUNT=$($MYSQL_CMD -N -s -e "SELECT COUNT(*) FROM $DATABASE.$TABLE;")
    if [[ $? -ne 0 ]]; then
        echo "错误: 获取目标表记录数失败!"
        exit 1
    fi
    
    if [[ "$TARGET_COUNT" -ne 0 ]]; then
        echo "警告: 目标表 $DATABASE.$TABLE 中已存在数据 ($TARGET_COUNT 条记录)"
        echo "注意: 本脚本执行的是数据追加操作，继续执行将导致数据重复"
        read -p "是否继续执行数据导入? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "用户取消操作，正在清理临时库..."
            $MYSQL_CMD -e "DROP DATABASE $TEMP_DATABASE;"
            if [[ $? -ne 0 ]]; then
                echo "警告: 删除临时库失败!"
            fi
            exit 1
        fi
    fi
    
    // 将数据从临时库导入目标库
    $MYSQL_CMD $DATABASE -e "INSERT INTO $TABLE SELECT * FROM $TEMP_DATABASE.$TABLE;"
    if [[ $? -ne 0 ]]; then
        echo "错误: 导入数据到目标库失败!"
        exit 1
    fi
    
    // 再次检查导入后的数据量
    NEW_TARGET_COUNT=$($MYSQL_CMD -N -s -e "SELECT COUNT(*) FROM $DATABASE.$TABLE;")
    if [[ $? -ne 0 ]]; then
        echo "警告: 导入后检查目标表记录数失败!"
    else
        echo "数据导入完成，目标表当前共有 $NEW_TARGET_COUNT 条记录"
    fi
fi

# 步骤6: 执行参数指定的后续SQL
if [[ -n "$POST_SQL" ]]; then
    echo "步骤6: 执行后续SQL"
    $MYSQL_CMD $DATABASE -e "$POST_SQL"
    if [[ $? -ne 0 ]]; then
        echo "错误: 执行后续SQL失败!"
        exit 1
    fi
else
    echo "步骤6: 跳过后续SQL执行（未提供后续SQL参数）"
fi

# 步骤7: 清理临时库
echo "步骤7: 删除临时库 $TEMP_DATABASE"
$MYSQL_CMD -e "DROP DATABASE $TEMP_DATABASE;"
if [[ $? -ne 0 ]]; then
    echo "警告: 删除临时库失败!"
fi

echo "所有步骤完成!"