# 配表影响分析 Docker 配置

在 Docker 容器中运行 Claude Code CLI，分析策划配表提交的影响范围，通过飞书通知团队。

## 目录结构

```
docker/
├── Dockerfile              # Node.js 运行时 + Claude Code CLI
├── docker-compose.yml      # 本地开发测试
├── entrypoint.sh           # 容器入口（仅运行时骨架）
├── notify.sh               # 飞书通知公共脚本（dm/group/monitor）
├── ci_pipeline.sh          # 宿主机 CI/CD 脚本
├── build.sh                # 构建推送镜像（搬运 mjs-skills + docker compose build）
├── local-ci-simulate.sh    # 本地模拟蓝鲸流水线（按 plugin 顺序执行）
├── required-skills.txt     # 核心 skill 清单（build.sh 构建时校验）
├── settings.json.example   # Claude Code 配置模板
├── .env.example            # 环境变量模板
├── skills/                 # Claude Code Skills 目录
│   ├── analyze-prompt/     # 配表提交影响分析（流水线默认 skill，含完整业务逻辑）
│   ├── check-config-commit-message/
│   ├── game-config-relations/
│   └── multica-cli/
└── README.md
```

## 快速开始

### 1. 准备配置

```bash
cd docker

# 环境变量
cp .env.example .env
# 编辑 .env 填入 CI_CLAUDE_FEISHU_ROBOT 等配置

# Claude Code 配置（连接国内模型）
cp settings.json.example settings.json
# 编辑 settings.json 填入模型和 API 地址
```

### 2. 构建镜像

build.sh 需要 mjs-skills 源目录（`MJS_SKILLS_SOURCE`，默认 `/root/ExcelChecker/mjs-skills`，流水线由蓝鲸 GIT 插件准备）。

```bash
# 推荐：用 local-ci-simulate.sh 一键拉取 mjs-skills + 构建 + 运行（见下方「本地模拟流水线」）

# 手动方式：先准备 mjs-skills 源，再构建
git clone https://git.devcloud.ztgame.com/Xcards/mjs-skills.git /tmp/mjs-skills
MJS_SKILLS_SOURCE=/tmp/mjs-skills ./build.sh
```

### 3. 本地模拟流水线（推荐，完整复现蓝鲸）

`local-ci-simulate.sh` 按蓝鲸 plugin 顺序执行（拉 config → mjs-skills → rain-qa-func → build → run），无需触发真实流水线即可验证改动：

```bash
# 一键模拟（命名参数）
./local-ci-simulate.sh --target-repo https://git.devcloud.ztgame.com/v-tangfangda/rain-qa-func.git \
                       --skill-repo  https://git.devcloud.ztgame.com/Xcards/mjs-skills.git

# 位置参数简写
./local-ci-simulate.sh <rain-qa-func-url> <mjs-skills-url>

# 只验证镜像构建（不跑 ci_pipeline.sh）
./local-ci-simulate.sh --target-repo <url> --skill-repo <url> --build-only
```

对应蓝鲸 plugin：`[Plugin 1] config → [Plugin 2] mjs-skills → [Plugin 3] rain-qa-func → [Plugin 4] Shell(build+run)`。

### 4. 容器本地调试（仅运行，不重新拉取）

```bash
# 默认：分析 HEAD，不发飞书通知，实时输出到终端
docker compose run --rm analyzer

# 指定 commit 测试（不发飞书）
CI_CLAUDE_TARGET_COMMIT=abc1234 docker compose run --rm analyzer

# 测试飞书通知（分析 HEAD）
CI_CLAUDE_FEISHU_ROBOT=db06f82a-4dad-43f6-bbef-97503e0b953a docker compose run --rm analyzer

# 指定 commit + 飞书通知
CI_CLAUDE_TARGET_COMMIT=abc1234 CI_CLAUDE_FEISHU_ROBOT=xxx docker compose run --rm analyzer

# 指定 Claude 执行的 skill（默认 analyze-prompt，即原有配表影响分析逻辑）
CI_INVOKE_SKILL=analyze-prompt docker compose run --rm analyzer

# 指定脚本层监控邮箱（默认 v-wangtengda@ztgame.com）
CI_CLAUDE_MONITOR_EMAIL=your-email@ztgame.com docker compose run --rm analyzer
```

