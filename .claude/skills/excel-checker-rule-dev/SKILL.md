---
name: excel-checker-rule-dev
description: |
  rain-excel-checker 校验规则开发全流程指导 — 涵盖新规则创建、现有规则修改、调试排错、单元测试和文档同步。

  触发条件（满足任一即触发）：
  - 用户要求为 rain-excel-checker 新增校验规则、检查规则、验证规则
  - 用户提到"增加校验"、"新增规则"、"补充规则"、"加一条检查"
  - 用户描述具体的配表校验需求，如"XX表的时间必须和YY表一致"、"XX字段必须在YY表中存在"
  - 用户要求修改、增强、修复现有的校验规则
  - 用户提到 DANQINGGE、战令、大将军、武将、掉落、配表检查、Excel 校验等关键词
  - 用户要求为校验规则编写或补充测试
  - 用户在校验规则开发过程中遇到 bug 或行为异常

  不要触发的场景：
  - 仅查看/查询配表数据内容（用 excel-parser skill）
  - 仅分析配表结构（用 game-config-analyzer skill）
  - 与 rain-excel-checker 无关的 Go 开发
---

# rain-excel-checker 校验规则开发指南

本技能指导 AI 完成 rain-excel-checker 中校验规则的全生命周期开发：从需求分析到实现、测试、调试和文档更新。

## 一、项目架构速览

```
rain-excel-checker/
├── main.go                              # 入口：merge commit 检测与遍历检查
├── gitutil/                             # Git 操作（commit 检测、diff）
├── src/                                 # 顶层流程控制（flow、load、result）
└── xlsx/
    ├── json_rule/
    │   ├── rule_def.go                  # ETableRule 常量、CheckParam、TableRuleMeta 等数据结构
    │   └── default_table_rules.go       # 默认表级规则映射 + GeneralRuleOverrides
    ├── check_manager/
    │   ├── table_checker_def.go         # TableChecker 接口定义
    │   ├── table_check_manager.go       # 表级规则注册（init 函数）
    │   ├── manager_def.go               # CheckAll/CheckWithFilter/CheckWithGitHistory 入口
    │   └── excel_check.go               # 列级检查执行
    ├── coded_rules/
    │   ├── cross_table/                 # 跨表级别规则（package coded_rules）
    │   ├── table/                       # 单表业务规则（package coded_rules）
    │   └── general/                     # 通用规则（通知等）+ column_check/ 列级
    ├── check_internal/                  # 工具函数、快照管理、差异检测
    │   ├── utils.go                     # FindSheetBySuffix、AutoDetectEndIndex、GetDataEndIndex
    │   ├── hero_rule_helper.go          # GetColIndexByName、GetColValue、武将/时间辅助
    │   ├── rule_helpers.go              # ParseIntWithError 等规则辅助
    │   ├── excel_diff.go                # Git 差异检测（快照构建、差异对比）
    │   └── diff_cache.go               # Git diff 缓存
    └── excel_internal/                  # Excel 解析、常量、ParseDate
```

### 核心接口

```go
// 表级检查器接口（所有表级规则必须实现）
// 定义于 check_manager/table_checker_def.go
type TableChecker interface {
    Check(param json_rule.CheckParam) *json_rule.TableCheckResult
    Meta() *json_rule.TableRuleMeta
}

// CheckParam 检查参数
// 定义于 json_rule/rule_def.go
type CheckParam struct {
    SheetName   string                    // 表名
    Cols        [][]string                // 列主序数据（cols[colIdx][rowIdx]）
    StartRowIdx int                       // 数据起始行（通常为 4）
    EndIndex    int                       // 数据结束行索引（不含），由调用方计算后传入
    Params      map[string]string         // 规则参数
    SheetMap    map[string]*excelize.File // 跨表数据（用于跨表检查）
    Now         time.Time                 // 注入当前时间（零值使用 time.Now()）。仅用于单元测试注入固定时间
}
```

**关键要点**：

- `EndIndex` 由框架在调用 `Check()` 之前计算好并传入，规则实现中**直接使用 `param.EndIndex`**，不需要自行调用 `AutoDetectEndIndex`
- 遍历循环使用 `for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++`（注意是 `<` 不是 `<=`）
- `Now` 字段用于时间相关规则：零值时框架自动使用 `time.Now()`，单元测试中可注入固定时间避免 flaky test
- 获取当前时间应使用 `check_internal.ResolveNow(param.Now)` 而非直接 `time.Now()`

### Excel 行结构常量

| 行号 | 索引 | 内容 | 常量 |
|------|------|------|------|
| 第 1 行 | 0 | 中文名 | `MJS_FIXED_ROWS_NAME = 0` |
| 第 2 行 | 1 | 数据类型 | — |
| 第 3 行 | 2 | **字段名**（`GetColIndexByName` 查找此行） | — |
| 第 4 行 | 3 | 导出标记 | — |
| 第 5 行+ | 4+ | **数据区** | `MJS_FIXED_ROWS_NUM = 4`（即 `startRowIdx = 4`） |

