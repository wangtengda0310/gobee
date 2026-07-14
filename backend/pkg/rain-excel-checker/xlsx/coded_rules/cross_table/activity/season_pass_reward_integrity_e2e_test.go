//go:build e2e

// Package activity 提供活动相关的跨表校验规则 E2E 测试
//
// 本文件使用 e2e build tag，只有指定 -tags=e2e 时才运行。
// 补充单元测试未覆盖的端到端场景：多赛季混合时间门控、多道具链、
// 二次跳转断裂、warnDays=0 边界行为、空 HighReward 等。
package activity

import (
	"strings"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// createMultiSeasonSheetMap 构造多赛季时间门控混合场景的测试数据
//
// 表结构：
//   - SeasonPassReward: 2行数据（level=51 关联赛季5, level=61 关联赛季6）
//   - Item: 4行数据（两组完整跳转链）
//   - SeasonPass: 2行数据（赛季5 StartTime=now+3d, 赛季6 StartTime=now+30d）
//   - Hero: 2行数据（11903 赵奢, 11904 测试武将B）
//
// 参数 now 用于计算动态的 StartTime 偏移
func createMultiSeasonSheetMap(t *testing.T, now time.Time) (map[string]*excelize.File, [][]string) {
	t.Helper()

	// --- SeasonPassReward 表 ---
	rewardFile := excelize.NewFile()
	rewardFile.SetSheetName("Sheet1", spRewardIntegritySheetName)

	// 表头：行1空、行2注释、行3列名、行4空
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A3", "Id")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B3", "SeasonPassId")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C3", "level")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D3", "HighReward")

	// 行5: Id=1, SeasonPassId=5, level=51, HighReward={1002001;1}
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A5", "1")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B5", "5")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C5", "51")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D5", "{1002001;1}")

	// 行6: Id=2, SeasonPassId=6, level=61, HighReward={1002002;1}
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A6", "2")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B6", "6")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C6", "61")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D6", "{1002002;1}")

	rewardCols, _ := rewardFile.GetCols(spRewardIntegritySheetName)

	// --- Item 表 ---
	itemFile := excelize.NewFile()
	itemFile.SetSheetName("Sheet1", spRewardIntegrityItemSheetName)

	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A3", "Id")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B3", "ItemParam")

	// 第一组跳转链: 1002001 → 1003001 → 11903
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A5", "1002001")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B5", "1003001")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A6", "1003001")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B6", "11903")

	// 第二组跳转链: 1002002 → 1003002 → 11904
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A7", "1002002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B7", "1003002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A8", "1003002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B8", "11904")

	// --- SeasonPass 表 ---
	spFile := excelize.NewFile()
	spFile.SetSheetName("Sheet1", spRewardIntegritySeasonPassSheetName)

	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A3", "Id")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B3", "StartTime")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C3", "HeroId")

	// 赛季5: StartTime=now+3天（在 warnDays=7 窗口内）
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A5", "5")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B5", now.AddDate(0, 0, 3).Format("2006-01-02 15:04:05"))
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C5", "11904")

	// 赛季6: StartTime=now+30天（超过 warnDays=7 窗口）
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A6", "6")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B6", now.AddDate(0, 0, 30).Format("2006-01-02 15:04:05"))
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C6", "11904")

	// --- Hero 表 ---
	heroFile := excelize.NewFile()
	heroFile.SetSheetName("Sheet1", spRewardIntegrityHeroSheetName)

	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A3", "Id")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B3", "Name")

	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A5", "11903")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B5", "赵奢")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A6", "11904")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B6", "测试武将B")

	sheetMap := map[string]*excelize.File{
		spRewardIntegritySheetName:           rewardFile,
		spRewardIntegrityItemSheetName:       itemFile,
		spRewardIntegritySeasonPassSheetName: spFile,
		spRewardIntegrityHeroSheetName:       heroFile,
	}

	return sheetMap, rewardCols
}

