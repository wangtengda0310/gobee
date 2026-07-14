package draw

import (
	"fmt"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// buildProtectionSheetMap 构建包含所有关联表的 sheetMap
//
// 参数：
//   - drawFixRows: DrawFix 表数据行，每行 [Id, Name, StartTime, EndTime, ItemIds, SlotNum]
//   - itemRows: Item 表数据行，每行 [Id, Type, ItemParam]
//   - spRewardRows: SeasonPassReward 表数据行，每行 [SeasonPassId, HighReward]
//   - spRows: SeasonPass 表数据行，每行 [Id, StartTime, EndTime]
//   - heroRows: Hero 表数据行，每行 [Id, Name]
func buildProtectionSheetMap(
	drawFixRows [][]string,
	itemRows [][]string,
	spRewardRows [][]string,
	spRows [][]string,
	heroRows [][]string,
) map[string]*excelize.File {
	sheetMap := make(map[string]*excelize.File)

	// DrawFix 表
	if drawFixRows != nil {
		dfFile := excelize.NewFile()
		dfSheet := "定向招募|DrawFix"
		_, _ = dfFile.NewSheet(dfSheet)
		// 字段名写在第3行
		_ = dfFile.SetCellValue(dfSheet, "A3", "Id")
		_ = dfFile.SetCellValue(dfSheet, "B3", "Name")
		_ = dfFile.SetCellValue(dfSheet, "C3", "StartTime")
		_ = dfFile.SetCellValue(dfSheet, "D3", "EndTime")
		_ = dfFile.SetCellValue(dfSheet, "E3", "ItemIds")
		_ = dfFile.SetCellValue(dfSheet, "F3", "SlotNum")
		// 数据从第5行开始
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

	// SeasonPassReward 表
	if spRewardRows != nil {
		sprFile := excelize.NewFile()
		sprSheet := "战令奖励|SeasonPassReward"
		_, _ = sprFile.NewSheet(sprSheet)
		_ = sprFile.SetCellValue(sprSheet, "A3", "SeasonPassId")
		_ = sprFile.SetCellValue(sprSheet, "B3", "HighReward")
		for i, row := range spRewardRows {
			rowNum := 5 + i
			cols := []string{"A", "B"}
			for j, col := range cols {
				if j < len(row) {
					_ = sprFile.SetCellValue(sprSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[sprSheet] = sprFile
	}

	// SeasonPass 表
	if spRows != nil {
		spFile := excelize.NewFile()
		spSheet := "战令|SeasonPass"
		_, _ = spFile.NewSheet(spSheet)
		_ = spFile.SetCellValue(spSheet, "A3", "Id")
		_ = spFile.SetCellValue(spSheet, "B3", "StartTime")
		_ = spFile.SetCellValue(spSheet, "C3", "EndTime")
		for i, row := range spRows {
			rowNum := 5 + i
			cols := []string{"A", "B", "C"}
			for j, col := range cols {
				if j < len(row) {
					_ = spFile.SetCellValue(spSheet, fmt.Sprintf("%s%d", col, rowNum), row[j])
				}
			}
		}
		sheetMap[spSheet] = spFile
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

// runProtectionCheck 从 sheetMap 中获取 DrawFix 数据并执行检查
func runProtectionCheck(sheetMap map[string]*excelize.File, params map[string]string) *json_rule.TableCheckResult {
	drawFixSheet := "定向招募|DrawFix"
	drawFixFile := sheetMap[drawFixSheet]
	cols, _ := drawFixFile.GetCols(drawFixSheet)

	rule := &DrawFixProtectionCheckRule{}
	return rule.Check(json_rule.CheckParam{
		SheetName:   drawFixSheet,
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    10,
		SheetMap:    sheetMap,
		Params:      params,
	})
}

// ==================== 测试用例 ====================

// TestDrawFixProtection_Meta 验证规则元数据
func TestDrawFixProtection_Meta(t *testing.T) {
	rule := &DrawFixProtectionCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.DRAWFIX_PROTECTION_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawFix")
	assert.Contains(t, meta.RequiredSheets, "Item")
	assert.Contains(t, meta.RequiredSheets, "SeasonPassReward")
	assert.Contains(t, meta.RequiredSheets, "SeasonPass")
	assert.Contains(t, meta.RequiredSheets, "Hero")
}

// TestDrawFixProtection_NoViolation DrawFix 结束时间在保护期之后，应通过
func TestDrawFixProtection_NoViolation(t *testing.T) {
	// SeasonPass StartTime=2025-01-01, protectMonths=4 → 保护期至 2025-05-01
	// DrawFix EndTime=2025-06-01 → 在保护期之后，通过
	sheetMap := buildProtectionSheetMap(
		// DrawFix 行: Id=1001, Name=安全招募, EndTime=2025-06-01, ItemIds={1010001;0}
		[][]string{{"1001", "安全招募", "2025-01-01 00:00:00", "2025-06-01 00:00:00", "{1010001;0}", "1"}},
		// Item: 道具ID 1010001, Type=Hero, ItemParam=10001(HeroId)
		[][]string{{"1010001", "Hero", "10001"}},
		// SeasonPassReward: SeasonPassId=1, HighReward 包含武将道具
		[][]string{{"1", "{1010001;1}"}},
		// SeasonPass: Id=1, StartTime=2025-01-01
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		// Hero: Id=10001
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.True(t, result.Ok, "EndTime 在保护期之后应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_Violation DrawFix 结束时间在保护期内，应失败
func TestDrawFixProtection_Violation(t *testing.T) {
	// SeasonPass StartTime=2025-01-01, protectMonths=4 → 保护期至 2025-05-01
	// DrawFix EndTime=2025-03-01 → 在保护期内，违规
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "测试招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.False(t, result.Ok, "EndTime 在保护期内应失败")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "测试招募")
	assert.Contains(t, result.ErrCells[0].Reason, "测试武将")
	assert.Contains(t, result.ErrCells[0].Reason, "保护期")
}

// TestDrawFixProtection_NonHeroItem ItemIds 包含非武将道具，应跳过
func TestDrawFixProtection_NonHeroItem(t *testing.T) {
	// 道具ID 2000001 不是 Hero 类型（Item 表 Type 不是 "Hero"），应跳过
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "道具招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{2000001;0}", "1"}},
		// Item: 2000001 Type=Material，不是 Hero
		[][]string{{"2000001", "Material", "0"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.True(t, result.Ok, "非武将道具应被跳过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_NonSeasonPassHero 武将不在战令列表中，应通过
func TestDrawFixProtection_NonSeasonPassHero(t *testing.T) {
	// ItemIds 中的道具 ID 在 Item 表标记为 Hero，但对应武将不在战令奖励中
	sheetMap := buildProtectionSheetMap(
		// DrawFix: 包含道具 1010002 (对应 HeroId 10002)
		[][]string{{"1001", "非战令招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010002;0}", "1"}},
		// Item: 1010002 是 Hero 道具，对应 HeroId=10002
		[][]string{{"1010002", "Hero", "10002"}},
		// SeasonPassReward: 战令奖励是 1010001（HeroId=10001），不包含 10002
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "战令武将"}, {"10002", "非战令武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.True(t, result.Ok, "非战令武将不在保护期映射中应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_EmptyItemIds ItemIds 为空，应跳过该行
func TestDrawFixProtection_EmptyItemIds(t *testing.T) {
	sheetMap := buildProtectionSheetMap(
		// DrawFix: ItemIds 为空
		[][]string{{"1001", "空招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.True(t, result.Ok, "ItemIds 为空应跳过该行")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_CustomProtectMonths 使用自定义保护期月数
func TestDrawFixProtection_CustomProtectMonths(t *testing.T) {
	// SeasonPass StartTime=2025-01-01, protectMonths=2 → 保护期至 2025-03-01
	// DrawFix EndTime=2025-04-01 → 在保护期之后，通过
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "自定义保护期", "2025-01-01 00:00:00", "2025-04-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, map[string]string{
		string(json_rule.PROTECT_MONTHS): "2",
	})
	assert.True(t, result.Ok, "protectMonths=2 时 EndTime=2025-04-01 在保护期(2025-03-01)之后应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_NoItemSheet Item 表不存在，应返回失败
func TestDrawFixProtection_NoItemSheet(t *testing.T) {
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "测试招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010001;0}", "1"}},
		nil, // 无 Item 表
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.False(t, result.Ok, "Item 表不存在应返回失败")
	assert.Contains(t, result.Reason, "Item")
}

// TestDrawFixProtection_NoSeasonPassSheet SeasonPass 表不存在，应通过（无战令武将可检查）
func TestDrawFixProtection_NoSeasonPassSheet(t *testing.T) {
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "测试招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		nil, // 无 SeasonPassReward
		nil, // 无 SeasonPass
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.True(t, result.Ok, "无战令数据时没有需要保护的武将应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_ProtectionDeadlineExact EndTime 恰好等于保护期截止时间，应通过
func TestDrawFixProtection_ProtectionDeadlineExact(t *testing.T) {
	// SeasonPass StartTime=2025-01-01, protectMonths=4 → 保护期至 2025-05-01
	// DrawFix EndTime=2025-05-01 → 恰好等于截止时间，通过（>= 保护期截止算通过）
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "截止日招募", "2025-01-01 00:00:00", "2025-05-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.True(t, result.Ok, "EndTime 恰好等于保护期截止时间应通过")
	assert.Empty(t, result.ErrCells)
}

// TestDrawFixProtection_HeroNameFallback Hero 表不存在时，武将名称应显示 "未知(ID=xxx)"
func TestDrawFixProtection_HeroNameFallback(t *testing.T) {
	// SeasonPass StartTime=2025-01-01, protectMonths=4 → 保护期至 2025-05-01
	// DrawFix EndTime=2025-03-01 → 违规，Hero 表不存在应显示 "未知(ID=10001)"
	sheetMap := buildProtectionSheetMap(
		[][]string{{"1001", "未知武将招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010001;0}", "1"}},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		nil, // 无 Hero 表
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.False(t, result.Ok, "Hero 表不存在时仍应正确检测违规")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "未知(ID=10001)")
}

// TestDrawFixProtection_MultipleHeroes 一行包含多个武将道具，部分违规部分不违规
func TestDrawFixProtection_MultipleHeroes(t *testing.T) {
	// 武将 10001: SeasonPass StartTime=2025-01-01, protectMonths=4 → 保护期至 2025-05-01
	// 武将 10002: 不在战令中
	// DrawFix EndTime=2025-03-01 → 只有 10001 违规
	sheetMap := buildProtectionSheetMap(
		// DrawFix: ItemIds 包含两个武将道具
		[][]string{{"1001", "多武将招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010001;0}{1010002;0}", "2"}},
		// Item: 两个 Hero 道具
		[][]string{{"1010001", "Hero", "10001"}, {"1010002", "Hero", "10002"}},
		// SeasonPassReward: 只包含 10001
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "战令武将"}, {"10002", "普通武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.False(t, result.Ok, "包含违规武将应失败")
	assert.Len(t, result.ErrCells, 1, "只有战令武将应报错")
	assert.Contains(t, result.ErrCells[0].Reason, "战令武将")
	assert.NotContains(t, result.ErrCells[0].Reason, "普通武将")
}

// TestDrawFixProtection_MultipleDrawFixRows 多行 DrawFix，部分违规部分合规
func TestDrawFixProtection_MultipleDrawFixRows(t *testing.T) {
	sheetMap := buildProtectionSheetMap(
		// 两行 DrawFix: 第一行违规，第二行合规
		[][]string{
			{"1001", "违规招募", "2025-01-01 00:00:00", "2025-03-01 00:00:00", "{1010001;0}", "1"},
			{"1002", "安全招募", "2025-06-01 00:00:00", "2025-07-01 00:00:00", "{1010001;0}", "1"},
		},
		[][]string{{"1010001", "Hero", "10001"}},
		[][]string{{"1", "{1010001;1}"}},
		[][]string{{"1", "2025-01-01 00:00:00", "2025-02-01 00:00:00"}},
		[][]string{{"10001", "测试武将"}},
	)

	result := runProtectionCheck(sheetMap, nil)
	assert.False(t, result.Ok, "有违规行应失败")
	assert.Len(t, result.ErrCells, 1, "只有一行违规")
	assert.Contains(t, result.ErrCells[0].Reason, "违规招募")
}
