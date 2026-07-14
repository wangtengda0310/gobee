package robot_action

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type RobotActionDiff struct {
	Id               string
	None             string
	Action           []string
	TargetNum        []int
	TargetType       []string
	CardNum          []int
	CardFromType     []string
	TransCardSkill   []int
	DefaultCardSkill []int
}

func (r RobotActionDiff) GetID() string {
	return r.Id
}

func (r RobotActionDiff) GetType() string {
	return "RobotActionDiff"
}

func (r RobotActionDiff) GetDisplayName() string {
	return r.Id
}

func GetRobotActionDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]RobotActionDiff, err error) {
	var robotActionCols [][]string // 人机行动表
	if sheet, exist := sheetMap["人机行动表|RobotAction"]; exist {
		var err error
		robotActionCols, err = sheet.GetCols("人机行动表|RobotAction")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建人机行动Map
	robotActionsDiff := make([]RobotActionDiff, 0, 100)

	for i, idStr := range robotActionCols[Id][startRow:helpers.AutoDetectEndIndex(robotActionCols, Id, startRow, 3)] {
		// 第一列是ESkillId类型（枚举），判断不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		robotActionDiff := RobotActionDiff{}

		// 获取所有列的值
		none := utils.GetCellValue(robotActionCols, None, startRow+i)
		action := utils.GetCellValue(robotActionCols, Action, startRow+i)
		targetNum := utils.GetCellValue(robotActionCols, TargetNum, startRow+i)
		targetType := utils.GetCellValue(robotActionCols, TargetType, startRow+i)
		cardNum := utils.GetCellValue(robotActionCols, CardNum, startRow+i)
		cardFromType := utils.GetCellValue(robotActionCols, CardFromType, startRow+i)
		transCardSkill := utils.GetCellValue(robotActionCols, TransCardSkill, startRow+i)
		defaultCardSkill := utils.GetCellValue(robotActionCols, DefaultCardSkill, startRow+i)

		// 录入数据
		robotActionDiff.Id = idStr
		robotActionDiff.None = none

		// Action - 数组类型
		robotActionDiff.Action = make([]string, 0, 5)
		for _, ss := range strings.Split(action, ",") {
			if ss != "" {
				robotActionDiff.Action = append(robotActionDiff.Action, ss)
			}
		}

		// TargetNum - 数组类型
		robotActionDiff.TargetNum = make([]int, 0, 5)
		for _, ss := range strings.Split(targetNum, ",") {
			if n, err := strconv.Atoi(ss); err == nil && ss != "" {
				robotActionDiff.TargetNum = append(robotActionDiff.TargetNum, n)
			}
		}

		// TargetType - 数组类型
		robotActionDiff.TargetType = make([]string, 0, 5)
		for _, ss := range strings.Split(targetType, ",") {
			if ss != "" {
				robotActionDiff.TargetType = append(robotActionDiff.TargetType, ss)
			}
		}

		// CardNum - 数组类型
		robotActionDiff.CardNum = make([]int, 0, 5)
		for _, ss := range strings.Split(cardNum, ",") {
			if n, err := strconv.Atoi(ss); err == nil && ss != "" {
				robotActionDiff.CardNum = append(robotActionDiff.CardNum, n)
			}
		}

		// CardFromType - 数组类型
		robotActionDiff.CardFromType = make([]string, 0, 5)
		for _, ss := range strings.Split(cardFromType, ",") {
			if ss != "" {
				robotActionDiff.CardFromType = append(robotActionDiff.CardFromType, ss)
			}
		}

		// TransCardSkill - 数组类型
		robotActionDiff.TransCardSkill = make([]int, 0, 5)
		for _, ss := range strings.Split(transCardSkill, ",") {
			if n, err := strconv.Atoi(ss); err == nil && ss != "" {
				robotActionDiff.TransCardSkill = append(robotActionDiff.TransCardSkill, n)
			}
		}

		// DefaultCardSkill - 数组类型
		robotActionDiff.DefaultCardSkill = make([]int, 0, 5)
		for _, ss := range strings.Split(defaultCardSkill, ",") {
			if n, err := strconv.Atoi(ss); err == nil && ss != "" {
				robotActionDiff.DefaultCardSkill = append(robotActionDiff.DefaultCardSkill, n)
			}
		}

		robotActionsDiff = append(robotActionsDiff, robotActionDiff)
	}

	return &robotActionsDiff, nil
}
