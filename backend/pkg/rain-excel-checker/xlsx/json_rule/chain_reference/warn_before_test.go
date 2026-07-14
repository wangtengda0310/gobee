package chain_reference

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestShouldSuppressByWarnBefore_未配置warnBefore 不静默
func TestShouldSuppressByWarnBefore_NoWarnBefore(t *testing.T) {
	ctx := &ChainContext{
		Params: map[string]string{},
	}
	result := ShouldSuppressByWarnBefore(ctx, time.Now())
	assert.False(t, result, "未配置 warnBefore 时不应该静默")
}

// TestShouldSuppressByWarnBefore_warnBefore为零 不静默
func TestShouldSuppressByWarnBefore_ZeroWarnBefore(t *testing.T) {
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "0h",
		},
	}
	result := ShouldSuppressByWarnBefore(ctx, time.Now())
	assert.False(t, result, "warnBefore 为 0 时不应该静默")
}

// TestShouldSuppressByWarnBefore_WarnValues为空 不静默
func TestShouldSuppressByWarnBefore_EmptyWarnValues(t *testing.T) {
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
		WarnValues: nil,
	}
	result := ShouldSuppressByWarnBefore(ctx, time.Now())
	assert.False(t, result, "WarnValues 为空时不应该静默")
}

// TestShouldSuppressByWarnBefore_未来时间超过warnBefore 静默
func TestShouldSuppressByWarnBefore_FutureTimeExceedsWarnBefore(t *testing.T) {
	now := time.Now()
	// 预警时间距今 200 小时，warnBefore=168h，超过 → 静默
	futureTime := now.Add(200 * time.Hour).Format("2006-01-02 15:04:05")

	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
		WarnValues: []string{futureTime},
	}
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.True(t, result, "未来时间超过 warnBefore 应该静默")
}

// TestShouldSuppressByWarnBefore_未来时间在warnBefore内 不静默
func TestShouldSuppressByWarnBefore_FutureTimeWithinWarnBefore(t *testing.T) {
	now := time.Now()
	// 预警时间距今 100 小时，warnBefore=168h，未超过 → 不静默
	futureTime := now.Add(100 * time.Hour).Format("2006-01-02 15:04:05")

	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
		WarnValues: []string{futureTime},
	}
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.False(t, result, "未来时间在 warnBefore 内不应该静默")
}

// TestShouldSuppressByWarnBefore_所有时间都是过去 不静默
func TestShouldSuppressByWarnBefore_AllPastTimes(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")

	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
		WarnValues: []string{pastTime},
	}
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.False(t, result, "所有时间都是过去时不应该静默")
}

// TestShouldSuppressByWarnBefore_多个未来时间取最近的
func TestShouldSuppressByWarnBefore_MultipleFutureTimes(t *testing.T) {
	now := time.Now()
	// 最近未来时间距今 100h，最远 300h，warnBefore=168h
	nearFuture := now.Add(100 * time.Hour).Format("2006-01-02 15:04:05")
	farFuture := now.Add(300 * time.Hour).Format("2006-01-02 15:04:05")

	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
		WarnValues: []string{farFuture, nearFuture},
	}
	// 最近未来时间 100h < warnBefore 168h → 不静默
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.False(t, result, "最近未来时间在 warnBefore 内不应该静默")
}

// TestShouldSuppressByWarnBefore_多个未来时间全部超过warnBefore 静默
func TestShouldSuppressByWarnBefore_AllFutureTimesExceedWarnBefore(t *testing.T) {
	now := time.Now()
	nearFuture := now.Add(200 * time.Hour).Format("2006-01-02 15:04:05")
	farFuture := now.Add(300 * time.Hour).Format("2006-01-02 15:04:05")

	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
		WarnValues: []string{farFuture, nearFuture},
	}
	// 最近未来时间 200h > warnBefore 168h → 静默
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.True(t, result, "所有未来时间超过 warnBefore 应该静默")
}

// TestWarnBefore_解析正常值
func TestWarnBefore_ValidDuration(t *testing.T) {
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
		},
	}
	assert.Equal(t, 168*time.Hour, ctx.WarnBefore())
}

