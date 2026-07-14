// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package hero

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// TestItemSynthesisCheckRule 测试武将合成检查
// 业务场景：验证武将合成配置正确性
func TestItemSynthesisCheckRule(t *testing.T) {
	cols, sheetMap := createHeroSynthesisTestData()

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 基本验证：测试能够正常运行
	if result == nil {
		t.Error("检查结果不应为空")
	}
}

// TestItemSynthesisCheckRule_GeneralHero_SheetNameMatch_BugReproduce
// 复现 Bug 1：ArenaScoreReward 表名单复数不匹配导致大将军武将合成检查完全失效
//
// Bug 详情（2026-04-14 线上流水线发现）：
//   - 文件名: ArenaScoreRewards.xlsx（复数带s）
//   - 实际 sheet 名: "竞技场积分奖励表|ArenaScoreReward"（单数无s）
//   - 代码中使用 FindSheetBySuffix(sheetMap, "ArenaScoreRewards")（复数带s）
//   - strings.HasSuffix("竞技场积分奖励表|ArenaScoreReward", "|ArenaScoreRewards") = false
//   - 导致 arenaScoreRewardsCols = nil, checkGeneralHeroes 直接 return
//   - 所有赛季的大将军武将合成检查全部失效
//
// 影响的线上问题：
//   - 李信(10613) ArenaSeason 3 大将军奖励，ValidDate为空（未配置合成），赛季已结束但未报错
//   - 霍光(10809) ArenaSeason 4 大将军奖励，ValidDate已配置，赛季未开始不应开放
//
// 修复：table_check_hero_synthesis.go:272 "ArenaScoreRewards" → "ArenaScoreReward"
func TestItemSynthesisCheckRule_GeneralHero_SheetNameMatch_BugReproduce(t *testing.T) {
	// 使用实际的 sheet 名（单数后缀）构造测试数据
	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10613",               // 武将ID（李信）
		"李信",                  // 武将名
		"",                    // ValidDate为空（未配置合成）
		"2026-01-01 00:00:00", // 赛季开始时间
		"2026-02-01 23:59:59", // 赛季结束时间（已结束）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 修复前：因表名不匹配，checkGeneralHeroes 被跳过，此处会 FAIL
	// 修复后：正确匹配表名，检测到大将军武将合成问题，此处 PASS
	if len(result.ErrCells) == 0 {
		t.Error("期望检测到大将军武将合成配置问题（赛季已结束但未配置合成），但 ErrCells 为空。" +
			"可能原因：ArenaScoreReward 表名单复数不匹配导致 checkGeneralHeroes 被跳过")
	}

	if len(result.ErrCells) > 0 {
		found := false
		for _, err := range result.ErrCells {
			if strings.Contains(err.Reason, "大将军") && strings.Contains(err.Reason, "10613") {
				found = true
				t.Logf("正确检测到大将军武将合成问题: %s", err.Reason)
				break
			}
		}
		if !found {
			t.Errorf("期望错误信息包含'大将军'和'10613'，实际错误: %v", result.ErrCells)
		}
	}
}

// TestItemSynthesisCheckRule_GeneralHero_SyntheticEnabled_NoError
// 验证修复后的正常场景：赛季已结束且 ValidDate 在保护期之后，不应报错
//
// 对应线上问题：霍光(10809) ArenaSeason 4 大将军奖励
//   - 保护期 = SeasonStartTime + 4个月 = 2026-01-01 + 4月 = 2026-05-01
//   - ValidDate=2026-06-01 在保护期之后，规则不应误报
func TestItemSynthesisCheckRule_GeneralHero_SyntheticEnabled_NoError(t *testing.T) {
	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10809",               // 武将ID（霍光）
		"霍光",                  // 武将名
		"2026-06-01 00:00:00", // ValidDate在保护期截止之后（SeasonStartTime+4月=2026-05-01）
		"2026-01-01 00:00:00",
		"2026-02-01 23:59:59", // 赛季已结束
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 赛季已结束且 ValidDate 已配置（在赛季结束之后）→ 不应有大将军相关错误
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "大将军") {
			t.Errorf("赛季已结束且 ValidDate 已配置（在赛季结束之后），不应报大将军合成错误: %s", err.Reason)
		}
	}
}

// TestItemSynthesisCheckRule_GeneralHero_SeasonEndingSoon_Warning
// 验证修复后的预警场景：赛季即将结束（≤7天）且 ValidDate 为空，应发送预警
//
// 对应线上问题：李信(10613) ArenaSeason 3 最后一天，ValidDate 为空
//   - 赛季 2026-03-15 ~ 2026-04-14 23:59:59
//   - 检查日 2026-04-14 距结束不到1天，在7天预警窗口内
//   - 应触发预警而非报错
func TestItemSynthesisCheckRule_GeneralHero_SeasonEndingSoon_Warning(t *testing.T) {
	// 赛季结束时间设为未来3天（在7天预警窗口内）
	cols, sheetMap := createSynthesisGeneralTestDataEndingSoon(
		"10613", "李信", "", // ValidDate为空（未配置合成）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 赛季即将结束 + ValidDate为空 → 应有预警
	if len(result.ErrCells) == 0 {
		t.Error("赛季即将结束（<=7天）且未配置合成时间，期望有预警")
	}

	if len(result.ErrCells) > 0 {
		found := false
		for _, err := range result.ErrCells {
			if strings.Contains(err.Reason, "大将军") {
				found = true
				t.Logf("正确触发大将军预警: %s", err.Reason)
				break
			}
		}
		if !found {
			t.Errorf("期望预警信息包含'大将军'，实际: %v", result.ErrCells)
		}
	}
}

// TestItemSynthesisCheckRule_GeneralHero_MultipleSeasons
// 验证多赛季场景：不同赛季的大将军武将应独立检查
//
// 复现线上场景：
//   - Season 3: 李信(10613) 大将军, 赛季已结束, ValidDate为空 → 应报错
//   - Season 4: 霍光(10809) 大将军, 赛季已结束, ValidDate已配置 → 不报错
func TestItemSynthesisCheckRule_GeneralHero_MultipleSeasons(t *testing.T) {
	cols, sheetMap := createSynthesisMultiSeasonTestData()

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 至少应检测到赛季已结束的李信未配置合成
	foundLixin := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "大将军") && strings.Contains(err.Reason, "10613") {
			foundLixin = true
			t.Logf("正确检测到李信合成问题: %s", err.Reason)
		}
	}
	if !foundLixin {
		t.Error("期望检测到李信(10613) 大将军武将赛季已结束但未配置合成")
	}
}

