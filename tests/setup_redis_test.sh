#!/bin/bash
# Redis 测试数据准备脚本
# 用途：为场景1、2、3 准备 Redis 测试数据

set -e

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_CLI="redis-cli -h $REDIS_HOST -p $REDIS_PORT"

echo "=== 准备 Redis 测试数据 ==="

# 清空数据库
echo "1️⃣  清空现有数据..."
$REDIS_CLI FLUSHDB

# ==================== 场景2: 业务数据 ====================

echo "2️⃣  准备业务数据..."

# String 类型 - 玩家会话
echo "   📝 创建玩家会话数据..."
$REDIS_CLI SET "session:user:10001" '{"token":"abc123","login":1706789123,"ip":"192.168.1.100"}'
$REDIS_CLI SET "session:user:10002" '{"token":"def456","login":1706789999,"ip":"192.168.1.101"}'
$REDIS_CLI SET "session:user:10003" '{"token":"ghi789","login":1706799999,"ip":"192.168.1.102"}'

# String 类型 - 游戏配置
echo "   ⚙️  创建游戏配置..."
$REDIS_CLI SET "config:game:version" "1.2.3"
$REDIS_CLI SET "config:drop:rates" '{"common":0.8,"rare":0.19,"epic":0.01}'
$REDIS_CLI SET "config:maintenance" "false"
$REDIS_CLI SET "config:announcement" "Welcome!"

# String 类型 - 限流计数
echo "   🛡️  创建限流计数..."
$REDIS_CLI SET "ratelimit:api:user:10001" "15"
$REDIS_CLI SET "ratelimit:api:user:10002" "20"
$REDIS_CLI SET "ratelimit:chat:user:10001" "3"
$REDIS_CLI SET "ratelimit:chat:user:10002" "5"

# Hash 类型 - 玩家背包
echo "   🎒 创建玩家背包..."
$REDIS_CLI HSET "inventory:user:10001" gold "10000" gems "500" capacity "100"
$REDIS_CLI HSET "inventory:user:10001" item:1 "sword:1" item:2 "potion:5" item:3 "shield:1"
$REDIS_CLI HSET "inventory:user:10002" gold "5000" gems "200" capacity "50"
$REDIS_CLI HSET "inventory:user:10002" item:1 "axe:1" item:2 "armor:1"
$REDIS_CLI HSET "inventory:user:10003" gold "0" gems "0" capacity "20"  # 空背包

# Hash 类型 - 玩家属性
echo "   👤 创建玩家属性..."
$REDIS_CLI HSET "user:profile:10001" level "99" exp "12345" guild "test" vip "true"
$REDIS_CLI HSET "user:profile:10001" class "warrior" hp "5000" mp "2000"
$REDIS_CLI HSET "user:profile:10002" level "85" exp "5432" guild "demo" vip "false"
$REDIS_CLI HSET "user:profile:10002" class "mage" hp "2000" mp "5000"

# ZSET 类型 - 排行榜
echo "   🏆 创建排行榜..."
$REDIS_CLI ZADD "leaderboard:level" 99 "user:10001"
$REDIS_CLI ZADD "leaderboard:level" 85 "user:10002"
$REDIS_CLI ZADD "leaderboard:level" 72 "user:10003"
$REDIS_CLI ZADD "leaderboard:level" 65 "user:10004"
$REDIS_CLI ZADD "leaderboard:level" 60 "user:10005"
$REDIS_CLI ZADD "leaderboard:score" 999999 "user:10001"
$REDIS_CLI ZADD "leaderboard:score" 85432 "user:10002"
$REDIS_CLI ZADD "leaderboard:score" 12345 "user:10003"

# Set 类型 - 好友列表
echo "   👥 创建好友关系..."
$REDIS_CLI SADD "friends:user:10001" "user:10002" "user:10003" "user:10005"
$REDIS_CLI SADD "friends:user:10002" "user:10001" "user:10004" "user:10006"
$REDIS_CLI SADD "friends:user:10003" "user:10001" "user:10005" "user:10007"

