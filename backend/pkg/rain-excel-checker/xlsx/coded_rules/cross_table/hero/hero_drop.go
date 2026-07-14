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

// HeroDropCheckRule 武将抽卡掉落检查规则
//
// 原始需求：https://ztgame.feishu.cn/wiki/TOJlw8ucyiNkfMkuoyjcJi8XnLc
//
// ## 校验规则（来源：需求确认）
// 1. 开放武将判断标准：IsOpen=true 且（OpenDate为空或已过）才算已开放
// 2. 掉落库检查范围：检查所有DropRule，武将出现在任何掉落规则中都算已加入掉落库
// 3. ValidDate/ExpireDate：只检查过期，不检查未生效（允许提前配置）
// 4. 提前5天提醒：每日构建报警 + 飞书即时通知
//
// ## 需求确认记录
// - 2024-xx-xx: 确认开放武将判断标准：IsOpen=true 且（OpenDate为空或已过）
// - 2024-xx-xx: 确认掉落库检查范围：检查所有DropRule，武将出现在任何掉落规则中都算已加入掉落库
// - 2024-xx-xx: 确认允许提前配置：掉落配置的ValidDate可以在未来
// - 2024-xx-xx: 确认提前5天提醒：每日构建报警 + 飞书即时通知
//
// ## 相关表结构
// - DropRule: Id, Name, Count, DropGroup, EnsureSmallGroup, EnsureBigGroup
// - DropGroup: Id, Name, Weight, ValidDate, ExpireDate
// - DropItem: Id, DropGroup, Item, ValidDate, ExpireDate
//   - Item 字段格式：{{物品ID;数量}...}
//
// - Hero: Id, Name, IsOpen, OpenDate
// - SeasonPass: Id, StartTime, EndTime
// - SeasonPassReward: SeasonPassId, HighReward
// - ArenaScoreRewards: Season, Dan, DanName, Reward
// - ArenaSeason: Id, SeasonStartTime, SeasonEndTime
//
// ## 数据流（7张表的关系）
//
// 战令路径：
//
//	Hero.IsOpen/OpenDate → 判断武将是否开放
//	SeasonPassReward.HighReward → 提取战令武将道具ID
//	SeasonPassReward.SeasonPassId → SeasonPass.Id → StartTime（保护期起算点）
//	DropItem.Item → 匹配武将道具ID → ValidDate/ExpireDate（掉落时间）
//	DropRule/DropGroup → 判断武将是否在掉落池中
//
// 大将军路径：
//
//	ArenaScoreReward.Reward(DanName含"大将军") → 提取武将道具ID
//	ArenaScoreReward.Season → ArenaSeason.Id → SeasonEndTime（保护期截止时间）
//	DropItem.Item → 匹配武将道具ID → ValidDate/ExpireDate
//
// ## 关键函数位置
//   - CalcSeasonPassProtectionDeadline → helpers/hero_rule_helper.go（战令保护期）
//   - CalcArenaProtectionDeadline     → helpers/hero_rule_helper.go（大将军保护期，返回SeasonEndTime）
//   - FindSeasonPassHeroes            → helpers/hero_rule_helper.go
//   - FindArenaGeneralHeroes          → helpers/hero_rule_helper.go
//   - GetHeroDropPoolStatus           → helpers/hero_rule_helper.go（掉落池状态判断）
//
// ## 物品ID规则
// - 武将道具ID = 1000000 + 武将ID
// - 物品ID前两位为10表示武将道具
type HeroDropCheckRule struct{}

