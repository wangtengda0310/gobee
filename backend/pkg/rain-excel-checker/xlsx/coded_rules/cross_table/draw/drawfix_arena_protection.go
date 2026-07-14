// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package draw

import (
	"fmt"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DrawFixArenaProtectionCheckRule 定向招募大将军保护期检查规则
//
// 校验规则：
//  1. 反向检查：DrawFix.EndTime < ArenaSeason.SeasonEndTime → 违规（赛季期间大将军武将不能出现在定向招募）
//
// 数据关系链：
//
//	DrawFix.ItemIds → Item(Type=Hero).ItemParam → Hero.Id
//	ArenaScoreReward.Reward → ArenaScoreReward.Season → ArenaSeason.Id
//	ArenaSeason.SeasonEndTime（保护期截止时间，直接使用）
//
// 相关表：
//   - DrawFix: Id, Name, StartTime, EndTime, ItemIds, SlotNum
//   - Item: Id, Type, ItemParam
//   - ArenaScoreReward: Season, Dan, DanName, Reward（单数，无s）
//   - ArenaSeason: Id, SeasonStartTime, SeasonEndTime
//   - Hero: Id, Name
type DrawFixArenaProtectionCheckRule struct{}

// 保护期计算：
//   - 大将军武将：CalcArenaProtectionDeadline(seasonEndTime) → 直接返回 SeasonEndTime
//   - 不使用 protectMonths 参数，赛季结束即保护期结束
//
// 关键函数位置：
//   - FindArenaGeneralHeroes          → helpers/hero_rule_helper.go
//   - CalcArenaProtectionDeadline     → helpers/hero_rule_helper.go（返回 SeasonEndTime）
//   - BuildHeroItemIdMap              → helpers/hero_rule_helper.go（语义映射：Item.Type=Hero → HeroId）
//
// arenaHeroProtection 记录大将军武将的保护期信息
type arenaHeroProtection struct {
	heroId          int
	heroName        string
	seasonId        int
	seasonStartTime time.Time
	seasonEndTime   time.Time // ArenaSeason.SeasonEndTime（保护期截止时间 = 赛季结束时间）
}

// Meta 返回规则元数据
func (c *DrawFixArenaProtectionCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DRAWFIX_ARENA_PROTECTION_CHECK,
		DisplayName:    "定向招募大将军保护期检查",
		Description:    "检查定向招募表(DrawFix)中是否错误配置了竞技场赛季保护期内的大将军武将。大将军武将在赛季期间不能出现在定向招募中。",
		TargetSheets:   []string{"DrawFix"},
		RequiredSheets: []string{"Item", "ArenaScoreReward", "ArenaSeason", "Hero"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行定向招募大将军保护期检查
//
// 执行流程：
//  1. 初始化检查结果结构体
//  2. 加载关联表数据（Item、ArenaScoreReward、ArenaSeason、Hero）
//  3. Item 表必须存在（语义映射依赖），否则直接返回失败
//  4. 构建 ItemId→HeroId 映射（通过 Item 表的 Type=Hero 行）
//  5. 获取大将军武将列表（通过 ArenaScoreReward + ArenaSeason）
//  6. 为每个大将军武将填充赛季时间（通过 GetArenaSeasonTime）
//  7. 构建 HeroId→保护期映射（同一HeroId取最晚SeasonEndTime）
//  8. 执行反向检查（DrawFix中的大将军武将，EndTime不能小于保护期截止时间）
//
// 10. 汇总检查结果
func (c *DrawFixArenaProtectionCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	// 步骤1: 初始化检查结果
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "定向招募大将军保护期检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 步骤2: 加载关联表
	itemCols, arenaScoreRewardCols, arenaSeasonCols, heroCols := c.loadArenaRelatedSheets(param.SheetMap)

	// 步骤3: Item 表必须存在（语义映射依赖）
	if itemCols == nil {
		result.Ok = false
		result.Reason = "未找到 Item 表数据，无法进行语义映射"
		return result
	}

	// 步骤4: 构建 ItemId→HeroId 映射
	heroItemIdMap := helpers.BuildHeroItemIdMap(itemCols, param.StartRowIdx)

	// 步骤5: 获取大将军武将列表
	if arenaScoreRewardCols == nil || arenaSeasonCols == nil {
		// 无大将军数据，直接返回通过
		return result
	}

	generalHeroes := helpers.FindArenaGeneralHeroes(arenaScoreRewardCols, param.StartRowIdx)

	// 步骤6: 为每个大将军武将填充赛季时间
	for _, hero := range generalHeroes {
		if hero.SeasonId > 0 {
			startTime, endTime := helpers.GetArenaSeasonTime(hero.SeasonId, arenaSeasonCols, param.StartRowIdx)
			hero.SeasonStartTime = startTime
			hero.SeasonEndTime = endTime
		}
	}

	// 步骤7: 构建 HeroId→保护期映射（同一HeroId取最晚SeasonEndTime）
	protectionMap := c.buildArenaProtectionMap(generalHeroes, heroCols, param.StartRowIdx)

	// 步骤8: 执行反向检查
	c.checkReverseProtection(param.Cols, param.StartRowIdx, heroItemIdMap, protectionMap, result)

	// 步骤9: 设置结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个定向招募大将军保护期配置问题", len(result.ErrCells))
	}

	return result
}

// loadArenaRelatedSheets 加载大将军相关的4张关联表
//
// 从 sheetMap 中查找 Item、ArenaScoreReward、ArenaSeason、Hero 四张表的数据，
// 使用 FindSheetBySuffix 进行后缀匹配查找。
// 注意：ArenaScoreReward 是单数，FindSheetBySuffix 必须使用 "ArenaScoreReward" 不带s
// 返回各表的列数据（二维字符串数组），未找到的表返回 nil。
func (c *DrawFixArenaProtectionCheckRule) loadArenaRelatedSheets(sheetMap map[string]*excelize.File) (itemCols, arenaScoreRewardCols, arenaSeasonCols, heroCols [][]string) {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Item"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			itemCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaScoreReward"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			arenaScoreRewardCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaSeason"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			arenaSeasonCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			heroCols = cols
		}
	}
	return
}

