package hero_skin_item

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

func GetHeroSkinsItemImgMap(sheetMap map[string]*excelize.File) (skinIdNameMap map[int]string, skinIdOpenMap map[int]bool, skinIdPinYinMap map[int]string, err error) {
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
	skinIdPinYinMap = make(map[int]string)
	skinIdNameMap = make(map[int]string)
	skinIdOpenMap = make(map[int]bool)

	for i, idStr := range heroSkinCols[HeroId][startRow:helpers.AutoDetectEndIndex(heroSkinCols, HeroId, startRow, 3)] {
		if _, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			if utils.GetCellValue(heroSkinCols, Lines, startRow+i) == "" {
				continue
			}
			if skinId, err := strconv.Atoi(utils.GetCellValue(heroSkinCols, SkinItemId, startRow+i)); err != nil {
				continue
			} else {
				if open, err := strconv.ParseBool(utils.GetCellValue(heroSkinCols, IsOpen, startRow+i)); err != nil {
					continue
				} else {
					skinIdOpenMap[skinId] = open
				}

				// 名称同步
				skinIdNameMap[skinId] = utils.GetCellValue(heroSkinCols, Name, startRow+i)
				skinIdPinYinMap[skinId] = utils.GetCellValue(heroSkinCols, SkinPinYin, startRow+i)
			}
		}
	}

	return skinIdNameMap, skinIdOpenMap, skinIdPinYinMap, nil
}