// ==================== 测试数据创建函数 ====================

// createHeroSynthesisTestData 创建武将合成检查的基础测试数据
func createHeroSynthesisTestData() (cols [][]string, sheetMap map[string]*excelize.File) {
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	heroFile.SetCellValue(sheetName, "A1", "")
	heroFile.SetCellValue(sheetName, "A2", "")
	heroFile.SetCellValue(sheetName, "A3", "Id")
	heroFile.SetCellValue(sheetName, "A4", "")

	heroFile.SetCellValue(sheetName, "A5", "10806")
	heroFile.SetCellValue(sheetName, "A6", "10807")

	cols, _ = heroFile.GetCols(sheetName)

	sheetMap = map[string]*excelize.File{
		sheetName: heroFile,
	}

	return cols, sheetMap
}

// createSynthesisGeneralHeroTestData 创建大将军武将合成检查的测试数据
// 使用实际的 sheet 名称格式 "中文名|英文后缀"
// 武将道具ID = 1000000 + heroId, 需在 IsHeroItem 有效范围 1001000-1099999 内
//
// 参数：
//   - heroId: 武将ID（如 "10613" 李信）
//   - heroName: 武将名称
//   - validDate: DropItem.ValidDate 的值，空字符串表示未配置合成
//   - seasonStart, seasonEnd: 赛季时间字符串
func createSynthesisGeneralHeroTestData(heroId, heroName, validDate, seasonStart, seasonEnd string) (cols [][]string, sheetMap map[string]*excelize.File) {
	// 武将道具ID
	itemId := 1000000
	id := 0
	for _, ch := range heroId {
		id = id*10 + int(ch-'0')
	}
	itemId += id

	// Hero 表
	heroFile := excelize.NewFile()
	heroSheet := "武将表|Hero"
	heroFile.SetSheetName("Sheet1", heroSheet)
	heroFile.SetCellValue(heroSheet, "A1", "")
	heroFile.SetCellValue(heroSheet, "A2", "")
	heroFile.SetCellValue(heroSheet, "A3", "Id")
	heroFile.SetCellValue(heroSheet, "A4", "")
	heroFile.SetCellValue(heroSheet, "B3", "Name")
	heroFile.SetCellValue(heroSheet, "A5", heroId)
	heroFile.SetCellValue(heroSheet, "B5", heroName)
	cols, _ = heroFile.GetCols(heroSheet)

	// DropItem 表：武将道具，ValidDate 由参数控制
	dropFile := excelize.NewFile()
	dropSheet := "掉落库|DropItem"
	dropFile.SetSheetName("Sheet1", dropSheet)
	dropFile.SetCellValue(dropSheet, "A1", "")
	dropFile.SetCellValue(dropSheet, "A2", "")
	dropFile.SetCellValue(dropSheet, "A3", "Id")
	dropFile.SetCellValue(dropSheet, "A4", "")
	dropFile.SetCellValue(dropSheet, "B3", "Item")
	dropFile.SetCellValue(dropSheet, "C3", "ValidDate")
	dropFile.SetCellValue(dropSheet, "A5", "1")
	dropFile.SetCellValue(dropSheet, "B5", fmt.Sprintf("{%d;1}", itemId))
	if validDate != "" {
		dropFile.SetCellValue(dropSheet, "C5", validDate)
	}

	// ArenaScoreReward 表（使用实际的单数后缀 "ArenaScoreReward"）
	arenaScoreFile := excelize.NewFile()
	arenaScoreSheet := "竞技场积分奖励表|ArenaScoreReward"
	arenaScoreFile.SetSheetName("Sheet1", arenaScoreSheet)
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A1", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A2", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A3", "Season")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A4", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B3", "Dan")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C3", "DanName")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D3", "Reward")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A5", "1")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B5", "23")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C5", "大将军")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D5", fmt.Sprintf("{1000023;100}{%d;1}", itemId))

	// ArenaSeason 表
	arenaSeasonFile := excelize.NewFile()
	arenaSeasonSheet := "竞技场积分表|ArenaSeason"
	arenaSeasonFile.SetSheetName("Sheet1", arenaSeasonSheet)
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A1", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A2", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A3", "Id")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A4", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B3", "SeasonStartTime")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C3", "SeasonEndTime")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A5", "1")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B5", seasonStart)
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C5", seasonEnd)

	sheetMap = map[string]*excelize.File{
		heroSheet:        heroFile,
		dropSheet:        dropFile,
		arenaScoreSheet:  arenaScoreFile,
		arenaSeasonSheet: arenaSeasonFile,
	}

	return cols, sheetMap
}

// createSynthesisGeneralTestDataEndingSoon 创建赛季即将结束的测试数据
// 赛季结束时间 = 当前时间 + 3天（在7天预警窗口内）
func createSynthesisGeneralTestDataEndingSoon(heroId, heroName, validDate string) (cols [][]string, sheetMap map[string]*excelize.File) {
	now := time.Now()
	seasonStart := now.AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	seasonEnd := now.AddDate(0, 0, 3).Format("2006-01-02 15:04:05")
	return createSynthesisGeneralHeroTestData(heroId, heroName, validDate, seasonStart, seasonEnd)
}

