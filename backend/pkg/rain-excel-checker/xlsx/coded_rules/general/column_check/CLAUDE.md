# coded_rules/general/column_check 目录文档

列级通用校验规则实现，按业务域组织为子目录，与前端 `rules/components/` 目录结构一致。

## 目录结构

```
column_check/
├── CLAUDE.md                       # 本文档
├── base/                           # package base — 基础/通用规则
├── datatype/                       # package datatype — 数据类型检查
├── business/
│   ├── calculation/                # package calculation — 计算类
│   ├── date/                       # package date — 日期类
│   ├── numeric/                    # package numeric — 数值类
│   ├── pinyin/                     # package pinyin — 拼音类
│   └── reference/                  # package reference — 引用/正则类
```

## 规则列表

### base/ — 基础/通用规则

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `ALL_BASE` | 全基础类型检查 | - |
| `NOT_EMPTY` | 非空检查 | ALLOW_COMMIT |
| `INCREASE` | 递增检查 | START, ALLOW_COMMIT |
| `UNIQUE` | 唯一性检查 | SCOPE, ALLOW_COMMIT |
| `CHS_ONLY` | 纯中文检查 | ALLOW_COMMIT |
| `RESOURCE` | 资源路径检查 | ALLOW_COMMIT |
| `SERVER_OR_CLIENT` | 服务端/客户端标识检查 | ALLOW_COMMIT |

### datatype/ — 数据类型检查

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `BOOLEAN` | 布尔值检查 | ALLOW_COMMIT |
| `DATE` | 日期格式检查 | FORMAT, ALLOW_COMMIT |
| `NUMERIC` | 数值检查 | ALLOW_COMMIT |
| `STRING` | 单元格应为字符串（基于 excelize 单元格类型），检测字符串列中是否存在数值格式单元格；前端对应 `CellTypeCheckParams.vue` | ALLOW_EMPTY, ALLOW_COMMIT, BREAK_LINE |
| `SPECIAL_FORMAT` | 特殊格式检查 | FORMAT, ALLOW_COMMIT |
| `RICH_TEXT` | 富文本检查 | ALLOW_COMMIT |

### business/calculation/ — 计算类

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `WEIGHT_SUM` | 权重和检查 | SUM, ALLOW_COMMIT |
| `DATE_CONSISTENCY` | 日期一致性检查 | TIME_COL_NAME, ALLOW_COMMIT |

### business/date/ — 日期类

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `DATE_DURATION` | 日期持续时间检查 | MIN_DURATION, MAX_DURATION, ALLOW_COMMIT |
| `DATE_RANGE` | 日期范围检查 | MIN, MAX, ALLOW_COMMIT |

### business/numeric/ — 数值类

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `ENUM` | 枚举值检查 | ENUM_VALUES, ALLOW_COMMIT |
| `NUMERIC_RANGE` | 数值范围检查 | MIN, MAX, ALLOW_COMMIT |

### business/pinyin/ — 拼音类

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `PIN_YIN_CHS` | 拼音中文检查 | ALLOW_COMMIT |

### business/reference/ — 引用/正则类

| 规则类型 | 说明 | 参数 |
|----------|------|------|
| `CROSS_REFERENCE` | 交叉引用检查 | REFERENCE_SHEET, ALLOW_COMMIT |
| `CHAIN_REFERENCE` | 跨表关系链检查 | CHAIN_STEPS, CHAIN_COMPARE, CHAIN_MATCH_COMPARE |
| `SPLIT_REFERENCE` | 分割引用检查 | SEPARATOR, REFERENCE_SHEET, ALLOW_COMMIT |
| `FOREIGN_KEY` | 外键检查 | REFERENCE_SHEET, ALLOW_COMMIT |
| `REGEX` | 正则表达式检查 | PATTERN, ALLOW_COMMIT |

## 文件命名规范

- 规则实现：`excel_check_{rule_type}.go`
- 单元测试：`excel_check_{rule_type}_test.go`

## 通用参数说明

| 参数 | 说明 | 适用规则 |
|------|------|----------|
| `ALLOW_COMMIT` | 是否允许提交时存在错误 | 所有规则 |
| `FORMAT` | 日期/特殊格式格式 | DATE, SPECIAL_FORMAT |
| `MIN/MAX` | 最小/最大值 | NUMERIC_RANGE, DATE_DURATION, STRING |
| `SCOPE` | 唯一性范围(global/sheet) | UNIQUE |

## 开发新规则

1. 确定规则所属子目录（base/datatype/business/xxx）
2. 在对应子目录创建 `excel_check_{rule_type}.go`，使用对应 package 声明
3. 实现 `Checker` 接口
4. 在 `json_rule/` 添加规则类型常量
5. 在 `engine/column_registry.go` 注册（import 对应子包）
6. 创建对应的单元测试文件（与源文件同目录）
7. **修改或新增规则后，同步更新前端参数组件**：
   - 若规则对应 `frontend/src/pages/excel-test/composables/rules/components/` 下的独立 `.vue` 参数组件，必须同步更新组件顶部注释和内部说明文本
   - 若规则在 `excel-rules-template.ts` 中注册，需确保注释中标注对应的 `EColRule` 枚举类型，便于后续定位
   - 示例：`STRING` 规则对应 `CellTypeCheckParams.vue`，其描述需与后端实际检查行为保持一致

## CHAIN_REFERENCE 参数详解

`CHAIN_REFERENCE` 规则用于跨表多跳转关系链检查，支持动态 N 步跳转和多种比较模式。

> 核心实现已迁移到 [chain_reference/](../../../chain_reference/) 目录，本文件仅保留薄适配层（`excel_check_chain_reference.go`）。
> 完整文档见 [chain_reference/CLAUDE.md](../../../chain_reference/CLAUDE.md)

