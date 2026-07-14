// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package hero

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// HeroMeltCheckRule 武将熔炼检查规则
//
// 原始需求：https://ztgame.feishu.cn/wiki/TOJlw8ucyiNkfMkuoyjcJi8XnLc
//
// ## 校验规则（来源：需求确认）
// 1. 开放武将必须可熔炼：IsOpen=true 且 OpenDate已过 的武将，CanMelt必须为true
// 2. 技能关联方式：Hero.Skill字段直接关联SkillMelt.Id
// 3. 所有技能必须配置熔炼：武将的每个技能都必须在SkillMelt表中有对应记录
// 4. 技能熔炼配置一致性：如果武将CanMelt=true，则其所有技能在SkillMelt表中的CanMelt也必须为true
// 5. 战令武将熔炼：战令结束后3个月开启熔炼，提前5天提醒
// 6. 大将军武将熔炼：赛季结束后开启熔炼，提前5天提醒
//
// ## 需求确认记录
//   - Q: 武将熔炼检查需要检查哪些字段/表？
//     A: 两者都需要检查（Hero.CanMelt + SkillMelt.CanMelt）
//     额外规则：开放的武将一定是可熔炼的，武将的所有技能都需要检查，不能在skillmelt表中漏配
//   - Q: 武将技能如何与SkillMelt表关联？
//     A: Hero.Skill字段直接关联SkillMelt.Id
//
// ## 相关表结构
// - Hero: Id, Name, CanMelt, MeltName, Skill, IsOpen, OpenDate
//   - Skill字段格式：[Skill1,Skill2] 或 Skill1,Skill2
//
// - SkillMelt: Id (ESkillId), MeltPower, CanMelt
type HeroMeltCheckRule struct{}

// Meta 返回规则元数据
func (c *HeroMeltCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.HERO_MELT_CHECK,
		DisplayName:    "武将熔炼检查",
		Description:    "检查武将熔炼配置是否完整。包括：1)开放武将必须可熔炼；2)所有技能必须在SkillMelt表配置；3)战令/大将军武将的熔炼时间检查。",
		TargetSheets:   []string{"Hero"}, // 适用于 Hero 表
		RequiredSheets: []string{"SkillMelt", "SeasonPassReward", "SeasonPass", "ArenaScoreReward", "ArenaSeason"},
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.WARN_DAYS_BEFORE,
				Label:       "提前警告天数",
				Description: "在需要开启熔炼前多少天开始警告（默认5天）",
				Type:        "number",
				Default:     "5",
				Required:    false,
			},
			{
				Key:         json_rule.PROTECT_MONTHS,
				Label:       "保护期月数",
				Description: "战令开始时间起多少个月内武将不应开启熔炼（默认4个月）",
				Type:        "number",
				Default:     "4",
				Required:    false,
			},
			{
				Key:         json_rule.CAN_MELT_COL_NAME,
				Label:       "可熔炼列名",
				Description: "Hero表中可熔炼的列名（默认: CanMelt）",
				Type:        "string",
				Default:     "CanMelt",
				Required:    false,
			},
			{
				Key:         json_rule.SKILL_COL_NAME,
				Label:       "技能列名",
				Description: "Hero表中技能的列名（默认: Skill）",
				Type:        "string",
				Default:     "Skill",
				Required:    false,
			},
		},
	}
}

// Check 执行表级检查
func (c *HeroMeltCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "武将熔炼检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 解析参数
	warnDaysBefore := 5
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
	heroCols, skillMeltCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols := c.loadRelatedSheets(param.SheetMap)

	if heroCols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	// 构建技能熔炼映射表
	skillMeltMap := helpers.BuildSkillMeltMap(skillMeltCols, param.StartRowIdx)

	now := helpers.ResolveNow(param.Now) // 单元测试可通过 param.Now 注入固定时间
	warnDuration := time.Duration(warnDaysBefore) * 24 * time.Hour

	// 1. 检查所有武将的熔炼完整性
	c.checkAllHeroesMelt(heroCols, skillMeltMap, param.StartRowIdx, now, result)

	// 2. 检查战令武将熔炼时间
	c.checkSeasonPassHeroesMelt(heroCols, skillMeltMap, seasonPassRewardCols, seasonPassCols, param.StartRowIdx, now, warnDuration, protectMonths, result)

	// 3. 检查大将军武将熔炼时间
	c.checkGeneralHeroesMelt(heroCols, skillMeltMap, arenaScoreRewardsCols, arenaSeasonCols, param.StartRowIdx, now, warnDuration, protectMonths, result)

	// 设置结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个武将熔炼配置问题", len(result.ErrCells))
	}

	return result
}