### 数据访问方式

`Cols` 是**列主序**二维数组：`cols[colIdx][rowIdx]`

```go
// 查找列索引（查找第3行，即索引2的字段名）
colIdx := check_internal.GetColIndexByName(cols, "FieldName")
// 安全获取值（越界返回空字符串）
value := check_internal.GetColValue(cols, colIdx, rowIdx)
// 查找其他表（支持后缀匹配："Hero" 可匹配 "武将|Hero"）
file, sheetName, ok := check_internal.FindSheetBySuffix(sheetMap, "TableName")
```

## 二、需求分析阶段

收到用户的规则需求后，**先查记忆再动手**：

```
1. 调用 mempalace_search 查询：
   - 关键词：目标表名、活动类型、关联表组合
   - 检查是否有已确认的字段名映射、数据格式陷阱、关联模式记录
2. 调用 claude-mem 查询（通过 /mem search 或查看历史摘要）：
   - 搜索：该表/该活动类型的规则开发历史、字段名变更记录
   - 检查近期会话中是否有已验证的关联逻辑或数据格式说明
3. 调用 forgetful query_memory 查询：
   - 搜索：检查引擎架构、规则清单、辅助函数速查
   - 检查是否有已记录的"规则设计模式"或"配表问题模式"
- 查看 .learnings/LEARNINGS.md 中是否有同类配表问题模式
- 如历史记录中有已验证的关联逻辑，直接复用而非重新推导
```

**再弄清以下问题**：

### 2.1 规则分类

| 规则类型 | 放置目录 | package 名 | 何时选择 |
|----------|----------|-----------|----------|
| **跨表规则** | `coded_rules/cross_table/` | `coded_rules` | 需要读取其他表的数据来验证 |
| **单表规则** | `coded_rules/table/` | `coded_rules` | 只涉及当前表内部的逻辑 |
| **列级规则** | `coded_rules/general/column_check/` | `column_check` | 通用的列数据格式检查 |

> **注意**：`cross_table/` 和 `table/` 目录下的 Go 文件 package 名都是 `coded_rules`（不是 `cross_table` 或 `table`）。

### 2.2 必须向用户确认的信息

向用户确认以下信息，缺少任何一条都不应开始编码：

1. **源表**：规则检查哪张表？（如 Activity、Hero、DropItem）
2. **目标字段**：检查哪些列？（如 StartTime、EndTime、CustomParma）
3. **验证逻辑**：具体的校验条件是什么？
   - 时间比较：精确到秒？还是只到天？
   - 存在性检查：在哪个表中查找？匹配哪个字段？
   - 枚举检查：合法值有哪些？
4. **依赖表**（跨表规则）：需要读取哪些其他表？
5. **错误级别**：是报错（`Ok=false`）还是警告/通知（`Ok=true`）？
6. **过滤条件**：是否只检查特定类型的行？（如 `ActivityType == "ActTypeSkinRaffle"`）

### 2.3 数据核对 — 查看实际 Excel 表格数据

**在确认需求信息后、动手编码前，必须查看实际 Excel 数据验证理解的正确性。** 仅凭用户描述开发规则是高风险行为——字段名记错、数据格式理解偏差、隐藏的边界情况都会导致规则实现错误。

#### 核对步骤

1. **定位并打开相关表格**
   - 读取 `d:\work\config` 目录下对应的 xlsx 文件
   - 使用 excel-parser skill 或直接读取文件

2. **核对字段名是否存在**
   - 用户说的字段名（如 `CustomParma`）在表头中是否真的存在？
   - 是否有同义字段名（如 `StartTime` vs `BeginTime`）？
   - 字段名的拼写、大小写是否与用户描述一致？

3. **观察实际数据格式**
   - 数据是纯数字、字符串、还是 `{id;count}` 格式？
   - 是否有空值？空值的含义是什么（未配置 vs 不适用）？
   - 是否有特殊值（如 `-1`、`0`、`NULL`）？
   - 数值范围是什么样的？

4. **验证跨表关系**
   - 如果涉及跨表检查，实际去源表和目标表中确认关联字段的数据是否能对应上
   - 例如：用户说"A 表的 HeroId 对应 B 表的 Id"，实际查看是否有对应不上的数据

5. **统计有效数据量**
   - 查看实际有多少行数据、多少种不同的值
   - 评估规则的检查范围（是检查全部行还是只检查特定类型？）

#### 质疑策略 — 主动发现矛盾

查看数据后，如果发现以下情况，**必须立即停下来向用户质疑**，不要假设：

