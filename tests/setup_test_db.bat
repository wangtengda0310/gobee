@echo off
REM LVAN Dumper 测试数据库准备脚本 (Windows)
REM 使用 Dolt 创建轻量级测试数据库

setlocal enabledelayedexpansion

set DB_NAME=lvan_dumper_test
set DB_DIR=.\tests\.dolt-data
set DOLT_PORT=3307

echo === LVAN Dumper 测试数据库准备 ===
echo 使用 Dolt 创建测试数据库...

REM 清理旧的测试数据库
if exist "%DB_DIR%" (
    echo 清理旧的测试数据...
    rmdir /s /q "%DB_DIR%"
)

REM 创建新的 Dolt 数据库
mkdir "%DB_DIR%"
cd /d "%DB_DIR%"

REM 初始化 Dolt 数据库
dolt init

REM 创建测试用户表（覆盖所有 MySQL 数据类型）
dolt sql <<EOF
CREATE TABLE user (
    uid INT PRIMARY KEY AUTO_INCREMENT,
    accountid VARCHAR(50) NOT NULL UNIQUE,
    username VARCHAR(100),
    data BLOB,
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
    text_val TEXT,
    blob_val BLOB,
    json_val JSON,
    ctime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    mtime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
EOF

echo ✅ 表结构创建完成

REM 插入测试数据
dolt sql --batch <<EOF
INSERT INTO user (accountid, username, data, int_val, float_val, bool_val, date_val, blob_val, text_val) VALUES
('test_user_001', '测试用户001', UNHEX('08019601010A0774657374696E67'), 2147483647, 3.14, TRUE, '2025-01-31', UNHEX('E8BDA6E69687E6B58BE8AF95'), '测试文本'),
('test_user_002', '边界值测试', NULL, -2147483648, -3.14, FALSE, '1970-01-01', NULL, '边界值'),
('test_user_003', 'NULL值测试', NULL, NULL, NULL, NULL, NULL, NULL, NULL),
('test_user_004', '特殊字符测试', UNHEX('000102'), 0, 0.0, TRUE, '2025-01-31', BINARY('data'), '包含"引号"和''单引号''');
EOF

echo ✅ 测试数据插入完成

REM 验证数据
echo.
echo === 验证测试数据 ===
dolt sql "SELECT COUNT(*) as total FROM user;"
dolt sql "SELECT uid, accountid, username FROM user;"

REM 启动 Dolt SQL Server
echo.
echo === 启动 Dolt SQL Server ===
echo 端口: %DOLT_PORT%

start /B dolt sql-server --port=%DOLT_PORT% --user=root --password="" --host=0.0.0.0

REM 等待服务启动
echo 等待 Dolt SQL Server 启动...
timeout /t 3 /nobreak

echo.
echo === 测试数据库准备完成 ===
echo 数据库目录: %DB_DIR%
echo.
echo 测试命令:
echo   go run lvan/cmd/dumper/main.go mysql dump -h 127.0.0.1 -P %DOLT_PORT% -u root -d %DB_NAME% -t user test_user_001 test_user_002
echo.

endlocal
