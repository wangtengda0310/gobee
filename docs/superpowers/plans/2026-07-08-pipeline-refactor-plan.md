# 流水线改造实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `mjs-skills` 集成到 Docker 构建流程，参数化 `ci_pipeline.sh` 以支持多流水线复用，简化 `entrypoint.sh`，并增强 `qa-pipeline` skill 的线上配置操作能力。

**Architecture:** 通过 `build.sh` 在构建时拉取 `mjs-skills` 并与镜像内置 skills 合并；`ci_pipeline.sh` 通过环境变量接收目标仓库和 skill 名称；`entrypoint.sh` 仅做运行时 skill 存在性检查；`qa-pipeline` skill 通过 `web-access` 操作蓝鲸流水线页面。

**Tech Stack:** Bash, Docker, Docker Compose, Claude Code Skills, Git

## Global Constraints

- `mjs-skills` 仓库地址：`https://git.devcloud.ztgame.com/Xcards/mjs-skills.git`
- 默认 skill：`analyze-prompt`
- `build.sh` 失败时直接退出（`set -e`）
- `mjs-skills` 覆盖镜像内置同名 skill，但 required skills 不能缺失
- 蓝鲸 Shell 插件 Content 保持：`build.sh` → `cp settings.json` → `ci_pipeline.sh` → `docker prune`
- 当前 `CI_INVOKE_SKILL` 变量名保持不变
- 本地开发环境：WSL + Docker

---

## File Structure

| 文件 | 职责 | 操作 |
|------|------|------|
| `docker/build.sh` | 构建镜像前拉取 `mjs-skills` 并校验 required skills | 修改 |
| `docker/Dockerfile` | 先 COPY 内置 skills，再 COPY `mjs-skills` | 修改 |
| `docker/.gitignore` | 排除构建时生成的 `skills-external/` | 修改 |
| `docker/required-skills.txt` | 核心 skill 清单，用于构建时校验 | 新增 |
| `docker/ci_pipeline.sh` | 参数化目标仓库路径和 skill 名称 | 修改 |
| `docker/entrypoint.sh` | 简化 skill 查找逻辑 | 修改 |
| `.claude/skills/qa-pipeline/SKILL.md` | 增加 show-config / run / set-plugin 能力 | 修改 |
| `.claude/skills/qa-pipeline/references/show-config.md` | 查看流水线配置详细操作 | 新增 |
| `.claude/skills/qa-pipeline/references/run-pipeline.md` | 触发流水线执行详细操作 | 新增 |
| `.claude/skills/qa-pipeline/references/set-plugins.md` | 配置 GIT/Shell 插件详细操作 | 新增 |
| `docker/CLAUDE.md` | 更新构建流程和常见问题 | 修改 |
| `docker/README.md` | 更新环境变量和本地测试说明 | 修改 |
| `docs/superpowers/specs/2026-07-08-pipeline-refactor-design.md` | 设计文档 | 只读参考 |

---

## Task 1: mjs-skills 集成到 build.sh

**Files:**
- Create: `docker/required-skills.txt`
- Modify: `docker/build.sh:1-89`

**Interfaces:**
- Consumes: `MJS_SKILLS_URL`（环境变量，可选）
- Produces: `docker/skills-external/` 目录（构建时生成，不提交）

### Step 1: 创建 required-skills.txt

- [ ] 在 `docker/required-skills.txt` 中写入核心 skill 清单：

```text
analyze-prompt
check-config-commit-message
excel-parser
game-config-relations
multica-cli
```

- [ ] 运行命令验证文件创建：

```bash
cat docker/required-skills.txt
```

Expected output:

```
analyze-prompt
check-config-commit-message
excel-parser
game-config-relations
multica-cli
```

### Step 2: 修改 build.sh 拉取 mjs-skills

