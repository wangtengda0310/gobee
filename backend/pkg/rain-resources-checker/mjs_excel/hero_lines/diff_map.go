package hero_lines

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroLinesDiff struct {
	Id           int
	Type         string // ELinesType
	TabName      string
	Text         string
	AudioId      string // EAudioId
	Achievements []int
	GroupId      string // ELinesGroupId
}

func (h HeroLinesDiff) GetType() string {
	return "HeroLinesDiff"
}

func (h HeroLinesDiff) GetDisplayName() string {
	return h.TabName
}

func GetHeroLinesDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroLinesDiff, err error) {
	var linesCols [][]string // 武将台词表
	if sheet, exist := sheetMap["武将台词|HeroLines"]; exist {
		var err error
		linesCols, err = sheet.GetCols("武将台词|HeroLines")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建武将台词Map
	linesDiff := make([]HeroLinesDiff, 0, 300)

	for i, idStr := range linesCols[Id][startRow:helpers.AutoDetectEndIndex(linesCols, Id, startRow, 3)] {
		// 表格第一列是int类型Id，用数字判断有效性
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			linesDiffItem := HeroLinesDiff{}

			type_ := utils.GetCellValue(linesCols, Type, startRow+i)
			tabName := utils.GetCellValue(linesCols, TabName, startRow+i)
			text := utils.GetCellValue(linesCols, Text, startRow+i)
			audioId := utils.GetCellValue(linesCols, AudioId, startRow+i)
			achievements := utils.GetCellValue(linesCols, Achievements, startRow+i)
			groupId := utils.GetCellValue(linesCols, GroupId, startRow+i)

			// 录入信息
			linesDiffItem.Id = id
			linesDiffItem.Type = type_
			linesDiffItem.TabName = tabName
			linesDiffItem.Text = text
			linesDiffItem.AudioId = audioId

			// 关联成就（int数组）
			linesDiffItem.Achievements = make([]int, 0, 5)
			for _, ss := range strings.Split(achievements, ",") {
				if n, err := strconv.Atoi(ss); err == nil && ss != "" {
					linesDiffItem.Achievements = append(linesDiffItem.Achievements, n)
				}
			}

			linesDiffItem.GroupId = groupId

			linesDiff = append(linesDiff, linesDiffItem)
		}
	}

	return &linesDiff, nil
}
