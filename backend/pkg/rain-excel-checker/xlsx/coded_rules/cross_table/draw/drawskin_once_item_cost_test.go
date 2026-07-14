package draw

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func buildOnceCostSheetMap(hasValidItem bool) map[string]*excelize.File {
	// 构建 Item 表
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	_ = itemFile.SetCellValue(itemSheetName, "C1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "int")
	if hasValidItem {
		_ = itemFile.SetCellValue(itemSheetName, "C5", "1000005")
	}

	// 构建 DrawSkin 表
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	// Id 列 (A)
	_ = drawFile.SetCellValue(drawSheetName, "A3", "Id")
	_ = drawFile.SetCellValue(drawSheetName, "A5", "1")
	// Name 列 (B)
	_ = drawFile.SetCellValue(drawSheetName, "B3", "Name")
	_ = drawFile.SetCellValue(drawSheetName, "B5", "测试池")
	// OnceItemCost 列 (C) - 在各测试中覆盖

	return map[string]*excelize.File{
		"道具表|Item":      itemFile,
		"皮肤抽奖|DrawSkin": drawFile,
	}
}

func TestDrawskinOnceItemCost_Meta(t *testing.T) {
	rule := &DrawskinOnceItemCostCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.DRAWSKIN_ONCE_ITEM_COST_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawSkin")
	assert.Contains(t, meta.RequiredSheets, "Item")
}

func TestDrawskinOnceItemCost_MissingOnceItemCost(t *testing.T) {
	sheetMap := buildOnceCostSheetMap(true)
	// 不设置 OnceItemCost 列值（留空）
	rule := &DrawskinOnceItemCostCheckRule{}

	drawFile := sheetMap["皮肤抽奖|DrawSkin"]
	cols, _ := drawFile.GetCols("皮肤抽奖|DrawSkin")

	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	// OnceItemCost 列不存在时跳过检查
	assert.True(t, result.Ok, "缺少 OnceItemCost 列时应跳过")
}

func TestDrawskinOnceItemCost_NoItemSheet(t *testing.T) {
	drawFile := excelize.NewFile()
	drawSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawFile.NewSheet(drawSheetName)
	_ = drawFile.SetCellValue(drawSheetName, "C3", "OnceItemCost")
	_ = drawFile.SetCellValue(drawSheetName, "C5", "{1000005;10}")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawFile,
	}

	cols, _ := drawFile.GetCols(drawSheetName)
	rule := &DrawskinOnceItemCostCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
		SheetMap:    sheetMap,
	})
	assert.False(t, result.Ok, "缺少 Item 表应报错")
}
