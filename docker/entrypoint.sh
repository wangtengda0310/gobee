#!/bin/bash
# ============================================================
# Docker 流水线 Claude 执行入口
#
# 职责：容器生命周期、CI 环境准备、启动 Claude 调用指定 skill。
# 不包含任何业务分析逻辑（git diff、分支判断、通知策略等），
# 全部下沉到 $CI_INVOKE_SKILL 对应的 skill 文档中。
#
# 行为由命令行参数控制：
#   docker run analyzer                    # 分析 HEAD，使用 CI_INVOKE_SKILL
#   docker run analyzer --commit abc123    # 分析指定 commit
#   docker run analyzer "自定义提示词"      # 使用自定义提示词（跳过 skill）
# ============================================================

set -euo pipefail

# 临时文件列表（供 cleanup 使用）
TEMP_FILES=""

cleanup() {
	if [ -n "$TEMP_FILES" ]; then
		# shellcheck disable=SC2086
		rm -f $TEMP_FILES
	fi
}
trap cleanup EXIT

# 先切换到非 git 目录，避免 worktree .git 文件路径解析失败
cd /app

# 解决挂载目录的 dubious ownership 问题（容器内 UID 与宿主机不同）
git config --global --add safe.directory /config-repo
# 中文文件名不转义（否则 grep .xlsx 匹配不到）
git config --global core.quotepath false

# ==================== 配置 ====================
# 读取并导出环境变量，确保 Claude 子进程能继承

export FEISHU_ROBOT="${FEISHU_ROBOT:-none}"
export CONFIG_REPO="${CONFIG_REPO:-/config-repo}"
export TARGET_COMMIT="${TARGET_COMMIT:-}"
export CASE_DIR="${CASE_DIR:-/workspace/cases}"
export CI_INVOKE_SKILL="${CI_INVOKE_SKILL:-analyze-prompt}"
SKILLS_DIR="/home/analyzer/.claude/skills"
export CI_CLAUDE_MONITOR_EMAIL="${CI_CLAUDE_MONITOR_EMAIL:-v-wangtengda@ztgame.com}"
# 兜底通知开关：1 时 entrypoint 在 Claude 执行完后把输出发到飞书，
# 用于 skill 自身不发通知的场景（如 mail-check）。默认关闭，避免与
# skill 内部通知（如 analyze-prompt）重复。
CI_ENTRYPOINT_NOTIFY="${CI_ENTRYPOINT_NOTIFY:-0}"
# notify.sh 依赖 MONITOR_EMAIL 变量名
export MONITOR_EMAIL="$CI_CLAUDE_MONITOR_EMAIL"

# 这些变量由 ci_pipeline.sh 透传，skill 中可能需要
export FEISHU_DM_APP_ID="${FEISHU_DM_APP_ID:-}"
export FEISHU_DM_APP_SECRET="${FEISHU_DM_APP_SECRET:-}"
export BK_CI_HOOK_BRANCH="${BK_CI_HOOK_BRANCH:-}"
export BK_CI_HOOK_REVISION="${BK_CI_HOOK_REVISION:-}"

# skill 中统一通过 CI_* 名字读取的关键变量
export CI_CONFIG_REPO="$CONFIG_REPO"
export CI_CASE_DIR="$CASE_DIR"
export CI_TARGET_COMMIT="$TARGET_COMMIT"
# 检查模式：full=强制全量检查（定时任务用），空=由 skill 按分支/merge 判断
export CI_CLAUDE_CHECK_MODE="${CI_CLAUDE_CHECK_MODE:-}"

# ==================== 命令行参数解析 ====================
# 用法:
#   docker run analyzer                    # 使用 skill 分析
#   docker run analyzer --commit abc123    # 分析指定 commit
#   docker run analyzer "自定义提示词"      # 使用自定义提示词
#
CUSTOM_PROMPT=""

while [ $# -gt 0 ]; do
	case "$1" in
		--commit)
			if [ -n "${2:-}" ]; then
				TARGET_COMMIT="$2"
				shift 2
			else
				log_error "--commit 需要指定 commit hash"
				exit 1
			fi
			;;
		*)
			# 非选项参数视为自定义提示词
			CUSTOM_PROMPT="$1"
			shift
			;;
	esac
done

# --commit 可能修改了 TARGET_COMMIT，同步到 skill 读取的 CI_* 变量
export CI_TARGET_COMMIT="$TARGET_COMMIT"

# ==================== 工具函数 ====================

log_info() {
	echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $*"
}

log_error() {
	echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $*" >&2
}

log_section() {
	echo ""
	echo "-------------------$*-------------------"
	echo ""
}

# ==================== Multica 登录状态检查 ====================
# 镜像构建时已预置 config.json，通常无需重复登录
# 如需覆盖（如 token 轮换），通过环境变量 MULTICA_TOKEN 传入
if [ -n "${MULTICA_TOKEN:-}" ]; then
	log_info "Multica 使用环境变量重新登录..."
	if printf '%s\n' "$MULTICA_TOKEN" | multica login --token > /dev/null 2>&1; then
		log_info "Multica 登录成功"
	else
		log_error "Multica 登录失败（token 可能无效）"
	fi
else
	# 验证预置的登录状态是否有效
	if multica auth status > /dev/null 2>&1; then
		log_info "Multica 已登录（预置配置）"
	else
		log_info "Multica 未登录（无预置配置）"
	fi
fi

# ==================== 前置检查 ====================

