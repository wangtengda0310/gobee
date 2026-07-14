# 代码模板参考

> 本文件包含活动Wiki开发中使用的所有代码模板，包括代码注释格式、Go/Vue文件模板、数据注册模板、前端页签模板等。
>
> 内容来源：SKILL.md 的"代码生成规范"和"开发步骤"章节。

## 目录

- [代码生成规范](#代码生成规范)
  - [Go代码注释格式](#go代码注释格式)
  - [Vue代码注释格式](#vue代码注释格式)
  - [新增字段注释格式](#新增字段注释格式)
- [开发步骤](#开发步骤)
  - [步骤1：新增配表解析模块](#步骤1新增配表解析模块如需要新表)
  - [步骤2：注册到数据容器](#步骤2注册到数据容器)
  - [步骤3：添加到ActivityCompleteData](#步骤3添加到activitycompletedata)
  - [步骤4：前端添加页签](#步骤4前端添加页签)
  - [步骤5：支持新活动类型](#步骤5支持新活动类型)

---

## 代码生成规范

**所有由 activity-wiki-dev 技能生成的代码必须包含技能标识注释**，以便日后技能升级时进行代码迁移和识别。

### Go代码注释格式

```go
// activity-wiki-dev: 活动Wiki开发技能生成
// 功能: {简要描述此文件的功能}
// 关联活动类型: {ActTypeXxx}
// 生成时间: {YYYY-MM-DD}

package xxx
```

### Vue代码注释格式

```vue
<!-- activity-wiki-dev: 活动Wiki开发技能生成 -->
<!-- 功能: {简要描述此组件/页签的功能} -->
<!-- 关联活动类型: {ActTypeXxx} -->
<!-- 生成时间: {YYYY-MM-DD} -->
```

### 新增字段注释格式

在现有结构体/组件中新增字段时：

```go
// activity-wiki-dev: 新增字段 - {字段用途说明}
Xxx *xxx.XxxDiff
```

```vue
<!-- activity-wiki-dev: 新增页签 - {页签用途说明} -->
<n-tab-pane name="xxx" ...>...</n-tab-pane>
```

---

## 开发步骤

### 步骤1：新增配表解析模块（如需要新表）

如果活动需要关联新的配表，先创建表解析模块：

**1.1 创建目录和文件**

在 `rain-resources-checker/mjs_excel/<表名>/` 下创建：
- `def.go` -- 列索引常量
- `diff_map.go` -- 解析逻辑

**1.2 def.go 模板**

```go
package <表名>

const (
    Id   = iota // ID列
    Name        // 名称列
    // ... 其他列
)
```

**1.3 diff_map.go 模板**

```go
package <表名>

import (
    "strconv"
    "strings"
    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-excel-checker/xlsx/check_internal"
    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-excel-checker/xlsx/excel_internal"
    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-resources-checker/mjs_excel/utils"
    "github.com/xuri/excelize/v2"
)

type XxxDiff struct {
    Id   int
    Name string
    // ... 其他字段
}

func (x XxxDiff) GetType() string {
    return "XxxDiff"
}

func (x XxxDiff) GetDisplayName() string {
    return x.Name
}

func GetXxxDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]XxxDiff, err error) {
    var cols [][]string
    if sheet, exist := sheetMap["表名|SheetName"]; exist {
        cols, err = sheet.GetCols("表名|SheetName")
        if err != nil {
            return nil, err
        }
    }

    startRow := excel_internal.MJS_FIXED_ROWS_NUM
    diffs := make([]XxxDiff, 0, 50)

    for i, idStr := range cols[Id][startRow:check_internal.AutoDetectEndIndex(cols, Id, startRow, 3)] {
        if id, err := strconv.Atoi(idStr); err != nil {
            continue
        } else {
            diff := XxxDiff{}
            diff.Id = id
            diff.Name = utils.GetCellValue(cols, Name, startRow+i)
            // ... 解析其他字段
            diffs = append(diffs, diff)
        }
    }
    return &diffs, nil
}
```

**关键约定：**
- Sheet名格式：`中文名|英文名`，如 `"商店表|Shop"`
- 使用 `excel_internal.MJS_FIXED_ROWS_NUM` 作为数据起始行（跳过表头）
- 使用 `check_internal.AutoDetectEndIndex` 自动检测数据结束位置
- 使用 `utils.GetCellValue` 安全获取单元格值
- 结构体必须实现 `GetType()` 和 `GetDisplayName()` 方法

### 步骤2：注册到数据容器

**2.1 在 `diff/interface.go` 的 DataContainer 中新增字段：**

```go
type DataContainer struct {
    // ... 现有字段 ...
    XxxDiff *[]xxx.XxxDiff  // 新增
}
```

**2.2 在 `mjs_excel/diff_excel_init.go` 的 InitDiffRefExcel 中初始化：**

```go
xxxDiff, err := xxx.GetXxxDiffMap(sheetMap)
if err != nil {
    return nil, err
}

excel := &diff.DataContainer{
    // ... 现有字段 ...
    XxxDiff: xxxDiff,  // 新增
}
```

### 步骤3：添加到ActivityCompleteData

**3.1 在 `activitywiki_def/def.go` 中新增关联字段：**

```go
type ActivityCompleteData struct {
    Basic *activity.ActivityDiff
    // ... 现有字段 ...
    Xxx *xxx.XxxDiff  // 新增关联表
}
```

**3.2 在 `activitywiki/format.go` 的 BuildActivityWikiDiff 中建立关联：**

```go
// 构建索引（在遍历活动之前）
xxxById := buildXxxIndex(container.XxxDiff)

// 在遍历活动中关联
data.Xxx = xxxById[关联条件]
```

添加对应的索引构建函数：

```go
func buildXxxIndex(xxxDiff *[]xxx.XxxDiff) map[int]*xxx.XxxDiff {
    index := make(map[int]*xxx.XxxDiff)
    if xxxDiff == nil {
        return index
    }
    for i := range *xxxDiff {
        x := &(*xxxDiff)[i]
        index[x.Id] = x
    }
    return index
}
```

### 步骤4：前端添加页签

**4.1 在 `activity-panel.vue` 的 header-extra 中添加打开Excel按钮**

在卡片右上角（ID左侧）增加一个打开Excel的按钮，点击时根据当前激活的页签打开对应的Excel文件。

```vue
<template #header-extra>
  <div style="display: flex; align-items: center; gap: 8px;">
    <!-- 打开Excel按钮 -->
    <n-button
        size="small"
        type="primary"
        ghost
        @click="handleOpenExcel"
    >
      <template #icon>
        <n-icon><TableIcon /></n-icon>
      </template>
      打开Excel
    </n-button>
    <div class="activity-id-badge">ID: {{ props.activityData.Basic?.Id }}</div>
  </div>
</template>
```

**4.2 页签与Excel文件映射关系**

```typescript
// 页签名称 -> Excel Sheet名 映射
const tabToSheetMap: Record<string, string> = {
  'basic': '活动表|Activity',
  'drawSkin': '皮肤抽奖|DrawSkin',
  'dropRule': '掉落规则表|DropRule',
  'timesRewards': '限时皮肤次数奖|LimitSkinTimesReward',
  'shop': '商店表|Shop',
  'shopGoods': '商品表|ShopGood',
  'heroSkinCollition': '英雄皮肤收藏|HeroSkinCollition',
  'itemHeroSkin': '武将皮肤展示表|ItemHeroSkin',
  'heroSkinItem': '英雄皮肤|HeroSkinItem',
  'heroSkinSpine': '英雄皮肤Spine|HeroSkinSpine',
}

// 打开Excel处理函数
const handleOpenExcel = () => {
  const sheetName = tabToSheetMap[activeTab.value]
  if (!sheetName) {
    return
  }
  // 调用后端API或Wails运行时打开Excel
  // 示例：OpenExcel(sheetName)
}
```

**4.3 在 `activity-panel.vue` 中添加新的 n-tab-pane**

```vue
<!-- 页签标题使用两行显示：第一行sheet名，第二行人类友好名称 -->
<n-tab-pane
    name="xxx"
    v-if="props.activityData.Xxx"
    :tab="() => h('div', { style: 'display: flex; flex-direction: column; align-items: center; line-height: 1.2;' },
        [h('span', { style: 'font-size: 10px; color: #888;' }, '表名|SheetName'),
         h('span', { style: 'font-size: 12px;' }, '显示名称')])"
>
  <n-grid :cols="24" :x-gap="16" :y-gap="16">
    <n-gi :span="24">
      <n-card title="标题" size="small" :bordered="false">
        <n-descriptions label-placement="left" :column="2" bordered>
          <n-descriptions-item label="字段名">
            <n-tag :bordered="false">{{ props.activityData.Xxx?.Field }}</n-tag>
          </n-descriptions-item>
        </n-descriptions>
      </n-card>
    </n-gi>
  </n-grid>
</n-tab-pane>
```

**页签标题规范：**
- 第一行（小字灰色）：`表名|SheetName`，如 `"商店表|Shop"`
- 第二行（正常字体）：人类友好名称，如 `"商店配置"`
- 使用 `v-if` 条件渲染，只在数据存在时显示

**新增页签的完整模板：**

```vue
<!-- 新页签模板 -->
<n-tab-pane
    name="xxx"
    v-if="props.activityData.Xxx"
    :tab="() => h('div', { style: 'display: flex; flex-direction: column; align-items: center; line-height: 1.2;' },
        [h('span', { style: 'font-size: 10px; color: #888;' }, '表名|SheetName'),
         h('span', { style: 'font-size: 12px;' }, '显示名称')])"
>
  <n-grid :cols="24" :x-gap="16" :y-gap="16">
    <!-- 单条信息展示 -->
    <n-gi :span="24">
      <n-card title="信息标题" size="small" :bordered="false">
        <n-descriptions label-placement="left" :column="2" bordered>
          <n-descriptions-item label="ID">
            <n-tag :bordered="false">{{ props.activityData.Xxx?.Id }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="名称">
            <n-text strong>{{ props.activityData.Xxx?.Name || '未命名' }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="布尔字段">
            <n-tag :bordered="false" :type="props.activityData.Xxx?.BoolField ? 'success' : 'default'">
              {{ formatBoolean(props.activityData.Xxx?.BoolField) }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="数组字段">
            <n-tag :bordered="false" type="info">{{ formatArray(props.activityData.Xxx?.ArrayField) }}</n-tag>
          </n-descriptions-item>
        </n-descriptions>
      </n-card>
    </n-gi>
  </n-grid>
</n-tab-pane>
```

### 步骤5：支持新活动类型

如果要支持新的活动类型（如 `ActTypeNewActivity`）：

1. 在 `BuildActivityWikiDiff` 中新增条件分支：

```go
if act.ActivityType == "ActTypeNewActivity" {
    // 该活动类型的关联逻辑
    if xxx, ok := xxxById[act.CustomParma[0]]; ok {
        data.Xxx = xxx
    }
}
```

2. 在前端 `getActivityTypeStyle` 中添加样式：

```typescript
const getActivityTypeStyle = (type: string) => {
  const styles: Record<string, { color: string; bgColor: string }> = {
    'ActTypeTest': {color: '#d32f2f', bgColor: '#ffebee'},
    'ActTypeNewActivity': {color: '#e65100', bgColor: '#fff3e0'}, // 新增
  };
  return styles[type] || {color: '#666', bgColor: '#f5f5f5'};
}
```

### 步骤5.5：时间匹配模式（多期数据活动）

对于有多期数据的活动（丹青阁、战令、结缘庭），需要使用时间匹配模式确定当前期，并构建三期数据（上一期/当前期/下一期）。

**适用场景**：
- 配表中有多条记录按 StartTime 排列，需要根据当前时间匹配"进行中的那一期"
- Activity.CustomParma[0] 指向的可能是过期的记录，不能直接使用
- 独立子系统（如战令）没有 ActivityType，全表数据按时间匹配

**后端实现**：

1. **构建排序切片**：

```go
// buildAllSortedXxx 构建全表按 StartTime 排序的切片
func buildAllSortedXxx(xxxDiff *[]xxx.XxxDiff) []*xxx.XxxDiff {
    var sorted []*xxx.XxxDiff
    if xxxDiff == nil {
        return sorted
    }
    for i := range *xxxDiff {
        sorted = append(sorted, &(*xxxDiff)[i])
    }
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].StartTime < sorted[j].StartTime
    })
    return sorted
}
```

2. **时间匹配找当前期**：

```go
// findCurrentXxx 从排序后的列表中找到当前期
// 优先选择正在进行中的（StartTime <= now <= EndTime），如果没有则选最新一期
func findCurrentXxx(sorted []*xxx.XxxDiff) *xxx.XxxDiff {
    if len(sorted) == 0 {
        return nil
    }
    now := time.Now()
    for _, x := range sorted {
        start, err1 := time.Parse("2006-01-02 15:04:05", x.StartTime)
        end, err2 := time.Parse("2006-01-02 15:04:05", x.EndTime)
        if err1 == nil && err2 == nil && (now.Equal(start) || now.After(start)) && (now.Equal(end) || now.Before(end)) {
            return x
        }
    }
    // 没有进行中的，返回最新一期（StartTime 最大的）
    return sorted[len(sorted)-1]
}
```

3. **获取上一期/下一期**：

```go
// findPrevNextXxx 在按 StartTime 排序的切片中找到当前记录的前一个和后一个
func findPrevNextXxx(sorted []*xxx.XxxDiff, currentId int) (prev *xxx.XxxDiff, next *xxx.XxxDiff) {
    for i, x := range sorted {
        if x.Id == currentId {
            if i > 0 {
                prev = sorted[i-1]
            }
            if i < len(sorted)-1 {
                next = sorted[i+1]
            }
            return
        }
    }
    return nil, nil
}
```

4. **在 CompleteData 中构建三期数据**：

```go
type XxxCompleteData struct {
    PrevXxx *xxx.XxxDiff
    Basic   *xxx.XxxDiff    // 当前期
    NextXxx *xxx.XxxDiff
    // ... 其他关联数据
}

// 使用
sortedXxx := buildAllSortedXxx(container.XxxDiff)
currentXxx := findCurrentXxx(sortedXxx)
if currentXxx != nil {
    prev, next := findPrevNextXxx(sortedXxx, currentXxx.Id)
    data := &XxxCompleteData{Basic: currentXxx}
    if prev != nil { data.PrevXxx = prev }
    if next != nil { data.NextXxx = next }
}
```

**前端展示**：

使用 PeriodCard 子组件展示三期数据（见设计模式D），在基础信息页签中纵向排列。


### 步骤6：规则覆盖角标

新增页签时，如果该页签对应的Sheet有规则配置，角标会自动显示。
如果新增了一个全新的Sheet（之前没有规则配置），需要确认是否需要新增对应的JSON规则文件。

**Tab角标自动工作**：`renderTabWithBadge` 从 `ruleCoverage.sheets[sheetName]` 读取统计。

**字段角标需要手动标注**：在 `n-descriptions-item` 中使用 `renderFieldWithBadge` 包裹label：

```vue
<n-descriptions-item>
  <template #label>
    <component :is="renderFieldWithBadge('FieldName', 'Sheet名|SheetName', '显示名称')" />
  </template>
  <n-tag :bordered="false">{{ props.activityData.Xxx?.Field }}</n-tag>
</n-descriptions-item>
```

**角标渲染函数使用 n-badge + n-popover**：
- `n-badge` 的 `value` 显示规则数量
- `n-badge` 的 `color` 控制颜色（绿/红）
- `n-popover` hover 显示规则列表详情

> 注意：只有有规则覆盖的字段才显示角标，没有规则的字段保持原样。


---

> 内容来源：SKILL.md 的"代码生成规范"和"开发步骤"章节。拆分时间：2026-04-30。
