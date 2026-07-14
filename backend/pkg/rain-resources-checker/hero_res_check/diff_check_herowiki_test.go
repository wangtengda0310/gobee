package hero_res_check

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel"
)

func TestHeroWikiDiff2(t *testing.T) {
	sheetMap, err := excelio.GetSheetMap("../config/excel")
	if err != nil {
		t.Fatal(err)
	}
	excels, err := mjs_excel.InitDiffRefExcel(sheetMap)
	if err != nil {
		t.Fatal(err)
	}
	excels.HeroWikiDiff = herowiki.BuildHeroWikiDiff(excels)

	comparator := NewComparator("../mjs_excel/tmp/test.json")

	oldWiki, _ := comparator.LoadPreviousData()

	result := comparator.CompareHeroWikiDiff(oldWiki.HeroWikiDiff, excels.HeroWikiDiff)

	// 打印详细结果
	printHeroWikiDiffResult(result)
}

func TestHeroWikiDiff(t *testing.T) {
	fmt.Println("========== 开始测试HeroWikiDiff对比功能 ==========")

	comparator := NewComparator("../mjs_excel/tmp/herowiki_test.json")

	// 创建旧版本的HeroWikiDiff数据
	oldWiki := createOldHeroWikiData()

	// 创建新版本的HeroWikiDiff数据（包含一些修改）
	newWiki := createNewHeroWikiData()

	newContainer := &diff.DataContainer{
		HeroWikiDiff: newWiki,
	}

	// 执行对比
	fmt.Println("\n📊 执行对比分析...")
	result := comparator.CompareHeroWikiDiff(oldWiki, newWiki)

	// 打印详细结果
	printHeroWikiDiffResult(result)

	// 保存到DataContainer
	newContainer.HeroWikiDiffResult = result

	// 生成JSON报告
	generateHeroWikiJSONReport(result)

	// 保存当前数据供下次对比使用
	saveHeroWikiData(newContainer, comparator)

	fmt.Println("\n✅ HeroWikiDiff测试完成！")
}

// 创建旧版本数据
func createOldHeroWikiData() *herowiki_def.HeroWikiDiff {
	heroId1 := "EHero_ZhaoYun"
	heroId2 := "EHero_GuanYu"

	return &herowiki_def.HeroWikiDiff{
		Indexes: &herowiki_def.HeroIndexes{
			HeroByID:      map[int]string{10001: heroId1, 10002: heroId2},
			HeroByEHeroId: map[string]int{heroId1: 10001, heroId2: 10002},
			HeroesByCountry: map[string][]string{
				"蜀": {heroId1, heroId2},
			},
		},
		Heroes: map[string]*herowiki_def.HeroCompleteData{
			heroId1: createZhaoYunData(false), // 旧版本
			heroId2: createGuanYuData(false),  // 旧版本
		},
	}
}

// 创建新版本数据（包含变化）
func createNewHeroWikiData() *herowiki_def.HeroWikiDiff {
	heroId1 := "EHero_ZhaoYun"
	heroId2 := "EHero_GuanYu"
	heroId3 := "EHero_ZhangFei" // 新增武将

	return &herowiki_def.HeroWikiDiff{
		Indexes: &herowiki_def.HeroIndexes{
			HeroByID:      map[int]string{10001: heroId1, 10002: heroId2, 10003: heroId3},
			HeroByEHeroId: map[string]int{heroId1: 10001, heroId2: 10002, heroId3: 10003},
			HeroesByCountry: map[string][]string{
				"蜀": {heroId1, heroId2, heroId3},
			},
		},
		Heroes: map[string]*herowiki_def.HeroCompleteData{
			heroId1: createZhaoYunData(true), // 新版本（有修改）
			heroId2: createGuanYuData(true),  // 新版本（无修改）
			heroId3: createZhangFeiData(),    // 新增武将
		},
	}
}

