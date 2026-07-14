// Package activity 提供活动相关的跨表校验规则
//
// 本文件实现 SEASON_PASS_REWARD_INTEGRITY_CHECK 规则：
// 检查战令高级奖励(level=51/61/71/81/91)的道具经两次Item跳转后的武将ID，
// 是否在SeasonPass.HeroId对应的武将集合中。
//
// ## 数据流
//
// 左链（SeasonPassReward → Item → Item → Hero）：
//
//	SeasonPassReward.HighReward → ParseItemCfg → 道具ID列表
//	Item表 Id=道具ID → ItemParam → 包装道具ID
//	Item表 Id=包装道具ID → ItemParam → 武将ID
//
// 右链（SeasonPass → Hero）：
//
//	SeasonPass.Id=SeasonPassId → HeroId
//	Hero表验证武将存在
//
// ## 关键方法
//
//   - Meta        → 规则元数据
//   - Check       → 执行检查主入口
//   - loadRelatedSheets → 加载 Item/SeasonPass/Hero 关联表
//   - findItemParamById → Item 表按 Id 查 ItemParam
//   - getSeasonPassStartTime → SeasonPass 表查 StartTime
package activity

import (
	"fmt"
	"strconv"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// targetLevels 需要检查的战令高级奖励 level 集合
var targetLevels = map[int]bool{51: true, 61: true, 71: true, 81: true, 91: true}

// SeasonPassRewardIntegrityCheckRule 战令高级奖励道具引用完整性检查规则
//
// 校验规则：
// 1. 从 SeasonPassReward 表过滤 level ∈ {51,61,71,81,91} 的行
// 2. 左链：HighReward → ParseItemCfg → Item表跳转1 → ItemParam(包装道具ID) → Item表跳转2 → ItemParam(武将ID)
// 3. 右链：SeasonPass.HeroId → Hero表验证存在
// 4. 时间门控：SeasonPass.StartTime 距今不足 warnDays 天才执行校验
// 5. 比较规则：左链武将ID集合与右链武将ID集合无交集且两边都有值 → 报错
type SeasonPassRewardIntegrityCheckRule struct{}

// Meta 返回规则元数据
func (c *SeasonPassRewardIntegrityCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.SEASON_PASS_REWARD_INTEGRITY_CHECK,
		DisplayName:    "战令高级奖励道具引用完整性检查",
		Description:    "检查战令高级奖励(level=51/61/71/81/91)的道具经两次Item跳转后的武将ID，是否在SeasonPass.HeroId对应的武将集合中",
		TargetSheets:   []string{"SeasonPassReward"},
		RequiredSheets: []string{"Item", "SeasonPass", "Hero"},
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.WARN_DAYS_BEFORE,
				Label:       "提前警告天数",
				Description: "赛季开始前多少天才执行校验并报错（默认7天）",
				Type:        "number",
				Default:     "7",
				Required:    false,
			},
		},
	}
}

