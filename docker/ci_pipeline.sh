#!/bin/bash
# ============================================================
# CI/CD 分析流水线启动脚本（通用，不绑定具体仓库类型）
# 在宿主机执行：前置检查 → docker run 启动分析容器
# 使用方法: CI_TARGET_REPO_PATH=... CI_TARGET_REPO_TYPE=... ./ci_pipeline.sh [--commit <hash>] [--help]
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ==================== 配置（除标注外均为环境变量传入）====================

# 镜像名称（需要先构建: ./build.sh）
IMAGE_NAME="${IMAGE_NAME:-rain-config-analyzer:latest}"

# 目标仓库路径（必填，各流水线不同：策划配表/前端/后端 各自传入）
# 目标仓库类型（必填：config/frontend/backend，用于日志和报告）
# ↑ 这两个变量下方做必填校验，无默认值

# rain-qa-func 工作空间路径（本工具仓库，基础设施，保留默认）
WORKSPACE_PATH="${WORKSPACE_PATH:-/root/ExcelChecker/rain-qa-func}"
CASE_DIR="${CASE_DIR:-/cases}"  # 容器内相对 WORKSPACE 的路径

# Claude Code 配置文件路径
SETTINGS_PATH="${SETTINGS_PATH:-$SCRIPT_DIR/settings.json}"

# ==================== 解析参数 ====================

TARGET_COMMIT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --commit)
            if [ -z "${2:-}" ]; then
                echo "错误: --commit 需要指定 commit hash"
                echo "使用方法: $0 [--commit <hash>] [--help]"
                exit 1
            fi
            TARGET_COMMIT="$2"
            shift 2
            ;;
        --help|-h)
            echo "CI/CD 分析流水线启动脚本"
            echo ""
            echo "使用方法: CI_TARGET_REPO_PATH=... CI_TARGET_REPO_TYPE=... $0 [--commit <hash>] [--help]"
            echo ""
            echo "选项:"
            echo "  --commit <hash>   分析指定 commit"
            echo "  --help, -h        显示此帮助信息"
            echo ""
            echo "必填环境变量:"
            echo "  CI_TARGET_REPO_PATH     目标仓库路径（各流水线不同，无默认）"
            echo "  CI_TARGET_REPO_TYPE     目标仓库类型：config / frontend / backend"
            echo ""
            echo "可选环境变量:"
            echo "  IMAGE_NAME              Docker 镜像名称（默认: rain-config-analyzer:latest）"
            echo "  WORKSPACE_PATH          rain-qa-func 项目路径（默认: /root/ExcelChecker/rain-qa-func）"
            echo "  CASE_DIR                cases 子目录名（默认: cases）"
            echo "  SETTINGS_PATH           Claude Code 配置文件路径"
            echo "  CI_INVOKE_SKILL         Claude 要执行的 skill（默认: analyze-prompt）"
            echo "  CI_ENTRYPOINT_NOTIFY    1=entrypoint 兜底把 Claude 输出发飞书（默认: 0）"
            echo "  CI_CLAUDE_MONITOR_EMAIL 脚本层监控通知接收邮箱（默认: v-wangtengda@ztgame.com，他人/其他流水线可覆盖）"
            echo "  CI_CLAUDE_FEISHU_ROBOT  飞书机器人 Webhook GUID"
            echo "  CI_CLAUDE_FEISHU_DM_APP_ID     飞书私聊应用 App ID"
            echo "  CI_CLAUDE_FEISHU_DM_APP_SECRET 飞书私聊应用 App Secret"
            echo "  CI_CLAUDE_CHECK_MODE           检查模式 (full=全量检查，定时任务用，空=增量)"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            echo "使用方法: $0 [--commit <hash>] [--help]"
            exit 1
            ;;
    esac
done

# ==================== 必填参数校验 ====================

if [ -z "${CI_TARGET_REPO_PATH:-}" ]; then
    echo "错误: 必填环境变量 CI_TARGET_REPO_PATH 未设置（目标仓库路径，各流水线不同）" >&2
    echo "  示例: CI_TARGET_REPO_PATH=/root/ExcelChecker/xcard-excel-config $0" >&2
    exit 3
fi
if [ -z "${CI_TARGET_REPO_TYPE:-}" ]; then
    echo "错误: 必填环境变量 CI_TARGET_REPO_TYPE 未设置（config / frontend / backend）" >&2
    exit 3
fi

# ==================== 前置检查 ====================

echo "========================================="
echo "CI/CD 分析流水线"
echo "========================================="
echo "目标仓库: $CI_TARGET_REPO_PATH"
echo "仓库类型: $CI_TARGET_REPO_TYPE"
echo "工作空间: $WORKSPACE_PATH"
echo "Cases 目录: $CASE_DIR"
echo "镜像: $IMAGE_NAME"
if [ -n "$TARGET_COMMIT" ]; then
    echo "目标 commit: $TARGET_COMMIT"
fi
echo "========================================="

# 检查镜像是否存在
if ! docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
    echo "错误: 镜像 $IMAGE_NAME 不存在"
    echo "请先构建: cd $SCRIPT_DIR && ./build.sh"
    exit 1
fi

# 检查 settings.json 是否存在
if [ ! -f "$SETTINGS_PATH" ]; then
    echo "错误: Claude Code 配置文件不存在: $SETTINGS_PATH"
    echo "请复制 settings.json.example 为 settings.json 并填入配置"
    exit 1
