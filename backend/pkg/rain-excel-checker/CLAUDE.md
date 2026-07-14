# rain-excel-checker 项目说明

策划配表规则检查工具，用于 CI/CD 流水线自动检查 Excel 配表是否符合预设规则。

## 技术栈
- Go 1.25
- 依赖: `pkg/common/feishu` (飞书通知)

## 注意事项
- **子模块依赖同步**：修改本模块或 pkg/common/feishu 后，需按根目录 CLAUDE.md 中的依赖更新流程同步下游 go.mod
- **目录位置**：本模块现在位于 `rain-qa-func/rain-excel-checker/` 下（已从根目录移入）
- **代码格式规范(Go)**：
  - 必须使用 tab 缩进(`gofmt` 默认行为)，禁止空格缩进
  - 结构体多字段初始化时每个字段缩进必须一致
  - 修改或新增 Go 代码后，提交前必须运行 `gofmt -d <file>` 检查格式差异
  - 禁止使用 `sed` 插入或修改 Go 代码
  - 多人/AI 协作生成的代码，review 时重点检查结构体字段对齐和缩进层级是否一致

## 目录结构
```
rain-excel-checker/
├── main.go           # 入口，merge commit 检测与遍历检查
├── dispatch/         # 通知分发策略（分支→通知通道映射 + handler 组装）
├── gitutil/          # Git 操作工具函数(仓库查询、merge 检测、变更文件获取)
└── xlsx/             # Excel 检查核心逻辑
    ├── check_internal/   # 快照管理、差异检测、规则辅助工具
    ├── check_manager/    # 检查器工厂管理
    ├── coded_rules/      # 规则实现(跨表/表级/通用)
    ├── json_rule/   # 规则类型定义、数据结构
    ├── excel_internal/   # Excel 解析、列辅助工具
    └── tests/            # 集成测试
```

## 命令行参数
| 参数 | 默认值 | 说明 |
|------|--------|------|
| -feishuRobot | vars.LJH | 飞书机器人 webhook ID（设为 `none` 不发群消息） |
| -excelPath | vars.ExcelPath | Excel 目录 |
| -casePath | vars.ExcelCheckRule | 检查规则目录 |
| -mode | incremental | 检查模式: incremental(增量), full(全量-本地磁盘) |
| -noDM | false | 禁用飞书私聊通知 |

程序自动通过 git 命令获取提交信息(作者、分支、变更文件列表)，不再需要手动传入。
测试时如无明确要求不要发送飞书消息避免消息污染

### 飞书通知配置

程序支持两种飞书通知通道，独立配置：

| 通道 | 凭证来源 | 用途 |
|------|----------|------|
| 群卡片消息 | `-feishuRobot` 命令行参数 | webhook 机器人向群发送检查结果卡片 |
| 私聊消息 | `FEISHU_DM_APP_ID` + `FEISHU_DM_APP_SECRET` 环境变量 | 向校验失败的 git 提交者发送私聊汇总 |

通知策略由 `dispatch` 包根据分支名决定（见 `dispatch/strategy.go`）：

| 分支 | 群消息 | 私聊 |
|------|--------|------|
| `v0.0.8-pre-release` | 发送 | 不发送 |
| 其他分支 | 不发送 | 发送 |

详见 [飞书私聊通知配置](docs/飞书私聊通知配置.md)。

### 中文文件名路径处理

详见 [gitutil/CLAUDE.md](gitutil/CLAUDE.md) — `decodeGitPath` 统一处理 Git 八进制转义。

### Merge Commit 处理

程序启动后自动判断当前 HEAD 是否为 merge commit：
- **普通 commit**：通过 `CheckWithFilter` 从本地磁盘加载文件执行检查
- **Merge commit**：遍历两个分支的所有子 commit，每个子 commit 通过 `CheckWithGitHistory` 从 git 历史获取文件版本执行检查(不读本地磁盘)，按需加载关联表(~15秒完成4个commit检查)

### 检查模式

详见 [xlsx/workflow/CLAUDE.md](xlsx/workflow/CLAUDE.md)。

## 关键入口
| 功能 | 位置 |
|------|------|
| 参数解析 | main.go:argParse |
| 全量检查 | main.go:handleFullCheck → CheckAll |
| merge 检测 | main.go:main (调用 gitutil.IsMergeCommit) |
| 普通 commit 检查 | main.go:runCheck → CheckWithGitHistory |
| merge commit 检查 | main.go:handleMergeCommit → CheckWithGitHistory |
| 通知分发 | main.go:dispatchResults |
| 通知策略 | dispatch/strategy.go `ResolveNotifyMode` |
| handler 组装 | dispatch/router.go `NotifyRouter.BuildDispatcher` |
| 增量检查(本地) | xlsx/check_manager/manager_def.go `CheckWithFilter` |
| 增量检查(git) | xlsx/check_manager/manager_def.go `CheckWithGitHistory` |
| 全量检查 | xlsx/check_manager/manager_def.go `CheckAll` |

## 输出
检查结果通过飞书机器人卡片消息通知。

