//go:build e2e

// Package coded_rules 提供跨表级别的校验规则 E2E 测试
// 本文件使用 e2e build tag，只有指定 -tags=e2e 时才运行

package draw

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestE2E_DrawFixArenaProtection_ReverseViolation 反向违规场景
//
// 场景描述：
// - ArenaSeason 赛季进行中（SeasonEndTime=2025-12-31）
// - ArenaScoreReward 大将军段位奖励包含卫青
// - DrawFix 在赛季期间开放，包含卫青
//
// 预期结果：Ok=false，错误信息包含"卫青"和"赛季"
func TestE2E_DrawFixArenaProtection_ReverseViolation(t *testing.T) {
	// 构造测试数据
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: EndTime=2025-06-30(在赛季期间), ItemIds 包含卫青
		[][]string{
			{"1001", "违规招募", "2025-01-01 00:00:00", "2025-06-30 00:00:00", "{1010804;0}", "1"},
		},
		// Item: 1010804 是 Hero 道具，对应 HeroId=10804（卫青）
		[][]string{
			{"1010804", "Hero", "10804"},
		},
		// ArenaScoreReward: Season=1, Dan=23, DanName="大将军", Reward 包含卫青
		[][]string{
			{"1", "23", "大将军", "{1010804;1}"},
		},
		// ArenaSeason: 赛季进行中，SeasonEndTime=2025-12-31
		[][]string{
			{"1", "2025-01-01 00:00:00", "2025-12-31 00:00:00"},
		},
		// Hero: 卫青
		[][]string{
			{"10804", "卫青"},
		},
	)

	// 注入 now=2025-06-15（赛季进行中）
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	result := runArenaProtectionCheck(sheetMap, now)

	// 验证结果
	assert.False(t, result.Ok, "赛季期间包含大将军武将应失败")
	assert.Len(t, result.ErrCells, 1, "应有一个错误单元格")
	assert.Contains(t, result.ErrCells[0].Reason, "卫青", "错误信息应包含武将名称")
	assert.Contains(t, result.ErrCells[0].Reason, "赛季", "错误信息应包含赛季相关信息")
}

// TestE2E_DrawFixArenaProtection_PostProtectionNoConfig 保护期后未配置大将军武将
//
// 场景描述：
// - ArenaSeason 赛季已结束（SeasonEndTime=2025-03-14）
// - ArenaScoreReward 大将军段位奖励包含王翦
// - DrawFix 已开放，但不包含王翦
//
// 预期结果：Ok=true，保护期后不强制要求配置大将军武将（只做反向检查）
func TestE2E_DrawFixArenaProtection_PostProtectionNoConfig(t *testing.T) {
	// 构造测试数据
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: ItemIds 不包含王翦（包含卫青）
		[][]string{
			{"1001", "其他招募", "2025-06-01 00:00:00", "2025-07-01 00:00:00", "{1010804;0}", "1"},
		},
		// Item: 两个 Hero 道具
		[][]string{
			{"1010804", "Hero", "10804"}, // 卫青
			{"1011601", "Hero", "11601"}, // 王翦
		},
		// ArenaScoreReward: Season=2, Dan=23, DanName="大将军", Reward 包含王翦
		[][]string{
			{"2", "23", "大将军", "{1011601;1}"},
		},
		// ArenaSeason: 赛季已结束，SeasonEndTime=2025-03-14
		[][]string{
			{"2", "2025-01-01 00:00:00", "2025-03-14 00:00:00"},
		},
		// Hero: 王翦和卫青
		[][]string{
			{"11601", "王翦"},
			{"10804", "卫青"},
		},
	)

	// 注入 now=2025-06-15（赛季已结束）
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	result := runArenaProtectionCheck(sheetMap, now)

	// 验证结果：保护期后不强制要求配置，应通过
	assert.True(t, result.Ok, "保护期后不强制要求大将军武将出现在定向招募中")
	assert.Empty(t, result.ErrCells, "应无错误单元格")
}

// TestE2E_DrawFixArenaProtection_NoViolation 无违规场景
//
// 场景描述：
// - ArenaSeason 赛季已结束（SeasonEndTime=2025-04-14）
// - ArenaScoreReward 大将军段位奖励包含李信
// - DrawFix 在赛季结束后才开放，包含李信
//
// 预期结果：Ok=true，无错误单元格
func TestE2E_DrawFixArenaProtection_NoViolation(t *testing.T) {
	// 构造测试数据
	sheetMap := buildArenaProtectionSheetMap(
		// DrawFix: EndTime=2025-06-01(赛季已结束), ItemIds 包含李信
		[][]string{
			{"1001", "安全招募", "2025-05-01 00:00:00", "2025-06-01 00:00:00", "{1010613;0}", "1"},
		},
		// Item: 1010613 是 Hero 道具，对应 HeroId=10613（李信）
		[][]string{
			{"1010613", "Hero", "10613"},
		},
		// ArenaScoreReward: Season=3, Dan=23, DanName="大将军", Reward 包含李信
		[][]string{
			{"3", "23", "大将军", "{1010613;1}"},
		},
		// ArenaSeason: 赛季已结束，SeasonEndTime=2025-04-14
		[][]string{
			{"3", "2025-01-01 00:00:00", "2025-04-14 00:00:00"},
		},
		// Hero: 李信
		[][]string{
			{"10613", "李信"},
		},
	)

	// 注入 now=2025-06-15（赛季已结束，但武将已在 DrawFix 开放）
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	result := runArenaProtectionCheck(sheetMap, now)

	// 验证结果
	assert.True(t, result.Ok, "赛季已结束且武已在定向招募开放应通过")
	assert.Empty(t, result.ErrCells, "应无错误单元格")
}