// checkAllHeroesMelt 检查所有武将的熔炼完整性
// 规则：
// 1. 开放武将必须可熔炼（仅检查 HeroType=1 的武将）
// 2. 所有技能必须在SkillMelt表配置
// 3. 技能熔炼配置一致性
// 4. MeltPower 值验证：CanMelt=true 时应为 16/24/32，CanMelt=false 时应为 100
func (c *HeroMeltCheckRule) checkAllHeroesMelt(heroCols [][]string, skillMeltMap map[string]*helpers.SkillMeltInfo, startRowIdx int, now time.Time, result *json_rule.TableCheckResult) {
	// 查找列索引
	idColIdx := helpers.GetColIndexByName(heroCols, "Id")
	nameColIdx := helpers.GetColIndexByName(heroCols, "Name")
	heroTypeColIdx := helpers.GetColIndexByName(heroCols, "HeroType")
	isOpenColIdx := helpers.GetColIndexByName(heroCols, "IsOpen")
	openDateColIdx := helpers.GetColIndexByName(heroCols, "OpenDate")
	canMeltColIdx := helpers.GetColIndexByName(heroCols, "CanMelt")
	skillColIdx := helpers.GetColIndexByName(heroCols, "Skill")

	if idColIdx < 0 {
		return
	}

	// 遍历所有武将
	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(heroCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(heroCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}

		heroName := ""
		if nameColIdx >= 0 {
			heroName = helpers.GetColValue(heroCols, nameColIdx, rowIdx)
		}

		// 检查武将类型
		heroType := 0
		if heroTypeColIdx >= 0 {
			heroType, _ = strconv.Atoi(helpers.GetColValue(heroCols, heroTypeColIdx, rowIdx))
		}

		// 检查是否开放
		isOpen := false
		if isOpenColIdx >= 0 {
			isOpen = parseBoolValue(helpers.GetColValue(heroCols, isOpenColIdx, rowIdx))
		}

		// 检查开放时间
		var openDate time.Time
		if openDateColIdx >= 0 {
			openDate = helpers.ParseDate(helpers.GetColValue(heroCols, openDateColIdx, rowIdx))
		}

		// 检查是否可熔炼
		canMelt := false
		if canMeltColIdx >= 0 {
			canMelt = parseBoolValue(helpers.GetColValue(heroCols, canMeltColIdx, rowIdx))
		}

		// 获取技能列表
		var skills []string
		if skillColIdx >= 0 {
			skills = parseSkillList(helpers.GetColValue(heroCols, skillColIdx, rowIdx))
		}

		// 规则1：开放武将必须可熔炼（仅检查 HeroType=1 的武将）
		// IsOpen=true 且 OpenDate已过 且 HeroType=1 的武将，CanMelt必须为true
		isHeroOpened := isOpen && !openDate.IsZero() && openDate.Before(now) && heroType == 1
		if isHeroOpened && !canMelt {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("武将【%s】已开放但CanMelt=false", heroName),
			})
		}

		// 规则2-4：技能熔炼配置检查（仅检查已开放的武将）
		// 未开放的武将（IsOpen=0 或 OpenDate未到）跳过技能配置验证
		if !isHeroOpened {
			continue // 跳过未开放武将的技能配置检查
		}

		// 规则2：所有技能必须在SkillMelt表中配置
		// 规则3：如果武将可熔炼，技能也必须可熔炼
		// 规则4：MeltPower 值验证
		for _, skillId := range skills {
			meltInfo, exists := skillMeltMap[skillId]
			if !exists {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("武将【%s】的技能 %s 在SkillMelt表中未配置", heroName, skillId),
				})
				continue
			}

			// 规则3：如果武将可熔炼，技能也必须可熔炼
			if canMelt && !meltInfo.CanMelt {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("武将【%s】可熔炼，但技能 %s 的SkillMelt.CanMelt=false", heroName, skillId),
				})
			}

			// 规则4：验证 MeltPower 值
			if meltInfo.CanMelt {
				// CanMelt=true 时，MeltPower 应为 16、24 或 32
				if meltInfo.MeltPower != 16 && meltInfo.MeltPower != 24 && meltInfo.MeltPower != 32 {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    rowIdx,
						ExcelRow: rowIdx + 1,
						Reason: fmt.Sprintf("武将【%s】的技能 %s CanMelt=true，但MeltPower=%d（应为16/24/32）",
							heroName, skillId, meltInfo.MeltPower),
					})
				}
			} else {
				// CanMelt=false 时，MeltPower 应为 100
				if meltInfo.MeltPower != 100 {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    rowIdx,
						ExcelRow: rowIdx + 1,
						Reason: fmt.Sprintf("武将【%s】的技能 %s CanMelt=false，但MeltPower=%d（应为100）",
							heroName, skillId, meltInfo.MeltPower),
					})
				}
			}
		}
	}
}

