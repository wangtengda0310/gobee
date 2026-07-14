package hero_skin_spine

import (
	"regexp"
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// GetHeroSkinSpineVoiceMap
// skinIdSpineAnimAudioMap 皮肤道具ID=>动画音效枚举
// skinIdKillAudioMap 皮肤道具ID=>击杀动画音效
func GetHeroSkinSpineVoiceMap(sheetMap map[string]*excelize.File) (skinIdSpineAnimAudioMap map[int][]string, skinIdKillAudioMap map[int]string, err error) {
	var heroSkinSpineCols [][]string // 技能台词表
	if sheet, exist := sheetMap["英雄皮肤Spine|HeroSkinSpine"]; exist {
		var err error
		heroSkinSpineCols, err = sheet.GetCols("英雄皮肤Spine|HeroSkinSpine")
		if err != nil {
			return nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建音频Map
	skinIdSpineAnimAudioMap = make(map[int][]string)
	skinIdKillAudioMap = make(map[int]string)

	// 需要用到的正则
	var extractPattern *regexp.Regexp
	extractPattern, err = regexp.Compile("\\{\\d+;([^}]+)}")
	if err != nil {
		return nil, nil, err
	}

	for i, idStr := range heroSkinSpineCols[SkinItemId][startRow:helpers.AutoDetectEndIndex(heroSkinSpineCols, SkinItemId, startRow, 3)] {
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {

			saa := utils.GetCellValue(heroSkinSpineCols, SpineAnimAudio, startRow+i)
			if saa != "" {
				matches := extractPattern.FindAllStringSubmatch(saa, -1)
				aids := make([]string, 0, len(matches))
				for _, match := range matches {
					aids = append(aids, match[1])
				}
				skinIdSpineAnimAudioMap[id] = aids
			}
			ka := utils.GetCellValue(heroSkinSpineCols, KillAudio, startRow+i)
			if ka != "" {
				skinIdKillAudioMap[id] = ka
			}
		}
	}

	return skinIdSpineAnimAudioMap, skinIdKillAudioMap, nil
}
