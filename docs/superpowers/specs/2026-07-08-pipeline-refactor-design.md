# 流水线改造设计：统一流程 + mjs-skills 集成 + qa-pipeline 增强

> 日期：2026-07-08
> 范围：docker/ 目录、蓝鲸流水线配置、qa-pipeline skill
> 状态：设计文档，待实现

---

## 1. 背景与目标

当前"配表检查-测试"流水线已基于 Claude Code + Docker 运行，职责分层如下：

```
蓝鲸流水线
  → 拉取目标仓库 + rain-qa-func
  → Shell 插件：build.sh → ci_pipeline.sh
    → build.sh：构建 Docker 镜像
    → ci_pipeline.sh：docker run 启动容器
      → entrypoint.sh：准备环境、启动 Claude
        → Claude 调用 $CI_INVOKE_SKILL（默认 analyze-prompt）
          → skill 执行分析并发送飞书通知
```

本次改造目标：

1. **mjs-skills 本地优先**：流水线构建时从 `https://git.devcloud.ztgame.com/Xcards/mjs-skills.git` 拉取 skill，本地缺失时回退 `npx @astron-team/skillhub`，两者都缺失则失败。
2. **流程统一复用**：前端代码、后端代码、策划配表三条流水线共用同一套脚本，只通过环境变量区分目标仓库和 skill。
3. **qa-pipeline skill 增强**：能查看线上配置、触发执行、配置 GIT/Shell 插件，后续支持批量同步多条流水线。
4. **保留回归测试用例**：将关键业务场景以测试用例形式沉淀，方便后续改造验证。

---

## 2. 现状分析

### 2.1 当前 Shell 插件 Content

来自 `qa-pipeline/references/cdp-operations.md`：

```bash
#!/usr/bin/env bash
bash /root/ExcelChecker/rain-qa-func/docker/build.sh
cp ~/.claude/settings.json.glm /root/ExcelChecker/rain-qa-func/docker/settings.json
bash /root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh
docker image prune -f
docker container prune -f
```

关键点：**每次构建都会执行 `build.sh`**，因此把 `mjs-skills` 集成到 `build.sh` 中不会增加额外构建步骤。

### 2.2 当前 skill 获取逻辑

`entrypoint.sh` 运行时才检查 skill：

1. 检查镜像内置 `/home/analyzer/.claude/skills/<skill>/SKILL.md`
2. 不存在则 `npx @astron-team/skillhub install`
3. 安装失败则退出

问题在于：
- 运行时下载依赖网络，增加失败点
- 无法使用内网 skill 仓库中的私有 skill
- 没有本地 skill 优先的降级路径

### 2.3 当前流水线参数

`ci_pipeline.sh` 中存在硬编码路径：

- `CONFIG_PATH=/root/ExcelChecker/xcard-excel-config`
- `WORKSPACE_PATH=/root/ExcelChecker/rain-qa-func`
- `CI_INVOKE_SKILL` 默认为 `analyze-prompt`

这些需要参数化，才能支持前端/后端/策划配表三条流水线复用。

### 2.4 定时任务现状

- 与当前"配表检查-测试"流水线是同一条
- 定时拉取 `v0.0.8-pre-release` 分支后触发
- 通过 `analyze-prompt` 的分支/merge 判断发送群消息
- 旧版变量（`CasePath` / `ExcelPath` / `ConfigPath` / `FeishuRobot` / `Mode`）已废弃

定时任务的全量检查机制：当 `v0.0.8-pre-release` 的 HEAD 是 merge commit 时，`analyze-prompt` 识别为 merge 并启用 `-mode=full`。

---

## 3. 详细设计

### 3.1 mjs-skills 集成到构建流程

#### 3.1.1 `build.sh` 改造

在 `docker compose build` 之前，clone 或更新 `mjs-skills`：

```bash
MJS_SKILLS_URL="${MJS_SKILLS_URL:-https://git.devcloud.ztgame.com/Xcards/mjs-skills.git}"
MJS_SKILLS_DIR="$SCRIPT_DIR/skills-external"

if [ -d "$MJS_SKILLS_DIR/.git" ]; then
    git -C "$MJS_SKILLS_DIR" pull
else
    git clone "$MJS_SKILLS_URL" "$MJS_SKILLS_DIR"
fi
```

失败时直接退出（`set -e`）。

#### 3.1.2 `Dockerfile` 改造

先 COPY 镜像内置 skills，再 COPY `mjs-skills`，后者覆盖同名 skill：

```dockerfile
COPY --chown=analyzer:analyzer docker/skills/ /home/analyzer/.claude/skills/
COPY --chown=analyzer:analyzer docker/skills-external/ /home/analyzer/.claude/skills/
```

