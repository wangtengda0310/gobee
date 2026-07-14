# Function Test - Component Hierarchy

> Parent: [index.md](./index.md)

## Component Details

### InitYanWuPanel

**Purpose**: Configure initial test case settings (武将、手牌、装备等)

**Location**: `components/init-yanwu-panel.vue`

**Sub-components**: None (uses Naive UI form components)

**Props**: N/A (uses shared `nowCaseData` ref)

**Emits**: None (directly mutates `nowCaseData`)

**State dependencies**:
- `nowCaseData` (use-case-data.ts)
- `actionsSelectOption` (StepActionsAndAssetsSelect.ts) - For card/hero/skill options

**Key features**:
- Dynamic form generation (1-8 seats)
- Multi-select for cards/skills
- Order adjustment switches
- Skill card zones

### StepsPanel

**Purpose**: Edit test case steps and assertions

**Location**: `components/steps-panel.vue`

**Sub-components**:
- `n-scrollbar` (2x) - Scroll containers
- `n-anchor` - Step navigation (right side, 120px)
- `ActionItem` (v-for) - Individual step
  - `AssetCard` (v-for) - Assertion cards

**Props**: None (uses shared `nowCaseData` ref)

**Emits**: None (directly mutates `nowCaseData.caseSteps`)

**State dependencies**:
- `nowCaseData` (use-case-data.ts)
- `actionsSelectOption` (StepActionsAndAssetsSelect.ts)

**Layout**:
```
┌─────────────────────────────┬──────────┐
│  ActionItem (v-for)         │ n-anchor │
│  ┌─────────────────────┐    │  (120px) │
│  │ Step controls       │    │          │
│  ├─────────────────────┤    │          │
│  │ AssetCard (v-for)   │    │          │
│  └─────────────────────┘    │          │
└─────────────────────────────┴──────────┘
```

### RobotTestLog

**Purpose**: Display robot execution logs in real-time

**Location**: `components/robot-test-log.vue`

**Sub-components**:
- `n-tabs` - Log tabs per test case
- `n-tab-pane` - Per test case log content
- `n-log` - Log output display

**Props**: None

**Emits**: None

**State dependencies**:
- `logCache` (RobotTestLog.ts) - Reactive log cache
- Event: `robotLog` - Subscribed in index.vue:45

**Data flow**:
```
Wails Backend
    │
    ▼
Events.Emit('robotLog', level, message)
    │
    ▼
Events.On('robotLog') (index.vue:45)
    │
    ▼
insertLogCache(level, message) (RobotTestLog.ts)
    │
    ▼
logCache reactive update
    │
    ▼
RobotTestLog component re-render
```

### FooterCaseLogStatistic

**Purpose**: Display test execution statistics

**Location**: `components/footer-case-log-statistic.vue`

**Sub-components**: Naive UI progress/statistic components

**Props**: None

**Emits**: None

**State dependencies**:
- `footerStatisticCaseNum` (FooterStatistic.ts)
- `footerStatisticStepNum` (FooterStatistic.ts)
- Execution state (running/paused/error)

## Event Flow

| Component | Event | Handler |
|-----------|-------|---------|
| n-tree | @node-click | nowCaseData update |
| n-tree | @drop | handleDrop() (Tree.ts) |
| n-dropdown | @select | Modal open/action |
| StepsPanel | step add/remove/update | nowCaseData mutation |
| AssetCard | assertion add/remove/update | nowCaseData mutation |
| InitYanWuPanel | seat add/remove | nowCaseData mutation |

## Modal Components

See [modals.md](./modals.md) for detailed modal component documentation.

### Quick Reference

| Modal | Trigger | Purpose | Location |
|-------|---------|---------|----------|
| AddCateModal | 右键 → 添加分类 | Add new category | modals/add-cate-modal.vue |
| AddCaseModal | 右键 → 添加用例 | Add new test case | modals/add-case-modal.vue |
| RenameCaseModal | 右键 → 重命名 (用例) | Rename test case | modals/rename-case-modal.vue |
| RenameCateModal | 右键 → 重命名 (分类) | Rename category | modals/rename-cate-modal.vue |
| OptionModal | 菜单 → 设置 | Configure options | modals/option-modal.vue |