log_section "Docker Claude 流水线 (skill: $CI_INVOKE_SKILL)"

if [ ! -d "$CONFIG_REPO/.git" ]; then
	log_error "CONFIG_REPO ($CONFIG_REPO) 不是有效的 git 仓库"
	exit 1
fi

if [ -n "$TARGET_COMMIT" ]; then
	log_info "模式: 指定 commit ($TARGET_COMMIT)"
else
	log_info "模式: 分析 HEAD"
fi

# ==================== 构建启动提示词 ====================
# 所有业务逻辑（git diff、分支判断、checker 执行、通知策略）
# 都在 $CI_INVOKE_SKILL 对应的 skill 中定义。
# 脚本只提供原始上下文变量和公共工具 /app/notify.sh。

PROMPT_FILE="/tmp/invoke-prompt.md"
TEMP_FILES="$TEMP_FILES $PROMPT_FILE"

if [ -n "$CUSTOM_PROMPT" ]; then
	log_info "使用自定义提示词"
	echo "$CUSTOM_PROMPT" > "$PROMPT_FILE"
else
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

	cat > "$PROMPT_FILE" <<EOF
请调用 /$CI_INVOKE_SKILL 技能，并严格按照该技能文档完成本次流水线任务。

## 可用公共资源

容器内已预置公共通知脚本 /app/notify.sh。若 skill 文档要求发送飞书消息，请使用 /app/notify.sh 并传入真实参数；不要在示例、测试或不确定的场景下调用，避免发送空消息干扰用户。

## 原始上下文环境变量

skill 中可通过 Bash 读取以下变量，但不要依赖 entrypoint 做任何业务计算：

- CI_CONFIG_REPO: 配表仓库路径（$CONFIG_REPO）
- CI_CASE_DIR: 测试用例目录（$CASE_DIR）
- CI_TARGET_COMMIT: 指定 commit（可能为空）
- CI_CLAUDE_CHECK_MODE: 检查模式（full=强制全量检查，定时任务用，空=按分支/merge 判断）
- CI_INVOKE_SKILL: 当前调用的 skill 名称
- BK_CI_HOOK_BRANCH / BK_CI_HOOK_REVISION: 流水线注入的分支/提交
- FEISHU_ROBOT: 群机器人 GUID
- FEISHU_DM_APP_ID / FEISHU_DM_APP_SECRET: 私聊应用凭证
- CI_CLAUDE_MONITOR_EMAIL: 监控通知邮箱
EOF
fi

# ==================== 运行 Claude Code ====================

log_section "Claude Code 执行中..."

CLAUDE_OUTPUT_FILE="/tmp/claude-text-output.txt"
TEMP_FILES="$TEMP_FILES $CLAUDE_OUTPUT_FILE /tmp/claude-output.json"

if [ "$FEISHU_ROBOT" = "none" ] || [ -z "$FEISHU_ROBOT" ]; then
	# 未配置飞书：本地调试模式，实时输出到终端（stream-json）
	log_info "开始分析（stream-json 模式，实时输出）..."
	claude -p "$(cat "$PROMPT_FILE")" \
		--dangerously-skip-permissions \
		--verbose \
		--output-format=stream-json \
		--debug-file /tmp/claude-debug.log 2>&1 | tee /tmp/claude-output.json
	EXIT_CODE=${PIPESTATUS[0]}
	# 从 stream-json 提取文本内容供兜底通知使用
	grep -o '"content":"[^"]*"' /tmp/claude-output.json 2>/dev/null \
		| sed 's/"content":"//;s/"$//' > "$CLAUDE_OUTPUT_FILE" || true
else
	# 配置了飞书：默认 text 输出，直接捕获
	claude -p "$(cat "$PROMPT_FILE")" \
		--dangerously-skip-permissions \
		--verbose \
		--debug-file /tmp/claude-debug.log > "$CLAUDE_OUTPUT_FILE" 2>&1
	EXIT_CODE=$?
	cat "$CLAUDE_OUTPUT_FILE"
fi

if [ $EXIT_CODE -ne 0 ]; then
	log_error "Claude Code 执行失败 (exit code: $EXIT_CODE)"
	exit 1
fi

# ==================== 兜底通知 ====================
# skill 自身不发通知时（CI_ENTRYPOINT_NOTIFY=1），由 entrypoint 把 Claude 输出发到飞书。
if [ "$CI_ENTRYPOINT_NOTIFY" = "1" ]; then
	log_section "entrypoint 兜底通知"
	CLAUDE_TEXT="$(cat "$CLAUDE_OUTPUT_FILE" 2>/dev/null || echo "")"
	if [ -z "$CLAUDE_TEXT" ]; then
		log_info "[兜底通知] Claude 输出为空，跳过"
	else
		TITLE="流水线执行结果 - ${CI_INVOKE_SKILL}"
		if [ "$FEISHU_ROBOT" != "none" ] && [ -n "$FEISHU_ROBOT" ]; then
			/app/notify.sh group "$TITLE" "$CLAUDE_TEXT" || true
		elif [ -n "${FEISHU_DM_APP_ID:-}" ] && [ -n "${FEISHU_DM_APP_SECRET:-}" ] && [ -n "$MONITOR_EMAIL" ]; then
			/app/notify.sh dm "$MONITOR_EMAIL" "$CLAUDE_TEXT" || true
		else
			log_info "[兜底通知] 未配置飞书群机器人或私聊凭证，跳过"
		fi
	fi
fi

log_section "Claude Code 执行完成"
exit 0
