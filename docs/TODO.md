# LVAN Dumper 待办事项

## 当前分支: lvan/dumper

**最后更新**: 2025-01-31

---

## 🔴 高优先级 (P0)

### 修复诊断错误

- [x] **TODO-001**: 修复 `encoding_examples.go` 未使用的 `strings` 导入
  - 状态: ✅ 已完成 (需手动处理)
  - 文件: `encoding_examples.go:10`
  - 问题: 导入了 `strings` 包但未使用
  - 修复: 移除未使用的导入

- [x] **TODO-002**: 审查并清理 go-spring 集成代码
  - 状态: ✅ 已完成
  - 文件: `lvan/cmd/dumper/cmd/dump.go`, `mysql.go`
  - 处理: 移除未完成的 `gs.Run()` 和 `gs.Object(&{})` 调用
  - 原因: go-spring 文档不完整，迁移困难，使用访问者模式作为临时替代

- [x] **TODO-010**: 修复 TIMESTAMP 导入失败问题
  - 状态: ✅ 已完成 (2025-01-31)
  - 问题: Error 1105 (HY000): 3905527098775974194 out of range for int
  - 原因: Dump() 将 TIMESTAMP 扫描为 []byte，导出为二进制表示
  - 修复: 添加 GetColumnTypes() 函数，时间类型使用 time.Time 扫描并转换为字符串

- [x] **TODO-011**: 修复 ZIP 文件名格式化问题
  - 状态: ✅ 已完成 (2025-01-31)
  - 问题: 文件名生成 %!s(MISSING).%!s(MISSING).zip
  - 原因: fmt.Sprintf("%s.%s.zip") 缺少参数
  - 修复: 从 viper 获取 database 和 table 参数

- [x] **TODO-012**: 架构重构 - 统一为 context 传递模式
  - 状态: ✅ 已完成 (2025-01-31)
  - 文件: `docs/ARCHITECTURE_ADR.md`, `docs/REFACTOR_PLAN.md`
  - 新增: `pkg/dump/service/`, `cmd/dumper/cmd/cmdcontext/`
  - 重构: 移除全局变量依赖，使用 context.Context 传递数据源
  - 详见: [架构决策记录](docs/ARCHITECTURE_ADR.md)

---

## 🟡 中优先级 (P1)

### 测试完善

- [ ] **TODO-003**: 补充 Redis dump/import 测试
  - 状态: Redis 数据源代码存在但缺少测试
  - 代码位置: `lvan/cmd/dumper/cmd/redis.go`
  - 参考: MySQL 测试用例实现

- [x] **TODO-004**: 创建 Dolt 测试数据库脚本
  - 状态: ✅ 已完成
  - 文件: `tests/setup_test_db.sh`, `tests/setup_test_db.bat`
  - 覆盖: 所有 MySQL 数据类型，重点 BLOB 字段

- [x] **TODO-005**: 创建端到端测试脚本
  - 状态: ✅ 已完成
  - 文件: `tests/e2e_test.sh`
  - 覆盖: 完整 dump → import 流程，数据一致性验证

- [x] **TODO-006**: 更新回归测试文档
  - 状态: ✅ 已完成
  - 文件: `docs/REGRESSION_TEST.md`
  - 新增: Dolt 环境准备、CLI 测试用例、BLOB 专项测试

- [ ] **TODO-007**: 运行完整回归测试
  - 状态: 📋 待执行
  - 任务: 运行 `tests/e2e_test.sh` 验证所有功能
  - 依赖: Dolt 正确安装

### 功能完善

- [ ] **TODO-008**: 实现 SQL 文件导入功能 (import-sql)
  - 设计文档: `docs/DESIGN_SQL_IMPORT.md`
  - 状态: 📋 待实现
  - 文件: `lvan/cmd/dumper/cmd/import_sql.go` (新建)
  - 文件: `lvan/pkg/dump/load/sql.go` (新建)
  - 方案: Hybrid (快速解析 + MySQL 回退)
  - 优先级: 高

  **子任务**:
  - [ ] 实现命令框架和参数解析
  - [ ] 实现 SQL 快速解析器
  - [ ] 实现 MySQL 回退方案
  - [ ] 添加临时数据库管理
  - [ ] 添加进度显示
  - [ ] 添加错误处理和日志

- [ ] **TODO-009**: 完善 SQL 模板导出格式
  - 代码位置: `lvan/cmd/dumper/cmd/sqlTpl.go:1-39`
  - 当前状态: 占位实现
  - 需要实现: 生成标准 INSERT SQL 语句

```go
// 当前实现 (需要完善)
case tpsql:
    return func(records []dump.Record, pks ...string) string {
        return fmt.Sprintf("%v", records)  // 占位代码
    }
```

- [ ] **TODO-007**: 实现重复主键导入策略
  - 问题: 导入时遇到重复主键如何处理？
  - 选项: SKIP / REPLACE / UPDATE
  - 代码位置: `lvan/pkg/dump/import.go`

