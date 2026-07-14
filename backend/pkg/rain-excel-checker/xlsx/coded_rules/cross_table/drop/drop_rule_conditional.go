package drop

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DropRuleConditionalCheckRule DropRule条件引用检查
//
// ## 校验规则
// 1. EnsureSmall > 0 时，EnsureSmallGroup 不应为空
// 2. EnsureSmall > 0 时，EnsureSmallGroup 中每个 ID 必须存在于 DropGroup 表
// 3. EnsureBig > 0 时，EnsureBigGroup 不应为空
// 4. EnsureBig > 0 时，EnsureBigGroup 中每个 ID 必须存在于 DropGroup 表
// 5. EnsureSmall > 0 时，EnsureSmallGroups 不应与 DropGroups 完全相同
// 6. EnsureItemCount > 0 时，EnsureItemID 必须 > 0
// 7. EnsureItemCount > 0 且 EnsureItemID 非 0 时，EnsureItemID 必须存在于 Item 表
//
// ## 相关表结构
// - DropRule: Id, Name, EnsureSmall, EnsureBig, EnsureSmallGroup, EnsureBigGroup, DropGroup, EnsureItemCount, EnsureItemID
// - DropGroup: Id（掉落分组表，与 DropRule 位于同一个 Drop.xlsx 中）
// - Item: Id（道具表）
//
// ## 检查流程
// 1. 加载 DropGroup 和 Item 表，构建有效 ID 集合
// 2. 遍历 DropRule 数据行，逐行检查七条条件规则
// 3. 汇总错误并返回
type DropRuleConditionalCheckRule struct{}