// TestE2E_SPRewardIntegrity_MultiSeason_TimeGateMixed 多赛季时间门控混合场景
//
// 场景描述：
//   - SeasonPassReward 表有2行：赛季5(level=51)和赛季6(level=61)
//   - 赛季5 StartTime 在 warnDays=7 窗口内（距今3天）→ 应触发校验
//   - 赛季6 StartTime 超过 warnDays=7 窗口（距今30天）→ 应跳过校验
//   - 两个赛季的 HeroId 都不匹配左链（左链=11903/11904, 右链=11904 交集不为空）
//   - 赛季5的左链=11903, 右链=11904 → 无交集 → 报错
//   - 赛季6被时间门控跳过 → 不报错
//
// 预期：Ok=false，只有赛季5报错
func TestE2E_SPRewardIntegrity_MultiSeason_TimeGateMixed(t *testing.T) {
	fixedNow := parseFixedTime(t, "2026-06-01 00:00:00")

	// 赛季5 StartTime = now+3d = 2026-06-04, threshold = 2026-06-04 - 7d = 2026-05-28
	// now(2026-06-01) >= threshold(2026-05-28) → 触发校验
	// 赛季6 StartTime = now+30d = 2026-07-01, threshold = 2026-07-01 - 7d = 2026-06-24
	// now(2026-06-01) < threshold(2026-06-24) → 跳过校验
	sheetMap, rewardCols := createMultiSeasonSheetMap(t, fixedNow)

	rule := new(SeasonPassRewardIntegrityCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   spRewardIntegritySheetName,
		Cols:        rewardCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      map[string]string{string(json_rule.WARN_DAYS_BEFORE): "7"},
		SheetMap:    sheetMap,
		Now:         fixedNow,
	})

	assert.False(t, result.Ok, "多赛季混合场景应有错误")
	// 只有赛季5报错，赛季6被时间门控跳过
	assert.Equal(t, 1, len(result.ErrCells), "应有且仅有1个错误（赛季5）")
	if len(result.ErrCells) > 0 {
		assert.Contains(t, result.ErrCells[0].Reason, "赛季5", "错误应来自赛季5")
	}
}

// TestE2E_SPRewardIntegrity_MultipleItemsInHighReward 多道具链场景
//
// 场景描述：
//   - HighReward 包含多个道具（如 "{1002001;1}{1002002;1}"）
//   - 两个道具都有完整的跳转链：1002001→11903, 1002002→11904
//   - 左链提取出2个武将ID: {11903, 11904}
//   - 右链 HeroId 只包含其中一个: {11903}
//   - 左右链有交集（11903）→ 通过
//
// 预期：Ok=true（有交集则通过）
func TestE2E_SPRewardIntegrity_MultipleItemsInHighReward(t *testing.T) {
	// 构造 SeasonPassReward 表：HighReward 包含两个道具
	rewardFile := excelize.NewFile()
	rewardFile.SetSheetName("Sheet1", spRewardIntegritySheetName)

	rewardFile.SetCellValue(spRewardIntegritySheetName, "A3", "Id")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B3", "SeasonPassId")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C3", "level")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D3", "HighReward")

	// HighReward 包含两个道具: {1002001;1}{1002002;1}
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A5", "1")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B5", "5")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C5", "51")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D5", "{1002001;1}{1002002;1}")

	rewardCols, _ := rewardFile.GetCols(spRewardIntegritySheetName)

	// 复用基础数据，HeroId=11903（只有其中一个）
	sheetMap, _ := createSPRewardIntegritySheetMap("11903")
	// 替换 rewardFile 为多道具版本
	sheetMap[spRewardIntegritySheetName] = rewardFile

	fixedNow := parseFixedTime(t, "2026-05-25 00:00:00")

	rule := new(SeasonPassRewardIntegrityCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   spRewardIntegritySheetName,
		Cols:        rewardCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      map[string]string{string(json_rule.WARN_DAYS_BEFORE): "7"},
		SheetMap:    sheetMap,
		Now:         fixedNow,
	})

	assert.True(t, result.Ok, "多道具链有交集应通过, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "多道具链有交集不应有错误")
}

// TestE2E_SPRewardIntegrity_SecondJumpBroken 第二次 Item 表跳转断裂
//
// 场景描述：
//   - Item 表有 1002001→1003001（第一次跳转成功）
//   - 但缺少 1003001 的映射行（第二次跳转失败）
//   - 时间门控开启（StartTime 在 warnDays 窗口内）
//
// 预期：Ok=false，错误信息包含 "Item表跳转2"
func TestE2E_SPRewardIntegrity_SecondJumpBroken(t *testing.T) {
	// 构造 Item 表：只有第一次跳转，缺少 1003001 的映射
	itemFile := excelize.NewFile()
	itemFile.SetSheetName("Sheet1", spRewardIntegrityItemSheetName)

	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A3", "Id")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B3", "ItemParam")

	// 只有 1002001 → 1003001（第一次跳转成功）
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A5", "1002001")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B5", "1003001")
	// 缺少 1003001 → ? 的行（第二次跳转断裂）

	// 复用基础数据
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11903")
	// 替换 Item 表为不完整的版本
	sheetMap[spRewardIntegrityItemSheetName] = itemFile

	fixedNow := parseFixedTime(t, "2026-05-25 00:00:00")

	rule := new(SeasonPassRewardIntegrityCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   spRewardIntegritySheetName,
		Cols:        rewardCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      map[string]string{string(json_rule.WARN_DAYS_BEFORE): "7"},
		SheetMap:    sheetMap,
		Now:         fixedNow,
	})

	assert.False(t, result.Ok, "第二次跳转断裂应报错")
	assert.NotEmpty(t, result.ErrCells, "应有错误信息")

	// 验证错误信息包含 "Item表跳转2"
	found := false
	for _, errCell := range result.ErrCells {
		if strings.Contains(errCell.Reason, "Item表跳转2") {
			found = true
			break
		}
	}
	assert.True(t, found, "错误信息应包含 'Item表跳转2'")
}

