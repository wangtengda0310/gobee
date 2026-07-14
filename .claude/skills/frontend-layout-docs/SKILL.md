---
name: frontend-layout-docs
description: |
  前端页面布局文档生成与维护技能。生成包含 ASCII 布局图、TypeScript 接口、组件层次分析的完整布局文档。
  触发条件（满足任一即触发）：
  (1) 开发新的前端页面或在已有页面上新增组件/面板/功能区域时
  (2) 用户提到"布局文档"、"layout"、"页面结构"、"组件层次"
  (3) 已有页面的组件结构发生显著变化（新增/删除/重组大面积UI区域）
  (4) 用户要求分析或review前端组件结构
  (5) 使用了 frontend/src/pages/ 下的 .vue 文件进行开发后
  重要：当开发工作涉及 frontend/src/pages/ 下任何页面的结构变更时，
  应在功能开发完成后主动使用此技能更新对应的布局文档。


# Frontend Layout Docs Skill

Generate comprehensive layout documentation to help AI quickly understand and modify frontend code.

## Use Cases

- New projects needing layout documentation for AI understanding
- Existing projects needing supplementary layout explanations
- Team collaboration requiring unified frontend structure documentation
- AI needing to quickly locate components and modify code
- Code review: identifying potential issues in component structure

## Document Structure

Generated layout docs are stored in `frontend/docs/layout/` directory, mirroring `src/` structure:

```
frontend/docs/
├── layout/
│   ├── CLAUDE.md              # Index file
│   ├── layouts/               # Layout components
│   │   └── normal-layout.md
│   ├── pages/                 # Per-page docs (mirrors src/pages/)
│   │   └── {page-name}/
│   │       ├── index.md       # Main layout ASCII + component tree
│   │       ├── components.md  # Component hierarchy
│   │       ├── composables.md # State management & logic
│   │       └── components/    # Nested component docs
│   └── shared/                # Shared components & data structures
│       ├── components.md
│       └── data-structures.md
└── screenshots/               # Optional: screenshots and annotations
```

> ⚠️ **禁止**在 `docs/layout/` 生成布局文档（旧路径已废弃），统一使用 `frontend/docs/layout/`。

> 完整的单页布局模板和索引文件模板详见 [references/doc-template.md](references/doc-template.md)

## Execution Steps

### 1. Analyze Project Structure

```bash
# View frontend directory structure
ls frontend/src/

# View page files
ls frontend/src/pages/

# View layout components
ls frontend/src/layout/ 2>/dev/null || ls frontend/src/layouts/ 2>/dev/null

# View router config
cat frontend/src/router/index.ts
```

### 2. Read Key Files

For each page, read:
- Page component file (`.vue` / `.tsx`)
- Related composables (`composables/`, `hooks/`, `stores/`)
- Sub-component files (`components/`)
- Type definitions (`types/`, `@bindings/`)

### 3. Extract TypeScript Interfaces

**IMPORTANT**: Always extract and document TypeScript interfaces:
- Props interfaces from component definition
- State interfaces from composables
- API response interfaces from service calls
- Type imports from `@bindings/` or local type files

> TypeScript 接口命名规范和必填章节详见 [references/doc-template.md](references/doc-template.md) 中的 TypeScript Interface Standards 章节

### 4. Generate Layout Docs

Create documents in this order:
1. **Layout Component** - Unified layout (e.g., Normal.vue)
2. **Main Pages** - Sorted by importance
3. **Index File** - Create CLAUDE.md last

> 生成文档时使用 [references/doc-template.md](references/doc-template.md) 中的 Single Page Layout Template 和 Index File Template

### 5. Code Review Analysis

During documentation, identify and note:
- Unused variables, functions, or computed properties
- Inconsistent patterns (styling, naming, structure)
- Performance concerns (missing virtualization, excessive re-renders)
- Accessibility issues
- Technical debt (commented code, hard-coded values)

### 6. Update Project CLAUDE.md

Add layout documentation reference in project root CLAUDE.md

## Code Review Notes Standards

### Issue Categories

1. **Unused Code**: Variables, functions, computed properties
2. **Inconsistent Patterns**: Styling, naming, structure differences
3. **Performance**: Rendering, memory, network concerns
4. **Accessibility**: ARIA, keyboard nav, contrast
5. **Technical Debt**: TODOs, commented code, hard-coded values

### Format

```markdown
### Potential Issues

1. **Category**: Description
   - Example: `variableName` at line XX - reason
```

> ASCII 布局图标准、组件层次树标准、数据流图标准详见 [references/doc-template.md](references/doc-template.md) 中的对应章节

## Notes

1. **Keep it concise** - Documentation should help understanding, not be verbose
2. **Reference code** - Use file paths and line numbers rather than copying code
3. **Focus on data flow** - How state flows between components is most important
4. **Mark dimensions** - Clearly distinguish fixed and flexible areas
5. **Include TypeScript types** - Always document interfaces for type safety
6. **Note issues** - Document potential problems discovered during analysis
7. **Use CLAUDE.md** - Use CLAUDE.md instead of README.md for AI recognition
8. **Verify dates** - Include verification date and status in each document
