package hero_skin_item

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroSkinItemDiff struct {
	SkinItemId          int
	HeroId              string
	None1               string
	RailyType           string
	SkinType            string
	Name                string
	GetWay              string
	SkinPinYin          string
	SeatSpecialImg      string
	HeroUIExtraIcons    []string
	Lines               []int
	DebutLines          []int
	KillLines           []int
	DeadLines           []int
	HeroAudio           []int
	LinesDubbed         string
	OriginalArtDesigner string
	CollitionType       string
	IsOpen              bool
	OpenDate            string
	CollitionTagImg     string
	KillShowTime        int
	BodyOffset          []int
	HasOutFrameIcon     bool
}

func (h HeroSkinItemDiff) GetID() string {
	return h.HeroId + "_" + strconv.Itoa(h.SkinItemId)
}

func (h HeroSkinItemDiff) GetType() string {
	return "HeroSkinItemDiff"
}

func (h HeroSkinItemDiff) GetDisplayName() string {
	return h.Name
}

func GetHeroSkinItemDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroSkinItemDiff, err error) {
	var skinCols [][]string // 英雄皮肤表
	if sheet, exist := sheetMap["英雄皮肤|HeroSkinItem"]; exist {
		var err error
		skinCols, err = sheet.GetCols("英雄皮肤|HeroSkinItem")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建皮肤Map
	skinsDiff := make([]HeroSkinItemDiff, 0, 500)

	for i, skinItemIdStr := range skinCols[SkinItemId][startRow:helpers.AutoDetectEndIndex(skinCols, SkinItemId, startRow, 3)] {
		// 第一列是int类型的SkinItemId，需要判断是否为有效数字
		if skinItemId, err := strconv.Atoi(skinItemIdStr); err != nil {
			continue
		} else {
			skinDiff := HeroSkinItemDiff{}

			// 获取所有字段的值
			heroId := utils.GetCellValue(skinCols, HeroId, startRow+i)
			none1 := utils.GetCellValue(skinCols, None, startRow+i)
			railyType := utils.GetCellValue(skinCols, RailyType, startRow+i)
			skinType := utils.GetCellValue(skinCols, SkinType, startRow+i)
			name := utils.GetCellValue(skinCols, Name, startRow+i)
			getWay := utils.GetCellValue(skinCols, GetWay, startRow+i)
			skinPinYin := utils.GetCellValue(skinCols, SkinPinYin, startRow+i)
			seatSpecialImg := utils.GetCellValue(skinCols, SeatSpecialImg, startRow+i)
			heroUIExtraIcons := utils.GetCellValue(skinCols, HeroUIExtraIcons, startRow+i)
			lines := utils.GetCellValue(skinCols, Lines, startRow+i)
			debutLines := utils.GetCellValue(skinCols, DebutLines, startRow+i)
			killLines := utils.GetCellValue(skinCols, KillLines, startRow+i)
			deadLines := utils.GetCellValue(skinCols, DeadLines, startRow+i)
			heroAudio := utils.GetCellValue(skinCols, HeroAudio, startRow+i)
			linesDubbed := utils.GetCellValue(skinCols, LinesDubbed, startRow+i)
			originalArtDesigner := utils.GetCellValue(skinCols, OriginalArtDesigner, startRow+i)
			collitionType := utils.GetCellValue(skinCols, CollitionType, startRow+i)
			isOpen := utils.GetCellValue(skinCols, IsOpen, startRow+i)
			openDate := utils.GetCellValue(skinCols, OpenDate, startRow+i)
			collitionTagImg := utils.GetCellValue(skinCols, CollitionTagImg, startRow+i)
			killShowTime := utils.GetCellValue(skinCols, KillShowTime, startRow+i)
			bodyOffset := utils.GetCellValue(skinCols, BodyOffset, startRow+i)
			hasOutFrameIcon := utils.GetCellValue(skinCols, HasOutFrameIcon, startRow+i)

			// 填充数据
			skinDiff.SkinItemId = skinItemId
			skinDiff.HeroId = heroId
			skinDiff.None1 = none1
			skinDiff.RailyType = railyType
			skinDiff.SkinType = skinType
			skinDiff.Name = name
			skinDiff.GetWay = getWay
			skinDiff.SkinPinYin = skinPinYin
			skinDiff.SeatSpecialImg = seatSpecialImg

			// HeroUIExtraIcons 是 []string 类型
			skinDiff.HeroUIExtraIcons = make([]string, 0, 5)
			for _, ss := range strings.Split(heroUIExtraIcons, ",") {
				if ss != "" {
					skinDiff.HeroUIExtraIcons = append(skinDiff.HeroUIExtraIcons, ss)
				}
			}

			// 台词类字段都是 []int 类型
			skinDiff.Lines = make([]int, 0, 10)
			for _, ss := range strings.Split(lines, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					skinDiff.Lines = append(skinDiff.Lines, n)
				}
			}

			skinDiff.DebutLines = make([]int, 0, 5)
			for _, ss := range strings.Split(debutLines, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					skinDiff.DebutLines = append(skinDiff.DebutLines, n)
				}
			}

			skinDiff.KillLines = make([]int, 0, 5)
			for _, ss := range strings.Split(killLines, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					skinDiff.KillLines = append(skinDiff.KillLines, n)
				}
			}

			skinDiff.DeadLines = make([]int, 0, 5)
			for _, ss := range strings.Split(deadLines, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					skinDiff.DeadLines = append(skinDiff.DeadLines, n)
				}
			}

			skinDiff.HeroAudio = make([]int, 0, 5)
			for _, ss := range strings.Split(heroAudio, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					skinDiff.HeroAudio = append(skinDiff.HeroAudio, n)
				}
			}

			skinDiff.LinesDubbed = linesDubbed
			skinDiff.OriginalArtDesigner = originalArtDesigner
			skinDiff.CollitionType = collitionType

			if b, err := strconv.ParseBool(isOpen); err == nil {
				skinDiff.IsOpen = b
			} else {
				skinDiff.IsOpen = false
			}

			skinDiff.OpenDate = openDate
			skinDiff.CollitionTagImg = collitionTagImg

			if n, err := strconv.Atoi(killShowTime); err == nil {
				skinDiff.KillShowTime = n
			} else {
				skinDiff.KillShowTime = -1
			}

			skinDiff.BodyOffset = make([]int, 0, 5)
			for _, ss := range strings.Split(bodyOffset, ",") {
				if n, err := strconv.Atoi(ss); err == nil {
					skinDiff.BodyOffset = append(skinDiff.BodyOffset, n)
				}
			}

			if b, err := strconv.ParseBool(hasOutFrameIcon); err == nil {
				skinDiff.HasOutFrameIcon = b
			} else {
				skinDiff.HasOutFrameIcon = false
			}

			skinsDiff = append(skinsDiff, skinDiff)
		}
	}

	return &skinsDiff, nil
}