// buildArenaProtectionMap 构建大将军武将保护期映射
//
// 将大将军武将列表转换为 HeroId → 保护期信息的映射。
// 同一武将出现在多个赛季中时，取最晚的 SeasonEndTime（保护期截止时间）。
// 保护期截止时间 = ArenaSeason.SeasonEndTime（赛季结束时间）
//
// 参数：
//   - generalHeroes: 大将军武将列表
//   - heroCols: Hero 表列数据（用于获取武将名称）
//   - startRowIdx: 数据起始行索引
func (c *DrawFixArenaProtectionCheckRule) buildArenaProtectionMap(generalHeroes []*helpers.GeneralHero, heroCols [][]string, startRowIdx int) map[int]*arenaHeroProtection {
	protectionMap := make(map[int]*arenaHeroProtection)

	for _, hero := range generalHeroes {
		// 获取武将名称
		heroName := hero.HeroName
		if heroName == "" && heroCols != nil {
			if h := helpers.FindHeroById(hero.HeroId, heroCols, startRowIdx); h != nil {
				heroName = h.Name
			}
		}
		if heroName == "" {
			heroName = fmt.Sprintf("未知(ID=%d)", hero.HeroId)
		}

		// 如果同一武将出现在多个赛季中，取最晚的 SeasonEndTime
		if existing, ok := protectionMap[hero.HeroId]; ok {
			if !hero.SeasonEndTime.IsZero() && hero.SeasonEndTime.After(existing.seasonEndTime) {
				existing.seasonId = hero.SeasonId
				existing.seasonStartTime = hero.SeasonStartTime
				existing.seasonEndTime = hero.SeasonEndTime
				// seasonEndTime 已更新，保护期截止时间 = seasonEndTime
			}
		} else {
			protectionMap[hero.HeroId] = &arenaHeroProtection{
				heroId:          hero.HeroId,
				heroName:        heroName,
				seasonId:        hero.SeasonId,
				seasonStartTime: hero.SeasonStartTime,
				seasonEndTime:   hero.SeasonEndTime,
				// 保护期截止时间 = seasonEndTime
			}
		}
	}

	return protectionMap
}

