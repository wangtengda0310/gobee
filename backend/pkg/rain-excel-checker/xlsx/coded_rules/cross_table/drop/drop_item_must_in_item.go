// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package drop

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DropItemMustInItemCheckRule 掉落道具必须在道具表中检查规则
//
// ## 校验规则
// 1. 掉落道具表|DropItem 中配置的道具必须在 道具表|Item 中有对应的配置
//
// ## 相关表结构
// - DropItem: Id, Name, DropGroup, Item (格式: {道具ID;数量}{道具ID;数量}...)
// - Item: Id, Name
//
// ## 检查流程
// 1. 加载 DropItem 表和 Item 表
// 2. 从 Item 表提取所有有效的道具 ID
// 3. 遍历 DropItem 表的每一行，解析 Item 列中的道具配置
// 4. 检查每个道具 ID 是否存在于 Item 表中
// 5. 记录所有不存在的道具 ID 到错误列表
type DropItemMustInItemCheckRule struct{}

// Meta 返回规则元数据
func (c *DropItemMustInItemCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DROP_ITEM_MUST_IN_ITEM_CHECK,
		DisplayName:    "掉落道具必须在道具表中",
		Description:    "检查掉落道具表|DropItem 中配置的道具是否在道具表|Item 中存在。如果道具不存在，可能是配置错误。",
		TargetSheets:   []string{"DropItem"},
		RequiredSheets: []string{"Item"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行掉落道具检查
func (c *DropItemMustInItemCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	// 步骤1: 初始化检查结果结构体
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "掉落道具必须在道具表中",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 步骤2: 加载 Item 表数据
	itemCols := c.loadItemSheet(param.SheetMap)
	if itemCols == nil {
		result.Ok = false
		result.Reason = "未找到 Item 表，无法执行道具存在性检查"
		return result
	}

	// 步骤3: 从 Item 表构建有效的道具 ID 集合
	validItemIds := c.buildValidItemIdSet(itemCols, param.StartRowIdx)

	// 步骤4: 查找 DropItem 表中的关键列索引
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	itemColIdx := helpers.GetColIndexByName(param.Cols, "Item")

	if itemColIdx < 0 {
		result.Ok = true
		result.Reason = "DropItem 表中未找到 Item 列"
		return result
	}

	// 步骤5-8: 遍历 DropItem 表，检查每行的 Item 列
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		// 获取 Item 列的值
		itemCfg := helpers.GetColValue(param.Cols, itemColIdx, rowIdx)
		if itemCfg == "" {
			continue // 空值跳过
		}

		// 步骤6: 解析 Item 列中的道具配置
		items := helpers.ParseItemCfg(itemCfg)
		if len(items) == 0 {
			continue // 无有效道具配置，跳过
		}

		// 获取当前行的 ID 和 Name（用于错误提示）
		rowId := ""
		rowName := ""
		if idColIdx >= 0 {
			rowId = helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		}
		if nameColIdx >= 0 {
			rowName = helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		}

		// 步骤7-8: 检查每个道具
		var missingItemIds []string

		for _, item := range items {
			// 检查道具 ID 是否存在
			if !validItemIds[item.ItemId] {
				missingItemIds = append(missingItemIds, fmt.Sprintf("%d", item.ItemId))
			}
		}

		// 步骤9: 记录错误
		if len(missingItemIds) > 0 {
			reason := fmt.Sprintf("掉落道具【%s】(ID=%s) 配置的道具 ID [%s] 在道具表|Item 中不存在",
				rowName, rowId, fmt.Sprintf("%v", missingItemIds))
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   reason,
			})
		}
	}

	// 步骤10: 设置检查结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个掉落道具配置问题", len(result.ErrCells))
	}

	return result
}

// loadItemSheet 加载 Item 表数据
//
// 执行流程：
// 1. 从 sheetMap 中查找 Item 表（支持后缀匹配）
// 2. 读取表数据并返回
func (c *DropItemMustInItemCheckRule) loadItemSheet(sheetMap map[string]*excelize.File) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Item"); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildValidItemIdSet 从 Item 表构建有效的道具 ID 集合
//
// 执行流程：
// 1. 查找 Item 表中的 Id 列索引
// 2. 遍历所有数据行，提取道具 ID
// 3. 构建 map[int]bool 类型的集合
func (c *DropItemMustInItemCheckRule) buildValidItemIdSet(itemCols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)

	// 查找 Id 列索引
	idColIdx := helpers.GetColIndexByName(itemCols, "Id")
	if idColIdx < 0 {
		return validIds
	}

	// 遍历所有数据行
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
