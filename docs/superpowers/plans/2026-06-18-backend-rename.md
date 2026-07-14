# 内层 rain-qa-func/ → backend/ 改名 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将内层 `rain-qa-func/` 目录改名为 `backend/`（与 `frontend/` 对称），更新所有 import 路径/bindings/前端引用/文档，module 名保留不变。

**Architecture:** 分阶段机械替换。阶段 A 用 sed 批量替换 290 个 Go 文件的 import 路径中段（`rain-qa-func/rain-qa-func/` → `rain-qa-func/backend/`）+ goimports 排序；阶段 B 重新生成 bindings + 前端 import；阶段 C 文档；阶段 D build + 全量验证。

**Tech Stack:** Go 1.26 + Wails v3 + sed + goimports

## Global Constraints

- **module 名 `git.devcloud.ztgame.com/v-tangfangda/rain-qa-func` 保留不变**（外层）
- **外层目录 `xcard-qa-tools/rain-qa-func/`（项目根）不改**
- **`cmd/`（含 cmd/tests/streamproxy）、`frontend/` 顶层目录不改**
- 各包 package 名不改（只改路径中段）
- sed 精确匹配双层 `rain-qa-func/rain-qa-func/`（唯一模式，不误伤单层 module 名前缀或外层目录）
- bindings 用 `wails3 generate bindings -ts`（命名重构教训：漏 -ts 生成 .js）
- 执行前确保 `wails3 dev` / `rain-qa-func.exe` 未运行（避免 bindings/exe 锁）

## File Structure

- Move: `rain-qa-func/` → `backend/`（含 `pkg/`）
- Modify: 290 Go 文件 import 路径中段
- Regenerate: `frontend/bindings/.../rain-qa-func/backend/`（替换旧 `rain-qa-func/rain-qa-func/`）
- Modify: 前端 `.ts`/`.vue` 的 `@bindings/.../rain-qa-func/rain-qa-func/` import
- Modify: 各 CLAUDE.md、layout docs、specs 中内层路径引用

---

## 阶段 A：Go 目录改名 + import 批量替换

### Task A1: git mv + sed import + goimports + 验证

**Files:**
- Move: `rain-qa-func/` → `backend/`
- Modify: 所有 `.go` 文件 import 路径

- [ ] **Step 1: 确认 wails3 dev / rain-qa-func.exe 未运行**

Run: `powershell.exe -Command "Get-Process wails3,rain-qa-func -ErrorAction SilentlyContinue"`
Expected: 无输出（有则 `Stop-Process -Force`）

- [ ] **Step 2: git mv 内层目录**

Run: `git mv rain-qa-func backend`
Expected: 无输出（rain-qa-func/ 整体移到 backend/，含 pkg/）

- [ ] **Step 3: sed 批量替换 Go import 路径中段**

Run: `find . -name "*.go" | xargs sed -i 's|rain-qa-func/rain-qa-func/|rain-qa-func/backend/|g'`
Expected: 无输出；所有 `.go` 文件中 `.../rain-qa-func/rain-qa-func/pkg/X` → `.../rain-qa-func/backend/pkg/X`

**说明：** 双层 `rain-qa-func/rain-qa-func/` 是唯一模式（module 名 `.../rain-qa-func` + 内层目录 `rain-qa-func`），sed 只匹配该子串，不误伤单层引用。处理**所有 import 含 blank import**（如 `_ ".../proto-test/msg"`，goimports 漏这类）。

- [ ] **Step 4: goimports 排序/清理**

Run: `find . -name "*.go" | xargs goimports -w 2>/dev/null; echo done`
Expected: goimports 排序 import，清理冗余。若个别文件报错（blank import 边角），忽略（sed 已处理路径）。

- [ ] **Step 5: gofmt**

Run: `gofmt -w $(find . -name "*.go")`
Expected: 无输出

- [ ] **Step 6: go build 全项目**

Run: `go build ./...`
Expected: 无输出（编译通过）。若报个别 import 未修复，Run `gofiximports` 兜底：`find . -name "*.go" | xargs gofiximports -w`，再 build。

- [ ] **Step 7: go test（关键包）**

Run: `go test ./backend/pkg/proto-test/server-config/... ./backend/pkg/proto-test/...`
Expected: PASS（server-config 8/8、proto-test 各测试通过）

- [ ] **Step 8: grep 验证无残留双层**

Run: `git grep -l "rain-qa-func/rain-qa-func/" -- '*.go' | head`
Expected: 无输出（所有 Go 文件双层路径已替换）

- [ ] **Step 9: 提交阶段 A**

