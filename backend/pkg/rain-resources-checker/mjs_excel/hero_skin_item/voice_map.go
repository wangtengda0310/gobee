package hero_skin_item

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkinLineVoiceMap map[int]SkinLineVoice

type SkinLineVoice struct {
	Lines      []int
	DebutLines []int
	KillLines  []int
	DeadLines  []int
	HeroAudio  []int
}

func GetHeroSkinsItemVoiceMap(sheetMap map[string]*excelize.File) (heroSkinLineVoiceMap map[int]SkinLineVoiceMap, skinIdNameMap map[int]string, skinIdOpenMap map[int]bool, err error) {
	var heroSkinCols [][]string // 武将皮肤表
	if sheet, exist := sheetMap["英雄皮肤|HeroSkinItem"]; exist {
		var err error
		heroSkinCols, err = sheet.GetCols("英雄皮肤|HeroSkinItem")
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建皮肤Map
	heroSkinLineVoiceMap = make(map[int]SkinLineVoiceMap)
	skinIdNameMap = make(map[int]string)
	skinIdOpenMap = make(map[int]bool)

	for i, idStr := range heroSkinCols[HeroId][startRow:helpers.AutoDetectEndIndex(heroSkinCols, HeroId, startRow, 3)] {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			if utils.GetCellValue(heroSkinCols, Lines, startRow+i) == "" {
				continue
			}
			if skinId, err := strconv.Atoi(utils.GetCellValue(heroSkinCols, SkinItemId, startRow+i)); err != nil {
				continue
			} else {
				if heroSkinLineVoiceMap[id] == nil {
					heroSkinLineVoiceMap[id] = make(SkinLineVoiceMap)
				}

				if open, err := strconv.ParseBool(utils.GetCellValue(heroSkinCols, IsOpen, startRow+i)); err != nil {
					continue
				} else {
					skinIdOpenMap[skinId] = open
				}

				// 格式化所有int数组
				lines := utils.GetIntArr(utils.GetCellValue(heroSkinCols, Lines, startRow+i))
				debutLines := utils.GetIntArr(utils.GetCellValue(heroSkinCols, DebutLines, startRow+i))
				killLines := utils.GetIntArr(utils.GetCellValue(heroSkinCols, KillLines, startRow+i))
				deadLines := utils.GetIntArr(utils.GetCellValue(heroSkinCols, DeadLines, startRow+i))
				heroAudio := utils.GetIntArr(utils.GetCellValue(heroSkinCols, HeroAudio, startRow+i))

				heroSkinLineVoiceMap[id][skinId] = SkinLineVoice{
					Lines:      lines,
					DebutLines: debutLines,
					KillLines:  killLines,
					DeadLines:  deadLines,
					HeroAudio:  heroAudio,
				}

				// 名称同步
				skinIdNameMap[skinId] = utils.GetCellValue(heroSkinCols, Name, startRow+i)
			}
		}
	}

	return heroSkinLineVoiceMap, skinIdNameMap, skinIdOpenMap, nil
}
