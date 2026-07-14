# Proto-test 页签目录重组 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `pages/proto-test/` 扁平 `components/`（25 文件）重组为 `stream-proxy/`/`cases/`/`replay-result/`/`shared/` 4 目录，UI 与组件逻辑不变。

**Architecture:** 机械 `git mv` + import 相对路径更新。严格按依赖归类（被多页签 import → shared/）。3 task：移文件+import+tsc 验证 / 文档同步 / E2E 冒烟。

**Tech Stack:** Vue 3 + TypeScript + Vite + Wails v3 + Playwright E2E

## Global Constraints

- **UI/组件逻辑不变**（纯目录移动 + import 路径），不改 `<template>`/`<script>` 业务逻辑
- **严格按依赖归类**：被多个页签 import → `shared/`，页签专属 → 各自目录
- **requirement.ts 跟随同名 .vue 同目录**
- **禁 sed 处理含中文 .vue/.md**（用 Edit；import 行是 ASCII，Edit 精确匹配 import 语句）
- **不改 index.vue 页签编排逻辑**（只改 import 路径）、不改路由/菜单/后端
- **无外部代码引用 `proto-test/components/`**（已 grep 确认，仅 CLAUDE.md 文档引用），重组只影响 proto-test 内部 import
- Git Bash 执行，**不用 `&&`**（用 `;` 或分行）

## File Structure（移动映射，25 文件）