```bash
git add -A
git commit -m "refactor(backend): 内层 rain-qa-func→backend 目录改名（Go import 路径中段）

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 阶段 B：bindings 重生成 + 前端 import

### Task B1: bindings + 前端 import + tsc

**Files:**
- Delete: `frontend/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/`（旧双层路径）
- Regenerate: `frontend/bindings/.../rain-qa-func/backend/`
- Modify: 前端 `.ts`/`.vue` 的 `@bindings` import

- [ ] **Step 1: 删除旧 bindings 目录**

Run: `git rm -r "frontend/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func"`
Expected: 删除旧双层路径 bindings

- [ ] **Step 2: 重新生成 bindings（-ts）**

Run: `wails3 generate bindings -ts`
Expected: 在 `frontend/bindings/.../rain-qa-func/backend/` 生成新路径 bindings（.ts）

- [ ] **Step 3: 确认新 bindings 路径**

Run: `ls "frontend/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/" | head`
Expected: 显示 `pkg/` 等子目录（新路径生成）

- [ ] **Step 4: 前端 import 路径替换**

Run: `find frontend/src -name "*.ts" -o -name "*.vue" | xargs sed -i 's|rain-qa-func/rain-qa-func/|rain-qa-func/backend/|g'`
Expected: 前端 `@bindings/.../rain-qa-func/rain-qa-func/` → `.../rain-qa-func/backend/`

- [ ] **Step 5: 前端 tsc 验证**

Run: `cd frontend && npx tsc --noEmit 2>&1 | grep -iE "rain-qa-func|backend" | head`
Expected: 无输出（无 backend 路径相关错误；注：全项目约 20 个预存在 tsc 错误与本次无关）

- [ ] **Step 6: grep 验证前端无残留双层**

Run: `git grep -l "rain-qa-func/rain-qa-func/" -- '*.ts' '*.vue' | head`
Expected: 无输出

- [ ] **Step 7: 提交阶段 B**

```bash
git add frontend/bindings frontend/src
git commit -m "refactor(backend): bindings 重生成 + 前端 import 到 rain-qa-func/backend

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 阶段 C：文档同步

### Task C1: 文档内层路径引用

**Files:** 各 CLAUDE.md、layout docs、specs、TODO.md 中内层路径引用

- [ ] **Step 1: 定位文档中的内层路径引用**

Run: `git grep -l "rain-qa-func/rain-qa-func/\|rain-qa-func/pkg/" -- '*.md' | head -20`
Expected: 列出含内层路径引用的文档

- [ ] **Step 2: 替换文档内层路径**

Run: `git grep -l "rain-qa-func/rain-qa-func/" -- '*.md' | xargs sed -i 's|rain-qa-func/rain-qa-func/|rain-qa-func/backend/|g'`
（对 `rain-qa-func/pkg/` 描述性引用，手动判断改为 `backend/pkg/` 或保留 —— 注意区分外层 `xcard-qa-tools/rain-qa-func/` 项目名引用不改）

- [ ] **Step 3: 更新各级 CLAUDE.md 的目录结构树**

各级 CLAUDE.md（根、backend/pkg/、frontend/ 等）的目录结构图里 `rain-qa-func/pkg/` → `backend/pkg/`。逐个检查，谨慎区分内层路径 vs 外层项目名。

- [ ] **Step 4: grep 验证文档无残留双层**

Run: `git grep -n "rain-qa-func/rain-qa-func/" -- '*.md' | head`
Expected: 无输出（外层 `xcard-qa-tools/rain-qa-func/` 项目名引用保留，不算残留）

- [ ] **Step 5: 提交阶段 C**

```bash
git add -A
git commit -m "docs(backend): 同步文档内层路径引用 rain-qa-func→backend

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 阶段 D：build + 全量验证

### Task D1: build .exe + 全量验证

- [ ] **Step 1: wails3 task build**

Run: `wails3 task build 2>&1 | tail -5`
Expected: 无错误，生成 `rain-qa-func.exe`

- [ ] **Step 2: 全量验证**

Run: `go build ./... && go test ./backend/pkg/... 2>&1 | tail -5`
Expected: go build 无输出，go test PASS

Run: `cd frontend && npx tsc --noEmit 2>&1 | grep -c "error TS"`
Expected: 约 20（预存在，无新增 backend 相关错误）

- [ ] **Step 3: 残留双层全局扫描**

Run: `git grep -n "rain-qa-func/rain-qa-func/" | head`
Expected: 无输出（全项目无残留双层路径）

- [ ] **Step 4: amend .exe + 提交**

```bash
git add rain-qa-func.exe
git commit -m "chore(backend): 重建 rain-qa-func.exe（backend 改名后）

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

- [ ] **Step 5: 冒烟（可选，人工或 E2E）**

启动 `wails3 dev`，确认应用运行、Proto测试页可达、注入按钮工作（与命名重构 B4 同）。

## Self-Review

**1. Spec coverage:** 映射表（目录/import/bindings/前端/module 保留）→ A1/B1/C1 全覆盖；边界（module/cmd/frontend/package 名）→ Global Constraints 列；验证 → D1。

**2. Placeholder scan:** 每步含具体 sed/build/grep 命令，无 TBD。

**3. 一致性:** `rain-qa-func/rain-qa-func/` → `rain-qa-func/backend/` 全程一致；module 名 `rain-qa-func` 保留一致。
