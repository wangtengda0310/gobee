package hero_res_check

import (
	"fmt"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/check_log"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/resource_utils"
)

type ImgCheckReport struct {
}

func CheckGitlabAndLocalImgInExcel(excelDir string, imgDir, spineDir string) (rep *ImgCheckReport, err error) {
	rep = new(ImgCheckReport)

	uiPaths := []string{
		"Global/touxiang",
		"Role",
		"HalfRole",
		"Skin",
		"Battle",
		"HeroCollection",
		"RoleTotalLineIcon",
		//"RoleOutFrame",
	}
	spinePaths := []string{"BattleSeatSpine", "HeroSpinePrefab/MainWindowBg", "HeroSpinePrefab/CollectionShow"}

	uiNames := map[string][]string{
		"Global/touxiang": {
			"ui_s1_touxiang_%s.png",
			"ui_s1_touxiang_zhujiemian_%s.png",
		},
		"Role": {
			"ui_s1_fight_km_%s.png",
			//"ui_s1_fight_km_%s_xiangao.png",
		},
		"HalfRole": {
			"ui_s1_fight_km_banshen_%s.png",
		},
		"Skin": {
			"ui_s1_gerenxinxi_pifu_%s.png",
		},
		"Battle/RoleHalfImage": {
			"ui_s1_fight__img_tu_ssqy_%s.png",
		},
		"HeroCollection/HeroIcon": {
			"ui_s1_shoucang_km_%s.png",
			//"ui_s1_shoucang_km_%s_xiangao.png",
		},
		"RoleTotalLineIcon": {
			//"ui_s1_quanshenlihui_%s.png", // 原画\动态原画\特殊皮肤不在这里
		},
		"RoleOutFrame": {
			"ui_s1_fight_km_%s_01.png",
		},
	}
	spineNames := map[string][]string{
		"BattleSeatSpine": {
			//"seat_%s_prefab.prefab",
		},
		"HeroSpinePrefab/MainWindowBg": {
			//"hero_%s_mainbg.prefab",
		},
		"HeroSpinePrefab/CollectionShow": {
			//"hero_%s_collection_prefab.prefab",
		},
	}

	wg := sync.WaitGroup{}

	filesMap := make(map[string]map[string]struct{})
	for _, p := range uiPaths {
		filesMap[p] = make(map[string]struct{})
	}
	for _, p := range spinePaths {
		filesMap[p] = make(map[string]struct{})
	}

	wg.Add(1)
	go func() {
		startTime := time.Now()
		defer func() {
			wg.Done()
			endTime := time.Now()
			check_log.Debug("get res list func cost:", endTime.Sub(startTime))
		}()

		for _, p := range uiPaths {
			filesInfo, err := resource_utils.GetFilesLocal(imgDir+p, "png")
			if err != nil {
				check_log.Error("扫描失败:", imgDir+p, err)
				continue
			}
			for _, f := range filesInfo {
				filesMap[p][f.Name] = struct{}{}
			}
		}

		for _, p := range spinePaths {
			filesInfo, err := resource_utils.GetFilesLocal(spineDir+p, "prefab")
			if err != nil {
				check_log.Error("扫描失败:", spineDir+p, err)
				continue
			}
			for _, f := range filesInfo {
				filesMap[p][f.Name] = struct{}{}
			}
		}
	}()

	var excels *mjs_excel.ImgRefExcel
	wg.Add(1)
	go func(err *error) {
		startTime := time.Now()
		defer func() {
			wg.Done()
			endTime := time.Now()
			check_log.Debug("get sheetMap func cost:", endTime.Sub(startTime))
		}()

		sheetMap, err1 := excelio.GetSheetMap(excelDir, "Hero.xlsx", "HeroSkinItem_英雄皮肤.xlsx")
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}

		excels, err1 = mjs_excel.InitImgRefExcel(sheetMap)
		if err1 != nil {
			*err = err1
			check_log.Panic(err1)
			return
		}
	}(&err)

	wg.Wait()
	if err != nil {
		return nil, err
	}

	count := 0
	for id, open := range excels.SkinIdOpenMap {
		if !open {
			continue
		}
		for _, p := range uiPaths {
			for _, n := range uiNames[p] {
				pinYin := excels.SkinIdPinYinMap[id]
				fullName := fmt.Sprintf(n, pinYin)
				if _, exist := filesMap[p][fullName]; !exist {
					check_log.Error("找不到对应资源:", fullName, ", SkinId:", id, ", SkinName:", excels.SkinIdNameMap[id])
				} else {
					count++
				}
			}
		}
		_ = spineNames
		for _, p := range spinePaths {
			for _, n := range spineNames[p] {
				pinYin := excels.SkinIdPinYinMap[id]
				fullName := fmt.Sprintf(n, pinYin)
				if _, exist := filesMap[p][fullName]; !exist {
					check_log.Error("找不到对应资源:", fullName, ", SkinId:", id, ", SkinName:", excels.SkinIdNameMap[id])
				} else {
					count++
				}
			}
		}
	}
	check_log.Info("完成检查:", count)

	return nil, err
}
