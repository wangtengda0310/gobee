# docker/ — CI/CD 分析流水线 Docker 镜像（通用基础设施）

> **文档归属约定（重要）**：本目录的脚本/skill 分两层——
> - **基础设施层**（`ci_pipeline.sh`、`build.sh`、`Dockerfile`、`qa-pipeline` skill 的通用 API/定位器）：通用，不绑定具体仓库类型，前端/后端/策划配表流水线复用。**不放项目特定值**（具体仓库 codeRepoId、分支、个人邮箱、内网域名等）。
> - **项目特定层**：具体仓库的配置（如 mjs-skills 的 codeRepoId=463、master 分支、`/root/ExcelChecker/xcard-excel-config` 路径）属于项目知识，记在本文档的「项目特定配置」章节，由各流水线以环境变量传入（如 `CI_TARGET_REPO_PATH`），**不硬编码进通用脚本**。
>
> 判定：换个项目/仓库这段还成立吗？成立→基础设施层；不成立→项目特定层。
> 这条规则跨机器/跨工具（Cursor 等）都生效——因为它在 git 管理的项目文档里，不依赖 `~/.claude/` 记忆。

流水线相关的任务
1 流水线通过docker启动claude执行ai分析，通过透传CI_INVOKE_SKILL控制claude调用技能
2 流水线中执行对策划配表校验以及git commit message的检查，见skills/analyze-prompt
  - 执行rain-qa-func校验策划配表
  - 校验区分普通git提交和merge，根据分支名是否为v0.0.8-pre-release区分是否发送飞书群消息或私聊
  - 通过analyze-prompt分析commit message是否与实际提交内容匹配，并发送飞书私聊或群消息
3 策划表仓库提交会触发流水线的运行用来对新提交的配置进行校验，每天还有一个定时任务对策划配表仓库进行全量校验

面向开发者的构建规范，使用者文档见 [README.md](README.md)。

## 服务器联网条件

流水线服务器（`host-10-254-114-174`）**无法直连 Docker Hub**（`registry-1.docker.io` 超时），只能通过国内镜像代理 `docker.1ms.run` 拉取基础镜像。

验证方式：
```bash
# 直连 Docker Hub — 会超时失败
docker pull node:22-alpine

# 通过代理 — 正常
docker pull docker.1ms.run/node:22-alpine
```

### 可用镜像代理

| 代理      | 地址               | 说明      |
|---------|------------------|---------|
| 1ms.run | `docker.1ms.run` | 当前使用，免费 |

如代理不可用，需在能访问 Docker Hub 的环境预先 `docker pull` 再 `docker save/load` 传输。

## 镜像命名规范

### 规则

1. **Dockerfile 中基础镜像必须带代理前缀**：`docker.1ms.run/<镜像名>:<标签>`
2. **不带 `/library/` 路径前缀**：`docker.1ms.run/node` 而非 `docker.1ms.run/library/node`（`/library/` 会导致 `not found`）
3. **构建产物镜像由 `build.sh` 统一打标签推送**
4. **统一使用 alpine 版本基础镜像**：减小镜像体积，包管理统一用 `apk`（不用 `apt-get`）

### 正确示例

```dockerfile
# ✅ 正确
FROM docker.1ms.run/node:22-alpine
FROM docker.1ms.run/golang:1.26.0-alpine

# ❌ 错误：多了 /library/ 前缀
FROM docker.1ms.run/library/node:22-alpine

# ❌ 错误：服务器无法直连 Docker Hub
FROM node:22-alpine
```

### 当前基础镜像

| 用途                   | 镜像                                    | 版本选择                       |
|----------------------|---------------------------------------|----------------------------|
| 运行时（Claude Code CLI） | `docker.1ms.run/node:XX-alpine`       | Claude Code 要求 Node.js 18+ |
| Go 编译（宿主机交叉编译）       | `docker.1ms.run/golang:1.26.0-alpine` | 非 Docker 内使用               |

