package hero

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

func GetHeroMap(sheetMap map[string]*excelize.File) (heroIdOpenMap map[int]bool, heroSkillsMap map[int][]int, heroNameMap map[int]string, err error) {
	var heroCols [][]string // 武将表
	if sheet, exist := sheetMap["武将表|Hero"]; exist {
		var err error
		heroCols, err = sheet.GetCols("武将表|Hero")
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建开放武将Map
	heroIdOpenMap = make(map[int]bool)
	heroSkillsMap = make(map[int][]int)
	heroNameMap = make(map[int]string)

	for i, idStr := range heroCols[Id][startRow:helpers.AutoDetectEndIndex(heroCols, Id, startRow, 3)] {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			if utils.GetCellValue(heroCols, IsOpen, startRow+i) == "" || utils.GetCellValue(heroCols, HeroType, startRow+i) == "" {
				continue
			}
			if utils.GetCellValue(heroCols, IsOpen, startRow+i) == "1" && utils.GetCellValue(heroCols, HeroType, startRow+i) == "1" {
				heroIdOpenMap[id] = true
			} else {
				heroIdOpenMap[id] = false
			}
			if utils.GetCellValue(heroCols, Skill, startRow+i) != "" {
				skills := utils.GetIntArr(utils.GetCellValue(heroCols, Skill, startRow+i))
				heroSkillsMap[id] = skills
			}
			if utils.GetCellValue(heroCols, Name, startRow+i) != "" {
				heroNameMap[id] = utils.GetCellValue(heroCols, Name, startRow+i)
			}

		}
	}

	return heroIdOpenMap, heroSkillsMap, heroNameMap, nil
}
