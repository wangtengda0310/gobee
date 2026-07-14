package hero_skin_spine

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type HeroSkinSpineDiff struct {
	SkinItemId                         int
	None1                              string
	IsHasSeatSpine                     bool
	IsHasBookSpine                     bool
	IsHasMainBgSpine                   bool
	MainBgFx                           string
	IsHasSeatKillSpine                 bool
	KillFxId                           int
	IsHasCollitionBgSpine              bool
	IsHasCollitionCardBgSpine          bool
	SpineAnimAudio                     string
	IsHasCollitionCardBgSpineDuplicate bool
	KillAudio                          string
}

func (h HeroSkinSpineDiff) GetType() string {
	return "HeroSkinSpineDiff"
}

func (h HeroSkinSpineDiff) GetDisplayName() string {
	return strconv.Itoa(h.SkinItemId)
}

func GetHeroSkinSpineDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]HeroSkinSpineDiff, err error) {
	var skinSpineCols [][]string
	if sheet, exist := sheetMap["英雄皮肤Spine|HeroSkinSpine"]; exist {
		var err error
		skinSpineCols, err = sheet.GetCols("英雄皮肤Spine|HeroSkinSpine")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建皮肤Spine Map
	skinSpinesDiff := make([]HeroSkinSpineDiff, 0, 300)

	for i, idStr := range skinSpineCols[SkinItemId][startRow:helpers.AutoDetectEndIndex(skinSpineCols, SkinItemId, startRow, 3)] {
		// 第一列是int类型，用strconv.Atoi判断有效行
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			skinSpineDiff := HeroSkinSpineDiff{}

			none1 := utils.GetCellValue(skinSpineCols, None, startRow+i)
			isHasSeatSpine := utils.GetCellValue(skinSpineCols, IsHasSeatSpine, startRow+i)
			isHasBookSpine := utils.GetCellValue(skinSpineCols, IsHasBookSpine, startRow+i)
			isHasMainBgSpine := utils.GetCellValue(skinSpineCols, IsHasMainBgSpine, startRow+i)
			mainBgFx := utils.GetCellValue(skinSpineCols, MainBgFx, startRow+i)
			isHasSeatKillSpine := utils.GetCellValue(skinSpineCols, IsHasSeatKillSpine, startRow+i)
			killFxId := utils.GetCellValue(skinSpineCols, KillFxId, startRow+i)
			isHasCollitionBgSpine := utils.GetCellValue(skinSpineCols, IsHasCollitionBgSpine, startRow+i)
			isHasCollitionCardBgSpine := utils.GetCellValue(skinSpineCols, IsHasCollitionCardBgSpine, startRow+i)
			spineAnimAudio := utils.GetCellValue(skinSpineCols, SpineAnimAudio, startRow+i)
			isHasCollitionCardBgSpineDuplicate := utils.GetCellValue(skinSpineCols, IsHasCollitionCardBgSpineDuplicate, startRow+i)
			killAudio := utils.GetCellValue(skinSpineCols, KillAudio, startRow+i)

			// 录入皮肤Spine信息
			skinSpineDiff.SkinItemId = id
			skinSpineDiff.None1 = none1

			if b, err := strconv.ParseBool(isHasSeatSpine); err == nil {
				skinSpineDiff.IsHasSeatSpine = b
			} else {
				skinSpineDiff.IsHasSeatSpine = false
			}

			if b, err := strconv.ParseBool(isHasBookSpine); err == nil {
				skinSpineDiff.IsHasBookSpine = b
			} else {
				skinSpineDiff.IsHasBookSpine = false
			}

			if b, err := strconv.ParseBool(isHasMainBgSpine); err == nil {
				skinSpineDiff.IsHasMainBgSpine = b
			} else {
				skinSpineDiff.IsHasMainBgSpine = false
			}

			skinSpineDiff.MainBgFx = mainBgFx

			if b, err := strconv.ParseBool(isHasSeatKillSpine); err == nil {
				skinSpineDiff.IsHasSeatKillSpine = b
			} else {
				skinSpineDiff.IsHasSeatKillSpine = false
			}

			if n, err := strconv.Atoi(killFxId); err == nil {
				skinSpineDiff.KillFxId = n
			} else {
				skinSpineDiff.KillFxId = -1
			}

			if b, err := strconv.ParseBool(isHasCollitionBgSpine); err == nil {
				skinSpineDiff.IsHasCollitionBgSpine = b
			} else {
				skinSpineDiff.IsHasCollitionBgSpine = false
			}

			if b, err := strconv.ParseBool(isHasCollitionCardBgSpine); err == nil {
				skinSpineDiff.IsHasCollitionCardBgSpine = b
			} else {
				skinSpineDiff.IsHasCollitionCardBgSpine = false
			}

			skinSpineDiff.SpineAnimAudio = spineAnimAudio

			if b, err := strconv.ParseBool(isHasCollitionCardBgSpineDuplicate); err == nil {
				skinSpineDiff.IsHasCollitionCardBgSpineDuplicate = b
			} else {
				skinSpineDiff.IsHasCollitionCardBgSpineDuplicate = false
			}

			skinSpineDiff.KillAudio = killAudio

			skinSpinesDiff = append(skinSpinesDiff, skinSpineDiff)
		}
	}

	return &skinSpinesDiff, nil
}
