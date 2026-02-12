package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelExporter Excel 导出器
type ExcelExporter struct {
	excelPath string
	outputDir string
	sheets    []string // 要导出的 Sheet 列表，为空则导出全部
}

// NewExcelExporter 创建 Excel 导出器
func NewExcelExporter(excelPath, outputDir string) *ExcelExporter {
	return &ExcelExporter{
		excelPath: excelPath,
		outputDir: outputDir,
		sheets:    nil, // 默认导出全部 Sheet
	}
}

// SetSheets 设置要导出的 Sheet 列表
func (e *ExcelExporter) SetSheets(sheets []string) {
	e.sheets = sheets
}

// Export 执行导出
// 将 Excel 的每个 Sheet 导出为独立的 CSV 文件
// 文件组织方式: {outputDir}/{Excel名}/{Sheet名}.csv
func (e *ExcelExporter) Export() error {
	// 打开 Excel 文件
	file, err := excelize.OpenFile(e.excelPath)
	if err != nil {
		return fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer file.Close()

	// 获取要导出的 Sheet 列表
	sheetsToExport := e.getSheetsToExport(file)

	// 创建输出目录
	if err := e.createOutputDir(); err != nil {
		return err
	}

	// 导出每个 Sheet
	for _, sheetName := range sheetsToExport {
		if err := e.exportSheet(file, sheetName); err != nil {
			return fmt.Errorf("导出 Sheet '%s' 失败: %w", sheetName, err)
		}
	}

	return nil
}

// getSheetsToExport 获取要导出的 Sheet 列表
func (e *ExcelExporter) getSheetsToExport(file *excelize.File) []string {
	if len(e.sheets) > 0 {
		return e.sheets
	}
	return file.GetSheetList()
}

// createOutputDir 创建输出目录
func (e *ExcelExporter) createOutputDir() error {
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	return nil
}

// exportSheet 导出单个 Sheet
func (e *ExcelExporter) exportSheet(file *excelize.File, sheetName string) error {
	// 读取 Sheet 数据
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取 Sheet 数据失败: %w", err)
	}

	// 过滤掉版本行和说明行
	rows = e.filterMetadataRows(rows)

	// 创建 CSV 文件路径
	csvPath := e.getCSVPath(sheetName)

	// 确保文件目录存在
	csvDir := filepath.Dir(csvPath)
	if err := os.MkdirAll(csvDir, 0755); err != nil {
		return fmt.Errorf("创建 CSV 文件目录失败: %w", err)
	}

	// 创建 CSV 文件
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("创建 CSV 文件失败: %w", err)
	}
	defer csvFile.Close()

	// 写入 CSV 数据
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	for _, row := range rows {
		// 写入行数据
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入 CSV 数据失败: %w", err)
		}
	}

	return nil
}

// filterMetadataRows 过滤掉元数据行（版本行、说明行等）
func (e *ExcelExporter) filterMetadataRows(rows [][]string) [][]string {
	if len(rows) == 0 {
		return rows
	}

	// 检查第一行是否是版本行
	startRow := 0
	if len(rows[0]) > 0 && rows[0][0] == "__version__" {
		// 跳过版本行
		startRow = 1
		// 检查第二行是否是变更说明行
		if len(rows) > 1 && len(rows[1]) > 0 && rows[1][0] == "__changes__" {
			startRow = 2
		}
	}

	return rows[startRow:]
}

// getCSVPath 获取 CSV 文件路径
// 格式: {outputDir}/{Excel名}/{Sheet名}.csv
func (e *ExcelExporter) getCSVPath(sheetName string) string {
	excelName := e.getExcelName()
	return filepath.Join(e.outputDir, excelName, sheetName+".csv")
}

// getExcelName 获取 Excel 文件名（不含扩展名）
func (e *ExcelExporter) getExcelName() string {
	base := filepath.Base(e.excelPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// ExportToMemory 导出到内存（用于测试）
// 返回 map[Sheet名][][]string
func (e *ExcelExporter) ExportToMemory() (map[string][][]string, error) {
	file, err := excelize.OpenFile(e.excelPath)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer file.Close()

	result := make(map[string][][]string)
	sheetsToExport := e.getSheetsToExport(file)

	for _, sheetName := range sheetsToExport {
		rows, err := file.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("读取 Sheet '%s' 失败: %w", sheetName, err)
		}
		result[sheetName] = e.filterMetadataRows(rows)
	}

	return result, nil
}
