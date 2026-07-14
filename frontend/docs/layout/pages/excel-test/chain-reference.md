# CHAIN_REFERENCE — 跨表关系链检查规则参数

> 主组件: `composables/rules/components/business/reference/ChainReferenceParams.vue`
> 步骤卡片: `composables/rules/components/business/reference/ChainStepCard.vue`（左右链共用）
> 正则行: `composables/rules/components/RegexFormatRow.vue`（正则下拉 + IsArray 开关）
> 共享工具: `composables/rules/chain-reference-params.ts`
> 规则分类: 业务关系规则 > 关联表 > 关系链检查

#### 三列 CSS Grid 布局

```
grid-template-columns: 1fr minmax(160px, auto) 1fr
grid-template-rows: auto 1fr ... 1fr   (标题行 + maxSteps 个 1fr)
```

中间列跨 `2 / -1`，覆盖所有步骤行。

```
┌──────────────────────┬────────────────────┬──────────────────────┐
│ 来源链 (left)        │       比较         │ 目标链 (right)       │
├──────────────────────┤                    ├──────────────────────┤
│ ┌─── 步骤1 ────────┐│  ┌──比较规则──┐   │┌─── 步骤1 ────────┐ │
│ │(无表名)           ││  │操作第一步  │   ││表: [sheet]        │ │
│ │正则+IsArray       ││  │ [下拉选择] │   ││列: [preCol]      │ │
│ │过滤: 列__ 值__    ││  └───────────┘   ││与左边比较         │ │
│ └───────────────────┘│                   ││正则+IsArray       │ │
│                      │  ┌──匹配规则──┐   ││过滤: 列__ 值__    │ │
│  (左链仅1步)         │  │操作最后一步│   │└───────────────────┘ │
│                      │  │ [下拉选择] │   │                      │
├──────────────────────┤  └───────────┘   ├──────────────────────┤
│ ┌─── 步骤2 ────────┐│                   │ ┌─── 步骤2 ────────┐│  ← 末行对齐
│ │表: [sheet]   [x]  ││                   │ │表: [sheet]   [x]  ││
│ │列: [preCol]      ││                   │ │列: [preCol]      ││
│ │与上一步匹配       ││                   │ │与上一步匹配       ││
│ │正则+IsArray       ││                   │ │正则+IsArray       ││
│ │过滤: 列__ 值__    ││                   │ │过滤: 列__ 值__    ││
│ │列 [nextCol]       ││                   │ │列 [nextCol]       ││
│ │与右边匹配         ││                   │ │与左边匹配         ││
│ └───────────────────┘│                   │ └───────────────────┘│
└──────────────────────┴───────────────────┴──────────────────────┘
```

**中间列两张卡片始终同时显示**：
- **比较规则卡片**（上方）：操作两侧链**第一个**步骤的数据，参数键 `chainCompare`
- **匹配规则卡片**（下方）：操作两侧链**最后一个**步骤的数据，参数键 `chainMatchCompare`
- 两张卡片各有独立的比较类型下拉（值匹配/时间匹配/存在性等）

**步骤级 IsArray 双重作用**：
- 步骤内：提取值后按逗号拆分（如 `"{1039016;1},{1040015;1}"` → `["{1039016;1}","{1040015;1}"]`）
- **左链第一步 IsArray** 还控制比较阶段当前列值的拆分（Phase 2 比较前拆分 BigAward 等多值字段）

**末行对齐规则**：当两侧步骤数不同时，短链的最后一步跳到网格末行（与长链最后一步水平对齐），确保来源链最后一步、匹配规则卡片底部、目标链最后一步三者水平对齐。

#### 悬停操作

鼠标悬停在 ChainStepCard 上时显示操作按钮：
- **向上插入**（顶部浮出）：在当前步骤前插入新步骤
- **向后追加**（底部浮出）：在当前步骤后追加新步骤
- **x 按钮**（标题行右侧）：删除当前步骤（至少保留1步）

#### 左侧（来源链）步骤卡片 — ChainStepCard side="left"

**第一步** (`i === 0`)：

| 行 | 字段 | 控件 | 说明 |
|----|------|------|------|
| 1 | 标题 "步骤1"（无表名 + 无删除） | 文本 | 来源链第一步无表名输入 |
| 2 | 正则格式 + IsArray 开关 | RegexFormatRow | 共用组件 |
| 3 | 过滤条件（列+值） | NInput + NInput | |

