// Package cross_table 提供跨表级别的校验规则
// 本包中的规则需要读取多个 Excel 表才能完成校验

package hero

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// TestHeroMeltCheckRule 测试武将熔炼检查
// 业务场景：验证武将熔炼配置正确性
func TestHeroMeltCheckRule(t *testing.T) {
	cols, sheetMap := createHeroMeltTestData()

	rule := new(HeroMeltCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 输出结果便于调试
	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Result:", string(jsonData))

	// 基本验证：测试能够正常运行
	if result == nil {
		t.Error("检查结果不应为空")
	}
}

// TestHeroMeltCheckRule_HeroTypeFilter 测试 HeroType 过滤
// 业务场景：只检查 HeroType=1 的武将，HeroType≠1 的武将应该被跳过
func TestHeroMeltCheckRule_HeroTypeFilter(t *testing.T) {
	// 测试数据：
	// 武将1: HeroType=1, 已开放, CanMelt=false → 应报错
	// 武将2: HeroType=2, 已开放, CanMelt=false → 应跳过，不报错
	cols, sheetMap := createHeroMeltTestDataWithHeroType()

	rule := new(HeroMeltCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 应该只有1个错误（HeroType=1 的武将）
	if len(result.ErrCells) != 1 {
		t.Errorf("期望 1 个错误，实际得到 %d 个: %v", len(result.ErrCells), result.ErrCells)
	}

	// 错误应该是关于 HeroType=1 的武将
	if len(result.ErrCells) > 0 {
		reason := result.ErrCells[0].Reason
		if !strings.Contains(reason, "已开放但CanMelt=false") {
			t.Errorf("错误原因不正确: %s", reason)
		}
	}
}

// TestHeroMeltCheckRule_MeltPowerValidation_CanMeltTrue 测试 MeltPower 验证（CanMelt=true）
// 业务场景：CanMelt=true 时，MeltPower 应为 16/24/32
func TestHeroMeltCheckRule_MeltPowerValidation_CanMeltTrue(t *testing.T) {
	// 创建测试数据：CanMelt=true 但 MeltPower=100（错误）
	cols, sheetMap := createHeroMeltTestDataWithInvalidMeltPower(true)

	rule := new(HeroMeltCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 应该有 MeltPower 验证错误
	hasMeltPowerError := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "MeltPower") {
			hasMeltPowerError = true
			if !strings.Contains(err.Reason, "应为16/24/32") {
				t.Errorf("MeltPower 错误原因不正确: %s", err.Reason)
			}
		}
	}

	if !hasMeltPowerError {
		t.Error("期望有 MeltPower 验证错误，但未找到")
	}
}

// TestHeroMeltCheckRule_MeltPowerValidation_CanMeltFalse 测试 MeltPower 验证（CanMelt=false）
// 业务场景：CanMelt=false 时，MeltPower 应为 100
func TestHeroMeltCheckRule_MeltPowerValidation_CanMeltFalse(t *testing.T) {
	// 创建测试数据：CanMelt=false 但 MeltPower=16（错误）
	cols, sheetMap := createHeroMeltTestDataWithInvalidMeltPower(false)

	rule := new(HeroMeltCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 应该有 MeltPower 验证错误
	hasMeltPowerError := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "MeltPower") {
			hasMeltPowerError = true
			if !strings.Contains(err.Reason, "应为100") {
				t.Errorf("MeltPower 错误原因不正确: %s", err.Reason)
			}
		}
	}

	if !hasMeltPowerError {
		t.Error("期望有 MeltPower 验证错误，但未找到")
	}
}