# ==================== 场景1: 中转数据 ====================

echo "3️⃣  准备中转数据 (MySQL → Redis)..."

# 序列化的 MySQL 用户数据
$REDIS_CLI SET "export:mysql:user:10001" '{"uid":10001,"accountid":"test_user_001","username":"测试用户001","data":"08019601010A0774657374696E67"}'
$REDIS_CLI SET "export:mysql:user:10002" '{"uid":10002,"accountid":"test_user_002","username":"边界值测试","data":null}'
$REDIS_CLI SET "export:mysql:user:10003" '{"uid":10003,"accountid":"test_user_003","username":"NULL测试","data":null}'
$REDIS_CLI SET "export:mysql:user:10004" '{"uid":10004,"accountid":"test_user_004","username":"特殊字符测试","data":"0001020304"}'

# ==================== 场景3: 降低门槛 - 测试数据 ====================

echo "4️⃣  准备易用性测试数据..."

# 各种字符串类型
$REDIS_CLI SET "test:string:empty" ""
$REDIS_CLI SET "test:string:number" "12345"
$REDIS_CLI SET "test:string:float" "123.45"
$REDIS_CLI SET "test:string:chinese" "中文测试数据"
$REDIS_CLI SET "test:string:special" "包含\"引号\"和'单引号'"

# 带过期时间的数据
echo "   ⏰ 创建带 TTL 的数据..."
$REDIS_CLI SET "cache:user:10001:profile" '{"level":99,"exp":12345}' EX 3600
$REDIS_CLI SET "cache:user:10002:profile" '{"level":85,"exp":5432}' EX 7200
$REDIS_CLI SET "cache:config:drop" '{"common":0.8}' EX 1800

# 二进制安全测试数据
echo "   🔢 创建二进制测试数据..."
$REDIS_CLI SET "test:binary:null" "$(printf 'abc\x00def\x00ghi')"
$REDIS_CLI SET "test:protobuf:user:10001" "$(printf '\x08\x96\x01\x01\x0A\x07testing')"

# 大 Value 测试
echo "   📦 创建大 Value 测试..."
# 创建 1KB 数据
large_data=$(python3 -c "print('A' * 1024)")
$REDIS_CLI SET "test:large:1kb" "$large_data"

# 创建 10KB 数据
large_data_10k=$(python3 -c "print('B' * 10240)")
$REDIS_CLI SET "test:large:10kb" "$large_data_10k"

# ==================== 验证 ====================

echo ""
echo "5️⃣  验证测试数据..."

# 统计 Key 数量
string_count=$($REDIS_CLI --no-auth-warning TYPE "session:user:10001" | grep -c "string" || echo 0)
hash_count=$($REDIS_CLI --no-auth-warning HLEN "inventory:user:10001")
zset_count=$($REDIS_CLI --no-auth-warning ZCARD "leaderboard:level")
set_count=$($REDIS_CLI --no-auth-warning SCARD "friends:user:10001")

echo "   String Keys: 4+ (会话)"
echo "   Hash Keys: $hash_count fields (背包)"
echo "   ZSET Keys: $zset_count members (排行榜)"
echo "   Set Keys: $set_count members (好友)"
echo "   Export Keys: 4 (中转数据)"
echo "   Cache Keys: 3 (带 TTL)"

total_keys=$($REDIS_CLI --no-auth-warning DBSIZE)
echo "   总 Keys: $total_keys"

echo ""
echo "✅ Redis 测试数据准备完成"
echo ""
echo "=== 数据预览 ==="
echo "📝 会话数据:"
$REDIS_CLI GET "session:user:10001"
echo ""
echo "🎒 背包数据:"
$REDIS_CLI HGETALL "inventory:user:10001"
echo ""
echo "🏆 排行榜 (前3):"
$REDIS_CLI ZRANGE "leaderboard:level" 0 2 WITHSCORES
echo ""
echo "👥 好友列表:"
$REDIS_CLI SMEMBERS "friends:user:10001"
