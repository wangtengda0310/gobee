// Package table 提供表级别的校验规则
// 本包中的规则针对单个 Excel 表的特定业务逻辑进行校验

package coded_rules

import (
	"encoding/json"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestSeasonPassHeroOpenCheckRule_Error 测试时间不一致场景（应该报错）
// 业务场景：武将开放时间(2026-02-15)与战令StartTime(2025-12-15)不是同一天
func TestSeasonPassHeroOpenCheckRule_Error(t *testing.T) {
	// 创建 SeasonPassReward 表数据
	cols, sheetMap := createSeasonPassHeroTestData_Error()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：时间不一致应该报错
	if result.Ok {
		t.Error("预期应该报错，但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}
}

// TestSeasonPassHeroOpenCheckRule_Success 测试时间一致场景（应该通过）
// 业务场景：武将开放时间与战令StartTime是同一天
func TestSeasonPassHeroOpenCheckRule_Success(t *testing.T) {
	// 创建 SeasonPassReward 表数据
	cols, sheetMap := createSeasonPassHeroTestData_Success()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：时间一致应该通过
	if !result.Ok {
		t.Errorf("预期应该通过，但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// createSeasonPassHeroTestData_Error 创建时间不一致的测试数据
// 战令StartTime: 2025-12-15, 武将OpenDate: 2026-02-15
func createSeasonPassHeroTestData_Error() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头（固定4行）
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行（第5行开始）
	// 武将道具ID = 1000000 + 武将ID，所以 1010803 = 武将ID 10803
	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", "2025-12-15 00:00:00")
	seasonPassFile.SetCellValue(spSheetName, "C5", "2026-02-14 23:59:59")

	// Hero 表
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-02-15 00:00:00") // 与战令StartTime不同天

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_Success 创建时间一致的测试数据
// 战令StartTime: 2025-12-15, 武将OpenDate: 2025-12-15 (同一天)
func createSeasonPassHeroTestData_Success() (cols [][]string, sheetMap map[string]*excelize.File) {
	// 复用 Error 数据，只修改 Hero 的 OpenDate
	cols, sheetMap = createSeasonPassHeroTestData_Error()

	// 修改 Hero 的 OpenDate 为与战令StartTime同一天
	heroFile := sheetMap["Hero"]
	heroFile.SetCellValue("Hero", "D5", "2025-12-15 00:00:00")

	return cols, sheetMap
}

// TestSeasonPassHeroOpenCheckRule_EmptyOpenDate_SeasonNotEnded 测试赛季未结束且OpenDate为空的场景
// 业务场景：赛季未结束（结束于未来时间），武将OpenDate为空，应该报错
func TestSeasonPassHeroOpenCheckRule_EmptyOpenDate_SeasonNotEnded(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_EmptyOpenDate_SeasonNotEnded()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：赛季未结束且OpenDate为空应该报错
	if result.Ok {
		t.Error("预期应该报错（赛季未结束时OpenDate不能为空），但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}
}

// TestSeasonPassHeroOpenCheckRule_EmptyOpenDate_SeasonEnded 测试赛季已结束且OpenDate为空的场景
// 业务场景：赛季已结束（结束于过去时间），武将OpenDate为空，应该通过
func TestSeasonPassHeroOpenCheckRule_EmptyOpenDate_SeasonEnded(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_EmptyOpenDate_SeasonEnded()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：赛季已结束且OpenDate为空应该通过
	if !result.Ok {
		t.Errorf("预期应该通过（赛季已结束时OpenDate可以为空），但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// createSeasonPassHeroTestData_EmptyOpenDate_SeasonNotEnded 模拟赛季未结束且OpenDate为空的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK（战令武将开放时间检查）- 赛季未结束时OpenDate不能为空
// 业务场景：赛季结束时间设置为未来时间（当前时间+30天），武将OpenDate为空
func createSeasonPassHeroTestData_EmptyOpenDate_SeasonNotEnded() (cols [][]string, sheetMap map[string]*excelize.File) {
	// 直接创建测试数据，不复用 Error 数据
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头（固定4行）
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行（第5行开始）
	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表 - 结束时间设置为未来30天
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", "2025-12-15 00:00:00")
	seasonPassFile.SetCellValue(spSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 为空
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "") // OpenDate 为空

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_EmptyOpenDate_SeasonEnded 模拟赛季已结束且OpenDate为空的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK（战令武将开放时间检查）- 赛季已结束时OpenDate可以为空
// 业务场景：赛季结束时间设置为过去时间（当前时间-30天），武将OpenDate为空
func createSeasonPassHeroTestData_EmptyOpenDate_SeasonEnded() (cols [][]string, sheetMap map[string]*excelize.File) {
	// 直接创建测试数据，不复用 Error 数据
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头（固定4行）
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行（第5行开始）
	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表 - 结束时间设置为过去30天
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	pastEndTime := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", "2025-12-15 00:00:00")
	seasonPassFile.SetCellValue(spSheetName, "C5", pastEndTime)

	// Hero 表 - OpenDate 为空
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "") // OpenDate 为空

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// TestSeasonPassHeroOpenCheckRule_NotEnded_MatchedTime 测试赛季未结束且时间匹配的场景
// 业务场景：赛季未结束，武将OpenDate与StartTime同一天，应该通过
func TestSeasonPassHeroOpenCheckRule_NotEnded_MatchedTime(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_NotEnded_MatchedTime()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：赛季未结束但时间匹配应该通过
	if !result.Ok {
		t.Errorf("预期应该通过（时间匹配），但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// TestSeasonPassHeroOpenCheckRule_Ended_MatchedTime 测试赛季已结束且时间匹配的场景
// 业务场景：赛季已结束，武将OpenDate与StartTime同一天，应该通过
func TestSeasonPassHeroOpenCheckRule_Ended_MatchedTime(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_Ended_MatchedTime()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：赛季已结束且时间匹配应该通过
	if !result.Ok {
		t.Errorf("预期应该通过（时间匹配），但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// TestSeasonPassHeroOpenCheckRule_Ended_MismatchedTime 测试赛季已结束但时间不匹配的场景
// 业务场景：赛季已结束，武将OpenDate与StartTime不是同一天，应该报错（原规则仍然有效）
func TestSeasonPassHeroOpenCheckRule_Ended_MismatchedTime(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_Ended_MismatchedTime()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：时间不匹配应该报错
	if result.Ok {
		t.Error("预期应该报错（时间不匹配），但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}
}

// createSeasonPassHeroTestData_NotEnded_MatchedTime 模拟赛季未结束且时间匹配的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 赛季未结束时OpenDate必须与StartTime同一天
// 业务场景：赛季结束时间为未来30天，武将OpenDate与StartTime同一天
func createSeasonPassHeroTestData_NotEnded_MatchedTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表 - 未来结束时间
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	startTime := "2026-03-25 00:00:00" // OpenDate将与此相同
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", startTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 与 StartTime 匹配
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", startTime) // 与 StartTime 同一天

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_Ended_MatchedTime 模拟赛季已结束且时间匹配的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 赛季已结束时OpenDate可以为空或匹配
// 业务场景：赛季结束时间为过去30天，武将OpenDate与StartTime同一天
func createSeasonPassHeroTestData_Ended_MatchedTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表 - 过去结束时间
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	startTime := "2026-02-01 00:00:00" // OpenDate将与此相同
	pastEndTime := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", startTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", pastEndTime)

	// Hero 表 - OpenDate 与 StartTime 匹配
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", startTime) // 与 StartTime 同一天

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// TestSeasonPassHeroOpenCheckRule_SecondPrecision_SameDayExactMatch 测试同一天精确时间匹配场景
// 业务场景：StartTime和OpenDate在同一天且完全相同（精确到秒），应该通过
func TestSeasonPassHeroOpenCheckRule_SecondPrecision_SameDayExactMatch(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_SecondPrecision_SameDayExactMatch()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：同一天且完全相同的时间应该通过
	if !result.Ok {
		t.Errorf("预期应该通过（同一天精确匹配），但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// TestSeasonPassHeroOpenCheckRule_SecondPrecision_SameDayDifferentSecond 测试同一天但秒数不同场景
// 业务场景：StartTime和OpenDate在同一天但秒数不同（12:30:45 vs 12:30:46），应该报错（精确到秒比较）
func TestSeasonPassHeroOpenCheckRule_SecondPrecision_SameDayDifferentSecond(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_SecondPrecision_SameDayDifferentSecond()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：同一天但秒数不同，应该报错（精确到秒比较）
	if result.Ok {
		t.Error("预期应该报错（秒数不匹配），但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// TestSeasonPassHeroOpenCheckRule_SecondPrecision_DifferentDaySameTime 测试不同天但时间相同场景
// 业务场景：StartTime和OpenDate在不同天但时间相同（都是00:00:00），应该报错
func TestSeasonPassHeroOpenCheckRule_SecondPrecision_DifferentDaySameTime(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_SecondPrecision_DifferentDaySameTime()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：不同天应该报错
	if result.Ok {
		t.Error("预期应该报错（不同天），但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}
}

// TestSeasonPassHeroOpenCheckRule_SecondPrecision_OneSecondDifferentDay 测试相差一秒但不同天场景
// 业务场景：StartTime和OpenDate相差一秒但属于不同天（23:59:59 vs 00:00:00），应该报错
func TestSeasonPassHeroOpenCheckRule_SecondPrecision_OneSecondDifferentDay(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_SecondPrecision_OneSecondDifferentDay()

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：不同天应该报错（即使只相差一秒）
	if result.Ok {
		t.Error("预期应该报错（不同天，即使只相差一秒），但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// createSeasonPassHeroTestData_SecondPrecision_SameDayExactMatch 模拟同一天精确时间匹配的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 时间比较精度
// 业务场景：StartTime和OpenDate完全相同（2026-03-25 12:30:45），验证精确匹配场景
func createSeasonPassHeroTestData_SecondPrecision_SameDayExactMatch() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表 - 使用精确时间
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	// 使用精确时间（包含时、分、秒）
	exactTime := "2026-03-25 12:30:45"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", exactTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 与 StartTime 完全相同
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", exactTime) // 与 StartTime 完全相同

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_SecondPrecision_SameDayDifferentSecond 模拟同一天但秒数不同的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 时间比较精度（当前只比较日期）
// 业务场景：StartTime=2026-03-25 12:30:45，OpenDate=2026-03-25 12:30:46，相差一秒但同一天
func createSeasonPassHeroTestData_SecondPrecision_SameDayDifferentSecond() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	startTime := "2026-03-25 12:30:45"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", startTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 相差一秒
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-25 12:30:46") // 相差一秒

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_SecondPrecision_DifferentDaySameTime 模拟不同天但时间相同的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 时间比较精度（跨天边界）
// 业务场景：StartTime=2026-03-25 00:00:00，OpenDate=2026-03-26 00:00:00，时间相同但不同天
func createSeasonPassHeroTestData_SecondPrecision_DifferentDaySameTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	startTime := "2026-03-25 00:00:00"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", startTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 不同天但时间相同
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-26 00:00:00") // 不同天但时间相同

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_SecondPrecision_OneSecondDifferentDay 模拟相差一秒但不同天的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 时间比较精度（跨天边界，相差一秒）
// 业务场景：StartTime=2026-03-25 23:59:59，OpenDate=2026-03-26 00:00:00，只相差一秒但跨越两天
func createSeasonPassHeroTestData_SecondPrecision_OneSecondDifferentDay() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	// 第一天的最后一秒
	startTime := "2026-03-25 23:59:59"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", startTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 为第二天的第一秒
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-26 00:00:00") // 只相差一秒但不同天

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// createSeasonPassHeroTestData_Ended_MismatchedTime 模拟赛季已结束但时间不匹配的测试数据
// 验证规则：SEASON_PASS_HERO_OPEN_CHECK - 已结束赛季的时间不匹配仍需报错（原规则仍然有效）
// 业务场景：赛季结束时间为过去30天，武将OpenDate与StartTime不是同一天
func createSeasonPassHeroTestData_Ended_MismatchedTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	sheetName := "SeasonPassReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "SeasonPassId")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "Level")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "HighReward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "1")
	rewardFile.SetCellValue(sheetName, "C5", "{1010803;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// SeasonPass 表 - 过去结束时间
	seasonPassFile := excelize.NewFile()
	spSheetName := "SeasonPass"
	seasonPassFile.SetSheetName("Sheet1", spSheetName)

	seasonPassFile.SetCellValue(spSheetName, "A1", "")
	seasonPassFile.SetCellValue(spSheetName, "A2", "")
	seasonPassFile.SetCellValue(spSheetName, "A3", "Id")
	seasonPassFile.SetCellValue(spSheetName, "A4", "")
	seasonPassFile.SetCellValue(spSheetName, "B1", "")
	seasonPassFile.SetCellValue(spSheetName, "B2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "B3", "StartTime")
	seasonPassFile.SetCellValue(spSheetName, "B4", "")
	seasonPassFile.SetCellValue(spSheetName, "C1", "")
	seasonPassFile.SetCellValue(spSheetName, "C2", "datetime")
	seasonPassFile.SetCellValue(spSheetName, "C3", "EndTime")
	seasonPassFile.SetCellValue(spSheetName, "C4", "")

	startTime := "2026-02-01 00:00:00"
	pastEndTime := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	seasonPassFile.SetCellValue(spSheetName, "A5", "1")
	seasonPassFile.SetCellValue(spSheetName, "B5", startTime)
	seasonPassFile.SetCellValue(spSheetName, "C5", pastEndTime)

	// Hero 表 - OpenDate 与 StartTime 不匹配
	heroFile := excelize.NewFile()
	heroSheetName := "Hero"
	heroFile.SetSheetName("Sheet1", heroSheetName)

	heroFile.SetCellValue(heroSheetName, "A1", "")
	heroFile.SetCellValue(heroSheetName, "A2", "")
	heroFile.SetCellValue(heroSheetName, "A3", "Id")
	heroFile.SetCellValue(heroSheetName, "A4", "")
	heroFile.SetCellValue(heroSheetName, "B1", "")
	heroFile.SetCellValue(heroSheetName, "B2", "")
	heroFile.SetCellValue(heroSheetName, "B3", "Name")
	heroFile.SetCellValue(heroSheetName, "B4", "")
	heroFile.SetCellValue(heroSheetName, "C1", "")
	heroFile.SetCellValue(heroSheetName, "C2", "")
	heroFile.SetCellValue(heroSheetName, "C3", "IsOpen")
	heroFile.SetCellValue(heroSheetName, "C4", "")
	heroFile.SetCellValue(heroSheetName, "D1", "")
	heroFile.SetCellValue(heroSheetName, "D2", "datetime")
	heroFile.SetCellValue(heroSheetName, "D3", "OpenDate")
	heroFile.SetCellValue(heroSheetName, "D4", "")

	heroFile.SetCellValue(heroSheetName, "A5", "10803")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-01 00:00:00") // 与 StartTime 不同天

	sheetMap = map[string]*excelize.File{
		sheetName:     rewardFile,
		spSheetName:   seasonPassFile,
		heroSheetName: heroFile,
	}

	return cols, sheetMap
}

// TestSeasonPassHeroOpenCheckRule_IsOpenFalse 测试战令武将 IsOpen=false 的场景
func TestSeasonPassHeroOpenCheckRule_IsOpenFalse(t *testing.T) {
	cols, sheetMap := createSeasonPassHeroTestData_EmptyOpenDate_SeasonNotEnded()
	heroFile := sheetMap["Hero"]
	heroFile.SetCellValue("Hero", "C5", "false")

	rule := new(SeasonPassHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "SeasonPassReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "IsOpen=false 应该报错")
	assert.NotEmpty(t, result.ErrCells, "应该有错误记录")
	assert.Contains(t, result.ErrCells[0].Reason, "IsOpen=false")
}
