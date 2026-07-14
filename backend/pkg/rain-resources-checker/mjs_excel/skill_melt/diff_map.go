package skill_melt

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkillMeltDiff struct {
	Id        string // ESkillId 枚举类型
	None1     string // 空列1
	None2     string // 空列2
	MeltPower int
	CanMelt   bool
}

func (s SkillMeltDiff) GetID() string {
	return s.Id
}

func (s SkillMeltDiff) GetType() string {
	return "SkillMeltDiff"
}

func (s SkillMeltDiff) GetDisplayName() string {
	return s.Id
}

func GetSkillMeltDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SkillMeltDiff, err error) {
	var skillMeltCols [][]string // 技能熔炼表
	if sheet, exist := sheetMap["技能熔炼表|SkillMelt"]; exist {
		var err error
		skillMeltCols, err = sheet.GetCols("技能熔炼表|SkillMelt")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能熔炼Map
	skillMeltsDiff := make([]SkillMeltDiff, 0, 100)

	for i, idStr := range skillMeltCols[Id][startRow:helpers.AutoDetectEndIndex(skillMeltCols, Id, startRow, 3)] {
		// 第一列是ESkillId枚举类型，判断规则：不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		skillMeltDiff := SkillMeltDiff{}

		// 获取所有列的值
		none1 := utils.GetCellValue(skillMeltCols, None1, startRow+i)
		none2 := utils.GetCellValue(skillMeltCols, None2, startRow+i)
		meltPower := utils.GetCellValue(skillMeltCols, MeltPower, startRow+i)
		canMelt := utils.GetCellValue(skillMeltCols, CanMelt, startRow+i)

		// 录入数据
		skillMeltDiff.Id = idStr
		skillMeltDiff.None1 = none1
		skillMeltDiff.None2 = none2

		if n, err := strconv.Atoi(meltPower); err == nil {
			skillMeltDiff.MeltPower = n
		} else {
			skillMeltDiff.MeltPower = -1
		}

		if b, err := strconv.ParseBool(canMelt); err == nil && b {
			skillMeltDiff.CanMelt = b
		} else {
			skillMeltDiff.CanMelt = false
		}

		skillMeltsDiff = append(skillMeltsDiff, skillMeltDiff)
	}

	return &skillMeltsDiff, nil
}
