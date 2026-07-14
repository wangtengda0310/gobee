# streamproxy / stream-proxy 命名重构 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `pkg/streamproxy` 重命名为 `pkg/proto-test/server-config` 子包、前端 `stream-proxy` 页面重命名为 `proto-test`，消除命名混淆（保留 `cmd/tests/streamproxy`）。

**Architecture:** 分两阶段独立提交。阶段 A 改后端 Go 包（git mv + package/类型改名 + bindings 重生成 + 前端 import）；阶段 B 改前端页面（目录 mv + 路由 + Page Object + E2E + 文档）。每阶段独立验证。

**Tech Stack:** Go 1.26 + Wails v3 + Vue 3 + TypeScript + Playwright

## Global Constraints

- **module 名 `git.devcloud.ztgame.com/v-tangfangda/rain-qa-func` 保留不变**（只改内层目录与包名）
- **`cmd/tests/streamproxy` 保留不动**（真·流量代理 CLI，名字正确）
- **路由直接替换，不做旧路由重定向**；菜单标签"Proto测试"不变
- package 目录用连字符 `server-config`，package 声明用 `serverconfig`（遵循项目 `proto-test`/`prototest` 惯例）
- **执行每阶段前确保 `wails3 dev` 与 `rain-qa-func.exe` 未运行**（避免 bindings 生成/构建被锁）
- 不含页签目录重组、不含 `rain-qa-func→backend`（均为后续独立工作）
- 改名映射全程一致：`streamproxy`→`serverconfig`、`StreamProxyService`→`ServerConfigService`、`NewStreamProxyService`→`NewServerConfigService`、`StreamProxyPage`→`ProtoTestPage`、`/StreamProxy`→`/ProtoTest`

## File Structure

**阶段 A 涉及文件：**
- 移动：`rain-qa-func/pkg/streamproxy/` → `rain-qa-func/pkg/proto-test/server-config/`（含 server_config.go、wails.go、server_config_test.go、CLAUDE.md）
- 修改：`cmd/rain-qa-func/wails.go`（import + 注册）
- 重新生成：`frontend/bindings/.../pkg/streamproxy/` → `frontend/bindings/.../pkg/proto-test/server-config/`
- 修改：`frontend/src/pages/stream-proxy/components/target-service-config.requirement.ts`（bindings import + 内部接口名）
- 修改：`rain-qa-func/pkg/proto-test/CLAUDE.md`（加子包索引）、`rain-qa-func/pkg/proto-test/server-config/CLAUDE.md`（路径引用）、`rain-qa-func/pkg/CLAUDE.md`（移除 streamproxy 行）、`CLAUDE.md`（子模块表）

**阶段 B 涉及文件：**
- 移动：`frontend/src/pages/stream-proxy/` → `frontend/src/pages/proto-test/`；`frontend/e2e/stream-proxy/` → `frontend/e2e/proto-test/`
- 修改：`frontend/src/App.vue`（路由 + 菜单 path）
- 移动+改类名：`frontend/e2e/shared/pages/StreamProxyPage.ts` → `ProtoTestPage.ts`；`frontend/e2e/shared/fixtures/index.ts`（fixture 名）
- 修改：约 20 个 E2E spec 文件（import + new + 类型注解 StreamProxyPage→ProtoTestPage）
- 修改：`frontend/e2e/shared/CLAUDE.md`、`frontend/e2e/CLAUDE.md`、`frontend/e2e/shared/docs/goto-analysis.md`、`frontend/src/e2e-index.md`、`frontend/src/pages/CLAUDE.md`、`frontend/docs/layout/pages/stream-proxy/`→`proto-test/`

---

## 阶段 A：后端 pkg/streamproxy → pkg/proto-test/server-config

### Task A1: 移动目录并改 package/类型名

**Files:**
- Move: `rain-qa-func/pkg/streamproxy/` → `rain-qa-func/pkg/proto-test/server-config/`
- Modify: `rain-qa-func/pkg/proto-test/server-config/server_config.go`、`wails.go`、`server_config_test.go`

**Interfaces:**
- Produces: package `serverconfig`；类型 `ServerConfigService`；构造 `NewServerConfigService(excelConfigSvc *settings.ExcelConfigService) *ServerConfigService`

- [ ] **Step 1: 确认 wails3 dev / rain-qa-func.exe 未运行**

