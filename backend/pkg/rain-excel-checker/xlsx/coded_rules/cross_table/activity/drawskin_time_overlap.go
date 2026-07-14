package activity

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ActivityDrawskinTimeOverlapCheckRule Activity与DrawSkin时间范围交集检查（XC-02）
// 检查 Activity 与关联的 DrawSkin 时间范围是否存在交集
type ActivityDrawskinTimeOverlapCheckRule struct{}

func (c *ActivityDrawskinTimeOverlapCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK,
		DisplayName:    "Activity与DrawSkin时间交集检查",
		Description:    "检查Activity与关联DrawSkin的时间范围是否存在交集(XC-02)",
		TargetSheets:   []string{"DrawSkin", "Activity"},
		RequiredSheets: []string{"Activity", "DrawSkin"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

func (c *ActivityDrawskinTimeOverlapCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "Activity与DrawSkin时间交集检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	isDrawSkin := strings.HasSuffix(param.SheetName, "|DrawSkin") || param.SheetName == "DrawSkin"

	if isDrawSkin {
		c.checkFromDrawSkin(param, result)
	}
	// Activity 上下文的检查在 DrawSkin 侧已覆盖（DrawSkin.ActivityId -> Activity 的时间）

	return result
}

// checkFromDrawSkin 从 DrawSkin 侧检查时间交集
func (c *ActivityDrawskinTimeOverlapCheckRule) checkFromDrawSkin(param json_rule.CheckParam, result *json_rule.TableCheckResult) {
	// 加载 Activity 表
	actCols := c.loadSheet(param.SheetMap, "Activity")
	if actCols == nil {
		result.Ok = false
		result.Reason = "未找到 Activity 表，无法执行时间交集检查"
		return
	}

	// 构建 Activity 时间映射：Id -> {startTime, endTime, name}
	type actTimeInfo struct {
		startTime string
		endTime   string
		name      string
	}
	actMap := make(map[int]*actTimeInfo)
	actIdCol := helpers.GetColIndexByName(actCols, "Id")
	actNameCol := helpers.GetColIndexByName(actCols, "Name")
	actStartCol := helpers.GetColIndexByName(actCols, "StartTime")
	actEndCol := helpers.GetColIndexByName(actCols, "EndTime")
	if actIdCol >= 0 {
		for rowIdx := param.StartRowIdx; rowIdx < helpers.GetDataEndIndex(actCols, param.StartRowIdx); rowIdx++ {
			idStr := helpers.GetColValue(actCols, actIdCol, rowIdx)
			if idStr == "" {
				continue
			}
			id, err := helpers.ParseIntWithError(idStr)
			if err != nil {
				continue
			}
			info := &actTimeInfo{}
			if actNameCol >= 0 {
				info.name = helpers.GetColValue(actCols, actNameCol, rowIdx)
			}
			if actStartCol >= 0 {
				info.startTime = helpers.GetColValue(actCols, actStartCol, rowIdx)
			}
			if actEndCol >= 0 {
				info.endTime = helpers.GetColValue(actCols, actEndCol, rowIdx)
			}
			actMap[id] = info
		}
	}

	// 遍历 DrawSkin 检查时间交集
	dsIdCol := helpers.GetColIndexByName(param.Cols, "Id")
	dsNameCol := helpers.GetColIndexByName(param.Cols, "Name")
	dsActIdCol := helpers.GetColIndexByName(param.Cols, "ActivityId")
	dsStartCol := helpers.GetColIndexByName(param.Cols, "StartTime")
	dsEndCol := helpers.GetColIndexByName(param.Cols, "EndTime")

	if dsActIdCol == -1 {
		return
	}

	warningCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := ""
		if dsIdCol >= 0 {
			rowId = helpers.GetColValue(param.Cols, dsIdCol, rowIdx)
		}
		dsName := ""
		if dsNameCol >= 0 {
			dsName = helpers.GetColValue(param.Cols, dsNameCol, rowIdx)
		}

		actIdStr := helpers.GetColValue(param.Cols, dsActIdCol, rowIdx)
		if actIdStr == "" {
			continue
		}
		actId, err := helpers.ParseIntWithError(actIdStr)
		if err != nil || actId == 0 {
			continue
		}

		actInfo, exists := actMap[actId]
		if !exists {
			continue // 引用不存在的检查由交叉引用规则负责
		}

		// 获取 DrawSkin 的时间
		var dsStartStr, dsEndStr string
		if dsStartCol >= 0 {
			dsStartStr = helpers.GetColValue(param.Cols, dsStartCol, rowIdx)
		}
		if dsEndCol >= 0 {
			dsEndStr = helpers.GetColValue(param.Cols, dsEndCol, rowIdx)
		}

		// 两者都有时间才检查交集
		if dsStartStr == "" || dsEndStr == "" || actInfo.startTime == "" || actInfo.endTime == "" {
			continue
		}

		dsStart := excelio.ParseDate(dsStartStr)
		dsEnd := excelio.ParseDate(dsEndStr)
		actStart := excelio.ParseDate(actInfo.startTime)
		actEnd := excelio.ParseDate(actInfo.endTime)

		if dsStart.IsZero() || dsEnd.IsZero() || actStart.IsZero() || actEnd.IsZero() {
			continue
		}

		// 时间区间交集检查：DrawSkin.StartTime <= Activity.EndTime && Activity.StartTime <= DrawSkin.EndTime
		if dsStart.After(actEnd) || actStart.After(dsEnd) {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason: fmt.Sprintf("丹青阁活动【%s】(ID=%d) 的时间范围与抽奖池【%s】(ID=%s) 无交集（活动: %s~%s，抽奖池: %s~%s）",
					actInfo.name, actId, dsName, rowId,
					actInfo.startTime, actInfo.endTime,
					dsStartStr, dsEndStr),
			})
			warningCount++
		}
	}

	if warningCount > 0 {
		result.Ok = true // Warning 性质
		result.Reason = fmt.Sprintf("Activity与DrawSkin时间交集检查发现 %d 个无交集警告", warningCount)
	}
}

func (c *ActivityDrawskinTimeOverlapCheckRule) loadSheet(sheetMap map[string]*excelize.File, tableName string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, tableName); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}