// Meta 返回规则元数据
func (c *HeroDropCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.HERO_DROP_CHECK,
		DisplayName:    "武将抽卡掉落检查",
		Description:    "检查武将是否正确配置到掉落库中。包括：1)已开放的武将必须加入掉落库；2)战令结束后3个月的武将需加入掉落库；3)大将军赛季结束后的武将需加入掉落库。",
		TargetSheets:   []string{"Hero"},                                                                          // 适用于 Hero 表
		RequiredSheets: []string{"DropItem", "SeasonPassReward", "SeasonPass", "ArenaScoreReward", "ArenaSeason"}, // 跨表关联
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.WARN_DAYS_BEFORE,
				Label:       "提前警告天数",
				Description: "在需要加入掉落库前多少天开始警告（默认5天）",
				Type:        "number",
				Default:     "5",
				Required:    false,
			},
			{
				Key:         json_rule.PROTECT_MONTHS,
				Label:       "保护期月数",
				Description: "战令开始时间起多少个月内武将不应在掉落池中（默认4个月）",
				Type:        "number",
				Default:     "4",
				Required:    false,
			},
			{
				Key:         json_rule.IS_OPEN_COL_NAME,
				Label:       "是否开放列名",
				Description: "Hero表中是否开放的列名（默认: IsOpen）",
				Type:        "string",
				Default:     "IsOpen",
				Required:    false,
			},
			{
				Key:         json_rule.OPEN_DATE_COL_NAME,
				Label:       "开放时间列名",
				Description: "Hero表中开放时间的列名（默认: OpenDate）",
				Type:        "string",
				Default:     "OpenDate",
				Required:    false,
			},
		},
	}
}

// Check 执行武将抽卡掉落检查
func (c *HeroDropCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	// 步骤1: 初始化检查结果结构体
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "武将抽卡掉落检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 步骤2: 解析参数
	warnDaysBefore := 5
	if val, ok := param.Params[string(json_rule.WARN_DAYS_BEFORE)]; ok && val != "" {
		if days, err := helpers.ParseIntWithError(val); err == nil {
			warnDaysBefore = days
		}
	}

	protectMonths := helpers.DefaultProtectMonths
	if val, ok := param.Params[string(json_rule.PROTECT_MONTHS)]; ok && val != "" {
		if months, err := helpers.ParseIntWithError(val); err == nil {
			protectMonths = months
		}
	}

	// 步骤3: 加载相关表数据
	heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols, err := c.loadRelatedSheets(param.SheetMap)
	if err != nil {
		result.Ok = false
		result.Reason = err.Error()
		return result
	}

	// 步骤4: 验证必需的表数据存在
	if heroCols == nil {
		result.Ok = false
		result.Reason = "未找到 Hero 表数据"
		return result
	}

	if dropItemCols == nil {
		result.Ok = false
		result.Reason = "未找到 DropItem 表数据"
		return result
	}

	// 步骤5: 计算时间阈值
	now := helpers.ResolveNow(param.Now) // 单元测试可通过 param.Now 注入固定时间
	warnDuration := time.Duration(warnDaysBefore) * 24 * time.Hour

	// 步骤6: 执行三类检查
	// 步骤6a: 检查已开放的武将是否在掉落库中（跳过战令武将）
	c.checkOpenedHeroes(heroCols, dropItemCols, seasonPassRewardCols, param.StartRowIdx, now, result)

	// 步骤6b: 检查战令武将掉落
	c.checkSeasonPassHeroes(heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, param.StartRowIdx, now, warnDuration, protectMonths, result)

	// 步骤6c: 检查大将军武将掉落
	c.checkGeneralHeroes(heroCols, dropItemCols, arenaScoreRewardsCols, arenaSeasonCols, param.StartRowIdx, now, warnDuration, protectMonths, result)

	// 步骤7: 设置检查结果并返回
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个武将掉落配置问题", len(result.ErrCells))
	}

	return result
}

