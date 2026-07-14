// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package hero

import (
	"fmt"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ItemSynthesisCheckRule 武将合成检查规则
//
// ## 校验规则
// 服务端武将合成判断走的是 DropItemsTemplate.ValidDate/ExpireDate（GetHeroCanComposeByDropConfig），
// 不依赖 Item.IsSynthetic 字段。本规则检查 DropItem.ValidDate 是否符合保护期要求。
//
// 1. 战令武将合成：保护期截止时间 = SeasonPass.StartTime + 4个月，DropItem.ValidDate 必须 >= 保护期截止时间
// 2. 大将军武将合成：保护期截止时间 = ArenaSeason.SeasonEndTime（赛季结束即保护期结束），DropItem.ValidDate 必须 >= 保护期截止时间
// 3. 反向检查：战令/赛季未开始时，不应提前配置 ValidDate（保护期内不应可合成）
//
// ## 保护期规则
// - 战令武将：SeasonPass.StartTime + protectMonths（默认4个月）
// - 大将军武将：ArenaSeason.SeasonEndTime（赛季时长约1个月，赛季结束即保护期结束）
//
// ## 数据流（6张表的关系）
//
// 战令路径：
//
//	SeasonPassReward.HighReward → 提取武将道具ID → Hero.Id（获取名称）
//	SeasonPassReward.SeasonPassId → SeasonPass.Id → StartTime（保护期起算点）
//	DropItem.Item → 匹配武将道具ID → DropItem.ValidDate（合成开放时间）
//
// 大将军路径：
//
//	ArenaScoreReward.Reward(DanName含"大将军") → 提取武将道具ID → Hero.Id
//	ArenaScoreReward.Season → ArenaSeason.Id → SeasonEndTime（保护期截止时间）
//	DropItem.Item → 匹配武将道具ID → DropItem.ValidDate（合成开放时间）
//
// ## 关键函数位置
//   - CalcSeasonPassProtectionDeadline → helpers/hero_rule_helper.go
//   - CalcArenaProtectionDeadline     → helpers/hero_rule_helper.go
//   - getHeroDropValidDate            → hero/hero_drop.go（同包内）
//   - FindSeasonPassHeroes            → helpers/hero_rule_helper.go
//   - FindArenaGeneralHeroes          → helpers/hero_rule_helper.go
type ItemSynthesisCheckRule struct{}

// Meta 返回规则元数据
func (c *ItemSynthesisCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.HERO_SYNTHESIS_CHECK,
		DisplayName:  "武将合成检查",
		Description:  "检查武将合成配置是否符合保护期规则。服务端合成判断走 DropItem.ValidDate/ExpireDate，本规则检查 ValidDate 不早于保护期截止时间。",
		TargetSheets: []string{"Item"}, // 适用于 Item 表（历史原因保留挂载点）
		RequiredSheets: []string{
			"Hero", "SeasonPassReward", "SeasonPass", "ArenaScoreReward", "ArenaSeason",
			"DropItem", // 合成判断实际依赖 DropItem.ValidDate
		},
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.WARN_DAYS_BEFORE,
				Label:       "提前警告天数",
				Description: "在保护期截止前多少天开始警告（默认7天）",
				Type:        "number",
				Default:     "7",
				Required:    false,
			},
			{
				Key:         json_rule.PROTECT_MONTHS,
				Label:       "保护期月数",
				Description: "保护期从活动开始时间算起的月数（默认4个月）",
				Type:        "number",
				Default:     "4",
				Required:    false,
			},
		},
	}
}

// Check 执行武将合成检查
func (c *ItemSynthesisCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "武将合成检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 解析参数
	warnDaysBefore := 7
	if val, ok := param.Params[string(json_rule.WARN_DAYS_BEFORE)]; ok && val != "" {
		if days, err := helpers.ParseIntParamWithError(val); err == nil {
			warnDaysBefore = days
		}
	}

	protectMonths := helpers.DefaultProtectMonths
	if val, ok := param.Params[string(json_rule.PROTECT_MONTHS)]; ok && val != "" {
		if months, err := helpers.ParseIntParamWithError(val); err == nil {
			protectMonths = months
		}
	}

	// 加载相关表数据
	heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols := c.loadRelatedSheets(param.SheetMap)

	if heroCols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	now := helpers.ResolveNow(param.Now)
	warnDuration := time.Duration(warnDaysBefore) * 24 * time.Hour

	// 1. 检查战令武将合成
	c.checkSeasonPassHeroes(heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, param.StartRowIdx, now, warnDuration, protectMonths, result)

	// 2. 检查大将军武将合成
	c.checkGeneralHeroes(heroCols, dropItemCols, arenaScoreRewardsCols, arenaSeasonCols, param.StartRowIdx, now, warnDuration, protectMonths, result)

	// 设置结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个武将合成配置问题", len(result.ErrCells))
	}

	return result
}

