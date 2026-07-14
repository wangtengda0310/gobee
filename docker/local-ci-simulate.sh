#!/usr/bin/env bash
# ============================================================
# 本地模拟蓝鲸流水线（按 plugin 顺序执行）
#
# 用途：在本地完整复现蓝鲸「配表检查-测试」流水线的执行流程，
#       便于在不触发真实流水线的情况下测试 build.sh / ci_pipeline.sh / Dockerfile 改动。
#
# 对应蓝鲸 plugin（执行顺序一致）：
#   [Plugin 1] config GIT 插件     → clone config → $WORKSPACE/xcard-excel-config
#   [Plugin 2] mjs-skills GIT 插件 → clone mjs-skills → $WORKSPACE/mjs-skills   (B1 架构新增)
#   [Plugin 3] rain-qa-func GIT 插件 → clone rain-qa-func → $WORKSPACE/rain-qa-func
#   [Plugin 4] claude Shell 插件   → cp settings.json + build.sh + ci_pipeline.sh + docker prune
#
# 用法：
#   # 命名参数（推荐）
#   local-ci-simulate.sh --target-repo <rain-qa-func-url> --skill-repo <mjs-skills-url>
#
#   # 位置参数简写（前两个位置参数 = target-repo, skill-repo）
#   local-ci-simulate.sh https://git.devcloud.ztgame.com/v-tangfangda/rain-qa-func.git \
#                        https://git.devcloud.ztgame.com/Xcards/mjs-skills.git
#
#   # 跳过 config 拉取（只测 build/run，用本地已有 config）
#   local-ci-simulate.sh --target-repo <url> --skill-repo <url> --skip-config
# ============================================================

set -euo pipefail

# ---------------- 默认值（可被环境变量/参数覆盖）----------------
TARGET_REPO=""
SKILL_REPO=""
CONFIG_REPO="${CONFIG_REPO:-https://git.devcloud.ztgame.com/Xcards/config.git}"
TARGET_BRANCH="${TARGET_BRANCH:-worktree-docker}"
SKILL_BRANCH="${SKILL_BRANCH:-master}"
CONFIG_BRANCH="${CONFIG_BRANCH:-v0.0.8-pre-release}"
WORKSPACE="${WORKSPACE:-/tmp/ExcelChecker-sim}"
SETTINGS_SRC="${SETTINGS_SRC:-$HOME/.claude/settings.json.kimi}"
SKIP_CONFIG="${SKIP_CONFIG:-0}"
# 到 ci_pipeline.sh 的运行步骤（默认执行；=0 时只到 build.sh，便于快速验证镜像构建）
RUN_PIPELINE="${RUN_PIPELINE:-1}"

usage() {
    cat <<EOF
本地模拟蓝鲸流水线（对应各 plugin 顺序）

用法:
  local-ci-simulate.sh --target-repo <url> --skill-repo <url> [选项]
  local-ci-simulate.sh <target-repo> <skill-repo> [选项]   # 位置参数简写

必需参数:
  --target-repo <url>    rain-qa-func 仓库 URL（对应蓝鲸 GIT 插件「（GIT）rain-qa-func」）
  --skill-repo  <url>    mjs-skills 仓库 URL（对应蓝鲸 GIT 插件「mjs-skills」，B1 新增）

可选参数:
  --config-repo <url>     config 仓库 URL（默认 Xcards/config）
  --target-branch <name>  rain-qa-func 分支（默认 worktree-docker）
  --skill-branch <name>   mjs-skills 分支（默认 master）
  --config-branch <name>  config 分支（默认 v0.0.8-pre-release）
  --workspace <path>      工作空间（默认 /tmp/ExcelChecker-sim）
  --settings <path>       settings.json 源文件（默认 ~/.claude/settings.json.kimi）
  --skip-config           跳过 Plugin 1（config 拉取），仅用已有或只测 build/run
  --build-only            只跑到 build.sh，不执行 ci_pipeline.sh（快速验证镜像构建）
  -h, --help              显示本帮助

环境变量（优先级低于命令行参数）:
  CONFIG_REPO / TARGET_BRANCH / SKILL_BRANCH / CONFIG_BRANCH
  WORKSPACE / SETTINGS_SRC / SKIP_CONFIG / RUN_PIPELINE

对应蓝鲸 plugin:
  [Plugin 1] config       → \$WORKSPACE/xcard-excel-config
  [Plugin 2] mjs-skills   → \$WORKSPACE/mjs-skills
  [Plugin 3] rain-qa-func → \$WORKSPACE/rain-qa-func
  [Plugin 4] Shell        → cp settings.json + build.sh(MJS_SKILLS_SOURCE) + ci_pipeline.sh
EOF
}

# ---------------- 参数解析（支持 --xxx 命名 + 位置参数）----------------
POSITIONAL=()
while [ $# -gt 0 ]; do
    case "$1" in
        --target-repo)    TARGET_REPO="$2"; shift 2 ;;
        --skill-repo)     SKILL_REPO="$2"; shift 2 ;;
        --config-repo)    CONFIG_REPO="$2"; shift 2 ;;
        --target-branch)  TARGET_BRANCH="$2"; shift 2 ;;
        --skill-branch)   SKILL_BRANCH="$2"; shift 2 ;;
        --config-branch)  CONFIG_BRANCH="$2"; shift 2 ;;
        --workspace)      WORKSPACE="$2"; shift 2 ;;
        --settings)       SETTINGS_SRC="$2"; shift 2 ;;
        --skip-config)    SKIP_CONFIG=1; shift ;;
        --build-only)     RUN_PIPELINE=0; shift ;;
        -h|--help)        usage; exit 0 ;;
        *)                POSITIONAL+=("$1"); shift ;;
    esac
