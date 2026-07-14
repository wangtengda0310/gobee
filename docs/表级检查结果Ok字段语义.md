# 表级检查结果 Ok 字段语义

## 概述

`Ok` 字段是 `TableCheckResult` 结构体的核心字段，用于区分**错误检测规则**和**通知规则**的检查结果类型。

## 字段定义

定义位置：`rain-excel-checker/xlsx/json_rule/rule_def.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| `Ok` | `bool` | 检查结果类型 |
| `Reason` | `string` | 结果说明 |
| `ErrCells` | `[]*CellError` | 错误/通知详情 |

## Ok 字段的语义

### `Ok = false`：检查失败（错误）

**含义**：检测到需要修复的问题或错误

**规则类型**：错误检测规则

**示例**：
- `ARENA_SEASON_CHECK` - 竞技场赛季已结束/即将结束
- `ARENA_GENERAL_HERO_OPEN_CHECK` - 武将开放时间不足
- `SEASON_PASS_HERO_OPEN_CHECK` - 战令武将开放时间不足

**输出表现**：
- 前端 UI：红色 ✗ 标签
- 命令行日志：`❌ 表级检查失败`
- 飞书通知：红色错误卡片

**代码示例**：`[table_check_arena_season.go](rain-excel-checker/xlsx/coded_rules/table/table_check_arena_season.go)`

### `Ok = true`：检查通过或仅为通知（非错误）

**含义**：
1. 检查通过，没有发现问题
2. **仅为通知**，记录变更信息但不视为错误

**规则类型**：
- 通知规则（`NEW_ROW_NOTIFY`、`ROW_CHANGE_NOTIFY`）

**输出表现**：
- 前端 UI：绿色 ✓ 标签 + 变更详情
- 命令行日志：**会输出**（修复完成）
- 飞书通知：**会发送**（修复完成）

**代码示例**：`[table_check_new_row_notify.go](rain-excel-checker/xlsx/coded_rules/general/table_check_new_row_notify.go)`

## ErrCells 与 Ok 的关系

| Ok 值 | ErrCells 含义 | 示例 |
|-------|--------------|------|
| `false` | **错误列表** - 需要修复的问题 | 赛季结束时间、武将开放时间不足 |
| `true` | **变更记录** - 仅通知，不视为错误 | 新增行、删除行、字段修改 |

**关键点**：`Ok=true` 的通知规则 **ErrCells 也可能有内容**！

## 前端 UI 显示逻辑

位置：`frontend/src/pages/excel-test/components/excel-check-log.vue:89-110`

- **Ok=true**：绿色 ✓ 标签，显示 DisplayName + Reason（不展开 ErrCells）
- **Ok=false**：红色 ✗ 标签，显示 DisplayName + Reason + 展开详细 ErrCells 列表

## 命令行日志输出逻辑

位置：`feishu-lib/notification/formatter.go:162-210`

- `FormatTableErrors` 只处理 Ok=false 的结果
- `FormatConsoleOutput` 同时输出错误和通知（修复后）

## 飞书通知逻辑

飞书通知发送条件：
- `HasErrors = true`（有错误）→ 发送红色卡片，显示错误详情
- `HasNotifications = true`（有通知）→ 发送红色卡片，显示变更通知

- **Ok=false 的错误**：会发送飞书卡片通知
- **Ok=true 的通知**：**也会发送飞书通知**（修复完成）

## 设计思路

### 为什么通知规则 Ok=true？

1. **语义清晰**：通知不是错误，需要区分处理
2. **前端友好**：可以用不同的视觉样式区分错误和通知
3. **业务合理**：配表变更需要通知给策划，同时错误也需要及时处理
4. **统计准确**：通知不应计入错误数，但需要独立统计

### 输出通道设计

| 通道 | 错误处理 | 通知处理 |
|------|----------|----------|
| 前端 UI | 红色 ✗，展开 ErrCells | 绿色 ✓，显示 Reason |
| 命令行 | `❌ 表级检查失败` | `📝 表级变更通知` |
| 飞书推送 | 红色卡片，显示 ErrCells | 红色卡片，显示 Reason |
| 错误统计 | 计入 TotalErrors | 不计入，独立统计 TableNotifications |


## 相关文件

| 文件 | 说明 |
|------|------|
| `rain-excel-checker/xlsx/json_rule/rule_def.go` | TableCheckResult 结构定义 |
| `rain-excel-checker/xlsx/coded_rules/general/table_check_new_row_notify.go` | 新增行/列通知规则 |
| `rain-excel-checker/xlsx/coded_rules/general/table_check_row_change_notify.go` | 行变更字段通知规则 |
| `rain-excel-checker/xlsx/coded_rules/table/table_check_arena_season.go` | 竞技场赛季检查规则（错误检测） |
| `rain-qa-func/frontend/src/pages/excel-test/components/excel-check-log.vue` | 前端检查结果显示组件 |
| `feishu-lib/notification/formatter.go` | 命令行日志格式化器 |
