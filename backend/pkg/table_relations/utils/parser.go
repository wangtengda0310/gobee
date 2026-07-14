package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// ParseDrawSkinData 解析皮肤抽奖表数据
//
// 从 Excel 表格中解析出皮肤抽奖配置，返回 DrawSkinDiff 切片
//
// 参数：
//   - sheetMap: 表映射表，key为表名，value为Excel文件对象
//
// 返回：
//   - *[]draw_skin.DrawSkinDiff: 皮肤抽奖配置数据
//   - error: 错误信息
//
// 示例：
//
//	drawSkins, err := ParseDrawSkinData(sheetMap)
//	if err != nil {
//	    return fmt.Errorf("解析皮肤抽奖表失败: %w", err)
//	}
//	for _, skin := range *drawSkins {
//	    fmt.Printf("抽奖ID: %d, 名称: %s\n", skin.Id, skin.Name)
//	}
func ParseDrawSkinData(sheetMap map[string]*excelize.File) (*[]draw_skin.DrawSkinDiff, error) {
	// 加载表数据
	cols, _, err := LoadSheetData(sheetMap, "DrawSkin")
	if err != nil {
		return nil, err
	}

	// 获取数据起始和结束行
	startRow := GetStartRowIndex()
	idColIdx := helpers.GetColIndexByName(cols, "Id")
	if idColIdx < 0 {
		return nil, fmt.Errorf("列 'Id' 不存在")
	}
	endRow := GetDataEndIndex(cols, idColIdx, startRow)

	// 构建皮肤抽奖配置
	drawSkins := make([]draw_skin.DrawSkinDiff, 0, 50)
	reg := regexp.MustCompile(`\{(\d+);\d+}`)

	for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
		idStr := helpers.GetColValue(cols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		skin := draw_skin.DrawSkinDiff{
			Id:        id,
			Name:      utils.GetCellValue(cols, draw_skin.Name, rowIdx),
			StartTime: utils.GetCellValue(cols, draw_skin.StartTime, rowIdx),
			EndTime:   utils.GetCellValue(cols, draw_skin.EndTime, rowIdx),
		}

		// 解析单抽掉落规则ID
		if onceDropRule, err := strconv.Atoi(utils.GetCellValue(cols, draw_skin.OnceDropRule, rowIdx)); err == nil {
			skin.OnceDropRule = onceDropRule
		} else {
			skin.OnceDropRule = -1
		}

		// 解析十抽掉落规则ID
		if tenDropRule, err := strconv.Atoi(utils.GetCellValue(cols, draw_skin.TenDropRule, rowIdx)); err == nil {
			skin.TenDropRule = tenDropRule
		} else {
			skin.TenDropRule = -1
		}

		// 解析活动ID
		if activityId, err := strconv.Atoi(utils.GetCellValue(cols, draw_skin.ActivityId, rowIdx)); err == nil {
			skin.ActivityId = activityId
		} else {
			skin.ActivityId = -1
		}

		// 解析大奖保底次数
		if bigAwardCount, err := strconv.Atoi(utils.GetCellValue(cols, draw_skin.BigAwardCount, rowIdx)); err == nil {
			skin.BigAwardCount = bigAwardCount
		} else {
			skin.BigAwardCount = -1
		}

		// 解析大奖道具ID
		if bigAwardItemId, err := strconv.Atoi(utils.GetCellValue(cols, draw_skin.BigAwardItemId, rowIdx)); err == nil {
			skin.BigAwardItemId = bigAwardItemId
		} else {
			skin.BigAwardItemId = -1
		}

		// 解析道具消耗
		onceItemCostStr := utils.GetCellValue(cols, draw_skin.OnceItemCost, rowIdx)
		skin.OnceItemCost = parseItemCost(onceItemCostStr, reg)

		tenItemCostStr := utils.GetCellValue(cols, draw_skin.TenItemCost, rowIdx)
		skin.TenItemCost = parseItemCost(tenItemCostStr, reg)

		drawSkins = append(drawSkins, skin)
	}

	return &drawSkins, nil
}

// parseItemCost 解析道具消耗字符串
//
// 格式为 {itemId;count}，多个用逗号分隔
//
// 参数：
//   - itemCostStr: 道具消耗字符串
//   - reg: 正则表达式对象（用于匹配 {数字;数字} 格式）
//
// 返回：
//   - []draw_skin.ItemCfg: 道具配置列表
func parseItemCost(itemCostStr string, reg *regexp.Regexp) []draw_skin.ItemCfg {
	itemCost := make([]draw_skin.ItemCfg, 0, 5)
	if itemCostStr == "" {
		return itemCost
	}

	// 使用正则找出所有匹配项
	matches := reg.FindAllStringSubmatch(itemCostStr, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			fullMatch := match[0]
			itemIdStr := match[1]

			// 解析数量
			count := 1
			parts := strings.Split(strings.Trim(fullMatch, "{}"), ";")
			if len(parts) >= 2 {
				if c, err := strconv.Atoi(parts[1]); err == nil {
					count = c
				}
			}

			itemId, _ := strconv.Atoi(itemIdStr)
			itemCost = append(itemCost, draw_skin.ItemCfg{
				ItemId: itemId,
				Count:  count,
			})
		}
	}

	return itemCost
}