Run: `powershell.exe -Command "Get-Process wails3,rain-qa-func -ErrorAction SilentlyContinue"`
Expected: 无输出（若有，先 `Stop-Process -Name wails3,rain-qa-func -Force`）

- [ ] **Step 2: git mv 目录**

Run: `git mv rain-qa-func/pkg/streamproxy rain-qa-func/pkg/proto-test/server-config`
Expected: 无输出，目录移动成功

- [ ] **Step 3: 改三个 .go 文件的 package 声明**

对 `server_config.go`、`wails.go`、`server_config_test.go` 三个文件，把首行：
```go
package streamproxy
```
改为：
```go
package serverconfig
```

- [ ] **Step 4: 改类型名 StreamProxyService → ServerConfigService**

在 `rain-qa-func/pkg/proto-test/server-config/wails.go` 中，把所有 `StreamProxyService` 替换为 `ServerConfigService`，`NewStreamProxyService` 替换为 `NewServerConfigService`。涉及：
- `type StreamProxyService struct` → `type ServerConfigService struct`
- `func NewStreamProxyService(...)` → `func NewServerConfigService(...)`
- `func (s *StreamProxyService)` → `func (s *ServerConfigService)`（两个方法 InjectUnityServer、ExportClientConfig）

- [ ] **Step 5: gofmt 格式化**

Run: `gofmt -w rain-qa-func/pkg/proto-test/server-config/`
Expected: 无输出

- [ ] **Step 6: 验证子包自身编译**

Run: `go build ./rain-qa-func/pkg/proto-test/server-config/...`
Expected: 无输出（子包自身编译通过；package 声明已统一为 serverconfig）

**说明：** 本命令只编译 server-config 子包，不含 `cmd/rain-qa-func/wails.go`（它仍引用旧 streamproxy，全项目编译在 Task A2 修复 import 后通过）。

---

### Task A2: 更新 wails.go import 与注册

**Files:**
- Modify: `cmd/rain-qa-func/wails.go`

**Interfaces:**
- Consumes: Task A1 的 `serverconfig.NewServerConfigService`

- [ ] **Step 1: 改 import 路径**

在 `cmd/rain-qa-func/wails.go` 第 29 行附近，把：
```go
	streamproxy "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/streamproxy"
```
改为：
```go
	serverconfig "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/proto-test/server-config"
```

- [ ] **Step 2: 改注册调用**

在第 186 行附近，把：
```go
	app.RegisterService(application.NewService(streamproxy.NewStreamProxyService(excelConfigSvc)))
```
改为：
```go
	app.RegisterService(application.NewService(serverconfig.NewServerConfigService(excelConfigSvc)))
```

- [ ] **Step 3: 全项目编译验证**

Run: `go build ./...`
Expected: 无输出（编译通过）

- [ ] **Step 4: 跑 server-config 包测试**

Run: `go test ./rain-qa-func/pkg/proto-test/server-config/...`
Expected: PASS（8 个测试，含 TestExportClientConfig_RunsBatch 等）

- [ ] **Step 5: 提交阶段 A 代码改动（暂不含 bindings/前端/文档）**

```bash
git add cmd/rain-qa-func/wails.go rain-qa-func/pkg/proto-test/server-config/
git commit -m "refactor(server-config): pkg/streamproxy 移入 pkg/proto-test/server-config 子包"
```

---

### Task A3: 重新生成 bindings 并更新前端 requirement.ts

**Files:**
- Regenerate: `frontend/bindings/.../pkg/proto-test/server-config/`（替换旧 `pkg/streamproxy/`）
- Modify: `frontend/src/pages/stream-proxy/components/target-service-config.requirement.ts`

**Interfaces:**
- Produces: bindings 导出 `ServerConfigService`（替换 `StreamProxyService`）

- [ ] **Step 1: 删除旧 bindings 目录**

Run: `git rm -r "frontend/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/streamproxy"`
Expected: 删除 index.ts、models.ts、streamproxyservice.ts

- [ ] **Step 2: 重新生成 bindings**

Run: `wails3 generate bindings -ts`
Expected: 在 `frontend/bindings/.../pkg/proto-test/server-config/` 下生成 index.ts、models.ts、serverconfigservice.ts

- [ ] **Step 3: 确认新 bindings 导出名**

