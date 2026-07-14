# Learnings Log

## [LRN-20260511-001] correction

**Logged**: 2026-05-11T10:30:00+08:00
**Priority**: high
**Status**: pending
**Area**: config

### Summary
MCP 工具优先于 CLI 命令；WSL 安装的工具不能在 Git Bash 中执行

### Details
用户要求"mempalace初始化当前项目"时，我直接在 Git Bash 中执行了 `mempalace init` CLI 命令，
而 mempalace 实际安装在 WSL 环境中。导致：
1. Git Bash 子进程使用 GBK 编码，无法解码 git log 中的中文提交消息（UTF-8），产生 `UnicodeDecodeError`
2. 即使 `PYTHONUTF8=1` 修正了编码问题，初始化的 palace 数据库创建在 Windows 路径下，
   MCP server 运行在 WSL 中，无法识别

正确做法：
1. 优先使用 MCP 工具（`mempalace_status`、`mempalace_search` 等）操作 mempalace
2. 如果 MCP 工具不覆盖所需操作（如 init），确认工具运行环境后再执行 CLI
3. mempalace 安装在 WSL 中，CLI 命令应在 WSL 中执行：`wsl mempalace init <path>`

### Suggested Action
1. 记录"mempalace 运行在 WSL 中"到环境信息
2. 执行任何工具命令前先检查其安装位置（`which xxx` / `where xxx`）
3. MCP 工具可完成的操作不使用 CLI

### Metadata
- Source: user_feedback
- Related Files: mempalace.yaml, entities.json（误创建，需清理）
- Tags: mempalace, wsl, encoding, mcp, tool-priority
---

## [LRN-20260519-002] best_practice

**Logged**: 2026-05-19T18:30:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
Go 扩展函数兼容模式：原函数委托新函数，避免逐一修改调用点

### Details
为 `FilterRowsByCondition` 增加 `withinDays` 时间过滤模式时，最初考虑新增 `FilterRowsByConditionEx` 并逐一修改 5 个调用点。审核后发现更好的方式：原函数直接委托给新函数（传入默认 mode=""），这样现有调用点零修改。

```go
func FilterRowsByCondition(cols, filterColName, filterVal, startRowIdx, filterIsArray) []int {
    return FilterRowsByConditionEx(cols, FilterOptions{
        FilterColName: filterColName, FilterVal: filterVal,
        StartRowIdx: startRowIdx, FilterIsArray: filterIsArray,
    })
}
```

新函数使用结构体参数（FilterOptions）避免 8 个位置参数的签名膨胀，同时通过 `Now time.Time` 字段注入时间支持单元测试。

### Suggested Action
Go 项目中扩展已有函数时，优先考虑"原函数委托+结构体参数"模式，而非逐一修改调用点。

### Metadata
- Source: plan_review
- Related Files: rain-excel-checker/xlsx/json_rule/chain_reference/helpers.go
- Tags: golang, backward-compat, refactoring, filter
- See Also: LRN-20260519-003
---

## [LRN-20260519-003] knowledge_gap

**Logged**: 2026-05-19T18:30:00+08:00
**Priority**: medium
**Status**: pending
**Area**: backend

### Summary
Go time.Parse 无时区信息返回 UTC，与 time.Now()（本地时间）比较有偏差

### Details
`helpers.ParseDate()` 使用 `time.Parse(format, dateStr)` 解析日期字符串，DateFormats 不含 MST 时区信息，所以返回 UTC 时间。但 `time.Now()` 返回本地时间（UTC+8），两者直接比较在边界情况下有 8 小时偏差。

现有项目中所有 ParseDate 调用都有同样问题（compare.go、warn_before.go 等），属于隐含假设。

解决方案：在 FilterRowsByConditionEx 中使用 `time.Now().UTC()` 对齐两侧 UTC 语义，并通过 FilterOptions.Now 字段注入时间以支持测试。

### Suggested Action
新代码中涉及时间比较时，统一使用 `time.Now().UTC()` 或通过参数注入时间。

### Metadata
- Source: plan_review
- Related Files: rain-excel-checker/xlsx/helpers/hero_rule_helper.go
- Tags: golang, timezone, time-parsing, utc
---

## [LRN-20260519-004] best_practice

