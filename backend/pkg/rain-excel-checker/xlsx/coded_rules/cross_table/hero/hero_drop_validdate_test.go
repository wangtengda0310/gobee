package hero

import (
	"fmt"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== getNextThursday5AM 测试 ====================

func TestGetNextThursday5AM_Wednesday(t *testing.T) {
	// 周三 → 本周四 5:00
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local) // 2026-05-13 是周三
	result := getNextThursday5AM(now)
	// 返回 UTC 时区，与 ParseDate 保持一致
	expected := time.Date(2026, 5, 14, 5, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result, "周三应返回本周四5:00")
}

func TestGetNextThursday5AM_ThursdayBefore5(t *testing.T) {
	// 周四 4:00 → 本周四 5:00
	now := time.Date(2026, 5, 14, 4, 0, 0, 0, time.Local) // 2026-05-14 是周四
	result := getNextThursday5AM(now)
	expected := time.Date(2026, 5, 14, 5, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result, "周四5点前应返回本周四5:00")
}

func TestGetNextThursday5AM_ThursdayAfter5(t *testing.T) {
	// 周四 6:00 → 下周四 5:00
	now := time.Date(2026, 5, 14, 6, 0, 0, 0, time.Local) // 2026-05-14 是周四
	result := getNextThursday5AM(now)
	expected := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result, "周四5点后应返回下周四5:00")
}

func TestGetNextThursday5AM_Friday(t *testing.T) {
	// 周五 → 下周四 5:00
	now := time.Date(2026, 5, 15, 10, 0, 0, 0, time.Local) // 2026-05-15 是周五
	result := getNextThursday5AM(now)
	expected := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result, "周五应返回下周四5:00")
}

// ==================== Check 规则测试 ====================

// TestHeroDropValidDate_NormalHeroPass 已开放普通武将 ValidDate <= 下周四5:00 → 通过
func TestHeroDropValidDate_NormalHeroPass(t *testing.T) {
	// 固定时间：周三，下一个周四是明天 5:00
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)
	nextThursday := time.Date(2026, 5, 14, 5, 0, 0, 0, time.Local)

	// DropItem ValidDate 设置在下周四之前
	validDate := nextThursday.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")

	cols, sheetMap := createDropValidDateTestData(validDate, "10805", "测试武将", true, "2020-01-01 00:00:00")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "ValidDate <= 下周四5:00 应该通过")
}

// TestHeroDropValidDate_EqualToNextThursday ValidDate 恰好等于下一个周四5:00 → 应该通过（<=）
// 复现 bug：时区不一致导致相等时间被误判为"超过"
func TestHeroDropValidDate_EqualToNextThursday(t *testing.T) {
	// 固定时间：2026-05-20 周三 10:01:30
	// 下一个周四 = 2026-05-21 05:00:00
	now := time.Date(2026, 5, 20, 10, 1, 30, 0, time.Local)

	// ValidDate 恰好等于下一个周四 5:00
	// 规则是 <=，相等应该通过
	validDate := "2026-05-21 05:00:00"

	cols, sheetMap := createDropValidDateTestData(validDate, "10805", "贾诩", true, "2020-01-01 00:00:00")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "ValidDate 等于下一个周四5:00 应该通过（规则是<=）")
	assert.Empty(t, result.ErrCells, "不应有错误单元格")
}

// TestHeroDropValidDate_NormalHeroFail 已开放普通武将 ValidDate > 下周四5:00 → 报错
func TestHeroDropValidDate_NormalHeroFail(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)
	nextThursday := time.Date(2026, 5, 14, 5, 0, 0, 0, time.Local)

	// ValidDate 设置在下周四之后
	validDate := nextThursday.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	cols, sheetMap := createDropValidDateTestData(validDate, "10805", "测试武将", true, "2020-01-01 00:00:00")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.False(t, result.Ok, "ValidDate > 下周四5:00 应该报错")
	assert.NotEmpty(t, result.ErrCells, "应该有错误单元格")
}

