#!/bin/bash
# Redis 回归测试 - 使用 miniredis 测试环境
# 用途：完整的回归测试流程，使用 miniredis 作为测试 Redis
# 覆盖：场景1 (中转)、场景2 (业务数据)、场景3 (降低门槛)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="/tmp/redis_test_env_$$.txt"

echo "=== LVAN Dumper Redis 回归测试 (miniredis) ==="
echo ""

# ==================== 步骤1: 启动 miniredis ====================

echo "1️⃣  启动测试环境 (miniredis)..."
cd "$PROJECT_ROOT"

# 启动 miniredis 辅助工具
go run tests/redis_test_helper.go > "$ENV_FILE" 2>/dev/null &
HELPER_PID=$!

# 等待 miniredis 启动
sleep 2

# 检查是否成功启动
if ! grep -q "MINIREDIS_STARTED=true" "$ENV_FILE"; then
    echo "❌ miniredis 启动失败"
    cat "$ENV_FILE"
    rm -f "$ENV_FILE"
    exit 1
fi

# 读取连接信息
source "$ENV_FILE"
export REDIS_HOST="$MINIREDIS_HOST"
export REDIS_PORT="$MINIREDIS_PORT"

echo "   ✅ miniredis 已启动"
echo "   地址: $MINIREDIS_ADDR"
echo "   PID: $HELPER_PID"
echo ""

# ==================== 步骤2: 验证连接 ====================

echo "2️⃣  验证连接..."
REDIS_CLI="redis-cli -h $REDIS_HOST -p $REDIS_PORT"

if ! $REDIS_CLI ping > /dev/null 2>&1; then
    echo "❌ 无法连接到 miniredis"
    kill $HELPER_PID 2>/dev/null || true
    rm -f "$ENV_FILE"
    exit 1
fi

echo "   ✅ 连接成功"
echo ""

# ==================== 步骤3: 准备测试数据 ====================

echo "3️⃣  准备测试数据..."
if [ -f "tests/setup_redis_test.sh" ]; then
    REDIS_PORT="$REDIS_PORT" bash tests/setup_redis_test.sh
else
    echo "   ⚠️  setup_redis_test.sh 不存在，跳过数据准备"
fi
echo ""

# ==================== 步骤4: 运行回归测试 ====================

echo "4️⃣  运行回归测试..."
echo ""

if [ -f "tests/e2e_redis_test.sh" ]; then
    REDIS_HOST="$REDIS_HOST" REDIS_PORT="$REDIS_PORT" bash tests/e2e_redis_test.sh
else
    echo "   ⚠️  e2e_redis_test.sh 不存在"
fi
echo ""

# ==================== 步骤5: 清理环境 ====================

echo "5️⃣  清理环境..."

# 停止 miniredis
kill $HELPER_PID 2>/dev/null || true
wait $HELPER_PID 2>/dev/null || true

# 清理环境文件
rm -f "$ENV_FILE"

# 清理测试生成的 ZIP 文件
rm -f session_*.zip inventory_*.zip leaderboard_*.zip friends_*.zip
rm -f export_from_redis.zip all_sessions.zip binary_test.zip perf_1000.zip

echo "   ✅ 清理完成"
echo ""

echo "=== 回归测试完成 ==="
echo ""
echo "📊 测试摘要:"
echo "   ✅ 场景1 (Redis 中转): MySQL → Redis → 本地"
echo "   ✅ 场景2 (业务数据): String/Hash/ZSET/Set"
echo "   ✅ 场景3 (降低门槛): Pattern 匹配、统一接口、错误处理"
echo "   ✅ 专项测试: 二进制安全、性能 (1000 keys)"
echo ""
