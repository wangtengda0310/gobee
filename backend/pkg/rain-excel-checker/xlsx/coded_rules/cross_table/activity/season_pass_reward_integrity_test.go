// Package activity 提供活动相关的跨表校验规则
//
// 本文件测试 SEASON_PASS_REWARD_INTEGRITY_CHECK 规则：
// 检查战令高级奖励(level=51/61/71/81/91)的道具经两次Item跳转后的武将ID，
// 是否在SeasonPass.HeroId对应的武将集合中。
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

// 固定测试时间常量
// SeasonPass.StartTime = 2026-06-01，warnDays=7，时间门控阈值 = 2026-05-25 00:00:00
const testStartTime = "2026-06-01 00:00:00"

// spRewardIntegritySheetName 测试中 SeasonPassReward 表的 sheet 键名
const spRewardIntegritySheetName = "赛季战令奖励表|SeasonPassReward"

// spRewardIntegrityItemSheetName 测试中 Item 表的 sheet 键名
const spRewardIntegrityItemSheetName = "道具表|Item"

// spRewardIntegritySeasonPassSheetName 测试中 SeasonPass 表的 sheet 键名
const spRewardIntegritySeasonPassSheetName = "赛季战令表|SeasonPass"

// spRewardIntegrityHeroSheetName 测试中 Hero 表的 sheet 键名
const spRewardIntegrityHeroSheetName = "武将|Hero"

// createSPRewardIntegritySheetMap 创建完整的4表测试数据
//
// 表结构：
//   - SeasonPassReward: Id, SeasonPassId, level, HighReward
//   - Item: Id, ItemParam (两行：信物→包装道具ID，包装道具→武将ID)
//   - SeasonPass: Id, StartTime, HeroId (HeroId 由参数指定)
//   - Hero: Id, Name
func createSPRewardIntegritySheetMap(heroId string) (map[string]*excelize.File, [][]string) {
	// --- SeasonPassReward 表 ---
	rewardFile := excelize.NewFile()
	rewardFile.SetSheetName("Sheet1", spRewardIntegritySheetName)

	// 表头：行1空、行2类型注释、行3列名、行4空
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A3", "Id")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A4", "")

	rewardFile.SetCellValue(spRewardIntegritySheetName, "B1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B3", "SeasonPassId")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B4", "")

	rewardFile.SetCellValue(spRewardIntegritySheetName, "C1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C3", "level")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C4", "")

	rewardFile.SetCellValue(spRewardIntegritySheetName, "D1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D3", "HighReward")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D4", "")

	// 数据行：SeasonPassId=5, level=51, HighReward="{1002001;1}"
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A5", "1")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B5", "5")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C5", "51")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D5", "{1002001;1}")

	rewardCols, _ := rewardFile.GetCols(spRewardIntegritySheetName)

	// --- Item 表 ---
	itemFile := excelize.NewFile()
	itemFile.SetSheetName("Sheet1", spRewardIntegrityItemSheetName)

	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A1", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A2", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A3", "Id")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A4", "")

	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B1", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B2", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B3", "ItemParam")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B4", "")

	// 信物道具 → 包装道具ID
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A5", "1002001")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B5", "1003001")
	// 包装道具 → 武将ID
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A6", "1003001")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B6", "11903")
	// 另一组道具链
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A7", "1002002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B7", "1003002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A8", "1003002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B8", "11904")

	// --- SeasonPass 表 ---
	spFile := excelize.NewFile()
	spFile.SetSheetName("Sheet1", spRewardIntegritySeasonPassSheetName)

	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A1", "")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A2", "")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A3", "Id")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A4", "")

	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B1", "")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B2", "")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B3", "StartTime")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B4", "")

	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C1", "")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C2", "")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C3", "HeroId")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C4", "")

	// 赛季5：StartTime=2026-06-01, HeroId 由参数指定
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "A5", "5")
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "B5", testStartTime)
	spFile.SetCellValue(spRewardIntegritySeasonPassSheetName, "C5", heroId)

	// --- Hero 表 ---
	heroFile := excelize.NewFile()
	heroFile.SetSheetName("Sheet1", spRewardIntegrityHeroSheetName)

	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A1", "")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A2", "")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A3", "Id")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "A4", "")

	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B1", "")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B2", "")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B3", "Name")
	heroFile.SetCellValue(spRewardIntegrityHeroSheetName, "B4", "")

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

// parseFixedTime 解析固定时间字符串（UTC），与 helpers.ParseDate 的 time.Parse 策略一致
func parseFixedTime(t *testing.T, timeStr string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		t.Fatalf("解析时间失败 %q: %v", timeStr, err)
	}
	return parsed
}

