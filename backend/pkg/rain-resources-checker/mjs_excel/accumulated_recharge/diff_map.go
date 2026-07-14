// activity-wiki-dev: 活动Wiki开发技能生成
// 功能: 累充奖励表解析逻辑，将Excel行解析为AccumulatedRechargeDiff结构体
// 关联活动类型: ActTypeAccumulatedRecharge
// 生成时间: 2026-05-05

package accumulated_recharge

import (
	"regexp"
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// ItemCfg 道具配置，从花括号格式 {itemId;count} 解析
type ItemCfg struct {
	ItemId int
	Count  int
}

// AccumulatedRechargeDiff 累充奖励表单行数据
type AccumulatedRechargeDiff struct {
	Id          int
	ActId       string    // EActivityId 枚举值，如 "Activity_AccumulatedRecharge1"
	RechargeNum int       // 累充门槛金额
	Reward      string    // 奖励原始字符串
	RewardItems []ItemCfg // 解析后的奖励道具列表
}

func (a AccumulatedRechargeDiff) GetType() string {
	return "AccumulatedRechargeDiff"
}

func (a AccumulatedRechargeDiff) GetDisplayName() string {
	return strconv.Itoa(a.Id)
}

// rewardRegex 匹配花括号格式的道具配置 {itemId;count}
var rewardRegex = regexp.MustCompile(`\{(\d+);(\d+)\}`)

// parseReward 解析奖励字符串，如 "{2000001;10}" 或 "{1030005;1},{1040006;1}"
func parseReward(s string) []ItemCfg {
	if s == "" {
		return nil
	}
	matches := rewardRegex.FindAllStringSubmatch(s, -1)
	items := make([]ItemCfg, 0, len(matches))
	for _, m := range matches {
		itemId, _ := strconv.Atoi(m[1])
		count, _ := strconv.Atoi(m[2])
		items = append(items, ItemCfg{ItemId: itemId, Count: count})
	}
	return items
}

// GetAccumulatedRechargeDiffMap 解析累充奖励表Excel
func GetAccumulatedRechargeDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]AccumulatedRechargeDiff, err error) {
	var cols [][]string
	if sheet, exist := sheetMap["累充奖励表|AccumulatedRechargeReward"]; exist {
		cols, err = sheet.GetCols("累充奖励表|AccumulatedRechargeReward")
		if err != nil {
			return nil, err
		}
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM
	diffs := make([]AccumulatedRechargeDiff, 0, 50)

	for i, idStr := range cols[Id][startRow:helpers.AutoDetectEndIndex(cols, Id, startRow, 3)] {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			diff := AccumulatedRechargeDiff{}
			diff.Id = id
			diff.ActId = utils.GetCellValue(cols, ActId, startRow+i)

			rechargeNumStr := utils.GetCellValue(cols, RechargeNum, startRow+i)
			if n, err := strconv.Atoi(rechargeNumStr); err == nil {
				diff.RechargeNum = n
			}

			rewardStr := utils.GetCellValue(cols, Reward, startRow+i)
			diff.Reward = rewardStr
			diff.RewardItems = parseReward(rewardStr)

			diffs = append(diffs, diff)
		}
	}
	return &diffs, nil
}