> 左链第一步无 preCol/nextCol 行。单步时 nextCol 不显示（hasMultiple=false）。

**中间步骤** (`0 < i < length-1`)：

| 行 | 字段 | 控件 | 说明 |
|----|------|------|------|
| 1 | 标题 "步骤n" + 表: [sheet] + [x] | 文本 + NInput + Button | 表名可编辑 |
| 2 | 列: [preCol] 与上一步匹配 | NInput | 查找列 |
| 3 | 正则格式 + IsArray 开关 | RegexFormatRow | 共用组件 |
| 4 | 过滤条件（列+值） | NInput + NInput | |
| 5 | 列 [nextCol] 与下一步匹配 | NInput | 提取列 |

**最后一步** (`i === length-1, length > 1`)：

| 行 | 字段 | 控件 | 说明 |
|----|------|------|------|
| 1 | 标题 "步骤n" + 表: [sheet] + [x] | 文本 + NInput + Button | 表名可编辑 |
| 2 | 列: [preCol] 与上一步匹配 | NInput | 查找列 |
| 3 | 正则格式 + IsArray 开关 | RegexFormatRow | 共用组件 |
| 4 | 过滤条件（列+值） | NInput + NInput | |
| 5 | 列 [nextCol] 与右边匹配 | NInput | 提取列，交汇匹配 |

#### 右侧（目标链）步骤卡片 — ChainStepCard side="right"

**第一步** (`i === 0`)：

| 行 | 字段 | 控件 | 说明 |
|----|------|------|------|
| 1 | 标题 "步骤1" + 表: [sheet] + [x] | 文本 + NInput + Button | 表名可编辑 |
| 2 | 列: [preCol] 与左边比较 | NInput | 全表扫描提取 preCol 值 |
| 3 | 正则格式 + IsArray 开关 | RegexFormatRow | 共用组件 |
| 4 | 过滤条件（列+值） | NInput + NInput | |
| 5 | 列 [nextCol] 与下一步匹配 | NInput | 提取列（多步时显示） |

**中间步骤** (`0 < i < length-1`)：与左侧中间步骤相同。

**最后一步** (`i === length-1, length > 1`)：

| 行 | 字段 | 控件 | 说明 |
|----|------|------|------|
| 1 | 标题 "步骤n" + 表: [sheet] + [x] | 文本 + NInput + Button | 表名可编辑 |
| 2 | 列: [preCol] 与上一步匹配 | NInput | 查找列 |
| 3 | 正则格式 + IsArray 开关 | RegexFormatRow | 共用组件 |
| 4 | 过滤条件（列+值） | NInput + NInput | |
| 5 | 列 [nextCol] 与左边匹配 | NInput | 提取列，交汇匹配 |

#### 数据结构

> 定义在 `chain-reference-params.ts`

```typescript
// 一侧链配置
interface ChainSideConfig {
    steps: ChainStep[]     // 步骤列表
    compareCol: string     // 预留比较列（当前未使用）
}

// 完整链对配置（序列化为 chainSteps 参数）
interface ChainPairConfig {
    left: ChainSideConfig   // 来源链
    right: ChainSideConfig  // 目标链
}

// 单个步骤
interface ChainStep {
    sheet: string       // 目标表名
    preCol: string      // 查找列名
    findVal: string     // 值来源: "col"=指定列, "self"=上一步结果
    nextCol: string     // 提取列
    pattern: string     // 正则模式（可选）
    groups: string      // 正则捕获组（可选）
    filterCol: string   // 过滤列（可选）
    filterVal: string   // 过滤值（可选）
    isArray: string     // "true"/"false" — 步骤级数组拆分控制
}
```

#### 两阶段模型

1. **比较阶段**（比较规则卡片）：操作两侧链**第一个**步骤的数据
2. **匹配阶段**（匹配规则卡片）：操作两侧链**最后一个**步骤的数据

#### 参数存储格式

| 参数键 | 说明 |
|--------|------|
| `chainSteps` | JSON 字符串，ChainPairConfig 结构（含 compareCol，当前序列化时留空） |
| `chainCompare` | 比较阶段类型：`verify_exists` / `time_overlap` / `date_equals` / ... |
| `chainMatchCompare` | 匹配阶段类型：`verify_exists` / `time_overlap` / `date_equals` / ... |

---
**Verification Date**: 2026-05-11
**Status**: 从 docs/layout/CLAUDE.md 迁移而来，已与代码同步
