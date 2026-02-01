# Redis 支持回归测试文档

## 概述

本文档定义 Redis 数据源支持的回归测试用例，覆盖以下核心场景：
- **场景1**: Redis 作为数据克隆中转
- **场景2**: Redis 中实现的业务数据
- **场景3**: 非 Redis 运维人员的操作

---

## 一、测试环境准备

### 1.1 测试 Redis 部署

**推荐方案**: 使用 Docker 快速部署测试 Redis

```bash
# 启动单实例 Redis
docker run -d --name redis-test \
  -p 6379:6379 \
  redis:7-alpine

# 或使用本地 Redis
redis-server --port 6379
```

**验证连接**:
```bash
redis-cli ping
# 预期输出: PONG
```

### 1.2 使用 miniredis 作为测试环境（推荐）

#### 为什么选择 miniredis？

**优势对比**:

| 特性 | Docker Redis | miniredis |
|------|-------------|-----------|
| 启动速度 | 🟡 中等 (2-5s) | ✅ 极快 (<100ms) |
| 外部依赖 | 🔴 Docker | ✅ 无依赖 |
| 测试隔离性 | ⚠️ 需要清理 | ✅ 自动隔离 |
| Redis 行为 | ✅ 完整 | ✅ 兼容 |
| CI/CD 集成 | ⚠️ 需要容器环境 | ✅ 开箱即用 |
| 并发测试 | ⚠️ 端口冲突 | ✅ 完美支持 |

**推荐**: 回归测试使用 miniredis（快速、无依赖），生产验证使用 Docker Redis

#### 安装 miniredis

```bash
# 安装 miniredis v2
go get github.com/alicebob/miniredis/v2

# 或
go mod tidy
```

#### 创建测试辅助脚本

**脚本位置**: `tests/redis_test_helper.go`

```go
// +build ignore

package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/alicebob/miniredis/v2"
)

// Redis 测试辅助工具：启动 miniredis 并输出连接信息
func main() {
    s := miniredis.RunMiniServer(nil)

    fmt.Printf("MINIREDIS_STARTED=true\n")
    fmt.Printf("MINIREDIS_HOST=%s\n", "127.0.0.1")
    fmt.Printf("MINIREDIS_PORT=%d\n", s.Port())
    fmt.Printf("MINIREDIS_ADDR=%s\n", s.Addr())

    // 等待终止信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    s.Close()
    fmt.Println("\nMINIREDIS_STOPPED=true")
}
```

**使用方式**:
```bash
# 1. 构建辅助工具
go build -o bin/redis_test_helper tests/redis_test_helper.go

# 2. 后台启动 miniredis
bin/redis_test_helper > redis_test_env.txt &
HELPER_PID=$!

# 3. 读取连接信息
source redis_test_env.txt
export REDIS_HOST=$MINIREDIS_HOST
export REDIS_PORT=$MINIREDIS_PORT

# 4. 运行回归测试
bash tests/e2e_redis_test.sh

# 5. 清理
kill $HELPER_PID
rm redis_test_env.txt
```

#### 回归测试流程

**测试脚本**: `tests/e2e_redis_with_miniredis.sh`

```bash
#!/bin/bash
# 使用 miniredis 的 Redis 回归测试

set -e

echo "=== Redis 回归测试 (miniredis) ==="

# 1. 启动 miniredis
echo "1️⃣  启动测试环境..."
go run tests/redis_test_helper.go > /tmp/redis_env.txt &
HELPER_PID=$!

# 等待 miniredis 启动
sleep 1

# 读取连接信息
source /tmp/redis_env.txt
export REDIS_PORT=$MINIREDIS_PORT

# 2. 验证连接
echo "2️⃣  验证连接..."
redis-cli -p $REDIS_PORT ping
# 预期输出: PONG

# 3. 准备测试数据
echo "3️⃣  准备测试数据..."
bash tests/setup_redis_test.sh $REDIS_PORT

# 4. 运行回归测试
echo "4️⃣  运行回归测试..."
REDIS_PORT=$REDIS_PORT bash tests/e2e_redis_test.sh

# 5. 清理环境
echo "5️⃣  清理环境..."
kill $HELPER_PID
rm /tmp/redis_env.txt

echo "=== 回归测试完成 ==="
```

#### 环境准备流程图

