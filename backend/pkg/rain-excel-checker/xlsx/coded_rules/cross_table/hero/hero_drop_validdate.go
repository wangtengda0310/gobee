// 普通武将掉落时间检查规则
// 检查已开放的普通武将（非战令、非大将军）在 DropItem 表中的 ValidDate 是否 <= 下一个周四 5:00
package hero

import (
	"fmt"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// HeroDropValidDateCheckRule 普通武将掉落时间检查规则
//
// 规则描述：已开放的普通武将（非战令、非大将军）在 DropItem 表中的 ValidDate 必须 <= 下一个周四 5:00
//
// 检查逻辑：
//  1. 构建战令武将ID集合和大将军武将ID集合
//  2. 遍历 DropItem 表，找到包含武将道具的行
//  3. 提取 HeroId，排除战令和大将军武将
//  4. 检查普通武将是否已开放（IsOpen=true 且 OpenDate 已过或为空）
//  5. 对已开放的普通武将，检查 ValidDate <= 下一个周四 5:00
//
// 跳过条件：
//   - 战令/大将军武将
//   - 未开放武将
//   - ValidDate 为空
type HeroDropValidDateCheckRule struct{}

// Meta 返回规则元数据
func (c *HeroDropValidDateCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.HERO_DROP_VALIDDATE_CHECK,
		DisplayName:    "普通武将掉落时间检查",
		Description:    "检查已开放的普通武将在掉落表中的ValidDate是否早于等于下一个周四5:00",
		TargetSheets:   []string{"DropItem"},
		RequiredSheets: []string{"Hero", "SeasonPassReward", "SeasonPass", "ArenaScoreReward"},
	}
}

// Check 执行检查
func (c *HeroDropValidDateCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "普通武将掉落时间检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	now := helpers.ResolveNow(param.Now)
	nextThursday := getNextThursday5AM(now)

	// 加载关联表
	dropItemCols, heroCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardCols := c.loadRelatedSheets(param.SheetMap)

	if dropItemCols == nil {
		result.Ok = false
		result.Reason = "未找到 DropItem 表数据"
		return result
	}
	if heroCols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	// 构建战令+大将军武将ID集合（用于排除）
	specialHeroIds := c.buildSpecialHeroIdSet(seasonPassRewardCols, seasonPassCols, arenaScoreRewardCols)

	// 遍历 DropItem 表
	itemColIdx := helpers.GetColIndexByName(dropItemCols, "Item")
	validDateColIdx := helpers.GetColIndexByName(dropItemCols, "ValidDate")
	nameColIdx := helpers.GetColIndexByName(dropItemCols, "Name")

	if itemColIdx < 0 {
		return result
	}

	for rowIdx := excelio.MJS_FIXED_ROWS_NUM; rowIdx < helpers.GetDataEndIndex(dropItemCols, excelio.MJS_FIXED_ROWS_NUM); rowIdx++ {
		itemStr := helpers.GetColValue(dropItemCols, itemColIdx, rowIdx)
		if itemStr == "" {
			continue
		}

		// 解析 Item 列中的物品
		items := helpers.ParseItemCfg(itemStr)
		for _, item := range items {
			// 只处理武将道具
			if !helpers.IsHeroItem(item.ItemId) {
				continue
			}

			heroId := helpers.ExtractHeroIdFromItemCfg(item.ItemId)
			if heroId <= 0 {
				continue
			}

			// 跳过战令/大将军武将
			if specialHeroIds[heroId] {
				continue
			}

			// 从 Hero 表获取武将信息
			hero := helpers.FindHeroById(heroId, heroCols, excelio.MJS_FIXED_ROWS_NUM)
			if hero == nil {
				continue
			}

			// 跳过未开放武将
			if !hero.IsOpen {
				continue
			}
			// 跳过 OpenDate 在未来的武将（尚未真正开放）
			if !hero.OpenDate.IsZero() && hero.OpenDate.After(now) {
				continue
			}

			// 读取 ValidDate
			if validDateColIdx < 0 {
				continue
			}
			validDateStr := helpers.GetColValue(dropItemCols, validDateColIdx, rowIdx)
			if validDateStr == "" {
				// ValidDate 为空，由 HERO_DROP_CHECK 负责检查是否应该在掉落库中
				continue
			}
			validDate := helpers.ParseDate(validDateStr)
			if validDate.IsZero() {
				continue
			}

			// 核心检查：ValidDate 必须 <= 下一个周四 5:00
			if validDate.After(nextThursday) {
				dropItemName := ""
				if nameColIdx >= 0 {
					dropItemName = helpers.GetColValue(dropItemCols, nameColIdx, rowIdx)
				}
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason: fmt.Sprintf("已开放普通武将【%s】(ID=%d)的掉落ValidDate(%s)超过下一个周四5:00(%s)（规则：已开放普通武将掉落时间必须<=下一个周四5:00）| 掉落项=%s",
						hero.Name, heroId,
						helpers.FormatDateTime(validDate), helpers.FormatDateTime(nextThursday),
						dropItemName),
				})
			}
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个普通武将掉落时间配置问题", len(result.ErrCells))
	}

	return result
}