Run: `grep -r "ServerConfigService" frontend/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/proto-test/server-config/`
Expected: 命中 `export class ServerConfigService`（确认导出名已随 Go 类型名更新）

- [ ] **Step 4: 更新 requirement.ts 的 bindings import**

在 `frontend/src/pages/stream-proxy/components/target-service-config.requirement.ts` 第 4-5 行，把：
```typescript
import { StreamProxyService } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/streamproxy'
import type { ServerXlsxConfig } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/streamproxy'
```
改为：
```typescript
import { ServerConfigService } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/proto-test/server-config'
import type { ServerXlsxConfig } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/rain-qa-func/pkg/proto-test/server-config'
```

- [ ] **Step 5: 更新 requirement.ts 内部引用 StreamProxyService → ServerConfigService**

在该文件中，把 `StreamProxyService.InjectUnityServer` → `ServerConfigService.InjectUnityServer`，`StreamProxyService.ExportClientConfig` → `ServerConfigService.ExportClientConfig`（第 28、31 行）。

- [ ] **Step 6: 更新 requirement.ts 内部接口/函数名（命名一致性）**

把该文件内前端自定义命名也统一（第 19、25、37 行等）：
- `StreamProxyServerConfigService` → `ServerConfigService`
- `createWailsStreamProxyServerConfigService` → `createWailsServerConfigService`
- `createMockStreamProxyServerConfigService` → `createMockServerConfigService`

- [ ] **Step 7: 更新组件 target-service-config.vue 的引用**

Run: `grep -n "StreamProxyServerConfigService\|createWailsStreamProxyServerConfigService\|createMockStreamProxyServerConfigService" frontend/src/pages/stream-proxy/components/target-service-config.vue`
把命中的引用同步改为 Task A3 Step 6 的新名字。

- [ ] **Step 8: 前端类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无类型错误

- [ ] **Step 9: 提交 bindings + 前端 import**

```bash
git add frontend/bindings frontend/src/pages/stream-proxy/components/
git commit -m "refactor(server-config): 同步 bindings 与前端 import 到 proto-test/server-config"
```

---

### Task A4: 阶段 A 文档同步

**Files:**
- Modify: `rain-qa-func/pkg/proto-test/CLAUDE.md`、`rain-qa-func/pkg/proto-test/server-config/CLAUDE.md`、`rain-qa-func/pkg/CLAUDE.md`、`CLAUDE.md`（根）

- [ ] **Step 1: pkg/proto-test/CLAUDE.md 加 server-config 子包索引**

在 `rain-qa-func/pkg/proto-test/CLAUDE.md` 的子包/模块索引表中，参照已有 `msg` 子包条目，新增一行指向 `server-config/CLAUDE.md`，说明"Unity 服务器注入 + 客户端配置导出"。

- [ ] **Step 2: server-config/CLAUDE.md 更新内部路径引用**

在 `rain-qa-func/pkg/proto-test/server-config/CLAUDE.md` 中，把标题/正文里的 "streamproxy" 包名描述更新为 "server-config（pkg/proto-test/server-config 子包）"；类型 `StreamProxyService` → `ServerConfigService`；确认代码引用行号仍有效。

- [ ] **Step 3: pkg/CLAUDE.md 移除独立 streamproxy 行**

在 `rain-qa-func/pkg/CLAUDE.md` 的模块索引表中，删除独立的 `streamproxy` 行（已归入 proto-test/server-config 子包）。

- [ ] **Step 4: 根 CLAUDE.md 子模块表同步**

在根 `CLAUDE.md` 子模块文档表中，把原 `streamproxy` 行的文档路径由 `pkg/streamproxy/CLAUDE.md` 改为 `pkg/proto-test/server-config/CLAUDE.md`，说明改为"Unity 服务器注入 + 客户端配置导出（proto-test 子包）"。

- [ ] **Step 5: 提交文档**

```bash
git add rain-qa-func/pkg/proto-test/CLAUDE.md rain-qa-func/pkg/proto-test/server-config/CLAUDE.md rain-qa-func/pkg/CLAUDE.md CLAUDE.md
git commit -m "docs(server-config): 同步 server-config 子包文档索引"
```

- [ ] **Step 6: 阶段 A 总验证**

Run: `go build ./... && go test ./rain-qa-func/pkg/proto-test/server-config/... && cd frontend && npx tsc --noEmit`
Expected: 全部通过，无错误