// checkSeasonPassHeroesMelt 检查战令武将熔炼时间
// 规则：战令结束后3个月需开启熔炼，提前5天提醒
// 注意：只检查 HeroType=1 的普通武将
func (c *HeroMeltCheckRule) checkSeasonPassHeroesMelt(heroCols [][]string, skillMeltMap map[string]*helpers.SkillMeltInfo, seasonPassRewardCols, seasonPassCols [][]string, startRowIdx int, now time.Time, warnDuration time.Duration, protectMonths int, result *json_rule.TableCheckResult) {
	if seasonPassRewardCols == nil || seasonPassCols == nil {
		return
	}

	// 查找 HeroType 列索引（只检查 HeroType=1 的普通武将）
	heroTypeColIdx := helpers.GetColIndexByName(heroCols, "HeroType")

	// 查找所有战令武将
	seasonPassHeroes := helpers.FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, startRowIdx)

	for _, spHero := range seasonPassHeroes {
		if spHero.EndTime.IsZero() {
			continue
		}

		// 获取武将信息
		hero := helpers.FindHeroById(spHero.HeroId, heroCols, startRowIdx)
		if hero == nil {
			continue
		}

		// 检查 HeroType 列（只检查 HeroType=1 的普通武将）
		if heroTypeColIdx >= 0 {
			heroType := 1 // 默认为普通武将
			if heroTypeStr := helpers.GetColValue(heroCols, heroTypeColIdx, hero.RowIndex); heroTypeStr != "" {
				heroType, _ = strconv.Atoi(heroTypeStr)
			}
			if heroType != 1 {
				continue // HeroType≠1 表示特殊武将，跳过熔炼检查
			}
		}

		heroName := spHero.HeroName
		if heroName == "" {
			heroName = hero.Name
		}

		// 计算需要开启熔炼的时间点：战令结束后N个月
		protectionDeadline := helpers.CalcSeasonPassProtectionDeadline(spHero.StartTime, protectMonths)

		// 战令已结束
		if spHero.EndTime.Before(now) {
			// 已过截止时间
			if protectionDeadline.Before(now) || protectionDeadline.Equal(now) {
				if !hero.CanMelt {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    hero.RowIndex,
						ExcelRow: hero.RowIndex + 1,
						Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 保护期已过（战令开始时间起%d个月，战令结束时间=%s），仍未开启熔炼",
							heroName, spHero.HeroId, protectMonths, helpers.FormatDate(spHero.EndTime)),
					})
				}
				// 检查技能熔炼配置
				for _, skillId := range hero.Skills {
					if !helpers.IsSkillMeltConfigured(skillId, skillMeltMap) {
						result.ErrCells = append(result.ErrCells, &json_rule.CellError{
							Index:    hero.RowIndex,
							ExcelRow: hero.RowIndex + 1,
							Reason: fmt.Sprintf("战令武将【%s】的技能 %s 未在SkillMelt表配置熔炼",
								heroName, skillId),
						})
					}
				}
			} else {
				// 临近截止时间，提前提醒
				timeUntilDeadline := protectionDeadline.Sub(now)
				if timeUntilDeadline <= warnDuration && !hero.CanMelt {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    hero.RowIndex,
						ExcelRow: hero.RowIndex + 1,
						Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 需在保护期截止 %s 前开启熔炼（剩余 %.0f 天）",
							heroName, spHero.HeroId, helpers.FormatDate(protectionDeadline), timeUntilDeadline.Hours()/24),
					})
				}
			}
		}
	}
}

