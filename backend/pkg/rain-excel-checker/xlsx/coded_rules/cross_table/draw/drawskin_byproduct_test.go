// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package draw

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestDrawskinByproduct_Meta 测试规则元数据
func TestDrawskinByproduct_Meta(t *testing.T) {
	rule := &DrawskinByproductCheckRule{}
	meta := rule.Meta()

	assert.Equal(t, json_rule.DRAWSKIN_BYPRODUCT_CHECK, meta.Type)
	assert.Equal(t, "皮肤抽奖副产品检查", meta.DisplayName)
	assert.NotEmpty(t, meta.Description)
}

// TestDrawskinByproduct_AllValid 测试所有道具都存在的情况
func TestDrawskinByproduct_AllValid(t *testing.T) {
	// 准备 Item 表数据
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "server/client")
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1022201")
	_ = itemFile.SetCellValue(itemSheetName, "A6", "")
	_ = itemFile.SetCellValue(itemSheetName, "B6", "")
	_ = itemFile.SetCellValue(itemSheetName, "C6", "1040010")

	// 准备 DrawSkin 表数据
	drawskinFile := excelize.NewFile()
	drawskinSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawskinFile.NewSheet(drawskinSheetName)
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A3", "Id")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B3", "Name")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C3", "byproduct")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A4", "int")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B4", "string")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C4", "int[]")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A5", "2001")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B5", "皮肤1")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C5", "1022201,1040010")

	sheetMap := map[string]*excelize.File{
		"道具表|Item":      itemFile,
		"皮肤抽奖|DrawSkin": drawskinFile,
	}

	drawskinCols, _ := drawskinFile.GetCols(drawskinSheetName)
	rule := &DrawskinByproductCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   drawskinSheetName,
		Cols:        drawskinCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "所有道具都存在时应该通过检查")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}

// TestDrawskinByproduct_ItemNotExist 测试道具不存在的情况
func TestDrawskinByproduct_ItemNotExist(t *testing.T) {
	// 准备 Item 表数据 - 只包含 1022201
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "server/client")
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1022201")

	// 准备 DrawSkin 表数据 - 包含不存在的道具 9999999
	drawskinFile := excelize.NewFile()
	drawskinSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawskinFile.NewSheet(drawskinSheetName)
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A3", "Id")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B3", "Name")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C3", "byproduct")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A4", "int")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B4", "string")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C4", "int[]")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A5", "2001")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B5", "皮肤1")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C5", "1022201,9999999")

	sheetMap := map[string]*excelize.File{
		"道具表|Item":      itemFile,
		"皮肤抽奖|DrawSkin": drawskinFile,
	}

	drawskinCols, _ := drawskinFile.GetCols(drawskinSheetName)
	rule := &DrawskinByproductCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   drawskinSheetName,
		Cols:        drawskinCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在不存在的道具时应该失败")
	assert.Len(t, result.ErrCells, 1, "应该有一个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "9999999", "错误信息应该包含不存在的道具ID")
	assert.Contains(t, result.ErrCells[0].Reason, "皮肤1", "错误信息应该包含皮肤名称")
}

// TestDrawskinByproduct_NoItemSheet 测试 Item 表不存在的情况
func TestDrawskinByproduct_NoItemSheet(t *testing.T) {
	drawskinFile := excelize.NewFile()
	drawskinSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawskinFile.NewSheet(drawskinSheetName)
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A3", "Id")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B3", "Name")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C3", "byproduct")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A4", "int")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B4", "string")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C4", "int[]")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A5", "2001")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B5", "皮肤1")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C5", "1022201")

	sheetMap := map[string]*excelize.File{
		"皮肤抽奖|DrawSkin": drawskinFile,
	}

	drawskinCols, _ := drawskinFile.GetCols(drawskinSheetName)
	rule := &DrawskinByproductCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   drawskinSheetName,
		Cols:        drawskinCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "Item 表不存在时应该失败")
	assert.Contains(t, result.Reason, "未找到 Item 表")
}

// TestDrawskinByproduct_EmptyByproduct 测试 byproduct 列为空的情况
func TestDrawskinByproduct_EmptyByproduct(t *testing.T) {
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "server/client")
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1022201")

	drawskinFile := excelize.NewFile()
	drawskinSheetName := "皮肤抽奖|DrawSkin"
	_, _ = drawskinFile.NewSheet(drawskinSheetName)
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C1", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C2", "")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A3", "Id")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B3", "Name")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C3", "byproduct")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A4", "int")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B4", "string")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C4", "int[]")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "A5", "2001")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "B5", "皮肤1")
	_ = drawskinFile.SetCellValue(drawskinSheetName, "C5", "")

	sheetMap := map[string]*excelize.File{
		"道具表|Item":      itemFile,
		"皮肤抽奖|DrawSkin": drawskinFile,
	}

	drawskinCols, _ := drawskinFile.GetCols(drawskinSheetName)
	rule := &DrawskinByproductCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   drawskinSheetName,
		Cols:        drawskinCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "byproduct 列为空时应该通过")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}
