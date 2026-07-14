package hero_skin_collition

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroSkinCollectionDiff struct {
	Type     string // ESkinCollitionType
	Name     string // 收藏册名称
	NameImg  string // 名称图片
	NameBg   string // 名称底图图片
	Desc     string // 诗词文案
	Weight   int    // 排序权重
	OpenDate string // 开放时间
}

func (h HeroSkinCollectionDiff) GetID() string {
	return h.Type
}

func (h HeroSkinCollectionDiff) GetType() string {
	return "HeroSkinCollectionDiff"
}

func (h HeroSkinCollectionDiff) GetDisplayName() string {
	return h.Name
}

func GetHeroSkinCollectionDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroSkinCollectionDiff, err error) {
	var skinCols [][]string // 英雄皮肤收藏表
	if sheet, exist := sheetMap["英雄皮肤收藏|HeroSkinCollition"]; exist {
		var err error
		skinCols, err = sheet.GetCols("英雄皮肤收藏|HeroSkinCollition")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建皮肤收藏Map
	skinsDiff := make([]HeroSkinCollectionDiff, 0, 100)

	// 第一列是ESkinCollitionType（枚举类型），判断规则：不为空且不以#开头
	for i, typeStr := range skinCols[Type][startRow:helpers.AutoDetectEndIndex(skinCols, Type, startRow, 3)] {
		// 跳过空行和注释行（以#开头）
		if typeStr == "" || strings.HasPrefix(typeStr, "#") {
			continue
		}

		skinDiff := HeroSkinCollectionDiff{}

		// 获取各列数据
		name := utils.GetCellValue(skinCols, Name, startRow+i)
		nameImg := utils.GetCellValue(skinCols, NameImg, startRow+i)
		nameBg := utils.GetCellValue(skinCols, NameBg, startRow+i)
		desc := utils.GetCellValue(skinCols, Desc, startRow+i)
		weight := utils.GetCellValue(skinCols, Weight, startRow+i)
		openDate := utils.GetCellValue(skinCols, OpenDate, startRow+i)

		// 赋值
		skinDiff.Type = typeStr
		skinDiff.Name = name
		skinDiff.NameImg = nameImg
		skinDiff.NameBg = nameBg
		skinDiff.Desc = desc

		// Weight转换为int
		if n, err := strconv.Atoi(weight); err == nil {
			skinDiff.Weight = n
		} else {
			skinDiff.Weight = -1
		}

		skinDiff.OpenDate = openDate

		skinsDiff = append(skinsDiff, skinDiff)
	}

	return &skinsDiff, nil
}
