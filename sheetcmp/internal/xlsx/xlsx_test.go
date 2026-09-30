package xlsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/wangtengda0310/gobee/sheetcmp/internal/engine"
)

// writeFile 写文本文件 (测试夹具用)。
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

// newFixture 在临时目录创建一个两工作表的测试文件:
// Sheet1: 表头 ID/Name + 3 行数据 + 一处公式; S2: 一个独立值。
// 返回文件路径。测试数据的"表"均指单个 sheet, 与文件概念区分 (项目术语约定)。
func newFixture(t *testing.T) string {
	t.Helper()
	f := excelize.NewFile()
	sheet1 := f.GetSheetName(0)
	set := func(cell, v string) {
		require.NoError(t, f.SetCellValue(sheet1, cell, v))
	}
	set("A1", "ID")
	set("B1", "Name")
	set("A2", "a")
	set("B2", "x")
	set("A3", "b")
	set("B3", "y")
	set("A4", "c")
	set("B4", "z")
	require.NoError(t, f.SetCellFormula(sheet1, "C4", "=1+1"))

	_, err := f.NewSheet("S2")
	require.NoError(t, err)
	require.NoError(t, f.SetCellValue("S2", "A1", "keep-me"))
	require.NoError(t, f.SetColWidth(sheet1, "A", "A", 20)) // 样式类信息, 写回应保留

	path := filepath.Join(t.TempDir(), "fixture.xlsx")
	require.NoError(t, f.SaveAs(path))
	return path
}

// ---------- Load 读取 ----------

func TestLoad_HeadersAndDataRows(t *testing.T) {
	// 首行作表头; 数据行 Index 从 2 开始 (源表绝对行号, 写回定位依赖)
	s, err := Load(newFixture(t), "")
	require.NoError(t, err)
	assert.Equal(t, []string{"ID", "Name"}, s.Headers)
	require.Len(t, s.Rows, 3)
	assert.Equal(t, 2, s.Rows[0].Index) // 数据首行在源表第 2 行
	assert.Equal(t, "a", s.Rows[0].Cells[0].Value)
	assert.Equal(t, "x", s.Rows[0].Cells[1].Value)
	assert.Equal(t, 4, s.Rows[2].Index)
}

func TestLoad_FormulaSeparated(t *testing.T) {
	// 公式单元格: Value 为缓存计算值 (无缓存时为空), Formula 单独存放
	s, err := Load(newFixture(t), "")
	require.NoError(t, err)
	c := s.Rows[2].Cells[2] // C4 = "=1+1"
	assert.Equal(t, "=1+1", c.Formula)
}

func TestLoad_NamedSheet(t *testing.T) {
	// 指定工作表名读取; 表名错误返回报错
	path := newFixture(t)
	s, err := Load(path, "S2")
	require.NoError(t, err)
	assert.Equal(t, "S2", s.Name)
	require.Len(t, s.Rows, 0) // S2 只有表头行 "keep-me" 被当作 Headers
	assert.Equal(t, []string{"keep-me"}, s.Headers)

	_, err = Load(path, "不存在")
	assert.Error(t, err)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "no-such.xlsx"), "")
	assert.Error(t, err)
}

func TestLoad_DefaultFirstSheet(t *testing.T) {
	// sheet 为空串时取第一个工作表, Sheet.Name 回填实际表名
	s, err := Load(newFixture(t), "")
	require.NoError(t, err)
	assert.Equal(t, "Sheet1", s.Name)
}

// ---------- ApplyCellEdits 写回 ----------