```
┌─────────────────────────────────────────────────────────┐
│  回归测试开始                                             │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤1: 启动 miniredis                                    │
│  ────────────────────────────────                        │
│  go run tests/redis_test_helper.go > env.txt &           │
│  输出: MINIREDIS_PORT=12345                              │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤2: 准备测试数据                                      │
│  ────────────────────────────────                        │
│  REDIS_PORT=12345 redis-cli SET "session:user:10001" ... │
│  REDIS_PORT=12345 redis-cli HSET "inventory:user:10001"  │
│  ... 准备所有测试数据                                     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤3: 运行 dump 命令                                    │
│  ────────────────────────────────                        │
│  go run main.go redis dump \                             │
│    -h 127.0.0.1 -p 12345 --pattern "session:*" \        │
│    -o sessions.zip                                       │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤4: 验证导出结果                                      │
│  ────────────────────────────────                        │
│  unzip -l sessions.zip                                    │
│  检查: session:user:10001/value 文件存在                  │
│  验证: 内容与 redis-cli GET 结果一致                      │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤5: 运行 import 命令                                  │
│  ────────────────────────────────                        │
│  redis-cli FLUSHDB (清空 DB0)                            │
│  go run main.go redis import sessions.zip                │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤6: 验证导入结果                                      │
│  ────────────────────────────────                        │
│  redis-cli GET "session:user:10001"                      │
│  对比: 导入值 = 原始值                                    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  步骤7: 清理环境                                          │
│  ────────────────────────────────                        │
│  kill $HELPER_PID (停止 miniredis)                       │
│  rm sessions.zip (清理测试文件)                          │
└─────────────────────────────────────────────────────────┘
```

#### 与 Docker Redis 的切换

**使用 Docker Redis** (生产验证):
```bash
# 启动 Docker Redis
docker run -d --name redis-test -p 6379:6379 redis:7-alpine

# 运行测试（使用默认端口 6379）
bash tests/e2e_redis_test.sh

# 清理
docker stop redis-test && docker rm redis-test
```

**使用 miniredis** (开发/CI):
```bash
# 启动 miniredis
go run tests/redis_test_helper.go > env.txt &
HELPER_PID=$!
source env.txt

# 运行测试
REDIS_PORT=$MINIREDIS_PORT bash tests/e2e_redis_test.sh

# 清理
kill $HELPER_PID
```

---

### 1.3 测试数据准备

```bash
#!/bin/bash
# 准备 Redis 测试数据

redis-cli FLUSHDB

# 场景2: 业务数据 - String 类型
redis-cli SET "session:user:10001" '{"token":"abc123","login":1706789123}'
redis-cli SET "session:user:10002" '{"token":"def456","login":1706789999}'
redis-cli SET "config:game:version" "1.2.3"
redis-cli SET "config:drop:rates" '{"common":0.8,"rare":0.19}'
redis-cli SET "ratelimit:api:user:10001" "15"
redis-cli SET "ratelimit:chat:user:10001" "3"

# 场景2: 业务数据 - Hash 类型
redis-cli HSET "inventory:user:10001" gold "10000" gems "500" item:1 "sword:1"
redis-cli HSET "inventory:user:10001" item:2 "potion:5" capacity "100"
redis-cli HSET "inventory:user:10002" gold "5000" gems "200"
redis-cli HSET "user:profile:10001" level "99" exp "12345" guild "test"
redis-cli HSET "user:profile:10002" level "85" exp "5432" guild "demo"

# 场景2: 业务数据 - ZSET 类型 (排行榜)
redis-cli ZADD "leaderboard:level" 99 "user:10001"
redis-cli ZADD "leaderboard:level" 85 "user:10002"
redis-cli ZADD "leaderboard:level" 72 "user:10003"
redis-cli ZADD "leaderboard:level" 65 "user:10004"
redis-cli ZADD "leaderboard:level" 60 "user:10005"

# 场景2: 业务数据 - Set 类型 (好友列表)
redis-cli SADD "friends:user:10001" "user:10002" "user:10003" "user:10005"
redis-cli SADD "friends:user:10002" "user:10001" "user:10004"

# 场景1: 中转数据 - 序列化的 MySQL 数据
redis-cli SET "export:mysql:user:10001" '{"uid":10001,"accountid":"test_user_001","username":"测试用户001"}'
redis-cli SET "export:mysql:user:10002" '{"uid":10002,"accountid":"test_user_002","username":"边界值测试"}'

# 带过期时间的数据
redis-cli SET "cache:user:10001:profile" '{"level":99}' EX 3600
redis-cli SET "session:user:10003" '{"token":"xyz789"}' EX 7200

echo "✅ Redis 测试数据准备完成"
```

