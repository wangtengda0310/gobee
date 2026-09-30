// Package service 是 GUI 前端与 engine/xlsx/textdiff 之间的桥接层:
// 接收前端的比对请求, 完成文件加载与比对, 转成前端友好的视图 DTO。
// 本包不依赖 wails (纯 Go 可测), wails 通过绑定直接暴露 CompareService 方法。
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wangtengda0310/gobee/sheetcmp/internal/engine"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/textdiff"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/xlsx"
)

// CompareService 比对服务 (wails 绑定: 方法被前端直接调用)。
type CompareService struct{}

// CompareRequest 前端比对请求。文件类型按左文件扩展名分流 (与 CLI 一致)。
type CompareRequest struct {
	LeftPath   string `json:"leftPath"`
	RightPath  string `json:"rightPath"`
	KeyColumns []int  `json:"keyColumns"` // 表格模式关键列 (0-based); 空则按行号位置配对
	SheetName  string `json:"sheetName"`  // xlsx 表名; 空则第一个
}

// CompareResult 前端渲染数据: 表格模式填 Headers/Rows, 文本模式填 Lines。
type CompareResult struct {
	Kind    string           `json:"kind"` // "sheet" | "text"
	Headers []string         `json:"headers"`
	Stats   engine.DiffStats `json:"stats"`
	Rows    []RowView        `json:"rows"`
	Lines   []LineView       `json:"lines"`
	// 展示名: 文件基名 (表格模式追加表名)
	LeftName  string `json:"leftName"`
	RightName string `json:"rightName"`
}

// RowView 一对对齐行的视图: LeftRow/RightRow 为源表行号 (0=该侧独有),
// Left/Right 为单元格显示值数组, Diffs 为该行的差异格 (前端高亮用),
// LeftFormulas/RightFormulas 为与值数组平行的公式数组 (空串=非公式格, 行同步写公式用)。
type RowView struct {
	LeftRow       int        `json:"leftRow"`
	RightRow      int        `json:"rightRow"`
	Left          []string   `json:"left"`
	Right         []string   `json:"right"`
	LeftFormulas  []string   `json:"leftFormulas"`
	RightFormulas []string   `json:"rightFormulas"`
	Diffs         []CellView `json:"diffs"`
}

// CellView 差异格视图。Kind: "modified" | "left" | "right"。
type CellView struct {
	Col            int    `json:"col"`
	Kind           string `json:"kind"`
	FormulaDiffers bool   `json:"formulaDiffers"`
}

// LineView 文本模式行视图。Kind: "equal" | "left" | "right"。
type LineView struct {
	Kind     string `json:"kind"`
	LeftIdx  int    `json:"leftIdx"`
	RightIdx int    `json:"rightIdx"`
	Text     string `json:"text"`
}

// Compare 执行比对: 按左文件扩展名分流 → 表格模式 (xlsx.Load/LoadCSV + engine)
// 或文本模式 (readLines + textdiff) → 转视图 DTO。
func (s *CompareService) Compare(req CompareRequest) (*CompareResult, error) {
	if req.LeftPath == "" || req.RightPath == "" {
		return nil, fmt.Errorf("需要左右两个文件")
	}
	if isSheetExt(req.LeftPath) {
		return compareAsSheet(req)
	}
	return compareAsText(req)
}

// EditOp 一次单元格同步编辑 (前端攒批后经 Apply 一次性写回)。
// Row/Col 为目标文件中的 1-based 绝对行列号; 三种互斥形态与 xlsx.CellEdit 一致:
// Formula 非空写公式, ClearFormula=true 清公式, 其余写值。
type EditOp struct {
	Row          int    `json:"row"`
	Col          int    `json:"col"`
	Value        string `json:"value"`
	Formula      string `json:"formula"`
	ClearFormula bool   `json:"clearFormula"`
}

// ApplyRequest 批量写回请求: Edits 全部落向同一个目标文件。
type ApplyRequest struct {
	TargetPath string   `json:"targetPath"`
	SheetName  string   `json:"sheetName"`
	Edits      []EditOp `json:"edits"`
}

// Apply 把前端累积的同步编辑原位写回目标文件 (GUI"保存"调用)。
// 只支持 xlsx 目标 (CSV 为只读比对); 空编辑列表直接返回。
func (s *CompareService) Apply(req ApplyRequest) error {
	if req.TargetPath == "" {
		return fmt.Errorf("缺少目标文件")
	}
	if len(req.Edits) == 0 {
		return nil
	}
	if strings.EqualFold(filepath.Ext(req.TargetPath), ".csv") {
		return fmt.Errorf("CSV 暂不支持写回 (只读比对)")
	}
	if !strings.EqualFold(filepath.Ext(req.TargetPath), ".xlsx") &&
		!strings.EqualFold(filepath.Ext(req.TargetPath), ".xlsm") {
		return fmt.Errorf("仅支持 xlsx/xlsm 写回, 收到 %s", filepath.Ext(req.TargetPath))
	}
	edits := make([]xlsx.CellEdit, 0, len(req.Edits))
	for _, e := range req.Edits {
		edits = append(edits, xlsx.CellEdit{
			Row: e.Row, Col: e.Col,
			Value: e.Value, Formula: e.Formula, ClearFormula: e.ClearFormula,
		})
	}
	return xlsx.ApplyCellEdits(req.TargetPath, req.SheetName, edits)
}

