// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package draw

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DrawskinByproductCheckRule 皮肤抽奖副产品检查规则
//
// ## 校验规则
// 1. 皮肤抽奖表|DrawSkin 中配置的副产品（byproduct）必须配置归属道具
// 2. byproduct 字段格式: 逗号分隔的道具ID，例如: 1022201,1040010
//
// ## 相关表结构
// - DrawSkin: Id, Name, byproduct (格式: 道具ID,道具ID,...)
// - Item: Id, Name (用于验证道具是否存在)
//
// ## 检查流程
// 1. 加载 DrawSkin 表和 Item 表
// 2. 从 Item 表提取所有有效的道具 ID
// 3. 遍历 DrawSkin 表的每一行，解析 byproduct 列中的道具ID
// 4. 检查每个道具 ID 是否存在于 Item 表中
// 5. 记录所有不存在的道具 ID 到错误列表
type DrawskinByproductCheckRule struct{}

// Meta 返回规则元数据
func (c *DrawskinByproductCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DRAWSKIN_BYPRODUCT_CHECK,
		DisplayName:    "皮肤抽奖副产品检查",
		Description:    "检查皮肤抽奖表|DrawSkin 中配置的副产品（byproduct）道具是否在道具表|Item 中存在。如果道具不存在，可能是配置错误。",
		TargetSheets:   []string{"DrawSkin"},
		RequiredSheets: []string{"Item"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行皮肤抽奖副产品检查
func (c *DrawskinByproductCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	// 步骤1: 初始化检查结果结构体
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "皮肤抽奖副产品检查",
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

	// 步骤4: 查找 DrawSkin 表中的关键列索引
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	byproductColIdx := helpers.GetColIndexByName(param.Cols, "byproduct")

	if byproductColIdx < 0 {
		result.Ok = true
		result.Reason = "DrawSkin 表中未找到 byproduct 列"
		return result
	}

	// 步骤5-8: 遍历 DrawSkin 表，检查每行的 byproduct 列
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		// 获取 byproduct 列的值
		byproductValue := helpers.GetColValue(param.Cols, byproductColIdx, rowIdx)
		if byproductValue == "" {
			continue // 空值跳过
		}

		// 步骤6: 解析 byproduct 列中的道具ID
		itemIds := helpers.ParseCommaSeparatedIds(byproductValue)
		if len(itemIds) == 0 {
			continue // 无有效道具ID配置，跳过
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

		for _, itemId := range itemIds {
			// 检查道具 ID 是否存在
			if !validItemIds[itemId] {
				missingItemIds = append(missingItemIds, fmt.Sprintf("%d", itemId))
			}
		}

		// 步骤9: 记录错误
		if len(missingItemIds) > 0 {
			reason := fmt.Sprintf("皮肤抽奖【%s】(ID=%s) 的 byproduct 字段配置的道具 ID [%s] 在道具表|Item 中不存在",
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
		result.Reason = fmt.Sprintf("发现 %d 个皮肤抽奖副产品配置问题", len(result.ErrCells))
	}

	return result
}

// loadItemSheet 加载 Item 表数据
//
// 执行流程：
// 1. 从 sheetMap 中查找 Item 表（支持后缀匹配）
// 2. 读取表数据并返回
func (c *DrawskinByproductCheckRule) loadItemSheet(sheetMap map[string]*excelize.File) [][]string {
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
func (c *DrawskinByproductCheckRule) buildValidItemIdSet(itemCols [][]string, startRowIdx int) map[int]bool {
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