**Logged**: 2026-05-19T18:30:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
嵌套在 JSON 字符串内部的字段不需要新增 ERuleParam 枚举常量

### Details
最初计划为 filterMode/filterDays 新增 `ERuleParam` 常量。审核时发现 ChainStep 的参数是通过 chainSteps JSON 字符串嵌套存储的（前端 serializeSteps → JSON.stringify → params["chainSteps"]，后端 JSON 反序列化），不属于 ColRule.Params 的顶层 key，因此不需要 ERuleParam 枚举。

同理，ChainStep 的 TypeScript 接口在 chain-reference-params.ts 中手动定义（不在 Wails 自动生成的 models.ts 中），不需要运行 bindings 生成。

### Suggested Action
区分"ColRule.Params 顶层 key"和"嵌套 JSON 内部字段"：前者需要 ERuleParam + bindings 生成，后者只需修改对应的结构体/接口。

### Metadata
- Source: plan_review
- Related Files: rain-excel-checker/xlsx/json_rule/rule_def.go, frontend/src/pages/excel-test/composables/rules/chain-reference-params.ts
- Tags: wails, bindings, eruleparam, json-nested, architecture
---

## [LRN-20260519-005] correction

**Logged**: 2026-05-19T18:30:00+08:00
**Priority**: medium
**Status**: pending
**Area**: tests

### Summary
E2E 测试通过 CDP 连接 WebView2 时，page.goto(相对路径) 会失败

### Details
项目 E2E 测试使用 Playwright 通过 CDP 连接 wails3 dev 的 WebView2 实例。现有 ExcelTestPage.goto() 调用 `page.goto("/Excel")`，但 CDP 模式下相对 URL 无效，导致 `Protocol error (Page.navigate): Cannot navigate to invalid URL`。

当前所有 E2E 测试（包括已有的 layout.spec.ts、excel-test.spec.ts 等）都因这个问题失败。这是全局性的环境问题，不是测试文件本身的问题。可能原因：wails3 dev 从非 worktree 目录启动，或 CDP 端口不匹配。

### Suggested Action
排查 CDP 模式下 page.goto 的正确用法。可能需要改为通过菜单点击导航，或使用 `wails3://localhost/Excel` 等完整 URL。

### Metadata
- Source: error
- Related Files: frontend/e2e/fixtures/index.ts, frontend/e2e/shared/pages/BasePage.ts
- Tags: playwright, e2e, cdp, wails, webview2
---

## [LRN-20260520-001] best_practice ✅ resolved

**Status**: resolved — `getNextThursday5AM` 统一使用 UTC 时区
**Summary**: Go time.Parse 返回 UTC 时区，与 time.Date(..., time.Local) 混用会导致跨时区比较偏差
**修复**: `hero_drop_validdate.go:223` 统一使用 `time.Now().UTC()`
**Tags**: timezone, go
- See Also: LRN-20260511-001, LRN-20260520-002
---

## [LRN-20260520-002] knowledge_gap

**Logged**: 2026-05-20T12:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
skill-creator 的 run_eval.py 在 Windows 下因 select.select() 不支持文件描述符而报错 WinError 10038

### Details
使用 skill-creator 的 `run_eval.py` 测试 fix-bug 技能触发条件时，所有 20 个查询都失败，报错：
```
Warning: query failed: [WinError 10038] 在一个非套接字上尝试了一个操作。
```

根因：`run_eval.py:108` 使用 `select.select([process.stdout], [], [], 1.0)` 读取子进程 stdout。
`select.select()` 在 Windows 上**只支持 socket，不支持文件描述符**（包括 subprocess.PIPE）。

这是 skill-creator 脚本本身的跨平台兼容性问题，不是使用方式问题。

影响：
- `run_eval.py` — 无法在 Windows 下运行触发评估
- `run_loop.py` — 同样受影响（依赖 run_eval）
- `quick_validate.py` — 不受影响（不依赖 select）
- `package_skill.py` — 不受影响

### Suggested Action
1. Windows 环境下做 skill 触发评估，需在 WSL/Linux 中运行 `run_eval.py`
2. 或向 skill-creator 项目反馈此兼容性问题，建议改用 `subprocess` 的 `communicate()` 或线程读取 stdout
3. 本地可临时修改 `run_eval.py`，将 `select.select()` 替换为 `process.stdout.readline()` 循环

