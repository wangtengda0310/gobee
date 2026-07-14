# Function Test Page Layout

> File path: `src/pages/function-test/index.vue`
> Route: `/function-test`

## Overview

功能测试用例编辑器页面，用于配置、编辑和执行测试用例。采用经典的 Header-Sider-Content-Footer 四段式布局，左侧为用例树形导航，右侧为 Tab 页签式内容区，支持用例配置、步骤编辑和执行日志查看。

## ASCII Layout Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [加载用例] [保存用例] [执行用例] [停止用例] [设置] [其他选项▼]              │ ← Header (34px)
├────────────────┬────────────────────────────────────────────────────────────┤
│ [搜索框]       │ ┌────────────────────────────────────────────────────────┐ │
│ [展示全部/仅过滤]│ │ [用例配置] [用例步骤] [执行日志]                       │ │ ← Tab栏
│ [显示描述/隐藏] │ └────────────────────────────────────────────────────────┘ │
│                │ ┌────────────────────────────────────────────────────────┐ │
│   用例树       │ │                                                        │ │
│   (Tree)       │ │                   Tab 内容区                           │ │
│                │ │                   (可滚动)                              │ │
│   ├─ 分类1 (3) │ │                                                        │ │
│   │  ├─ 用例A  │ │                                                        │ │
│   │  └─ 用例B  │ │                                                        │ │
│   └─ 分类2     │ └────────────────────────────────────────────────────────┘ │
│                │                                                            │
│ [右键菜单]      │                                                            │
├────────────────┴────────────────────────────────────────────────────────────┤
│ [XX 条用例] [XX 个动作] [进度条] [当前运行用例] [断言错误数目]              │ ← Footer (64px)
└─────────────────────────────────────────────────────────────────────────────┘
```

## Layout Dimensions

| Area | Size | Description |
|------|------|-------------|
| Header | Height 34px | Top toolbar with menu |
| Footer | Height 64px | Bottom status bar |
| Left Sider | Width 240px (50px collapsed) | Test case tree。**移动端(pointer:coarse)默认折叠 50px**（`isTouchDevice` + `:default-collapsed`，详见 [Android 适配规范](../../android-adaptation.md)） |
| Content | Adaptive | Tab panel area |
| Anchor Nav | Width 120px | Step navigation (steps tab only) |

## Data Flow

```
User clicks tree node
    │
    ▼
nodeProps.onClick() (Tree.ts)
    │
    ▼
nowCaseData.value = option (use-case-data.ts)
    │
    ├──► InitYanWuPanel reactive update
    ├──► StepsPanel reactive update
    └──► n-anchor reactive update
```

```
User right-clicks tree node
    │
    ▼
nodeProps.onContextmenu() (Tree.ts)
    │
    ▼
Show n-dropdown menu
    │
    ▼
handleSelect() (TreeDropDown.ts)
    │
    ├──► 添加分类 → showAddCateModal
    ├──► 添加用例 → showAddCaseModal
    ├──► 重命名 → showRenameXxxModal
    └──► 删除 → JsonCaseService.DeleteJSONFile
```

```
Robot execution
    │
    ▼
Events.Emit('robotLog', ...) (Backend)
    │
    ▼
Events.On('robotLog') (index.vue:45)
    │
    ▼
insertLogCache() (RobotTestLog.ts)
    │
    └──► RobotTestLog component update
