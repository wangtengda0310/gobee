package hero

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroDiff struct {
	Id                  int // EHeroId
	Name                string
	EHeroId             string // EHeroId枚举定义列
	IsOpen              bool
	OpenDate            string
	Gender              int
	Point               int
	HpLimit             int
	HandLimit           int
	EquipLimit          int
	Country             string
	IsAlwaysZhuGong     bool
	Skill               []int
	ExcludeIdentity     []int
	NotUseModeType      []int
	HeroType            int
	EHeroType           string
	CanMelt             bool
	MeltName            []string
	IsNewHero           bool
	IsGacha             bool
	BelongExpansionPack string
}

func (h HeroDiff) GetType() string {
	return "HeroDiff"
}

func (h HeroDiff) GetDisplayName() string {
	return h.Name
}

func GetHeroDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroDiff, err error) {
	var heroCols [][]string // 武将表
	if sheet, exist := sheetMap["武将表|Hero"]; exist {
		var err error
		heroCols, err = sheet.GetCols("武将表|Hero")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建开放武将Map
	herosDiff := make([]HeroDiff, 0, 300)

	for i, idStr := range heroCols[Id][startRow:helpers.AutoDetectEndIndex(heroCols, Id, startRow, 3)] {
		// 表格第一列是否有id作为读取这一列的标准
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			heroDiff := HeroDiff{}

			name := utils.GetCellValue(heroCols, Name, startRow+i)
			eHeroId := utils.GetCellValue(heroCols, EHeroId, startRow+i)
			isOpen := utils.GetCellValue(heroCols, IsOpen, startRow+i)
			openDate := utils.GetCellValue(heroCols, OpenDate, startRow+i)
			gender := utils.GetCellValue(heroCols, Gender, startRow+i)
			point := utils.GetCellValue(heroCols, Point, startRow+i)
			hpLimit := utils.GetCellValue(heroCols, HpLimit, startRow+i)
			handLimit := utils.GetCellValue(heroCols, HandLimit, startRow+i)
			equipLimit := utils.GetCellValue(heroCols, EquipLimit, startRow+i)
			country := utils.GetCellValue(heroCols, Country, startRow+i)
			isAlwaysZhuGong := utils.GetCellValue(heroCols, IsAlwaysZhuGong, startRow+i)
			skill := utils.GetCellValue(heroCols, Skill, startRow+i)
			excludeIdentity := utils.GetCellValue(heroCols, ExcludeIdentity, startRow+i)
			notUseModeType := utils.GetCellValue(heroCols, NotUseModeType, startRow+i)
			heroType := utils.GetCellValue(heroCols, HeroType, startRow+i)
			eHeroType := utils.GetCellValue(heroCols, EHeroType, startRow+i)
			canMelt := utils.GetCellValue(heroCols, CanMelt, startRow+i)
			meltName := utils.GetCellValue(heroCols, MeltName, startRow+i)
			isNewHero := utils.GetCellValue(heroCols, IsNewHero, startRow+i)
			isGacha := utils.GetCellValue(heroCols, IsGacha, startRow+i)
			belongExpansionPack := utils.GetCellValue(heroCols, BelongExpansionPack, startRow+i)

			// 特殊，排除未开放且非Normal类型的武将
			if isOpen != "1" || heroType != "1" {
				continue
			}

			// 其他武将录入信息
			heroDiff.Id = id
			heroDiff.Name = name
			heroDiff.EHeroId = eHeroId
			heroDiff.IsOpen = true
			heroDiff.OpenDate = openDate
			if n, err := strconv.Atoi(gender); err == nil {
				heroDiff.Gender = n
			} else {
				heroDiff.Gender = -1
			}
			if n, err := strconv.Atoi(point); err == nil {
				heroDiff.Point = n
			} else {
				heroDiff.Point = -1
			}
			if n, err := strconv.Atoi(hpLimit); err == nil {
				heroDiff.HpLimit = n
			} else {
				heroDiff.HpLimit = -1
			}
			if n, err := strconv.Atoi(handLimit); err == nil {
				heroDiff.HandLimit = n
			} else {
				heroDiff.HandLimit = -1
			}
			if n, err := strconv.Atoi(equipLimit); err == nil {
				heroDiff.EquipLimit = n
			} else {
				heroDiff.EquipLimit = -1
			}
			heroDiff.Country = country
			if b, err := strconv.ParseBool(isAlwaysZhuGong); err == nil && b {
				heroDiff.IsAlwaysZhuGong = b
			} else {
				heroDiff.IsAlwaysZhuGong = false
			}
			heroDiff.Skill = make([]int, 0, 5)
			for _, ss := range strings.Split(skill, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					heroDiff.Skill = append(heroDiff.Skill, n)
				}
			}
			heroDiff.ExcludeIdentity = make([]int, 0, 10)
			for _, ss := range strings.Split(excludeIdentity, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					heroDiff.ExcludeIdentity = append(heroDiff.ExcludeIdentity, n)
				}
			}
			heroDiff.NotUseModeType = make([]int, 0, 10)
			for _, ss := range strings.Split(notUseModeType, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					heroDiff.NotUseModeType = append(heroDiff.NotUseModeType, n)
				}
			}
			if n, err := strconv.Atoi(heroType); err == nil {
				heroDiff.HeroType = n
			} else {
				heroDiff.HeroType = -1
			}
			heroDiff.EHeroType = eHeroType
			if b, err := strconv.ParseBool(canMelt); err == nil && b {
				heroDiff.CanMelt = b
			} else {
				heroDiff.CanMelt = false
			}
			heroDiff.MeltName = strings.Split(meltName, ",")
			if b, err := strconv.ParseBool(isNewHero); err == nil && b {
				heroDiff.IsGacha = b
			} else {
				heroDiff.IsGacha = false
			}
			if b, err := strconv.ParseBool(isGacha); err == nil && b {
				heroDiff.IsGacha = b
			} else {
				heroDiff.IsGacha = false
			}
			heroDiff.BelongExpansionPack = belongExpansionPack

			herosDiff = append(herosDiff, heroDiff)
		}
	}

	return &herosDiff, nil
}
