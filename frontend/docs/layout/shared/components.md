# Shared Components - 共享组件

> Parent: [../CLAUDE.md](../CLAUDE.md)

## Overview

共享组件位于 `src/shared/components/`，供多个页面复用。

## Component List

```
shared/components/
├── path-config-input/
│   └── index.vue           # 路径配置输入组件
└── status-bar/
    └── index.vue           # 状态栏组件
```

## Component Details

### PathConfigInput

**Purpose**: 可复用的配置路径输入框组件，统一多个页面的路径配置界面

**Location**: `src/shared/components/path-config-input/index.vue`

**Props**:
```typescript
interface Props {
  excelDir: string                    // Excel 目录路径 (v-model:excel-dir)
  excelLabel?: string                 // 第一个输入框的 label，默认"配表"
  excelPlaceholder?: string           // 第一个输入框的 placeholder
  secondValue: string                 // 第二个输入框的值 (v-model:second-value)
  secondLabel?: string                // 第二个输入框的 label
  secondPlaceholder?: string          // 第二个输入框的 placeholder
  onSave?: () => void                 // 失焦时的保存回调
  inputWidth?: string                 // 输入框宽度，默认"180px"（仅 inline 模式）
  size?: 'small' | 'medium' | 'large' // 尺寸，默认"small"
  layout?: 'inline' | 'flex'          // 布局模式，默认"inline"
}
```

**Emits**:
- `update:excelDir` - Excel 目录路径更新
- `update:secondValue` - 第二个值更新

**Layout Modes**:
- **inline**（默认）: 使用 n-space 水平排列，固定宽度输入框
- **flex**: 输入框自适应宽度（用于 hero-voice-resource-check）

**Usage Examples**:

```vue
<!-- Label 模式 -->
<PathConfigInput
  v-model:excel-dir="excelDir"
  v-model:second-value="caseDir"
  excel-label="配表"
  second-label="用例"
/>

<!-- Placeholder 模式 -->
<PathConfigInput
  v-model:excel-dir="excelDir"
  v-model:second-value="jsonPath"
  excel-label=""
  second-label=""
  excel-placeholder="Excel 配置目录路径"
  second-placeholder="JSON 文件路径"
/>

<!-- 带保存回调 -->
<PathConfigInput
  v-model:excel-dir="excelDir"
  v-model:second-value="caseDir"
  :on-save="saveConfig"
/>

<!-- flex 布局模式 -->
<PathConfigInput
  v-model:excel-dir="excelDir"
  v-model:second-value="cardDir"
  excel-label="配表位置"
  second-label="Card文件夹位置"
  layout="flex"
/>
```

**Used By**:
- `pages/excel-test/index.vue` - 配表目录、用例目录
- `pages/hero-wiki-check/index.vue` - Excel 配置目录、JSON 历史文件
- `pages/hero-voice-resource-check/index.vue` - 配表位置、Card 文件夹位置

**State Dependencies**:
- None (stateless component, controlled by parent)

**Component Tree**:
```
PathConfigInput
├── (layout="inline") n-space
│   ├── n-input-group
│   │   ├── n-input-group-label (excelLabel)
│   │   └── n-input (v-model:excelDir)
│   └── n-input-group
│       ├── n-input-group-label (secondLabel)
│       └── n-input (v-model:secondValue)
└── (layout="flex") template
    ├── span (excelLabel || '配表')
    ├── n-input (flex: 1, v-model:excelDir)
    ├── span (secondLabel)
    └── n-input (flex: 1, v-model:secondValue)
```

### StatusBar

**Purpose**: 应用状态栏组件，显示应用版本信息和最近提交记录

**Location**: `src/shared/components/status-bar/index.vue`

**Layout Structure**:
```
┌─────────────────────────────────────────────────────────────────┐
│ rain-qa-func              自定义信息插槽          最近更新: abc1234 │
│                           (可扩展)                [已构建]        │
└─────────────────────────────────────────────────────────────────┘
```

**Component Tree**:
```
StatusBar
├── .status-left
│   └── n-tag → "rain-qa-func"
├── .status-center
│   └── slot name="custom-info" → 自定义信息插槽
└── .status-right
    └── n-tooltip (trigger="hover")
        ├── trigger: span.version-tip
        │   ├── "最近更新: {{ latestHash }}"
        │   └── span.build-time-badge (v-if="hasBuildTime") → "已构建"
        └── content: .commit-list
            ├── .build-info (v-if="hasBuildTime")
            │   ├── "构建时间: {{ buildTime }}"
            │   └── (divider)
            └── .commit-item (v-for commit in commits)
                ├── commit-hash (e.g., "abc1234")
                ├── commit-msg
                └── commit-meta (author · date)
```

**Data**:
```typescript
interface CommitInfo {
  hash: string        // 提交哈希
  message: string     // 提交消息
  author: string      // 作者
  date: string        // 日期
}

interface BuildInfo {
  commitHash: string
  commitMsg: string
  buildTime: string   // 构建时间（生产构建时才有）
}
```

**Backend Services**:
- `VersionService.GetRecentCommits(5)` - 获取最近 5 条提交
- `VersionService.GetBuildInfo()` - 获取构建信息

**Used By**:
- `layouts/normal-layout/index.vue` - Footer

**Slots**:
- `custom-info` - 自定义信息区（可扩展）

**Notes**:
- 左侧显示应用名称
- 中间可插入自定义内容（通过 slot）
- 右侧显示版本信息，hover 显示详细提交记录
- 生产构建时显示"已构建"徽章
- 自动加载最近 5 条提交记录

## Usage Summary

| Component | Used By Pages | Purpose |
|-----------|---------------|---------|
| PathConfigInput | excel-test, hero-wiki-check, hero-voice-resource-check | 路径配置输入 |
| StatusBar | All pages (via NormalLayout) | 版本信息显示 |

## State Management

这些是无状态或有状态但独立的组件，由父组件通过 props 和事件控制。

## Related Files

- `src/shared/components/path-config-input/index.vue`
- `src/shared/components/status-bar/index.vue`
- `src/layouts/normal-layout/index.vue` - StatusBar 使用位置
