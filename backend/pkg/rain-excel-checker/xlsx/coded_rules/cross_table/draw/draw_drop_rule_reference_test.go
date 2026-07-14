package draw

import (
	"fmt"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestDrawDropRuleReference_AllValid 全部有效数据应通过
func TestDrawDropRuleReference_AllValid(t *testing.T) {
	// DrawSkin 表
	drawFile := createDrawDropRuleTestExcel("皮肤抽奖|DrawSkin", []string{"Id", "Name", "OnceDropRule", "TenDropRule"}, [][]string{
		{"1001", "丹青阁抽奖", "1", "2"},
	})
	drawCols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	// DropRule 表
	dropRuleFile := createDrawDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "Count"}, [][]string{
		{"1", "单抽规则", "1"},
		{"2", "十连抽规则", "10"},
	})

	sheetMap := map[string]*excelize.File{
		"掉落规则表|DropRule": dropRuleFile,
	}

	rule := &DrawDropRuleReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        drawCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "全部有效时应通过，实际原因: %s", result.Reason)
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// TestDrawDropRuleReference_InvalidOnceDropRule 无效单抽掉落规则应报错
func TestDrawDropRuleReference_InvalidOnceDropRule(t *testing.T) {
	drawFile := createDrawDropRuleTestExcel("皮肤抽奖|DrawSkin", []string{"Id", "Name", "OnceDropRule", "TenDropRule"}, [][]string{
		{"1001", "丹青阁抽奖", "999", "2"},
	})
	drawCols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	dropRuleFile := createDrawDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "Count"}, [][]string{
		{"1", "单抽规则", "1"},
		{"2", "十连抽规则", "10"},
	})

	sheetMap := map[string]*excelize.File{
		"掉落规则表|DropRule": dropRuleFile,
	}

	rule := &DrawDropRuleReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        drawCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在无效掉落规则ID时应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "999", "错误信息应包含无效ID")
	assert.Contains(t, result.ErrCells[0].Reason, "不存在", "错误信息应说明不存在")
	assert.Contains(t, result.ErrCells[0].Reason, "单抽", "错误信息应说明是单抽")
}

// TestDrawDropRuleReference_InvalidTenDropRule 无效十连抽掉落规则应报错
func TestDrawDropRuleReference_InvalidTenDropRule(t *testing.T) {
	drawFile := createDrawDropRuleTestExcel("皮肤抽奖|DrawSkin", []string{"Id", "Name", "OnceDropRule", "TenDropRule"}, [][]string{
		{"1001", "丹青阁抽奖", "1", "888"},
	})
	drawCols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	dropRuleFile := createDrawDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "Count"}, [][]string{
		{"1", "单抽规则", "1"},
	})

	sheetMap := map[string]*excelize.File{
		"掉落规则表|DropRule": dropRuleFile,
	}

	rule := &DrawDropRuleReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        drawCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在无效十连抽规则ID时应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "888", "错误信息应包含无效ID")
	assert.Contains(t, result.ErrCells[0].Reason, "十连抽", "错误信息应说明是十连抽")
}

// TestDrawDropRuleReference_NoDropRuleSheet 缺少 DropRule 表应报错
func TestDrawDropRuleReference_NoDropRuleSheet(t *testing.T) {
	drawFile := createDrawDropRuleTestExcel("皮肤抽奖|DrawSkin", []string{"Id", "Name", "OnceDropRule", "TenDropRule"}, [][]string{
		{"1001", "丹青阁抽奖", "1", "2"},
	})
	drawCols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	sheetMap := map[string]*excelize.File{}

	rule := &DrawDropRuleReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        drawCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "缺少 DropRule 表应报错")
	assert.Contains(t, result.Reason, "DropRule", "应说明缺少 DropRule 表")
}

// TestDrawDropRuleReference_EmptyData 空数据应通过
func TestDrawDropRuleReference_EmptyData(t *testing.T) {
	drawFile := createDrawDropRuleTestExcel("皮肤抽奖|DrawSkin", []string{"Id", "Name", "OnceDropRule", "TenDropRule"}, nil)
	drawCols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	dropRuleFile := createDrawDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name"}, nil)

	sheetMap := map[string]*excelize.File{
		"掉落规则表|DropRule": dropRuleFile,
	}

	rule := &DrawDropRuleReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        drawCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "空数据应通过")
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// TestDrawDropRuleReference_EmptyDropRuleRef 空的掉落规则字段应跳过不报错
func TestDrawDropRuleReference_EmptyDropRuleRef(t *testing.T) {
	drawFile := createDrawDropRuleTestExcel("皮肤抽奖|DrawSkin", []string{"Id", "Name", "OnceDropRule", "TenDropRule"}, [][]string{
		{"1001", "未配置抽奖", "", ""},
	})
	drawCols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	dropRuleFile := createDrawDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name"}, nil)

	sheetMap := map[string]*excelize.File{
		"掉落规则表|DropRule": dropRuleFile,
	}

	rule := &DrawDropRuleReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        drawCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "空的掉落规则字段应跳过不报错")
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// createDrawDropRuleTestExcel 创建测试用 Excel 文件
func createDrawDropRuleTestExcel(sheetName string, headers []string, dataRows [][]string) *excelize.File {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)

	// 行3: 列名
	for i, h := range headers {
		col := string(rune('A' + i))
		_ = f.SetCellValue(sheetName, col+"3", h)
	}
	// 行4: 类型标记
	for i := range headers {
		col := string(rune('A' + i))
		_ = f.SetCellValue(sheetName, col+"4", "string")
	}
	// 行5+: 数据
	for rowIdx, row := range dataRows {
		for colIdx, val := range row {
			col := string(rune('A' + colIdx))
			_ = f.SetCellValue(sheetName, col+fmt.Sprintf("%d", 5+rowIdx), val)
		}
	}
	return f
}
