package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- 测试数据构造 ----------

// mkRows 便捷构造行: 每个参数是一行的单元格值, Index 自动按 1-based 顺序编号。
func mkRows(rows ...[]string) []Row {
	out := make([]Row, 0, len(rows))
	for i, cells := range rows {
		r := Row{Index: i + 1, Cells: make([]Cell, len(cells))}
		for j, v := range cells {
			r.Cells[j] = Cell{Value: v}
		}
		out = append(out, r)
	}
	return out
}

// pairKeys 把对齐结果压成可断言的紧凑形式:
// "a|b"=配对(左键|右键), "<a"=左独有, "b>"=右独有。
// 行键取第 keyCol 列值 (缺省第 0 列; 测试数据里该列均为行标识)。
func pairKeys(pairs []AlignPair) []string { return pairKeysCol(pairs, 0) }

func pairKeysCol(pairs []AlignPair, keyCol int) []string {
	cell := func(r *Row) string {
		if keyCol < len(r.Cells) {
			return r.Cells[keyCol].Value
		}
		return ""
	}
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		switch {
		case p.Left != nil && p.Right != nil:
			out = append(out, cell(p.Left)+"|"+cell(p.Right))
		case p.Left != nil:
			out = append(out, "<"+cell(p.Left))
		default:
			out = append(out, cell(p.Right)+">")
		}
	}
	return out
}

var hdr = []string{"ID", "Name", "Value"}

// ---------- 无关键列: 行序 LCS 对齐 ----------

func TestAlignByRowOrder_IdenticalSheets(t *testing.T) {
	// 完全一致的两表: 全部配对, 无独有行
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"a", "x", "1"}, []string{"b", "y", "2"}, []string{"c", "z", "3"},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"a", "x", "1"}, []string{"b", "y", "2"}, []string{"c", "z", "3"},
	)}
	got := pairKeys(AlignRows(left, right, AlignOptions{}))
	assert.Equal(t, []string{"a|a", "b|b", "c|c"}, got)
}

func TestAlignByRowOrder_MiddleInsertion(t *testing.T) {
	// 行号位置配对语义: 第 i 行对第 i 行。右表中间插入行会引起连锁错位
	// (b↔X, c↔b 各成配对, 右表尾部 c 落为独有)——这正是需要选关键列的场景,
	// 关键列模式下同样的输入能精确标出唯一插入行 (见 TestAlignByKey_*)。
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"a", "", ""}, []string{"b", "", ""}, []string{"c", "", ""},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"a", "", ""}, []string{"X", "", ""}, []string{"b", "", ""}, []string{"c", "", ""},
	)}
	got := pairKeys(AlignRows(left, right, AlignOptions{}))
	assert.Equal(t, []string{"a|a", "b|X", "c|b", "c>"}, got)
}

func TestAlignByRowOrder_EmptySheets(t *testing.T) {
	// 边界: 双空表 / 单侧空表
	empty := &Sheet{Name: "E", Headers: hdr}
	other := &Sheet{Name: "O", Headers: hdr, Rows: mkRows([]string{"a", "", ""})}

	assert.Empty(t, AlignRows(empty, &Sheet{Name: "E2", Headers: hdr}, AlignOptions{}))
	assert.Equal(t, []string{"<a"}, pairKeys(AlignRows(other, empty, AlignOptions{})))
	assert.Equal(t, []string{"a>"}, pairKeys(AlignRows(empty, other, AlignOptions{})))
}

// ---------- 关键列对齐: 行序无关 ----------

func TestAlignByKey_RowOrderIndependence(t *testing.T) {
	// 核心特性: 右表行序完全打乱, 按 ID 列依然正确配对
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"a", "", ""}, []string{"b", "", ""}, []string{"c", "", ""},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"c", "", ""}, []string{"a", "", ""}, []string{"b", "", ""},
	)}
	got := pairKeys(AlignRows(left, right, AlignOptions{KeyColumns: []int{0}}))
	assert.Equal(t, []string{"a|a", "b|b", "c|c"}, got)
}

func TestAlignByKey_OnlyOneSide(t *testing.T) {
	// 单侧独有行: 左独有(删除候选)与右独有(新增)都要保留为单侧对
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"a", "", ""}, []string{"b", "", ""}, []string{"d", "", ""},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"b", "", ""}, []string{"c", "", ""}, []string{"d", "", ""},
	)}
	got := pairKeys(AlignRows(left, right, AlignOptions{KeyColumns: []int{0}}))
	assert.ElementsMatch(t, []string{"<a", "b|b", "c>", "d|d"}, got)
}

func TestAlignByKey_DuplicateKeys(t *testing.T) {
	// 同 key 多行: 按各自出现顺序依次配对 (不做值偏好), 多出者落为单侧独有
	// (配置表场景: 同一 ID 允许多行记录; 语义在 align.go 注释中固定)
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"a", "1", ""}, []string{"a", "2", ""}, []string{"a", "3", ""},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"a", "9", ""}, []string{"a", "1", ""},
	)}
	// 用第 1 列 (Name) 作行标识区分同 key 的多行
	got := pairKeysCol(AlignRows(left, right, AlignOptions{KeyColumns: []int{0}}), 1)
	// 左 a1↔右第一个 a9, 左 a2↔右 a1, 左 a3 落为左独有
	assert.Equal(t, []string{"1|9", "2|1", "<3"}, got)
}

func TestAlignByKey_EmptyKeyFallback(t *testing.T) {
	// 关键列值为空的行: 不参与键匹配, 左右空键行之间按行序兜底配对
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"", "x1", ""}, []string{"a", "k", ""}, []string{"", "x2", ""},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"", "y1", ""}, []string{"a", "k", ""}, []string{"", "y2", ""}, []string{"", "y3", ""},
	)}
	// 用第 1 列 (Name) 作行标识: 空键行按行序 x1↔y1, x2↔y2, y3 落为右独有; 有键行 a 正常配对
	got := pairKeysCol(AlignRows(left, right, AlignOptions{KeyColumns: []int{0}}), 1)
	assert.ElementsMatch(t, []string{"x1|y1", "x2|y2", "y3>", "k|k"}, got)
}

func TestAlignByKey_KeyColumnOutOfRange(t *testing.T) {
	// 越界关键列: 不 panic, 该列视为空 (全部行落入空键兜底), 返回非 nil
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows([]string{"a", "", ""})}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows([]string{"b", "", ""})}
	got := AlignRows(left, right, AlignOptions{KeyColumns: []int{5}})
	require.NotNil(t, got)
	assert.NotEmpty(t, got)
}

// ---------- 行号保持 ----------

func TestAlign_PreservesSourceRowIndexes(t *testing.T) {
	// Row.Index 必须保留源表行号 (写回定位依赖), 对齐不重编号
	left := &Sheet{Name: "L", Headers: hdr, Rows: mkRows(
		[]string{"a", "", ""}, []string{"b", "", ""},
	)}
	right := &Sheet{Name: "R", Headers: hdr, Rows: mkRows(
		[]string{"b", "", ""},
	)} // 右表 b 在第 1 行
	pairs := AlignRows(left, right, AlignOptions{KeyColumns: []int{0}})
	for _, p := range pairs {
		if p.Left != nil && p.Right != nil {
			assert.Equal(t, 2, p.Left.Index)  // 左表 b 是第 2 行
			assert.Equal(t, 1, p.Right.Index) // 右表 b 是第 1 行
		}
	}
}