---

## 二、场景2测试：Redis业务数据克隆

### 2.1 String 类型测试

#### 测试用例组 RS: Redis String Dump

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| RS01 | `connect redis dump --pattern "session:user:10001" -o session.zip` | 生成 ZIP | 检查文件存在 |
| RS02 | 同上，多 key | ZIP 包含多个目录 | 解压验证 |
| RS03 | `--pattern "session:*"` | 导出所有会话 | 检查记录数 |
| RS04 | `--pattern "config:*"` | 导出配置 | 验证内容 |

**执行示例**:
```bash
# 导出单个 String
connect redis dump \
  -h 127.0.0.1 -p 6379 \
  --pattern "session:user:10001" \
  -o session_10001.zip

# 验证 ZIP 内容
unzip -l session_10001.zip
# 预期: session:user:10001/value

# 验证数据内容
unzip -p session_10001.zip "session:user:10001/value"
# 预期: {"token":"abc123","login":1706789123}
```

#### 测试用例组 RSI: Redis String Import

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| RSI01 | `connect redis import session.zip` | 数据导入成功 | redis-cli GET 验证 |
| RSI02 | 覆盖已存在的 key | 根据策略处理 | 检查最终值 |
| RSI03 | 导入到不同 DB | 使用目标 DB | SELECT 验证 |

**执行示例**:
```bash
# 导入到不同数据库
connect redis import \
  -h 127.0.0.1 -p 6379 \
  --db 1 \
  session_10001.zip

# 验证导入结果
redis-cli -n 1 GET "session:user:10001"
# 预期: {"token":"abc123","login":1706789123}
```

### 2.2 Hash 类型测试

#### 测试用例组 RH: Redis Hash Dump

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| RH01 | `connect redis dump --pattern "inventory:user:10001"` | 导出 Hash | 检查 ZIP 结构 |
| RH02 | 同上 | 所有字段都在 ZIP 中 | 解压验证字段数 |
| RH03 | 导出多个 Hash | 包含多个目录 | 计数验证 |

**ZIP 结构验证**:
```
inventory:user:10001.zip
└── inventory:user:10001/
    ├── gold       # 内容: "10000"
    ├── gems       # 内容: "500"
    ├── item:1     # 内容: "sword:1"
    ├── item:2     # 内容: "potion:5"
    ├── capacity   # 内容: "100"
    └── ...
```

**执行示例**:
```bash
# 导出背包数据
connect redis dump \
  -h 127.0.0.1 -p 6379 \
  --pattern "inventory:user:10001" \
  -o inventory_10001.zip

# 验证所有字段
unzip -l inventory_10001.zip
```

### 2.3 ZSET 类型测试

#### 测试用例组 RZ: Redis ZSET Dump

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| RZ01 | `connect redis dump --pattern "leaderboard:level"` | 导出 ZSET | 检查结构 |
| RZ02 | 保持分数排序 | 按 score 排序 | 验证顺序 |
| RZ03 | 保持 member-score 对 | 成对存储 | 验证数据 |

**数据格式**:
```
leaderboard:level.zip
└── leaderboard:level/
    ├── 000001_user:10001    # score: 99
    ├── 000002_user:10002    # score: 85
    ├── 000003_user:10003    # score: 72
    └── ...
```

### 2.4 Set 类型测试

#### 测试用例组 RS: Redis Set Dump

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| RS01 | `connect redis dump --pattern "friends:user:10001"` | 导出 Set | 检查结构 |
| RS02 | 所有 member 都导出 | 计数正确 | SCARD 验证 |

### 2.5 混合类型测试

#### 测试用例组 RMIX: 混合类型 Dump

**场景**: 一次导出多种数据类型

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| RMIX01 | `connect redis dump --pattern "user:10001:*"` | 导出所有匹配 Key | 类型自动识别 |
| RMIX02 | 导出结果分类 | 按类型组织目录 | 目录结构验证 |
| RMIX03 | 保持数据完整性 | String/Hash/ZSET 都正确 | 逐类验证 |