// checkOpenedHeroes 检查已开放的武将是否在掉落库中
// 规则：IsOpen=true 且（OpenDate为空或已过）的武将必须加入掉落库
// 注意：跳过战令武将（SeasonPassReward中的武将），由checkSeasonPassHeroes单独处理
//
// 执行流程：
//  1. 查找 Hero 表中所需列的索引位置
//  2. 遍历所有武将数据：
//     a. 读取武将ID，跳过空值或解析失败的行
//     b. 读取武将名称
//     c. 检查 IsOpen 列，未开放则跳过
//     d. 检查 OpenDate 列：为空或已过都认为已开放
//     e. 跳过战令武将（由单独方法处理）
//     f. 获取武将在掉落库中的状态
//     g. 根据状态添加相应的错误信息（未加入或已过期报错，未生效不报错）
//  3. 返回检查结果
func (c *HeroDropCheckRule) checkOpenedHeroes(heroCols, dropItemCols, seasonPassRewardCols [][]string, startRowIdx int, now time.Time, result *json_rule.TableCheckResult) {
	// 步骤1: 查找列索引
	idColIdx := helpers.GetColIndexByName(heroCols, "Id")
	nameColIdx := helpers.GetColIndexByName(heroCols, "Name")
	heroTypeColIdx := helpers.GetColIndexByName(heroCols, "HeroType")
	isOpenColIdx := helpers.GetColIndexByName(heroCols, "IsOpen")
	openDateColIdx := helpers.GetColIndexByName(heroCols, "OpenDate")

	if idColIdx < 0 {
		return
	}

	// 步骤2: 遍历所有武将数据
	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(heroCols, startRowIdx); rowIdx++ {
		// 步骤2a: 读取武将ID，跳过空值或解析失败的行
		idStr := helpers.GetColValue(heroCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}

		heroId, err := helpers.ParseIntWithError(idStr)
		if err != nil {
			continue
		}

		// 步骤2b: 读取武将名称
		heroName := ""
		if nameColIdx >= 0 {
			heroName = helpers.GetColValue(heroCols, nameColIdx, rowIdx)
		}

		// 步骤2c: 检查 HeroType 列（只检查 HeroType=1 的普通武将）
		heroType := 1 // 默认为普通武将
		if heroTypeColIdx >= 0 {
			heroType, _ = helpers.ParseIntWithError(helpers.GetColValue(heroCols, heroTypeColIdx, rowIdx))
		}
		if heroType != 1 {
			continue // HeroType≠1 表示特殊武将，跳过掉落库检查
		}

		// 步骤2d: 检查 IsOpen 列
		isOpen := false
		if isOpenColIdx >= 0 {
			isOpen = helpers.ParseBool(helpers.GetColValue(heroCols, isOpenColIdx, rowIdx))
		}
		if !isOpen {
			continue // 未开放，跳过
		}

		// 步骤2e: 检查 OpenDate 列
		var openDate time.Time
		if openDateColIdx >= 0 {
			openDate = helpers.ParseDate(helpers.GetColValue(heroCols, openDateColIdx, rowIdx))
		}
		// 开放时间未配置或已过都认为已开放
		if !openDate.IsZero() && openDate.After(now) {
			continue // 开放时间在未来，跳过
		}

		// 步骤2f: 跳过战令武将
		if helpers.IsHeroInSeasonPassReward(heroId, seasonPassRewardCols, startRowIdx) {
			continue
		}

		// 准备开放时间字符串用于日志
		openDateStr := "未配置"
		if !openDate.IsZero() {
			openDateStr = helpers.FormatDateTime(openDate)
		}

		// 步骤2f: 获取武将在掉落库中的状态
		// 注意：允许掉落配置在未来生效（提前配置），不检查 HeroDropConfigNotEffective
		status, _, expireDateStr := helpers.GetHeroDropPoolStatus(heroId, dropItemCols, startRowIdx, now)

		// 步骤2g: 根据状态添加相应的错误信息
		switch status {
		case helpers.HeroNotInDropPool:
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("已开放武将【%s】(ID=%d, OpenDate=%s) 未加入掉落库（规则：已开放武将必须加入掉落库）", heroName, heroId, openDateStr),
			})
		case helpers.HeroDropConfigExpired:
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("已开放武将【%s】(ID=%d, OpenDate=%s) 掉落配置已过期（过期时间: %s）（规则：掉落配置必须有效）", heroName, heroId, openDateStr, expireDateStr),
			})
		}
		// HeroInDropPool 和 HeroDropConfigNotEffective 不需要报错（允许提前配置）
	}
}

