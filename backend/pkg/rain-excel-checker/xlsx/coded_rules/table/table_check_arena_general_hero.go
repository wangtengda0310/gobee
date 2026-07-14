// Package table 提供表级别的校验规则
// 本包中的规则针对单个 Excel 表的特定业务逻辑进行校验

package coded_rules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ArenaGeneralHeroOpenCheckRule 大将军武将开放时间检查规则
//
// 原始需求：https://ztgame.feishu.cn/wiki/TOJlw8ucyiNkfMkuoyjcJi8XnLc
//
// 需求确认记录：
// - 2024-xx-xx: 确认校验规则：从 ArenaScoreRewards 表获取 DanName 包含"大将军"的奖励
// - 2024-xx-xx: 确认对比规则：武将 OpenDate 与 ArenaSeason.SeasonStartTime 必须是同一天
// - 2024-xx-xx: 确认检查范围：只检查当前进行中或未来的赛季
//
// 相关表结构：
// - ArenaScoreRewards: Season, Dan, DanName, Reward (格式: {物品ID;数量})
// - ArenaSeason: Id, SeasonStartTime, SeasonEndTime
// - Hero: Id, Name, IsOpen, OpenDate
//
// 物品ID规则：
// - 物品ID前两位为10表示武将道具
// - 武将ID = 物品ID % 100000
type ArenaGeneralHeroOpenCheckRule struct{}

// Meta 返回规则元数据
func (c *ArenaGeneralHeroOpenCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.ARENA_GENERAL_HERO_OPEN_CHECK,
		DisplayName:    "大将军武将开放时间检查",
		Description:    "检查大将军段位奖励中的武将开放时间是否与竞技场赛季开始时间一致",
		TargetSheets:   []string{"ArenaScoreReward"},    // 适用于 ArenaScoreReward 表（需访问 Hero, ArenaSeason）
		RequiredSheets: []string{"Hero", "ArenaSeason"}, // 执行检查需要加载的关联表
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.WARN_DAYS_BEFORE,
				Label:       "提前警告天数",
				Description: "在赛季开始前多少天开始警告（默认5天）",
				Type:        "number",
				Default:     "5",
				Required:    false,
			},
			{
				Key:         json_rule.OPEN_DATE_COL_NAME,
				Label:       "武将开放时间列名",
				Description: "Hero表中武将开放时间的列名（默认: OpenDate）",
				Type:        "string",
				Default:     "OpenDate",
				Required:    false,
			},
		},
	}
}

// Check 执行表级检查
// 校验规则：
// 1. 从 ArenaScoreRewards 表获取 DanName 包含"大将军"的奖励
// 2. 从 Hero 表获取该武将的 OpenDate
// 3. 对比 OpenDate 与 ArenaSeason.SeasonStartTime 是否同一天
// 4. 只检查当前进行中或未来的赛季
func (c *ArenaGeneralHeroOpenCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "大将军武将开放时间检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 解析参数
	warnDaysBefore := 5
	if val, ok := param.Params[string(json_rule.WARN_DAYS_BEFORE)]; ok && val != "" {
		if days, err := helpers.ParseIntParamWithError(val); err == nil {
			warnDaysBefore = days
		}
	}

	// 获取当前时间
	now := helpers.ResolveNow(param.Now) // 单元测试可通过 param.Now 注入固定时间
	warnThreshold := now.AddDate(0, 0, -warnDaysBefore)

	// 加载相关表数据
	arenaScoreRewardsCols, arenaSeasonCols, heroCols := c.loadRelatedSheets(param.SheetMap)

	if arenaScoreRewardsCols == nil {
		result.Ok = false
		result.Reason = "未找到 ArenaScoreReward 表数据"
		return result
	}

	if arenaSeasonCols == nil {
		result.Ok = false
		result.Reason = "未找到 ArenaSeason 表数据"
		return result
	}

	if heroCols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	// 查找所有大将军武将
	generalHeroes := findArenaGeneralHeroes(arenaScoreRewardsCols, excelio.MJS_FIXED_ROWS_NUM)

	// 检查每个大将军武将
	for _, genHero := range generalHeroes {
		// 获取赛季时间
		seasonStartTime, seasonEndTime := getArenaSeasonTime(genHero.SeasonId, arenaSeasonCols, excelio.MJS_FIXED_ROWS_NUM)

		// 查找武将信息
		hero := findHeroById(genHero.HeroId, heroCols, excelio.MJS_FIXED_ROWS_NUM)
		if hero == nil {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  genHero.RowIndex,
				Reason: fmt.Sprintf("大将军武将ID=%d 在Hero表中不存在", genHero.HeroId),
			})
			continue
		}

		// 业务规则：
		// 0. 大将军武将必须 IsOpen=true
		// 1. 赛季未结束的武将必须配置 OpenDate
		// 2. 赛季已结束时，OpenDate 可以为空
		// 3. OpenDate 不为空时，必须与赛季 StartTime 精确匹配（精确到秒）

		// 检查 IsOpen：大将军武将必须开放
		if !hero.IsOpen {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: genHero.RowIndex,
				Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) IsOpen=false，大将军武将必须设置为IsOpen=true（规则：大将军武将必须开放）| 段位=%d",
					hero.Name, genHero.HeroId, genHero.Dan),
			})
			continue
		}

		if hero.OpenDate.IsZero() {
			// OpenDate 为空时，检查赛季是否已结束
			if seasonEndTime.Before(warnThreshold) {
				// 赛季已结束（相对于 warnThreshold），OpenDate 可以为空，跳过检查
			} else {
				// 赛季未结束，OpenDate 不能为空，报错
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index: genHero.RowIndex,
					Reason: fmt.Sprintf("赛季未结束(结束于%s)，大将军武将【%s】(ID=%d, OpenDate未配置)必须配置OpenDate（规则：赛季未结束时大将军武将必须配置OpenDate）| 段位=%d",
						helpers.FormatDateTime(seasonEndTime), hero.Name, genHero.HeroId, genHero.Dan),
				})
			}
		} else {
			// OpenDate 不为空，检查是否与赛季开始时间一致（精确到秒）
			if !helpers.TimeEquals(hero.OpenDate, seasonStartTime) {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index: genHero.RowIndex,
					Reason: fmt.Sprintf("大将军武将【%s】OpenDate(%s) 与赛季StartTime(%s) 不匹配（规则：大将军武将开放时间必须与赛季开始时间一致）",
						hero.Name,
						helpers.FormatDateTime(hero.OpenDate),
						helpers.FormatDateTime(seasonStartTime)),
				})
			}
		}
	}

	// 设置结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个大将军武将开放时间配置问题", len(result.ErrCells))
	}

	return result
}

