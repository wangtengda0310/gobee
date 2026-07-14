package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestParseDrawSkinData 测试皮肤抽奖表数据解析
func TestParseDrawSkinData(t *testing.T) {
	tests := []struct {
		name        string
		setupSheet  map[string]*excelize.File
		wantCount   int
		wantFirstId int
		wantErr     bool
		errContains string
	}{
		{
			name:        "正常解析",
			setupSheet:  createMockDrawSkinSheet(),
			wantCount:   2,
			wantFirstId: 1001,
			wantErr:     false,
		},
		{
			name:        "表不存在",
			setupSheet:  map[string]*excelize.File{},
			wantCount:   0,
			wantFirstId: 0,
			wantErr:     true,
			errContains: "表 'DrawSkin' 不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDrawSkinData(tt.setupSheet)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(*result))

			if len(*result) > 0 {
				assert.Equal(t, tt.wantFirstId, (*result)[0].Id)
			}
		})
	}
}

// TestParseDropRuleData 测试掉落规则表数据解析
func TestParseDropRuleData(t *testing.T) {
	tests := []struct {
		name        string
		setupSheet  map[string]*excelize.File
		wantCount   int
		wantFirstId int
		wantErr     bool
		errContains string
	}{
		{
			name:        "正常解析",
			setupSheet:  createMockDropRuleSheet(),
			wantCount:   2,
			wantFirstId: 2001,
			wantErr:     false,
		},
		{
			name:        "表不存在",
			setupSheet:  map[string]*excelize.File{},
			wantCount:   0,
			wantFirstId: 0,
			wantErr:     true,
			errContains: "表 'DropRule' 不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDropRuleData(tt.setupSheet)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(*result))

			if len(*result) > 0 {
				assert.Equal(t, tt.wantFirstId, (*result)[0].Id)
			}
		})
	}
}

// TestParseItemCost 测试道具消耗解析
func TestParseItemCost(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantIds []int
	}{
		{
			name:    "单个道具",
			input:   "{1001;10}",
			wantLen: 1,
			wantIds: []int{1001},
		},
		{
			name:    "多个道具",
			input:   "{1001;10},{1002;20},{1003;30}",
			wantLen: 3,
			wantIds: []int{1001, 1002, 1003},
		},
		{
			name:    "空字符串",
			input:   "",
			wantLen: 0,
			wantIds: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试需要的正则表达式
			reg := regexp.MustCompile(`\{(\d+);\d+}`)
			result := parseItemCost(tt.input, reg)

			assert.Equal(t, tt.wantLen, len(result))

			if len(result) > 0 {
				for i, itemId := range tt.wantIds {
					assert.Equal(t, itemId, result[i].ItemId)
				}
			}
		})
	}
}

