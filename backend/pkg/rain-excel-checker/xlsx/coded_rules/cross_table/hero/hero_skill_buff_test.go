package hero

import (
	"fmt"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// createHeroSkillBuffTestExcel 创建测试用 Excel 文件
// sheetName: Excel 中的 sheet 名（需与 sheetMap key 一致）
// headers: 列名列表（写入第3行）
// dataRows: 数据行（从第5行开始）
func createHeroSkillBuffTestExcel(sheetName string, headers []string, dataRows [][]string) *excelize.File {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)

	// 行3: 列名
	for i, h := range headers {
		col := string(rune('A' + i))
		_ = f.SetCellValue(sheetName, col+"3", h)
	}
	// 行4: 类型标记
	for i := range headers {
		col := string(rune('A' + i))
		_ = f.SetCellValue(sheetName, col+"4", "string")
	}
	// 行5+: 数据
	for rowIdx, row := range dataRows {
		for colIdx, val := range row {
			col := string(rune('A' + colIdx))
			_ = f.SetCellValue(sheetName, col+fmt.Sprintf("%d", 5+rowIdx), val)
		}
	}
	return f
}

// TestHeroSkillBuffCheck_AllValid 全部有效数据应通过
func TestHeroSkillBuffCheck_AllValid(t *testing.T) {
	// Hero 表
	heroFile := createHeroSkillBuffTestExcel("武将表|Hero", []string{"Id", "Name", "Skill", "Buff"}, [][]string{
		{"10001", "曹操", "1001,1002", ""},
		{"10002", "刘备", "1003", ""},
		{"19017", "Boss", "8040", "BuffBoss2026"},
	})
	heroCols, _ := heroFile.GetCols("武将表|Hero")

	// Skill 表
	skillFile := createHeroSkillBuffTestExcel("技能表|Skill", []string{"Id", "SkillName"}, [][]string{
		{"1001", "技能1"},
		{"1002", "技能2"},
		{"1003", "技能3"},
		{"8040", "Boss技能"},
	})

	// Buff 表（第2列 B 是英文字符串标识符，没有列名）
	buffFile := excelize.NewFile()
	buffFile.SetSheetName("Sheet1", "Buff表|Buff")
	_ = buffFile.SetCellValue("Buff表|Buff", "A3", "Id")
	_ = buffFile.SetCellValue("Buff表|Buff", "A4", "int")
	_ = buffFile.SetCellValue("Buff表|Buff", "B4", "string")
	_ = buffFile.SetCellValue("Buff表|Buff", "A5", "1")
	_ = buffFile.SetCellValue("Buff表|Buff", "B5", "BuffBoss2026")

	sheetMap := map[string]*excelize.File{
		"技能表|Skill":  skillFile,
		"Buff表|Buff": buffFile,
	}

	rule := &HeroSkillBuffCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        heroCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "全部有效时应通过，实际原因: %s", result.Reason)
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}

// TestHeroSkillBuffCheck_InvalidSkill 无效技能ID应报错
func TestHeroSkillBuffCheck_InvalidSkill(t *testing.T) {
	heroFile := createHeroSkillBuffTestExcel("武将表|Hero", []string{"Id", "Name", "Skill", "Buff"}, [][]string{
		{"10001", "曹操", "1001,9999", ""},
	})
	heroCols, _ := heroFile.GetCols("武将表|Hero")

	skillFile := createHeroSkillBuffTestExcel("技能表|Skill", []string{"Id", "SkillName"}, [][]string{
		{"1001", "技能1"},
	})

	buffFile := excelize.NewFile()
	buffFile.SetSheetName("Sheet1", "Buff表|Buff")
	_ = buffFile.SetCellValue("Buff表|Buff", "A3", "Id")

	sheetMap := map[string]*excelize.File{
		"技能表|Skill":  skillFile,
		"Buff表|Buff": buffFile,
	}

	rule := &HeroSkillBuffCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        heroCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在无效技能ID时应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "9999", "错误信息应包含无效ID")
	assert.Contains(t, result.ErrCells[0].Reason, "不存在", "错误信息应说明不存在")
}

// TestHeroSkillBuffCheck_InvalidBuff 无效Buff标识符应报错
func TestHeroSkillBuffCheck_InvalidBuff(t *testing.T) {
	heroFile := createHeroSkillBuffTestExcel("武将表|Hero", []string{"Id", "Name", "Skill", "Buff"}, [][]string{
		{"10001", "曹操", "1001", "BuffNotExist"},
	})
	heroCols, _ := heroFile.GetCols("武将表|Hero")

	skillFile := createHeroSkillBuffTestExcel("技能表|Skill", []string{"Id", "SkillName"}, [][]string{
		{"1001", "技能1"},
	})

	buffFile := excelize.NewFile()
	buffFile.SetSheetName("Sheet1", "Buff表|Buff")
	_ = buffFile.SetCellValue("Buff表|Buff", "A3", "Id")
	_ = buffFile.SetCellValue("Buff表|Buff", "A5", "1")
	_ = buffFile.SetCellValue("Buff表|Buff", "B5", "BuffBoss2026")

	sheetMap := map[string]*excelize.File{
		"技能表|Skill":  skillFile,
		"Buff表|Buff": buffFile,
	}

	rule := &HeroSkillBuffCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        heroCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "存在无效Buff时应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "BuffNotExist", "错误信息应包含无效标识符")
}

// TestHeroSkillBuffCheck_NoSkillSheet 缺少 Skill 表应报错
func TestHeroSkillBuffCheck_NoSkillSheet(t *testing.T) {
	heroFile := createHeroSkillBuffTestExcel("武将表|Hero", []string{"Id", "Name", "Skill", "Buff"}, [][]string{
		{"10001", "曹操", "1001", ""},
	})
	heroCols, _ := heroFile.GetCols("武将表|Hero")

	buffFile := excelize.NewFile()
	buffFile.SetSheetName("Sheet1", "Buff表|Buff")

	sheetMap := map[string]*excelize.File{
		"Buff表|Buff": buffFile,
	}

	rule := &HeroSkillBuffCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        heroCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.False(t, result.Ok, "缺少 Skill 表应报错")
	assert.Contains(t, result.Reason, "未找到 Skill 表", "应说明缺少 Skill 表")
}

// TestHeroSkillBuffCheck_EmptyData 空数据应通过
func TestHeroSkillBuffCheck_EmptyData(t *testing.T) {
	heroFile := createHeroSkillBuffTestExcel("武将表|Hero", []string{"Id", "Name", "Skill", "Buff"}, nil)
	heroCols, _ := heroFile.GetCols("武将表|Hero")

	skillFile := createHeroSkillBuffTestExcel("技能表|Skill", []string{"Id", "SkillName"}, nil)

	buffFile := excelize.NewFile()
	buffFile.SetSheetName("Sheet1", "Buff表|Buff")

	sheetMap := map[string]*excelize.File{
		"技能表|Skill":  skillFile,
		"Buff表|Buff": buffFile,
	}

	rule := &HeroSkillBuffCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "武将表|Hero",
		Cols:        heroCols,
		StartRowIdx: 4,
		EndIndex:    1000,
		Params:      map[string]string{},
		SheetMap:    sheetMap,
	})

	assert.True(t, result.Ok, "空数据应通过")
	assert.Len(t, result.ErrCells, 0, "不应有错误")
}