### Metadata
- Source: error
- Related Files: skill-creator/scripts/run_eval.py
- Tags: skill-creator, windows, select, subprocess, cross-platform, claude-code
- See Also: LRN-20260520-001
---

## [LRN-20260520-003] bug_fix ✅ resolved

**Status**: resolved — `robot-test-log.vue` 已修复
**Summary**: Vue3 reactive 对象重置后，依赖 key 集合的 UI 选中状态未同步失效
**修复**: 验证 `defaultVal` 是否仍存在于 `logCache` 的 keys 中，否则自动切换到第一个可用标签页
**Tags**: vue3, reactive, watch, frontend
- See Also: LRN-20260520-002
---

## [LRN-20260525-001] correction

**Logged**: 2026-05-25T22:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
飞书 bot 通过 email 发送私聊消息无需预先建立单聊关系，但要求用户在 bot 可用范围内

### Details
我错误地告诉用户"飞书要求用户必须先给 bot 发过消息（建立单聊关系），bot 才能给用户发私聊消息"。
这基于对飞书旧版 API 或某些博客文章的错误记忆，与当前文档不符。

用户正确指出：如果需要预先建立单聊关系，那向陌生 git 提交者发消息的功能就无法实现。

实际飞书 OpenAPI 文档明确规定：
1. `receive_id_type=email` 可以直接给任何用户发消息，无需预先互动
2. 前提条件只有两条：应用开启机器人能力 + 用户在 bot 可用范围内
3. 错误码 230013 表示用户不在可用范围内（非"无单聊关系"）

### Suggested Action
1. 遇到 API 行为不确定时，先查官方文档而非依赖记忆
2. 飞书发送消息的可用范围需要在开发者后台配置为"全体员工"

### Metadata
- Source: user_feedback
- Related Files: feishu-lib/feishu_openapi.go, feishu-lib/notification/handlers/feishu_dm.go
- Tags: feishu, openapi, dm, email, bot-availability, correction
---

## [LRN-20260525-002] best_practice

**Logged**: 2026-05-25T22:30:00+08:00
**Priority**: medium
**Status**: pending
**Area**: tests

### Summary
E2E 测试涉及长时间运行的操作时，应提前评估耗时并向用户反馈

### Details
测试 merge 场景的飞书私聊通知时，我直接运行了 `go run rain-excel-checker/main.go`，
没有检查 merge 包含多少子 commit。结果 merge 有大量子 commit（遍历整个分支历史），
程序运行了 5 分钟以上仍未完成，导致用户长时间等待。

用户指出"这种情况你可以向我反馈让我制作一个简单的合并"。

正确做法：
1. 运行前先用 `git log --oneline` 评估 merge 子 commit 数量
2. 如果子 commit 超过 5 个，向用户反馈并建议制作简化测试场景
3. 或预估耗时并设置合理的 timeout

### Suggested Action
涉及 git 遍历的 E2E 测试前，先评估数据规模再决定是否需要简化测试数据

### Metadata
- Source: user_feedback
- Related Files: rain-excel-checker/main.go
- Tags: e2e, testing, merge, performance, user-feedback
---

## [LRN-20260525-003] best_practice

**Logged**: 2026-05-25T23:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: backend

### Summary
Go 程序中成功路径也应有关键日志，避免"静默成功"导致运维排查困难

### Details
飞书私聊通知功能中，`sendDM` 方法只在发送失败时打日志：
```go
if err := h.client.SendText(email, "email", content); err != nil {
    log.Printf("[FeishuDM] 发送失败 (%s): %v", email, err)
}
```

E2E 测试时发送成功但没有日志输出，导致我误以为 handler 没被触发，加了多个 debug 日志排查。
最终发现是"发送成功不输出日志"而非"handler 未注册"。

### Suggested Action
对外部 API 调用（飞书、HTTP 等），成功和失败都应有日志：
```go
log.Printf("[FeishuDM] 发送成功 (%s)", email)
```

### Metadata
- Source: error
- Related Files: feishu-lib/notification/handlers/feishu_dm.go
- Tags: logging, observability, silent-success, go
---

## [LRN-20260527-001] correction

**Logged**: 2026-05-27T10:00:00+08:00
**Priority**: critical
**Status**: pending
**Area**: backend