- [ ] 在 `docker/build.sh` 中，找到 `CLAUDE_CODE_VERSION="${CLAUDE_CODE_VERSION:-2.1.141}"` 行之后，插入 mjs-skills 配置：

```bash
# mjs-skills 仓库配置
MJS_SKILLS_URL="${MJS_SKILLS_URL:-https://git.devcloud.ztgame.com/Xcards/mjs-skills.git}"
MJS_SKILLS_DIR="$SCRIPT_DIR/skills-external"
```

- [ ] 在 `echo "========================================="` 之前，插入 mjs-skills 拉取逻辑：

```bash
# 拉取/更新 mjs-skills 仓库
if [ -d "$MJS_SKILLS_DIR/.git" ]; then
    echo "更新 mjs-skills..."
    git -C "$MJS_SKILLS_DIR" pull
else
    echo "克隆 mjs-skills..."
    git clone "$MJS_SKILLS_URL" "$MJS_SKILLS_DIR"
fi
```

- [ ] 在 `docker compose build` 之前，插入 required skills 校验逻辑：

```bash
# 校验 required skills 不缺失
if [ -f "$SCRIPT_DIR/required-skills.txt" ]; then
    echo "校验核心 skill..."
    MISSING_SKILLS=""
    while IFS= read -r skill_name || [ -n "$skill_name" ]; do
        [ -z "$skill_name" ] && continue
        if [ ! -d "$MJS_SKILLS_DIR/$skill_name" ] && [ ! -d "$SCRIPT_DIR/skills/$skill_name" ]; then
            MISSING_SKILLS="$MISSING_SKILLS $skill_name"
        fi
    done < "$SCRIPT_DIR/required-skills.txt"

    if [ -n "$MISSING_SKILLS" ]; then
        echo "错误: 以下核心 skill 缺失:$MISSING_SKILLS"
        echo "请确认 mjs-skills 或 docker/skills/ 中存在这些 skill"
        exit 1
    fi
    echo "核心 skill 校验通过"
fi
```

### Step 3: 测试 build.sh 语法

- [ ] 运行 bash 语法检查：

```bash
bash -n docker/build.sh
```

Expected output: 无输出（退出码 0）

- [ ] 验证退出码：

```bash
echo $?
```

Expected output:

```
0
```

### Step 4: 测试 mjs-skills 拉取（可选，需要网络）

- [ ] 运行 build.sh 的前置部分（仅拉取 skills，不构建镜像）：

```bash
cd docker && bash -c '
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MJS_SKILLS_URL="${MJS_SKILLS_URL:-https://git.devcloud.ztgame.com/Xcards/mjs-skills.git}"
MJS_SKILLS_DIR="$SCRIPT_DIR/skills-external"
if [ -d "$MJS_SKILLS_DIR/.git" ]; then
    git -C "$MJS_SKILLS_DIR" pull
else
    git clone "$MJS_SKILLS_URL" "$MJS_SKILLS_DIR"
fi
'
```

Expected output: 克隆成功或已是最新

### Step 5: Commit

- [ ] 提交变更：

