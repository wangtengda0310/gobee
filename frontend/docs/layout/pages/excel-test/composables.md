# Excel Test - Composables (Logic Layer)

> Parent: [index.md](./index.md)

## Overview

Composables 包含 Excel 测试页面的响应式逻辑和状态管理。

## Composables Structure

```
composables/
├── menu.ts                        # 顶部菜单配置
├── use-tree.ts                    # 树交互逻辑
├── use-tree-search.ts             # 搜索过滤逻辑
├── use-tree-drop-down.ts          # 右键菜单逻辑
├── use-tree-and-history.ts        # 树数据和历史记录
├── use-excel-check-log.ts         # 检查日志管理
├── use-excel-check-data.ts        # Excel 检查数据管理
├── option.ts                      # 配置管理
├── func.ts                        # 功能函数
└── params-components/             # 参数组件目录
    ├── all-base-params.ts
    ├── boolean-params.ts
    ├── chs-only-params.ts
    ├── cross-reference-params.ts
    ├── custom-params.ts
    ├── date-consistency-params.ts
    ├── date-duration-params.ts
    ├── date-params.ts
    ├── date-range-params.ts
    ├── enum-params.ts
    ├── increase-params.ts
    ├── not-empty-params.ts
    ├── numeric-params.ts
    ├── numeric-range-params.ts
    ├── pin-yin-chs-params.ts
    ├── regex-params.ts
    ├── resource-params.ts
    ├── rich-text-params.ts
    ├── server-or-client-params.ts
    ├── special-format-params.ts
    ├── string-params.ts
    ├── unique-params.ts
    ├── weight-sum-params.ts
    └── excel-rules-template.ts
```

## State Management

### Global State (Shared across components)

详见 [index.md#Key State](./index.md#key-state)

### Composable Details

#### menu.ts

**Exports**: `menuOptions`, `activeKey`

**Purpose**: 顶部菜单配置

**Actions**:
- 加载配置 → `loadAllSettings()` + 刷新树
- 保存配置 → `saveExcelConfig()`
- 执行检查 → 执行 Excel 检查
- 停止检查 → 停止检查
- 设置 → 打开 `option-modal`

#### use-tree.ts

**Exports**: `expandedKeysRef`, `checkedKeysRef`, `nodeProps`, `handleDrop`, etc.

**Purpose**: 树组件交互逻辑

**Features**:
- 展开/折叠节点
- 复选框选择
- 拖拽重排序
- 自定义节点渲染（带图标）

#### use-tree-search.ts

**Exports**: `pattern`, `showExcelCheckDesc`, `showIrrelevantNodes`

**Purpose**: 搜索和显示过滤

**Logic**: 根据 `pattern` 过滤 `dataRef`

#### use-tree-and-history.ts

**Exports**: `dataRef`, `succAndFailSheetNum`, 历史记录相关

**Purpose**: 树数据和历史记录管理

**Features**:
- 树数据加载和保存
- Sheet 成功/失败统计
- 检查结果历史

#### use-excel-check-log.ts

**Exports**: `checkLog`, 日志相关函数

**Purpose**: 检查日志管理

**Features**:
- 日志缓存
- 日志筛选
- 日志导出

#### use-excel-check-data.ts

**Exports**: Excel 检查数据相关

**Purpose**: Excel 检查数据管理

**Features**:
- 检查规则配置
- 规则参数管理
- 检查结果数据

#### option.ts

**Exports**: `ExcelResourceDir`, `ExcelCaseDir`, `saveExcelConfig`

**Purpose**: 配置管理

**Features**:
- 配置路径管理
- 配置保存/加载
- 配置持久化

#### func.ts

**Exports**: 功能函数

**Purpose**: 功能函数集合

**Features**:
- 加载配置
- 保存配置
- 执行检查
- 停止检查

### Params Components (params-components/)

参数组件目录，包含 24 个检查规则参数配置文件。详见目录内文件列表。

## Data Flow Diagrams

### 检查执行流程

```
User clicks "执行检查"
    │
    ▼
menuOptions.onClick()
    │
    ▼
func.ts - executeCheck()
    │
    ├──► 获取选中的 Sheet
    ├──► 获取配置的规则
    │
    ▼
ExcelCheckService.Check()
    │
    ▼
检查结果 → checkLog
    │
    └──► ExcelCheckLog component update
```

### 配置保存流程

```
User edits path input
    │
    ▼
@blur event
    │
    ▼
saveExcelConfig()
    │
    ├──► SettingsService.SetExcelCheckConfig()
    │
    ▼
Config saved to backend
```

### 树节点选择流程

```
User clicks tree node
    │
    ▼
nodeProps.onClick()
    │
    ▼
Load Sheet/Rule data
    │
    ├──► ExcelCheckManager update (if 负责人 tab)
    ├──► ExcelCheckPanel update (if 用例配置 tab)
    │
    └──► nowSelectedNode update
```

## Related Files

### Components
- [components.md](./components.md) - 组件层级文档

### Main Page
- [index.md](./index.md) - 主页面布局

## Notes

- 参数组件目录包含 24+ 个参数配置文件
- 每个参数组件对应一种检查规则类型
- 参数组件动态加载，根据 `RuleType` 决定使用哪个
- 树组件支持拖拽排序
- 检查结果实时更新到日志组件
- 配置自动保存到后端
