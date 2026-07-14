package skill_lines

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkillLineVoice struct {
	SkillFirstLine  []int
	SkillSecondLine []int
	SkillThirdLine  []int
	SkillForthLine  []int
	SpecialAudio    []int
}

func GetSkillLinesMap(sheetMap map[string]*excelize.File, skillEnumIdMap map[string]int) (skillIdLineVoiceMap map[int]SkillLineVoice, err error) {
	var skillLinesCols [][]string // 技能台词表
	if sheet, exist := sheetMap["技能台词配置表|SkillLines"]; exist {
		var err error
		skillLinesCols, err = sheet.GetCols("技能台词配置表|SkillLines")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能Map
	skillIdLineVoiceMap = make(map[int]SkillLineVoice)

	for i, idStr := range skillLinesCols[Id][startRow:helpers.AutoDetectEndIndex(skillLinesCols, Id, startRow, 3)] {
		if _, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			if utils.GetCellValue(skillLinesCols, SkillId, startRow+i) == "" {
				continue
			}
			if skillId, exist := skillEnumIdMap[utils.GetCellValue(skillLinesCols, SkillId, startRow+i)]; exist {

				first := utils.GetIntArr(utils.GetCellValue(skillLinesCols, SkillFirstLine, startRow+i))
				second := utils.GetIntArr(utils.GetCellValue(skillLinesCols, SkillSecondLine, startRow+i))
				third := utils.GetIntArr(utils.GetCellValue(skillLinesCols, SkillThirdLine, startRow+i))
				forth := utils.GetIntArr(utils.GetCellValue(skillLinesCols, SkillForthLine, startRow+i))
				spec := utils.GetIntArr(utils.GetCellValue(skillLinesCols, SpecialAudio, startRow+i))

				skillIdLineVoiceMap[skillId] = SkillLineVoice{
					SkillFirstLine:  first,
					SkillSecondLine: second,
					SkillThirdLine:  third,
					SkillForthLine:  forth,
					SpecialAudio:    spec,
				}
			} else {

			}
		}
	}

	return skillIdLineVoiceMap, nil
}