// TestHeroDropValidDate_SeasonPassHero 战令武将 → 跳过
func TestHeroDropValidDate_SeasonPassHero(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)

	// 武将ID=1001，在 SeasonPassReward 中作为战令武将
	heroItemId := 1001001
	validDate := "2026-06-01 00:00:00" // 远超下周四

	cols, sheetMap := createDropValidDateTestDataWithHeroId(validDate, heroItemId, "1001", "战令武将", true, "2020-01-01 00:00:00")
	// 添加 SeasonPassReward 表，将 1001 标记为战令武将
	addSeasonPassRewardSheet(sheetMap, 1001001, 1)

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "战令武将应该被跳过")
}

// TestHeroDropValidDate_GeneralHero 大将军武将 → 跳过
func TestHeroDropValidDate_GeneralHero(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)

	heroItemId := 1002002
	validDate := "2026-06-01 00:00:00"

	cols, sheetMap := createDropValidDateTestDataWithHeroId(validDate, heroItemId, "2002", "大将军武将", true, "2020-01-01 00:00:00")
	// 添加 ArenaScoreReward 表，将 2002 标记为大将军武将
	addArenaScoreRewardSheet(sheetMap, 1002002, "大将军段位")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "大将军武将应该被跳过")
}

// TestHeroDropValidDate_NotOpenedHero 未开放武将 → 跳过
func TestHeroDropValidDate_NotOpenedHero(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)
	validDate := "2026-06-01 00:00:00"

	cols, sheetMap := createDropValidDateTestData(validDate, "10805", "未开放武将", false, "")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "未开放武将应该被跳过")
}

// TestHeroDropValidDate_EmptyValidDate ValidDate 为空 → 跳过
func TestHeroDropValidDate_EmptyValidDate(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)

	cols, sheetMap := createDropValidDateTestData("", "10805", "测试武将", true, "2020-01-01 00:00:00")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "ValidDate 为空应该被跳过")
}

// TestHeroDropValidDate_NoHeroItem DropItem 不含武将道具 → 跳过
func TestHeroDropValidDate_NoHeroItem(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)

	// DropItem 中配置普通道具（非武将道具）
	cols, sheetMap := createDropValidDateTestDataWithItem("{1001;10}", "2026-06-01 00:00:00")
	addHeroSheet(sheetMap, "10805", "测试武将", true, "2020-01-01 00:00:00")

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.True(t, result.Ok, "非武将道具应该被跳过")
}

// TestHeroDropValidDate_MissingHeroSheet Hero 表不存在 → 报错
func TestHeroDropValidDate_MissingHeroSheet(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.Local)

	// 只创建 DropItem 表，不创建 Hero 表
	dropItemFile := excelize.NewFile()
	dropItemFile.SetSheetName("Sheet1", "DropItem")
	heroItemId := 1010805
	dropItemFile.SetCellValue("DropItem", "A1", "")
	dropItemFile.SetCellValue("DropItem", "A2", "")
	dropItemFile.SetCellValue("DropItem", "A3", "Item")
	dropItemFile.SetCellValue("DropItem", "A4", "")
	dropItemFile.SetCellValue("DropItem", "A5", "{1010805;1}")
	dropItemFile.SetCellValue("DropItem", "B1", "")
	dropItemFile.SetCellValue("DropItem", "B2", "")
	dropItemFile.SetCellValue("DropItem", "B3", "ValidDate")
	dropItemFile.SetCellValue("DropItem", "B4", "")
	dropItemFile.SetCellValue("DropItem", "B5", "2026-06-01 00:00:00")
	_ = heroItemId

	cols, _ := dropItemFile.GetCols("DropItem")
	sheetMap := map[string]*excelize.File{
		"DropItem": dropItemFile,
	}

	rule := new(HeroDropValidDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "DropItem",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		SheetMap:    sheetMap,
		Now:         now,
	})

	assert.False(t, result.Ok, "缺少 Hero 表应该报错")
	assert.Contains(t, result.Reason, "Hero", "错误信息应提及 Hero 表")
}

// ==================== 测试数据构建辅助函数 ====================

// createDropValidDateTestData 创建标准的测试数据（武将道具ID=1010805，对应 heroId=10805）
func createDropValidDateTestData(validDate, heroId, heroName string, isOpen bool, openDate string) ([][]string, map[string]*excelize.File) {
	heroItemId := 1010805
	return createDropValidDateTestDataWithHeroId(validDate, heroItemId, heroId, heroName, isOpen, openDate)
}