**stream-proxy/**（发包改包，2）：
- `components/packet-tab.vue`、`components/protocol-content.requirement.ts`

**cases/**（测试用例，1）：
- `components/testcase-tab.vue`

**replay-result/**（重放结果，2）：
- `components/replay-result-tab.vue`、`components/replay-result-selector.vue`（无 requirement.ts）

**shared/**（跨页签共享，15）：
- `components/message-table.vue`(+req)、`paired-payload-editor.vue`(+req)、`replay-control.vue`(+req)、`target-service-config.vue`(+req)、`case-selector.vue`(+req)、`req-card-editor.vue`(+req)、`payload-editor.vue`(孤儿保留)、`field-item.vue`、`combo-select.vue`、`enum-select.vue`、`range-input.vue`、`variable-select.vue`(+req)

**shared/composables/**（2）：
- `composables/use-paired-messages.ts`、`composables/use-selected-entry.ts`

**删除**：`components/CLAUDE.md`（Task 2）+ 空 `components/` + 空 `composables/`

---

## Task 1: git mv + import 更新 + tsc/grep 验证

**Files:**
- Move: 25 文件（见 File Structure）
- Modify: `index.vue`、`packet-tab.vue`、`testcase-tab.vue`、`replay-result-tab.vue`、`shared/` 内组件 import

- [ ] **Step 1: 创建目录 + git mv stream-proxy/cases/replay-result**

```bash
cd "D:/work/xcard-qa-tools/rain-qa-func/.claude/worktrees/streamfilter/frontend/src/pages/proto-test"
mkdir -p stream-proxy cases replay-result shared/composables
git mv components/packet-tab.vue stream-proxy/
git mv components/protocol-content.requirement.ts stream-proxy/
git mv components/testcase-tab.vue cases/
git mv components/replay-result-tab.vue replay-result/
git mv components/replay-result-selector.vue replay-result/
```

- [ ] **Step 2: git mv shared（15 文件）**

```bash
git mv components/message-table.vue components/message-table.requirement.ts shared/
git mv components/paired-payload-editor.vue components/paired-payload-editor.requirement.ts shared/
git mv components/replay-control.vue components/replay-control.requirement.ts shared/
git mv components/target-service-config.vue components/target-service-config.requirement.ts shared/
git mv components/case-selector.vue components/case-selector.requirement.ts shared/
git mv components/req-card-editor.vue components/req-card-editor.requirement.ts shared/
git mv components/variable-select.vue components/variable-select.requirement.ts shared/
git mv components/payload-editor.vue shared/
git mv components/field-item.vue shared/
git mv components/combo-select.vue shared/
git mv components/enum-select.vue shared/
git mv components/range-input.vue shared/
```

- [ ] **Step 3: git mv composables → shared/composables**

```bash
git mv composables/use-paired-messages.ts shared/composables/
git mv composables/use-selected-entry.ts shared/composables/
```

- [ ] **Step 4: 更新 index.vue import（5 行，Edit）**

`index.vue` 在 `proto-test/` 根，Edit 这 5 个 import（ASCII 精确匹配）：
- `'./components/packet-tab.vue'` → `'./stream-proxy/packet-tab.vue'`
- `'./components/testcase-tab.vue'` → `'./cases/testcase-tab.vue'`
- `'./components/replay-result-tab.vue'` → `'./replay-result/replay-result-tab.vue'`
- `'./components/target-service-config.vue'` → `'./shared/target-service-config.vue'`
- `'./components/case-selector.requirement'` → `'./shared/case-selector.requirement'`

- [ ] **Step 5: 更新 3 个 tab 的 import（跨目录 → shared）**

对 `stream-proxy/packet-tab.vue`、`cases/testcase-tab.vue`、`replay-result/replay-result-tab.vue`，用 grep 定位 + Edit：

| 旧（`./` 同目录） | 新（`../shared/`） |
|---|---|
| `'./message-table.vue'` | `'../shared/message-table.vue'` |
| `'./paired-payload-editor.vue'` | `'../shared/paired-payload-editor.vue'` |
| `'./replay-control.vue'` | `'../shared/replay-control.vue'` |
| `'./replay-control.requirement'` | `'../shared/replay-control.requirement'` |
| `'./case-selector.requirement'` | `'../shared/case-selector.requirement'` |
| `'./paired-payload-editor.requirement'` | `'../shared/paired-payload-editor.requirement'` |

| 旧（`../composables/`） | 新 |
|---|---|
| `'../composables/use-selected-entry'` | `'../shared/composables/use-selected-entry'` |
| `'../composables/use-paired-messages'` | `'../shared/composables/use-paired-messages'` |

**不变**（同目录）：`packet-tab.vue` 的 `'./protocol-content.requirement'`（stream-proxy/ 同目录）、`replay-result-tab.vue` 的 `'./replay-result-selector.vue'`（replay-result/ 同目录）。

定位命令：
```bash
cd "D:/work/xcard-qa-tools/rain-qa-func/.claude/worktrees/streamfilter"
grep -rn "from '\./\(message-table\|paired-payload-editor\|replay-control\|case-selector\)" frontend/src/pages/proto-test/stream-proxy/ frontend/src/pages/proto-test/cases/ frontend/src/pages/proto-test/replay-result/
grep -rn "from '\.\./composables/" frontend/src/pages/proto-test/
```

- [ ] **Step 6: 更新 shared 内组件 → composables import**

`shared/message-table.vue`、`shared/paired-payload-editor.vue`、`shared/replay-control.vue`（在 shared/，import composables）：
- `'../composables/use-paired-messages'` → `'./composables/use-paired-messages'`

`shared/composables/use-selected-entry.ts`：`'./use-paired-messages'` **不变**（同 shared/composables/ 目录）。

定位命令：
```bash
grep -rn "from '\.\./composables/" frontend/src/pages/proto-test/shared/
```

- [ ] **Step 7: tsc 验证**

```bash
cd "D:/work/xcard-qa-tools/rain-qa-func/.claude/worktrees/streamfilter/frontend"
npx tsc --noEmit 2>&1 | tee /tmp/tsc-tab.log | tail -5
cd "D:/work/xcard-qa-tools/rain-qa-func/.claude/worktrees/streamfilter"
echo "=== 总错误（应 ≈20 预存在）==="
grep -c "error TS" /tmp/tsc-tab.log
echo "=== components 旧路径/Cannot find module（应无）==="
grep -iE "proto-test/components|Cannot find module" /tmp/tsc-tab.log | head
```
预期：总错误 ≈20（预存在，与重组无关）；无 `proto-test/components` 或 `Cannot find module` 错误。

- [ ] **Step 8: grep 验证无旧路径残留**

```bash
echo "=== components/ 旧引用（应无）==="
git grep -n "from '.*components/" frontend/src/pages/proto-test/ -- '*.vue' '*.ts' | head
echo "=== ../composables/ 旧引用（应无，已改 shared/composables）==="
git grep -n "from '\.\./composables/" frontend/src/pages/proto-test/ | head
```
预期：两条均无输出。

- [ ] **Step 9: commit**

```bash
git add -A
git status --short | head    # 确认 rename + import 修改
git commit -m "refactor(proto-test): 页签组件重组为 stream-proxy/cases/replay-result/shared 目录

- 严格按依赖归类：被多页签 import 的归 shared/，页签专属归各自目录
- 25 文件 git mv + import 相对路径更新，UI/逻辑不变

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: 文档同步

**Files:**
- Modify: `pages/proto-test/CLAUDE.md`（§2 索引表 + §7 目录树）
- Delete: `components/CLAUDE.md`、空 `components/`、空 `composables/`
- Check: `frontend/docs/layout/pages/proto-test/index.md`

- [ ] **Step 1: 更新 proto-test/CLAUDE.md**

- **§2 requirement.ts 组件列表表**：组件路径列从 `components/xxx.vue` 改为新目录（stream-proxy/cases/replay-result/shared）
- **§7 目录结构树**：改为新 4 目录结构（参考规格 §6）
- §4 E2E、§3 设计决策：不变（与目录无关）

- [ ] **Step 2: 删除旧 components/CLAUDE.md + 空目录**

```bash
cd "D:/work/xcard-qa-tools/rain-qa-func/.claude/worktrees/streamfilter"
git rm frontend/src/pages/proto-test/components/CLAUDE.md
# 确认 components/ 和 composables/ 已空（Task 1 移走所有文件），删除空目录
rmdir frontend/src/pages/proto-test/components frontend/src/pages/proto-test/composables 2>/dev/null
```

- [ ] **Step 3: 检查 + 更新布局文档**

```bash
git grep -n "proto-test/components/" -- 'frontend/docs/**/*.md' | head
```
若 `frontend/docs/layout/pages/proto-test/index.md`（或 data-flow.md）有 `components/` 路径引用，Edit 更新为新结构。

- [ ] **Step 4: grep 验证文档无旧路径**

```bash
git grep -n "proto-test/components/" -- '*.md' | grep -v "superpowers/" | head
```
预期：无输出（superpowers/ 下的历史 specs/plans 保留）。

- [ ] **Step 5: commit**

```bash
git add -A
git commit -m "docs(proto-test): 同步页签重组后目录结构

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: E2E 冒烟验证

