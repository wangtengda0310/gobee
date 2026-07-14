package activity

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestActivityDrawskinCrossRef_Meta(t *testing.T) {
	rule := &ActivityDrawskinCrossReferenceCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawSkin")
	assert.Contains(t, meta.TargetSheets, "Activity")
}

func buildCrossRefSheetMap() map[string]*excelize.File {
	// Activity 表
	actFile := excelize.NewFile()
	actSheet := "活动表|Activity"
	_, _ = actFile.NewSheet(actSheet)
	_ = actFile.SetCellValue(actSheet, "A3", "Id")
	_ = actFile.SetCellValue(actSheet, "A5", "100")
	_ = actFile.SetCellValue(actSheet, "B3", "Name")
	_ = actFile.SetCellValue(actSheet, "B5", "丹青阁活动")
	_ = actFile.SetCellValue(actSheet, "C3", "ActivityType")
	_ = actFile.SetCellValue(actSheet, "C5", "ActTypeSkinRaffle")
	_ = actFile.SetCellValue(actSheet, "D3", "CustomParma")
	_ = actFile.SetCellValue(actSheet, "D5", "1")

	// DrawSkin 表
	drawFile := excelize.NewFile()
	drawSheet := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheet)
	_ = drawFile.SetCellValue(drawSheet, "A3", "Id")
	_ = drawFile.SetCellValue(drawSheet, "A5", "1")
	_ = drawFile.SetCellValue(drawSheet, "B3", "Name")
	_ = drawFile.SetCellValue(drawSheet, "B5", "测试池")
	_ = drawFile.SetCellValue(drawSheet, "C3", "ActivityId")
	_ = drawFile.SetCellValue(drawSheet, "C5", "100")

	return map[string]*excelize.File{
		"活动表|Activity":  actFile,
		"皮肤抽奖|DrawSkin": drawFile,
	}
}

func TestActivityDrawskinCrossRef_DrawSkinContext(t *testing.T) {
	sheetMap := buildCrossRefSheetMap()
	drawFile := sheetMap["皮肤抽奖|DrawSkin"]
	cols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	rule := &ActivityDrawskinCrossReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	assert.True(t, result.Ok, "ActivityId 正确引用丹青阁活动应通过")
	assert.Empty(t, result.ErrCells)
}

func TestActivityDrawskinCrossRef_ActivityIdNotExist(t *testing.T) {
	sheetMap := buildCrossRefSheetMap()
	// 修改 DrawSkin 的 ActivityId 为不存在的值
	drawFile := sheetMap["皮肤抽奖|DrawSkin"]
	_ = drawFile.SetCellValue("皮肤抽奖|DrawSkin", "C5", "9999")
	cols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	rule := &ActivityDrawskinCrossReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	assert.False(t, result.Ok, "ActivityId 不存在应报错")
}

func TestActivityDrawskinCrossRef_ActivityContext(t *testing.T) {
	sheetMap := buildCrossRefSheetMap()
	actFile := sheetMap["活动表|Activity"]
	cols, _ := actFile.GetCols("活动表|Activity")

	rule := &ActivityDrawskinCrossReferenceCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "活动表|Activity",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	// 双向引用一致（Activity 100 -> DrawSkin 1 -> ActivityId 100）
	assert.True(t, result.Ok)
}
