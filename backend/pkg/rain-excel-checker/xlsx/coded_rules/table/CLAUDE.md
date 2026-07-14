# coded_rules/table 目录文档

表级别校验规则实现，针对单个 Excel 表的特定业务规则。

## 规则列表

| 规则类型 | 说明 | 检查表 |
|----------|------|--------|
| `ARENA_SEASON_CHECK` | 竞技场赛季时间检查 | ArenaSeason |
| `ARENA_GENERAL_HERO_OPEN_CHECK` | 大将军武将开放时间检查 | ArenaScoreReward |
| `SEASON_PASS_HERO_OPEN_CHECK` | 战令武将开放时间检查 | SeasonPassReward |
| `DANQINGGE_TIME_ACTIVE_CHECK` | 丹青阁活动时间校验 | Activity |
| `COL_CONTINUOUS_CHECK` | 列连续性与唯一性检查（严格递增、单调递增、日期间隔、日期月模式、{id;count}格式、提取数字递增、拆分后全局唯一）。**支持同类型多实例** | 所有表 |
| `HERO_ISOPEN_OPENDATE_CHECK` | 战令/大将军武将IsOpen与OpenDate一致性检查（仅检查战令和大将军武将，普通武将不检查） | Hero |

## 文件说明

| 文件 | 说明 |
|------|------|
| `table_check_arena_season.go` | 竞技场赛季检查 |
| `table_check_arena_general_hero.go` | 大将军武将开放时间检查 |
| `table_check_season_pass_hero.go` | 战令武将开放时间检查 |
| `table_check_activity_danqingge_time.go` | 丹青阁活动时间校验 |
| `table_check_hero_isopen_opendate.go` | 战令/大将军武将IsOpen与OpenDate一致性检查（需跨表获取战令/大将军武将列表） |

## 规则参数

### ARENA_SEASON_CHECK

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `timeRangeBefore` | 提前警告时间（Go duration 格式，如 `168h` 表示7天） | `168h` |

**容错处理**：参数解析失败时自动使用默认值 `168h`（7天），不会中断检查。支持的格式示例：
- `168h`（168小时 = 7天）
- `720h`（720小时 = 30天）
- `1.5h`（1.5小时）
- `2h30m`（2小时30分钟）

### ARENA_GENERAL_HERO_OPEN_CHECK

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `WARN_DAYS_BEFORE` | 提前警告天数 | 30 |
| `OPEN_DATE_COL_NAME` | 武将开放时间列名 | OpenDate |

### SEASON_PASS_HERO_OPEN_CHECK

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `WARN_DAYS_BEFORE` | 提前警告天数 | 7 |

### DANQINGGE_TIME_ACTIVE_CHECK

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `timeRangeBefore` | 提前警告时间(duration 格式) | 168h(7天) |

### COL_CONTINUOUS_CHECK

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `targetCol` | 要检查连续性的列字段名（必填） | - |
| `checkMode` | INCREASE_STRICT/INCREASE_MONOTONE/DATE_CONTINUOUS/DATE_MONTHLY_PATTERN/ID_FORMAT_CONTINUOUS/EXTRACT_NUMBER_STRICT/SPLIT_UNIQUE | INCREASE_STRICT |
| `scope` | added(仅新增行) 或 all(全量) | all |
| `startValue` | 期望的起始值 | 自动检测 |
| `tolerance` | 日期模式容差天数 | 0 |
| `allowEmpty` | 跳过空值 | true |
| `allowCommit` | 跳过注释行 | false |
| `excludeRows` | 排除行号（如 "3,7" 或 "3-5"） | - |
| `separator` | SPLIT_UNIQUE 模式的分隔符 | , |

## 依赖表

| 规则 | 依赖表 |
|------|--------|
| `ARENA_SEASON_CHECK` | ArenaSeason |
| `ARENA_GENERAL_HERO_OPEN_CHECK` | ArenaScoreReward, Hero |
| `SEASON_PASS_HERO_OPEN_CHECK` | SeasonPassReward, SeasonPass, Hero |
| `HERO_ISOPEN_OPENDATE_CHECK` | Hero, SeasonPassReward, SeasonPass, ArenaScoreReward |
| `DANQINGGE_TIME_ACTIVE_CHECK` | Activity |

## 开发注意事项

- 表级规则需要检查规则是否适用于当前表(使用 `Meta().TargetSheets`)
- 时间解析使用 `excel_internal.ParseTime` 或 `excel_internal.ParseDate`
- 检查结果通过 `json_rule.TableCheckResult` 返回

## 丹青阁

服务端校验逻辑参考 `d:\work\server\app\src\services\lobby\user\useract\useractBase.go:30` 的 `CheckConfig` 方法。

### 勘误记录

**错误**: 初次实现时将丹青阁活动的检查目标表设为 `Draw.xlsx`，将 `ActivityType` 误判为 `Draw`。

**原因**: subagent 调研时从 `d:\work\config\docs\丹青阁活动配置指南.md` 中读取到丹青阁配置流程涉及 `Draw.xlsx`(抽卡池配置表)，将 Draw.xlsx 的表名错误等同于 ActivityType 的值。实际上 `Draw` 是 `ActivityPrefabType`(客户端 UI 预制类型)，而非 `ActivityType`(服务端活动类型枚举)。

**正确信息**(经服务端代码和用户确认):
- **目标表**: `Activity_活动表.xlsx`(非 Draw.xlsx)
- **ActivityType**: `ActTypeSkinRaffle`(枚举值 13)，服务端代码见 `useractManager.go:98` `case excel.EActivityType_ActTypeSkinRaffle`
- **时间字段**: `StartTime`、`EndTime`、`RewardTime`(TimeType=1 为绝对时间)
- **服务端校验逻辑**: `useractBase.go:30` CheckConfig 方法检查 `StartTimeStamp < nowTime < EndTimeStamp`