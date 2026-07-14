// L6 战斗测试结构化报告：RunRobotTest 结束后由 services.go 调用 persistFightTestReportAfterRun，
// 写入 {casesDir}/../fight_test_reports/latest.{md,json}（见 fight_test_report_test.go）。
// 与 services.go 的 hook 为同一功能，git 提交时须一并包含。
package functiontest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FightTestRunReport L6 执行结果结构化报告（供 skill 批量归因，避免从 GUI 手抄日志）。
type FightTestRunReport struct {
	FinishedAt string                 `json:"finishedAt"`
	JobDesc    string                 `json:"jobDesc,omitempty"`
	Prefix     string                 `json:"prefix,omitempty"`
	Summary    FightTestReportSummary `json:"summary"`
	Cases      []CaseRunResult        `json:"cases"`
}

// FightTestReportSummary 汇总统计。
type FightTestReportSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// CaseRunResult 单用例执行结果（失败时带首错步骤与错误列表）。
type CaseRunResult struct {
	CaseName        string             `json:"caseName"`
	Passed          bool               `json:"passed"`
	FirstFailStepID int                `json:"firstFailStepId,omitempty"`
	FailureKind     string             `json:"failureKind,omitempty"`
	Errors          []LogEntrySnapshot `json:"errors,omitempty"`
}

// LogEntrySnapshot 失败相关日志快照（不含 assetOk 等 INFO 噪音）。
type LogEntrySnapshot struct {
	StepID       int    `json:"stepId"`
	Level        string `json:"level"`
	Type         string `json:"type"`
	Msg          string `json:"msg"`
	CodeLocation string `json:"codeLocation,omitempty"`
}

// buildFightTestReport 从 GetTestLogs 结果构建 L6 报告。
func buildFightTestReport(logs map[string][]LogEntry, jobDesc, prefix string, finishedAt time.Time) FightTestRunReport {
	report := FightTestRunReport{
		FinishedAt: finishedAt.Format(time.RFC3339),
		JobDesc:    jobDesc,
		Prefix:     prefix,
		Cases:      make([]CaseRunResult, 0, len(logs)),
	}

	for caseName, entries := range logs {
		cr := CaseRunResult{CaseName: caseName, Passed: true}
		for _, e := range entries {
			if !isFailureLogEntry(e) {
				continue
			}
			cr.Passed = false
			if cr.FirstFailStepID == 0 || (e.ID > 0 && e.ID < cr.FirstFailStepID) {
				cr.FirstFailStepID = e.ID
			}
			cr.Errors = append(cr.Errors, LogEntrySnapshot{
				StepID:       e.ID,
				Level:        e.Level,
				Type:         e.Type,
				Msg:          e.Msg,
				CodeLocation: e.CodeLocation,
			})
		}
		if !cr.Passed && len(cr.Errors) > 0 {
			cr.FailureKind = classifyFailureKind(cr.Errors[0])
		}
		report.Cases = append(report.Cases, cr)
		if cr.Passed {
			report.Summary.Passed++
		} else {
			report.Summary.Failed++
		}
	}
	report.Summary.Total = len(report.Cases)
	sort.Slice(report.Cases, func(i, j int) bool {
		if report.Cases[i].Passed != report.Cases[j].Passed {
			return !report.Cases[i].Passed // 失败在前
		}
		return report.Cases[i].CaseName < report.Cases[j].CaseName
	})
	return report
}

func isFailureLogLevel(level string) bool {
	return level == "ERROR" || level == "FAIL"
}

// isFailureLogEntry 判定一条日志是否代表用例失败。
//
// 默认只认 ERROR/FAIL 级别。但 AttrChange 的值不匹配走 assetMaybeWrong
// （module_yanwu_qa_function.go），其日志级别是 WARN 而非 ERROR——因为 robot
// 会继续等待下一条同消息（最多 30s）期望匹配。若整个用例结束时该 asset 仍处于
// MaybeWrong（最终未匹配到正确值），它不会被升级为 ERROR，导致用例虚假通过。
//
// 因此额外把「WARN + Type=ASSET + Msg 含"判断不匹配"」纳入失败，
// 精准捕获 AttrChange 值断言错误，不误伤其他非断言 WARN（如 SeasonPassInfoNtf）。
func isFailureLogEntry(e LogEntry) bool {
	if isFailureLogLevel(e.Level) {
		return true
	}
	// assetMaybeWrong：值不匹配但 robot 容忍继续等。用例结束仍 MaybeWrong 即视为失败。
	if e.Level == "WARN" && e.Type == "ASSET" && strings.Contains(e.Msg, "判断不匹配") {
		return true
	}
	return false
}

// classifyFailureKind 按日志文案粗分类，便于 skill 按模式修文档而非逐条手抄。
func classifyFailureKind(entry LogEntrySnapshot) string {
	msg := entry.Msg
	typ := strings.ToUpper(entry.Type)

	switch {
	case strings.Contains(msg, "等待消息超时"):
		return "asset_timeout"
	case strings.Contains(msg, "等待消息错误"), strings.Contains(msg, "未匹配到的Asset"):
		return "asset_unmatched"
	case strings.Contains(msg, "判断错误"), strings.Contains(msg, "判断不匹配"):
		return "asset_mismatch"
	case typ == "CORE" || strings.Contains(msg, "[CORE]"):
		return "core_error"
	case strings.Contains(msg, "Result="), strings.Contains(msg, "ErrCode"):
		return "ack_error"
	default:
		return "other"
	}
}