```bash
git add docker/required-skills.txt docker/build.sh
git commit -m "feat(docker): build.sh 构建时拉取 mjs-skills 并校验核心 skill

- 新增 docker/required-skills.txt 核心 skill 清单
- build.sh 构建前 clone/pull mjs-skills 到 docker/skills-external/
- 校验 required skills 在 mjs-skills 或 docker/skills/ 中存在
- 缺失核心 skill 时构建失败

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Task 2: Dockerfile 合并内置 skills 和 mjs-skills

**Files:**
- Modify: `docker/Dockerfile:62-65`
- Modify: `docker/.gitignore:1-2`

**Interfaces:**
- Consumes: `docker/skills/`（镜像内置）、`docker/skills-external/`（构建时生成）
- Produces: 容器内 `/home/analyzer/.claude/skills/`

### Step 1: 修改 Dockerfile COPY 顺序

- [ ] 找到 `docker/Dockerfile` 中的以下行：

```dockerfile
COPY --chown=analyzer:analyzer docker/skills/ /home/analyzer/.claude/skills/
```

- [ ] 在其后插入：

```dockerfile
# 合并 mjs-skills，同名 skill 覆盖镜像内置版本
COPY --chown=analyzer:analyzer docker/skills-external/ /home/analyzer/.claude/skills/
```

修改后的片段应类似：

```dockerfile
COPY --chown=analyzer:analyzer docker/skills/ /home/analyzer/.claude/skills/
COPY --chown=analyzer:analyzer docker/skills-external/ /home/analyzer/.claude/skills/
COPY --chown=analyzer:analyzer docker/entrypoint.sh /app/entrypoint.sh
```

### Step 2: 更新 .gitignore

- [ ] 在 `docker/.gitignore` 末尾追加：

```text
skills-external/
```

修改后的 `docker/.gitignore`：

```text
settings.json
.env
skills-external/
```

### Step 3: 本地构建测试

- [ ] 确保 `docker/skills-external/` 目录存在（运行过 Task 1 的 build.sh 或手动创建模拟目录）

- [ ] 运行 docker compose build：

```bash
cd docker && docker compose build
```

Expected output: 构建成功，无 ERROR

### Step 4: 验证镜像内 skills

- [ ] 运行容器并检查 skills 目录：

```bash
cd docker && docker run --rm rain-config-analyzer:latest ls -la /home/analyzer/.claude/skills/
```

Expected output: 同时包含 `docker/skills/` 和 `docker/skills-external/` 中的 skill 目录

### Step 5: Commit

- [ ] 提交变更：

```bash
git add docker/Dockerfile docker/.gitignore
git commit -m "feat(docker): Dockerfile 合并 mjs-skills 与内置 skills

- 先 COPY docker/skills/ 再 COPY docker/skills-external/
- mjs-skills 同名 skill 覆盖镜像内置版本
- .gitignore 排除构建生成的 skills-external/

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Task 3: ci_pipeline.sh 参数化

**Files:**
- Modify: `docker/ci_pipeline.sh:1-208`

**Interfaces:**
- Consumes: `CI_TARGET_REPO_PATH`, `CI_TARGET_REPO_TYPE`, `CI_INVOKE_SKILL`, `CI_ENTRYPOINT_NOTIFY`, `CI_CLAUDE_MONITOR_EMAIL`, `CI_CLAUDE_FEISHU_ROBOT`, `CI_CLAUDE_FEISHU_DM_APP_ID`, `CI_CLAUDE_FEISHU_DM_APP_SECRET`, `CI_CLAUDE_TARGET_COMMIT`
- Produces: `docker run` 命令透传上述变量

### Step 1: 替换硬编码路径为环境变量

- [ ] 将 `docker/ci_pipeline.sh` 中的默认配置部分：

```bash
CONFIG_PATH="${CONFIG_PATH:-/root/ExcelChecker/xcard-excel-config}"
WORKSPACE_PATH="${WORKSPACE_PATH:-/root/ExcelChecker/rain-qa-func}"
```

替换为：

```bash
# 目标仓库路径（各流水线不同）
CI_TARGET_REPO_PATH="${CI_TARGET_REPO_PATH:-/root/ExcelChecker/xcard-excel-config}"
# rain-qa-func 工作空间路径
WORKSPACE_PATH="${WORKSPACE_PATH:-/root/ExcelChecker/rain-qa-func}"
# 仓库类型（config/frontend/backend），仅用于日志和报告
CI_TARGET_REPO_TYPE="${CI_TARGET_REPO_TYPE:-config}"
```

### Step 2: 更新所有 CONFIG_PATH 引用

- [ ] 将脚本中所有 `CONFIG_PATH` 替换为 `CI_TARGET_REPO_PATH`（共 3 处左右）

- [ ] 修改帮助信息中的环境变量说明：

