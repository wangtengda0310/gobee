# Hero Wiki Check - Component Hierarchy

> Parent: [index.md](./index.md)

## Component Dependency Tree

```
index.vue (Main Page)
│
├─── Config Section (Fixed)
│    └── n-card (config-card)
│         ├── PathConfigInput → @shared/path-config-input
│         ├── n-button (执行检查)
│         ├── n-button (保存结果)
│         └── n-text (error message)
│
├─── Summary Section (Fixed, Conditional)
│    └── n-card (global-diff-summary)
│         └── n-tag × 4
│              ├── 总变化
│              ├── 新增
│              ├── 删除
│              └── 修改
│
├─── Filter Section (Fixed)
│    └── n-card (filter-card)
│         ├── n-input (searchName)
│         ├── n-select (filterCountry)
│         ├── n-checkbox (新武将)
│         ├── n-checkbox (抽奖)
│         └── n-checkbox (已开放)
│
├─── Hero List (Scrollable)
│    └── n-scrollbar
│         ├── HeroPanel (v-for, normal heroes)
│         │    ├── HeroDiffDisplay
│         │    ├── BuffDisplay
│         │    ├── DropDisplay
│         │    └── [其他 Display 组件]
│         │
│         └── n-divider (removed heroes section)
│              └── HeroPanel (v-for, removed heroes)
│
└─── Anchor Navigation (Fixed, Conditional)
     └── n-scrollbar
          └── n-anchor-link (v-for)
```

## Component Details

### PathConfigInput

**Purpose**: 路径配置输入组件

**Location**: `@shared/path-config-input/index.vue`

**Props**:
- `v-model:excel-dir` - Excel 配置目录路径
- `v-model:second-value` - 历史数据 JSON 文件路径
- `excel-label` - 第一个输入框标签
- `second-label` - 第二个输入框标签
- `excel-placeholder` - 第一个输入框占位符
- `second-placeholder` - 第二个输入框占位符
- `input-width` - 输入框宽度

**Usage**:
```vue
<PathConfigInput
  v-model:excel-dir="excelDir"
  v-model:second-value="oldJsonPath"
  excel-label=""
  second-label=""
  excel-placeholder="Excel 配置目录路径"
  second-placeholder="历史数据 JSON 文件路径(可选)"
  input-width="280px"
/>
```

### HeroPanel

**Purpose**: 武将面板主组件，显示单个武将的完整信息

**Location**: `components/hero-panel.vue`

**Props**:
- `seq` - 序号
- `hero-info` - 武将信息 (HeroDiff)
- `diff-excels` - 差异数据容器
- `diff-index-map` - 差异索引映射
- `hero-wiki-data` - Wiki 数据 (可选)
- `is-removed` - 是否为已删除武将

**Sub-components**:
- `HeroDiffDisplay` - 武将基础信息显示
- `BuffDisplay` - 技能加成显示
- `DropDisplay` - 掉落显示
- 其他 Display 组件

**Features**:
- 显示武将基础属性（名称、势力、性别、点数等）
- 显示技能加成变化
- 显示掉落变化
- 支持已删除武将的特殊显示样式

### HeroDiffDisplay

**Purpose**: 武将基础信息显示组件

**Location**: `components/hero-diff-display.vue`

**Features**:
- 显示武将名称、ID
- 显示势力、性别
- 显示点数、血量等属性
- 高亮显示变化的字段

### BuffDisplay

**Purpose**: 技能加成显示组件

**Location**: `components/buff-display.vue`

**Features**:
- 显示技能列表
- 高亮新增/修改/删除的技能
- 显示技能效果变化

### DropDisplay

**Purpose**: 掉落显示组件

**Location**: `components/drop-display.vue`

**Features**:
- 显示掉落列表
- 高亮新增/修改/删除的掉落
- 显示掉落率变化

## Props Flow

```
index.vue (Parent)
    │
    ├──► HeroPanel (v-for)
    │       ├── hero-info: HeroDiff (from diffExcels.HeroDiff)
    │       ├── diff-excels: DataContainer
    │       ├── diff-index-map: transMap()
    │       └── (removed heroes use hero-wiki-data)
    │
    └──► Filter inputs (v-model)
         ├── searchName
         ├── filterCountry
         ├── filterIsNewHero
         ├── filterIsGacha
         └── filterIsOpen
```

## Event Flow

```
Component          Event                          Handler
────────────────────────────────────────────────────────────────
n-button           @click                         → runCheck()
n-button           @click                         → saveResult()
n-input            v-model:update                  → searchName update
n-select           v-model:update                  → filterCountry update
n-checkbox          @update:checked                 → filter update
n-tag              @click                         → setDiffTypeFilter()
```

## Nested Directory Structure

此页面没有嵌套的组件目录，所有组件都直接位于 `components/` 目录下。

```
src/pages/hero-wiki-check/
├── index.vue
├── components/
│   ├── hero-panel.vue
│   ├── hero-diff-display.vue
│   ├── buff-display.vue
│   └── drop-display.vue
└── composables/
    └── hero-wiki.types.ts
```

对应的文档结构：
```
frontend/docs/layout/pages/hero-wiki-check/
├── index.md
└── components.md
```

## File Sizes

```
components/
├── hero-panel.vue         # 52540 bytes (largest, most complex)
├── hero-diff-display.vue  # 17501 bytes
├── buff-display.vue        # 11838 bytes
└── drop-display.vue        # 24950 bytes
```

## Related Files

- **Main Page**: [../index.vue](../index.vue)
- **Shared Components**: [../../../../shared/components.md](../../../../shared/components.md)
- **Composables**: [../composables/hero-wiki.types.ts](../composables/hero-wiki.types.ts)
