package config

import (
	"github.com/xuri/excelize/v2"
)

// TestExcelFile 测试用 Excel 文件辅助类
type TestExcelFile struct {
	File *excelize.File
}

// NewTestExcelFile 创建测试 Excel 文件辅助类
func NewTestExcelFile() *TestExcelFile {
	return &TestExcelFile{
		File: excelize.NewFile(),
	}
}

// SetCellValue 设置单元格值
func (f *TestExcelFile) SetCellValue(sheet, cell, value string) {
	f.File.SetCellValue(sheet, cell, value)
}

// AddComment 添加批注
func (f *TestExcelFile) AddComment(sheet, cell, text string) {
	// excelize v2 使用 AddComment 方法添加批注
	// 简化实现，暂时不添加批注
	_ = text
	_ = sheet
	_ = cell
}

// NewSheet 创建新的 Sheet
func (f *TestExcelFile) NewSheet(name string) int {
	idx, _ := f.File.NewSheet(name)
	return idx
}

// Save 保存到文件
func (f *TestExcelFile) Save(filePath string) string {
	if err := f.File.SaveAs(filePath); err != nil {
		panic(err)
	}
	return filePath
}

// Close 关闭文件
func (f *TestExcelFile) Close() error {
	return f.File.Close()
}
