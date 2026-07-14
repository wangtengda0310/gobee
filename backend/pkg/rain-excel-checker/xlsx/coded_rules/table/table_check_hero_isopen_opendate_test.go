package coded_rules

import (
	"fmt"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 辅助函数 ====================

// setHeroHeader 设置 Hero 表的列头（复用）
// MJS_FIXED_ROWS_NUM=4，前4行是元数据，第3行(index=2)是列名
func setHeroHeader(f *excelize.File, s string) {
	f.SetCellValue(s, "A1", "")
	f.SetCellValue(s, "A2", "")
	f.SetCellValue(s, "A3", "Id")
	f.SetCellValue(s, "A4", "")
	f.SetCellValue(s, "B1", "")
	f.SetCellValue(s, "B2", "")
	f.SetCellValue(s, "B3", "Name")
	f.SetCellValue(s, "B4", "")
	f.SetCellValue(s, "C1", "")
	f.SetCellValue(s, "C2", "")
	f.SetCellValue(s, "C3", "IsOpen")
	f.SetCellValue(s, "C4", "")
	f.SetCellValue(s, "D1", "")
	f.SetCellValue(s, "D2", "datetime")
	f.SetCellValue(s, "D3", "OpenDate")
	f.SetCellValue(s, "D4", "")
}

// buildIsOpenOpendateSheetMap 构建基础 SheetMap（Hero + 可选 SeasonPassReward + 可选 ArenaScoreReward）
func buildIsOpenOpendateSheetMap(heroFile *excelize.File, rewardFile *excelize.File, asrFile *excelize.File) map[string]*excelize.File {
	sheetMap := map[string]*excelize.File{
		"武将|Hero": heroFile,
	}
	if rewardFile != nil {
		sheetMap["SeasonPassReward"] = rewardFile
	}
	if asrFile != nil {
		sheetMap["ArenaScoreReward"] = asrFile
	}
	return sheetMap
}

// buildSeasonPassSheetMap 构造包含战令武将的完整 SheetMap
// heroId: 战令武将ID（如 10618）
// spName/spStart/spEnd: 战令信息
func buildSeasonPassSheetMap(heroFile *excelize.File, heroId int, spName, spStart, spEnd string) map[string]*excelize.File {
	// 武将道具ID = 1000000 + 武将ID
	heroItemId := 1000000 + heroId

	// SeasonPassReward 表
	rewardFile := excelize.NewFile()
	rewardFile.SetSheetName("Sheet1", "SeasonPassReward")
	rs := "SeasonPassReward"
	rewardFile.SetCellValue(rs, "A1", "")
	rewardFile.SetCellValue(rs, "A2", "")
	rewardFile.SetCellValue(rs, "A3", "SeasonPassId")
	rewardFile.SetCellValue(rs, "A4", "")
	rewardFile.SetCellValue(rs, "B1", "")
	rewardFile.SetCellValue(rs, "B2", "")
	rewardFile.SetCellValue(rs, "B3", "HighReward")
	rewardFile.SetCellValue(rs, "B4", "")
	rewardFile.SetCellValue(rs, "A5", "1")
	rewardFile.SetCellValue(rs, "B5", fmt.Sprintf("{%d;1}", heroItemId))

	// SeasonPass 表
	spFile := excelize.NewFile()
	spFile.SetSheetName("Sheet1", "SeasonPass")
	sps := "SeasonPass"
	spFile.SetCellValue(sps, "A1", "")
	spFile.SetCellValue(sps, "A2", "")
	spFile.SetCellValue(sps, "A3", "Id")
	spFile.SetCellValue(sps, "A4", "")
	spFile.SetCellValue(sps, "B1", "")
	spFile.SetCellValue(sps, "B2", "")
	spFile.SetCellValue(sps, "B3", "Name")
	spFile.SetCellValue(sps, "B4", "")
	spFile.SetCellValue(sps, "C1", "")
	spFile.SetCellValue(sps, "C2", "datetime")
	spFile.SetCellValue(sps, "C3", "StartTime")
	spFile.SetCellValue(sps, "C4", "")
	spFile.SetCellValue(sps, "D1", "")
	spFile.SetCellValue(sps, "D2", "datetime")
	spFile.SetCellValue(sps, "D3", "EndTime")
	spFile.SetCellValue(sps, "D4", "")
	spFile.SetCellValue(sps, "A5", "1")
	spFile.SetCellValue(sps, "B5", spName)
	spFile.SetCellValue(sps, "C5", spStart)
	spFile.SetCellValue(sps, "D5", spEnd)

	sheetMap := buildIsOpenOpendateSheetMap(heroFile, rewardFile, nil)
	sheetMap["SeasonPass"] = spFile
	return sheetMap
}

// buildArenaGeneralSheetMap 构造包含大将军武将的完整 SheetMap
// heroId: 大将军武将ID（如 10613）
func buildArenaGeneralSheetMap(heroFile *excelize.File, heroId int) map[string]*excelize.File {
	heroItemId := 1000000 + heroId

	// ArenaScoreReward 表
	asrFile := excelize.NewFile()
	asrFile.SetSheetName("Sheet1", "ArenaScoreReward")
	as := "ArenaScoreReward"
	asrFile.SetCellValue(as, "A1", "")
	asrFile.SetCellValue(as, "A2", "")
	asrFile.SetCellValue(as, "A3", "Season")
	asrFile.SetCellValue(as, "A4", "")
	asrFile.SetCellValue(as, "B1", "")
	asrFile.SetCellValue(as, "B2", "")
	asrFile.SetCellValue(as, "B3", "Dan")
	asrFile.SetCellValue(as, "B4", "")
	asrFile.SetCellValue(as, "C1", "")
	asrFile.SetCellValue(as, "C2", "")
	asrFile.SetCellValue(as, "C3", "DanName")
	asrFile.SetCellValue(as, "C4", "")
	asrFile.SetCellValue(as, "D1", "")
	asrFile.SetCellValue(as, "D2", "")
	asrFile.SetCellValue(as, "D3", "Reward")
	asrFile.SetCellValue(as, "D4", "")
	asrFile.SetCellValue(as, "A5", "1")
	asrFile.SetCellValue(as, "B5", "1")
	asrFile.SetCellValue(as, "C5", "大将军")
	asrFile.SetCellValue(as, "D5", fmt.Sprintf("{%d;1}", heroItemId))

	return buildIsOpenOpendateSheetMap(heroFile, nil, asrFile)
}

// createHeroData 创建 Hero 表测试数据并返回列数据
// heroId, name, isOpen, openDate: 武将行数据
func createHeroData(heroId, name, isOpen, openDate string) (*excelize.File, [][]string) {
	heroFile := excelize.NewFile()
	heroFile.SetSheetName("Sheet1", "武将|Hero")
	s := "武将|Hero"
	setHeroHeader(heroFile, s)
	heroFile.SetCellValue(s, "A5", heroId)
	heroFile.SetCellValue(s, "B5", name)
	heroFile.SetCellValue(s, "C5", isOpen)
	heroFile.SetCellValue(s, "D5", openDate)
	cols, _ := heroFile.GetCols(s)
	return heroFile, cols
}

// ==================== Bug 复现测试 ====================

// TestHeroIsOpenOpenDate_NormalHeroNotChecked 验证普通武将（非战令/大将军）IsOpen=true 但无 OpenDate 不报错
// 复现 bug：之前所有武将 IsOpen=true 无 OpenDate 都报错，普通武将（如曹操 ID=10001）一直在线上是正常的
func TestHeroIsOpenOpenDate_NormalHeroNotChecked(t *testing.T) {
	// 构造 Hero 表：普通武将（ID=10001 曹操）IsOpen=true，OpenDate 为空
	heroFile, heroCols := createHeroData("10001", "曹操", "true", "")

	// SheetMap 中只有 Hero 表，无战令/大将军关联表
	sheetMap := buildIsOpenOpendateSheetMap(heroFile, nil, nil)

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "普通武将 IsOpen=true 无 OpenDate 不应报错")
	assert.Empty(t, result.ErrCells)
}

