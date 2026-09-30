// sheetcmp-merge — sheetcmp 的命令行比对入口 (Git/UGit merge tool 形态)。
//
// 用法:
//
//	sheetcmp-merge [flags] <left-file> <right-file>
//
// 文件按扩展名分流: .xlsx/.xlsm/.csv 走表格模式 (engine 行对齐+单元格 diff),
// 其余按文本行 diff。退出码: 0=无差异, 1=有差异, 2=用法/错误。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wangtengda0310/gobee/sheetcmp"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/engine"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/textdiff"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/xlsx"
)

// moduleName go module 名, -h 输出中明示 (CLI 使用者溯源入口)。
const moduleName = "github.com/wangtengda0310/gobee/sheetcmp"

func main() {
	key := flag.String("key", "", "关键列索引, 逗号分隔 (0-based), 如 \"0,2\"; 缺省按行号位置配对 (仅表格模式)")
	sheet := flag.String("sheet", "", "xlsx 中的表名; 缺省第一个")
	doc := flag.Bool("doc", false, "打印内嵌的项目 README")
	flag.Usage = usage
	flag.Parse()

	if *doc {
		fmt.Print(sheetcmp.Readme())
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) != 2 {
		usage()
		os.Exit(2)
	}

	var hasDiff bool
	var err error
	if isSheetFile(args[0]) {
		hasDiff, err = runTableMode(args[0], args[1], *sheet, *key)
	} else {
		hasDiff, err = runTextMode(args[0], args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(2)
	}
	if hasDiff {
		os.Exit(1)
	}
	os.Exit(0)
}

func usage() {
	fmt.Fprintf(os.Stderr, "sheetcmp-merge — 表格/文本双向比对 (git merge tool 形态)\n"+
		"用法: sheetcmp-merge [flags] <left-file> <right-file>\n"+
		"  -key    关键列索引, 逗号分隔 (0-based), 如 \"0,2\"; 缺省按行号位置配对\n"+
		"  -sheet  xlsx 中的表名; 缺省第一个\n"+
		"  -doc    打印内嵌的项目 README\n"+
		"退出码: 0=无差异, 1=有差异, 2=用法/错误\n"+
		"go module: %s\n", moduleName)
}

// isSheetFile 按扩展名判断是否走表格模式。
func isSheetFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xlsx", ".xlsm", ".csv":
		return true
	default:
		return false
	}
}

// runTableMode 表格模式: 解析关键列 → 加载两文件 → engine 比对 → 输出统计与逐项差异。
// 返回是否存在差异。
func runTableMode(leftPath, rightPath, sheetName, key string) (bool, error) {
	opt := engine.AlignOptions{}
	if key != "" {
		for _, s := range strings.Split(key, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return false, fmt.Errorf("-key 非法 (%q): %w", key, err)
			}
			opt.KeyColumns = append(opt.KeyColumns, n)
		}
	}
	left, right, err := loadPair(leftPath, rightPath, sheetName)
	if err != nil {
		return false, err
	}
	d, err := engine.CompareSheets(left, right, opt)
	if err != nil {
		return false, err
	}

	fmt.Printf("表格比对: %s[%s] vs %s[%s]  (关键列: %s)\n",
		filepath.Base(leftPath), left.Name, filepath.Base(rightPath), right.Name, keyOrDefault(key))
	st := d.Stats
	fmt.Printf("  配对行 %d | 左独有行 %d | 右独有行 %d | 修改格 %d | 左独格 %d | 右独格 %d\n",
		st.MatchedRows, st.LeftOnlyRows, st.RightOnlyRows,
		st.ModifiedCells, st.LeftOnlyCells, st.RightOnlyCells)

	for _, c := range d.Diffs {
		p := d.Pairs[c.PairIndex]
		row := 0
		if p.Left != nil {
			row = p.Left.Index
		} else if p.Right != nil {
			row = p.Right.Index
		}
		col := colName(left.Headers, c.Col)
		switch {
		case c.FormulaDiffers:
			fmt.Printf("  [公式] R%d %s: 值相同公式不同\n", row, col)
		case c.Kind == engine.DiffLeftOnly:
			fmt.Printf("  [左独格] R%d %s: %q → (空)\n", row, col, c.Left)
		case c.Kind == engine.DiffRightOnly:
			fmt.Printf("  [右独格] R%d %s: (空) → %q\n", row, col, c.Right)
		default:
			fmt.Printf("  [修改] R%d %s: %q → %q\n", row, col, c.Left, c.Right)
		}
	}
	for _, p := range d.Pairs {
		if p.Left != nil && p.Right == nil {
			fmt.Printf("  [左独有行] R%d: %s\n", p.Left.Index, firstCell(p.Left))
		} else if p.Right != nil && p.Left == nil {
			fmt.Printf("  [右独有行] R%d: %s\n", p.Right.Index, firstCell(p.Right))
		}
	}
	return d.HasRowDifference(), nil
}

// runTextMode 文本模式: 读两文件分行 → 行级 diff → 输出统计与单侧行。
// 返回是否存在差异。
func runTextMode(leftPath, rightPath string) (bool, error) {
	l, err := readLines(leftPath)
	if err != nil {
		return false, err
	}
	r, err := readLines(rightPath)
	if err != nil {
		return false, err
	}
	pairs := textdiff.CompareLines(l, r)

	equal, del, ins := 0, 0, 0
	for _, p := range pairs {
		switch {
		case p.Left != 0 && p.Right != 0:
			equal++
		case p.Left != 0:
			del++
		default:
			ins++
		}
	}
	fmt.Printf("文本比对: %s vs %s\n  相同行 %d | 左独有(删) %d | 右独有(增) %d\n",
		filepath.Base(leftPath), filepath.Base(rightPath), equal, del, ins)
	for _, p := range pairs {
		switch {
		case p.Left != 0 && p.Right == 0:
			fmt.Printf("  [-] L%d: %s\n", p.Left, l[p.Left-1])
		case p.Right != 0 && p.Left == 0:
			fmt.Printf("  [+] R%d: %s\n", p.Right, r[p.Right-1])
		}
	}
	return del > 0 || ins > 0, nil
}

// loadPair 按扩展名分别加载左右文件 (xlsx 与 CSV 可混用, 统一进 engine.Sheet)。
func loadPair(leftPath, rightPath, sheetName string) (*engine.Sheet, *engine.Sheet, error) {
	load := func(path string) (*engine.Sheet, error) {
		if strings.EqualFold(filepath.Ext(path), ".csv") {
			return xlsx.LoadCSV(path)
		}
		return xlsx.Load(path, sheetName)
	}
	l, err := load(leftPath)
	if err != nil {
		return nil, nil, err
	}
	r, err := load(rightPath)
	if err != nil {
		return nil, nil, err
	}
	return l, r, nil
}

// readLines 读文本文件分行: 容忍 CRLF 与结尾换行, 丢尾空行。
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

// keyOrDefault 关键列参数的展示形式。
func keyOrDefault(key string) string {
	if key == "" {
		return "(行号位置配对)"
	}
	return key
}

// colName 列展示名: 有表头用表头, 否则 C1 形式列号。
func colName(headers []string, col int) string {
	if col < len(headers) && headers[col] != "" {
		return headers[col]
	}
	return fmt.Sprintf("C%d", col+1)
}

// firstCell 行首格内容 (独有行的摘要展示; 空行显示占位符)。
func firstCell(r *engine.Row) string {
	if len(r.Cells) > 0 && r.Cells[0].Value != "" {
		return r.Cells[0].Value
	}
	return "(空)"
}
