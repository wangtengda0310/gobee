package country

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type CountryDiff struct {
	ECountry                  string
	Name                      string
	KingName                  string
	SeatFlagImagePath         string
	PlayerInfoFlagImagePath   string
	PlayerInfoFlagTextPath    string
	IsWhiteBg                 bool
	IsOpen                    bool
	NameArtSelectedFontPath   string
	NameArtUnselectedFontPath string
	ShiLiSmallTopIcon         string
	ShiLiSmallBottomIcon      string
	SortPriority              int
	GuildIcon                 string
}

func (c CountryDiff) GetID() string {
	return c.ECountry
}

func (c CountryDiff) GetType() string {
	return "CountryDiff"
}

func (c CountryDiff) GetDisplayName() string {
	return c.Name
}

func GetCountryDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]CountryDiff, err error) {
	var countryCols [][]string // 势力表
	if sheet, exist := sheetMap["势力表|Country"]; exist {
		var err error
		countryCols, err = sheet.GetCols("势力表|Country")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建势力Map
	countriesDiff := make([]CountryDiff, 0, 50)

	for i, idStr := range countryCols[ECountry][startRow:helpers.AutoDetectEndIndex(countryCols, ECountry, startRow, 3)] {
		// 第一列是ECountry枚举类型，判断规则：不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		countryDiff := CountryDiff{}

		// 获取所有字段的值
		eCountry := idStr
		name := utils.GetCellValue(countryCols, Name, startRow+i)
		kingName := utils.GetCellValue(countryCols, KingName, startRow+i)
		seatFlagImagePath := utils.GetCellValue(countryCols, SeatFlagImagePath, startRow+i)
		playerInfoFlagImagePath := utils.GetCellValue(countryCols, PlayerInfoFlagImagePath, startRow+i)
		playerInfoFlagTextPath := utils.GetCellValue(countryCols, PlayerInfoFlagTextPath, startRow+i)
		isWhiteBg := utils.GetCellValue(countryCols, IsWhiteBg, startRow+i)
		isOpen := utils.GetCellValue(countryCols, IsOpen, startRow+i)
		nameArtSelectedFontPath := utils.GetCellValue(countryCols, NameArtSelectedFontPath, startRow+i)
		nameArtUnselectedFontPath := utils.GetCellValue(countryCols, NameArtUnselectedFontPath, startRow+i)
		shiLiSmallTopIcon := utils.GetCellValue(countryCols, ShiLiSmallTopIcon, startRow+i)
		shiLiSmallBottomIcon := utils.GetCellValue(countryCols, ShiLiSmallBottomIcon, startRow+i)
		sortPriority := utils.GetCellValue(countryCols, SortPriority, startRow+i)
		guildIcon := utils.GetCellValue(countryCols, GuildIcon, startRow+i)

		// 录入势力信息
		countryDiff.ECountry = eCountry
		countryDiff.Name = name
		countryDiff.KingName = kingName
		countryDiff.SeatFlagImagePath = seatFlagImagePath
		countryDiff.PlayerInfoFlagImagePath = playerInfoFlagImagePath
		countryDiff.PlayerInfoFlagTextPath = playerInfoFlagTextPath

		if b, err := strconv.ParseBool(isWhiteBg); err == nil {
			countryDiff.IsWhiteBg = b
		} else {
			countryDiff.IsWhiteBg = false
		}

		if b, err := strconv.ParseBool(isOpen); err == nil {
			countryDiff.IsOpen = b
		} else {
			countryDiff.IsOpen = false
		}

		countryDiff.NameArtSelectedFontPath = nameArtSelectedFontPath
		countryDiff.NameArtUnselectedFontPath = nameArtUnselectedFontPath
		countryDiff.ShiLiSmallTopIcon = shiLiSmallTopIcon
		countryDiff.ShiLiSmallBottomIcon = shiLiSmallBottomIcon

		if n, err := strconv.Atoi(sortPriority); err == nil {
			countryDiff.SortPriority = n
		} else {
			countryDiff.SortPriority = -1
		}

		countryDiff.GuildIcon = guildIcon

		countriesDiff = append(countriesDiff, countryDiff)
	}

	return &countriesDiff, nil
}