// compareAsSheet 表格模式: 加载两表 (xlsx 与 CSV 可混用) → engine 比对 → DTO。
func compareAsSheet(req CompareRequest) (*CompareResult, error) {
	load := func(path string) (*engine.Sheet, error) {
		if strings.EqualFold(filepath.Ext(path), ".csv") {
			return xlsx.LoadCSV(path)
		}
		return xlsx.Load(path, req.SheetName)
	}
	left, err := load(req.LeftPath)
	if err != nil {
		return nil, err
	}
	right, err := load(req.RightPath)
	if err != nil {
		return nil, err
	}
	d, err := engine.CompareSheets(left, right, engine.AlignOptions{KeyColumns: req.KeyColumns})
	if err != nil {
		return nil, err
	}
	res := FromSheetDiff(d, left)
	res.Kind = "sheet"
	res.LeftName = displayName(req.LeftPath, left.Name)
	res.RightName = displayName(req.RightPath, right.Name)
	return res, nil
}

// compareAsText 文本模式: 读两文件分行 → 行 diff → DTO。
func compareAsText(req CompareRequest) (*CompareResult, error) {
	l, err := readLines(req.LeftPath)
	if err != nil {
		return nil, err
	}
	r, err := readLines(req.RightPath)
	if err != nil {
		return nil, err
	}
	res := FromLinePairs(textdiff.CompareLines(l, r), l, r)
	res.Kind = "text"
	res.LeftName = filepath.Base(req.LeftPath)
	res.RightName = filepath.Base(req.RightPath)
	return res, nil
}

// FromSheetDiff 把引擎比对结果转成行视图列表:
// 逐对齐对生成 RowView, 差异格按 PairIndex 分桶挂到所属行 (前端高亮定位)。
func FromSheetDiff(d *engine.SheetDiff, left *engine.Sheet) *CompareResult {
	// 差异分桶: PairIndex → 该行的差异格
	buckets := make(map[int][]CellView, len(d.Diffs))
	for _, c := range d.Diffs {
		k := "modified"
		switch c.Kind {
		case engine.DiffLeftOnly:
			k = "left"
		case engine.DiffRightOnly:
			k = "right"
		}
		buckets[c.PairIndex] = append(buckets[c.PairIndex], CellView{
			Col: c.Col, Kind: k, FormulaDiffers: c.FormulaDiffers,
		})
	}

	rows := make([]RowView, 0, len(d.Pairs))
	for i, p := range d.Pairs {
		rv := RowView{Diffs: buckets[i]}
		if p.Left != nil {
			rv.LeftRow = p.Left.Index
			rv.Left = cellValues(p.Left)
			rv.LeftFormulas = cellFormulas(p.Left)
		}
		if p.Right != nil {
			rv.RightRow = p.Right.Index
			rv.Right = cellValues(p.Right)
			rv.RightFormulas = cellFormulas(p.Right)
		}
		rows = append(rows, rv)
	}
	return &CompareResult{Headers: left.Headers, Stats: d.Stats, Rows: rows}
}

// FromLinePairs 把行 diff 结果转成行视图列表 (equal/单侧逐行展开, 保持顺序)。
func FromLinePairs(pairs []textdiff.LinePair, left, right []string) *CompareResult {
	lines := make([]LineView, 0, len(pairs))
	var st engine.DiffStats
	for _, p := range pairs {
		switch {
		case p.Left != 0 && p.Right != 0:
			st.MatchedRows++
			lines = append(lines, LineView{Kind: "equal", LeftIdx: p.Left, RightIdx: p.Right, Text: left[p.Left-1]})
		case p.Left != 0:
			st.LeftOnlyRows++
			lines = append(lines, LineView{Kind: "left", LeftIdx: p.Left, Text: left[p.Left-1]})
		default:
			st.RightOnlyRows++
			lines = append(lines, LineView{Kind: "right", RightIdx: p.Right, Text: right[p.Right-1]})
		}
	}
	return &CompareResult{Stats: st, Lines: lines}
}

// cellValues 行的显示值数组。
func cellValues(r *engine.Row) []string {
	out := make([]string, len(r.Cells))
	for i, c := range r.Cells {
		out[i] = c.Value
	}
	return out
}

// cellFormulas 行的公式数组 (与 cellValues 平行, 空串=非公式格)。
func cellFormulas(r *engine.Row) []string {
	out := make([]string, len(r.Cells))
	for i, c := range r.Cells {
		out[i] = c.Formula
	}
	return out
}

// isSheetExt 表格类扩展名 (与 CLI 分流一致)。
func isSheetExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xlsx", ".xlsm", ".csv":
		return true
	default:
		return false
	}
}

// displayName 表格模式展示名: 文件基名[表名]。
func displayName(path, sheet string) string {
	if sheet == "" {
		return filepath.Base(path)
	}
	return fmt.Sprintf("%s[%s]", filepath.Base(path), sheet)
}

// readLines 读文本文件分行: 容忍 CRLF 与结尾换行, 丢尾空行 (与 CLI 同语义)。
func readLines(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 %s: %w", path, err)
	}
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines, nil
}