| 发现的问题 | 质疑方式 |
|-----------|---------|
| 用户说的字段名在表中不存在 | "我查看了 XX 表，没有找到名为 YY 的列，实际列名是 ZZ，请确认是否是这列？" |
| 数据格式与用户描述不符 | "您描述 XX 字段是纯数字，但实际数据中我看到 `{1;100}{2;200}` 格式，请确认解析方式" |
| 存在大量空值或特殊值 | "XX 字段有约 30% 的行为空，是否应该跳过这些行？还是应该报错？" |
| 跨表关联数据对不上 | "A 表中 ID=1001 的记录引用了 B 表中的 ID=999，但 B 表中不存在此 ID，这是预期行为还是配置错误？" |
| 数据与规则逻辑矛盾 | "您说'结束时间必须大于开始时间'，但我看到第 15 行数据的结束时间早于开始时间，是否是已知问题？" |
| 过滤条件不够精确 | "您说检查 ActivityType=X 的行，但数据中我发现还有 ActivityType=Y 也包含类似配置，是否也需要检查？" |

**原则**：宁可多问一句，也不要基于错误假设写出一整条规则。质疑时附带具体数据（行号、值），让用户能快速定位和确认。

### 2.4 参考已有规则

在动手前，先阅读现有规则中**最相似的实现**作为参考模板。常见场景对应表：

| 需求场景 | 参考规则文件 |
|----------|-------------|
| 时间匹配（精确到秒） | `table/table_check_season_pass_hero.go` |
| 时间预警（即将到期） | `table/table_check_activity_danqingge_time.go` |
| 跨表存在性检查 | `cross_table/table_check_drop_item_must_in_item.go` |
| 武将相关查询 | `check_internal/hero_rule_helper.go` 中的辅助函数 |
| 字段值解析（{id;count}格式） | `cross_table/table_check_drawskin_byproduct.go` |
| 按名称/类型过滤行 | `table/table_check_arena_general_hero.go` |
| 数据有效性检查（多条件） | `table/table_check_drop_rule_data_validity.go` |
| 条件引用检查 | `cross_table/table_check_drop_rule_conditional.go` |
| 时间范围/交集检查 | `cross_table/table_check_activity_drawskin_time_overlap.go` |
| 日期有效性检查（简单单表） | `table/table_check_date_valid_expire.go` |

### 2.5 TargetSheets 挂载原则

规则的 `TargetSheets` 决定增量模式下哪些表变更会触发该规则。**必须挂载在核心检查字段所在的表上**，否则增量检查不会触发。

判断方法：规则检查的核心字段在哪个表，就挂载到哪个表。跨表读取的辅助数据放入 `RequiredSheets`。

| 场景 | 正确挂载 | 原因 |
|------|---------|------|
| 检查 Item.IsSynthetic | **Item** | 核心字段在 Item，改 Item 应触发 |
| 检查 Hero.CanMelt | **Hero** | 核心字段在 Hero |
| 检查 Hero.OpenDate vs 赛季时间 | **Hero** | 核心字段在 Hero |

## 三、新规则创建流程

### 前置步骤：使用测试技能设计测试用例（测试先行）

**在开始编码之前**，必须使用 `test-case-generator` 技能设计测试用例。这是 TDD 思想在规则开发中的具体应用——先定义"什么是正确的"，再实现"如何检查正确性"。

#### 测试设计流程

1. **调用 test-case-generator 技能**
   - 基于需求分析阶段（第二节）收集的信息和对实际数据的观察
   - 明确告知技能：这是 Excel 检查规则的测试，使用 Go + testify/assert

2. **设计测试用例清单**（只输出清单，不写代码）
   - 基于实际数据中的真实值构造测试数据
   - 覆盖以下场景（至少）：

   | 测试场景 | 优先级 | 设计要点 |
   |---------|--------|---------|
   | 正向数据全部通过 | P0 | 使用与实际数据格式一致的合法数据 |
   | 无效数据报错 | P0 | 针对每条校验规则设计一个违反的用例 |
   | 空值/缺失值处理 | P1 | 明确空值是跳过还是报错 |
   | 边界值 | P1 | 最小值、最大值、临界条件 |
   | 依赖表不存在 | P1 | 跨表规则必须测试 |
   | 字段缺失 | P2 | 关键列不存在时的降级处理 |
   | 过滤条件 | P2 | 验证只检查符合条件的行 |

3. **提交测试用例清单给用户审核**
   - 以表格形式呈现，包含场景、输入数据、预期结果
   - 用户确认后再进入编码步骤

4. **审核通过后开始编码**，按以下步骤 1-5 执行

### 编码步骤（步骤 1-5）

必须**严格按顺序**完成以下 5 个步骤。遗漏任何步骤都会导致规则无法正常工作。

### 步骤 1：添加规则类型常量

文件：`json_rule/rule_def.go`

在 `ETableRule` 枚举中添加（约第 29 行之后的 const 块中）：

```go
YOUR_RULE_CHECK ETableRule = "YOUR_RULE_CHECK" // 规则中文描述
```

命名规范：`大写蛇形_CHECK`，简洁且有描述性。

### 步骤 2：创建规则实现

根据规则类型选择目录：
- 跨表规则 → `coded_rules/cross_table/table_check_{描述}.go`
- 单表规则 → `coded_rules/table/table_check_{描述}.go`