```bash
echo "  CI_TARGET_REPO_PATH     目标仓库路径（默认: /root/ExcelChecker/xcard-excel-config）"
echo "  CI_TARGET_REPO_TYPE     目标仓库类型（默认: config）"
```

### Step 3: 透传 CI_TARGET_REPO_TYPE 到容器

- [ ] 在 `docker run` 命令的环境变量列表中，新增：

```bash
    -e CI_TARGET_REPO_TYPE="$CI_TARGET_REPO_TYPE" \
```

### Step 4: 更新前置检查输出

- [ ] 将输出信息从：

```bash
echo "配置仓库: $CONFIG_PATH"
```

改为：

```bash
echo "目标仓库: $CI_TARGET_REPO_PATH"
echo "仓库类型: $CI_TARGET_REPO_TYPE"
```

### Step 5: 语法检查

- [ ] 运行：

```bash
bash -n docker/ci_pipeline.sh
```

Expected output: 无输出，退出码 0

### Step 6: 本地模拟测试

- [ ] 准备本地测试环境：

```bash
cp docker/settings.json.example docker/settings.json
# 编辑 settings.json 填入实际 API 配置（如本地测试可保持 example）
```

- [ ] 设置环境变量并运行 ci_pipeline.sh --help：

```bash
export CI_TARGET_REPO_PATH=/tmp/test-config
export CI_TARGET_REPO_TYPE=config
export CI_INVOKE_SKILL=analyze-prompt
cd docker && bash ci_pipeline.sh --help
```

Expected output: 帮助信息中显示 `CI_TARGET_REPO_PATH` 和 `CI_TARGET_REPO_TYPE`

### Step 7: Commit

- [ ] 提交变更：

```bash
git add docker/ci_pipeline.sh
git commit -m "feat(docker): ci_pipeline.sh 支持参数化目标仓库

- 使用 CI_TARGET_REPO_PATH 替代硬编码的 CONFIG_PATH
- 新增 CI_TARGET_REPO_TYPE 用于区分 config/frontend/backend
- docker run 透传 CI_TARGET_REPO_TYPE
- 更新帮助信息

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Task 4: 简化 entrypoint.sh

**Files:**
- Modify: `docker/entrypoint.sh:156-197`

**Interfaces:**
- Consumes: `CI_INVOKE_SKILL`, `SKILLS_DIR`
- Produces: Claude 启动 prompt

### Step 1: 简化 skill 查找逻辑

- [ ] 找到 `entrypoint.sh` 中的 skill 安装逻辑（约 156-197 行）

- [ ] 替换为更简洁的版本：

```bash
# skill 不存在时尝试从 registry 自动安装
if [ ! -f "$SKILLS_DIR/$CI_INVOKE_SKILL/SKILL.md" ]; then
    log_info "Skill 不存在，尝试从 registry 安装: $CI_INVOKE_SKILL (namespace: mjs)"
    if npx @astron-team/skillhub@latest install "$CI_INVOKE_SKILL" \
        --namespace mjs \
        --registry http://192.168.116.60:8090 \
        --dir "$SKILLS_DIR" \
        2>&1; then
        log_info "Skill 安装成功: $CI_INVOKE_SKILL"
    else
        log_error "Skill 安装失败: $CI_INVOKE_SKILL"
        exit 3
    fi
fi

# 安装后再校验一次
if [ ! -f "$SKILLS_DIR/$CI_INVOKE_SKILL/SKILL.md" ]; then
    log_error "Skill 安装后仍不存在: $SKILLS_DIR/$CI_INVOKE_SKILL/SKILL.md"
    exit 3