// Check 执行战令高级奖励道具引用完整性检查
//
// 执行流程：
//  1. 解析 warnDays 参数（默认7天）
//  2. 加载 Item、SeasonPass、Hero 关联表
//  3. 构建 Hero 表的武将ID集合（用于右链验证）
//  4. 遍历 SeasonPassReward 表，过滤目标 level 行
//  5. 对每行执行左链和右链解析
//  6. 时间门控检查
//  7. 比较左右链武将ID集合是否有交集
//  8. 汇总错误并返回结果
func (c *SeasonPassRewardIntegrityCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "战令高级奖励道具引用完整性检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 步骤1: 解析 warnDays 参数
	warnDays := 7
	if val, ok := param.Params[string(json_rule.WARN_DAYS_BEFORE)]; ok && val != "" {
		if days, err := helpers.ParseIntWithError(val); err == nil && days > 0 {
			warnDays = days
		}
	}

	// 步骤2: 加载关联表
	itemCols, seasonPassCols, heroCols := c.loadRelatedSheets(param.SheetMap)

	if itemCols == nil {
		result.Ok = false
		result.Reason = "未找到 Item 表数据"
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

	// 步骤3: 构建 Hero 表武将ID集合（右链验证用）
	heroIdSet := buildHeroIdSet(heroCols, param.StartRowIdx)

	// 步骤4: 获取 SeasonPassReward 表的列索引
	// level 列名可能为 "level" 或 "Level"，优先小写（JSON 配置中为小写）
	cols := param.Cols
	levelColIdx := helpers.GetColIndexByName(cols, "level")
	if levelColIdx < 0 {
		levelColIdx = helpers.GetColIndexByName(cols, "Level")
	}
	highRewardColIdx := helpers.GetColIndexByName(cols, "HighReward")
	seasonPassIdColIdx := helpers.GetColIndexByName(cols, "SeasonPassId")

	if highRewardColIdx < 0 {
		result.Ok = false
		result.Reason = "SeasonPassReward 表未找到 HighReward 列"
		return result
	}

	// 步骤5: 解析当前时间
	now := helpers.ResolveNow(param.Now)

	// 步骤6: 遍历 SeasonPassReward 数据行
	dataEnd := helpers.GetDataEndIndex(cols, param.StartRowIdx)
	for rowIdx := param.StartRowIdx; rowIdx < dataEnd; rowIdx++ {
		// 过滤目标 level
		if levelColIdx < 0 {
			continue
		}
		levelStr := helpers.GetColValue(cols, levelColIdx, rowIdx)
		level, err := strconv.Atoi(levelStr)
		if err != nil || !targetLevels[level] {
			continue
		}

		// 读取 HighReward 和 SeasonPassId
		highReward := helpers.GetColValue(cols, highRewardColIdx, rowIdx)
		if highReward == "" {
			continue
		}

		seasonPassId := 0
		if seasonPassIdColIdx >= 0 {
			seasonPassId, _ = strconv.Atoi(helpers.GetColValue(cols, seasonPassIdColIdx, rowIdx))
		}
		if seasonPassId == 0 {
			continue
		}

		// 步骤6a: 时间门控 — 获取 SeasonPass.StartTime
		startTime := getSeasonPassStartTime(seasonPassId, seasonPassCols, param.StartRowIdx)
		if startTime.IsZero() {
			continue
		}
		if now.Before(startTime.AddDate(0, 0, -warnDays)) {
			continue
		}

		// 步骤6b: 左链解析 — HighReward → Item → Item → 武将ID集合
		leftHeroIds, leftErrors := c.resolveLeftChain(highReward, itemCols, param.StartRowIdx, seasonPassId, startTime)
		for _, errCell := range leftErrors {
			errCell.Index = rowIdx
			errCell.ExcelRow = rowIdx + 1
			result.ErrCells = append(result.ErrCells, errCell)
		}

		// 步骤6c: 右链解析 — SeasonPass.HeroId → Hero 表验证
		rightHeroIds := getSeasonPassHeroIds(seasonPassId, seasonPassCols, param.StartRowIdx, heroIdSet)

		// 步骤6d: 比较左右链 — 两边都有值但无交集则报错
		if len(leftHeroIds) > 0 && len(rightHeroIds) > 0 {
			hasIntersection := false
			for id := range leftHeroIds {
				if rightHeroIds[id] {
					hasIntersection = true
					break
				}
			}
			if !hasIntersection {
				leftIds := mapKeysSorted(leftHeroIds)
				rightIds := mapKeysSorted(rightHeroIds)
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason: fmt.Sprintf("赛季%d StartTime=%s: 左链武将ID=%v(HighReward→Item→Item→Hero), 右链武将ID=%v(SeasonPass.HeroId→Hero), 无交集",
						seasonPassId, helpers.FormatDateTime(startTime), leftIds, rightIds),
				})
			}
		}
	}

	// 步骤7: 汇总结果
	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个战令高级奖励道具引用完整性问题", len(result.ErrCells))
	}

	return result
}

