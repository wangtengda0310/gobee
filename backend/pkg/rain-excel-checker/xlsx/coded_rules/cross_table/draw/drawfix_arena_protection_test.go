package draw

import (
	"fmt"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// buildArenaProtectionSheetMap 构建包含大将军保护期检查所有关联表的 sheetMap
//
// 参数：
//   - drawFixRows: DrawFix 表数据行，每行 [Id, Name, StartTime, EndTime, ItemIds, SlotNum]
//   - itemRows: Item 表数据行，每行 [Id, Type, ItemParam]
//   - arenaScoreRewardRows: ArenaScoreReward 表数据行，每行 [Season, Dan, DanName, Reward]
//   - arenaSeasonRows: ArenaSeason 表数据行，每行 [Id, SeasonStartTime, SeasonEndTime]
//   - heroRows: Hero 表数据行，每行 [Id, Name]
func buildArenaProtectionSheetMap(
	drawFixRows [][]string,
	itemRows [][]string,
	arenaScoreRewardRows [][]string,
	arenaSeasonRows [][]string,
	heroRows [][]string,
) map[string]*excelize.File {
	sheetMap := make(map[string]*excelize.File)

	// DrawFix 表
	if drawFixRows != nil {
		dfFile := excelize.NewFile()
		dfSheet := "定向招募|DrawFix"
		_, _ = dfFile.NewSheet(dfSheet)
		_ = dfFile.SetCellValue(dfSheet, "A3", "Id")
		_ = dfFile.SetCellValue(dfSheet, "B3", "Name")
		_ = dfFile.SetCellValue(dfSheet, "C3", "StartTime")
		_ = dfFile.SetCellValue(dfSheet, "D3", "EndTime")
		_ = dfFile.SetCellValue(dfSheet, "E3", "ItemIds")
		_ = dfFile.SetCellValue(dfSheet, "F3", "SlotNum")
		for i, row := range drawFixRows {
			rowNum := 5 + i
			cols := []string{"A", "B", "C", "D", "E", "F"}
			for j, col := range cols {
				if j < len(row) {
					_ = dfFile.SetCellValue(dfSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[dfSheet] = dfFile
	}

	// Item 表
	if itemRows != nil {
		itemFile := excelize.NewFile()
		itemSheet := "道具|Item"
		_, _ = itemFile.NewSheet(itemSheet)
		_ = itemFile.SetCellValue(itemSheet, "A3", "Id")
		_ = itemFile.SetCellValue(itemSheet, "B3", "Type")
		_ = itemFile.SetCellValue(itemSheet, "C3", "ItemParam")
		for i, row := range itemRows {
			rowNum := 5 + i
			cols := []string{"A", "B", "C"}
			for j, col := range cols {
				if j < len(row) {
					_ = itemFile.SetCellValue(itemSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[itemSheet] = itemFile
	}

	// ArenaScoreReward 表（注意：单数 Reward）
	if arenaScoreRewardRows != nil {
		asrFile := excelize.NewFile()
		asrSheet := "竞技场积分奖励表|ArenaScoreReward"
		_, _ = asrFile.NewSheet(asrSheet)
		_ = asrFile.SetCellValue(asrSheet, "A3", "Season")
		_ = asrFile.SetCellValue(asrSheet, "B3", "Dan")
		_ = asrFile.SetCellValue(asrSheet, "C3", "DanName")
		_ = asrFile.SetCellValue(asrSheet, "D3", "Reward")
		for i, row := range arenaScoreRewardRows {
			rowNum := 5 + i
			cols := []string{"A", "B", "C", "D"}
			for j, col := range cols {
				if j < len(row) {
					_ = asrFile.SetCellValue(asrSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[asrSheet] = asrFile
	}

	// ArenaSeason 表
	if arenaSeasonRows != nil {
		asFile := excelize.NewFile()
		asSheet := "竞技场赛季|ArenaSeason"
		_, _ = asFile.NewSheet(asSheet)
		_ = asFile.SetCellValue(asSheet, "A3", "Id")
		_ = asFile.SetCellValue(asSheet, "B3", "SeasonStartTime")
		_ = asFile.SetCellValue(asSheet, "C3", "SeasonEndTime")
		for i, row := range arenaSeasonRows {
			rowNum := 5 + i
			cols := []string{"A", "B", "C"}
			for j, col := range cols {
				if j < len(row) {
					_ = asFile.SetCellValue(asSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[asSheet] = asFile
	}

	// Hero 表
	if heroRows != nil {
		heroFile := excelize.NewFile()
		heroSheet := "武将|Hero"
		_, _ = heroFile.NewSheet(heroSheet)
		_ = heroFile.SetCellValue(heroSheet, "A3", "Id")
		_ = heroFile.SetCellValue(heroSheet, "B3", "Name")
		for i, row := range heroRows {
			rowNum := 5 + i
			cols := []string{"A", "B"}
			for j, col := range cols {
				if j < len(row) {
					_ = heroFile.SetCellValue(heroSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[heroSheet] = heroFile
	}

	return sheetMap
}

// runArenaProtectionCheck 执行大将军保护期检查
func runArenaProtectionCheck(sheetMap map[string]*excelize.File, now time.Time) *json_rule.TableCheckResult {
	drawFixSheet := "定向招募|DrawFix"
	drawFixFile := sheetMap[drawFixSheet]
	cols, _ := drawFixFile.GetCols(drawFixSheet)

	rule := &DrawFixArenaProtectionCheckRule{}
	return rule.Check(json_rule.CheckParam{
		SheetName:   drawFixSheet,
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    20,
		SheetMap:    sheetMap,
		Now:         now,
	})
}

// ==================== 测试用例 ====================

// TestDrawFixArenaProtection_Meta 验证规则元数据
func TestDrawFixArenaProtection_Meta(t *testing.T) {
	rule := &DrawFixArenaProtectionCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.DRAWFIX_ARENA_PROTECTION_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawFix")
	assert.Contains(t, meta.RequiredSheets, "Item")
	assert.Contains(t, meta.RequiredSheets, "ArenaScoreReward")
	assert.Contains(t, meta.RequiredSheets, "ArenaSeason")
	assert.Contains(t, meta.RequiredSheets, "Hero")
}

// TestDrawFixArenaProtection_ReverseViolation 反向违规 - DrawFix 在保护期内且包含大将军武将
func TestDrawFixArenaProtection_ReverseViolation(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: EndTime=2025-06-01, ItemIds 包含大将军武将道具 1010001
		[][]string{{"1001", "违规招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		// Item: 1010001 是 Hero 道具，对应 HeroId=10001
		[][]string{{"1010001", "Hero", "10001"}},
		// ArenaScoreReward: Season=1, Dan=23, DanName="大将军", Reward 包含武将道具
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		// ArenaSeason: SeasonEndTime=2025-12-01 (保护期至赛季结束)
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		// Hero
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.False(t, result.Ok, "DrawFix 在保护期内包含大将军武将应失败")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "大将军")
}

// TestDrawFixArenaProtection_ReverseNoViolation 反向通过 - 赛季已结束
func TestDrawFixArenaProtection_ReverseNoViolation(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: EndTime=2025-06-01
		[][]string{{"1001", "安全招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		// ArenaSeason: SeasonEndTime=2025-03-01 (赛季已结束)
		[][]string{{"1", "2025-01-01 00:00:00", "2025-03-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, result.Ok, "赛季已结束应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_ReverseExactBoundary 边界测试 - EndTime 恰好等于 SeasonEndTime
func TestDrawFixArenaProtection_ReverseExactBoundary(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: EndTime=2025-06-01
		[][]string{{"1001", "边界招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		// ArenaSeason: SeasonEndTime=2025-06-01 (与 DrawFix.EndTime 相同)
		[][]string{{"1", "2025-01-01 00:00:00", "2025-06-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, result.Ok, "EndTime == SeasonEndTime 应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_PostProtectionNoConfig 保护期后未配置 - 赛季已结束，DrawFix 不包含大将军武将，应通过（保护期后不强制配置）
func TestDrawFixArenaProtection_PostProtectionNoConfig(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 不包含大将军武将
		[][]string{{"1001", "无大将军招募", "2025-01-01 00:00:00", "2025-07-01 00:00:00", "{1020001;0}", "1"}},
		[][]string{{"1020001", "Hero", "10002"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		// ArenaSeason: SeasonEndTime=2025-01-01
		[][]string{{"1", "2025-01-01 00:00:00", "2025-01-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}, {"10002", "普通武将"}},
	)

	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	result := runArenaProtectionCheck(sheetMap, now)
	assert.True(t, result.Ok, "赛季结束后不强制要求大将军武将出现在定向招募中")
	assert.Len(t, result.ErrCells, 0)
	// 赛季结束后不报错，不检查 ErrCells
}

// TestDrawFixArenaProtection_PostProtectionWithConfig 保护期后已配置 - 赛季已结束，DrawFix 包含大将军武将，应通过
func TestDrawFixArenaProtection_PostProtectionWithConfig(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 包含大将军武将道具
		[][]string{{"1001", "含大将军招募", "2025-01-01 00:00:00", "2025-07-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		// ArenaSeason: SeasonEndTime=2025-01-01
		[][]string{{"1", "2025-01-01 00:00:00", "2025-01-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	result := runArenaProtectionCheck(sheetMap, now)
	assert.True(t, result.Ok, "赛季已结束且大将军武将已出现应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_ActiveSeasonNoGeneralHero 赛季进行中无大将军武将 - DrawFix 不包含大将军武将，应通过（反向检查不违规）
func TestDrawFixArenaProtection_ActiveSeasonNoGeneralHero(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 不包含大将军武将
		[][]string{{"1001", "赛季中招募", "2025-01-01 00:00:00", "2025-07-01 00:00:00", "{1020001;0}", "1"}},
		[][]string{{"1020001", "Hero", "10002"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		// ArenaSeason: SeasonEndTime=2025-12-01 (在未来)
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}, {"10002", "普通武将"}},
	)

	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	result := runArenaProtectionCheck(sheetMap, now)
	assert.True(t, result.Ok, "赛季未结束应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_NonHeroItem 非武将道具应跳过
func TestDrawFixArenaProtection_NonHeroItem(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 包含非 Hero 类型道具
		[][]string{{"1001", "道具招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{2000001;0}", "1"}},
		// Item: 2000001 Type=Material
		[][]string{{"2000001", "Material", "0"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, result.Ok, "非 Hero 类型道具应跳过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_NonGeneralHero 武将不在大将军名单中应通过
func TestDrawFixArenaProtection_NonGeneralHero(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 包含 HeroId=10002 的道具
		[][]string{{"1001", "非大将军招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1020001;0}", "1"}},
		// Item: 1020001 对应 HeroId=10002
		[][]string{{"1020001", "Hero", "10002"}},
		// ArenaScoreReward: 只有 HeroId=10001 是大将军
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}, {"10002", "普通武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, result.Ok, "非大将军武将应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_NoItemSheet Item 表缺失应失败
func TestDrawFixArenaProtection_NoItemSheet(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		[][]string{{"1001", "测试招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		nil, // 无 Item 表
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.False(t, result.Ok, "Item 表缺失应失败")
	assert.Contains(t, result.Reason, "Item")
}

// TestDrawFixArenaProtection_NoArenaSheets Arena 表全部缺失应通过（无大将军可检查）
func TestDrawFixArenaProtection_NoArenaSheets(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		[][]string{{"1001", "测试招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		nil, // 无 ArenaScoreReward
		nil, // 无 ArenaSeason
		[][]string{{"10001", "武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, result.Ok, "无 Arena 数据时应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_EmptyItemIds ItemIds 为空应跳过
func TestDrawFixArenaProtection_EmptyItemIds(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 为空
		[][]string{{"1001", "空招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, result.Ok, "ItemIds 为空应跳过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixArenaProtection_MultipleGeneralHeroes 多个大将军武将部分违规
func TestDrawFixArenaProtection_MultipleGeneralHeroes(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 包含两个大将军武将
		[][]string{{"1001", "多武将招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}{1010002;0}", "2"}},
		// Item: 两个 Hero 道具
		[][]string{{"1010001", "Hero", "10001"}, {"1010002", "Hero", "10002"}},
		// ArenaScoreReward: 两个赛季的大将军
		[][]string{{"1", "23", "大将军", "{1010001;1}"}, {"2", "23", "大将军", "{1010002;1}"}},
		// ArenaSeason: Season=2 已结束
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}, {"2", "2025-01-01 00:00:00", "2025-03-01 00:00:00"}},
		[][]string{{"10001", "赛季1大将军"}, {"10002", "赛季2大将军"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.False(t, result.Ok, "有违规的大将军武将应失败")
	// 赛季1未结束，HeroId=10001 违规；赛季2已结束，HeroId=10002 不违规
	assert.Contains(t, result.ErrCells[0].Reason, "赛季1大将军")
}

// TestDrawFixArenaProtection_MultipleDrawFixRows 多行 DrawFix，部分违规部分合规
func TestDrawFixArenaProtection_MultipleDrawFixRows(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// 两行 DrawFix: 第一行违规，第二行合规
		[][]string{
			{"1001", "违规招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"},
			{"1002", "安全招募", "2025-06-02 00:00:00", "2025-07-01 00:00:00", "{1010001;0}", "1"},
		},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		[][]string{{"10001", "大将军武将"}},
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.False(t, result.Ok, "有违规行应失败")
	assert.Len(t, result.ErrCells, 2, "两行都违规")
	assert.Contains(t, result.ErrCells[0].Reason, "违规招募")
	assert.Contains(t, result.ErrCells[1].Reason, "安全招募")
}

// TestDrawFixArenaProtection_HeroNameFallback Hero 表缺失时显示 "未知(ID=xxx)"
func TestDrawFixArenaProtection_HeroNameFallback(t *testing.T) {
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 包含大将军武将
		[][]string{{"1001", "未知武将招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		// Item: 1010001 对应 HeroId=10001
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "23", "大将军", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-12-01 00:00:00"}},
		nil, // 无 Hero 表
	)

	result := runArenaProtectionCheck(sheetMap, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.False(t, result.Ok, "Hero 表缺失时仍应检测违规")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "未知(ID=10001)")
}