// resolveFightTestReportDir 报告输出目录：优先 cases 同级 fight_test_reports。
func resolveFightTestReportDir(casesDir *string) string {
	if casesDir != nil && *casesDir != "" {
		return filepath.Join(*casesDir, "..", "fight_test_reports")
	}
	return filepath.Join("cases", "fight_test_reports")
}

// writeFightTestReport 写入 timestamp 文件与 latest.json / latest.md（覆盖）。
func writeFightTestReport(report FightTestRunReport, outputDir string) (timestampPath, latestJSONPath, latestMDPath string, err error) {
	if err = os.MkdirAll(outputDir, 0o755); err != nil {
		return "", "", "", fmt.Errorf("创建报告目录失败: %w", err)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", "", "", fmt.Errorf("序列化报告失败: %w", err)
	}

	md := formatFightTestReportMarkdown(report)

	ts := time.Now().Format("20060102_150405")
	safeDesc := sanitizeFileName(report.JobDesc)
	baseName := ts
	if safeDesc != "" {
		baseName = ts + "_" + safeDesc
	}
	timestampPath = filepath.Join(outputDir, baseName+".json")
	if err = os.WriteFile(timestampPath, data, 0o644); err != nil {
		return "", "", "", fmt.Errorf("写入报告失败: %w", err)
	}
	if err = os.WriteFile(strings.TrimSuffix(timestampPath, ".json")+".md", []byte(md), 0o644); err != nil {
		return timestampPath, "", "", fmt.Errorf("写入归档 Markdown 失败: %w", err)
	}

	latestJSONPath = filepath.Join(outputDir, "latest.json")
	if err = os.WriteFile(latestJSONPath, data, 0o644); err != nil {
		return timestampPath, "", "", fmt.Errorf("写入 latest.json 失败: %w", err)
	}
	latestMDPath = filepath.Join(outputDir, "latest.md")
	if err = os.WriteFile(latestMDPath, []byte(md), 0o644); err != nil {
		return timestampPath, latestJSONPath, "", fmt.Errorf("写入 latest.md 失败: %w", err)
	}
	return timestampPath, latestJSONPath, latestMDPath, nil
}

// formatFightTestReportMarkdown 生成供人工/AI 扫读的失败摘要表。
func formatFightTestReportMarkdown(report FightTestRunReport) string {
	var b strings.Builder
	b.WriteString("# L6 战斗测试报告\n\n")
	if report.JobDesc != "" {
		fmt.Fprintf(&b, "- **任务**: %s\n", report.JobDesc)
	}
	fmt.Fprintf(&b, "- **完成时间**: %s\n", report.FinishedAt)
	fmt.Fprintf(&b, "- **汇总**: 通过 %d / 失败 %d / 共 %d\n\n",
		report.Summary.Passed, report.Summary.Failed, report.Summary.Total)

	if report.Summary.Failed == 0 {
		b.WriteString("## 结果\n\n全部通过。\n")
		return b.String()
	}

	b.WriteString("## 失败用例\n\n")
	b.WriteString("| 用例 | stepId | failureKind | 首条错误 |\n")
	b.WriteString("|------|--------|-------------|----------|\n")
	for _, c := range report.Cases {
		if c.Passed {
			continue
		}
		firstMsg := ""
		if len(c.Errors) > 0 {
			firstMsg = escapeMarkdownTableCell(c.Errors[0].Msg)
			if len(firstMsg) > 120 {
				firstMsg = firstMsg[:120] + "…"
			}
		}
		fmt.Fprintf(&b, "| %s | %d | %s | %s |\n",
			escapeMarkdownTableCell(c.CaseName), c.FirstFailStepID, c.FailureKind, firstMsg)
	}

	b.WriteString("\n## 失败详情\n\n")
	for _, c := range report.Cases {
		if c.Passed {
			continue
		}
		fmt.Fprintf(&b, "### %s\n\n", c.CaseName)
		fmt.Fprintf(&b, "- stepId: %d\n", c.FirstFailStepID)
		fmt.Fprintf(&b, "- failureKind: `%s`\n\n", c.FailureKind)
		for _, e := range c.Errors {
			fmt.Fprintf(&b, "- `[step %d]` %s\n", e.StepID, e.Msg)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func escapeMarkdownTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func sanitizeFileName(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ', r == '，', r == '。', r == '|', r == '/':
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if len(out) > 40 {
		out = out[:40]
	}
	return out
}

// persistFightTestReportAfterRun RunRobotTest 结束后落盘 L6 报告；无日志时跳过。
func persistFightTestReportAfterRun(logs map[string][]LogEntry, jobDesc, prefix string, casesDir *string) {
	if len(logs) == 0 {
		return
	}
	report := buildFightTestReport(logs, jobDesc, prefix, time.Now())
	outputDir := resolveFightTestReportDir(casesDir)
	tsPath, latestJSON, latestMD, err := writeFightTestReport(report, outputDir)
	if err != nil {
		fmt.Println("战斗测试报告写入失败:", err)
		return
	}
	fmt.Printf("战斗测试报告: %s | latest.json: %s | latest.md: %s — 通过 %d / 失败 %d / 共 %d\n",
		tsPath, latestJSON, latestMD, report.Summary.Passed, report.Summary.Failed, report.Summary.Total)
}
