# Composables - Logic Layer

> Parent: [index.md](./index.md)
> Source: `src/pages/function-test/composables/`

## Overview

Composables contain the reactive logic and state management for the function test page. They are organized by functionality and follow Vue 3 Composition API patterns.

## Composables Structure

```
composables/
├── Menu.ts                       # Top menu options
├── Tree.ts                       # Tree expand/check/drag/render
├── TreeSearch.ts                 # Search filtering
├── TreeDropDown.ts               # Right-click menu
├── TreeAndHistory.ts             # Tree data and history
├── use-case-data.ts              # Current case data
├── Modals.ts                     # Modal visibility states
├── Func.ts                       # Load/save/execute functions
├── FooterStatistic.ts            # Footer statistics
├── RobotTestLog.ts               # Log cache management
├── StepActionsAndAssetsSelect.ts  # Step action options
├── AssetMapTrans.ts              # Asset mapping
├── HeroAndCardsAndSkillsSelect.ts # Hero/card/skill selection
├── AssetProtoOptions.ts          # Asset proto definitions
├── Option.ts                     # Test options
└── Utils.ts                      # Utility functions
```

## State Management

### Global State (Shared across components)

详见 [index.md#Key State](./index.md#key-state)

### Modal States (Modals.ts)

| State | Type | Description |
|-------|------|-------------|
| `showAddCateModal` | Ref\<boolean\> | Add category modal |
| `showAddCaseModal` | Ref\<boolean\> | Add case modal |
| `showRenameCaseModal` | Ref\<boolean\> | Rename case modal |
| `showRenameCateModal` | Ref\<boolean\> | Rename category modal |
| `showOptionModal` | Ref\<boolean\> | Options modal |

## Composable Details

### Menu.ts

**Exports**: `menuOptions`, `activeKey`

顶部菜单配置（5 项：加载/保存/执行/停止用例、设置）。

### Tree.ts

**Exports**: `expandedKeysRef`, `checkedKeysRef`, `nodeProps`, `handleDrop`, etc.

树组件交互逻辑：展开/折叠、复选框、拖拽重排、自定义渲染、右键菜单触发。

### TreeSearch.ts

**Exports**: `pattern`, `showCasesDesc`, `showIrrelevantNodes`

搜索过滤：将 `pattern` 和 `showIrrelevantNodes` 传递给 n-tree。

### TreeDropDown.ts

**Exports**: `showDropdownRef`, `optionsRef`, `xRef`, `yRef`, `handleSelect`

右键菜单：分类节点（添加分类/用例、重命名、删除），用例节点（复制、重命名、删除）。

### TreeAndHistory.ts

**Exports**: `dataRef`

用例树数据和操作历史。

### use-case-data.ts

**Exports**: `nowCaseData`

当前编辑用例数据。类型定义见 [use-case-data.ts](../../composables/use-case-data.ts)。

### Modals.ts

弹窗可见性状态控制。

### Func.ts

加载/保存/执行/停止用例的核心 CRUD 操作。

### FooterStatistic.ts

**Exports**: `footerStatisticCaseNum`, `footerStatisticStepNum`

底部统计：计算用例总数和步骤总数。

### RobotTestLog.ts

**Exports**: `insertLogCache`, `logCache`

机器人执行日志缓存管理。

### StepActionsAndAssetsSelect.ts

**Exports**: `actionsSelectOption`

步骤动作类型和断言选项（动作类型、卡牌、技能、武将选择）。

### 辅助文件

| 文件 | 用途 |
|------|------|
| AssetMapTrans.ts | 断言 ID 到名称映射 |
| HeroAndCardsAndSkillsSelect.ts | 武将/卡牌/技能选择选项 |
| AssetProtoOptions.ts | 断言 protobuf 定义 |
| Option.ts | 测试执行配置 |
| Utils.ts | 工具函数 |

## Related Files

- **Main Page**: [../index.vue](../index.vue)
- **Components**: [../components.md](../components.md)
- **Modals**: [../components/modals.md](../components/modals.md)
