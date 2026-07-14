package activity

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type ActivityDiff struct {
	Id                 int
	EActivityId        string // 枚举值，如 Activity_Test
	ActivityType       string // 如 ActTypeSkinRaffle
	ActivityPrefabType string
	BelongId           int
	Name               string
	ShowTab            bool
	Weight             int
	TimeType           int
	StartTime          string
	EndTime            string
	RewardTime         string
	RewardEndTime      string
	CustomParma        []int
	CustomParma2       string
}

func (a ActivityDiff) GetType() string {
	return "ActivityDiff"
}

func (a ActivityDiff) GetDisplayName() string {
	return a.Name
}

func GetActivityDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]ActivityDiff, err error) {
	var activityCols [][]string // 活动表
	if sheet, exist := sheetMap["活动表|Activity"]; exist {
		var err error
		activityCols, err = sheet.GetCols("活动表|Activity")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建活动Map
	activityDiffs := make([]ActivityDiff, 0, 100)

	for i, idStr := range activityCols[Id][startRow:helpers.AutoDetectEndIndex(activityCols, Id, startRow, 3)] {
		// 表格第一列是否有id作为读取这一列的标准
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			activityDiff := ActivityDiff{}

			eActivityId := utils.GetCellValue(activityCols, EActivityId, startRow+i)
			activityType := utils.GetCellValue(activityCols, ActivityType, startRow+i)
			activityPrefabType := utils.GetCellValue(activityCols, ActivityPrefabType, startRow+i)
			belongId := utils.GetCellValue(activityCols, BelongId, startRow+i)
			name := utils.GetCellValue(activityCols, Name, startRow+i)
			showTab := utils.GetCellValue(activityCols, ShowTab, startRow+i)
			weight := utils.GetCellValue(activityCols, Weight, startRow+i)
			timeType := utils.GetCellValue(activityCols, TimeType, startRow+i)
			startTime := utils.GetCellValue(activityCols, StartTime, startRow+i)
			endTime := utils.GetCellValue(activityCols, EndTime, startRow+i)
			rewardTime := utils.GetCellValue(activityCols, RewardTime, startRow+i)
			rewardEndTime := utils.GetCellValue(activityCols, RewardEndTime, startRow+i)
			customParma := utils.GetCellValue(activityCols, CustomParma, startRow+i)
			customParma2 := utils.GetCellValue(activityCols, CustomParma2, startRow+i)

			// 录入活动信息
			activityDiff.Id = id
			activityDiff.EActivityId = eActivityId
			activityDiff.ActivityType = activityType
			activityDiff.ActivityPrefabType = activityPrefabType

			if n, err := strconv.Atoi(belongId); err == nil {
				activityDiff.BelongId = n
			} else {
				activityDiff.BelongId = -1
			}

			activityDiff.Name = name

			if b, err := strconv.ParseBool(showTab); err == nil {
				activityDiff.ShowTab = b
			} else {
				activityDiff.ShowTab = false
			}

			if n, err := strconv.Atoi(weight); err == nil {
				activityDiff.Weight = n
			} else {
				activityDiff.Weight = -1
			}

			if n, err := strconv.Atoi(timeType); err == nil {
				activityDiff.TimeType = n
			} else {
				activityDiff.TimeType = -1
			}

			activityDiff.StartTime = startTime
			activityDiff.EndTime = endTime
			activityDiff.RewardTime = rewardTime
			activityDiff.RewardEndTime = rewardEndTime

			activityDiff.CustomParma = make([]int, 0, 10)
			for _, ss := range strings.Split(customParma, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					activityDiff.CustomParma = append(activityDiff.CustomParma, n)
				}
			}

			activityDiff.CustomParma2 = customParma2

			activityDiffs = append(activityDiffs, activityDiff)
		}
	}

	return &activityDiffs, nil
}
