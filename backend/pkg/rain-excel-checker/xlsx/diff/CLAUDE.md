# diff 包 — Excel 配表差异检测与上下文管理

专为游戏配表增量检查设计的差异检测引擎，将 Git 历史版本对比与 Excel 差异检测结合。

## 文件结构

```
diff/
├── adaptor.go      # 上下文管理器、列规则适配器
├── diff_cache.go   # 差异结果缓存层（避免重复 git show + Excel 解析）
└── excel_diff.go   # Excel 差异检测核心逻辑（快照构建、行对比、列变化检测）
```

## 核心组件

### excel_diff.go — 差异检测核心

| 导出类型 | 说明 |
|----------|------|
| `ExcelSnapshot` | Excel 数据快照（按 ID 列索引的行数据映射） |
| `RowSnapshot` | 单行快照（ID、名称、行号、原始数据） |
| `ExcelDiffResult` | 差异检测结果（新增/删除/修改行、字段变更明细） |
| `RowChange` | 行变更信息（ID、名称、行号、变更类型、字段变更列表） |
| `FieldChange` | 字段变更（列名、旧值、新值） |
| `ColDiffResult` | 列变化检测（新增列、删除列） |

| 导出函数 | 说明 |
|----------|------|
| `BuildSnapshot` | 按单列 ID 构建快照 |
| `BuildSnapshotWithCompositeKey` | 按多列复合主键构建快照（`\x01` 分隔） |
| `DetectDiff` | 单列 ID 差异检测 |
| `DetectDiffWithCompositeKey` | 复合主键差异检测 |
| `DetectColChanges` | 列增删检测 |
| `ParseExcelFromBytes` | 从字节流解析 Excel（用于 git show 获取的历史版本） |
| `FormatDiffNotification` | 格式化差异通知文本 |

**关键行为**：
- 自动检测枚举表格式（`name`/`value` 开头的列）
- 使用 `helpers.AutoDetectEndIndex` 实现"连续 3 个空行注释"截断规则
- ID 列不存在时自动降级到第一列
- 字段变更按 Excel 原始列顺序排序

### diff_cache.go — 差异结果缓存

| 导出类型 | 说明 |
|----------|------|
| `DiffCacheKey` | 缓存键（ExcelPath + BaseCommit + SheetName + IdColKey + NameColName） |
| `DiffCacheEntry` | 缓存条目（DiffResult + GitCtx + IsNewFile + Err） |
| `GitNotifyContext` | Git 提交元数据（committer、commitTime、hash 等） |
| `DiffComputeParam` | 差异计算参数（git 路径、commit、sheet、cols 等） |

| 导出函数 | 说明 |
|----------|------|
| `GetOrComputeDiff` | 带缓存的差异计算（核心入口） |
| `ClearDiffCache` | 清理缓存 |
| `PrefillDiffCache` | 预填充缓存（单元测试用） |
| `BuildDiffCacheKey` | 构建规范化缓存键 |
| `ParseDiffParams` | 从规则参数解析差异检测所需字段 |
| `ResolveExcelPath` | 从 sheetMap 定位 Excel 文件路径 |
| `BuildDiffComputeParam` | 构建完整的差异计算参数 |

**缓存策略**：`sync.Map` 包级缓存，多个通知规则共享同一份 DiffResult，避免重复执行 git show + Excel 解析。

### adaptor.go — 上下文管理与适配器

| 导出类型 | 说明 |
|----------|------|
| `ContextAdaptor` | 上下文适配器接口（Store/Get/Clear/Stats） |
| `CheckContext` | 单次检查的完整上下文（参数、列名、sheet 名、colIndex 等） |
| `ColRuleAdapter` | 列规则适配器（将简洁接口适配到 Checker 接口） |
| `ColRule` | 列规则接口（CheckCol 方法） |
| `AdaptorStats` | 适配器统计（存储数、获取数、清理数） |

| 导出变量/函数 | 说明 |
|---------------|------|
| `GlobalAdaptor` | 全局适配器实例 |
| `NewColRuleAdapter` | 创建列规则适配器 |

**上下文 key 格式**：`"sheetName:colName:uuid"`，使用原子计数器生成 UUID，确保并发安全。自动清理机制通过 `context.WithValue` + `defer cancel()` 防止内存泄漏。

## 包依赖

### 依赖
- `gitutil` — Git 操作（获取历史文件、提交信息）
- `helpers` — 通用辅助（自动检测结束行、列索引查找）
- `excelio` — Excel 解析、常量（MJS_FIXED_ROWS_NUM 等）
- `json_rule` — 规则参数结构

### 被依赖
- `engine` — 上下文管理（`GlobalAdaptor`）和缓存清理（`ClearDiffCache`）
- `coded_rules/general` — 所有通知规则使用差异检测和缓存
- `coded_rules/general/column_check/datatype` — `cell_type_check_adaptor` 获取检查上下文

## 注意事项

- **循环依赖**：`coded_rules` 包不要直接导入 `adaptor.go`，已有注释标注
- **复合主键分隔符**：使用不可见字符 `\x01`，避免与数据内容冲突
- **枚举表检测**：自动识别 `name`/`value` 开头的列名模式
- **空行截断**：`BREAK_LINE = 3`（连续 3 个空行视为注释区，停止快照构建）
