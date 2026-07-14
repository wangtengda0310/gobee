package cases

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveTestCaseToFile_ReqOnlyWithoutDirectionOffset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "case.json")

	rec := &Recording{
		Version:    RecordingVersion,
		RecordedAt: "2026-06-11T00:00:00Z",
		ServerAddr: "127.0.0.1:18000",
		Messages: []*RecordEntry{
			{
				OffsetMs:    100,
				MsgID:       1001,
				MsgName:     "GmCommandReq",
				SeqID:       1,
				Direction:   DirClientToServer,
				Descript:    "添加粮草",
				PayloadJSON: jsonRaw(`{"content":"//AddItem 1"}`),
			},
			{
				OffsetMs:    200,
				MsgID:       1002,
				MsgName:     "GmCommandAck",
				SeqID:       2,
				Direction:   "←",
				PayloadJSON: jsonRaw(`{}`),
			},
		},
	}

	require.NoError(t, SaveTestCaseToFile(path, rec))

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(raw)
	assert.NotContains(t, content, "direction")
	assert.NotContains(t, content, "offset_ms")
	assert.NotContains(t, content, "GmCommandAck")
	assert.Contains(t, content, "GmCommandReq")
	assert.Contains(t, content, "添加粮草")

	loaded, err := LoadTestCaseFromFile(path)
	require.NoError(t, err)
	require.Len(t, loaded.Messages, 1)
	assert.Equal(t, "GmCommandReq", loaded.Messages[0].MsgName)
	assert.Equal(t, DirClientToServer, loaded.Messages[0].Direction)
	assert.Equal(t, 0, loaded.Messages[0].OffsetMs)
}

func TestLoadTestCaseFromFile_LegacyFormatFiltersAck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")
	legacy := `{
  "version": 1,
  "recorded_at": "2026-06-08T07:03:51.429Z",
  "server_addr": "10.254.114.204:18000",
  "messages": [
    {
      "offset_ms": 4897,
      "msg_id": 1001,
      "msg_name": "GmCommandReq",
      "seq_id": 7,
      "direction": "→",
      "payload_json": {"content": "//AddItem 1"}
    },
    {
      "offset_ms": 4898,
      "msg_id": 1002,
      "msg_name": "GmCommandAck",
      "seq_id": 41,
      "direction": "←",
      "payload_json": {}
    }
  ]
}`
	require.NoError(t, os.WriteFile(path, []byte(legacy), 0644))

	loaded, err := LoadTestCaseFromFile(path)
	require.NoError(t, err)
	require.Len(t, loaded.Messages, 1)
	assert.Equal(t, "GmCommandReq", loaded.Messages[0].MsgName)
}

func TestAppendTestCaseToFile_AppendsReqToExistingCase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "case.json")

	// 第一次保存：包含 HelloReq
	first := &Recording{
		Version:    RecordingVersion,
		RecordedAt: "2026-06-11T00:00:00Z",
		ServerAddr: "127.0.0.1:18000",
		Messages: []*RecordEntry{
			{
				OffsetMs:    100,
				MsgID:       1001,
				MsgName:     "HelloReq",
				SeqID:       1,
				Direction:   DirClientToServer,
				Descript:    "hello",
				PayloadJSON: jsonRaw(`{"greeting":"hello"}`),
			},
		},
	}
	require.NoError(t, SaveTestCaseToFile(path, first))

	// 第二次追加：包含 LoginReq（以及应被过滤的 LoginAck）
	second := &Recording{
		Version:    RecordingVersion,
		RecordedAt: "2026-06-11T01:00:00Z",
		ServerAddr: "127.0.0.1:18001",
		Messages: []*RecordEntry{
			{
				OffsetMs:    200,
				MsgID:       2001,
				MsgName:     "LoginReq",
				SeqID:       2,
				Direction:   DirClientToServer,
				Descript:    "login",
				PayloadJSON: jsonRaw(`{"username":"test"}`),
			},
			{
				OffsetMs:    250,
				MsgID:       2002,
				MsgName:     "LoginAck",
				SeqID:       2,
				Direction:   DirServerToClient,
				PayloadJSON: jsonRaw(`{}`),
			},
		},
	}
	require.NoError(t, AppendTestCaseToFile(path, second))

	loaded, err := LoadTestCaseFromFile(path)
	require.NoError(t, err)
	require.Len(t, loaded.Messages, 2)
	assert.Equal(t, "HelloReq", loaded.Messages[0].MsgName)
	assert.Equal(t, "LoginReq", loaded.Messages[1].MsgName)

	// 追加后仍应只保留 Req，省略 direction / offset_ms
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(raw)
	assert.NotContains(t, content, "direction")
	assert.NotContains(t, content, "offset_ms")
	assert.NotContains(t, content, "LoginAck")
}

func jsonRaw(s string) json.RawMessage {
	return json.RawMessage(s)
}