**Files:** 无代码改动（验证性）；若发现 Task 1 import 遗漏则补改。

- [ ] **Step 1: 确认 wails3 dev 运行（前置）**

```bash
powershell.exe -Command "Get-Process wails3,rain-qa-func -ErrorAction SilentlyContinue"
```
若无运行，提示用户启动 `wails3 dev`（或调度者后台启动），等 CDP 就绪。

- [ ] **Step 2: 跑 proto-test-base.spec**

```bash
cd "D:/work/xcard-qa-tools/rain-qa-func/.claude/worktrees/streamfilter/frontend"
npx playwright test proto-test/proto-test-base.spec.ts 2>&1 | tail -20
```
预期：核心用例通过（页签切换、录制按钮、页面结构）。locator 是 CSS/text，子目录移动**不影响**。预存在 flaky（约 6/12 失败：目标服务 TCP/HTTP/最大并发，与重组无关）。

- [ ] **Step 3: 分析结果**

- 失败若是**预存在 flaky**（TCP/HTTP/最大并发输入框）→ 与重组无关，记录
- 失败若是**组件找不到/import 错误**（新引入）→ Task 1 import 有遗漏，grep 定位 + Edit 补改，回到 Step 2 重跑

- [ ] **Step 4: 报告 / commit fix**

无改动 → 报告冒烟通过（核心用例 + 预存在 flaky 说明）。有 fix → commit `fix(proto-test): 补充遗漏 import`。

---

## Self-Review

**1. Spec coverage:**
- 归类原则（严格按依赖）→ Task 1 Step 1-3（git mv 分组）✓
- shared 扁平 + composables/ → Task 1 Step 2-3 ✓
- import 相对路径更新（index/tab/shared→composables）→ Task 1 Step 4-6 ✓
- 文档同步（CLAUDE.md 目录树+索引表、删旧 components/CLAUDE.md、布局文档）→ Task 2 ✓
- E2E 冒烟（locator 不变验证）→ Task 3 ✓
- YAGNI 边界（不改逻辑/编排/路由/后端/不新建 shared/CLAUDE.md）→ Global Constraints ✓

**2. Placeholder scan:** 每步含确切 git mv 命令 / import 改动表 / grep+ tsc 验证命令，无 TBD/TODO ✓

**3. Type consistency:** 组件名（packet-tab/testcase-tab/replay-result-tab/message-table 等）与路径（stream-proxy/cases/replay-result/shared）全程一致；import 改动表旧→新明确 ✓
