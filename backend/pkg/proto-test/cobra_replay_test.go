package prototest

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// newMockCmd 创建一个输出重定向到 buffer 的 cobra.Command，用于捕获 Println。
func newMockCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	return cmd, buf
}

func TestMakePrintAckCallback_DisabledReturnsNil(t *testing.T) {
	if cb := makePrintAckCallback(false, nil); cb != nil {
		t.Fatalf("enabled=false 应返回 nil，got %v", cb)
	}
}

func TestMakePrintAckCallback_FiltersClientToServer(t *testing.T) {
	cmd, buf := newMockCmd()
	cb := makePrintAckCallback(true, cmd)
	if cb == nil {
		t.Fatal("enabled=true 应返回非 nil 回调")
	}

	// 客户端→服务端的消息应被忽略
	cb("GmCommandReq", 1001, 5, `{"content":"//AddItem 1 1"}`, 0, protocol.DirClientToServer, "test1")
	if buf.Len() > 0 {
		t.Fatalf("客户端→服务端消息不应输出，got %q", buf.String())
	}
}

func TestMakePrintAckCallback_OutputsServerToClient(t *testing.T) {
	cmd, buf := newMockCmd()
	cb := makePrintAckCallback(true, cmd)

	// 服务端→客户端的 Ack 应输出为 NDJSON
	cb("GmCommandAck", 9001, 5, `{"code":0}`, 0, protocol.DirServerToClient, "test3")

	var entry struct {
		Account string         `json:"account"`
		MsgName string         `json:"msg_name"`
		MsgID   uint16         `json:"msg_id"`
		SeqID   uint32         `json:"seq_id"`
		Payload map[string]any `json:"payload"`
	}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("输出不是合法 JSON: %v, raw=%q", err, buf.String())
	}
	if entry.Account != "test3" {
		t.Errorf("account = %q, want test3", entry.Account)
	}
	if entry.MsgName != "GmCommandAck" {
		t.Errorf("msg_name = %q, want GmCommandAck", entry.MsgName)
	}
	if entry.MsgID != 9001 {
		t.Errorf("msg_id = %d, want 9001", entry.MsgID)
	}
	if entry.Payload["code"] != float64(0) {
		t.Errorf("payload.code = %v, want 0", entry.Payload["code"])
	}
}

func TestMakePrintAckCallback_EmptyPayload(t *testing.T) {
	cmd, buf := newMockCmd()
	cb := makePrintAckCallback(true, cmd)

	// payload 为空时不应崩溃，payload 字段为 null
	cb("HeartBeatNtf", 5000, 1, "", 0, protocol.DirServerToClient, "test1")

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("应有输出")
	}
	if !strings.Contains(line, `"payload":null`) {
		t.Errorf("空 payload 应为 null, got %q", line)
	}
}