// createDropValidDateTestDataWithHeroId 使用指定武将道具ID创建测试数据
func createDropValidDateTestDataWithHeroId(validDate string, heroItemId int, heroId, heroName string, isOpen bool, openDate string) ([][]string, map[string]*excelize.File) {
	itemCfg := fmt.Sprintf("{%d;1}", heroItemId)
	cols, dropFile := createDropItemFile(itemCfg, validDate)

	sheetMap := map[string]*excelize.File{
		"DropItem": dropFile,
	}
	addHeroSheet(sheetMap, heroId, heroName, isOpen, openDate)

	return cols, sheetMap
}

// createDropValidDateTestDataWithItem 使用指定 Item 字符串创建测试数据
func createDropValidDateTestDataWithItem(itemCfg, validDate string) ([][]string, map[string]*excelize.File) {
	cols, dropFile := createDropItemFile(itemCfg, validDate)
	sheetMap := map[string]*excelize.File{
		"DropItem": dropFile,
	}
	return cols, sheetMap
}

func createDropItemFile(itemCfg, validDate string) ([][]string, *excelize.File) {
	dropFile := excelize.NewFile()
	dropFile.SetSheetName("Sheet1", "DropItem")
	dropFile.SetCellValue("DropItem", "A1", "")
	dropFile.SetCellValue("DropItem", "A2", "")
	dropFile.SetCellValue("DropItem", "A3", "Item")
	dropFile.SetCellValue("DropItem", "A4", "")
	dropFile.SetCellValue("DropItem", "A5", itemCfg)
	dropFile.SetCellValue("DropItem", "B1", "")
	dropFile.SetCellValue("DropItem", "B2", "")
	dropFile.SetCellValue("DropItem", "B3", "ValidDate")
	dropFile.SetCellValue("DropItem", "B4", "")
	dropFile.SetCellValue("DropItem", "B5", validDate)
	dropFile.SetCellValue("DropItem", "C1", "")
	dropFile.SetCellValue("DropItem", "C2", "")
	dropFile.SetCellValue("DropItem", "C3", "Name")
	dropFile.SetCellValue("DropItem", "C4", "")
	dropFile.SetCellValue("DropItem", "C5", "测试掉落项")

	cols, _ := dropFile.GetCols("DropItem")
	return cols, dropFile
}

func addHeroSheet(sheetMap map[string]*excelize.File, heroId, heroName string, isOpen bool, openDate string) {
	heroFile := excelize.NewFile()
	heroFile.SetSheetName("Sheet1", "Hero")
	heroFile.SetCellValue("Hero", "A1", "")
	heroFile.SetCellValue("Hero", "A2", "")
	heroFile.SetCellValue("Hero", "A3", "Id")
	heroFile.SetCellValue("Hero", "A4", "")
	heroFile.SetCellValue("Hero", "A5", heroId)
	heroFile.SetCellValue("Hero", "B1", "")
	heroFile.SetCellValue("Hero", "B2", "")
	heroFile.SetCellValue("Hero", "B3", "Name")
	heroFile.SetCellValue("Hero", "B4", "")
	heroFile.SetCellValue("Hero", "B5", heroName)
	heroFile.SetCellValue("Hero", "C1", "")
	heroFile.SetCellValue("Hero", "C2", "")
	heroFile.SetCellValue("Hero", "C3", "IsOpen")
	heroFile.SetCellValue("Hero", "C4", "")
	isOpenStr := "false"
	if isOpen {
		isOpenStr = "true"
	}
	heroFile.SetCellValue("Hero", "C5", isOpenStr)
	heroFile.SetCellValue("Hero", "D1", "")
	heroFile.SetCellValue("Hero", "D2", "")
	heroFile.SetCellValue("Hero", "D3", "OpenDate")
	heroFile.SetCellValue("Hero", "D4", "")
	heroFile.SetCellValue("Hero", "D5", openDate)

	sheetMap["Hero"] = heroFile
}

