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

// TestArenaGeneralHeroOpenCheckRule_TimeMismatch 测试时间不匹配场景
// 业务场景：武将开放时间与赛季StartTime不是同一天
func TestArenaGeneralHeroOpenCheckRule_TimeMismatch(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_TimeMismatch()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：时间不匹配应该报错
	if result.Ok {
		t.Error("预期应该报错，但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// TestArenaGeneralHeroOpenCheckRule_Success 测试时间匹配场景
// 业务场景：武将开放时间与赛季StartTime是同一天
func TestArenaGeneralHeroOpenCheckRule_Success(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_Success()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：时间匹配应该通过
	if !result.Ok {
		t.Errorf("预期应该通过，但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// createArenaGeneralHeroTestData_TimeMismatch 创建时间不匹配的测试数据
func createArenaGeneralHeroTestData_TimeMismatch() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行：大将军奖励包含武将
	rewardFile.SetCellValue(sheetName, "A5", "1") // Season ID
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}") // 武将ID 10804

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 使用未来时间，确保赛季不会被时间过滤逻辑跳过
	// 赛季开始时间：当前时间+10天
	seasonStartTime := time.Now().AddDate(0, 0, 10).Format("2006-01-02 15:04:05")
	// 赛季结束时间：当前时间+30天
	seasonEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", seasonStartTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", seasonEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	// OpenDate 与赛季StartTime不同天（赛季StartTime+10天，OpenDate设置为+15天）
	openDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02 15:04:05")
	heroFile.SetCellValue(heroSheetName, "D5", openDate)

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_Success 创建时间匹配的测试数据
func createArenaGeneralHeroTestData_Success() (cols [][]string, sheetMap map[string]*excelize.File) {
	cols, sheetMap = createArenaGeneralHeroTestData_TimeMismatch()

	// 修改 Hero 的 OpenDate 为与赛季StartTime同一天
	seasonStartTime := time.Now().AddDate(0, 0, 10).Format("2006-01-02 15:04:05")
	heroFile := sheetMap["Hero"]
	heroFile.SetCellValue("Hero", "D5", seasonStartTime)

	return cols, sheetMap
}

// TestArenaGeneralHeroOpenCheckRule_EmptyOpenDate_SeasonNotEnded 测试赛季未结束且OpenDate为空的场景
// 业务场景：赛季未结束（结束于未来时间），武将OpenDate为空，应该报错
func TestArenaGeneralHeroOpenCheckRule_EmptyOpenDate_SeasonNotEnded(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_EmptyOpenDate_SeasonNotEnded()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// TestArenaGeneralHeroOpenCheckRule_EmptyOpenDate_SeasonEnded 测试赛季已结束且OpenDate为空的场景
// 业务场景：赛季已结束（结束于过去时间），武将OpenDate为空，应该通过
func TestArenaGeneralHeroOpenCheckRule_EmptyOpenDate_SeasonEnded(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_EmptyOpenDate_SeasonEnded()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// createArenaGeneralHeroTestData_EmptyOpenDate_SeasonNotEnded 模拟赛季未结束且OpenDate为空的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK（大将军武将开放时间检查）- 赛季未结束时OpenDate不能为空
// 业务场景：赛季结束时间设置为未来时间（当前时间+30天），武将OpenDate为空
func createArenaGeneralHeroTestData_EmptyOpenDate_SeasonNotEnded() (cols [][]string, sheetMap map[string]*excelize.File) {
	// 直接创建测试数据
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行：大将军奖励包含武将
	rewardFile.SetCellValue(sheetName, "A5", "1") // Season ID
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}") // 武将ID 10804

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 赛季结束时间：未来30天
	seasonEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonStartTime := time.Now().AddDate(0, 0, 10).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", seasonStartTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", seasonEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "") // OpenDate 为空

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_EmptyOpenDate_SeasonEnded 模拟赛季已结束且OpenDate为空的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK（大将军武将开放时间检查）- 赛季已结束时OpenDate可以为空
// 业务场景：赛季结束时间设置为过去时间（当前时间-30天，确保小于warnThreshold），武将OpenDate为空
func createArenaGeneralHeroTestData_EmptyOpenDate_SeasonEnded() (cols [][]string, sheetMap map[string]*excelize.File) {
	// 直接创建测试数据
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行：大将军奖励包含武将
	rewardFile.SetCellValue(sheetName, "A5", "1") // Season ID
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}") // 武将ID 10804

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 赛季结束时间：过去30天（确保小于 warnThreshold = 当前时间-5天）
	seasonEndTime := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	seasonStartTime := time.Now().AddDate(0, 0, -60).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", seasonStartTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", seasonEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "") // OpenDate 为空

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// TestArenaGeneralHeroOpenCheckRule_NotEnded_MatchedTime 测试赛季未结束且时间匹配场景
// 业务场景：赛季未结束时，武将OpenDate与赛季StartTime同一天，应该通过
func TestArenaGeneralHeroOpenCheckRule_NotEnded_MatchedTime(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_NotEnded_MatchedTime()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：赛季未结束且时间匹配应该通过
	if !result.Ok {
		t.Errorf("预期应该通过，但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// TestArenaGeneralHeroOpenCheckRule_Ended_MatchedTime 测试赛季已结束且时间匹配场景
// 业务场景：赛季已结束时，武将OpenDate与赛季StartTime同一天，应该通过
func TestArenaGeneralHeroOpenCheckRule_Ended_MatchedTime(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_Ended_MatchedTime()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：赛季已结束且时间匹配应该通过
	if !result.Ok {
		t.Errorf("预期应该通过，但检查失败了: %s", result.Reason)
	}
	if len(result.ErrCells) > 0 {
		t.Errorf("预期不应该有错误记录，但有 %d 个", len(result.ErrCells))
	}
}

// TestArenaGeneralHeroOpenCheckRule_Ended_MismatchedTime 测试赛季已结束但时间不匹配场景
// 业务场景：赛季已结束时，武将OpenDate与赛季StartTime不是同一天，应该报错
func TestArenaGeneralHeroOpenCheckRule_Ended_MismatchedTime(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_Ended_MismatchedTime()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 验证结果：即使赛季已结束，OpenDate不为空时也需要检查时间是否匹配
	if result.Ok {
		t.Error("预期应该报错（时间不匹配），但检查通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("预期应该有错误记录，但没有")
	}

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))
}

// createArenaGeneralHeroTestData_NotEnded_MatchedTime 模拟赛季未结束且时间匹配的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK（大将军武将开放时间检查）- 赛季未结束时OpenDate必须配置且匹配
// 业务场景：赛季结束时间设置为未来时间（当前时间+30天），武将OpenDate与StartTime同一天
func createArenaGeneralHeroTestData_NotEnded_MatchedTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行：大将军奖励包含武将
	rewardFile.SetCellValue(sheetName, "A5", "99") // Season ID
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1099904;1}") // 武将ID 99904

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 赛季结束时间：未来30天（赛季未结束）
	seasonEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonStartTime := time.Now().AddDate(0, 0, 10).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "99")
	seasonFile.SetCellValue(seasonSheetName, "B5", seasonStartTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", seasonEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "99904")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", seasonStartTime) // OpenDate 与 StartTime 同一天

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_Ended_MatchedTime 模拟赛季已结束且时间匹配的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK（大将军武将开放时间检查）- 赛季已结束时OpenDate可以为空或匹配
// 业务场景：赛季结束时间设置为过去时间（当前时间-30天），武将OpenDate与StartTime同一天
func createArenaGeneralHeroTestData_Ended_MatchedTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行：大将军奖励包含武将
	rewardFile.SetCellValue(sheetName, "A5", "98") // Season ID
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1099903;1}") // 武将ID 99903

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 赛季结束时间：过去30天（赛季已结束）
	seasonEndTime := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	seasonStartTime := time.Now().AddDate(0, 0, -60).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "98")
	seasonFile.SetCellValue(seasonSheetName, "B5", seasonStartTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", seasonEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "99903")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", seasonStartTime) // OpenDate 与 StartTime 同一天

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// TestArenaGeneralHeroOpenCheckRule_SecondPrecision_SameDayExactMatch 测试同一天精确时间匹配场景
// 业务场景：SeasonStartTime和OpenDate在同一天且完全相同（精确到秒），应该通过
func TestArenaGeneralHeroOpenCheckRule_SecondPrecision_SameDayExactMatch(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_SecondPrecision_SameDayExactMatch()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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

// TestArenaGeneralHeroOpenCheckRule_SecondPrecision_SameDayDifferentSecond 测试同一天但秒数不同场景
// 业务场景：SeasonStartTime和OpenDate在同一天但秒数不同（12:30:45 vs 12:30:46），应该报错（精确到秒比较）
func TestArenaGeneralHeroOpenCheckRule_SecondPrecision_SameDayDifferentSecond(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_SecondPrecision_SameDayDifferentSecond()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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
	t.Logf("Result: Ok=%v, Reason=%s", result.Ok, result.Reason)
	if len(result.ErrCells) > 0 {
		for _, err := range result.ErrCells {
			t.Logf("Error: %s", err.Reason)
		}
	}
}

// TestArenaGeneralHeroOpenCheckRule_SecondPrecision_DifferentDaySameTime 测试不同天但时间相同场景
// 业务场景：SeasonStartTime和OpenDate在不同天但时间相同（都是00:00:00），应该报错
func TestArenaGeneralHeroOpenCheckRule_SecondPrecision_DifferentDaySameTime(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_SecondPrecision_DifferentDaySameTime()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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

// TestArenaGeneralHeroOpenCheckRule_SecondPrecision_OneSecondDifferentDay 测试相差一秒但不同天场景
// 业务场景：SeasonStartTime和OpenDate相差一秒但属于不同天（23:59:59 vs 00:00:00），应该报错
func TestArenaGeneralHeroOpenCheckRule_SecondPrecision_OneSecondDifferentDay(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_SecondPrecision_OneSecondDifferentDay()

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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
	t.Logf("Result: Ok=%v, Reason=%s", result.Ok, result.Reason)
	if len(result.ErrCells) > 0 {
		for _, err := range result.ErrCells {
			t.Logf("Error: %s", err.Reason)
		}
	}
}

// createArenaGeneralHeroTestData_SecondPrecision_SameDayExactMatch 模拟同一天精确时间匹配的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK - 时间比较精度
// 业务场景：SeasonStartTime和OpenDate完全相同（2026-03-25 12:30:45），验证精确匹配场景
func createArenaGeneralHeroTestData_SecondPrecision_SameDayExactMatch() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表 - 使用精确时间
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 使用精确时间（包含时、分、秒）
	exactTime := "2026-03-25 12:30:45"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", exactTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", futureEndTime)

	// Hero 表 - OpenDate 与 SeasonStartTime 完全相同
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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", exactTime) // 与 SeasonStartTime 完全相同

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_SecondPrecision_SameDayDifferentSecond 模拟同一天但秒数不同的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK - 时间比较精度（当前只比较日期）
// 业务场景：SeasonStartTime=2026-03-25 12:30:45，OpenDate=2026-03-25 12:30:46，相差一秒但同一天
func createArenaGeneralHeroTestData_SecondPrecision_SameDayDifferentSecond() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	startTime := "2026-03-25 12:30:45"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", startTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", futureEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-25 12:30:46") // 相差一秒

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_SecondPrecision_DifferentDaySameTime 模拟不同天但时间相同的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK - 时间比较精度（跨天边界）
// 业务场景：SeasonStartTime=2026-03-25 00:00:00，OpenDate=2026-03-26 00:00:00，时间相同但不同天
func createArenaGeneralHeroTestData_SecondPrecision_DifferentDaySameTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	startTime := "2026-03-25 00:00:00"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", startTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", futureEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-26 00:00:00") // 不同天但时间相同

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_SecondPrecision_OneSecondDifferentDay 模拟相差一秒但不同天的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK - 时间比较精度（跨天边界，相差一秒）
// 业务场景：SeasonStartTime=2026-03-25 23:59:59，OpenDate=2026-03-26 00:00:00，只相差一秒但跨越两天
func createArenaGeneralHeroTestData_SecondPrecision_OneSecondDifferentDay() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	rewardFile.SetCellValue(sheetName, "A5", "1")
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1010804;1}")

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 第一天的最后一秒
	startTime := "2026-03-25 23:59:59"
	futureEndTime := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "1")
	seasonFile.SetCellValue(seasonSheetName, "B5", startTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", futureEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "10804")
	heroFile.SetCellValue(heroSheetName, "B5", "卫青")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	heroFile.SetCellValue(heroSheetName, "D5", "2026-03-26 00:00:00") // 只相差一秒但不同天

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// createArenaGeneralHeroTestData_Ended_MismatchedTime 模拟赛季已结束但时间不匹配的测试数据
// 验证规则：ARENA_GENERAL_HERO_OPEN_CHECK（大将军武将开放时间检查）- 即使赛季已结束，OpenDate不为空时也需要检查
// 业务场景：赛季结束时间设置为过去时间（当前时间-30天），武将OpenDate与StartTime不是同一天
func createArenaGeneralHeroTestData_Ended_MismatchedTime() (cols [][]string, sheetMap map[string]*excelize.File) {
	// ArenaScoreReward 表
	rewardFile := excelize.NewFile()
	sheetName := "ArenaScoreReward"
	rewardFile.SetSheetName("Sheet1", sheetName)

	// 表头
	rewardFile.SetCellValue(sheetName, "A1", "")
	rewardFile.SetCellValue(sheetName, "A2", "")
	rewardFile.SetCellValue(sheetName, "A3", "Season")
	rewardFile.SetCellValue(sheetName, "A4", "")
	rewardFile.SetCellValue(sheetName, "B1", "")
	rewardFile.SetCellValue(sheetName, "B2", "")
	rewardFile.SetCellValue(sheetName, "B3", "DanName")
	rewardFile.SetCellValue(sheetName, "B4", "")
	rewardFile.SetCellValue(sheetName, "C1", "")
	rewardFile.SetCellValue(sheetName, "C2", "")
	rewardFile.SetCellValue(sheetName, "C3", "Reward")
	rewardFile.SetCellValue(sheetName, "C4", "")

	// 数据行：大将军奖励包含武将
	rewardFile.SetCellValue(sheetName, "A5", "98") // Season ID
	rewardFile.SetCellValue(sheetName, "B5", "大将军")
	rewardFile.SetCellValue(sheetName, "C5", "{1099902;1}") // 武将ID 99902

	cols, _ = rewardFile.GetCols(sheetName)

	// ArenaSeason 表
	seasonFile := excelize.NewFile()
	seasonSheetName := "ArenaSeason"
	seasonFile.SetSheetName("Sheet1", seasonSheetName)

	seasonFile.SetCellValue(seasonSheetName, "A1", "")
	seasonFile.SetCellValue(seasonSheetName, "A2", "")
	seasonFile.SetCellValue(seasonSheetName, "A3", "Id")
	seasonFile.SetCellValue(seasonSheetName, "A4", "")
	seasonFile.SetCellValue(seasonSheetName, "B1", "")
	seasonFile.SetCellValue(seasonSheetName, "B2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "B3", "SeasonStartTime")
	seasonFile.SetCellValue(seasonSheetName, "B4", "")
	seasonFile.SetCellValue(seasonSheetName, "C1", "")
	seasonFile.SetCellValue(seasonSheetName, "C2", "datetime")
	seasonFile.SetCellValue(seasonSheetName, "C3", "SeasonEndTime")
	seasonFile.SetCellValue(seasonSheetName, "C4", "")

	// 赛季结束时间：过去30天（赛季已结束）
	seasonEndTime := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	seasonStartTime := time.Now().AddDate(0, 0, -60).Format("2006-01-02 15:04:05")
	seasonFile.SetCellValue(seasonSheetName, "A5", "98")
	seasonFile.SetCellValue(seasonSheetName, "B5", seasonStartTime)
	seasonFile.SetCellValue(seasonSheetName, "C5", seasonEndTime)

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

	heroFile.SetCellValue(heroSheetName, "A5", "99902")
	heroFile.SetCellValue(heroSheetName, "B5", "测试武将")
	heroFile.SetCellValue(heroSheetName, "C5", "true")
	// OpenDate 与 StartTime 不同天（相差10天）
	mismatchedOpenDate := time.Now().AddDate(0, 0, -50).Format("2006-01-02 15:04:05")
	heroFile.SetCellValue(heroSheetName, "D5", mismatchedOpenDate)

	sheetMap = map[string]*excelize.File{
		sheetName:       rewardFile,
		seasonSheetName: seasonFile,
		heroSheetName:   heroFile,
	}

	return cols, sheetMap
}

// TestArenaGeneralHeroOpenCheckRule_IsOpenFalse 测试大将军武将 IsOpen=false 的场景
func TestArenaGeneralHeroOpenCheckRule_IsOpenFalse(t *testing.T) {
	cols, sheetMap := createArenaGeneralHeroTestData_EmptyOpenDate_SeasonNotEnded()
	heroFile := sheetMap["Hero"]
	heroFile.SetCellValue("Hero", "C5", "false")

	rule := new(ArenaGeneralHeroOpenCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaScoreReward",
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
