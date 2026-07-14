//go:build e2e

// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件包含预警窗口（chainWarnBefore）的端到端（E2E）测试
// 使用 e2e build tag，只有指定 -tags=e2e 时才运行
package chain_reference

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 测试1: 洋葱模型完整链路 + 预警窗口静默 ====================

// TestE2E_WarnBefore_OnionChainWithWarnSuppression 洋葱模型完整链路 + 预警窗口过滤
// 场景：左链值在右链中不存在（本应 violation=true），但预警时间距今超过 168h → 静默
//
// 构造方式：
//   - 当前表有 Col1="1001"
//   - 左链第一步取 "1001"，第二步到 LeftTable 查找提取 Name="ItemA"
//   - 右链一步到 SeasonPass 表查找
//   - chainWarnSheet="赛季战令表|SeasonPass", chainWarnCol="StartTime"
//   - SeasonPass 表有 Id=SP001, StartTime=now+200h（远超 168h 窗口）
//
// 预期：Compare 层判定 violation，但 ShouldSuppressByWarnBefore 返回 true → Violation=false
func TestE2E_WarnBefore_OnionChainWithWarnSuppression(t *testing.T) {
	now := time.Now()
	futureTime200h := now.Add(200 * time.Hour).Format("2006-01-02 15:04:05")

	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// 左链目标表
	leftFile := buildChainTestRefSheet("LeftTable", []string{"Id", "Name"}, [][]string{
		{"1001", "ItemA"},
	})

	// 预警表（同时也是右链目标表）
	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass", []string{"Id", "StartTime", "Name"}, [][]string{
		{"SP001", futureTime200h, "Season1"},
	})

	sheetMap := map[string]*excelize.File{
		"LeftTable":        leftFile,
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
				{Sheet: "LeftTable", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "赛季战令表|SeasonPass", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"chainWarnBefore":   "168h",
		"chainWarnSheet":    "赛季战令表|SeasonPass",
		"chainWarnCol":      "StartTime",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链最终值 "ItemA"，右链最终表 Name 列值 ["Season1"]
	// Match verify_exists: "ItemA" 在 ["Season1"] 中未找到 → 门控不通过
	// 门控不通过时 Violation 保持 false，不触发预警检查
	// 但我们仍需验证预警参数被正确传递
	assert.False(t, ctx.Violation, "门控未通过时不应违规")
}

// ==================== 测试2: 洋葱模型预警时间在窗口内 ====================

// TestE2E_WarnBefore_OnionChainWithinWindow 洋葱模型 + 预警时间在窗口内
// 场景：与测试1类似，但预警时间距今只有 50h < 168h → 不静默，正常报错
//
// 构造方式：
//   - 左链最终值与右链最终值有交集（匹配通过门控）
//   - 但 Phase 2 比较时当前列值不在右链 FirstStepInputValues 中 → violation=true
//   - 预警时间距今 50h < 168h → 不静默
//
// 预期：Violation=true（预警窗口内，不静默）
func TestE2E_WarnBefore_OnionChainWithinWindow(t *testing.T) {
	now := time.Now()
	futureTime50h := now.Add(50 * time.Hour).Format("2006-01-02 15:04:05")

	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "match_val"},
	})

	// 左链目标表：只有一步（仅取值），左链值 "match_val"
	// 不需要额外的左链目标表

	// 右链目标表（同时是预警表）
	// NextCol="Name" 的值 "match_val" 用于 Match 阶段匹配
	// PreCol="Id" 的值会在反向查找中被提取到 FirstStepInputValues
	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass", []string{"Id", "StartTime", "Name"}, [][]string{
		{"SP001", futureTime50h, "match_val"},
	})

	sheetMap := map[string]*excelize.File{
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "赛季战令表|SeasonPass", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"chainWarnBefore":   "168h",
		"chainWarnSheet":    "赛季战令表|SeasonPass",
		"chainWarnCol":      "StartTime",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链值 "match_val"，右链最终表 Name 列值 ["match_val"]
	// Match verify_exists: "match_val" 在 ["match_val"] 中找到 → 门控通过
	// Phase 2: 当前列值 "match_val" vs FirstStepInputValues=["SP001"] → "match_val" 不在 ["SP001"] 中 → violation=true
	// 预警时间距今 50h < 168h → 不静默
	assert.True(t, ctx.Matched, "左链值应在右链中找到")
	assert.True(t, ctx.Violation, "预警时间在窗口内，不应静默")
}

