package draw_skin

import (
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type DrawSkinDiff struct {
	Id             int
	Name           string
	OnceDropRule   int
	TenDropRule    int
	OnceItemCost   []ItemCfg
	TenItemCost    []ItemCfg
	ActivityId     int
	StartTime      string
	EndTime        string
	BigAwardCount  int
	BigAwardItemId int
}

type ItemCfg struct {
	ItemId int
	Count  int
}

func (d DrawSkinDiff) GetType() string {
	return "DrawSkinDiff"
}

func (d DrawSkinDiff) GetDisplayName() string {
	return d.Name
}

func GetDrawSkinDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]DrawSkinDiff, err error) {
	var drawSkinCols [][]string // 皮肤抽奖表
	if sheet, exist := sheetMap["皮肤抽奖|DrawSkin"]; exist {
		var err error
		drawSkinCols, err = sheet.GetCols("皮肤抽奖|DrawSkin")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建皮肤抽奖Map
	drawSkinsDiff := make([]DrawSkinDiff, 0, 50)

	reg := regexp.MustCompile(`\{(\d+);\d+}`)

	for i, idStr := range drawSkinCols[Id][startRow:helpers.AutoDetectEndIndex(drawSkinCols, Id, startRow, 3)] {
		// 表格第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			drawSkinDiff := DrawSkinDiff{}

			name := utils.GetCellValue(drawSkinCols, Name, startRow+i)
			onceDropRule := utils.GetCellValue(drawSkinCols, OnceDropRule, startRow+i)
			tenDropRule := utils.GetCellValue(drawSkinCols, TenDropRule, startRow+i)
			onceItemCost := utils.GetCellValue(drawSkinCols, OnceItemCost, startRow+i)
			tenItemCost := utils.GetCellValue(drawSkinCols, TenItemCost, startRow+i)
			activityId := utils.GetCellValue(drawSkinCols, ActivityId, startRow+i)
			startTime := utils.GetCellValue(drawSkinCols, StartTime, startRow+i)
			endTime := utils.GetCellValue(drawSkinCols, EndTime, startRow+i)
			bigAwardCount := utils.GetCellValue(drawSkinCols, BigAwardCount, startRow+i)
			bigAwardItemId := utils.GetCellValue(drawSkinCols, BigAwardItemId, startRow+i)

			// 录入皮肤抽奖信息
			drawSkinDiff.Id = id
			drawSkinDiff.Name = name

			if n, err := strconv.Atoi(onceDropRule); err == nil {
				drawSkinDiff.OnceDropRule = n
			} else {
				drawSkinDiff.OnceDropRule = -1
			}

			if n, err := strconv.Atoi(tenDropRule); err == nil {
				drawSkinDiff.TenDropRule = n
			} else {
				drawSkinDiff.TenDropRule = -1
			}

			// 解析单抽消耗道具，格式为 {itemId;count}，多个用逗号分隔
			drawSkinDiff.OnceItemCost = parseItemCost(onceItemCost, reg)

			// 解析十抽消耗道具，格式为 {itemId;count}，多个用逗号分隔
			drawSkinDiff.TenItemCost = parseItemCost(tenItemCost, reg)

			if n, err := strconv.Atoi(activityId); err == nil {
				drawSkinDiff.ActivityId = n
			} else {
				drawSkinDiff.ActivityId = -1
			}

			drawSkinDiff.StartTime = startTime
			drawSkinDiff.EndTime = endTime

			if n, err := strconv.Atoi(bigAwardCount); err == nil {
				drawSkinDiff.BigAwardCount = n
			} else {
				drawSkinDiff.BigAwardCount = -1
			}

			if n, err := strconv.Atoi(bigAwardItemId); err == nil {
				drawSkinDiff.BigAwardItemId = n
			} else {
				drawSkinDiff.BigAwardItemId = -1
			}

			drawSkinsDiff = append(drawSkinsDiff, drawSkinDiff)
		}
	}

	return &drawSkinsDiff, nil
}

// parseItemCost 解析道具消耗字符串，格式为 {itemId;count}，多个用逗号分隔
func parseItemCost(itemCostStr string, reg *regexp.Regexp) []ItemCfg {
	itemCost := make([]ItemCfg, 0, 5)
	if itemCostStr == "" {
		return itemCost
	}

	// 使用正则找出所有匹配项
	matches := reg.FindAllStringSubmatch(itemCostStr, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			// 解析完整格式 {itemId;count}
			fullMatch := match[0] // 完整的 {数字;数字}
			itemIdStr := match[1] // 第一个数字

			// 解析数量
			count := 1
			parts := strings.Split(strings.Trim(fullMatch, "{}"), ";")
			if len(parts) >= 2 {
				if c, err := strconv.Atoi(parts[1]); err == nil {
					count = c
				}
			}

			itemId, _ := strconv.Atoi(itemIdStr)

			itemCost = append(itemCost, ItemCfg{
				ItemId: itemId,
				Count:  count,
			})
		}
	}

	return itemCost
}