> **重要**：两个目录下的 package 名都是 `coded_rules`。

#### 实现模板

```go
package coded_rules

import (
    "fmt"

    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-excel-checker/xlsx/check_internal"
    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-excel-checker/xlsx/excel_internal"
    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-excel-checker/xlsx/json_rule"
    "github.com/xuri/excelize/v2"
)

// YourRuleCheckRule 规则中文描述
//
// ## 校验规则
// 1. 具体规则描述
//
// ## 相关表结构
// - SourceTable: 列A, 列B
// - DependTable（可选）: 列C, 列D
type YourRuleCheckRule struct{}

// Meta 返回规则元数据
func (c *YourRuleCheckRule) Meta() *json_rule.TableRuleMeta {
    return &json_rule.TableRuleMeta{
        Type:           json_rule.YOUR_RULE_CHECK,
        DisplayName:    "规则中文显示名",
        Description:    "规则详细描述（用于飞书消息和MCP展示）",
        TargetSheets:   []string{"SourceTable"},    // 源表（支持后缀匹配）
        RequiredSheets: []string{"DependTable"},     // 跨表依赖（仅跨表规则需要）
        ParamDefs:      []json_rule.TableRuleParamDef{},
    }
}

// Check 执行检查
func (c *YourRuleCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
    // 1. 初始化结果
    result := &json_rule.TableCheckResult{
        Ok:          true,
        SheetName:   &param.SheetName,
        DisplayName: "规则中文显示名",
        ErrCells:    make([]*json_rule.CellError, 0),
    }

    // 2. 查找源表关键列
    targetColIdx := check_internal.GetColIndexByName(param.Cols, "TargetColumn")
    if targetColIdx == -1 {
        result.Ok = true // 列不存在时通常跳过检查，而非报错
        result.Reason = "未找到 TargetColumn 列，跳过检查"
        return result
    }

    // 3. 加载依赖表（跨表规则）
    dependFile, dependSheetName, ok := check_internal.FindSheetBySuffix(param.SheetMap, "DependTable")
    if !ok {
        result.Ok = false
        result.Reason = "未找到 DependTable 表，无法执行检查"
        return result
    }
    dependCols, err := dependFile.GetCols(dependSheetName)
    if err != nil {
        result.Ok = false
        result.Reason = fmt.Sprintf("读取 DependTable 失败: %v", err)
        return result
    }

    // 4. 构建验证数据
    validSet := make(map[int]bool) // 或其他合适的数据结构
    // ... 构建逻辑
    // 从其他表构建验证数据时，使用 GetDataEndIndex 获取结束位置：
    // endIdx := check_internal.GetDataEndIndex(dependCols, param.StartRowIdx)

    // 5. 遍历检查（使用 param.EndIndex，不要自行计算）
    for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
        value := check_internal.GetColValue(param.Cols, targetColIdx, rowIdx)
        if value == "" {
            continue // 跳过空行
        }
        // ... 校验逻辑
        if !validSet[someId] {
            result.ErrCells = append(result.ErrCells, &json_rule.CellError{
                Index:  rowIdx,
                Reason: fmt.Sprintf("具体错误描述，包含足够定位信息：行ID=%s, 值=%s", rowId, value),
            })
        }
    }

    // 6. 设置最终状态
    if len(result.ErrCells) > 0 {
        result.Ok = false
        result.Reason = fmt.Sprintf("发现 %d 个问题", len(result.ErrCells))
    }
    return result
}
```

#### 关键注意事项

- **package 名必须是 `coded_rules`**（cross_table/ 和 table/ 目录均如此）
- **错误信息必须使用中文**，包含足够定位信息（表名、行ID、具体值）
- **空值处理**：空单元格不应报错，用 `continue` 跳过
- **越界安全**：始终使用 `check_internal.GetColValue` 而非直接索引
- **依赖表缺失**：返回 `Ok=false` 并说明原因，**不要 panic**
- **列不存在**：通常应跳过检查（`Ok=true`）而非报错，除非该列是规则的核心前提
- `FindSheetBySuffix` 支持后缀匹配：传入 `"Hero"` 可匹配 `"武将|Hero"`
- **时间相关规则**：使用 `check_internal.ResolveNow(param.Now)` 获取当前时间，而非 `time.Now()`
- **日期解析**：使用 `excel_internal.ParseDate(str)` 解析日期字符串

### 步骤 3：注册检查器

文件：`check_manager/table_check_manager.go`

确认文件顶部 import 中已有对应的 alias：
- `crosstable ".../coded_rules/cross_table"`
- `tablecheck ".../coded_rules/table"`

在 `init()` 函数末尾添加：

```go
// 你的规则描述
TableManager.Reg(json_rule.YOUR_RULE_CHECK, new(crosstable.YourRuleCheckRule))
// 或 new(tablecheck.YourRuleCheckRule)（单表规则）
```

注册时会自动将 `Meta()` 返回的元数据同步到 `json_rule.TableRuleMetas`。

### 步骤 4：配置默认规则（最容易遗漏！）

