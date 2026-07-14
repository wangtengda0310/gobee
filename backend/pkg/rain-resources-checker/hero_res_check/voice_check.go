package hero_res_check

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/check_log"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/resource_utils"
)

type ExistInfo struct {
	Path   string
	Exist  bool
	Reason string
}

type LineVoiceRepeatInfo struct {
	RepeatNum  int
	Location   []string
	ExistInfos []ExistInfo
}
type AudioRepeatInfo struct {
	RepeatNum  int
	Location   []string
	ExistInfos []ExistInfo
}

type VoiceCheckReport struct {
	ExcelMaps                   *mjs_excel.AudioRefExcel
	HeroLineVoiceRepeatVoiceMap map[int]map[int]*LineVoiceRepeatInfo // heroid 台词id 重复信息
	HeroAudioRepeatVoiceMap     map[int]map[string]*AudioRepeatInfo  // heroid 音频枚举 重复信息
}

func MarkLineVoice2RepeatMap(voicesRepeatMap map[int]*LineVoiceRepeatInfo, vid int, location string) {
	if repeatInfo, ok := voicesRepeatMap[vid]; !ok {
		voicesRepeatMap[vid] = &LineVoiceRepeatInfo{
			RepeatNum: 1,
			Location:  []string{location},
		}
	} else {
		repeatInfo.RepeatNum++
		repeatInfo.Location = append(repeatInfo.Location, location)
		voicesRepeatMap[vid] = repeatInfo
	}
}

func MarkAudio2RepeatMap(audioRepeatMap map[string]*AudioRepeatInfo, vEnum string, location string) {
	if repeatInfo, ok := audioRepeatMap[vEnum]; !ok {
		audioRepeatMap[vEnum] = &AudioRepeatInfo{
			RepeatNum: 1,
			Location:  []string{location},
		}
	} else {
		repeatInfo.RepeatNum++
		repeatInfo.Location = append(repeatInfo.Location, location)
		audioRepeatMap[vEnum] = repeatInfo
	}
}

