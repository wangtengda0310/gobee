#!/bin/bash

# MySQL 表数据导出脚本
# 用法: ./dump.sh [选项] [ID列表]

# 默认参数
HOST="localhost"
USER="root"
PASSWORD=""
DATABASE="gforge"
TABLE="user"
WHERE_COLUMN="uid"
IDS=()
SIMULATE=false

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --host HOST          MySQL 主机名 (默认: localhost)"
    echo "  -u, --user USER          MySQL 用户名 (默认: root)"
    echo "  -p, --password PASSWORD  MySQL 密码 (空密码时可不提供，支持MySQL风格的stdin输入方式)"
    echo "  -d, --database DATABASE  数据库名 (默认: gforge)"
    echo "  -t, --table TABLE        表名 (默认: user)"
    echo "  -w, --where column       列名"
    echo "  --simulate               模拟模式，不实际连接数据库"
    echo "  --help                   显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  导出account数据: $0 -h localhost -u root -p password -d gforge -t account 1 2 3"
    echo "  导出user数据: $0 -h localhost -u root -p password -d gforge -t user 1 2 3"
    echo "  指定条件导出user数据: $0 -h localhost -u root -p password -d gforge -t user -w uid 1 2 3"
    echo "  默认导出user数据: $0 -h localhost -u root -p password 1 2 3"
    echo "  MySQL风格的stdin输入密码: $0 -h localhost -u root -p -d gforge -t user 1 2 3"
    echo "  模拟模式测试: $0 --simulate -d testdb -t users 1 2 3"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    # 检查是否是选项参数
    if [[ $1 == -* ]]; then
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
                    # 没有密码值参数，使用stdin输入密码
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
            -w|--where)
                WHERE_COLUMN="$2"
                shift
                shift
                ;;
            --simulate)
                SIMULATE=true
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
    else
        # 不是选项参数，作为ID处理
        IDS+=("$1")
        shift
    fi
done

# 验证必需参数
# HOST和USER有默认值，DATABASE和TABLE也有默认值，不需要强制指定
# 只有在非模拟模式下才需要确保mysqldump/mariadb-dump命令可用

# 验证ID列表（如果提供）
if [[ ${#IDS[@]} -gt 0 ]]; then
    for id in "${IDS[@]}"; do
        if ! [[ "$id" =~ ^[0-9]+$ ]]; then
            echo "错误: ID必须是数字，无效ID: $id"
            exit 1
        fi
    done
fi

# 模拟模式下直接生成测试文件
if [[ "$SIMULATE" == true ]]; then
    # 如果有ID列表，构建WHERE条件
    if [[ ${#IDS[@]} -gt 0 ]]; then
        # 构建ID列表字符串，用逗号分隔
        ID_LIST=$(IFS=,; echo "${IDS[*]}")
        # 设置输出文件名，包含ID列表信息
        OUTPUT_FILE="${TABLE}_${WHERE_COLUMN}_${ID_LIST}.sql"
    else
        # 如果没有ID列表，导出整个表
        OUTPUT_FILE="${TABLE}_all.sql"
    fi
    
    echo "模拟模式: 生成测试文件 $OUTPUT_FILE"
    echo "-- 模拟的数据库导出文件" > "$OUTPUT_FILE"
    echo "-- 数据库: $DATABASE" >> "$OUTPUT_FILE"
    echo "-- 表: $TABLE" >> "$OUTPUT_FILE"
    if [[ ${#IDS[@]} -gt 0 ]]; then
        echo "-- 条件: $WHERE_COLUMN IN ($(IFS=,; echo "${IDS[*]}"))" >> "$OUTPUT_FILE"
    fi
    echo "CREATE TABLE $TABLE (id INT, name VARCHAR(255));" >> "$OUTPUT_FILE"
    echo "INSERT INTO $TABLE VALUES (1, 'test');" >> "$OUTPUT_FILE"
    echo "成功生成模拟文件: $OUTPUT_FILE"
    exit 0
fi

# 构建导出命令
if command -v mariadb-dump &> /dev/null; then
    DUMP_CMD="mariadb-dump"
elif command -v mysqldump &> /dev/null; then
    DUMP_CMD="mysqldump"
else
    echo "错误: 未找到 mariadb-dump 或 mysqldump 命令!"
    exit 1
fi

# 构建基本导出命令
DUMP_CMD="$DUMP_CMD -h $HOST -u $USER"
if [[ -n "$PASSWORD" ]]; then
    DUMP_CMD="$DUMP_CMD -p$PASSWORD"
else
    DUMP_CMD="$DUMP_CMD -p"
fi

# 添加数据库和表名
DUMP_CMD="$DUMP_CMD $DATABASE $TABLE"

# 如果有ID列表，构建WHERE条件
if [[ ${#IDS[@]} -gt 0 ]]; then
    # 构建ID列表字符串，用逗号分隔
    ID_LIST=$(IFS=,; echo "${IDS[*]}")
    WHERE_CONDITION="$WHERE_COLUMN IN ($ID_LIST)"
    
    # 设置输出文件名，包含ID列表信息
    OUTPUT_FILE="${TABLE}_${WHERE_COLUMN}_${ID_LIST}.sql"
else
    # 如果没有ID列表，导出整个表
    OUTPUT_FILE="${TABLE}_all.sql"
    WHERE_CONDITION=""
fi

# 显示执行的命令（不包含密码）
DUMP_CMD_DISPLAY="$DUMP_CMD"
if [[ -n "$PASSWORD" ]]; then
    DUMP_CMD_DISPLAY="$DUMP_CMD_DISPLAY -p***"
fi
if [[ -n "$WHERE_CONDITION" ]]; then
    DUMP_CMD_DISPLAY="$DUMP_CMD_DISPLAY --where=\"$WHERE_CONDITION\""
fi
DUMP_CMD_DISPLAY="$DUMP_CMD_DISPLAY > $OUTPUT_FILE"

echo "执行导出命令: $DUMP_CMD_DISPLAY"

# 执行导出命令（不使用eval）
if [[ -n "$WHERE_CONDITION" ]]; then
    $DUMP_CMD --where="$WHERE_CONDITION" > "$OUTPUT_FILE"
else
    $DUMP_CMD > "$OUTPUT_FILE"
fi

# 检查导出结果
if [[ $? -eq 0 ]]; then
    echo "成功导出数据到文件: $OUTPUT_FILE"
else
    echo "导出数据失败!"
    exit 1
fi