---

## 三、场景1测试：Redis作为中转

### 3.1 MySQL → Redis 中转

#### 测试用例组 MR1: MySQL 导出到 Redis

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| MR101 | `connect mysql dump --output-redis redis:6379 --key-prefix "export:"` | 数据写入 Redis | redis-cli EXISTS 验证 |
| MR102 | 指定 TTL | 设置过期时间 | redis-cli TTL 验证 |
| MR103 | 批量导出 | 多个记录写入 | redis-cli KEYS 计数 |

**执行示例**:
```bash
# 导出 MySQL 用户数据到 Redis
connect mysql dump \
  -h mysql.production.com \
  -d gforge -t user \
  --output-format redis \
  --redis-host cache.internal \
  --redis-port 6379 \
  --redis-key-prefix "export:user:" \
  --redis-ttl 3600 \
  user10001 user10002

# 验证 Redis 中的数据
redis-cli -h cache.internal KEYS "export:user:*"
# 预期: export:user:10001, export:user:10002

redis-cli -h cache.internal GET "export:user:10001"
# 预期: JSON 格式的用户数据
```

#### 测试用例组 MR2: Redis 导出到本地 MySQL

| 用例ID | 命令 | 预期结果 | 验证方法 |
|--------|------|---------|---------|
| MR201 | `connect redis dump --pattern "export:user:*" -o export.zip` | 生成 ZIP | 文件检查 |
| MR202 | `connect mysql import export.zip` | 导入成功 | MySQL 查询验证 |
| MR203 | 数据一致性 | MySQL = 原始数据 | 对比验证 |

**完整流程**:
```bash
# 步骤1: 生产 → Redis (生产环境执行)
connect mysql dump \
  --host production.mysql \
  --database gforge --table user \
  --output-format redis \
  --redis-host cache.internal \
  --redis-key-prefix "export:user:" \
  10001 10002

# 步骤2: Redis → 本地 (本地执行)
connect redis dump \
  --host cache.internal \
  --pattern "export:user:*" \
  -o user_export.zip

# 步骤3: 本地 MySQL 导入
connect mysql import \
  --host 127.0.0.1 --database gforge_test \
  user_export.zip

# 步骤4: 验证数据一致性
dolt sql "SELECT * FROM user WHERE uid IN (10001, 10002);"
```

### 3.2 数据完整性验证

#### 测试用例组 V1: 中转完整性

| 用例ID | 场景 | 验证方法 | 预期结果 |
|--------|------|---------|---------|
| V101 | String 数据 | 值完全相等 | JSON 对比 |
| V102 | Hash 字段数量 | HLEN 相等 | 计数对比 |
| V103 | ZSET 分数 | score 精度 | 误差 < 1e-6 |
| V104 | Set member 数量 | SCARD 相等 | 计数对比 |
| V105 | 二进制安全 | HEX 相等 | HEX 对比 |

---

## 四、场景3测试：降低使用门槛

### 4.1 统一接口测试

#### 测试用例组 UX1: 与 MySQL 命令一致性

| 用例ID | MySQL 命令 | Redis 等效命令 | 验证点 |
|--------|-----------|---------------|--------|
| UX101 | `mysql dump -t user` | `redis dump --pattern "*"` | 参数结构一致 |
| UX102 | `mysql import data.zip` | `redis import data.zip` | 操作流程一致 |
| UX103 | `--help` 输出 | 提供清晰示例 | 文档友好 |

**接口对比**:
```bash
# MySQL 操作
connect mysql dump -h localhost -d db -t user key1 key2
connect mysql import -h localhost -d db -t user data.zip

# Redis 操作 (相同的命令结构)
connect redis dump -h localhost --pattern "session:*" -o session.zip
connect redis import -h localhost session.zip
```

### 4.2 Pattern 匹配测试

#### 测试用例组 UX2: 智能匹配

| 用例ID | 命令 | 预期匹配 | 验证方法 |
|--------|------|---------|---------|
| UX201 | `--pattern "session:*"` | 所有 session key | KEYS 验证 |
| UX202 | `--pattern "user:*"` | 匹配多种类型 | 类型自动识别 |
| UX203 | `--pattern "*:10001"` | 匹配所有后缀 | 覆盖验证 |
| UX204 | 无匹配键 | 友好错误提示 | 错误信息清晰 |