// loadRelatedSheets 加载相关的表数据
func (c *ArenaGeneralHeroOpenCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (arenaScoreRewardsCols, arenaSeasonCols, heroCols [][]string) {
	// ArenaScoreReward 表（注意：实际 sheet 名为单数 ArenaScoreReward）
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaScoreReward"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			arenaScoreRewardsCols = cols
		}
	}

	// ArenaSeason 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaSeason"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			arenaSeasonCols = cols
		}
	}

	// Hero 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			heroCols = cols
		}
	}

	return arenaScoreRewardsCols, arenaSeasonCols, heroCols
}

// ==================== 辅助类型和函数 ====================

// GeneralHero 大将军武将信息
type GeneralHero struct {
	HeroId          int
	HeroName        string
	SeasonId        int
	SeasonStartTime time.Time
	SeasonEndTime   time.Time
	Dan             int
	DanName         string
	RowIndex        int
}

// findArenaGeneralHeroes 查找所有大将军武将
// 规则：从ArenaScoreRewards表筛选DanName包含"大将军"的行
func findArenaGeneralHeroes(arenaScoreRewardsCols [][]string, startRowIdx int) []*GeneralHero {
	heroes := make([]*GeneralHero, 0)

	seasonColIdx := helpers.GetColIndexByName(arenaScoreRewardsCols, "Season")
	danColIdx := helpers.GetColIndexByName(arenaScoreRewardsCols, "Dan")
	danNameColIdx := helpers.GetColIndexByName(arenaScoreRewardsCols, "DanName")
	rewardColIdx := helpers.GetColIndexByName(arenaScoreRewardsCols, "Reward")

	if danNameColIdx < 0 || rewardColIdx < 0 {
		return heroes
	}

	// 用于去重
	foundHeroIds := make(map[int]bool)

	// 遍历ArenaScoreRewards表
	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(arenaScoreRewardsCols, startRowIdx); rowIdx++ {
		danName := helpers.GetColValue(arenaScoreRewardsCols, danNameColIdx, rowIdx)
		if !strings.Contains(danName, "大将军") {
			continue
		}

		reward := helpers.GetColValue(arenaScoreRewardsCols, rewardColIdx, rowIdx)
		if reward == "" {
			continue
		}

		// 解析Reward中的物品
		items := helpers.ParseItemCfg(reward)
		for _, item := range items {
			if helpers.IsHeroItem(item.ItemId) {
				heroId := helpers.ExtractHeroIdFromItemCfg(item.ItemId)
				if heroId > 0 && !foundHeroIds[heroId] {
					hero := &GeneralHero{
						HeroId:   heroId,
						DanName:  danName,
						RowIndex: rowIdx,
					}

					if seasonColIdx >= 0 {
						hero.SeasonId, _ = strconv.Atoi(helpers.GetColValue(arenaScoreRewardsCols, seasonColIdx, rowIdx))
					}
					if danColIdx >= 0 {
						hero.Dan, _ = strconv.Atoi(helpers.GetColValue(arenaScoreRewardsCols, danColIdx, rowIdx))
					}

					heroes = append(heroes, hero)
					foundHeroIds[heroId] = true
					break
				}
			}
		}
	}

	return heroes
}

// getArenaSeasonTime 获取竞技场赛季时间
func getArenaSeasonTime(seasonId int, arenaSeasonCols [][]string, startRowIdx int) (start, end time.Time) {
	idColIdx := helpers.GetColIndexByName(arenaSeasonCols, "Id")
	startTimeColIdx := helpers.GetColIndexByName(arenaSeasonCols, "SeasonStartTime")
	endTimeColIdx := helpers.GetColIndexByName(arenaSeasonCols, "SeasonEndTime")

	if idColIdx < 0 {
		return
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(arenaSeasonCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(arenaSeasonCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id == seasonId {
			if startTimeColIdx >= 0 {
				start = helpers.ParseDate(helpers.GetColValue(arenaSeasonCols, startTimeColIdx, rowIdx))
			}
			if endTimeColIdx >= 0 {
				end = helpers.ParseDate(helpers.GetColValue(arenaSeasonCols, endTimeColIdx, rowIdx))
			}
			return
		}
	}
	return
}