### npm/pip 国内镜像

Dockerfile 中已配置国内镜像源，无需修改：

```bash
# npm — 淘宝镜像
npm config set registry https://registry.npmmirror.com

# pip — 清华镜像
pip3 install --index-url https://pypi.tuna.tsinghua.edu.cn/simple
```

## 二进制构建方式（开发机编译 + 提交 git）

`docker/rain-excel-checker` 是 **Linux amd64 预编译二进制，提交在 git 里**。Dockerfile `COPY` 进镜像，流水线宿主机**不需要 go 工具链**，`build.sh` 也不做任何编译。

### 开发机编译流程

改 Go 代码后，推荐用 Taskfile 编译（权威编译入口，`build/Taskfile.yml` 的 `build:rain-excel-checker` task）：

```bash
wails3 task build:rain-excel-checker
git add docker/rain-excel-checker
```

等价的手动交叉编译命令（在仓库根执行）：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o docker/rain-excel-checker -ldflags="-w -s" ./cmd/rain-excel-checker/
git add docker/rain-excel-checker
```

> ⚠️ Windows 上 go build 产出的二进制**无 execute bit**（NTFS 不支持 Unix 权限），COPY 到 Linux 容器后权限为 `-rw-r--r--`。Dockerfile 已用 `RUN chmod +x /usr/local/bin/rain-excel-checker` 处理，**不要删除这行 chmod**。

### 二进制位置与路径约束（维护必读）

> 沉淀 2026-07-09 路径审查结论，防止后续维护把二进制挪错位置、或把编译路径改回已删除的旧目录。

**1. `docker/rain-excel-checker` 必须留在 `docker/` 根目录，禁止移入 `docker/skills/` 任何技能子目录**（包括唯一运行时调用方 `analyze-prompt/`）。理由：

- **职责不同** — `docker/skills/` 是 Claude 扫描的 `SKILL.md` 文档区（Dockerfile COPY 到 `/home/analyzer/.claude/skills/`）；二进制是系统级工具，COPY 到 `/usr/local/bin/` + `chmod +x`。两套路径、两套权限模型，不能混。
- **多技能共享** — `analyze-prompt` 是唯一运行时调用方（`/usr/local/bin/rain-excel-checker ...`），但 `game-config-relations` 也基于其校验规则构建知识库。放进某技能私有目录会制造虚假耦合。
- **Dockerfile COPY 顺序** — `docker/skills/`（内置）先、`docker/skills-external/`（mjs-skills）后，靠顺序实现覆盖语义；二进制混入 skills 目录会污染此语义。

**2. 源码与编译入口（权威路径，曾因重构漂移）**：

| 项       | 权威路径                                                   | 说明                                                    |
|---------|--------------------------------------------------------|-------------------------------------------------------|
| main 入口 | `cmd/rain-excel-checker/main.go`                       | 完整 main 包，import `backend/pkg/rain-excel-checker/...` |
| 库代码     | `backend/pkg/rain-excel-checker/`                      | 被多个 backend 子包依赖                                      |
| 编译入口    | `build/Taskfile.yml` 的 `build:rain-excel-checker` task | `wails3 task build:rain-excel-checker`                |
| 产物      | `docker/rain-excel-checker`（Linux amd64，无后缀）           | 提交 git                                                |

> ⚠️ 项目根的旧 `rain-excel-checker/` 目录已于 `e11645d` 重构删除。2026-07-09 审查发现 `build/Taskfile.yml`（编译路径）和 `.gitignore`（exe 忽略条目）仍引用旧路径，已修复。**编译路径必须是 `./cmd/rain-excel-checker/`，切勿回退为 `./rain-excel-checker/`。**

### 为什么不用多阶段 Docker 构建（曾尝试，未采用）

历史上 36e1080 想加 `golang` builder 阶段做多阶段构建，最终没采用，两个原因：

1. **docker.1ms.run 不支持 `/library/` 前缀**——当时的 Dockerfile 用 `docker.1ms.run/library/golang:...`，builder 镜像拉取失败。此问题 37c5ae6 已修复（去掉 `/library/`），现在 `docker.1ms.run/golang:1.26-alpine` 可用。
2. **rain-robot 外部 module 依赖**（关键障碍）——rain-qa-func 的 `go.mod` require 了 `git.devcloud.ztgame.com/v-tangfangda/rain-robot`（私有 module，模块名不规范需 replace）。多阶段 builder 的 `go mod download` 会拉 rain-robot，要求**流水线服务器能访问内网 git**（`git.devcloud.ztgame.com`），引入网络/认证依赖。

因此采用「开发机编译提交 git」——开发机有完整依赖环境（go.work/replace 已配好），产物提交 git，流水线零依赖。

### 为什么不在流水线宿主机现场编译（历史教训）

e11645d4 曾在 build.sh 加「宿主机现场 `go build`」预编译段，导致流水线报 `go.mod file not found in current directory`：

- 蓝鲸把 Shell 脚本生成到 `/tmp/devops_script*.sh` 执行，build.sh 继承 **PWD=/tmp**
- `go build` 从 **PWD**（非包路径参数）往上找 go.mod，`/tmp→/` 都没有 → 报错
- 改 `SCRIPT_DIR`/`$0`/`BASH_SOURCE` 都没用——它本来就解析正确，问题在 PWD

此方案已废弃。**维护提醒：不要在 build.sh 里加任何 `go build`；二进制只由开发机编译提交**。

## 构建流程

### Docker 构建

```
build.sh（宿主机）
  → docker compose build（调用 Dockerfile）
    → FROM docker.1ms.run/node:XX-alpine
    → 安装依赖（apk + pip + npm）
    → 复制预编译二进制 + skills + 配置
  → docker tag → docker push（可选）