#### 3.1.3 核心 skill 不缺失校验

新增 `docker/required-skills.txt`（实现时移除了 `excel-parser`，因该 skill 在 `docker/skills/` 与 `mjs-skills` 仓库中均不存在）：

```text
analyze-prompt
check-config-commit-message
game-config-relations
multica-cli
```

`build.sh` 构建时校验：

1. 对 required skill，优先从 `mjs-skills` 取
2. 如果 `mjs-skills` 中缺失，但 `docker/skills/` 中有，保留内置版本
3. 如果两边都缺失，构建失败

#### 3.1.4 skill 查找优先级（运行时）

`entrypoint.sh` 中简化为：

1. 镜像内 `/home/analyzer/.claude/skills/<skill>/SKILL.md`
2. `npx @astron-team/skillhub install`（回退）
3. 失败退出（exit code 3）

因为构建时已经合并了 `docker/skills/` 和 `mjs-skills/`，运行时只需要检查镜像内 skill。

### 3.2 `ci_pipeline.sh` 参数化

改为通过环境变量读取配置：

| 环境变量 | 必填 | 默认值 | 说明 |
|---------|------|--------|------|
| `CI_TARGET_REPO_PATH` | 是 | - | 要分析的目标仓库路径 |
| `CI_TARGET_REPO_URL` | 否 | - | 目标仓库 git 地址（用于文档/日志） |
| `CI_INVOKE_SKILL` | 否 | `analyze-prompt` | Claude 调用的 skill |
| `CI_TARGET_REPO_TYPE` | 否 | - | 仓库类型：`config` / `frontend` / `backend`，用于报告 |
| `CI_ENTRYPOINT_NOTIFY` | 否 | `0` | entrypoint 兜底通知开关 |
| `CI_CLAUDE_MONITOR_EMAIL` | 否 | `v-wangtengda@ztgame.com` | 监控通知邮箱 |
| `CI_CLAUDE_FEISHU_ROBOT` | 否 | `none` | 群机器人 GUID |
| `CI_CLAUDE_FEISHU_DM_APP_ID` | 否 | - | 私聊 App ID |
| `CI_CLAUDE_FEISHU_DM_APP_SECRET` | 否 | - | 私聊 App Secret |
| `CI_CLAUDE_TARGET_COMMIT` | 否 | - | 指定 commit |

移除硬编码：

- `/root/ExcelChecker/xcard-excel-config` → `CI_TARGET_REPO_PATH`
- `/root/ExcelChecker/rain-qa-func` → 保持为 `WORKSPACE_PATH`，但可配置

### 3.3 多流水线复用方案

由于当前暂无权限创建新流水线，本次先完成**脚本层参数化**。等权限到位后，蓝鲸侧配置如下：

| 流水线 | GIT 插件 1 | `CI_TARGET_REPO_PATH` | `CI_INVOKE_SKILL` | `CI_TARGET_REPO_TYPE` |
|--------|-----------|----------------------|-------------------|----------------------|
| 策划配表 | `xcard-excel-config` | `/root/ExcelChecker/xcard-excel-config` | `analyze-prompt` | `config` |
| 前端代码 | `xcard-frontend` | `/root/ExcelChecker/xcard-frontend` | `analyze-prompt-frontend` | `frontend` |
| 后端代码 | `xcard-backend` | `/root/ExcelChecker/xcard-backend` | `analyze-prompt-backend` | `backend` |

三条流水线共用同一个 `rain-qa-func`（GIT 插件 2）和同一个 `ci_pipeline.sh`。

### 3.4 定时任务

无需改造 `analyze-prompt`，但需要在文档中明确机制：

- 定时任务拉取 `v0.0.8-pre-release` 分支
- `analyze-prompt` 判断 `COMMIT_BRANCH=v0.0.8-pre-release` → 发群消息
- 若 HEAD 是 merge commit → `-mode=full` 全量检查
- 若 HEAD 是普通提交 → 增量检查

### 3.5 `qa-pipeline` skill 增强

新增以下能力：

#### 3.5.1 查看流水线配置

```bash
/qa-pipeline show-config --pipeline=<流水线ID>
```

输出：
- GIT 插件列表及仓库/分支配置
- Shell 插件 Content
- 环境变量设置
- 触发器设置（手动/远程/定时）

#### 3.5.2 触发流水线执行

```bash
/qa-pipeline run --pipeline=<流水线ID> --branch=<分支> --skill=<skill>
```

#### 3.5.3 配置插件

```bash
/qa-pipeline set-git-plugin --pipeline=<流水线ID> --repo=<仓库> --branch=<分支>
/qa-pipeline set-shell-plugin --pipeline=<流水线ID> --content=<脚本路径或内容>
```