// checkSeasonPassHeroes 检查战令武将掉落
// 规则：战令结束后3个月需加入掉落库，提前5天提醒；保护期内不应在掉落池中
//
// 执行流程：
//  1. 检查 SeasonPassReward 和 SeasonPass 表数据是否存在
//  2. 查找 HeroType 列索引（只检查 HeroType=1 的普通武将）
//  3. 查找所有战令武将
//  4. 遍历每个战令武将：
//     a. 获取武将名称（从SeasonPassReward或Hero表）
//     b. 检查 HeroType 列，跳过特殊武将（HeroType≠1）
//     c. 计算需要加入掉落库的时间点（战令结束后N个月）
//     d. 获取武将在掉落库中的状态
//     e. 战令已结束：
//     - 超过截止时间且未在掉落库 → 报错
//     - 临近截止时间 → 提前提醒
//     f. 反向检查：ValidDate早于保护期截止时间 → 报错
//  5. 返回检查结果
func (c *HeroDropCheckRule) checkSeasonPassHeroes(heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols [][]string, startRowIdx int, now time.Time, warnDuration time.Duration, protectMonths int, result *json_rule.TableCheckResult) {
	// 步骤1: 检查表数据是否存在
	if seasonPassRewardCols == nil || seasonPassCols == nil {
		return
	}

	// 步骤2: 查找 HeroType 列索引（只检查 HeroType=1 的普通武将）
	heroTypeColIdx := helpers.GetColIndexByName(heroCols, "HeroType")

	// 步骤3: 查找所有战令武将
	seasonPassHeroes := helpers.FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols, startRowIdx)

	// 步骤4: 遍历每个战令武将
	for _, spHero := range seasonPassHeroes {
		if spHero.EndTime.IsZero() {
			continue
		}

		// 步骤4a: 获取武将名称
		heroRow := helpers.FindHeroById(spHero.HeroId, heroCols, startRowIdx)

		heroName := spHero.HeroName

		if heroName == "" && heroRow != nil {

			heroName = heroRow.Name

		}

		// 步骤4b: 检查 HeroType 列（只检查 HeroType=1 的普通武将）
		if heroRow != nil {
			if heroTypeColIdx >= 0 {
				heroType := 1 // 默认为普通武将
				if heroTypeStr := helpers.GetColValue(heroCols, heroTypeColIdx, heroRow.RowIndex); heroTypeStr != "" {
					heroType, _ = helpers.ParseIntWithError(heroTypeStr)
				}
				if heroType != 1 {
					continue // HeroType≠1 表示特殊武将，跳过掉落库检查
				}
			}
		}

		// 步骤4c: 计算保护期截止时间
		protectionDeadline := helpers.CalcSeasonPassProtectionDeadline(spHero.StartTime, protectMonths)

		// 步骤4d: 获取武将在掉落库中的状态
		// 注意：允许掉落配置在未来生效（提前配置），不检查 HeroDropConfigNotEffective
		status, _, _ := helpers.GetHeroDropPoolStatus(spHero.HeroId, dropItemCols, startRowIdx, now)

		// 步骤4e: 战令已结束（正向检查）
		if spHero.EndTime.Before(now) {
			// 已过截止时间
			if protectionDeadline.Before(now) || protectionDeadline.Equal(now) {
				if status == helpers.HeroNotInDropPool {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    heroRow.RowIndex,
						ExcelRow: heroRow.RowIndex + 1,
						Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 战令已于 %s 结束，保护期已过（战令开始时间起%d个月），仍未加入掉落库（规则：保护期结束后必须加入掉落库）",
							heroName, spHero.HeroId, helpers.FormatDateTime(spHero.EndTime), protectMonths),
					})
				}
			} else {
				// 临近截止时间，提前提醒
				timeUntilDeadline := protectionDeadline.Sub(now)
				if timeUntilDeadline <= warnDuration && status == helpers.HeroNotInDropPool {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    heroRow.RowIndex,
						ExcelRow: heroRow.RowIndex + 1,
						Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 需在保护期截止 %s 前加入掉落库（剩余 %.0f 天）（规则：战令开始时间起%d个月保护期）",
							heroName, spHero.HeroId, helpers.FormatDateTime(protectionDeadline), timeUntilDeadline.Hours()/24, protectMonths),
					})
				}
			}
		}

		// 步骤4f: 反向检查 — 战令武将保护期内不应在掉落池中
		// 如果掉落ValidDate早于保护期截止时间，说明武将在保护期内就可以被掉落获取
		validDate := getHeroDropValidDate(spHero.HeroId, dropItemCols, startRowIdx)
		if !validDate.IsZero() && validDate.Before(protectionDeadline) {
			errIdx := spHero.RowIndex
			if heroRow != nil {
				errIdx = heroRow.RowIndex
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    errIdx,
				ExcelRow: errIdx + 1,
				Reason: fmt.Sprintf("战令武将【%s】(ID=%d) 的掉落ValidDate(%s)早于保护期截止时间(%s)（规则：战令武将保护期内不应在掉落池中）",
					heroName, spHero.HeroId,
					helpers.FormatDateTime(validDate), helpers.FormatDateTime(protectionDeadline)),
			})
		}
	}
}