### Summary
目录重组（feishu-lib → pkg/common/feishu）后，所有引用旧路径的 Go 文件必须同步更新 import，否则编译失败

### Details
`078b3ad` 提交将 `feishu-lib/` 合并到 `backend/pkg/common/feishu/`，更新了 9 个文件的 import 路径。但在 worktree 分支上，后续新增的 `rain-excel-checker/dispatch/` 包仍引用旧路径 `feishu-lib/notification`，导致：
1. `main.go` 同时存在新旧两套 import（同名包冲突）
2. `dispatch/router.go` 和 `router_test.go` 引用旧路径，类型不兼容
3. 交叉编译失败：`no required module provides package .../feishu-lib`

排查过程浪费了大量时间（尝试从主仓库构建、检查 go.mod 等），根因很简单：import 路径没同步。

### Suggested Action
1. 目录重组（rename/move）后，立即 `grep -rn "old_path" --include="*.go"` 确认所有引用已更新
2. 在 worktree 上基于重组后的代码开发新功能时，先 `go build ./...` 确认编译通过再开始
3. squash 前确认所有新增文件（如 debug.go）已被 git add 并出现在 diff --stat 中

### Metadata
- Source: error
- Related Files: rain-excel-checker/main.go, rain-excel-checker/dispatch/router.go
- Tags: golang, import, refactoring, worktree, directory-restructure
- Pattern-Key: harden.import_sync_after_rename
---

## [LRN-20260527-002] correction

**Logged**: 2026-05-27T10:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: infra

### Summary
Dockerfile COPY 不保留源文件的可执行权限，必须在 RUN 中显式 chmod +x

### Details
容器中 rain-excel-checker 报 "权限不足，无法执行 /usr/local/bin/rain-excel-checker"。

Dockerfile 中 `COPY --chown=analyzer:analyzer` 只改变文件属主，不赋予可执行权限。Windows 上 `go build` 产出的文件默认没有 execute bit（NTFS 不支持 Unix 权限），COPY 到 Linux 容器后权限为 `-rw-r--r--`。

修复：`RUN chmod +x /usr/local/bin/rain-excel-checker /app/entrypoint.sh`

### Suggested Action
Dockerfile 中复制二进制或脚本文件后，始终添加 `chmod +x`，不要依赖源文件权限。

### Metadata
- Source: error
- Related Files: docker/Dockerfile
- Tags: docker, permissions, chmod, copy, executable
- Pattern-Key: harden.dockerfile_executable_permissions
---

## [LRN-20260527-003] best_practice

**Logged**: 2026-05-27T10:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
squash 前必须确认所有新增文件已被 git add，否则文件会丢失（git reset --soft 不恢复未跟踪文件）

### Details
当前会话中创建了 `handlers/debug.go`（DMMonitorDecorator），文件存在于磁盘但从未被 `git add`。squash 时 `git reset --soft` 回到 base commit，然后 `git commit` 只提交已 staged 的文件。结果 `debug.go` 没有出现在提交中。

后续编译时 `main.go` 引用了 `handlers.WrapDMMonitor`，但函数不存在，导致编译失败。排查了很久才发现文件丢失。

### Suggested Action
1. squash 前执行 `git status` 确认无未跟踪的新文件
2. 或在创建新文件后立即 `git add`，不要等 squash 时才处理
3. squash 后执行 `go build` 验证编译通过

### Metadata
- Source: error
- Related Files: backend/pkg/common/feishu/notification/handlers/debug.go
- Tags: git, squash, untracked-files, workflow
- Pattern-Key: harden.verify_all_files_before_squash
---

## [LRN-20260527-004] best_practice

**Logged**: 2026-05-27T10:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
CI 环境变量在宿主机侧统一前缀，容器内使用无前缀名，由 ci_pipeline.sh 做映射

### Details
蓝鲸流水线中环境变量命名约定：
- 宿主机侧：`CI_CLAUDE_` 前缀（`CI_CLAUDE_FEISHU_ROBOT`、`CI_CLAUDE_FEISHU_DM_APP_ID`）
- 容器内：无前缀（`FEISHU_ROBOT`、`FEISHU_DM_APP_ID`）
- 映射方式：`ci_pipeline.sh` 中 `docker run -e FEISHU_DM_APP_ID="${CI_CLAUDE_FEISHU_DM_APP_ID:-}"`