// createSynthesisMultiSeasonTestData 创建多赛季大将军武将测试数据
// 模拟线上实际场景：
//   - Season 3: 李信(10613) 大将军, 赛季已结束, ValidDate为空 → 应报错
//   - Season 4: 霍光(10809) 大将军, 赛季已结束, ValidDate已配置 → 不报错
func createSynthesisMultiSeasonTestData() (cols [][]string, sheetMap map[string]*excelize.File) {
	// Hero 表：两个大将军武将
	heroFile := excelize.NewFile()
	heroSheet := "武将表|Hero"
	heroFile.SetSheetName("Sheet1", heroSheet)
	heroFile.SetCellValue(heroSheet, "A1", "")
	heroFile.SetCellValue(heroSheet, "A2", "")
	heroFile.SetCellValue(heroSheet, "A3", "Id")
	heroFile.SetCellValue(heroSheet, "A4", "")
	heroFile.SetCellValue(heroSheet, "B3", "Name")
	heroFile.SetCellValue(heroSheet, "A5", "10613")
	heroFile.SetCellValue(heroSheet, "B5", "李信")
	heroFile.SetCellValue(heroSheet, "A6", "10809")
	heroFile.SetCellValue(heroSheet, "B6", "霍光")
	cols, _ = heroFile.GetCols(heroSheet)

	// DropItem 表：李信 ValidDate为空（未配置合成），霍光 ValidDate已配置
	dropFile := excelize.NewFile()
	dropSheet := "掉落库|DropItem"
	dropFile.SetSheetName("Sheet1", dropSheet)
	dropFile.SetCellValue(dropSheet, "A1", "")
	dropFile.SetCellValue(dropSheet, "A2", "")
	dropFile.SetCellValue(dropSheet, "A3", "Id")
	dropFile.SetCellValue(dropSheet, "A4", "")
	dropFile.SetCellValue(dropSheet, "B3", "Item")
	dropFile.SetCellValue(dropSheet, "C3", "ValidDate")
	dropFile.SetCellValue(dropSheet, "A5", "1")
	dropFile.SetCellValue(dropSheet, "B5", "{1010613;1}")
	// 李信：ValidDate为空（不设置 C5），未配置合成
	dropFile.SetCellValue(dropSheet, "A6", "2")
	dropFile.SetCellValue(dropSheet, "B6", "{1010809;1}")
	dropFile.SetCellValue(dropSheet, "C6", "2026-06-01 00:00:00") // 霍光：ValidDate已配置（赛季结束后）

	// ArenaScoreReward 表：两个赛季的大将军段位
	arenaScoreFile := excelize.NewFile()
	arenaScoreSheet := "竞技场积分奖励表|ArenaScoreReward"
	arenaScoreFile.SetSheetName("Sheet1", arenaScoreSheet)
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A1", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A2", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A3", "Season")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A4", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B3", "Dan")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C3", "DanName")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D3", "Reward")
	// Season 3 大将军：李信
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A5", "3")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B5", "23")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C5", "大将军")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D5", "{1000023;100}{1010613;1}")
	// Season 4 大将军：霍光
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A6", "4")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B6", "23")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C6", "大将军")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D6", "{1000023;100}{1010809;1}")

	// ArenaSeason 表：两个赛季
	arenaSeasonFile := excelize.NewFile()
	arenaSeasonSheet := "竞技场积分表|ArenaSeason"
	arenaSeasonFile.SetSheetName("Sheet1", arenaSeasonSheet)
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A1", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A2", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A3", "Id")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A4", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B3", "SeasonStartTime")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C3", "SeasonEndTime")
	// Season 3 已结束
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A5", "3")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B5", "2026-03-15 00:00:00")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C5", "2026-04-14 23:59:59")
	// Season 4 已结束
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A6", "4")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B6", "2026-04-15 00:00:00")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C6", "2026-05-14 23:59:59")

	sheetMap = map[string]*excelize.File{
		heroSheet:        heroFile,
		dropSheet:        dropFile,
		arenaScoreSheet:  arenaScoreFile,
		arenaSeasonSheet: arenaSeasonFile,
	}

	return cols, sheetMap
}

// ==================== 战令武将截止日测试 ====================

// TestItemSynthesisCheckRule_SeasonPassHero_DeadlineApproaching_Warn
// 验证截止日前 7 天触发预警：战令已结束，合成截止日距今在 warnDuration 内
//
// 场景：嬴政(10618) SP 结束 2026-03-14，截止日 2026-06-14
// 注入时间 2026-06-07（截止日前 7 天），应触发预警
func TestItemSynthesisCheckRule_SeasonPassHero_DeadlineApproaching_Warn(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData(
		"10618",               // 武将ID（嬴政）
		"嬴政",                  // 武将名
		"",                    // ValidDate为空（未配置合成）
		"2026-01-15 00:00:00", // SP 开始时间
		"2026-03-14 23:59:59", // SP 结束时间
	)

	// 注入时间：截止日前约 7 天（截止日 2026-06-14 23:59:59 UTC，使用 UTC 避免时区偏差）
	injectedNow := time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落库|DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         injectedNow,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	if len(result.ErrCells) == 0 {
		t.Error("截止日前 7 天且未配置合成时间，期望触发预警")
	}

	if len(result.ErrCells) > 0 {
		found := false
		for _, err := range result.ErrCells {
			if strings.Contains(err.Reason, "嬴政") || strings.Contains(err.Reason, "10618") {
				found = true
				t.Logf("截止日前 7 天正确触发预警: %s", err.Reason)
				break
			}
		}
		if !found {
			t.Errorf("期望预警信息包含'嬴政'或'10618'，实际: %v", result.ErrCells)
		}
	}
}

