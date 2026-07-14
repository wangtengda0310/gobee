package skill

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

func GetSkillMap(sheetMap map[string]*excelize.File) (skillEnumIdMap map[string]int, skillIdEnumMap, skillIdNameMap map[int]string, err error) {
	var heroSkillCols [][]string // 技能表
	if sheet, exist := sheetMap["技能表|Skill"]; exist {
		var err error
		heroSkillCols, err = sheet.GetCols("技能表|Skill")
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能Map
	skillEnumIdMap = make(map[string]int)
	skillIdEnumMap = make(map[int]string)
	skillIdNameMap = make(map[int]string)

	for i, idStr := range heroSkillCols[Id][startRow:helpers.AutoDetectEndIndex(heroSkillCols, Id, startRow, 3)] {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			if utils.GetCellValue(heroSkillCols, ESkillId, startRow+i) == "" {
				continue
			}
			skillIdEnumMap[id] = utils.GetCellValue(heroSkillCols, ESkillId, startRow+i)
			skillIdNameMap[id] = utils.GetCellValue(heroSkillCols, SkillName, startRow+i)
			skillEnumIdMap[utils.GetCellValue(heroSkillCols, ESkillId, startRow+i)] = id
		}
	}

	return skillEnumIdMap, skillIdEnumMap, skillIdNameMap, nil
}