ci_pipeline.sh（流水线宿主机）
  → 前置检查（镜像、配置、仓库）
  → docker run（挂载仓库 + 配置）
    → entrypoint.sh（容器内）
      → rain-excel-checker → claude -p → 飞书通知
```

### mjs-skills 集成（B1 架构：蓝鲸 GIT 插件拉取 + build.sh 搬运）

mjs-skills 的 **git 操作交给蓝鲸 GIT 插件**（自带 git 运行时，不受流水线宿主机 git 1.8.3.1 限制），`build.sh` 只负责把已拉取的目录搬运到 docker context 内供 Dockerfile COPY。

```
蓝鲸流水线（GIT 插件拉取）
  → mjs-skills 拉到 /root/ExcelChecker/mjs-skills（localPath=mjs-skills）

build.sh（Shell 插件）
  → cp -r $MJS_SKILLS_SOURCE(默认 /root/ExcelChecker/mjs-skills) → docker/skills-external/
  → 校验 required-skills.txt
  → docker compose build
    → COPY docker/skills/          → /skills/（内置）
    → COPY docker/skills-external/ → /skills/（mjs-skills 覆盖同名）
```

- `MJS_SKILLS_SOURCE` 环境变量指定 mjs-skills 源目录（流水线默认 `/root/ExcelChecker/mjs-skills`，本地模拟脚本会 export 指向本地 clone）
- `build.sh` 不再做任何 `git clone/pull`（历史教训：build.sh 用 `git -C` 在宿主机 git 1.8.3.1 报 `Unknown option: -C` → 改为蓝鲸插件拉取）
- `docker/skills-external/` 不提交 git，由 build.sh 搬运生成（`.gitignore` 排除）

> **实测（2026-07-09）**：当前 mjs-skills 内容（Excel-check、card-skin-config、configure-* 等策划配表 skill）与内置 skills（analyze-prompt 等）**互补**——mjs-skills 未包含 analyze-prompt，合并后无覆盖冲突。COPY 顺序（内置先、mjs-skills 后）保证未来同名时 mjs-skills 优先。

### mjs-skills 蓝鲸配置（项目特定，2026-07-09 实测）

| 项                            | 值                                                                              |
|------------------------------|--------------------------------------------------------------------------------|
| 仓库                           | `https://git.devcloud.ztgame.com/Xcards/mjs-skills.git`                        |
| 默认主分支                        | **master**（不是 main）                                                            |
| 蓝鲸代码库 scm id（= `codeRepoId`） | **463**                                                                        |
| 关联凭据                         | `buildpipeline`（已关联，scm 列表可查）                                                  |
| 流水线 GIT 插件配置                 | `codeRepoId=463, refName=master, localPath=mjs-skills, strategy=REVERT_UPDATE` |

