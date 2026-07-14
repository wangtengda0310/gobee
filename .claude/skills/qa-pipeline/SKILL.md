---
name: qa-pipeline
description: |
  QA 配表检查流水线操作指南 — 蓝鲸 CI/CD 流水线的配置、构建、测试、排查全流程。

  **必须使用此技能的场景**（满足任一）：
  - 用户提到"流水线"、"pipeline"、"CI"、"CD"、"蓝鲸"
  - 用户要求启动构建、查看构建结果、排查构建失败
  - 用户要求修改流水线配置（Job、插件、变量、触发器）
  - 用户提到"配表检查流水线"、"配表影响分析"
  - 用户要求更新 Docker 镜像、修改 entrypoint.sh 或 ci_pipeline.sh
  - 用户要求本地测试流水线逻辑（docker compose）
  - 用户提到飞书通知（群消息/私聊）与流水线的关系

  **不要触发**：
  - 用户只是写代码、修改配表检查规则（与流水线无关）
  - 用户操作其他项目的流水线（本 skill 仅适用于 rain-qa-func 配表检查流水线）
  - 用户讨论 Git 操作本身（→ git-safe）

  **判定关键**：涉及蓝鲸 CI/CD 平台上的"配表检查-测试"流水线操作时使用。
---

# QA 配表检查流水线操作指南

> 蓝鲸 CI/CD 平台上的配表检查流水线，用于分析策划配表提交的影响范围并通过飞书通知团队。

## 流水线信息

| 项目 | 内容 |
|------|------|
| **名称** | 配表检查-测试 |
| **平台** | 蓝鲸 CI/CD（devops.devcloud.ztgame.com） |
| **项目** | xcard |
| **地址** | `https://devops.devcloud.ztgame.com/projects/xcard/pipelines/p-5e007d6a5e3e472e9b3abb99b48e1064` |

**安全约束**：只操作这一条流水线，避免误操作其他流水线。

**问题排查**：遇到流水线相关问题（配置细节、操作步骤、历史决策）时，先到 forgetful 记忆系统搜索关键词"流水线"、"pipeline"查找上下文线索（记忆 ID 66-68 为用户口述原文备份）。

---

## 页面布局与配置详情

流水线页面的完整布局描述（主页、编辑页、启动构建对话框）、流程配置（Job/插件结构）、变量设置、触发设置、代码源、执行通知等详细内容见 [references/page-layout.md](references/page-layout.md)。

核心要点速览：

- **2 个 Stage**：Stage「构建触发」（3 个触发器：手动/远程/定时）+ Stage「使用新版校验工具增量校验全表规则」（Job「Linux」，**4 个插件**）
- **Job「Linux」的 4 个插件**：`config commit`(git) + `config schedule`(git) 都拉取 xcard-excel-config、`（GIT）rain-qa-func`(git)、`claude`/`test`(linuxScript)。两个 git 拉取同一仓库分别服务于 commit 触发和定时触发
- **6 个启动变量**：`CI_CLAUDE_FEISHU_ROBOT`（群消息）、`CI_CLAUDE_FEISHU_DM_APP_ID`（私聊）、`CI_CLAUDE_FEISHU_DM_APP_SECRET`（私聊）、`CI_CLAUDE_CHECK_MODE`（检查模式，`full`=全量）、`CI_INVOKE_SKILL`（调用的 skill，默认 analyze-prompt）、`CI_ENTRYPOINT_NOTIFY`（1=飞书兜底发送结果）
- **代码源**：GitLab（xcard-excel-config），`^.*$` 匹配所有分支
- **触发方式**：手动、远程、定时（周一至周五 10:00，定时任务需设 `CI_CLAUDE_CHECK_MODE=full` 触发全量检查）
- **当前分支配置（2026-07-09 实测）**：rain-qa-func → `worktree-docker`；xcard-excel-config → `v0.0.8-pre-release`

---

## Docker 镜像与容器执行

Docker 相关的完整技术细节（目录结构、镜像内容、构建命令、本地测试、容器内执行流程、通知策略、退出码、环境变量、挂载说明、debug 日志）详见 [references/docker-detail.md](references/docker-detail.md)。

以下是核心要点速览：

### 容器执行流程概要

```
ci_pipeline.sh (宿主机: 前置检查 → docker run) → entrypoint.sh (容器内: git diff → 过滤Excel → claude -p 分析 → 飞书通知)
```

无 Excel 变更则跳过分析，发通知后退出。

### 通知策略概要

