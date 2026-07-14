# Hero Wiki Check Page Layout

> File path: `src/pages/hero-wiki-check/index.vue`
> Route: `/hero-wiki-check`

## Overview

武将 Wiki 检查页面，用于对比 Excel 配置中的武将数据变化。采用简单的单列布局，顶部为配置和筛选区域，下方为可滚动的武将面板列表，右侧为锚点导航。

## ASCII Layout Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [配置区域]                                                     [执行检查] [保存结果] │ ← Config Card
├─────────────────────────────────────────────────────────────────────────────┤
│ [总变化: X] [新增: X] [删除: X] [修改: X]                                    │ ← Summary Card
├─────────────────────────────────────────────────────────────────────────────┤
│ [搜索武将] [势力] [新武将] [抽奖] [已开放]                        筛选: X/Y    │ ← Filter Card
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ HeroPanel 1                                                       │  │
│  │ ┌─────────────────────────────────────────────────────────────────┐ │  │
│  │ │ HeroDiffDisplay | BuffDisplay | DropDisplay | SkillDisplay  │ │  │
│  │ └─────────────────────────────────────────────────────────────────┘ │  │
│  ├─────────────────────────────────────────────────────────────────────┤  │
│  │ HeroPanel 2                                                       │  │
│  │ ...                                                              │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│  (可滚动)                                                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │ ← Anchor Nav
│  [武将1]                                                            │   (120px)
│  [武将2]                                                            │
│  [武将3]                                                            │
│  ...                                                                │
└─────────────────────────────────────────────────────────────────────┘
```

## Layout Dimensions

| Area | Size | Description |
|------|------|-------------|
| Config Card | Auto (fixed) | 配置输入和操作按钮 |
| Summary Card | Auto (fixed) | 差异统计标签 |
| Filter Card | Auto (fixed) | 筛选条件输入 |
| Hero List | Adaptive (scrollable) | 武将面板容器 |
| Anchor Nav | 120px (fixed) | 快速跳转锚点 |

## Component Tree Structure

```
pages/hero-wiki-check/index.vue                    # Main page container
├── Config Card (n-card)                          # 顶部配置区域
│   ├── PathConfigInput → @shared/path-config-input/index.vue
│   ├── n-button (执行检查) → index.vue:244
│   ├── n-button (保存结果) → index.vue:247
│   └── n-text (error msg)
├── Summary Card (n-card)                         # 差异统计（条件显示）
│   └── n-tag × 4 → index.vue:262-288
│       ├── 总变化
│       ├── 新增
│       ├── 删除
│       └── 修改
├── Filter Card (n-card)                          # 筛选条件区域
│   ├── n-input (searchName) → index.vue:294
│   ├── n-select (filterCountry) → index.vue:300
│   ├── n-checkbox (新武将) → index.vue:308
│   ├── n-checkbox (抽奖) → index.vue:311
│   └── n-checkbox (已开放) → index.vue:314
├── Hero List Container (div)                     # 可滚动列表
│   └── n-scrollbar → index.vue:324
│       ├── HeroPanel (v-for) → components/hero-panel.vue
│       │   ├── HeroDiffDisplay → components/hero-diff-display.vue
│       │   ├── BuffDisplay → components/buff-display.vue
│       │   ├── DropDisplay → components/drop-display.vue
│       │   └── [各种 Display 组件]
│       └── HeroPanel (removed, v-for)
└── Anchor Container (div)                         # 锚点导航（条件显示）
    └── n-scrollbar → index.vue:349
        └── n-anchor-link (v-for)
```

## Component File Mapping

| Component | File Path | Line | Description |
|-----------|-----------|------|-------------|
| Main Page | pages/hero-wiki-check/index.vue | 1-400+ | 页面容器 |
| PathConfigInput | @shared/path-config-input/index.vue | - | 路径配置输入组件 |
| HeroPanel | components/hero-panel.vue | - | 武将面板主组件 |
| HeroDiffDisplay | components/hero-diff-display.vue | - | 武将差异显示 |
| BuffDisplay | components/buff-display.vue | - | 技能加成显示 |
| DropDisplay | components/drop-display.vue | - | 掉落显示 |

## Data Flow

```
User clicks "执行检查"
    │
    ▼
runCheck() (index.vue:34)
    │
    ▼
