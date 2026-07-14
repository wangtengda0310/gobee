package functiontest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// buildFightTestResult 测试。
//
// 复现 run_fight_test "空跑 success=true" bug：runTestFunc 返回空 logs
// （filterCases/caseName 未命中任何用例，0 用例执行）且无 error 时，
// 当前实现 Success=err==nil 会误报 true。期望 Success=false 并给出空跑提示。
// 根因详见 memory l6-run-fight-test-fake-pass：老会话 L6 空跑审计真实实跑 1/39。
// =====================================================================

// TestBuildFightTestResult_空跑空日志应判失败 复现 bug：
// 空 logs（0 用例执行）+ 无 error 时，应判 Success=false，而非误报 true。
func TestBuildFightTestResult_空跑空日志应判失败(t *testing.T) {
	logs := map[string][]LogEntry{} // 空跑：未命中任何用例，0 条协议日志

	result := buildFightTestResult(logs, nil)

	assert.False(t, result.Success, "空 logs（空跑）应判失败，不应误报 success=true")
	assert.Contains(t, result.Message, "未执行", "空跑应给出未执行提示，而非\"战斗测试完成\"")
}

// TestBuildFightTestResult_有协议日志判成功 回归保护：实跑（有协议日志）应判成功。
func TestBuildFightTestResult_有协议日志判成功(t *testing.T) {
	logs := map[string][]LogEntry{
		"TC-17_主动打出杀参与累计计数": {
			{Case: "TC-17_主动打出杀参与累计计数", Msg: "PlayCardAck Result=0"},
		},
	}

	result := buildFightTestResult(logs, nil)

	assert.True(t, result.Success, "有协议日志（实跑）应判成功")
	assert.Contains(t, result.Message, "战斗测试完成")
}

// TestBuildFightTestResult_执行出错判失败 错误路径：runTestFunc 返 error 时应判失败并带错误信息。
func TestBuildFightTestResult_执行出错判失败(t *testing.T) {
	result := buildFightTestResult(nil, fmt.Errorf("连接游戏服失败"))

	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "连接游戏服失败")
}

// =====================================================================
// validateFilterCaseNamesRan 测试。
//
// 复现 run_robot_test 空跑 success=true bug：RunRobotTest 恒返 nil（services.go:230），
// filterCaseNames 未命中时 0 用例执行仍 success。validate 在 handler 调 RunRobotTest 后
// 核对日志（log_service 缓存）：空日志或 filterCaseNames 缺失则返 error。
// =====================================================================

// TestValidateFilterCaseNamesRan_空日志判失败 空 logs（0 用例执行）应判失败。
func TestValidateFilterCaseNamesRan_空日志判失败(t *testing.T) {
	err := validateFilterCaseNamesRan(map[string][]LogEntry{}, nil)
	assert.ErrorContains(t, err, "未执行")
}

// TestValidateFilterCaseNamesRan_未指定filterCaseNames且有日志通过 未指定 filterCaseNames + 有日志 → 不核对，通过。
func TestValidateFilterCaseNamesRan_未指定filterCaseNames且有日志通过(t *testing.T) {
	logs := map[string][]LogEntry{"TC-17": {{Case: "TC-17"}}}
	err := validateFilterCaseNamesRan(logs, nil)
	assert.NoError(t, err)
}

// TestValidateFilterCaseNamesRan_指定case全命中通过 filterCaseNames 都在日志里 → 通过。
func TestValidateFilterCaseNamesRan_指定case全命中通过(t *testing.T) {
	logs := map[string][]LogEntry{
		"TC-17": {{Case: "TC-17"}},
		"TC-31": {{Case: "TC-31"}},
	}
	err := validateFilterCaseNamesRan(logs, []string{"TC-17", "TC-31"})
	assert.NoError(t, err)
}

// TestValidateFilterCaseNamesRan_指定case部分缺失判失败 filterCaseNames 有缺失 → 失败并指出缺失项。
func TestValidateFilterCaseNamesRan_指定case部分缺失判失败(t *testing.T) {
	logs := map[string][]LogEntry{"TC-17": {{Case: "TC-17"}}}
	err := validateFilterCaseNamesRan(logs, []string{"TC-17", "TC-999_不存在"})
	assert.ErrorContains(t, err, "TC-999_不存在")
}
