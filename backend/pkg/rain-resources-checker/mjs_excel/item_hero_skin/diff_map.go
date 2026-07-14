package item_hero_skin

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroSkinDiff struct {
	SkinItemId int
	None       string
	Path       string
}

func (h HeroSkinDiff) GetType() string {
	return "HeroSkinDiff"
}

func (h HeroSkinDiff) GetDisplayName() string {
	return strconv.Itoa(h.SkinItemId)
}

func GetHeroSkinDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroSkinDiff, err error) {
	var skinCols [][]string // 武将皮肤展示表
	if sheet, exist := sheetMap["武将皮肤展示表|ItemHeroSkin"]; exist {
		var err error
		skinCols, err = sheet.GetCols("武将皮肤展示表|ItemHeroSkin")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建皮肤Map
	skinsDiff := make([]HeroSkinDiff, 0, 300)

	for i, idStr := range skinCols[SkinItemId][startRow:helpers.AutoDetectEndIndex(skinCols, SkinItemId, startRow, 3)] {
		// 第一列是int类型，通过是否能转为int来判断是否有效行
		if skinItemId, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			skinDiff := HeroSkinDiff{}

			none := utils.GetCellValue(skinCols, None, startRow+i)
			path := utils.GetCellValue(skinCols, Path, startRow+i)

			skinDiff.SkinItemId = skinItemId
			skinDiff.None = none
			skinDiff.Path = path

			skinsDiff = append(skinsDiff, skinDiff)
		}
	}

	return &skinsDiff, nil
}