### Merge 场景消息汇聚

Merge commit 的检查结果汇聚为一条飞书消息(而非每个 commit 各发一条)。详见 [merge 消息汇聚实现计划](docs/merge消息汇聚实现计划.md)。

## 规则系统

### 列级规则 vs 表级规则

| 特性 | 列级规则 | 表级规则 |
|------|----------|----------|
| 检查范围 | 单列数据 | 整张表 |
| 配置位置 | 每列一个规则卡片 | "0. 表级规则" 虚拟列 |
| 典型用途 | 唯一性、非空、枚举等 | 跨列逻辑、时间预警等 |
| 接口 | `Checker` | `TableChecker` |
| 已注册数量 | 25 个(engine/column_registry.go:57-97) | 17 个(check_manager/table_check_manager.go:72-121) |

### 新增表级规则步骤

详见 [xlsx/coded_rules/CLAUDE.md](xlsx/coded_rules/CLAUDE.md)。

### 表级检查结果

表级规则分为**错误检测规则**(Ok=false)和**通知规则**(Ok=true)，详见 [表级检查结果 Ok 字段语义](../../docs/表级检查结果Ok字段语义.md)。

`CellError` 结构包含 `Index`(行索引)、`Reason`(兜底显示)、`Detail`(结构化详情)，数据流向：规则实现 -> ErrCells -> 前端UI/MCP(用Detail), 飞书/命令行(用Reason)。详见 [规则类型定义](xlsx/json_rule/CLAUDE.md)。

## 更多文档

| 文档 | 说明 |
|------|------|
| [校验规则实现](xlsx/coded_rules/CLAUDE.md) | 三层结构、新增规则流程、单元测试要求 |
| [规则类型定义](xlsx/json_rule/CLAUDE.md) | 所有规则类型、参数定义、数据结构(CellError/Detail 等) |
| [检查工作流](xlsx/workflow/) | 检查流程编排、Checker/TableChecker 接口、增量检查 |
| [NewRowNotifyRule.ParamDefs.深度分析](docs/NewRowNotifyRule.ParamDefs.%E6%B7%B1%E5%BA%A6%E5%88%86%E6%9E%90) | ParamDef 相关逻辑分析，供重构参考 |
| [merge 消息汇聚实现计划](docs/merge消息汇聚实现计划.md) | ~~Merge 场景下飞书消息汇聚为一条的设计与实现~~ (已实现) |
| [性能优化记录](docs/性能优化记录.md) | **性能优化历程、已实施方案、进一步优化方向** |
| [merge 跨表规则数据完整性 TODO](docs/todo-merge跨表规则数据完整性.md) | ~~merge 遍历中跨表规则缺少关联数据的问题和长期方案~~ (已解决,归档) |
| [性能优化指导手册](docs/性能优化指导手册.md) | ~~性能瓶颈分析、优化路线图、立即可实施的优化方案~~ (已解决,归档) |
| [规则验证记录](docs/规则验证记录.md) | 跨表规则验证结果汇总(单元测试、历史提交测试、命令速查) |
| [跨表规则开发手册](docs/跨表规则开发手册.md) | 新增跨表规则的完整步骤、实现模板、测试规范 |
| [丹青阁活动和掉落逻辑验证规则设计](docs/丹青阁活动和掉落逻辑验证规则设计.md) | 10个待实现规则的设计(含服务端代码引用、数据结构) |
| [测试规则补充需求-丹青阁活动和掉落逻辑](docs/测试规则补充需求-丹青阁活动和掉落逻辑.md) | 8个待实现规则的需求、验收标准、优先级 |
| [规则映射表](docs/规则映射表.md) | 所有规则的前端操作路径与代码文件对应关系，方便AI Agent快速定位 |
| [武将保护期检查框架设计](docs/武将保护期检查框架设计.md) | 战令武将保护期通用检查框架：数据关系链、折中方案设计、关系链配置结构、通用求值器 |
| [规则重构计划-拆分增删改通知规则](docs/规则重构计划-拆分增删改通知规则.md) | ~~将 NEW_ROW_NOTIFY/ROW_CHANGE_NOTIFY 拆分为 5 个独立规则，含缓存层、向后兼容、迁移指南~~ (已完成) |
| [掉落道具和皮肤抽奖规则实现总结](docs/RULE_IMPLEMENTATION_SUMMARY.md) | DropItem/DrawFix 等规则实现总结 |
| [配置验收清单](docs/excel配置验收清单/CLAUDE.md) | 各表校验规则清单（按表组织，含实现状态追踪） |
| [表级规则变更检测-实现详解](docs/表级规则变更检测-实现详解.md) | 快照机制、差异检测核心逻辑 |
| [流水线知识手册](docs/流水线知识手册.md) | 流水线检查流程、命令行参数、CI集成 |
| [关系链检查数据流模型](docs/关系链检查数据流模型.md) | 跨表引用检查的数据流设计 |
| [跨表规则影响评估报告](docs/跨表规则影响评估报告.md) | 跨表规则变更对现有检查的影响评估 |
