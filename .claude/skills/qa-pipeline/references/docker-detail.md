# Docker 镜像与容器详细文档

> 从 `qa-pipeline/SKILL.md` 提取，流水线涉及 Docker 的完整技术细节。

## 目录结构

```
docker/
├── Dockerfile              # Node.js Alpine + Claude Code CLI + rain-excel-checker
├── docker-compose.yml      # 本地开发测试
├── entrypoint.sh           # 容器入口脚本
├── ci_pipeline.sh          # 宿主机 CI/CD 脚本
├── build.sh                # 构建推送镜像
├── settings.json.example   # Claude Code 配置模板
├── .env.example            # 环境变量模板
├── multica-config.json     # Multica 预置认证
├── prebuilt/               # 预置二进制（multica-cli）
├── skills/                 # 容器内 Claude Code Skills
│   ├── CLAUDE.md           # Skills 开发指导
│   ├── analyze-prompt/     # 配表提交影响分析（流水线默认 skill）
│   ├── check-config-commit-message/
│   ├── excel-parser/
│   ├── game-config-relations/
│   └── multica-cli/
└── README.md
```

## 镜像内容

- **基础镜像**：`node:26-alpine`
- **运行时**：Node.js + Claude Code CLI（`@anthropic-ai/claude-code`）+ multica CLI
- **工具**：git, bash, jq, curl, python3, openpyxl
- **业务二进制**：`rain-excel-checker`（宿主机预编译，COPY 进容器）
- **npm 镜像**：淘宝 `registry.npmmirror.com`
- **pip 镜像**：清华 `pypi.tuna.tsinghua.edu.cn`

## 容器内执行流程（entrypoint.sh）

```
ci_pipeline.sh (宿主机)
    │  前置检查: 镜像、配置、仓库
    │  docker run 启动容器（传入 DM 凭证、CI_INVOKE_SKILL 等）
    │
    ▼
entrypoint.sh (容器内)
    │  1. 容器环境准备（git safe.directory、quotepath）
    │  2. 验证 skill 存在
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

`/app/notify.sh` 是容器内公共通知工具，提供：

- `dm`：飞书私聊
- `group`：飞书群卡片
- `monitor`：监控通知

由 skill 在业务需要时直接调用。`entrypoint.sh` 仅负责在启动 prompt 中说明其存在。通知策略（发给谁、何时发）由各 skill 自行决定。

## 通知策略

| 分支 | 提交类型 | 群消息 | 私聊 | rain-excel-checker |
|------|---------|--------|------|-------------------|
| `v0.0.8-pre-release` | merge/普通 | Claude Code 输出发群 | 不发 | 增量模式 |
| 其他分支 | merge | 不发 | curl 发私聊（合并者） | `-mode=full` 全量 |
| 其他分支 | 普通 | 不发 | curl 发私聊（提交者） | 增量模式 |

rain-excel-checker 始终以 `-feishuRobot=none -noDM` 运行，不发任何飞书消息。

## 退出码

| 退出码 | 状态 | 说明 |
|--------|------|------|
| 0 | 成功 | 分析完成，无异常 |
| 1 | 失败 | Claude Code 错误或内部异常 |
| 2 | 警告 | 无 Excel 文件变更，跳过分析 |
| 3 | 配置错误 | 环境变量或配置缺失 |
| 125 | Docker 错误 | Docker 守护进程问题 |
| 130 | 中断 | 用户或系统中断 |

## 构建命令

```bash
# 构建
cd docker && ./build.sh

# 构建并推送
./build.sh docker.1ms.run v1.0.0 --push

# 本地 docker compose
docker compose build
```

## 本地测试

```bash
cd docker
cp .env.example .env       # 编辑填入配置
cp settings.json.example settings.json  # 编辑填入模型和 API 地址

# 默认：分析 HEAD，不发飞书
docker compose run --rm analyzer

# 指定 commit
CI_CLAUDE_TARGET_COMMIT=abc1234 docker compose run --rm analyzer

# 测试飞书通知
CI_CLAUDE_FEISHU_ROBOT=xxx docker compose run --rm analyzer
```

## 环境变量

| 变量 | 必填 | 默认值 | 来源 | 说明 |
|------|------|--------|------|------|
| `CI_INVOKE_SKILL` | 否 | `analyze-prompt` | CI/CD 变量 | 流水线触发 Claude 需要执行的 skill（`docker/skills/` 下的目录名） |
| `CI_CLAUDE_FEISHU_ROBOT` | 否 | `none` | CI/CD 变量 | 飞书机器人 GUID（v0.0.8 分支群消息用） |
| `CI_CLAUDE_TARGET_COMMIT` | 否 | - | CI/CD 参数 | 指定 commit hash（不指定则分析 HEAD） |
| `CI_CLAUDE_FEISHU_DM_APP_ID` | 否 | - | CI/CD 变量 | 飞书私聊应用 App ID（非 v0.0.8 分支私聊用） |
| `CI_CLAUDE_FEISHU_DM_APP_SECRET` | 否 | - | CI/CD 变量 | 飞书私聊应用 App Secret |
| `BK_CI_START_USER_NAME` | 否 | 自动 | 流水线注入 | 触发人 |
| `BK_CI_HOOK_BRANCH` | 否 | 自动 | 流水线注入 | 触发分支（用于分支判断） |
| `BK_CI_HOOK_REVISION` | 否 | 自动 | 流水线注入 | 提交 ID |

## 挂载说明

```bash
# 必须挂载完整 git 仓库（含 .git 目录）
-v "/path/to/xcard-excel-config:/config-repo"

# rain-qa-func 项目代码（含 skills、前端源码、测试用例）
-v "/path/to/rain-qa-func:/workspace"

# Claude Code 配置（只读挂载）
-v "./settings.json:/home/analyzer/.claude/settings.json:ro"
```

## 容器内 Skills

| Skill | 说明 | 状态 |
|-------|------|------|
| `analyze-prompt` | 配表提交影响分析（`CI_INVOKE_SKILL` 默认 skill） | 可用 |
| `check-config-commit-message` | Commit message 质量检查 | 可用 |
| `excel-parser` | Excel 配表解析查询 | 可用 |
| `game-config-relations` | 配表关系知识库 | 可用 |
| `multica-cli` | 创建 Multica issue | 可用 |

## Debug 日志标签

容器内飞书通知操作带 debug 标签，可通过日志过滤：
- `[DEBUG-DM]` — 私聊通知相关（调用、跳过原因、token 获取、发送结果）
- `[DEBUG-群消息]` — 群消息通知相关（调用、跳过原因、发送结果）

## 飞书通知内容限制

- **群卡片**：内容截断到 15000 字符
- **私聊文本**：截断到 100KB（保守取 100KB，飞书限制 150KB）

## Docker 镜像和容器清理

Shell 脚本最后两行执行 `docker image prune -f` 和 `docker container prune -f`，清理无用的镜像和容器。这是为了防止构建机磁盘被历史镜像占满。注意这会删除**所有**未使用的镜像和已停止的容器，不仅仅是本次构建的。
