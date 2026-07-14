// Package table 提供表级别的校验规则
// 本包中的规则针对单个 Excel 表的特定业务逻辑进行校验

package coded_rules

import (
	"fmt"
	"strconv"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// SeasonPassHeroOpenCheckRule 战令武将开放时间检查规则
//
// 原始需求：https://ztgame.feishu.cn/wiki/TOJlw8ucyiNkfMkuoyjcJi8XnLc
//
// 需求确认记录：
// - 2024-xx-xx: 确认校验规则：从 SeasonPassReward 表获取 HighReward 第一个包含武将道具的奖励
// - 2024-xx-xx: 确认对比规则：武将 OpenDate 与战令 StartTime 必须是同一天（精确匹配）
// - 2024-xx-xx: 确认检查范围：检查所有赛季（配置错误应在配置时就发现）
//
// 相关表结构：
// - SeasonPass: Id, StartTime, EndTime
// - SeasonPassReward: SeasonPassId, Level, HighReward (格式: {物品ID;数量})
// - Hero: Id, Name, IsOpen, OpenDate
//
// 物品ID规则：
// - 物品ID前两位为10表示武将道具
// - 武将ID = 物品ID % 100000
type SeasonPassHeroOpenCheckRule struct{}

// Meta 返回规则元数据
func (c *SeasonPassHeroOpenCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.SEASON_PASS_HERO_OPEN_CHECK,
		DisplayName:    "战令武将开放时间检查",
		Description:    "检查战令奖励中的武将开放时间是否与战令开始时间一致",
		TargetSheets:   []string{"SeasonPassReward"},   // 适用于 SeasonPassReward 表（需访问 Hero, SeasonPass）
		RequiredSheets: []string{"Hero", "SeasonPass"}, // 执行检查需要加载的关联表
		ParamDefs: []json_rule.TableRuleParamDef{
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
// 1. 从 SeasonPassReward 表获取 HighReward 第一个包含武将道具的奖励
// 2. 从 Hero 表获取该武将的 OpenDate
// 3. 对比 OpenDate 与战令 StartTime 是否同一天（精确匹配）
// 4. 检查所有赛季，包括已过期的（配置错误应在配置时就发现）
func (c *SeasonPassHeroOpenCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "战令武将开放时间检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 加载相关表数据
	seasonPassRewardCols, seasonPassCols, heroCols := c.loadRelatedSheets(param.SheetMap)

	if seasonPassRewardCols == nil {
		result.Ok = false
		result.Reason = "未找到 SeasonPassReward 表数据"
		return result
	}

	if seasonPassCols == nil {
		result.Ok = false
		result.Reason = "未找到 SeasonPass 表数据"
		return result
	}

	if heroCols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	// 查找所有战令武将
	seasonPassHeroes := findSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, excelio.MJS_FIXED_ROWS_NUM)

	// 解析当前时间（单元测试可通过 param.Now 注入固定时间）
	now := helpers.ResolveNow(param.Now)

	// 检查每个战令武将
	// 注意：配置错误应在配置时就发现，不应该跳过已过期的赛季
	for _, spHero := range seasonPassHeroes {
		// 查找武将信息
		hero := findHeroById(spHero.HeroId, heroCols, excelio.MJS_FIXED_ROWS_NUM)
		if hero == nil {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: spHero.RowIndex,
				Reason: fmt.Sprintf("[赛季%d] 武将道具ID=%d(武将ID=%d)在Hero表中不存在 | HighReward=%s",
					spHero.SeasonPassId, spHero.ItemId, spHero.HeroId, spHero.HighRewardValue),
			})
			continue
		}

		// 业务规则：
		// 0. 战令武将必须 IsOpen=true
		// 1. 赛季未结束的武将必须配置 OpenDate
		// 2. 赛季已结束时，OpenDate 可以为空
		// 3. OpenDate 不为空时，必须与战令 StartTime 精确匹配（精确到秒）

		// 检查 IsOpen：战令武将必须开放
		if !hero.IsOpen {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: spHero.RowIndex,
				Reason: fmt.Sprintf("[赛季%d] 战令武将【%s】(ID=%d) IsOpen=false，战令武将必须设置为IsOpen=true（规则：战令武将必须开放）| 道具ID=%d | HighReward=%s",
					spHero.SeasonPassId, hero.Name, spHero.HeroId, spHero.ItemId, spHero.HighRewardValue),
			})
			continue
		}

		if hero.OpenDate.IsZero() {
			// OpenDate 为空时，检查赛季是否已结束
			if spHero.EndTime.Before(now) {
				// 赛季已结束，OpenDate 可以为空，跳过检查
			} else {
				// 赛季未结束，OpenDate 不能为空，报错
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index: spHero.RowIndex,
					Reason: fmt.Sprintf("[赛季%d] 赛季未结束(结束于%s)，战令武将【%s】(ID=%d, OpenDate未配置)必须配置OpenDate（规则：赛季未结束时战令武将必须配置OpenDate）| 道具ID=%d | HighReward=%s",
						spHero.SeasonPassId, helpers.FormatDateTime(spHero.EndTime),
						hero.Name, spHero.HeroId, spHero.ItemId, spHero.HighRewardValue),
				})
			}
		} else {
			// OpenDate 不为空，检查是否与战令开始时间一致（精确到秒）
			if !helpers.TimeEquals(hero.OpenDate, spHero.StartTime) {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index: spHero.RowIndex,
					Reason: fmt.Sprintf("[赛季%d] 战令武将【%s】(ID=%d)开放时间不匹配 | OpenDate=%s vs StartTime=%s（规则：战令武将开放时间必须与战令开始时间一致）| 道具ID=%d | HighReward=%s",
						spHero.SeasonPassId, hero.Name, spHero.HeroId,
						helpers.FormatDateTime(hero.OpenDate), helpers.FormatDateTime(spHero.StartTime),
						spHero.ItemId, spHero.HighRewardValue),
				})
			}
		}
	}

	// 设置结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个战令武将开放时间配置问题", len(result.ErrCells))
	}

	return result
}

