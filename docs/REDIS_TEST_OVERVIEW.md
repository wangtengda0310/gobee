# Redis 支持测试用例总览

> 针对场景1、2、3 的完整测试覆盖分析

---

## 📊 测试覆盖总览

### 整体覆盖情况

| 场景 | 描述 | 测试用例 | 数据类型 | 优先级 | 测试状态 |
|------|------|---------|---------|--------|---------|
| **场景1** | Redis 中转 | 11 | 4 种 | P1 | 📄 已设计 |
| **场景2** | Redis 业务数据 | 38 | 5 种 | **P0** | 📄 已设计 |
| **场景3** | 降低使用门槛 | 11 | 4 种 | **P0** | 📄 已设计 |
| **专项测试** | 二进制/TTL/性能 | 8 | 特殊 | P1-P2 | 📄 已设计 |
| **总计** | - | **68+** | - | - | **📄 已设计** |

---

## 场景1: Redis 作为数据中转

### 测试流程

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  生产环境 MySQL  │────▶│  中转环境 Redis  │────▶│  本地环境 MySQL  │
│  (受限访问)      │     │  (开放访问)      │     │  (开发调试)      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                            ↑                            ↑
                      Redis 导出/导入              数据验证
```

### 测试用例清单

| ID | 测试用例 | 命令示例 | 验证方法 | 状态 |
|----|---------|---------|---------|------|
| **MR101** | MySQL → Redis 导出 | `connect mysql dump --output-format redis --redis-key-prefix "export:"` | `redis-cli EXISTS` | 📄 |
| **MR102** | 设置 TTL | `--redis-ttl 3600` | `redis-cli TTL` | 📄 |
| **MR103** | 批量导出 | 导入多个记录 | `redis-cli KEYS` 计数 | 📄 |
| **MR201** | Redis → 本地导出 | `connect redis dump --pattern "export:*" -o export.zip` | 检查 ZIP | 📄 |
| **MR202** | 本地 MySQL 导入 | `connect mysql import export.zip` | MySQL 查询 | 📄 |
| **V101** | String 完整性 | 值完全相等 | JSON 对比 | 📄 |
| **V102** | Hash 字段数 | HLEN 相等 | 计数对比 | 📄 |
| **V103** | ZSET 分数精度 | 误差 < 1e-6 | 浮点对比 | 📄 |
| **V104** | Set member 数量 | SCARD 相等 | 计数对比 | 📄 |
| **V105** | 二进制安全 | HEX 相等 | HEX 对比 | 📄 |

**测试数据准备**:
```bash
# 模拟 MySQL 导出到 Redis 的数据
redis-cli SET "export:mysql:user:10001" '{"uid":10001,"accountid":"test_user_001"}'
redis-cli SET "export:mysql:user:10002" '{"uid":10002,"accountid":"test_user_002"}'
```

---

## 场景2: Redis 业务数据

### 数据分布

```
Redis 数据分布:
├── String (30%)
│   ├── session:user:*     # 玩家会话
│   ├── config:*            # 游戏配置
│   └── ratelimit:*         # 限流计数
├── Hash (40%)
│   ├── inventory:user:*    # 玩家背包
│   └── user:profile:*      # 玩家属性
├── ZSET (20%)
│   └── leaderboard:*       # 排行榜
└── Set (10%)
    └── friends:user:*      # 好友列表
