package activity

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestActivityDrawskinTimeOverlap_Meta(t *testing.T) {
	rule := &ActivityDrawskinTimeOverlapCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawSkin")
	assert.Contains(t, meta.TargetSheets, "Activity")
}

func buildTimeOverlapSheetMap(actStart, actEnd, dsStart, dsEnd string) map[string]*excelize.File {
	// Activity 表
	actFile := excelize.NewFile()
	actSheet := "活动表|Activity"
	_, _ = actFile.NewSheet(actSheet)
	_ = actFile.SetCellValue(actSheet, "A3", "Id")
	_ = actFile.SetCellValue(actSheet, "A5", "100")
	_ = actFile.SetCellValue(actSheet, "B3", "Name")
	_ = actFile.SetCellValue(actSheet, "B5", "丹青阁活动")
	_ = actFile.SetCellValue(actSheet, "C3", "StartTime")
	_ = actFile.SetCellValue(actSheet, "C5", actStart)
	_ = actFile.SetCellValue(actSheet, "D3", "EndTime")
	_ = actFile.SetCellValue(actSheet, "D5", actEnd)

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
	_ = drawFile.SetCellValue(drawSheet, "D3", "StartTime")
	_ = drawFile.SetCellValue(drawSheet, "D5", dsStart)
	_ = drawFile.SetCellValue(drawSheet, "E3", "EndTime")
	_ = drawFile.SetCellValue(drawSheet, "E5", dsEnd)

	return map[string]*excelize.File{
		"活动表|Activity":  actFile,
		"皮肤抽奖|DrawSkin": drawFile,
	}
}

func TestActivityDrawskinTimeOverlap_HasOverlap(t *testing.T) {
	sheetMap := buildTimeOverlapSheetMap(
		"2025-01-01 00:00:00", "2025-06-01 00:00:00",
		"2025-02-01 00:00:00", "2025-05-01 00:00:00",
	)
	drawFile := sheetMap["皮肤抽奖|DrawSkin"]
	cols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	rule := &ActivityDrawskinTimeOverlapCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	assert.True(t, result.Ok, "有交集时应通过")
	assert.Empty(t, result.ErrCells)
}

func TestActivityDrawskinTimeOverlap_NoOverlap(t *testing.T) {
	sheetMap := buildTimeOverlapSheetMap(
		"2025-01-01 00:00:00", "2025-03-01 00:00:00",
		"2025-04-01 00:00:00", "2025-06-01 00:00:00",
	)
	drawFile := sheetMap["皮肤抽奖|DrawSkin"]
	cols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	rule := &ActivityDrawskinTimeOverlapCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	assert.True(t, result.Ok, "无交集是 Warning 性质")
	assert.NotEmpty(t, result.ErrCells, "无交集应有警告")
}
