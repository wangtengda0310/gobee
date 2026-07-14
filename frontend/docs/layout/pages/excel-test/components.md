# Excel Test - Component Hierarchy

> Parent: [index.md](./index.md)

## Component Details

### PathConfigInput

**Purpose**: 路径配置输入组件

**Location**: `@shared/components/path-config-input/index.vue`

**Props**:
- `v-model:excel-dir` - Excel 目录路径
- `v-model:second-value` - 用例目录路径
- `excel-label` - "配表"
- `second-label` - "用例"
- `:on-save` - 保存配置回调

**Usage**:
```vue
<PathConfigInput
  v-model:excel-dir="ExcelResourceDir"
  v-model:second-value="ExcelCaseDir"
  excel-label="配表"
  second-label="用例"
  :on-save="saveConfig"
/>
```

### ExcelCheckManager

**Purpose**: 负责人管理面板

**Location**: `components/excel-check-manager.vue`

**Features**:
- 管理 Sheet 负责人
- 显示未分配负责人的 Sheet
- 批量分配负责人

### ExcelCheckPanel

**Purpose**: 检查规则配置面板

**Location**: `components/excel-check-panel.vue`

**Features**:
- 显示检查规则树
- 配置规则参数
- 动态加载不同规则类型的参数组件
- 支持启用/禁用规则

**Sub-components**:
- `n-tree` - 规则树结构
- `n-form` - 规则参数表单
- 动态参数组件 - 根据 `RuleType` 动态加载

### ExcelCheckLog

**Purpose**: 执行日志显示面板

**Location**: `components/excel-check-log.vue`

**Features**:
- 显示检查结果日志
- 支持筛选不同级别的日志
- 高亮错误信息
- 支持导出日志

### OptionModal

**Purpose**: 设置弹窗

**Location**: `components/option-modal.vue`

**Features**:
- 配置检查选项
- 配置显示选项

## Nested Directory Structure

```
src/pages/excel-test/
├── index.vue
├── components/
│   ├── excel-check-manager.vue
│   ├── excel-check-panel.vue
│   ├── excel-check-log.vue
│   └── option-modal.vue
└── composables/
    ├── menu.ts
    ├── use-tree.ts
    ├── use-tree-search.ts
    ├── use-tree-drop-down.ts
    ├── use-tree-and-history.ts
    ├── use-excel-check-log.ts
    ├── use-excel-check-data.ts
    ├── option.ts
    ├── func.ts
    └── params-components/
        └── [24 个参数配置文件]
```

对应的文档结构：
```
frontend/docs/layout/pages/excel-test/
├── index.md
├── components.md
└── composables/
    └── params-components/
        └── (参数组件文档待补充)
```

## Related Files

- **Main Page**: [index.md](./index.md)
- **Shared Components**: [../../../../shared/components.md](../../../../shared/components.md)
