package limit_skin_times_reward

import (
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// LimitSkinTimesRewardDiff 限时皮肤次数奖励差异结构体
type LimitSkinTimesRewardDiff struct {
	Id        int
	ActId     int    // 数值ID（枚举时为-1）
	ActIdStr  string // 原始字符串值（枚举类型如 Activity_LimitTimeSkin）
	DrawTimes int
	Reward    []ItemCfg
}

// ItemCfg 物品配置结构体
type ItemCfg struct {
	ItemId int
	Count  int
}

// GetType 返回类型名称
func (d LimitSkinTimesRewardDiff) GetType() string {
	return "LimitSkinTimesRewardDiff"
}

// GetDisplayName 返回显示名称
func (d LimitSkinTimesRewardDiff) GetDisplayName() string {
	return strconv.Itoa(d.Id)
}

const (
	Id        = 0
	ActId     = 1
	DrawTimes = 2
	Reward    = 3
)

// GetLimitSkinTimesRewardDiffMap 从sheetMap中加载限时皮肤次数奖励配置
func GetLimitSkinTimesRewardDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]LimitSkinTimesRewardDiff, err error) {
	var rewardCols [][]string // 限时皮肤次数奖表
	if sheet, exist := sheetMap["限时皮肤次数奖|LimitSkinTimesReward"]; exist {
		var err error
		rewardCols, err = sheet.GetCols("限时皮肤次数奖|LimitSkinTimesReward")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建限时皮肤次数奖励Map
	rewardsDiff := make([]LimitSkinTimesRewardDiff, 0, 100)

	// 正则匹配 {itemId;count} 格式
	braceReg := regexp.MustCompile(`\{(\d+);(\d+)}`)

	for i, idStr := range rewardCols[Id][startRow:helpers.AutoDetectEndIndex(rewardCols, Id, startRow, 3)] {
		// 表格第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			rewardDiff := LimitSkinTimesRewardDiff{}

			actId := utils.GetCellValue(rewardCols, ActId, startRow+i)
			drawTimes := utils.GetCellValue(rewardCols, DrawTimes, startRow+i)
			reward := utils.GetCellValue(rewardCols, Reward, startRow+i)

			// 录入限时皮肤次数奖励信息
			rewardDiff.Id = id
			rewardDiff.ActIdStr = actId

			if n, err := strconv.Atoi(actId); err == nil {
				rewardDiff.ActId = n
			} else {
				rewardDiff.ActId = -1
			}

			if n, err := strconv.Atoi(drawTimes); err == nil {
				rewardDiff.DrawTimes = n
			} else {
				rewardDiff.DrawTimes = -1
			}

			// 解析Reward字段，支持 {itemId;count} 或 itemId:count 格式，多个用逗号分隔
			rewardDiff.Reward = make([]ItemCfg, 0, 5)
			if reward != "" {
				for _, itemStr := range strings.Split(reward, ",") {
					itemStr = strings.TrimSpace(itemStr)
					if itemStr == "" {
						continue
					}

					itemCfg := ItemCfg{}

					// 尝试匹配 {itemId;count} 格式
					matches := braceReg.FindStringSubmatch(itemStr)
					if len(matches) >= 3 {
						// {itemId;count} 格式
						if itemId, err := strconv.Atoi(matches[1]); err == nil {
							itemCfg.ItemId = itemId
						}
						if count, err := strconv.Atoi(matches[2]); err == nil {
							itemCfg.Count = count
						}
					} else {
						// 尝试 itemId:count 格式
						parts := strings.Split(itemStr, ":")
						if len(parts) >= 2 {
							if itemId, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
								itemCfg.ItemId = itemId
							}
							if count, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
								itemCfg.Count = count
							}
						} else {
							// 纯数字，只解析itemId
							if itemId, err := strconv.Atoi(itemStr); err == nil {
								itemCfg.ItemId = itemId
								itemCfg.Count = 1
							}
						}
					}

					if itemCfg.ItemId > 0 {
						rewardDiff.Reward = append(rewardDiff.Reward, itemCfg)
					}
				}
			}

			rewardsDiff = append(rewardsDiff, rewardDiff)
		}
	}

	return &rewardsDiff, nil
}