// TestItemSynthesisCheckRule_SeasonPassHero_DeadlineFarAway_NoWarn
// 验证截止日前 30 天不触发预警：合成截止日远未到来
//
// 场景：嬴政(10618) SP 开始 2026-02-15，保护期截止日 = StartTime+4月 = 2026-06-15
// 注入时间 2026-05-15（截止日前 31 天，> 7天预警窗口），不应触发预警
func TestItemSynthesisCheckRule_SeasonPassHero_DeadlineFarAway_NoWarn(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData(
		"10618",               // 武将ID（嬴政）
		"嬴政",                  // 武将名
		"",                    // ValidDate为空（未配置合成）
		"2026-02-15 00:00:00", // SP 开始时间（保护期截止 = 2026-06-15）
		"2026-04-14 23:59:59", // SP 结束时间
	)

	// 注入时间：截止日前 30 天（2026-05-15）
	injectedNow := time.Date(2026, 5, 15, 10, 0, 0, 0, time.Local)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落库|DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         injectedNow,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 截止日前 30 天（> 7 天），不应触发预警
	if len(result.ErrCells) > 0 {
		t.Errorf("截止日前 30 天（超出7天预警窗口），不应触发预警，实际: %v", result.ErrCells)
	}
}

// createSeasonPassHeroTestData 创建战令武将合成检查测试数据
//
// 参数：
//   - heroId: 武将ID（如 "10618" 嬴政）
//   - heroName: 武将名称
//   - validDate: DropItem.ValidDate 的值，空字符串表示未配置合成
//   - spStart, spEnd: 战令时间
func createSeasonPassHeroTestData(
	heroId, heroName, validDate string,
	spStart, spEnd string,
) (cols [][]string, sheetMap map[string]*excelize.File) {
	// 武将道具ID
	itemId := 1000000
	id := 0
	for _, ch := range heroId {
		id = id*10 + int(ch-'0')
	}
	itemId += id

	// Hero 表
	heroFile := excelize.NewFile()
	heroSheet := "武将表|Hero"
	heroFile.SetSheetName("Sheet1", heroSheet)
	heroFile.SetCellValue(heroSheet, "A3", "Id")
	heroFile.SetCellValue(heroSheet, "B3", "Name")
	heroFile.SetCellValue(heroSheet, "A5", heroId)
	heroFile.SetCellValue(heroSheet, "B5", heroName)
	cols, _ = heroFile.GetCols(heroSheet)

	// DropItem 表
	dropFile := excelize.NewFile()
	dropSheet := "掉落库|DropItem"
	dropFile.SetSheetName("Sheet1", dropSheet)
	dropFile.SetCellValue(dropSheet, "A3", "Id")
	dropFile.SetCellValue(dropSheet, "B3", "Item")
	dropFile.SetCellValue(dropSheet, "C3", "ValidDate")
	dropFile.SetCellValue(dropSheet, "A5", "1")
	dropFile.SetCellValue(dropSheet, "B5", fmt.Sprintf("{%d;1}", itemId))
	if validDate != "" {
		dropFile.SetCellValue(dropSheet, "C5", validDate)
	}

	// SeasonPassReward 表
	spRewardFile := excelize.NewFile()
	spRewardSheet := "赛季战令奖励表|SeasonPassReward"
	spRewardFile.SetSheetName("Sheet1", spRewardSheet)
	spRewardFile.SetCellValue(spRewardSheet, "A3", "SeasonPassId")
	spRewardFile.SetCellValue(spRewardSheet, "B3", "level")
	spRewardFile.SetCellValue(spRewardSheet, "C3", "HighReward")
	// Level=1 的战令武将奖励
	spRewardFile.SetCellValue(spRewardSheet, "A5", "1")
	spRewardFile.SetCellValue(spRewardSheet, "B5", "1")
	spRewardFile.SetCellValue(spRewardSheet, "C5", fmt.Sprintf("{%d;1}", itemId))

	// SeasonPass 表
	spFile := excelize.NewFile()
	spSheet := "赛季战令表|SeasonPass"
	spFile.SetSheetName("Sheet1", spSheet)
	spFile.SetCellValue(spSheet, "A3", "Id")
	spFile.SetCellValue(spSheet, "B3", "StartTime")
	spFile.SetCellValue(spSheet, "C3", "EndTime")
	spFile.SetCellValue(spSheet, "A5", "1")
	spFile.SetCellValue(spSheet, "B5", spStart)
	spFile.SetCellValue(spSheet, "C5", spEnd)

	sheetMap = map[string]*excelize.File{
		heroSheet:     heroFile,
		dropSheet:     dropFile,
		spRewardSheet: spRewardFile,
		spSheet:       spFile,
	}

	return cols, sheetMap
}

// ==================== Bug 3 & Bug 4: 反向检查测试 ====================

// TestItemSynthesisCheckRule_GeneralHero_SeasonNotStarted_SyntheticEnabled_Warn
// Bug 3: 赛季未开始但 ValidDate 已配置（提前开放合成）
//
// 场景：霍光(10809) ArenaSeason 5 大将军奖励
//   - ArenaSeason 5 赛季未开始（开始时间在未来）
//   - DropItem.ValidDate 已配置（早于赛季开始时间）
//   - 服务端合成判断走 DropItem.ValidDate
//   - ValidDate 早于赛季开始意味着玩家可以立即合成，但赛季还没开始，不应该提前开放
//
// 预期：应检测到"赛季未开始不应提前开放合成"的错误
func TestItemSynthesisCheckRule_GeneralHero_SeasonNotStarted_SyntheticEnabled_Warn(t *testing.T) {
	// 赛季在未来 30 天后开始，尚未开始
	now := time.Now()
	seasonStart := now.AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonEnd := now.AddDate(0, 1, 30).Format("2006-01-02 15:04:05")

	// ValidDate 设为赛季开始之前（2026-04-10 早于赛季开始时间）
	validDate := now.AddDate(0, 0, 29).Format("2006-01-02") + " 00:00:00"

	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10809",     // 武将ID（霍光）
		"霍光",        // 武将名
		validDate,   // ValidDate早于赛季开始时间
		seasonStart, // 赛季开始时间（未来）
		seasonEnd,   // 赛季结束时间（未来）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// Bug 3 修复后：检测到赛季未开始但已配置合成
	if len(result.ErrCells) == 0 {
		t.Error("期望检测到大将军武将赛季未开始但已配置合成的错误，但 ErrCells 为空。" +
			"Bug 3：缺少反向检查（赛季未开始不应开放合成）")
	}

	if len(result.ErrCells) > 0 {
		found := false
		for _, err := range result.ErrCells {
			if strings.Contains(err.Reason, "大将军") && strings.Contains(err.Reason, "10809") {
				found = true
				t.Logf("正确检测到赛季未开始但已配置合成: %s", err.Reason)
				break
			}
		}
		if !found {
			t.Errorf("期望错误信息包含'大将军'和'10809'，实际: %v", result.ErrCells)
		}
	}
}

