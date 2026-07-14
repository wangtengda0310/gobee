# json_rule - 规则类型定义和默认配置

## 核心文件

| 文件 | 说明 |
|------|------|
| `rule_def.go` | 规则类型枚举、数据结构、工具函数 |
| `default_table_rules.go` | 默认表级规则配置 |

## 规则类型枚举

### 表级规则 (ETableRule)

定义于 `rule_def.go:24-41`，共 20 种(18 个已注册，ACTIVITY_TASK_REWARD_CHECK 已定义未注册)：

| 规则类型 | 说明 | 适用表 |
|----------|------|--------|
| `ARENA_SEASON_CHECK` | 竞技场赛季检查 | ArenaSeason |
| `NEW_ROW_NOTIFY` | 新增行/列通知(旧版，保留兼容) | 所有表 |
| `ADDED_ROW_NOTIFY` | 新增行通知 | 所有表 |
| `REMOVED_ROW_NOTIFY` | 删除行通知 | 所有表 |
| `ADDED_COL_NOTIFY` | 新增列通知 | 所有表 |
| `REMOVED_COL_NOTIFY` | 删除列通知 | 所有表 |
| `MODIFIED_ROW_NOTIFY` | 修改行通知 | 所有表 |
| `ROW_CHANGE_NOTIFY` | 行变更字段通知 | 所有表 |
| `SEASON_PASS_HERO_OPEN_CHECK` | 战令武将开放时间检查 | SeasonPassReward |
| `ARENA_GENERAL_HERO_OPEN_CHECK` | 大将军武将开放时间检查 | ArenaScoreReward |
| `HERO_DROP_CHECK` | 武将抽卡掉落检查 | Hero |
| `HERO_SYNTHESIS_CHECK` | 武将合成检查 | Item(IsSynthetic 字段在 Item 表) |
| `HERO_MELT_CHECK` | 武将熔炼检查 | Hero |
| `HERO_ISOPEN_OPENDATE_CHECK` | 武将IsOpen与OpenDate一致性检查 | Hero |
| `HERO_DROP_VALIDDATE_CHECK` | 普通武将掉落时间检查 | DropItem |
| `ACTIVITY_TASK_REWARD_CHECK` | 活动任务奖励检查 | ActivityTask(已定义未注册) |
| `DROP_ITEM_MUST_IN_ITEM_CHECK` | 掉落道具检查 | DropItem |
| `DRAWSKIN_BYPRODUCT_CHECK` | 皮肤抽奖副产品检查 | DrawSkin |
| `DANQINGGE_TIME_ACTIVE_CHECK` | 丹青阁活动时间校验 | Activity |
| `DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK` | 丹青阁自定义参数检查 | Activity |
| `DRAWFIX_PROTECTION_CHECK` | 定向招募战令保护期检查 | DrawFix |
| `DRAWFIX_ARENA_PROTECTION_CHECK` | 定向招募大将军保护期检查 | DrawFix |

### 列级规则 (EColRule)

定义于 `rule_def.go:205-279`，分为四类：

| 类别 | 规则 |
|------|------|
| 基础规则 | TEST, INCREASE_ID, UNIQUE, CHS_ONLY, NOT_EMPTY, SERVER_OR_CLIENT, ALL_BASE |
| 数据类型 | NUMERIC, DATE, BOOLEAN, STRING |
| 业务关系 | DATE_RANGE, DATE_DURATION, NUMERIC_RANGE, ENUM, FOREIGN_KEY, CROSS_REFERENCE, CHAIN_REFERENCE, SPLIT_REGERENCE, SPECIAL_FORMAT, REGEX |
| 其他 | WEIGHT_SUM, DATE_CONSISTENCY, RESOURCE, PIN_YIN_CHS, RICH_TEXT |

### 规则参数 (ERuleParam)

定义于 `rule_def.go:286-321`，包括：
- **通用参数**: allowEmpty, excepts, allowCommit, enums, breakLine, strict, compareRule, tolerance, min, max, pattern, groups
- **日期参数**: startDate, endDate
- **表级规则参数**: seasonEndTimeCol, timeRangeBefore, idColName, nameColName, oldDataPath, notifyAddedRows 等

