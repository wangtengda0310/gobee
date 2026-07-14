package season_pass_task

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// SeasonPassTaskDiff 赛季战令任务表差异数据结构
type SeasonPassTaskDiff struct {
	Id              int
	SeasonPassId    int
	Name            string
	Class           string
	SubType         string
	IsSpecial       bool
	IsAutoAccept    bool
	AcceptCond      int
	AcceptCondParam string
	CompleteCond    string
	Reward          []ItemCfg
	PassExp         int
	IgnoreExpLimit  bool
	IsAutoSubmit    bool
	ResetType       int
	SendStatus      int
	ShowType        int
	ShowPos         string
	ShowDate        string
	BeginDate       string
	EndDate         string
	ExpireDate      string
}

// ItemCfg 道具配置
type ItemCfg struct {
	ItemId int
	Count  int
}

func (s SeasonPassTaskDiff) GetType() string {
	return "SeasonPassTaskDiff"
}

func (s SeasonPassTaskDiff) GetDisplayName() string {
	return s.Name
}

// GetSeasonPassTaskDiffMap 解析赛季战令任务表数据
func GetSeasonPassTaskDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SeasonPassTaskDiff, err error) {
	var taskCols [][]string
	if sheet, exist := sheetMap["战令任务表|SeasonPassTask"]; exist {
		var err error
		taskCols, err = sheet.GetCols("战令任务表|SeasonPassTask")
		if err != nil {
			return nil, err
		}
	}

	// 如果没有找到sheet，返回空数组
	if taskCols == nil || len(taskCols) == 0 {
		return &[]SeasonPassTaskDiff{}, nil
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM

	tasksDiff := make([]SeasonPassTaskDiff, 0, 200)

	// 检查Id列是否有数据
	if len(taskCols) <= Id || len(taskCols[Id]) <= startRow {
		return &tasksDiff, nil
	}

	endIndex := helpers.AutoDetectEndIndex(taskCols, Id, startRow, 3)
	idCol := taskCols[Id][startRow:endIndex]

	for i, idStr := range idCol {
		// 跳过注释行（以#开头的ID）
		idStr = strings.TrimSpace(idStr)
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			taskDiff := SeasonPassTaskDiff{}

			seasonPassId := utils.GetCellValue(taskCols, SeasonPassId, startRow+i)
			name := utils.GetCellValue(taskCols, Name, startRow+i)
			class := utils.GetCellValue(taskCols, Class, startRow+i)
			subType := utils.GetCellValue(taskCols, SubType, startRow+i)
			isSpecial := utils.GetCellValue(taskCols, IsSpecial, startRow+i)
			isAutoAccept := utils.GetCellValue(taskCols, IsAutoAccept, startRow+i)
			acceptCond := utils.GetCellValue(taskCols, AcceptCond, startRow+i)
			acceptCondParam := utils.GetCellValue(taskCols, AcceptCondParam, startRow+i)
			completeCond := utils.GetCellValue(taskCols, CompleteCond, startRow+i)
			reward := utils.GetCellValue(taskCols, Reward, startRow+i)
			passExp := utils.GetCellValue(taskCols, PassExp, startRow+i)
			ignoreExpLimit := utils.GetCellValue(taskCols, IgnoreExpLimit, startRow+i)
			isAutoSubmit := utils.GetCellValue(taskCols, IsAutoSubmit, startRow+i)
			resetType := utils.GetCellValue(taskCols, ResetType, startRow+i)
			sendStatus := utils.GetCellValue(taskCols, SendStatus, startRow+i)
			showType := utils.GetCellValue(taskCols, ShowType, startRow+i)
			showPos := utils.GetCellValue(taskCols, ShowPos, startRow+i)
			showDate := utils.GetCellValue(taskCols, ShowDate, startRow+i)
			beginDate := utils.GetCellValue(taskCols, BeginDate, startRow+i)
			endDate := utils.GetCellValue(taskCols, EndDate, startRow+i)
			expireDate := utils.GetCellValue(taskCols, ExpireDate, startRow+i)

			taskDiff.Id = id

			if n, err := strconv.Atoi(seasonPassId); err == nil {
				taskDiff.SeasonPassId = n
			} else {
				taskDiff.SeasonPassId = -1
			}

			taskDiff.Name = name
			taskDiff.Class = class
			taskDiff.SubType = subType

			taskDiff.IsSpecial = strings.ToLower(isSpecial) == "true"
			taskDiff.IsAutoAccept = strings.ToLower(isAutoAccept) == "true"

			if n, err := strconv.Atoi(acceptCond); err == nil {
				taskDiff.AcceptCond = n
			} else {
				taskDiff.AcceptCond = -1
			}

			taskDiff.AcceptCondParam = acceptCondParam
			taskDiff.CompleteCond = completeCond

			taskDiff.Reward = parseItemCfgList(reward)

			if n, err := strconv.Atoi(passExp); err == nil {
				taskDiff.PassExp = n
			} else {
				taskDiff.PassExp = -1
			}

			taskDiff.IgnoreExpLimit = strings.ToLower(ignoreExpLimit) == "true"
			taskDiff.IsAutoSubmit = strings.ToLower(isAutoSubmit) == "true"

			if n, err := strconv.Atoi(resetType); err == nil {
				taskDiff.ResetType = n
			} else {
				taskDiff.ResetType = -1
			}

			if n, err := strconv.Atoi(sendStatus); err == nil {
				taskDiff.SendStatus = n
			} else {
				taskDiff.SendStatus = -1
			}

			if n, err := strconv.Atoi(showType); err == nil {
				taskDiff.ShowType = n
			} else {
				taskDiff.ShowType = -1
			}

			taskDiff.ShowPos = showPos
			taskDiff.ShowDate = showDate
			taskDiff.BeginDate = beginDate
			taskDiff.EndDate = endDate
			taskDiff.ExpireDate = expireDate

			tasksDiff = append(tasksDiff, taskDiff)
		}
	}

	return &tasksDiff, nil
}

// parseItemCfgList 解析多条道具配置字符串，格式为 {itemId;count},{itemId;count}
func parseItemCfgList(itemStr string) []ItemCfg {
	result := make([]ItemCfg, 0, 10)
	if itemStr == "" {
		return result
	}

	itemStr = strings.TrimSpace(itemStr)

	// 按逗号分割多个道具配置
	items := strings.Split(itemStr, "},{")
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.TrimPrefix(item, "{")
		item = strings.TrimSuffix(item, "}")

		parts := strings.Split(item, ";")
		if len(parts) >= 2 {
			itemId, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			if count == 0 {
				count = 1
			}
			result = append(result, ItemCfg{
				ItemId: itemId,
				Count:  count,
			})
		}
	}

	return result
}