// checkSeasonPassHeroes 检查战令武将合成
// 规则：战令开始后N个月需配置合成（DropItem.ValidDate >= 保护期截止时间）
//
// 保护期截止时间 = StartTime + protectMonths
// - ValidDate 为空 → 报错（未配置合成时间）
// - ValidDate <= 保护期截止时间 → 报错（保护期内不应可合成）
// - ValidDate > 保护期截止时间 → 正常
//
// 反向检查：战令未开始时 ValidDate 不应早于战令开始时间
func (c *ItemSynthesisCheckRule) checkSeasonPassHeroes(heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols [][]string, startRowIdx int, now time.Time, warnDuration time.Duration, protectMonths int, result *json_rule.TableCheckResult) {
	if seasonPassRewardCols == nil || seasonPassCols == nil {
		return
	}

	seasonPassHeroes := helpers.FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, startRowIdx)

	for _, spHero := range seasonPassHeroes {
		if spHero.StartTime.IsZero() {
			continue
		}

		// 计算保护期截止时间：战令开始时间 + protectMonths
		protectionDeadline := helpers.CalcSeasonPassProtectionDeadline(spHero.StartTime, protectMonths)

		// 获取武将名称和行信息
		heroRow := helpers.FindHeroById(spHero.HeroId, heroCols, startRowIdx)
		heroName := spHero.HeroName
		if heroName == "" && heroRow != nil {
			heroName = heroRow.Name
		}
		sourceInfo := fmt.Sprintf("来源:SeasonPassReward#%d", spHero.RowIndex+1)

		// 获取 DropItem 中该武将的 ValidDate
		validDate := getHeroDropValidDate(spHero.HeroId, dropItemCols, startRowIdx)

		if validDate.IsZero() {
			// 未配置掉落/合成时间
			// 只在保护期截止已过或临近时才报错
			if protectionDeadline.Before(now) || protectionDeadline.Equal(now) {
				errIdx := spHero.RowIndex
				if heroRow != nil {
					errIdx = heroRow.RowIndex
				}
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    errIdx,
					ExcelRow: errIdx + 1,
					Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 战令已于 %s 开始(SeasonPass.StartTime)，保护期(StartTime+%d月)已过仍未配置合成时间（DropItem.ValidDate为空）（%s）",
						heroName, spHero.HeroId, helpers.FormatDate(spHero.StartTime), protectMonths, sourceInfo),
				})
			} else {
				// 临近截止时间，提前提醒
				timeUntilDeadline := protectionDeadline.Sub(now)
				if timeUntilDeadline <= warnDuration {
					errIdx := spHero.RowIndex
					if heroRow != nil {
						errIdx = heroRow.RowIndex
					}
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    errIdx,
						ExcelRow: errIdx + 1,
						Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 需在 %s 前配置合成时间(SeasonPass.StartTime+%d月截止，剩余 %.0f 天)（%s）",
							heroName, spHero.HeroId, helpers.FormatDate(protectionDeadline), protectMonths, timeUntilDeadline.Hours()/24, sourceInfo),
					})
				}
			}
			continue
		}

		// ValidDate 已配置，检查是否在保护期之后
		// ValidDate == 保护期截止时间视为合法（保护期当天生效）
		if validDate.Before(protectionDeadline) {
			// ValidDate < 保护期截止时间，保护期内不应可合成
			errIdx := spHero.RowIndex
			if heroRow != nil {
				errIdx = heroRow.RowIndex
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    errIdx,
				ExcelRow: errIdx + 1,
				Reason: fmt.Sprintf("战令武将【%s】(ID=%d) DropItem.ValidDate(%s) 早于保护期截止时间(%s, SeasonPass.StartTime+%d月)（规则：战令武将保护期内不应可合成）（%s）",
					heroName, spHero.HeroId,
					helpers.FormatDateTime(validDate), helpers.FormatDateTime(protectionDeadline), protectMonths, sourceInfo),
			})
		}

		// 反向检查：战令未开始时 ValidDate 不应早于战令开始时间
		if !spHero.StartTime.IsZero() && spHero.StartTime.After(now) && validDate.Before(spHero.StartTime) {
			errIdx := spHero.RowIndex
			if heroRow != nil {
				errIdx = heroRow.RowIndex
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    errIdx,
				ExcelRow: errIdx + 1,
				Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 战令尚未开始（%s 开始），DropItem.ValidDate(%s) 早于战令开始时间(SeasonPass.StartTime)（规则：战令未开始不应提前配置合成）（%s）",
					heroName, spHero.HeroId, helpers.FormatDate(spHero.StartTime),
					helpers.FormatDateTime(validDate), sourceInfo),
			})
		}
	}
}