## 行为控制

由分支名和提交类型共同决定通知方式：

| 分支 | 提交类型 | 群消息 | 私聊 | rain-excel-checker |
|------|---------|--------|------|-------------------|
| v0.0.8-pre-release | merge/普通 | claude code 输出发群 | 不发 | 增量模式 |
| 其他分支 | merge | 不发 | curl 发私聊（合并者） | `-mode=full` 全量 |
| 其他分支 | 普通 | 不发 | curl 发私聊（提交者） | 增量模式 |

rain-excel-checker 始终以 `-feishuRobot=none -noDM` 运行，不发任何飞书消息。

### 分析 HEAD（CI/CD 自动）

```bash
docker run --rm \
  -e CI_CLAUDE_FEISHU_ROBOT=xxx \
  -v /path/to/config:/config-repo \
  -v ./settings.json:/home/analyzer/.claude/settings.json:ro \
  rain-config-analyzer:latest
```

### 分析指定 commit

```bash
docker run --rm \
  -e CI_CLAUDE_TARGET_COMMIT=abc1234 \
  -e CI_CLAUDE_FEISHU_ROBOT=xxx \
  -v /path/to/config:/config-repo \
  -v ./settings.json:/home/analyzer/.claude/settings.json:ro \
  rain-config-analyzer:latest
```

### 本地调试

```bash
docker compose run --rm analyzer
```

## CI/CD 集成

### 流水线步骤

```
配表影响分析
├── 阶段1: 准备
│   ├── 拉取代码(GIT): xcard-excel-config → /root/ExcelChecker/xcard-excel-config
│   └── 拉取代码(GIT): rain-qa-func → /root/ExcelChecker/rain-qa-func
├── 阶段2: 分析
│   └── Shell: /root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh
└── 触发器: Webhook (配表仓库代码变更) / 手动 / 定时
```

### 环境变量

| 变量 | 必填 | 默认值 | 来源 | 说明 |
|------|------|--------|------|------|
| `CI_TARGET_REPO_PATH` | 是 | `/root/ExcelChecker/xcard-excel-config` | CI/CD 变量 | 要分析的目标仓库路径 |
| `CI_TARGET_REPO_TYPE` | 否 | `config` | CI/CD 变量 | 仓库类型：`config` / `frontend` / `backend` |
| `MJS_SKILLS_URL` | 否 | `https://git.devcloud.ztgame.com/Xcards/mjs-skills.git` | build.sh | mjs-skills 仓库地址 |
| `CI_INVOKE_SKILL` | 否 | `analyze-prompt` | CI/CD 变量 | 流水线触发 Claude 需要执行的 skill（`docker/skills/` 下的目录名），不存在时会尝试从 registry 自动安装 |
| `CI_ENTRYPOINT_NOTIFY` | 否 | `0` | CI/CD 变量 | `1` 时 entrypoint 在 Claude 执行完后兜底把输出发飞书（skill 自身不发通知时用，如 mail-check）；analyze-prompt 自己发通知应保持 `0` |
| `CI_CLAUDE_MONITOR_EMAIL` | 否 | `v-wangtengda@ztgame.com` | CI/CD 变量 | 脚本层监控通知接收邮箱 |
| `CI_CLAUDE_FEISHU_ROBOT` | 否 | `none` | CI/CD 变量 | 飞书机器人 GUID（v0.0.8 分支群消息用） |
| `CI_CLAUDE_TARGET_COMMIT` | 否 | - | CI/CD 参数 | 指定 commit hash（不指定则分析 HEAD） |
| `CI_CLAUDE_FEISHU_DM_APP_ID` | 否 | - | CI/CD 变量 | 飞书私聊应用 App ID（非 v0.0.8 分支私聊用） |
| `CI_CLAUDE_FEISHU_DM_APP_SECRET` | 否 | - | CI/CD 变量 | 飞书私聊应用 App Secret |
| `BK_CI_START_USER_NAME` | 否 | 自动 | 流水线注入 | 触发人 |
| `BK_CI_HOOK_BRANCH` | 否 | 自动 | 流水线注入 | 触发分支（用于分支判断，替代 git rev-parse） |
| `BK_CI_HOOK_REVISION` | 否 | 自动 | 流水线注入 | 提交 ID |

