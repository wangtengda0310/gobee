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

// DrawFixProtectionCheckRule 定向招募战令保护期检查规则
//
// 校验规则：战令武将在保护期内（StartTime + protectMonths）不能出现在定向招募表中
// 如果 DrawFix.ItemIds 包含战令武将，且 DrawFix.EndTime < 保护期截止时间 → 错误
//
// 数据关系链：
//
//	DrawFix.ItemIds → Item(Type=Hero).ItemParam → Hero.Id
//	SeasonPassReward.HighReward → SeasonPass.StartTime/EndTime
//
// 相关表：
//   - DrawFix: Id, Name, StartTime, EndTime, ItemIds, SlotNum
//   - Item: Id, Type, ItemParam
//   - SeasonPassReward: SeasonPassId, HighReward
//   - SeasonPass: Id, StartTime, EndTime
//   - Hero: Id, Name
type DrawFixProtectionCheckRule struct{}

// 保护期计算：
//   - 战令武将：CalcSeasonPassProtectionDeadline(StartTime, protectMonths) → helpers/hero_rule_helper.go
//
// 关键函数位置：
//   - FindSeasonPassHeroes            → helpers/hero_rule_helper.go
//   - CalcSeasonPassProtectionDeadline → helpers/hero_rule_helper.go
//   - BuildHeroItemIdMap              → helpers/hero_rule_helper.go（语义映射：Item.Type=Hero → HeroId）
//
// heroProtection 记录武将的保护期信息
type heroProtection struct {
	heroId             int
	heroName           string
	seasonPassId       int
	seasonStartTime    time.Time
	protectionDeadline time.Time // SeasonPass.StartTime + protectMonths
}

// Meta 返回规则元数据
func (c *DrawFixProtectionCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DRAWFIX_PROTECTION_CHECK,
		DisplayName:    "定向招募战令保护期检查",
		Description:    "检查定向招募表(DrawFix)中是否错误配置了战令保护期内的武将。战令武将在保护期（战令开始时间起N个月）内不能出现在其他系统中。",
		TargetSheets:   []string{"DrawFix"},
		RequiredSheets: []string{"Item", "SeasonPassReward", "SeasonPass", "Hero"},
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.PROTECT_MONTHS,
				Label:       "保护期月数",
				Description: "战令开始时间起多少个月内武将不能出现在其他系统（默认4个月）",
				Type:        "number",
				Default:     "4",
				Required:    false,
			},
		},
	}
}

// Check 执行定向招募战令保护期检查
//
// 执行流程：
//  1. 初始化检查结果结构体
//  2. 解析保护期月数参数（默认4个月）
//  3. 加载关联表数据（Item、SeasonPassReward、SeasonPass、Hero）
//  4. Item 表必须存在（语义映射依赖），否则直接返回失败
//  5. 构建 ItemId→HeroId 映射（通过 Item 表的 Type=Hero 行）
//  6. 获取战令武将列表（通过 SeasonPassReward + SeasonPass）
//  7. 构建保护期映射（HeroId → 保护期信息）
//  8. 遍历 DrawFix 行，检查每个 ItemId 是否包含在保护期内的战令武将
//  9. 汇总检查结果
func (c *DrawFixProtectionCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	// 步骤1: 初始化检查结果
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "定向招募战令保护期检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 步骤2: 解析参数
	protectMonths := 4
	if val, ok := param.Params[string(json_rule.PROTECT_MONTHS)]; ok && val != "" {
		if months, err := helpers.ParseIntWithError(val); err == nil {
			protectMonths = months
		}
	}

	// 步骤3: 加载关联表
	itemCols, seasonPassRewardCols, seasonPassCols, heroCols := c.loadRelatedSheets(param.SheetMap)

	// 步骤4: Item 表必须存在（语义映射依赖）
	if itemCols == nil {
		result.Ok = false
		result.Reason = "未找到 Item 表数据，无法进行语义映射"
		return result
	}

	// 步骤5: 构建 ItemId→HeroId 映射
	heroItemIdMap := helpers.BuildHeroItemIdMap(itemCols, param.StartRowIdx)

	// 步骤6: 获取战令武将列表
	now := helpers.ResolveNow(param.Now)
	if seasonPassRewardCols == nil || seasonPassCols == nil {
		// 无战令数据，没有需要保护的武将，直接返回通过
		return result
	}

	seasonPassHeroes := helpers.FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, param.StartRowIdx)

	// 步骤7: 构建保护期映射
	heroProtectionMap := c.buildHeroProtectionMap(seasonPassHeroes, heroCols, protectMonths, param.StartRowIdx)

	// 步骤8: 遍历 DrawFix 行检查
	c.checkDrawFixRows(param.Cols, param.StartRowIdx, heroItemIdMap, heroProtectionMap, now, result)

	// 步骤9: 设置结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个定向招募战令保护期配置问题", len(result.ErrCells))
	}

	return result
}

