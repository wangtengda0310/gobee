#!/bin/bash
# check-sync.sh - 检查 skill 文件是否同步
#
# 用途：CI/CD 或 pre-commit hook 中检查嵌入文件与源文件是否同步

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SOURCE_DIR="$PROJECT_ROOT/.claude/skills/gameconfig"
EMBEDDED_DIR="$PROJECT_ROOT/cmd/install-skill/skills"

echo "🔍 检查 skill 文件同步状态..."

# 检查源目录是否存在
if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ 源目录不存在: $SOURCE_DIR"
    exit 1
fi

# 检查嵌入目录是否存在
if [ ! -d "$EMBEDDED_DIR" ]; then
    echo "❌ 嵌入目录不存在: $EMBEDDED_DIR"
    echo "请运行: cd cmd/install-skill && go generate"
    exit 1
fi

# 比较关键文件
check_file() {
    local file="$1"
    local src_path="$SOURCE_DIR/$file"
    local emb_path="$EMBEDDED_DIR/$file"

    if [ ! -f "$src_path" ]; then
        return 0
    fi

    if [ ! -f "$emb_path" ]; then
        echo "❌ 嵌入文件缺失: $file"
        return 1
    fi

    # 比较文件大小（简单检查）
    src_size=$(stat -c%s "$src_path" 2>/dev/null || stat -f%z "$src_path")
    emb_size=$(stat -c%s "$emb_path" 2>/dev/null || stat -f%z "$emb_path")

    if [ "$src_size" -ne "$emb_size" ]; then
        echo "❌ 文件不同步: $file"
        echo "   源文件大小: $src_size"
        echo "   嵌入文件大小: $emb_size"
        echo ""
        echo "请运行: cd cmd/install-skill && go generate"
        return 1
    fi
}

# 检查关键文件
check_file "SKILL.md"
check_file "abilities/AI指导.md"

echo "✅ skill 文件同步正常"
