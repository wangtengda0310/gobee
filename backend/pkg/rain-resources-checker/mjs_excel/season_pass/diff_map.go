package season_pass

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// SeasonPassDiff 赛季战令表差异数据结构
type SeasonPassDiff struct {
	Id                   int
	Name                 string
	LevelLimit           int
	StartTime            string
	EndTime              string
	ExtTime              string
	ExpPerLevel          int
	ExpShowItemId        int
	PerLevelBuyCost      []ItemCfg
	WeekExpLimit         int
	NormalLuckyBag       []ItemCfg
	HighLuckyBag         []ItemCfg
	LoopExp              int
	MailId               int
	HeroId               int
	SkinId               int
	SeasonIcon           string
	SeasonBg             string
	SeasonStartIcon      string
	GoldBagId            int
	TreasureBagId        int
	RechargeGoldId       int
	RechargeUpTreasureId int
	RechargeTreasureId   int
	SeasonGoldBg         string
	SeasonTreasureBg     string
}

// ItemCfg 道具配置
type ItemCfg struct {
	ItemId int
	Count  int
}

func (s SeasonPassDiff) GetType() string {
	return "SeasonPassDiff"
}

func (s SeasonPassDiff) GetDisplayName() string {
	return s.Name
}

// GetSeasonPassDiffMap 解析赛季战令表数据
func GetSeasonPassDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SeasonPassDiff, err error) {
	var seasonPassCols [][]string
	if sheet, exist := sheetMap["赛季战令表|SeasonPass"]; exist {
		var err error
		seasonPassCols, err = sheet.GetCols("赛季战令表|SeasonPass")
		if err != nil {
			return nil, err
		}
	}

	// 如果没有找到sheet，返回空数组
	if seasonPassCols == nil || len(seasonPassCols) == 0 {
		return &[]SeasonPassDiff{}, nil
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM

	seasonPassesDiff := make([]SeasonPassDiff, 0, 50)

	// 检查Id列是否有数据
	if len(seasonPassCols) <= Id || len(seasonPassCols[Id]) <= startRow {
		return &seasonPassesDiff, nil
	}

	endIndex := helpers.AutoDetectEndIndex(seasonPassCols, Id, startRow, 3)
	idCol := seasonPassCols[Id][startRow:endIndex]

	for i, idStr := range idCol {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			seasonPassDiff := SeasonPassDiff{}

			name := utils.GetCellValue(seasonPassCols, Name, startRow+i)
			levelLimit := utils.GetCellValue(seasonPassCols, LevelLimit, startRow+i)
			startTime := utils.GetCellValue(seasonPassCols, StartTime, startRow+i)
			endTime := utils.GetCellValue(seasonPassCols, EndTime, startRow+i)
			extTime := utils.GetCellValue(seasonPassCols, ExtTime, startRow+i)
			expPerLevel := utils.GetCellValue(seasonPassCols, ExpPerLevel, startRow+i)
			expShowItemId := utils.GetCellValue(seasonPassCols, ExpShowItemId, startRow+i)
			perLevelBuyCost := utils.GetCellValue(seasonPassCols, PerLevelBuyCost, startRow+i)
			weekExpLimit := utils.GetCellValue(seasonPassCols, WeekExpLimit, startRow+i)
			normalLuckyBag := utils.GetCellValue(seasonPassCols, NormalLuckyBag, startRow+i)
			highLuckyBag := utils.GetCellValue(seasonPassCols, HighLuckyBag, startRow+i)
			loopExp := utils.GetCellValue(seasonPassCols, LoopExp, startRow+i)
			mailId := utils.GetCellValue(seasonPassCols, MailId, startRow+i)
			heroId := utils.GetCellValue(seasonPassCols, HeroId, startRow+i)
			skinId := utils.GetCellValue(seasonPassCols, SkinId, startRow+i)
			seasonIcon := utils.GetCellValue(seasonPassCols, SeasonIcon, startRow+i)
			seasonBg := utils.GetCellValue(seasonPassCols, SeasonBg, startRow+i)
			seasonStartIcon := utils.GetCellValue(seasonPassCols, SeasonStartIcon, startRow+i)
			goldBagId := utils.GetCellValue(seasonPassCols, GoldBagId, startRow+i)
			treasureBagId := utils.GetCellValue(seasonPassCols, TreasureBagId, startRow+i)
			rechargeGoldId := utils.GetCellValue(seasonPassCols, RechargeGoldId, startRow+i)
			rechargeUpTreasureId := utils.GetCellValue(seasonPassCols, RechargeUpTreasureId, startRow+i)
			rechargeTreasureId := utils.GetCellValue(seasonPassCols, RechargeTreasureId, startRow+i)
			seasonGoldBg := utils.GetCellValue(seasonPassCols, SeasonGoldBg, startRow+i)
			seasonTreasureBg := utils.GetCellValue(seasonPassCols, SeasonTreasureBg, startRow+i)

			seasonPassDiff.Id = id
			seasonPassDiff.Name = name

			if n, err := strconv.Atoi(levelLimit); err == nil {
				seasonPassDiff.LevelLimit = n
			} else {
				seasonPassDiff.LevelLimit = -1
			}

			seasonPassDiff.StartTime = startTime
			seasonPassDiff.EndTime = endTime
			seasonPassDiff.ExtTime = extTime

			if n, err := strconv.Atoi(expPerLevel); err == nil {
				seasonPassDiff.ExpPerLevel = n
			} else {
				seasonPassDiff.ExpPerLevel = -1
			}

			if n, err := strconv.Atoi(expShowItemId); err == nil {
				seasonPassDiff.ExpShowItemId = n
			} else {
				seasonPassDiff.ExpShowItemId = -1
			}

			seasonPassDiff.PerLevelBuyCost = parseItemCfg(perLevelBuyCost)

			if n, err := strconv.Atoi(weekExpLimit); err == nil {
				seasonPassDiff.WeekExpLimit = n
			} else {
				seasonPassDiff.WeekExpLimit = -1
			}

			seasonPassDiff.NormalLuckyBag = parseItemCfg(normalLuckyBag)
			seasonPassDiff.HighLuckyBag = parseItemCfg(highLuckyBag)

			if n, err := strconv.Atoi(loopExp); err == nil {
				seasonPassDiff.LoopExp = n
			} else {
				seasonPassDiff.LoopExp = -1
			}

			if n, err := strconv.Atoi(mailId); err == nil {
				seasonPassDiff.MailId = n
			} else {
				seasonPassDiff.MailId = -1
			}

			if n, err := strconv.Atoi(heroId); err == nil {
				seasonPassDiff.HeroId = n
			} else {
				seasonPassDiff.HeroId = -1
			}

			if n, err := strconv.Atoi(skinId); err == nil {
				seasonPassDiff.SkinId = n
			} else {
				seasonPassDiff.SkinId = -1
			}

			seasonPassDiff.SeasonIcon = seasonIcon
			seasonPassDiff.SeasonBg = seasonBg
			seasonPassDiff.SeasonStartIcon = seasonStartIcon

			if n, err := strconv.Atoi(goldBagId); err == nil {
				seasonPassDiff.GoldBagId = n
			} else {
				seasonPassDiff.GoldBagId = -1
			}

			if n, err := strconv.Atoi(treasureBagId); err == nil {
				seasonPassDiff.TreasureBagId = n
			} else {
				seasonPassDiff.TreasureBagId = -1
			}

			if n, err := strconv.Atoi(rechargeGoldId); err == nil {
				seasonPassDiff.RechargeGoldId = n
			} else {
				seasonPassDiff.RechargeGoldId = -1
			}

			if n, err := strconv.Atoi(rechargeUpTreasureId); err == nil {
				seasonPassDiff.RechargeUpTreasureId = n
			} else {
				seasonPassDiff.RechargeUpTreasureId = -1
			}

			if n, err := strconv.Atoi(rechargeTreasureId); err == nil {
				seasonPassDiff.RechargeTreasureId = n
			} else {
				seasonPassDiff.RechargeTreasureId = -1
			}

			seasonPassDiff.SeasonGoldBg = seasonGoldBg
			seasonPassDiff.SeasonTreasureBg = seasonTreasureBg

			seasonPassesDiff = append(seasonPassesDiff, seasonPassDiff)
		}
	}

	return &seasonPassesDiff, nil
}

// parseItemCfg 解析单条道具配置字符串，格式为 {itemId;count}
func parseItemCfg(itemStr string) []ItemCfg {
	result := make([]ItemCfg, 0, 5)
	if itemStr == "" {
		return result
	}

	// 去除花括号
	itemStr = strings.TrimSpace(itemStr)
	itemStr = strings.TrimPrefix(itemStr, "{")
	itemStr = strings.TrimSuffix(itemStr, "}")

	parts := strings.Split(itemStr, ";")
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

	return result
}