// TestHeroMeltCheckRule_ValidMeltPower 测试正确的 MeltPower 值
// 业务场景：MeltPower=16/24/32 且 CanMelt=true，或 MeltPower=100 且 CanMelt=false，应该通过
func TestHeroMeltCheckRule_ValidMeltPower(t *testing.T) {
	cols, sheetMap := createHeroMeltTestDataWithValidMeltPower()

	rule := new(HeroMeltCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 筛选出 MeltPower 相关的错误
	hasMeltPowerError := false
	for _, err := range result.ErrCells {
		if strings.Contains(err.Reason, "MeltPower") {
			hasMeltPowerError = true
			break
		}
	}

	if hasMeltPowerError {
		t.Errorf("期望 MeltPower 验证通过，但有错误: %v", result.ErrCells)
	}
}

// ==================== 测试数据创建函数 ====================

// createHeroMeltTestData 创建武将熔炼检查的基础测试数据
// 验证规则：HERO_MELT_CHECK（武将熔炼检查）
// 业务场景：基础的 Hero 和 SkillMelt 表数据
func createHeroMeltTestData() (cols [][]string, sheetMap map[string]*excelize.File) {
	// Hero 表
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	// 表头（4行表头 + 数据）
	heroFile.SetCellValue(sheetName, "A1", "")   // 空
	heroFile.SetCellValue(sheetName, "A2", "")   // 空
	heroFile.SetCellValue(sheetName, "A3", "Id") // 字段名
	heroFile.SetCellValue(sheetName, "A4", "")   // 导出标识
	heroFile.SetCellValue(sheetName, "B3", "Name")
	heroFile.SetCellValue(sheetName, "C3", "HeroType")
	heroFile.SetCellValue(sheetName, "D3", "IsOpen")
	heroFile.SetCellValue(sheetName, "E3", "OpenDate")
	heroFile.SetCellValue(sheetName, "F3", "CanMelt")
	heroFile.SetCellValue(sheetName, "G3", "Skill")

	// 数据行（从第5行开始）
	pastDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05")

	// 武将1: HeroType=1, 已开放, CanMelt=true
	heroFile.SetCellValue(sheetName, "A5", "10001")
	heroFile.SetCellValue(sheetName, "B5", "测试武将1")
	heroFile.SetCellValue(sheetName, "C5", "1")
	heroFile.SetCellValue(sheetName, "D5", "true")
	heroFile.SetCellValue(sheetName, "E5", pastDate)
	heroFile.SetCellValue(sheetName, "F5", "true")
	heroFile.SetCellValue(sheetName, "G5", "[1001,1002]")

	// SkillMelt 表
	skillMeltFile := excelize.NewFile()
	meltSheetName := "SkillMelt"
	skillMeltFile.SetSheetName("Sheet1", meltSheetName)

	// 表头
	skillMeltFile.SetCellValue(meltSheetName, "A1", "")
	skillMeltFile.SetCellValue(meltSheetName, "A2", "")
	skillMeltFile.SetCellValue(meltSheetName, "A3", "Id")
	skillMeltFile.SetCellValue(meltSheetName, "A4", "")
	skillMeltFile.SetCellValue(meltSheetName, "B3", "CanMelt")
	skillMeltFile.SetCellValue(meltSheetName, "B4", "")
	skillMeltFile.SetCellValue(meltSheetName, "C3", "MeltPower")
	skillMeltFile.SetCellValue(meltSheetName, "C4", "")

	// 数据行
	skillMeltFile.SetCellValue(meltSheetName, "A5", "1001")
	skillMeltFile.SetCellValue(meltSheetName, "B5", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C5", "16")

	skillMeltFile.SetCellValue(meltSheetName, "A6", "1002")
	skillMeltFile.SetCellValue(meltSheetName, "B6", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C6", "24")

	cols, _ = heroFile.GetCols(sheetName)

	sheetMap = map[string]*excelize.File{
		sheetName:     heroFile,
		meltSheetName: skillMeltFile,
	}

	return cols, sheetMap
}

// createHeroMeltTestDataWithHeroType 创建包含 HeroType 字段的测试数据
// 验证规则：HERO_MELT_CHECK - HeroType 过滤（只检查 HeroType=1）
// 业务场景：
//   - 武将1: HeroType=1, 已开放, CanMelt=false → 应报错
//   - 武将2: HeroType=2, 已开放, CanMelt=false → 应跳过
func createHeroMeltTestDataWithHeroType() (cols [][]string, sheetMap map[string]*excelize.File) {
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	// 表头
	heroFile.SetCellValue(sheetName, "A1", "")
	heroFile.SetCellValue(sheetName, "A2", "")
	heroFile.SetCellValue(sheetName, "A3", "Id")
	heroFile.SetCellValue(sheetName, "A4", "")
	heroFile.SetCellValue(sheetName, "B3", "Name")
	heroFile.SetCellValue(sheetName, "C3", "HeroType")
	heroFile.SetCellValue(sheetName, "D3", "IsOpen")
	heroFile.SetCellValue(sheetName, "E3", "OpenDate")
	heroFile.SetCellValue(sheetName, "F3", "CanMelt")
	heroFile.SetCellValue(sheetName, "G3", "Skill")

	pastDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05")

	// 武将1: HeroType=1, 已开放, CanMelt=false → 应报错
	heroFile.SetCellValue(sheetName, "A5", "10001")
	heroFile.SetCellValue(sheetName, "B5", "普通武将")
	heroFile.SetCellValue(sheetName, "C5", "1") // HeroType=1
	heroFile.SetCellValue(sheetName, "D5", "true")
	heroFile.SetCellValue(sheetName, "E5", pastDate)
	heroFile.SetCellValue(sheetName, "F5", "false")
	heroFile.SetCellValue(sheetName, "G5", "[1001]")

	// 武将2: HeroType=2, 已开放, CanMelt=false → 应跳过，不报错
	heroFile.SetCellValue(sheetName, "A6", "10002")
	heroFile.SetCellValue(sheetName, "B6", "特殊武将")
	heroFile.SetCellValue(sheetName, "C6", "2") // HeroType=2
	heroFile.SetCellValue(sheetName, "D6", "true")
	heroFile.SetCellValue(sheetName, "E6", pastDate)
	heroFile.SetCellValue(sheetName, "F6", "false")
	heroFile.SetCellValue(sheetName, "G6", "[1002]")

	// SkillMelt 表
	skillMeltFile := excelize.NewFile()
	meltSheetName := "SkillMelt"
	skillMeltFile.SetSheetName("Sheet1", meltSheetName)

	skillMeltFile.SetCellValue(meltSheetName, "A1", "")
	skillMeltFile.SetCellValue(meltSheetName, "A2", "")
	skillMeltFile.SetCellValue(meltSheetName, "A3", "Id")
	skillMeltFile.SetCellValue(meltSheetName, "A4", "")
	skillMeltFile.SetCellValue(meltSheetName, "B3", "CanMelt")
	skillMeltFile.SetCellValue(meltSheetName, "B4", "")
	skillMeltFile.SetCellValue(meltSheetName, "C3", "MeltPower")
	skillMeltFile.SetCellValue(meltSheetName, "C4", "")

	skillMeltFile.SetCellValue(meltSheetName, "A5", "1001")
	skillMeltFile.SetCellValue(meltSheetName, "B5", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C5", "16")

	skillMeltFile.SetCellValue(meltSheetName, "A6", "1002")
	skillMeltFile.SetCellValue(meltSheetName, "B6", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C6", "16")

	cols, _ = heroFile.GetCols(sheetName)

	sheetMap = map[string]*excelize.File{
		sheetName:     heroFile,
		meltSheetName: skillMeltFile,
	}

	return cols, sheetMap
}

// createHeroMeltTestDataWithInvalidMeltPower 创建包含无效 MeltPower 的测试数据
// 验证规则：HERO_MELT_CHECK - MeltPower 值验证
// 业务场景：
//   - canMelt=true: MeltPower 应为 16/24/32，但设为 100 触发错误
//   - canMelt=false: MeltPower 应为 100，但设为 16 触发错误
func createHeroMeltTestDataWithInvalidMeltPower(canMelt bool) (cols [][]string, sheetMap map[string]*excelize.File) {
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	// 表头
	heroFile.SetCellValue(sheetName, "A1", "")
	heroFile.SetCellValue(sheetName, "A2", "")
	heroFile.SetCellValue(sheetName, "A3", "Id")
	heroFile.SetCellValue(sheetName, "A4", "")
	heroFile.SetCellValue(sheetName, "B3", "Name")
	heroFile.SetCellValue(sheetName, "C3", "HeroType")
	heroFile.SetCellValue(sheetName, "D3", "IsOpen")
	heroFile.SetCellValue(sheetName, "E3", "OpenDate")
	heroFile.SetCellValue(sheetName, "F3", "CanMelt")
	heroFile.SetCellValue(sheetName, "G3", "Skill")

	pastDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05")
	canMeltStr := "false"
	if canMelt {
		canMeltStr = "true"
	}

	heroFile.SetCellValue(sheetName, "A5", "10001")
	heroFile.SetCellValue(sheetName, "B5", "测试武将")
	heroFile.SetCellValue(sheetName, "C5", "1")
	heroFile.SetCellValue(sheetName, "D5", "true")
	heroFile.SetCellValue(sheetName, "E5", pastDate)
	heroFile.SetCellValue(sheetName, "F5", canMeltStr)
	heroFile.SetCellValue(sheetName, "G5", "[1001]")

	// SkillMelt 表 - 设置无效的 MeltPower
	skillMeltFile := excelize.NewFile()
	meltSheetName := "SkillMelt"
	skillMeltFile.SetSheetName("Sheet1", meltSheetName)

	skillMeltFile.SetCellValue(meltSheetName, "A1", "")
	skillMeltFile.SetCellValue(meltSheetName, "A2", "")
	skillMeltFile.SetCellValue(meltSheetName, "A3", "Id")
	skillMeltFile.SetCellValue(meltSheetName, "A4", "")
	skillMeltFile.SetCellValue(meltSheetName, "B3", "CanMelt")
	skillMeltFile.SetCellValue(meltSheetName, "B4", "")
	skillMeltFile.SetCellValue(meltSheetName, "C3", "MeltPower")
	skillMeltFile.SetCellValue(meltSheetName, "C4", "")

	// 设置无效的 MeltPower
	invalidMeltPower := "100"
	if canMelt {
		// CanMelt=true 时，MeltPower 应为 16/24/32
		invalidMeltPower = "100"
	} else {
		// CanMelt=false 时，MeltPower 应为 100
		invalidMeltPower = "16"
	}

	skillMeltFile.SetCellValue(meltSheetName, "A5", "1001")
	skillMeltFile.SetCellValue(meltSheetName, "B5", canMeltStr)
	skillMeltFile.SetCellValue(meltSheetName, "C5", invalidMeltPower)

	cols, _ = heroFile.GetCols(sheetName)

	sheetMap = map[string]*excelize.File{
		sheetName:     heroFile,
		meltSheetName: skillMeltFile,
	}

	return cols, sheetMap
}

// createHeroMeltTestDataWithValidMeltPower 创建包含有效 MeltPower 的测试数据
// 验证规则：HERO_MELT_CHECK - MeltPower 值验证
// 业务场景：
//   - CanMelt=true + MeltPower=16 → 通过
//   - CanMelt=true + MeltPower=24 → 通过
//   - CanMelt=true + MeltPower=32 → 通过
//   - CanMelt=false + MeltPower=100 → 通过
func createHeroMeltTestDataWithValidMeltPower() (cols [][]string, sheetMap map[string]*excelize.File) {
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	// 表头
	heroFile.SetCellValue(sheetName, "A1", "")
	heroFile.SetCellValue(sheetName, "A2", "")
	heroFile.SetCellValue(sheetName, "A3", "Id")
	heroFile.SetCellValue(sheetName, "A4", "")
	heroFile.SetCellValue(sheetName, "B3", "Name")
	heroFile.SetCellValue(sheetName, "C3", "HeroType")
	heroFile.SetCellValue(sheetName, "D3", "IsOpen")
	heroFile.SetCellValue(sheetName, "E3", "OpenDate")
	heroFile.SetCellValue(sheetName, "F3", "CanMelt")
	heroFile.SetCellValue(sheetName, "G3", "Skill")

	pastDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05")

	// 武将1: CanMelt=true + MeltPower=16
	heroFile.SetCellValue(sheetName, "A5", "10001")
	heroFile.SetCellValue(sheetName, "B5", "武将16")
	heroFile.SetCellValue(sheetName, "C5", "1")
	heroFile.SetCellValue(sheetName, "D5", "true")
	heroFile.SetCellValue(sheetName, "E5", pastDate)
	heroFile.SetCellValue(sheetName, "F5", "true")
	heroFile.SetCellValue(sheetName, "G5", "[1001]")

	// 武将2: CanMelt=true + MeltPower=24
	heroFile.SetCellValue(sheetName, "A6", "10002")
	heroFile.SetCellValue(sheetName, "B6", "武将24")
	heroFile.SetCellValue(sheetName, "C6", "1")
	heroFile.SetCellValue(sheetName, "D6", "true")
	heroFile.SetCellValue(sheetName, "E6", pastDate)
	heroFile.SetCellValue(sheetName, "F6", "true")
	heroFile.SetCellValue(sheetName, "G6", "[1002]")

	// 武将3: CanMelt=true + MeltPower=32
	heroFile.SetCellValue(sheetName, "A7", "10003")
	heroFile.SetCellValue(sheetName, "B7", "武将32")
	heroFile.SetCellValue(sheetName, "C7", "1")
	heroFile.SetCellValue(sheetName, "D7", "true")
	heroFile.SetCellValue(sheetName, "E7", pastDate)
	heroFile.SetCellValue(sheetName, "F7", "true")
	heroFile.SetCellValue(sheetName, "G7", "[1003]")

	// 武将4: CanMelt=false + MeltPower=100
	heroFile.SetCellValue(sheetName, "A8", "10004")
	heroFile.SetCellValue(sheetName, "B8", "武将100")
	heroFile.SetCellValue(sheetName, "C8", "2") // HeroType=2，跳过检查
	heroFile.SetCellValue(sheetName, "D8", "false")
	heroFile.SetCellValue(sheetName, "E8", pastDate)
	heroFile.SetCellValue(sheetName, "F8", "false")
	heroFile.SetCellValue(sheetName, "G8", "[1004]")

	// SkillMelt 表
	skillMeltFile := excelize.NewFile()
	meltSheetName := "SkillMelt"
	skillMeltFile.SetSheetName("Sheet1", meltSheetName)

	skillMeltFile.SetCellValue(meltSheetName, "A1", "")
	skillMeltFile.SetCellValue(meltSheetName, "A2", "")
	skillMeltFile.SetCellValue(meltSheetName, "A3", "Id")
	skillMeltFile.SetCellValue(meltSheetName, "A4", "")
	skillMeltFile.SetCellValue(meltSheetName, "B3", "CanMelt")
	skillMeltFile.SetCellValue(meltSheetName, "B4", "")
	skillMeltFile.SetCellValue(meltSheetName, "C3", "MeltPower")
	skillMeltFile.SetCellValue(meltSheetName, "C4", "")

	skillMeltFile.SetCellValue(meltSheetName, "A5", "1001")
	skillMeltFile.SetCellValue(meltSheetName, "B5", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C5", "16")

	skillMeltFile.SetCellValue(meltSheetName, "A6", "1002")
	skillMeltFile.SetCellValue(meltSheetName, "B6", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C6", "24")

	skillMeltFile.SetCellValue(meltSheetName, "A7", "1003")
	skillMeltFile.SetCellValue(meltSheetName, "B7", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C7", "32")

	skillMeltFile.SetCellValue(meltSheetName, "A8", "1004")
	skillMeltFile.SetCellValue(meltSheetName, "B8", "false")
	skillMeltFile.SetCellValue(meltSheetName, "C8", "100")

	cols, _ = heroFile.GetCols(sheetName)

	sheetMap = map[string]*excelize.File{
		sheetName:     heroFile,
		meltSheetName: skillMeltFile,
	}

	return cols, sheetMap
}

// TestHeroMeltCheckRule_IsOpenFilter 测试 IsOpen 过滤
// 业务场景：未开放武将（IsOpen=0 或 OpenDate未到）的技能配置不应被检查
// 验证规则：HERO_MELT_CHECK - 技能配置检查仅针对已开放武将
// 业务数据：
//   - 武将1: IsOpen=0, HeroType=1, 技能未配置 → 应跳过，不报错
//   - 武将2: IsOpen=1, OpenDate已过, HeroType=1, 技能未配置 → 应报错
func TestHeroMeltCheckRule_IsOpenFilter(t *testing.T) {
	cols, sheetMap := createHeroMeltTestDataWithIsOpenFilter()

	rule := new(HeroMeltCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Hero",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      make(map[string]string),
		SheetMap:    sheetMap,
	})

	// 应该只有1个错误（已开放武将的技能未配置错误）
	if len(result.ErrCells) != 1 {
		t.Errorf("期望 1 个错误（已开放武将），实际得到 %d 个: %v", len(result.ErrCells), result.ErrCells)
	}

	// 错误应该是关于已开放武将的
	if len(result.ErrCells) > 0 {
		reason := result.ErrCells[0].Reason
		if !strings.Contains(reason, "未配置") {
			t.Errorf("错误原因不正确，期望'未配置'，实际: %s", reason)
		}
	}
}

// createHeroMeltTestDataWithIsOpenFilter 创建包含 IsOpen 字段过滤的测试数据
// 验证规则：HERO_MELT_CHECK - IsOpen 过滤（只检查已开放武将的技能配置）
// 业务场景：
//   - 武将1: IsOpen=0, HeroType=1, 技能未配置 → 应跳过
//   - 武将2: IsOpen=1, OpenDate已过, HeroType=1, 技能未配置 → 应报错
func createHeroMeltTestDataWithIsOpenFilter() (cols [][]string, sheetMap map[string]*excelize.File) {
	heroFile := excelize.NewFile()
	sheetName := "Hero"
	heroFile.SetSheetName("Sheet1", sheetName)

	// 表头
	heroFile.SetCellValue(sheetName, "A1", "")
	heroFile.SetCellValue(sheetName, "A2", "")
	heroFile.SetCellValue(sheetName, "A3", "Id")
	heroFile.SetCellValue(sheetName, "A4", "")
	heroFile.SetCellValue(sheetName, "B3", "Name")
	heroFile.SetCellValue(sheetName, "C3", "HeroType")
	heroFile.SetCellValue(sheetName, "D3", "IsOpen")
	heroFile.SetCellValue(sheetName, "E3", "OpenDate")
	heroFile.SetCellValue(sheetName, "F3", "CanMelt")
	heroFile.SetCellValue(sheetName, "G3", "Skill")

	pastDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05")

	// 武将1: IsOpen=0, HeroType=1 → 未开放，应跳过技能配置检查
	heroFile.SetCellValue(sheetName, "A5", "10001")
	heroFile.SetCellValue(sheetName, "B5", "未开放武将")
	heroFile.SetCellValue(sheetName, "C5", "1")
	heroFile.SetCellValue(sheetName, "D5", "false") // IsOpen=0
	heroFile.SetCellValue(sheetName, "E5", pastDate)
	heroFile.SetCellValue(sheetName, "F5", "true")
	heroFile.SetCellValue(sheetName, "G5", "[9999]") // 技能9999不存在，但因为IsOpen=0应跳过检查

	// 武将2: IsOpen=1, OpenDate已过, HeroType=1 → 已开放，应检查技能配置
	heroFile.SetCellValue(sheetName, "A6", "10002")
	heroFile.SetCellValue(sheetName, "B6", "已开放武将")
	heroFile.SetCellValue(sheetName, "C6", "1")
	heroFile.SetCellValue(sheetName, "D6", "true") // IsOpen=1
	heroFile.SetCellValue(sheetName, "E6", pastDate)
	heroFile.SetCellValue(sheetName, "F6", "true")
	heroFile.SetCellValue(sheetName, "G6", "[8888]") // 技能8888不存在，应报错

	// SkillMelt 表（不包含9999和8888）
	skillMeltFile := excelize.NewFile()
	meltSheetName := "SkillMelt"
	skillMeltFile.SetSheetName("Sheet1", meltSheetName)

	skillMeltFile.SetCellValue(meltSheetName, "A1", "")
	skillMeltFile.SetCellValue(meltSheetName, "A2", "")
	skillMeltFile.SetCellValue(meltSheetName, "A3", "Id")
	skillMeltFile.SetCellValue(meltSheetName, "A4", "")
	skillMeltFile.SetCellValue(meltSheetName, "B3", "CanMelt")
	skillMeltFile.SetCellValue(meltSheetName, "B4", "")
	skillMeltFile.SetCellValue(meltSheetName, "C3", "MeltPower")
	skillMeltFile.SetCellValue(meltSheetName, "C4", "")

	skillMeltFile.SetCellValue(meltSheetName, "A5", "1001")
	skillMeltFile.SetCellValue(meltSheetName, "B5", "true")
	skillMeltFile.SetCellValue(meltSheetName, "C5", "16")

	cols, _ = heroFile.GetCols(sheetName)

	sheetMap = map[string]*excelize.File{
		sheetName:     heroFile,
		meltSheetName: skillMeltFile,
	}

	return cols, sheetMap
}
