# coded_rules/general 目录文档

通用表级别校验规则实现，适用于所有表的通用功能。

## 规则列表

| 规则类型 | 说明 | 实现文件 |
|----------|------|----------|
| `NEW_ROW_NOTIFY` | 新增行/列通知(Ok=true，旧版保留兼容) | `table_check_new_row_notify.go` |
| `ROW_CHANGE_NOTIFY` | 行变更字段通知(Ok=true，旧版保留兼容) | `table_check_row_change_notify.go` |
| `ADDED_ROW_NOTIFY` | 新增行通知(Ok=true) | `table_check_added_row_notify.go` |
| `REMOVED_ROW_NOTIFY` | 删除行通知(Ok=true) | `table_check_removed_row_notify.go` |
| `ADDED_COL_NOTIFY` | 新增列通知(Ok=true) | `table_check_added_col_notify.go` |
| `REMOVED_COL_NOTIFY` | 删除列通知(Ok=true) | `table_check_removed_col_notify.go` |
| `MODIFIED_ROW_NOTIFY` | 修改行通知(Ok=true) | `table_check_modified_row_notify.go` |

## 规则参数

### NEW_ROW_NOTIFY

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `GIT_REPO_PATH` | Git 仓库路径 | Excel 文件所在目录 |
| `BASE_COMMIT` | 对比的基准 commit | HEAD~1 |
| `ID_COL_NAME` | ID 列名 | Id |
| `NAME_COL_NAME` | 名称列名 | Name |
| `NOTIFY_ADDED_ROWS` | 是否通知新增行 | true |
| `NOTIFY_REMOVED_ROWS` | 是否通知删除行 | true |
| `NOTIFY_ADDED_COLS` | 是否通知新增列 | true |
| `NOTIFY_REMOVED_COLS` | 是否通知删除列 | true |

### ROW_CHANGE_NOTIFY

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `GIT_REPO_PATH` | Git 仓库路径 | Excel 文件所在目录 |
| `BASE_COMMIT` | 对比的基准 commit | HEAD~1 |
| `ID_COL_NAME` | ID 列名 | Id |
| `NAME_COL_NAME` | 名称列名 | Name |

输出格式：`{行名}行，id为{ID}，字段{列名}从{旧值}改成了{新值}`

## 返回数据结构

通知规则返回的 `TableCheckResult.ErrCells` 中 `CellError.Detail` 为结构化数据：

| Detail 类型 | 说明 | 定义位置 |
|-------------|------|----------|
| `RowChangeDetail` | 行变更(新增/删除)，字段：changeType/rowId/rowName | json_rule/rule_def.go:152 |
| `FieldChangeDetail` | 字段变更(修改)，字段：rowId/rowName/colName/oldValue/newValue | json_rule/rule_def.go:160 |
| `ColumnChangeDetail` | 列变更(新增/删除)，字段：changeType/colName | json_rule/rule_def.go:169 |

数据流向：规则实现 -> JSON 序列化 -> 前端/MCP(使用 Detail)/ 飞书/命令行(使用 Reason)

## Git Diff 模式

变更检测规则使用 Git Diff 模式：

- 不再依赖本地快照文件
- 直接对比 Git commit 之间的 Excel 文件差异
- 更适合 CI/CD 流水线场景

### 工作原理

```
当前版本 Excel                     上一个 commit
D:/work/config/excel/Hero.xlsx  →  git show HEAD~1:Hero.xlsx
        │                                   │
        │ excelize.OpenFile()               │ git show + excelize.OpenReader()
        ▼                                   ▼
   currentCols                       oldCols ([]byte 解析)
        │                                   │
        └───────────┬───────────────────────┘
                    │ BuildSnapshot() + DetectDiff()
                    ▼
            ExcelDiffResult
            - AddedRows
            - RemovedRows
            - ModifiedRows
            - AddedCols
            - RemovedCols
```

## 功能说明

### 新增行/列通知

首次运行时只保存快照，不触发通知。后续运行时：
- 检测新增的行
- 检测删除的行
- 检测新增的列
- 检测删除的列

### 行变更字段通知

检测字段值的变化：
- 按行分组记录每个字段的变更(原值 → 新值)
- 输出格式(飞书结构化卡片)：
  ```
  📋 工作表变更通知 - {表名}

  📁 表格名称: {表名}
  🔄 变更类型: 修改行
  变更范围: 共 N 行记录发生变更
  🔑 关键标识: {ID列名} (第一列)

  【变更记录 N】
  主键值: {ID}
  行号: 第 N 行
  字段变更明细:
    ✏️ {列名}: [原值] {旧值} → [新值] {新值}

  ⏰ 变更时间: {时间}
  👤 提交人: {作者}
  🔗 对比版本: {baseCommit} → {headCommit}
  ```

### 新增/删除行/列通知格式

通知消息使用结构化卡片格式，包含标题、总览、预览表格和元数据：
- 新增行：预览表格(ID/名称/其他列)
- 删除行：已删除记录表格 + 影响范围 + 建议
- 新增列：列名预览
- 删除列：已删除列详情 + 影响范围 + 建议

## 依赖

- 依赖 `check_internal` 模块的快照管理功能
- 快照文件命名：`{SheetName}_snapshot.json`
- 详细实现见 [check_internal/CLAUDE.md](../../check_internal/CLAUDE.md)

## 共享模块

| 模块 | 文件 | 说明 |
|------|------|------|
| 通知格式化 | `notify_format.go` | 所有通知规则的格式化函数 |
| Diff缓存 | `check_internal/diff_cache.go` | 共享的 git diff 结果缓存，避免重复计算 |

## 开发注意事项

- 通知规则不算错误，`result.Ok` 仍为 `true`
- 首次运行(上一个 commit 不存在该文件)不触发通知，只记录当前状态
- 支持"中文|英文"格式的 Sheet 名称匹配