```

## Interactions

| Action | Trigger | Handler | Description |
|--------|---------|----------|-------------|
| Click tree node | @node-click | nodeProps.onClick() | Load case data to right panel |
| Right-click node | onContextmenu | nodeProps.onContextmenu() | Show context menu |
| Drag tree node | @drop | handleDrop() | Reorder or move case |
| Click "加载用例" | menuOptions.onClick | handleLoad() (Func.ts) | Load cases from work directory |
| Click "保存用例" | menuOptions.onClick | Save all cases | |
| Click "执行用例" | menuOptions.onClick | Execute current case | |
| Click "停止用例" | menuOptions.onClick | Stop execution | |
| Search input | v-model | pattern | Filter tree |
| Toggle switch | v-model | showIrrelevantNodes / showCasesDesc | Display options |
| Tab switch | n-tabs | Active tab change | Switch editor view |

## Related Files

### Components (components/)
- [components/init-yanwu-panel.vue](../components/init-yanwu-panel.vue) - 初始演武配置面板
- [components/steps-panel.vue](../components/steps-panel.vue) - 测试步骤编辑面板
- [components/robot-test-log.vue](../components/robot-test-log.vue) - 执行日志面板
- [components/asset-card.vue](../components/asset-card.vue) - 断言卡片组件
- [components/footer-case-log-statistic.vue](../components/footer-case-log-statistic.vue) - 底部运行状态统计
- **[components/modals/](./modals.md)** - Modal dialogs (see modals.md)

### Composables (composables/)
- [composables/Menu.ts](../../composables/Menu.ts) - 顶部菜单配置
- [composables/Tree.ts](../../composables/Tree.ts) - 用例树逻辑
- [composables/TreeSearch.ts](../../composables/TreeSearch.ts) - 树搜索过滤
- [composables/TreeDropDown.ts](../../composables/TreeDropDown.ts) - 右键菜单逻辑
- [composables/TreeAndHistory.ts](../../composables/TreeAndHistory.ts) - 树数据和历史
- [composables/use-case-data.ts](../../composables/use-case-data.ts) - 当前用例数据
- [composables/Modals.ts](../../composables/Modals.ts) - 弹窗状态控制
- [composables/Func.ts](../../composables/Func.ts) - 加载/保存/执行功能
- [composables/FooterStatistic.ts](../../composables/FooterStatistic.ts) - 底部统计数据
- [composables/RobotTestLog.ts](../../composables/RobotTestLog.ts) - 机器人日志缓存
- [composables/StepActionsAndAssetsSelect.ts](../../composables/StepActionsAndAssetsSelect.ts) - 步骤动作选择

## Sub-page Layouts

### 用例配置 Tab (InitYanWuPanel)

```
┌─────────────────────────────────────────────────────────────────┐
│ 用例名称                                    [用例描述输入框]    │
├─────────────────────────────────────────────────────────────────┤
│ 负责人: [输入框]                                                 │
├─────────────────────────────────────────────────────────────────┤
│ 牌堆组: [数字输入]                                               │
│ 摸牌堆: [多选下拉] [顺序调整开关]                                │
│ 弃牌堆: [多选下拉] (禁用)                                        │
├─────────────────────────────────────────────────────────────────┤
│ 座位 1                                    [拖动] [×]            │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ [武将选择] [身份选择] [势力选择]  初始技能: xxx              │ │
│ │ 初始手牌: [多选下拉] [顺序调整开关]                          │ │
│ │ 初始装备: [多选下拉] [顺序调整开关]                          │ │
│ │ 触发装备: [多选下拉] [顺序调整开关]                          │ │
│ │ 初始卜卦: [多选下拉] [顺序调整开关]                          │ │
│ │ 删除技能: [多选下拉] [顺序调整开关]                          │ │
│ │ 增加技能: [下拉选择] [已选技能标签...]                       │ │
│ │ 技能牌区: [下拉选择] [已选技能标签...]                       │ │
│ │   └─ xxx牌区: [多选下拉]                                    │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ 座位 2 ...                                                      │
│ [增加武将]                                                      │
└─────────────────────────────────────────────────────────────────┘
```

### 用例步骤 Tab (StepsPanel)

```
┌──────────────────────────────────────────┬─────────────┐
│ 动作 1               [拖动] [应用智能描述->] [描述输入] [复制] │ ← 锚点导航
│ ┌────────────────────────────────────────────────────────────┐ │   (120px)
│ │ [动作类型] [座位选择] [等待秒数/确认开关/...]              │ │
│ │ [技能选择] [目标选择] [卡牌选择] [当xx牌打出] [超时时间]   │ │
│ ├────────────────────────────────────────────────────────────┤ │
│ │ 断言 1                                    [拖动] [×]       │ │
│ │ [AssetCard 组件]                                          │ │
│ ├────────────────────────────────────────────────────────────┤ │
│ │ [新增断言] [新增]                                          │ │
│ └────────────────────────────────────────────────────────────┘ │
│ 动作 2 ...                                                     │
├──────────────────────────────────────────┴─────────────┘
```

### 执行日志 Tab (RobotTestLog)

```
┌─────────────────────────────────────────────────────────────────┐
│ [用例名1(红色/白色)] [用例名2] ...                              │ ← Tab栏
├─────────────────────────────────────────────────────────────────┤
│ [时间], ID[x], Case[x], name[x], [Level], 消息内容             │
│ [时间], ID[x], Case[x], name[x], [Level], 消息内容             │
│ ...                                                             │
│ (可滚动，自动滚动到底部)                                        │
└─────────────────────────────────────────────────────────────────┘
```

## Special Notes

1. **Tree Node Types**: 分为 Categories(分类) 和 Cases(用例) 两种 levelType
2. **Modification Indicator**: 被修改的用例/分类会以绿色显示并带有 `*` 前缀
3. **Event Subscription**: 页面加载时订阅 `robotLog` 事件 (index.vue:45)，卸载时取消订阅
4. **Drag and Drop**: 使用 Naive UI n-tree 的 draggable 属性实现用例和步骤的拖拽排序
5. **Anchor Navigation**: 用例步骤 Tab 右侧有锚点导航 (120px宽)，可快速跳转到指定步骤
6. **Modal Components**: 所有弹窗组件在 components/modals/ 目录下，详见 [modals.md](./modals.md)