这样做的好处：
1. 流水线变量与系统变量不冲突
2. 容器内代码（entrypoint.sh）不关心 CI 前缀，保持通用
3. 本地测试可直接设置无前缀变量

### Suggested Action
CI 相关环境变量统一 `CI_CLAUDE_` 前缀，新变量遵循此约定。

### Metadata
- Source: user_feedback
- Related Files: docker/ci_pipeline.sh, docker/entrypoint.sh, docker/.env.example
- Tags: ci, env-vars, naming-convention, pipeline
---

## [LRN-20260527-005] knowledge_gap ✅ resolved

**Status**: resolved — `sync-rain-robot-pb` skill 已修复
**Summary**: 同步 rain-robot bytes 配表时必须同步对应的 pb.go proto 定义，否则 proto 反序列化失败
**根因**: skill 遗漏了 `*.pb.go` → `xcard_excel/excel/` 同步步骤
**修复**: skill 已新增配表 pb 文件同步和 `git pull` 前置检查
**Tags**: protobuf, rain-robot, sync
- See Also: LRN-20260519-005
---

## [LRN-20260617-001] best_practice

**Logged**: 2026-06-17T15:55:00+08:00
**Priority**: high
**Status**: resolved
**Area**: backend

### Summary
proto-test 保存用例需区分"覆盖"与"追加"语义，不可直接修改 SaveTestCase 为追加

### Details
用户反馈：从前端 packet-tab.vue 多选 Req 后"保存到用例"，若选择已存在用例，原有数据会被新选中的 Req 覆盖，期望改为追加。

根因分析：
- 前端 `packet-tab.vue` 的保存对话框按钮文案和提示都是"追加"，但实际调用 `TestCaseService.SaveTestCase`。
- 后端 `SaveTestCase` 底层调用 `cases.SaveTestCaseToFile`，使用 `os.WriteFile` 覆盖写入文件。

为何不把 `SaveTestCase` 直接改成追加：
- `testcase-tab.vue` 的"新增模块"/"删除消息并保存"/"保存顺序"等操作依赖覆盖语义。
- 若 `SaveTestCase` 改为追加，删除消息后旧消息仍会保留、排序后新旧消息混排、新增模块时混入旧数据。

解决方案：
1. 后端新增 `AppendTestCase` 方法，底层使用 `cases.AppendTestCaseToFile`。
2. `SaveTestCase` 保持覆盖语义不变。
3. 前端 `case-selector.requirement.ts` 暴露 `appendTestCase` 与 `saveTestCase`，注释明确语义。
4. `packet-tab.vue` 的 `confirmSaveCase` 改调 `appendTestCase`。

### Suggested Action
当同一操作在不同页面有不同语义时，优先新增独立方法而非修改现有方法语义；在注释和项目文档中明确区分覆盖/追加的适用场景。

### Resolution
- **Resolved**: 2026-06-17T16:00:00+08:00
- **Commit**: 待提交
- **Notes**: 已完成后端 `AppendTestCase`、前端调用切换、bindings 重新生成、E2E 测试补充、CLAUDE.md 文档更新、forgetful 记忆入库。

### Metadata
- Source: bug_fix
- Related Files: backend/pkg/proto-test/wails_test_case.go, backend/pkg/proto-test/cases/testcase.go, frontend/src/pages/stream-proxy/components/case-selector.requirement.ts, frontend/src/pages/stream-proxy/components/packet-tab.vue, frontend/e2e/stream-proxy/stream-proxy-save-case.spec.ts
- Tags: proto-test, testcase, append, overwrite, wails, design-decision
- See Also: forgetful memory #104
---

---

## [LRN-20260617-002] bug_fix

**Logged**: 2026-06-17T17:45:00+08:00
**Priority**: high
**Status**: resolved
**Area**: frontend

### Summary
Vue 3 子组件 emit 更新 props 后需 await nextTick()，否则读取到的仍是旧值

### Details
用户反馈：proto-test 顶部「目标服务」输入框填写 `10.254.114.174:20144`，但实际连接的是 `10.254.114.17:20144`，IP 末尾数字被截断。