底层基于 `web-access` skill 操作蓝鲸页面。

#### 3.5.4 批量同步（后续）

```bash
/qa-pipeline sync --source=<源流水线ID> --targets=<目标流水线ID列表>
```

### 3.6 `.gitignore` 更新

排除构建时生成的 `docker/skills-external/`：

```text
settings.json
.env
skills-external/
```

---

## 4. 回归测试用例

以下测试用例需要保留，用于验证改造后的流水线行为。

> 注：若以下描述与实际代码逻辑不符，以实际代码逻辑为准。

### 4.1 非 `v0.0.8-pre-release` 分支普通提交

**前置条件**：
- 目标仓库分支：`feature/xxx`（非 `v0.0.8-pre-release`）
- 提交类型：普通提交（非 merge）
- 有 Excel 文件变更
- `CI_INVOKE_SKILL=analyze-prompt`

**预期行为**：
1. 流水线拉取目标仓库和 `rain-qa-func`
2. `build.sh` 拉取 `mjs-skills` 并构建镜像
3. `ci_pipeline.sh` 启动容器
4. `entrypoint.sh` 调用 `analyze-prompt`
5. `analyze-prompt` 执行增量配表校验
6. 飞书私聊发送给提交者

### 4.2 `v0.0.8-pre-release` 分支普通提交

**前置条件**：
- 目标仓库分支：`v0.0.8-pre-release`
- 提交类型：普通提交
- 有 Excel 文件变更
- `CI_CLAUDE_FEISHU_ROBOT` 配置为有效群机器人 GUID

**预期行为**：
1-4 同上
5. `analyze-prompt` 执行增量配表校验
6. 飞书群消息，并 @提交者

### 4.3 `v0.0.8-pre-release` 分支 merge 提交

**前置条件**：
- 目标仓库分支：`v0.0.8-pre-release`
- 提交类型：merge commit（有两个父提交）
- 有 Excel 文件变更
- `CI_CLAUDE_FEISHU_ROBOT` 配置为有效群机器人 GUID

**预期行为**：
1-4 同上
5. `analyze-prompt` 执行全量配表校验（`-mode=full`）
6. 飞书群消息，并 @合并范围内每个 commit 的提交者

### 4.4 定时任务

**前置条件**：
- 蓝鲸定时任务配置为拉取 `v0.0.8-pre-release` 分支
- 定时触发时间为周一至周五 10:00
- `CI_CLAUDE_FEISHU_ROBOT` 配置为有效群机器人 GUID

**预期行为**：
1. 定时拉取 `v0.0.8-pre-release` 最新代码
2. 触发 `build.sh` 和 `ci_pipeline.sh`
3. `analyze-prompt` 对 `v0.0.8-pre-release` HEAD 执行全量配表校验（要求 HEAD 为 merge commit）
4. 飞书群消息，不 @任何人

### 4.5 `CI_INVOKE_SKILL` 控制 skill 调用

**前置条件**：
- `CI_INVOKE_SKILL=analyze-prompt`

**预期行为**：
- 容器内 Claude 调用 `/analyze-prompt` skill
- 若 `CI_INVOKE_SKILL` 指定的 skill 在镜像内置 skills 和 `mjs-skills` 中均不存在，则尝试 `npx @astron-team/skillhub install`
- 仍不存在则构建失败（exit code 3）

### 4.6 mjs-skills 本地优先

**前置条件**：
- `mjs-skills` 仓库中存在 `analyze-prompt/SKILL.md`
- 镜像内置 `docker/skills/analyze-prompt/SKILL.md` 也存在

**预期行为**：
- 构建镜像时，`mjs-skills` 版本覆盖镜像内置版本
- 运行时 Claude 使用 `mjs-skills` 版本

### 4.7 mjs-skills 缺失核心 skill 时的兜底

**前置条件**：
- `mjs-skills` 仓库中缺失 `analyze-prompt`
- 镜像内置 `docker/skills/analyze-prompt/` 存在

**预期行为**：
- 构建时使用镜像内置版本
- 任务可正常执行

### 4.8 build.sh 失败处理

**前置条件**：
- `mjs-skills` 仓库不可访问

**预期行为**：
- `build.sh` 中 `git clone`/`git pull` 失败
- 脚本直接退出，不再执行后续 `docker compose build`

---

## 5. 实施步骤

### 5.1 第一阶段：mjs-skills 集成 + build.sh 改造

1. 新增 `docker/required-skills.txt`
2. 修改 `docker/build.sh`：
   - clone/pull `mjs-skills`
   - 校验 required skills