// TestWarnBefore_无效格式返回零
func TestWarnBefore_InvalidFormat(t *testing.T) {
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "invalid",
		},
	}
	assert.Equal(t, time.Duration(0), ctx.WarnBefore())
}

// TestWarnSheet_And_WarnCol
func TestWarnSheet_WarnCol(t *testing.T) {
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnSheet": "赛季战令表|SeasonPass",
			"chainWarnCol":   "StartTime",
		},
	}
	assert.Equal(t, "赛季战令表|SeasonPass", ctx.WarnSheet())
	assert.Equal(t, "StartTime", ctx.WarnCol())
}

// TestSheetNameMatch 各种匹配场景
func TestSheetNameMatch(t *testing.T) {
	assert.True(t, sheetNameMatch("SeasonPass", "SeasonPass"), "精确匹配")
	assert.True(t, sheetNameMatch("赛季战令表|SeasonPass", "SeasonPass"), "后缀匹配")
	assert.True(t, sheetNameMatch("SeasonPass", "赛季战令表|SeasonPass"), "反向后缀匹配")
	assert.False(t, sheetNameMatch("SeasonPass", "OtherTable"), "不匹配")
	assert.False(t, sheetNameMatch("SeasonPassX", "SeasonPass"), "前缀不匹配")
}

// TestShouldSuppressByCurrentValue_第五期在窗口内不静默_第六期超窗口静默
// 模拟场景：SeasonPassReward 中第5期和第6期奖励配了不存在的道具
// 第5赛季 StartTime 距今3天（7天窗口内）→ 不静默（报错）
// 第6赛季 StartTime 距今30天（超过7天窗口）→ 静默（不报错）
func TestShouldSuppressByCurrentValue_Season5WarnSeason6Suppress(t *testing.T) {
	now := time.Now()
	season5Start := now.Add(3 * 24 * time.Hour).Format("2006-01-02 15:04:05")  // 3天后
	season6Start := now.Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05") // 30天后

	// 构造 SeasonPass 表：Id + StartTime 两列
	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass",
		[]string{"Id", "StartTime"},
		[][]string{
			{"5", season5Start},
			{"6", season6Start},
		},
	)
	sheetMap := map[string]*excelize.File{
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	// 第5赛季（SeasonPassId="5"）→ StartTime 距今3天 < 7天 → 不静默
	ctx5 := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
			"chainWarnSheet":  "赛季战令表|SeasonPass",
			"chainWarnCol":    "StartTime",
		},
		SheetMap:   sheetMap,
		WarnValues: nil, // 链路断开，WarnValues 为空
		MyColData:  []string{"5", "5"},
		DataIdx:    0,
	}
	result5 := ShouldSuppressByWarnBefore(ctx5, now)
	assert.False(t, result5, "第5赛季 StartTime 在7天窗口内，不应该静默")

	// 第6赛季（SeasonPassId="6"）→ StartTime 距今30天 > 7天 → 静默
	ctx6 := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
			"chainWarnSheet":  "赛季战令表|SeasonPass",
			"chainWarnCol":    "StartTime",
		},
		SheetMap:   sheetMap,
		WarnValues: nil,
		MyColData:  []string{"6", "6"},
		DataIdx:    0,
	}
	result6 := ShouldSuppressByWarnBefore(ctx6, now)
	assert.True(t, result6, "第6赛季 StartTime 超过7天窗口，应该静默")
}

// TestShouldSuppressByCurrentValue_过去时间不静默
// 模拟场景：赛季已开始（StartTime 在过去），应该报错
func TestShouldSuppressByCurrentValue_PastTimeNotSuppress(t *testing.T) {
	now := time.Now()
	pastStart := now.Add(-10 * 24 * time.Hour).Format("2006-01-02 15:04:05") // 10天前

	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass",
		[]string{"Id", "StartTime"},
		[][]string{
			{"5", pastStart},
		},
	)
	sheetMap := map[string]*excelize.File{
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
			"chainWarnSheet":  "赛季战令表|SeasonPass",
			"chainWarnCol":    "StartTime",
		},
		SheetMap:   sheetMap,
		WarnValues: nil,
		MyColData:  []string{"5"},
		DataIdx:    0,
	}
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.False(t, result, "过去时间不应该静默")
}