// checkGeneralHeroes 检查大将军武将掉落
// 规则：赛季结束后需加入掉落库，提前5天提醒；赛季结束前不应在掉落池中
//
// 执行流程：
//  1. 检查 ArenaScoreRewards 和 ArenaSeason 表数据是否存在
//  2. 查找 HeroType 列索引（只检查 HeroType=1 的普通武将）
//  3. 查找所有大将军武将
//  4. 遍历每个大将军武将：
//     a. 获取赛季时间
//     b. 获取武将名称（从ArenaScoreRewards或Hero表）
//     c. 检查 HeroType 列，跳过特殊武将（HeroType≠1）
//     d. 获取武将在掉落库中的状态
//     e. 赛季已结束且未在掉落库 → 报错
//     f. 赛季未结束但临近结束 → 提前提醒
//     g. 反向检查：ValidDate早于赛季结束时间 → 报错
//  5. 返回检查结果
func (c *HeroDropCheckRule) checkGeneralHeroes(heroCols, dropItemCols, arenaScoreRewardsCols, arenaSeasonCols [][]string, startRowIdx int, now time.Time, warnDuration time.Duration, protectMonths int, result *json_rule.TableCheckResult) {
	// 步骤1: 检查表数据是否存在
	if arenaScoreRewardsCols == nil || arenaSeasonCols == nil {
		return
	}

	// 步骤2: 查找 HeroType 列索引（只检查 HeroType=1 的普通武将）
	heroTypeColIdx := helpers.GetColIndexByName(heroCols, "HeroType")

	// 步骤3: 查找所有大将军武将
	generalHeroes := helpers.FindArenaGeneralHeroes(arenaScoreRewardsCols, startRowIdx)

	// 步骤4: 遍历每个大将军武将
	for _, genHero := range generalHeroes {
		// 步骤4a: 获取赛季时间
		seasonStartTime, seasonEndTime := helpers.GetArenaSeasonTime(genHero.SeasonId, arenaSeasonCols, startRowIdx)

		if seasonEndTime.IsZero() {
			continue
		}

		// 步骤4b: 获取武将名称
		heroRow := helpers.FindHeroById(genHero.HeroId, heroCols, startRowIdx)

		heroName := genHero.HeroName

		if heroName == "" && heroRow != nil {

			heroName = heroRow.Name

		}

		// 步骤4c: 检查 HeroType 列（只检查 HeroType=1 的普通武将）
		if heroRow != nil {
			if heroTypeColIdx >= 0 {
				heroType := 1 // 默认为普通武将
				if heroTypeStr := helpers.GetColValue(heroCols, heroTypeColIdx, heroRow.RowIndex); heroTypeStr != "" {
					heroType, _ = helpers.ParseIntWithError(heroTypeStr)
				}
				if heroType != 1 {
					continue // HeroType≠1 表示特殊武将，跳过掉落库检查
				}
			}
		}

		// 步骤4d: 获取武将在掉落库中的状态
		// 注意：允许掉落配置在未来生效（提前配置），不检查 HeroDropConfigNotEffective
		status, _, _ := helpers.GetHeroDropPoolStatus(genHero.HeroId, dropItemCols, startRowIdx, now)

		// 步骤4e: 赛季已结束（正向检查）
		if seasonEndTime.Before(now) {
			if status == helpers.HeroNotInDropPool {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    heroRow.RowIndex,
					ExcelRow: heroRow.RowIndex + 1,
					Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 赛季已于 %s 结束，仍未加入掉落库（规则：赛季结束后必须加入掉落库）",
						heroName, genHero.HeroId, helpers.FormatDateTime(seasonEndTime)),
				})
			}
		} else {
			// 步骤4f: 赛季未结束，提前提醒
			timeUntilEnd := seasonEndTime.Sub(now)
			if timeUntilEnd <= warnDuration && status == helpers.HeroNotInDropPool {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    heroRow.RowIndex,
					ExcelRow: heroRow.RowIndex + 1,
					Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 赛季将于 %s 结束，请提前准备加入掉落库（剩余 %.0f 天）（规则：赛季结束后必须加入掉落库）",
						heroName, genHero.HeroId, helpers.FormatDateTime(seasonEndTime), timeUntilEnd.Hours()/24),
				})
			}
		}

		// 步骤4g: 反向检查 — 大将军武将赛季结束前不应在掉落池中
		// 如果掉落ValidDate早于赛季结束时间，说明武将在赛季结束前就可以被掉落获取
		validDate := getHeroDropValidDate(genHero.HeroId, dropItemCols, startRowIdx)
		if !validDate.IsZero() && validDate.Before(seasonEndTime) {
			errIdx := genHero.RowIndex
			if heroRow != nil {
				errIdx = heroRow.RowIndex
			}
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    errIdx,
				ExcelRow: errIdx + 1,
				Reason: fmt.Sprintf("大将军武将【%s】(ID=%d) 的掉落ValidDate(%s)早于保护期截止时间(%s)（规则：大将军武将保护期内不应在掉落池中）",
					heroName, genHero.HeroId,
					helpers.FormatDateTime(validDate), helpers.FormatDateTime(seasonEndTime)),
			})
		}

		_ = seasonStartTime // 避免未使用警告
	}
}