// checkGeneralHeroes 检查大将军武将合成
// 规则：保护期截止后需配置合成（DropItem.ValidDate >= 保护期截止时间）
//
// 保护期截止时间 = ArenaSeason.SeasonEndTime（赛季结束即保护期结束）
// - ValidDate 为空 → 报错（未配置合成时间）
// - ValidDate <= 保护期截止时间 → 报错（保护期内不应可合成）
// - ValidDate > 保护期截止时间 → 正常
//
// 反向检查：赛季未开始时 ValidDate 不应早于赛季开始时间
func (c *ItemSynthesisCheckRule) checkGeneralHeroes(heroCols, dropItemCols, arenaScoreRewardsCols, arenaSeasonCols [][]string, startRowIdx int, now time.Time, warnDuration time.Duration, protectMonths int, result *json_rule.TableCheckResult) {
	if arenaScoreRewardsCols == nil || arenaSeasonCols == nil {
		return
	}

	generalHeroes := helpers.FindArenaGeneralHeroes(arenaScoreRewardsCols, startRowIdx)

	for _, genHero := range generalHeroes {
		seasonStartTime, seasonEndTime := helpers.GetArenaSeasonTime(genHero.SeasonId, arenaSeasonCols, startRowIdx)

		if seasonEndTime.IsZero() {
			continue
		}

		// 计算保护期截止时间：赛季结束时间即保护期截止
		arenaProtectionDeadline := helpers.CalcArenaProtectionDeadline(seasonEndTime)

		// 获取武将名称和行信息
		heroRow := helpers.FindHeroById(genHero.HeroId, heroCols, startRowIdx)
		heroName := genHero.HeroName
		if heroName == "" && heroRow != nil {
			heroName = heroRow.Name
		}
		sourceInfo := fmt.Sprintf("来源:ArenaScoreReward#%d", genHero.RowIndex+1)

		// 获取 DropItem 中该武将的 ValidDate
		validDate := getHeroDropValidDate(genHero.HeroId, dropItemCols, startRowIdx)

		if validDate.IsZero() {
			// 未配置掉落/合成时间
			if arenaProtectionDeadline.Before(now) {
				// 保护期已过，仍未配置
				errIdx := genHero.RowIndex
				if heroRow != nil {
					errIdx = heroRow.RowIndex
				}
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    errIdx,
					ExcelRow: errIdx + 1,
					Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 保护期已于 %s 截止(ArenaSeason.SeasonEndTime)，仍未配置合成时间（DropItem.ValidDate为空）（%s）",
						heroName, genHero.HeroId, helpers.FormatDate(arenaProtectionDeadline), sourceInfo),
				})
			} else {
				// 保护期未结束，临近截止时提前提醒
				timeUntilEnd := arenaProtectionDeadline.Sub(now)
				if timeUntilEnd <= warnDuration {
					errIdx := genHero.RowIndex
					if heroRow != nil {
						errIdx = heroRow.RowIndex
					}
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    errIdx,
						ExcelRow: errIdx + 1,
						Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 保护期将于 %s 截止(ArenaSeason.SeasonEndTime)，请提前配置合成时间（剩余 %.0f 天）（%s）",
							heroName, genHero.HeroId, helpers.FormatDate(arenaProtectionDeadline), timeUntilEnd.Hours()/24, sourceInfo),
					})
				}
			}
			continue
		}

		// ValidDate 已配置，检查是否在保护期截止时间之后
		// ValidDate == 保护期截止时间视为合法（保护期当天生效）
		if validDate.Before(arenaProtectionDeadline) {
			// ValidDate < 保护期截止时间，保护期内不应可合成
			errIdx := genHero.RowIndex
			if heroRow != nil {
				errIdx = heroRow.RowIndex
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    errIdx,
				ExcelRow: errIdx + 1,
				Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) DropItem.ValidDate(%s) 早于保护期截止时间(%s, ArenaSeason.SeasonEndTime)（规则：大将军武将保护期内不应可合成）（%s）",
					heroName, genHero.HeroId,
					helpers.FormatDateTime(validDate), helpers.FormatDateTime(arenaProtectionDeadline), sourceInfo),
			})
		}

		// 反向检查：赛季未开始时 ValidDate 不应早于赛季开始时间
		if !seasonStartTime.IsZero() && seasonStartTime.After(now) && validDate.Before(seasonStartTime) {
			errIdx := genHero.RowIndex
			if heroRow != nil {
				errIdx = heroRow.RowIndex
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    errIdx,
				ExcelRow: errIdx + 1,
				Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 赛季尚未开始（%s 开始），DropItem.ValidDate(%s) 早于赛季开始时间(ArenaSeason.SeasonStartTime)（规则：赛季未开始不应提前配置合成）（%s）",
					heroName, genHero.HeroId, helpers.FormatDate(seasonStartTime),
					helpers.FormatDateTime(validDate), sourceInfo),
			})
		}
	}
}

// loadRelatedSheets 加载相关的表数据
func (c *ItemSynthesisCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols [][]string) {
	// Hero 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			heroCols = cols
		}
	}

	// DropItem 表（合成判断实际依赖 ValidDate）
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "DropItem"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			dropItemCols = cols
		}
	}

	// SeasonPassReward 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPassReward"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			seasonPassRewardCols = cols
		}
	}

	// SeasonPass 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPass"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			seasonPassCols = cols
		}
	}

	// ArenaScoreReward 表（注意：实际 sheet 名为单数 ArenaScoreReward，不是复数 ArenaScoreRewards）
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

	return
}
