package coded_rules

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// HeroIsOpenOpenDateCheckRule 武将 IsOpen 与 OpenDate 一致性检查（仅战令/大将军武将）
//
// 规则：
//   - 仅对战令武将和大将军武将检查：IsOpen=true 时 OpenDate 不允许为空
//   - 普通武将（非战令、非大将军）IsOpen=true 时 OpenDate 可以为空（一直在线上运行）
//   - IsOpen=false 时，OpenDate 可以为空（合法）
//
// 关联表：
//   - SeasonPassReward + SeasonPass：获取战令武将ID列表
//   - ArenaScoreReward：获取大将军武将ID列表
type HeroIsOpenOpenDateCheckRule struct{}

func (c *HeroIsOpenOpenDateCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.HERO_ISOPEN_OPENDATE_CHECK,
		DisplayName:    "武将IsOpen与OpenDate一致性检查",
		Description:    "仅检查战令/大将军武将 IsOpen=true 时 OpenDate 是否已配置，普通武将不检查",
		TargetSheets:   []string{"Hero"},
		RequiredSheets: []string{"SeasonPassReward", "SeasonPass", "ArenaScoreReward"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行 Hero 表 IsOpen+OpenDate 一致性检查
// 仅对战令武将和大将军武将检查 IsOpen=true 时 OpenDate 是否为空
func (c *HeroIsOpenOpenDateCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "武将IsOpen与OpenDate一致性检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	if param.Cols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	// 构建战令/大将军武将ID集合
	specialHeroIds := c.buildSpecialHeroIdSet(param)

	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	isOpenColIdx := helpers.GetColIndexByName(param.Cols, "IsOpen")
	openDateColIdx := helpers.GetColIndexByName(param.Cols, "OpenDate")

	if idColIdx < 0 {
		result.Ok = false
		result.Reason = "Hero 表缺少 Id 列"
		return result
	}

	startRowIdx := param.StartRowIdx
	if startRowIdx == 0 {
		startRowIdx = excelio.MJS_FIXED_ROWS_NUM
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(param.Cols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}

		// 跳过非战令/大将军武将
		heroId, _ := helpers.ParseIntParamWithError(idStr)
		if !specialHeroIds[heroId] {
			continue
		}

		isOpen := false
		if isOpenColIdx >= 0 {
			isOpen = helpers.ParseBool(helpers.GetColValue(param.Cols, isOpenColIdx, rowIdx))
		}

		if !isOpen {
			continue
		}

		openDateStr := ""
		if openDateColIdx >= 0 {
			openDateStr = helpers.GetColValue(param.Cols, openDateColIdx, rowIdx)
		}

		if openDateStr == "" {
			heroName := ""
			if nameColIdx >= 0 {
				heroName = helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason: fmt.Sprintf("战令/大将军武将【%s】(ID=%s) IsOpen=true 但 OpenDate 未配置",
					heroName, idStr),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个战令/大将军武将 IsOpen/OpenDate 配置问题", len(result.ErrCells))
	}

	return result
}

// buildSpecialHeroIdSet 从关联表构建战令+大将军武将ID集合
// 返回 map[int]bool，key 为武将ID
func (c *HeroIsOpenOpenDateCheckRule) buildSpecialHeroIdSet(param json_rule.CheckParam) map[int]bool {
	heroIds := make(map[int]bool)

	if param.SheetMap == nil {
		return heroIds
	}

	startRowIdx := excelio.MJS_FIXED_ROWS_NUM

	// 战令武将：从 SeasonPassReward + SeasonPass 获取
	seasonPassRewardCols := c.loadSheetCols(param.SheetMap, "SeasonPassReward")
	seasonPassCols := c.loadSheetCols(param.SheetMap, "SeasonPass")
	if seasonPassRewardCols != nil && seasonPassCols != nil {
		spHeroes := helpers.FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, startRowIdx)
		for _, h := range spHeroes {
			heroIds[h.HeroId] = true
		}
	}

	// 大将军武将：从 ArenaScoreReward 获取
	arenaScoreRewardCols := c.loadSheetCols(param.SheetMap, "ArenaScoreReward")
	if arenaScoreRewardCols != nil {
		generalHeroes := helpers.FindArenaGeneralHeroes(arenaScoreRewardCols, startRowIdx)
		for _, h := range generalHeroes {
			heroIds[h.HeroId] = true
		}
	}

	return heroIds
}

// loadSheetCols 从 SheetMap 加载指定表的列数据
func (c *HeroIsOpenOpenDateCheckRule) loadSheetCols(sheetMap map[string]*excelize.File, suffix string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, suffix); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}