---

## 阶段 B：前端 stream-proxy → proto-test

### Task B1: 移动前端目录并改路由/菜单

**Files:**
- Move: `frontend/src/pages/stream-proxy/` → `frontend/src/pages/proto-test/`；`frontend/e2e/stream-proxy/` → `frontend/e2e/proto-test/`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: git mv 页面目录**

Run: `git mv frontend/src/pages/stream-proxy frontend/src/pages/proto-test`
Expected: 无输出

- [ ] **Step 2: git mv E2E 目录**

Run: `git mv frontend/e2e/stream-proxy frontend/e2e/proto-test`
Expected: 无输出

- [ ] **Step 3: 改 App.vue 路由**

在 `frontend/src/App.vue` 第 17 行，把：
```typescript
    {path: '/StreamProxy', component: () => import("@/pages/stream-proxy/index.vue")},
```
改为：
```typescript
    {path: '/ProtoTest', component: () => import("@/pages/proto-test/index.vue")},
```

- [ ] **Step 4: 改 App.vue 菜单 path**

在 `frontend/src/App.vue` 第 75 行，把：
```typescript
  { label: 'Proto测试', path: '/StreamProxy' },
```
改为：
```typescript
  { label: 'Proto测试', path: '/ProtoTest' },
```

- [ ] **Step 5: 前端类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无类型错误（组件内相对引用随目录移动，无需改）

---

### Task B2: Page Object 改名 StreamProxyPage → ProtoTestPage

**Files:**
- Move+Rename: `frontend/e2e/shared/pages/StreamProxyPage.ts` → `ProtoTestPage.ts`
- Modify: `frontend/e2e/shared/fixtures/index.ts`、约 20 个 `frontend/e2e/proto-test/*.spec.ts`、`frontend/e2e/settings/debug-layout.spec.ts`

**Interfaces:**
- Produces: `ProtoTestPage` 类（替换 `StreamProxyPage`）、fixture `protoTestPage`（替换 `streamProxyPage`）

- [ ] **Step 1: 重命名 Page Object 文件**

Run: `git mv frontend/e2e/shared/pages/StreamProxyPage.ts frontend/e2e/shared/pages/ProtoTestPage.ts`

- [ ] **Step 2: 改类名**

在 `frontend/e2e/shared/pages/ProtoTestPage.ts` 第 16 行，把：
```typescript
export class StreamProxyPage extends BasePage {
```
改为：
```typescript
export class ProtoTestPage extends BasePage {
```

- [ ] **Step 3: 批量改 E2E spec 的 import 与类名引用**

对所有 `frontend/e2e/proto-test/*.spec.ts` 和 `frontend/e2e/settings/debug-layout.spec.ts`，执行以下替换（每文件逐个确认）：
- `import { StreamProxyPage } from '../shared/pages/StreamProxyPage'` → `import { ProtoTestPage } from '../shared/pages/ProtoTestPage'`
- 注意 `debug-layout.spec.ts` 和部分文件 import 路径前缀不同（`../shared/pages/` vs `'../shared/pages/StreamProxyPage'`），按各文件实际 import 语句调整
- 类型注解 `StreamProxyPage` → `ProtoTestPage`（如 `let page: StreamProxyPage`）
- `new StreamProxyPage(p)` → `new ProtoTestPage(p)`

Run（验证无残留）: `grep -rn "StreamProxyPage" frontend/e2e/`
Expected: 无输出（全部已替换；文档 CLAUDE.md/goto-analysis.md 中的引用在 Task B3 处理）

- [ ] **Step 4: 改 fixtures/index.ts**

在 `frontend/e2e/shared/fixtures/index.ts`：
- 第 21 行 `import { StreamProxyPage } from '../pages/StreamProxyPage'` → `import { ProtoTestPage } from '../pages/ProtoTestPage'`
- 第 45 行 `streamProxyPage: StreamProxyPage;` → `protoTestPage: ProtoTestPage;`
- 第 135 行 `const streamProxyPage = new StreamProxyPage(page);` → `const protoTestPage = new ProtoTestPage(page);`，并同步把导出/return 的键名 `streamProxyPage` → `protoTestPage`

- [ ] **Step 5: 前端类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无类型错误

- [ ] **Step 6: 提交阶段 B 代码改动**

