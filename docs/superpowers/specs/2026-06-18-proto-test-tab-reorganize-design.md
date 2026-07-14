# Proto-test 页签目录重组设计

> **日期**：2026-06-18
> **状态**：待用户审查
> **范围**：`frontend/src/pages/proto-test/` 扁平 `components/` 重组为按页签内聚

## 1. 背景

`pages/proto-test/` 下扁平 `components/` 含 25 个组件，3 个页签（发包改包 / 测试用例 / 重放结果）的专属组件与跨页签共享组件混在一起，导航与归属判断困难。

当前结构：
```
pages/proto-test/
├── index.vue                    # 页面入口（页签编排）
├── components/                  # 扁平，25 组件混放
├── composables/
└── protocol-content.requirement.ts
```

## 2. 目标

按页签内聚重组为 4 个目录：`stream-proxy/`（发包改包）、`cases/`（测试用例）、`replay-result/`（重放结果）、`shared/`（跨页签共享）。**UI 与组件逻辑不变**，纯目录移动 + import 相对路径更新。

## 3. 归类原则

**严格按依赖关系**：被多个页签 import 的组件归 `shared/`，页签专属组件归各自目录。依赖关系基于实际 import（非猜测）。

## 4. 归属表

### stream-proxy/（发包改包页签专属）
| 组件 | 依据 |
|------|------|
| `packet-tab.vue` | 页签容器，index.vue import |
| `protocol-content.requirement.ts` | packet-tab 专属（RecordControlService 封装） |

### cases/（测试用例页签专属）
| 组件 | 依据 |
|------|------|
| `testcase-tab.vue` | 页签容器，index.vue import |

### replay-result/（重放结果页签专属）
| 组件 | 依据 |
|------|------|
| `replay-result-tab.vue` | 页签容器 |
| `replay-result-selector.vue` | replay-result-tab 专属（无独立 requirement.ts） |

### shared/（跨页签共享，~15）
| 组件 | 依据（被哪些页签 import） |
|------|------|
| `message-table.vue` + req | packet + testcase + replay-result（3 页签，variant 区分） |
| `paired-payload-editor.vue` + req | packet + testcase + replay-result |
| `replay-control.vue` + req | packet + testcase + replay-result |
| `target-service-config.vue` + req | index.vue（页面级顶部配置，3 页签共享） |
| `case-selector.vue` + req | testcase + packet（保存用例）+ index |
| `req-card-editor.vue` + req | paired-payload-editor 子组件 |
| `payload-editor.vue` | 无引用（疑似 paired-payload-editor 前身，孤儿保留到 shared/） |
| `field-item.vue` | req-card-editor 子组件 |
| `combo-select.vue` / `enum-select.vue` / `range-input.vue` / `variable-select.vue`(+req) | 字段类型子组件，被 field-item/编辑器用 |
| `composables/use-paired-messages.ts` | 配对算法，被编辑器/tab 用 |
| `composables/use-selected-entry.ts` | 被 tab/编辑器用 |

## 5. shared/ 内部组织

**扁平 + composables/ 独立**（与现有 pages/components/ + composables/ 风格一致）：
- `shared/` 根：所有共享 .vue + 同名 .requirement.ts
- `shared/composables/`：use-*.ts

## 6. 目标结构

```
pages/proto-test/
├── index.vue
├── CLAUDE.md
├── stream-proxy/
│   ├── packet-tab.vue
│   └── protocol-content.requirement.ts
├── cases/
│   └── testcase-tab.vue
├── replay-result/
│   ├── replay-result-tab.vue
│   └── replay-result-selector.vue
└── shared/
    ├── message-table.vue + .requirement.ts
    ├── paired-payload-editor.vue + .requirement.ts
    ├── replay-control.vue + .requirement.ts
    ├── target-service-config.vue + .requirement.ts
    ├── case-selector.vue + .requirement.ts
    ├── req-card-editor.vue + .requirement.ts
    ├── payload-editor.vue
    ├── field-item.vue
    ├── combo-select.vue
    ├── enum-select.vue
    ├── range-input.vue
    ├── variable-select.vue + .requirement.ts
    └── composables/
        ├── use-paired-messages.ts
        └── use-selected-entry.ts
```