// Meta 返回规则元数据
func (c *DropRuleConditionalCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DROP_RULE_CONDITIONAL_CHECK,
		DisplayName:    "DropRule条件引用检查",
		Description:    "检查DropRule表中保底机制和EnsureItem的条件引用有效性",
		TargetSheets:   []string{"DropRule"},
		RequiredSheets: []string{"DropGroup", "Item"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行DropRule条件引用检查
//
// 执行流程：
// 1. 加载 DropGroup 和 Item 表，构建有效 ID 集合
// 2. 查找 DropRule 表中的关键列
// 3. 遍历数据行，逐行检查七条条件规则
// 4. 汇总错误并返回
func (c *DropRuleConditionalCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DropRule条件引用检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 查找关键列索引
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	ensureSmallColIdx := helpers.GetColIndexByName(param.Cols, "EnsureSmall")
	ensureBigColIdx := helpers.GetColIndexByName(param.Cols, "EnsureBig")
	ensureSmallGroupColIdx := helpers.GetColIndexByName(param.Cols, "EnsureSmallGroup")
	ensureBigGroupColIdx := helpers.GetColIndexByName(param.Cols, "EnsureBigGroup")
	dropGroupColIdx := helpers.GetColIndexByName(param.Cols, "DropGroup")
	ensureItemCountColIdx := helpers.GetColIndexByName(param.Cols, "EnsureItemCount")
	ensureItemIdColIdx := helpers.GetColIndexByName(param.Cols, "EnsureItemID")

	// 加载 DropGroup 表构建有效 ID 集合
	var validGroupIds map[int]bool
	dropGroupCols := c.loadSheet(param.SheetMap, "DropGroup")
	if dropGroupCols != nil {
		validGroupIds = c.buildIdSet(dropGroupCols, param.StartRowIdx)
	}

	// 加载 Item 表构建有效道具 ID 集合
	var validItemIds map[int]bool
	itemCols := c.loadSheet(param.SheetMap, "Item")
	if itemCols != nil {
		validItemIds = c.buildIdSet(itemCols, param.StartRowIdx)
	}

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		rowName := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		var errors []string

		// 获取 EnsureSmall 和 EnsureBig 的值
		ensureSmall := 0
		if ensureSmallColIdx >= 0 {
			val := helpers.GetColValue(param.Cols, ensureSmallColIdx, rowIdx)
			if val != "" {
				if n, err := strconv.Atoi(val); err == nil {
					ensureSmall = n
				}
			}
		}
		ensureBig := 0
		if ensureBigColIdx >= 0 {
			val := helpers.GetColValue(param.Cols, ensureBigColIdx, rowIdx)
			if val != "" {
				if n, err := strconv.Atoi(val); err == nil {
					ensureBig = n
				}
			}
		}

		// 1. EnsureSmall > 0 时，EnsureSmallGroup 不应为空
		if ensureSmall > 0 && ensureSmallGroupColIdx >= 0 {
			val := helpers.GetColValue(param.Cols, ensureSmallGroupColIdx, rowIdx)
			if val == "" {
				errors = append(errors, "EnsureSmall > 0 时，EnsureSmallGroup 不应为空")
			}
		}

		// 2. EnsureSmall > 0 时，EnsureSmallGroup 中每个 ID 必须存在于 DropGroup 表
		if ensureSmall > 0 && ensureSmallGroupColIdx >= 0 && validGroupIds != nil {
			val := helpers.GetColValue(param.Cols, ensureSmallGroupColIdx, rowIdx)
			if val != "" {
				if missing := c.checkIdsExist(val, validGroupIds, "EnsureSmallGroup"); len(missing) > 0 {
					errors = append(errors, missing...)
				}
			}
		}

		// 3. EnsureBig > 0 时，EnsureBigGroup 不应为空
		if ensureBig > 0 && ensureBigGroupColIdx >= 0 {
			val := helpers.GetColValue(param.Cols, ensureBigGroupColIdx, rowIdx)
			if val == "" {
				errors = append(errors, "EnsureBig > 0 时，EnsureBigGroup 不应为空")
			}
		}

		// 4. EnsureBig > 0 时，EnsureBigGroup 中每个 ID 必须存在于 DropGroup 表
		if ensureBig > 0 && ensureBigGroupColIdx >= 0 && validGroupIds != nil {
			val := helpers.GetColValue(param.Cols, ensureBigGroupColIdx, rowIdx)
			if val != "" {
				if missing := c.checkIdsExist(val, validGroupIds, "EnsureBigGroup"); len(missing) > 0 {
					errors = append(errors, missing...)
				}
			}
		}

		// 5. EnsureSmall > 0 时，EnsureSmallGroups 不应与 DropGroups 完全相同
		if ensureSmall > 0 && ensureSmallGroupColIdx >= 0 && dropGroupColIdx >= 0 {
			smallGroupVal := strings.TrimSpace(helpers.GetColValue(param.Cols, ensureSmallGroupColIdx, rowIdx))
			dropGroupVal := strings.TrimSpace(helpers.GetColValue(param.Cols, dropGroupColIdx, rowIdx))
			if smallGroupVal != "" && dropGroupVal != "" && smallGroupVal == dropGroupVal {
				errors = append(errors, "EnsureSmallGroup 与 DropGroup 完全相同，小保底组应与普通组有所区别")
			}
		}

		// 6. EnsureItemCount > 0 时，EnsureItemID 必须 > 0
		ensureItemCount := 0
		if ensureItemCountColIdx >= 0 {
			val := helpers.GetColValue(param.Cols, ensureItemCountColIdx, rowIdx)
			if val != "" {
				if n, err := strconv.Atoi(val); err == nil {
					ensureItemCount = n
				}
			}
		}
		ensureItemId := 0
		if ensureItemIdColIdx >= 0 {
			val := helpers.GetColValue(param.Cols, ensureItemIdColIdx, rowIdx)
			if val != "" {
				if n, err := strconv.Atoi(val); err == nil {
					ensureItemId = n
				}
			}
		}

		if ensureItemCount > 0 && ensureItemId <= 0 {
			errors = append(errors, fmt.Sprintf("EnsureItemCount=%d > 0 时，EnsureItemID 必须 > 0", ensureItemCount))
		}

		// 7. EnsureItemCount > 0 且 EnsureItemID 非 0 时，EnsureItemID 必须存在于 Item 表
		if ensureItemCount > 0 && ensureItemId > 0 && validItemIds != nil {
			if !validItemIds[ensureItemId] {
				errors = append(errors, fmt.Sprintf("EnsureItemID=%d 在道具表中不存在", ensureItemId))
			}
		}

		if len(errors) > 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("掉落规则【%s】(ID=%s) %s", rowName, rowId, strings.Join(errors, "; ")),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个掉落规则条件引用问题", len(result.ErrCells))
	}

	return result
}

// loadSheet 从 sheetMap 中加载指定表的数据
func (c *DropRuleConditionalCheckRule) loadSheet(sheetMap map[string]*excelize.File, suffix string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, suffix); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildIdSet 从指定表的 Id 列构建有效 ID 集合
func (c *DropRuleConditionalCheckRule) buildIdSet(cols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(cols, "Id")
	if idColIdx < 0 {
		return validIds
	}
	endIdx := helpers.GetDataEndIndex(cols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < endIdx; rowIdx++ {
		idStr := helpers.GetColValue(cols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		validIds[id] = true
	}
	return validIds
}

// checkIdsExist 检查逗号分隔的 ID 列表中每个 ID 是否存在于有效 ID 集合中
func (c *DropRuleConditionalCheckRule) checkIdsExist(value string, validIds map[int]bool, label string) []string {
	parts := strings.Split(value, ",")
	var missing []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			missing = append(missing, fmt.Sprintf("%s中的ID【%s】不是有效数字", label, part))
			continue
		}
		if !validIds[id] {
			missing = append(missing, fmt.Sprintf("%s中的ID【%d】在掉落分组表中不存在", label, id))
		}
	}
	return missing
}