// TestItemSynthesisCheckRule_GeneralHero_SeasonNotStarted_SyntheticDisabled_NoWarn
// 验证 Bug 3 反向检查不会误报：赛季未开始且 ValidDate 为空（正确配置），不应报错
func TestItemSynthesisCheckRule_GeneralHero_SeasonNotStarted_SyntheticDisabled_NoWarn(t *testing.T) {
	now := time.Now()
	seasonStart := now.AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonEnd := now.AddDate(0, 1, 30).Format("2006-01-02 15:04:05")

	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10809",
		"霍光",
		"", // ValidDate为空（未配置合成，正确配置）
		seasonStart,
		seasonEnd,
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 赛季未开始 + ValidDate为空 = 正确配置，不应有大将军相关错误
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "大将军") {
			t.Errorf("赛季未开始且 ValidDate 为空，不应报大将军错误: %s", err.Reason)
		}
	}
}

// TestItemSynthesisCheckRule_SeasonPassHero_SpNotStarted_SyntheticEnabled_Warn
// Bug 4: 战令未开始但 ValidDate 已配置（提前开放合成）
//
// 场景：平阳公主(10813) SeasonPass 2 战令武将
//   - SeasonPass 2 未开始（开始时间在未来）
//   - DropItem.ValidDate 已配置（早于战令开始时间）
//   - 服务端合成判断走 DropItem.ValidDate
//   - ValidDate 早于战令开始意味着战令还没开始就可以合成，不应该提前开放
//
// 预期：应检测到"战令未开始不应提前开放合成"的错误
func TestItemSynthesisCheckRule_SeasonPassHero_SpNotStarted_SyntheticEnabled_Warn(t *testing.T) {
	now := time.Now()
	// 战令在未来 30 天后开始
	spStart := now.AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	spEnd := now.AddDate(0, 2, 30).Format("2006-01-02 15:04:05")

	// ValidDate 设为战令开始之前（29天后，早于战令开始30天）
	validDate := now.AddDate(0, 0, 29).Format("2006-01-02") + " 00:00:00"

	cols, sheetMap := createSeasonPassHeroTestData(
		"10813",   // 武将ID（平阳公主）
		"平阳公主",    // 武将名
		validDate, // ValidDate早于战令开始时间
		spStart,   // 战令开始时间（未来）
		spEnd,     // 战令结束时间（未来）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落库|DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// Bug 4 修复后：检测到战令未开始但已配置合成
	if len(result.ErrCells) == 0 {
		t.Error("期望检测到战令武将战令未开始但已配置合成的错误，但 ErrCells 为空。" +
			"Bug 4：缺少反向检查（战令未开始不应开放合成）")
	}

	if len(result.ErrCells) > 0 {
		found := false
		for _, err := range result.ErrCells {
			if strings.Contains(err.Reason, "平阳公主") || strings.Contains(err.Reason, "10813") {
				if strings.Contains(err.Reason, "尚未开始") {
					found = true
					t.Logf("正确检测到战令未开始但已配置合成: %s", err.Reason)
					break
				}
			}
		}
		if !found {
			t.Errorf("期望错误信息包含'平阳公主'或'10813'，实际: %v", result.ErrCells)
		}
	}
}

// TestItemSynthesisCheckRule_SeasonPassHero_SpNotStarted_SyntheticDisabled_NoWarn
// 验证 Bug 4 反向检查不会误报：战令未开始且 ValidDate 为空（正确配置），不应报错
func TestItemSynthesisCheckRule_SeasonPassHero_SpNotStarted_SyntheticDisabled_NoWarn(t *testing.T) {
	now := time.Now()
	spStart := now.AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	spEnd := now.AddDate(0, 2, 30).Format("2006-01-02 15:04:05")

	cols, sheetMap := createSeasonPassHeroTestData(
		"10813",
		"平阳公主",
		"", // ValidDate为空（未配置合成，正确配置）
		spStart,
		spEnd,
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落库|DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 战令未开始 + ValidDate为空 = 正确配置，不应有错误
	if len(result.ErrCells) > 0 {
		t.Errorf("战令未开始且 ValidDate 为空，不应触发任何错误，实际: %v", result.ErrCells)
	}
}

// ==================== 线上精确复现测试 ====================
// 使用线上真实的日期和ID构造数据，通过 CheckParam.Now 注入固定时间（2026-04-14），
// 确保测试结果确定性的，不依赖于实际运行时间。

// productionNow 线上 Bug 发现日的注入时间：2026-04-14 10:00:00
var productionNow = time.Date(2026, 4, 14, 10, 0, 0, 0, time.Local)

// TestItemSynthesisCheckRule_Production_Bug3_HuoGuang_10809
// 精确复现 Bug 3 线上场景：霍光(10809)
//   - ArenaSeason 4: 2026-04-15 ~ 2026-05-14
//   - DropItem.ValidDate="2026-04-10 00:00:00"（早于赛季开始时间）
//   - 注入时间 2026-04-14：赛季明天开始 → 应报"赛季未开始不应开放合成"
func TestItemSynthesisCheckRule_Production_Bug3_HuoGuang_10809(t *testing.T) {
	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10809",               // 武将ID（霍光）
		"霍光",                  // 武将名
		"2026-04-10 00:00:00", // ValidDate早于赛季开始时间（线上实际配置）
		"2026-04-15 00:00:00", // ArenaSeason 4 开始时间（线上真实）
		"2026-05-14 23:59:59", // ArenaSeason 4 结束时间（线上真实）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         productionNow, // 注入线上 Bug 发现日
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 确定性断言：赛季 2026-04-15 开始，注入时间 2026-04-14，一定检测到反向检查错误
	found := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "大将军") && strings.Contains(err.Reason, "10809") &&
			strings.Contains(err.Reason, "尚未开始") {
			found = true
			t.Logf("正确检测到: %s", err.Reason)
			break
		}
	}
	if !found {
		t.Errorf("赛季尚未开始（2026-04-15），期望检测到反向检查错误，实际: %v", result.ErrCells)
	}
}