HeroWikiResCheckService.Check(excelDir, oldJsonPath)
    │
    ▼
diffExcels.value = res
    │
    ├──► Summary Card reactive update
    ├──► Filter options computed update
    └──► Hero list renders

User filters heroes
    │
    ▼
filteredHeroes computed (index.vue:180)
    │
    ├──► Hero list updates
    └──► Anchor list updates

User clicks "保存结果"
    │
    ▼
saveResult() (index.vue:49)
    │
    ▼
HeroWikiResCheckService.Save(oldJsonPath, diffExcels)
```

## Key State

| State | Type | Description |
|-------|------|-------------|
| `excelDir` | Ref\<string\> | Excel 配置目录路径 |
| `oldJsonPath` | Ref\<string\> | 历史数据 JSON 文件路径 |
| `diffExcels` | Ref\<DataContainer \| null\> | 检查结果数据 |
| `isLoading` | Ref\<boolean\> | 执行检查加载状态 |
| `isSaving` | Ref\<boolean\> | 保存结果加载状态 |
| `errorMsg` | Ref\<string\> | 错误消息 |
| `searchName` | Ref\<string\> | 武将名称搜索 |
| `filterCountry` | Ref\<string[]\> | 势力筛选 |
| `filterIsNewHero` | Ref\<boolean \| null\> | 新武将筛选 |
| `filterIsGacha` | Ref\<boolean \| null\> | 抽卡武将筛选 |
| `filterIsOpen` | Ref\<boolean \| null\> | 已开放筛选 |
| `filterDiffType` | Ref\<DiffTypeFilter\> | diff类型筛选 |

## Computed Properties

| Property | Type | Description |
|----------|------|-------------|
| `hasDiffResult` | ComputedRef\<boolean\> | 是否有 diff 结果 |
| `diffSummary` | ComputedRef\<Summary\> | 全局 diff 统计 |
| `countryOptions` | ComputedRef\<Option[]\> | 国家选项列表 |
| `removedHeroesDetail` | ComputedRef\<RemovedHeroItem[]\> | 删除的武将详情列表 |
| `filteredHeroes` | ComputedRef\<HeroDiff[]\> | 过滤后的武将列表 |
| `filteredAnchors` | ComputedRef\<{seq, hero}[]\> | 过滤后的锚点列表 |

## Interactions

| Action | Trigger | Handler | Description |
|--------|---------|----------|-------------|
| 执行检查 | Button click | runCheck() | 调用后端检查服务 |
| 保存结果 | Button click | saveResult() | 保存检查结果到文件 |
| 搜索武将 | Input v-model | searchName | 实时过滤武将列表 |
| 势力筛选 | Select v-model | filterCountry | 按势力过滤 |
| 新武将筛选 | Checkbox | filterIsNewHero | 按是否新武将过滤 |
| 点击统计标签 | Tag click | setDiffTypeFilter() | 按 diff 类型筛选 |
| 点击锚点 | Anchor click | Scroll to hero | 滚动到对应武将 |

## Related Files

### Components (components/)
- [components/hero-panel.vue](../components/hero-panel.vue) - 武将面板主组件
- [components/hero-diff-display.vue](../components/hero-diff-display.vue) - 武将基础信息显示
- [components/buff-display.vue](../components/buff-display.vue) - 技能加成显示
- [components/drop-display.vue](../components/drop-display.vue) - 掉落显示

### Composables (composables/)
- [composables/hero-wiki.types.ts](../composables/hero-wiki.types.ts) - 类型定义

### Shared Components
- [@shared/path-config-input/index.vue](../../../../shared/components/path-config-input) - 路径配置输入组件

## Special Notes

1. **差异数据结构**: 使用 `DataContainer` 包含多种 Diff 类型（HeroDiff, SkillDiff, BuffDiff, DropDiff 等）
2. **删除武将处理**: 删除的武会有单独的显示区域，带有分割线标识
3. **筛选联动**: 多个筛选条件可以组合使用，实时更新列表
4. **锚点导航**: 仅在有 diff 结果时显示，快速跳转到指定武将
5. **共享组件**: 使用 `@shared` 路径引用共享组件
6. **条件渲染**: 统计卡片和锚点导航都是条件渲染（基于 `hasDiffResult`）
