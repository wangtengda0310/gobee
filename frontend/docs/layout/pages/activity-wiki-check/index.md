# Activity Wiki Check Page Layout

> File path: `src/pages/activity-wiki-check/index.vue`
> Route: `/ActivityWiki`

## Overview

活动 Wiki 检查页面，展示 Excel 配置中的活动数据和战令数据。顶部为配置和筛选区域，中间为可滚动的活动卡片列表 + 战令卡片，右侧为锚点导航。

## ASCII Layout Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ [配置区域]  Excel目录路径输入 + 执行检查                                          │ ← Config Card
├─────────────────────────────────────────────────────────────────────────────────┤
│ [搜索活动名称]  [活动类型▼]  ☑显示页签                       筛选结果: X/Y       │ ← Filter Card
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │ 1. 活动名称  [ActType标签] [EActivityId标签]          [打开Excel] [ID:X] │  │  ← ActivityPanel
│  │ ┌─────────────────────────────────────────────────────────────────────┐  │  │
│  │ │ [基础信息] [抽奖配置] [掉落规则] [商店配置] ...                    │  │  │  ← n-tabs
│  │ │                                                                     │  │  │
│  │ │  基础信息页签:                                                      │  │  │
│  │ │  ┌──────────────────────┐  ┌──────────────────────┐               │  │  │
│  │ │  │ Activity 信息        │  │ 关联关系链 (n-steps) │               │  │  │
│  │ │  └──────────────────────┘  └──────────────────────┘               │  │  │
│  │ └─────────────────────────────────────────────────────────────────────┘  │  │
│  ├───────────────────────────────────────────────────────────────────────────┤  │
│  │ 2. 活动名称  [ActType标签] [EActivityId标签]          [打开Excel] [ID:X] │  │  ← ActivityPanel
│  │ ...                                                                     │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                 │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │ 1. 战令名称  [战令]                                   [打开Excel] [ID:X] │  │  ← SeasonPassPanel
│  │ ┌─────────────────────────────────────────────────────────────────────┐  │  │
│  │ │ [基础信息] [战令礼包] [等级奖励] [战令任务]                         │  │  │  ← n-tabs
│  │ │                                                                     │  │  │
│  │ │  基础信息页签:                                                      │  │  │
│  │ │  ┌──────────────────────┐                                           │  │  │
│  │ │  │ 上一期 PeriodCard    │ (灰色左边框)                               │  │  │
│  │ │  └──────────────────────┘                                           │  │  │
│  │ │  ┌──────────────────────┐                                           │  │  │
│  │ │  │ 本期 PeriodCard      │ (蓝色左边框, 高亮)                        │  │  │
│  │ │  └──────────────────────┘                                           │  │  │
│  │ │  ┌──────────────────────┐                                           │  │  │
│  │ │  │ 下一期 PeriodCard    │ (灰色左边框)                               │  │  │
│  │ │  └──────────────────────┘                                           │  │  │
│  │ │  ┌──────────────────────┐                                           │  │  │
│  │ │  │ 关联关系链 (n-steps) │                                           │  │  │
│  │ │  └──────────────────────┘                                           │  │  │
│  │ └─────────────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│  (统一可滚动区域)                                                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │ ← Anchor Nav
│  1. 活动名称                                                         │   (150px)
│  2. 活动名称                                                         │
│  ...                                                                  │
│  N. 战令                                                              │
└─────────────────────────────────────────────────────────────────────┘
```

## Layout Dimensions

| Area | Size | Description |
|------|------|-------------|
| Config Card | Auto (fixed) | Excel目录路径 + 执行检查按钮 |
| Filter Card | Auto (fixed) | 搜索、活动类型筛选、ShowTab复选框 |
| Activity/SeasonPass List | Adaptive (scrollable) | 活动卡片 + 战令卡片，统一滚动区域 |
| Anchor Nav | 150px (fixed) | 快速跳转锚点 |

## Component Tree Structure

```
pages/activity-wiki-check/index.vue                   # Main page container
├── Config Card (n-card)                             # 顶部配置区域
│   ├── PathConfigInput → @shared/path-config-input/index.vue
│   └── n-button (执行检查)
├── Filter Card (n-card)                             # 筛选条件
│   ├── n-input (搜索活动名称)
│   ├── n-select (活动类型)
│   ├── TooltipCheckbox (显示页签) → @shared/components/tooltip-checkbox/index.vue
│   └── n-text (筛选结果计数)
├── Activity List (n-scrollbar)                      # 统一可滚动区域
│   ├── ActivityPanel × N                            # 活动卡片
│   │   ├── n-card (hoverable, segmented)
│   │   │   ├── header: 序号.活动名 + ActType标签 + EActivityId标签
│   │   │   ├── header-extra: 打开Excel + ID badge
│   │   │   └── content: n-tabs
│   │   │       ├── 基础信息 (Activity描述 + 关联关系链 n-steps)
│   │   │       ├── 抽奖配置 (DrawSkin 三期 PeriodCard × 3)
│   │   │       ├── 掉落规则/组/项 (DropRule/Group/Item)
│   │   │       ├── 结缘亭 (DrawPet 三期 PeriodCard × 3 + Pet列表)
│   │   │       ├── 次数奖励 (LimitSkinTimesReward 表格)
│   │   │       ├── 商店/商品 (Shop/ShopGoods)
│   │   │       ├── 皮肤展示 (ItemHeroSkin)
│   │   │       ├── 皮肤信息 (HeroSkinItem/Spine/Collition)
│   │   │       └── 累充奖励 (AccumulatedRecharge)
│   ├── SeasonPassPanel × 1                          # 战令卡片（只展示当前期）
│   │   ├── n-card (hoverable, segmented)
│   │   │   ├── header: 序号.战令名 + "战令"标签
│   │   │   ├── header-extra: 打开Excel + ID badge
│   │   │   └── content: n-tabs
│   │   │       ├── 基础信息 (SeasonPassPeriodCard × 3 + 关联关系链 n-steps)
│   │   │       ├── 战令礼包 (SeasonPassBag 表格)
│   │   │       ├── 等级奖励 (SeasonPassReward 表格)
│   │   │       └── 战令任务 (SeasonPassTask 表格)
│   └── n-empty (无匹配数据时显示)
└── Anchor Nav (n-anchor)                            # 右侧锚点导航
    └── n-anchor-link × (活动数 + 战令)
```

## Data Flow

```
composables/use-activity-wiki.ts
  ├── runCheck() → 后端 ActivityWikiCheckService → activityWikiDiff
  ├── filteredActivities (computed) → ActivityPanel × N
  ├── filteredSeasonPasses (computed) → SeasonPassPanel × 1
  ├── filteredAnchors (computed) → 右侧锚点导航
  └── ruleCoverage (ref) → provide → BadgeLabel (角标)
```

## Card Header Specification

| 组件 | header | header-extra |
|------|--------|--------------|
| ActivityPanel | 序号.活动名 + ActivityType彩色标签 + EActivityId蓝色标签 | 打开Excel按钮 + ID badge |
| SeasonPassPanel | 序号.战令名 + "战令"蓝色标签 | 打开Excel按钮 + ID badge |

---

**Generated**: 2026-05-13 | **Skill**: manual creation
