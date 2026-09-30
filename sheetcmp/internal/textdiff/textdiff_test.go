package textdiff

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pairsCompact 把对齐结果压成可断言的紧凑形式: "L|R"=配对(左行号|右行号),
// "<L"=左独有, "R>"=右独有。行号均为 1-based。
func pairsCompact(pairs []LinePair) []string {
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		l, r := strconv.Itoa(p.Left), strconv.Itoa(p.Right)
		switch {
		case p.Left != 0 && p.Right != 0:
			out = append(out, l+"|"+r)
		case p.Left != 0:
			out = append(out, "<"+l)
		default:
			out = append(out, r+">")
		}
	}
	return out
}

func TestCompareLines_Identical(t *testing.T) {
	// 完全相同: 全部按行号 1:1 配对
	got := CompareLines([]string{"a", "b", "c"}, []string{"a", "b", "c"})
	assert.Equal(t, []string{"1|1", "2|2", "3|3"}, pairsCompact(got))
}

func TestCompareLines_RightInsertion(t *testing.T) {
	// 右侧中间插入一行: 插入行标为右独有, 其余行不错位
	// (与表格行号位置配对不同, 文本行 diff 是内容 LCS——错位会让 diff 淹没在噪声里)
	got := CompareLines([]string{"a", "b", "c"}, []string{"a", "X", "b", "c"})
	assert.Equal(t, []string{"1|1", "2>", "2|3", "3|4"}, pairsCompact(got))
}

func TestCompareLines_LeftDeletion(t *testing.T) {
	// 左侧独有行 (对右侧是删除): 标为左独有
	got := CompareLines([]string{"a", "X", "b"}, []string{"a", "b"})
	assert.Equal(t, []string{"1|1", "<2", "3|2"}, pairsCompact(got))
}

func TestCompareLines_TailDiffs(t *testing.T) {
	// 两侧尾部各有独有块: 都要按位置保留
	got := CompareLines([]string{"a", "b", "L1", "L2"}, []string{"a", "b", "R1", "R2", "R3"})
	assert.Equal(t, []string{"1|1", "2|2", "<3", "<4", "3>", "4>", "5>"}, pairsCompact(got))
}

func TestCompareLines_EmptySide(t *testing.T) {
	// 边界: 单侧为空
	got := CompareLines(nil, []string{"a"})
	require.Len(t, got, 1)
	assert.Equal(t, LinePair{Left: 0, Right: 1}, got[0])
	got = CompareLines([]string{"a"}, nil)
	require.Len(t, got, 1)
	assert.Equal(t, LinePair{Left: 1, Right: 0}, got[0])
	assert.Empty(t, CompareLines(nil, nil))
}

func TestCompareLines_ModifiedLine(t *testing.T) {
	// 行内容修改 = 一删一增两个单侧对 (行级 diff 不做行内字符级拆分)
	got := CompareLines([]string{"a", "old", "c"}, []string{"a", "new", "c"})
	assert.Equal(t, []string{"1|1", "<2", "2>", "3|3"}, pairsCompact(got))
}
