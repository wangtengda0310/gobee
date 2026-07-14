# Frontend Layout Documentation Templates and Standards

> 本文件包含前端布局文档生成所需的完整模板和标准规范。
> 内容来源：从 SKILL.md 中拆分而来，便于独立维护和引用。

## 目录

- [Single Page Layout Template](#single-page-layout-template)
- [Index File Template](#index-file-template)
- [ASCII Layout Diagram Standards](#ascii-layout-diagram-standards)
- [Component Hierarchy Tree Standards](#component-hierarchy-tree-standards)
- [TypeScript Interface Standards](#typescript-interface-standards)
- [Data Flow Diagram Standards](#data-flow-diagram-standards)

---

## Single Page Layout Template

```markdown
# [Page Name] Layout

> File path: `frontend/src/pages/[FileName].vue`
> Route: `/[route]`

## Overview

[Brief description of page functionality - what the page does and its purpose]

## ASCII Layout Diagram

\`\`\`
┌─────────────────────────────────────────────────────────────┐
│ [Top Navigation]                                            │ ← Header
├────────────────┬────────────────────────────────────────────┤
│                │ ┌────────────────────────────────────────┐ │
│   Left Sidebar │ │ Tab: [Tab1] [Tab2] [Tab3]             │ │
│   (240px)      │ └────────────────────────────────────────┘ │
│                │ ┌────────────────────────────────────────┐ │
│   - Menu 1     │ │                                        │ │
│   - Menu 2     │ │         Tab Content Area               │ │
│                │ │         (scrollable)                   │ │
│                │ └────────────────────────────────────────┘ │
├────────────────┴────────────────────────────────────────────┤
│ [Statistics]                                                │ ← Footer
└─────────────────────────────────────────────────────────────┘
\`\`\`

## Layout Dimensions

| Area | Size | Notes |
|------|------|-------|
| Header | Height XXpx | Fixed |
| Footer | Height XXpx | Fixed |
| Left Sider | Width XXpx | Collapsible to XXpx |
| Content | Adaptive | Scrollable |

## Component Hierarchy Tree

\`\`\`
PageComponent (index.vue)
├── n-layout (root container)
│   ├── n-layout-header
│   │   └── n-menu (horizontal navigation)
│   │
│   ├── n-layout (has-sider)
│   │   ├── n-layout-sider (left sidebar)
│   │   │   ├── SearchInput
│   │   │   ├── FilterSwitches
│   │   │   └── TreeComponent
│   │   │
│   │   └── n-layout-content
│   │       └── TabContainer
│   │           ├── Tab1: ConfigPanel
│   │           │   └── FormComponents...
│   │           ├── Tab2: StepsPanel
│   │           │   └── DraggableCards...
│   │           └── Tab3: LogPanel
│   │
│   └── n-layout-footer
│       └── StatusBar
\`\`\`

## Component Mapping

### Main Structure

| Layout Area | Component/Element | File Location | Function |
|-------------|-------------------|---------------|----------|
| Root Container | n-layout | Page.vue:10 | Root layout wrapper |
| Header | n-layout-header | Page.vue:15 | Top navigation bar |
| Left Sidebar | n-layout-sider | Page.vue:20 | Side navigation |
| Content | n-layout-content | Page.vue:35 | Main content area |
| Footer | n-layout-footer | Page.vue:50 | Status bar |

### Sub-components

| Component | File Location | Props | Function |
|-----------|---------------|-------|----------|
| ComponentA | components/ComponentA.vue | prop1, prop2 | Description |
| ComponentB | components/ComponentB.vue | - | Description |

## TypeScript Interface Definitions

### Props Interface

\`\`\`typescript
interface PageProps {
  headerTitle?: string      // Header title text
  sideOption?: SideConfig   // Sidebar configuration
}

interface SideConfig {
  title: string
  collapsed?: boolean
}
\`\`\`

### State Interface

\`\`\`typescript
interface PageState {
  loading: Ref<boolean>
  data: Ref<DataType[]>
  selectedItem: Ref<string | null>
}

interface DataType {
  id: string
  name: string
  // ... other fields
}
\`\`\`

### API Response Interface (if applicable)

\`\`\`typescript
interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

// Example usage
interface UserInfoResponse extends ApiResponse<UserInfo> {}
\`\`\`

## Data Flow Diagram

### Phase 1: User Interaction Flow

\`\`\`
┌──────────────────────────────────────────────────────────────────────────┐
│                          User Interaction Flow                            │
└──────────────────────────────────────────────────────────────────────────┘

User Input                      Component State                Actions
─────────────────────────────────────────────────────────────────────────────

[Input Field] ────────────────> inputValue (ref)
    │                                │
    │ onInput                        │
    ▼                                ▼
handleInput() ────────────────> Validate & Update
    │
    │ [Submit Button]
    ▼
handleSubmit() ───────────────> loading.value = true
    │
    │ API Call
    ▼
Backend Service
\`\`\`

### Phase 2: State Update Flow

\`\`\`
┌──────────────────────────────────────────────────────────────────────────┐
│                          State Update Flow                                │
└──────────────────────────────────────────────────────────────────────────┘

Backend Response
    │
    ▼
Update State (data.value = response.data)
    │
    ├──► ComponentA (auto-reactive update)
    │         │
    │         └──► Re-render with new data
    │
    ├──► ComponentB (computed recalculation)
    │         │
    │         └──► Derived values updated
    │
    └──► ComponentC (watch effect)
              │
              └──► Side effects executed
\`\`\`

### Phase 3: Data Display Flow

\`\`\`
┌──────────────────────────────────────────────────────────────────────────┐
│                          Data Display Flow                                │
└──────────────────────────────────────────────────────────────────────────┘

State (data)
    │
    ├──► List Rendering (v-for)
    │         │
    │         └──► ItemComponent x N
    │                   │
    │                   └──► Display item details
    │
    ├──► Conditional Rendering (v-if)
    │         │
    │         └──► Show/hide based on conditions
    │
    └──► Computed Properties
              │
              └──► Filtered/sorted/transformed data
\`\`\`

## Key State

| State | File | Type | Default | Description |
|-------|------|------|---------|-------------|
| stateName | composables/State.ts | `Ref<Type>` | - | Description |
| computedValue | composables/Computed.ts | `Computed<Type>` | - | Computed from... |

## Backend Services (if applicable)

| Service | Method | Parameters | Return Type | Description |
|---------|--------|------------|-------------|-------------|
| ServiceName | methodName | (param: Type) | Promise<Result> | Description |

## Interactions

| Action | Trigger | Handler | Description |
|--------|---------|---------|-------------|
| Click button | @click="handleClick" | handleClick() | Description |
| Form submit | @submit="onSubmit" | onSubmit() | Description |
| Route change | watch($route) | onRouteChange() | Description |

## External Dependencies

### Shared Components

| Component | Location | Props | Usage |
|-----------|----------|-------|-------|
| SharedComponent | shared/components/ | prop1, prop2 | Description |

### Composables/Hooks

| Composable | Location | Exports | Purpose |
|------------|----------|---------|---------|
| useFeature | composables/useFeature.ts | state, actions | Description |

### Configuration Files

| Config | Location | Purpose |
|--------|----------|---------|
| configName | config/Config.ts | Description |

## Styling Notes

| Selector | Style | Description |
|----------|-------|-------------|
| .className | display: flex; | Layout method |
| #idName | position: fixed; | Positioning |

## Code Review Notes

### Potential Issues

> ⚠️ Document any code issues discovered during analysis

1. **Unused Variables/Functions**: List any defined but unused code
   - Example: `unusedVar` defined at line 20 but never referenced

2. **Inconsistent Patterns**: Note any inconsistent coding patterns
   - Example: Different styling applied to similar components

3. **Performance Concerns**: Identify potential performance issues
   - Example: Large list without virtualization
   - Example: Missing debounce on frequent updates

4. **Accessibility Issues**: Note any accessibility concerns
   - Example: Missing aria labels
   - Example: Insufficient color contrast

5. **Technical Debt**: Note areas that could be improved
   - Example: Commented-out code that should be removed
   - Example: Hard-coded values that should be configurable

### Best Practices Followed

- ✅ Proper TypeScript typing
- ✅ Composables for reusable logic
- ✅ Proper cleanup in onUnmounted

## Related Files

### Component Files
- `components/Xxx.vue` - Description

### Logic Files
- `composables/Xxx.ts` - Description

### Type Definitions
- `@bindings/xxx` - Auto-generated types from backend

---
**Verification Date**: YYYY-MM-DD
**Status**: Document matches / needs update for current code implementation
```

---

## Index File Template

用于生成 `frontend/docs/layout/CLAUDE.md` 索引文件。

```markdown
# [Project Name] Frontend Layout Documentation Index

## Document List

| Document | Page | Route | Description |
|----------|------|-------|-------------|
| [Normal-layout.md](./Normal-layout.md) | Layout Component | - | Outer layout |

## Tech Stack

- Framework: Vue 3 / React
- UI Library: Naive UI / Ant Design / Element Plus
- Router: Vue Router / React Router
- State: Pinia / Redux / Composables

## Common Components

| Component | Location | Usage |
|-----------|----------|-------|
| SharedComponent | shared/components/ | Description |

## Maintenance

- Update layout docs when page structure changes significantly
- Run verification when dependencies are updated
- Keep TypeScript interfaces in sync with actual code
```

---

## ASCII Layout Diagram Standards

### Common Layout Patterns

**Three-Section Layout (Header + Sider/Content + Footer)**
```
┌─────────────────────────────────────────────────────────────┐
│ Header (fixed height)                                       │
├────────────────┬────────────────────────────────────────────┤
│                │                                            │
│   Sider        │            Content                         │
│   (240px)      │            (adaptive, scrollable)          │
│   collapsible  │                                            │
│                │                                            │
├────────────────┴────────────────────────────────────────────┤
│ Footer (fixed height)                                       │
└─────────────────────────────────────────────────────────────┘
```

**Two-Column Layout with Anchor Navigation**
```
┌──────────────────────────────────────────────┬─────────────┐
│                                              │             │
│              Main Content                    │   Anchor    │
│              (scrollable)                    │   Nav       │
│                                              │   (120px)   │
│                                              │   fixed     │
└──────────────────────────────────────────────┴─────────────┘
```

**Tab-Based Layout**
```
┌─────────────────────────────────────────────────────────────┐
│ [Tab 1] [Tab 2] [Tab 3]                                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    Tab Content                              │
│                    (changes based on selection)             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Annotation Standards

- Use `←` to mark area names
- Use `(size)` for fixed dimensions (e.g., `(240px)`)
- Use `(adaptive)` for flexible areas
- Use `(scrollable)` for scrollable areas
- Use `(collapsible)` for collapsible sections
- Include component names in brackets `[ComponentName]`

---

## Component Hierarchy Tree Standards

组件层次树使用 tree 格式展示，示例如下：

```
PageComponent (filename.vue)
├── LayoutComponent
│   ├── HeaderSection
│   │   └── NavigationMenu
│   │       └── MenuItem x N
│   │
│   └── ContentArea
│       ├── SidebarComponent
│       │   ├── SearchInput
│       │   └── TreeView
│       │
│       └── MainContent
│           └── TabContainer
│               ├── Tab1: PanelA
│               └── Tab2: PanelB
```

---

## TypeScript Interface Standards

### Naming Conventions

- Props: `ComponentNameProps`
- State: `ComponentNameState` or describe purpose (e.g., `UserFormData`)
- API Response: `ApiResponseType` or `EndpointNameResponse`
- Generic types: Use meaningful names like `TData`, `TItem`

### Required Sections

1. **Props Interface** - All component props with types and descriptions
2. **State Interface** - Reactive state variables
3. **API Response Interface** - Backend data structures (if applicable)

---

## Data Flow Diagram Standards

### Three-Phase Approach

1. **User Interaction Flow**: How user actions trigger state changes
2. **State Update Flow**: How state changes propagate to components
3. **Data Display Flow**: How data is rendered in the UI

### Diagram Elements

- Use `───►` for direct flow
- Use `├──►` and `└──►` for branching
- Use boxes (`┌─┐`) for grouping related steps
- Label each phase clearly

---

> 内容来源：从 SKILL.md (frontend-layout-docs) 拆分而来，包含完整的文档模板、ASCII 布局图标准、组件层次树标准、TypeScript 接口标准和数据流图标准。
