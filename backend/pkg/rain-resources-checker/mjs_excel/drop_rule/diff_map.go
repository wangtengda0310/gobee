package drop_rule

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type DropRuleDiff struct {
	Id               int
	Name             string
	Count            int
	DropGroup        []int
	EnsureSmall      int
	EnsureSmallGroup []int
	EnsureBig        int
	EnsureBigGroup   []int
	EnsureItemCount  int
	EnsureItemID     int
	ItemCheckExist   bool
}

func (d DropRuleDiff) GetType() string {
	return "DropRuleDiff"
}

func (d DropRuleDiff) GetDisplayName() string {
	return d.Name
}

func GetDropRuleDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]DropRuleDiff, err error) {
	var dropRuleCols [][]string // 掉落规则表
	if sheet, exist := sheetMap["掉落规则表|DropRule"]; exist {
		var err error
		dropRuleCols, err = sheet.GetCols("掉落规则表|DropRule")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建掉落规则Map
	dropRulesDiff := make([]DropRuleDiff, 0, 300)

	for i, idStr := range dropRuleCols[Id][startRow:helpers.AutoDetectEndIndex(dropRuleCols, Id, startRow, 3)] {
		// 表格第一列是否有id作为读取这一列的标准
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			dropRuleDiff := DropRuleDiff{}

			name := utils.GetCellValue(dropRuleCols, Name, startRow+i)
			count := utils.GetCellValue(dropRuleCols, Count, startRow+i)
			dropGroup := utils.GetCellValue(dropRuleCols, DropGroup, startRow+i)
			ensureSmall := utils.GetCellValue(dropRuleCols, EnsureSmall, startRow+i)
			ensureSmallGroup := utils.GetCellValue(dropRuleCols, EnsureSmallGroup, startRow+i)
			ensureBig := utils.GetCellValue(dropRuleCols, EnsureBig, startRow+i)
			ensureBigGroup := utils.GetCellValue(dropRuleCols, EnsureBigGroup, startRow+i)
			ensureItemCount := utils.GetCellValue(dropRuleCols, EnsureItemCount, startRow+i)
			ensureItemID := utils.GetCellValue(dropRuleCols, EnsureItemID, startRow+i)
			itemCheckExist := utils.GetCellValue(dropRuleCols, ItemCheckExist, startRow+i)

			// 录入掉落规则信息
			dropRuleDiff.Id = id
			dropRuleDiff.Name = name

			if n, err := strconv.Atoi(count); err == nil {
				dropRuleDiff.Count = n
			} else {
				dropRuleDiff.Count = -1
			}

			dropRuleDiff.DropGroup = make([]int, 0, 10)
			for _, ss := range strings.Split(dropGroup, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					dropRuleDiff.DropGroup = append(dropRuleDiff.DropGroup, n)
				}
			}

			if n, err := strconv.Atoi(ensureSmall); err == nil {
				dropRuleDiff.EnsureSmall = n
			} else {
				dropRuleDiff.EnsureSmall = -1
			}

			dropRuleDiff.EnsureSmallGroup = make([]int, 0, 10)
			for _, ss := range strings.Split(ensureSmallGroup, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					dropRuleDiff.EnsureSmallGroup = append(dropRuleDiff.EnsureSmallGroup, n)
				}
			}

			if n, err := strconv.Atoi(ensureBig); err == nil {
				dropRuleDiff.EnsureBig = n
			} else {
				dropRuleDiff.EnsureBig = -1
			}

			dropRuleDiff.EnsureBigGroup = make([]int, 0, 10)
			for _, ss := range strings.Split(ensureBigGroup, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					dropRuleDiff.EnsureBigGroup = append(dropRuleDiff.EnsureBigGroup, n)
				}
			}

			if n, err := strconv.Atoi(ensureItemCount); err == nil {
				dropRuleDiff.EnsureItemCount = n
			} else {
				dropRuleDiff.EnsureItemCount = -1
			}

			if n, err := strconv.Atoi(ensureItemID); err == nil {
				dropRuleDiff.EnsureItemID = n
			} else {
				dropRuleDiff.EnsureItemID = -1
			}

			if b, err := strconv.ParseBool(itemCheckExist); err == nil {
				dropRuleDiff.ItemCheckExist = b
			} else {
				dropRuleDiff.ItemCheckExist = false
			}

			dropRulesDiff = append(dropRulesDiff, dropRuleDiff)
		}
	}

	return &dropRulesDiff, nil
}
