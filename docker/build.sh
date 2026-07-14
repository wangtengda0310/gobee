#!/bin/bash
# ============================================================
# 构建并推送配表分析 Docker 镜像
# 基于 docker compose build，只做 tag 和 push
# 使用方法: ./build.sh [registry] [tag] [--push]
# 示例: ./build.sh docker.1ms.run v1.0.0 --push
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REGISTRY="${1:-docker.1ms.run}"
TAG="${2:-latest}"
PUSH_FLAG="${3:-}"
IMAGE_NAME="rain-config-analyzer"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${TAG}"

# 默认 Claude Code 版本（与 Dockerfile 默认值一致）
CLAUDE_CODE_VERSION="${CLAUDE_CODE_VERSION:-2.1.141}"

# mjs-skills 源目录：由蓝鲸 GIT 插件拉取（流水线环境）或本地 simulate.sh 准备。
# B1 架构：git 操作交给蓝鲸 GIT 插件（自带 git，不受宿主机 git 1.8.3.1 限制），
# build.sh 只负责把已拉取的 mjs-skills 搬运到 docker context 内供 Dockerfile COPY。
MJS_SKILLS_SOURCE="${MJS_SKILLS_SOURCE:-/root/ExcelChecker/mjs-skills}"
MJS_SKILLS_DIR="$SCRIPT_DIR/skills-external"

# 构建缓存参数（默认启用）
BUILD_CACHE="${BUILD_CACHE:-1}"

# 解析可选的 --push 参数（可能在任意位置）
for arg in "$@"; do
    if [ "$arg" = "--push" ]; then
        PUSH_FLAG="--push"
    fi
done

# 搬运 mjs-skills 到 docker context 内（供 Dockerfile COPY docker/skills-external/）
# 流水线：蓝鲸 GIT 插件拉到 $MJS_SKILLS_SOURCE（默认 /root/ExcelChecker/mjs-skills）
# 本地：   simulate.sh 设 MJS_SKILLS_SOURCE 指向本地 clone，或手动 export
if [ ! -d "$MJS_SKILLS_SOURCE" ]; then
    echo "错误: mjs-skills 源目录不存在: $MJS_SKILLS_SOURCE"
    echo "  流水线环境: 确认已添加 mjs-skills 的 GIT 插件（localPath=mjs-skills）"
    echo "  本地环境:   用 docker/local-ci-simulate.sh，或手动 clone 后 export MJS_SKILLS_SOURCE=<dir>"
    exit 1
fi
echo "搬运 mjs-skills: $MJS_SKILLS_SOURCE → $MJS_SKILLS_DIR"
rm -rf "$MJS_SKILLS_DIR"
cp -r "$MJS_SKILLS_SOURCE" "$MJS_SKILLS_DIR"
# cp 会带上 mjs-skills 的 .git 目录，镜像不需要，删掉减小 context
rm -rf "$MJS_SKILLS_DIR/.git"

echo "========================================="
echo "构建配表分析镜像"
echo "  镜像: ${FULL_IMAGE}"
echo "  Claude Code 版本: ${CLAUDE_CODE_VERSION}"
echo "  构建缓存: $([ "$BUILD_CACHE" = "1" ] && echo '启用' || echo '禁用')"
echo "  自动推送: $([ "$PUSH_FLAG" = "--push" ] && echo '是' || echo '否')"
echo "========================================="

# 登录检查：如果指定了 --push，先检查是否已登录到目标仓库
if [ "$PUSH_FLAG" = "--push" ]; then
    if ! docker info --format '{{.IndexServerAddress}}' > /dev/null 2>&1; then
        echo "错误: Docker 守护进程未运行"
        exit 1
    fi

    # 检查目标 registry 的登录状态
    REGISTRY_HOST="${REGISTRY%%/*}"
    if [ "$REGISTRY_HOST" != "docker.io" ] && [ "$REGISTRY_HOST" != "" ]; then
        if ! grep -q "\"$REGISTRY_HOST\"" ~/.docker/config.json 2>/dev/null; then
            echo "警告: 未检测到 ${REGISTRY_HOST} 的登录凭证"
            echo "请先执行: docker login ${REGISTRY_HOST}"
            exit 1
        fi
    fi
fi

# 构建 docker compose build 参数
BUILD_ARGS=""
BUILD_ARGS="${BUILD_ARGS} --build-arg CLAUDE_CODE_VERSION=${CLAUDE_CODE_VERSION}"

# 缓存控制
if [ "$BUILD_CACHE" = "0" ]; then
    BUILD_ARGS="${BUILD_ARGS} --no-cache"
fi

# rain-excel-checker 二进制由开发机交叉编译后提交 git，Dockerfile 直接 COPY。
# 不在 build.sh / 流水线宿主机现场编译。详见 CLAUDE.md「二进制构建方式」。

# 校验 required skills 不缺失
if [ ! -f "$SCRIPT_DIR/required-skills.txt" ]; then
    echo "错误: 核心 skill 清单文件不存在: $SCRIPT_DIR/required-skills.txt"
    exit 1
fi

echo "校验核心 skill..."
MISSING_SKILLS=""
while IFS= read -r skill_name || [ -n "$skill_name" ]; do
    # 去除首尾空白，避免行尾空格导致误报缺失
    skill_name=$(echo "$skill_name" | xargs)
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

# 使用 docker compose 构建（复用 docker-compose.yml 配置）
cd "$SCRIPT_DIR"
# shellcheck disable=SC2086
docker compose build ${BUILD_ARGS}

# 打标签（docker-compose.yml 中定义的镜像名为 rain-config-analyzer:latest）
docker tag "${IMAGE_NAME}:latest" "${FULL_IMAGE}"

echo ""
echo "========================================="
echo "构建完成: ${FULL_IMAGE}"
echo "========================================="
echo ""

# 自动推送（如果指定了 --push）
if [ "$PUSH_FLAG" = "--push" ]; then
    echo "推送到远程仓库..."
    docker push "${FULL_IMAGE}"
    echo "推送完成: ${FULL_IMAGE}"
fi