### 跨表引用参数名

`ReferenceSheetParamKeys` (rule_def.go) — 列级规则中引用其他表的参数名数组：
- `targetSheet` — FOREIGN_KEY 规则使用
- `refSheet` — CROSS_REFERENCE/SPLIT_REFERENCE 规则使用
- `descSheet` — DATE_CONSISTENCY 规则使用

## 核心数据结构

所有结构定义于 `rule_def.go`：

| 结构体 | 行号 | 用途 |
|--------|------|------|
| `TableRule` | :24 | 表级校验规则实例(Type/DisplayName/Uuid/Description/Params/Enabled) |
| `TableRuleMeta` | :55 | 表级规则元数据(含 TargetSheets/DefaultParams/ParamDefs) |
| `CellError` | :146 | 单元格错误/变更信息(Index/Reason/Detail) |
| `RowChangeDetail` | :153 | 行变更详情(changeType/rowId/rowName) |
| `FieldChangeDetail` | :160 | 字段变更详情(rowId/rowName/colName/oldValue/newValue) |
| `ColumnChangeDetail` | :169 | 列变更详情(changeType/colName) |
| `SheetRule` | :184 | Sheet 规则配置(Sheet/ManagerList/Rules/TableRules) |
| `ColRule` | :197 | 列级规则(Type/Params) |

**CellError.Detail 数据流向**：规则实现 -> TableCheckResult.ErrCells -> 前端UI/MCP(用Detail), 飞书/命令行(用Reason)

## 工具函数

| 函数 | 位置 | 说明 |
|------|------|------|
| `GetAllTableRuleMetas()` | rule_def.go:82 | 获取所有表级规则元数据 |
| `GetTableRuleMetasForSheet(sheetName)` | rule_def.go:92 | 获取指定表适用的元数据 |
| `GetDefaultTableRulesForSheet(sheetName)` | default_table_rules.go:29 | 获取指定表的默认规则 |
| `HasDefaultTableRule(sheetName, ruleType)` | default_table_rules.go:49 | 检查是否有指定默认规则 |

## 默认表级规则

`default_table_rules.go` 中自动应用的规则：

| 表名 | 自动应用的规则 |
|------|----------------|
| SeasonPassReward | SEASON_PASS_HERO_OPEN_CHECK |
| ArenaScoreReward | ARENA_GENERAL_HERO_OPEN_CHECK |
| ArenaSeason | ARENA_SEASON_CHECK |
| Hero | HERO_DROP_CHECK, HERO_MELT_CHECK, SEASON_PASS_HERO_OPEN_CHECK, ARENA_GENERAL_HERO_OPEN_CHECK, HERO_SKILL_BUFF_CHECK, HERO_ISOPEN_OPENDATE_CHECK |
| DropItem | DROP_ITEM_MUST_IN_ITEM_CHECK, DROP_ITEM_VALIDITY_CHECK, DATE_VALID_EXPIRE_CHECK, HERO_DROP_VALIDDATE_CHECK |
| Item | HERO_SYNTHESIS_CHECK |
| ActivityTask | ACTIVITY_TASK_REWARD_CHECK |
| DrawSkin | DRAWSKIN_BYPRODUCT_CHECK |
| Activity | DANQINGGE_TIME_ACTIVE_CHECK, DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK |
| DrawFix | DRAWFIX_PROTECTION_CHECK, DRAWFIX_ARENA_PROTECTION_CHECK |

匹配规则：精确匹配 sheetName，或后缀匹配 "中文|Key" 格式。

## 新增规则流程

### 新增表级规则

1. 在 `rule_def.go:10` 的 `ETableRule` 枚举中添加常量
2. 在 `coded_rules/table/` 目录实现规则(实现 `TableChecker` 接口)
3. 在 `check_manager/table_check_manager.go` 的 `init()` 注册元数据

### 新增列级规则

1. 在 `rule_def.go:205` 的 `EColRule` 枚举中添加常量
2. 在 `coded_rules/general/column_check/` 目录实现规则
3. 在 `check_manager/excel_check_factory.go` 注册检查器