**对比传统方式**:
```bash
# 传统 Redis 方式 (需要专业知识)
redis-cli KEYS "session:*" | while read key; do
    redis-cli GET "$key"
    redis-cli TTL "$key"
done

# LVAN Dumper 方式 (统一接口)
connect redis dump --pattern "session:*" -o sessions.zip
```

### 4.3 错误处理测试

#### 测试用例组 UX3: 友好错误提示

| 用例ID | 场景 | 预期行为 | 验证方法 |
|--------|------|---------|---------|
| UX301 | Redis 连接失败 | 清晰的错误信息 | 提示主机/端口 |
| UX302 | Pattern 无匹配 | "未找到匹配的 Key" | 不崩溃 |
| UX303 | 权限不足 | "需要 AUTH" | 提示认证 |
| UX304 | 内存不足 | "Redis 内存警告" | 优雅处理 |

---

## 五、数据类型专项测试

### 5.1 二进制安全测试

#### 测试用例组 BIN: 二进制数据

| 用例ID | 场景 | 数据 | 验证方法 |
|--------|------|------|---------|
| BIN01 | Protobuf 数据 | 任意字节 | HEX 对比 |
| BIN02 | 包含 NULL 字节 | \x00 | 字节级对比 |
| BIN03 | UTF-8 编码 | 中文文本 | 编码保持 |
| BIN04 | 大 Value (1MB) | 大块数据 | 完整性验证 |

**测试数据准备**:
```bash
# 生成包含 NULL 字节的测试数据
redis-cli SET "binary:test:001" "$(printf 'abc\x00def\x00ghi')"

# 生成中文测试数据
redis-cli SET "chinese:test" "测试中文数据"

# 导出并验证
connect redis dump --pattern "binary:test:*" -o binary_test.zip
unzip -p binary_test.zip "binary:test:001/value" | xxd
```

### 5.2 过期时间测试

#### 测试用例组 TTL: 过期时间处理

| 用例ID | 场景 | 预期行为 | 验证方法 |
|--------|------|---------|---------|
| TTL01 | 导出带 TTL 的 Key | 保留 TTL 信息 | 元数据检查 |
| TTL02 | 导入时恢复 TTL | TTL 正确设置 | redis-cli TTL 验证 |
| TTL03 | 已过期的 Key | 跳过或警告 | 不导出 |
| TTL04 | 无 TTL 的 Key | TTL = -1 | 标记为永不过期 |

**元数据格式**:
```json
{
  "key": "cache:user:10001:profile",
  "type": "string",
  "value": "{...}",
  "ttl": 3600,
  "encoding": "utf-8"
}
```

---

## 六、性能测试

### 6.1 大数据量测试

| 用例ID | 场景 | 数据量 | 预期性能 |
|--------|------|--------|----------|
| PERF01 | 导出大量 String | 10,000 keys | < 10秒 |
| PERF02 | 导出大型 Hash | 10,000 fields | < 15秒 |
| PERF03 | 导出大型 ZSET | 100,000 members | < 30秒 |
| PERF04 | 导入大量数据 | 10,000 keys | < 20秒 |

**性能测试脚本**:
```bash
# 生成测试数据
for i in {1..10000}; do
  redis-cli SET "perf:test:$i" "value_$i"
done

# 测试导出性能
time connect redis dump \
  --pattern "perf:test:*" \
  -o perf_test.zip

# 测试导入性能
time connect redis import perf_test.zip
```

---

## 七、端到端测试

### 7.1 完整工作流测试

#### 测试用例组 E2E: 完整流程

**场景**: 玩家充值问题调试 - MySQL + Redis 完整克隆

