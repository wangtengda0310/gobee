// Package engine 是表格双向比对的纯逻辑核心: 行对齐 + 单元格差异。
// 本包不依赖文件格式与 GUI——xlsx/CSV 加载层把数据转成这里的值对象,
// GUI/CLI/merge 三种形态共用同一份对齐与 diff 结果。
package engine

import "strings"

// Cell 单元格: 显示值与公式分离存储, 二者都参与比对
// (公式不同即使计算值相同也算差异——静默改公式是配置表的典型事故)。
type Cell struct {
	Value   string // 显示值 (单元格渲染后的文本)
	Formula string // 公式原文; 非公式单元格为空串
}

// Row 一行数据。Index 是该行在源表中的 1-based 行号,
// 行同步写回时靠它定位目标表的实际位置。
type Row struct {
	Index int
	Cells []Cell
}

// Sheet 工作表抽象: 任意来源 (xlsx sheet / CSV) 统一加载成此结构。
// Headers 为首行列名 (无表头的来源填 Col1..ColN 占位)。
type Sheet struct {
	Name    string
	Headers []string
	Rows    []Row
}

// RowKey 提取一行在关键列上的拼接键: 关键列越界时该列按空值参与。
// 所有参与列的值均为空 → 返回空串 (调用方以此识别"空键行", 走兜底配对)。
// 用 \x1f (单元分隔符) 拼接, 避免多列值恰好拼接出歧义键。
func RowKey(r *Row, keyColumns []int) string {
	if r == nil {
		return ""
	}
	parts := make([]string, 0, len(keyColumns))
	any := false
	for _, c := range keyColumns {
		v := ""
		if c >= 0 && c < len(r.Cells) {
			v = r.Cells[c].Value
		}
		if v != "" {
			any = true
		}
		parts = append(parts, v)
	}
	if !any {
		return ""
	}
	return strings.Join(parts, "\x1f")
}