// 创建赵云数据
func createZhaoYunData(isNew bool) *herowiki_def.HeroCompleteData {
	data := &herowiki_def.HeroCompleteData{
		Basic: &herowiki_def.HeroBasicInfo{
			Id:              10001,
			EHeroId:         "EHero_ZhaoYun",
			Name:            "赵云",
			IsOpen:          true,
			OpenDate:        "2026-01-01",
			Gender:          1,
			Point:           100,
			HpLimit:         1000,
			HandLimit:       4,
			EquipLimit:      3,
			Country:         "蜀",
			IsAlwaysZhuGong: false,
			HeroType:        1,
			EHeroType:       "力量",
			CanMelt:         true,
			MeltName:        []string{"龙胆"},
			IsNewHero:       false,
			IsGacha:         true,
		},
		UI: &herowiki_def.HeroUIInfo{
			HeroDiffId:        10001,
			Name:              "赵云",
			LongIntroduction:  "常山赵子龙，一身是胆",
			ShortIntroduction: "七进七出",
			GetWay:            "抽卡",
			Position:          "前锋",
		},
		Skills: []*herowiki_def.HeroSkillInfo{
			createQiangSkill(isNew),
			createLongSkill(isNew),
		},
		Country: &herowiki_def.CountryInfo{
			ECountry: "蜀",
			Name:     "蜀国",
			KingName: "刘备",
			IsOpen:   true,
		},
	}

	// 如果是新版本，修改一些字段
	if isNew {
		data.Basic.HpLimit = 1200                       // 体力从1000改为1200
		data.Basic.Point = 150                          // 珠点从100改为150
		data.UI.LongIntroduction = "常山赵子龙，一身是胆，七进七出救阿斗" // 修改描述
	}

	return data
}

// 创建关羽数据
func createGuanYuData(isNew bool) *herowiki_def.HeroCompleteData {
	return &herowiki_def.HeroCompleteData{
		Basic: &herowiki_def.HeroBasicInfo{
			Id:         10002,
			EHeroId:    "EHero_GuanYu",
			Name:       "关羽",
			IsOpen:     true,
			OpenDate:   "2026-01-01",
			Gender:     1,
			Point:      120,
			HpLimit:    1100,
			HandLimit:  4,
			EquipLimit: 3,
			Country:    "蜀",
			HeroType:   1,
			EHeroType:  "力量",
			CanMelt:    true,
			MeltName:   []string{"武圣"},
			IsNewHero:  false,
			IsGacha:    true,
		},
		UI: &herowiki_def.HeroUIInfo{
			HeroDiffId:        10002,
			Name:              "关羽",
			LongIntroduction:  "美髯公，武圣关羽",
			ShortIntroduction: "千里走单骑",
			GetWay:            "抽卡",
			Position:          "主力",
		},
		Country: &herowiki_def.CountryInfo{
			ECountry: "蜀",
			Name:     "蜀国",
			KingName: "刘备",
			IsOpen:   true,
		},
	}
}

// 创建张飞数据（新增）
func createZhangFeiData() *herowiki_def.HeroCompleteData {
	return &herowiki_def.HeroCompleteData{
		Basic: &herowiki_def.HeroBasicInfo{
			Id:         10003,
			EHeroId:    "EHero_ZhangFei",
			Name:       "张飞",
			IsOpen:     true,
			OpenDate:   "2026-03-01",
			Gender:     1,
			Point:      110,
			HpLimit:    1300,
			HandLimit:  4,
			EquipLimit: 3,
			Country:    "蜀",
			HeroType:   1,
			EHeroType:  "力量",
			CanMelt:    true,
			MeltName:   []string{"咆哮"},
			IsNewHero:  true,
			IsGacha:    true,
		},
		UI: &herowiki_def.HeroUIInfo{
			HeroDiffId:        10003,
			Name:              "张飞",
			LongIntroduction:  "燕人张翼德，当阳桥上一声吼",
			ShortIntroduction: "喝断当阳桥",
			GetWay:            "抽卡",
			Position:          "前锋",
		},
		Country: &herowiki_def.CountryInfo{
			ECountry: "蜀",
			Name:     "蜀国",
			KingName: "刘备",
			IsOpen:   true,
		},
	}
}

