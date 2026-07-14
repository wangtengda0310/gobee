# 前端页面设计模式参考

> 本文件包含活动Wiki前端页面的设计规范，包括页面布局结构、组件使用规范、页签内容设计模式、关联关系展示规范和样式规范。
>
> 内容来源：SKILL.md 的"前端页面设计规范"章节。

## 目录

- [页面布局结构](#页面布局结构)
- [组件使用规范](#组件使用规范)
- [页签内容设计模式](#页签内容设计模式)
- [关联关系展示规范](#关联关系展示规范)
- [样式规范](#样式规范)

---

## 页面布局结构

活动Wiki页面采用**卡片 + 页签**的层级结构：

```
activity-panel.vue (单个活动卡片)
|-- n-card (活动卡片)
|   |-- header (活动标题区)
|   |   |-- 活动名称 + 序号
|   |   |-- ActivityType 标签 (彩色)
|   |   |-- EActivityId 标签 (蓝色)
|   |   +-- header-extra: 打开Excel按钮 + 活动ID
|   +-- content (活动内容区)
|       +-- n-tabs (页签容器)
|           |-- 基础信息页签 (始终显示)
|           |-- 配表A页签 (条件渲染)
|           |-- 配表B页签 (条件渲染)
|           +-- ...
```

## 组件使用规范

**1. 外层卡片 (n-card)**
- 使用 `hoverable` 添加悬停效果
- `segmented` 开启内容分隔线
- **禁止设置 `:title` 属性**：当 `:title` 被设置（即使返回 null），Naive UI 会接管标题渲染并阻止 `#header` slot 生效，导致标题区域不显示
- 标题和标签使用 `<template #header>` slot 构建，右侧按钮用 `<template #header-extra>` slot
- ActivityPanel 和 SeasonPassPanel（及其他独立子系统面板）必须保持统一的 header 风格：左侧文字标题+标签，右侧打开Excel+ID badge

**2. 页签容器 (n-tabs)**
- `type="line"` 线性页签样式
- `animated` 开启切换动画
- 每个页签使用 `v-if` 条件渲染

**3. 内容布局 (n-grid)**
- 使用 `:cols="24"` 的网格系统
- 常用列宽：`:span="12"`（半宽）、`:span="24"`（全宽）
- 间距：`:x-gap="16" :y-gap="16"`

**4. 信息展示组件**

| 场景 | 组件 | 属性 |
|------|------|------|
| 键值对信息 | n-descriptions | `label-placement="left"` `:column="2"` `bordered` |
| 列表数据 | n-table | `:bordered="true"` `:single-line="false"` `size="small"` |
| 标签值 | n-tag | `:bordered="false"` `size="small"` `type="primary/info/success/warning/error"` |
| 普通文本 | n-text | `strong` 加粗, `type="success"` 着色 |
| 卡片分组 | n-card | `size="small"` `:bordered="false"` |

**5. 标签类型使用规范**

| 数据类型 | 标签类型 | 示例 |
|----------|----------|------|
| ID类 | 默认/无 | 活动ID、道具ID |
| 类型枚举 | primary | ActivityType、ShopType |
| 布尔值-是 | success | true、IsOpen |
| 布尔值-否 | warning/default | false |
| 数值/数量 | success | Price、Count |
| 重要ID | error | 大奖ID、保底道具ID |
| 数组 | info | UseCurrency、DropGroup |

## 页签内容设计模式

### 页签显示条件规则（重要）

页签的 `v-if` 条件必须基于**父级关联数据**是否存在，而非子级数据是否存在。当链式关联中某个环节缺失时，页签仍需显示，内部提示缺失原因。

```
❌ 错误：v-if="props.activityData.Pets && props.activityData.Pets.length > 0"
✅ 正确：v-if="props.activityData.DrawPet"
```

### 关联数据缺失提示模式

在每个页签内，用 `n-alert type="warning"` 按关联链层级逐级提示缺失原因：

```vue
<!-- 页签的 v-if 基于父级关联 -->
<n-tab-pane name="pet" v-if="props.activityData.DrawPet" ...>
  <!-- 第1级：关联字段为空 -->
  <n-gi :span="24" v-if="!props.activityData.DrawPet?.BigAwardItemId || props.activityData.DrawPet.BigAwardItemId <= 0">
    <n-alert type="warning" size="small">
      <template #header>关联数据缺失</template>
      DrawPet.BigAwardItemId 为空或为 0，无法关联灵宠数据。请检查「结缘亭|DrawPet」表中该记录的 BigAwardItemId 列。
    </n-alert>
  </n-gi>

  <!-- 第2级：关联字段有值但目标表无匹配 -->
  <n-gi :span="24" v-else-if="!props.activityData.Pets || props.activityData.Pets.length === 0">
    <n-alert type="warning" size="small">
      <template #header>关联数据缺失</template>
      DrawPet.BigAwardItemId = {{ props.activityData.DrawPet?.BigAwardItemId }}，但在「灵宠|Pet」表中未找到 Id={{ props.activityData.DrawPet?.BigAwardItemId }} 的记录。
    </n-alert>
  </n-gi>

  <!-- 数据正常时展示 -->
  <n-gi :span="24" v-for="..." :key="...">...</n-gi>
</n-tab-pane>
```

**提示文案必须包含**：哪个字段为空/无匹配、期望的关联逻辑、实际值、用户该检查什么。

### 数据展示模式

**模式A：纯信息展示（单条记录）**

适用于：DrawSkin、DropRule、Shop、HeroSkinItem 等单条数据

```vue
<n-gi :span="24">
  <n-card title="卡片标题" size="small" :bordered="false">
    <n-descriptions label-placement="left" :column="2" bordered>
      <n-descriptions-item label="字段名">
        <n-tag :bordered="false">{{ props.activityData.Xxx?.Field }}</n-tag>
      </n-descriptions-item>
      <!-- 布尔值字段 -->
      <n-descriptions-item label="是否XX">
        <n-tag :bordered="false" :type="props.activityData.Xxx?.BoolField ? 'success' : 'default'">
          {{ formatBoolean(props.activityData.Xxx?.BoolField) }}
        </n-tag>
      </n-descriptions-item>
      <!-- 数组字段 -->
      <n-descriptions-item label="数组字段">
        <n-tag :bordered="false" type="info">{{ formatArray(props.activityData.Xxx?.ArrayField) }}</n-tag>
      </n-descriptions-item>
    </n-descriptions>
  </n-card>
</n-gi>
```

**模式B：表格列表（多条记录）**

适用于：DropGroup、DropItem、ShopGoods、TimesRewards 等数组数据

```vue
<n-gi :span="24" v-if="props.activityData.XxxList && props.activityData.XxxList.length > 0">
  <n-card title="列表标题" size="small" :bordered="false">
    <n-table :bordered="true" :single-line="false" size="small">
      <thead>
      <tr>
        <th>ID</th>
        <th>名称</th>
        <th>道具</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(item, idx) in props.activityData.XxxList" :key="item?.Id || idx">
        <template v-if="item">
          <td>{{ item.Id }}</td>
          <td>{{ item.Name || '-' }}</td>
          <td>
            <n-tag v-for="(cfg, cidx) in item.Item" :key="cidx" size="small" :bordered="false">
              {{ cfg?.ItemId }}x{{ cfg?.Count }}
            </n-tag>
            <span v-if="!item.Item || item.Item.length === 0">-</span>
          </td>
        </template>
      </tr>
      </tbody>
    </n-table>
  </n-card>
</n-gi>
```

**模式C：复合布局（信息+列表）**

适用于：DropRule（规则信息 + DropGroup列表 + DropItem列表）

```vue
<n-grid :cols="24" :x-gap="16" :y-gap="16">
  <!-- 上半部分：规则信息 -->
  <n-gi :span="24">
    <n-card title="规则信息" size="small" :bordered="false">
      <n-descriptions ...>...</n-descriptions>
    </n-card>
  </n-gi>
  <!-- 下半部分：关联列表 -->
  <n-gi :span="24" v-if="props.activityData.SubList?.length > 0">
    <n-card title="子列表" size="small" :bordered="false">
      <n-table ...>...</n-table>
    </n-card>
  </n-gi>
</n-grid>
```

**模式D：三期数据卡片（PeriodCard子组件）**

适用于：丹青阁（DrawSkinCard）、战令（SeasonPassPeriodCard）、结缘亭等有多期数据的活动

**后端数据结构**：CompleteData 中包含 Prev/Basic/Next 三期字段：
```go
type XxxCompleteData struct {
    PrevXxx *xxx.XxxDiff  // 上一期
    Basic   *xxx.XxxDiff  // 当前期（时间匹配确定）
    NextXxx *xxx.XxxDiff  // 下一期
}
```

**PeriodCard 子组件模板**：

```vue
<!-- xxx-period-card.vue -->
<script setup lang="ts">
import BadgeLabel from "@shared/components/badge-label/index.vue";
const SHEET = '表名|SheetName'

const props = defineProps<{
  title: string
  periodData: any
  highlight: boolean
  periodLabel: string
}>()
</script>

<template>
  <n-card
      :title="title"
      size="small"
      :bordered="false"
      :class="highlight ? 'period-card--current' : 'period-card--other'"
  >
    <template #header-extra>
      <n-tag v-if="highlight" type="primary" size="small" :bordered="false">
        当前关联
      </n-tag>
      <n-tag v-else size="small" :bordered="false">
        {{ periodLabel }}
      </n-tag>
    </template>

    <n-descriptions label-placement="left" :column="2" bordered>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="Id" label="ID" /></template>
        <n-tag :bordered="false">{{ props.periodData?.Id }}</n-tag>
      </n-descriptions-item>
      <!-- 其他字段... -->
    </n-descriptions>
  </n-card>
</template>

<style scoped>
.period-card--current {
  border-left: 4px solid var(--n-primary-color, #2080f0) !important;
  background-color: var(--n-primary-color-hover, rgba(32, 128, 240, 0.08)) !important;
}
.period-card--other {
  border-left: 4px solid var(--n-border-color, rgba(255, 255, 255, 0.09)) !important;
  background-color: var(--n-action-color, rgba(255, 255, 255, 0.06)) !important;
}
</style>
```

**父组件中使用**（基础信息页签内纵向排列）：

```vue
<script setup lang="ts">
// 构建三期展示数据
interface Period {
  title: string
  periodData: any
  highlight: boolean
  periodLabel: string
}

const getPeriods = computed(() => {
  const periods: Period[] = []
  if (props.completeData.PrevXxx) {
    periods.push({ title: '上一期', periodData: props.completeData.PrevXxx, highlight: false, periodLabel: '上一期' })
  }
  if (props.completeData.Basic) {
    periods.push({ title: '本期 (当前关联)', periodData: props.completeData.Basic, highlight: true, periodLabel: '当前关联' })
  }
  if (props.completeData.NextXxx) {
    periods.push({ title: '下一期', periodData: props.completeData.NextXxx, highlight: false, periodLabel: '下一期' })
  }
  return periods
})
</script>

<template>
  <!-- 基础信息页签 -->
  <n-tab-pane name="basic" :tab="() => tBadge('表名|SheetName', '基础信息')">
    <n-grid :cols="24" :x-gap="16" :y-gap="16">
      <n-gi :span="24" v-for="period in getPeriods" :key="period.title">
        <XxxPeriodCard
          :title="period.title"
          :period-data="period.periodData"
          :highlight="period.highlight"
          :period-label="period.periodLabel"
        />
      </n-gi>
      <n-gi :span="24" v-if="getPeriods.length === 0">
        <n-empty description="无数据" size="small" />
      </n-gi>
    </n-grid>
  </n-tab-pane>
</template>
```

## 关联关系展示规范

**设计目标**：让用户清楚理解每个页签的数据是如何关联到当前活动的，特别是字段级别的关联逻辑。

**展示方式**：在每个页签内容顶部添加"关联说明"提示条。

**实现方案**：

```vue
<!-- 关联说明组件（放在每个页签的 n-grid 最前面） -->
<!--
  activity-wiki-dev: 关联说明提示
  本提示由 activity-wiki-dev 技能生成，用于展示当前页签数据与活动的关联逻辑。
  如果此关联说明的展示方式不符合需求（如位置、样式、内容详细程度），
  可反馈给 AI 以优化 activity-wiki-dev 技能中的"关联关系展示规范"章节。
-->
<n-gi :span="24">
  <n-alert type="info" :show-icon="false" size="small">
    <template #header>
      <n-text strong>关联说明</n-text>
    </template>
    <n-text code size="small">
      Activity.CustomParma[0] -> DrawSkin.Id = {{ props.activityData.DrawSkin?.Id }}
    </n-text>
  </n-alert>
</n-gi>
```

**各页签关联说明示例**：

| 页签 | 关联说明内容 |
|------|-------------|
| 抽奖配置 | `Activity.CustomParma[0] -> DrawSkin.Id` |
| 掉落规则 | `DrawSkin.OnceDropRule -> DropRule.Id` |
| 掉落组 | `DropRule.DropGroup[] -> DropGroup.Id` |
| 掉落项 | `DropItem.DropGroup -> DropGroup.Id` |
| 次数奖励 | `LimitSkinTimesReward.ActIdStr == Activity.EActivityId` |
| 商店配置 | `Shop.ShopType == "ShopTypeSkinRaffle"` |
| 商品配置 | `ShopGoods.ShopType == "ShopTypeSkinRaffle"` |
| 皮肤展示 | `ItemHeroSkin.SkinItemId == DrawSkin.BigAwardItemId` |
| 皮肤信息 | `HeroSkinItem.SkinItemId == DrawSkin.BigAwardItemId` |
| Spine配置 | `HeroSkinSpine.SkinItemId == DrawSkin.BigAwardItemId` |
| 皮肤收藏 | `HeroSkinCollition.Type == HeroSkinItem.CollitionType` |

**进阶方案：关联关系链可视化**

在基础信息页签底部增加"关联关系链"卡片，用步骤条展示完整关联路径：

```vue
<!--
  activity-wiki-dev: 关联关系链可视化
  本步骤条由 activity-wiki-dev 技能生成，用于展示活动与配表的完整关联路径。
  如果此可视化方式不符合需求（如步骤过多、信息密度、交互方式），
  可反馈给 AI 以优化 activity-wiki-dev 技能中的"关联关系展示规范"章节。
-->
<n-gi :span="24">
  <n-card title="关联关系链" size="small" :bordered="false">
    <n-steps :current="999" size="small">
      <n-step title="活动表|Activity" description="CustomParma[0] = 1001" />
      <n-step title="皮肤抽奖|DrawSkin" description="Id = 1001" />
      <n-step title="掉落规则|DropRule" description="OnceDropRule = 2001" />
      <n-step title="掉落组|DropGroup" description="DropGroup = [3001, 3002]" />
      <n-step title="掉落项|DropItem" description="DropGroup in [3001, 3002]" />
    </n-steps>
  </n-card>
</n-gi>
```

**关联说明设计原则**：
1. **简洁性**：用 `表名.字段 -> 目标表.字段` 的格式表达
2. **可读性**：使用人类友好的字段名，而非代码变量名
3. **条件清晰**：如果是条件匹配（如枚举、类型），用 `==` 或 `in` 表达
4. **动态值**：展示实际的关联值（如 `Id = 1001`），而非仅展示字段名
5. **位置统一**：放在每个页签内容的最顶部，用户切换页签时第一眼就能看到
6. **可反馈性**：代码注释中标注技能标识，便于后续不满意时反馈优化技能

## 样式规范

**1. 活动类型标签颜色**

在 `getActivityTypeStyle` 函数中定义：

```typescript
const getActivityTypeStyle = (type: string) => {
  const styles: Record<string, { color: string; bgColor: string }> = {
    'ActTypeTest':     {color: '#d32f2f', bgColor: '#ffebee'},
    'ActTypeGacha':    {color: '#7b1fa2', bgColor: '#f3e5f5'},
    'ActTypeLogin':    {color: '#1976d2', bgColor: '#e3f2fd'},
    'ActTypeRecharge': {color: '#388e3c', bgColor: '#e8f5e9'},
    'ActTypeSkinRaffle': {color: '#e65100', bgColor: '#fff3e0'},
  };
  return styles[type] || {color: '#666', bgColor: '#f5f5f5'};
}
```

**2. 空值显示**
- 字符串空值：显示 `'-'`
- 数值空值：显示 `0` 或 `'无'`
- 布尔空值：使用 `formatBoolean()` 处理

**3. 工具函数**

```typescript
// 引入已有的格式化函数
import {formatArray, formatBoolean, formatItemArray} from "@shared/composables/use-format-utils";

// formatArray: 将数组格式化为字符串，如 [1,2,3] -> "1, 2, 3"
// formatBoolean: 将布尔值格式化为 "是"/"否"
// formatItemArray: 将道具数组格式化为 "道具IDx数量" 的tag列表
```


## 规则覆盖角标设计规范

### 设计目标
在活动Wiki中直观展示哪些表/字段被配表测试规则覆盖，覆盖程度如何。

### 角标规范

**Tab角标**：
- 位置：Tab标题右侧（n-badge superscript样式）
- 数字：表级规则数量（仅Enabled）
- 0个规则时不显示角标

**字段角标**：
- 位置：n-descriptions-item 的 label 旁
- 数字：列级规则数量 + 行级规则数量
- 0个规则时不显示角标

**颜色规范**：
- 绿色（`#18a058`）：默认状态/全部通过
- 红色（`#d03050`）：有校验失败

**Popover内容规范**：
- Tab角标：列出该表所有表级规则的名称和类型
- 字段角标：列出该字段所有列级规则的类型和参数摘要

### Naive UI组件使用

```vue
<n-popover trigger="hover" placement="top">
  <template #trigger>
    <n-badge :value="count" :color="color" :show-zero="false">
      <span>label文本</span>
    </n-badge>
  </template>
  <!-- 规则列表内容 -->
</n-popover>
```

### 数据来源
- 规则统计：后端 `GetRuleCoverage()` 方法，读取 `cases/excel_cases/*.json`
- 检查结果：前端全局状态中的检查结果数据


---

> 内容来源：SKILL.md 的"前端页面设计规范"章节。拆分时间：2026-04-30。
