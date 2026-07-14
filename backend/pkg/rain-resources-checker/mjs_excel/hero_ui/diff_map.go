package hero_ui

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroUIDiff struct {
	Id                  int
	Name                string
	VoiceType           int
	BelongExpansionPack string
	AwardDes            string
	LongIntroduction    string
	ShortIntroduction   string
	Evidence            string
	Evaluation          string
	IsNew               bool
	CopyWriter          string
	SkillDesigner       string
	GetWay              string
	Position            string
	ExclusiveCard       []int
	NewbieShowSkillTag  []int
	WinningRateIn2v2    int
	WinRateShowPriority int
}

func (h HeroUIDiff) GetType() string {
	return "HeroUIDiff"
}

func (h HeroUIDiff) GetDisplayName() string {
	return h.Name
}

func GetHeroUIDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroUIDiff, err error) {
	var heroUICols [][]string // 武将表现配置表
	if sheet, exist := sheetMap["武将表现配置|HeroUI"]; exist {
		var err error
		heroUICols, err = sheet.GetCols("武将表现配置|HeroUI")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建武将表现配置Map
	heroUIDiff := make([]HeroUIDiff, 0, 300)

	for i, idStr := range heroUICols[Id][startRow:helpers.AutoDetectEndIndex(heroUICols, Id, startRow, 3)] {
		// 第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			heroUIDiffItem := HeroUIDiff{}

			name := utils.GetCellValue(heroUICols, Name, startRow+i)
			voiceType := utils.GetCellValue(heroUICols, VoiceType, startRow+i)
			belongExpansionPack := utils.GetCellValue(heroUICols, BelongExpansionPack, startRow+i)
			awardDes := utils.GetCellValue(heroUICols, AwardDes, startRow+i)
			longIntroduction := utils.GetCellValue(heroUICols, LongIntroduction, startRow+i)
			shortIntroduction := utils.GetCellValue(heroUICols, ShortIntroduction, startRow+i)
			evidence := utils.GetCellValue(heroUICols, Evidence, startRow+i)
			evaluation := utils.GetCellValue(heroUICols, Evaluation, startRow+i)
			isNew := utils.GetCellValue(heroUICols, IsNew, startRow+i)
			copyWriter := utils.GetCellValue(heroUICols, CopyWriter, startRow+i)
			skillDesigner := utils.GetCellValue(heroUICols, SkillDesigner, startRow+i)
			getWay := utils.GetCellValue(heroUICols, GetWay, startRow+i)
			position := utils.GetCellValue(heroUICols, Position, startRow+i)
			exclusiveCard := utils.GetCellValue(heroUICols, ExclusiveCard, startRow+i)
			newbieShowSkillTag := utils.GetCellValue(heroUICols, NewbieShowSkillTag, startRow+i)
			winningRateIn2v2 := utils.GetCellValue(heroUICols, WinningRateIn2v2, startRow+i)
			winRateShowPriority := utils.GetCellValue(heroUICols, WinRateShowPriority, startRow+i)

			// 武将信息录入
			heroUIDiffItem.Id = id
			heroUIDiffItem.Name = name

			if n, err := strconv.Atoi(voiceType); err == nil {
				heroUIDiffItem.VoiceType = n
			} else {
				heroUIDiffItem.VoiceType = -1
			}

			heroUIDiffItem.BelongExpansionPack = belongExpansionPack
			heroUIDiffItem.AwardDes = awardDes
			heroUIDiffItem.LongIntroduction = longIntroduction
			heroUIDiffItem.ShortIntroduction = shortIntroduction
			heroUIDiffItem.Evidence = evidence
			heroUIDiffItem.Evaluation = evaluation

			if b, err := strconv.ParseBool(isNew); err == nil && b {
				heroUIDiffItem.IsNew = b
			} else {
				heroUIDiffItem.IsNew = false
			}

			heroUIDiffItem.CopyWriter = copyWriter
			heroUIDiffItem.SkillDesigner = skillDesigner
			heroUIDiffItem.GetWay = getWay
			heroUIDiffItem.Position = position

			// 处理数组字段
			heroUIDiffItem.ExclusiveCard = make([]int, 0, 5)
			for _, ss := range strings.Split(exclusiveCard, ",") {
				if ss != "" {
					if n, err := strconv.Atoi(ss); err == nil {
						heroUIDiffItem.ExclusiveCard = append(heroUIDiffItem.ExclusiveCard, n)
					}
				}
			}

			heroUIDiffItem.NewbieShowSkillTag = make([]int, 0, 5)
			for _, ss := range strings.Split(newbieShowSkillTag, ",") {
				if ss != "" {
					if n, err := strconv.Atoi(ss); err == nil {
						heroUIDiffItem.NewbieShowSkillTag = append(heroUIDiffItem.NewbieShowSkillTag, n)
					}
				}
			}

			if n, err := strconv.Atoi(winningRateIn2v2); err == nil {
				heroUIDiffItem.WinningRateIn2v2 = n
			} else {
				heroUIDiffItem.WinningRateIn2v2 = -1
			}

			if n, err := strconv.Atoi(winRateShowPriority); err == nil {
				heroUIDiffItem.WinRateShowPriority = n
			} else {
				heroUIDiffItem.WinRateShowPriority = -1
			}

			heroUIDiff = append(heroUIDiff, heroUIDiffItem)
		}
	}

	return &heroUIDiff, nil
}
