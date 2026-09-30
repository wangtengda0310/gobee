package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wangtengda0310/gobee/sheetcmp/internal/engine"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/textdiff"
)

// ---------- FromSheetDiff ----------

func TestFromSheetDiff_RowViews(t *testing.T) {
	// 配对行/单侧独有行 → RowView 行号与值正确
	left := &engine.Sheet{Headers: []string{"ID", "V"}, Rows: []engine.Row{
		{Index: 2, Cells: []engine.Cell{{Value: "a"}, {Value: "1"}}},
		{Index: 3, Cells: []engine.Cell{{Value: "b"}, {Value: "2"}}},
	}}
	right := &engine.Sheet{Headers: []string{"ID", "V"}, Rows: []engine.Row{
		{Index: 2, Cells: []engine.Cell{{Value: "a"}, {Value: "9"}}},
	}}
	d, err := engine.CompareSheets(left, right, engine.AlignOptions{})
	require.NoError(t, err)

	res := FromSheetDiff(d, left)
	require.Len(t, res.Rows, 2)

	paired := res.Rows[0]
	assert.Equal(t, 2, paired.LeftRow)
	assert.Equal(t, 2, paired.RightRow)
	assert.Equal(t, []string{"a", "1"}, paired.Left)
	assert.Equal(t, []string{"a", "9"}, paired.Right)
	// 差异格挂到所属行: V 列 (col=1) modified
	require.Len(t, paired.Diffs, 1)
	assert.Equal(t, 1, paired.Diffs[0].Col)
	assert.Equal(t, "modified", paired.Diffs[0].Kind)

	onlyLeft := res.Rows[1]
	assert.Equal(t, 3, onlyLeft.LeftRow)
	assert.Equal(t, 0, onlyLeft.RightRow)
	assert.Empty(t, onlyLeft.Right)
	assert.Empty(t, onlyLeft.Diffs) // 独有行无单元格差异

	assert.Equal(t, []string{"ID", "V"}, res.Headers)
	assert.Equal(t, 1, res.Stats.LeftOnlyRows)
}

func TestFromSheetDiff_FormulaMarker(t *testing.T) {
	// 值同公式异: CellView.FormulaDiffers 透传 (前端特别强调样式依据)
	left := &engine.Sheet{Rows: []engine.Row{
		{Index: 2, Cells: []engine.Cell{{Value: "a"}, {Value: "1", Formula: "=1"}}},
	}}
	right := &engine.Sheet{Rows: []engine.Row{
		{Index: 2, Cells: []engine.Cell{{Value: "a"}, {Value: "1", Formula: "=2"}}},
	}}
	d, err := engine.CompareSheets(left, right, engine.AlignOptions{})
	require.NoError(t, err)

	res := FromSheetDiff(d, left)
	require.Len(t, res.Rows, 1)
	require.Len(t, res.Rows[0].Diffs, 1)
	assert.True(t, res.Rows[0].Diffs[0].FormulaDiffers)
}

// ---------- FromLinePairs ----------

func TestFromLinePairs(t *testing.T) {
	// 行 diff → LineView 序列: equal/left/right 三类与统计正确
	l := []string{"a", "old", "c"}
	r := []string{"a", "new", "c"}
	res := FromLinePairs(textdiff.CompareLines(l, r), l, r)

	kinds := make([]string, 0, len(res.Lines))
	for _, lv := range res.Lines {
		kinds = append(kinds, lv.Kind)
	}
	assert.Equal(t, []string{"equal", "left", "right", "equal"}, kinds)
	assert.Equal(t, "old", res.Lines[1].Text)
	assert.Equal(t, 2, res.Lines[2].RightIdx)
	assert.Equal(t, 2, res.Stats.MatchedRows)
	assert.Equal(t, 1, res.Stats.LeftOnlyRows)
	assert.Equal(t, 1, res.Stats.RightOnlyRows)
}

// ---------- Compare (端到端, 走真实文件) ----------

func TestCompare_TextModeEndToEnd(t *testing.T) {
	// 文本模式端到端: 临时文件 → Compare → Kind/Lines 正确
	dir := t.TempDir()
	lf, rf := dir+"/l.txt", dir+"/r.txt"
	require.NoError(t, os.WriteFile(lf, []byte("a\nb\n"), 0o644))
	require.NoError(t, os.WriteFile(rf, []byte("a\nc\n"), 0o644))

	s := &CompareService{}
	res, err := s.Compare(CompareRequest{LeftPath: lf, RightPath: rf})
	require.NoError(t, err)
	assert.Equal(t, "text", res.Kind)
	assert.Equal(t, 1, res.Stats.LeftOnlyRows)
	assert.Equal(t, 1, res.Stats.RightOnlyRows)
	assert.Equal(t, "l.txt", res.LeftName)
}

func TestCompare_MissingPath(t *testing.T) {
	// 参数校验: 缺文件路径报错
	s := &CompareService{}
	_, err := s.Compare(CompareRequest{LeftPath: "x"})
	assert.Error(t, err)
}