// TestSPRewardIntegrity_Match 测试左右链武将ID匹配
// SeasonPass.HeroId="11903"，左链提取武将ID=11903，now >= startTime-7d
// 预期：Ok=true
func TestSPRewardIntegrity_Match(t *testing.T) {
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11903")
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

	assert.True(t, result.Ok, "左右链武将ID匹配应通过, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "不应有错误单元格")
}

// TestSPRewardIntegrity_NoMatch_TimeGateOpen 测试左右链不匹配且时间门控开启
// SeasonPass.HeroId="11904"，左链提取武将ID=11903，now >= startTime-7d
// 预期：Ok=false，ErrCells 包含链路信息
func TestSPRewardIntegrity_NoMatch_TimeGateOpen(t *testing.T) {
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11904")
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

	assert.False(t, result.Ok, "左右链不匹配且时间门控开启应报错")
	assert.NotEmpty(t, result.ErrCells, "应有错误信息")
}

// TestSPRewardIntegrity_NoMatch_TimeGateClosed 测试左右链不匹配但时间门控关闭
// SeasonPass.HeroId="11904"，左链提取武将ID=11903，now < startTime-7d
// 预期：Ok=true（时间门控关闭，跳过校验）
func TestSPRewardIntegrity_NoMatch_TimeGateClosed(t *testing.T) {
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11904")
	fixedNow := parseFixedTime(t, "2026-05-24 00:00:00")

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

	assert.True(t, result.Ok, "时间门控关闭时应跳过校验而通过, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "时间门控关闭时不应有错误")
}

// TestSPRewardIntegrity_TimeGateEdge 测试时间门控边界值
// SeasonPass.HeroId="11904"，左链提取武将ID=11903
// now = startTime - 7d 恰好等于阈值（now 不 Before threshold，即 >= 触发校验）
// 预期：Ok=false
func TestSPRewardIntegrity_TimeGateEdge(t *testing.T) {
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11904")
	// startTime = 2026-06-01, warnDays=7, threshold = 2026-05-25 00:00:00
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

	assert.False(t, result.Ok, "now 等于 startTime-7d 应触发校验并报错（不匹配）")
	assert.NotEmpty(t, result.ErrCells, "应有错误信息")
}

// TestSPRewardIntegrity_LeftChainBroken 测试左链断裂
// Item 表中没有 Id=1002001 的行（信物道具不存在），时间门控开启
// 预期：Ok=false，错误信息包含 "道具ID=1002001" 和 "Item表跳转1"
func TestSPRewardIntegrity_LeftChainBroken(t *testing.T) {
	// 创建没有 1002001 的 Item 表
	itemFile := excelize.NewFile()
	itemFile.SetSheetName("Sheet1", spRewardIntegrityItemSheetName)

	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A1", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A2", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A3", "Id")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A4", "")

	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B1", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B2", "")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B3", "ItemParam")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B4", "")

	// 只放 1002002 → 1003002 → 11904 的数据，不放 1002001
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A5", "1002002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B5", "1003002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "A6", "1003002")
	itemFile.SetCellValue(spRewardIntegrityItemSheetName, "B6", "11904")

	// 用默认数据构建其他表，HeroId=11903 让右链有值
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11903")
	// 替换 Item 表为缺少 1002001 的版本
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

	assert.False(t, result.Ok, "左链断裂应报错")
	assert.NotEmpty(t, result.ErrCells, "应有错误信息")

	// 验证错误信息包含关键词
	found := false
	for _, errCell := range result.ErrCells {
		if strings.Contains(errCell.Reason, "道具ID=1002001") && strings.Contains(errCell.Reason, "Item表跳转1") {
			found = true
			break
		}
	}
	assert.True(t, found, "错误信息应包含 '道具ID=1002001' 和 'Item表跳转1'")
}

// TestSPRewardIntegrity_MissingSheets 测试缺少关联表
// sheetMap 中没有 Item 表
// 预期：Ok=false，Reason 包含 "Item"
func TestSPRewardIntegrity_MissingSheets(t *testing.T) {
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("11903")
	// 删除 Item 表
	delete(sheetMap, spRewardIntegrityItemSheetName)

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

	assert.False(t, result.Ok, "缺少 Item 表应报错")
	assert.True(t, strings.Contains(result.Reason, "Item"), "Reason 应包含 'Item', 实际: %s", result.Reason)
}

// TestSPRewardIntegrity_NoTargetLevels 测试没有目标 level 行
// SeasonPassReward 表中没有 level=51/61/71/81/91 的行（只有 level=1）
// 预期：Ok=true（所有行都被过滤，无检查目标）
func TestSPRewardIntegrity_NoTargetLevels(t *testing.T) {
	// 创建只有 level=1 的 SeasonPassReward 表
	rewardFile := excelize.NewFile()
	rewardFile.SetSheetName("Sheet1", spRewardIntegritySheetName)

	rewardFile.SetCellValue(spRewardIntegritySheetName, "A1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A3", "Id")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A4", "")

	rewardFile.SetCellValue(spRewardIntegritySheetName, "B1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B3", "SeasonPassId")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B4", "")

	rewardFile.SetCellValue(spRewardIntegritySheetName, "C1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C3", "level")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C4", "")

	rewardFile.SetCellValue(spRewardIntegritySheetName, "D1", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D2", "")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D3", "HighReward")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D4", "")

	// 只有 level=1 的行，不在目标集合 {51,61,71,81,91} 中
	rewardFile.SetCellValue(spRewardIntegritySheetName, "A5", "1")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "B5", "5")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "C5", "1")
	rewardFile.SetCellValue(spRewardIntegritySheetName, "D5", "{1002001;1}")

	rewardCols, _ := rewardFile.GetCols(spRewardIntegritySheetName)

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

	assert.True(t, result.Ok, "没有目标 level 行应通过, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "没有目标 level 行不应有错误")
}

// TestSPRewardIntegrity_EmptyHeroId 测试 SeasonPass.HeroId 为空
// SeasonPass 表的 HeroId 为空字符串，时间门控开启
// 预期：Ok=true（右链为空，跳过比较）
func TestSPRewardIntegrity_EmptyHeroId(t *testing.T) {
	sheetMap, rewardCols := createSPRewardIntegritySheetMap("")
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

	assert.True(t, result.Ok, "HeroId 为空时应跳过比较而通过, Reason: %s", result.Reason)
	assert.Empty(t, result.ErrCells, "HeroId 为空时不应有错误")
}
