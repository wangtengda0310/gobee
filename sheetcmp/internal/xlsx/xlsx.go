// Package xlsx 负责 xlsx/CSV 文件与 engine.Sheet 之间的转换:
// 读取 (显示值与公式分离) 与单元格级原位写回 (保留原文件其余内容与样式)。
// 术语约定: "文件"指一个 xlsx/CSV 文件, "表"指其中一个 sheet。
package xlsx

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/wangtengda0310/gobee/sheetcmp/internal/engine"
)

// Load 读取 xlsx 文件的指定表 (sheet 为空时取第一个)。
// 流程: 打开文件 → 定位表 → 逐行读取显示值 → 非空单元格补读公式。
// 首行视为表头; 数据行 Index 为源表 1-based 绝对行号 (从 2 开始, 写回定位依赖)。
// 公式单元格的 Value 为缓存计算值 (无缓存时为空), 公式原文在 Cell.Formula。
// 性能注: 公式按单元格逐个查询, 万行大表有开销, 性能优化阶段再批量化。
func Load(path, sheet string) (*engine.Sheet, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开 %s: %w", filepath.Base(path), err)
	}
	defer f.Close()

	if sheet == "" {
		sheet, err = firstSheet(f, path)
		if err != nil {
			return nil, err
		}
	}
	return loadSheet(f, sheet)
}

// firstSheet 取文件第一个表名 (空文件报错)。
func firstSheet(f *excelize.File, path string) (string, error) {
	list := f.GetSheetList()
	if len(list) == 0 {
		return "", fmt.Errorf("%s 不含任何工作表", filepath.Base(path))
	}
	return list[0], nil
}

// loadSheet 从已打开的文件读取单个表: 行迭代器逐行取值, 表头行消费后数据行从 2 号行计数。
func loadSheet(f *excelize.File, sheet string) (*engine.Sheet, error) {
	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取表 %s: %w", sheet, err)
	}
	defer rows.Close()

	s := &engine.Sheet{Name: sheet}
	absRow := 0
	for rows.Next() {
		absRow++
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("读取表 %s 第 %d 行: %w", sheet, absRow, err)
		}
		if absRow == 1 {
			s.Headers = cols
			continue
		}
		r := engine.Row{Index: absRow, Cells: make([]engine.Cell, len(cols))}
		for i, v := range cols {
			r.Cells[i] = engine.Cell{Value: v}
		}
		// 逐格补公式: 不能按"值非空"过滤——工具生成的 xlsx 公式无缓存计算值
		// (excelize 不计算), 值空公式非空是常态。性能注: 逐格查询有开销,
		// 万行大表优化时再批量化 (如先解析 sheet XML 收集公式集合)。
		for i := range r.Cells {
			cell, err := excelize.CoordinatesToCellName(i+1, absRow)
			if err != nil {
				continue
			}
			if formula, err := f.GetCellFormula(sheet, cell); err == nil && formula != "" {
				r.Cells[i].Formula = formula
			}
		}
		s.Rows = append(s.Rows, r)
	}
	return s, nil
}

// CellEdit 单元格写回操作。Row/Col 为源表 1-based 绝对行列号 (与 engine.Row.Index 同一坐标系)。
// 三种互斥形态: Formula 非空写公式; ClearFormula=true 清除公式 (值不动, 对端无公式时同步用);
// 其余写值 (Value 为空串即清空该格)。
type CellEdit struct {
	Row, Col     int
	Value        string
	Formula      string
	ClearFormula bool
}

// ApplyCellEdits 把单元格修改原位写回文件的指定表 (sheet 为空时取第一个):
// 只改命中单元格, 其余单元格、样式、其他表全部保留。edits 为空时直接返回 (不动文件)。
func ApplyCellEdits(path, sheet string, edits []CellEdit) error {
	if len(edits) == 0 {
		return nil
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		return fmt.Errorf("打开 %s: %w", filepath.Base(path), err)
	}
	defer f.Close()

	if sheet == "" {
		sheet, err = firstSheet(f, path)
		if err != nil {
			return err
		}
	}
	for _, e := range edits {
		cell, err := excelize.CoordinatesToCellName(e.Col, e.Row)
		if err != nil {
			return fmt.Errorf("非法坐标 (%d,%d): %w", e.Row, e.Col, err)
		}
		switch {
		case e.Formula != "":
			if err := f.SetCellFormula(sheet, cell, e.Formula); err != nil {
				return fmt.Errorf("写公式 %s!%s: %w", sheet, cell, err)
			}
		case e.ClearFormula:
			// 空公式串在 excelize 语义里即清除该格公式
			if err := f.SetCellFormula(sheet, cell, ""); err != nil {
				return fmt.Errorf("清公式 %s!%s: %w", sheet, cell, err)
			}
		default:
			if err := f.SetCellValue(sheet, cell, e.Value); err != nil {
				return fmt.Errorf("写值 %s!%s: %w", sheet, cell, err)
			}
		}
	}
	return f.Save()
}

// LoadCSV 读取 CSV 文件为 engine.Sheet: 语义与 xlsx 一致 (首行表头, 数据行 Index 从 2 开始)。
func LoadCSV(path string) (*engine.Sheet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 %s: %w", filepath.Base(path), err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1 // 列数不齐不报错 (与引擎容错语义一致)
	s := &engine.Sheet{Name: filepath.Base(path)}
	absRow := 0
	for {
		cols, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取 %s 第 %d 行: %w", path, absRow+1, err)
		}
		absRow++
		if absRow == 1 {
			s.Headers = cols
			continue
		}
		row := engine.Row{Index: absRow, Cells: make([]engine.Cell, len(cols))}
		for i, v := range cols {
			row.Cells[i] = engine.Cell{Value: strings.TrimSpace(v)}
		}
		s.Rows = append(s.Rows, row)
	}
	return s, nil
}