> 代码库有两套 API（数据源不同）：`/ms/scm/api/user/v1/repository/xcard`（scm，流水线 GIT 插件用这套，17 个）vs `/ms/repository/api/user/repositories/xcard`（凭据视角，13 个）。查 codeRepoId 用 scm 那套。

> 关联新代码库的通用流程见 `qa-pipeline` skill 的 [register-repo.md](../.claude/skills/qa-pipeline/references/register-repo.md)（GitLab 强制 OAuth 凭据、后台 tab 受限需前台手动）。本仓库已关联完，无需重复。

## ⚠️ 流水线调整后必须同步本地测试脚本

**每次调整蓝鲸流水线（增删 plugin、改 Shell Content、改变量、改分支）后，必须同步检查/更新 `docker/local-ci-simulate.sh`**，确保本地模拟脚本的 plugin 顺序和行为与蓝鲸一致。

同步检查点：
- 蓝鲸新增 GIT 插件 → simulate.sh 加对应 clone 步骤
- Shell Content 改动（如 settings.json 文件名、新增命令）→ simulate.sh 的 Plugin 4 同步
- 变量/分支改动 → simulate.sh 默认值同步
- plugin 执行顺序调整 → simulate.sh 步骤顺序同步

> 原因：simulate.sh 的价值在于本地复现蓝鲸行为，一旦不同步就失去意义，还会误导排查。

> **实测（2026-07-09）**：当前 mjs-skills 内容（Excel-check、card-skin-config、configure-* 等策划配表 skill）与内置 skills（analyze-prompt、check-config-commit-message 等）**互补**——mjs-skills 未包含 analyze-prompt，合并后无覆盖冲突，容器拥有两者完整 skill 集。COPY 顺序（内置先、mjs-skills 后）保证未来同名时 mjs-skills 优先。

## 实测经验与陷阱（2026-07-09 流水线改造验证）

### skills-external 不提交 git（不要加 .gitkeep）

`docker/skills-external/` 是 build.sh 运行时 clone 的产物，**不提交 git**（`.gitignore` 排除）。

⚠️ **不要为「Dockerfile COPY 空目录失败」加 `.gitkeep` 占位**——这是伪防御：
- 正常流程 build.sh 先 clone 创建目录，COPY 永远不会遇到空目录
- build.sh clone 失败时 `set -e` 退出，根本到不了 docker build
- `.gitkeep` 让目录非空，反而导致 build.sh 的 `git clone <url> <dir>` 报 `destination path already exists and is not an empty directory`
- 唯一「触发 COPY 空目录失败」的场景是「跳过 build.sh 直接 docker build」，这本身就是错误用法（mjs-skills 没进去），应明确报错而非静默产出残缺镜像

### 本地 Docker 测试（simulate.sh）

`docker/local-ci-simulate.sh` 按蓝鲸 plugin 顺序在本地完整模拟流水线（拉 config/mjs-skills/rain-qa-func → build.sh → ci_pipeline.sh），无需触发真实流水线即可验证改动。