根因分析：
- `target-service-config.vue` 的 `handleServerAddrChange` / `handleHttpAddrChange` 中，先 `emit('update:serverAddr', value)` 再立即调用 `saveProtoTestSettings()`。
- `saveProtoTestSettings()` 内部从 `props.serverAddr` / `props.httpAddr` 读取值，用于保存到 `ProtoTestConfig` 并重启监听。
- Vue 3 子组件的 props 在 emit 事件后不会同步更新，父组件的状态更新要等到下一次渲染周期（next tick）才会反映到子组件 props。
- 因此实际保存的是上一次（可能不完整）的旧值，表现为新输入的末尾字符被"截断"。

解决方案：
1. 将 `handleServerAddrChange`、`handleHttpAddrChange`、`handleTcpListenPortChange`、`handleHttpListenPortChange` 改为 `async`。
2. emit 后 `await nextTick()`，确保 props 与父组件状态同步后再执行保存和监听重启。

### Suggested Action
在 Vue 3 子组件中，如果 emit 更新父组件状态后需要立即读取 props 进行副作用（如持久化、调用后端），务必 `await nextTick()`；或直接在事件处理函数中使用参数 `value` 而非依赖 props。

### Resolution
- **Resolved**: 2026-06-17T17:45:00+08:00
- **Commit**: 待提交
- **Notes**: 已修改 target-service-config.vue，前端编译通过，重新构建 rain-qa-func.exe。

### Metadata
- Source: bug_fix
- Related Files: frontend/src/pages/stream-proxy/components/target-service-config.vue, frontend/src/pages/stream-proxy/CLAUDE.md
- Tags: vue3, props, nextTick, emit, async, proto-test, target-service
- See Also: LRN-20260617-001
---

---

## [LRN-20260617-003] bug_fix

**Logged**: 2026-06-17T19:25:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: frontend

### Summary
proto-test 多选保存用例误选 Ntf/Ack 行时会创建空文件，且无任何提示

### Details
用户反馈：多选保存到用例和右键增加到用例都没有成功保存数据到用例文件，但选择不存在的用例名会创建对应的空文件。

调查过程：
1. 后端 `AppendTestCase` 和 `AppendTestCaseToFile` 逻辑正常，单元测试通过。
2. 添加端到端日志后发现，Go 端收到的 `data.Messages` 为空。
3. 前端日志显示 `handleSaveToCase` 匹配到了选中的行，但提取到的 Req 消息数为 0。
4. 最终确认用户实际选中的是 Ntf 行；用例文件只保存 Req，因此 messages 被过滤为空，但前端没有任何提示，反而弹出保存对话框并创建空文件。

解决方案：
1. `message-table.vue` 多选模式下禁用非 Req 行的 checkbox，并提示"仅可选择 Req 消息"。
2. `packet-tab.vue` 的 `handleSaveToCase` 在提取到 0 条 Req 消息时立即提示用户，不再打开保存对话框。
3. 保留右键"增加到用例"对无 Req 行的禁用和提示。

### Suggested Action
在涉及"选择部分内容并保存"的 UI 中，应在选择器层面就限制可选范围（禁用/隐藏不可保存项），并在保存入口做二次兜底提示，避免用户产生"保存成功但实际为空"的误解。

### Resolution
- **Resolved**: 2026-06-17T19:25:00+08:00
- **Commit**: 8a42899
- **Notes**: 已修复多选 checkbox 禁用逻辑、添加保存前提示、补充 E2E 测试。

### Metadata
- Source: bug_fix
- Related Files: frontend/src/pages/stream-proxy/components/message-table.vue, frontend/src/pages/stream-proxy/components/packet-tab.vue, frontend/e2e/stream-proxy/stream-proxy-save-case.spec.ts
- Tags: proto-test, testcase, multi-select, req-only, ux, message-table
- See Also: LRN-20260617-001, LRN-20260617-002
---

## [LRN-20260623-001] correction

**Logged**: 2026-06-23T17:10:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
不能凭"翻译/映射代码怎么用字段"反推字段语义，应以权威定义（配表/proto）为准

### Details
调查战斗测试座位 `CustomHero.Color` 字段语义时，看到后端 `reverse_translate.go:716` 用 `countryMap`（ECountry 国家表）翻译 `hero.Color`，便推断"color=势力ID=ECountry 国家编号"并写进前后端 4 处注释。

