// fight_test_report.go 的单元测试；该包与 services.go 末尾 persistFightTestReportAfterRun 须同次提交。
package functiontest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyFailureKind(t *testing.T) {
	tests := []struct {
		msg  string
		typ  string
		want string
	}{
		{"PropValue判断错误:2 累计 实际值为:1", "ASSET", "asset_mismatch"},
		{"等待消息超时 PlayCardAck", "ASSET", "asset_timeout"},
		{"存在未匹配到的Asset", "ASSET", "asset_unmatched"},
		{"[CORE] 传参 SkillUUID 类型错误", "CORE", "core_error"},
		{"PlayCardAck Result=1530", "None", "ack_error"},
	}
	for _, tt := range tests {
		got := classifyFailureKind(LogEntrySnapshot{Msg: tt.msg, Type: tt.typ})
		assert.Equal(t, tt.want, got, "msg=%q", tt.msg)
	}
}

func TestBuildFightTestReport_只收录失败日志(t *testing.T) {
	logs := map[string][]LogEntry{
		"TC-18": {
			{Case: "TC-18", ID: 1, Level: "INFO", Type: "ASSET", Msg: "PropValue判断正确"},
			{Case: "TC-18", ID: 3, Level: "ERROR", Type: "ASSET", Msg: "PropValue判断错误:2 实际值为:1"},
		},
		"TC-01": {
			{Case: "TC-01", ID: 1, Level: "INFO", Type: "ASSET", Msg: "ok"},
		},
	}

	report := buildFightTestReport(logs, "七进七出", "test", time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC))

	assert.Equal(t, 2, report.Summary.Total)
	assert.Equal(t, 1, report.Summary.Passed)
	assert.Equal(t, 1, report.Summary.Failed)
	require.Len(t, report.Cases, 2)

	var tc18 *CaseRunResult
	for i := range report.Cases {
		if report.Cases[i].CaseName == "TC-18" {
			tc18 = &report.Cases[i]
			break
		}
	}
	require.NotNil(t, tc18)
	assert.False(t, tc18.Passed)
	assert.Equal(t, 3, tc18.FirstFailStepID)
	assert.Equal(t, "asset_mismatch", tc18.FailureKind)
	require.Len(t, tc18.Errors, 1)
}

// TestIsFailureLogEntry_AttrChangeMaybeWrong 验证 AttrChange 值不匹配（assetMaybeWrong）
// 走 WARN 级别时仍被判定为失败。这是流水线场景的关键：robot 对 AttrChange 值不匹配
// 只打 WARN（继续等下一条），若不纳入失败判定会导致用例虚假通过。
func TestIsFailureLogEntry_AttrChangeMaybeWrong(t *testing.T) {
	tests := []struct {
		name    string
		entry   LogEntry
		want    bool
	}{
		{
			name:  "WARN+ASSET+判断不匹配 → 失败（assetMaybeWrong）",
			entry: LogEntry{Level: "WARN", Type: "ASSET", Msg: "ShaCount判断不匹配:999 实际值为: 1"},
			want:  true,
		},
		{
			name:  "ERROR+ASSET → 失败（assetWrong）",
			entry: LogEntry{Level: "ERROR", Type: "ASSET", Msg: "判断错误"},
			want:  true,
		},
		{
			name:  "INFO+ASSET → 不失败（assetOk）",
			entry: LogEntry{Level: "INFO", Type: "ASSET", Msg: "判断正确"},
			want:  false,
		},
		{
			name:  "WARN+非ASSET → 不失败（如 SeasonPassInfoNtf is Nil）",
			entry: LogEntry{Level: "WARN", Type: "OTHER", Msg: "*SeasonPassInfoNtf is Nil"},
			want:  false,
		},
		{
			name:  "WARN+ASSET+非判断不匹配 → 不失败",
			entry: LogEntry{Level: "WARN", Type: "ASSET", Msg: "某条普通警告"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isFailureLogEntry(tt.entry))
		})
	}
}

// TestBuildFightTestReport_AttrChangeMaybeWrong判失败 验证完整报告链路：
// 含 assetMaybeWrong（WARN）的用例应被判失败且 failureKind=asset_mismatch。
func TestBuildFightTestReport_AttrChangeMaybeWrong判失败(t *testing.T) {
	logs := map[string][]LogEntry{
		"TC-attr": {
			{Case: "TC-attr", ID: 3, Level: "WARN", Type: "ASSET", Msg: "HandLimit判断不匹配:999 实际值为: 3"},
		},
	}
	report := buildFightTestReport(logs, "job", "pfx", time.Now())
	require.Len(t, report.Cases, 1)
	assert.False(t, report.Cases[0].Passed, "含 MaybeWrong 的用例应判失败")
	assert.Equal(t, "asset_mismatch", report.Cases[0].FailureKind)
	assert.Equal(t, 3, report.Cases[0].FirstFailStepID)
}

func TestWriteFightTestReport(t *testing.T) {
	dir := t.TempDir()
	report := buildFightTestReport(map[string][]LogEntry{
		"TC-1": {{Case: "TC-1", ID: 1, Level: "ERROR", Type: "ASSET", Msg: "判断错误"}},
	}, "job", "pfx", time.Now())

	tsPath, latestPath, latestMD, err := writeFightTestReport(report, dir)
	require.NoError(t, err)
	assert.FileExists(t, tsPath)
	assert.FileExists(t, latestPath)
	assert.FileExists(t, latestMD)

	data, err := os.ReadFile(latestPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"caseName": "TC-1"`)

	md, err := os.ReadFile(latestMD)
	require.NoError(t, err)
	assert.Contains(t, string(md), "asset_mismatch")
	assert.Contains(t, string(md), "TC-1")
}

func TestResolveFightTestReportDir(t *testing.T) {
	casesDir := filepath.Join("cases", "fight_cases")
	got := resolveFightTestReportDir(&casesDir)
	assert.Equal(t, filepath.Join("cases", "fight_test_reports"), got)
}