```bash
# 一键模拟（命名参数）
./local-ci-simulate.sh --target-repo https://git.devcloud.ztgame.com/v-tangfangda/rain-qa-func.git \
                       --skill-repo  https://git.devcloud.ztgame.com/Xcards/mjs-skills.git

# 位置参数简写
./local-ci-simulate.sh <rain-qa-func-url> <mjs-skills-url>

# 只验证镜像构建（不跑 ci_pipeline.sh）
./local-ci-simulate.sh --target-repo <url> --skill-repo <url> --build-only

# 跳过 config 拉取（用本地已有）
./local-ci-simulate.sh --target-repo <url> --skill-repo <url> --skip-config
```

> **环境提示**：mjs-skills 需内网 GitLab 认证。WSL 的 git 无 credential，可用 Windows git bash 跑 simulate.sh（有 credential），或先手动 clone 后用 `--skill-repo <本地路径>`（脚本支持本地路径，clone_or_pull 会识别）。

> `docker compose build` 本身不执行 git 命令，build.sh 现在只 `cp`（不 clone），所以 build 阶段不受任何 git 版本/认证影响。

### 流水线变量设置：按 label 定位，不要按 idx

蓝鲸启动对话框的变量由「disabled input（变量名）+ editable input（值）」成对组成，显示顺序随「显示选项」配置变化。**不要假设 `inputs[0]` 是固定变量**——按 disabled label 的 value 定位。且**不要在同一次 eval 里连续 setVal 多个变量**（React 批处理会丢失前几个，实测只最后一个生效）。详见 `qa-pipeline` skill 的 [references/cdp-operations.md](../.claude/skills/qa-pipeline/references/cdp-operations.md)。

### Shell Content 中 settings.json 文件名随模型变化

流水线 Shell 插件 Content 里 `cp ~/.claude/settings.json.<model> .../settings.json` 的 `<model>` 后缀随当前使用的模型变化（曾用 `.glm`，现为 `.kimi`）。源文件由流水线宿主机 `~/.claude/` 提供，修改模型时需同步更新 Shell Content 和宿主机文件。

## 常见构建问题

| 错误                                                              | 原因                                                                             | 修复                                                                                   |
|-----------------------------------------------------------------|--------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| `go.mod file not found in current directory`（历史）                | e11645d4 曾让 build.sh 在宿主机现场编译，PWD=/tmp 导致 go 找不到 go.mod                        | 已改用「开发机编译提交 git」，build.sh 不再编译（见「二进制构建方式」）                                           |
| `not found` 拉基础镜像失败                                             | Dockerfile 用了 `/library/` 前缀                                                   | 去掉 `/library/`，用 `docker.1ms.run/node`                                               |
| `request canceled` 超时                                           | 服务器无法直连 Docker Hub                                                             | 所有 `FROM` 必须带 `docker.1ms.run` 前缀                                                    |
| `apk add` 包找不到                                                  | alpine 版本过旧，软件源过期                                                              | 更新基础镜像版本                                                                             |
| `mjs-skills clone 失败`                                           | 内网 git 不可用或仓库地址错误                                                              | 检查 `MJS_SKILLS_URL` 和网络；如无法访问，确保 `docker/skills/` 中包含所需 skill                        |
| `destination path already exists and is not an empty directory` | `skills-external/` 目录非空（如误加 `.gitkeep`）导致 `git clone` 失败                       | 不要给 skills-external 加 `.gitkeep`；B1 架构下 build.sh 不再 clone，改为 cp 搬运（见「mjs-skills 集成」） |
| `Unknown option: -C`（历史）                                        | build.sh 曾用 `git -C pull` 更新 mjs-skills，流水线宿主机 git 1.8.3.1 不支持 `-C`（1.8.5+ 才有） | 已改用 B1 架构：git 操作交给蓝鲸 GIT 插件，build.sh 只 cp 搬运，不再调用 git                                |
