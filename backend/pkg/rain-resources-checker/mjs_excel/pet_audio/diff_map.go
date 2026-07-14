package pet_audio

// activity-wiki-dev: 活动Wiki开发技能生成

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// PetAudioDiff 灵宠音效表差异数据结构
type PetAudioDiff struct {
	AnimationState string
	ItemCfgId      int
	LobbyAudio     string
	PetWindowAudio string
}

func (p PetAudioDiff) GetType() string {
	return "PetAudioDiff"
}

func (p PetAudioDiff) GetDisplayName() string {
	return p.AnimationState
}

// GetPetAudioDiffMap 解析灵宠音效表数据
func GetPetAudioDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]PetAudioDiff, err error) {
	var petAudioCols [][]string
	if sheet, exist := sheetMap["灵宠音效|PetAudio"]; exist {
		var err error
		petAudioCols, err = sheet.GetCols("灵宠音效|PetAudio")
		if err != nil {
			return nil, err
		}
	}

	// 如果没有找到sheet，返回空数组
	if len(petAudioCols) == 0 {
		return &[]PetAudioDiff{}, nil
	}

	startRow := excelio.MJS_FIXED_ROWS_NUM

	petAudiosDiff := make([]PetAudioDiff, 0, 100)

	// 检查AnimationState列是否有数据
	if len(petAudioCols) <= AnimationState || len(petAudioCols[AnimationState]) <= startRow {
		return &petAudiosDiff, nil
	}

	endIndex := helpers.AutoDetectEndIndex(petAudioCols, AnimationState, startRow, 3)
	animationStateCol := petAudioCols[AnimationState][startRow:endIndex]

	for i, animationState := range animationStateCol {
		if animationState == "" {
			continue
		}

		petAudioDiff := PetAudioDiff{}

		itemCfgId := utils.GetCellValue(petAudioCols, ItemCfgId, startRow+i)
		lobbyAudio := utils.GetCellValue(petAudioCols, LobbyAudio, startRow+i)
		petWindowAudio := utils.GetCellValue(petAudioCols, PetWindowAudio, startRow+i)

		petAudioDiff.AnimationState = animationState

		if n, err := strconv.Atoi(itemCfgId); err == nil {
			petAudioDiff.ItemCfgId = n
		} else {
			petAudioDiff.ItemCfgId = -1
		}

		petAudioDiff.LobbyAudio = lobbyAudio
		petAudioDiff.PetWindowAudio = petWindowAudio

		petAudiosDiff = append(petAudiosDiff, petAudioDiff)
	}

	return &petAudiosDiff, nil
}