// CheckGitlabAndLocalVoicesInExcel
// excelDir = "../"+vars.ExcelPath
// audioDir = "../../../client/Master/Card/Audio/"
func CheckGitlabAndLocalVoicesInExcel(excelDir string, audioDir string) (rep *VoiceCheckReport, err error) {
	rep = new(VoiceCheckReport)

	wg := sync.WaitGroup{}

	filesMap := make(map[string]map[string]string) // 第一层是Audio目录内的文件夹名, 第二层是wav文件名
	filesMap["Amb"] = make(map[string]string)
	filesMap["Bgm"] = make(map[string]string)
	filesMap["Pet"] = make(map[string]string)
	filesMap["Se"] = make(map[string]string)
	filesMap["UI"] = make(map[string]string)
	filesMap["Voice"] = make(map[string]string)
	wg.Add(1)
	go func(err *error) {
		startTime := time.Now()
		defer func() {
			wg.Done()
			endTime := time.Now()
			check_log.Debug("get res list func cost:", endTime.Sub(startTime))
		}()

		//filesVoiceInfo, err := checker.GetFilesGitlab("v0.0.8-pre-release", "Xcards/client", "Master/Card/Audio/Voice", "jMD7RXcbMLosvHoPZTbS", "wav")
		//if err != nil {
		//	t.Fatal(err)
		//}
		filesAmbInfo, err1 := resource_utils.GetFilesLocal(audioDir+"Amb", "wav")
		if err1 != nil {
			*err = err1
			check_log.Panic(err)
			return
		}
		filesBgmInfo, err1 := resource_utils.GetFilesLocal(audioDir+"Bgm", "wav")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
		filesPetInfo, err1 := resource_utils.GetFilesLocal(audioDir+"Pet", "wav")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
		filesVoiceInfo, err1 := resource_utils.GetFilesLocal(audioDir+"Voice", "wav")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
		filesUIInfo, err1 := resource_utils.GetFilesLocal(audioDir+"UI", "wav")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
		filesSeInfo, err1 := resource_utils.GetFilesLocal(audioDir+"Se", "wav")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
		for _, fileInfo := range filesAmbInfo {
			filesMap["Amb"][fileInfo.Name] = fileInfo.Path
		}
		for _, fileInfo := range filesBgmInfo {
			filesMap["Bgm"][fileInfo.Name] = fileInfo.Path
		}
		for _, fileInfo := range filesPetInfo {
			filesMap["Pet"][fileInfo.Name] = fileInfo.Path
		}
		for _, fileInfo := range filesVoiceInfo {
			filesMap["Voice"][fileInfo.Name] = fileInfo.Path
		}
		for _, fileInfo := range filesUIInfo {
			filesMap["UI"][fileInfo.Name] = fileInfo.Path
		}
		for _, fileInfo := range filesSeInfo {
			filesMap["Se"][fileInfo.Name] = fileInfo.Path
		}
	}(&err)

	var excels *mjs_excel.AudioRefExcel
	wg.Add(1)
	go func(err *error) {
		startTime := time.Now()
		defer func() {
			wg.Done()
			endTime := time.Now()
			check_log.Debug("get sheetMap func cost:", endTime.Sub(startTime))
		}()

		sheetMap, err1 := excelio.GetSheetMap(excelDir, "Audio_音频配置表.xlsx", "Hero.xlsx", "HeroLines_武将台词表.xlsx", "HeroSkinItem_英雄皮肤.xlsx", "HeroSkinSpine_英雄皮肤Spine.xlsx", "Skill.xlsx", "SkillLines_技能台词表.xlsx", "SkillUI_技能表现表.xlsx")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
		excels, err1 = mjs_excel.InitAudioRefExcel(sheetMap)
		if err != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
	}(&err)

	wg.Wait()
	if err != nil {
		return nil, err
	}

	// 各音频的重复使用数量及使用位置
	heroLineVoiceRepeatVoiceMap := make(map[int]map[int]*LineVoiceRepeatInfo) // heroid 台词id 重复信息
	//heroLinesTestMap := make(map[int]map[int]LineVoiceRepeatInfo)            // 配表里填写的所有信息
	heroAudioRepeatVoiceMap := make(map[int]map[string]*AudioRepeatInfo) // heroid 音频枚举 重复信息
	for id, open := range excels.HeroIdOpenMap {
		if !open {
			continue
		}
		voicesRepeatMap := make(map[int]*LineVoiceRepeatInfo)
		//voicesRepeatTestMap := make(map[int]LineVoiceRepeatInfo)
		audioRepeatMap := make(map[string]*AudioRepeatInfo)

		// 技能音频
		for _, skillId := range excels.HeroSkillsMap[id] {
			voice, exist := excels.SkillIdLineVoiceIdMap[skillId]
			if !exist {
				check_log.Error("voice not ExistInfos, Id:", id, "skillId:", skillId)
				continue
			}
			for _, vid := range voice.SkillFirstLine {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能第一段台词(SkillLines表)[SkillFirstLine]"))
			}
			for _, vid := range voice.SkillSecondLine {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能第二段台词(SkillLines表)[SkillSecondLine]"))
			}
			for _, vid := range voice.SkillThirdLine {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能第三段台词(SkillLines表)[SkillThirdLine]"))
			}
			for _, vid := range voice.SkillForthLine {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能第四段台词(SkillLines表)[SkillForthLine]"))
			}
			for _, vid := range voice.SpecialAudio {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能特殊音效(SkillLines表)[SpecialAudio]"))
			}
			skillEnum, exist := excels.SkillIdEnumMap[skillId]
			if !exist {
				check_log.Error("skillEnum not ExistInfos, Id:", id, "skillId:", skillId)
				continue
			}
			// skill一般找不到playCard
			if ae, exist := excels.SkillEnumPlayCardAudioEnumMap[skillEnum]; exist {
				MarkAudio2RepeatMap(audioRepeatMap, ae, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "出牌音效(SkillUI表)[PlayCardAudio]"))
			}
			// skill一般找不到身份台词
			if arArr, exist := excels.SkillEnumIdentityLineVoiceEnumMap[skillEnum]; exist {
				for _, vEnum := range arArr {
					MarkLineVoice2RepeatMap(voicesRepeatMap, excels.LineVoiceEnumIdMap[vEnum], fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "身份技能台词(SkillUI表)[IdentityLine]"))
				}
			}
			if vArr, exist := excels.SkillEnumLineVoiceIdMap[skillEnum]; exist {
				for _, vid := range vArr {
					MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "打牌时触发技能的台词(SkillUI表)[Audio]"))
				}
			}
			// skill一般找不到特殊音效
			if ae, exist := excels.SkillEnumSpecialAudioEnumMap[skillEnum]; exist {
				if vid, exist := excels.LineVoiceEnumIdMap[ae]; exist {
					MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能特殊音效(SkillUI表)[SpecialAudio]"))
				} else {
					MarkLineVoice2RepeatMap(voicesRepeatMap, -1, fmt.Sprintf("skill: %s[%d]:%s", excels.SkillIdNameMap[skillId], skillId, "技能特殊音效(SkillUI表)[SpecialAudio], 找不到枚举对应audioId"))
				}
			}
		}
		// 皮肤音频
		for skinId, skinVoice := range excels.HeroSkinLineVoiceIdMap[id] {
			if !excels.SkinIdOpenMap[skinId] {
				check_log.Warn("武将检测到未开放皮肤, Id:", id, "skinId:", skinId)
				continue
			}
			//for _, vid := range skinVoice.Lines {
			//	MarkLineVoice2RepeatMap(voicesRepeatTestMap, vid, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将台词[Lines]"))
			//}
			for _, vid := range skinVoice.DebutLines {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将登场台词[DebutLines]"))
			}
			for _, vid := range skinVoice.KillLines {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将击杀台词[KillLines]"))
			}
			for _, vid := range skinVoice.DeadLines {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将阵亡台词[DeadLines]"))
			}
			for _, vid := range skinVoice.HeroAudio {
				MarkLineVoice2RepeatMap(voicesRepeatMap, vid, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将自选台词[HeroAudio]"))
			}
			for _, vEnum := range excels.SkinIdSpineAnimAudioMap[skinId] {
				MarkAudio2RepeatMap(audioRepeatMap, vEnum, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将皮肤Spine动画音效[SpineAnimAudio]"))
			}
			if enum, exist := excels.SkinIdKillAudioMap[skinId]; exist {
				MarkAudio2RepeatMap(audioRepeatMap, enum, fmt.Sprintf("skin: %s[%d]:%s", excels.SkinIdNameMap[skinId], skinId, "武将皮肤Kill动画音效[KillAudio]"))
			}
		}

		heroLineVoiceRepeatVoiceMap[id] = voicesRepeatMap
		//heroLinesTestMap[id] = voicesRepeatTestMap
		heroAudioRepeatVoiceMap[id] = audioRepeatMap
	}

	// 音频英雄音频统计打印输出
	//indent1, err := json.MarshalIndent(HeroLineVoiceRepeatVoiceMap[10806], "", "\t")
	//if err != nil {
	//	t.Error(err)
	//}
	//fmt.Println(string(indent1))

	rep.ExcelMaps = excels
	rep.HeroLineVoiceRepeatVoiceMap = heroLineVoiceRepeatVoiceMap
	rep.HeroAudioRepeatVoiceMap = heroAudioRepeatVoiceMap

	// 音频文件查询是否存在
	okCount := 0
	for id, lineVoiceRepeatInfo := range heroLineVoiceRepeatVoiceMap {
		if !excels.HeroIdOpenMap[id] {
			check_log.Warn("武将未开放, Id:", id)
			continue
		}
		for vid, rInfo := range lineVoiceRepeatInfo {
			ve, exist := excels.LineVoiceIdEnumMap[vid]
			if !exist {
				content := fmt.Sprint("武将台词表未配置该台词ID, Id:", id, ", vId:", vid)
				check_log.Warn(content)
				lineVoiceRepeatInfo[vid].ExistInfos = []ExistInfo{
					{
						Reason: content,
						Exist:  false,
					},
				}
				continue
			}
			paths, exist := excels.AudioEnumPathMap[ve]
			if !exist {
				content := fmt.Sprint("音频配置表未配置该台词枚举, Enum:", ve)
				check_log.Warn(content)
				lineVoiceRepeatInfo[vid].ExistInfos = []ExistInfo{
					{
						Reason: content,
						Exist:  false,
					},
				}
				continue
			}
			typeE := excels.AudioEnumTypeEnumMap[ve]
			existRes, reasonRes, err := compareLocalFile(&okCount, typeE, paths, filesMap, id, ve, rInfo.Location)
			existInfo := make([]ExistInfo, len(paths))
			if err != nil {
				for i := range existInfo {
					existInfo[i].Path = paths[i]
					existInfo[i].Reason = reasonRes[i]
					existInfo[i].Exist = false
				}
			} else {
				for i := range existInfo {
					existInfo[i].Path = paths[i]
					existInfo[i].Reason = reasonRes[i]
					existInfo[i].Exist = existRes[i]
				}
			}
			lineVoiceRepeatInfo[vid].ExistInfos = existInfo
		}
	}
	for id, audioEnumRepeatInfo := range heroAudioRepeatVoiceMap {
		if !excels.HeroIdOpenMap[id] {
			check_log.Warn("武将未开放, Id:", id)
			continue
		}
		for enum, rInfo := range audioEnumRepeatInfo {
			paths, exist := excels.AudioEnumPathMap[enum]
			if !exist {
				content := fmt.Sprint("音频配置表未配置该台词枚举, Enum:", enum)
				check_log.Warn(content)
				audioEnumRepeatInfo[enum].ExistInfos = []ExistInfo{
					{
						Reason: content,
						Exist:  false,
					},
				}
				continue
			}
			typeE, exist := excels.AudioEnumTypeEnumMap[enum]
			if !exist {
				content := fmt.Sprint("音频配置表未配置该台词枚举, Enum:", enum)
				check_log.Warn(content)
				audioEnumRepeatInfo[enum].ExistInfos = []ExistInfo{
					{
						Reason: content,
						Exist:  false,
					},
				}
				continue
			}
			existRes, reasonRes, err := compareLocalFile(&okCount, typeE, paths, filesMap, id, enum, rInfo.Location)
			existInfo := make([]ExistInfo, len(paths))
			if err != nil {
				for i := range existInfo {
					existInfo[i].Path = paths[i]
					existInfo[i].Reason = reasonRes[i]
					existInfo[i].Exist = false
				}
			} else {
				for i := range existInfo {
					existInfo[i].Path = paths[i]
					existInfo[i].Reason = reasonRes[i]
					existInfo[i].Exist = existRes[i]
				}
			}
			audioEnumRepeatInfo[enum].ExistInfos = existInfo
		}
	}
	check_log.Info("okCount:", okCount)

	return rep, nil
}