// loadRelatedSheets 加载关联表数据
//
// 从 sheetMap 中查找 Item、SeasonPassReward、SeasonPass、Hero 四张表的数据，
// 使用 FindSheetBySuffix 进行后缀匹配查找。
// 返回各表的列数据（二维字符串数组），未找到的表返回 nil。
func (c *DrawFixProtectionCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (itemCols, seasonPassRewardCols, seasonPassCols, heroCols [][]string) {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Item"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			itemCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPassReward"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			seasonPassRewardCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPass"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			seasonPassCols = cols
		}
	}
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		if cols, e := file.GetCols(sheetName); e == nil {
			heroCols = cols
		}
	}
	return
}

// buildHeroProtectionMap 构建武将保护期映射
//
// 将战令武将列表转换为 HeroId → 保护期信息的映射。
// 同一武将出现在多个战令中时，取最晚的保护期截止时间。
// 保护期截止时间 = SeasonPass.StartTime + protectMonths
//
// 参数：
//   - seasonPassHeroes: 战令武将列表
//   - heroCols: Hero 表列数据（用于获取武将名称）
//   - protectMonths: 保护期月数
//   - startRowIdx: 数据起始行索引
func (c *DrawFixProtectionCheckRule) buildHeroProtectionMap(seasonPassHeroes []*helpers.SeasonPassHero, heroCols [][]string, protectMonths, startRowIdx int) map[int]*heroProtection {
	protectionMap := make(map[int]*heroProtection)

	for _, spHero := range seasonPassHeroes {
		if spHero.StartTime.IsZero() {
			continue
		}

		deadline := spHero.StartTime.AddDate(0, protectMonths, 0)

		// 获取武将名称
		heroName := spHero.HeroName
		if heroName == "" && heroCols != nil {
			if hero := helpers.FindHeroById(spHero.HeroId, heroCols, startRowIdx); hero != nil {
				heroName = hero.Name
			}
		}
		if heroName == "" {
			heroName = fmt.Sprintf("未知(ID=%d)", spHero.HeroId)
		}

		// 如果同一武将出现在多个战令中，取最晚的 protectionDeadline
		if existing, ok := protectionMap[spHero.HeroId]; ok {
			if deadline.After(existing.protectionDeadline) {
				existing.protectionDeadline = deadline
				existing.seasonPassId = spHero.SeasonPassId
				existing.seasonStartTime = spHero.StartTime
			}
		} else {
			protectionMap[spHero.HeroId] = &heroProtection{
				heroId:             spHero.HeroId,
				heroName:           heroName,
				seasonPassId:       spHero.SeasonPassId,
				seasonStartTime:    spHero.StartTime,
				protectionDeadline: deadline,
			}
		}
	}

	return protectionMap
}

// checkDrawFixRows 遍历 DrawFix 行检查战令保护期冲突
//
// 执行流程：
//  1. 查找 DrawFix 表中所需列的索引（Id、Name、EndTime、ItemIds）
//  2. 遍历每一行数据：
//     a. 读取 DrawFix 的 Id 和 Name
//     b. 解析 EndTime，跳过无结束时间的行
//     c. 解析 ItemIds（格式：{ItemId;Count},...）
//     d. 通过 heroItemIdMap 将 ItemId 映射为 HeroId
//     e. 检查 HeroId 是否在保护期映射中
//     f. 比较 DrawFix.EndTime 与保护期截止时间
func (c *DrawFixProtectionCheckRule) checkDrawFixRows(cols [][]string, startRowIdx int, heroItemIdMap map[int]int, heroProtectionMap map[int]*heroProtection, now time.Time, result *json_rule.TableCheckResult) {
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
			protection, inProtection := heroProtectionMap[heroId]
			if !inProtection {
				continue // 非战令武将，跳过
			}

			// 比较 DrawFix.EndTime 与保护期截止时间
			// EndTime >= protectionDeadline 算通过（保护期已过）
			if endTime.Before(protection.protectionDeadline) {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason: fmt.Sprintf("定向招募【%s】(ID=%d) 结束时间 %s 在战令保护期内，包含战令武将【%s】(HeroID=%d，保护期至 %s)",
						drawFixName, drawFixId,
						helpers.FormatDateTime(endTime),
						protection.heroName, heroId,
						helpers.FormatDateTime(protection.protectionDeadline)),
				})
			}
		}
	}
}
