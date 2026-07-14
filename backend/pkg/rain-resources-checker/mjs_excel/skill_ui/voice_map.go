package skill_ui

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// GetSkillUIMap
// skillEnumPlayCardAudioMap 技能ESkillId枚举=>出牌音效EAudioId
// skillEnumIdentityLineVoiceEnumMap 技能ESkillId枚举=>身份技能台词枚举数组
// skillEnumLineVoiceIdMap 技能ESkillId枚举=>出发技能台词ID数组
// skillEnumSpecialAudioEnumMap 技能ESkillId枚举=>特殊台词枚举
func GetSkillUIMap(sheetMap map[string]*excelize.File) (skillEnumPlayCardAudioEnumMap map[string]string, skillEnumIdentityLineVoiceEnumMap map[string][]string, skillEnumLineVoiceIdMap map[string][]int, skillEnumSpecialAudioEnumMap map[string]string, err error) {
	var skillUICols [][]string // 技能台词表
	if sheet, exist := sheetMap["技能表现配置表|SkillUI"]; exist {
		var err error
		skillUICols, err = sheet.GetCols("技能表现配置表|SkillUI")
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建音频Map
	skillEnumPlayCardAudioEnumMap = make(map[string]string)
	skillEnumIdentityLineVoiceEnumMap = make(map[string][]string)
	skillEnumLineVoiceIdMap = make(map[string][]int)
	skillEnumSpecialAudioEnumMap = make(map[string]string)

	for i, idStr := range skillUICols[Id][startRow:helpers.AutoDetectEndIndex(skillUICols, Id, startRow, 3)] {
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		} else {
			pca := utils.GetCellValue(skillUICols, PlayCardAudio, startRow+i)
			if pca != "" {
				skillEnumPlayCardAudioEnumMap[idStr] = pca
			}
			il := utils.GetCellValue(skillUICols, IdentityLine, startRow+i)
			if il != "" {
				skillEnumIdentityLineVoiceEnumMap[idStr] = strings.Split(il, ",")
			}
			a := utils.GetCellValue(skillUICols, Audio, startRow+i)
			if a != "" {
				vidStrArr := strings.Split(a, ",")
				vids := make([]int, 0, len(vidStrArr))
				for _, vidStr := range vidStrArr {
					if vid, err := strconv.Atoi(vidStr); err == nil {
						vids = append(vids, vid)
					}
				}
				skillEnumLineVoiceIdMap[idStr] = vids
			}
			sa := utils.GetCellValue(skillUICols, SpecialAudio, startRow+i)
			if sa != "" {
				skillEnumSpecialAudioEnumMap[idStr] = sa
			}
		}
	}

	return skillEnumPlayCardAudioEnumMap, skillEnumIdentityLineVoiceEnumMap, skillEnumLineVoiceIdMap, skillEnumSpecialAudioEnumMap, nil
}
