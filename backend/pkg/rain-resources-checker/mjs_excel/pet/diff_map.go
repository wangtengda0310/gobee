package pet

// activity-wiki-dev: 活动Wiki开发技能生成

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// PetDiff 灵宠表差异数据结构
type PetDiff struct {
	Id                int
	Name              string
	WuXingType        int
	PrefabPath        string
	Skills            []KVPair
	BattleAttrWeights []KVPair
	HighBattleAttrs   []int
	UpgradeLikeItems  []int
	UpgradeHateItems  []int
	SquareHeadIcon    string
	HeadIcon          string
	Silhouette        string
	PopBg             string
	PopIcon           string
	PopTitle          string
	PetWeekTaskBg     string
	InfoTextID        string
}

// KVPair 键值对结构
type KVPair struct {
	Key   int
	Value int
}

func (p PetDiff) GetType() string {
	return "PetDiff"
}

func (p PetDiff) GetDisplayName() string {
	return p.Name
}

// GetPetDiffMap 解析灵宠表数据
func GetPetDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]PetDiff, err error) {
	var petCols [][]string
	if sheet, exist := sheetMap["灵宠|Pet"]; exist {
		var err error
		petCols, err = sheet.GetCols("灵宠|Pet")
		if err != nil {
			return nil, err
		}
	}

	// 如果没有找到sheet，返回空数组
	if len(petCols) == 0 {
		return &[]PetDiff{}, nil
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM

	petsDiff := make([]PetDiff, 0, 100)

	// 检查Id列是否有数据
	if len(petCols) <= Id || len(petCols[Id]) <= startRow {
		return &petsDiff, nil
	}

	endIndex := helpers.AutoDetectEndIndex(petCols, Id, startRow, 3)
	idCol := petCols[Id][startRow:endIndex]

	for i, idStr := range idCol {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			petDiff := PetDiff{}

			name := utils.GetCellValue(petCols, Name, startRow+i)
			wuXingType := utils.GetCellValue(petCols, WuXingType, startRow+i)
			prefabPath := utils.GetCellValue(petCols, PrefabPath, startRow+i)
			skills := utils.GetCellValue(petCols, Skills, startRow+i)
			battleAttrWeights := utils.GetCellValue(petCols, BattleAttrWeights, startRow+i)
			highBattleAttrs := utils.GetCellValue(petCols, HighBattleAttrs, startRow+i)
			upgradeLikeItems := utils.GetCellValue(petCols, UpgradeLikeItems, startRow+i)
			upgradeHateItems := utils.GetCellValue(petCols, UpgradeHateItems, startRow+i)
			squareHeadIcon := utils.GetCellValue(petCols, SquareHeadIcon, startRow+i)
			headIcon := utils.GetCellValue(petCols, HeadIcon, startRow+i)
			silhouette := utils.GetCellValue(petCols, Silhouette, startRow+i)
			popBg := utils.GetCellValue(petCols, PopBg, startRow+i)
			popIcon := utils.GetCellValue(petCols, PopIcon, startRow+i)
			popTitle := utils.GetCellValue(petCols, PopTitle, startRow+i)
			petWeekTaskBg := utils.GetCellValue(petCols, PetWeekTaskBg, startRow+i)
			infoTextID := utils.GetCellValue(petCols, InfoTextID, startRow+i)

			petDiff.Id = id
			petDiff.Name = name

			if n, err := strconv.Atoi(wuXingType); err == nil {
				petDiff.WuXingType = n
			} else {
				petDiff.WuXingType = -1
			}

			petDiff.PrefabPath = prefabPath
			petDiff.Skills = parseKVPair(skills)
			petDiff.BattleAttrWeights = parseKVPair(battleAttrWeights)
			petDiff.HighBattleAttrs = utils.GetIntArr(highBattleAttrs)
			petDiff.UpgradeLikeItems = utils.GetIntArr(upgradeLikeItems)
			petDiff.UpgradeHateItems = utils.GetIntArr(upgradeHateItems)
			petDiff.SquareHeadIcon = squareHeadIcon
			petDiff.HeadIcon = headIcon
			petDiff.Silhouette = silhouette
			petDiff.PopBg = popBg
			petDiff.PopIcon = popIcon
			petDiff.PopTitle = popTitle
			petDiff.PetWeekTaskBg = petWeekTaskBg
			petDiff.InfoTextID = infoTextID

			petsDiff = append(petsDiff, petDiff)
		}
	}

	return &petsDiff, nil
}

// parseKVPair 解析KVPair字符串，格式为 key:value,key:value
func parseKVPair(str string) []KVPair {
	result := make([]KVPair, 0, 5)
	if str == "" {
		return result
	}

	pairs := strings.Split(str, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, ":")
		if len(parts) >= 2 {
			key, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			value, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			result = append(result, KVPair{Key: key, Value: value})
		}
	}

	return result
}
