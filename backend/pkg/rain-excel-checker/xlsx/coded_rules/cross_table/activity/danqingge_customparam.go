// Package activity 提供活动相关的跨表校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package activity

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ActivityDanqinggeCustomParamIsItemidCheckRule 丹青阁活动自定义参数检查规则
//
// ## 校验规则
// 1. ACT-02: 丹青阁活动的 CustomParma 不应为空（Warning）
// 2. 活动表|Activity 中丹青阁活动的 CustomParma 列值必须是 DrawSkin 表中存在的 ID
// 3. 丹青阁活动的识别：Name 列包含"丹青阁"关键字
// 4. CustomParma 列值对应 DrawSkin 表中的 Id 列
//
// ## 相关表结构
// - Activity: Id, Name, CustomParma (丹青阁活动的自定义参数)
// - DrawSkin: Id, Name (皮肤抽奖池ID和名称)
//
// ## 检查流程
// 1. 加载 Activity 表和 DrawSkin 表
// 2. 从 DrawSkin 表提取所有有效的皮肤抽奖池ID
// 3. 在 Activity 表中查找丹青阁活动
// 4. ACT-02: 检查 CustomParma 是否为空（Warning）
// 5. 检查丹青阁活动的 CustomParma 值是否存在于 DrawSkin 表中
// 6. 如果不存在，记录错误
type ActivityDanqinggeCustomParamIsItemidCheckRule struct{}

// Meta 返回规则元数据
func (c *ActivityDanqinggeCustomParamIsItemidCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK,
		DisplayName:    "丹青阁活动自定义参数检查",
		Description:    "检查活动表|Activity 中丹青阁活动的 CustomParma 非空(ACT-02)且值在 DrawSkin 表中存在(ACT-03)",
		TargetSheets:   []string{"Activity"},
		RequiredSheets: []string{"DrawSkin"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行丹青阁活动自定义参数检查
func (c *ActivityDanqinggeCustomParamIsItemidCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "丹青阁活动自定义参数检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 加载 DrawSkin 表数据
	drawCols := c.loadDrawSkinSheet(param.SheetMap)
	if drawCols == nil {
		result.Ok = false
		result.Reason = "未找到 DrawSkin 表，无法执行皮肤抽奖池ID存在性检查"
		return result
	}

	// 从 DrawSkin 表构建有效的皮肤抽奖池ID 集合
	validDrawIds := c.buildValidDrawIdSet(drawCols, param.StartRowIdx)

	// 查找 Activity 表中的关键列索引
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	customParmaColIdx := helpers.GetColIndexByName(param.Cols, "CustomParma")

	if customParmaColIdx < 0 {
		result.Ok = true
		result.Reason = "Activity 表中未找到 CustomParma 列"
		return result
	}

	// 遍历 Activity 表，查找丹青阁活动并检查
	danqinggeFound := false
	hasError := false // 区分 Error（引用不存在/格式错误）和 Warning（ACT-02 未配置）
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		nameValue := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		if nameValue == "" {
			continue
		}

		if !c.isDanqinggeActivity(nameValue) {
			continue
		}

		danqinggeFound = true

		rowId := ""
		if idColIdx >= 0 {
			rowId = helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		}
		customParmaValue := helpers.GetColValue(param.Cols, customParmaColIdx, rowIdx)

		// ACT-02: 丹青阁活动的 CustomParma 不应为空（Warning，Ok 保持 true）
		if customParmaValue == "" {
			reason := fmt.Sprintf("丹青阁活动【%s】(ID=%s) 的 CustomParma 未配置，丹青阁活动应关联皮肤抽奖池",
				nameValue, rowId)
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   reason,
			})
			continue
		}

		// 解析 CustomParma 值
		drawId, err := helpers.ParseIntWithError(customParmaValue)
		if err != nil {
			reason := fmt.Sprintf("活动【%s】(ID=%s) 的 CustomParma 值【%s】不是有效的数字格式",
				nameValue, rowId, customParmaValue)
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   reason,
			})
			hasError = true
			continue
		}

		// 检查皮肤抽奖池ID 是否存在于 DrawSkin 表中
		if !validDrawIds[drawId] {
			reason := fmt.Sprintf("活动【%s】(ID=%s) 的 CustomParma 值【%d】在 DrawSkin 表（皮肤抽奖表）中不存在，请检查配置",
				nameValue, rowId, drawId)
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   reason,
			})
			hasError = true
		}
	}

	// 如果没有找到丹青阁活动，返回成功（不强制要求）
	if !danqinggeFound {
		result.Ok = true
		result.Reason = "Activity 表中未找到丹青阁活动"
		return result
	}

	// 设置检查结果：Warning（ACT-02）不影响 Ok，Error（引用不存在）设 Ok=false
	if len(result.ErrCells) > 0 {
		if hasError {
			result.Ok = false
		}
		result.Reason = fmt.Sprintf("丹青阁活动自定义参数检查发现 %d 个问题", len(result.ErrCells))
	}

	return result
}

// loadDrawSkinSheet 加载 DrawSkin 表数据
func (c *ActivityDanqinggeCustomParamIsItemidCheckRule) loadDrawSkinSheet(sheetMap map[string]*excelize.File) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "DrawSkin"); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildValidDrawIdSet 从 DrawSkin 表构建有效的皮肤抽奖池ID 集合
func (c *ActivityDanqinggeCustomParamIsItemidCheckRule) buildValidDrawIdSet(drawCols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)

	idColIdx := helpers.GetColIndexByName(drawCols, "Id")
	if idColIdx < 0 {
		return validIds
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(drawCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(drawCols, idColIdx, rowIdx)
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

// isDanqinggeActivity 判断是否是丹青阁活动
func (c *ActivityDanqinggeCustomParamIsItemidCheckRule) isDanqinggeActivity(activityName string) bool {
	return strings.Contains(activityName, "丹青阁")
}