done
# 位置参数兜底：前两个分别赋给 target-repo / skill-repo
if [ -z "$TARGET_REPO" ] && [ "${#POSITIONAL[@]}" -ge 1 ]; then TARGET_REPO="${POSITIONAL[0]}"; fi
if [ -z "$SKILL_REPO" ] && [ "${#POSITIONAL[@]}" -ge 2 ]; then SKILL_REPO="${POSITIONAL[1]}"; fi

if [ -z "$TARGET_REPO" ] || [ -z "$SKILL_REPO" ]; then
    echo "错误: 必须指定 --target-repo 和 --skill-repo（或两个位置参数）" >&2
    echo ""
    usage
    exit 1
fi

# ---------------- 工具函数 ----------------
# 模拟蓝鲸 GIT 插件的拉取（增量优先，首次克隆）
clone_or_pull() {
    local url="$1" dir="$2" branch="$3"
    if [ -d "$dir/.git" ]; then
        echo "  更新: $dir"
        # 用 (cd && git) 形式，兼容低版本 git（不依赖 git -C）
        (cd "$dir" && git fetch --all && git checkout "$branch" && git pull origin "$branch") || true
    else
        echo "  克隆: $url → $dir"
        git clone -b "$branch" "$url" "$dir"
    fi
}

echo "========================================="
echo "本地模拟蓝鲸流水线"
echo "  workspace:      $WORKSPACE"
echo "  target-repo:    $TARGET_REPO@$TARGET_BRANCH"
echo "  skill-repo:     $SKILL_REPO@$SKILL_BRANCH"
echo "  config-repo:    $CONFIG_REPO@$CONFIG_BRANCH (skip=$SKIP_CONFIG)"
echo "  settings 源:    $SETTINGS_SRC"
echo "  执行 ci_pipeline: $RUN_PIPELINE"
echo "========================================="

mkdir -p "$WORKSPACE"

# ---------------- [Plugin 1] config GIT 插件 ----------------
# 对应蓝鲸「config commit」「config schedule」插件，拉取策划配表到 xcard-excel-config
if [ "$SKIP_CONFIG" != "1" ]; then
    echo ""
    echo "[Plugin 1] 拉取 config → $WORKSPACE/xcard-excel-config"
    clone_or_pull "$CONFIG_REPO" "$WORKSPACE/xcard-excel-config" "$CONFIG_BRANCH"
else
    echo ""
    echo "[Plugin 1] 跳过 config 拉取（--skip-config）"
fi

# ---------------- [Plugin 2] mjs-skills GIT 插件（B1 新增）----------------
# 对应蓝鲸「mjs-skills」GIT 插件，build.sh 会从这里搬运到 docker/skills-external/
echo ""
echo "[Plugin 2] 拉取 mjs-skills → $WORKSPACE/mjs-skills"
clone_or_pull "$SKILL_REPO" "$WORKSPACE/mjs-skills" "$SKILL_BRANCH"

# ---------------- [Plugin 3] rain-qa-func GIT 插件 ----------------
# 对应蓝鲸「（GIT）rain-qa-func」插件，拉取本工具仓库
echo ""
echo "[Plugin 3] 拉取 rain-qa-func → $WORKSPACE/rain-qa-func"
clone_or_pull "$TARGET_REPO" "$WORKSPACE/rain-qa-func" "$TARGET_BRANCH"

# ---------------- [Plugin 4] claude Shell 插件 ----------------
# 对应蓝鲸「claude」Shell 插件 Content：
#   bash build.sh && cp settings.json && bash ci_pipeline.sh && docker prune
echo ""
echo "[Plugin 4] Shell: cp settings + build.sh + ci_pipeline.sh"
DOCKER_DIR="$WORKSPACE/rain-qa-func/docker"

# 模拟 Shell Content: cp ~/.claude/settings.json.<model> docker/settings.json
if [ -f "$SETTINGS_SRC" ]; then
    cp "$SETTINGS_SRC" "$DOCKER_DIR/settings.json"
    echo "  cp settings.json: $SETTINGS_SRC → docker/settings.json"
else
    echo "  ⚠ 跳过 settings.json（源不存在: $SETTINGS_SRC）"
    echo "    ci_pipeline.sh 需要它，若缺失容器内 Claude 会启动失败"
fi

# 模拟 Shell Content: bash build.sh
# build.sh 通过 MJS_SKILLS_SOURCE 找到 Plugin 2 拉的 mjs-skills
export MJS_SKILLS_SOURCE="$WORKSPACE/mjs-skills"
echo "  bash build.sh (MJS_SKILLS_SOURCE=$MJS_SKILLS_SOURCE)"
(cd "$DOCKER_DIR" && bash build.sh)

# 模拟 Shell Content: bash ci_pipeline.sh
if [ "$RUN_PIPELINE" = "1" ]; then
    export CI_TARGET_REPO_PATH="$WORKSPACE/xcard-excel-config"
    export CI_INVOKE_SKILL="${CI_INVOKE_SKILL:-analyze-prompt}"
    export CI_TARGET_REPO_TYPE="${CI_TARGET_REPO_TYPE:-config}"
    echo "  bash ci_pipeline.sh (CI_TARGET_REPO_PATH=$CI_TARGET_REPO_PATH, SKILL=$CI_INVOKE_SKILL)"
    (cd "$DOCKER_DIR" && bash ci_pipeline.sh)
else
    echo "  跳过 ci_pipeline.sh（--build-only）"
fi

echo ""
echo "========================================="
echo "模拟完成"
echo "  workspace: $WORKSPACE"
echo "  镜像:      rain-config-analyzer:latest"
echo "========================================="
