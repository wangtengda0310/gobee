package draw_pet

// activity-wiki-dev: 活动Wiki开发技能生成

import (
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// DrawPetDiff 结缘亭表差异数据结构
type DrawPetDiff struct {
	Id                  int
	Name                string
	EnsureNewer         int
	EnsureCount         int
	FirstEnsureNewer    int
	OnceDropRule        int
	TenDropRule         int
	OnceItemCost        []ItemCfg
	TenItemCost         []ItemCfg
	DailyLimit          int
	ActivityId          []int
	BigAwardCount       int
	BigAwardItemId      int
	StartTime           string
	EndTime             string
	BigAwards           []ItemCfg
	PartnerItem         *ItemCfg
	Byproducts          []int
	ProbabilityContents []string
	ProbabilityPercents []float64
	DrawPetTitleIcon    string
	DrawPetTitleDescBg  string
	DrawRuleContent     string
}

// ItemCfg 道具配置
type ItemCfg struct {
	ItemId int
	Count  int
}

func (d DrawPetDiff) GetType() string {
	return "DrawPetDiff"
}

func (d DrawPetDiff) GetDisplayName() string {
	return d.Name
}

// GetDrawPetDiffMap 解析结缘亭表数据
func GetDrawPetDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]DrawPetDiff, err error) {
	var drawPetCols [][]string
	if sheet, exist := sheetMap["结缘亭|DrawPet"]; exist {
		var err error
		drawPetCols, err = sheet.GetCols("结缘亭|DrawPet")
		if err != nil {
			return nil, err
		}
	}

	// 如果没有找到sheet，返回空数组
	if drawPetCols == nil || len(drawPetCols) == 0 {
		return &[]DrawPetDiff{}, nil
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM

	drawPetsDiff := make([]DrawPetDiff, 0, 50)

	reg := regexp.MustCompile(`\{(\d+);\d+}`)

	// 检查Id列是否有数据
	if len(drawPetCols) <= Id || len(drawPetCols[Id]) <= startRow {
		return &drawPetsDiff, nil
	}

	endIndex := helpers.AutoDetectEndIndex(drawPetCols, Id, startRow, 3)
	idCol := drawPetCols[Id][startRow:endIndex]

	for i, idStr := range idCol {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			drawPetDiff := DrawPetDiff{}

			name := utils.GetCellValue(drawPetCols, Name, startRow+i)
			ensureNewer := utils.GetCellValue(drawPetCols, EnsureNewer, startRow+i)
			ensureCount := utils.GetCellValue(drawPetCols, EnsureCount, startRow+i)
			firstEnsureNewer := utils.GetCellValue(drawPetCols, FirstEnsureNewer, startRow+i)
			onceDropRule := utils.GetCellValue(drawPetCols, OnceDropRule, startRow+i)
			tenDropRule := utils.GetCellValue(drawPetCols, TenDropRule, startRow+i)
			onceItemCost := utils.GetCellValue(drawPetCols, OnceItemCost, startRow+i)
			tenItemCost := utils.GetCellValue(drawPetCols, TenItemCost, startRow+i)
			dailyLimit := utils.GetCellValue(drawPetCols, DailyLimit, startRow+i)
			activityId := utils.GetCellValue(drawPetCols, ActivityId, startRow+i)
			bigAwardCount := utils.GetCellValue(drawPetCols, BigAwardCount, startRow+i)
			bigAwardItemId := utils.GetCellValue(drawPetCols, BigAwardItemId, startRow+i)
			startTime := utils.GetCellValue(drawPetCols, StartTime, startRow+i)
			endTime := utils.GetCellValue(drawPetCols, EndTime, startRow+i)
			bigAwards := utils.GetCellValue(drawPetCols, BigAwards, startRow+i)
			partnerItem := utils.GetCellValue(drawPetCols, PartnerItem, startRow+i)
			byproducts := utils.GetCellValue(drawPetCols, Byproducts, startRow+i)
			probabilityContents := utils.GetCellValue(drawPetCols, ProbabilityContents, startRow+i)
			probabilityPercents := utils.GetCellValue(drawPetCols, ProbabilityPercents, startRow+i)
			drawPetTitleIcon := utils.GetCellValue(drawPetCols, DrawPetTitleIcon, startRow+i)
			drawPetTitleDescBg := utils.GetCellValue(drawPetCols, DrawPetTitleDescBg, startRow+i)
			drawRuleContent := utils.GetCellValue(drawPetCols, DrawRuleContent, startRow+i)

			drawPetDiff.Id = id
			drawPetDiff.Name = name

			if n, err := strconv.Atoi(ensureNewer); err == nil {
				drawPetDiff.EnsureNewer = n
			} else {
				drawPetDiff.EnsureNewer = -1
			}

			if n, err := strconv.Atoi(ensureCount); err == nil {
				drawPetDiff.EnsureCount = n
			} else {
				drawPetDiff.EnsureCount = -1
			}

			if n, err := strconv.Atoi(firstEnsureNewer); err == nil {
				drawPetDiff.FirstEnsureNewer = n
			} else {
				drawPetDiff.FirstEnsureNewer = -1
			}

			if n, err := strconv.Atoi(onceDropRule); err == nil {
				drawPetDiff.OnceDropRule = n
			} else {
				drawPetDiff.OnceDropRule = -1
			}

			if n, err := strconv.Atoi(tenDropRule); err == nil {
				drawPetDiff.TenDropRule = n
			} else {
				drawPetDiff.TenDropRule = -1
			}

			drawPetDiff.OnceItemCost = parseItemCost(onceItemCost, reg)
			drawPetDiff.TenItemCost = parseItemCost(tenItemCost, reg)

			if n, err := strconv.Atoi(dailyLimit); err == nil {
				drawPetDiff.DailyLimit = n
			} else {
				drawPetDiff.DailyLimit = -1
			}

			// ActivityId 为单个整数
			if n, err := strconv.Atoi(activityId); err == nil {
				drawPetDiff.ActivityId = []int{n}
			} else {
				drawPetDiff.ActivityId = []int{}
			}

			drawPetDiff.StartTime = startTime
			drawPetDiff.EndTime = endTime

			if n, err := strconv.Atoi(bigAwardCount); err == nil {
				drawPetDiff.BigAwardCount = n
			} else {
				drawPetDiff.BigAwardCount = -1
			}

			if n, err := strconv.Atoi(bigAwardItemId); err == nil {
				drawPetDiff.BigAwardItemId = n
			} else {
				drawPetDiff.BigAwardItemId = -1
			}

			drawPetDiff.BigAwards = parseItemCost(bigAwards, reg)
			drawPetDiff.PartnerItem = parseSingleItemCfg(partnerItem, reg)
			drawPetDiff.Byproducts = parseIntArray(byproducts)
			drawPetDiff.ProbabilityContents = parseStringArray(probabilityContents)
			drawPetDiff.ProbabilityPercents = parseFloatArray(probabilityPercents)
			drawPetDiff.DrawPetTitleIcon = drawPetTitleIcon
			drawPetDiff.DrawPetTitleDescBg = drawPetTitleDescBg
			drawPetDiff.DrawRuleContent = drawRuleContent

			drawPetsDiff = append(drawPetsDiff, drawPetDiff)
		}
	}

	return &drawPetsDiff, nil
}