### 参数说明

| 参数 | 说明 | 必填 | 示例 |
|------|------|------|------|
| `CHAIN_STEPS` | JSON 格式的两链配置（left/right 各包含 steps 和 compareCol） | 是 | 见下方配置示例 |
| `CHAIN_COMPARE` | 比较阶段类型（操作第一步数据）：verify_exists/time_overlap/date 类型 | 否，默认 verify_exists | `verify_exists` |
| `CHAIN_MATCH_COMPARE` | 匹配阶段类型（操作最后一步数据）：verify_exists/time_overlap/date 类型 | 否，默认 verify_exists | `verify_exists` |

### CHAIN_STEPS 配置格式

```json
{
  "left": {
    "steps": [
      {
        "sheet": "",
        "preCol": "",
        "findVal": "col",
        "nextCol": "OnceDropRule"
      },
      {
        "sheet": "掉落规则表|DropRule",
        "preCol": "Id",
        "findVal": "self",
        "nextCol": "DropGroup"
      }
    ]
  },
  "right": {
    "steps": [
      {
        "sheet": "掉落道具表|DropItem",
        "preCol": "Item",
        "findVal": "self",
        "nextCol": "DropGroup"
      },
      {
        "sheet": "掉落分组表|DropGroup",
        "preCol": "Id",
        "findVal": "self",
        "nextCol": "Id"
      }
    ]
  }
}
```

### Step 配置字段

| 字段 | 说明 | 左链第一步 | 右链第一步 | 后续步骤 |
|------|------|-----------|-----------|---------|
| `sheet` | 目标表名 | 强制为空 | 目标表名 | 目标表名 |
| `preCol` | 在目标表中用哪列匹配/提取 | 强制为空 | 全表扫描时提取 FirstStepInputValues | 匹配输入值 |
| `findVal` | 匹配值来源 | "col" | "self" | "self" |
| `nextCol` | 从当前表哪列取值 / 提取哪列传给下一步 | 指定列名取值（如 OnceDropRule） | 提取值传给下一步 | 提取值传给下一步 |
| `pattern` | 正则提取模式 | 可选 | 不使用 | 可选 |
| `groups` | 正则捕获组 | 可选 | 不使用 | 可选 |
| `filterCol` | 过滤条件列 | 不使用 | 可选 | 可选 |
| `filterVal` | 过滤条件值 | 不使用 | 可选 | 可选 |

### compareCol 字段

- 位于 `left` 或 `right` 配置的顶层
- 指定最终提取的比较列名（如 "StartTime"、"EndTime"）
- 仅在 `CHAIN_COMPARE=time_overlap` 时使用
- 当两链都有 `compareCol` 时，启用两阶段比较（先匹配 Match 值，再比较 Compare 时间值）

### 两阶段门控模型

当 `CHAIN_MATCH_COMPARE` 非空时，启用两阶段门控比较：

1. **Phase 1（匹配阶段）**：用 `chainMatchCompare` 比较两链**最后一步**的值，判断是否"交汇"
   - 交汇 → 进入 Phase 2
   - 不交汇 → 跳过此行，不报错
2. **Phase 2（比较阶段）**：对交汇成功的行对，用 `chainCompare` 比较当前列值与右链**第一步 preCol** 的提取值，判断是否报错

**退化情况**：`chainMatchCompare` 为空时，退化为单阶段比较（使用最后一步值），保持向后兼容。

**SQL 类比**：
```sql
SELECT 左链第一步值, 右链第一步值     -- chainCompare 操作的值
FROM ...
WHERE 两链最后一步值匹配             -- chainMatchCompare 门控条件
  AND 比较规则检查(第一步值)          -- chainCompare 实际检查
```

### 比较类型

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| `verify_exists` | **Match 阶段为门控语义**：左侧值不在右侧集合时静默跳过（不报错）<br>**Compare 阶段为报错语义**：左侧值不在右侧时报错并报告缺失值 | 配合两阶段门控使用（先 Match 判断交汇，再 Compare 检查） |
| `verify_must_exist` | **引用完整性强制检查**：左侧值不在右侧集合时报错并报告缺失值（Match 与 Compare 阶段语义一致） | 检查"A.byproduct 必须在 B.Item 中存在"等强制引用完整性约束 |
| `time_overlap` | 时间点匹配：左链时间点与右链时间点有相同值 → 报错 | 战令保护期检查 |
| `date_equals` | 日期相同：秒级比较 | 日期精确匹配 |
| `date_before_or_equal` | 日期早于或等于：右链时间早于或等于左链 → 报错 | 时间顺序检查 |
| `date_after_or_equal` | 日期晚于或等于：右链时间晚于或等于左链 → 报错 | 时间顺序检查 |

**verify_exists vs verify_must_exist 选型**：
- 用 `verify_exists` 作 chainMatchCompare：两链可能无交汇时（如不同活动期间），静默跳过；只在交汇时进入比较
- 用 `verify_must_exist` 作 chainMatchCompare：左链每个值都必须在右链中存在，缺失即报错（不可降级为门控）

### 预警窗口参数

| 参数 | 说明 | 必填 | 示例 |
|------|------|------|------|
| `chainWarnBefore` | Go duration，仅在预警时间距今不足此时长时报错 | 否 | `168h`（7天） |
| `chainWarnSheet` | 预警时间来源表名，需与某条链最后一步的 sheet 匹配 | 否 | `赛季战令表\|SeasonPass` |
| `chainWarnCol` | 预警时间来源列名 | 否 | `StartTime` |

三参数同时配置时生效。逐行精确：从链最后一步匹配到的行中提取预警时间，而非全表扫描。超出预警窗口的错误被静默（不产生 CellError）。