// loadRelatedSheets 加载相关的表数据
//
// 执行流程：
// 1. 从 sheetMap 中获取 Hero 表数据
// 2. 从 sheetMap 中获取 DropItem 表数据
// 3. 从 sheetMap 中获取 SeasonPassReward 表数据
// 4. 从 sheetMap 中获取 SeasonPass 表数据
// 5. 从 sheetMap 中获取 ArenaScoreRewards 表数据（注意Sheet名称是ArenaScoreReward）
// 6. 从 sheetMap 中获取 ArenaSeason 表数据
// 7. 返回所有加载的表数据
//
// 注意：使用 FindSheetBySuffix 方法支持后缀匹配（如"武将|Hero"可通过"Hero"找到）
func (c *HeroDropCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols [][]string, err error) {
	// 步骤1: Hero 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Hero"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			heroCols = cols
		}
	}

	// 步骤2: DropItem 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "DropItem"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			dropItemCols = cols
		}
	}

	// 步骤3: SeasonPassReward 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPassReward"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			seasonPassRewardCols = cols
		}
	}

	// 步骤4: SeasonPass 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "SeasonPass"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			seasonPassCols = cols
		}
	}

	// 步骤5: ArenaScoreRewards 表（注意：Sheet名称是 ArenaScoreReward，不是 ArenaScoreRewards）
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaScoreReward"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			arenaScoreRewardsCols = cols
		}
	}

	// 步骤6: ArenaSeason 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "ArenaSeason"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			arenaSeasonCols = cols
		}
	}

	// 步骤7: 返回所有加载的表数据
	return heroCols, dropItemCols, seasonPassRewardCols, seasonPassCols, arenaScoreRewardsCols, arenaSeasonCols, nil
}

// getHeroDropValidDate 从 DropItem 表获取指定武将道具的 ValidDate
// 遍历 DropItem 的 Item 列，找到包含指定武将道具的行，返回其 ValidDate
// 注意：此函数被同包的 item_synthesis.go 复用（武将合成检查也需要 DropItem.ValidDate）
func getHeroDropValidDate(heroId int, dropItemCols [][]string, startRowIdx int) time.Time {
	heroItemId := helpers.MakeHeroItemId(heroId)
	itemColIdx := helpers.GetColIndexByName(dropItemCols, "Item")
	validDateColIdx := helpers.GetColIndexByName(dropItemCols, "ValidDate")

	if itemColIdx < 0 {
		return time.Time{}
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(dropItemCols, startRowIdx); rowIdx++ {
		itemStr := helpers.GetColValue(dropItemCols, itemColIdx, rowIdx)
		if itemStr == "" {
			continue
		}
		items := helpers.ParseItemCfg(itemStr)
		for _, item := range items {
			if item.ItemId == heroItemId {
				if validDateColIdx < 0 {
					return time.Time{}
				}
				return helpers.ParseDate(helpers.GetColValue(dropItemCols, validDateColIdx, rowIdx))
			}
		}
	}
	return time.Time{}
}
