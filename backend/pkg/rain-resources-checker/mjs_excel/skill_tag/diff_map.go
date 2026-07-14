package skill_tag

import (
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkillTagDiff struct {
	SkillTag string
	TagName  string
	TagColor string
	TagDes   string
	None     string
	None1    string
}

func (s SkillTagDiff) GetID() string {
	return s.SkillTag
}

func (s SkillTagDiff) GetType() string {
	return "SkillTagDiff"
}

func (s SkillTagDiff) GetDisplayName() string {
	return s.TagName
}

func GetSkillTagDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SkillTagDiff, err error) {
	var skillTagCols [][]string // 技能标签表
	if sheet, exist := sheetMap["技能标签表|SkillTag"]; exist {
		var err error
		skillTagCols, err = sheet.GetCols("技能标签表|SkillTag")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能标签Map
	skillTagsDiff := make([]SkillTagDiff, 0, 100)

	for i, skillTagStr := range skillTagCols[SkillTag][startRow:helpers.AutoDetectEndIndex(skillTagCols, SkillTag, startRow, 3)] {
		// 第一列是ESkillTag枚举类型，不为空且不以#开头即为有效行
		if skillTagStr == "" || strings.HasPrefix(skillTagStr, "#") {
			continue
		}

		skillTagDiff := SkillTagDiff{}

		tagName := utils.GetCellValue(skillTagCols, TagName, startRow+i)
		tagColor := utils.GetCellValue(skillTagCols, TagColor, startRow+i)
		tagDes := utils.GetCellValue(skillTagCols, TagDes, startRow+i)
		none := utils.GetCellValue(skillTagCols, None, startRow+i)
		none1 := utils.GetCellValue(skillTagCols, None1, startRow+i)

		// 录入技能标签信息
		skillTagDiff.SkillTag = skillTagStr
		skillTagDiff.TagName = tagName
		skillTagDiff.TagColor = tagColor
		skillTagDiff.TagDes = tagDes
		skillTagDiff.None = none
		skillTagDiff.None1 = none1

		skillTagsDiff = append(skillTagsDiff, skillTagDiff)
	}

	return &skillTagsDiff, nil
}