func addSeasonPassRewardSheet(sheetMap map[string]*excelize.File, heroItemId, seasonPassId int) {
	spFile := excelize.NewFile()
	spFile.SetSheetName("Sheet1", "SeasonPassReward")
	spFile.SetCellValue("SeasonPassReward", "A1", "")
	spFile.SetCellValue("SeasonPassReward", "A2", "")
	spFile.SetCellValue("SeasonPassReward", "A3", "SeasonPassId")
	spFile.SetCellValue("SeasonPassReward", "A4", "")
	spFile.SetCellValue("SeasonPassReward", "A5", seasonPassId)
	spFile.SetCellValue("SeasonPassReward", "B1", "")
	spFile.SetCellValue("SeasonPassReward", "B2", "")
	spFile.SetCellValue("SeasonPassReward", "B3", "HighReward")
	spFile.SetCellValue("SeasonPassReward", "B4", "")
	spFile.SetCellValue("SeasonPassReward", "B5", fmt.Sprintf("{%d;1}", heroItemId))

	sheetMap["SeasonPassReward"] = spFile

	// 同时需要 SeasonPass 表
	spInfoFile := excelize.NewFile()
	spInfoFile.SetSheetName("Sheet1", "SeasonPass")
	spInfoFile.SetCellValue("SeasonPass", "A1", "")
	spInfoFile.SetCellValue("SeasonPass", "A2", "")
	spInfoFile.SetCellValue("SeasonPass", "A3", "Id")
	spInfoFile.SetCellValue("SeasonPass", "A4", "")
	spInfoFile.SetCellValue("SeasonPass", "A5", seasonPassId)
	spInfoFile.SetCellValue("SeasonPass", "B1", "")
	spInfoFile.SetCellValue("SeasonPass", "B2", "")
	spInfoFile.SetCellValue("SeasonPass", "B3", "StartTime")
	spInfoFile.SetCellValue("SeasonPass", "B4", "")
	spInfoFile.SetCellValue("SeasonPass", "B5", "2026-01-01 00:00:00")
	spInfoFile.SetCellValue("SeasonPass", "C1", "")
	spInfoFile.SetCellValue("SeasonPass", "C2", "")
	spInfoFile.SetCellValue("SeasonPass", "C3", "EndTime")
	spInfoFile.SetCellValue("SeasonPass", "C4", "")
	spInfoFile.SetCellValue("SeasonPass", "C5", "2026-02-01 00:00:00")

	sheetMap["SeasonPass"] = spInfoFile
}

func addArenaScoreRewardSheet(sheetMap map[string]*excelize.File, heroItemId int, danName string) {
	arFile := excelize.NewFile()
	arFile.SetSheetName("Sheet1", "ArenaScoreReward")
	arFile.SetCellValue("ArenaScoreReward", "A1", "")
	arFile.SetCellValue("ArenaScoreReward", "A2", "")
	arFile.SetCellValue("ArenaScoreReward", "A3", "Season")
	arFile.SetCellValue("ArenaScoreReward", "A4", "")
	arFile.SetCellValue("ArenaScoreReward", "A5", "1")
	arFile.SetCellValue("ArenaScoreReward", "B1", "")
	arFile.SetCellValue("ArenaScoreReward", "B2", "")
	arFile.SetCellValue("ArenaScoreReward", "B3", "Dan")
	arFile.SetCellValue("ArenaScoreReward", "B4", "")
	arFile.SetCellValue("ArenaScoreReward", "B5", "10")
	arFile.SetCellValue("ArenaScoreReward", "C1", "")
	arFile.SetCellValue("ArenaScoreReward", "C2", "")
	arFile.SetCellValue("ArenaScoreReward", "C3", "DanName")
	arFile.SetCellValue("ArenaScoreReward", "C4", "")
	arFile.SetCellValue("ArenaScoreReward", "C5", danName)
	arFile.SetCellValue("ArenaScoreReward", "D1", "")
	arFile.SetCellValue("ArenaScoreReward", "D2", "")
	arFile.SetCellValue("ArenaScoreReward", "D3", "Reward")
	arFile.SetCellValue("ArenaScoreReward", "D4", "")
	arFile.SetCellValue("ArenaScoreReward", "D5", fmt.Sprintf("{%d;1}", heroItemId))

	sheetMap["ArenaScoreReward"] = arFile
}
