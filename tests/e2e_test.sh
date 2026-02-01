#!/bin/bash
# LVAN Dumper 端到端测试
# 验证完整的 dump → import 流程和数据一致性

set -e

DB_NAME="lvan_dumper_test"
DB_DIR="./tests/.dolt-data"
DOLT_PORT=3307

echo "=== LVAN Dumper 端到端测试 ==="
echo ""

# 检查 Dolt 是否安装
if ! command -v dolt &> /dev/null; then
    echo "❌ Dolt 未安装，请先安装 Dolt"
    echo "下载地址: https://github.com/dolthub/dolt/releases"
    exit 1
fi

# 1. 准备测试数据库
echo "1️⃣  准备测试数据库..."
if [ -f "tests/setup_test_db.sh" ]; then
    bash tests/setup_test_db.sh
else
    echo "❌ 测试数据库准备脚本不存在: tests/setup_test_db.sh"
    exit 1
fi

# 等待数据库启动
sleep 2

# 2. 验证测试数据
echo ""
echo "2️⃣  验证测试数据..."
cd "$DB_DIR"
RECORD_COUNT=$(dolt sql -q "SELECT COUNT(*) FROM user;" 2>/dev/null | tail -n 1)
echo "   测试数据记录数: $RECORD_COUNT"

if [ -z "$RECORD_COUNT" ] || [ "$RECORD_COUNT" -lt 4 ]; then
    echo "❌ 测试数据不足，至少需要 4 条记录"
    exit 1
fi
echo "   ✅ 测试数据准备完成"

# 3. 导出数据（ZIP 格式）
echo ""
echo "3️⃣  导出数据 (ZIP 格式)..."
cd - > /dev/null

go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P $DOLT_PORT -u root \
  -d $DB_NAME -t user \
  test_user_001 test_user_002 test_user_003 test_user_004

# 验证 ZIP 文件生成
ZIP_FILE="${DB_NAME}.user.zip"
if [ ! -f "$ZIP_FILE" ]; then
    echo "❌ 导出失败: ZIP 文件未生成"
    exit 1
fi

echo "   ✅ ZIP 文件已生成: $ZIP_FILE"

# 检查 ZIP 内容
echo ""
echo "   ZIP 内容结构:"
unzip -l "$ZIP_FILE"

# 4. 验证 BLOB 字段完整性
echo ""
echo "4️⃣  验证 BLOB 字段完整性..."

# 从数据库获取原始 HEX
cd "$DB_DIR"
ORIGINAL_HEX=$(dolt sql -q "SELECT HEX(data) FROM user WHERE accountid='test_user_001';" 2>/dev/null | tail -n 1)
cd - > /dev/null

if [ -z "$ORIGINAL_HEX" ]; then
    echo "   ⚠️  无法获取原始 BLOB 数据（可能为 NULL）"
else
    echo "   原始 BLOB HEX: $ORIGINAL_HEX"
fi

# 从 ZIP 中提取 BLOB 并验证
ZIP_BLOB=$(unzip -p "$ZIP_FILE" test_user_001/data 2>/dev/null | xxd -p -c 256)
if [ -n "$ZIP_BLOB" ]; then
    echo "   ZIP 中 BLOB HEX: $ZIP_BLOB"
    echo "   ✅ BLOB 字段已导出"
fi

# 5. 创建导入测试表
echo ""
echo "5️⃣  创建导入测试表..."
cd "$DB_DIR"
dolt sql "CREATE TABLE user_import LIKE user;"
dolt sql "CREATE TABLE user_verify LIKE user;"
echo "   ✅ 测试表已创建"

# 6. 导入数据
echo ""
echo "6️⃣  导入数据..."
cd - > /dev/null

go run lvan/cmd/dumper/main.go mysql import \
  -h 127.0.0.1 -P $DOLT_PORT -u root \
  -d $DB_NAME -t user_import \
  "./$ZIP_FILE"

echo "   ✅ 数据导入完成"

# 7. 数据一致性验证
echo ""
echo "7️⃣  数据一致性验证..."
cd "$DB_DIR"

# 验证记录数
IMPORT_COUNT=$(dolt sql -q "SELECT COUNT(*) FROM user_import;" 2>/dev/null | tail -n 1)
echo "   导入记录数: $IMPORT_COUNT"

if [ "$IMPORT_COUNT" -lt 4 ]; then
    echo "❌ 导入记录数不正确: 预期 >= 4, 实际 = $IMPORT_COUNT"
    exit 1
fi

# 验证各字段一致性
echo ""
echo "   字段一致性检查:"

dolt sql "
SELECT
    o.accountid,
    HEX(o.data) = HEX(i.data) as blob_match,
    o.int_val = i.int_val as int_match,
    o.float_val = i.float_val as float_match,
    o.username = i.username as text_match,
    o.ctime = i.ctime as time_match
FROM user o
LEFT JOIN user_import i ON o.accountid = i.accountid
WHERE o.accountid IN ('test_user_001', 'test_user_002', 'test_user_003', 'test_user_004')
ORDER BY o.accountid;
"

# 检查是否所有匹配都为 TRUE
MISMATCH_COUNT=$(dolt sql -q "
SELECT COUNT(*)
FROM user o
LEFT JOIN user_import i ON o.accountid = i.accountid
WHERE
    o.accountid IN ('test_user_001', 'test_user_002', 'test_user_003', 'test_user_004') AND
    (
        HEX(o.data) != HEX(i.data) OR
        o.int_val != i.int_val OR
        ABS(o.float_val - i.float_val) > 0.000001 OR
        o.username != i.username
    );
" 2>/dev/null | tail -n 1)

if [ "$MISMATCH_COUNT" -eq 0 ]; then
    echo "   ✅ 所有字段一致性验证通过"
else
    echo "❌ 发现 $MISMATCH_COUNT 个字段不一致"
    exit 1
fi

cd - > /dev/null

# 8. 测试目录格式导出
echo ""
echo "8️⃣  测试目录格式导出..."

go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P $DOLT_PORT -u root \
  -d $DB_NAME -t user \
  --in dir \
  test_user_001

DIR_EXPORT="${DB_NAME}.user"
if [ -d "$DIR_EXPORT" ]; then
    echo "   ✅ 目录格式导出成功"
    echo "   目录结构:"
    ls -la "$DIR_EXPORT/test_user_001/" | head -10
else
    echo "❌ 目录格式导出失败"
    exit 1
fi

# 9. 清理
echo ""
echo "9️⃣  清理测试文件..."

# 停止 Dolt SQL Server
pkill -f "dolt sql-server" 2>/dev/null || true

# 保留 ZIP 文件供手动检查
echo "   保留测试文件: $ZIP_FILE"
echo "   如需清理，请手动删除: $ZIP_FILE, $DIR_EXPORT, $DB_DIR"

# 10. 测试总结
echo ""
echo "=== 端到端测试完成 ==="
echo ""
echo "✅ 所有测试通过"
echo ""
echo "测试覆盖:"
echo "  ✅ Dolt 测试数据库准备"
echo "  ✅ MySQL Dump 命令 (ZIP 格式)"
echo "  ✅ MySQL Dump 命令 (目录格式)"
echo "  ✅ MySQL Import 命令"
echo "  ✅ BLOB 字段完整性"
echo "  ✅ 数据一致性验证"
echo ""
echo "生成的文件:"
echo "  📦 $ZIP_FILE"
echo "  📁 $DIR_EXPORT"
echo ""