// 创建枪技能
func createQiangSkill(isNew bool) *herowiki_def.HeroSkillInfo {
	skill := &herowiki_def.HeroSkillInfo{
		Basic: &herowiki_def.SkillBasicInfo{
			Id:              "1001",
			SkillName:       "枪术精通",
			ShortSkillName:  "枪术",
			ESkillId:        "ESkill_Qiang",
			SkillType:       "主动",
			SkillLimitTimes: 3,
			TotalLimitTimes: 3,
		},
		UI: &herowiki_def.SkillUIInfo{
			Id:             "ESkill_Qiang",
			SkillName:      "枪术精通",
			SkillText:      "使用枪类武器时，攻击力提升20%",
			ShortSkillText: "枪类武器攻击+20%",
		},
		Melt: &herowiki_def.SkillMeltInfo{
			Id:        "1001",
			MeltPower: 100,
			CanMelt:   true,
		},
		Tags: []*herowiki_def.SkillTagInfo{
			{
				SkillTag: "1",
				TagName:  "攻击",
				TagColor: "#f5222d",
				TagDes:   "提升攻击力",
			},
		},
	}

	// 如果是新版本，修改技能效果
	if isNew {
		skill.UI.SkillText = "使用枪类武器时，攻击力提升25%" // 从20%改为25%
		skill.Basic.SkillLimitTimes = 4         // 从3次改为4次
	}

	return skill
}

// 创建龙技能
func createLongSkill(isNew bool) *herowiki_def.HeroSkillInfo {
	return &herowiki_def.HeroSkillInfo{
		Basic: &herowiki_def.SkillBasicInfo{
			Id:              "1002",
			SkillName:       "龙胆",
			ShortSkillName:  "龙胆",
			ESkillId:        "ESkill_Long",
			SkillType:       "被动",
			SkillLimitTimes: 1,
			TotalLimitTimes: 1,
		},
		UI: &herowiki_def.SkillUIInfo{
			Id:             "ESkill_Long",
			SkillName:      "龙胆",
			SkillText:      "每回合开始时，有30%概率获得1点怒气",
			ShortSkillText: "概率获得怒气",
		},
	}
}

