package audio

import (
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

func GetAudioMap(sheetMap map[string]*excelize.File) (audioEnumPathMap map[string][]string, audioEnumTypeEnumMap map[string]string, err error) {
	var audioCols [][]string // 技能台词表
	if sheet, exist := sheetMap["音频配置表|Audio"]; exist {
		var err error
		audioCols, err = sheet.GetCols("音频配置表|Audio")
		if err != nil {
			return nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建音频Map
	audioEnumPathMap = make(map[string][]string)
	audioEnumTypeEnumMap = make(map[string]string)

	for i, idStr := range audioCols[Id][startRow:helpers.AutoDetectEndIndex(audioCols, Id, startRow, 3)] {
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		} else {
			if utils.GetCellValue(audioCols, Path, startRow+i) == "" {
				continue
			}
			if utils.GetCellValue(audioCols, PathRandom, startRow+i) != "" {
				audioEnumPathMap[idStr] = strings.Split(utils.GetCellValue(audioCols, PathRandom, startRow+i), ",")
			} else {
				audioEnumPathMap[idStr] = []string{utils.GetCellValue(audioCols, Path, startRow+i)}
			}
			audioEnumTypeEnumMap[idStr] = utils.GetCellValue(audioCols, AudioType, startRow+i)
		}
	}

	return audioEnumPathMap, audioEnumTypeEnumMap, nil
}
