#!/bin/bash
# 分支安全守卫 — 拦截在 main/dev/master 分支上的 commit/push 操作
# 由 PreToolUse hook 调用，接收工具输入作为 stdin
# 注意：此 hook 只对 rain-qa-func 仓库生效

# 读取工具输入
INPUT=$(cat)

# 检查是否包含 git commit 或 git push 命令
if echo "$INPUT" | grep -qE 'git[[:space:]]+(commit|push)'; then
    # 从命令中提取工作目录
    GIT_DIR=""

    # 尝试提取 "cd /path &&" 格式
    CD_PATH=$(echo "$INPUT" | sed -n 's/.*cd[[:space:]][[:space:]]*\([^&|;]*\).*/\1/p' | head -1 | tr -d ' ')
    if [ -n "$CD_PATH" ] && [ -d "$CD_PATH" ]; then
        GIT_DIR="$CD_PATH"
    fi

    # 如果没有 cd，尝试提取 "git -C /path" 格式
    if [ -z "$GIT_DIR" ]; then
        C_PATH=$(echo "$INPUT" | sed -n 's/.*git[[:space:]][[:space:]]*-C[[:space:]][[:space:]]*\([^[:space:]]*\).*/\1/p' | head -1)
        if [ -n "$C_PATH" ] && [ -d "$C_PATH" ]; then
            GIT_DIR="$C_PATH"
        fi
    fi

    # 获取分支和仓库名
    if [ -n "$GIT_DIR" ]; then
        BRANCH=$(cd "$GIT_DIR" 2>/dev/null && git branch --show-current 2>/dev/null)
        REPO_NAME=$(cd "$GIT_DIR" 2>/dev/null && git rev-parse --show-toplevel 2>/dev/null | xargs basename)
    else
        BRANCH=$(git branch --show-current 2>/dev/null)
        REPO_NAME=$(git rev-parse --show-toplevel 2>/dev/null | xargs basename)
    fi

    # 只对 rain-qa-func 仓库生效
    case "$REPO_NAME" in
        rain-qa-func|xcard-qa-tools)
            # 检查是否在受保护分支上
            case "$BRANCH" in
                main|dev|master)
                    echo "BLOCKED: 当前在 '$BRANCH' 分支，禁止直接 commit/push。请使用 worktree 分支。" >&2
                    echo "提示: 使用 claude -w 创建 worktree，或在 worktree 分支上操作" >&2
                    exit 2
                    ;;
            esac
            ;;
    esac
fi

# 允许操作
exit 0
