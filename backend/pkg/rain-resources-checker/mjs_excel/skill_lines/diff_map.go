package skill_lines

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkillLinesDiff struct {
	Id              int
	SkillId         string // ESkillId类型
	None            string // 空列
	SkinId          int
	SkillFirstLine  []int
	SkillSecondLine []int
	SkillThirdLine  []int
	SkillForthLine  []int
	SpecialAudio    string // EAudioId类型
}

func (s SkillLinesDiff) GetType() string {
	return "SkillLinesDiff"
}

func (s SkillLinesDiff) GetDisplayName() string {
	return strconv.Itoa(s.Id)
}

func GetSkillLinesDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SkillLinesDiff, err error) {
	var skillLinesCols [][]string // 技能台词配置表
	if sheet, exist := sheetMap["技能台词配置表|SkillLines"]; exist {
		var err error
		skillLinesCols, err = sheet.GetCols("技能台词配置表|SkillLines")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能台词Map
	skillLinesDiff := make([]SkillLinesDiff, 0, 300)

	for i, idStr := range skillLinesCols[Id][startRow:helpers.AutoDetectEndIndex(skillLinesCols, Id, startRow, 3)] {
		// 第一列是int类型的ID，按数字方式判断
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			skillLineDiff := SkillLinesDiff{}

			skillId := utils.GetCellValue(skillLinesCols, SkillId, startRow+i)
			none := utils.GetCellValue(skillLinesCols, None, startRow+i)
			skinId := utils.GetCellValue(skillLinesCols, SkinId, startRow+i)
			skillFirstLine := utils.GetCellValue(skillLinesCols, SkillFirstLine, startRow+i)
			skillSecondLine := utils.GetCellValue(skillLinesCols, SkillSecondLine, startRow+i)
			skillThirdLine := utils.GetCellValue(skillLinesCols, SkillThirdLine, startRow+i)
			skillForthLine := utils.GetCellValue(skillLinesCols, SkillForthLine, startRow+i)
			specialAudio := utils.GetCellValue(skillLinesCols, SpecialAudio, startRow+i)

			// 录入技能台词信息
			skillLineDiff.Id = id
			skillLineDiff.SkillId = skillId
			skillLineDiff.None = none

			if n, err := strconv.Atoi(skinId); err == nil {
				skillLineDiff.SkinId = n
			} else {
				skillLineDiff.SkinId = -1
			}

			// 技能第一段台词（int数组）
			skillLineDiff.SkillFirstLine = make([]int, 0, 5)
			for _, ss := range strings.Split(skillFirstLine, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					skillLineDiff.SkillFirstLine = append(skillLineDiff.SkillFirstLine, n)
				}
			}

			// 技能第二段台词（int数组）
			skillLineDiff.SkillSecondLine = make([]int, 0, 5)
			for _, ss := range strings.Split(skillSecondLine, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					skillLineDiff.SkillSecondLine = append(skillLineDiff.SkillSecondLine, n)
				}
			}

			// 技能第三段台词（int数组）
			skillLineDiff.SkillThirdLine = make([]int, 0, 5)
			for _, ss := range strings.Split(skillThirdLine, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					skillLineDiff.SkillThirdLine = append(skillLineDiff.SkillThirdLine, n)
				}
			}

			// 技能第四段台词（int数组）
			skillLineDiff.SkillForthLine = make([]int, 0, 5)
			for _, ss := range strings.Split(skillForthLine, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					skillLineDiff.SkillForthLine = append(skillLineDiff.SkillForthLine, n)
				}
			}

			skillLineDiff.SpecialAudio = specialAudio

			skillLinesDiff = append(skillLinesDiff, skillLineDiff)
		}
	}

	return &skillLinesDiff, nil
}