- `CI_CLAUDE_CHECK_MODE=full`：强制全量检查（定时任务用），优先级最高，通知走私聊
- `v0.0.8-pre-release` 分支：Claude 输出发群消息，增量模式
- 其他分支 merge：curl 发私聊（合并者），`-mode=full` 全量检查
- 其他分支普通：curl 发私聊（提交者），增量模式
- rain-excel-checker 始终 `-feishuRobot=none -noDM`

### 退出码概要

| 退出码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | Claude Code 错误 |
| 2 | 无 Excel 变更，跳过 |
| 3 | 配置缺失 |
| 125/130 | Docker 错误/中断 |

### 本地测试快速命令

```bash
cd docker
docker compose run --rm analyzer                    # 分析 HEAD，不发飞书
CI_CLAUDE_TARGET_COMMIT=abc1234 docker compose run --rm analyzer  # 指定 commit
```

---

## 浏览器操作

通过 **web-access** skill（CDP Proxy）直接操作用户已登录的 Chrome 浏览器，天然携带登录态。

### ⚠️ 首次连接需用户在浏览器手动授权

CDP Proxy 首次连接 Chrome 时，浏览器会弹出「允许此站点调试」确认对话框（Chrome 原生安全机制，**无法通过程序默认同意**）。AI 执行操作前应提醒用户：**请到 Chrome 中点击「允许」**，否则 CDP 连接会超时失败。授权一次后，Proxy 持续运行，后续会话不再弹窗。

读取/修改流水线配置时**优先用蓝鲸 API**（fetch 直接获取 JSON model），比 UI 操作可靠得多——详见 [references/blueking-api.md](references/blueking-api.md)。

UI 操作的选择器参考见 [references/page-selectors.md](references/page-selectors.md)。

**安装**：`https://github.com/eze-is/web-access`

**重要**：不要使用 Playwright 工具操作蓝鲸平台。Playwright 打开独立浏览器（Chrome），不共享用户的登录态。web-access 通过 CDP 直连用户日常 Chrome。

详细的 CDP 操作步骤、选择器、变量修改方法见 [references/cdp-operations.md](references/cdp-operations.md)。

## 常见操作

### 启动新构建

1. 进入流水线主页 → 点击"启动新构建"
2. 确认变量值（飞书机器人 GUID、私聊应用凭证）
3. 确认 rain-qa-func 拉取分支（应为当前 worktree 分支）
4. 点击启动

### 更新代码后触发

1. 将代码推送到 rain-qa-func 的 worktree 分支
2. 在流水线编辑页面确认插件2的分支指向正确
3. 手动启动构建或等待代码源自动触发

### 修改 entrypoint.sh / skills 后

1. 修改 `docker/entrypoint.sh` 或 `docker/skills/` 下的文件
2. 提交代码并推送
3. 流水线会通过 `build.sh` 重新构建 Docker 镜像（因为 COPY 了这些文件）
4. 无需手动推送镜像（流水线每次都重新构建）

### 修改 Dockerfile 后

1. 修改 `docker/Dockerfile`
2. 提交推送后，流水线 `build.sh` 会重新构建
3. **注意**：Dockerfile 变更可能影响构建时间（如新增依赖）

### 排查构建失败

1. 进入流水线主页 → 点击失败的构建记录
2. 在"执行详情"页签中，点击各插件查看日志
3. 常见失败原因：
   - **build.sh 失败**：Docker 构建错误（网络/依赖问题）
   - **ci_pipeline.sh 失败**：前置检查不通过（镜像不存在、settings.json 缺失、仓库不存在）
   - **容器内失败**：Claude Code 执行错误、Excel 检查异常
4. 使用退出码对照表判断问题类型

### 查看 debug 日志

容器内飞书通知操作带 debug 标签（`[DEBUG-DM]` 私聊、`[DEBUG-群消息]` 群消息），可在流水线日志中过滤查看。详见 [references/docker-detail.md](references/docker-detail.md)。

---

## 高级操作

### 关联（注册）代码库

流水线 GIT 插件要引用新仓库前，需先在「代码库」页关联。GitLab 类型强制 OAuth 凭据。

```bash
/qa-pipeline register-repo --type=GitLab --name=<别名>
```

> **注意**：关联表单含 Ant Select 下拉（OAuth 凭据为远程加载），CDP 后台 tab 操作不稳定，**推荐前台手动**。详细步骤、字段选择器、CDP 限制见 [references/register-repo.md](references/register-repo.md)。

> 关联后拿到的 `codeRepoId` 属于**项目特定信息**，应记在项目文档（如 `docker/CLAUDE.md`），不放本 skill。