// ==================== 验证修复测试 ====================

// TestHeroIsOpenOpenDate_SeasonPassHeroMissingOpenDate 验证战令武将 IsOpen=true 无 OpenDate 会报错
func TestHeroIsOpenOpenDate_SeasonPassHeroMissingOpenDate(t *testing.T) {
	heroFile, heroCols := createHeroData("10618", "嬴政", "true", "")
	sheetMap := buildSeasonPassSheetMap(heroFile, 10618, "战令1", "2026-03-01 00:00:00", "2026-03-14 00:00:00")

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "战令武将 IsOpen=true 无 OpenDate 应报错")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "10618")
	assert.Contains(t, result.ErrCells[0].Reason, "嬴政")
}

// TestHeroIsOpenOpenDate_GeneralHeroMissingOpenDate 验证大将军武将 IsOpen=true 无 OpenDate 会报错
func TestHeroIsOpenOpenDate_GeneralHeroMissingOpenDate(t *testing.T) {
	heroFile, heroCols := createHeroData("10613", "李信", "true", "")
	sheetMap := buildArenaGeneralSheetMap(heroFile, 10613)

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "大将军武将 IsOpen=true 无 OpenDate 应报错")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "10613")
}

// TestHeroIsOpenOpenDate_MixedHeroes 混合场景：普通武将+战令武将同时存在
// 验证只有战令武将被检查，普通武将被跳过
func TestHeroIsOpenOpenDate_MixedHeroes(t *testing.T) {
	heroFile := excelize.NewFile()
	heroFile.SetSheetName("Sheet1", "武将|Hero")
	s := "武将|Hero"
	setHeroHeader(heroFile, s)
	// 普通武将：曹操 IsOpen=true，无 OpenDate → 不应报错
	heroFile.SetCellValue(s, "A5", "10001")
	heroFile.SetCellValue(s, "B5", "曹操")
	heroFile.SetCellValue(s, "C5", "true")
	heroFile.SetCellValue(s, "D5", "")
	// 战令武将：嬴政 IsOpen=true，无 OpenDate → 应报错
	heroFile.SetCellValue(s, "A6", "10618")
	heroFile.SetCellValue(s, "B6", "嬴政")
	heroFile.SetCellValue(s, "C6", "true")
	heroFile.SetCellValue(s, "D6", "")
	// 普通武将：赵云 IsOpen=true，无 OpenDate → 不应报错
	heroFile.SetCellValue(s, "A7", "10003")
	heroFile.SetCellValue(s, "B7", "赵云")
	heroFile.SetCellValue(s, "C7", "true")
	heroFile.SetCellValue(s, "D7", "")
	heroCols, _ := heroFile.GetCols(s)

	// 构造 SeasonPassReward：嬴政(10618) 是战令武将
	sheetMap := buildSeasonPassSheetMap(heroFile, 10618, "战令1", "2026-03-01 00:00:00", "2026-03-14 00:00:00")

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 应只报嬴政（战令武将），不报曹操和赵云（普通武将）
	assert.False(t, result.Ok)
	assert.Len(t, result.ErrCells, 1, "应只有1个错误（嬴政）")
	assert.Contains(t, result.ErrCells[0].Reason, "10618")
}