fi
```

注意：保留原有逻辑即可，因为构建时已经合并了 `docker/skills/` 和 `mjs-skills/`，所以运行时大部分时间会命中镜像内 skill。

### Step 2: 语法检查

- [ ] 运行：

```bash
bash -n docker/entrypoint.sh
```

Expected output: 无输出，退出码 0

### Step 3: 本地测试

- [ ] 使用 docker compose 运行本地测试：

```bash
cd docker
cp .env.example .env
# 编辑 .env 设置 CI_CLAUDE_CONFIG_REPO_PATH 为本地 config 仓库路径
docker compose run --rm analyzer
```

Expected output: 容器启动，Claude 调用 `analyze-prompt`，输出到终端

### Step 4: Commit

- [ ] 提交变更：

```bash
git add docker/entrypoint.sh
git commit -m "refactor(docker): 简化 entrypoint.sh skill 查找逻辑

- 构建时已合并 mjs-skills，运行时优先使用镜像内 skill
- 保留 npx registry 回退安装逻辑
- 安装失败返回 exit code 3

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Task 5: 增强 qa-pipeline skill

**Files:**
- Create: `.claude/skills/qa-pipeline/references/show-config.md`
- Create: `.claude/skills/qa-pipeline/references/run-pipeline.md`
- Create: `.claude/skills/qa-pipeline/references/set-plugins.md`
- Modify: `.claude/skills/qa-pipeline/SKILL.md`

**Interfaces:**
- Consumes: `web-access` skill, 蓝鲸流水线 URL
- Produces: 流水线配置读取/修改能力

### Step 1: 创建 show-config.md

- [ ] 在 `.claude/skills/qa-pipeline/references/show-config.md` 中写入：

```markdown
# 查看流水线配置

使用 `web-access` skill 读取蓝鲸流水线配置。

## 操作步骤

1. 打开流水线编辑页：
   ```bash
   curl -s "http://localhost:3456/new?url=https://devops.devcloud.ztgame.com/projects/xcard/pipelines/<PIPELINE_ID>/edit"
   ```

2. 切换到"流程配置"页签（默认打开）。

3. 读取 GIT 插件配置：
   - 插件 title 选择器：`.name-wrapper .name`
   - 仓库来源、分支、保存路径在插件展开后的表单中

4. 读取 Shell 插件 Content：
   - 选中 Shell 插件
   - 切换到"配置"页签（`.menus span` 第 2 个）
   - 由于 monaco editor 在后台 tab 不渲染，需用户在前台手动复制 Content
   - 替代方案：通过最近一次构建详情页的"配置"页签读取

5. 切换到"变量设置"页签读取环境变量。

6. 切换到"触发设置"页签读取手动/远程/定时触发配置。

## 输出格式

```json
{
  "pipeline_id": "p-xxx",
  "name": "配表检查-测试",
  "plugins": [
    {"type": "git", "repo": "xcard-excel-config", "branch": "", "path": "xcard-excel-config"},
    {"type": "git", "repo": "rain-qa-func", "branch": "main", "path": "rain-qa-func"},
    {"type": "shell", "content": "bash /root/ExcelChecker/rain-qa-func/docker/build.sh\n..."}
  ],
  "variables": [
    {"name": "CI_CLAUDE_FEISHU_ROBOT", "default": "..."}
  ],
  "triggers": ["manual", "remote", "scheduled"]
}
```
```

### Step 2: 创建 run-pipeline.md

- [ ] 在 `.claude/skills/qa-pipeline/references/run-pipeline.md` 中写入：

