package activitywikicheck

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTempRuleFile 在临时目录创建规则JSON文件，返回目录路径
func createTempRuleFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "test_rule.json"), []byte(content), 0644)
	assert.NoError(t, err)
	return dir
}

// TestGetRuleCoverage_EmptyDir 空目录不报错
func TestGetRuleCoverage_EmptyDir(t *testing.T) {
	svc := NewActivityWikiCheckService()
	dir := t.TempDir()

	result, err := svc.GetRuleCoverage(dir)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// 空目录无JSON文件，但DefaultTableRules仍会产生条目
	assert.NotEmpty(t, result.Sheets)
}

// TestGetRuleCoverage_EnabledFilter 验证Enabled==false的规则不被计数
func TestGetRuleCoverage_EnabledFilter(t *testing.T) {
	ruleJSON := `{
		"sheet": "测试表|TestSheet",
		"managerList": null,
		"rules": {
			"字段A": {
				"PropName": "字段A",
				"PropType": "int",
				"PropRules": [
					{"Type": "UNIQUE", "Uuid": "u1", "DisplayName": "唯一性", "Enabled": true, "Params": {}},
					{"Type": "NOT_EMPTY", "Uuid": "u2", "DisplayName": "非空", "Enabled": false, "Params": {}}
				]
			}
		},
		"tableRules": [
			{"type": "ARENA_SEASON_CHECK", "displayName": "赛季检查", "uuid": "t1", "description": "", "params": {}, "enabled": true},
			{"type": "ARENA_SEASON_CHECK", "displayName": "赛季检查2", "uuid": "t2", "description": "", "params": {}, "enabled": false}
		]
	}`
	dir := createTempRuleFile(t, ruleJSON)

	svc := NewActivityWikiCheckService()
	result, err := svc.GetRuleCoverage(dir)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	stats, ok := result.Sheets["测试表|TestSheet"]
	assert.True(t, ok, "应包含Sheet '测试表|TestSheet'")

	// 只有1个表级规则Enabled=true
	assert.Equal(t, 1, stats.TableRuleCount)

	// 字段A只有1个列级规则Enabled=true
	fieldStat, ok := stats.FieldRuleStats["字段A"]
	assert.True(t, ok)
	assert.Equal(t, 1, fieldStat.ColRuleCount)
	assert.Equal(t, 1, fieldStat.TotalCount)
}

// TestGetRuleCoverage_BasicCount 验证基本计数正确
func TestGetRuleCoverage_BasicCount(t *testing.T) {
	ruleJSON := `{
		"sheet": "角色成就|HeroAchieve",
		"managerList": null,
		"rules": {
			"Id": {
				"PropName": "Id",
				"PropType": "int",
				"PropRules": [
					{"Type": "UNIQUE", "Uuid": "u1", "DisplayName": "唯一性", "Enabled": true, "Params": {}},
					{"Type": "NOT_EMPTY", "Uuid": "u2", "DisplayName": "非空", "Enabled": true, "Params": {}}
				]
			},
			"Name": {
				"PropName": "Name",
				"PropType": "string",
				"PropRules": [
					{"Type": "NOT_EMPTY", "Uuid": "u3", "DisplayName": "非空", "Enabled": true, "Params": {}}
				]
			}
		},
		"tableRules": [
			{"type": "ARENA_SEASON_CHECK", "displayName": "赛季检查", "uuid": "t1", "description": "", "params": {}, "enabled": true}
		]
	}`
	dir := createTempRuleFile(t, ruleJSON)

	svc := NewActivityWikiCheckService()
	result, err := svc.GetRuleCoverage(dir)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	stats := result.Sheets["角色成就|HeroAchieve"]
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.TableRuleCount)
	assert.Equal(t, 2, stats.FieldRuleStats["Id"].ColRuleCount)
	assert.Equal(t, 2, stats.FieldRuleStats["Id"].TotalCount)
	assert.Equal(t, 1, stats.FieldRuleStats["Name"].ColRuleCount)
}

// TestGetRuleCoverage_DefaultTableRules 验证默认规则被计入
func TestGetRuleCoverage_DefaultTableRules(t *testing.T) {
	// 创建一个DrawSkin的SheetRule，不配置表级规则
	// DefaultTableRules["DrawSkin"] 有多个默认规则，应被累加
	ruleJSON := `{
		"sheet": "皮肤抽奖|DrawSkin",
		"managerList": null,
		"rules": {
			"Id": {
				"PropName": "Id",
				"PropType": "int",
				"PropRules": [
					{"Type": "UNIQUE", "Uuid": "u1", "DisplayName": "唯一性", "Enabled": true, "Params": {}}
				]
			}
		},
		"tableRules": []
	}`
	dir := createTempRuleFile(t, ruleJSON)

	svc := NewActivityWikiCheckService()
	result, err := svc.GetRuleCoverage(dir)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	stats := result.Sheets["皮肤抽奖|DrawSkin"]
	assert.NotNil(t, stats)

	// tableRules为空数组(0个)，DefaultTableRules["DrawSkin"]有8个默认规则
	// 总共应为 0 + 8 = 8
	assert.Equal(t, 8, stats.TableRuleCount, "应包含DrawSkin的8个默认表级规则")

	// 列级规则：Id字段有1个Enabled规则
	assert.Equal(t, 1, stats.FieldRuleStats["Id"].ColRuleCount)
}

// TestGetRuleCoverage_DefaultTableRules_NoSheetRule 验证无SheetRule时默认规则仍被计入
func TestGetRuleCoverage_DefaultTableRules_NoSheetRule(t *testing.T) {
	// 空目录，无任何JSON规则文件
	dir := t.TempDir()

	svc := NewActivityWikiCheckService()
	result, err := svc.GetRuleCoverage(dir)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// DefaultTableRules中每个key都会创建一个条目
	// 检查几个已知的短名
	_, hasHero := result.Sheets["Hero"]
	assert.True(t, hasHero, "应包含Hero的默认规则统计")

	_, hasDrawSkin := result.Sheets["DrawSkin"]
	assert.True(t, hasDrawSkin, "应包含DrawSkin的默认规则统计")
}
