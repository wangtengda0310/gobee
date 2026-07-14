# ============================================================
# 配表影响分析 Docker 镜像（精简版）
# Node.js 运行时 + Claude Code CLI + rain-excel-checker
#
# 优化策略：
#   1. 使用 alpine 基础镜像（体积更小）
#   2. 合并 RUN 层减少层数
#   3. 清理构建缓存减少镜像体积
#   4. 宿主机构建 rain-excel-checker，复制二进制进容器
# ============================================================

# --- 运行镜像（alpine 精简版）---
FROM docker.1ms.run/node:22-alpine

LABEL maintainer="QA Team"
LABEL description="配表影响分析 Docker 镜像 - Node.js + Claude Code CLI + rain-excel-checker"
LABEL version="1.3"

# 合并安装所有依赖，清理缓存
RUN apk add --no-cache \
        git \
        bash \
        ca-certificates \
        jq \
        curl \
        python3 \
        py3-pip \
        file \
    && apk cache clean

# 安装 Python 包（使用阿里云镜像）
RUN pip3 install --no-cache-dir \
    --break-system-packages \
    --index-url https://pypi.tuna.tsinghua.edu.cn/simple \
    openpyxl

# 全局安装 npm 包（使用淘宝镜像）
ARG CLAUDE_CODE_VERSION="2.1.141"
RUN npm config set registry https://registry.npmmirror.com && \
    npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION} xlsx && \
    npm cache clean --force

# 安装 multica CLI（从预置二进制安装，避免流水线网络问题）
# 预置文件来源：github.com/multica-ai/multica/releases
# 如需更新版本，在能访问 GitHub 的环境中执行：
#   curl -fsSL "https://github.com/multica-ai/multica/releases/download/v${VERSION}/multica-cli-${VERSION}-linux-amd64.tar.gz" \
#     -o docker/prebuilt/multica-cli-${VERSION}-linux-amd64.tar.gz
ARG MULTICA_VERSION="0.3.10"
COPY docker/prebuilt/multica-cli-${MULTICA_VERSION}-linux-amd64.tar.gz /tmp/multica.tar.gz
RUN tar -xzf /tmp/multica.tar.gz -C /usr/local/bin multica && \
    chmod +x /usr/local/bin/multica && \
    rm /tmp/multica.tar.gz && \
    multica version

# 创建用户和目录（合并到一层）
RUN addgroup -S analyzer && \
    adduser -S -G analyzer -h /home/analyzer -s /bin/sh analyzer && \
    mkdir -p /home/analyzer/.claude/skills /workspace /config-repo /app /cases && \
    chown -R analyzer:analyzer /home/analyzer/.claude /workspace /config-repo /app /cases

# 复制文件（从 docker/ 复制构建产物）
COPY --chown=analyzer:analyzer docker/rain-excel-checker /usr/local/bin/rain-excel-checker
COPY --chown=analyzer:analyzer docker/skills/ /home/analyzer/.claude/skills/
# 合并 mjs-skills（由 build.sh 克隆到 docker/skills-external/，覆盖同名内置 skill）
# skills-external/ 不提交 git，由 build.sh 在 docker build 前创建；故必须先跑 build.sh 再 docker build
COPY --chown=analyzer:analyzer docker/skills-external/ /home/analyzer/.claude/skills/
COPY --chown=analyzer:analyzer docker/entrypoint.sh /app/entrypoint.sh
COPY --chown=analyzer:analyzer docker/notify.sh /app/notify.sh

# 复制 multica 登录配置（预置认证状态，避免容器内重复登录）
COPY --chown=analyzer:analyzer docker/multica-config.json /home/analyzer/.multica/config.json

# 确保二进制和脚本有可执行权限
RUN chmod +x /usr/local/bin/rain-excel-checker /app/entrypoint.sh /app/notify.sh

# 运行时环境
ENV PATH="/app:/usr/local/bin:${PATH}"
ENV NODE_PATH="/usr/local/lib/node_modules:${NODE_PATH}"

USER analyzer
WORKDIR /workspace

ENTRYPOINT ["bash", "/app/entrypoint.sh"]