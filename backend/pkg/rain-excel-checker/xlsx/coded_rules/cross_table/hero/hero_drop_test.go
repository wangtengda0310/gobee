// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package hero

import (
	"encoding/json"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// TestHeroDropCheckRule_NotInPool 测试武将未加入掉落库
// 业务场景：战令已结束超过N个月，武将仍未加入掉落库
func TestHeroDropCheckRule_NotInPool(t *testing.T) {
	cols, sheetMap := createHeroDropTestData_NotInPool()

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	// 验证结果：武将未加入掉落库应该报错
	if result.Ok {
		t.Error("武将未加入掉落库应该报错")
	}

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// TestHeroDropCheckRule_InPool 测试武将已加入掉落库
// 业务场景：武将已正常加入掉落库
func TestHeroDropCheckRule_InPool(t *testing.T) {
	cols, sheetMap := createHeroDropTestData_InPool()

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	// 验证结果：武将已加入掉落库应该通过
	if !result.Ok {
		t.Errorf("武将已加入掉落库应该通过: %s", result.Reason)
	}
}

// createHeroDropTestData_NotInPool 创建武将未加入掉落库的测试数据
func createHeroDropTestData_NotInPool() (cols [][]string, sheetMap map[string]*excelize.File) {
	// Hero 表
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	// 表头
	heroFile.SetCellValue(sheetName, "A1", "")
	heroFile.SetCellValue(sheetName, "A2", "")
	heroFile.SetCellValue(sheetName, "A3", "Id")
	heroFile.SetCellValue(sheetName, "A4", "")
	heroFile.SetCellValue(sheetName, "B1", "")
	heroFile.SetCellValue(sheetName, "B2", "")
	heroFile.SetCellValue(sheetName, "B3", "Name")
	heroFile.SetCellValue(sheetName, "B4", "")
	heroFile.SetCellValue(sheetName, "C1", "")
	heroFile.SetCellValue(sheetName, "C2", "")
	heroFile.SetCellValue(sheetName, "C3", "IsOpen")
	heroFile.SetCellValue(sheetName, "C4", "")
	heroFile.SetCellValue(sheetName, "D1", "")
	heroFile.SetCellValue(sheetName, "D2", "datetime")
	heroFile.SetCellValue(sheetName, "D3", "OpenDate")
	heroFile.SetCellValue(sheetName, "D4", "")

	// 武已开放超过6个月（超过3个月延迟+当前时间）
	heroFile.SetCellValue(sheetName, "A5", "10805")
	heroFile.SetCellValue(sheetName, "B5", "测试武将")
	heroFile.SetCellValue(sheetName, "C5", "true")
	openDate := time.Now().AddDate(0, -6, 0).Format("2006-01-02 15:04:05")
	heroFile.SetCellValue(sheetName, "D5", openDate)

	cols, _ = heroFile.GetCols(sheetName)

	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	rewardSheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", rewardSheetName)

	rewardFile.SetCellValue(rewardSheetName, "A1", "")
	rewardFile.SetCellValue(rewardSheetName, "A2", "")
	rewardFile.SetCellValue(rewardSheetName, "A3", "Id")
	rewardFile.SetCellValue(rewardSheetName, "A4", "")
	rewardFile.SetCellValue(rewardSheetName, "B1", "")
	rewardFile.SetCellValue(rewardSheetName, "B2", "")
	rewardFile.SetCellValue(rewardSheetName, "B3", "StartTime")
	rewardFile.SetCellValue(rewardSheetName, "B4", "")
	rewardFile.SetCellValue(rewardSheetName, "C1", "")
	rewardFile.SetCellValue(rewardSheetName, "C2", "")
	rewardFile.SetCellValue(rewardSheetName, "C3", "EndTime")
	rewardFile.SetCellValue(rewardSheetName, "C4", "")

	// 战令已结束
	endTime := time.Now().AddDate(0, -5, 0).Format("2006-01-02 15:04:05")
	rewardFile.SetCellValue(rewardSheetName, "A5", "1")
	rewardFile.SetCellValue(rewardSheetName, "B5", endTime)
	rewardFile.SetCellValue(rewardSheetName, "C5", endTime)

	// DropItem 表（掉落库，不包含该武将）
	dropFile := excelize.NewFile()
	dropSheetName := "DropItem"
	dropFile.SetSheetName("Sheet1", dropSheetName)

	dropFile.SetCellValue(dropSheetName, "A1", "")
	dropFile.SetCellValue(dropSheetName, "A2", "")
	dropFile.SetCellValue(dropSheetName, "A3", "Id")
	dropFile.SetCellValue(dropSheetName, "A4", "")
	dropFile.SetCellValue(dropSheetName, "B1", "")
	dropFile.SetCellValue(dropSheetName, "B2", "")
	dropFile.SetCellValue(dropSheetName, "B3", "Item")
	dropFile.SetCellValue(dropSheetName, "B4", "")

	dropFile.SetCellValue(dropSheetName, "A5", "1")
	dropFile.SetCellValue(dropSheetName, "B5", "{1002001;1}") // 其他物品

	sheetMap = map[string]*excelize.File{
		sheetName:       heroFile,
		rewardSheetName: rewardFile,
		dropSheetName:   dropFile,
	}

	return cols, sheetMap
}

// createHeroDropTestData_InPool 创建武将已加入掉落库的测试数据
func createHeroDropTestData_InPool() (cols [][]string, sheetMap map[string]*excelize.File) {
	cols, sheetMap = createHeroDropTestData_NotInPool()

	// 修改 DropItem 表，添加该武将
	dropFile := sheetMap["DropItem"]
	dropFile.SetCellValue("DropItem", "A6", "2")
	dropFile.SetCellValue("DropItem", "B6", "{1010805;1}") // 武将道具ID = 1000000 + 10805

	return cols, sheetMap
}

// --- 反向保护期测试 ---

// TestHeroDropCheck_SeasonPassProtectionViolation 战令武将ValidDate早于保护期截止时间
// 预期：报错，错误信息包含保护期截止时间
func TestHeroDropCheck_SeasonPassProtectionViolation(t *testing.T) {
	cols, sheetMap := createSeasonPassProtectionTestData()

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	// now = 2026-05-14，战令EndTime=2026-06-14，保护期截止=2026-09-14
	// ValidDate=2026-02-15 早于保护期截止，应报错
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	if result.Ok {
		t.Error("战令武将ValidDate早于保护期截止时间应该报错")
	}

	// 验证错误信息包含关键词
	found := false
	for _, err := range result.ErrCells {
		if len(err.Reason) > 0 {
			t.Logf("Error: %s", err.Reason)
			if containsStr(err.Reason, "保护期内不应在掉落池中") {
				found = true
			}
		}
	}
	if !found {
		t.Error("应包含反向保护期检查的错误信息")
	}
}

// TestHeroDropCheck_SeasonPassProtectionPass 战令武将ValidDate在保护期之后
// 预期：通过（不报反向保护期错误）
func TestHeroDropCheck_SeasonPassProtectionPass(t *testing.T) {
	cols, sheetMap := createSeasonPassProtectionTestData()

	// 修改 ValidDate 为保护期截止之后
	dropFile := sheetMap["DropItem"]
	dropFile.SetCellValue("DropItem", "C5", "2026-10-01 00:00:00") // 保护期截止=2026-09-14

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	// 不应有反向保护期错误
	for _, err := range result.ErrCells {
		if containsStr(err.Reason, "保护期内不应在掉落池中") {
			t.Errorf("ValidDate在保护期之后不应报反向保护期错误: %s", err.Reason)
		}
	}
}

// TestHeroDropCheck_SeasonPassProtectionEmptyValidDate 战令武将ValidDate为空
// 预期：跳过反向检查（由正向检查负责）
func TestHeroDropCheck_SeasonPassProtectionEmptyValidDate(t *testing.T) {
	cols, sheetMap := createSeasonPassProtectionTestData()

	// 清空 ValidDate
	dropFile := sheetMap["DropItem"]
	dropFile.SetCellValue("DropItem", "C5", "")

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	// 不应有反向保护期错误
	for _, err := range result.ErrCells {
		if containsStr(err.Reason, "保护期内不应在掉落池中") {
			t.Errorf("ValidDate为空不应报反向保护期错误: %s", err.Reason)
		}
	}
}

// TestHeroDropCheck_GeneralProtectionViolation 大将军武将ValidDate早于赛季结束时间
// 预期：报错
func TestHeroDropCheck_GeneralProtectionViolation(t *testing.T) {
	cols, sheetMap := createGeneralProtectionTestData()

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	// now = 2026-05-14，赛季EndTime=2026-06-14，ValidDate=2026-03-01 早于赛季结束
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	if result.Ok {
		t.Error("大将军武将ValidDate早于赛季结束时间应该报错")
	}

	found := false
	for _, err := range result.ErrCells {
		t.Logf("Error: %s", err.Reason)
		if containsStr(err.Reason, "保护期内不应在掉落池中") {
			found = true
		}
	}
	if !found {
		t.Error("应包含大将军反向保护期检查的错误信息")
	}
}

// TestHeroDropCheck_GeneralProtectionPass 大将军武将ValidDate在赛季结束之后
// 预期：不报反向保护期错误
func TestHeroDropCheck_GeneralProtectionPass(t *testing.T) {
	cols, sheetMap := createGeneralProtectionTestData()

	// 修改 ValidDate 为赛季结束之后
	dropFile := sheetMap["DropItem"]
	dropFile.SetCellValue("DropItem", "C5", "2026-07-01 00:00:00") // 赛季结束=2026-06-14

	params := make(map[string]string)
	params[string(json_rule.WARN_DAYS_BEFORE)] = "5"
	params[string(json_rule.DROP_MONTHS_DELAY)] = "3"

	rule := new(HeroDropCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	for _, err := range result.ErrCells {
		if containsStr(err.Reason, "保护期内不应在掉落池中") {
			t.Errorf("ValidDate在赛季结束之后不应报反向保护期错误: %s", err.Reason)
		}
	}
}

// containsStr 检查字符串是否包含子串
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createSeasonPassProtectionTestData 创建战令武将反向保护期测试数据
// 场景：战令武将（HeroId=11903，赵奢），ValidDate=2026-02-15，战令EndTime=2026-06-14
// 保护期截止 = EndTime + 3个月 = 2026-09-14，ValidDate早于保护期截止
func createSeasonPassProtectionTestData() (cols [][]string, sheetMap map[string]*excelize.File) {
	// Hero 表
	heroFile := excelize.NewFile()
	heroSheet := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheet)

	heroFile.SetCellValue(heroSheet, "A1", "")
	heroFile.SetCellValue(heroSheet, "A2", "")
	heroFile.SetCellValue(heroSheet, "A3", "Id")
	heroFile.SetCellValue(heroSheet, "A4", "")
	heroFile.SetCellValue(heroSheet, "B1", "")
	heroFile.SetCellValue(heroSheet, "B2", "")
	heroFile.SetCellValue(heroSheet, "B3", "Name")
	heroFile.SetCellValue(heroSheet, "B4", "")
	heroFile.SetCellValue(heroSheet, "C1", "")
	heroFile.SetCellValue(heroSheet, "C2", "")
	heroFile.SetCellValue(heroSheet, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheet, "C4", "")
	heroFile.SetCellValue(heroSheet, "D1", "")
	heroFile.SetCellValue(heroSheet, "D2", "datetime")
	heroFile.SetCellValue(heroSheet, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheet, "D4", "")
	heroFile.SetCellValue(heroSheet, "E1", "")
	heroFile.SetCellValue(heroSheet, "E2", "")
	heroFile.SetCellValue(heroSheet, "E3", "HeroType")
	heroFile.SetCellValue(heroSheet, "E4", "")

	// 赵奢：HeroId=11903, IsOpen=true, OpenDate=2026-05-15
	heroFile.SetCellValue(heroSheet, "A5", "11903")
	heroFile.SetCellValue(heroSheet, "B5", "赵奢")
	heroFile.SetCellValue(heroSheet, "C5", "true")
	heroFile.SetCellValue(heroSheet, "D5", "2026-05-15 00:00:00")
	heroFile.SetCellValue(heroSheet, "E5", "1")

	cols, _ = heroFile.GetCols(heroSheet)

	// SeasonPass 表
	spFile := excelize.NewFile()
	spSheet := "SeasonPass"
	spFile.SetSheetName("Sheet1", spSheet)

	spFile.SetCellValue(spSheet, "A1", "")
	spFile.SetCellValue(spSheet, "A2", "")
	spFile.SetCellValue(spSheet, "A3", "Id")
	spFile.SetCellValue(spSheet, "A4", "")
	spFile.SetCellValue(spSheet, "B1", "")
	spFile.SetCellValue(spSheet, "B2", "datetime")
	spFile.SetCellValue(spSheet, "B3", "StartTime")
	spFile.SetCellValue(spSheet, "B4", "")
	spFile.SetCellValue(spSheet, "C1", "")
	spFile.SetCellValue(spSheet, "C2", "datetime")
	spFile.SetCellValue(spSheet, "C3", "EndTime")
	spFile.SetCellValue(spSheet, "C4", "")

	// 第5赛季：StartTime=2026-05-15, EndTime=2026-06-14
	spFile.SetCellValue(spSheet, "A5", "5")
	spFile.SetCellValue(spSheet, "B5", "2026-05-15 00:00:00")
	spFile.SetCellValue(spSheet, "C5", "2026-06-14 23:59:59")

	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	rewardSheet := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", rewardSheet)

	rewardFile.SetCellValue(rewardSheet, "A1", "")
	rewardFile.SetCellValue(rewardSheet, "A2", "")
	rewardFile.SetCellValue(rewardSheet, "A3", "Id")
	rewardFile.SetCellValue(rewardSheet, "A4", "")
	rewardFile.SetCellValue(rewardSheet, "B1", "")
	rewardFile.SetCellValue(rewardSheet, "B2", "")
	rewardFile.SetCellValue(rewardSheet, "B3", "SeasonPassId")
	rewardFile.SetCellValue(rewardSheet, "B4", "")
	rewardFile.SetCellValue(rewardSheet, "C1", "")
	rewardFile.SetCellValue(rewardSheet, "C2", "")
	rewardFile.SetCellValue(rewardSheet, "C3", "HighReward")
	rewardFile.SetCellValue(rewardSheet, "C4", "")

	// 赵奢作为第5赛季战令奖励：道具ID = 1000000 + 11903 = 1011903
	rewardFile.SetCellValue(rewardSheet, "A5", "1")
	rewardFile.SetCellValue(rewardSheet, "B5", "5")
	rewardFile.SetCellValue(rewardSheet, "C5", "{1011903;1}")

	// DropItem 表
	dropFile := excelize.NewFile()
	dropSheet := "DropItem"
	dropFile.SetSheetName("Sheet1", dropSheet)

	dropFile.SetCellValue(dropSheet, "A1", "")
	dropFile.SetCellValue(dropSheet, "A2", "")
	dropFile.SetCellValue(dropSheet, "A3", "Id")
	dropFile.SetCellValue(dropSheet, "A4", "")
	dropFile.SetCellValue(dropSheet, "B1", "")
	dropFile.SetCellValue(dropSheet, "B2", "")
	dropFile.SetCellValue(dropSheet, "B3", "Item")
	dropFile.SetCellValue(dropSheet, "B4", "")
	dropFile.SetCellValue(dropSheet, "C1", "")
	dropFile.SetCellValue(dropSheet, "C2", "datetime")
	dropFile.SetCellValue(dropSheet, "C3", "ValidDate")
	dropFile.SetCellValue(dropSheet, "C4", "")

	// 赵奢在掉落库中，ValidDate=2026-02-15（早于保护期截止2026-09-14）
	dropFile.SetCellValue(dropSheet, "A5", "1")
	dropFile.SetCellValue(dropSheet, "B5", "{1011903;1}")
	dropFile.SetCellValue(dropSheet, "C5", "2026-02-15 00:00:00")

	sheetMap = map[string]*excelize.File{
		heroSheet:   heroFile,
		spSheet:     spFile,
		rewardSheet: rewardFile,
		dropSheet:   dropFile,
	}

	return cols, sheetMap
}

// createGeneralProtectionTestData 创建大将军武将反向保护期测试数据
// 场景：大将军武将（HeroId=11904），ValidDate=2026-03-01，赛季EndTime=2026-06-14
func createGeneralProtectionTestData() (cols [][]string, sheetMap map[string]*excelize.File) {
	// Hero 表
	heroFile := excelize.NewFile()
	heroSheet := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheet)

	heroFile.SetCellValue(heroSheet, "A1", "")
	heroFile.SetCellValue(heroSheet, "A2", "")
	heroFile.SetCellValue(heroSheet, "A3", "Id")
	heroFile.SetCellValue(heroSheet, "A4", "")
	heroFile.SetCellValue(heroSheet, "B1", "")
	heroFile.SetCellValue(heroSheet, "B2", "")
	heroFile.SetCellValue(heroSheet, "B3", "Name")
	heroFile.SetCellValue(heroSheet, "B4", "")
	heroFile.SetCellValue(heroSheet, "E1", "")
	heroFile.SetCellValue(heroSheet, "E2", "")
	heroFile.SetCellValue(heroSheet, "E3", "HeroType")
	heroFile.SetCellValue(heroSheet, "E4", "")

	// 大将军武将
	heroFile.SetCellValue(heroSheet, "A5", "11904")
	heroFile.SetCellValue(heroSheet, "B5", "大将军武将A")
	heroFile.SetCellValue(heroSheet, "E5", "1")

	cols, _ = heroFile.GetCols(heroSheet)

	// ArenaScoreReward 表
	asrFile := excelize.NewFile()
	asrSheet := "ArenaScoreReward"
	asrFile.SetSheetName("Sheet1", asrSheet)

	asrFile.SetCellValue(asrSheet, "A1", "")
	asrFile.SetCellValue(asrSheet, "A2", "")
	asrFile.SetCellValue(asrSheet, "A3", "Id")
	asrFile.SetCellValue(asrSheet, "A4", "")
	asrFile.SetCellValue(asrSheet, "B1", "")
	asrFile.SetCellValue(asrSheet, "B2", "")
	asrFile.SetCellValue(asrSheet, "B3", "Season")
	asrFile.SetCellValue(asrSheet, "B4", "")
	asrFile.SetCellValue(asrSheet, "C1", "")
	asrFile.SetCellValue(asrSheet, "C2", "")
	asrFile.SetCellValue(asrSheet, "C3", "Dan")
	asrFile.SetCellValue(asrSheet, "C4", "")
	asrFile.SetCellValue(asrSheet, "D1", "")
	asrFile.SetCellValue(asrSheet, "D2", "")
	asrFile.SetCellValue(asrSheet, "D3", "DanName")
	asrFile.SetCellValue(asrSheet, "D4", "")
	asrFile.SetCellValue(asrSheet, "E1", "")
	asrFile.SetCellValue(asrSheet, "E2", "")
	asrFile.SetCellValue(asrSheet, "E3", "Reward")
	asrFile.SetCellValue(asrSheet, "E4", "")

	// 大将军奖励：DanName包含"大将军"
	asrFile.SetCellValue(asrSheet, "A5", "1")
	asrFile.SetCellValue(asrSheet, "B5", "1")
	asrFile.SetCellValue(asrSheet, "C5", "10")
	asrFile.SetCellValue(asrSheet, "D5", "大将军")
	asrFile.SetCellValue(asrSheet, "E5", "{1011904;1}")

	// ArenaSeason 表
	asFile := excelize.NewFile()
	asSheet := "ArenaSeason"
	asFile.SetSheetName("Sheet1", asSheet)

	asFile.SetCellValue(asSheet, "A1", "")
	asFile.SetCellValue(asSheet, "A2", "")
	asFile.SetCellValue(asSheet, "A3", "Id")
	asFile.SetCellValue(asSheet, "A4", "")
	asFile.SetCellValue(asSheet, "B1", "")
	asFile.SetCellValue(asSheet, "B2", "datetime")
	asFile.SetCellValue(asSheet, "B3", "SeasonStartTime")
	asFile.SetCellValue(asSheet, "B4", "")
	asFile.SetCellValue(asSheet, "C1", "")
	asFile.SetCellValue(asSheet, "C2", "datetime")
	asFile.SetCellValue(asSheet, "C3", "SeasonEndTime")
	asFile.SetCellValue(asSheet, "C4", "")

	// 赛季：EndTime=2026-06-14
	asFile.SetCellValue(asSheet, "A5", "1")
	asFile.SetCellValue(asSheet, "B5", "2026-05-01 00:00:00")
	asFile.SetCellValue(asSheet, "C5", "2026-06-14 23:59:59")

	// DropItem 表
	dropFile := excelize.NewFile()
	dropSheet := "DropItem"
	dropFile.SetSheetName("Sheet1", dropSheet)

	dropFile.SetCellValue(dropSheet, "A1", "")
	dropFile.SetCellValue(dropSheet, "A2", "")
	dropFile.SetCellValue(dropSheet, "A3", "Id")
	dropFile.SetCellValue(dropSheet, "A4", "")
	dropFile.SetCellValue(dropSheet, "B1", "")
	dropFile.SetCellValue(dropSheet, "B2", "")
	dropFile.SetCellValue(dropSheet, "B3", "Item")
	dropFile.SetCellValue(dropSheet, "B4", "")
	dropFile.SetCellValue(dropSheet, "C1", "")
	dropFile.SetCellValue(dropSheet, "C2", "datetime")
	dropFile.SetCellValue(dropSheet, "C3", "ValidDate")
	dropFile.SetCellValue(dropSheet, "C4", "")

	// 大将军武将在掉落库中，ValidDate=2026-03-01（早于赛季结束2026-06-14）
	dropFile.SetCellValue(dropSheet, "A5", "1")
	dropFile.SetCellValue(dropSheet, "B5", "{1011904;1}")
	dropFile.SetCellValue(dropSheet, "C5", "2026-03-01 00:00:00")

	sheetMap = map[string]*excelize.File{
		heroSheet: heroFile,
		asrSheet:  asrFile,
		asSheet:   asFile,
		dropSheet: dropFile,
	}

	return cols, sheetMap
}
