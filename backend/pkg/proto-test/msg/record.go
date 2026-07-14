package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/cases"
	proto "github.com/gogo/protobuf/proto"
)

// 纯数据类型已移至 cases 包，此处保留类型别名以兼容包内引用
type RecordEntry = cases.RecordEntry
type Recording = cases.Recording
type OnRecordCallback = func(msgCount int, latestMsg *RecordEntry)

const RecordingVersion = cases.RecordingVersion

// DirClientToServer / DirServerToClient 已在 decoder.go 中定义

var LoadRecordingFromFile = cases.LoadRecordingFromFile
var SaveRecordingToFile = cases.SaveRecordingToFile

// Recorder 录制器（运行时，依赖 DecodedFrame 等帧解码类型）
// 录制数据驻留内存，不自动落盘；由前端"保存为用例"流程通过 SaveRecordingToFile 手动持久化
type Recorder struct {
	mu              sync.Mutex
	serverAddr      string
	messages        []*RecordEntry
	startTime       time.Time
	started         bool
	loginPayloadB64 string
	onRecord        OnRecordCallback
}

func NewRecorder(serverAddr string) *Recorder {
	return &Recorder{
		serverAddr: serverAddr,
		messages:   make([]*RecordEntry, 0),
	}
}

func (r *Recorder) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.startTime = time.Now()
	r.started = true
}

func (r *Recorder) RecordFrame(frame *DecodedFrame, dir string) error {
	return r.recordFrame(frame, dir, true)
}

func (r *Recorder) RecordFrameSilent(frame *DecodedFrame, dir string) error {
	return r.recordFrame(frame, dir, false)
}

func (r *Recorder) recordFrame(frame *DecodedFrame, dir string, notify bool) error {
	if frame.MsgID < 1000 {
		return nil
	}
	msg, ok := NewMessage(frame.MsgID)
	if !ok {
		return fmt.Errorf("消息未注册: MsgID=%d", frame.MsgID)
	}
	var protoData []byte
	if len(frame.Payload) >= 2 {
		dataLen := int(frame.Payload[0]) | int(frame.Payload[1])<<8
		if dataLen <= len(frame.Payload)-2 {
			protoData = frame.Payload[2 : 2+dataLen]
		}
	}
	if len(protoData) == 0 {
		protoData = []byte{}
	}
	if err := proto.Unmarshal(protoData, msg); err != nil {
		return fmt.Errorf("反序列化失败 MsgID=%d: %v", frame.MsgID, err)
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("JSON 序列化失败 MsgID=%d: %v", frame.MsgID, err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.started {
		r.startTime = time.Now()
		r.started = true
	}
	if len(r.messages) == 0 {
		r.startTime = time.Now()
	}
	entry := &RecordEntry{
		OffsetMs:    int(time.Since(r.startTime).Milliseconds()),
		MsgID:       frame.MsgID,
		MsgName:     GetMsgName(frame.MsgID),
		SeqID:       frame.SeqID,
		Direction:   dir,
		PayloadJSON: jsonBytes,
	}
	r.messages = append(r.messages, entry)
	if notify && r.onRecord != nil {
		r.onRecord(len(r.messages), entry)
	}
	return nil
}

func (r *Recorder) SetOnRecord(cb OnRecordCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRecord = cb
}

func (r *Recorder) RecordLoginPayload(payload []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loginPayloadB64 = base64.StdEncoding.EncodeToString(payload)
}

func (r *Recorder) GetMessageCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.messages)
}

func (r *Recorder) UpdatePayloadBySeqID(seqID uint32, dir string, payloadJSON json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.messages {
		if entry.SeqID == seqID && entry.Direction == dir {
			entry.PayloadJSON = payloadJSON
			return
		}
	}
}

// ToRecording 导出当前内存中的录制数据为 Recording 结构
// 供"保存为用例"流程使用（前端 GUI 和 AI Agent 工具）
func (r *Recorder) ToRecording() *Recording {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Recording{
		Version:         RecordingVersion,
		RecordedAt:      r.startTime.Format(time.RFC3339),
		ServerAddr:      r.serverAddr,
		LoginPayloadB64: r.loginPayloadB64,
		Messages:        r.messages,
	}
}
