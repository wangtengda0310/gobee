package drop

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DropItemValidityCheckRule DropItem条件和互斥检查
//
// ## 校验规则
// 1. ReplaceGroup = 0 或必须存在于 DropGroup 表的 Id 列
// 2. CheckExist = true 时，ReplaceGroup 必须 > 0
// 3. MustHave 和 ExcludeExist 不能同时为 true
// 4. Item 中每个 ItemId 必须 > 0 且存在于 Item 表的 Id 列
// 5. Item 中每个 ItemNum 必须 > 0
//
// ## 相关表结构
// - DropItem: Id, Name, ReplaceGroup, CheckExist, ExcludeExist, MustHave, Item
// - DropGroup: Id（掉落分组表）
// - Item: Id（道具表）
//
// ## 检查流程
// 1. 加载 DropGroup 和 Item 表，构建有效 ID 集合
// 2. 查找 DropItem 表中的关键列
// 3. 遍历数据行，逐行检查五条规则
// 4. 汇总错误并返回
//
// ## 已移除的检查
//   - ~~CheckExist 和 ExcludeExist 不能同时为 true~~：两者作用时机不同，
//     ExcludeExist 在候选构建阶段排除(roll.go:323-328)，CheckExist 在抽奖后递归重抽(roll.go:367-372)，
//     不是逻辑互斥。同时设置时 ExcludeExist 优先级更高，CheckExist 不会生效，属于配置冗余而非错误。
type DropItemValidityCheckRule struct{}

// itemCfgRegex 匹配 {道具ID;数量} 格式
var itemCfgRegex = regexp.MustCompile(`\{(\d+);(\d+)\}`)