文件：`json_rule/default_table_rules.go`

在 `DefaultTableRules` map 中，找到目标表对应的条目，追加新规则：

```go
"SourceTable": {EXISTING_RULE, YOUR_RULE_CHECK},  // 追加到已有数组
```

如果目标表还没有条目，新建一行：

```go
"SourceTable": {YOUR_RULE_CHECK},
```

**如果跳过这一步，规则虽然注册了但不会自动执行，飞书消息中看不到任何输出。**

### 步骤 5：编写单元测试

测试文件与实现文件放在**同一目录**，命名为 `table_check_{描述}_test.go`。

#### 测试必须覆盖的场景

| 场景 | 测试名模式 | 预期结果 |
|------|-----------|----------|
| 全部有效数据 | `TestXxx_AllValid` | `Ok=true, ErrCells 为空` |
| 无效数据应报错 | `TestXxx_InvalidData` | `Ok=false, ErrCells 非空` |
| 依赖表不存在 | `TestXxx_NoDependSheet` | `Ok=false, Reason 含"未找到"` |
| 空数据 | `TestXxx_EmptyData` | `Ok=true` |
| 关键列缺失 | `TestXxx_MissingColumn` | `Ok=true, Reason 含"跳过"` |
| 时间注入（时间相关规则） | `TestXxx_WithFixedTime` | 使用 `Now` 字段注入固定时间 |

#### 测试数据构造

```go
func setupTestExcel(headers []string, dataRows [][]string) (*excelize.File, string) {
    f := excelize.NewFile()
    sheetName := "测试表|TestTable"
    _, _ = f.NewSheet(sheetName)

    // 行 1-2: 空
    // 行 3（索引 2）: 列名
    for i, h := range headers {
        col := string(rune('A' + i))
        _ = f.SetCellValue(sheetName, col+"3", h)
    }
    // 行 4（索引 3）: 类型标记
    for i := range headers {
        col := string(rune('A' + i))
        _ = f.SetCellValue(sheetName, col+"4", "string")
    }
    // 行 5+（索引 4+）: 数据
    for rowIdx, row := range dataRows {
        for colIdx, val := range row {
            col := string(rune('A' + colIdx))
            _ = f.SetCellValue(sheetName, col+fmt.Sprintf("%d", 5+rowIdx), val)
        }
    }
    return f, sheetName
}
```

#### 断言规范

使用 `testify/assert`，**禁止** `t.Log`/`t.Errorf`：

```go
assert.True(t, result.Ok, "全部有效时应通过")
assert.False(t, result.Ok, "存在无效数据时应失败")
assert.Len(t, result.ErrCells, 1, "应有1个错误")
assert.Contains(t, result.ErrCells[0].Reason, "关键字", "错误信息应包含定位信息")
```

#### 时间注入测试示例

对于使用 `param.Now` 的规则：

```go
func TestXxx_WithFixedTime(t *testing.T) {
    fixedTime := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)
    param := json_rule.CheckParam{
        Now:         fixedTime,  // 注入固定时间
        // ...其他参数
    }
    result := rule.Check(param)
    assert.True(t, result.Ok)
}
```

#### 运行测试

```bash
go test ./xlsx/coded_rules/table/... -v -run TestYourRule
go test ./xlsx/coded_rules/cross_table/... -v -run TestYourRule
```

## 四、现有规则修改

修改现有规则时，除了修改实现代码外，还必须：

1. **更新测试**：确保现有测试仍然通过，并根据新逻辑补充测试用例
2. **更新 Meta() 描述**：如果规则行为发生变化，更新 `Description` 和 `DisplayName`
3. **更新 ParamDefs**：如果新增参数，添加对应的 `TableRuleParamDef` 和 `ERuleParam` 常量
4. **更新文档**：更新对应目录的 CLAUDE.md 中的规则列表

### 常见修改场景

| 修改类型 | 操作 |
|----------|------|
| 增加校验条件 | 在 `Check()` 中追加检查逻辑，添加对应测试 |
| 新增参数 | 在 `rule_def.go` 添加 `ERuleParam`，在 `Meta().ParamDefs` 中声明，在 `Check()` 中通过 `param.Params` 读取 |
| 更换目标表 | 修改 `Meta().TargetSheets` 和 `default_table_rules.go` |
| 增加依赖表 | 修改 `Meta().RequiredSheets`，在 `Check()` 中通过 `sheetMap` 加载 |

### 参数读取模式

推荐使用 `Meta().ResolveParams()` 统一处理默认值（已由框架在 `CheckWithFilter` 中调用 `SupplementDefaultParams` 自动补充）。在规则内部读取参数：

```go
// 读取带默认值的参数
warnDays := 7 // 默认值
if val, ok := param.Params[string(json_rule.WARN_DAYS_BEFORE)]; ok && val != "" {
    if days, err := check_internal.ParseIntWithError(val); err == nil {
        warnDays = days
    }
}
```

