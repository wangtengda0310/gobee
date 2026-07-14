package recommend_bd

import (
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type RecommendBdDiff struct {
	HeroId string   // 武将ID
	Level1 []string // 壹地1_4 (KVPair[])
	Level2 []string // 贰北2_8 (KVPair[])
	Level3 []string // 叁西3_12 (KVPair[])
	Level4 []string // 肆南4_16 (KVPair[])
	Level5 []string // 伍东5_20 (KVPair[])
	Level6 []string // 陆天6_24 (KVPair[])
}

func (r RecommendBdDiff) GetID() string {
	return r.HeroId
}

func (r RecommendBdDiff) GetType() string {
	return "RecommendBdDiff"
}

func (r RecommendBdDiff) GetDisplayName() string {
	return r.HeroId
}

func GetRecommendBdDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]RecommendBdDiff, err error) {
	var recommendBdCols [][]string
	if sheet, exist := sheetMap["推荐加点表|RecommendBd"]; exist {
		var err error
		recommendBdCols, err = sheet.GetCols("推荐加点表|RecommendBd")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建推荐加点Map
	recommendBdDiff := make([]RecommendBdDiff, 0, 300)

	for i, heroIdStr := range recommendBdCols[HeroId][startRow:helpers.AutoDetectEndIndex(recommendBdCols, HeroId, startRow, 3)] {
		// 第一列是HeroId(string类型)，判断规则：不为空且不以#开头
		if heroIdStr == "" || strings.HasPrefix(heroIdStr, "#") {
			continue
		}

		recommendBd := RecommendBdDiff{}

		// 读取各列数据
		level1 := utils.GetCellValue(recommendBdCols, Level1, startRow+i)
		level2 := utils.GetCellValue(recommendBdCols, Level2, startRow+i)
		level3 := utils.GetCellValue(recommendBdCols, Level3, startRow+i)
		level4 := utils.GetCellValue(recommendBdCols, Level4, startRow+i)
		level5 := utils.GetCellValue(recommendBdCols, Level5, startRow+i)
		level6 := utils.GetCellValue(recommendBdCols, Level6, startRow+i)

		// 填充数据
		recommendBd.HeroId = heroIdStr

		// KVPair[]类型，按逗号分割成字符串数组
		recommendBd.Level1 = parseKVPair(level1)
		recommendBd.Level2 = parseKVPair(level2)
		recommendBd.Level3 = parseKVPair(level3)
		recommendBd.Level4 = parseKVPair(level4)
		recommendBd.Level5 = parseKVPair(level5)
		recommendBd.Level6 = parseKVPair(level6)

		recommendBdDiff = append(recommendBdDiff, recommendBd)
	}

	return &recommendBdDiff, nil
}

// parseKVPair 解析KVPair类型的数据，按逗号分割字符串
func parseKVPair(data string) []string {
	if data == "" {
		return []string{}
	}
	return strings.Split(data, ",")
}
