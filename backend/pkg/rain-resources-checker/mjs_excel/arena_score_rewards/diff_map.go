package arena_score_rewards

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type ArenaScoreRewardDiff struct {
	Id      int
	Season  int
	Dan     int
	DanName string
	Reward  []ItemCfg // 注意：这里需要根据ItemCfg的实际类型定义
}

// ItemCfg 类型定义 - 根据实际需求可能需要调整
type ItemCfg struct {
	// 根据实际数据结构定义字段
}

func (a ArenaScoreRewardDiff) GetType() string {
	return "ArenaScoreRewardDiff"
}

func (a ArenaScoreRewardDiff) GetDisplayName() string {
	return strconv.Itoa(a.Id)
}

func GetArenaScoreRewardDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]ArenaScoreRewardDiff, err error) {
	var rewardCols [][]string
	if sheet, exist := sheetMap["竞技场积分奖励表|ArenaScoreReward"]; exist {
		var err error
		rewardCols, err = sheet.GetCols("竞技场积分奖励表|ArenaScoreReward")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	rewardsDiff := make([]ArenaScoreRewardDiff, 0, 100)

	for i, idStr := range rewardCols[Id][startRow:helpers.AutoDetectEndIndex(rewardCols, Id, startRow, 3)] {
		// 第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			rewardDiff := ArenaScoreRewardDiff{}

			season := utils.GetCellValue(rewardCols, Season, startRow+i)
			dan := utils.GetCellValue(rewardCols, Dan, startRow+i)
			danName := utils.GetCellValue(rewardCols, DanName, startRow+i)
			reward := utils.GetCellValue(rewardCols, Reward, startRow+i)

			rewardDiff.Id = id

			if n, err := strconv.Atoi(season); err == nil {
				rewardDiff.Season = n
			} else {
				rewardDiff.Season = -1
			}

			if n, err := strconv.Atoi(dan); err == nil {
				rewardDiff.Dan = n
			} else {
				rewardDiff.Dan = -1
			}

			rewardDiff.DanName = danName

			// 处理ItemCfg[]类型的奖励
			// 这里需要根据ItemCfg的实际结构来解析
			// 目前先按字符串数组处理
			rewardDiff.Reward = make([]ItemCfg, 0)
			if reward != "" {
				// TODO: 根据ItemCfg的实际格式解析reward字符串
				// 例如：可能格式是 "itemId1,count1;itemId2,count2" 或 JSON 格式
			}

			rewardsDiff = append(rewardsDiff, rewardDiff)
		}
	}

	return &rewardsDiff, nil
}