// loadRelatedSheets 加载相关的表数据
func (c *SeasonPassHeroOpenCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (seasonPassRewardCols, seasonPassCols, heroCols [][]string) {
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

	// Hero 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			heroCols = cols
		}
	}

	return seasonPassRewardCols, seasonPassCols, heroCols
}

// ==================== 辅助类型和函数 ====================

// HeroRow 武将行数据
type HeroRow struct {
	Id       int
	Name     string
	IsOpen   bool
	OpenDate time.Time
	RowIndex int
}

// SeasonPassHero 战令武将信息
type SeasonPassHero struct {
	HeroId          int
	HeroName        string
	SeasonPassId    int
	StartTime       time.Time
	EndTime         time.Time
	RowIndex        int
	ItemId          int    // 武将道具ID
	HighRewardValue string // HighReward原始值
}

// findHeroById 根据武将ID查找武将行
func findHeroById(heroId int, heroCols [][]string, startRowIdx int) *HeroRow {
	idColIdx := helpers.GetColIndexByName(heroCols, "Id")
	nameColIdx := helpers.GetColIndexByName(heroCols, "Name")
	isOpenColIdx := helpers.GetColIndexByName(heroCols, "IsOpen")
	openDateColIdx := helpers.GetColIndexByName(heroCols, "OpenDate")

	if idColIdx < 0 {
		return nil
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(heroCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(heroCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id == heroId {
			hero := &HeroRow{
				Id:       heroId,
				RowIndex: rowIdx,
			}

			if nameColIdx >= 0 {
				hero.Name = helpers.GetColValue(heroCols, nameColIdx, rowIdx)
			}
			if isOpenColIdx >= 0 {
				hero.IsOpen = helpers.ParseBool(helpers.GetColValue(heroCols, isOpenColIdx, rowIdx))
			}
			if openDateColIdx >= 0 {
				hero.OpenDate = helpers.ParseDate(helpers.GetColValue(heroCols, openDateColIdx, rowIdx))
			}

			return hero
		}
	}
	return nil
}

// findSeasonPassHeroes 查找所有战令武将
// 规则：按SeasonPassReward表行顺序，找到第一个HighReward包含武将道具的数据
func findSeasonPassHeroes(seasonPassRewardCols, seasonPassCols [][]string, startRowIdx int) []*SeasonPassHero {
	heroes := make([]*SeasonPassHero, 0)

	seasonPassIdColIdx := helpers.GetColIndexByName(seasonPassRewardCols, "SeasonPassId")
	highRewardColIdx := helpers.GetColIndexByName(seasonPassRewardCols, "HighReward")

	if highRewardColIdx < 0 {
		return heroes
	}

	// 用于去重
	foundHeroIds := make(map[int]bool)

	// 遍历SeasonPassReward表
	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(seasonPassRewardCols, startRowIdx); rowIdx++ {
		highReward := helpers.GetColValue(seasonPassRewardCols, highRewardColIdx, rowIdx)
		if highReward == "" {
			continue
		}

		// 解析HighReward中的物品
		items := helpers.ParseItemCfg(highReward)
		for _, item := range items {
			// 检查是否是武将道具
			if helpers.IsHeroItem(item.ItemId) {
				heroId := helpers.ExtractHeroIdFromItemCfg(item.ItemId)
				if heroId > 0 && !foundHeroIds[heroId] {
					// 获取战令信息
					seasonPassId := 0
					if seasonPassIdColIdx >= 0 {
						seasonPassId, _ = strconv.Atoi(helpers.GetColValue(seasonPassRewardCols, seasonPassIdColIdx, rowIdx))
					}

					hero := &SeasonPassHero{
						HeroId:          heroId,
						SeasonPassId:    seasonPassId,
						RowIndex:        rowIdx,
						ItemId:          item.ItemId,
						HighRewardValue: highReward,
					}

					// 获取战令时间
					if seasonPassCols != nil && seasonPassId > 0 {
						startTime, endTime := getSeasonPassTime(seasonPassId, seasonPassCols, startRowIdx)
						hero.StartTime = startTime
						hero.EndTime = endTime
					}

					heroes = append(heroes, hero)
					foundHeroIds[heroId] = true
					break // 只取第一个武将
				}
			}
		}
	}

	return heroes
}

// getSeasonPassTime 获取战令时间
func getSeasonPassTime(seasonPassId int, seasonPassCols [][]string, startRowIdx int) (start, end time.Time) {
	idColIdx := helpers.GetColIndexByName(seasonPassCols, "Id")
	startTimeColIdx := helpers.GetColIndexByName(seasonPassCols, "StartTime")
	endTimeColIdx := helpers.GetColIndexByName(seasonPassCols, "EndTime")

	if idColIdx < 0 {
		return
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(seasonPassCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(seasonPassCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id == seasonPassId {
			if startTimeColIdx >= 0 {
				start = helpers.ParseDate(helpers.GetColValue(seasonPassCols, startTimeColIdx, rowIdx))
			}
			if endTimeColIdx >= 0 {
				end = helpers.ParseDate(helpers.GetColValue(seasonPassCols, endTimeColIdx, rowIdx))
			}
			return
		}
	}
	return
}
