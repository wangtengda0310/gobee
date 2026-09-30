package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mkCellRow 构造带公式的一行 (值与公式成对出现: v1, f1, v2, f2, ... 公式可为空串)。
func mkCellRow(index int, vf ...string) Row {
	r := Row{Index: index}
	for i := 0; i+1 < len(vf); i += 2 {
		r.Cells = append(r.Cells, Cell{Value: vf[i], Formula: vf[i+1]})
	}
	return r
}

func sheet2(l, r []Row) (*Sheet, *Sheet) {
	// 测试专用: 两表共用表头, 行 Index 已在构造时指定
	return &Sheet{Name: "L", Headers: hdr, Rows: l}, &Sheet{Name: "R", Headers: hdr, Rows: r}
}

// ---------- 基础差异 ----------

func TestCompare_IdenticalSheets_NoDiff(t *testing.T) {
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "x", "", "1", "")},
		[]Row{mkCellRow(1, "a", "", "x", "", "1", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	assert.Empty(t, d.Diffs)
	assert.Equal(t, DiffStats{MatchedRows: 1}, d.Stats)
	assert.False(t, d.HasRowDifference())
}

func TestCompare_CellModified(t *testing.T) {
	// 值不同 → DiffModified, 差异记录定位到正确行列
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "x", "", "1", "")},
		[]Row{mkCellRow(1, "a", "", "y", "", "1", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	require.Len(t, d.Diffs, 1)
	diff := d.Diffs[0]
	assert.Equal(t, 1, diff.Col) // 第 2 列 (0-based=1)
	assert.Equal(t, DiffModified, diff.Kind)
	assert.Equal(t, "x", diff.Left)
	assert.Equal(t, "y", diff.Right)
	assert.Equal(t, 1, d.Stats.ModifiedCells)
	assert.True(t, d.HasRowDifference())
}

func TestCompare_CellOnlyOneSide(t *testing.T) {
	// 一侧空一侧有值 → LeftOnly / RightOnly (空串不算"有值")
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "x", "", "", "")},
		[]Row{mkCellRow(1, "a", "", "", "", "9", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	require.Len(t, d.Diffs, 2)
	kinds := map[int]DiffKind{}
	for _, c := range d.Diffs {
		kinds[c.Col] = c.Kind
	}
	assert.Equal(t, DiffLeftOnly, kinds[1])  // 左有 x 右空
	assert.Equal(t, DiffRightOnly, kinds[2]) // 左空 右有 9
	assert.Equal(t, 1, d.Stats.LeftOnlyCells)
	assert.Equal(t, 1, d.Stats.RightOnlyCells)
}

// ---------- 公式差异 (静默事故防线) ----------

func TestCompare_FormulaDiffers_ValueSame(t *testing.T) {
	// 值相同但公式不同 → 仍算差异且 FormulaDiffers=true
	// (配置表事故形态: 改了公式但恰好算出同值, 肉眼比对值完全看不出)
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "", "", "1", "")},
		[]Row{mkCellRow(1, "a", "", "", "=SUM(1,0)", "1", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	require.Len(t, d.Diffs, 1)
	assert.Equal(t, DiffModified, d.Diffs[0].Kind)
	assert.True(t, d.Diffs[0].FormulaDiffers)
}

func TestCompare_FormulaSameNoDiff(t *testing.T) {
	// 值与公式都相同 → 无差异 (公式一致不算差异)
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "", "=SUM(1,0)", "1", "")},
		[]Row{mkCellRow(1, "a", "", "", "=SUM(1,0)", "1", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	assert.Empty(t, d.Diffs)
}

// ---------- 列数不齐 ----------

func TestCompare_ColumnCountMismatch(t *testing.T) {
	// 配对行列数不同: 短行缺失列按空单元格处理 (多出的列自然落为单侧差异), 不报错
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "x", "")},
		[]Row{mkCellRow(1, "a", "", "x", "", "9", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	require.Len(t, d.Diffs, 1)
	assert.Equal(t, 2, d.Diffs[0].Col)
	assert.Equal(t, DiffRightOnly, d.Diffs[0].Kind)
}

// ---------- 独有行与统计 ----------

func TestCompare_OnlyRowsCounted(t *testing.T) {
	// 独有行计入统计, 不产生单元格差异 (没有配对对象无从 diff)
	left, right := sheet2(
		[]Row{mkCellRow(1, "a", "", "", ""), mkCellRow(2, "b", "", "", "")},
		[]Row{mkCellRow(1, "a", "", "", "")},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	assert.Empty(t, d.Diffs)
	assert.Equal(t, 1, d.Stats.MatchedRows)
	assert.Equal(t, 1, d.Stats.LeftOnlyRows)
	assert.Equal(t, 0, d.Stats.RightOnlyRows)
	assert.True(t, d.HasRowDifference())
}

func TestCompare_StatsAggregation(t *testing.T) {
	// 多行多列混合: 统计聚合正确性
	left, right := sheet2(
		[]Row{
			mkCellRow(1, "a", "", "1", "", "1", ""),
			mkCellRow(2, "b", "", "2", "", "2", ""),
			mkCellRow(3, "c", "", "3", "", "3", ""),
		},
		[]Row{
			mkCellRow(1, "a", "", "9", "", "1", ""), // 行1: 1 处 modified
			mkCellRow(2, "b", "", "2", "", "9", ""), // 行2: 1 处 modified
			mkCellRow(4, "d", "", "4", "", "4", ""), // 行3: 左 c 独有, 右 d 独有
		},
	)
	d, err := CompareSheets(left, right, AlignOptions{KeyColumns: []int{0}})
	require.NoError(t, err)
	assert.Equal(t, 2, d.Stats.MatchedRows)
	assert.Equal(t, 1, d.Stats.LeftOnlyRows)
	assert.Equal(t, 1, d.Stats.RightOnlyRows)
	assert.Equal(t, 2, d.Stats.ModifiedCells)
}

// ---------- PairIndex 回溯 ----------

func TestCompare_PairIndexResolvesToPair(t *testing.T) {
	// CellDiff.PairIndex 必须能回溯到正确的 AlignPair (前端高亮定位依赖)
	left, right := sheet2(
		[]Row{
			mkCellRow(1, "a", "", "1", ""),
			mkCellRow(2, "b", "", "2", ""),
		},
		[]Row{
			mkCellRow(1, "a", "", "1", ""),
			mkCellRow(2, "b", "", "9", ""),
		},
	)
	d, err := CompareSheets(left, right, AlignOptions{})
	require.NoError(t, err)
	require.Len(t, d.Diffs, 1)
	pair := d.Pairs[d.Diffs[0].PairIndex]
	require.NotNil(t, pair.Left)
	require.NotNil(t, pair.Right)
	assert.Equal(t, "b", pair.Left.Cells[0].Value) // 差异确实落在 b 行
}
