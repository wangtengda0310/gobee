// Package activity 提供活动相关的跨表校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package activity

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestActivityDanqinggeCustomParamIsItemid_Meta 测试规则元数据
func TestActivityDanqinggeCustomParamIsItemid_Meta(t *testing.T) {
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	meta := rule.Meta()

	assert.Equal(t, json_rule.DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK, meta.Type)
	assert.Equal(t, "丹青阁活动自定义参数检查", meta.DisplayName)
	assert.NotEmpty(t, meta.Description)
}

// TestActivityDanqinggeCustomParamIsItemid_ValidDrawId 测试 DrawId 存在的情况
func TestActivityDanqinggeCustomParamIsItemid_ValidDrawId(t *testing.T) {
	// 准备 DrawSkin 表数据
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	_ = drawFile.SetCellValue(drawSheetName, "A1", "")
	_ = drawFile.SetCellValue(drawSheetName, "B1", "")
	_ = drawFile.SetCellValue(drawSheetName, "C1", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A2", "")
	_ = drawFile.SetCellValue(drawSheetName, "B2", "")
	_ = drawFile.SetCellValue(drawSheetName, "C2", "int")
	_ = drawFile.SetCellValue(drawSheetName, "A3", "")
	_ = drawFile.SetCellValue(drawSheetName, "B3", "")
	_ = drawFile.SetCellValue(drawSheetName, "C3", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A4", "")
	_ = drawFile.SetCellValue(drawSheetName, "B4", "")
	_ = drawFile.SetCellValue(drawSheetName, "C4", "server/client")
	_ = drawFile.SetCellValue(drawSheetName, "A5", "")
	_ = drawFile.SetCellValue(drawSheetName, "B5", "")
	_ = drawFile.SetCellValue(drawSheetName, "C5", "2001")
	_ = drawFile.SetCellValue(drawSheetName, "A6", "")
	_ = drawFile.SetCellValue(drawSheetName, "B6", "")
	_ = drawFile.SetCellValue(drawSheetName, "C6", "2002")

	// 准备 Activity 表数据
	activityFile := excelize.NewFile()
	activitySheetName := "活动表|Activity"
	_, _ = activityFile.NewSheet(activitySheetName)
	_ = activityFile.SetCellValue(activitySheetName, "A1", "")
	_ = activityFile.SetCellValue(activitySheetName, "B1", "")
	_ = activityFile.SetCellValue(activitySheetName, "C1", "")
	_ = activityFile.SetCellValue(activitySheetName, "A2", "")
	_ = activityFile.SetCellValue(activitySheetName, "B2", "")
	_ = activityFile.SetCellValue(activitySheetName, "C2", "")
	_ = activityFile.SetCellValue(activitySheetName, "A3", "Id")
	_ = activityFile.SetCellValue(activitySheetName, "B3", "Name")
	_ = activityFile.SetCellValue(activitySheetName, "C3", "CustomParma")
	_ = activityFile.SetCellValue(activitySheetName, "A4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "B4", "string")
	_ = activityFile.SetCellValue(activitySheetName, "C4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "A5", "1")
	_ = activityFile.SetCellValue(activitySheetName, "B5", "丹青阁活动")
	_ = activityFile.SetCellValue(activitySheetName, "C5", "2002")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawFile,
		"活动表|Activity":  activityFile,
	}

	activityCols, _ := activityFile.GetCols(activitySheetName)
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activitySheetName,
		Cols:        activityCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "DrawId 存在时应该通过检查")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}

// TestActivityDanqinggeCustomParamIsItemid_DrawIdNotExist 测试 DrawId 不存在的情况
func TestActivityDanqinggeCustomParamIsItemid_DrawIdNotExist(t *testing.T) {
	// 准备 DrawSkin 表数据 - 只包含 2001
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	_ = drawFile.SetCellValue(drawSheetName, "A1", "")
	_ = drawFile.SetCellValue(drawSheetName, "B1", "")
	_ = drawFile.SetCellValue(drawSheetName, "C1", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A2", "")
	_ = drawFile.SetCellValue(drawSheetName, "B2", "")
	_ = drawFile.SetCellValue(drawSheetName, "C2", "int")
	_ = drawFile.SetCellValue(drawSheetName, "A3", "")
	_ = drawFile.SetCellValue(drawSheetName, "B3", "")
	_ = drawFile.SetCellValue(drawSheetName, "C3", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A4", "")
	_ = drawFile.SetCellValue(drawSheetName, "B4", "")
	_ = drawFile.SetCellValue(drawSheetName, "C4", "server/client")
	_ = drawFile.SetCellValue(drawSheetName, "A5", "")
	_ = drawFile.SetCellValue(drawSheetName, "B5", "")
	_ = drawFile.SetCellValue(drawSheetName, "C5", "2001")

	// 准备 Activity 表数据 - 丹青阁活动的 CustomParma 为 9999（不存在）
	activityFile := excelize.NewFile()
	activitySheetName := "活动表|Activity"
	_, _ = activityFile.NewSheet(activitySheetName)
	_ = activityFile.SetCellValue(activitySheetName, "A1", "")
	_ = activityFile.SetCellValue(activitySheetName, "B1", "")
	_ = activityFile.SetCellValue(activitySheetName, "C1", "")
	_ = activityFile.SetCellValue(activitySheetName, "A2", "")
	_ = activityFile.SetCellValue(activitySheetName, "B2", "")
	_ = activityFile.SetCellValue(activitySheetName, "C2", "")
	_ = activityFile.SetCellValue(activitySheetName, "A3", "Id")
	_ = activityFile.SetCellValue(activitySheetName, "B3", "Name")
	_ = activityFile.SetCellValue(activitySheetName, "C3", "CustomParma")
	_ = activityFile.SetCellValue(activitySheetName, "A4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "B4", "string")
	_ = activityFile.SetCellValue(activitySheetName, "C4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "A5", "1")
	_ = activityFile.SetCellValue(activitySheetName, "B5", "丹青阁活动")
	_ = activityFile.SetCellValue(activitySheetName, "C5", "9999")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawFile,
		"活动表|Activity":  activityFile,
	}

	activityCols, _ := activityFile.GetCols(activitySheetName)
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activitySheetName,
		Cols:        activityCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "DrawId 不存在时应该失败")
	assert.Len(t, result.ErrCells, 1, "应该有一个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "9999", "错误信息应该包含不存在的DrawId")
	assert.Contains(t, result.ErrCells[0].Reason, "丹青阁活动", "错误信息应该包含活动名称")
}

// TestActivityDanqinggeCustomParamIsItemid_NoDrawSheet 测试 DrawSkin 表不存在的情况
func TestActivityDanqinggeCustomParamIsItemid_NoDrawSheet(t *testing.T) {
	activityFile := excelize.NewFile()
	activitySheetName := "活动表|Activity"
	_, _ = activityFile.NewSheet(activitySheetName)
	_ = activityFile.SetCellValue(activitySheetName, "A1", "")
	_ = activityFile.SetCellValue(activitySheetName, "B1", "")
	_ = activityFile.SetCellValue(activitySheetName, "C1", "")
	_ = activityFile.SetCellValue(activitySheetName, "A2", "")
	_ = activityFile.SetCellValue(activitySheetName, "B2", "")
	_ = activityFile.SetCellValue(activitySheetName, "C2", "")
	_ = activityFile.SetCellValue(activitySheetName, "A3", "Id")
	_ = activityFile.SetCellValue(activitySheetName, "B3", "Name")
	_ = activityFile.SetCellValue(activitySheetName, "C3", "CustomParma")
	_ = activityFile.SetCellValue(activitySheetName, "A4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "B4", "string")
	_ = activityFile.SetCellValue(activitySheetName, "C4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "A5", "1")
	_ = activityFile.SetCellValue(activitySheetName, "B5", "丹青阁活动")
	_ = activityFile.SetCellValue(activitySheetName, "C5", "2002")

	sheetMap := map[string]*excelize.File{
		"活动表|Activity": activityFile,
	}

	activityCols, _ := activityFile.GetCols(activitySheetName)
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activitySheetName,
		Cols:        activityCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "DrawSkin 表不存在时应该失败")
	assert.Contains(t, result.Reason, "未找到 DrawSkin 表")
}

// TestActivityDanqinggeCustomParamIsItemid_NoDanqinggeActivity 测试没有丹青阁活动的情况
func TestActivityDanqinggeCustomParamIsItemid_NoDanqinggeActivity(t *testing.T) {
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	_ = drawFile.SetCellValue(drawSheetName, "A1", "")
	_ = drawFile.SetCellValue(drawSheetName, "B1", "")
	_ = drawFile.SetCellValue(drawSheetName, "C1", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A2", "")
	_ = drawFile.SetCellValue(drawSheetName, "B2", "")
	_ = drawFile.SetCellValue(drawSheetName, "C2", "int")
	_ = drawFile.SetCellValue(drawSheetName, "A3", "")
	_ = drawFile.SetCellValue(drawSheetName, "B3", "")
	_ = drawFile.SetCellValue(drawSheetName, "C3", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A4", "")
	_ = drawFile.SetCellValue(drawSheetName, "B4", "")
	_ = drawFile.SetCellValue(drawSheetName, "C4", "server/client")
	_ = drawFile.SetCellValue(drawSheetName, "A5", "")
	_ = drawFile.SetCellValue(drawSheetName, "B5", "")
	_ = drawFile.SetCellValue(drawSheetName, "C5", "2001")

	activityFile := excelize.NewFile()
	activitySheetName := "活动表|Activity"
	_, _ = activityFile.NewSheet(activitySheetName)
	_ = activityFile.SetCellValue(activitySheetName, "A1", "")
	_ = activityFile.SetCellValue(activitySheetName, "B1", "")
	_ = activityFile.SetCellValue(activitySheetName, "C1", "")
	_ = activityFile.SetCellValue(activitySheetName, "A2", "")
	_ = activityFile.SetCellValue(activitySheetName, "B2", "")
	_ = activityFile.SetCellValue(activitySheetName, "C2", "")
	_ = activityFile.SetCellValue(activitySheetName, "A3", "Id")
	_ = activityFile.SetCellValue(activitySheetName, "B3", "Name")
	_ = activityFile.SetCellValue(activitySheetName, "C3", "CustomParma")
	_ = activityFile.SetCellValue(activitySheetName, "A4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "B4", "string")
	_ = activityFile.SetCellValue(activitySheetName, "C4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "A5", "1")
	_ = activityFile.SetCellValue(activitySheetName, "B5", "其他活动")
	_ = activityFile.SetCellValue(activitySheetName, "C5", "2001")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawFile,
		"活动表|Activity":  activityFile,
	}

	activityCols, _ := activityFile.GetCols(activitySheetName)
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activitySheetName,
		Cols:        activityCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "没有丹青阁活动时应该通过")
	assert.Contains(t, result.Reason, "未找到丹青阁活动")
}

// TestActivityDanqinggeCustomParamIsItemid_EmptyCustomParma 测试 CustomParma 为空的情况
func TestActivityDanqinggeCustomParamIsItemid_EmptyCustomParma(t *testing.T) {
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	_ = drawFile.SetCellValue(drawSheetName, "A1", "")
	_ = drawFile.SetCellValue(drawSheetName, "B1", "")
	_ = drawFile.SetCellValue(drawSheetName, "C1", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A2", "")
	_ = drawFile.SetCellValue(drawSheetName, "B2", "")
	_ = drawFile.SetCellValue(drawSheetName, "C2", "int")
	_ = drawFile.SetCellValue(drawSheetName, "A3", "")
	_ = drawFile.SetCellValue(drawSheetName, "B3", "")
	_ = drawFile.SetCellValue(drawSheetName, "C3", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A4", "")
	_ = drawFile.SetCellValue(drawSheetName, "B4", "")
	_ = drawFile.SetCellValue(drawSheetName, "C4", "server/client")
	_ = drawFile.SetCellValue(drawSheetName, "A5", "")
	_ = drawFile.SetCellValue(drawSheetName, "B5", "")
	_ = drawFile.SetCellValue(drawSheetName, "C5", "2001")

	activityFile := excelize.NewFile()
	activitySheetName := "活动表|Activity"
	_, _ = activityFile.NewSheet(activitySheetName)
	_ = activityFile.SetCellValue(activitySheetName, "A1", "")
	_ = activityFile.SetCellValue(activitySheetName, "B1", "")
	_ = activityFile.SetCellValue(activitySheetName, "C1", "")
	_ = activityFile.SetCellValue(activitySheetName, "A2", "")
	_ = activityFile.SetCellValue(activitySheetName, "B2", "")
	_ = activityFile.SetCellValue(activitySheetName, "C2", "")
	_ = activityFile.SetCellValue(activitySheetName, "A3", "Id")
	_ = activityFile.SetCellValue(activitySheetName, "B3", "Name")
	_ = activityFile.SetCellValue(activitySheetName, "C3", "CustomParma")
	_ = activityFile.SetCellValue(activitySheetName, "A4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "B4", "string")
	_ = activityFile.SetCellValue(activitySheetName, "C4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "A5", "1")
	_ = activityFile.SetCellValue(activitySheetName, "B5", "丹青阁活动")
	_ = activityFile.SetCellValue(activitySheetName, "C5", "")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawFile,
		"活动表|Activity":  activityFile,
	}

	activityCols, _ := activityFile.GetCols(activitySheetName)
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activitySheetName,
		Cols:        activityCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	// ACT-02: CustomParma 为空时报告 Warning（Ok=true 但 ErrCells 非空）
	assert.True(t, result.Ok, "CustomParma 为空时 Ok 应为 true（Warning 性质）")
	assert.NotEmpty(t, result.ErrCells, "CustomParma 为空时应报告 Warning")
	assert.Contains(t, result.ErrCells[0].Reason, "CustomParma 未配置")
}

// TestActivityDanqinggeCustomParamIsItemid_InvalidCustomParma 测试 CustomParma 不是有效数字的情况
func TestActivityDanqinggeCustomParamIsItemid_InvalidCustomParma(t *testing.T) {
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	_ = drawFile.SetCellValue(drawSheetName, "A1", "")
	_ = drawFile.SetCellValue(drawSheetName, "B1", "")
	_ = drawFile.SetCellValue(drawSheetName, "C1", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A2", "")
	_ = drawFile.SetCellValue(drawSheetName, "B2", "")
	_ = drawFile.SetCellValue(drawSheetName, "C2", "int")
	_ = drawFile.SetCellValue(drawSheetName, "A3", "")
	_ = drawFile.SetCellValue(drawSheetName, "B3", "")
	_ = drawFile.SetCellValue(drawSheetName, "C3", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A4", "")
	_ = drawFile.SetCellValue(drawSheetName, "B4", "")
	_ = drawFile.SetCellValue(drawSheetName, "C4", "server/client")
	_ = drawFile.SetCellValue(drawSheetName, "A5", "")
	_ = drawFile.SetCellValue(drawSheetName, "B5", "")
	_ = drawFile.SetCellValue(drawSheetName, "C5", "2001")

	activityFile := excelize.NewFile()
	activitySheetName := "活动表|Activity"
	_, _ = activityFile.NewSheet(activitySheetName)
	_ = activityFile.SetCellValue(activitySheetName, "A1", "")
	_ = activityFile.SetCellValue(activitySheetName, "B1", "")
	_ = activityFile.SetCellValue(activitySheetName, "C1", "")
	_ = activityFile.SetCellValue(activitySheetName, "A2", "")
	_ = activityFile.SetCellValue(activitySheetName, "B2", "")
	_ = activityFile.SetCellValue(activitySheetName, "C2", "")
	_ = activityFile.SetCellValue(activitySheetName, "A3", "Id")
	_ = activityFile.SetCellValue(activitySheetName, "B3", "Name")
	_ = activityFile.SetCellValue(activitySheetName, "C3", "CustomParma")
	_ = activityFile.SetCellValue(activitySheetName, "A4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "B4", "string")
	_ = activityFile.SetCellValue(activitySheetName, "C4", "int")
	_ = activityFile.SetCellValue(activitySheetName, "A5", "1")
	_ = activityFile.SetCellValue(activitySheetName, "B5", "丹青阁活动")
	_ = activityFile.SetCellValue(activitySheetName, "C5", "abc")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawFile,
		"活动表|Activity":  activityFile,
	}

	activityCols, _ := activityFile.GetCols(activitySheetName)
	rule := &ActivityDanqinggeCustomParamIsItemidCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activitySheetName,
		Cols:        activityCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "CustomParma 不是有效数字时应该失败")
	assert.Len(t, result.ErrCells, 1, "应该有一个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "abc", "错误信息应该包含无效的CustomParma值")
	assert.Contains(t, result.ErrCells[0].Reason, "不是有效的数字格式", "错误信息应该说明不是有效数字")
}