// resolveLeftChain 解析左链：HighReward → Item表跳转1 → 包装道具ID → Item表跳转2 → 武将ID集合
//
// 执行流程：
//  1. ParseItemCfg 解析 HighReward 得到道具ID列表
//  2. 对每个道具ID，在 Item 表查 ItemParam 得包装道具ID
//  3. 对包装道具ID，在 Item 表查 ItemParam 得武将ID
//  4. 返回武将ID集合和链路断裂错误
func (c *SeasonPassRewardIntegrityCheckRule) resolveLeftChain(highReward string, itemCols [][]string, startRowIdx int, seasonPassId int, startTime time.Time) (heroIds map[int]bool, errors []*json_rule.CellError) {
	heroIds = make(map[int]bool)

	// 步骤1: 解析 HighReward 中的道具列表
	items := helpers.ParseItemCfg(highReward)
	if len(items) == 0 {
		return heroIds, nil
	}

	for _, item := range items {
		// 步骤2: 第一次 Item 表跳转 — 道具ID → 包装道具ID
		itemParam1 := findItemParamById(item.ItemId, itemCols, startRowIdx)
		if itemParam1 == "" {
			errors = append(errors, &json_rule.CellError{
				Reason: fmt.Sprintf("赛季%d StartTime=%s: 道具ID=%d(HighReward) → Item表跳转1 → 未找到ItemParam",
					seasonPassId, helpers.FormatDateTime(startTime), item.ItemId),
			})
			continue
		}

		// ItemParam 可能是逗号分隔的多个值，取第一个有效数字
		wrapperItemId, err := parseFirstInt(itemParam1)
		if err != nil {
			errors = append(errors, &json_rule.CellError{
				Reason: fmt.Sprintf("赛季%d StartTime=%s: 道具ID=%d → ItemParam=%s 解析包装道具ID失败",
					seasonPassId, helpers.FormatDateTime(startTime), item.ItemId, itemParam1),
			})
			continue
		}
		if wrapperItemId == 0 {
			continue
		}

		// 步骤3: 第二次 Item 表跳转 — 包装道具ID → 武将ID
		itemParam2 := findItemParamById(wrapperItemId, itemCols, startRowIdx)
		if itemParam2 == "" {
			errors = append(errors, &json_rule.CellError{
				Reason: fmt.Sprintf("赛季%d StartTime=%s: 包装道具ID=%d → Item表跳转2 → 未找到ItemParam",
					seasonPassId, helpers.FormatDateTime(startTime), wrapperItemId),
			})
			continue
		}

		heroId, err := parseFirstInt(itemParam2)
		if err != nil {
			errors = append(errors, &json_rule.CellError{
				Reason: fmt.Sprintf("赛季%d StartTime=%s: 包装道具ID=%d → ItemParam=%s 解析武将ID失败",
					seasonPassId, helpers.FormatDateTime(startTime), wrapperItemId, itemParam2),
			})
			continue
		}
		if heroId > 0 {
			heroIds[heroId] = true
		}
	}

	return heroIds, errors
}