// TestHeroIsOpenOpenDate_SpecialHeroWithOpenDate 验证战令武将 IsOpen=true 且有 OpenDate 不报错
func TestHeroIsOpenOpenDate_SpecialHeroWithOpenDate(t *testing.T) {
	heroFile, heroCols := createHeroData("10618", "嬴政", "true", "2026-03-15 00:00:00")
	sheetMap := buildSeasonPassSheetMap(heroFile, 10618, "战令1", "2026-03-01 00:00:00", "2026-03-14 00:00:00")

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "战令武将 IsOpen=true 且有 OpenDate 不应报错")
	assert.Empty(t, result.ErrCells)
}

// TestHeroIsOpenOpenDate_NoSheetMap 验证无 SheetMap 时不报错（无法确定特殊武将列表）
func TestHeroIsOpenOpenDate_NoSheetMap(t *testing.T) {
	_, heroCols := createHeroData("10001", "曹操", "true", "")

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
	})

	assert.True(t, result.Ok, "无 SheetMap 时无法获取特殊武将列表，不应报错")
}

// TestHeroIsOpenOpenDate_IsOpenFalseWithEmptyDate 验证 IsOpen=false 时无论 OpenDate 是否为空都不报错
func TestHeroIsOpenOpenDate_IsOpenFalseWithEmptyDate(t *testing.T) {
	heroFile, heroCols := createHeroData("10618", "嬴政", "false", "")
	sheetMap := buildSeasonPassSheetMap(heroFile, 10618, "战令1", "2026-03-01 00:00:00", "2026-03-14 00:00:00")

	rule := new(HeroIsOpenOpenDateCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将|Hero",
		Cols:        heroCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "战令武将 IsOpen=false 时不应报错")
	assert.Empty(t, result.ErrCells)
}