```

### 测试用例清单

#### 2.1 String 类型测试

| ID | 测试用例 | 测试数据 | 验证点 | 状态 |
|----|---------|---------|--------|------|
| **RS01** | 单个 String dump | `session:user:10001` | ZIP 结构正确 | 📄 |
| **RS02** | 多个 String dump | `session:user:10001, 10002` | 多目录 ZIP | 📄 |
| **RS03** | Pattern 匹配 | `session:*` | 导出所有会话 | 📄 |
| **RS04** | 配置数据 | `config:*` | 配置完整性 | 📄 |
| **RSI01** | String import | 导入会话数据 | `redis-cli GET` 验证 | 📄 |
| **RSI02** | 覆盖已存在 key | 根据策略处理 | 行为验证 | 📄 |
| **RSI03** | 导入到不同 DB | `--db 1` | SELECT 验证 | 📄 |

#### 2.2 Hash 类型测试

| ID | 测试用例 | 测试数据 | 验证点 | 状态 |
|----|---------|---------|--------|------|
| **RH01** | Hash dump | `inventory:user:10001` | 所有字段在 ZIP | 📄 |
| **RH02** | 字段完整性 | gold=10000,gems=500 | 值保持一致 | 📄 |
| **RH03** | 多个 Hash | 背包 x 3 | 目录计数正确 | 📄 |
| **RHI01** | Hash import | 导入背包数据 | `redis-cli HGETALL` | 📄 |
| **RHI02** | 字段类型 | 数值/字符串 | 类型保持 | 📄 |

#### 2.3 ZSET 类型测试

| ID | 测试用例 | 测试数据 | 验证点 | 状态 |
|----|---------|---------|--------|------|
| **RZ01** | ZSET dump | `leaderboard:level` | 结构正确 | 📄 |
| **RZ02** | 分数保持排序 | 按 score 排序 | 顺序验证 | 📄 |
| **RZ03** | member-score 对 | 成对存储 | 数据完整性 | 📄 |

#### 2.4 Set 类型测试

| ID | 测试用例 | 测试数据 | 验证点 | 状态 |
|----|---------|---------|--------|------|
| **RS01** | Set dump | `friends:user:10001` | 所有 member 导出 | 📄 |
| **RS02** | member 数量 | SCARD 验证 | 计数正确 | 📄 |

---

## 场景3: 降低使用门槛

### 核心价值

```
传统 Redis 操作:
❌ redis-cli KEYS "session:*"     # 需要知道 KEYS 命令
❌ redis-cli GET $key            # 需要循环处理
❌ redis-cli HGETALL $key        # 需要知道 HGETALL 命令
❌ 手动复制粘贴输出               # 容易出错

LVAN Dumper:
✅ connect redis dump --pattern "session:*" -o sessions.zip  # 统一接口
✅ connect redis import sessions.zip                        # 统一接口
✅ 自动处理所有类型                                          # 无需专业知识
```

### 测试用例清单

#### 3.1 统一接口测试

| ID | 测试用例 | 对比 | 状态 |
|----|---------|-----|------|
| **UX101** | 命令结构一致性 | MySQL vs Redis 参数结构 | 📄 |
| **UX102** | 操作流程一致性 | dump/import 流程相同 | 📄 |
| **UX103** | 帮助文档 | `--help` 提供示例 | 📄 |

#### 3.2 Pattern 匹配测试

| ID | 测试用例 | Pattern | 预期匹配 | 状态 |
|----|---------|---------|---------|------|
| **UX201** | 前缀匹配 | `session:*` | 所有 session key | 📄 |
| **UX202** | 多类型匹配 | `user:*` | String/Hash 混合 | 📄 |
| **UX203** | 后缀匹配 | `*:10001` | 所有用户10001数据 | 📄 |
| **UX204** | 无匹配处理 | `nonexistent:*` | 友好错误提示 | 📄 |

#### 3.3 错误处理测试

| ID | 测试用例 | 场景 | 预期行为 | 状态 |
|----|---------|------|---------|------|
| **UX301** | 连接失败 | Redis 未运行 | "无法连接到 Redis" | 📄 |
| **UX302** | 无匹配 Key | Pattern 无结果 | "未找到匹配的 Key" | 📄 |
| **UX303** | 认证失败 | 需要 AUTH | "需要密码认证" | 📄 |
| **UX304** | 内存不足 | OOM | "Redis 内存不足" | 📄 |

---

## 专项测试

### 二进制安全测试

| ID | 场景 | 数据 | 验证方法 | 状态 |
|----|------|------|---------|------|
| **BIN01** | Protobuf 数据 | 任意字节 | HEX 对比 | 📄 |
| **BIN02** | NULL 字节 | `\x00` | 字节级对比 | 📄 |
| **BIN03** | UTF-8 编码 | 中文 | 编码保持 | 📄 |
| **BIN04** | 大 Value (1MB) | 大块数据 | 完整性 | 📄 |

### TTL 测试

| ID | 场景 | 行为 | 验证方法 | 状态 |
|----|------|------|---------|------|
| **TTL01** | 导出 TTL | 保留元数据 | 检查 TTL 字段 | 📄 |
| **TTL02** | 导入恢复 TTL | 设置过期时间 | `redis-cli TTL` | 📄 |
| **TTL03** | 已过期 Key | 跳过 | 不导出 | 📄 |
| **TTL04** | 无 TTL Key | 标记为永久 | TTL = -1 | 📄 |

### 性能测试

| ID | 场景 | 数据量 | 预期性能 | 状态 |
|----|------|--------|----------|------|
| **PERF01** | String 导出 | 10,000 keys | < 10s | 📄 |
| **PERF02** | Hash 导出 | 10,000 fields | < 15s | 📄 |
| **PERF03** | ZSET 导出 | 100,000 members | < 30s | 📄 |
| **PERF04** | 批量导入 | 10,000 keys | < 20s | 📄 |

---

## 测试执行流程

### 快速验证流程

```bash
# 1. 启动 Redis
docker run -d --name redis-test -p 6379:6379 redis:7-alpine