fi

# 检查 settings.json 权限（不应是全局可写）
SETTINGS_PERMS=$(stat -c "%a" "$SETTINGS_PATH" 2>/dev/null || stat -f "%Lp" "$SETTINGS_PATH" 2>/dev/null || echo "unknown")
if [ "$SETTINGS_PERMS" != "unknown" ]; then
    # 提取其他用户的写权限（最后一位）
    OTHERS_WRITE="${SETTINGS_PERMS: -1}"
    if [ "$((OTHERS_WRITE % 2))" -eq 1 ] || [ "$((OTHERS_WRITE / 2 % 2))" -eq 1 ]; then
        echo "警告: settings.json 权限过于开放 (${SETTINGS_PERMS})，建议设置为 600 或 644"
        echo "  修复: chmod 600 $SETTINGS_PATH"
    fi
fi

# 检查目标仓库（通用，不假设是配表）
if [ ! -d "$CI_TARGET_REPO_PATH/.git" ]; then
    echo "错误: 目标仓库不存在或不是 git 仓库: $CI_TARGET_REPO_PATH"
    exit 1
fi

# 检查工作空间（rain-qa-func）
if [ ! -d "$WORKSPACE_PATH" ]; then
    echo "错误: 工作空间目录不存在: $WORKSPACE_PATH"
    exit 1
fi

# 检查 cases 目录（可选，如果不存在警告但不阻止）
CASE_FULL_PATH="$WORKSPACE_PATH/$CASE_DIR"
if [ ! -d "$CASE_FULL_PATH" ]; then
    echo "警告: Cases 目录不存在: $CASE_FULL_PATH"
    echo "提示: 分析器将在空 cases 目录下运行"
fi

# ==================== 获取流水线环境变量 ====================

# CI 环境变量（如果存在则覆盖）
BK_USER="${BK_CI_START_USER_NAME:-}"
BK_BRANCH="${BK_CI_HOOK_BRANCH:-}"
BK_REVISION="${BK_CI_HOOK_REVISION:-}"

echo "流水线环境变量:"
echo "  BK_CI_START_USER_NAME: ${BK_USER:-未设置}"
echo "  BK_CI_HOOK_BRANCH: ${BK_BRANCH:-未设置}"
echo "  BK_CI_HOOK_REVISION: ${BK_REVISION:-未设置}"

# ==================== 启动容器 ====================

echo ""
echo "启动分析容器..."

# 运行容器并捕获退出码
# --add-host: 项目内网 AI 服务域名解析（基础设施层，随环境调整）
set +e
docker run --rm \
    --add-host ai-office.ztgame.com:10.254.40.145 \
    -e TARGET_COMMIT="$TARGET_COMMIT" \
    -e CI_INVOKE_SKILL="${CI_INVOKE_SKILL:-analyze-prompt}" \
    -e CI_ENTRYPOINT_NOTIFY="${CI_ENTRYPOINT_NOTIFY:-0}" \
    -e CI_CLAUDE_MONITOR_EMAIL="${CI_CLAUDE_MONITOR_EMAIL:-v-wangtengda@ztgame.com}" \
    -e FEISHU_ROBOT="${CI_CLAUDE_FEISHU_ROBOT:-none}" \
    -e FEISHU_DM_APP_ID="${CI_CLAUDE_FEISHU_DM_APP_ID:-}" \
    -e FEISHU_DM_APP_SECRET="${CI_CLAUDE_FEISHU_DM_APP_SECRET:-}" \
    -e CI_CLAUDE_CHECK_MODE="${CI_CLAUDE_CHECK_MODE:-}" \
    -e BK_CI_START_USER_NAME="$BK_USER" \
    -e BK_CI_HOOK_BRANCH="$BK_BRANCH" \
    -e BK_CI_HOOK_REVISION="$BK_REVISION" \
    -e CI_TARGET_REPO_TYPE="$CI_TARGET_REPO_TYPE" \
    -e CASE_DIR="/workspace/$CASE_DIR" \
    -v "$CI_TARGET_REPO_PATH:/config-repo:ro" \
    -v "$WORKSPACE_PATH:/workspace:ro" \
    -v "$SETTINGS_PATH:/home/analyzer/.claude/settings.json:ro" \
    "$IMAGE_NAME"

EXIT_CODE=$?
set -e

# ==================== 结构化退出码处理 ====================

echo ""
echo "========================================="

case $EXIT_CODE in
    0)
        echo "状态: 成功"
        echo "说明: 分析完成，无异常"
        ;;
    1)
        echo "状态: 失败"
        echo "说明: 分析执行失败（Claude Code 错误或内部异常）"
        ;;
    2)
        echo "状态: 警告"
        echo "说明: 无文件变更或无需分析内容，跳过分析"
        ;;
    3)
        echo "状态: 配置错误"
        echo "说明: 环境变量或配置缺失"
        ;;
    125)
        echo "状态: Docker 错误"
        echo "说明: Docker 守护进程问题"
        ;;
    130)
        echo "状态: 中断"
        echo "说明: 用户或系统中断 (Ctrl+C)"
        ;;
    *)
        echo "状态: 未知 (exit code: $EXIT_CODE)"
        echo "说明: 未定义的退出状态"
        ;;
esac

echo "========================================="

exit $EXIT_CODE
