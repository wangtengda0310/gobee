package prototest

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

func isProtoCasePath(path string) bool {
	return strings.Contains(filepath.ToSlash(path), protoCaseDir)
}

// RecordFileService 录制文件加载/保存/Payload 编辑服务
// 对应前端 packet-tab.vue, testcase-tab.vue, replay-result-tab.vue 的文件操作
//
// 时序图（加载文件）：
// ┌──────────────────┐  LoadRecordFile   ┌───────────────────┐
// │ packet-tab.vue   │ ────────────────> │ RecordFileService │
// │ testcase-tab.vue │                   │   (后端)          │
// └──────────────────┘                   └───────────────────┘
//
//	     │                                        │
//	     │                                        │ LoadRecordingFromFile()
//	     │                                        ▼
//	     │                                  ┌──────────┐
//	     │                                  │streamproto│
//	     │                                  │  解析JSON │
//	     │                                  │recordingToViews()
//	     │                                  └──────────┘
//	     │                                        │
//	     │        返回 RecordFileData             │
//	     │ <──────────────────────────────────────┘
//	     │                                        │
//	     ▼                                        ▼
//	设置到组件状态                      解析录制文件
//
// 时序图（Payload 编辑）：
// ┌──────────────────────┐  UpdateMessagePayload  ┌───────────────────┐
// │ paired-payload-editor│ ──────────────────────> │ RecordFileService │
// └──────────────────────┘                        │   (后端)          │
//
//	     │                                            └───────────────────┘
//	     │                                                  │
//	     │                                                  │ 加载录制文件
//	     │                                                  │ 更新 PayloadJSON
//	     │                                                  │ 保存回文件
//	     │                                                  │
//	     │            保存成功                                │
//	     │ <─────────────────────────────────────────────────┘
//	     │
//	     ▼
//	刷新表格数据
type RecordFileService struct{}

// viewsToEntries 将 RecordEntryView 切片转换为 RecordEntry 切片
// 辅助函数：用于将前端编辑后的数据保存回录制文件
func viewsToEntries(msgs []RecordEntryView) []*protocol.RecordEntry {
	entries := make([]*protocol.RecordEntry, len(msgs))
	for i, m := range msgs {
		// 转换 FieldValues: map[string]FieldValues → map[string]any
		fieldValues := make(map[string]any)
		for k, v := range m.FieldValues {
			fieldValues[k] = v
		}

		entries[i] = &protocol.RecordEntry{
			OffsetMs:    m.OffsetMs,
			MsgID:       m.MsgID,
			MsgName:     m.MsgName,
			SeqID:       m.SeqID,
			Direction:   m.Direction,
			Descript:    m.Descript,
			PayloadJSON: payloadToRawMessage(m.Payload),
			FieldValues: fieldValues,
		}
	}
	return entries
}

// payloadToRawMessage 将 map[string]any 序列化为 json.RawMessage
func payloadToRawMessage(p map[string]any) json.RawMessage {
	if p == nil {
		return json.RawMessage("{}")
	}
	b, _ := json.Marshal(p)
	return json.RawMessage(b)
}

// NewRecordFileService 创建录制文件服务实例
func NewRecordFileService() *RecordFileService {
	return &RecordFileService{}
}

// LoadRecordFile 加载录制文件，返回消息列表
func (s *RecordFileService) LoadRecordFile(path string) (*RecordFileData, error) {
	var rec *protocol.Recording
	var err error
	if isProtoCasePath(path) {
		rec, err = LoadTestCaseFromFile(path)
	} else {
		rec, err = protocol.LoadRecordingFromFile(path)
	}
	if err != nil {
		return nil, err
	}
	views := recordingToViews(rec)
	return &RecordFileData{
		Version:      rec.Version,
		RecordedAt:   rec.RecordedAt,
		ServerAddr:   rec.ServerAddr,
		MessageCount: len(views),
		Messages:     views,
	}, nil
}

// SaveRecordFile 将修改后的录制数据保存到文件
func (s *RecordFileService) SaveRecordFile(path string, data *RecordFileData) error {
	rec := &protocol.Recording{
		Version:    data.Version,
		RecordedAt: data.RecordedAt,
		ServerAddr: data.ServerAddr,
		Messages:   viewsToEntries(data.Messages),
	}
	if isProtoCasePath(path) {
		return SaveTestCaseToFile(path, rec)
	}
	// 录制文件保留 login_payload_b64
	if orig, err := protocol.LoadRecordingFromFile(path); err == nil {
		rec.LoginPayloadB64 = orig.LoginPayloadB64
	}
	return protocol.SaveRecordingToFile(path, rec)
}

// UpdateMessagePayload 更新指定 index 消息的 payload（前端直接传 map[string]any 对象）
func (s *RecordFileService) UpdateMessagePayload(path string, index int, payload map[string]any) (*RecordFileData, error) {
	data, err := s.LoadRecordFile(path)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(data.Messages) {
		return nil, fmt.Errorf("消息索引越界: %d, 总数: %d", index, len(data.Messages))
	}
	data.Messages[index].Payload = payload
	if err := s.SaveRecordFile(path, data); err != nil {
		return nil, err
	}
	return s.LoadRecordFile(path)
}

// UpdateMessageFieldValues 更新指定 index 消息的字段4态值
func (s *RecordFileService) UpdateMessageFieldValues(path string, index int, fieldValues map[string]FieldValues) (*RecordFileData, error) {
	data, err := s.LoadRecordFile(path)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(data.Messages) {
		return nil, fmt.Errorf("消息索引越界: %d, 总数: %d", index, len(data.Messages))
	}
	data.Messages[index].FieldValues = fieldValues
	if err := s.SaveRecordFile(path, data); err != nil {
		return nil, err
	}
	return s.LoadRecordFile(path)
}

// UpdateMessageDescript 更新指定 index 消息的描述信息
func (s *RecordFileService) UpdateMessageDescript(path string, index int, descript string) (*RecordFileData, error) {
	data, err := s.LoadRecordFile(path)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(data.Messages) {
		return nil, fmt.Errorf("消息索引越界: %d, 总数: %d", index, len(data.Messages))
	}
	data.Messages[index].Descript = descript
	if err := s.SaveRecordFile(path, data); err != nil {
		return nil, err
	}
	return s.LoadRecordFile(path)
}
