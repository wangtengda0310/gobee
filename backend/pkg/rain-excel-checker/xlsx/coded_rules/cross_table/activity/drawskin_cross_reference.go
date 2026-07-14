package activity

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ActivityDrawskinCrossReferenceCheckRule Activity与DrawSkin交叉引用检查
// DSK-05: DrawSkin.ActivityId 引用 Activity 表
// XC-01: Activity↔DrawSkin 双向引用一致性
// XC-04: DrawSkin.ActivityId 对应的 Activity 应为丹青阁类型
type ActivityDrawskinCrossReferenceCheckRule struct{}

// 丹青阁活动类型标识，对应服务端 EActivityType_ActTypeSkinRaffle
const crossRefDanQingGeActivityType = "ActTypeSkinRaffle"

func (c *ActivityDrawskinCrossReferenceCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK,
		DisplayName:    "Activity与DrawSkin交叉引用检查",
		Description:    "检查Activity和DrawSkin之间的引用完整性：ActivityId引用(DSK-05)、双向引用一致(XC-01)、活动类型校验(XC-04)",
		TargetSheets:   []string{"DrawSkin", "Activity"},
		RequiredSheets: []string{"Activity", "DrawSkin"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

func (c *ActivityDrawskinCrossReferenceCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "Activity与DrawSkin交叉引用检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	isDrawSkin := strings.HasSuffix(param.SheetName, "|DrawSkin") || param.SheetName == "DrawSkin"

	if isDrawSkin {
		c.checkFromDrawSkin(param, result)
	} else {
		c.checkFromActivity(param, result)
	}

	return result
}

// checkFromDrawSkin DrawSkin 表变更时检查 DSK-05 + XC-04
func (c *ActivityDrawskinCrossReferenceCheckRule) checkFromDrawSkin(param json_rule.CheckParam, result *json_rule.TableCheckResult) {
	// 加载 Activity 表
	actCols := c.loadSheet(param.SheetMap, "Activity")
	if actCols == nil {
		result.Ok = false
		result.Reason = "未找到 Activity 表，无法执行交叉引用检查"
		return
	}

	// 构建 Activity 数据映射：Id -> {ActivityType, Name}
	type actInfo struct {
		activityType string
		name         string
	}
	actMap := make(map[int]*actInfo)
	idCol := helpers.GetColIndexByName(actCols, "Id")
	typeCol := helpers.GetColIndexByName(actCols, "ActivityType")
	nameCol := helpers.GetColIndexByName(actCols, "Name")
	if idCol >= 0 {
		for rowIdx := param.StartRowIdx; rowIdx < helpers.GetDataEndIndex(actCols, param.StartRowIdx); rowIdx++ {
			idStr := helpers.GetColValue(actCols, idCol, rowIdx)
			if idStr == "" {
				continue
			}
			id, err := helpers.ParseIntWithError(idStr)
			if err != nil {
				continue
			}
			info := &actInfo{}
			if typeCol >= 0 {
				info.activityType = helpers.GetColValue(actCols, typeCol, rowIdx)
			}
			if nameCol >= 0 {
				info.name = helpers.GetColValue(actCols, nameCol, rowIdx)
			}
			actMap[id] = info
		}
	}

	// 遍历 DrawSkin 检查
	dsIdCol := helpers.GetColIndexByName(param.Cols, "Id")
	dsNameCol := helpers.GetColIndexByName(param.Cols, "Name")
	dsActIdCol := helpers.GetColIndexByName(param.Cols, "ActivityId")

	if dsActIdCol == -1 {
		return
	}

	errorCount := 0
	warningCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := ""
		if dsIdCol >= 0 {
			rowId = helpers.GetColValue(param.Cols, dsIdCol, rowIdx)
		}
		name := ""
		if dsNameCol >= 0 {
			name = helpers.GetColValue(param.Cols, dsNameCol, rowIdx)
		}

		actIdStr := helpers.GetColValue(param.Cols, dsActIdCol, rowIdx)
		if actIdStr == "" {
			continue
		}
		actId, err := helpers.ParseIntWithError(actIdStr)
		if err != nil || actId == 0 {
			continue
		}

		// DSK-05: ActivityId 引用 Activity 表
		actInfo, exists := actMap[actId]
		if !exists {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 ActivityId=%d 在活动表中不存在", name, rowId, actId),
			})
			errorCount++
			continue
		}

		// XC-04: ActivityId 对应的 Activity 应为丹青阁类型
		if actInfo.activityType != crossRefDanQingGeActivityType {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 ActivityId=%d 对应的活动【%s】类型为 %s，不是丹青阁类型(ActTypeSkinRaffle)", name, rowId, actId, actInfo.name, actInfo.activityType),
			})
			warningCount++
		}
	}

	if errorCount > 0 {
		result.Ok = false
	}
	if errorCount > 0 || warningCount > 0 {
		parts := make([]string, 0, 2)
		if errorCount > 0 {
			parts = append(parts, fmt.Sprintf("%d 个引用不存在", errorCount))
		}
		if warningCount > 0 {
			parts = append(parts, fmt.Sprintf("%d 个类型不一致", warningCount))
		}
		result.Reason = fmt.Sprintf("Activity与DrawSkin交叉引用检查: %s", strings.Join(parts, "，"))
	}
}

