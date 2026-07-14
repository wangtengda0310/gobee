---
name: hero-wiki-filter
description: 指导在武将 Wiki 检查页面（HeroWikiCheck.vue）中添加新的过滤条件。当用户请求在武将检查页面添加筛选功能、过滤条件、或需要根据武将属性进行筛选时使用此 skill。
---

# 武将 Wiki 检查页面 - 过滤条件扩展

## 何时使用

- 用户请求在武将 Wiki 检查页面添加新的筛选功能
- 用户需要根据武将属性（如性别、类型等）进行过滤
- 用户提到"过滤"、"筛选"、"filter"相关需求
- 需要为 HeroWikiCheck.vue 页面扩展过滤条件

## 核心文件

- 主页面: `frontend/src/pages/HeroWikiCheck.vue`
- 武将面板: `frontend/src/components/HeroWiki/HeroPanel.vue`
- 类型定义: `frontend/src/scripts/HeroPanel/TypeDef.ts`
- HeroDiff 类型: `bindings/rain-resources-checker/mjs_excel/hero.ts`

## 架构概述

过滤系统由三部分组成：
1. **筛选状态变量** - 存储用户选择的筛选值（ref）
2. **filteredHeroes 计算属性** - 根据筛选条件过滤武将列表
3. **UI 筛选控件** - 用户交互界面（Naive UI 组件）

## 添加过滤条件的四个步骤

### 步骤 1: 添加筛选状态变量

在 `// 筛选状态` 注释下添加新的 ref：

```typescript
// 筛选状态
const searchName = ref('')
const filterCountry = ref<string[]>([])
// ... 现有变量
const filterYourNewField = ref<YourType>(defaultValue)
```

**常见类型选择**：
- 布尔筛选: `ref<boolean | null>(null)` - 三态筛选（未筛选/是/否）
- 单选筛选: `ref<string | null>(null)` - 枚举或字符串
- 多选筛选: `ref<string[]>([])` - 多个值

### 步骤 2: 在 filteredHeroes 中添加过滤逻辑

在 `filteredHeroes` 计算属性的 filter 回调中添加判断：

```typescript
const filteredHeroes = computed(() => {
  if (!diffExcels.value?.HeroDiff) return []

  return diffExcels.value.HeroDiff.filter((hero) => {
    // ... 现有过滤条件

    // 新的过滤条件
    if (filterYourNewField.value !== defaultValue && hero.YourField !== filterYourNewField.value) {
      return false
    }

    return true
  })
})
```

**关键点**：先检查筛选值是否为默认值（未筛选状态），再进行比较。

### 步骤 3: 添加 UI 筛选控件

在筛选条件区域的 `<n-space>` 中添加对应的 UI 组件。

#### 布尔值筛选（复选框）

```vue
<n-checkbox
  :checked="filterYourBool === true"
  @update:checked="filterYourBool = $event ? true : null"
>
  显示名称
</n-checkbox>
```

#### 枚举/字符串筛选（下拉选择）

```vue
<n-select
  v-model:value="filterYourEnum"
  placeholder="占位文本"
  clearable
  style="width: 150px"
  :options="yourOptions"
/>
```

**选项数据源**：

静态选项：
```typescript
const yourOptions = [
  { label: '选项1', value: 'value1' },
  { label: '选项2', value: 'value2' },
]
```

动态选项（从数据源获取）：
```typescript
const yourOptions = computed(() => {
  if (!diffExcels.value?.YourDataSource) return []
  return diffExcels.value.YourDataSource.map(item => ({
    label: item.Name,
    value: item.EnumValue
  }))
})
```

#### 多选筛选

```vue
<n-select
  v-model:value="filterYourMulti"
  multiple
  placeholder="占位文本"
  clearable
  style="width: 150px"
  :options="yourOptions"
/>
```

### 步骤 4: 更新删除武将列表（如需要）

如果新的过滤条件需要影响删除武将列表的显示，需要修改两处 `v-if` 条件：

1. 删除武将面板列表
2. 删除武将导航列表

```vue
<template v-if="removedHeroesDetail.length > 0 && (filterDiffType === null || filterDiffType === 'removed')">
```

## 完整示例：添加"性别"筛选

### 1. 添加筛选状态变量

```typescript
// 在 // 筛选状态 注释下
const filterGender = ref<string | null>(null)
```

### 2. 添加选项

```typescript
const genderOptions = [
  { label: '男', value: 'male' },
  { label: '女', value: 'female' },
]
```

### 3. 在 filteredHeroes 中添加过滤逻辑

```typescript
const filteredHeroes = computed(() => {
  return diffExcels.value.HeroDiff.filter((hero) => {
    // ... 其他条件
    if (filterGender.value !== null && hero.Gender !== filterGender.value) {
      return false
    }
    return true
  })
})
```

### 4. 添加 UI 控件

```vue
<n-select
  v-model:value="filterGender"
  placeholder="性别"
  clearable
  style="width: 100px"
  :options="genderOptions"
/>
```

## 注意事项

1. **三态筛选**：布尔筛选使用 `boolean | null` 类型，`null` 表示未筛选
2. **响应式更新**：使用 `ref` 或 `computed` 确保筛选条件变化时视图自动更新
3. **性能考虑**：`filteredHeroes` 是计算属性，会自动缓存
4. **数据源检查**：动态选项需要检查数据源是否存在
5. **导航同步**：导航列表使用 `filteredAnchors` 计算属性，会自动同步

## 相关文档

详细实现说明请参考：
- `rain-qa-func/docs/武将Wiki检查-过滤条件扩展指南.md`
- `rain-qa-func/docs/武将Wiki检查-实现详解.md`