// checkGeneralHeroesMelt 检查大将军武将熔炼时间
// 规则：赛季结束后需开启熔炼，提前5天提醒
// 注意：只检查 HeroType=1 的普通武将
func (c *HeroMeltCheckRule) checkGeneralHeroesMelt(heroCols [][]string, skillMeltMap map[string]*helpers.SkillMeltInfo, arenaScoreRewardsCols, arenaSeasonCols [][]string, startRowIdx int, now time.Time, warnDuration time.Duration, protectMonths int, result *json_rule.TableCheckResult) {
	if arenaScoreRewardsCols == nil || arenaSeasonCols == nil {
		return
	}

	// 查找 HeroType 列索引（只检查 HeroType=1 的普通武将）
	heroTypeColIdx := helpers.GetColIndexByName(heroCols, "HeroType")

	// 查找所有大将军武将
	generalHeroes := helpers.FindArenaGeneralHeroes(arenaScoreRewardsCols, startRowIdx)

	for _, genHero := range generalHeroes {
		// 获取赛季时间
		_, seasonEndTime := helpers.GetArenaSeasonTime(genHero.SeasonId, arenaSeasonCols, startRowIdx)
		arenaProtectionDeadline := helpers.CalcArenaProtectionDeadline(seasonEndTime)

		if seasonEndTime.IsZero() {
			continue
		}

		// 获取武将信息
		hero := helpers.FindHeroById(genHero.HeroId, heroCols, startRowIdx)
		if hero == nil {
			continue
		}

		// 检查 HeroType 列（只检查 HeroType=1 的普通武将）
		if heroTypeColIdx >= 0 {
			heroType := 1 // 默认为普通武将
			if heroTypeStr := helpers.GetColValue(heroCols, heroTypeColIdx, hero.RowIndex); heroTypeStr != "" {
				heroType, _ = strconv.Atoi(heroTypeStr)
			}
			if heroType != 1 {
				continue // HeroType≠1 表示特殊武将，跳过熔炼检查
			}
		}

		heroName := genHero.HeroName
		if heroName == "" {
			heroName = hero.Name
		}

		// 保护期已过
		if !arenaProtectionDeadline.IsZero() && arenaProtectionDeadline.Before(now) {
			if !hero.CanMelt {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    hero.RowIndex,
					ExcelRow: hero.RowIndex + 1,
					Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 赛季已结束(%s)，仍未开启熔炼",
						heroName, genHero.HeroId, helpers.FormatDate(seasonEndTime)),
				})
			}
			// 检查技能熔炼配置
			for _, skillId := range hero.Skills {
				if !helpers.IsSkillMeltConfigured(skillId, skillMeltMap) {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    hero.RowIndex,
						ExcelRow: hero.RowIndex + 1,
						Reason: fmt.Sprintf("大将军武将【%s】的技能 %s 未在SkillMelt表配置熔炼",
							heroName, skillId),
					})
				}
			}
		} else {
			// 赛季未结束，提前提醒
			timeUntilEnd := seasonEndTime.Sub(now)
			if timeUntilEnd <= warnDuration && !hero.CanMelt {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    hero.RowIndex,
					ExcelRow: hero.RowIndex + 1,
					Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 保护期将于 %s 截止，请提前准备开启熔炼（剩余 %.0f 天）",
						heroName, genHero.HeroId, helpers.FormatDate(seasonEndTime), timeUntilEnd.Hours()/24),
				})
			}
		}
	}
}

// loadRelatedSheets 加载相关的表数据
func (c *HeroMeltCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (heroCols, skillMeltCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols [][]string) {
	// Hero 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			heroCols = cols
		}
	}

	// SkillMelt 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SkillMelt"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			skillMeltCols = cols
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

// parseBoolValue 解析布尔值
func parseBoolValue(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

// parseSkillList 解析技能列表
// 格式示例: [Skill1,Skill2] 或 Skill1,Skill2
func parseSkillList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// 去除方括号
	s = strings.Trim(s, "[]")

	// 按逗号分割
	parts := strings.Split(s, ",")
	skills := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			skills = append(skills, part)
		}
	}
	return skills
}