// checkFromActivity Activity 表变更时检查 XC-01
func (c *ActivityDrawskinCrossReferenceCheckRule) checkFromActivity(param json_rule.CheckParam, result *json_rule.TableCheckResult) {
	// 加载 DrawSkin 表
	drawCols := c.loadSheet(param.SheetMap, "DrawSkin")
	if drawCols == nil {
		result.Ok = false
		result.Reason = "未找到 DrawSkin 表，无法执行交叉引用检查"
		return
	}

	// 构建 DrawSkin 数据：Id -> ActivityId
	drawToAct := make(map[int]int)
	dsIdCol := helpers.GetColIndexByName(drawCols, "Id")
	dsActIdCol := helpers.GetColIndexByName(drawCols, "ActivityId")
	if dsIdCol >= 0 && dsActIdCol >= 0 {
		for rowIdx := param.StartRowIdx; rowIdx < helpers.GetDataEndIndex(drawCols, param.StartRowIdx); rowIdx++ {
			dsIdStr := helpers.GetColValue(drawCols, dsIdCol, rowIdx)
			actIdStr := helpers.GetColValue(drawCols, dsActIdCol, rowIdx)
			if dsIdStr == "" {
				continue
			}
			dsId, err := helpers.ParseIntWithError(dsIdStr)
			if err != nil {
				continue
			}
			actId, _ := helpers.ParseIntWithError(actIdStr)
			drawToAct[dsId] = actId
		}
	}

	// 遍历 Activity 表中丹青阁活动
	actIdCol := helpers.GetColIndexByName(param.Cols, "Id")
	actNameCol := helpers.GetColIndexByName(param.Cols, "Name")
	actTypeCol := helpers.GetColIndexByName(param.Cols, "ActivityType")
	customCol := helpers.GetColIndexByName(param.Cols, "CustomParma")

	warningCount := 0

	// 缺少 ActivityType 列时无法识别丹青阁活动，跳过检查
	if actTypeCol == -1 {
		return
	}

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		actType := helpers.GetColValue(param.Cols, actTypeCol, rowIdx)
		if actType != crossRefDanQingGeActivityType {
			continue
		}

		if customCol == -1 {
			continue
		}
		customStr := helpers.GetColValue(param.Cols, customCol, rowIdx)
		if customStr == "" {
			continue
		}
		drawId, err := helpers.ParseIntWithError(customStr)
		if err != nil {
			continue
		}

		rowId := ""
		if actIdCol >= 0 {
			rowId = helpers.GetColValue(param.Cols, actIdCol, rowIdx)
		}
		name := ""
		if actNameCol >= 0 {
			name = helpers.GetColValue(param.Cols, actNameCol, rowIdx)
		}

		// XC-01: 检查 DrawSkin 的 ActivityId 是否反向指向该 Activity
		linkedActId, exists := drawToAct[drawId]
		if !exists {
			continue // 引用不存在的问题由 ACT-03 规则检查
		}
		rowActId, _ := helpers.ParseIntWithError(rowId)
		if rowActId > 0 && linkedActId != rowActId {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】(ID=%s) 引用 DrawSkin(ID=%d)，但该 DrawSkin 的 ActivityId=%d 不指向原活动(ID=%s)", name, rowId, drawId, linkedActId, rowId),
			})
			warningCount++
		}
	}

	if warningCount > 0 {
		result.Ok = true // Warning 性质
		result.Reason = fmt.Sprintf("Activity与DrawSkin交叉引用检查发现 %d 个双向引用不一致", warningCount)
	}
}

func (c *ActivityDrawskinCrossReferenceCheckRule) loadSheet(sheetMap map[string]*excelize.File, tableName string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, tableName); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}
