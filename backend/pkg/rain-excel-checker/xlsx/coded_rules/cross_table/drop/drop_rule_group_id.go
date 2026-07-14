package drop

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DropRuleGroupIdCheckRule 掉落规则组ID引用检查
//
// ## 校验规则
// 1. DropRule 表中 DropGroup 列引用的掉落组ID必须存在于 DropGroup 表的 Id 列
// 2. DropRule 表中 EnsureSmallGroup 列引用的保底组ID必须存在于 DropGroup 表的 Id 列
// 3. DropRule 表中 EnsureBigGroup 列引用的保底组ID必须存在于 DropGroup 表的 Id 列
//
// ## 相关表结构
// - DropRule: Id, Name, DropGroup, EnsureSmallGroup, EnsureBigGroup
// - DropGroup: Id（掉落分组表，与 DropRule 位于同一个 Drop.xlsx 中）
//
// ## 检查流程
// 1. 加载 DropGroup 表，构建有效掉落组 ID 集合
// 2. 遍历 DropRule 表的每一行
// 3. 分别检查 DropGroup、EnsureSmallGroup、EnsureBigGroup 列的引用
// 4. 记录所有不存在的引用到错误列表
type DropRuleGroupIdCheckRule struct{}

// Meta 返回规则元数据
func (c *DropRuleGroupIdCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DROP_RULE_GROUP_ID_CHECK,
		DisplayName:    "掉落规则组ID引用检查",
		Description:    "检查 DropRule 表中 DropGroup、EnsureSmallGroup、EnsureBigGroup 引用的掉落组ID是否在 DropGroup 表中存在",
		TargetSheets:   []string{"DropRule"},
		RequiredSheets: []string{"DropGroup"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行掉落规则组ID引用检查
//
// 执行流程：
// 1. 加载 DropGroup 表数据，构建有效掉落组 ID 集合
// 2. 查找 DropRule 表中的关键列
// 3. 遍历数据行，逐行检查三列引用
// 4. 汇总错误并返回
func (c *DropRuleGroupIdCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "掉落规则组ID引用检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 加载 DropGroup 表（位于 Drop.xlsx 的 "掉落分组表|DropGroup" sheet 中）
	dropGroupCols := c.loadSheet(param.SheetMap, "DropGroup")
	if dropGroupCols == nil {
		result.Ok = false
		result.Reason = "未找到 DropGroup 表（掉落分组表），无法执行组ID引用检查"
		return result
	}

	// 构建有效掉落组 ID 集合
	validGroupIds := c.buildValidGroupIdSet(dropGroupCols, param.StartRowIdx)

	// 查找 DropRule 表关键列
	dropGroupColIdx := helpers.GetColIndexByName(param.Cols, "DropGroup")
	ensureSmallColIdx := helpers.GetColIndexByName(param.Cols, "EnsureSmallGroup")
	ensureBigColIdx := helpers.GetColIndexByName(param.Cols, "EnsureBigGroup")
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")

	// 遍历检查
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		rowName := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		var errors []string

		// 检查 DropGroup 列（主掉落组，可能为逗号分隔的多个ID）
		if dropGroupColIdx >= 0 {
			if errs := c.checkGroupIds(param.Cols, dropGroupColIdx, rowIdx, rowId, rowName, "掉落组", validGroupIds); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		// 检查 EnsureSmallGroup 列（小保底组）
		if ensureSmallColIdx >= 0 {
			if errs := c.checkGroupIds(param.Cols, ensureSmallColIdx, rowIdx, rowId, rowName, "小保底组", validGroupIds); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		// 检查 EnsureBigGroup 列（大保底组）
		if ensureBigColIdx >= 0 {
			if errs := c.checkGroupIds(param.Cols, ensureBigColIdx, rowIdx, rowId, rowName, "大保底组", validGroupIds); len(errs) > 0 {
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
		result.Reason = fmt.Sprintf("发现 %d 个掉落规则的组ID引用问题", len(result.ErrCells))
	}

	return result
}

// loadSheet 从 sheetMap 中加载指定表的数据
func (c *DropRuleGroupIdCheckRule) loadSheet(sheetMap map[string]*excelize.File, suffix string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, suffix); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildValidGroupIdSet 从 DropGroup 表构建有效的掉落组 ID 集合
func (c *DropRuleGroupIdCheckRule) buildValidGroupIdSet(dropGroupCols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(dropGroupCols, "Id")
	if idColIdx < 0 {
		return validIds
	}

	endIdx := helpers.GetDataEndIndex(dropGroupCols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < endIdx; rowIdx++ {
		idStr := helpers.GetColValue(dropGroupCols, idColIdx, rowIdx)
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

// checkGroupIds 检查单行的组ID引用（支持逗号分隔的多个ID）
func (c *DropRuleGroupIdCheckRule) checkGroupIds(cols [][]string, colIdx, rowIdx int, rowId, rowName, label string, validGroupIds map[int]bool) []string {
	value := helpers.GetColValue(cols, colIdx, rowIdx)
	if value == "" {
		return nil
	}

	// 解析逗号分隔的组ID
	parts := strings.Split(value, ",")
	var missing []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			missing = append(missing, fmt.Sprintf("%sID【%s】不是有效数字", label, part))
			continue
		}
		if !validGroupIds[id] {
			missing = append(missing, fmt.Sprintf("%sID【%s】在掉落分组表中不存在", label, part))
		}
	}

	if len(missing) > 0 {
		return []string{fmt.Sprintf("掉落规则【%s】(ID=%s) %s", rowName, rowId, strings.Join(missing, ", "))}
	}
	return nil
}
