#!/bin/bash
# LVAN Dumper 测试数据库准备脚本
# 使用 Dolt 创建轻量级测试数据库，覆盖所有 MySQL 数据类型
# 重点测试 BLOB 字段（protobuf 数据）

set -e

DB_NAME="lvan_dumper_test"
DB_DIR="./tests/.dolt-data"

echo "=== LVAN Dumper 测试数据库准备 ==="
echo "使用 Dolt 创建测试数据库..."

# 清理旧的测试数据库
if [ -d "$DB_DIR" ]; then
    echo "清理旧的测试数据..."
    rm -rf "$DB_DIR"
fi

# 创建新的 Dolt 数据库
mkdir -p "$DB_DIR"
cd "$DB_DIR"

# 初始化 Dolt 数据库
dolt init

# 创建测试用户表（简化版本，专注核心测试字段）
dolt sql <<'EOF'
-- 创建测试用户表
CREATE TABLE user (
    uid INT PRIMARY KEY AUTO_INCREMENT,
    accountid VARCHAR(50) NOT NULL UNIQUE,
    username VARCHAR(100),
    data BLOB,                    -- 🔑 BLOB 字段：存储 protobuf 数据
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
    timestamp_val TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    year_val YEAR,
    char_val CHAR(10),
    varchar_val VARCHAR(255),
    blob_val BLOB,                -- 🔑 BLOB 字段：二进制数据
    text_val TEXT,
    json_val JSON,
    enum_val ENUM('A', 'B', 'C'),
    set_val SET('X', 'Y', 'Z'),
    ctime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    mtime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
EOF

echo "✅ 表结构创建完成"

# 生成测试 protobuf 数据（二进制）
# 模拟真实场景的 protobuf 序列化数据
# 直接使用硬编码的十六进制值，避免 Python 兼容性问题
# Protobuf 数据: 089601010A0774657374696E67 (field 1: 150, field 2: 1, field 3: "testing")
PROTOBUF_DATA="089601010A0774657374696E67"

# 插入测试数据（简化版本，专注核心测试字段）
dolt sql --batch <<EOF
-- 插入测试记录 1：完整数据（包含 BLOB 和 TIMESTAMP）
INSERT INTO user (accountid, username, data, blob_val, text_val, enum_val, set_val)
VALUES (
    'test_user_001', '测试用户001',
    UNHEX('$PROTOBUF_DATA'),
    UNHEX('E8BDA6E69687E6B58BE8AF95'),
    '测试文本内容',
    'B', 'X,Y'
);

-- 插入测试记录 2：边界值测试（包含 TIMESTAMP）
INSERT INTO user (accountid, username, data, blob_val, text_val, enum_val, set_val)
VALUES (
    'test_user_002', '边界值测试',
    NULL, NULL, '', 'A', 'X'
);

-- 插入测试记录 3：NULL 值测试
INSERT INTO user (accountid, username) VALUES
('test_user_003', 'NULL值测试');

-- 插入测试记录 4：特殊字符测试
INSERT INTO user (accountid, username, data, blob_val, text_val)
VALUES (
    'test_user_004', '特殊字符测试',
    UNHEX('000102030405'),
    UNHEX('E8BDA6E69687E6B58BE8AF95'),
    '包含"引号"和''单引号'''
);

-- 插入测试记录 5：大 BLOB 测试（~200 字节）
INSERT INTO user (accountid, username, blob_val)
VALUES ('test_user_005', '大BLOB测试', UNHEX('$(printf "ABCD%.0s" {1..50})'));
EOF

echo "✅ 测试数据插入完成"

# 验证数据
echo ""
echo "=== 验证测试数据 ==="
dolt sql -q "SELECT COUNT(*) as total FROM user;"
dolt sql -q "SELECT uid, accountid, username, HEX(data) as data_hex, blob_val IS NOT NULL as has_blob FROM user;"

# 启动 Dolt SQL Server（后台运行）
DOLT_PORT=3307
echo ""
echo "=== 启动 Dolt SQL Server ==="
echo "端口: $DOLT_PORT"

# 停止可能存在的旧进程
pkill -f "dolt sql-server" || true

# 启动 Dolt SQL Server
dolt sql-server --port=$DOLT_PORT --host=0.0.0.0 --loglevel=info &
DOLT_PID=$!

# 等待服务启动
echo "等待 Dolt SQL Server 启动..."
sleep 3

# 测试连接
echo "测试连接..."
mysql -h 127.0.0.1 -P $DOLT_PORT -u root -e "SELECT VERSION();" 2>/dev/null || {
    echo "⚠️  警告: 无法连接到 Dolt SQL Server"
    echo "请检查 Dolt 是否正确安装"
    kill $DOLT_PID 2>/dev/null || true
    exit 1
}

echo ""
echo "=== 测试数据库准备完成 ==="
echo "数据库目录: $DB_DIR"
echo "连接参数:"
echo "  Host: 127.0.0.1"
echo "  Port: $DOLT_PORT"
echo "  User: root"
echo "  Password: (空)"
echo "  Database: $DB_NAME"
echo ""
echo "Dolt SQL Server PID: $DOLT_PID"
echo ""
echo "测试命令:"
echo "  go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P $DOLT_PORT -u root -d $DB_NAME -t user test_user_001 test_user_002"
echo ""
echo "停止服务器: kill $DOLT_PID"
