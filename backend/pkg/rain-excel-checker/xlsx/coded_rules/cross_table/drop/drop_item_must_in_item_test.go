// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package drop

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestDropItemMustInItem_Meta 测试规则元数据
func TestDropItemMustInItem_Meta(t *testing.T) {
	rule := &DropItemMustInItemCheckRule{}
	meta := rule.Meta()

	assert.Equal(t, json_rule.DROP_ITEM_MUST_IN_ITEM_CHECK, meta.Type)
	assert.Equal(t, "掉落道具必须在道具表中", meta.DisplayName)
	assert.NotEmpty(t, meta.Description)
}

// TestDropItemMustInItem_AllValid 测试所有道具都存在的情况
func TestDropItemMustInItem_AllValid(t *testing.T) {
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
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")
	_ = itemFile.SetCellValue(itemSheetName, "A6", "")
	_ = itemFile.SetCellValue(itemSheetName, "B6", "")
	_ = itemFile.SetCellValue(itemSheetName, "C6", "1000002")

	// 准备 DropItem 表数据
	dropItemFile := excelize.NewFile()
	dropItemSheetName := "掉落道具表|DropItem"
	_, _ = dropItemFile.NewSheet(dropItemSheetName)
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A3", "Id")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B3", "Name")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C3", "DropGroup")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D3", "Item")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B4", "string")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D4", "ItemCfg[]")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A5", "101")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B5", "测试掉落1")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C5", "10001")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D5", "{1000001;1}")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A6", "102")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B6", "测试掉落2")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C6", "10002")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D6", "{1000002;1}{1000001;5}")

	sheetMap := map[string]*excelize.File{
		"道具表|Item":       itemFile,
		"掉落道具表|DropItem": dropItemFile,
	}

	dropItemCols, _ := dropItemFile.GetCols(dropItemSheetName)
	rule := &DropItemMustInItemCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   dropItemSheetName,
		Cols:        dropItemCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "所有道具都存在时应该通过检查")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}

// TestDropItemMustInItem_ItemNotExist 测试道具不存在的情况
func TestDropItemMustInItem_ItemNotExist(t *testing.T) {
	// 准备 Item 表数据 - 只包含 1000001
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
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")

	// 准备 DropItem 表数据 - 包含不存在的道具 9999999
	dropItemFile := excelize.NewFile()
	dropItemSheetName := "掉落道具表|DropItem"
	_, _ = dropItemFile.NewSheet(dropItemSheetName)
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A3", "Id")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B3", "Name")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C3", "DropGroup")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D3", "Item")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B4", "string")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D4", "ItemCfg[]")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A5", "101")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B5", "测试掉落1")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C5", "10001")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D5", "{1000001;1}{9999999;10}")

	sheetMap := map[string]*excelize.File{
		"道具表|Item":       itemFile,
		"掉落道具表|DropItem": dropItemFile,
	}

	dropItemCols, _ := dropItemFile.GetCols(dropItemSheetName)
	rule := &DropItemMustInItemCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   dropItemSheetName,
		Cols:        dropItemCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在不存在的道具时应该失败")
	assert.Len(t, result.ErrCells, 1, "应该有一个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "9999999", "错误信息应该包含不存在的道具ID")
	assert.Contains(t, result.ErrCells[0].Reason, "测试掉落1", "错误信息应该包含掉落名称")
}

// TestDropItemMustInItem_NoItemSheet 测试 Item 表不存在的情况
func TestDropItemMustInItem_NoItemSheet(t *testing.T) {
	dropItemFile := excelize.NewFile()
	dropItemSheetName := "掉落道具表|DropItem"
	_, _ = dropItemFile.NewSheet(dropItemSheetName)
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A3", "Id")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B3", "Name")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C3", "DropGroup")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D3", "Item")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B4", "string")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D4", "ItemCfg[]")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A5", "101")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B5", "测试掉落1")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C5", "10001")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D5", "{1000001;1}")

	sheetMap := map[string]*excelize.File{
		"掉落道具表|DropItem": dropItemFile,
	}

	dropItemCols, _ := dropItemFile.GetCols(dropItemSheetName)
	rule := &DropItemMustInItemCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   dropItemSheetName,
		Cols:        dropItemCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "Item 表不存在时应该失败")
	assert.Contains(t, result.Reason, "未找到 Item 表")
}

// TestDropItemMustInItem_EmptyItemColumn 测试 Item 列为空的情况
func TestDropItemMustInItem_EmptyItemColumn(t *testing.T) {
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
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")

	dropItemFile := excelize.NewFile()
	dropItemSheetName := "掉落道具表|DropItem"
	_, _ = dropItemFile.NewSheet(dropItemSheetName)
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D1", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D2", "")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A3", "Id")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B3", "Name")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C3", "DropGroup")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D3", "Item")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B4", "string")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C4", "int")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D4", "ItemCfg[]")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "A5", "101")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "B5", "测试掉落1")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "C5", "10001")
	_ = dropItemFile.SetCellValue(dropItemSheetName, "D5", "")

	sheetMap := map[string]*excelize.File{
		"道具表|Item":       itemFile,
		"掉落道具表|DropItem": dropItemFile,
	}

	dropItemCols, _ := dropItemFile.GetCols(dropItemSheetName)
	rule := &DropItemMustInItemCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   dropItemSheetName,
		Cols:        dropItemCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "Item 列为空时应该通过")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}