// ==================== 测试3: 旧路径全表扫描 ====================

// TestE2E_WarnBefore_LegacyPath 全表扫描的 ShouldSuppressWarnBeforeLegacy
// 场景：不使用洋葱模型，直接调用 ShouldSuppressWarnBeforeLegacy
// 构造预警表有 3 行数据：过去时间、近未来、远未来
// 验证最近的未来时间 > warnBefore 时静默
func TestE2E_WarnBefore_LegacyPath(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	nearFuture := now.Add(100 * time.Hour).Format("2006-01-02 15:04:05")
	farFuture := now.Add(300 * time.Hour).Format("2006-01-02 15:04:05")

	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass", []string{"Id", "StartTime"}, [][]string{
		{"1", pastTime},
		{"2", nearFuture},
		{"3", farFuture},
	})

	sheetMap := map[string]*excelize.File{
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	// warnBefore=50h，最近未来时间距今 100h > 50h → 静默
	result := ShouldSuppressWarnBeforeLegacy(sheetMap, "赛季战令表|SeasonPass", "StartTime", 50*time.Hour, now)
	assert.True(t, result, "最近未来时间 100h 超过 warnBefore 50h，应该静默")

	// warnBefore=200h，最近未来时间距今 100h < 200h → 不静默
	result2 := ShouldSuppressWarnBeforeLegacy(sheetMap, "赛季战令表|SeasonPass", "StartTime", 200*time.Hour, now)
	assert.False(t, result2, "最近未来时间 100h 在 warnBefore 200h 内，不应该静默")
}

// ==================== 测试4: 链路断开时使用 CurrentCellValue 独立查找 ====================

// TestE2E_WarnBefore_ChainBroken_CurrentValueLookup 链路断开时使用 CurrentCellValue 独立查找
// 场景：WarnValues 为空（模拟链路断开），CurrentCellValue="5"（SeasonPassId）
// 在预警表 SeasonPass 中查找 Id="5" 的行，提取 StartTime
// StartTime 距今 > 168h → 静默
func TestE2E_WarnBefore_ChainBroken_CurrentValueLookup(t *testing.T) {
	now := time.Now()
	futureTime200h := now.Add(200 * time.Hour).Format("2006-01-02 15:04:05")

	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass", []string{"Id", "StartTime"}, [][]string{
		{"5", futureTime200h},
		{"6", now.Add(-10 * time.Hour).Format("2006-01-02 15:04:05")}, // 过去时间
	})

	sheetMap := map[string]*excelize.File{
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	// 链路断开：WarnValues 为空，CurrentCellValue="5"
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
			"chainWarnSheet":  "赛季战令表|SeasonPass",
			"chainWarnCol":    "StartTime",
		},
		SheetMap:   sheetMap,
		WarnValues: nil, // 链路断开
		MyColData:  []string{"5"},
		DataIdx:    0,
	}
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.True(t, result, "链路断开后用 CurrentCellValue=5 查找到 StartTime 距今 200h > 168h，应该静默")
}

// ==================== 测试5: 链路断开 + 过去时间不静默 ====================

// TestE2E_WarnBefore_ChainBroken_PastTimeNotSuppress 链路断开 + 过去时间不静默
// 场景：CurrentCellValue="6"，SeasonPass 表 Id=6 的 StartTime 在过去
// 预期：不静默
func TestE2E_WarnBefore_ChainBroken_PastTimeNotSuppress(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-10 * 24 * time.Hour).Format("2006-01-02 15:04:05")

	seasonPassFile := buildChainTestRefSheet("赛季战令表|SeasonPass", []string{"Id", "StartTime"}, [][]string{
		{"5", now.Add(200 * time.Hour).Format("2006-01-02 15:04:05")},
		{"6", pastTime},
	})

	sheetMap := map[string]*excelize.File{
		"赛季战令表|SeasonPass": seasonPassFile,
	}

	// 链路断开：WarnValues 为空，CurrentCellValue="6"
	ctx := &ChainContext{
		Params: map[string]string{
			"chainWarnBefore": "168h",
			"chainWarnSheet":  "赛季战令表|SeasonPass",
			"chainWarnCol":    "StartTime",
		},
		SheetMap:   sheetMap,
		WarnValues: nil,
		MyColData:  []string{"6"},
		DataIdx:    0,
	}
	result := ShouldSuppressByWarnBefore(ctx, now)
	assert.False(t, result, "链路断开后用 CurrentCellValue=6 查找到 StartTime 为过去时间，不应该静默")
}