// 打印HeroWikiDiff对比结果
func printHeroWikiDiffResult(result *diff.HeroWikiDiffResult) {
	if result == nil {
		fmt.Println("❌ 没有对比结果")
		return
	}

	fmt.Println("\n📋 ========== HeroWikiDiff 对比报告 ==========")
	fmt.Printf("🕐 对比时间: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("📊 总武将数: %d\n", result.Summary.TotalHeroes)

	fmt.Println("\n📈 变化统计:")
	fmt.Printf("  ✅ 新增: %d 个\n", len(result.Summary.AddedHeroes))
	fmt.Printf("  ❌ 删除: %d 个\n", len(result.Summary.RemovedHeroes))
	fmt.Printf("  ⚠️ 修改: %d 个\n", len(result.Summary.ModifiedHeroes))
	fmt.Printf("  📍 总变化数: %d 处\n", result.Summary.TotalChanges)

	if len(result.Summary.AddedHeroes) > 0 {
		fmt.Println("\n✨ 新增武将:")
		for _, heroId := range result.Summary.AddedHeroes {
			if detail, ok := result.HeroesDiff[heroId]; ok {
				fmt.Printf("  + %s (%s)\n", detail.Name, heroId)
			}
		}
	}

	if len(result.Summary.RemovedHeroes) > 0 {
		fmt.Println("\n🗑️ 删除武将:")
		for _, heroId := range result.Summary.RemovedHeroes {
			if detail, ok := result.HeroesDiff[heroId]; ok {
				fmt.Printf("  - %s (%s)\n", detail.Name, heroId)
			}
		}
	}

	if len(result.Summary.ModifiedHeroes) > 0 {
		fmt.Println("\n📝 修改详情:")
		for _, heroId := range result.Summary.ModifiedHeroes {
			if detail, ok := result.HeroesDiff[heroId]; ok {
				fmt.Printf("\n  📌 %s (%s) - %d处变化:\n", detail.Name, heroId, detail.ChangeCount)

				// 打印字段变化
				for _, change := range detail.FieldChanges {
					fmt.Printf("    • %s:\n", change.FieldName)
					fmt.Printf("      旧值: %v\n", change.OldValue)
					fmt.Printf("      新值: %v\n", change.NewValue)
				}

				// 打印嵌套结构变化
				for path, nested := range detail.NestedChanges {
					fmt.Printf("    📁 %s [%s]:\n", path, nested.StructType)
					for _, change := range nested.Changes {
						fmt.Printf("      • %s: %v → %v\n",
							change.FieldName, change.OldValue, change.NewValue)
					}
				}
			}
		}
	}

	fmt.Println("\n==========================================")
}

// 生成JSON报告
func generateHeroWikiJSONReport(result *diff.HeroWikiDiffResult) {
	report := map[string]interface{}{
		"timestamp": result.Timestamp,
		"summary": map[string]interface{}{
			"total_heroes":    result.Summary.TotalHeroes,
			"added_heroes":    result.Summary.AddedHeroes,
			"removed_heroes":  result.Summary.RemovedHeroes,
			"modified_heroes": result.Summary.ModifiedHeroes,
			"total_changes":   result.Summary.TotalChanges,
		},
		"details": map[string]interface{}{},
	}

	// 添加详细信息
	for heroId, detail := range result.HeroesDiff {
		heroDetail := map[string]interface{}{
			"name":          detail.Name,
			"change_type":   detail.ChangeType,
			"change_count":  detail.ChangeCount,
			"field_changes": []map[string]interface{}{},
		}

		// 添加字段变化
		for _, change := range detail.FieldChanges {
			heroDetail["field_changes"] = append(heroDetail["field_changes"].([]map[string]interface{}), map[string]interface{}{
				"field":       change.FieldName,
				"old_value":   change.OldValue,
				"new_value":   change.NewValue,
				"change_type": change.ChangeType,
			})
		}

		report["details"].(map[string]interface{})[heroId] = heroDetail
	}

	// 保存JSON文件
	reportJSON, _ := json.MarshalIndent(report, "", "  ")

	if err := os.MkdirAll("./tmp", os.ModePerm); err != nil {
		fmt.Printf("❌ 创建目录失败: %v\n", err)
	} else {
		err := os.WriteFile("./tmp/herowiki_diff_report.json", reportJSON, 0644)
		if err != nil {
			fmt.Printf("❌ 保存报告失败: %v\n", err)
		} else {
			fmt.Println("✅ JSON报告已保存: ./tmp/herowiki_diff_report.json")
		}
	}
}

// 保存HeroWiki数据
func saveHeroWikiData(container *diff.DataContainer, comparator *Comparator) {
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		return
	}

	if err := os.MkdirAll("../mjs_excel/tmp", os.ModePerm); err != nil {
		fmt.Printf("❌ 创建目录失败: %v\n", err)
		return
	}

	err = os.WriteFile("../mjs_excel/tmp/herowiki_test.json", data, 0644)
	if err != nil {
		fmt.Printf("❌ 保存数据失败: %v\n", err)
	} else {
		fmt.Println("✅ HeroWiki数据已保存: ../mjs_excel/tmp/herowiki_test.json")
	}
}

// 运行测试
func TestHeroWikiDiffWithRealData(t *testing.T) {
	fmt.Println("\n🎯 运行HeroWikiDiff真实数据测试")
	TestHeroWikiDiff(t)
}
