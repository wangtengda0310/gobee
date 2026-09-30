package engine

// DiffKind 单元格差异类型。
type DiffKind int

const (
	DiffModified  DiffKind = iota // 双侧有值且不同 (值或公式)
	DiffLeftOnly                  // 仅左侧有值 (右侧空)
	DiffRightOnly                 // 仅右侧有值 (左侧空)
)

// CellDiff 单元格级差异, PairIndex 指向 CompareSheets 结果中对齐序列的下标,
// 前端高亮与行同步都靠它回溯到具体行列。
type CellDiff struct {
	PairIndex int // AlignPair 在结果 Pairs 中的下标
	Col       int // 列索引 (0-based)
	Kind      DiffKind
	Left      string // 左侧值 (显示用; 公式差异时为值, 公式差异信息见 FormulaDiffers)
	Right     string
	// FormulaDiffers: 值相同但公式不同的静默差异 (true 时 Kind 仍为 DiffModified,
	// 界面须用不同样式强调——这类差异肉眼看值发现不了)
	FormulaDiffers bool
}

// DiffStats 差异统计 (驱动界面汇总栏与"有差异才允许写回"的门槛判断)。
type DiffStats struct {
	MatchedRows    int // 配对行数 (含内容有差异的)
	LeftOnlyRows   int // 左独有行数
	RightOnlyRows  int // 右独有行数
	ModifiedCells  int
	LeftOnlyCells  int
	RightOnlyCells int
}

// SheetDiff 全表比对结果。
type SheetDiff struct {
	Pairs []AlignPair
	Diffs []CellDiff
	Stats DiffStats
}

// HasRowDifference 是否存在任何行级差异 (独有行或单元格差异), 写回门槛用。
func (d *SheetDiff) HasRowDifference() bool {
	return len(d.Diffs) > 0 || d.Stats.LeftOnlyRows > 0 || d.Stats.RightOnlyRows > 0
}

// CompareSheets 全表比对: 先 AlignRows 对齐, 再对配对行逐单元格 diff。
// 列数不齐时短行按缺省空单元格处理 (不视为错误);
// 单元格空值以显示值为准 (值空公式非空由 FormulaDiffers 分支兜住)。
func CompareSheets(left, right *Sheet, opt AlignOptions) (*SheetDiff, error) {
	d := &SheetDiff{Pairs: AlignRows(left, right, opt)}
	for pi := range d.Pairs {
		p := &d.Pairs[pi]
		switch {
		case p.Left != nil && p.Right != nil:
			d.Stats.MatchedRows++
			cols := len(p.Left.Cells)
			if len(p.Right.Cells) > cols {
				cols = len(p.Right.Cells)
			}
			for c := 0; c < cols; c++ {
				var lc, rc Cell
				if c < len(p.Left.Cells) {
					lc = p.Left.Cells[c]
				}
				if c < len(p.Right.Cells) {
					rc = p.Right.Cells[c]
				}
				if lc.Value == rc.Value && lc.Formula == rc.Formula {
					continue // 值与公式都一致, 无差异
				}
				cd := CellDiff{PairIndex: pi, Col: c, Left: lc.Value, Right: rc.Value}
				switch {
				case lc.Value != rc.Value && rc.Value == "":
					cd.Kind = DiffLeftOnly
					d.Stats.LeftOnlyCells++
				case lc.Value != rc.Value && lc.Value == "":
					cd.Kind = DiffRightOnly
					d.Stats.RightOnlyCells++
				case lc.Value != rc.Value:
					cd.Kind = DiffModified
					d.Stats.ModifiedCells++
				default:
					// 值相同公式不同: 仍是 DiffModified, 标 FormulaDiffers 供界面特别强调
					cd.Kind = DiffModified
					cd.FormulaDiffers = true
					d.Stats.ModifiedCells++
				}
				d.Diffs = append(d.Diffs, cd)
			}
		case p.Left != nil:
			d.Stats.LeftOnlyRows++
		default:
			d.Stats.RightOnlyRows++
		}
	}
	return d, nil
}
