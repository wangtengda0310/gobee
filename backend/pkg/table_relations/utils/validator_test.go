package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestValidateTableExists 测试表存在性验证
func TestValidateTableExists(t *testing.T) {
	tests := []struct {
		name        string
		sheetMap    map[string]*excelize.File
		tableName   string
		wantExist   bool
		wantErr     bool
		errContains string
	}{
		{
			name: "表存在（精确匹配）",
			sheetMap: map[string]*excelize.File{
				"Activity": excelize.NewFile(),
			},
			tableName: "Activity",
			wantExist: true,
			wantErr:   false,
		},
		{
			name: "表存在（后缀匹配）",
			sheetMap: map[string]*excelize.File{
				"活动表|Activity": excelize.NewFile(),
			},
			tableName: "Activity",
			wantExist: true,
			wantErr:   false,
		},
		{
			name:        "表不存在",
			sheetMap:    map[string]*excelize.File{},
			tableName:   "NonExistent",
			wantExist:   false,
			wantErr:     true,
			errContains: "表 'NonExistent' 不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, actualName, err := ValidateTableExists(tt.sheetMap, tt.tableName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantExist, exists)
			assert.NotEmpty(t, actualName)
		})
	}
}

// TestValidateFieldExists 测试字段存在性验证
func TestValidateFieldExists(t *testing.T) {
	tests := []struct {
		name        string
		sheetMap    map[string]*excelize.File
		tableName   string
		fieldName   string
		wantExist   bool
		wantErr     bool
		errContains string
	}{
		{
			name:      "普通字段存在",
			sheetMap:  createMockActivitySheet(),
			tableName: "Activity",
			fieldName: "ActivityType",
			wantExist: true,
			wantErr:   false,
		},
		{
			name:      "数组字段存在",
			sheetMap:  createMockActivitySheet(),
			tableName: "Activity",
			fieldName: "CustomParma[0]",
			wantExist: true,
			wantErr:   false,
		},
		{
			name:        "字段不存在",
			sheetMap:    createMockActivitySheet(),
			tableName:   "Activity",
			fieldName:   "NonExistent",
			wantExist:   false,
			wantErr:     true,
			errContains: "字段 'NonExistent' 在表 'Activity' 中不存在",
		},
		{
			name:        "数组索引越界",
			sheetMap:    createMockActivitySheet(),
			tableName:   "Activity",
			fieldName:   "CustomParma[10]",
			wantExist:   false,
			wantErr:     true,
			errContains: "索引 10 超出数组范围",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := ValidateFieldExists(tt.sheetMap, tt.tableName, tt.fieldName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantExist, exists)
		})
	}
}

// TestParseFieldName 测试字段名解析
func TestParseFieldName(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		wantBase    string
		wantIndex   *int
		wantErr     bool
		errContains string
	}{
		{
			name:      "普通字段名",
			fieldName: "ActivityType",
			wantBase:  "ActivityType",
			wantIndex: nil,
			wantErr:   false,
		},
		{
			name:      "数组字段名（索引0）",
			fieldName: "CustomParma[0]",
			wantBase:  "CustomParma",
			wantIndex: intPtr(0),
			wantErr:   false,
		},
		{
			name:      "数组字段名（索引10）",
			fieldName: "CustomParma[10]",
			wantBase:  "CustomParma",
			wantIndex: intPtr(10),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseName, arrayIndex, err := parseFieldName(tt.fieldName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, arrayIndex)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantBase, baseName)
			assert.Equal(t, tt.wantIndex, arrayIndex)
		})
	}
}

// TestParseArrayValue 测试数组值解析
func TestParseArrayValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "逗号分隔",
			input: "1,2,3,4,5",
			want:  []string{"1", "2", "3", "4", "5"},
		},
		{
			name:  "分号分隔",
			input: "1;2;3;4;5",
			want:  []string{"1", "2", "3", "4", "5"},
		},
		{
			name:  "单个值",
			input: "1001",
			want:  []string{"1001"},
		},
		{
			name:  "空字符串",
			input: "",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArrayValue(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

// createMockActivitySheet 创建模拟的活动表
func createMockActivitySheet() map[string]*excelize.File {
	f := excelize.NewFile()
	sheetName := "活动表|Activity"
	f.NewSheet(sheetName)

	// 第0行：中文名称
	f.SetCellValue(sheetName, "A1", "ID")
	f.SetCellValue(sheetName, "B1", "活动类型")
	f.SetCellValue(sheetName, "C1", "自定义参数")

	// 第1行：类型定义
	f.SetCellValue(sheetName, "A2", "int")
	f.SetCellValue(sheetName, "B2", "string")
	f.SetCellValue(sheetName, "C2", "string")

	// 第2行：属性名称
	f.SetCellValue(sheetName, "A3", "Id")
	f.SetCellValue(sheetName, "B3", "ActivityType")
	f.SetCellValue(sheetName, "C3", "CustomParma")

	// 第3行：导出标识
	for j := 0; j < 3; j++ {
		f.SetCellValue(sheetName, getCellName(j, 3), "server")
	}

	// 数据行（从第4行开始）
	f.SetCellValue(sheetName, "A5", "3001")
	f.SetCellValue(sheetName, "B5", "BossBattle")
	f.SetCellValue(sheetName, "C5", "100,200,300")

	f.SetCellValue(sheetName, "A6", "3002")
	f.SetCellValue(sheetName, "B6", "GuildWar")
	f.SetCellValue(sheetName, "C6", "400,500,600")

	return map[string]*excelize.File{
		"活动表|Activity": f,
	}
}

// intPtr 返回整数指针的辅助函数
func intPtr(i int) *int {
	return &i
}