如果需要在规则外部构建参数，使用 `ResolveParams`：

```go
params := meta.ResolveParams(map[string]string{
    string(json_rule.WARN_DAYS_BEFORE): "14", // 覆盖默认值
})
```

## 五、常用辅助函数速查

### check_internal/utils.go — 表查找与数据边界

| 函数 | 用途 | 示例 |
|------|------|------|
| `FindSheetBySuffix(sheetMap, suffix)` | 后缀匹配查找表 | `"Hero"` 匹配 `"武将\|Hero"` |
| `MatchSheetBySuffix(sheetName, requiredFilter)` | 检查表名是否在需求集合中 | 缓存按需加载时使用 |
| `AutoDetectEndIndex(cols, cIdx, c1idx, spaceNum)` | 自动检测数据结束行（4参数） | 列级规则使用，表级规则不需要直接调用 |
| `GetDataEndIndex(cols, startRowIdx)` | 从其他表加载时的数据结束位置 | 表级规则构建验证数据集合时使用 |
| `ResolveNow(now time.Time)` | 获取当前时间（零值时用 `time.Now()`） | 时间相关规则应使用此函数 |

### check_internal/hero_rule_helper.go — 列数据读取与解析

| 函数 | 用途 | 示例 |
|------|------|------|
| `GetColIndexByName(cols, name)` | 按列名查列索引 | `GetColIndexByName(cols, "Id")` |
| `GetColValue(cols, colIdx, rowIdx)` | 安全读取单元格（越界返回 `""`） | — |
| `GetColValues(cols, colIndex, startRowIdx)` | 获取列的所有数据值 | — |
| `ParseCommaSeparatedIds(idsStr)` | 解析 `"1,2,3"` 为 `[]int` | — |
| `ParseItemCfg(itemCfgStr)` | 解析 `{id;count}{id;count}` 为 `[]*ItemConfig` | — |

### check_internal/rule_helpers.go — 规则辅助

| 函数 | 用途 |
|------|------|
| `ParseIntWithError(s)` | 安全整数解析，返回 `(int, error)` |

### check_internal/hero_rule_helper.go — 武将与时间相关

| 函数 | 用途 |
|------|------|
| `ParseDate(dateStr)` | 解析日期字符串为 `time.Time`（定义于 hero_rule_helper.go，**非** excel_internal） |
| `TimeEquals(t1, t2)` | 精确到秒的时间比较 |
| `TimeIsZero(t)` | 检查时间是否为零值 |
| `TimeIsBefore(t, target)` | 时间先后比较 |
| `TimeIsAfter(t, target)` | 时间先后比较 |
| `TimeIsInRange(t, start, end)` | 时间范围判断 |
| `FormatDate(t)` | 格式化日期为字符串 |
| `FormatDateTime(t)` | 格式化日期时间为字符串 |
| `FormatChangeMessage(rowName, rowId, colName, oldValue, newValue)` | 格式化变更消息 |
| `FormatAddRowMessage(rowId, rowName)` | 格式化新增行消息 |
| `FindHeroById(heroId, heroCols, startRowIdx)` | 在 Hero 表中按 ID 查找（返回 `*HeroRow`） |
| `IsHeroOpened(hero *HeroRow)` | 检查武将是否已开放 |
| `FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, startRowIdx)` | 从战令奖励中找出战令武将 |
| `GetSeasonPassTime(seasonPassId, seasonPassCols, startRowIdx)` | 获取战令开始/结束时间 |
| `FindArenaGeneralHeroes(arenaScoreRewardsCols, startRowIdx)` | 找出大将军段位武将 |
| `FindGeneralDan(arenaScoreRewardsCols, startRowIdx)` | 获取大将军段位编号 |
| `GetArenaSeasonTime(seasonId, arenaSeasonCols, startRowIdx)` | 获取竞技场赛季时间 |
| `GetHeroDropPoolStatus(heroId, dropItemCols, startRowIdx)` | 获取武将掉落池状态 |
| `IsHeroInDropPool(heroId, dropItemCols, startRowIdx)` | 武将是否在掉落池中 |
| `GetDropItemInfos(dropItemCols, startRowIdx)` | 获取所有掉落道具信息 |
| `IsHeroSynthesisEnabled(heroId, itemCols, startRowIdx)` | 武将合成是否启用 |
| `IsHeroMeltEnabled(heroId, heroCols, startRowIdx)` | 武将熔炼是否启用 |
| `BuildSkillMeltMap(skillMeltCols, startRowIdx)` | 构建技能熔炼映射 |
| `IsSkillMeltConfigured(skillId, skillMeltMap)` | 技能是否配置了熔炼 |
| `IsHeroInSeasonPassReward(heroId, seasonPassRewardCols, startRowIdx)` | 武将是否在战令奖励中 |
| `ExtractHeroIdFromItemCfg(itemId)` | 从道具ID提取武将ID |
| `IsHeroItem(itemId)` | 判断道具ID是否为武将道具 |
| `MakeHeroItemId(heroId)` | 根据武将ID构造道具ID |