3. 修改 `docker/Dockerfile`：
   - 先 COPY `docker/skills/`
   - 再 COPY `docker/skills-external/`
4. 更新 `docker/.gitignore`：排除 `skills-external/`
5. 本地 `docker compose build` 验证

### 5.2 第二阶段：ci_pipeline.sh 参数化

1. 修改 `docker/ci_pipeline.sh`：
   - 环境变量读取 `CI_TARGET_REPO_PATH`
   - 环境变量读取 `CI_TARGET_REPO_TYPE`
   - 移除硬编码路径
2. 更新 `docker/README.md` 环境变量表格
3. 本地测试：模拟蓝鲸环境变量运行 `ci_pipeline.sh`

### 5.3 第三阶段：entrypoint.sh 简化

1. 修改 `docker/entrypoint.sh`：
   - 移除复杂的 skill 安装逻辑（构建时已合并）
   - 保留镜像内检查和 npx 回退
2. 本地测试：指定不同 `CI_INVOKE_SKILL` 验证

### 5.4 第四阶段：qa-pipeline skill 增强

1. 修改 `.claude/skills/qa-pipeline/SKILL.md`：
   - 增加 `show-config` 操作
   - 增加 `run` 操作
   - 增加 `set-git-plugin` / `set-shell-plugin` 操作
2. 在现有旧流水线中做部分测试
3. 更新 `qa-pipeline/references/` 文档

### 5.5 第五阶段：文档与回归测试

1. 更新 `docker/CLAUDE.md`
2. 更新 `docker/README.md`
3. 执行 4.1-4.8 回归测试用例
4. git 提交

---

## 6. 风险与回滚

| 风险 | 影响 | 应对措施 |
|------|------|---------|
| `mjs-skills` 仓库不可访问 | 构建失败 | `build.sh` 失败即退出，不会用不完整镜像启动流水线 |
| `mjs-skills` 中 skill 与镜像内置不兼容 | 运行时错误 | 通过 required-skills 校验 + 版本锁定机制缓解 |
| 参数化后旧流水线配置未同步 | 现有流水线失败 | 旧流水线保持默认参数兼容，逐步迁移 |
| `qa-pipeline` web-access 操作不稳定 | 无法查看/修改配置 | 保留人工操作作为 fallback |
| 定时任务 HEAD 不是 merge commit | 全量变增量 | 这是现有行为，本次不改造；如需强制全量，后续增加 `CI_FORCE_FULL_MODE` |

---

## 7. 附录

### 7.1 环境变量汇总

| 变量 | 来源 | 说明 |
|------|------|------|
| `CI_TARGET_REPO_PATH` | 蓝鲸/手动 | 要分析的目标仓库本地路径 |
| `CI_TARGET_REPO_URL` | 蓝鲸/手动 | 目标仓库 git 地址（日志用） |
| `CI_TARGET_REPO_TYPE` | 蓝鲸/手动 | `config` / `frontend` / `backend` |
| `CI_INVOKE_SKILL` | 蓝鲸/手动 | Claude 调用的 skill |
| `CI_ENTRYPOINT_NOTIFY` | 蓝鲸/手动 | entrypoint 兜底通知开关 |
| `CI_CLAUDE_MONITOR_EMAIL` | 蓝鲸/手动 | 监控通知邮箱 |
| `CI_CLAUDE_FEISHU_ROBOT` | 蓝鲸变量 | 群机器人 GUID |
| `CI_CLAUDE_FEISHU_DM_APP_ID` | 蓝鲸变量 | 私聊 App ID |
| `CI_CLAUDE_FEISHU_DM_APP_SECRET` | 蓝鲸变量 | 私聊 App Secret |
| `CI_CLAUDE_TARGET_COMMIT` | 蓝鲸/手动 | 指定 commit |
| `BK_CI_START_USER_NAME` | 蓝鲸注入 | 触发人 |
| `BK_CI_HOOK_BRANCH` | 蓝鲸注入 | 触发分支 |
| `BK_CI_HOOK_REVISION` | 蓝鲸注入 | 提交 ID |
| `MJS_SKILLS_URL` | build.sh 默认/覆盖 | mjs-skills 仓库地址 |

### 7.2 相关文件索引

- 宿主机脚本：`docker/build.sh`、`docker/ci_pipeline.sh`
- 容器入口：`docker/entrypoint.sh`
- 镜像定义：`docker/Dockerfile`
- 本地测试：`docker/docker-compose.yml`
- 默认 skills：`docker/skills/`
- 外部 skills：`docker/skills-external/`（构建时生成，不提交）
- 流水线 skill：`.claude/skills/qa-pipeline/SKILL.md`
- 默认业务 skill：`docker/skills/analyze-prompt/SKILL.md`