随后查游戏配表 `IdentityEncodeRule_身份编码规则表`（`config/excel/`），发现前端 `excelIdentityColorMap` 与该配表 100% 吻合——color 实际是"身份阵营颜色编号"（主公 1~8、反贼 65、内奸 97…），与 ECountry 国家是完全不同的维度。后端用 `countryMap` 翻译 color 本身就是 bug（color∈1~16 恰好落在国家表范围内，被误译为秦/西楚…），我却被这段"能翻译"的代码误导，把 bug 行为当成正确语义写进注释，二次放大了错误。

最终：纠正 4 处注释为正确语义并标注 bug（用户选择暂不修翻译逻辑），bug 记入 `docs/TODO.md`。

### Suggested Action
1. 字段语义以权威定义（配表/proto 枚举/生成代码注释）为准，不能由"业务代码如何使用该字段"反推——使用代码本身可能就是错的。
2. 看到"A 用 B 表翻译/映射 C"时，警惕 A、C 语义是否真的一致；尤其当数值范围恰好重叠时（color 1~16 vs 国家 1~16）最易误判。
3. 发现代码语义与权威定义矛盾时，优先怀疑代码（可能是 bug），而非用代码"修正"自己的理解。

### Metadata
- Source: self_correction
- Related Files: backend/pkg/function-test/reverse_translate.go, backend/pkg/function-test/services.go, frontend/src/pages/function-test/config/Identity.ts, config/excel/IdentityEncodeRule_身份编码规则表.xlsx
- Tags: semantics, config, identity-color, reverse-translate, correction, bug
- Pattern-Key: verify.field_semantics_from_authoritative_source
---

## [LRN-20260623-002] correction

**Logged**: 2026-06-23T22:30:00+08:00
**Priority**: medium
**Status**: pending
**Area**: frontend

### Summary
naive-ui n-card 标题 slot 是 #header，不是 #title；用 UI 库 slot 前必须查文档确认 slot 名

### Details
战斗测试座位颜色扩展中，给 n-card 标题加身份色圆点时，我把原来的 `:title` prop 删掉、改用 `<template #title>`。但 naive-ui n-card 没有 `#title` slot（标题用 `title` prop 或 `#header` slot），Vue 把该具名 slot 内容直接丢弃，导致「座位N」「动作N」标题文本和色点全部消失。

用户反馈「座位号丢失、颜色没看到」后定位：Grep 项目发现 30+ 处 n-card/n-collapse-item 都用 `<template #header>` 做标题，确认 naive-ui 标题 slot 是 `#header`。改回 `#header` 后恢复。

根因：受其他 UI 库（element-plus 的 `#title`）影响想当然，未查 naive-ui 文档。

### Suggested Action
1. 用 naive-ui 组件 slot 前查官方文档确认 slot 名：n-card / n-collapse-item 标题用 `#header`（不是 `#title`）。
2. 改动 UI 组件的 slot/prop 后，必须 `wails3 dev` 实际渲染验证，不能只靠 vue-tsc——类型检查不校验 template 具名 slot 是否存在。
3. 项目内已有大量 `#header` 用法可作参考（`Grep "<template #header>"`）。

### Metadata
- Source: self_correction
- Related Files: frontend/src/pages/function-test/components/init-yanwu-panel.vue, frontend/src/pages/function-test/components/steps-panel.vue
- Tags: naive-ui, n-card, slot, frontend, correction
- Pattern-Key: verify.naive_ui_slot_name
---

## [LRN-20260709-001] best_practice

**Logged**: 2026-07-09T19:30:00+08:00
**Priority**: critical
**Status**: resolved
**Area**: infra

### Summary
Wails v3 构建 Android APK 要求 CLI、go.mod 模块、前端 @wailsio/runtime 三者版本同代,否则 android 编译失败或运行时 422

