package utils

import (
	"strconv"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestLoadSheetData 测试表数据加载功能
func TestLoadSheetData(t *testing.T) {
	tests := []struct {
		name        string
		setupSheet  map[string]*excelize.File
		sheetName   string
		wantCols    int
		wantExist   bool
		wantErr     bool
		errContains string
	}{
		{
			name: "精确匹配表名",
			setupSheet: map[string]*excelize.File{
				"Activity": createMockSheet("Activity"),
			},
			sheetName: "Activity",
			wantCols:  2,
			wantExist: true,
			wantErr:   false,
		},
		{
			name: "后缀匹配表名",
			setupSheet: map[string]*excelize.File{
				"活动表|Activity": createMockSheet("活动表|Activity"),
			},
			sheetName: "Activity",
			wantCols:  2,
			wantExist: true,
			wantErr:   false,
		},
		{
			name:        "表不存在",
			setupSheet:  map[string]*excelize.File{},
			sheetName:   "NonExistent",
			wantCols:    0,
			wantExist:   false,
			wantErr:     true,
			errContains: "表 'NonExistent' 不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, actualName, err := LoadSheetData(tt.setupSheet, tt.sheetName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCols, len(cols))
			assert.NotEmpty(t, actualName)
		})
	}
}

// TestGetStartRowIndex 测试数据起始行索引
func TestGetStartRowIndex(t *testing.T) {
	startRow := GetStartRowIndex()
	assert.Equal(t, excelio.MJS_FIXED_ROWS_NUM, startRow)
}

// TestGetDataEndIndex 测试数据结束行索引
func TestGetDataEndIndex(t *testing.T) {
	tests := []struct {
		name        string
		cols        [][]string
		idColIdx    int
		startRow    int
		expectedEnd int
	}{
		{
			name: "从空开始",
			cols: [][]string{
				{"Id", "Name"},
				{"", ""}, // 索引0：表头
				{"", ""}, // 索引1：空1
				{"", ""}, // 索引2：空2
				{"", ""}, // 索引3：空3
			},
			idColIdx:    0,
			startRow:    1,
			expectedEnd: 2, // 立即遇到3个空，返回起始位置（1+1）
		},
		{
			name: "数据后有空行",
			cols: [][]string{
				{"Id", "Name"},
				{"", ""},       // 索引0：表头
				{"1", "Test1"}, // 索引1：数据1
				{"", ""},       // 索引2：空1
				{"", ""},       // 索引3：空2
				{"", ""},       // 索引4：空3
			},
			idColIdx:    0,
			startRow:    1,
			expectedEnd: 2, // 遇到3个空，返回第一个空的位置（2）
		},
		{
			name: "没有空行",
			cols: [][]string{
				{"Id", "Name"},
				{"", ""},       // 索引0：表头
				{"1", "Test1"}, // 索引1：数据1
				{"2", "Test2"}, // 索引2：数据2
				{"3", "Test3"}, // 索引3：数据3
			},
			idColIdx:    0,
			startRow:    1,
			expectedEnd: 2, // 返回第一个空的位置（索引2）
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endRow := GetDataEndIndex(tt.cols, tt.idColIdx, tt.startRow)
			assert.Equal(t, tt.expectedEnd, endRow)
		})
	}
}

// createMockSheet 创建模拟的 Excel 表
//
// 用于测试，创建一个简单的表结构
func createMockSheet(sheetName string) *excelize.File {
	f := excelize.NewFile()
	f.NewSheet(sheetName)

	// 创建固定表头行（MJS 格式）
	// 第0行：中文名称
	f.SetCellValue(sheetName, "A1", "ID")
	f.SetCellValue(sheetName, "B1", "名称")

	// 第1行：类型定义
	f.SetCellValue(sheetName, "A2", "int")
	f.SetCellValue(sheetName, "B2", "string")

	// 第2行：属性名称
	f.SetCellValue(sheetName, "A3", "Id")
	f.SetCellValue(sheetName, "B3", "Name")

	// 第3行：导出标识
	f.SetCellValue(sheetName, "A4", "server")
	f.SetCellValue(sheetName, "B4", "server")

	// 第4行开始：数据
	for i := 0; i < 10; i++ {
		f.SetCellValue(sheetName, getCellName(0, i+4), i+1)                      // Id
		f.SetCellValue(sheetName, getCellName(1, i+4), "Test"+strconv.Itoa(i+1)) // Name
	}

	return f
}

// getCellName 将列索引和行索引转换为单元格名称
//
// 例如：colIdx=0, rowIdx=0 → "A1"
// 例如：colIdx=1, rowIdx=4 → "B5"
func getCellName(colIdx, rowIdx int) string {
	colName := ""
	for colIdx >= 0 {
		remainder := colIdx % 26
		colName = string(rune('A'+remainder)) + colName
		colIdx = colIdx/26 - 1
	}
	return colName + strconv.Itoa(rowIdx+1)
}
