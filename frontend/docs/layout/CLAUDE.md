# Rain QA Frontend Layout Documentation

前端布局文档索引，各页面的 ASCII 布局可视化。

## Document Structure

```
frontend/docs/layout/
├── CLAUDE.md                  # 本文件（索引）
├── layouts/                   # 布局组件
│   └── normal-layout.md
├── pages/                     # 页面文档
│   ├── activity-wiki-check/
│   ├── function-test/
│   ├── excel-test/
│   ├── hero-wiki-check/
│   ├── home/
│   ├── settings/
│   └── hero-voice-resource-check/
└── shared/                    # 共享组件
    └── components.md
```

## Pages

| Page | Route | Document |
|------|-------|----------|
| Function Test | `/function-test` | [pages/function-test/index.md](./pages/function-test/index.md) |
| Excel Test | `/excel-test` | [pages/excel-test/index.md](./pages/excel-test/index.md) |
| Hero Wiki Check | `/hero-wiki-check` | [pages/hero-wiki-check/index.md](./pages/hero-wiki-check/index.md) |
| Home | `/` | [pages/home/index.md](./pages/home/index.md) |
| Settings | `/settings` | [pages/settings/index.md](./pages/settings/index.md) |
| Hero Voice Resource Check | `/hero-voice-resource-check` | [pages/hero-voice-resource-check/index.md](./pages/hero-voice-resource-check/index.md) |
| Activity Wiki Check | `/ActivityWiki` | [pages/activity-wiki-check/index.md](./pages/activity-wiki-check/index.md) |
| Proto Test | `/ProtoTest` | [pages/proto-test/index.md](./pages/proto-test/index.md) |

## Layout Components

| Component | Document |
|-----------|----------|
| Normal Layout | [layouts/normal-layout.md](./layouts/normal-layout.md) |

## Tech Stack

- Vue 3 (Composition API, `<script setup>`) + TypeScript + Naive UI + Wails

## Documentation Convention

每个页面的文档结构：
1. **index.md** — 主布局 ASCII 图 + 组件树 + 数据流
2. **components.md** — 组件层级和依赖关系
3. **composables.md** — 状态管理和逻辑层
4. **components/[nested].md** — 嵌套组件（如 modals.md）

## Source vs Doc Mapping

| Source Path | Doc Path |
|-------------|----------|
| `src/layouts/normal-layout/index.vue` | `layouts/normal-layout.md` |
| `src/pages/{page}/index.vue` | `pages/{page}/index.md` |
| `src/pages/{page}/components/*` | `pages/{page}/components.md` |
| `src/pages/{page}/composables/*` | `pages/{page}/composables.md` |
| `src/shared/components/*` | `shared/components.md` |
| 规则参数组件 (chain-reference) | `pages/excel-test/chain-reference.md` |
| 设计文档 (组件级) | `pages/{page}/components/{component}.md` |

## Additional Documents

以下文档从 `docs/layout/` 合并而来，记录组件设计或规则参数的详细信息：

| 文档 | 来源 | 说明 |
|------|------|------|
| [pages/excel-test/chain-reference.md](./pages/excel-test/chain-reference.md) | 原独立节 | CHAIN_REFERENCE 关系链规则参数完整说明 |
| [pages/excel-test/components/excel-check-log.md](./pages/excel-test/components/excel-check-log.md) | 原 ExcelCheckLog-layout.md | 检查日志面板优化设计（已实现） |
| [pages/settings/components/roadmap.md](./pages/settings/components/roadmap.md) | 原 Roadmap-layout.md | 开发路线图布局（已合并到设置页） |
| [shared/data-structures.md](./shared/data-structures.md) | 原索引内联 | TableCheckResult / ColCheckResult 共享数据结构 |

## Maintenance

1. 页面布局变更时更新对应文档
2. 文档目录结构与 `src/` 保持镜像
3. 新增/删除组件时更新组件树
4. **E2E 测试同步**：前端变更须同步更新 E2E 测试，详见 [E2E 测试同步规范](../../e2e-sync-rules.md)

---

**Generated**: 2026-03-25 | **Skill**: frontend-layout-docs v2.0