```bash
#!/bin/bash
# e2e_redis_test.sh

echo "=== Redis E2E 测试：玩家充值问题 ==="

# 1. 准备测试环境
redis-cli FLUSHDB
dolt sql "DROP TABLE IF EXISTS user_import;"

# 2. 导出 MySQL 订单数据
echo "步骤1: 导出 MySQL 订单数据"
connect mysql dump \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user \
  test_user_001 test_user_002 \
  -o user_mysql.zip

# 3. 导出 Redis 余额数据
echo "步骤2: 导出 Redis 余额数据"
connect redis dump \
  -h 127.0.0.1 -p 6379 \
  --pattern "balance:user:*" \
  -o balance_redis.zip

# 4. 验证导出结果
echo "步骤3: 验证导出结果"
if [ -f "user_mysql.zip" ] && [ -f "balance_redis.zip" ]; then
  echo "✅ 导出成功"
else
  echo "❌ 导出失败"
  exit 1
fi

# 5. 导入到测试环境
echo "步骤4: 导入到测试环境"
connect mysql import user_mysql.zip
connect redis import balance_redis.zip --db 1

# 6. 数据一致性验证
echo "步骤5: 数据一致性验证"
# MySQL 查询
mysql_result=$(dolt sql "SELECT gems FROM user WHERE accountid='test_user_001';")
# Redis 查询
redis_result=$(redis-cli -n 1 HGET "balance:user:10001" "gems")

# 对比
if [ "$mysql_result" = "$redis_result" ]; then
  echo "✅ 数据一致性验证通过"
else
  echo "❌ 数据不一致：MySQL=$mysql_result, Redis=$redis_result"
  exit 1
fi

echo "=== E2E 测试完成 ==="
```

---

## 八、回归测试检查清单

### 8.1 场景覆盖检查

| 场景 | 测试用例组 | 覆盖率 | 状态 |
|------|-----------|--------|------|
| **场景1: Redis 中转** | MR101-MR103, V101-V105 | 100% | 待实现 |
| **场景2: 业务数据** | RS01-RS04, RH01-RH03, RZ01-RZ03 | 100% | 待实现 |
| **场景3: 降低门槛** | UX101-UX104, UX201-UX204, UX301-UX304 | 100% | 待实现 |

### 8.2 数据类型覆盖

| Redis 类型 | 测试用例数 | 覆盖场景 | 状态 |
|-----------|-----------|---------|------|
| String | 12+ | 会话、配置、中转 | 待实现 |
| Hash | 8+ | 背包、属性 | 待实现 |
| ZSET | 6+ | 排行榜 | 待实现 |
| Set | 4+ | 好友列表 | 待实现 |
| List | 2+ | 消息队列(可选) | 待实现 |
| TTL | 4+ | 过期处理 | 待实现 |
| 二进制安全 | 4+ | Protobuf | 待实现 |

### 8.3 核心功能验证

✅ **必须通过的测试**:
- [ ] String dump/import (会话数据)
- [ ] Hash dump/import (背包数据)
- [ ] Pattern 匹配 (降低门槛)
- [ ] MySQL → Redis → MySQL (中转场景)
- [ ] 二进制数据完整性 (Protobuf)

⚠️ **重要测试**:
- [ ] ZSET dump/import (排行榜)
- [ ] Set dump/import (好友)
- [ ] TTL 处理
- [ ] 性能基准 (10,000 keys)

📊 **可选测试**:
- [ ] List 类型
- [ ] HyperLogLog
- [ ] Bitmap
- [ ] Geo

---

## 九、测试执行顺序

### Phase 1: 基础功能 (P0)

```bash
# 1. String 类型基础测试
connect redis dump --pattern "session:user:10001" -o test.zip
connect redis import test.zip

# 2. Hash 类型基础测试
connect redis dump --pattern "inventory:user:10001" -o test.zip
connect redis import test.zip
```

### Phase 2: 中转场景 (P1)

```bash
# MySQL → Redis
connect mysql dump --output-format redis ...

# Redis → MySQL
connect redis dump --pattern "*" -o export.zip
connect mysql import export.zip
```

### Phase 3: 完整功能 (P2)

```bash
# 运行所有测试用例
bash tests/run_redis_regression.sh
```

---

## 十、故障排除

### 问题1: Redis 连接失败

```bash
# 检查 Redis 是否运行
redis-cli ping

# 检查端口
netstat -an | grep 6379

# 启动 Redis
redis-server --port 6379
```

### 问题2: Pattern 无匹配

```bash
# 检查实际存在的 Key
redis-cli KEYS "*"

# 验证 Pattern 语法
# 支持 * 通配符和简单前缀/后缀匹配
```

### 问题3: 导入失败

```bash
# 检查 ZIP 结构
unzip -l test.zip

# 检查 Redis 内存
redis-cli INFO memory

# 清理测试数据
redis-cli FLUSHDB
```

---

*维护者: LVAN Dumper 开发团队*
*最后更新: 2025-02-02*
*Redis 版本要求: >= 5.0*
