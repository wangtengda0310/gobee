package cases

import (
	"encoding/json"
	"fmt"
	"os"
)

// TestCaseEntry 测试用例单条消息（仅 Req；不持久化 direction / offset_ms）
type TestCaseEntry struct {
	MsgID       uint16          `json:"msg_id"`
	MsgName     string          `json:"msg_name"`
	SeqID       uint32          `json:"seq_id"`
	Descript    string          `json:"descript,omitempty"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	FieldValues map[string]any  `json:"field_values,omitempty"`
}

// IsReqDirection 是否为客户端请求方向。空 direction 视为 Req（兼容精简用例格式）。
func IsReqDirection(dir string) bool {
	return dir == "" || dir == DirClientToServer || dir == "→"
}

func recordEntryToTestCaseEntry(e *RecordEntry) *TestCaseEntry {
	if e == nil || !IsReqDirection(e.Direction) {
		return nil
	}
	return &TestCaseEntry{
		MsgID:       e.MsgID,
		MsgName:     e.MsgName,
		SeqID:       e.SeqID,
		Descript:    e.Descript,
		PayloadJSON: e.PayloadJSON,
		FieldValues: e.FieldValues,
	}
}

// NormalizeTestCaseRecording 加载用例后规范化：只保留 Req，补全 direction 供重放使用。
func NormalizeTestCaseRecording(rec *Recording) {
	if rec == nil {
		return
	}
	msgs := make([]*RecordEntry, 0, len(rec.Messages))
	for _, e := range rec.Messages {
		if e == nil || !IsReqDirection(e.Direction) {
			continue
		}
		if e.Direction == "" {
			e.Direction = DirClientToServer
		}
		e.OffsetMs = 0
		msgs = append(msgs, e)
	}
	rec.Messages = msgs
}

type testCaseFile struct {
	Version    int              `json:"version"`
	RecordedAt string           `json:"recorded_at"`
	ServerAddr string           `json:"server_addr"`
	Messages   []*TestCaseEntry `json:"messages"`
}

// SaveTestCaseToFile 保存测试用例文件（仅 Req，省略 direction / offset_ms）
func SaveTestCaseToFile(path string, rec *Recording) error {
	if rec == nil {
		return fmt.Errorf("录制数据不能为空")
	}
	tc := buildTestCaseFile(rec, nil)
	return writeTestCaseFile(path, tc)
}

// AppendTestCaseToFile 向已存在用例追加 Req 消息。
// 若文件不存在则等效于 SaveTestCaseToFile；存在则保留原文件元信息，仅追加新 Req。
func AppendTestCaseToFile(path string, rec *Recording) error {
	if rec == nil {
		return fmt.Errorf("录制数据不能为空")
	}

	var existing []*TestCaseEntry
	tc := buildTestCaseFile(rec, nil)
	if _, err := os.Stat(path); err == nil {
		loaded, err := LoadTestCaseFromFile(path)
		if err != nil {
			return fmt.Errorf("加载已存在用例失败: %v", err)
		}
		for _, e := range loaded.Messages {
			existing = append(existing, recordEntryToTestCaseEntry(e))
		}
	}

	// 合并：已存在消息在前，新消息在后
	tc.Messages = append(existing, tc.Messages...)
	return writeTestCaseFile(path, tc)
}

func buildTestCaseFile(rec *Recording, existing []*TestCaseEntry) testCaseFile {
	tc := testCaseFile{
		Version:    rec.Version,
		RecordedAt: rec.RecordedAt,
		ServerAddr: rec.ServerAddr,
		Messages:   make([]*TestCaseEntry, 0, len(rec.Messages)+len(existing)),
	}
	if tc.Version == 0 {
		tc.Version = RecordingVersion
	}
	if existing != nil {
		tc.Messages = append(tc.Messages, existing...)
	}
	for _, e := range rec.Messages {
		if entry := recordEntryToTestCaseEntry(e); entry != nil {
			tc.Messages = append(tc.Messages, entry)
		}
	}
	return tc
}

func writeTestCaseFile(path string, tc testCaseFile) error {
	out, err := json.MarshalIndent(tc, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("写入用例文件失败: %v", err)
	}
	return nil
}

// LoadTestCaseFromFile 加载测试用例并规范化为 Recording（兼容旧格式含 Ack/Ntf）
func LoadTestCaseFromFile(path string) (*Recording, error) {
	rec, err := LoadRecordingFromFile(path)
	if err != nil {
		return nil, err
	}
	NormalizeTestCaseRecording(rec)
	return rec, nil
}