// loadRelatedSheets 加载 Item、SeasonPass、Hero 关联表
func (c *SeasonPassRewardIntegrityCheckRule) loadRelatedSheets(sheetMap map[string]*excelize.File) (itemCols, seasonPassCols, heroCols [][]string) {
	// Item 表
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, "Item"); ok {
		cols, e := file.GetCols(sheetName)
		if e == nil {
			itemCols = cols
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

	return itemCols, seasonPassCols, heroCols
}

// findItemParamById 在 Item 表中按 Id 查找 ItemParam 值
// 遍历 Item 表所有数据行，找到 Id 匹配的行后返回其 ItemParam 列的值
func findItemParamById(targetId int, itemCols [][]string, startRowIdx int) string {
	idColIdx := helpers.GetColIndexByName(itemCols, "Id")
	itemParamColIdx := helpers.GetColIndexByName(itemCols, "ItemParam")

	if idColIdx < 0 || itemParamColIdx < 0 {
		return ""
	}

	dataEnd := helpers.GetDataEndIndex(itemCols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < dataEnd; rowIdx++ {
		idStr := helpers.GetColValue(itemCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id == targetId {
			return helpers.GetColValue(itemCols, itemParamColIdx, rowIdx)
		}
	}
	return ""
}

// getSeasonPassStartTime 从 SeasonPass 表查询指定赛季的 StartTime
//
// 执行流程：
//  1. 查找 Id 列和 StartTime 列索引
//  2. 遍历 SeasonPass 数据行，匹配 seasonPassId
//  3. 用 helpers.ParseDate 解析 StartTime
func getSeasonPassStartTime(seasonPassId int, seasonPassCols [][]string, startRowIdx int) time.Time {
	idColIdx := helpers.GetColIndexByName(seasonPassCols, "Id")
	startTimeColIdx := helpers.GetColIndexByName(seasonPassCols, "StartTime")

	if idColIdx < 0 {
		return time.Time{}
	}

	dataEnd := helpers.GetDataEndIndex(seasonPassCols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < dataEnd; rowIdx++ {
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
				return helpers.ParseDate(helpers.GetColValue(seasonPassCols, startTimeColIdx, rowIdx))
			}
			return time.Time{}
		}
	}
	return time.Time{}
}

// getSeasonPassHeroIds 从 SeasonPass 表获取指定赛季的 HeroId，并验证其在 Hero 表中存在
//
// 执行流程：
//  1. 在 SeasonPass 表中匹配 seasonPassId
//  2. 解析 HeroId 字段（可能是逗号分隔的多个ID）
//  3. 过滤掉 Hero 表中不存在的ID
func getSeasonPassHeroIds(seasonPassId int, seasonPassCols [][]string, startRowIdx int, heroIdSet map[int]bool) map[int]bool {
	result := make(map[int]bool)

	idColIdx := helpers.GetColIndexByName(seasonPassCols, "Id")
	heroIdColIdx := helpers.GetColIndexByName(seasonPassCols, "HeroId")

	if idColIdx < 0 || heroIdColIdx < 0 {
		return result
	}

	dataEnd := helpers.GetDataEndIndex(seasonPassCols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < dataEnd; rowIdx++ {
		idStr := helpers.GetColValue(seasonPassCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id == seasonPassId {
			heroIdStr := helpers.GetColValue(seasonPassCols, heroIdColIdx, rowIdx)
			if heroIdStr == "" {
				return result
			}
			// HeroId 可能是单个数字或逗号分隔
			heroId, err := strconv.Atoi(heroIdStr)
			if err == nil && heroId > 0 && heroIdSet[heroId] {
				result[heroId] = true
			}
			return result
		}
	}
	return result
}

// buildHeroIdSet 构建 Hero 表中所有武将ID的集合
// 遍历 Hero 表的 Id 列，收集所有有效武将ID
func buildHeroIdSet(heroCols [][]string, startRowIdx int) map[int]bool {
	result := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(heroCols, "Id")
	if idColIdx < 0 {
		return result
	}

	dataEnd := helpers.GetDataEndIndex(heroCols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < dataEnd; rowIdx++ {
		idStr := helpers.GetColValue(heroCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id > 0 {
			result[id] = true
		}
	}
	return result
}

// parseFirstInt 从字符串中解析第一个有效整数
// ItemParam 可能包含逗号分隔的多个值，只取第一个有效数字
func parseFirstInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("空字符串")
	}
	return strconv.Atoi(s)
}

// mapKeysSorted 提取 map 的 key 并排序返回，用于错误信息中的稳定输出
func mapKeysSorted(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