// Meta 返回规则元数据
func (c *DropItemValidityCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.DROP_ITEM_VALIDITY_CHECK,
		DisplayName:    "DropItem条件和互斥检查",
		Description:    "检查DropItem表中ReplaceGroup条件引用、布尔互斥、Item道具有效性",
		TargetSheets:   []string{"DropItem"},
		RequiredSheets: []string{"DropGroup", "Item"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行DropItem条件和互斥检查
//
// 执行流程：
// 1. 加载 DropGroup 和 Item 表，构建有效 ID 集合
// 2. 查找 DropItem 表中的关键列
// 3. 遍历数据行，逐行检查六条规则
// 4. 汇总错误并返回
func (c *DropItemValidityCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DropItem条件和互斥检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 加载 DropGroup 表，构建有效掉落组 ID 集合
	var validGroupIds map[int]bool
	dropGroupCols := c.loadSheet(param.SheetMap, "DropGroup")
	if dropGroupCols != nil {
		validGroupIds = c.buildIdSet(dropGroupCols, param.StartRowIdx)
	}

	// 加载 Item 表，构建有效道具 ID 集合
	var validItemIds map[int]bool
	itemCols := c.loadSheet(param.SheetMap, "Item")
	if itemCols != nil {
		validItemIds = c.buildIdSet(itemCols, param.StartRowIdx)
	}

	// 查找 DropItem 表关键列
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	replaceGroupColIdx := helpers.GetColIndexByName(param.Cols, "ReplaceGroup")
	checkExistColIdx := helpers.GetColIndexByName(param.Cols, "CheckExist")
	excludeExistColIdx := helpers.GetColIndexByName(param.Cols, "ExcludeExist")
	mustHaveColIdx := helpers.GetColIndexByName(param.Cols, "MustHave")
	itemColIdx := helpers.GetColIndexByName(param.Cols, "Item")

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		rowName := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		var errors []string

		// 获取布尔列的值
		checkExist := c.parseBool(helpers.GetColValue(param.Cols, checkExistColIdx, rowIdx))
		excludeExist := c.parseBool(helpers.GetColValue(param.Cols, excludeExistColIdx, rowIdx))
		mustHave := c.parseBool(helpers.GetColValue(param.Cols, mustHaveColIdx, rowIdx))

		// 获取 ReplaceGroup 的值
		replaceGroup := 0
		if replaceGroupColIdx >= 0 {
			val := strings.TrimSpace(helpers.GetColValue(param.Cols, replaceGroupColIdx, rowIdx))
			if val != "" {
				if n, err := strconv.Atoi(val); err == nil {
					replaceGroup = n
				}
			}
		}

		// 规则1: ReplaceGroup = 0 或必须存在于 DropGroup 表
		if replaceGroup != 0 && validGroupIds != nil {
			if !validGroupIds[replaceGroup] {
				errors = append(errors, fmt.Sprintf("ReplaceGroup=%d 在掉落分组表中不存在", replaceGroup))
			}
		}

		// 规则2: CheckExist = true 时，ReplaceGroup 必须 > 0
		if checkExist && replaceGroup <= 0 {
			errors = append(errors, "CheckExist=true 时，ReplaceGroup 必须 > 0")
		}

		// 规则3: MustHave 和 ExcludeExist 不能同时为 true
		if mustHave && excludeExist {
			errors = append(errors, "MustHave 和 ExcludeExist 不能同时为 true")
		}

		// 规则4-5: Item 字段检查
		if itemColIdx >= 0 {
			itemVal := helpers.GetColValue(param.Cols, itemColIdx, rowIdx)
			if itemVal != "" {
				items := c.parseItemCfg(itemVal)
				for _, item := range items {
					// 规则4: ItemId 必须 > 0 且存在于 Item 表
					if item.itemId <= 0 {
						errors = append(errors, fmt.Sprintf("Item 中 ItemId=%d 必须 > 0", item.itemId))
					} else if validItemIds != nil && !validItemIds[item.itemId] {
						errors = append(errors, fmt.Sprintf("Item 中 ItemId=%d 在道具表中不存在", item.itemId))
					}
					// 规则5: ItemNum 必须 > 0
					if item.itemNum <= 0 {
						errors = append(errors, fmt.Sprintf("Item 中 ItemNum=%d 必须 > 0", item.itemNum))
					}
				}
			}
		}

		if len(errors) > 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("掉落道具【%s】(ID=%s) %s", rowName, rowId, strings.Join(errors, "; ")),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个掉落道具条件和互斥问题", len(result.ErrCells))
	}

	return result
}

// itemPair 解析后的道具ID和数量对
type itemPair struct {
	itemId  int
	itemNum int
}

// parseItemCfg 解析 Item 配置字段
// 格式: {道具ID;数量}{道具ID;数量}...
// 例如: {1000005;1}{1000011;10}{9000001;5}
func (c *DropItemValidityCheckRule) parseItemCfg(itemCfg string) []itemPair {
	var items []itemPair
	matches := itemCfgRegex.FindAllStringSubmatch(itemCfg, -1)
	for _, match := range matches {
		if len(match) == 3 {
			itemId, err1 := strconv.Atoi(match[1])
			itemNum, err2 := strconv.Atoi(match[2])
			if err1 == nil && err2 == nil {
				items = append(items, itemPair{itemId: itemId, itemNum: itemNum})
			}
		}
	}
	return items
}

// parseBool 解析布尔值（支持多种格式）
func (c *DropItemValidityCheckRule) parseBool(val string) bool {
	val = strings.TrimSpace(strings.ToLower(val))
	return val == "true" || val == "1" || val == "yes"
}

// loadSheet 从 sheetMap 中加载指定表的数据
func (c *DropItemValidityCheckRule) loadSheet(sheetMap map[string]*excelize.File, suffix string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, suffix); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildIdSet 从指定表的 Id 列构建有效 ID 集合
func (c *DropItemValidityCheckRule) buildIdSet(cols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(cols, "Id")
	if idColIdx < 0 {
		return validIds
	}
	endIdx := helpers.GetDataEndIndex(cols, startRowIdx)
	for rowIdx := startRowIdx; rowIdx < endIdx; rowIdx++ {
		idStr := helpers.GetColValue(cols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		validIds[id] = true
	}
	return validIds
}
