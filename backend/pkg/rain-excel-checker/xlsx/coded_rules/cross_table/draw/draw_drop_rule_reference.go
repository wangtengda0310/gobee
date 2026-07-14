package draw

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DrawDropRuleReferenceCheckRule 抽奖池掉落规则引用检查
//
// ## 校验规则
// 1. Draw/DrawSkin 表中 OnceDropRule 列的掉落规则ID必须存在于 DropRule 表的 Id 列
// 2. Draw/DrawSkin 表中 TenDropRule 列的掉落规则ID必须存在于 DropRule 表的 Id 列
//
// ## 相关表结构
// - Draw/DrawSkin: Id, Name, OnceDropRule, TenDropRule
// - DropRule: Id（掉落规则表，位于 Drop.xlsx 中）
//
// ## 检查流程
// 1. 加载 DropRule 表，构建有效掉落规则 ID 集合
// 2. 遍历 Draw/DrawSkin 表的每一行
// 3. 分别检查 OnceDropRule 和 TenDropRule 列的引用
// 4. 记录所有不存在的引用到错误列表
type DrawDropRuleReferenceCheckRule struct{}

// Meta 返回规则元数据
func (c *DrawDropRuleReferenceCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DRAW_DROP_RULE_REFERENCE_CHECK,
		DisplayName:    "抽奖池掉落规则引用检查",
		Description:    "检查 Draw/DrawSkin 表中 OnceDropRule 和 TenDropRule 引用的掉落规则ID是否在 DropRule 表中存在",
		TargetSheets:   []string{"Draw", "DrawSkin"},
		RequiredSheets: []string{"DropRule"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行抽奖池掉落规则引用检查
//
// 执行流程：
// 1. 加载 DropRule 表数据，构建有效掉落规则 ID 集合
// 2. 查找 Draw/DrawSkin 表中的关键列（OnceDropRule、TenDropRule、Id、Name）
// 3. 遍历数据行，逐行检查 OnceDropRule 和 TenDropRule 引用
// 4. 汇总错误并返回
func (c *DrawDropRuleReferenceCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "抽奖池掉落规则引用检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 加载 DropRule 表（位于 Drop.xlsx 的 "掉落规则表|DropRule" sheet 中）
	dropRuleCols := c.loadSheet(param.SheetMap, "DropRule")
	if dropRuleCols == nil {
		result.Ok = false
		result.Reason = "未找到 DropRule 表（掉落规则表），无法执行掉落规则引用检查"
		return result
	}

	// 构建有效掉落规则 ID 集合
	validDropRuleIds := c.buildValidDropRuleIdSet(dropRuleCols, param.StartRowIdx)

	// 查找源表关键列
	onceDropColIdx := helpers.GetColIndexByName(param.Cols, "OnceDropRule")
	tenDropColIdx := helpers.GetColIndexByName(param.Cols, "TenDropRule")
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")

	// 遍历检查
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		rowName := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		var errors []string

		// 检查 OnceDropRule 列
		if onceDropColIdx >= 0 {
			if errs := c.checkDropRuleRef(param.Cols, onceDropColIdx, rowIdx, rowId, rowName, "单抽掉落规则", validDropRuleIds); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		// 检查 TenDropRule 列
		if tenDropColIdx >= 0 {
			if errs := c.checkDropRuleRef(param.Cols, tenDropColIdx, rowIdx, rowId, rowName, "十连抽掉落规则", validDropRuleIds); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		if len(errors) > 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   strings.Join(errors, "; "),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个抽奖池的掉落规则引用问题", len(result.ErrCells))
	}

	return result
}

// loadSheet 从 sheetMap 中加载指定表的数据
func (c *DrawDropRuleReferenceCheckRule) loadSheet(sheetMap map[string]*excelize.File, suffix string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, suffix); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildValidDropRuleIdSet 从 DropRule 表构建有效的掉落规则 ID 集合
func (c *DrawDropRuleReferenceCheckRule) buildValidDropRuleIdSet(dropRuleCols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(dropRuleCols, "Id")
	if idColIdx < 0 {
		return validIds
	}

	endIdx := helpers.GetDataEndIndex(dropRuleCols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < endIdx; rowIdx++ {
		idStr := helpers.GetColValue(dropRuleCols, idColIdx, rowIdx)
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

// checkDropRuleRef 检查单行的掉落规则引用
func (c *DrawDropRuleReferenceCheckRule) checkDropRuleRef(cols [][]string, colIdx, rowIdx int, rowId, rowName, label string, validDropRuleIds map[int]bool) []string {
	value := helpers.GetColValue(cols, colIdx, rowIdx)
	if value == "" {
		return nil
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return []string{fmt.Sprintf("抽奖池【%s】(ID=%s) 的%s值【%s】不是有效数字", rowName, rowId, label, value)}
	}

	if !validDropRuleIds[id] {
		return []string{fmt.Sprintf("抽奖池【%s】(ID=%s) 的%sID【%s】在掉落规则表中不存在", rowName, rowId, label, value)}
	}
	return nil
}