### 查看流水线配置

```bash
/qa-pipeline show-config --pipeline=p-5e007d6a5e3e472e9b3abb99b48e1064
```

详细步骤见 [references/show-config.md](references/show-config.md)。

### 触发流水线执行

```bash
/qa-pipeline run --pipeline=p-5e007d6a5e3e472e9b3abb99b48e1064 --skill=analyze-prompt
```

详细步骤见 [references/run-pipeline.md](references/run-pipeline.md)。

### 配置 GIT 插件

```bash
/qa-pipeline set-git-plugin --pipeline=p-xxx --repo=xcard-frontend --path=xcard-frontend
```

详细步骤见 [references/set-plugins.md](references/set-plugins.md)。

### 配置 Shell 插件

```bash
/qa-pipeline set-shell-plugin --pipeline=p-xxx --script-path=/root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh
```

详细步骤见 [references/set-plugins.md](references/set-plugins.md)。

### 批量同步流水线（实验性）

```bash
/qa-pipeline sync --source=p-source --targets=p-target1,p-target2
```

将源流水线的 GIT 插件、Shell Content、变量设置同步到目标流水线。

---

## 测试飞书通知

测试流水线时注意屏蔽飞书消息，避免垃圾消息打扰其他用户：

| 测试目标 | GUID | 用途 |
|----------|------|------|
| 私聊测试 | `v-wangtengda`（邮箱） | 发送给自己，不影响他人 |
| 群消息测试 | `db06f82a-4dad-43f6-bbef-97503e0b953a` | 测试用群机器人 |

**测试时建议**：将 `CI_CLAUDE_FEISHU_ROBOT` 设为 `none` 关闭群消息，仅用私聊测试；或使用上述测试专用 GUID。

---

## 注意事项

### 1. 飞书消息屏蔽（测试时必须注意）

测试流水线时，根据需要屏蔽飞书消息避免打扰用户。流水线有两种通知方式（群消息和私聊），测试时应分别控制：

- **私聊测试**：收件人邮箱设为 `v-wangtengda`，仅发给自己
- **群消息测试**：机器人 GUID 用 `db06f82a-4dad-43f6-bbef-97503e0b953a`（测试专用群）
- **关闭群消息**：将启动变量 `CI_CLAUDE_FEISHU_ROBOT` 设为 `none`
- **关闭私聊**：将 `CI_CLAUDE_FEISHU_DM_APP_ID` 和 `CI_CLAUDE_FEISHU_DM_APP_SECRET` 留空

容器内通知由 `entrypoint.sh` 根据分支策略自动选择发送方式（见"通知策略"章节），因此即使只想测一种通知方式，也需要了解两种都会被触发。

### 2. 代码源与触发关系

流水线的"代码源"页签配置的是**策划配表仓库**（xcard-excel-config），这意味着：
- **会触发**：策划提交配表代码到 xcard-excel-config 仓库（通过 GitLab webhook 自动触发）
- **不会触发**：开发者提交 rain-qa-func 代码（代码源未关联此仓库）

因此修改 rain-qa-func 的代码（如 entrypoint.sh、skills、Dockerfile）后，需要**手动启动构建**来验证，不会自动触发。

### 3. 分支默认行为（可能导致拉取非预期分支）

GIT 插件的分支/TAG/COMMIT 字段悬浮提示说明：

> 可为空，指定拉取的分支 默认 master。如果是来源为流水线中配置的项目触发 push/mr 触发，此项为空时拉取对应的触发分支。

这意味着：
- **手动启动**时若字段留空，默认拉取 **master** 分支
- **代码源自动触发**时若字段留空，会拉取**触发推送对应的分支**（可能是任意分支）

因此手动执行时需确认策划配表拉取的分支是否符合预期，特别是 xcard-excel-config 插件的分支配置。

### 4. rain-qa-func 分支需随 worktree 同步

插件 `（GIT）rain-qa-func` 的分支/TAG/COMMIT 随当前 worktree 变化（2026-07-09 实测为 `worktree-docker`）。开发期间会临时指向 worktree 分支以验证改动；调试完毕后需改回 `main`，否则流水线会持续拉取过时的 worktree 分支代码。**每次切换 worktree 后，必须同步修改此分支配置**。

### 5. xcard-excel-config 分支

插件 `config commit` / `config schedule` 的分支/TAG/COMMIT 当前为 `v0.0.8-pre-release`（2026-07-09 实测）。正常情况下此字段应**留空**（默认 master），由代码源自动触发时拉取对应分支；当前值是测试期间临时指定。