func compareLocalFile(okCount *int, typeE string, paths []string, filesMap map[string]map[string]string, heroId int, audioE string, location []string) (existRes []bool, reasonRes []string, err error) {
	existRes = make([]bool, len(paths))
	reasonRes = make([]string, len(paths))
	if typeE == "CommonVoice" || typeE == "HeroVoice" {
		for i, path := range paths {
			if _, exist := filesMap["Voice"][path+".wav"]; !exist {
				content := fmt.Sprint("资源不存在, Id:", heroId, ", audioEnum:", audioE, ", wavPath:", path)
				check_log.Error(content)
				existRes[i] = false
				reasonRes[i] = content
			} else {
				*okCount++
				existRes[i] = true
				reasonRes[i] = ""
			}
		}
	} else if typeE == "Amb" || typeE == "Bgm" || typeE == "Pet" || typeE == "Se" || typeE == "UI" {
		for i, path := range paths {
			if _, exist := filesMap[typeE][path+".wav"]; !exist {
				content := fmt.Sprint("资源不存在, Id:", heroId, ", audioEnum:", audioE, ", wavPath:", path)
				check_log.Error(content)
				existRes[i] = false
				reasonRes[i] = content
			} else {
				*okCount++
				existRes[i] = true
				reasonRes[i] = ""
			}
		}
	} else {
		check_log.Error("台词类型不在Voice/Se分类内, Id:", heroId, ", audioEnum:", audioE)
		return nil, nil, errors.New("台词类型不在Voice/Se分类内")
	}
	return existRes, reasonRes, nil
}