// TestItemSynthesisCheckRule_Production_Bug4_PingYang_10813
// 精确复现 Bug 4 线上场景：平阳公主(10813)
//   - SeasonPass 4: 2026-04-15 ~ 2026-05-14
//   - DropItem.ValidDate="2026-04-10 00:00:00"（早于战令开始时间）
//   - 注入时间 2026-04-14：战令明天开始 → 应报"战令未开始不应开放合成"
func TestItemSynthesisCheckRule_Production_Bug4_PingYang_10813(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData(
		"10813",               // 武将ID（平阳公主）
		"平阳公主",                // 武将名
		"2026-04-10 00:00:00", // ValidDate早于战令开始时间（线上实际配置）
		"2026-04-15 00:00:00", // SeasonPass 4 开始时间（线上真实）
		"2026-05-14 23:59:59", // SeasonPass 4 结束时间（线上真实）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落库|DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         productionNow, // 注入线上 Bug 发现日
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 确定性断言：战令 2026-04-15 开始，注入时间 2026-04-14，一定检测到反向检查错误
	found := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "平阳公主") || strings.Contains(err.Reason, "10813") {
			if strings.Contains(err.Reason, "尚未开始") {
				found = true
				t.Logf("正确检测到: %s", err.Reason)
				break
			}
		}
	}
	if !found {
		t.Errorf("战令尚未开始（2026-04-15），期望检测到反向检查错误，实际: %v", result.ErrCells)
	}
}