// parseItemCost 解析道具消耗字符串，格式为 {itemId;count}，多个用逗号分隔
func parseItemCost(itemCostStr string, reg *regexp.Regexp) []ItemCfg {
	itemCost := make([]ItemCfg, 0, 5)
	if itemCostStr == "" {
		return itemCost
	}

	matches := reg.FindAllStringSubmatch(itemCostStr, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			fullMatch := match[0]
			itemIdStr := match[1]

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

// parseSingleItemCfg 解析单条道具配置
func parseSingleItemCfg(itemStr string, reg *regexp.Regexp) *ItemCfg {
	if itemStr == "" {
		return nil
	}
	matches := reg.FindStringSubmatch(itemStr)
	if len(matches) >= 2 {
		fullMatch := matches[0]
		itemIdStr := matches[1]
		count := 1
		parts := strings.Split(strings.Trim(fullMatch, "{}"), ";")
		if len(parts) >= 2 {
			if c, err := strconv.Atoi(parts[1]); err == nil {
				count = c
			}
		}
		itemId, _ := strconv.Atoi(itemIdStr)
		return &ItemCfg{
			ItemId: itemId,
			Count:  count,
		}
	}
	return nil
}

// parseIntArray 解析逗号分隔的int数组
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
		if n, err := strconv.Atoi(part); err == nil {
			result = append(result, n)
		}
	}
	return result
}

// parseStringArray 解析逗号分隔的string数组
func parseStringArray(str string) []string {
	if str == "" {
		return []string{}
	}
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// parseFloatArray 解析逗号分隔的float数组
func parseFloatArray(str string) []float64 {
	result := make([]float64, 0, 10)
	if str == "" {
		return result
	}
	parts := strings.Split(str, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if f, err := strconv.ParseFloat(part, 64); err == nil {
			result = append(result, f)
		}
	}
	return result
}

// parseIntArrayFromBraces 解析花括号格式的ID数组，格式为 {id;weight},{id;weight}
func parseIntArrayFromBraces(str string) []int {
	result := make([]int, 0, 5)
	if str == "" {
		return result
	}
	// 匹配 {数字;数字} 中的第一个数字
	reg := regexp.MustCompile(`\{(\d+);\d+}`)
	matches := reg.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			if n, err := strconv.Atoi(match[1]); err == nil {
				result = append(result, n)
			}
		}
	}
	return result
}