### Details
将 rain-qa-func(Wails v3 桌面应用)扩展到 Android 时连续踩坑,根因都是版本错配:
1. **CLI 与模块错配**:项目钉 `wails/v3 alpha.96`(旧命名),CLI 是 `alpha2.109`。旧模块的 android 代码与核心 API 脱节——`undefined: events.Android`/`iosMethodNames`、`runtime.Core` 参数不足。升级模块到 `alpha2.117`(与 CLI 同代)后解决。
2. **前端 runtime 与后端错配(422 根因)**:前端 `@wailsio/runtime` 钉 `alpha.78`,而 alpha2.x 模板用 npm `latest`(=`alpha.97`)。运行时协议不匹配导致 `/wails/runtime` 返回 `422 Invalid runtime call: missing object value`,所有服务调用失败。升级 `@wailsio/runtime → alpha.97` 后 422 消失、服务调用成功返回。
3. **npm 上 `@wailsio/runtime` 无 alpha2.x 版本**,最高 `alpha.97`(=模板 latest)。判断"正确版本"的权威方法:用当前 CLI `wails3 init -t vanilla` 建空项目,看它 package.json 里 `@wailsio/runtime` 的值。

### Suggested Action
1. Wails v3 项目定期对齐三者:CLI(`wails3 version`)、go.mod 模块、前端 `@wailsio/runtime`,保持在同一代(alpha2.x + runtime alpha.97)。
2. 遇 android 编译报"未定义符号/签名漂移"或运行时 422,**第一时间查三者版本是否错配**,而非深挖代码。
3. 不确定某 wails 版本对应的 runtime 时,`wails3 init -t vanilla` 空项目查 package.json。

### Metadata
- Source: bug_fix
- Related Files: go.mod, frontend/package.json, docs/Android-APK构建.md
- Tags: wails, android, version-skew, runtime, 422, protocol-mismatch
- Pattern-Key: verify.wails_version_alignment
- See Also: LRN-20260709-002
---

## [LRN-20260709-002] correction

**Logged**: 2026-07-09T19:30:00+08:00
**Priority**: high
**Status**: resolved
**Area**: infra

### Summary
官方 Wails android 构建模板只支持 macOS/Linux host,Windows 上构建需多处工具链适配

### Details
Windows host 构建 android APK 踩的工具链坑(均非显而易见,完整流程见 `docs/Android-APK构建.md`):
1. **`compile:go:shared` 缺 Windows host 分支**:官方 `build/android/Taskfile.yml` 的 host 探测 `case` 只有 Darwin/Linux,Windows 落到 `*) exit 1`。需补 `MINGW*|MSYS*|CYGWIN*) HOST_TAG="windows-x86_64"`。
2. **Windows NDK clang 是 `.cmd` 包装**:Go cgo 经 CreateProcess 直接 exec,只能运行显式 `.cmd/.exe`,不能运行同目录无扩展名的 POSIX 脚本版。故 CC/CXX 须带 `.cmd` 后缀。
3. **go-task 在 Windows 的 shell 选择不稳定**:同一份 Taskfile,`compile:go:shared` 走 sh 正常,`assemble:apk` 却走 cmd.exe 使 `./gradlew` 报 `'.' 不是内部或外部命令`。**Gradle 步骤直接用 `gradlew.bat`(PowerShell/cmd)最稳**,别依赖 task runner。
4. **`wails3 update build-assets` 不生成 `build/android/`**:它只重生成 darwin/ios/linux/windows 资产清单。android gradle 工程要用 `wails3 init -t vanilla` 建空项目后收割 `build/android/` 目录。
5. **`alpha2.109` 的 `linux_cgo.c/.h` 漏 `&& !android` 守卫**(android implies linux → 拉进 GTK 找 `<gtk/gtk.h>` 失败),需改模块缓存(非复现)。**`alpha2.117` 上游已修**,故勿降级到 alpha2.109。

### Suggested Action
1. Windows 上构建 wails android 按 `docs/Android-APK构建.md` 命令序列(已含上述适配)。
2. Gradle 步骤一律用 `gradlew.bat` 直接跑,不用 `wails3 task android:assemble:apk`。
3. wails 保持在 `alpha2.117+`(linux_cgo 守卫已修);遇 GTK 头找不到先查 wails 版本。

### Metadata
- Source: bug_fix
- Related Files: build/android/Taskfile.yml, build/android/app/build.gradle, docs/Android-APK构建.md
- Tags: wails, android, windows, ndk, cgo, go-task, gradle, host-adapter
- Pattern-Key: verify.wails_android_windows_toolchain
- See Also: LRN-20260709-001
---