```markdown
# 触发流水线执行

使用 `web-access` skill 在蓝鲸流水线页面点击"启动新构建"。

## 操作步骤

1. 打开流水线主页：
   ```bash
   curl -s "http://localhost:3456/new?url=https://devops.devcloud.ztgame.com/projects/xcard/pipelines/<PIPELINE_ID>"
   ```

2. 点击"启动新构建"按钮：
   ```bash
   curl -s -X POST "http://localhost:3456/click?target=<TARGET_ID>" -d 'button.ant-btn-primary'
   ```

3. 修改变量值（如需）：
   - 变量输入框：`.ant-modal input.ant-input.w-full`
   - 需要触发 React input/change 事件才能生效

4. 点击对话框中的"执行"按钮：
   ```bash
   curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d "document.querySelector('.ant-modal button.ant-btn-primary').click()"
   ```

5. 确认 URL 中出现 `buildId=b-xxx` 表示启动成功。

## 参数

- `--pipeline=<PIPELINE_ID>`：流水线 ID
- `--branch=<BRANCH>`：可选，覆盖 GIT 插件分支
- `--skill=<SKILL>`：可选，设置启动变量 `CI_INVOKE_SKILL`
```

### Step 3: 创建 set-plugins.md

- [ ] 在 `.claude/skills/qa-pipeline/references/set-plugins.md` 中写入：

```markdown
# 配置 GIT/Shell 插件

## 配置 GIT 插件

1. 打开流水线编辑页。
2. 找到目标 GIT 插件或点击"+"添加新插件。
3. 设置：
   - 仓库来源：代码库
   - 仓库：选择已有代码库
   - 分支/TAG/COMMIT：留空（拉取默认分支）或指定分支
   - 代码保存路径：如 `xcard-frontend`、`xcard-backend`

## 配置 Shell 插件

1. 在编辑页添加/选中 Shell 插件。
2. 在 Content 中写入统一脚本：
   ```bash
   #!/usr/bin/env bash
   bash /root/ExcelChecker/rain-qa-func/docker/build.sh
   cp ~/.claude/settings.json.glm /root/ExcelChecker/rain-qa-func/docker/settings.json
   bash /root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh
   docker image prune -f
   docker container prune -f
   ```
3. 保存流水线。

## 注意事项

- 修改后保存，蓝鲸可能弹出到执行历史页（已知 bug），需重新进入流水线页面确认
- 不要在生产流水线直接测试，先在旧流水线验证
```

### Step 4: 修改 qa-pipeline/SKILL.md

- [ ] 在 `.claude/skills/qa-pipeline/SKILL.md` 的"常见操作"章节后增加"高级操作"章节：

```markdown
## 高级操作

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
```

### Step 5: 语法和链接检查

- [ ] 验证所有新创建文件存在：

```bash
ls -la .claude/skills/qa-pipeline/references/
```

Expected output: 包含 `show-config.md`、`run-pipeline.md`、`set-plugins.md`

- [ ] 验证 SKILL.md 中的相对链接正确

### Step 6: Commit

- [ ] 提交变更：

```bash
git add .claude/skills/qa-pipeline/
git commit -m "feat(qa-pipeline): 增强 skill，支持查看/触发/配置流水线

- 新增 references/show-config.md：查看流水线配置
- 新增 references/run-pipeline.md：触发流水线执行
- 新增 references/set-plugins.md：配置 GIT/Shell 插件
- SKILL.md 增加高级操作章节

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Task 6: 文档更新与回归测试

**Files:**
- Modify: `docker/CLAUDE.md`
- Modify: `docker/README.md`

### Step 1: 更新 docker/CLAUDE.md

- [ ] 在"构建流程"章节增加 `mjs-skills` 说明：

```markdown
### mjs-skills 集成

构建镜像时，`build.sh` 会自动从 `https://git.devcloud.ztgame.com/Xcards/mjs-skills.git` 拉取最新 skills，与 `docker/skills/` 合并后 COPY 进镜像。

```
build.sh
  → git clone/pull mjs-skills → docker/skills-external/
  → 校验 required-skills.txt 中的核心 skill
  → docker compose build
    → COPY docker/skills/ /home/analyzer/.claude/skills/
    → COPY docker/skills-external/ /home/analyzer/.claude/skills/
```