// TestItemSynthesisCheckRule_Production_AllBugs
// 综合复现：使用线上全部武将的真实数据，注入 2026-04-14 固定时间，验证修复后的完整检查结果。
//   - 李信(10613): S3 大将军, S3 已结束, ValidDate为空 → 应报"赛季已结束未配置合成时间"
//   - 霍去病(10803): SP1 战令, SP1 已结束, ValidDate为空 → 保护期截止 2026-05-15（StartTime+4月），距注入时间 31 天 > 7 天，不应触发预警
//   - 霍光(10809): S4 大将军, S4 未开始, ValidDate早于赛季开始 → 应报"赛季尚未开始"
//   - 平阳公主(10813): SP4 战令, SP4 未开始, ValidDate早于战令开始 → 应报"战令尚未开始"
func TestItemSynthesisCheckRule_Production_AllBugs(t *testing.T) {

	// Hero 表：4 个武将
	heroFile := excelize.NewFile()
	heroSheet := "武将表|Hero"
	heroFile.SetSheetName("Sheet1", heroSheet)
	heroFile.SetCellValue(heroSheet, "A1", "")
	heroFile.SetCellValue(heroSheet, "A2", "")
	heroFile.SetCellValue(heroSheet, "A3", "Id")
	heroFile.SetCellValue(heroSheet, "A4", "")
	heroFile.SetCellValue(heroSheet, "B3", "Name")
	heroFile.SetCellValue(heroSheet, "A5", "10613")
	heroFile.SetCellValue(heroSheet, "B5", "李信")
	heroFile.SetCellValue(heroSheet, "A6", "10803")
	heroFile.SetCellValue(heroSheet, "B6", "霍去病")
	heroFile.SetCellValue(heroSheet, "A7", "10809")
	heroFile.SetCellValue(heroSheet, "B7", "霍光")
	heroFile.SetCellValue(heroSheet, "A8", "10813")
	heroFile.SetCellValue(heroSheet, "B8", "平阳公主")
	cols, _ := heroFile.GetCols(heroSheet)

	// DropItem 表：线上真实 ValidDate 值
	dropFile := excelize.NewFile()
	dropSheet := "掉落库|DropItem"
	dropFile.SetSheetName("Sheet1", dropSheet)
	dropFile.SetCellValue(dropSheet, "A1", "")
	dropFile.SetCellValue(dropSheet, "A2", "")
	dropFile.SetCellValue(dropSheet, "A3", "Id")
	dropFile.SetCellValue(dropSheet, "A4", "")
	dropFile.SetCellValue(dropSheet, "B3", "Item")
	dropFile.SetCellValue(dropSheet, "C3", "ValidDate")
	// 李信：ValidDate为空（未配置合成）
	dropFile.SetCellValue(dropSheet, "A5", "1")
	dropFile.SetCellValue(dropSheet, "B5", "{1010613;1}")
	// 霍去病：ValidDate为空（未配置合成）
	dropFile.SetCellValue(dropSheet, "A6", "2")
	dropFile.SetCellValue(dropSheet, "B6", "{1010803;1}")
	// 霍光：ValidDate早于赛季开始时间（S4 开始 2026-04-15，ValidDate 为 2026-04-10）
	dropFile.SetCellValue(dropSheet, "A7", "3")
	dropFile.SetCellValue(dropSheet, "B7", "{1010809;1}")
	dropFile.SetCellValue(dropSheet, "C7", "2026-04-10 00:00:00")
	// 平阳公主：ValidDate早于战令开始时间（SP4 开始 2026-04-15，ValidDate 为 2026-04-10）
	dropFile.SetCellValue(dropSheet, "A8", "4")
	dropFile.SetCellValue(dropSheet, "B8", "{1010813;1}")
	dropFile.SetCellValue(dropSheet, "C8", "2026-04-10 00:00:00")

	// ArenaScoreReward 表：S3/S4 大将军奖励（线上真实）
	arenaScoreFile := excelize.NewFile()
	arenaScoreSheet := "竞技场积分奖励表|ArenaScoreReward"
	arenaScoreFile.SetSheetName("Sheet1", arenaScoreSheet)
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A1", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A2", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A3", "Season")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A4", "")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B3", "Dan")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C3", "DanName")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D3", "Reward")
	// S3 大将军：李信
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A5", "3")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B5", "23")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C5", "大将军")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D5", "{1000023;100}{1010613;1}")
	// S4 大将军：霍光
	arenaScoreFile.SetCellValue(arenaScoreSheet, "A6", "4")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "B6", "23")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "C6", "大将军")
	arenaScoreFile.SetCellValue(arenaScoreSheet, "D6", "{1000023;100}{1010809;1}")

	// ArenaSeason 表：S3/S4（线上真实日期）
	arenaSeasonFile := excelize.NewFile()
	arenaSeasonSheet := "竞技场积分表|ArenaSeason"
	arenaSeasonFile.SetSheetName("Sheet1", arenaSeasonSheet)
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A1", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A2", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A3", "Id")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A4", "")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B3", "SeasonStartTime")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C3", "SeasonEndTime")
	// S3 已结束
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A5", "3")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B5", "2026-03-15 00:00:00")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C5", "2026-04-14 23:59:59")
	// S4
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "A6", "4")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "B6", "2026-04-15 00:00:00")
	arenaSeasonFile.SetCellValue(arenaSeasonSheet, "C6", "2026-05-14 23:59:59")

	// SeasonPassReward 表：SP1/SP4 战令武将奖励（线上真实）
	spRewardFile := excelize.NewFile()
	spRewardSheet := "赛季战令奖励表|SeasonPassReward"
	spRewardFile.SetSheetName("Sheet1", spRewardSheet)
	spRewardFile.SetCellValue(spRewardSheet, "A1", "")
	spRewardFile.SetCellValue(spRewardSheet, "A2", "")
	spRewardFile.SetCellValue(spRewardSheet, "A3", "SeasonPassId")
	spRewardFile.SetCellValue(spRewardSheet, "A4", "")
	spRewardFile.SetCellValue(spRewardSheet, "B3", "level")
	spRewardFile.SetCellValue(spRewardSheet, "C3", "HighReward")
	// SP1 Level=1：霍去病
	spRewardFile.SetCellValue(spRewardSheet, "A5", "1")
	spRewardFile.SetCellValue(spRewardSheet, "B5", "1")
	spRewardFile.SetCellValue(spRewardSheet, "C5", "{1010803;1}")
	// SP4 Level=1：平阳公主
	spRewardFile.SetCellValue(spRewardSheet, "A6", "4")
	spRewardFile.SetCellValue(spRewardSheet, "B6", "1")
	spRewardFile.SetCellValue(spRewardSheet, "C6", "{1010813;1}")

	// SeasonPass 表：SP1/SP4（线上真实日期）
	spFile := excelize.NewFile()
	spSheet := "赛季战令表|SeasonPass"
	spFile.SetSheetName("Sheet1", spSheet)
	spFile.SetCellValue(spSheet, "A1", "")
	spFile.SetCellValue(spSheet, "A2", "")
	spFile.SetCellValue(spSheet, "A3", "Id")
	spFile.SetCellValue(spSheet, "A4", "")
	spFile.SetCellValue(spSheet, "B3", "StartTime")
	spFile.SetCellValue(spSheet, "C3", "EndTime")
	// SP1 已结束（StartTime=2026-01-15 → 保护期截止 = 2026-05-15，距注入时间31天 > 7天，不触发预警）
	spFile.SetCellValue(spSheet, "A5", "1")
	spFile.SetCellValue(spSheet, "B5", "2026-01-15 00:00:00")
	spFile.SetCellValue(spSheet, "C5", "2026-02-14 23:59:59")
	// SP4
	spFile.SetCellValue(spSheet, "A6", "4")
	spFile.SetCellValue(spSheet, "B6", "2026-04-15 00:00:00")
	spFile.SetCellValue(spSheet, "C6", "2026-05-14 23:59:59")

	sheetMap := map[string]*excelize.File{
		heroSheet:        heroFile,
		dropSheet:        dropFile,
		arenaScoreSheet:  arenaScoreFile,
		arenaSeasonSheet: arenaSeasonFile,
		spRewardSheet:    spRewardFile,
		spSheet:          spFile,
	}

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "掉落库|DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         productionNow, // 注入线上 Bug 发现日 2026-04-14
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Full production result:", string(jsonData))

	// 确定性验证：注入时间 2026-04-14

	// Bug 1：李信 S3 已结束，ValidDate为空 → 应报"仍未配置合成时间"
	foundLiXin := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "李信") || strings.Contains(err.Reason, "10613") {
			foundLiXin = true
			t.Logf("Bug1 李信: %s", err.Reason)
			break
		}
	}
	if !foundLiXin {
		t.Error("Bug1 未检测到：李信(10613) 赛季已结束但未配置合成时间")
	}

	// 霍去病：保护期截止 2026-05-15（StartTime+4月），距注入时间 31 天 > 7 天，不应触发预警（保底规则已移除）
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "霍去病") || strings.Contains(err.Reason, "10803") {
			t.Errorf("霍去病(10803) 保护期截止还有31天（>7天），不应触发预警（保底规则已移除）: %s", err.Reason)
		}
	}

	// Bug 3：霍光 S4 未开始，ValidDate早于赛季开始 → 应报反向检查错误
	foundHuoGuang := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "霍光") || strings.Contains(err.Reason, "10809") {
			if strings.Contains(err.Reason, "尚未开始") {
				foundHuoGuang = true
				t.Logf("Bug3 霍光: %s", err.Reason)
				break
			}
		}
	}
	if !foundHuoGuang {
		t.Error("Bug3 未检测到：霍光(10809) 赛季未开始但已配置合成")
	}

	// Bug 4：平阳公主 SP4 未开始，ValidDate早于战令开始 → 应报反向检查错误
	foundPingYang := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "平阳公主") || strings.Contains(err.Reason, "10813") {
			if strings.Contains(err.Reason, "尚未开始") {
				foundPingYang = true
				t.Logf("Bug4 平阳公主: %s", err.Reason)
				break
			}
		}
	}
	if !foundPingYang {
		t.Error("Bug4 未检测到：平阳公主(10813) 战令未开始但已配置合成")
	}

	t.Logf("总检测错误数: %d", len(result.ErrCells))
}

