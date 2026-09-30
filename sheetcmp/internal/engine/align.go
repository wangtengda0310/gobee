package engine

// AlignPair 一对对齐的行: 单侧独有时另一侧为 nil
// (Left==nil 表示右侧独有, Right==nil 表示左侧独有, 双侧非 nil 表示配对)。
type AlignPair struct {
	Left  *Row
	Right *Row
}

// AlignOptions 行对齐配置。
type AlignOptions struct {
	// KeyColumns 关键列索引 (0-based, 相对 Headers)。
	// 为空 → 按行号位置配对 (第 i 行对第 i 行; 中间插行会引起连锁错位,
	//   需要处理插入/乱序时请指定关键列——这是关键列功能存在的核心理由)。
	// 非空 → 按关键列拼接值匹配, 与行序无关。
	KeyColumns []int
}

// AlignRows 行对齐, 返回保持可读顺序的对齐序列:
//
//	无关键列: 按行号位置配对 (i↔i), 多出的行落为单侧独有。
//	有关键列: 先输出双侧都有的配对与各自独有行, 按左表行序为主线;
//	  关键列值为空的行不参与键匹配, 左右两侧的空键行之间按行序依次兜底配对;
//	  同一 key 出现多行时, 同 key 行按各自出现顺序依次配对, 多出者落为单侧独有。
//
// 注意: 无关键列时不做"整行内容 LCS"——那会让任何有差异的行都落成
// 删除+新增对, 单元格级 diff 永远没有配对对象可用 (TDD 首轮实证的设计错误)。
func AlignRows(left, right *Sheet, opt AlignOptions) []AlignPair {
	var lr, rr []Row
	if left != nil {
		lr = left.Rows
	}
	if right != nil {
		rr = right.Rows
	}
	if len(opt.KeyColumns) == 0 {
		return alignByOrder(lr, rr)
	}
	return alignByKey(lr, rr, opt.KeyColumns)
}

// alignByOrder 行号位置配对: 第 i 行对第 i 行, 长表多出的尾部行落为单侧独有。
func alignByOrder(lr, rr []Row) []AlignPair {
	n := len(lr)
	if len(rr) < n {
		n = len(rr)
	}
	pairs := make([]AlignPair, 0, len(lr)+len(rr)-n)
	for k := 0; k < n; k++ {
		pairs = append(pairs, AlignPair{Left: &lr[k], Right: &rr[k]})
	}
	for i := n; i < len(lr); i++ {
		pairs = append(pairs, AlignPair{Left: &lr[i]})
	}
	for j := n; j < len(rr); j++ {
		pairs = append(pairs, AlignPair{Right: &rr[j]})
	}
	return pairs
}

// alignByKey 关键列对齐: 右表按 key 建行队列 (保持出现顺序), 左表按行序消费;
// 空 key 行不进队列, 由左右空键序列按行序依次兜底配对; 未消费的右行落为右独有。
func alignByKey(lr, rr []Row, keyCols []int) []AlignPair {
	rightQueues := make(map[string][]int, len(rr))
	var rightEmpty []int
	for j := range rr {
		k := RowKey(&rr[j], keyCols)
		if k == "" {
			rightEmpty = append(rightEmpty, j)
		} else {
			rightQueues[k] = append(rightQueues[k], j)
		}
	}

	usedRight := make([]bool, len(rr))
	pairs := make([]AlignPair, 0, len(lr))
	emptyCursor := 0 // 右空键行的兜底配对游标
	for i := range lr {
		k := RowKey(&lr[i], keyCols)
		switch {
		case k != "":
			if q := rightQueues[k]; len(q) > 0 {
				j := q[0]
				rightQueues[k] = q[1:]
				usedRight[j] = true
				pairs = append(pairs, AlignPair{Left: &lr[i], Right: &rr[j]})
			} else {
				pairs = append(pairs, AlignPair{Left: &lr[i]})
			}
		default:
			if emptyCursor < len(rightEmpty) {
				j := rightEmpty[emptyCursor]
				emptyCursor++
				usedRight[j] = true
				pairs = append(pairs, AlignPair{Left: &lr[i], Right: &rr[j]})
			} else {
				pairs = append(pairs, AlignPair{Left: &lr[i]})
			}
		}
	}
	// 右独有: 未被任何左行消费的右行, 按右表行序追加在末尾
	for j := range rr {
		if !usedRight[j] {
			pairs = append(pairs, AlignPair{Right: &rr[j]})
		}
	}
	return pairs
}
