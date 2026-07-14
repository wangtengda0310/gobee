# 武将 Wiki 检查页面 - 过滤条件扩展指南

本文档说明如何在 `HeroWikiCheck.vue` 页面中添加新的过滤条件。

## 架构概述

过滤系统由三部分组成：
1. **筛选状态变量** - 存储用户选择的筛选值
2. **filteredHeroes 计算属性** - 根据筛选条件过滤武将列表
3. **UI 筛选控件** - 用户交互界面

## 现有过滤条件

| 变量名 | 类型 | 说明 | UI 组件 |
|--------|------|------|---------|
| `searchName` | `string` | 名称搜索 | `n-input` |
| `filterCountry` | `string[]` | 势力筛选 | `n-select multiple` |
| `filterIsNewHero` | `boolean \| null` | 新武将 | `n-checkbox` |
| `filterIsGacha` | `boolean \| null` | 抽卡武将 | `n-checkbox` |
| `filterIsOpen` | `boolean \| null` | 已开放 | `n-checkbox` |
| `filterDiffType` | `'added' \| 'modified' \| 'removed' \| null` | diff 类型 | `n-tag` (顶部统计) |

## 添加新过滤条件步骤

### 步骤 1: 添加筛选状态变量

在 `// 筛选状态` 注释下添加新的 `ref`：

```typescript
const filterYourNewField = ref<YourType>(defaultValue)
```

### 步骤 2: 在 filteredHeroes 中添加过滤逻辑

```typescript
const filteredHeroes = computed(() => {
  return diffExcels.value.HeroDiff.filter((hero) => {
    // ... 现有过滤条件
    if (filterYourNewField.value !== defaultValue && hero.YourField !== filterYourNewField.value) {
      return false
    }
    return true
  })
})
```

### 步骤 3: 添加 UI 筛选控件

| 筛选类型 | 组件 | 关键属性 |
|----------|------|----------|
| 布尔值 | `n-checkbox` | `:checked="filter === true"`，`@update:checked` 设置 `true/null` |
| 枚举/字符串 | `n-select` | `v-model:value`、`clearable`、`:options` |
| 多选 | `n-select multiple` | 同上，加 `multiple` |

选项数据：静态数组或从 `diffExcels.value?.YourDataSource` 动态计算。

### 步骤 4: 更新删除武将列表（如需要）

修改删除武将列表和导航的 `v-if` 条件。显示条件：`removedHeroesDetail.length > 0 && (filterDiffType === null || filterDiffType === 'removed')`

## 示例：添加"性别"筛选

```typescript
// 1. 添加状态变量 + 选项
const filterGender = ref<string | null>(null)
const genderOptions = [{ label: '男', value: 'male' }, { label: '女', value: 'female' }]

// 2. 在 filteredHeroes 中添加判断
if (filterGender.value !== null && hero.Gender !== filterGender.value) return false
```

```vue
<n-select v-model:value="filterGender" placeholder="性别" clearable :options="genderOptions" />
```

## 注意事项

1. **三态筛选**：布尔筛选通常使用 `boolean | null` 类型，`null` 表示未筛选，`true/false` 表示筛选特定值

2. **响应式更新**：使用 `ref` 或 `computed` 确保筛选条件变化时视图自动更新

3. **性能考虑**：`filteredHeroes` 是计算属性，会自动缓存，避免不必要的重复计算

4. **数据源可用性**：动态选项需要检查数据源是否存在，如：
   ```typescript
   if (!diffExcels.value?.CountryDiff) return []
   ```

5. **导航同步**：添加过滤条件后，导航列表使用 `filteredAnchors` 计算属性，会自动同步

## 相关文件

- `frontend/src/pages/HeroWikiCheck.vue` - 主页面，包含所有筛选逻辑
- `frontend/src/components/HeroWiki/HeroPanel.vue` - 武将面板组件
- `bindings/rain-resources-checker/mjs_excel/hero.ts` - HeroDiff 类型定义