### excel_internal — Excel 解析与日期

| 函数 | 用途 |
|------|------|
| `ParseDate(str)` | 解析日期字符串为 `time.Time` |
| `ParseTime(str)` | 解析时间字符串 |
| `GetSheetMap(dir)` | 加载目录下所有 Excel 文件 |
| `GetSheetMapFromBytes(filesData, repoPath)` | 从字节数据构建 sheetMap |
| `MJS_FIXED_ROWS_NUM` | 常量 = 4（数据起始行索引） |
| `MJS_FIXED_ROWS_NAME` | 常量 = 0（中文名行索引），字段名在索引 2 |

### 使用辅助函数的注意事项

- `GetColIndexByName`、`GetColValue`、`GetColValues` 定义在 `hero_rule_helper.go` 中（非 `utils.go`），但通过 `check_internal` 包统一导出
- `ParseIntWithError` 定义在 `rule_helpers.go` 中
- 日期解析有两个位置：`check_internal.ParseDate`（hero_rule_helper.go）和 `excel_internal.ParseDate`，规则实现中优先使用 `excel_internal.ParseDate`
- 如果现有辅助函数不满足需求，优先在对应文件中扩展而非在规则文件中写本地版本
- 如果辅助函数是规则特有的逻辑，放在规则文件中即可
- **表级规则的遍历循环**始终使用 `for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++`，`EndIndex` 由框架传入
- **从其他表构建验证数据**时使用 `GetDataEndIndex(cols, startRowIdx)` 获取结束位置（非 `AutoDetectEndIndex`）

## 六、调试排错

### 编译验证

```bash
# 在 rain-excel-checker 目录下
go build ./...
```

### 运行测试

```bash
# 运行指定规则的测试
go test ./xlsx/coded_rules/cross_table/... -v -run TestYourRule
go test ./xlsx/coded_rules/table/... -v -run TestYourRule

# 运行所有测试
go test ./... -v
```

### 常见问题排查

| 症状 | 原因 | 修复 |
|------|------|------|
| 规则不执行（飞书无消息） | `default_table_rules.go` 未配置 | 步骤 4 补上 |
| panic: index out of range | 直接用 `cols[i][j]` 而非 `GetColValue` | 替换为安全函数 |
| 找不到依赖表 | `FindSheetBySuffix` 参数不对 | 检查表名的英文后缀 |
| 规则注册了但 Meta 返回 nil | `init()` 中未注册 | 步骤 3 补上 |
| 编译错误：undefined | 常量未添加到 `rule_def.go` | 步骤 1 补上 |
| 测试数据读取不到 | 行号/列号构造不正确 | 检查 Excel 行结构常量 |
| 增量检查不触发规则 | TargetSheets 挂载在了错误的表 | 参见 2.5 TargetSheets 挂载原则 |
| 时间相关测试 flaky | 直接使用 `time.Now()` | 使用 `param.Now` 注入固定时间 |

### 全量检查（通过日志验证规则）

开发完新规则后，通过全量模式对本地 `d:\work\config` 中的 Excel 执行检查，通过控制台日志验证规则是否正常工作：

```bash
# 在 rain-excel-checker 目录下
go run main.go -feishuRobot=none -mode=full -excelPath=d:\work\config
```

参数说明：
- `-feishuRobot=none` — 禁用飞书通知（开发调试时避免消息污染）
- `-mode=full` — 全量检查模式，扫描 excelPath 下所有 xlsx 文件
- `-excelPath=d:\work\config` — 策划配置目录

注意：全量模式下通用通知规则（NEW_ROW_NOTIFY/ROW_CHANGE_NOTIFY 及其拆分规则）被跳过（无 diff 基准）。

### 通过 MCP 接口调试

如果 rain-qa-func 正在运行，可以通过 MCP 接口触发单条规则检查，验证结果是否符合预期。详见 `rain-qa-func/docs/MCP-USAGE.md`。

### 记录配表问题模式（使用 self-improvement）

如果新规则揭示了一种新的配表问题模式（如"又一个系统存在相同的过期引用问题"），记录到 `.learnings/LEARNINGS.md`：
- 问题模式描述
- 影响了哪些表/系统
- 校验规则的核心逻辑

这类模式可能影响其他系统，值得跨项目传播。

## 七、完成后的自检清单

开发完成后，逐项确认：