```bash
git add frontend/src/App.vue frontend/src/pages/proto-test/ frontend/e2e/proto-test/ frontend/e2e/shared/ frontend/e2e/settings/
git commit -m "refactor(proto-test): 前端 stream-proxy 页面与 Page Object 改名为 proto-test"
```

---

### Task B3: 阶段 B 文档同步

**Files:**
- Move: `frontend/docs/layout/pages/stream-proxy/` → `frontend/docs/layout/pages/proto-test/`
- Modify: `frontend/e2e/shared/CLAUDE.md`、`frontend/e2e/CLAUDE.md`、`frontend/e2e/proto-test/CLAUDE.md`、`frontend/e2e/shared/docs/goto-analysis.md`、`frontend/src/e2e-index.md`、`frontend/src/pages/CLAUDE.md`、`frontend/src/pages/proto-test/CLAUDE.md`、`frontend/src/pages/proto-test/components/CLAUDE.md`

- [ ] **Step 1: 移动布局文档目录**

Run: `git mv frontend/docs/layout/pages/stream-proxy frontend/docs/layout/pages/proto-test`

- [ ] **Step 2: 更新各 CLAUDE.md 中的 stream-proxy 引用**

逐文件把 `stream-proxy` → `proto-test`、`StreamProxyPage` → `ProtoTestPage`（涉及上述文件中描述页面路径、Page Object、E2E 目录的语句）。重点：
- `frontend/e2e/shared/CLAUDE.md` 第 11 行 Page Object 表
- `frontend/e2e/CLAUDE.md` 模块对应表与目录结构
- `frontend/src/pages/CLAUDE.md` 模块索引表（stream-proxy → proto-test）
- `frontend/src/e2e-index.md`、`frontend/src/pages/proto-test/CLAUDE.md`（原 stream-proxy CLAUDE.md，标题与内部路径）

Run（验证）: `grep -rn "stream-proxy\|StreamProxyPage" frontend/src/pages/ frontend/e2e/ frontend/docs/`
Expected: 仅剩历史 specs 引用（如 `docs/superpowers/specs/2026-06-09-stream-proxy-intercept-filter-mode.md` 这类历史文档名不改）；活文档无残留

- [ ] **Step 3: 提交文档**

```bash
git add frontend/docs/ frontend/e2e/ frontend/src/
git commit -m "docs(proto-test): 同步前端 proto-test 页面相关文档"
```

---

### Task B4: 冒烟验证

- [ ] **Step 1: 确认无进程占用**

Run: `powershell.exe -Command "Get-Process wails3,rain-qa-func -ErrorAction SilentlyContinue"`
Expected: 无输出

- [ ] **Step 2: 启动应用**

Run: `wails3 dev`（后台运行，等待应用窗口启动）
Expected: 应用启动，无编译错误

- [ ] **Step 3: 人工冒烟（或 CDP 验证）**

验证：
- 菜单"Proto测试"可点击，跳转到 `/ProtoTest` 路由
- 三个页签（发包改包/测试用例/重放结果）切换正常
- "注入 unity 服务器列表"按钮在设置抽屉中存在（阶段 A 的 server-config 功能正常）

- [ ] **Step 4: 跑关键 E2E**

Run（另开终端，应用保持运行）: `cd frontend && npx playwright test proto-test/stream-proxy-base.spec.ts`
Expected: PASS（CDP 连接 9223，页面结构与页签测试通过）

- [ ] **Step 5: 收尾**

停止 `wails3 dev`。最终全量验证：
Run: `go build ./... && cd frontend && npx tsc --noEmit`
Expected: 全部通过

---

## Self-Review

**1. Spec coverage:**
- 映射表 7 项：pkg 目录移动（A1）、package/类型改名（A1）、import 路径（A2）、bindings 重生成（A3）、前端 import（A3）、前端页面/路由（B1）、Page Object（B2）— 全覆盖
- 阶段 A 文档（A4）、阶段 B 文档（B3）、验证关卡（A4 Step6、B4）— 全覆盖
- 边界（cmd/tests 保留、路由不重定向、菜单不变、module 名保留）— Global Constraints 已列

**2. Placeholder scan:** 无 TBD/TODO；每步含具体命令或代码。

**3. Type consistency:** `ServerConfigService`/`NewServerConfigService`/`serverconfig`/`ProtoTestPage`/`protoTestPage`/`/ProtoTest` 全程一致。