// ==================== 边界条件：ValidDate 等于保护期截止时间 ====================

// TestItemSynthesisCheckRule_SeasonPass_ValidDateEqualsDeadline_NoError
// 边界条件：战令武将 DropItem.ValidDate 恰好等于保护期截止时间（StartTime + 4个月）
// 保护期截止当天生效应视为合法，不应报"保护期内不应可合成"
//
// 场景：
//   - SeasonPass.StartTime = 2026-04-15
//   - 保护期截止 = 2026-04-15 + 4月 = 2026-08-15 00:00:00
//   - DropItem.ValidDate = 2026-08-15 00:00:00（恰好等于保护期截止）
//   - 预期：不报"保护期内不应可合成"错误
func TestItemSynthesisCheckRule_SeasonPass_ValidDateEqualsDeadline_NoError(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData(
		"10803", // 霍去病
		"霍去病",
		"2026-08-15 00:00:00", // ValidDate == 保护期截止
		"2026-04-15 00:00:00", // StartTime
		"2026-07-15 23:59:59", // EndTime（已结束）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local), // 保护期已过
	})

	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "保护期内不应可合成") {
			t.Errorf("ValidDate == 保护期截止时间应视为合法，不应报错: %s", err.Reason)
		}
	}
}

// TestItemSynthesisCheckRule_GeneralHero_ValidDateEqualsDeadline_NoError
// 边界条件：大将军武将 DropItem.ValidDate 恰好等于保护期截止时间
//
// 场景：
//   - ArenaSeason.SeasonStartTime = 2026-04-15
//   - 保护期截止 = 2026-04-15 + 4月 = 2026-08-15 00:00:00
//   - DropItem.ValidDate = 2026-08-15 00:00:00（恰好等于保护期截止）
//   - 预期：不报"保护期内不应可合成"错误
func TestItemSynthesisCheckRule_GeneralHero_ValidDateEqualsDeadline_NoError(t *testing.T) {
	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10804", // 卫青
		"卫青",
		"2026-08-15 00:00:00", // ValidDate == 保护期截止
		"2026-04-15 00:00:00", // SeasonStartTime
		"2026-07-15 23:59:59", // SeasonEndTime（已结束）
	)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local),
	})

	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "保护期内不应可合成") {
			t.Errorf("ValidDate == 保护期截止时间应视为合法，不应报错: %s", err.Reason)
		}
	}
}

// ==================== 大将军保护期截止时间 = SeasonEndTime 测试 ====================

// TestItemSynthesisCheckRule_ArenaProtectionDeadline
// 验证 CalcArenaProtectionDeadline 使用 SeasonEndTime（而非旧逻辑 StartTime+4月）后的行为
//
// 场景：
//   - ArenaScoreReward 配置一个大将军武将（HeroId=10804，卫青）
//   - ArenaSeason 赛季：SeasonStartTime=2026-04-15, SeasonEndTime=2026-05-14
//   - DropItem 配置 ValidDate=2026-02-15（早于赛季结束时间 2026-05-14）
//   - 注入时间 2026-06-01（赛季已结束）
//   - 期望：报错"保护期内不应可合成"，保护期截止时间为 2026-05-14（SeasonEndTime）
//   - 而非旧逻辑的保护期截止时间 2026-08-15（StartTime+4月）
func TestItemSynthesisCheckRule_ArenaProtectionDeadline(t *testing.T) {
	cols, sheetMap := createSynthesisGeneralHeroTestData(
		"10804",               // 武将ID（卫青）
		"卫青",                  // 武将名
		"2026-02-15 00:00:00", // ValidDate 早于赛季结束时间
		"2026-04-15 00:00:00", // SeasonStartTime
		"2026-05-14 23:59:59", // SeasonEndTime（赛季已结束）
	)

	// 注入时间：赛季已结束
	injectedNow := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)

	rule := new(ItemSynthesisCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
		Now:         injectedNow,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 期望：ValidDate(2026-02-15) < 保护期截止时间(2026-05-14) → 报错"保护期内不应可合成"
	found := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "保护期内不应可合成") {
			found = true
			t.Logf("正确检测到保护期内合成: %s", err.Reason)
			break
		}
	}
	if !found {
		t.Errorf("ValidDate(2026-02-15) 早于保护期截止时间(2026-05-14)，期望报错'保护期内不应可合成'，实际: %v", result.ErrCells)
	}
}