- [ ] **TODO-008**: 添加导出进度显示
  - 大量数据导出时显示进度
  - 使用进度条或百分比

---

## 🟢 低优先级 (P2)

### 代码优化

- [ ] **TODO-009**: 统一错误处理模式
  - 当前: 混用 `log.Panic()` 和错误返回
  - 目标: 统一使用错误返回链

- [ ] **TODO-010**: 优化日志输出
  - 移除调试用的 `log.Println`
  - 使用结构化日志库 (如 zap)

- [ ] **TODO-011**: 代码文档完善
  - 为导出的函数添加 godoc 注释
  - 添加包级别的文档说明

### 新功能

- [ ] **TODO-012**: PostgreSQL 数据源支持
  - 实现 `PostgresDatasource`
  - 参考 MySQL 数据源实现

- [ ] **TODO-013**: 数据加密功能
  - 导出时支持 AES/GCM 加密
  - 导入时自动解密

- [ ] **TODO-014**: 增量导出
  - 基于时间戳的增量导出
  - 支持仅导出变更数据

- [ ] **TODO-015**: 并行导出优化
  - 大表分块并行导出
  - 使用 worker pool 模式

---

## 📋 技术债务

### 架构相关

- [ ] **DEBT-001**: v1 和 v2 数据源并存
  - v1: `lvan/pkg/dump/` 中的旧实现
  - v2: `lvan/pkg/dump/datasource/v2/` 新实现
  - 计划: 逐步迁移到 v2，废弃 v1

- [ ] **DEBT-002**: 访问者模式使用不一致
  - `internal/accept.go` 定义了访问者模式
  - 但部分代码未使用此模式
  - 需要统一架构

### 测试相关

- [ ] **DEBT-003**: 测试覆盖率报告
  - 当前: 无统一覆盖率跟踪
  - 计划: 添加 CI/CD 覆盖率检查

- [ ] **DEBT-004**: 集成测试依赖真实数据库
  - `main_test.go` 硬编码了远程数据库地址
  - 需要改为使用 Mock 框架

```go
// 问题代码: lvan/cmd/dumper/main_test.go:42
host := "101.34.211.79:32533"  // 硬编码
username := "root"
password := "p_mysql"
```

---

## 📝 文档完善

- [ ] **DOC-001**: 补充 API 文档
  - 使用 godoc 生成 API 文档
  - 添加使用示例

- [ ] **DOC-002**: 编写贡献指南
  - 代码风格规范
  - PR 提交流程
  - 测试要求

- [ ] **DOC-003**: 添加故障排除指南
  - 常见问题及解决方案
  - 调试技巧

---

## 🗓️ 里程碑

### v1.0 - 当前版本
- [x] 修复所有 P0 级别问题
- [x] 完成基础功能测试
- [x] 创建 Dolt 测试环境
- [x] 完善回归测试文档
- [ ] 运行完整回归测试验证

### v1.1 - 下一版本
- [ ] SQL 模板导出
- [ ] 重复主键处理策略
- [ ] Redis 功能完善
- [ ] 性能基准测试

### v2.0 - 未来版本
- [ ] PostgreSQL 支持
- [ ] 数据加密
- [ ] 增量导出
- [ ] 并行优化

---

## 🔍 问题跟踪

### 已知问题

| ID | 描述 | 严重程度 | 状态 | 负责人 |
|----|------|---------|------|--------|
| TODO-001 | encoding_examples.go 未使用的导入 | 低 | ✅ 已处理 | - |
| TODO-002 | go-spring 集成代码清理 | 中 | ✅ 已完成 | - |
| DEBT-004 | 集成测试硬编码数据库 | 中 | 🟡 待重构 | - |

### 最近完成 (2025-01-31)

| ID | 描述 | 完成日期 |
|----|------|---------|
| TODO-001 | 移除 go-spring 未完成代码 | 2025-01-31 |
| TODO-004 | 创建 Dolt 测试数据库脚本 | 2025-01-31 |
| TODO-005 | 创建端到端测试脚本 | 2025-01-31 |
| TODO-006 | 更新回归测试文档 | 2025-01-31 |
| TODO-010 | 修复 TIMESTAMP 导入失败 | 2025-01-31 |
| TODO-011 | 修复 ZIP 文件名格式化 | 2025-01-31 |
| TODO-012 | 架构重构：统一为 context 传递模式 | 2025-01-31 |
| - | MySQL Mock 框架实现 | 2025-01-10 |
| - | v2 数据源抽象 | 2025-11-10 |
| - | ZIP 格式测试 | 已完成 |

---

## 📊 统计

- **总任务数**: 20+
- **高优先级**: 2
- **中优先级**: 8
- **低优先级**: 6
- **技术债务**: 4

---

*维护者: LVAN Dumper 开发团队*
*最后更新: 2025-01-31*
