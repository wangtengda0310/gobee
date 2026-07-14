package hero_lines

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

func GetHeroLinesMap(sheetMap map[string]*excelize.File) (lineVoiceIdEnumMap map[int]string, lineVoiceEnumIdMap map[string]int, err error) {
	var heroLinesCols [][]string // 技能台词表
	if sheet, exist := sheetMap["武将台词|HeroLines"]; exist {
		var err error
		heroLinesCols, err = sheet.GetCols("武将台词|HeroLines")
		if err != nil {
			return nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建台词Map
	lineVoiceIdEnumMap = make(map[int]string)
	lineVoiceEnumIdMap = make(map[string]int)

	for i, idStr := range heroLinesCols[Id][startRow:helpers.AutoDetectEndIndex(heroLinesCols, Id, startRow, 3)] {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			if utils.GetCellValue(heroLinesCols, AudioId, startRow+i) == "" {
				continue
			}
			lineVoiceIdEnumMap[id] = utils.GetCellValue(heroLinesCols, AudioId, startRow+i)
			lineVoiceEnumIdMap[utils.GetCellValue(heroLinesCols, AudioId, startRow+i)] = id
		}
	}

	return lineVoiceIdEnumMap, lineVoiceEnumIdMap, nil
}
