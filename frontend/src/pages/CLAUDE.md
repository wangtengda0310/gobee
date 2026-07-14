# pages/ 目录规范

本目录包含所有页面级功能模块。每个子目录是一个完整的功能单元。

## 模块文档组织规范

每个页面模块的 `CLAUDE.md` 必须包含以下章节，顺序固定，**页面布局有修改时需要同步更新相关章节、相关文档**：

```markdown
# <页面名称>

## 1. 前端布局文档
文档链接：`docs/layout/pages/<page-name>/index.md`

## 2. e2e测试用例

- 测试文件：`frontend/e2e/<page-name>.spec.ts`
- Page Object：`frontend/e2e/shared/pages/<page-name>Page.ts`
- 运行方式：`wails3 dev` 启动后 `npx playwright test <page-name>.spec.ts`
- 详细说明：`frontend/e2e/<page-name>-replay-tests.md` — xxx文档


## 3. requirement.ts 规范

每个 .vue 组件如有 Wails 调用，必须同目录声明同名 `.requirement.ts` 文件，禁止直接 import Wails bindings。

## 4. wails.go 索引

| Service | 后端文件 | 对应前端组件 | requirement.ts |
|---------|---------|------------|---------------|

## 5. 设计决策 / 时序图 / 已知问题

每个页面组件的 vue 文件中必须包含 ASCII 图表示当前组件在页面中的位置关系和调用关系，标明组件名称和调用的 requirement.ts 文件。详见 `proto-test/CLAUDE.md` 中的示例。

## 组件开发规范

### requirement.ts 强制规则

1. **存在性**：任何 `.vue` 文件如果调用 Wails 方法，必须同目录下有同名 `.requirement.ts`
2. **禁止直接 import bindings**：组件只能 import 自己的 `.requirement.ts`
3. **接口隔离**：每个 requirement.ts 只声明该组件需要的方法，不暴露多余接口

### requirement.ts 文件模板

```typescript
// <组件名>.requirement.ts

// ========== DTO ==========
export interface XxxDTO { ... }

// ========== Service 接口 ==========
export interface XxxService {
  method(): Promise<XxxDTO>
}

// ========== Wails 实现 ==========
export function createWailsXxxService(): XxxService { ... }

// ========== Mock 实现 ==========
export function createMockXxxService(): XxxService { ... }
```

### 组件中使用方式

```typescript
// <组件名>.vue
import { createWailsXxxService, type XxxDTO } from './<组件名>.requirement'

const service = createWailsXxxService() // 或 createMockXxxService()
```

## 目录结构示例

```
pages/<page-name>/
├── index.vue                   # 页面入口
├── index.requirement.ts        # 页面级依赖（如有）
├── components/
│   ├── xxx.vue                 # UI 组件
│   ├── xxx.requirement.ts      # 组件依赖的 Service 接口
│   └── CLAUDE.md               # 组件目录手册
├── CLAUDE.md                   # 模块级文档（按本规范组织）
```

## 现有模块索引

| 模块 | 路径 | 状态 |
|------|------|------|
| proto-test | proto-test/ | requirement.ts 迁移中 |
| settings | settings/ | 待迁移 |
| function-test | function-test/ | 待迁移 |
| excel-test | excel-test/ | 待迁移 |
| hero-wiki-check | hero-wiki-check/ | 待迁移 |
| activity-wiki-check | activity-wiki-check/ | 待迁移 |
| hero-voice-resource-check | hero-voice-resource-check/ | 待迁移 |
| llm | llm/ | 待迁移 |