// TestE2E_SPRewardIntegrity_WarnDaysZero warnDays=0 时使用默认值7天
//
// 场景描述：
//   - SeasonPass.StartTime 距今很远（now 之前365天 → 赛季已开始很久）
//   - warnDays="0"，解析后 days=0 不满足 >0 条件，使用默认值7
//   - StartTime 距今365天，远超7天窗口
//   - 左右链不匹配
//
// 预期：Ok=true（时间门控关闭，跳过校验）
// 验证行为：warnDays="0" 时 days=0，不满足 days>0 条件，不会覆盖默认值7
func TestE2E_SPRewardIntegrity_WarnDaysZero(t *testing.T) {
	// 构造 SeasonPass 表：StartTime=now+365天（距今非常远）
	spFile := excelize.NewFile()
	spFile.SetSheetName("Sheet1", spRewardIntegritySeasonPassSheetName)

	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A3", "Id")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B3", "StartTime")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C3", "HeroId")

	// StartTime 设为 now+365天
	fixedNow := parseFixedTime(t, "2026-06-01 00:00:00")
	startTime365dLater := fixedNow.AddDate(0, 0, 365).Format("2006-01-02 15:04:05")

	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A5", "5")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B5", startTime365dLater)
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C5", "11904") // 与左链不匹配

	// 复用基础数据（HeroId=11904 使右链与左链11903不匹配）
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11904")
	// 替换 SeasonPass 表
	sheetMap[spRewardIntegritySeasonPassSheetName] = spFile

	// warnDays="0" → days=0, 不满足 days>0, 使用默认值7
	// threshold = startTime - 7d
	// now(2026-06-01) < threshold → 跳过校验
	rule := new(SeasonPassRewardIntegrityCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   spRewardIntegritySheetName,
		Cols:        rewardCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      map[string]string{string(json_rule.WARN_DAYS_BEFORE): "0"},
		SheetMap:    sheetMap,
		Now:         fixedNow,
	})

	assert.True(t, result.Ok, "warnDays=0 使用默认值7天后，StartTime距今365天超出窗口应跳过校验, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "时间门控关闭时不应有错误")
}

// TestE2E_SPRewardIntegrity_EmptyHighReward HighReward 为空字符串
//
// 场景描述：
//   - 目标 level 行存在但 HighReward=""
//   - 时间门控开启（StartTime 在 warnDays 窗口内）
//   - 空字符串在 Check 方法中被 continue 跳过
//
// 预期：Ok=true（空字符串被跳过）
func TestE2E_SPRewardIntegrity_EmptyHighReward(t *testing.T) {
	// 构造 SeasonPassReward 表：level=51 但 HighReward 为空
	rewardFile := excelize.NewFile()
	rewardFile.SetSheetName("Sheet1", spRewardIntegritySheetName)

	rewardFile.SetCellValue(spRewardIntegritySheetName, "A3", "Id")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B3", "SeasonPassId")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C3", "level")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D3", "HighReward")

	// level=51 是目标行，但 HighReward 为空
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A5", "1")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B5", "5")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C5", "51")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D5", "")

	rewardCols, _ := rewardFile.GetCols(spRewardIntegritySheetName)

	// 复用基础数据
	sheetMap, _ := createSPRewardIntegritySheetMap("11903")
	// 替换 rewardFile
	sheetMap[spRewardIntegritySheetName] = rewardFile

	fixedNow := parseFixedTime(t, "2026-05-25 00:00:00")

	rule := new(SeasonPassRewardIntegrityCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   spRewardIntegritySheetName,
		Cols:        rewardCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      map[string]string{string(json_rule.WARN_DAYS_BEFORE): "7"},
		SheetMap:    sheetMap,
		Now:         fixedNow,
	})

	assert.True(t, result.Ok, "HighReward 为空应跳过该行而通过, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "HighReward 为空不应有错误")
}