## 7. 连锁更新

### 7.1 文件移动（git mv）
- 25 个文件（.vue + .requirement.ts + composables）git mv 到 4 目录
- requirement.ts 跟随同名 .vue 同目录
- 删除空的旧 `components/` + 旧 `composables/`（内容移到 shared/composables/）

### 7.2 import 相对路径更新
| 场景 | 旧 | 新 |
|------|-----|-----|
| tab → shared 组件 | `./message-table.vue` | `../shared/message-table.vue` |
| tab → shared requirement | `./replay-control.requirement` | `../shared/replay-control.requirement` |
| index.vue → tab | `./components/packet-tab.vue` | `./stream-proxy/packet-tab.vue` |
| index.vue → shared | `./components/case-selector.requirement` | `./shared/case-selector.requirement` |
| index.vue → target-service-config | `./components/target-service-config.vue` | `./shared/target-service-config.vue` |
| shared 内部 | `./xxx` | `./xxx`（不变） |
| shared → composables | `../composables/use-xxx` | `./composables/use-xxx` |
| tab → composables | `../composables/use-xxx` | `../shared/composables/use-xxx` |

### 7.3 文档同步
- `pages/proto-test/CLAUDE.md`：目录结构树（第 7 节）+ requirement.ts 索引表（第 2 节）+ 组件列表
- `frontend/docs/layout/pages/proto-test/index.md`：组件树/路径引用（如有）
- 旧 `components/CLAUDE.md`：组件职责已在 `proto-test/CLAUDE.md` §2 索引表覆盖，重组后更新该表即可；旧 `components/CLAUDE.md` 删除（不新建 `shared/CLAUDE.md`，YAGNI）
- `backend/pkg/proto-test/CLAUDE.md`：前端组件路径引用（如有）

### 7.4 E2E
- `frontend/e2e/shared/pages/ProtoTestPage.ts` 用 CSS/text locator（不 import 组件路径）→ 子目录移动**不影响 locator**
- 各 spec 不 import 组件路径 → 不受影响
- **需冒烟验证**：proto-test-base.spec 核心（页签/录制按钮）确认 locator 仍工作

## 8. 验证

1. **tsc --noEmit**：import 路径全部正确，预存在 ~20 错误不新增、无 `proto-test/components` 旧路径错误
2. **grep 无旧路径残留**：`git grep "pages/proto-test/components/" -- '*.ts' '*.vue'` 无输出
3. **E2E 冒烟**：proto-test-base.spec 核心用例通过（页签切换、录制按钮）

## 9. 边界（YAGNI）

**不做**：
- 不改组件内部逻辑（纯目录移动 + import 路径）
- 不改 index.vue 页签编排逻辑（只改 import 路径）
- 不重组 E2E Page Object（locator 不变）
- 不改路由/菜单（/ProtoTest 不变）
- 不改后端

## 10. 执行策略（按层分 task）

- **Task 1**：git mv 25 文件到 4 目录 + 更新所有 import 相对路径 + tsc 验证 + grep 验证
- **Task 2**：文档同步（CLAUDE.md 目录树 + 索引表 + 布局文档 + 删旧 components/CLAUDE.md）
- **Task 3**：E2E 冒烟验证（proto-test-base.spec 核心）

每个 task 独立提交，subagent-driven 执行。

## 11. 风险与缓解

- **import 路径错误**：tsc + grep 双验证
- **requirement.ts 漏移**：grep 同名 .vue 确认每个 .vue 的 requirement.ts 同目录
- **文档遗漏**：Task 2 专门处理，grep 旧路径 `pages/proto-test/components/` 兜底
- **E2E locator 意外失效**：Task 3 冒烟验证（理论上不影响，因 locator 是 CSS/text）