### 6. 启动构建界面

点击"保存并执行"按钮后弹出执行流水线界面：
- **左侧**：环境变量配置区域
- **右侧**：Stage/Job/插件列表

变量是否显示受"变量设置"页签中每个变量的**显示选项**控制：
- 不显示：变量存在但界面上不可见（使用默认值）
- 显示且必填：界面上可见且必须填写
- 满足全部条件时显示：需满足配置的条件才显示

### 7. 保存后页面 bug

修改流水线配置保存后，页面直接弹出到执行历史界面，此时**不显示执行值**（变量等配置信息为空）。需要手动重新进入当前流水线页面才能正常显示。这是蓝鲸平台的已知问题，不是配置错误。

### 8. settings.json（不要提交到代码库）

`settings.json` 包含 Claude Code 的模型和 API 配置。流水线 Shell 脚本中通过 `cp ~/.claude/settings.json.<model> /root/ExcelChecker/rain-qa-func/docker/settings.json` 复制到 docker 目录，其中 `<model>` 后缀随当前使用的模型变化（2026-07-09 实测为 `.kimi`，曾用 `.glm`）。该源文件由流水线宿主机 `~/.claude/` 提供，包含敏感信息（API 地址、权限配置），**不应提交到代码库**。`docker/.gitignore` 已排除 `settings.json`。

### 9. rain-excel-checker 二进制（开发机编译 + 提交 git）

`docker/rain-excel-checker` 是预编译的 Linux amd64 二进制，**提交在 git 里**（通过 `COPY` 进镜像）。源码位于 `cmd/rain-excel-checker/`。流水线宿主机不需要 go，`build.sh` 不做编译。

**开发机编译**（改 Go 代码后执行，然后提交二进制）：
```bash
GOOS=linux GOARCH=amd64 go build -o docker/rain-excel-checker ./cmd/rain-excel-checker/
git add docker/rain-excel-checker
```

> 不用多阶段 Docker 构建（rain-robot 外部 module 依赖阻碍）、不在流水线宿主机现场编译（PWD=/tmp 导致 `go.mod not found`）。详见 [docker/CLAUDE.md](docker/CLAUDE.md)「二进制构建方式」。

### 10. multica 预置二进制

`docker/prebuilt/multica-cli-*.tar.gz` 预置在仓库中避免流水线网络问题（流水线环境可能无法访问 GitHub）。更新版本需在能访问 GitHub 的环境手动下载，同时更新 Dockerfile 中的 `MULTICA_VERSION`。详见 [references/docker-detail.md](references/docker-detail.md)。

### 11. 流水线页面已知 bug

- **运行期间切换历史记录**：执行详情可能显示为其他执行日志
- **保存后执行历史**：保存配置后弹出执行历史不显示执行值，需重新进入

### 12. 触发分支规则

代码源页签的触发分支规则**不能为空**，否则流水线无法触发。经过测试验证，当前配置 `^.*$` 匹配所有分支提交。悬浮提示说明支持正则匹配 ref 全称或简称，如 `^refs/heads/(master|feature/.*)$` 可精确匹配 master 和 feature 开头的分支。

### 13. 定时任务变量（旧版格式）

定时触发任务使用旧版变量格式，与新 Claude 流水线变量**完全不同**，不要混淆：

| 定时任务变量 | 对应关系 |
|-------------|---------|
| `CasePath=../rain-qa-func/excel_cases` | 旧版 Go 运行模式的 cases 路径 |
| `ExcelPath=../xcard-excel-config/excel` | 旧版 Go 运行模式的 Excel 路径 |
| `ConfigPath=v0.0.8-pre-release` | 旧版分支配置 |
| `FeishuRobot=36732a0b-...` | 旧版飞书通知 |
| `Mode=full` | 旧版检查模式 |

这些变量服务于旧版纯 Go 运行模式的流水线，新版 Claude Code 流水线使用 `CI_CLAUDE_*` 系列变量。

**新版定时全量检查**：在定时任务中设置环境变量 `CI_CLAUDE_CHECK_MODE=full`，`analyze-prompt` skill 会据此强制 `-mode=full`，优先级高于分支/merge 判断。

### 14. 构建超时

默认超时 900 分钟。Claude Code 分析可能耗时较长（取决于配表变更复杂度），一般 5-15 分钟内完成，复杂场景可能更久。

### 15. 飞书通知内容限制与镜像清理

详见 [references/docker-detail.md](references/docker-detail.md) 中的"飞书通知内容限制"和"Docker 镜像和容器清理"章节。
