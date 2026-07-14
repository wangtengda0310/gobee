# streamproxy / stream-proxy 命名重构设计

> **日期**：2026-06-18
> **状态**：待用户审查
> **范围**：后端包归入 + 前端页面改名（不含页签目录重组）

## 1. 背景

项目中存在三个名字相近但功能不同的实体，导致 AI 与开发者混淆：

| 名字 | 实际功能 | 名字合理性 |
|------|---------|-----------|
| `cmd/tests/streamproxy` | 真·流量代理 CLI（TCP/HTTP 代理、录制、重放、server_addr 劫持） | ✅ 合理 |
| `pkg/streamproxy` | Unity 服务器注入 + 客户端配置导出 | ❌ 误导（与流量代理无关，只因服务前端页面得名） |
| 前端 `pages/stream-proxy/` | Proto 测试页面（协议录制/重放/拦截，菜单"Proto测试"） | ⚠️ 名不副实 |

引用规模：`pkg/streamproxy` 约 10 文件（本次新增）；前端 `stream-proxy` 约 40+ 文件（页面 + 组件 + E2E + 文档）。

## 2. 目标

- 消除 `pkg/streamproxy` 的误导命名（Unity 注入叫 streamproxy）
- 前端页面名与功能/菜单对齐（stream-proxy → proto-test）
- 保留 `cmd/tests/streamproxy`（名字正确）
- 不引入不必要的兼容/重定向（YAGNI）

## 3. 重命名映射

| 当前 | 目标 |
|------|------|
| `rain-qa-func/pkg/streamproxy/` | `rain-qa-func/pkg/proto-test/server-config/`（package `serverconfig`） |
| `StreamProxyService` / `NewStreamProxyService` | `ServerConfigService` / `NewServerConfigService` |
| import `.../pkg/streamproxy` | `.../pkg/proto-test/server-config` |
| 前端 `pages/stream-proxy/` | `pages/proto-test/` |
| 路由 `/StreamProxy` | `/ProtoTest` |
| `e2e/stream-proxy/` | `e2e/proto-test/` |
| `StreamProxyPage.ts` | `ProtoTestPage.ts`（类名 `StreamProxyPage` → `ProtoTestPage`） |
| `cmd/tests/streamproxy` | **保留不动** |

## 4. 阶段 A — 后端子包归入（独立提交）

1. `git mv rain-qa-func/pkg/streamproxy rain-qa-func/pkg/proto-test/server-config`
2. package 声明 `streamproxy` → `serverconfig`（server_config.go、wails.go、server_config_test.go）
3. 类型重命名：`StreamProxyService` → `ServerConfigService`，`NewStreamProxyService` → `NewServerConfigService`
4. `cmd/rain-qa-func/wails.go`：import 改 `serverconfig ".../pkg/proto-test/server-config"`，注册改 `app.RegisterService(application.NewService(serverconfig.NewServerConfigService(excelConfigSvc)))`
5. `wails3 generate bindings` 重新生成（新路径 `bindings/.../pkg/proto-test/server-config/`）
6. 前端 `target-service-config.requirement.ts`：更新 bindings import 路径（`.../pkg/streamproxy/...` → `.../pkg/proto-test/server-config/...`）
7. `gofmt -w` + `go build ./...` + `go test ./backend/pkg/proto-test/server-config/...`
8. 文档同步：
   - `pkg/proto-test/CLAUDE.md`：新增 server-config 子包索引（与 msg 子包并列）
   - `pkg/proto-test/server-config/CLAUDE.md`：原 streamproxy CLAUDE.md 内容，更新内部路径引用
   - `pkg/CLAUDE.md`：移除独立 streamproxy 模块行（已归入 proto-test）
   - 根 `CLAUDE.md`：子模块文档表 streamproxy 行 → 指向 `pkg/proto-test/server-config/`

## 5. 阶段 B — 前端页面改名（独立提交）

1. `git mv frontend/src/pages/stream-proxy frontend/src/pages/proto-test`
2. `git mv frontend/e2e/stream-proxy frontend/e2e/proto-test`
3. `frontend/src/App.vue`：路由 `/StreamProxy` → `/ProtoTest`；`menuItems` 中对应项 path 更新（菜单标签"Proto测试"不变）
4. `frontend/e2e/shared/pages/StreamProxyPage.ts` → `ProtoTestPage.ts`：类名 `StreamProxyPage` → `ProtoTestPage`；更新所有 E2E 文件的 `import` 与 `new ProtoTestPage(...)`
5. 组件内相对引用（`./components/...`、`../composables/...`）随目录整体移动，无需逐个改
6. 文档同步：
   - `frontend/docs/layout/pages/stream-proxy/` → `proto-test/`（含 index.md、data-flow.md）
   - `frontend/src/pages/proto-test/CLAUDE.md`（原 stream-proxy CLAUDE.md，更新标题与内部引用）
   - `frontend/src/pages/proto-test/components/CLAUDE.md`
   - `frontend/e2e/CLAUDE.md`、`frontend/e2e/proto-test/CLAUDE.md`
   - `frontend/src/e2e-index.md`、`frontend/src/pages/CLAUDE.md`
   - `docs/superpowers/specs/` 下引用 stream-proxy 路径的 specs
7. 前端类型检查 `tsc` + `wails3 dev` 冒烟（页面可达、3 页签切换正常）+ 跑 1-2 个关键 E2E

## 6. 策略与边界

- **路由**：直接替换 `/StreamProxy` → `/ProtoTest`，不做旧路由重定向（内部工具，无外部链接，YAGNI）
- **菜单**：标签"Proto测试"保持不变
- **保留不动**：`cmd/tests/streamproxy`（真流量代理 CLI，名字正确）、`export_client.bat`/`export.py`（策划文件）、`pkg/proto-test/` 主包（server-config 仅作其子包）
- **执行顺序**：先 A 后 B；各自独立验证与提交，降低风险

## 7. 验证关卡

| 阶段 | 验证命令/动作 |
|------|--------------|
| A 完成 | `go build ./...`、`go test ./backend/pkg/proto-test/server-config/...`、`wails3 generate bindings`（确认新路径生成）、前端 `tsc`（requirement.ts import 正确） |
| B 完成 | 前端 `tsc`、`wails3 dev` 冒烟（/ProtoTest 可达、页签正常）、关键 E2E（如 `proto-test-base.spec.ts`） |

## 8. 不在范围（YAGNI）

- 页签目录重组（`pages/proto-test/{stream-proxy,cases,replay-result,shared}`）—— 见后续工作
- `cmd/tests/streamproxy` 重命名
- 策划文件改动

## 9. 后续工作（独立会话）

- **内层目录 `rain-qa-func/` → `backend/`（独立重构，不并入本次）**：将 worktree 根下的 `rain-qa-func/`（含 `pkg/`）改名为 `backend/`，与 `frontend/` 对称，消除内外层 rain-qa-func 同名混淆。影响 250+ Go 文件 import 路径中段（`.../rain-qa-func/pkg/` → `.../backend/pkg/`）+ bindings + 前端 import + 文档；module 名 `rain-qa-func` 保留。本次 streamproxy 重构先落地，此重构作为紧接的下一个独立工作。

- **前端页签目录重组**：将扁平 `components/` 重组为按页签内聚的 `pages/proto-test/{stream-proxy(发包改包), cases(测试用例), replay-result(重放结果), shared(跨页签共享如 message-table)}`。需先梳理每个组件的页签归属与共享关系，作为独立结构优化工作。注意：届时 `stream-proxy` 作为发包改包页签子目录名将名实相符（流量代理/拦截改包）。