// checkReverseProtection 反向检查：DrawFix中的大将军武不能在赛季期间出现
//
// 执行流程：
//  1. 查找 DrawFix 表中所需列的索引（Id、Name、EndTime、ItemIds）
//  2. 遍历 DrawFix 每一行数据：
//     a. 读取 Id、Name、EndTime
//     b. 解析 ItemIds（格式：{ItemId;Count},...）
//     c. 通过 heroItemIdMap 将 ItemId 映射为 HeroId
//     d. 检查 HeroId 是否在保护期映射中
//     e. 比较 DrawFix.EndTime 与保护期截止时间（SeasonStartTime + 4个月）
//     f. 如果 EndTime < SeasonEndTime，说明在赛季期间，添加错误
func (c *DrawFixArenaProtectionCheckRule) checkReverseProtection(cols [][]string, startRowIdx int, heroItemIdMap map[int]int, protectionMap map[int]*arenaHeroProtection, result *json_rule.TableCheckResult) {
	idColIdx := helpers.GetColIndexByName(cols, "Id")
	nameColIdx := helpers.GetColIndexByName(cols, "Name")
	endTimeColIdx := helpers.GetColIndexByName(cols, "EndTime")
	itemIdsColIdx := helpers.GetColIndexByName(cols, "ItemIds")

	if idColIdx < 0 || itemIdsColIdx < 0 {
		return
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(cols, startRowIdx); rowIdx++ {
		// 读取 DrawFix 行数据
		drawFixIdStr := helpers.GetColValue(cols, idColIdx, rowIdx)
		if drawFixIdStr == "" {
			continue
		}
		drawFixId, err := helpers.ParseIntWithError(drawFixIdStr)
		if err != nil {
			continue
		}

		drawFixName := ""
		if nameColIdx >= 0 {
			drawFixName = helpers.GetColValue(cols, nameColIdx, rowIdx)
		}

		// 解析 EndTime
		var endTime time.Time
		if endTimeColIdx >= 0 {
			endTime = helpers.ParseDate(helpers.GetColValue(cols, endTimeColIdx, rowIdx))
		}
		if endTime.IsZero() {
			continue // 无结束时间，跳过
		}

		// 解析 ItemIds
		itemIdsStr := helpers.GetColValue(cols, itemIdsColIdx, rowIdx)
		if itemIdsStr == "" {
			continue
		}

		items := helpers.ParseItemCfg(itemIdsStr)
		for _, item := range items {
			// 通过语义映射获取 HeroId
			heroId, isHero := heroItemIdMap[item.ItemId]
			if !isHero {
				continue // 非武将道具，跳过
			}

			// 检查是否在保护期映射中
			protection, inProtection := protectionMap[heroId]
			if !inProtection {
				continue // 非大将军武将，跳过
			}

			// 检查是否在保护期内：DrawFix.EndTime < ArenaSeason.SeasonEndTime（赛季结束即保护期结束）
			// 规则：保护期内大将军武将不能出现在定向招募中
			if endTime.Before(protection.seasonEndTime) && !protection.seasonEndTime.IsZero() {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason: fmt.Sprintf("定向招募【%s】(ID=%d) 结束时间 %s 在竞技场赛季期间，包含大将军武将【%s】(HeroID=%d，赛季ID=%d，保护期至 %s)",
						drawFixName, drawFixId,
						helpers.FormatDateTime(endTime),
						protection.heroName, heroId,
						protection.seasonId,
						helpers.FormatDateTime(protection.seasonEndTime)),
				})
			}
		}
	}
}
