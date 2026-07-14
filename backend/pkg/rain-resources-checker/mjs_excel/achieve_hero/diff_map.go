package achieve_hero

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroAchieveDiff struct {
	Id           string   // 局内成就类型 (E#HeroAchieve)
	Name         string   // 成就名称
	IsMult       bool     // 是否复用
	Mode         []int    // 房间模式
	MinPlayerNum int      // 房间人数
	UseHero      []string // 使用英雄 (E#HeroId列表)
	Class        int      // 身份类型
	Camp         int      // 阵营
	Identity     int      // 身份
	Hooker       []int    // 对应钩子
	HookerTarget []int    // 对应钩子
	CondParam    []int    // 条件一
}

func (h HeroAchieveDiff) GetID() string {
	return h.Id
}

func (h HeroAchieveDiff) GetType() string {
	return "HeroAchieveDiff"
}

func (h HeroAchieveDiff) GetDisplayName() string {
	return h.Name
}

func GetHeroAchieveDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroAchieveDiff, err error) {
	var achieveCols [][]string // 角色成就表
	if sheet, exist := sheetMap["角色成就|HeroAchieve"]; exist {
		var err error
		achieveCols, err = sheet.GetCols("角色成就|HeroAchieve")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建成就Map
	achievesDiff := make([]HeroAchieveDiff, 0, 100)

	for i, idStr := range achieveCols[Id][startRow:helpers.AutoDetectEndIndex(achieveCols, Id, startRow, 3)] {
		// 第一列是E#开头的枚举类型，判断规则：不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		achieveDiff := HeroAchieveDiff{}

		name := utils.GetCellValue(achieveCols, Name, startRow+i)
		isMult := utils.GetCellValue(achieveCols, IsMult, startRow+i)
		mode := utils.GetCellValue(achieveCols, Mode, startRow+i)
		minPlayerNum := utils.GetCellValue(achieveCols, MinPlayerNum, startRow+i)
		useHero := utils.GetCellValue(achieveCols, UseHero, startRow+i)
		class := utils.GetCellValue(achieveCols, Class, startRow+i)
		camp := utils.GetCellValue(achieveCols, Camp, startRow+i)
		identity := utils.GetCellValue(achieveCols, Identity, startRow+i)
		hooker := utils.GetCellValue(achieveCols, Hooker, startRow+i)
		hookerTarget := utils.GetCellValue(achieveCols, HookerTarget, startRow+i)
		condParam := utils.GetCellValue(achieveCols, CondParam, startRow+i)

		// 录入成就信息
		achieveDiff.Id = idStr
		achieveDiff.Name = name

		if b, err := strconv.ParseBool(isMult); err == nil && b {
			achieveDiff.IsMult = b
		} else {
			achieveDiff.IsMult = false
		}

		// 房间模式（数组）
		achieveDiff.Mode = make([]int, 0, 5)
		for _, ss := range strings.Split(mode, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					achieveDiff.Mode = append(achieveDiff.Mode, n)
				}
			}
		}

		// 房间人数
		if n, err := strconv.Atoi(minPlayerNum); err == nil {
			achieveDiff.MinPlayerNum = n
		} else {
			achieveDiff.MinPlayerNum = -1
		}

		// 使用英雄（E#HeroId数组）
		achieveDiff.UseHero = make([]string, 0, 5)
		for _, ss := range strings.Split(useHero, ",") {
			if ss != "" {
				achieveDiff.UseHero = append(achieveDiff.UseHero, ss)
			}
		}

		// 身份类型
		if n, err := strconv.Atoi(class); err == nil {
			achieveDiff.Class = n
		} else {
			achieveDiff.Class = -1
		}

		// 阵营
		if n, err := strconv.Atoi(camp); err == nil {
			achieveDiff.Camp = n
		} else {
			achieveDiff.Camp = -1
		}

		// 身份
		if n, err := strconv.Atoi(identity); err == nil {
			achieveDiff.Identity = n
		} else {
			achieveDiff.Identity = -1
		}

		// 对应钩子（数组）
		achieveDiff.Hooker = make([]int, 0, 5)
		for _, ss := range strings.Split(hooker, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					achieveDiff.Hooker = append(achieveDiff.Hooker, n)
				}
			}
		}

		// 对应钩子（数组）
		achieveDiff.HookerTarget = make([]int, 0, 5)
		for _, ss := range strings.Split(hookerTarget, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					achieveDiff.HookerTarget = append(achieveDiff.HookerTarget, n)
				}
			}
		}

		// 条件一（数组）
		achieveDiff.CondParam = make([]int, 0, 5)
		for _, ss := range strings.Split(condParam, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					achieveDiff.CondParam = append(achieveDiff.CondParam, n)
				}
			}
		}

		achievesDiff = append(achievesDiff, achieveDiff)
	}

	return &achievesDiff, nil
}
