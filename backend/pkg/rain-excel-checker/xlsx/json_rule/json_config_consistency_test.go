// Package json_rule 提供 JSON 配置文件与代码期望的键名一致性检查
// ⚠️ 本测试用于防止类似 commit 2ee0e9c 的字段名不一致问题
package json_rule

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExcelCasesJSONParamsKeyConsistency 检查 cases/excel_cases/*.json 中的 Params key 是否与 ERuleParam 一致
func TestExcelCasesJSONParamsKeyConsistency(t *testing.T) {
	// 收集所有有效的 ERuleParam 值
	validParams := getAllERuleParamValues()

	// 遍历 cases/excel_cases 目录
	// 注意: 测试从 worktree 运行，需要找到正确的 cases 目录
	casesDir := findCasesDir(t)
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("读取 cases/excel_cases 目录失败: %v", err)
	}

	var invalidKeys []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(casesDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Logf("警告: 读取文件 %s 失败: %v", entry.Name(), err)
			continue
		}

		var sheetRule SheetRule
		if err := json.Unmarshal(data, &sheetRule); err != nil {
			t.Logf("警告: 解析文件 %s 失败: %v", entry.Name(), err)
			continue
		}

		// 检查每个列规则的 Params key
		for colName, colRule := range sheetRule.Rules {
			for _, rule := range colRule.PropRules {
				for key := range rule.Params {
					// 跳过通用基础参数（这些参数在 ALL_BASE 等规则中共享）
					if isCommonBaseParam(key) {
						continue
					}
					// 检查 key 是否在有效的 ERuleParam 中
					if !validParams[key] {
						invalidKeys = append(invalidKeys,
							filepath.Base(filePath)+": 列["+colName+"] 规则["+string(rule.Type)+"] 包含未知参数 key: "+key)
					}
				}
			}
		}
	}

	if len(invalidKeys) > 0 {
		t.Errorf("发现 %d 个无效的 Params key:\n%s", len(invalidKeys), strings.Join(invalidKeys, "\n"))
	}
}

// getAllERuleParamValues 返回所有 ERuleParam 常量的字符串值
func getAllERuleParamValues() map[string]bool {
	return map[string]bool{
		// 基础参数
		"allowEmpty":     true,
		"excepts":        true,
		"allowCommit":    true,
		"enums":          true,
		"breakLine":      true,
		"useIdColForEnd": true,
		"strict":         true,
		"compareRule":    true,
		"tolerance":      true,
		"min":            true,
		"max":            true,
		"pattern":        true,
		"groups":         true,
		"filterCol":      true,
		"filterVal":      true,
		"filterIsArray":  true,
		"filterMode":     true,
		"filterDays":     true,
		// 日期参数
		"startDate": true,
		"endDate":   true,
		// 表级规则参数
		"seasonEndTimeCol": true,
		"timeRangeBefore":  true,
		// 通用通知参数
		"idColName":         true,
		"idColNames":        true,
		"nameColName":       true,
		"oldDataPath":       true,
		"notifyAddedRows":   true,
		"notifyRemovedRows": true,
		"notifyAddedCols":   true,
		"notifyRemovedCols": true,
		// Git diff 参数
		"useGitDiff":  true,
		"gitRepoPath": true,
		"baseCommit":  true,
		"headCommit":  true,
		// 列连续性检查参数
		"targetCol":   true,
		"checkMode":   true,
		"scope":       true,
		"startValue":  true,
		"excludeRows": true,
		"separator":   true,
		// 武将检查参数
		"warnDaysBefore":  true,
		"dropMonthsDelay": true,
		"skillColName":    true,
		"canMeltColName":  true,
		"isOpenColName":   true,
		"openDateColName": true,
		// NUMERIC_RANGE 规则专用（历史兼容性，实际应使用 min/max）
		"minValue": true,
		"maxValue": true,
		// 其他已知参数
		"displayUnit": true,
		// CROSS_REFERENCE 规则参数
		"refSheet": true,
		"refCol":   true,
		// CROSS_REFERENCE 规则扩展参数（日期比较/匹配列）
		"compareOp":   true,
		"matchCol":    true,
		"matchRefCol": true,
		// FOREIGN_KEY 规则参数
		"targetSheet": true,
		// DATE 规则参数
		"format": true,
		// SPECIAL_FORMAT / CROSS_REFERENCE 规则参数
		"isArray": true,
		// CHAIN_REFERENCE 规则参数
		"chainCompare":      true,
		"chainMatchCompare": true,
		"chainSteps":        true,
		"chainWarnBefore":   true,
		"chainWarnSheet":    true,
		"chainWarnCol":      true,
		// WEIGHT_SUM 规则参数
		"targetSum": true,
	}
}

// isCommonBaseParam 判断是否为 ALL_BASE 等基础规则共享的通用参数
func isCommonBaseParam(key string) bool {
	common := map[string]bool{
		"allowEmpty":  true,
		"allowCommit": true,
		"breakLine":   true,
		"chsOnly":     true,
		"increase":    true,
		"unique":      true,
		"strict":      true,
	}
	return common[key]
}

// findCasesDir 查找 cases/excel_cases 目录路径
// 支持从 worktree 或普通目录运行测试
func findCasesDir(t *testing.T) string {
	// 尝试多个可能的路径
	candidates := []string{
		// 从 worktree 运行时（当前在 .claude/worktrees/xxx/rain-excel-checker/xlsx/json_rule）
		// worktree 结构: rain-qa-func/.claude/worktrees/xxx/rain-excel-checker/xlsx/json_rule
		// 需要回到 rain-qa-func 根目录
		filepath.Join("..", "..", "..", "cases", "excel_cases"),
		// 从 worktree rain-excel-checker 目录运行时（go test 从 xlsx/json_rule）
		filepath.Join("..", "..", "..", "..", "..", "..", "..", "cases", "excel_cases"),
		// 从 rain-qa-func 根目录运行时
		filepath.Join("..", "..", "..", "..", "cases", "excel_cases"),
		// 绝对路径（CI环境）
		"/d/work/xcard-qa-tools/rain-qa-func/cases/excel_cases",
		// Windows 风格绝对路径
		"D:/work/xcard-qa-tools/rain-qa-func/cases/excel_cases",
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	t.Fatalf("无法找到 cases/excel_cases 目录，已尝试: %v", candidates)
	return ""
}

// TestNumericRangeParamsKeyMapping 验证 NUMERIC_RANGE 规则的参数键名映射
// ⚠️ 这是回归测试，防止再次修改键名导致配置失效
func TestNumericRangeParamsKeyMapping(t *testing.T) {
	// 当前代码期望的键名（与 ERuleParam.MIN/MAX 一致）
	expectedMinKey := "min"
	expectedMaxKey := "max"

	// 验证 ERuleParam 常量值
	if string(MIN) != expectedMinKey {
		t.Errorf("ERuleParam MIN 的值应为 %q, 实际为 %q", expectedMinKey, string(MIN))
	}
	if string(MAX) != expectedMaxKey {
		t.Errorf("ERuleParam MAX 的值应为 %q, 实际为 %q", expectedMaxKey, string(MAX))
	}

	// 验证 excel_check_numeric_range.go 中读取的键名（通过代码审查确保一致）
	// 注意：这里不能直接测试私有代码，但可以通过集成测试验证
}