func TestApplyCellEdits_PreservesOthers(t *testing.T) {
	// 原位更新: 修改的单元格生效, 其他工作表/其他单元格/列宽全部保留
	path := newFixture(t)
	err := ApplyCellEdits(path, "", []CellEdit{
		{Row: 2, Col: 2, Value: "modified"}, // B2: x -> modified
		{Row: 3, Col: 1, Value: ""},         // A3: b -> 清空
		{Row: 4, Col: 2, Formula: "=2+2"},   // B4: z -> 公式
	})
	require.NoError(t, err)

	s, err := Load(path, "")
	require.NoError(t, err)
	require.Len(t, s.Rows, 3)
	assert.Equal(t, "modified", s.Rows[0].Cells[1].Value) // B2 已改
	assert.Equal(t, "", s.Rows[1].Cells[0].Value)         // A3 已清空
	assert.Equal(t, "=2+2", s.Rows[2].Cells[1].Formula)   // B4 变公式

	// 其他工作表未被破坏
	f, err := excelize.OpenFile(path)
	require.NoError(t, err)
	defer f.Close()
	v, err := f.GetCellValue("S2", "A1")
	require.NoError(t, err)
	assert.Equal(t, "keep-me", v)
	// 列宽保留 (原位更新的证明)
	w, err := f.GetColWidth("Sheet1", "A")
	require.NoError(t, err)
	assert.Equal(t, float64(20), w)
}

func TestApplyCellEdits_EmptyNoop(t *testing.T) {
	// 空 edits 直接返回, 不动文件
	path := newFixture(t)
	require.NoError(t, ApplyCellEdits(path, "", nil))
	s, err := Load(path, "")
	require.NoError(t, err)
	assert.Equal(t, "x", s.Rows[0].Cells[1].Value) // 原值未动
}

// ---------- 引擎联动: 端到端 ----------

func TestEndToEnd_LoadCompareApply(t *testing.T) {
	// 端到端: 两文件加载 → engine 比对 → 差异写回左文件 → 重比无差异
	leftPath := newFixture(t)

	f := excelize.NewFile()
	sh := f.GetSheetName(0)
	for cell, v := range map[string]string{
		"A1": "ID", "B1": "Name", "A2": "a", "B2": "x", "A3": "b", "B3": "y", "A4": "c", "B4": "w",
	} {
		require.NoError(t, f.SetCellValue(sh, cell, v))
	}
	rightPath := filepath.Join(t.TempDir(), "right.xlsx")
	require.NoError(t, f.SaveAs(rightPath))

	left, err := Load(leftPath, "")
	require.NoError(t, err)
	right, err := Load(rightPath, "")
	require.NoError(t, err)

	d, err := engine.CompareSheets(left, right, engine.AlignOptions{})
	require.NoError(t, err)
	require.True(t, d.HasRowDifference())
	// 两处差异: B4 值 z vs w; C4 公式 =1+1 vs 无 (值同公式异 → FormulaDiffers,
	// 右文件没写该公式——引擎公式防线正确暴露了这处静默差异)
	require.Len(t, d.Diffs, 2)
	assert.False(t, d.Diffs[0].FormulaDiffers) // B4 z vs w
	assert.True(t, d.Diffs[1].FormulaDiffers)  // C4 公式差异

	// 把两处差异都写回左文件消除 (值写回 + 公式清除), GUI 行同步将复用此模式
	var edits []CellEdit
	for _, c := range d.Diffs {
		edit := CellEdit{Row: d.Pairs[c.PairIndex].Left.Index, Col: c.Col + 1}
		if c.FormulaDiffers {
			edit.ClearFormula = true // 右侧无公式 → 清除左侧公式
		} else {
			edit.Value = c.Right
		}
		edits = append(edits, edit)
	}
	require.NoError(t, ApplyCellEdits(leftPath, "", edits))

	left2, err := Load(leftPath, "")
	require.NoError(t, err)
	d2, err := engine.CompareSheets(left2, right, engine.AlignOptions{})
	require.NoError(t, err)
	assert.False(t, d2.HasRowDifference())
}

// ---------- CSV ----------

func TestLoadCSV(t *testing.T) {
	// CSV 与 xlsx 统一进 engine.Sheet: 首行表头, 数据行 Index 从 2 开始
	dir := t.TempDir()
	path := filepath.Join(dir, "t.csv")
	require.NoError(t, writeFile(path, "ID,Name\na,x\nb,y\n"))

	s, err := LoadCSV(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"ID", "Name"}, s.Headers)
	require.Len(t, s.Rows, 2)
	assert.Equal(t, 2, s.Rows[0].Index)
	assert.Equal(t, "a", s.Rows[0].Cells[0].Value)
	assert.Equal(t, "y", s.Rows[1].Cells[1].Value)
}
