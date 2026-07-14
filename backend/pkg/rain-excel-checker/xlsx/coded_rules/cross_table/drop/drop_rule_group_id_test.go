package drop

import (
	"fmt"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestDropRuleGroupId_AllValid 全部有效数据应通过
func TestDropRuleGroupId_AllValid(t *testing.T) {
	// DropRule 表
	dropRuleFile := createDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "DropGroup", "EnsureSmallGroup", "EnsureBigGroup"}, [][]string{
		{"1", "单抽规则", "10,20", "30", "40"},
	})
	dropRuleCols, _ := dropRuleFile.GetCols("掉落规则表|DropRule")

	// DropGroup 表
	dropGroupFile := createDropRuleTestExcel("掉落分组表|DropGroup", []string{"Id", "Name", "Weight"}, [][]string{
		{"10", "普通组", "100"},
		{"20", "稀有组", "50"},
		{"30", "小保底组", "100"},
		{"40", "大保底组", "100"},
	})

	sheetMap := map[string]*excelize.File{
		"掉落分组表|DropGroup": dropGroupFile,
	}

	rule := &DropRuleGroupIdCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落规则表|DropRule",
		Cols:        dropRuleCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "全部有效时应通过，实际原因: %s", result.Reason)
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// TestDropRuleGroupId_InvalidDropGroup 无效掉落组ID应报错
func TestDropRuleGroupId_InvalidDropGroup(t *testing.T) {
	dropRuleFile := createDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "DropGroup", "EnsureSmallGroup", "EnsureBigGroup"}, [][]string{
		{"1", "单抽规则", "10,999", "", ""},
	})
	dropRuleCols, _ := dropRuleFile.GetCols("掉落规则表|DropRule")

	dropGroupFile := createDropRuleTestExcel("掉落分组表|DropGroup", []string{"Id", "Name", "Weight"}, [][]string{
		{"10", "普通组", "100"},
	})

	sheetMap := map[string]*excelize.File{
		"掉落分组表|DropGroup": dropGroupFile,
	}

	rule := &DropRuleGroupIdCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落规则表|DropRule",
		Cols:        dropRuleCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在无效掉落组ID时应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "999", "错误信息应包含无效ID")
	assert.Contains(t, result.ErrCells[0].Reason, "不存在", "错误信息应说明不存在")
	assert.Contains(t, result.ErrCells[0].Reason, "掉落组", "错误信息应说明是掉落组")
}

// TestDropRuleGroupId_InvalidEnsureGroup 无效保底组ID应报错
func TestDropRuleGroupId_InvalidEnsureGroup(t *testing.T) {
	dropRuleFile := createDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "DropGroup", "EnsureSmallGroup", "EnsureBigGroup"}, [][]string{
		{"1", "单抽规则", "10", "999", "888"},
	})
	dropRuleCols, _ := dropRuleFile.GetCols("掉落规则表|DropRule")

	dropGroupFile := createDropRuleTestExcel("掉落分组表|DropGroup", []string{"Id", "Name", "Weight"}, [][]string{
		{"10", "普通组", "100"},
	})

	sheetMap := map[string]*excelize.File{
		"掉落分组表|DropGroup": dropGroupFile,
	}

	rule := &DropRuleGroupIdCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落规则表|DropRule",
		Cols:        dropRuleCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在无效保底组ID时应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "999", "错误信息应包含无效小保底ID")
	assert.Contains(t, result.ErrCells[0].Reason, "888", "错误信息应包含无效大保底ID")
}

// TestDropRuleGroupId_NoDropGroupSheet 缺少 DropGroup 表应报错
func TestDropRuleGroupId_NoDropGroupSheet(t *testing.T) {
	dropRuleFile := createDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "DropGroup"}, [][]string{
		{"1", "单抽规则", "10"},
	})
	dropRuleCols, _ := dropRuleFile.GetCols("掉落规则表|DropRule")

	sheetMap := map[string]*excelize.File{}

	rule := &DropRuleGroupIdCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落规则表|DropRule",
		Cols:        dropRuleCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "缺少 DropGroup 表应报错")
	assert.Contains(t, result.Reason, "DropGroup", "应说明缺少 DropGroup 表")
}

// TestDropRuleGroupId_EmptyData 空数据应通过
func TestDropRuleGroupId_EmptyData(t *testing.T) {
	dropRuleFile := createDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "DropGroup"}, nil)
	dropRuleCols, _ := dropRuleFile.GetCols("掉落规则表|DropRule")

	dropGroupFile := createDropRuleTestExcel("掉落分组表|DropGroup", []string{"Id", "Name"}, nil)

	sheetMap := map[string]*excelize.File{
		"掉落分组表|DropGroup": dropGroupFile,
	}

	rule := &DropRuleGroupIdCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落规则表|DropRule",
		Cols:        dropRuleCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "空数据应通过")
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// TestDropRuleGroupId_EmptyGroupFields 空的组字段应跳过不报错
func TestDropRuleGroupId_EmptyGroupFields(t *testing.T) {
	dropRuleFile := createDropRuleTestExcel("掉落规则表|DropRule", []string{"Id", "Name", "DropGroup", "EnsureSmallGroup", "EnsureBigGroup"}, [][]string{
		{"1", "未配置组", "", "", ""},
	})
	dropRuleCols, _ := dropRuleFile.GetCols("掉落规则表|DropRule")

	dropGroupFile := createDropRuleTestExcel("掉落分组表|DropGroup", []string{"Id", "Name"}, nil)

	sheetMap := map[string]*excelize.File{
		"掉落分组表|DropGroup": dropGroupFile,
	}

	rule := &DropRuleGroupIdCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落规则表|DropRule",
		Cols:        dropRuleCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "空的组字段应跳过不报错")
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// createDropRuleTestExcel 创建测试用的 Excel 文件
func createDropRuleTestExcel(sheetName string, headers []string, dataRows [][]string) *excelize.File {
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