### 多流水线复用

前端、后端、策划配表三条流水线可共用同一套 `docker/` 脚本，只通过蓝鲸环境变量区分：

| 流水线 | CI_TARGET_REPO_PATH | CI_INVOKE_SKILL | CI_TARGET_REPO_TYPE |
|--------|---------------------|-----------------|---------------------|
| 策划配表 | `/root/ExcelChecker/xcard-excel-config` | `analyze-prompt` | `config` |
| 前端代码 | `/root/ExcelChecker/xcard-frontend` | `analyze-prompt-frontend` | `frontend` |
| 后端代码 | `/root/ExcelChecker/xcard-backend` | `analyze-prompt-backend` | `backend` |

### 调试

容器内所有飞书通知操作带 debug 日志标签，可通过日志过滤：
- `[DEBUG-DM]` — 私聊通知相关（调用、跳过原因、token 获取、发送结果）
- `[DEBUG-群消息]` — 群消息通知相关（调用、跳过原因、发送结果）

## 挂载说明

### 必须挂载完整 git 仓库

```bash
# 正确: 挂载完整仓库（含 .git 目录）
-v "/path/to/xcard-excel-config:/config-repo"

# 错误: 只挂载子目录（无法执行 git 命令）
-v "/path/to/xcard-excel-config/excel:/config-repo"
```

### settings.json

容器内路径: `/home/analyzer/.claude/settings.json`

用于配置 Claude Code 的模型、API 地址、权限等。必须以只读方式挂载。

## Skills

`docker/skills/` 在构建镜像时 COPY 到容器内 `/home/analyzer/.claude/skills/`（用户级 skill，`claude -p` 可直接调用）。

| Skill | 说明 | 状态 |
|-------|------|------|
| `analyze-prompt` | 配表提交影响分析（流水线默认，`CI_INVOKE_SKILL` 默认值） | 可用 |
| `check-config-commit-message` | Commit message 质量检查 | 可用 |
| `game-config-relations` | 配表关系知识库 | 可用 |
| `multica-cli` | 创建 Multica issue | 可用 |

开发规范见 [skills/CLAUDE.md](skills/CLAUDE.md)。

## 执行流程

```
CI/CD 触发
    │
    ▼
ci_pipeline.sh (宿主机)
    │  前置检查: 镜像、配置、仓库
    │  docker run 启动容器（传入 DM 凭证、CI_INVOKE_SKILL 等）
    │
    ▼
entrypoint.sh (容器内)
    │  1. 容器环境准备（git safe.directory、quotepath）
    │  2. 校验 skill 是否存在；不存在则调用 npx @astron-team/skillhub 从 registry 安装
    │  3. 启动 Claude，提示词告知：
    │     - 调用 $CI_INVOKE_SKILL 技能
    │     - 公共工具 /app/notify.sh 可用
    │     - 原始上下文环境变量清单
    │
    ▼
Claude 调用 skill（默认 analyze-prompt）
    │  skill 内部完成完整业务逻辑：
    │    - git diff 过滤 Excel 文件
    │    - 分支/merge 判断
    │    - 构建并执行 rain-excel-checker
    │    - 调用 /check-config-commit-message 等辅助 skill
    │    - 生成影响分析报告
    │    - 调用 /app/notify.sh 发送飞书通知
    │    - 调用 /app/notify.sh monitor 发送监控通知
    │
    ▼
飞书通知（群消息或私聊） + 监控通知
```

## 公共通知脚本

`/app/notify.sh` 是容器内公共通知工具，由 `entrypoint.sh` 在启动 prompt 中向 Claude 说明。

```bash
# 私聊
/app/notify.sh dm "houzhensong@ztgame.com" "消息内容"

# 群卡片
/app/notify.sh group "标题" "markdown 内容"

# 监控通知
/app/notify.sh monitor "类型" "目标" "摘要"
```

通知失败不会阻塞流水线。

## 与旧 scripts/ 的关系

`scripts/` 目录下的旧配置（纯 Go 运行模式）保留不动，现有流水线继续使用。本目录（`docker/`）是新的 Claude Code 方案，两者独立运行。

后续确认新方案稳定后，可逐步迁移旧流水线。
