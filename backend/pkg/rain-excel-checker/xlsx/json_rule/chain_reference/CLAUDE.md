# chain_reference 目录文档

关系链检查（CHAIN_REFERENCE）的完整实现，包括数据结构、执行引擎、比较/匹配函数和洋葱模型。

## 职责

- **数据结构定义**：ChainStep/ChainConfig/ChainResult 等关系链核心类型
- **正向执行引擎**：ExecuteChain（单链前向执行）
- **洋葱模型引擎**：BuildOnionChain（左链前向 + 右链反向 + 两阶段门控）
- **比较/匹配函数**：CompareByType/MatchByType 及所有子类型
- **预警窗口**：ShouldSuppressByWarnBefore 逐行精确时间过滤
- **参数校验**：ValidateHandler 8 步校验

## 文件说明

### 基础设施

| 文件 | 说明 |
|------|------|
| `types.go` | ChainStep/ChainConfig/ChainValue/ChainResult/ChainPairConfig 数据结构及方法 |
| `engine.go` | ParseChainPairConfig 解析 + ExecuteChain 单链前向执行引擎 |
| `helpers.go` | ExtractByRegexFromString 正则提取 + FilterRowsByCondition 行过滤 |
| `compare.go` | CompareByType 比较分发 + 所有比较函数 + CompareTwoPhase + CompareTimeMatch |
| `match.go` | MatchByType 匹配分发 + 所有匹配函数 |
| `sheets.go` | ExtractChainStepSheets 从 chainSteps JSON 提取引用表名 |
| `warn_before.go` | ShouldSuppressByWarnBefore 逐行预警过滤 + ShouldSuppressWarnBeforeLegacy 旧路径全表过滤 |

### 洋葱模型

| 文件 | 说明 |
|------|------|
| `onion.go` | ChainHandler 接口、NextFunc 类型、ChainContext 数据上下文（含 WarnValues 预警字段） |
| `onion_builder.go` | BuildOnionChain 构建器、wrapHandler 闭包链机制 |
| `onion_validate.go` | ValidateHandler 参数校验（最外层，最先执行，含预警参数校验） |
| `onion_left_step.go` | LeftStepHandler 左链前向执行 + 预警时间提取（前置处理） |
| `onion_match.go` | MatchHandler 两阶段门控（左链完成后判断是否交汇） |
| `onion_right_step.go` | RightStepHandler 右链反向查找 + 预警时间提取（后置处理，核心创新） |
| `onion_compare.go` | CompareHandler 最终比较判定 + 预警窗口过滤（含 time_overlap 特化） |

### 测试

| 文件 | 说明 |
|------|------|
| `check_test.go` | 基础单元测试（ExecuteChain + ExtractChainStepSheets） |
| `onion_validate_test.go` | 参数校验测试（9 + 3 个预警参数场景） |
| `onion_left_test.go` | 左链 + 比较测试（15 个场景） |
| `onion_integration_test.go` | 洋葱模型集成测试（20 个场景） |
| `realdata_test.go` | DrawPet 真实数据测试（3 个场景） |
| `warn_before_test.go` | 预警窗口过滤测试（12 个场景） |

## 洋葱模型架构

### 嵌套结构

```
Validate → Compare → LeftStep0 → ... → LeftStepM → Match → RightStep0 → ... → RightStepN → terminal
```

执行顺序：LeftStep0 → ... → LeftStepM → Match → RightStepN → ... → RightStep0 → Compare → Validate

### 右链反向查找（核心创新）

传统方式：右链正向执行，第一步全表扫描，每行重复扫描同一张表。

洋葱模型：右链反向执行，利用 Match 的过滤结果精确查找：
- 反向查找：在 NextCol 列匹配上一步输出值 → 提取 PreCol 值
- 每步只需要处理匹配行，避免全表扫描

### 启用条件

`chainMatchCompare` 参数非空时使用洋葱模型路径，否则使用旧路径（向后兼容）。

### 预警窗口（chainWarnBefore）

当配置了 `chainWarnBefore` + `chainWarnSheet` + `chainWarnCol` 三参数时启用：

1. 链最后一步执行时，如果 `chainWarnSheet` 匹配当前步骤的 sheet，从匹配行的 `chainWarnCol` 列提取预警时间写入 `ChainContext.WarnValues`
2. 比较阶段 violation=true 后，检查 WarnValues 中最近的未来时间是否距今超过 chainWarnBefore
3. 超过则静默（不报错），在窗口内则正常报错

旧路径使用 `ShouldSuppressWarnBeforeLegacy` 全表扫描方案作为保守降级。

## 依赖

- `check_internal` — FindSheetBySuffix, GetColIndexByName, GetColValue 等工具函数
- `excel_internal` — MJS_FIXED_ROWS_NUM 常量
- `excelize/v2` — Excel 文件操作

## 外部调用

- `coded_rules/general/column_check/excel_check_chain_reference.go` — Check 方法委托调用
- `check_manager/filter.go` — filterSheetMapByRules 调用 ExtractChainStepSheets
