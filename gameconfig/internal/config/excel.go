package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelReader Excel 读取器
type ExcelReader struct {
	filePath string
	file     *excelize.File
}

// NewExcelReader 创建 Excel 读取器
func NewExcelReader(filePath string) (*ExcelReader, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, filePath)
	}

	// 打开 Excel 文件
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}

	return &ExcelReader{
		filePath: filePath,
		file:     file,
	}, nil
}

// Path 获取文件路径
func (r *ExcelReader) Path() string {
	return r.filePath
}

// Close 关闭读取器
func (r *ExcelReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// GetSheetNames 获取所有 Sheet 名称
func (r *ExcelReader) GetSheetNames() []string {
	return r.file.GetSheetList()
}

// ReadSheet 读取指定 Sheet 的数据
// 返回二维字符串数组，第一行是表头
func (r *ExcelReader) ReadSheet(sheetName string) ([][]string, error) {
	// 检查 Sheet 是否存在
	sheets := r.GetSheetNames()
	exists := false
	for _, name := range sheets {
		if name == sheetName {
			exists = true
			break
		}
	}
	if !exists {
		return nil, fmt.Errorf("%w: Sheet '%s' 不存在", ErrSheetNotFound, sheetName)
	}

	// 获取 Sheet 的所有行
	rows, err := r.file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取 Sheet 失败: %w", err)
	}

	return rows, nil
}

// ReadAllSheets 读取所有 Sheet 的数据
// 返回 map[Sheet名][][]string
func (r *ExcelReader) ReadAllSheets() (map[string][][]string, error) {
	result := make(map[string][][]string)
	sheets := r.GetSheetNames()

	for _, sheetName := range sheets {
		data, err := r.ReadSheet(sheetName)
		if err != nil {
			return nil, err
		}
		result[sheetName] = data
	}

	return result, nil
}

// ReadCell 读取指定单元格的值
// 支持公式计算后的值
func (r *ExcelReader) ReadCell(sheetName, cell string) (string, error) {
	// 使用 RawCellValue=false 获取公式计算后的值
	value, err := r.file.GetCellValue(sheetName, cell, excelize.Options{
		RawCellValue: false,
	})
	if err != nil {
		return "", fmt.Errorf("读取单元格失败: %w", err)
	}

	return value, nil
}

// SetCellValue 设置单元格的值（用于测试）
func (r *ExcelReader) SetCellValue(sheetName, cell, value string) error {
	return r.file.SetCellValue(sheetName, cell, value)
}

// GetVersion 获取 Sheet 的版本号
// 版本号格式：第一行第一列为 "__version__"，第二列为版本号
// 例如: A1="__version__", B1="2"
func (r *ExcelReader) GetVersion(sheetName string) (int, error) {
	// 读取第一行
	rows, err := r.file.GetRows(sheetName)
	if err != nil || len(rows) == 0 {
		return 0, nil // 没有数据，返回默认版本 0
	}

	firstRow := rows[0]
	if len(firstRow) < 2 {
		return 0, nil // 第一行数据不足，返回默认版本 0
	}

	// 检查第一列是否是 __version__
	if firstRow[0] != "__version__" {
		return 0, nil // 没有版本信息，返回默认版本 0
	}

	// 解析版本号
	versionStr := strings.TrimSpace(firstRow[1])
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, fmt.Errorf("%w: 无法解析版本号 '%s': %w", ErrInvalidVersion, versionStr, err)
	}

	return version, nil
}

// GetCellValue 获取单元格值（带格式）
func (r *ExcelReader) GetCellValue(sheetName, cell string, raw bool) (string, error) {
	value, err := r.file.GetCellValue(sheetName, cell, excelize.Options{
		RawCellValue: raw,
	})
	if err != nil {
		return "", fmt.Errorf("获取单元格值失败: %w", err)
	}
	return value, nil
}

// GetMergedCells 获取合并单元格信息
func (r *ExcelReader) GetMergedCells(sheetName string) ([]string, error) {
	mergedCells, err := r.file.GetMergeCells(sheetName)
	if err != nil {
		return nil, fmt.Errorf("获取合并单元格失败: %w", err)
	}

	result := make([]string, len(mergedCells))
	for i, mc := range mergedCells {
		result[i] = mc.GetStartAxis() + ":" + mc.GetEndAxis()
	}
	return result, nil
}

// GetSheetName 获取 Excel 文件名（不含扩展名）作为默认 Sheet 名
func (r *ExcelReader) GetSheetName() string {
	base := filepath.Base(r.filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// GetRowCount 获取指定 Sheet 的行数
func (r *ExcelReader) GetRowCount(sheetName string) (int, error) {
	rows, err := r.file.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("获取行数失败: %w", err)
	}
	return len(rows), nil
}

// GetColCount 获取指定 Sheet 的列数
func (r *ExcelReader) GetColCount(sheetName string) (int, error) {
	rows, err := r.file.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("获取列数失败: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}

	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	return maxCols, nil
}

// CellToCoord 将单元格坐标转换为行列索引
// 例如: "A1" -> (0, 0), "B2" -> (1, 1)
func CellToCoord(cell string) (col, row int, err error) {
	return excelize.CellNameToCoordinates(cell)
}

// CoordToCell 将行列索引转换为单元格坐标
// 例如: (0, 0) -> "A1", (1, 1) -> "B2"
func CoordToCell(col, row int) string {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	return cell
}