`mjs-skills` 中的同名 skill 会覆盖镜像内置版本。
```

- [ ] 在"常见构建问题"表格中增加一行：

| 错误 | 原因 | 修复 |
|------|------|------|
| `mjs-skills clone 失败` | 内网 git 不可用或仓库地址错误 | 检查 `MJS_SKILLS_URL` 和网络；如无法访问，确保 `docker/skills/` 中包含所需 skill |

### Step 2: 更新 docker/README.md

- [ ] 在"环境变量"表格中新增/更新：

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `CI_TARGET_REPO_PATH` | 是 | `/root/ExcelChecker/xcard-excel-config` | 要分析的目标仓库路径 |
| `CI_TARGET_REPO_TYPE` | 否 | `config` | 仓库类型：`config` / `frontend` / `backend` |
| `MJS_SKILLS_URL` | 否 | `https://git.devcloud.ztgame.com/Xcards/mjs-skills.git` | mjs-skills 仓库地址 |

- [ ] 在"CI/CD 集成"章节说明多流水线复用：

```markdown
### 多流水线复用

前端、后端、策划配表三条流水线可共用同一套 `docker/` 脚本，只通过蓝鲸环境变量区分：

| 流水线 | CI_TARGET_REPO_PATH | CI_INVOKE_SKILL | CI_TARGET_REPO_TYPE |
|--------|---------------------|-----------------|---------------------|
| 策划配表 | `/root/ExcelChecker/xcard-excel-config` | `analyze-prompt` | `config` |
| 前端代码 | `/root/ExcelChecker/xcard-frontend` | `analyze-prompt-frontend` | `frontend` |
| 后端代码 | `/root/ExcelChecker/xcard-backend` | `analyze-prompt-backend` | `backend` |
```

### Step 3: 执行回归测试

- [ ] **测试 4.1**：本地模拟非 `v0.0.8-pre-release` 分支提交

```bash
cd docker
docker compose run --rm analyzer
```

验证：终端输出中不出现 `[DEBUG-群消息]`，或 skill 输出中说明私聊发送。

- [ ] **测试 4.2**：本地模拟 `v0.0.8-pre-release` 普通提交

```bash
cd docker
CI_CLAUDE_FEISHU_ROBOT=db06f82a-4dad-43f6-bbef-97503e0b953a \
BK_CI_HOOK_BRANCH=v0.0.8-pre-release \
docker compose run --rm analyzer
```

验证：终端输出中出现 `[DEBUG-群消息]` 且内容包含 @提交者。

- [ ] **测试 4.6**：验证 mjs-skills 覆盖内置 skill

```bash
cd docker
docker run --rm rain-config-analyzer:latest \
  cat /home/analyzer/.claude/skills/analyze-prompt/SKILL.md | head -5
```

验证：内容与 `docker/skills-external/analyze-prompt/SKILL.md` 一致（如果存在）。

### Step 4: Commit

- [ ] 提交文档更新：

```bash
git add docker/CLAUDE.md docker/README.md
git commit -m "docs(docker): 更新构建流程和多流水线复用文档

- CLAUDE.md 增加 mjs-skills 集成说明和常见问题
- README.md 更新环境变量表格和多流水线复用配置

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review Checklist

- [ ] **Spec coverage**: 设计文档中所有章节（mjs-skills 集成、参数化、entrypoint 简化、qa-pipeline 增强、测试用例）都有对应任务
- [ ] **Placeholder scan**: 计划中没有 TBD/TODO/"实现 later"/"适当处理"等占位描述
- [ ] **Type consistency**: 环境变量名在所有任务中一致（`CI_TARGET_REPO_PATH`, `CI_TARGET_REPO_TYPE`, `CI_INVOKE_SKILL`）
- [ ] **File paths**: 所有文件路径使用绝对路径或从仓库根目录的相对路径
- [ ] **Commit messages**: 每个任务末尾都包含符合项目规范的 commit 命令

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-07-08-pipeline-refactor-plan.md`.

Two execution options:

**1. Subagent-Driven (recommended)** - Dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints for review.

Which approach?