// ParseDropRuleData 解析掉落规则表数据
//
// 从 Excel 表格中解析出掉落规则配置，返回 DropRuleDiff 切片
//
// 参数：
//   - sheetMap: 表映射表，key为表名，value为Excel文件对象
//
// 返回：
//   - *[]drop_rule.DropRuleDiff: 掉落规则配置数据
//   - error: 错误信息
//
// 示例：
//
//	dropRules, err := ParseDropRuleData(sheetMap)
//	if err != nil {
//	    return fmt.Errorf("解析掉落规则表失败: %w", err)
//	}
//	for _, rule := range *dropRules {
//	    fmt.Printf("掉落ID: %d, 名称: %s\n", rule.Id, rule.Name)
//	}
func ParseDropRuleData(sheetMap map[string]*excelize.File) (*[]drop_rule.DropRuleDiff, error) {
	// 加载表数据
	cols, _, err := LoadSheetData(sheetMap, "DropRule")
	if err != nil {
		return nil, err
	}

	// 获取数据起始和结束行
	startRow := GetStartRowIndex()
	idColIdx := helpers.GetColIndexByName(cols, "Id")
	if idColIdx < 0 {
		return nil, fmt.Errorf("列 'Id' 不存在")
	}
	endRow := GetDataEndIndex(cols, idColIdx, startRow)

	// 构建掉落规则配置
	dropRules := make([]drop_rule.DropRuleDiff, 0, 300)

	for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
		idStr := helpers.GetColValue(cols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		rule := drop_rule.DropRuleDiff{
			Id:   id,
			Name: utils.GetCellValue(cols, drop_rule.Name, rowIdx),
		}

		// 解析掉落次数
		if count, err := strconv.Atoi(utils.GetCellValue(cols, drop_rule.Count, rowIdx)); err == nil {
			rule.Count = count
		} else {
			rule.Count = -1
		}

		// 解析掉落组ID
		rule.DropGroup = parseIntArray(utils.GetCellValue(cols, drop_rule.DropGroup, rowIdx))

		// 解析小保底次数
		if ensureSmall, err := strconv.Atoi(utils.GetCellValue(cols, drop_rule.EnsureSmall, rowIdx)); err == nil {
			rule.EnsureSmall = ensureSmall
		} else {
			rule.EnsureSmall = -1
		}

		// 解析小保底组ID
		rule.EnsureSmallGroup = parseIntArray(utils.GetCellValue(cols, drop_rule.EnsureSmallGroup, rowIdx))

		// 解析大保底次数
		if ensureBig, err := strconv.Atoi(utils.GetCellValue(cols, drop_rule.EnsureBig, rowIdx)); err == nil {
			rule.EnsureBig = ensureBig
		} else {
			rule.EnsureBig = -1
		}

		// 解析大保底组ID
		rule.EnsureBigGroup = parseIntArray(utils.GetCellValue(cols, drop_rule.EnsureBigGroup, rowIdx))

		// 解析道具保底次数
		if ensureItemCount, err := strconv.Atoi(utils.GetCellValue(cols, drop_rule.EnsureItemCount, rowIdx)); err == nil {
			rule.EnsureItemCount = ensureItemCount
		} else {
			rule.EnsureItemCount = -1
		}

		// 解析保底道具ID
		if ensureItemID, err := strconv.Atoi(utils.GetCellValue(cols, drop_rule.EnsureItemID, rowIdx)); err == nil {
			rule.EnsureItemID = ensureItemID
		} else {
			rule.EnsureItemID = -1
		}

		// 解析道具保底排重
		if itemCheckExist, err := strconv.ParseBool(utils.GetCellValue(cols, drop_rule.ItemCheckExist, rowIdx)); err == nil {
			rule.ItemCheckExist = itemCheckExist
		} else {
			rule.ItemCheckExist = false
		}

		dropRules = append(dropRules, rule)
	}

	return &dropRules, nil
}

// parseIntArray 解析逗号分隔的整数字符串
//
// 参数：
//   - str: 逗号分隔的整数字符串（如 "1,2,3"）
//
// 返回：
//   - []int: 整数数组
func parseIntArray(str string) []int {
	result := make([]int, 0, 10)
	if str == "" {
		return result
	}

	parts := strings.Split(str, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if num, err := strconv.Atoi(part); err == nil {
			result = append(result, num)
		}
	}

	return result
}
