package season_pass_reward

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// ItemCfg 道具配置（格式：{物品ID;数量}）
type ItemCfg struct {
	ItemId int
	Count  int
}

type SeasonPassRewardDiff struct {
	Id           int
	SeasonPassId int
	Level        int
	NormalReward []ItemCfg // ItemCfg[] 类型，格式 {物品ID;数量}
	HighReward   []ItemCfg // ItemCfg[] 类型，格式 {物品ID;数量}
}

func (s SeasonPassRewardDiff) GetType() string {
	return "SeasonPassRewardDiff"
}

func (s SeasonPassRewardDiff) GetDisplayName() string {
	return strconv.Itoa(s.Id)
}

// parseItemCfgList 解析道具配置字符串，格式为 {itemId;count},{itemId;count}
func parseItemCfgList(itemStr string) []ItemCfg {
	result := make([]ItemCfg, 0, 10)
	if itemStr == "" {
		return result
	}

	itemStr = strings.TrimSpace(itemStr)
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
		} else if len(parts) == 1 {
			itemId, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			if itemId != 0 {
				result = append(result, ItemCfg{
					ItemId: itemId,
					Count:  1,
				})
			}
		}
	}
	return result
}

func GetSeasonPassRewardDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SeasonPassRewardDiff, err error) {
	var rewardCols [][]string // 赛季战令奖励表
	if sheet, exist := sheetMap["赛季战令奖励表|SeasonPassReward"]; exist {
		var err error
		rewardCols, err = sheet.GetCols("赛季战令奖励表|SeasonPassReward")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建战令奖励Map
	rewardsDiff := make([]SeasonPassRewardDiff, 0, 100)

	// 第一列是int类型的Id
	for i, idStr := range rewardCols[Id][startRow:helpers.AutoDetectEndIndex(rewardCols, Id, startRow, 3)] {
		// 表格第一列是否有id作为读取这一列的标准
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			rewardDiff := SeasonPassRewardDiff{}

			seasonPassId := utils.GetCellValue(rewardCols, SeasonPassId, startRow+i)
			level := utils.GetCellValue(rewardCols, Level, startRow+i)
			normalReward := utils.GetCellValue(rewardCols, NormalReward, startRow+i)
			highReward := utils.GetCellValue(rewardCols, HighReward, startRow+i)

			// 录入战令奖励信息
			rewardDiff.Id = id

			if n, err := strconv.Atoi(seasonPassId); err == nil {
				rewardDiff.SeasonPassId = n
			} else {
				rewardDiff.SeasonPassId = -1
			}

			if n, err := strconv.Atoi(level); err == nil {
				rewardDiff.Level = n
			} else {
				rewardDiff.Level = -1
			}

			rewardDiff.NormalReward = parseItemCfgList(normalReward)
			rewardDiff.HighReward = parseItemCfgList(highReward)

			rewardsDiff = append(rewardsDiff, rewardDiff)
		}
	}

	return &rewardsDiff, nil
}
