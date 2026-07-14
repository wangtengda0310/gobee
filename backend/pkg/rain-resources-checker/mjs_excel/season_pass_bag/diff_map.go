package season_pass_bag

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// SeasonPassBagDiff 赛季战令礼包表差异数据结构
type SeasonPassBagDiff struct {
	Id                    int
	Name                  string
	SeasonPassId          int
	BagType               string
	IsUnlockTask          bool
	UnlockHighlReward     bool
	PassExtWeeklyExpLimit int
	UnlockHeroQuality     []int
	AddLevel              int
	FirstReward           []ItemCfg
	UpReward              []ItemCfg
	ShowReward            []ItemCfg
}

// ItemCfg 道具配置
type ItemCfg struct {
	ItemId int
	Count  int
}

func (s SeasonPassBagDiff) GetType() string {
	return "SeasonPassBagDiff"
}

func (s SeasonPassBagDiff) GetDisplayName() string {
	return s.Name
}

// GetSeasonPassBagDiffMap 解析赛季战令礼包表数据
func GetSeasonPassBagDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SeasonPassBagDiff, err error) {
	var bagCols [][]string
	if sheet, exist := sheetMap["战令礼包表|SeasonPassBag"]; exist {
		var err error
		bagCols, err = sheet.GetCols("战令礼包表|SeasonPassBag")
		if err != nil {
			return nil, err
		}
	}

	// 如果没有找到sheet，返回空数组
	if bagCols == nil || len(bagCols) == 0 {
		return &[]SeasonPassBagDiff{}, nil
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM

	bagsDiff := make([]SeasonPassBagDiff, 0, 50)

	// 检查Id列是否有数据
	if len(bagCols) <= Id || len(bagCols[Id]) <= startRow {
		return &bagsDiff, nil
	}

	endIndex := helpers.AutoDetectEndIndex(bagCols, Id, startRow, 3)
	idCol := bagCols[Id][startRow:endIndex]

	for i, idStr := range idCol {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			bagDiff := SeasonPassBagDiff{}

			name := utils.GetCellValue(bagCols, Name, startRow+i)
			seasonPassId := utils.GetCellValue(bagCols, SeasonPassId, startRow+i)
			bagType := utils.GetCellValue(bagCols, BagType, startRow+i)
			isUnlockTask := utils.GetCellValue(bagCols, IsUnlockTask, startRow+i)
			unlockHighlReward := utils.GetCellValue(bagCols, UnlockHighlReward, startRow+i)
			passExtWeeklyExpLimit := utils.GetCellValue(bagCols, PassExtWeeklyExpLimit, startRow+i)
			unlockHeroQuality := utils.GetCellValue(bagCols, UnlockHeroQuality, startRow+i)
			addLevel := utils.GetCellValue(bagCols, AddLevel, startRow+i)
			firstReward := utils.GetCellValue(bagCols, FirstReward, startRow+i)
			upReward := utils.GetCellValue(bagCols, UpReward, startRow+i)
			showReward := utils.GetCellValue(bagCols, ShowReward, startRow+i)

			bagDiff.Id = id
			bagDiff.Name = name

			if n, err := strconv.Atoi(seasonPassId); err == nil {
				bagDiff.SeasonPassId = n
			} else {
				bagDiff.SeasonPassId = -1
			}

			bagDiff.BagType = bagType

			bagDiff.IsUnlockTask = strings.ToLower(isUnlockTask) == "true"
			bagDiff.UnlockHighlReward = strings.ToLower(unlockHighlReward) == "true"

			if n, err := strconv.Atoi(passExtWeeklyExpLimit); err == nil {
				bagDiff.PassExtWeeklyExpLimit = n
			} else {
				bagDiff.PassExtWeeklyExpLimit = -1
			}

			// 解析品质数组，格式为 "1,2,3"
			bagDiff.UnlockHeroQuality = make([]int, 0, 5)
			if unlockHeroQuality != "" {
				for _, qStr := range strings.Split(unlockHeroQuality, ",") {
					qStr = strings.TrimSpace(qStr)
					if qStr == "" {
						continue
					}
					if n, err := strconv.Atoi(qStr); err == nil {
						bagDiff.UnlockHeroQuality = append(bagDiff.UnlockHeroQuality, n)
					}
				}
			}

			if n, err := strconv.Atoi(addLevel); err == nil {
				bagDiff.AddLevel = n
			} else {
				bagDiff.AddLevel = -1
			}

			bagDiff.FirstReward = parseItemCfgList(firstReward)
			bagDiff.UpReward = parseItemCfgList(upReward)
			bagDiff.ShowReward = parseItemCfgList(showReward)

			bagsDiff = append(bagsDiff, bagDiff)
		}
	}

	return &bagsDiff, nil
}

// parseItemCfgList 解析多条道具配置字符串，格式为 {itemId;count},{itemId;count}
func parseItemCfgList(itemStr string) []ItemCfg {
	result := make([]ItemCfg, 0, 10)
	if itemStr == "" {
		return result
	}

	itemStr = strings.TrimSpace(itemStr)

	// 按逗号分割多个道具配置
	items := strings.Split(itemStr, "},{")
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.TrimPrefix(item, "{")
		item = strings.TrimSuffix(item, "}")

		parts := strings.Split(item, ";")
		if len(parts) >= 2 {
			itemId, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			if count == 0 {
				count = 1
			}
			result = append(result, ItemCfg{
				ItemId: itemId,
				Count:  count,
			})
		}
	}

	return result
}