- [ ] **需求核对** — 已查看实际 Excel 数据，确认字段名、数据格式与需求描述一致
- [ ] **矛盾质疑** — 发现的矛盾/不一致已与用户确认并解决
- [ ] **测试先行** — 已使用 test-case-generator 技能设计测试用例清单并经用户审核
- [ ] `rule_def.go` — `ETableRule` 常量已添加
- [ ] 规则实现文件 — `Check()` 和 `Meta()` 已实现
- [ ] 规则实现文件 — package 名为 `coded_rules`（非 `table` 或 `cross_table`）
- [ ] 规则实现文件 — `Meta().TargetSheets` 设置正确（参见 2.5 挂载原则）
- [ ] 规则实现文件 — `Meta().RequiredSheets` 包含所有跨表依赖
- [ ] `table_check_manager.go` — `init()` 中已 `TableManager.Reg()`
- [ ] `default_table_rules.go` — `DefaultTableRules` 中已配置
- [ ] 测试文件 — 覆盖审核通过的测试用例清单中的所有场景
- [ ] 时间相关规则 — 使用 `ResolveNow(param.Now)` 而非 `time.Now()`
- [ ] `go build ./...` 编译通过
- [ ] `go test ./... -v` 全部通过
- [ ] 错误信息中文，包含足够定位信息
- [ ] 对应目录的 CLAUDE.md 已更新规则列表

## 八、典型规则场景模板

以下是常见业务场景的实现思路，供快速参考。

### 场景 A：时间匹配检查

"表A的时间字段必须和表B的时间字段一致"

参考：`table/table_check_season_pass_hero.go`（战令武将开放时间 vs 战令时间）

关键实现要点：
1. 从源表提取目标行（可能需要按条件过滤）
2. 从依赖表提取参照时间
3. 使用 `TimeEquals()` 精确到秒比较
4. 报告不匹配的行，显示期望值和实际值

### 场景 B：跨表存在性检查

"表A中的 XX 字段值必须存在于表B中"

参考：`cross_table/table_check_drop_item_must_in_item.go`（掉落道具必须在道具表）

关键实现要点：
1. 从依赖表构建 `map[KeyType]bool` 集合
2. 源表中可能有复杂格式需要先解析（如 `{id;count}`）
3. 查找失败的记录要包含原始值和所在行信息

### 场景 C：字段归属检查

"表A中的 XX 必须关联到表B中的特定类型记录"

参考：`cross_table/table_check_drawskin_byproduct.go`（皮肤副产品必须配置归属武将）

关键实现要点：
1. 解析源表中的 ID 列表（逗号分隔）
2. 在目标表中查找每个 ID 并验证属性
3. 报告未找到或属性不匹配的记录

### 场景 D：活动相关时间检查

"活动开启/结束时间必须符合特定模式（如周四开始、周三结束）"

参考：`table/table_check_activity_danqingge_time.go`（丹青阁活动时间校验）

关键实现要点：
1. 通过 `ActivityType` 过滤特定活动类型
2. 检查 `TimeType` 区分绝对时间和相对时间
3. 比较时间是否在合理范围内
4. 可配置预警阈值（如提前 N 天警告即将到期）

### 场景 E：数据有效性检查（简单单表）

"表内某字段必须满足条件"

参考：`table/table_check_date_valid_expire.go`（ValidDate <= ExpireDate）

关键实现要点：
1. 查找关键列，不存在则跳过检查
2. 遍历数据行，两列均非空时比较
3. 简洁的错误信息包含行ID和实际值

### 场景 F：条件引用检查

"表A中的字段值必须在表B的特定条件记录中存在"

参考：`cross_table/table_check_drop_rule_conditional.go`（DropRule 条件引用检查）

关键实现要点：
1. 从依赖表构建带条件的验证集合
2. 源表数据需要按条件分组
3. 报告引用失败的记录，包含条件和值信息

## 九、通知规则系统（已拆分）

原 `NEW_ROW_NOTIFY` 和 `ROW_CHANGE_NOTIFY` 已拆分为 5 个独立规则：

| 旧枚举（Deprecated） | 拆分后的新枚举 | 说明 |
|---------------------|---------------|------|
| `NEW_ROW_NOTIFY` | `ADDED_ROW_NOTIFY` | 新增行通知 |
| | `REMOVED_ROW_NOTIFY` | 删除行通知 |
| | `ADDED_COL_NOTIFY` | 新增列通知 |
| | `REMOVED_COL_NOTIFY` | 删除列通知 |
| `ROW_CHANGE_NOTIFY` | `MODIFIED_ROW_NOTIFY` | 修改行字段通知 |

- 旧枚举仍保留注册（向后兼容），但在通用规则执行时被跳过
- `GeneralRuleOverrides` 支持按表覆盖通知参数（如 ID 列名），新枚举自动回退到旧枚举的配置
- 通知规则基于 Git diff 或快照对比检测变更，全量模式（无 diff 基准）时跳过

## 十、文档更新

完成规则开发后，更新以下文档：

| 文档 | 更新内容 |
|------|----------|
| `rain-excel-checker/CLAUDE.md` | 已注册规则数量（如有变化） |
| `xlsx/coded_rules/CLAUDE.md` | 如果是新类型的规则 |
| `xlsx/coded_rules/table/CLAUDE.md` 或 `cross_table/CLAUDE.md` | 规则列表、依赖表 |
| `xlsx/json_rule/CLAUDE.md` | 新增的 ETableRule 常量（如有） |
| `xlsx/check_manager/CLAUDE.md` | 已注册表级检查器数量（如有变化） |
