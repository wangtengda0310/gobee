// Package helpers 提供校验规则的内部辅助工具
// 本文件包含 hero_rule_helper.go 中 CalcArenaProtectionDeadline 的单元测试
package helpers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== CalcArenaProtectionDeadline 测试 ====================

// TestCalcArenaProtectionDeadline_Basic
// 正常赛季结束时间，返回值等于输入
func TestCalcArenaProtectionDeadline_Basic(t *testing.T) {
	seasonEndTime := time.Date(2026, 5, 14, 23, 59, 59, 0, time.Local)

	result := CalcArenaProtectionDeadline(seasonEndTime)

	// 新逻辑：保护期截止时间 = 赛季结束时间
	assert.Equal(t, seasonEndTime, result, "保护期截止时间应等于赛季结束时间")
}

// TestCalcArenaProtectionDeadline_ZeroTime
// 零值时间输入，返回零值
func TestCalcArenaProtectionDeadline_ZeroTime(t *testing.T) {
	result := CalcArenaProtectionDeadline(time.Time{})

	// 零值时间传入，返回的也是零值（seasonEndTime 本身就是零值）
	assert.True(t, result.IsZero(), "零值时间输入应返回零值")
}

// TestCalcArenaProtectionDeadline_RealWorldScenario
// 使用真实场景数据验证保护期截止时间计算
//
// 场景：
//   - 赛季开始 2026-04-15，赛季结束 2026-05-14
//   - 保护期截止时间应为 2026-05-14（即赛季结束时间）
//   - 而非旧逻辑的 2026-08-15（StartTime + 4个月）
func TestCalcArenaProtectionDeadline_RealWorldScenario(t *testing.T) {
	seasonStartTime := time.Date(2026, 4, 15, 0, 0, 0, 0, time.Local)
	seasonEndTime := time.Date(2026, 5, 14, 23, 59, 59, 0, time.Local)

	result := CalcArenaProtectionDeadline(seasonEndTime)

	// 旧逻辑会计算 SeasonStartTime + 4个月 = 2026-08-15
	// 新逻辑直接返回 SeasonEndTime = 2026-05-14
	assert.Equal(t, seasonEndTime, result,
		"保护期截止时间应等于赛季结束时间 2026-05-14，而非旧逻辑的 SeasonStartTime+4月")

	// 确保不是旧逻辑的结果
	oldLogicResult := seasonStartTime.AddDate(0, 4, 0)
	assert.NotEqual(t, oldLogicResult, result,
		"保护期截止时间不应等于旧逻辑的 SeasonStartTime+4月(2026-08-15)")
}

// TestCalcArenaProtectionDeadline_DifferentSeasonTimes
// 验证不同赛季时间都能正确返回
func TestCalcArenaProtectionDeadline_DifferentSeasonTimes(t *testing.T) {
	tests := []struct {
		name    string
		endTime time.Time
	}{
		{
			name:    "三月赛季",
			endTime: time.Date(2026, 3, 14, 23, 59, 59, 0, time.Local),
		},
		{
			name:    "六月赛季",
			endTime: time.Date(2026, 6, 15, 23, 59, 59, 0, time.Local),
		},
		{
			name:    "十二月赛季",
			endTime: time.Date(2026, 12, 31, 23, 59, 59, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcArenaProtectionDeadline(tt.endTime)
			assert.Equal(t, tt.endTime, result, "保护期截止时间应等于赛季结束时间")
		})
	}
}

// ==================== ParseDate 测试 ====================

// TestParseDate_ZeroPadVariants 验证 ParseDate 兼容补零和不补零两种日期格式
// 回归保护：策划在 Excel 中可能写 "2026-6-1" 或 "2026-06-1" 等不完整补零的日期，
// 修复前 DateFormats 仅支持 "2006-01-02"，单位月/日的字符串解析为零值时间，
// 导致 withinDays 过滤等基于时间的逻辑静默跳过这些行
func TestParseDate_ZeroPadVariants(t *testing.T) {
	want := time.Date(2026, 6, 1, 5, 0, 0, 0, time.UTC)
	wantDateOnly := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		input string
		want  time.Time
	}{
		{"2026-06-01 05:00:00", want},
		{"2026-06-1 05:00:00", want}, // 日不补零
		{"2026-6-1 05:00:00", want},  // 月日都不补零
		{"2026-6-01 05:00:00", want}, // 月不补零、日补零
		{"2026/06/01 05:00:00", want},
		{"2026/6/1 05:00:00", want},
		{"2026-06-01", wantDateOnly},
		{"2026-6-1", wantDateOnly},
		{"2026/6/1", wantDateOnly},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got := ParseDate(c.input)
			assert.Equal(t, c.want, got,
				"ParseDate(%q) 应解析成功（修复前对非补零格式返回零值时间）", c.input)
		})
	}
}

// TestParseDate_Invalid 验证非法格式返回零值
func TestParseDate_Invalid(t *testing.T) {
	cases := []string{
		"",
		"not-a-date",
		"2026/13/40", // 非法月日
		"2026年6月1日",  // 中文日期，未支持
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			got := ParseDate(input)
			assert.True(t, got.IsZero(), "ParseDate(%q) 应返回零值时间", input)
		})
	}
}
