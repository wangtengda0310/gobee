# Modal Components

> Parent: [../components.md](../components.md)
> Source: `src/pages/function-test/components/modals/`

## Overview

All modal components are located in `src/pages/function-test/components/modals/`. These modals are used for creating, editing, and managing test cases and categories in the function test page.

## Modal Components Tree

```
components/modals/
├── add-cate-modal.vue      # Add category dialog
├── add-case-modal.vue      # Add case dialog
├── rename-case-modal.vue   # Rename case dialog
├── rename-cate-modal.vue   # Rename category dialog
└── option-modal.vue        # Settings/options dialog
```

## Component Details

### AddCateModal

**Purpose**: Add a new category to the test case tree

**Location**: `src/pages/function-test/components/modals/add-cate-modal.vue`

**Props**:
```typescript
// None (uses v-model for visibility)
```

**Emits**:
- None (directly calls service functions)

**State**:
- `showModal` (Ref\<boolean\>) - Controls modal visibility

**Usage in parent**:
```vue
<!-- In index.vue -->
<AddCateModal />
```

**Trigger**: Right-click tree node → "添加分类"

**Logic**:
1. User clicks "添加分类" in context menu
2. Modal opens with name input field
3. User enters category name and confirms
4. `JsonCaseService.CreateCategory()` called
5. `dataRef` updated (TreeAndHistory.ts)
6. Modal closes

### AddCaseModal

**Purpose**: Add a new test case to a category

**Location**: `src/pages/function-test/components/modals/add-case-modal.vue`

**Props**: None

**Emits**: None

**State**:
- `showModal` (Ref\<boolean\>) - Controls modal visibility
- `categoryKey` (Ref\<string\>) - Parent category key

**Usage in parent**:
```vue
<!-- In index.vue -->
<AddCaseModal />
```

**Trigger**: Right-click category node → "添加用例"

**Logic**:
1. User clicks "添加用例" in context menu
2. Modal opens with name input field
3. User enters case name and confirms
4. `JsonCaseService.CreateJSONFile()` called
5. `dataRef` updated (TreeAndHistory.ts)
6. Modal closes

### RenameCaseModal

**Purpose**: Rename an existing test case

**Location**: `src/pages/function-test/components/modals/rename-case-modal.vue`

**Props**: None

**Emits**: None

**State**:
- `showModal` (Ref\<boolean\>)
- `nodeKey` (Ref\<string\>) - Node to rename
- `currentName` (Ref\<string\>) - Current name (default value)

**Usage in parent**:
```vue
<!-- In index.vue -->
<RenameCaseModal />
```

**Trigger**: Right-click case node → "重命名"

**Logic**:
1. User clicks "重命名" in context menu
2. Modal opens with current name pre-filled
3. User edits name and confirms
4. File renamed via `JsonCaseService.RenameJSONFile()`
5. `dataRef` updated
6. Modal closes

### RenameCateModal

**Purpose**: Rename an existing category

**Location**: `src/pages/function-test/components/modals/rename-cate-modal.vue`

**Props**: None

**Emits**: None

**State**:
- `showModal` (Ref\<boolean\>)
- `nodeKey` (Ref\<string\>) - Category node to rename
- `currentName` (Ref\<string\>) - Current name (default value)

**Usage in parent**:
```vue
<!-- In index.vue -->
<RenameCateModal />
```

**Trigger**: Right-click category node → "重命名"

**Logic**:
1. User clicks "重命名" in context menu
2. Modal opens with current name pre-filled
3. User edits name and confirms
4. Category renamed via service
5. `dataRef` updated
6. Modal closes

### OptionModal

**Purpose**: Configure test execution options and settings

**Location**: `src/pages/function-test/components/modals/option-modal.vue`

**Props**: None

**Emits**: None

**State**:
- `showModal` (Ref\<boolean\>)
- Various option refs (from composables/Option.ts)

**Usage in parent**:
```vue
<!-- In index.vue -->
<OptionModal />
```

**Trigger**: Top menu → "设置"

**Options**:
- Execution timeout settings
- Log verbosity
- Error handling behavior
- Other test execution parameters

## Usage Flow

```
User right-clicks tree node
    │
    ▼
nodeProps.onContextmenu() (Tree.ts)
    │
    ▼
n-dropdown menu shows (TreeDropDown.ts)
    │
    ├──► "添加分类" → showAddCateModal = true → AddCateModal opens
    │                    │
    │                    └──► Confirm → JsonCaseService.CreateCategory()
    │                              │
    │                              └──► dataRef updated → Tree re-renders
    │
    ├──► "添加用例" → showAddCaseModal = true → AddCaseModal opens
    │                  │
    │                  └──► Confirm → JsonCaseService.CreateJSONFile()
    │                            │
    │                            └──► dataRef updated → Tree re-renders
    │
    ├──► "重命名" → showRenameXxxModal = true → RenameModal opens
    │               │
    │               └──► Confirm → Rename service called
    │                         │
    │                         └──► dataRef updated → Tree re-renders
    │
    └──► "删除" → No modal, direct action → JsonCaseService.Delete()
                  │
                  └──► dataRef updated → Tree re-renders
```

## Modal State Management

All modals use a simple visibility pattern controlled by refs in `composables/Modals.ts`:

```typescript
// composables/Modals.ts
export const showAddCateModal = ref(false)
export const showAddCaseModal = ref(false)
export const showRenameCaseModal = ref(false)
export const showRenameCateModal = ref(false)
export const showOptionModal = ref(false)
```

In parent template:
```vue
<AddCateModal v-if="showAddCateModal" />
<AddCaseModal v-if="showAddCaseModal" />
<!-- ... -->
```

## Common Modal Structure

Each modal follows a similar pattern:

```
┌─────────────────────────────────────┐
│ Modal Title                         │
├─────────────────────────────────────┤
│                                     │
│  [Form fields for user input]       │
│                                     │
│  - Name input                       │
│  - (Additional options)             │
│                                     │
├─────────────────────────────────────┤
│            [取消]     [确认]         │
└─────────────────────────────────────┘
```

## File Locations

```
src/pages/function-test/components/modals/
├── add-cate-modal.vue       # 1680 bytes
├── add-case-modal.vue       # 1743 bytes
├── rename-case-modal.vue    # 1561 bytes
├── rename-cate-modal.vue    # 1563 bytes
└── option-modal.vue         # 5467 bytes (largest, most complex)
```

## Related Files

- **State Management**: [composables/Modals.ts](../../composables/Modals.ts)
- **Tree Data**: [composables/TreeAndHistory.ts](../../composables/TreeAndHistory.ts)
- **Context Menu**: [composables/TreeDropDown.ts](../../composables/TreeDropDown.ts)
- **Service Functions**: [composables/Func.ts](../../composables/Func.ts)
