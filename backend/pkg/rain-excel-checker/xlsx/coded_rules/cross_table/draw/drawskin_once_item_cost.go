package draw

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DrawskinOnceItemCostCheckRule DSK-06/07/08 OnceItemCost 检查
// DSK-06: OnceItemCost 必须配置（非空）
// DSK-07: OnceItemCost 中的 ItemId 必须在 Item 表中存在
// DSK-08: OnceItemCost 中的 ItemNum 必须 > 0
type DrawskinOnceItemCostCheckRule struct{}

func (c *DrawskinOnceItemCostCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DRAWSKIN_ONCE_ITEM_COST_CHECK,
		DisplayName:    "DrawSkin单抽消耗配置检查",
		Description:    "检查DrawSkin表的OnceItemCost非空(DSK-06)、ItemId引用Item表(DSK-07)、ItemNum>0(DSK-08)",
		TargetSheets:   []string{"DrawSkin"},
		RequiredSheets: []string{"Item"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

func (c *DrawskinOnceItemCostCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DrawSkin单抽消耗配置检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	onceCostColIdx := helpers.GetColIndexByName(param.Cols, "OnceItemCost")

	if onceCostColIdx == -1 {
		result.Ok = true
		result.Reason = "未找到 OnceItemCost 列，跳过检查"
		return result
	}

	// 加载 Item 表
	itemCols := c.loadItemSheet(param.SheetMap)
	if itemCols == nil {
		result.Ok = false
		result.Reason = "未找到 Item 表，无法执行道具存在性检查"
		return result
	}
	validItemIds := c.buildValidItemIdSet(itemCols, param.StartRowIdx)

	errorCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := ""
		if idColIdx != -1 {
			rowId = helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		}
		name := ""
		if nameColIdx != -1 {
			name = helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		}

		costStr := strings.TrimSpace(helpers.GetColValue(param.Cols, onceCostColIdx, rowIdx))

		// DSK-06: OnceItemCost 必须配置
		if costStr == "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 OnceItemCost 未配置", name, rowId),
			})
			errorCount++
			continue
		}

		// 解析 OnceItemCost
		items := helpers.ParseDrawSkinItemCfg(costStr)
		if len(items) == 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 OnceItemCost【%s】格式错误，应为{itemId;count}格式", name, rowId, costStr),
			})
			errorCount++
			continue
		}

		for _, item := range items {
			// DSK-07: ItemId 引用 Item 表
			if !validItemIds[item.ItemId] {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的单抽消耗道具ID【%d】在道具表中不存在", name, rowId, item.ItemId),
				})
				errorCount++
			}

			// DSK-08: ItemNum > 0
			if item.Count <= 0 {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的单抽消耗数量 ItemNum=%d 必须 > 0", name, rowId, item.Count),
				})
				errorCount++
			}
		}
	}

	if errorCount > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("DrawSkin单抽消耗配置检查发现 %d 个问题", errorCount)
	}

	return result
}

// loadItemSheet 从 SheetMap 中加载 Item 表的列数据
func (c *DrawskinOnceItemCostCheckRule) loadItemSheet(sheetMap map[string]*excelize.File) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Item"); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildValidItemIdSet 从 Item 表列数据中构建有效道具ID集合
func (c *DrawskinOnceItemCostCheckRule) buildValidItemIdSet(itemCols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(itemCols, "Id")
	if idColIdx < 0 {
		return validIds
	}
	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(itemCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(itemCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := helpers.ParseIntWithError(idStr)
		if err != nil {
			continue
		}
		validIds[id] = true
	}
	return validIds
}