# 2. 准备测试数据
bash tests/setup_redis_test.sh

# 3. 运行 E2E 测试 (需要功能实现后)
bash tests/e2e_redis_test.sh

# 4. 查看测试报告
cat e2e_test_report.txt
```

### 完整回归测试

```bash
# 运行所有回归测试
bash tests/run_regression.sh

# 包含:
# - MySQL 测试 (现有)
# - Redis 测试 (新增)
# - 性能测试
# - 二进制安全测试
```

---

## 实现与测试的对应关系

### Phase 1: String + Hash 基础功能

**实现内容**:
- `RedisManager` 基础类
- String 类型的 dump/import
- Hash 类型的 dump/import

**对应测试**:
- ✅ RS01-RS04, RSI01- RSI03 (String)
- ✅ RH01-RH03, RHI01- RHI02 (Hash)
- ✅ UX101-UX103 (统一接口)
- ✅ UX201-UX204 (Pattern)

### Phase 2: 中转功能

**实现内容**:
- MySQL → Redis 输出格式
- Redis → MySQL 导入流程

**对应测试**:
- ✅ MR101-MR103 (MySQL → Redis)
- ✅ MR201-MR202 (Redis → MySQL)
- ✅ V101-V105 (数据完整性)

### Phase 3: 完整支持

**实现内容**:
- ZSET/Set 类型支持
- TTL 处理
- 错误处理优化
- 性能优化

**对应测试**:
- ✅ RZ01-RZ03 (ZSET)
- ✅ RS01-RS02 (Set)
- ✅ TTL01-TTL04 (过期时间)
- ✅ UX301-UX304 (错误处理)
- ✅ PERF01-PERF04 (性能)

---

## 测试用例覆盖对比表

| 测试类别 | MySQL 测试 | Redis 测试 | 覆盖率对比 |
|---------|-----------|-----------|-----------|
| 基础 dump/import | ✅ 完整 | 🔶 待实现 | Redis 需要实现 |
| 二进制安全 | ✅ 重点 (BLOB) | 🔶 专项设计 | Protobuf 场景相同 |
| 性能测试 | ✅ 1000 条记录 | 🔶 1000 keys | 规模相似 |
| 端到端测试 | ✅ 完整脚本 | 🔶 完整脚本 | 流程一致 |
| 错误处理 | ✅ 基础覆盖 | 🔶 友好化设计 | Redis 更友好 |

---

## 总结

### ✅ 已完成

1. **测试文档** (`docs/REGRESSION_TEST_REDIS.md`)
   - 68+ 详细测试用例
   - 清晰的执行流程
   - 故障排除指南

2. **测试脚本** (`tests/setup_redis_test.sh`, `tests/e2e_redis_test.sh`)
   - 自动化数据准备
   - 完整 E2E 测试流程
   - 测试报告生成

3. **覆盖分析** (`docs/REDIS_TEST_COVERAGE_ANALYSIS.md`)
   - 3 个场景完整覆盖
   - 与现有测试对比
   - 实现优先级规划

### 🔶 待实现

1. **Redis 功能实现** (Phase 1-4)
2. **测试脚本执行** (功能实现后)
3. **持续集成集成** (CI/CD)

### 📊 测试用例统计

- **总测试用例数**: 68+
- **场景1 测试**: 11 个
- **场景2 测试**: 38 个
- **场景3 测试**: 11 个
- **专项测试**: 8 个

---

*维护者: LVAN Dumper 开发团队*
*最后更新: 2025-02-02*