// loadRelatedSheets 加载关联表数据
func (c *HeroDropValidDateCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (
	dropItemCols, heroCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardCols [][]string,
) {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "DropItem"); ok {
		if cols, err := file.GetCols(sheetName); err == nil {
			dropItemCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		if cols, err := file.GetCols(sheetName); err == nil {
			heroCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPassReward"); ok {
		if cols, err := file.GetCols(sheetName); err == nil {
			seasonPassRewardCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPass"); ok {
		if cols, err := file.GetCols(sheetName); err == nil {
			seasonPassCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaScoreReward"); ok {
		if cols, err := file.GetCols(sheetName); err == nil {
			arenaScoreRewardCols = cols
		}
	}
	return
}

// buildSpecialHeroIdSet 构建战令+大将军武将ID集合
func (c *HeroDropValidDateCheckRule) buildSpecialHeroIdSet(
	seasonPassRewardCols, seasonPassCols, arenaScoreRewardCols [][]string,
) map[int]bool {
	result := make(map[int]bool)

	// 战令武将
	if seasonPassRewardCols != nil {
		spHeroes := helpers.FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, excelio.MJS_FIXED_ROWS_NUM)
		for _, h := range spHeroes {
			result[h.HeroId] = true
		}
	}

	// 大将军武将
	if arenaScoreRewardCols != nil {
		genHeroes := helpers.FindArenaGeneralHeroes(arenaScoreRewardCols, excelio.MJS_FIXED_ROWS_NUM)
		for _, h := range genHeroes {
			result[h.HeroId] = true
		}
	}

	return result
}

// getNextThursday5AM 计算下一个周四 5:00
// 规则：如果当前时间 <= 本周四 5:00 则返回本周四 5:00，否则返回下周四 5:00
//
// 注意：返回的时间使用 UTC 时区，与 helpers.ParseDate 保持一致。
// ParseDate 使用 time.Parse 解析无时区的日期字符串，返回 UTC 时区的时间。
// 如果此处使用本地时区，会导致跨时区比较时产生偏差（如 CST 环境下相差 8 小时）。
func getNextThursday5AM(now time.Time) time.Time {
	// 获取本周四（基于本地时间的年月日，但时区统一为 UTC）
	weekday := now.Weekday()
	// time.Sunday = 0, time.Monday = 1, ..., time.Thursday = 4, ..., time.Saturday = 6
	daysToThursday := (4 - int(weekday) + 7) % 7 // 本周距离周四的天数
	// 使用 UTC 时区，与 ParseDate 解析的日期保持同一时区基准
	thisThursday := time.Date(now.Year(), now.Month(), now.Day(), 5, 0, 0, 0, time.UTC).AddDate(0, 0, daysToThursday)

	// 比较时也使用相同方式构造 now（本地时间的年月日时分秒 + UTC 时区）
	nowForCompare := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
	// 如果本周四 5:00 还没过，返回本周四
	if !nowForCompare.After(thisThursday) {
		return thisThursday
	}

	// 否则返回下周四
	return thisThursday.AddDate(0, 0, 7)
}
