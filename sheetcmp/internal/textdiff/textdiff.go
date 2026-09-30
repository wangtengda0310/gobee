// Package textdiff 提供文本按行 diff: 代码/JSON/CSV 等普通文本的行级对齐,
// 输出结构与 engine 的行对齐 (配对/单侧独有) 同构, 供前端统一渲染。
package textdiff

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// LinePair 对齐的行对。Left/Right 为各自文本中的 1-based 行号, 0 表示该侧无对应行
// (Left==0 右独有/插入, Right==0 左独有/删除, 双非 0 相等配对)。
type LinePair struct {
	Left  int
	Right int
}

// CompareLines 按行对齐两段文本: 内容 LCS 语义 (经 Myers 算法),
// 相等行成对, 修改行拆成一删一增两个单侧对 (不做行内字符级拆分, 保持与表格模式一致的行级颗粒度)。
// 返回序列保持双方行序, 可直接驱动双栏渲染。
// 注意: 底层把每行映射为一个 rune, 超 RuneMax 数量级唯一行的极端文件不适用 (常规源码/配置远达不到)。
func CompareLines(left, right []string) []LinePair {
	dmp := diffmatchpatch.New()
	r1, r2, _ := dmp.DiffLinesToRunes(joinLines(left), joinLines(right))
	diffs := dmp.DiffMainRunes(r1, r2, false)

	pairs := make([]LinePair, 0, len(left)+len(right))
	li, ri := 1, 1 // 双侧行号游标 (1-based)
	for _, d := range diffs {
		for k := 0; k < len([]rune(d.Text)); k++ { // 每个 rune 恰好一行
			switch d.Type {
			case diffmatchpatch.DiffEqual:
				pairs = append(pairs, LinePair{Left: li, Right: ri})
				li++
				ri++
			case diffmatchpatch.DiffDelete:
				pairs = append(pairs, LinePair{Left: li})
				li++
			case diffmatchpatch.DiffInsert:
				pairs = append(pairs, LinePair{Right: ri})
				ri++
			}
		}
	}
	return pairs
}

// joinLines 行数组拼回带换行符的文本 (DiffLinesToRunes 要求以行结尾的文本)。
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
