// Package cases 提供录制/用例的纯数据类型与文件 I/O。
//
// 职责：
//   - RecordEntry / Recording：录制数据结构
//   - RecordingVersion：录制文件版本号
//   - LoadRecordingFromFile / SaveRecordingToFile：文件读写
//   - TestCaseEntry / SaveTestCaseToFile / LoadTestCaseFromFile：用例读写与规范化
//
// 本包不依赖 streamproto（帧解码、连接管理），避免循环导入。
package cases

import (
	"encoding/json"
	"fmt"
	"os"
)

// RecordingVersion 录制文件版本号
// 修改 Recording 结构体时需同步递增此版本号，并更新 proto_cases/CLAUDE.md 中的说明
const RecordingVersion = 1

// DirClientToServer 客户端→服务端方向标记
const DirClientToServer = "→"

// DirServerToClient 服务端→客户端方向标记
const DirServerToClient = "←"

// RecordEntry 单条消息记录
type RecordEntry struct {
	OffsetMs    int             `json:"offset_ms"`
	MsgID       uint16          `json:"msg_id"`
	MsgName     string          `json:"msg_name"`
	SeqID       uint32          `json:"seq_id"`
	Direction   string          `json:"direction"`          // "→" 客户端→服务端, "←" 服务端→客户端
	Descript    string          `json:"descript,omitempty"` // 消息描述信息（用户可编辑）
	PayloadJSON json.RawMessage `json:"payload_json"`
	FieldValues map[string]any  `json:"field_values,omitempty"` // 字段路径→4态值映射（前端编辑状态）
}

// Recording 录制文件结构
type Recording struct {
	Version    int    `json:"version"`
	RecordedAt string `json:"recorded_at"`
	ServerAddr string `json:"server_addr"`
	// LoginPayloadB64 保存客户端 LoginReq 的解密后 payload（base64 编码）
	// 重放时替换其中的 Token 字段后重新加密发送
	LoginPayloadB64 string         `json:"login_payload_b64,omitempty"`
	Messages        []*RecordEntry `json:"messages"`
}

// LoadRecordingFromFile 从文件加载 Recording
func LoadRecordingFromFile(path string) (*Recording, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("解析录制文件失败: %v", err)
	}
	return &rec, nil
}

// SaveRecordingToFile 保存 Recording 到文件
func SaveRecordingToFile(path string, rec *Recording) error {
	out, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	return nil
}