// TestParseIntArray 测试整数字符串解析
func TestParseIntArray(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []int
	}{
		{
			name:  "正常整数数组",
			input: "1,2,3,4,5",
			want:  []int{1, 2, 3, 4, 5},
		},
		{
			name:  "带空格的整数数组",
			input: "1, 2, 3, 4, 5",
			want:  []int{1, 2, 3, 4, 5},
		},
		{
			name:  "空字符串",
			input: "",
			want:  []int{},
		},
		{
			name:  "只有空格",
			input: "   ",
			want:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntArray(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

// createMockDrawSkinSheet 创建模拟的皮肤抽奖表
func createMockDrawSkinSheet() map[string]*excelize.File {
	f := excelize.NewFile()
	sheetName := "皮肤抽奖|DrawSkin"
	f.NewSheet(sheetName)

	// 第0行：中文名称
	f.SetCellValue(sheetName, "A1", "ID")
	f.SetCellValue(sheetName, "B1", "名称")
	f.SetCellValue(sheetName, "C1", "单抽掉落规则")
	f.SetCellValue(sheetName, "D1", "十抽掉落规则")
	f.SetCellValue(sheetName, "E1", "单抽消耗")
	f.SetCellValue(sheetName, "F1", "十抽消耗")
	f.SetCellValue(sheetName, "G1", "活动ID")
	f.SetCellValue(sheetName, "H1", "开始时间")
	f.SetCellValue(sheetName, "I1", "结束时间")

	// 第1行：类型定义
	f.SetCellValue(sheetName, "A2", "int")
	f.SetCellValue(sheetName, "B2", "string")
	f.SetCellValue(sheetName, "C2", "int")
	f.SetCellValue(sheetName, "D2", "int")
	f.SetCellValue(sheetName, "E2", "string")
	f.SetCellValue(sheetName, "F2", "string")
	f.SetCellValue(sheetName, "G2", "int")
	f.SetCellValue(sheetName, "H2", "string")
	f.SetCellValue(sheetName, "I2", "string")

	// 第2行：属性名称
	f.SetCellValue(sheetName, "A3", "Id")
	f.SetCellValue(sheetName, "B3", "Name")
	f.SetCellValue(sheetName, "C3", "OnceDropRule")
	f.SetCellValue(sheetName, "D3", "TenDropRule")
	f.SetCellValue(sheetName, "E3", "OnceItemCost")
	f.SetCellValue(sheetName, "F3", "TenItemCost")
	f.SetCellValue(sheetName, "G3", "ActivityId")
	f.SetCellValue(sheetName, "H3", "StartTime")
	f.SetCellValue(sheetName, "I3", "EndTime")

	// 第3行：导出标识
	for j := 0; j < 9; j++ {
		f.SetCellValue(sheetName, getCellName(j, 3), "server")
	}

	// 数据行（从第4行开始）
	// 第一行数据
	f.SetCellValue(sheetName, "A5", "1001")
	f.SetCellValue(sheetName, "B5", "皮肤抽奖1")
	f.SetCellValue(sheetName, "C5", "2001")
	f.SetCellValue(sheetName, "D5", "2002")
	f.SetCellValue(sheetName, "E5", "{1001;100}")
	f.SetCellValue(sheetName, "F5", "{1001;1000}")
	f.SetCellValue(sheetName, "G5", "3001")
	f.SetCellValue(sheetName, "H5", "2026-01-01 00:00:00")
	f.SetCellValue(sheetName, "I5", "2026-12-31 23:59:59")

	// 第二行数据
	f.SetCellValue(sheetName, "A6", "1002")
	f.SetCellValue(sheetName, "B6", "皮肤抽奖2")
	f.SetCellValue(sheetName, "C6", "2003")
	f.SetCellValue(sheetName, "D6", "2004")
	f.SetCellValue(sheetName, "E6", "{1002;50}")
	f.SetCellValue(sheetName, "F6", "{1002;500}")
	f.SetCellValue(sheetName, "G6", "3002")
	f.SetCellValue(sheetName, "H6", "2026-02-01 00:00:00")
	f.SetCellValue(sheetName, "I6", "2026-06-30 23:59:59")

	return map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": f,
	}
}

// createMockDropRuleSheet 创建模拟的掉落规则表
func createMockDropRuleSheet() map[string]*excelize.File {
	f := excelize.NewFile()
	sheetName := "掉落规则表|DropRule"
	f.NewSheet(sheetName)

	// 第0行：中文名称
	f.SetCellValue(sheetName, "A1", "ID")
	f.SetCellValue(sheetName, "B1", "名称")
	f.SetCellValue(sheetName, "C1", "次数")
	f.SetCellValue(sheetName, "D1", "掉落组")
	f.SetCellValue(sheetName, "E1", "小保底次数")
	f.SetCellValue(sheetName, "F1", "小保底组")
	f.SetCellValue(sheetName, "G1", "大保底次数")
	f.SetCellValue(sheetName, "H1", "大保底组")

	// 第1行：类型定义
	f.SetCellValue(sheetName, "A2", "int")
	f.SetCellValue(sheetName, "B2", "string")
	f.SetCellValue(sheetName, "C2", "int")
	f.SetCellValue(sheetName, "D2", "string")
	f.SetCellValue(sheetName, "E2", "int")
	f.SetCellValue(sheetName, "F2", "string")
	f.SetCellValue(sheetName, "G2", "int")
	f.SetCellValue(sheetName, "H2", "string")

	// 第2行：属性名称
	f.SetCellValue(sheetName, "A3", "Id")
	f.SetCellValue(sheetName, "B3", "Name")
	f.SetCellValue(sheetName, "C3", "Count")
	f.SetCellValue(sheetName, "D3", "DropGroup")
	f.SetCellValue(sheetName, "E3", "EnsureSmall")
	f.SetCellValue(sheetName, "F3", "EnsureSmallGroup")
	f.SetCellValue(sheetName, "G3", "EnsureBig")
	f.SetCellValue(sheetName, "H3", "EnsureBigGroup")

	// 第3行：导出标识
	for j := 0; j < 8; j++ {
		f.SetCellValue(sheetName, getCellName(j, 3), "server")
	}

	// 数据行（从第4行开始）
	// 第一行数据
	f.SetCellValue(sheetName, "A5", "2001")
	f.SetCellValue(sheetName, "B5", "掉落规则1")
	f.SetCellValue(sheetName, "C5", "10")
	f.SetCellValue(sheetName, "D5", "1001,1002,1003")
	f.SetCellValue(sheetName, "E5", "5")
	f.SetCellValue(sheetName, "F5", "2001,2002")
	f.SetCellValue(sheetName, "G5", "20")
	f.SetCellValue(sheetName, "H5", "3001,3002")

	// 第二行数据
	f.SetCellValue(sheetName, "A6", "2002")
	f.SetCellValue(sheetName, "B6", "掉落规则2")
	f.SetCellValue(sheetName, "C6", "5")
	f.SetCellValue(sheetName, "D6", "1004,1005")
	f.SetCellValue(sheetName, "E6", "3")
	f.SetCellValue(sheetName, "F6", "2003")
	f.SetCellValue(sheetName, "G6", "10")
	f.SetCellValue(sheetName, "H6", "3003")

	return map[string]*excelize.File{
		"掉落规则表|DropRule": f,
	}
}
